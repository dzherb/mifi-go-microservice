package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/dzherb/mifi-go-microservice/model"
	"github.com/dzherb/mifi-go-microservice/storage"
)

var (
	ErrUserDoesNotExist = errors.New("user does not exist")
)

type Storage[T any] interface {
	Set(context.Context, string, T) error
	Get(context.Context, string) (T, error)
	GetAll(context.Context) ([]T, error)
	Delete(context.Context, string) error
}

type UserService struct {
	log     *slog.Logger
	storage Storage[model.User]
}

func NewUserService(
	log *slog.Logger,
	storage Storage[model.User],
) *UserService {
	return &UserService{
		log:     log,
		storage: storage,
	}
}

var (
	ErrMissingUserID = errors.New("missing user ID")
)

const userOperationsTimeout = 10 * time.Second

func (u *UserService) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	createdUser := user
	createdUser.ID = u.generateID()

	err := u.storage.Set(ctx, u.buildStorageKey(createdUser.ID), createdUser)
	if err != nil {
		return user, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (u *UserService) Get(ctx context.Context, id string) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	user, err := u.storage.Get(ctx, u.buildStorageKey(id))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return user, ErrUserDoesNotExist
		}

		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (u *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)

	defer cancel()

	users, err := u.storage.GetAll(ctx)
	if err != nil {

		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

func (u *UserService) Update(ctx context.Context, user model.User) error {
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

func (u *UserService) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, userOperationsTimeout)
	defer cancel()

	err := u.storage.Delete(ctx, u.buildStorageKey(id))

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (u *UserService) generateID() string {
	return uuid.New().String()
}

func (u *UserService) buildStorageKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}
