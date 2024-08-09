package flatservice

import (
	"fmt"
	"log/slog"

	"github.com/zanzhit/flat-seller/internal/domain/constants"
	"github.com/zanzhit/flat-seller/internal/domain/errs"
	"github.com/zanzhit/flat-seller/internal/domain/models"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
)

type FlatService struct {
	log  *slog.Logger
	flat Flat
}

func New(log *slog.Logger, flat Flat) *FlatService {
	return &FlatService{
		log:  log,
		flat: flat,
	}
}

type Flat interface {
	SaveFlat(houseID, price, rooms int) (models.Flat, error)
	UpdateFlat(flatID, price, rooms int, status string) (models.Flat, error)
}

func (s *FlatService) SaveFlat(houseID, price, rooms int) (models.Flat, error) {
	const op = "service.flat.SaveFlat"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("house_id", houseID),
	)

	log.Info("saving flat")

	flat, err := s.flat.SaveFlat(houseID, price, rooms)
	if err != nil {
		log.Error("failed to save flat", sl.Err(err))

		return models.Flat{}, fmt.Errorf("%s: %w", op, err)
	}

	return flat, nil
}

func (s *FlatService) UpdateFlat(flatID, price, rooms int, status string) (models.Flat, error) {
	const op = "service.flat.UpdateFlat"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("flat_id", flatID),
	)

	if status != constants.Declined && status != constants.Approved && status != constants.Moderation && status != constants.Created {
		log.Warn("invalid status", sl.Err(errs.ErrFlatStatus))

		return models.Flat{}, fmt.Errorf("%s: %w", op, errs.ErrFlatStatus)
	}

	log.Info("updating flat")

	flat, err := s.flat.UpdateFlat(flatID, price, rooms, status)
	if err != nil {
		log.Error("failed to update flat", sl.Err(err))

		return models.Flat{}, fmt.Errorf("%s: %w", op, err)
	}

	return flat, nil
}
