package storage

import (
	"context"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestRedisUserStorage_CreateUser(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	userStorage := NewRedisUserStorage(redisClient)
	ctx := context.Background()

	user := &model.User{
		ID:    "test-user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := userStorage.CreateUser(ctx, user)
	assert.NoError(t, err)

	exists := mini.Exists("user:" + user.ID)
	assert.True(t, exists)

	exists = mini.Exists("user:email:" + user.Email)
	assert.True(t, exists)

	emailIndexValue, _ := mini.Get("user:email:" + user.Email)
	assert.Equal(t, user.ID, emailIndexValue)
}

func TestRedisUserStorage_GetUser(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	userStorage := NewRedisUserStorage(redisClient)
	ctx := context.Background()

	now := time.Now().UTC()
	user := &model.User{
		ID:        "test-user-123",
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := userStorage.CreateUser(ctx, user)
	assert.NoError(t, err)

	fetchedUser, err := userStorage.GetUser(ctx, user.ID)
	assert.NoError(t, err)

	assert.Equal(t, user.ID, fetchedUser.ID)
	assert.Equal(t, user.Name, fetchedUser.Name)
	assert.Equal(t, user.Email, fetchedUser.Email)
}

func TestRedisUserStorage_UpdateUser(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	userStorage := NewRedisUserStorage(redisClient)
	ctx := context.Background()

	user := &model.User{
		ID:    "test-user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := userStorage.CreateUser(ctx, user)
	assert.NoError(t, err)

	user.Name = "Updated Name"
	user.Email = "updated@example.com"
	err = userStorage.UpdateUser(ctx, user)
	assert.NoError(t, err)

	updatedUser, err := userStorage.GetUser(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.Equal(t, "updated@example.com", updatedUser.Email)

	exists := mini.Exists("user:email:test@example.com")
	assert.False(t, exists)

	exists = mini.Exists("user:email:updated@example.com")
	assert.True(t, exists)
}

func TestRedisUserStorage_GetUserByEmail(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	userStorage := NewRedisUserStorage(redisClient)
	ctx := context.Background()

	user := &model.User{
		ID:    "test-user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := userStorage.CreateUser(ctx, user)
	assert.NoError(t, err)

	fetchedUser, err := userStorage.GetUserByEmail(ctx, user.Email)
	assert.NoError(t, err)

	assert.Equal(t, user.ID, fetchedUser.ID)
	assert.Equal(t, user.Name, fetchedUser.Name)
	assert.Equal(t, user.Email, fetchedUser.Email)

	_, err = userStorage.GetUserByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
}
