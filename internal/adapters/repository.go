package adapters

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/eatonphil/gosqlite"
	"github.com/pkg/errors"
	"time"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS device_tokens
		(
			device_token TEXT PRIMARY KEY NOT NULL,
			user_id	     TEXT,
			group_id     TEXT,
			last_sync    INTEGER
		) WITHOUT ROWID;`,
	`CREATE TABLE IF NOT EXISTS operations
		(
			id             INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			device_token   TEXT                              NOT NULL,
			group_id       TEXT                              NOT NULL,
			operation_type TEXT                              NOT NULL,
			sql            TEXT                              NOT NULL,
			created_at     INTEGER                           NOT NULL
		);`,
	`CREATE TABLE IF NOT EXISTS related_entities
		(
			operation_id INTEGER NOT NULL,
			entity_id    TEXT    NOT NULL,
			entity_name  TEXT    NOT NULL,
			PRIMARY KEY (operation_id, entity_id, entity_name),
			FOREIGN KEY (operation_id) REFERENCES operations (id)
		) WITHOUT ROWID;`,
	`CREATE INDEX operations_group_id_idx ON operations(group_id, created_at);`,
	`CREATE INDEX operations_conflicts_query_idx ON operations(operation_type, group_id, id);`,
}

type repository struct {
	db *gosqlite.Conn
}

func NewRepository(db *gosqlite.Conn) (Repository, error) {
	if err := migrate(db); err != nil {
		return nil, errors.Wrap(err, "failed to run migrations")
	}

	return &repository{
		db: db,
	}, nil
}

func NewDB() (*gosqlite.Conn, error) {
	conn, err := gosqlite.Open("main.db")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open db")
	}

	conn.BusyTimeout(5 * time.Second)

	return conn, nil
}

func migrate(db *gosqlite.Conn) error {
	err := db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return errors.Wrap(err, "failed to enable foreign keys")
	}
	err = db.Exec(`PRAGMA journal_mode=WAL;`)
	if err != nil {
		return errors.Wrap(err, "failed to enable WAL logs")
	}
	err = db.Exec(`PRAGMA synchronous = NORMAL`)
	if err != nil {
		return errors.Wrap(err, "failed to update synchronous mod")
	}

	var version int
	stmt, err := db.Prepare(`SELECT version FROM migrations`)
	if err != nil && err.Error() == "sqlite3: no such table: migrations [1]" {
		err = db.Exec(`CREATE TABLE migrations (version INTEGER)`)
		if err != nil {
			return errors.Wrap(err, "failed to create migrations table")
		}
		err = db.Exec(`INSERT INTO migrations (version) VALUES (0)`)
		if err != nil {
			return errors.Wrap(err, "failed to create migrations table")
		}
		stmt, err = db.Prepare(`SELECT version FROM migrations`)
	}
	if err != nil {
		return errors.Wrap(err, "failed to prepare migrations select")
	}
	defer stmt.Close()

	hasRow, err := stmt.Step()
	if err != nil {
		return errors.Wrap(err, "failed to step migrations select")
	}
	if hasRow {
		err = stmt.Scan(&version)
		if err != nil {
			return errors.Wrap(err, "failed to scan migrations select")
		}
	}

	for i := version; i < len(migrations); i++ {
		err := db.Exec(migrations[i])
		if err != nil {
			return errors.Wrapf(err, "failed to run migration %d", i)
		}
	}

	err = db.Exec("UPDATE migrations SET version = ?", len(migrations))
	if err != nil {
		return errors.Wrap(err, "failed to update migrations version")
	}

	return nil
}

func (r repository) UpdateDeviceTokenTime(deviceToken, userID, groupID string) error {
	err := r.db.Exec(
		`INSERT INTO device_tokens(device_token, user_id, group_id, last_sync) VALUES(?, ?) 
				ON CONFLICT(device_token) DO UPDATE SET last_sync=excluded.last_sync;`,
		deviceToken, userID, groupID, time.Now().Unix())
	if err != nil {
		return errors.Wrap(err, "failed to update device token time")
	}

	return nil
}

func (r repository) InsertData(deviceToken, groupID string, operations []*proto.Operation) error {
	err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer r.db.Rollback()

	operationsStmt, err := r.db.Prepare(
		`INSERT INTO operations(device_token, group_id, operation_type, sql, created_at) 
				VALUES(?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return errors.Wrap(err, "failed to prepare operations insert statement")
	}
	defer operationsStmt.Close()

	relatedEntitiesStmt, err := r.db.Prepare(
		`INSERT INTO related_entities (operation_id, entity_id, entity_name) 
				VALUES (last_insert_rowid(), ?, ?);`,
	)
	if err != nil {
		return errors.Wrap(err, "failed to prepare related_entities insert statement")
	}
	defer relatedEntitiesStmt.Close()

	for _, op := range operations {
		err = operationsStmt.Exec(deviceToken, groupID, op.Type.String(), op.Sql, time.Now().Unix())
		if err != nil {
			return errors.Wrap(err, "failed to insert data")
		}

		for _, entity := range op.RelatedEntities {
			err = relatedEntitiesStmt.Exec(entity.Id, entity.Name)
			if err != nil {
				return errors.Wrap(err, "failed to insert related entities")
			}
		}
	}

	err = r.db.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r repository) CleanConflicted(deviceToken, groupID string) error {
	err := r.db.Exec(
		`DELETE FROM operations
				WHERE group_id = ?
				  AND created_at > coalesce((SELECT last_sync
											 FROM device_tokens
											 WHERE device_token = ?), 0)
				  AND EXISTS (SELECT null
							  FROM operations AS op2
									   JOIN related_entities re2 on op2.id = re2.operation_id
									   JOIN related_entities re on operations.id = re.operation_id
							  WHERE op2.operation_type = 'DELETE'
								AND re2.entity_name = re.entity_name
								AND re2.entity_id = re.entity_id
								AND op2.group_id = operations.group_id
								AND op2.id < operations.id);`,
		groupID, deviceToken,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete conflicts")
	}

	return nil
}

