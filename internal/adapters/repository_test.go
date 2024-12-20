package adapters_test

import (
	"github.com/Gregmus2/sync-service/internal/adapters"
	"testing"
	"time"

	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/eatonphil/gosqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_UpdateDeviceTokenTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	t.Run("inserts new device token", func(t *testing.T) {
		err := repo.UpdateDeviceTokenTime("device1", "user1", "group1")
		require.NoError(t, err)

		var lastSync int
		stmt, err := db.Prepare("SELECT last_sync FROM device_tokens WHERE device_token = ?", "device1")
		require.NoError(t, err)
		defer stmt.Close()

		hasRow, err := stmt.Step()
		require.NoError(t, err)
		require.True(t, hasRow)

		err = stmt.Scan(&lastSync)
		require.NoError(t, err)
		assert.NotZero(t, lastSync)
	})

	t.Run("updates existing device token", func(t *testing.T) {
		err := repo.UpdateDeviceTokenTime("device1", "user1", "group1")
		require.NoError(t, err)

		var lastSync int
		stmt, err := db.Prepare("SELECT last_sync FROM device_tokens WHERE device_token = ?", "device1")
		require.NoError(t, err)
		defer stmt.Close()

		hasRow, err := stmt.Step()
		require.NoError(t, err)
		require.True(t, hasRow)

		err = stmt.Scan(&lastSync)
		require.NoError(t, err)
		assert.NotZero(t, lastSync)

		time.Sleep(time.Second)

		// Update again
		err = repo.UpdateDeviceTokenTime("device1", "user1", "group1")
		require.NoError(t, err)

		var updatedLastSync int
		stmt, err = db.Prepare("SELECT last_sync FROM device_tokens WHERE device_token = ?", "device1")
		require.NoError(t, err)
		defer stmt.Close()

		hasRow, err = stmt.Step()
		require.NoError(t, err)
		require.True(t, hasRow)

		err = stmt.Scan(&updatedLastSync)
		require.NoError(t, err)
		assert.Greater(t, updatedLastSync, lastSync)
	})
}

func TestRepository_InsertData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	t.Run("inserts data successfully", func(t *testing.T) {
		operations := []*proto.Operation{
			{
				Type: proto.Operation_OPERATION_CREATE,
				Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
				RelatedEntities: []*proto.Operation_Entity{
					{Id: "1", Name: "table1"},
				},
			},
		}

		err := repo.InsertData("device1", "group1", operations)
		require.NoError(t, err)

		var count int
		stmt, err := db.Prepare("SELECT COUNT(*) FROM operations")
		require.NoError(t, err)
		defer stmt.Close()

		hasRow, err := stmt.Step()
		require.NoError(t, err)
		require.True(t, hasRow)

		err = stmt.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestRepository_GetGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	t.Run("returns user id if group id is not set", func(t *testing.T) {
		groupID, err := repo.GetGroupID("user1")
		require.NoError(t, err)
		assert.Equal(t, "user1", groupID)
	})

	t.Run("returns group id if set", func(t *testing.T) {
		err := repo.UpdateDeviceTokenTime("device1", "user1", "group1")
		require.NoError(t, err)

		groupID, err := repo.GetGroupID("user1")
		require.NoError(t, err)
		assert.Equal(t, "group1", groupID)
	})
}

func TestRepository_GetData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	err = repo.UpdateDeviceTokenTime("device1", "user1", "group1")
	require.NoError(t, err)

	time.Sleep(time.Second)

	// Insert data for different devices
	operationsDevice2 := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_CREATE,
			Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device2", "group1", operationsDevice2)
	require.NoError(t, err)

	operationsDevice3 := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_UPDATE,
			Sql:  "UPDATE table1 SET name = 'updated' WHERE id = 1",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device3", "group1", operationsDevice3)
	require.NoError(t, err)

	// Get data for device1
	data, err := repo.GetData("device1", "group1")
	require.NoError(t, err)

	// Verify that data from device2 and device3 is returned
	assert.Len(t, data, 2)
	assert.Equal(t, proto.Operation_OPERATION_CREATE.String(), data[0].Type.String())
	assert.Equal(t, proto.Operation_OPERATION_UPDATE.String(), data[1].Type.String())
}

func TestRepository_UpdateGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	err = repo.UpdateDeviceTokenTime("device1", "user1", "group1")
	require.NoError(t, err)

	// Update group id
	err = repo.UpdateGroupID("user1", "newGroup")
	require.NoError(t, err)

	// Verify that group id was updated
	var groupID string
	stmt, err := db.Prepare("SELECT group_id FROM device_tokens WHERE user_id = ?", "user1")
	require.NoError(t, err)
	defer stmt.Close()

	hasRow, err := stmt.Step()
	require.NoError(t, err)
	require.True(t, hasRow)

	err = stmt.Scan(&groupID)
	require.NoError(t, err)
	assert.Equal(t, "newGroup", groupID)
}

func TestRepository_MigrateData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	// Insert data for group1
	operations := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_CREATE,
			Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device1", "group1", operations)
	require.NoError(t, err)

	// Migrate data from group1 to group2
	err = repo.MigrateData("group1", "group2")
	require.NoError(t, err)

	// Verify that data was migrated to group2
	var count int
	stmt, err := db.Prepare("SELECT COUNT(*) FROM operations WHERE group_id = ?", "group2")
	require.NoError(t, err)
	defer stmt.Close()

	hasRow, err := stmt.Step()
	require.NoError(t, err)
	require.True(t, hasRow)

	err = stmt.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRepository_RemoveData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	// Insert data for user1
	operations := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_CREATE,
			Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device1", "user1", operations)
	require.NoError(t, err)

	// Remove data for user1
	err = repo.RemoveData("user1")
	require.NoError(t, err)

	// Verify that data was removed
	var count int
	stmt, err := db.Prepare("SELECT COUNT(*) FROM operations WHERE group_id = ?", "user1")
	require.NoError(t, err)
	defer stmt.Close()

	hasRow, err := stmt.Step()
	require.NoError(t, err)
	require.True(t, hasRow)

	err = stmt.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRepository_GetAllData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	// Insert data for group1
	operations := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_CREATE,
			Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device1", "group1", operations)
	require.NoError(t, err)

	// Get all data for group1
	data, err := repo.GetAllData("group1")
	require.NoError(t, err)

	// Verify that data was returned
	assert.Len(t, data, 1)
	assert.Equal(t, proto.Operation_OPERATION_CREATE.String(), data[0].Type.String())
}

func TestRepository_CopyOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo, err := adapters.NewRepository(db)
	require.NoError(t, err)

	// Insert data for group1
	operations := []*proto.Operation{
		{
			Type: proto.Operation_OPERATION_CREATE,
			Sql:  "INSERT INTO table1 (id, name) VALUES (1, 'test')",
			RelatedEntities: []*proto.Operation_Entity{
				{Id: "1", Name: "table1"},
			},
		},
	}
	err = repo.InsertData("device1", "group1", operations)
	require.NoError(t, err)

	// Copy operations from group1 to group2
	err = repo.CopyOperations("group1", "group2")
	require.NoError(t, err)

	// Verify that operations were copied to group2
	var count int
	stmt, err := db.Prepare("SELECT COUNT(*) FROM operations WHERE group_id = ?", "group2")
	require.NoError(t, err)
	defer stmt.Close()

	hasRow, err := stmt.Step()
	require.NoError(t, err)
	require.True(t, hasRow)

	err = stmt.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func setupTestDB(t *testing.T) *gosqlite.Conn {
	t.Helper()

	db, err := gosqlite.Open(":memory:")
	require.NoError(t, err)

	return db
}
