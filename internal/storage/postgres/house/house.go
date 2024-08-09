package housestorage

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanzhit/flat-seller/internal/domain/constants"
	"github.com/zanzhit/flat-seller/internal/domain/models"
	"github.com/zanzhit/flat-seller/internal/storage/postgres"
)

type HouseStorage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *HouseStorage {
	return &HouseStorage{db: db}
}

func (s *HouseStorage) SaveHouse(address, developer string, year int) (models.House, error) {
	const op = "storage.postgres.house.SaveHouse"

	query := fmt.Sprintf("INSERT INTO %s (address, year, developer, created_at) VALUES ($1, $2, $3, $4) RETURNING *", postgres.HousesTable)

	var house models.House
	err := s.db.QueryRowx(query, address, year, developer, time.Now()).StructScan(&house)
	if err != nil {
		return house, fmt.Errorf("%s: %w", op, err)
	}

	return house, nil
}

func (s *HouseStorage) HouseUser(houseID int) ([]models.Flat, error) {
	const op = "storage.postgres.house.HouseUser"

	query := fmt.Sprintf("SELECT * FROM %s WHERE house_id = $1 AND status = '%s'", postgres.FlatsTable, constants.Approved)

	var flats []models.Flat
	err := s.db.Select(&flats, query, houseID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flats, nil
}

func (s *HouseStorage) HouseAdmin(houseID int) ([]models.Flat, error) {
	const op = "storage.postgres.house.HouseModerator"

	query := fmt.Sprintf("SELECT * FROM %s WHERE house_id = $1", postgres.FlatsTable)

	var flats []models.Flat
	err := s.db.Select(&flats, query, houseID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flats, nil
}
