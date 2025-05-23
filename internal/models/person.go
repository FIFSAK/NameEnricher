package models

import (
	"NameEnricher/pkg/logger"
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
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
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
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		persons = append(persons, person)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return persons, nil
}

func DeletePersonByID(ctx context.Context, id uint, db *sql.DB) (Person, error) {
	var deletedPerson Person
	err := db.QueryRowContext(ctx, "DELETE FROM persons WHERE id = $1", id).Scan(&deletedPerson)
	if err != nil {
		return Person{}, err
	}
	return deletedPerson, nil
}

func UpdatePerson
