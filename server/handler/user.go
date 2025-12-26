package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/dzherb/mifi-go-microservice/model"
	"github.com/dzherb/mifi-go-microservice/server/response"
	"github.com/dzherb/mifi-go-microservice/service"
)

type UserService interface {
	Create(context.Context, model.User) (model.User, error)
	Get(context.Context, string) (model.User, error)
	GetAll(context.Context) ([]model.User, error)
	Update(context.Context, model.User) error
	Delete(context.Context, string) error
}

type Notifier interface {
	Send(string, map[string]any)
}

type UserHandler struct {
	log      *slog.Logger
	service  UserService
	notifier Notifier
}

func NewUserHandler(
	log *slog.Logger,
	userService UserService,
	notifier Notifier,
) *UserHandler {
	return &UserHandler{
		log:      log,
		service:  userService,
		notifier: notifier,
	}
}

type UserCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserResponse model.User

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req UserCreateRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Write(
			w, h.log,
			response.NewError(err.Error()),
			http.StatusUnprocessableEntity,
		)

		return
	}

	user := model.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if !h.validateIncomingUserOrWriteError(w, user) {
		return
	}

	created, err := h.service.Create(r.Context(), user)
	if err != nil {
		h.log.Error(
			"failed to create user",
			slog.String("error", err.Error()),
		)

		response.WriteDefaultError(w, h.log)

		return
	}

	go h.notifier.Send(
		"user_created",
		map[string]any{
			"user_id": created.ID,
		},
	)

	response.Write(
		w,
		h.log,
		UserResponse(created),
		http.StatusCreated,
	)
}

type AllUsersResponse struct {
	Users []model.User `json:"users"`
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAll(r.Context())
	if err != nil {
		response.WriteDefaultError(w, h.log)

		return
	}

	response.Write(w, h.log, AllUsersResponse{Users: users}, http.StatusOK)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDParamOrWriteError(w, r)
	if !ok {
		return
	}

	user, err := h.service.Get(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserDoesNotExist) {
			response.Write(
				w, h.log,
				response.NewError(service.ErrUserDoesNotExist.Error()),
				http.StatusNotFound,
			)

			return
		}

		response.WriteDefaultError(w, h.log)

		return
	}

	response.Write(
		w, h.log,
		UserResponse(user),
		http.StatusOK,
	)
}

type UserUpdateRequest struct {
	UserCreateRequest
}

type UserUpdateResponse model.User

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDParamOrWriteError(w, r)
	if !ok {
		return
	}

	var req UserUpdateRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Write(
			w, h.log,
			response.NewError(err.Error()),
			http.StatusUnprocessableEntity,
		)

		return
	}

	user := model.User{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
	}

	if !h.validateIncomingUserOrWriteError(w, user) {
		return
	}

	err = h.service.Update(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrUserDoesNotExist) {
			response.Write(
				w, h.log,
				response.NewError(service.ErrUserDoesNotExist.Error()),
				http.StatusNotFound,
			)

			return
		}

		response.WriteDefaultError(w, h.log)

		return
	}

	response.Write(
		w, h.log,
		UserUpdateResponse(user),
		http.StatusOK,
	)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDParamOrWriteError(w, r)
	if !ok {
		return
	}

	err := h.service.Delete(r.Context(), userID)
	if err != nil {
		response.WriteDefaultError(w, h.log)

		return
	}

	response.Write(w, h.log, nil, http.StatusNoContent)
}

type UserValidationFailedResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (h *UserHandler) getUserIDParamOrWriteError(
	w http.ResponseWriter,
	r *http.Request,
) (string, bool) {
	vars := mux.Vars(r)
	userID := vars["id"]

	_, err := uuid.Parse(userID)
	if err != nil {
		response.Write(
			w, h.log,
			response.NewError("invalid user ID"),
			http.StatusBadRequest,
		)

		return "", false
	}

	return userID, true
}

func (h *UserHandler) validateIncomingUserOrWriteError(
	w http.ResponseWriter,
	user model.User,
) (ok bool) {
	if user.Name == "" {
		response.Write(
			w, h.log,
			UserValidationFailedResponse{
				Field:   "name",
				Message: "name is required",
			},
			http.StatusBadRequest,
		)

		return
	}

	if user.Email == "" {
		response.Write(
			w, h.log,
			UserValidationFailedResponse{
				Field:   "email",
				Message: "email is required",
			},
			http.StatusBadRequest,
		)
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		response.Write(
			w, h.log,
			UserValidationFailedResponse{
				Field:   "email",
				Message: "email not valid: " + err.Error(),
			},
			http.StatusBadRequest,
		)
	}

	return true
}
