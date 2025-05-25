package handlers

import (
	"NameEnricher/internal/models"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetPersonsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := models.PersonFilter{}

		if idStr := c.Query("id"); idStr != "" {
			if idVal, err := strconv.ParseUint(idStr, 10, 32); err == nil && idVal > 0 {
				filter.ID = uint(idVal)
			}
		}

		if name := c.Query("name"); name != "" {
			filter.Name = name
		}

		if surname := c.Query("surname"); surname != "" {
			filter.Surname = surname
		}

		if patronymic := c.Query("patronymic"); patronymic != "" {
			filter.Patronymic = patronymic
		}

		if ageFromStr := c.Query("age_from"); ageFromStr != "" {
			if ageFromVal, err := strconv.Atoi(ageFromStr); err == nil && ageFromVal > 0 {
				filter.AgeFrom = ageFromVal
			}
		}

		if ageToStr := c.Query("age_to"); ageToStr != "" {
			if ageToVal, err := strconv.Atoi(ageToStr); err == nil && ageToVal > 0 {
				filter.AgeTo = ageToVal
			}
		}

		if genderIDStr := c.Query("gender_id"); genderIDStr != "" {
			if genderIDVal, err := strconv.Atoi(genderIDStr); err == nil && genderIDVal > 0 {
				filter.GenderID = genderIDVal
			}
		}

		if nationalityIDStr := c.Query("nationality_id"); nationalityIDStr != "" {
			if nationalityIDVal, err := strconv.Atoi(nationalityIDStr); err == nil && nationalityIDVal > 0 {
				filter.NationalityID = nationalityIDVal
			}
		}

		persons, err := models.GetPersons(c.Request.Context(), db, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		c.JSON(http.StatusOK, persons)
	}
}

func GetPersonByIDHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		filter := models.PersonFilter{ID: uint(id)}
		persons, err := models.GetPersons(c.Request.Context(), db, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		if len(persons) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
			return
		}

		c.JSON(http.StatusOK, persons[0])
	}
}

func CreatePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var person models.Person

		if err := c.ShouldBindJSON(&person); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		age, err := ageFromExternalApi(person.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting age": err.Error()})
			return
		}
		person.Age = age

		genderName, err := genderFromExternalApi(person.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting gender": err.Error()})
			return
		}

		genders, err := models.GetGenders(db, c.Request.Context(), models.GenderFilter{Name: genderName})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during checking gender": err.Error()})
			return
		}

		var genderID int
		if len(genders) == 0 {
			newGender, err := models.CreateGender(db, c.Request.Context(), genderName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error during creating gender": err.Error()})
				return
			}
			genderID = newGender.ID
		} else {
			genderID = genders[0].ID
		}
		person.Gender.ID = genderID

		nationalityCode, err := nationalityFromExternalApi(person.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting nationality": err.Error()})
			return
		}

		nationalities, err := models.GetNationalities(db, c.Request.Context(), models.NationalityFilter{Name: nationalityCode})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during checking nationality": err.Error()})
			return
		}

		var nationalityID int
		if len(nationalities) == 0 {
			newNationality, err := models.CreateNationality(db, c.Request.Context(), nationalityCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error during creating nationality": err.Error()})
				return
			}
			nationalityID = newNationality.ID
		} else {
			nationalityID = nationalities[0].ID
		}
		person.Nationality.ID = nationalityID

		createdPerson, err := models.CreatePerson(c.Request.Context(), person, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during creation": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, createdPerson)
	}
}

func UpdatePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		var patch models.PersonPatch
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		updatedPerson, err := models.UpdatePerson(c.Request.Context(), uint(id), patch, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during update": err.Error()})
			return
		}

		c.JSON(http.StatusOK, updatedPerson)
	}
}

func DeletePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		deletedPerson, err := models.DeletePersonByID(c.Request.Context(), uint(id), db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during delete": err.Error()})
			return
		}

		c.JSON(http.StatusOK, deletedPerson)
	}
}
