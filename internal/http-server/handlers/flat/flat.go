package flathandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/zanzhit/flat-seller/internal/domain/models"
	"github.com/zanzhit/flat-seller/internal/http-server/handlers"
	resp "github.com/zanzhit/flat-seller/internal/lib/api/response"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
)

type Request struct {
	Price  int    `json:"price" validate:"required"`
	Room   int    `json:"room" validate:"required"`
	Status string `json:"status"`
}

type UpdateRequest struct {
	Id int `json:"id" validate:"required"`
	Request
}

type SaveRequest struct {
	Id int `json:"house_id" validate:"required"`
	Request
}

type FlatHandler struct {
	log  *slog.Logger
	flat Flat
}

type Flat interface {
	SaveFlat(houseID, price, rooms int) (models.Flat, error)
	UpdateFlat(flatID, price, rooms int, status string) (models.Flat, error)
}

func New(
	log *slog.Logger,
	flat Flat,
) *FlatHandler {
	return &FlatHandler{
		log:  log,
		flat: flat,
	}
}

func (h *FlatHandler) SaveFlat(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.flat.SaveFlat"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req SaveRequest
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		handlers.Error(w, r, http.StatusBadRequest, resp.Error("request body is empty", ""))

		return
	}

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to decode request body", middleware.GetReqID(r.Context())))

		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

		return
	}

	flat, err := h.flat.SaveFlat(req.Id, req.Price, req.Room)
	if err != nil {
		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to save flat", middleware.GetReqID(r.Context())))
		return
	}

	render.JSON(w, r, flat)
}

func (h *FlatHandler) UpdateFlat(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.flat.UpdateFlat"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req UpdateRequest
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		handlers.Error(w, r, http.StatusBadRequest, resp.Error("request body is empty", ""))

		return
	}

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))

		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to decode request body", middleware.GetReqID(r.Context())))

		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("invalid request", sl.Err(err))

		handlers.Error(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

		return
	}

	flat, err := h.flat.UpdateFlat(req.Id, req.Price, req.Room, req.Status)
	if err != nil {
		handlers.Error(w, r, http.StatusInternalServerError, resp.Error("failed to save flat", middleware.GetReqID(r.Context())))
		return
	}

	render.JSON(w, r, flat)
}
