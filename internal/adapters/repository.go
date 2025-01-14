package adapters

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/common"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type repository struct {
	client *gorm.DB
}

func NewDB(cfg *common.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseFQDN), &gorm.Config{
		FullSaveAssociations: true,
		Logger:               logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewRepository(db *gorm.DB) (Repository, error) {
	return &repository{
		client: db,
	}, nil
}

func (r repository) UpdateDeviceTokenTime(deviceToken, userID, groupID string) error {
	err := r.client.Exec(
		`INSERT INTO device_tokens(device_token, user_id, group_id, last_sync) VALUES(?, ?, ?, ?) 
				ON CONFLICT(device_token) DO UPDATE SET last_sync=excluded.last_sync;`,
		deviceToken, userID, groupID, time.Now().Unix()).Error
	if err != nil {
		return errors.Wrap(err, "failed to update device token time")
	}

	return nil
}

func (r repository) InsertData(deviceToken, groupID string, op *proto.Operation) error {
	return r.client.Transaction(func(tx *gorm.DB) error {
		operation := &common.Operation{
			DeviceToken:   deviceToken,
			GroupId:       groupID,
			OperationType: op.Type.String(),
			Sql:           op.Sql,
			Args:          op.Args,
			CreatedAt:     time.Now().Unix(),
		}
		err := tx.Create(operation).Error
		if err != nil {
			return errors.Wrap(err, "failed to insert data")
		}

		for _, entity := range op.RelatedEntities {
			err = tx.Exec(`INSERT INTO related_entities (operation_id, entity_id, entity_name) 
				VALUES (?, ?, ?);`, operation.ID, entity.Id, entity.Name).Error
			if err != nil {
				return errors.Wrap(err, "failed to insert related entities")
			}
		}

		return nil
	})
}

func (r repository) CleanConflicted(deviceToken, groupID string) error {
	err := r.client.Exec(
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
	).Error
	if err != nil {
		return errors.Wrap(err, "failed to delete conflicts")
	}

	return nil
}

func (r repository) GetGroupID(deviceToken, userID string) (string, error) {
	var groupID string
	err := r.client.Raw(`SELECT CASE WHEN group_id = '' THEN user_id ELSE group_id END FROM device_tokens 
		WHERE device_token = ? LIMIT 1`, deviceToken).Scan(&groupID).Error
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare select group id")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// return user id if group id is not set to keep user in own group
		return userID, nil
	}

	return groupID, nil
}

type data struct {
	ID            int
	OperationType string
	Sql           string
	Args          string
	EntityID      string
	EntityName    string
}

func (r repository) GetData(deviceToken, groupID string) ([]*proto.Operation, error) {
	return r.queryData(r.client.Raw(
		`SELECT id, operation_type, sql, args, entity_id, entity_name
				FROM operations 
				JOIN related_entities ON operations.id = related_entities.operation_id
				WHERE group_id = ? and 
				      device_token != ? and 
				      created_at > (SELECT last_sync FROM device_tokens WHERE device_token = ?)`,
		groupID, deviceToken, deviceToken,
	))
}

func (r repository) queryData(tx *gorm.DB) ([]*proto.Operation, error) {
	operations := make([]*data, 0)
	err := tx.Scan(&operations).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare select data")
	}

	rows := make([]*proto.Operation, 0)
	var lastOperation *proto.Operation
	var lastID int
	for _, op := range operations {
		if lastOperation == nil || lastID != op.ID {
			lastOperation = &proto.Operation{
				Type:            proto.Operation_OperationType(proto.Operation_OperationType_value[op.OperationType]),
				Sql:             op.Sql,
				Args:            op.Args,
				RelatedEntities: make([]*proto.Operation_Entity, 0),
			}
			lastID = op.ID
			rows = append(rows, lastOperation)
		}

		lastOperation.RelatedEntities = append(lastOperation.RelatedEntities, &proto.Operation_Entity{
			Id:   op.EntityID,
			Name: op.EntityName,
		})
	}

	return rows, nil
}

func (r repository) UpdateGroupID(userID, newGroupID string) error {
	err := r.client.Exec(
		`UPDATE device_tokens SET group_id = ? WHERE user_id = ?`, newGroupID, userID,
	).Error
	if err != nil {
		return errors.Wrap(err, "failed to update group id")
	}

	return nil
}

func (r repository) MigrateData(fromID, toID string) error {
	err := r.client.Exec(
		`UPDATE operations SET group_id = ?, created_at = unixepoch() WHERE group_id = ?`, toID, fromID,
	).Error
	if err != nil {
		return errors.Wrap(err, "failed to migrate data")
	}

	return nil
}

func (r repository) RemoveData(userID string) error {
	err := r.client.Exec(
		`DELETE FROM operations WHERE group_id = ?`, userID,
	).Error
	if err != nil {
		return errors.Wrap(err, "failed to remove data")
	}

	return nil
}

func (r repository) GetAllData(groupID string) ([]*proto.Operation, error) {
	return r.queryData(r.client.Raw(
		`SELECT id, operation_type, sql, args, entity_id, entity_name
				FROM operations 
				JOIN related_entities ON operations.id = related_entities.operation_id
				WHERE group_id = ?`,
		groupID,
	))
}

func (r repository) CopyOperations(fromID, toID string) error {
	return r.client.Transaction(func(tx *gorm.DB) error {
		operations := make([]common.Operation, 0)
		err := r.client.Raw(
			`SELECT id, device_token, operation_type, sql, args, created_at
				FROM operations
				WHERE group_id = ?`,
			fromID,
		).Scan(&operations).Error
		if err != nil {
			return errors.Wrap(err, "failed to prepare select data")
		}

		for _, op := range operations {
			operation := &common.Operation{
				DeviceToken:   op.DeviceToken,
				GroupId:       toID,
				OperationType: op.OperationType,
				Sql:           op.Sql,
				Args:          op.Args,
				CreatedAt:     op.CreatedAt,
			}
			err = tx.Create(operation).Error
			if err != nil {
				return errors.Wrap(err, "failed to insert data")
			}

			err = r.client.Exec(
				`INSERT INTO related_entities (operation_id, entity_id, entity_name) 
				SELECT ?, entity_id, entity_name FROM related_entities
					WHERE operation_id = ?;`, operation.ID, op.ID,
			).Error
			if err != nil {
				return errors.Wrap(err, "failed to insert related entities")
			}
		}

		return nil
	})
}

func (r repository) GetCurrentGroup(userID string) (*string, error) {
	var groupID string
	err := r.client.Raw(`SELECT group_id FROM device_tokens WHERE user_id = ? LIMIT 1`, userID).
		Scan(&groupID).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare select group id")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &groupID, nil
}
