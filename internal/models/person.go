package models

import (
	"context"
	"database/sql"
	"fmt"
)

type Person struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Patronymic    string `json:"patronymic,omitempty"`
	Age           int    `json:"age"`
	GenderID      int    `json:"gender_id"`
	NationalityID int    `json:"nationality_id"`
}

type PersonFilter struct {
	ID            uint
	Name          string
	Surname       string
	Patronymic    string
	AgeFrom       int
	AgeTo         int
	GenderID      int
	NationalityID int
}

type PersonPatch struct {
	Name          *string `json:"name,omitempty"`
	Surname       *string `json:"surname,omitempty"`
	Patronymic    *string `json:"patronymic,omitempty"`
	Age           *int    `json:"age,omitempty"`
	GenderID      *int    `json:"gender_id,omitempty"`
	NationalityID *int    `json:"nationality_id,omitempty"`
}

func GetPersons(ctx context.Context, db *sql.DB, filter PersonFilter) ([]Person, error) {
	var persons []Person

	query := "SELECT id, name, surname, patronymic, age, gender_id, nationality_id FROM persons WHERE 1=1"
	var args []interface{}
	var conditions []string

	paramCounter := 1

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", paramCounter))
		args = append(args, "%"+filter.Name+"%")
		paramCounter++
	}

	if filter.Surname != "" {
		conditions = append(conditions, fmt.Sprintf("surname ILIKE $%d", paramCounter))
		args = append(args, "%"+filter.Surname+"%")
		paramCounter++
	}

	if filter.AgeTo > 0 {
		conditions = append(conditions, fmt.Sprintf("age <= $%d", paramCounter))
		args = append(args, filter.AgeTo)
		paramCounter++
	}

	if filter.AgeFrom > 0 {
		conditions = append(conditions, fmt.Sprintf("age >= $%d", paramCounter))
		args = append(args, filter.AgeFrom)
		paramCounter++
	}

	if filter.GenderID > 0 {
		conditions = append(conditions, fmt.Sprintf("gender_id = $%d", paramCounter))
		args = append(args, filter.GenderID)
		paramCounter++
	}

	if filter.NationalityID > 0 {
		conditions = append(conditions, fmt.Sprintf("nationality_id = $%d", paramCounter))
		args = append(args, filter.NationalityID)
		paramCounter++
	}

	for _, condition := range conditions {
		query += " AND " + condition
	}

	query += " ORDER BY id"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var person Person
		err = rows.Scan(
			&person.ID,
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.GenderID,
			&person.NationalityID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		persons = append(persons, person)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	return persons, nil
}

func DeletePersonByID(ctx context.Context, id uint, db *sql.DB) (Person, error) {
	var deletedPerson Person
	err := db.QueryRowContext(ctx, "DELETE FROM persons WHERE id = $1", id).Scan(&deletedPerson)
	if err != nil {
		return Person{}, fmt.Errorf("error deleting person: %w", err)
	}
	return deletedPerson, nil
}

func UpdatePerson(ctx context.Context, id uint, patch PersonPatch, db *sql.DB) (Person, error) {
	var currentPerson Person
	err := db.QueryRowContext(ctx,
		"SELECT id, name, surname, patronymic, age, gender_id, nationality_id FROM persons WHERE id = $1",
		id).Scan(
		&currentPerson.ID,
		&currentPerson.Name,
		&currentPerson.Surname,
		&currentPerson.Patronymic,
		&currentPerson.Age,
		&currentPerson.GenderID,
		&currentPerson.NationalityID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Person{}, fmt.Errorf("record with id=%d not found", id)
		}
		return Person{}, fmt.Errorf("error while receiving data: %w", err)
	}

	// Формируем SQL запрос и аргументы динамически
	query := "UPDATE persons SET"
	var args []interface{}
	paramCounter := 1
	needUpdate := false

	// Проверяем каждое поле на необходимость обновления
	if patch.Name != nil {
		query += fmt.Sprintf(" name = $%d,", paramCounter)
		args = append(args, *patch.Name)
		paramCounter++
		needUpdate = true
	}

	if patch.Surname != nil {
		query += fmt.Sprintf(" surname = $%d,", paramCounter)
		args = append(args, *patch.Surname)
		paramCounter++
		needUpdate = true
	}

	if patch.Patronymic != nil {
		query += fmt.Sprintf(" patronymic = $%d,", paramCounter)
		args = append(args, *patch.Patronymic)
		paramCounter++
		needUpdate = true
	}

	if patch.Age != nil {
		query += fmt.Sprintf(" age = $%d,", paramCounter)
		args = append(args, *patch.Age)
		paramCounter++
		needUpdate = true
	}

	if patch.GenderID != nil {
		query += fmt.Sprintf(" gender_id = $%d,", paramCounter)
		args = append(args, *patch.GenderID)
		paramCounter++
		needUpdate = true
	}

	if patch.NationalityID != nil {
		query += fmt.Sprintf(" nationality_id = $%d,", paramCounter)
		args = append(args, *patch.NationalityID)
		paramCounter++
		needUpdate = true
	}

	if !needUpdate {
		return currentPerson, nil
	}

	query = query[:len(query)-1]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, surname, patronymic, age, gender_id, nationality_id", paramCounter)
	args = append(args, id)

	var updatedPerson Person
	err = db.QueryRowContext(ctx, query, args...).Scan(
		&updatedPerson.ID,
		&updatedPerson.Name,
		&updatedPerson.Surname,
		&updatedPerson.Patronymic,
		&updatedPerson.Age,
		&updatedPerson.GenderID,
		&updatedPerson.NationalityID,
	)

	if err != nil {
		return Person{}, fmt.Errorf("ошибка при обновлении: %w", err)
	}

	return updatedPerson, nil
}

func CreatePerson(ctx context.Context, person Person, db *sql.DB) (Person, error) {
	var createdPerson Person
	err := db.QueryRowContext(ctx,
		"INSERT INTO persons (name, surname, patronymic, age, gender_id, nationality_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, name, surname, patronymic, age, gender_id, nationality_id",
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.GenderID,
		person.NationalityID,
	).Scan(createdPerson)
	if err != nil {
		return Person{}, fmt.Errorf("error inserting person: %w", err)
	}
	return createdPerson, nil
}
