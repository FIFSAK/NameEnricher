package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Person struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Surname     string      `json:"surname"`
	Patronymic  string      `json:"patronymic,omitempty"`
	Age         int         `json:"age"`
	Gender      Gender      `json:"gender"`
	Nationality Nationality `json:"nationality"`
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

	query := `SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1`

	var args []interface{}
	var conditions []string

	paramCounter := 1

	if filter.ID > 0 {
		conditions = append(conditions, fmt.Sprintf("p.id = $%d", paramCounter))
		args = append(args, filter.ID)
		paramCounter++
	}

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("p.name ILIKE $%d", paramCounter))
		args = append(args, "%"+filter.Name+"%")
		paramCounter++
	}

	if filter.Surname != "" {
		conditions = append(conditions, fmt.Sprintf("p.surname ILIKE $%d", paramCounter))
		args = append(args, "%"+filter.Surname+"%")
		paramCounter++
	}

	if filter.AgeTo > 0 {
		conditions = append(conditions, fmt.Sprintf("p.age <= $%d", paramCounter))
		args = append(args, filter.AgeTo)
		paramCounter++
	}

	if filter.AgeFrom > 0 {
		conditions = append(conditions, fmt.Sprintf("p.age >= $%d", paramCounter))
		args = append(args, filter.AgeFrom)
		paramCounter++
	}

	if filter.GenderID > 0 {
		conditions = append(conditions, fmt.Sprintf("p.gender_id = $%d", paramCounter))
		args = append(args, filter.GenderID)
		paramCounter++
	}

	if filter.NationalityID > 0 {
		conditions = append(conditions, fmt.Sprintf("p.nationality_id = $%d", paramCounter))
		args = append(args, filter.NationalityID)
		paramCounter++
	}

	for _, condition := range conditions {
		query += " AND " + condition
	}

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
			&person.Gender.ID,
			&person.Gender.Name,
			&person.Nationality.ID,
			&person.Nationality.Name,
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
	query := "DELETE FROM persons WHERE id = $1 RETURNING id, name, surname"
	err := db.QueryRowContext(ctx, query, id).Scan(&deletedPerson.ID, &deletedPerson.Name, &deletedPerson.Surname)
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
		&currentPerson.Gender.ID,
		&currentPerson.Nationality.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Person{}, fmt.Errorf("record with id=%d not found", id)
		}
		return Person{}, fmt.Errorf("error while receiving data: %w", err)
	}

	query := "UPDATE persons SET"
	var args []interface{}
	paramCounter := 1
	needUpdate := false

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
		&updatedPerson.Gender.ID,
		&updatedPerson.Nationality.ID,
	)

	if err != nil {
		return Person{}, fmt.Errorf("error during update: %w", err)
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
		person.Gender.ID,
		person.Nationality.ID,
	).Scan(
		&createdPerson.ID,
		&createdPerson.Name,
		&createdPerson.Surname,
		&createdPerson.Patronymic,
		&createdPerson.Age,
		&createdPerson.Gender.ID,
		&createdPerson.Nationality.ID,
	)
	if err != nil {
		return Person{}, fmt.Errorf("error inserting person: %w", err)
	}
	return createdPerson, nil
}
