package models

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

func TestGetGenders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	genders := []Gender{
		{Name: "Male"},
		{Name: "Female"},
	}

	t.Run("GetAllGenders", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})
		for _, g := range genders {
			rows.AddRow(g.ID, g.Name)
		}

		mock.ExpectQuery("^SELECT id, name FROM genders WHERE 1=1$").
			WillReturnRows(rows)

		result, err := GetGenders(db, ctx, GenderFilter{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, genders) {
			t.Errorf("Results not matching received: %v, expected: %v", result, genders)
		}
	})

	t.Run("FilterById", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Male")

		mock.ExpectQuery("^SELECT id, name FROM genders WHERE 1=1 AND id = \\$1$").
			WithArgs(1).
			WillReturnRows(rows)

		result, err := GetGenders(db, ctx, GenderFilter{ID: 1})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Gender{{ID: 1, Name: "Male"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterByName", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Female")

		mock.ExpectQuery("^SELECT id, name FROM genders WHERE 1=1 AND name ILIKE \\$1$").
			WithArgs("%Female%").
			WillReturnRows(rows)

		result, err := GetGenders(db, ctx, GenderFilter{Name: "Female"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Gender{{ID: 2, Name: "Female"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("QueryError", func(t *testing.T) {
		mock.ExpectQuery("^SELECT id, name FROM genders WHERE 1=1$").
			WillReturnError(errors.New("error executing query"))

		_, err := GetGenders(db, ctx, GenderFilter{})
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestCreateGender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulCreate", func(t *testing.T) {
		name := "Non-binary"
		expectedGender := Gender{ID: 3, Name: name}

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(expectedGender.ID, expectedGender.Name)

		mock.ExpectQuery("^INSERT INTO genders").
			WithArgs(name).
			WillReturnRows(rows)

		result, err := CreateGender(db, ctx, name)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, expectedGender) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expectedGender)
		}
	})

	t.Run("CreateError", func(t *testing.T) {
		name := "Test"

		mock.ExpectQuery("^INSERT INTO genders").
			WithArgs(name).
			WillReturnError(errors.New("unique key violation"))

		_, err := CreateGender(db, ctx, name)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUpdateGender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		id := 1
		newName := "Other"
		patch := PatchGender{Name: &newName}
		currentGender := Gender{ID: id, Name: "Male"}
		updatedGender := Gender{ID: id, Name: newName}

		selRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(currentGender.ID, currentGender.Name)
		mock.ExpectQuery("^SELECT id, name FROM genders WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		updateRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(updatedGender.ID, updatedGender.Name)
		mock.ExpectQuery("^UPDATE genders SET name = \\$1 WHERE id = \\$2 RETURNING id, name$").
			WithArgs(newName, id).
			WillReturnRows(updateRows)

		result, err := UpdateGender(db, ctx, id, patch)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, updatedGender) {
			t.Errorf("Results not matching received: %v, expected: %v", result, updatedGender)
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		id := 1
		currentGender := Gender{ID: id, Name: "Male"}

		selRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(currentGender.ID, currentGender.Name)
		mock.ExpectQuery("^SELECT id, name FROM genders WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		patch := PatchGender{}

		result, err := UpdateGender(db, ctx, id, patch)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, currentGender) {
			t.Errorf("Results not matching received: %v, expected: %v", result, currentGender)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := 999
		newName := "Other"
		patch := PatchGender{Name: &newName}

		mock.ExpectQuery("^SELECT id, name FROM genders WHERE id = \\$1$").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := UpdateGender(db, ctx, id, patch)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestDeleteGender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulDelete", func(t *testing.T) {
		id := 1
		deletedGender := Gender{ID: id, Name: "Male"}

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(deletedGender.ID, deletedGender.Name)

		mock.ExpectQuery("^DELETE FROM genders WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(rows)

		result, err := DeleteGender(db, ctx, id)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, deletedGender) {
			t.Errorf("Results not matching received: %v, expected: %v", result, deletedGender)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := 999

		mock.ExpectQuery("^DELETE FROM genders WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := DeleteGender(db, ctx, id)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
