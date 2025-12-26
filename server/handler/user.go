package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

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
	err := json.NewEncoder(w).Encode(req)
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

		response.Write(
			w,
			h.log,
			response.NewError("something went wrong"),
			http.StatusInternalServerError,
		)

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

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := vars["id"]

	_, err := uuid.Parse(userID)
	if err != nil {
		response.Write(
			w, h.log,
			response.NewError("invalid user ID"),
			http.StatusBadRequest,
		)

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

		response.Write(
			w, h.log,
			response.NewError("something went wrong"),
			http.StatusInternalServerError,
		)

		return
	}

	response.Write(
		w, h.log,
		UserResponse(user),
		http.StatusOK,
	)
}

type UserValidationFailedResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (h *UserHandler) validateIncomingUserOrWriteError(
	w http.ResponseWriter,
	user model.User,
) (ok bool) {
	switch {
	case user.Name == "":
		response.Write(
			w, h.log,
			UserValidationFailedResponse{
				Field:   "name",
				Message: "name is required",
			},
			http.StatusBadRequest,
		)

		return
	case user.Email == "": // todo add more complex validation
		response.Write(
			w, h.log,
			UserValidationFailedResponse{
				Field:   "email",
				Message: "email is required",
			},
			http.StatusBadRequest,
		)

		return
	}

	return true
}
