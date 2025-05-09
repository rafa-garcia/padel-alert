package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
)

// UserStorage defines operations for user persistence
type UserStorage interface {
	GetUser(ctx context.Context, userID string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

// RedisUserStorage implements UserStorage using Redis
type RedisUserStorage struct {
	redis *RedisClient
}

// NewRedisUserStorage creates a new Redis user storage
func NewRedisUserStorage(redis *RedisClient) *RedisUserStorage {
	return &RedisUserStorage{redis: redis}
}

// GetUser gets a user by ID
func (s *RedisUserStorage) GetUser(ctx context.Context, userID string) (*model.User, error) {
	key := fmt.Sprintf("user:%s", userID)
	data, err := s.redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}

	return &user, nil
}

// CreateUser stores a new user
func (s *RedisUserStorage) CreateUser(ctx context.Context, user *model.User) error {
	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Serialize
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	// Store user and index by email
	userKey := fmt.Sprintf("user:%s", user.ID)
	emailKey := fmt.Sprintf("user:email:%s", user.Email)

	pipe := s.redis.Client.Pipeline()
	pipe.Set(ctx, userKey, data, 0)
	pipe.Set(ctx, emailKey, user.ID, 0)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (s *RedisUserStorage) UpdateUser(ctx context.Context, user *model.User) error {
	// Check if user exists
	key := fmt.Sprintf("user:%s", user.ID)
	exists, err := s.redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("check user existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	// Get the current user data to check if email changed
	currentUser, err := s.GetUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("get current user: %w", err)
	}

	// Update timestamp
	user.CreatedAt = currentUser.CreatedAt
	user.UpdatedAt = time.Now()

	// Serialize
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	pipe := s.redis.Client.Pipeline()
	pipe.Set(ctx, key, data, 0)

	// Update email index if changed
	if currentUser.Email != user.Email {
		oldEmailKey := fmt.Sprintf("user:email:%s", currentUser.Email)
		newEmailKey := fmt.Sprintf("user:email:%s", user.Email)
		pipe.Del(ctx, oldEmailKey)
		pipe.Set(ctx, newEmailKey, user.ID, 0)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// GetUserByEmail gets a user by email
func (s *RedisUserStorage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	// Get user ID from email index
	emailKey := fmt.Sprintf("user:email:%s", email)
	userID, err := s.redis.Client.Get(ctx, emailKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	// Get user from ID
	return s.GetUser(ctx, userID)
}
