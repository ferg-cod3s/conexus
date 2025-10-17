package connectors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	// Test that we can list connectors (should be empty)
	connectors, err := store.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, connectors)
}

func TestStore_Add(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	connector := &Connector{
		ID:     "test-connector",
		Name:   "Test Connector",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}

	// Test successful add
	err = store.Add(ctx, connector)
	assert.NoError(t, err)

	// Test duplicate add fails
	err = store.Add(ctx, connector)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test validation - empty ID
	connector2 := &Connector{
		Name:   "Test Connector 2",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}
	err = store.Add(ctx, connector2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// Test validation - invalid type
	connector3 := &Connector{
		ID:     "test-connector-3",
		Name:   "Test Connector 3",
		Type:   "invalid-type",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}
	err = store.Add(ctx, connector3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid connector type")
}

func TestStore_Get(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	connector := &Connector{
		ID:     "test-connector",
		Name:   "Test Connector",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}

	// Add connector first
	err = store.Add(ctx, connector)
	require.NoError(t, err)

	// Test successful get
	retrieved, err := store.Get(ctx, "test-connector")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, connector.ID, retrieved.ID)
	assert.Equal(t, connector.Name, retrieved.Name)
	assert.Equal(t, connector.Type, retrieved.Type)
	assert.Equal(t, connector.Status, retrieved.Status)
	assert.Equal(t, connector.Config, retrieved.Config)

	// Test get non-existent connector
	_, err = store.Get(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test get with empty ID
	_, err = store.Get(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestStore_Update(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	connector := &Connector{
		ID:     "test-connector",
		Name:   "Test Connector",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}

	// Add connector first
	err = store.Add(ctx, connector)
	require.NoError(t, err)

	// Update connector
	updated := &Connector{
		Name:   "Updated Test Connector",
		Type:   "github",
		Config: map[string]interface{}{"repo": "owner/repo"},
		Status: "inactive",
	}

	err = store.Update(ctx, "test-connector", updated)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := store.Get(ctx, "test-connector")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Connector", retrieved.Name)
	assert.Equal(t, "github", retrieved.Type)
	assert.Equal(t, "inactive", retrieved.Status)
	assert.Equal(t, map[string]interface{}{"repo": "owner/repo"}, retrieved.Config)

	// Test update non-existent connector
	err = store.Update(ctx, "non-existent", updated)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test update with empty ID
	err = store.Update(ctx, "", updated)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestStore_Remove(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	connector := &Connector{
		ID:     "test-connector",
		Name:   "Test Connector",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}

	// Add connector first
	err = store.Add(ctx, connector)
	require.NoError(t, err)

	// Verify it exists
	_, err = store.Get(ctx, "test-connector")
	assert.NoError(t, err)

	// Remove connector
	err = store.Remove(ctx, "test-connector")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = store.Get(ctx, "test-connector")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test remove non-existent connector
	err = store.Remove(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test remove with empty ID
	err = store.Remove(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestStore_List(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Add multiple connectors
	connectors := []*Connector{
		{
			ID:     "connector-1",
			Name:   "Connector 1",
			Type:   "filesystem",
			Config: map[string]interface{}{"path": "/tmp"},
			Status: "active",
		},
		{
			ID:     "connector-2",
			Name:   "Connector 2",
			Type:   "github",
			Config: map[string]interface{}{"repo": "owner/repo"},
			Status: "active",
		},
	}

	for _, conn := range connectors {
		err = store.Add(ctx, conn)
		require.NoError(t, err)
	}

	// List all connectors
	listed, err := store.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, listed, 2)

	// Verify connectors are returned (order may vary due to ORDER BY created_at DESC)
	found := make(map[string]bool)
	for _, conn := range listed {
		found[conn.ID] = true
		assert.NotZero(t, conn.CreatedAt)
		assert.NotZero(t, conn.UpdatedAt)
	}
	assert.True(t, found["connector-1"])
	assert.True(t, found["connector-2"])
}

func TestValidateConnector(t *testing.T) {
	tests := []struct {
		name      string
		connector *Connector
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid connector",
			connector: &Connector{
				ID:     "test",
				Name:   "Test",
				Type:   "filesystem",
				Status: "active",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			connector: &Connector{
				Name:   "Test",
				Type:   "filesystem",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "ID cannot be empty",
		},
		{
			name: "empty name",
			connector: &Connector{
				ID:     "test",
				Type:   "filesystem",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "empty type",
			connector: &Connector{
				ID:     "test",
				Name:   "Test",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "type cannot be empty",
		},
		{
			name: "invalid type",
			connector: &Connector{
				ID:     "test",
				Name:   "Test",
				Type:   "invalid",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "invalid connector type",
		},
		{
			name: "invalid status",
			connector: &Connector{
				ID:   "test",
				Name: "Test",
				Type: "filesystem",
			},
			wantErr: true,
			errMsg:  "invalid connector status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConnector(tt.connector)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStore_Timestamps(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	connector := &Connector{
		ID:        "test-connector",
		Name:      "Test Connector",
		Type:      "filesystem",
		Config:    map[string]interface{}{"path": "/tmp"},
		Status:    "active",
		CreatedAt: customTime,
		UpdatedAt: customTime,
	}

	// Add connector with custom timestamps
	err = store.Add(ctx, connector)
	assert.NoError(t, err)

	// Retrieve and verify timestamps are preserved
	retrieved, err := store.Get(ctx, "test-connector")
	assert.NoError(t, err)
	assert.Equal(t, customTime.Unix(), retrieved.CreatedAt.Unix())
	assert.Equal(t, customTime.Unix(), retrieved.UpdatedAt.Unix())

	// Update and verify updated_at is updated
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	updated := &Connector{
		Name:   "Updated Name",
		Type:   "filesystem",
		Config: map[string]interface{}{"path": "/tmp"},
		Status: "active",
	}

	err = store.Update(ctx, "test-connector", updated)
	assert.NoError(t, err)

	retrieved2, err := store.Get(ctx, "test-connector")
	assert.NoError(t, err)
	assert.Equal(t, customTime.Unix(), retrieved2.CreatedAt.Unix()) // CreatedAt should remain the same
	assert.True(t, retrieved2.UpdatedAt.After(customTime))          // UpdatedAt should be newer
}
