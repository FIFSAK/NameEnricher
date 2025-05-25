package models

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

func TestGetPersons(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	persons := []Person{
		{
			ID:         1,
			Name:       "John",
			Surname:    "Doe",
			Patronymic: "Smith",
			Age:        30,
			Gender: Gender{
				ID:   1,
				Name: "Male",
			},
			Nationality: Nationality{
				ID:   1,
				Name: "American",
			},
		},
		{
			ID:         2,
			Name:       "Jane",
			Surname:    "Smith",
			Patronymic: "Doe",
			Age:        25,
			Gender: Gender{
				ID:   2,
				Name: "Female",
			},
			Nationality: Nationality{
				ID:   2,
				Name: "Canadian",
			},
		},
	}

	t.Run("GetAllPersons", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		for _, p := range persons {
			rows.AddRow(p.ID, p.Name, p.Surname, p.Patronymic, p.Age, p.Gender.ID, p.Gender.Name, p.Nationality.ID, p.Nationality.Name)
		}

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1$`).WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, persons) {
			t.Errorf("Results not matching received: %v, expected: %v", result, persons)
		}
	})

	t.Run("FilterByName", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		rows.AddRow(persons[0].ID, persons[0].Name, persons[0].Surname, persons[0].Patronymic, persons[0].Age,
			persons[0].Gender.ID, persons[0].Gender.Name, persons[0].Nationality.ID, persons[0].Nationality.Name)

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1 AND p.name ILIKE \$1$`).WithArgs("%John%").WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{Name: "John"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Person{persons[0]}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterBySurname", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		rows.AddRow(persons[1].ID, persons[1].Name, persons[1].Surname, persons[1].Patronymic, persons[1].Age,
			persons[1].Gender.ID, persons[1].Gender.Name, persons[1].Nationality.ID, persons[1].Nationality.Name)

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1 AND p.surname ILIKE \$1$`).WithArgs("%Smith%").WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{Surname: "Smith"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Person{persons[1]}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterByAgeRange", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		rows.AddRow(persons[1].ID, persons[1].Name, persons[1].Surname, persons[1].Patronymic, persons[1].Age,
			persons[1].Gender.ID, persons[1].Gender.Name, persons[1].Nationality.ID, persons[1].Nationality.Name)

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1 AND p.age <= \$1 AND p.age >= \$2$`).WithArgs(26, 20).WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{AgeFrom: 20, AgeTo: 26})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Person{persons[1]}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterByGenderID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		rows.AddRow(persons[0].ID, persons[0].Name, persons[0].Surname, persons[0].Patronymic, persons[0].Age,
			persons[0].Gender.ID, persons[0].Gender.Name, persons[0].Nationality.ID, persons[0].Nationality.Name)

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1 AND p.gender_id = \$1$`).WithArgs(1).WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{GenderID: 1})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Person{persons[0]}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterByNationalityID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "gender_name", "nationality_id", "nationality_name"})
		rows.AddRow(persons[1].ID, persons[1].Name, persons[1].Surname, persons[1].Patronymic, persons[1].Age,
			persons[1].Gender.ID, persons[1].Gender.Name, persons[1].Nationality.ID, persons[1].Nationality.Name)

		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1 AND p.nationality_id = \$1$`).WithArgs(2).WillReturnRows(rows)

		result, err := GetPersons(ctx, db, PersonFilter{NationalityID: 2})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Person{persons[1]}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("QueryError", func(t *testing.T) {
		mock.ExpectQuery(`^SELECT p.id, p.name, p.surname, p.patronymic, p.age, p.gender_id, g.name as gender_name,
p.nationality_id, n.name as nationality_name
FROM persons p
LEFT JOIN nationalities n ON n.id = p.nationality_id
LEFT JOIN genders g ON g.id = p.gender_id
WHERE 1=1$`).WillReturnError(errors.New("database connection error"))

		_, err := GetPersons(ctx, db, PersonFilter{})
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestCreatePerson(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulCreate", func(t *testing.T) {
		person := Person{
			Name:       "Alex",
			Surname:    "Johnson",
			Patronymic: "Robert",
			Age:        35,
			Gender: Gender{
				ID: 1,
			},
			Nationality: Nationality{
				ID: 1,
			},
		}

		expectedPerson := person
		expectedPerson.ID = 3

		rows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "nationality_id"}).
			AddRow(expectedPerson.ID, expectedPerson.Name, expectedPerson.Surname, expectedPerson.Patronymic,
				expectedPerson.Age, expectedPerson.Gender.ID, expectedPerson.Nationality.ID)

		mock.ExpectQuery("INSERT INTO persons").
			WithArgs(person.Name, person.Surname, person.Patronymic, person.Age, person.Gender.ID, person.Nationality.ID).
			WillReturnRows(rows)

		result, err := CreatePerson(ctx, person, db)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != expectedPerson.ID || result.Name != expectedPerson.Name ||
			result.Surname != expectedPerson.Surname || result.Age != expectedPerson.Age {
			t.Errorf("Results not matching received: %v, expected: %v", result, expectedPerson)
		}
	})

	t.Run("CreateError", func(t *testing.T) {
		person := Person{
			Name:       "Error",
			Surname:    "Test",
			Patronymic: "",
			Age:        0,
			Gender: Gender{
				ID: 0,
			},
			Nationality: Nationality{
				ID: 0,
			},
		}

		mock.ExpectQuery("INSERT INTO persons").
			WithArgs(person.Name, person.Surname, person.Patronymic, person.Age, person.Gender.ID, person.Nationality.ID).
			WillReturnError(errors.New("constraint violation"))

		_, err := CreatePerson(ctx, person, db)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUpdatePerson(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		id := uint(1)
		name := "UpdatedName"
		age := 31
		patch := PersonPatch{
			Name: &name,
			Age:  &age,
		}

		currentPerson := Person{
			ID:         id,
			Name:       "John",
			Surname:    "Doe",
			Patronymic: "Smith",
			Age:        30,
			Gender: Gender{
				ID: 1,
			},
			Nationality: Nationality{
				ID: 1,
			},
		}

		updatedPerson := currentPerson
		updatedPerson.Name = name
		updatedPerson.Age = age

		selRows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "nationality_id"}).
			AddRow(currentPerson.ID, currentPerson.Name, currentPerson.Surname, currentPerson.Patronymic,
				currentPerson.Age, currentPerson.Gender.ID, currentPerson.Nationality.ID)
		mock.ExpectQuery("^SELECT id, name, surname, patronymic, age, gender_id, nationality_id FROM persons WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		updateRows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "nationality_id"}).
			AddRow(updatedPerson.ID, updatedPerson.Name, updatedPerson.Surname, updatedPerson.Patronymic,
				updatedPerson.Age, updatedPerson.Gender.ID, updatedPerson.Nationality.ID)
		mock.ExpectQuery("^UPDATE persons SET name = \\$1, age = \\$2 WHERE id = \\$3 RETURNING").
			WithArgs(name, age, id).
			WillReturnRows(updateRows)

		result, err := UpdatePerson(ctx, id, patch, db)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != updatedPerson.ID || result.Name != updatedPerson.Name || result.Age != updatedPerson.Age {
			t.Errorf("Results not matching received: %v, expected: %v", result, updatedPerson)
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		id := uint(1)
		patch := PersonPatch{} // Empty patch

		currentPerson := Person{
			ID:         id,
			Name:       "John",
			Surname:    "Doe",
			Patronymic: "Smith",
			Age:        30,
			Gender: Gender{
				ID: 1,
			},
			Nationality: Nationality{
				ID: 1,
			},
		}

		selRows := sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender_id", "nationality_id"}).
			AddRow(currentPerson.ID, currentPerson.Name, currentPerson.Surname, currentPerson.Patronymic,
				currentPerson.Age, currentPerson.Gender.ID, currentPerson.Nationality.ID)
		mock.ExpectQuery("^SELECT id, name, surname, patronymic, age, gender_id, nationality_id FROM persons WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		result, err := UpdatePerson(ctx, id, patch, db)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != currentPerson.ID || result.Name != currentPerson.Name || result.Age != currentPerson.Age {
			t.Errorf("Results not matching received: %v, expected: %v", result, currentPerson)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := uint(999)
		name := "UpdatedName"
		patch := PersonPatch{
			Name: &name,
		}

		mock.ExpectQuery("^SELECT id, name, surname, patronymic, age, gender_id, nationality_id FROM persons WHERE id = \\$1$").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := UpdatePerson(ctx, id, patch, db)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestDeletePersonByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulDelete", func(t *testing.T) {
		id := uint(1)
		deletedPerson := Person{
			ID:      id,
			Name:    "John",
			Surname: "Doe",
		}

		rows := sqlmock.NewRows([]string{"id", "name", "surname"}).
			AddRow(deletedPerson.ID, deletedPerson.Name, deletedPerson.Surname)

		mock.ExpectQuery("^DELETE FROM persons WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(rows)

		result, err := DeletePersonByID(ctx, id, db)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != deletedPerson.ID || result.Name != deletedPerson.Name {
			t.Errorf("Results not matching received: %v, expected: %v", result, deletedPerson)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := uint(999)

		mock.ExpectQuery("^DELETE FROM persons WHERE id = \\$1$").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := DeletePersonByID(ctx, id, db)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
