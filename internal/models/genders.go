package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Gender struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GenderFilter struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type PatchGender struct {
	Name *string `json:"name,omitempty"`
}

func GetGenders(db *sql.DB, ctx context.Context, filter GenderFilter) ([]Gender, error) {
	var genders []Gender
	query := "SELECT id, name FROM genders WHERE 1=1"

	var args []interface{}
	var conditions []string
	paramCounter := 1

	if filter.ID > 0 {
		conditions = append(conditions, fmt.Sprintf("id = $%d", paramCounter))
		args = append(args, filter.ID)
		paramCounter++
	}
	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", paramCounter))
		args = append(args, "%"+filter.Name+"%")
		paramCounter++
	}

	for _, condition := range conditions {
		query += " AND " + condition
	}

	if filter.Page > 0 && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramCounter)
		args = append(args, filter.Limit)
		paramCounter++
		query += fmt.Sprintf(" OFFSET $%d", paramCounter)
		args = append(args, (filter.Page-1)*filter.Limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var gender Gender
		err = rows.Scan(
			&gender.ID,
			&gender.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		genders = append(genders, gender)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	return genders, nil
}

func DeleteGender(db *sql.DB, ctx context.Context, id int) (Gender, error) {
	var deletedGender Gender
	err := db.QueryRowContext(ctx, "DELETE FROM genders WHERE id = $1 RETURNING id, name", id).Scan(
		&deletedGender.ID,
		&deletedGender.Name,
	)
	if err != nil {
		return Gender{}, fmt.Errorf("error deleting gender: %w", err)
	}
	return deletedGender, nil
}

func UpdateGender(db *sql.DB, ctx context.Context, id int, patch PatchGender) (Gender, error) {
	var currentGender Gender
	err := db.QueryRowContext(ctx, "SELECT id, name FROM genders WHERE id = $1", id).Scan(
		&currentGender.ID,
		&currentGender.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Gender{}, fmt.Errorf("record with id=%d not found", id)
		}
		return Gender{}, fmt.Errorf("error while receiving data: %w", err)
	}

	query := "UPDATE genders SET"
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
		return currentGender, nil
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name", paramCounter)
	args = append(args, id)

	var updatedGender Gender
	err = db.QueryRowContext(ctx, query, args...).Scan(
		&updatedGender.ID,
		&updatedGender.Name,
	)

	if err != nil {
		return Gender{}, fmt.Errorf("error during update: %w", err)
	}

	return updatedGender, nil
}

func CreateGender(db *sql.DB, ctx context.Context, name string) (Gender, error) {
	var createdGender Gender
	err := db.QueryRowContext(ctx, "INSERT INTO genders (name) VALUES ($1) RETURNING id, name", name).Scan(
		&createdGender.ID,
		&createdGender.Name,
	)
	if err != nil {
		return Gender{}, fmt.Errorf("error inserting gender: %w", err)
	}
	return createdGender, nil
}
