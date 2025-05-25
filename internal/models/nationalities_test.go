package models

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

func TestGetNationalities(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	nationalities := []Nationality{
		{ID: 1, Name: "Russian"},
		{ID: 2, Name: "American"},
	}

	t.Run("GetAllNationalities", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})
		for _, n := range nationalities {
			rows.AddRow(n.ID, n.Name)
		}

		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE 1=1$").
			WillReturnRows(rows)

		result, err := GetNationalities(db, ctx, NationalityFilter{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, nationalities) {
			t.Errorf("Results not matching received: %v, expected: %v", result, nationalities)
		}
	})

	t.Run("FilterById", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Russian")

		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE 1=1 AND id = \\$1$").
			WithArgs(1).
			WillReturnRows(rows)

		result, err := GetNationalities(db, ctx, NationalityFilter{ID: 1})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Nationality{{ID: 1, Name: "Russian"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("FilterByName", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "American")

		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE 1=1 AND name ILIKE \\$1$").
			WithArgs("%American%").
			WillReturnRows(rows)

		result, err := GetNationalities(db, ctx, NationalityFilter{Name: "American"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Nationality{{ID: 2, Name: "American"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "American")

		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE 1=1 LIMIT \\$1 OFFSET \\$2$").
			WithArgs(10, 10).
			WillReturnRows(rows)

		result, err := GetNationalities(db, ctx, NationalityFilter{Page: 2, Limit: 10})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []Nationality{{ID: 2, Name: "American"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expected)
		}
	})

	t.Run("QueryError", func(t *testing.T) {
		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE 1=1$").
			WillReturnError(errors.New("error executing query"))

		_, err := GetNationalities(db, ctx, NationalityFilter{})
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestCreateNationality(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulCreate", func(t *testing.T) {
		name := "French"
		expectedNationality := Nationality{ID: 3, Name: name}

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(expectedNationality.ID, expectedNationality.Name)

		mock.ExpectQuery("^INSERT INTO nationalities").
			WithArgs(name).
			WillReturnRows(rows)

		result, err := CreateNationality(db, ctx, name)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, expectedNationality) {
			t.Errorf("Results not matching received: %v, expected: %v", result, expectedNationality)
		}
	})

	t.Run("CreateError", func(t *testing.T) {
		name := "Test"

		mock.ExpectQuery("^INSERT INTO nationalities").
			WithArgs(name).
			WillReturnError(errors.New("unique key violation"))

		_, err := CreateNationality(db, ctx, name)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUpdateNationality(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		id := 1
		newName := "Italian"
		patch := PatchNationality{Name: &newName}
		currentNationality := Nationality{ID: id, Name: "Russian"}
		updatedNationality := Nationality{ID: id, Name: newName}

		selRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(currentNationality.ID, currentNationality.Name)
		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		updateRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(updatedNationality.ID, updatedNationality.Name)
		mock.ExpectQuery("^UPDATE nationalities SET name = \\$1 WHERE id = \\$2 RETURNING id, name$").
			WithArgs(newName, id).
			WillReturnRows(updateRows)

		result, err := UpdateNationality(db, ctx, id, patch)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, updatedNationality) {
			t.Errorf("Results not matching received: %v, expected: %v", result, updatedNationality)
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		id := 1
		currentNationality := Nationality{ID: id, Name: "Russian"}

		selRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(currentNationality.ID, currentNationality.Name)
		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE id = \\$1$").
			WithArgs(id).
			WillReturnRows(selRows)

		patch := PatchNationality{}

		result, err := UpdateNationality(db, ctx, id, patch)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, currentNationality) {
			t.Errorf("Results not matching received: %v, expected: %v", result, currentNationality)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := 999
		newName := "German"
		patch := PatchNationality{Name: &newName}

		mock.ExpectQuery("^SELECT id, name FROM nationalities WHERE id = \\$1$").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := UpdateNationality(db, ctx, id, patch)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestDeleteNationality(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("SuccessfulDelete", func(t *testing.T) {
		id := 1
		deletedNationality := Nationality{ID: id, Name: "Russian"}

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(deletedNationality.ID, deletedNationality.Name)

		mock.ExpectQuery("^DELETE FROM nationalities WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(rows)

		result, err := DeleteNationality(db, ctx, id)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(result, deletedNationality) {
			t.Errorf("Results not matching received: %v, expected: %v", result, deletedNationality)
		}
	})

	t.Run("NonExistentId", func(t *testing.T) {
		id := 999

		mock.ExpectQuery("^DELETE FROM nationalities WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := DeleteNationality(db, ctx, id)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
