package househandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/zanzhit/flat-seller/internal/domain/constants"
	"github.com/zanzhit/flat-seller/internal/domain/models"
	"github.com/zanzhit/flat-seller/internal/http-server/handlers"
	authmid "github.com/zanzhit/flat-seller/internal/http-server/middleware/auth"
	resp "github.com/zanzhit/flat-seller/internal/lib/api/response"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
)

type Request struct {
	Year      int    `json:"year" validate:"required"`
	Developer string `json:"developer,omitempty"`
	Address   string `json:"address" validate:"required"`
}

type HouseHandler struct {
	log   *slog.Logger
	house House
}

type House interface {
	SaveHouse(address, developer string, year int) (models.House, error)
	HouseUser(houseID int) ([]models.Flat, error)
	HouseAdmin(houseID int) ([]models.Flat, error)
}

func New(log *slog.Logger, house House) *HouseHandler {
	return &HouseHandler{
		log:   log,
		house: house,
	}
}

func (h *HouseHandler) SaveHouse(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.house.SaveFlat"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req Request
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		handlers.Error(w, r, http.StatusBadRequest, resp.Error("request body is empty", ""))

		return
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

		return
	}

	house, err := h.house.SaveHouse(req.Address, req.Developer, req.Year)
	if err != nil {
		log.Error("failed to save house", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to save house", ""))

		return
	}

	render.JSON(w, r, house)
}

func (h *HouseHandler) House(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.house.HouseUser"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("request received", slog.Any("request", r))

	houseID := chi.URLParam(r, "id")
	if houseID == "" {
		log.Error("houseID is empty")

		handlers.Error(w, r, http.StatusBadRequest, resp.Error("houseID is empty", ""))

		return
	}

	id, err := strconv.Atoi(houseID)
	if err != nil {
		log.Error("house id is not a number", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.Error("house id is not a number", ""))

		return
	}

	user, ok := r.Context().Value(authmid.UserContextKey).(models.User)
	if !ok {
		h.log.Error("user not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var flats []models.Flat
	if user.UserType == constants.Admin {
		flats, err = h.house.HouseAdmin(id)
	} else {
		flats, err = h.house.HouseUser(id)
	}
	if err != nil {
		log.Error("failed to get flats", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to get flats", ""))

		return
	}

	render.JSON(w, r, flats)
}
