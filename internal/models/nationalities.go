package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Nationality struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type NationalityCreateRequest struct {
	Name string `json:"name"`
}

type NationalityFilter struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type PatchNationality struct {
	Name *string `json:"name,omitempty"`
}

func GetNationalities(db *sql.DB, ctx context.Context, filter NationalityFilter) ([]Nationality, error) {
	var nationalities []Nationality
	query := "SELECT id, name FROM nationalities WHERE 1=1"

	var args []interface{}
	var conditions []string
	paramCounter := 1

	if filter.ID > 0 {
		conditions = append(conditions, "id = $"+strconv.Itoa(paramCounter))
		args = append(args, filter.ID)
		paramCounter++
	}
	if filter.Name != "" {
		conditions = append(conditions, "name ILIKE $"+strconv.Itoa(paramCounter))
		args = append(args, "%"+filter.Name+"%")
		paramCounter++
	}

	for _, condition := range conditions {
		query += " AND " + condition
	}

	if filter.Page > 0 && filter.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(paramCounter)
		args = append(args, filter.Limit)
		paramCounter++
		query += " OFFSET $" + strconv.Itoa(paramCounter)
		args = append(args, (filter.Page-1)*filter.Limit)
		paramCounter++
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nationality Nationality
		err = rows.Scan(
			&nationality.ID,
			&nationality.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		nationalities = append(nationalities, nationality)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	return nationalities, nil
}

func DeleteNationality(db *sql.DB, ctx context.Context, id int) (Nationality, error) {
	var deletedNationality Nationality
	query := "DELETE FROM nationalities WHERE id = $1 RETURNING id, name"
	err := db.QueryRowContext(ctx, query, id).Scan(&deletedNationality.ID, &deletedNationality.Name)
	if err != nil {
		return Nationality{}, fmt.Errorf("error deleting nationality: %w", err)
	}
	return deletedNationality, nil
}

func UpdateNationality(db *sql.DB, ctx context.Context, id int, patch PatchNationality) (Nationality, error) {
	var currentNationality Nationality
	err := db.QueryRowContext(ctx, "SELECT id, name FROM nationalities WHERE id = $1", id).Scan(
		&currentNationality.ID,
		&currentNationality.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Nationality{}, fmt.Errorf("record with id=%d not found", id)
		}
		return Nationality{}, fmt.Errorf("error while receiving data: %w", err)
	}

	query := "UPDATE nationalities SET"
	var args []interface{}
	paramCounter := 1
	needUpdate := false

	if patch.Name != nil {
		query += fmt.Sprintf(" name = $%d", paramCounter)
		args = append(args, *patch.Name)
		paramCounter++
		needUpdate = true
	}

	if !needUpdate {
		return currentNationality, nil
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name", paramCounter)
	args = append(args, id)

	var updatedNationality Nationality
	err = db.QueryRowContext(ctx, query, args...).Scan(
		&updatedNationality.ID,
		&updatedNationality.Name,
	)

	if err != nil {
		return Nationality{}, fmt.Errorf("error during update: %w", err)
	}

	return updatedNationality, nil
}

func CreateNationality(db *sql.DB, ctx context.Context, name string) (Nationality, error) {
	var createdNationality Nationality
	err := db.QueryRowContext(ctx, "INSERT INTO nationalities (name) VALUES ($1) RETURNING id, name", name).Scan(
		&createdNationality.ID,
		&createdNationality.Name,
	)
	if err != nil {
		return Nationality{}, fmt.Errorf("error inserting nationality: %w", err)
	}
	return createdNationality, nil
}
