// Package handlers provides HTTP request handlers for the API endpoints.
package handlers

import (
	"NameEnricher/internal/models"
	"NameEnricher/pkg/logger"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetPersonsHandler godoc
// @Summary List persons
// @Description Get a list of persons with optional filtering
// @Tags persons
// @Accept json
// @Produce json
// @Param id query integer false "Person ID"
// @Param name query string false "Person name"
// @Param surname query string false "Person surname"
// @Param patronymic query string false "Person patronymic"
// @Param age_from query integer false "Minimum age"
// @Param age_to query integer false "Maximum age"
// @Param gender_id query integer false "Gender ID"
// @Param nationality_id query integer false "Nationality ID"
// @Success 200 {array} models.Person "Successfully retrieved person list"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /persons [get]
func GetPersonsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing get persons request")
		filter := models.PersonFilter{}

		if idStr := c.Query("id"); idStr != "" {
			if idVal, err := strconv.ParseUint(idStr, 10, 32); err == nil && idVal > 0 {
				filter.ID = uint(idVal)
				logger.Log.Debugf("Filtering by ID: %d", idVal)
			}
		}

		if name := c.Query("name"); name != "" {
			filter.Name = name
			logger.Log.Debugf("Filtering by name: %s", name)
		}

		if surname := c.Query("surname"); surname != "" {
			filter.Surname = surname
			logger.Log.Debugf("Filtering by surname: %s", surname)
		}

		if patronymic := c.Query("patronymic"); patronymic != "" {
			filter.Patronymic = patronymic
			logger.Log.Debugf("Filtering by patronymic: %s", patronymic)
		}

		if ageFromStr := c.Query("age_from"); ageFromStr != "" {
			if ageFromVal, err := strconv.Atoi(ageFromStr); err == nil && ageFromVal > 0 {
				filter.AgeFrom = ageFromVal
				logger.Log.Debugf("Filtering by age from: %d", ageFromVal)
			}
		}

		if ageToStr := c.Query("age_to"); ageToStr != "" {
			if ageToVal, err := strconv.Atoi(ageToStr); err == nil && ageToVal > 0 {
				filter.AgeTo = ageToVal
				logger.Log.Debugf("Filtering by age to: %d", ageToVal)
			}
		}

		if genderIDStr := c.Query("gender_id"); genderIDStr != "" {
			if genderIDVal, err := strconv.Atoi(genderIDStr); err == nil && genderIDVal > 0 {
				filter.GenderID = genderIDVal
				logger.Log.Debugf("Filtering by gender ID: %d", genderIDVal)
			}
		}

		if nationalityIDStr := c.Query("nationality_id"); nationalityIDStr != "" {
			if nationalityIDVal, err := strconv.Atoi(nationalityIDStr); err == nil && nationalityIDVal > 0 {
				filter.NationalityID = nationalityIDVal
				logger.Log.Debugf("Filtering by nationality ID: %d", nationalityIDVal)
			}
		}

		logger.Log.Debugf("Executing GetPersons with filter: %+v", filter)
		persons, err := models.GetPersons(c.Request.Context(), db, filter)
		if err != nil {
			logger.Log.Errorf("Failed to get persons: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		logger.Log.Infof("Successfully retrieved %d persons", len(persons))
		c.JSON(http.StatusOK, persons)
	}
}

// CreatePersonHandler godoc
// @Summary Create a new person
// @Description Create a new person with automatic enrichment of age, gender, and nationality
// @Tags persons
// @Accept json
// @Produce json
// @Param person body models.Person true "Person data (name is required for enrichment)"
// @Success 201 {object} models.Person "Successfully created person"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /persons [post]
func CreatePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing create person request")

		var person models.Person
		if err := c.ShouldBindJSON(&person); err != nil {
			logger.Log.Errorf("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Creating person with name: %s, surname: %s", person.Name, person.Surname)

		age, err := ageFromExternalApi(person.Name)
		if err != nil {
			logger.Log.Errorf("Failed to get age for name %s: %v", person.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting age": err.Error()})
			return
		}
		logger.Log.Debugf("Retrieved age %d for name %s", age, person.Name)
		person.Age = age

		genderName, err := genderFromExternalApi(person.Name)
		if err != nil {
			logger.Log.Errorf("Failed to get gender for name %s: %v", person.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting gender": err.Error()})
			return
		}
		logger.Log.Debugf("Retrieved gender '%s' for name %s", genderName, person.Name)

		logger.Log.Debugf("Looking up gender '%s' in database", genderName)
		genders, err := models.GetGenders(db, c.Request.Context(), models.GenderFilter{Name: genderName})
		if err != nil {
			logger.Log.Errorf("Failed to check gender in database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during checking gender": err.Error()})
			return
		}

		var genderID int
		if len(genders) == 0 {
			logger.Log.Infof("Gender '%s' not found in database, creating new entry", genderName)
			newGender, err := models.CreateGender(db, c.Request.Context(), genderName)
			if err != nil {
				logger.Log.Errorf("Failed to create gender '%s': %v", genderName, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error during creating gender": err.Error()})
				return
			}
			genderID = newGender.ID
			logger.Log.Infof("Created new gender '%s' with ID %d", genderName, genderID)
		} else {
			genderID = genders[0].ID
			logger.Log.Infof("Using existing gender '%s' with ID %d", genderName, genderID)
		}
		person.Gender.ID = genderID

		nationalityCode, err := nationalityFromExternalApi(person.Name)
		if err != nil {
			logger.Log.Errorf("Failed to get nationality for name %s: %v", person.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting nationality": err.Error()})
			return
		}
		logger.Log.Debugf("Retrieved nationality '%s' for name %s", nationalityCode, person.Name)

		logger.Log.Debugf("Looking up nationality '%s' in database", nationalityCode)
		nationalities, err := models.GetNationalities(db, c.Request.Context(), models.NationalityFilter{Name: nationalityCode})
		if err != nil {
			logger.Log.Errorf("Failed to check nationality in database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during checking nationality": err.Error()})
			return
		}

		var nationalityID int
		if len(nationalities) == 0 {
			logger.Log.Infof("Nationality '%s' not found in database, creating new entry", nationalityCode)
			newNationality, err := models.CreateNationality(db, c.Request.Context(), nationalityCode)
			if err != nil {
				logger.Log.Errorf("Failed to create nationality '%s': %v", nationalityCode, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error during creating nationality": err.Error()})
				return
			}
			nationalityID = newNationality.ID
			logger.Log.Infof("Created new nationality '%s' with ID %d", nationalityCode, nationalityID)
		} else {
			nationalityID = nationalities[0].ID
			logger.Log.Infof("Using existing nationality '%s' with ID %d", nationalityCode, nationalityID)
		}
		person.Nationality.ID = nationalityID

		logger.Log.Debugf("Saving person to database")
		createdPerson, err := models.CreatePerson(c.Request.Context(), person, db)
		if err != nil {
			logger.Log.Errorf("Failed to create person: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during creation": err.Error()})
			return
		}

		logger.Log.Infof("Successfully created person with ID %d", createdPerson.ID)
		c.JSON(http.StatusCreated, createdPerson)
	}
}

// UpdatePersonHandler godoc
// @Summary Update a person
// @Description Update an existing person by ID
// @Tags persons
// @Accept json
// @Produce json
// @Param id path integer true "Person ID"
// @Param person body models.PersonPatch true "Person update data"
// @Success 200 {object} models.Person "Successfully updated person"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /persons/{id} [put]
func UpdatePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing update person request for ID: %s", idStr)

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		var patch models.PersonPatch
		if err := c.ShouldBindJSON(&patch); err != nil {
			logger.Log.Errorf("Failed to bind JSON for person update: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Updating person ID %d with patch: %+v", id, patch)

		updatedPerson, err := models.UpdatePerson(c.Request.Context(), uint(id), patch, db)
		if err != nil {
			logger.Log.Errorf("Failed to update person ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during update": err.Error()})
			return
		}

		logger.Log.Infof("Successfully updated person with ID %d", id)
		c.JSON(http.StatusOK, updatedPerson)
	}
}

// DeletePersonHandler godoc
// @Summary Delete a person
// @Description Delete a person by their ID
// @Tags persons
// @Accept json
// @Produce json
// @Param id path integer true "Person ID"
// @Success 200 {object} models.Person "Successfully deleted person"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /persons/{id} [delete]
func DeletePersonHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing delete person request for ID: %s", idStr)

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		logger.Log.Debugf("Deleting person with ID: %d", id)
		deletedPerson, err := models.DeletePersonByID(c.Request.Context(), uint(id), db)
		if err != nil {
			logger.Log.Errorf("Failed to delete person ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during delete": err.Error()})
			return
		}

		logger.Log.Infof("Successfully deleted person with ID %d", id)
		c.JSON(http.StatusOK, deletedPerson)
	}
}

// todo: fix swagger example in post
// todo: delete neccessary filters from swagger
