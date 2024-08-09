package flatstorage

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanzhit/flat-seller/internal/domain/constants"
	"github.com/zanzhit/flat-seller/internal/domain/models"
	"github.com/zanzhit/flat-seller/internal/storage/postgres"
)

type FlatStorage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *FlatStorage {
	return &FlatStorage{db: db}
}

func (s *FlatStorage) SaveFlat(houseID, price, rooms int) (models.Flat, error) {
	const op = "storage.postgres.flat.SaveFlat"

	query := fmt.Sprintf(`
		INSERT INTO %s (house_id, flat_number, price, rooms, status, created_at, updated_at)
		VALUES ($1, (SELECT COALESCE(MAX(flat_number), 0) + 1 FROM %s WHERE house_id = $1), $2, $3, '%s', $4, $5)
		RETURNING *`, postgres.FlatsTable, postgres.FlatsTable, constants.Created)

	now := time.Now()
	var flat models.Flat
	err := s.db.QueryRowx(query, houseID, price, rooms, now, now).StructScan(&flat)
	if err != nil {
		return models.Flat{}, fmt.Errorf("%s: %w", op, err)
	}

	return flat, nil
}

func (s *FlatStorage) UpdateFlat(flatID, price, rooms int, status string) (models.Flat, error) {
	const op = "storage.postgres.flat.UpdateFlat"

	query := fmt.Sprintf("UPDATE %s SET status = $1, updated_at = $2, price = $3, rooms = $4 WHERE id = $5 RETURNING *", postgres.FlatsTable)

	now := time.Now()
	var flat models.Flat
	err := s.db.QueryRowx(query, status, now, price, rooms, flatID).StructScan(&flat)
	if err != nil {
		return models.Flat{}, fmt.Errorf("%s: %w", op, err)
	}

	return flat, nil
}