func (r repository) GetGroupID(userID string) (string, error) {
	stmt, err := r.db.Prepare(`SELECT coalesce(group_id, user_id) FROM device_tokens`)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare select group id")
	}
	defer stmt.Close()

	hasRow, err := stmt.Step()
	if err != nil {
		return "", errors.Wrap(err, "failed to step select group id")
	}

	if !hasRow {
		// return user id if group id is not set to keep user in own group
		return userID, nil
	}

	var groupID string
	err = stmt.Scan(&groupID)
	if err != nil {
		return "", errors.Wrap(err, "failed to scan select group id")
	}

	return groupID, nil
}

func (r repository) GetData(deviceToken, groupID string) ([]*proto.Operation, error) {
	stmt, err := r.db.Prepare(
		`SELECT id, operation_type, sql, entity_id, entity_name
				FROM operations 
				JOIN related_entities ON operations.id = related_entities.operation_id
				WHERE group_id = ? and 
				      device_token != ? and 
				      created_at > (SELECT last_sync FROM device_tokens WHERE device_token = ?)`,
		groupID, deviceToken, deviceToken,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare select data")
	}
	defer stmt.Close()

	rows := make([]*proto.Operation, 0)
	var lastOperation *proto.Operation
	var lastID int
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, errors.Wrap(err, "failed to step select data")
		}
		if !hasRow {
			break
		}

		var id int
		var opType, sql, entityID, entityName string
		err = stmt.Scan(&id, &opType, &sql, &entityID, &entityName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan select data")
		}

		if lastOperation == nil || lastID != id {
			lastOperation = &proto.Operation{
				Type:            proto.Operation_OperationType(proto.Operation_OperationType_value[opType]),
				Sql:             sql,
				RelatedEntities: make([]*proto.Operation_Entity, 0),
			}
			lastID = id
			rows = append(rows, lastOperation)
		}

		lastOperation.RelatedEntities = append(lastOperation.RelatedEntities, &proto.Operation_Entity{
			Id:   entityID,
			Name: entityName,
		})
	}

	return rows, nil
}
