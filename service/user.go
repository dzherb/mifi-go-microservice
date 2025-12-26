package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/dzherb/mifi-go-microservice/model"
)

type Storage[T any] interface {
	Set(context.Context, string, T) error
	Get(context.Context, string) (T, error)
	Delete(context.Context, string) error
}

type User struct {
	log     *slog.Logger
	storage Storage[model.User]
}

func NewUser(log *slog.Logger, storage Storage[model.User]) *User {
	return &User{
		log:     log,
		storage: storage,
	}
}

var (
	ErrMissingUserID = errors.New("missing user ID")
)

const userOperationsTimeout = 10 * time.Second

func (u *User) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	//if user.ID == "" {
	//	return user, ErrMissingUserID
	//}

	createdUser := user
	createdUser.ID = u.generateID()

	err := u.storage.Set(ctx, u.buildStorageKey(user.ID), user)
	if err != nil {
		return user, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (u *User) Get(ctx context.Context, id string) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	user, err := u.storage.Get(ctx, u.buildStorageKey(id))
	if err != nil {
		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (u *User) Update(ctx context.Context, user model.User) error {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	if user.ID == "" {
		return ErrMissingUserID
	}

	err := u.storage.Set(ctx, u.buildStorageKey(user.ID), user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (u *User) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)
	defer cancel()

	err := u.storage.Delete(ctx, u.buildStorageKey(id))

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (u *User) generateID() string {
	return uuid.New().String()
}

func (u *User) buildStorageKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}
