package authhandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/zanzhit/flat-seller/internal/domain/errs"
	"github.com/zanzhit/flat-seller/internal/http-server/handlers"
	resp "github.com/zanzhit/flat-seller/internal/lib/api/response"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
)

type RequestRegister struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	UserType string `json:"user_type" validate:"required"`
}

type RequestLogin struct {
	Id       string `json:"id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthHandler struct {
	log  *slog.Logger
	user User
}

type User interface {
	Login(email, password string) (string, error)
	RegisterNewUser(email, password, userType string) (string, error)
}

func New(
	log *slog.Logger,
	user User,
) *AuthHandler {
	return &AuthHandler{
		log:  log,
		user: user,
	}
}

func (h *AuthHandler) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Register"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req RequestRegister
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			handlers.Error(w, r, http.StatusBadRequest, resp.Error("empty request", ""))

			return
		}

		log.Error("failed to decode request body", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to decode request", middleware.GetReqID(r.Context())))

		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

		return
	}

	id, err := h.user.RegisterNewUser(req.Email, req.Password, req.UserType)
	if err != nil {
		if errors.Is(err, errs.ErrUserExists) {
			handlers.Error(w, r, http.StatusBadRequest, resp.Error("user with this email already exists", ""))

			return
		}
		if errors.Is(err, errs.ErrUserType) {
			handlers.Error(w, r, http.StatusBadRequest, resp.Error("invalid user_type", ""))

			return
		}

		log.Error("failed to register new user", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to register new user", middleware.GetReqID(r.Context())))

		return
	}

	render.JSON(w, r, map[string]string{"id": id})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Login"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req RequestLogin
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			handlers.Error(w, r, http.StatusBadRequest, resp.Error("empty request", ""))

			return
		}

		log.Error("failed to decode request body", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to decode request", middleware.GetReqID(r.Context())))

		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

		return
	}

	token, err := h.user.Login(req.Id, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidCredentials) {
			handlers.Error(w, r, http.StatusBadRequest, resp.Error("invalid credentials", ""))

			return
		}

		log.Error("failed to login", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to login", middleware.GetReqID(r.Context())))

		return
	}

	render.JSON(w, r, map[string]string{"token": token})
}
