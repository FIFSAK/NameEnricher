// Package handlers provides HTTP request handlers for the API endpoints.
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"NameEnricher/internal/models"
	"NameEnricher/pkg/logger"
	"github.com/gin-gonic/gin"
)

// GetGendersHandler godoc
// @Summary List genders
// @Description Get a list of genders with optional filtering
// @Tags genders
// @Accept json
// @Produce json
// @Param id query integer false "Gender ID"
// @Param name query string false "Gender name"
// @Param page query integer false "Page number for pagination"
// @Param limit query integer false "Number of items per page"
// @Success 200 {array} models.Gender "Successfully retrieved gender list"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genders [get]
func GetGendersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing get genders request")
		filter := models.GenderFilter{}

		if idStr := c.Query("id"); idStr != "" {
			if idVal, err := strconv.Atoi(idStr); err == nil && idVal > 0 {
				filter.ID = idVal
				logger.Log.Debugf("Filtering by ID: %d", idVal)
			}
		}

		if name := c.Query("name"); name != "" {
			filter.Name = name
			logger.Log.Debugf("Filtering by name: %s", name)
		}

		if pageStr := c.Query("page"); pageStr != "" {
			if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
				filter.Page = pageVal
				logger.Log.Debugf("Filtering by page: %d", pageVal)
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
				filter.Limit = limitVal
				logger.Log.Debugf("Filtering by limit: %d", limitVal)
			}
		}

		logger.Log.Debugf("Executing GetGenders with filter: %+v", filter)
		genders, err := models.GetGenders(db, c.Request.Context(), filter)
		if err != nil {
			logger.Log.Errorf("Failed to get genders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		logger.Log.Infof("Successfully retrieved %d genders", len(genders))
		c.JSON(http.StatusOK, genders)
	}
}

// GetGenderByIDHandler godoc
// @Summary Get a gender by ID
// @Description Get a single gender by its ID
// @Tags genders
// @Accept json
// @Produce json
// @Param id path integer true "Gender ID"
// @Success 200 {object} models.Gender "Successfully retrieved gender"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 404 {object} map[string]string "Gender not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genders/{id} [get]
func GetGenderByIDHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing get gender by ID request: %s", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		logger.Log.Debugf("Looking up gender with ID: %d", id)
		genders, err := models.GetGenders(db, c.Request.Context(), models.GenderFilter{ID: id})
		if err != nil {
			logger.Log.Errorf("Failed to get gender by ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		if len(genders) == 0 {
			logger.Log.Warnf("Gender with ID %d not found", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "gender not found"})
			return
		}

		logger.Log.Infof("Successfully retrieved gender with ID %d", id)
		c.JSON(http.StatusOK, genders[0])
	}
}

// CreateGenderHandler godoc
// @Summary Create a new gender
// @Description Create a new gender entry in the database
// @Tags genders
// @Accept json
// @Produce json
// @Param gender body object true "Gender object with name field"
// @Success 201 {object} models.Gender "Successfully created gender"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genders [post]
func CreateGenderHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing create gender request")

		var request struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Log.Errorf("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Creating gender with name: %s", request.Name)

		gender, err := models.CreateGender(db, c.Request.Context(), request.Name)
		if err != nil {
			logger.Log.Errorf("Failed to create gender '%s': %v", request.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during creation": err.Error()})
			return
		}

		logger.Log.Infof("Successfully created gender '%s' with ID %d", request.Name, gender.ID)
		c.JSON(http.StatusCreated, gender)
	}
}

// UpdateGenderHandler godoc
// @Summary Update a gender
// @Description Update an existing gender by ID
// @Tags genders
// @Accept json
// @Produce json
// @Param id path integer true "Gender ID"
// @Param gender body models.PatchGender true "Gender update data"
// @Success 200 {object} models.Gender "Successfully updated gender"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genders/{id} [put]
func UpdateGenderHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing update gender request for ID: %s", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		var patch models.PatchGender
		if err := c.ShouldBindJSON(&patch); err != nil {
			logger.Log.Errorf("Failed to bind JSON for gender update: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Updating gender ID %d with patch: %+v", id, patch)

		gender, err := models.UpdateGender(db, c.Request.Context(), id, patch)
		if err != nil {
			logger.Log.Errorf("Failed to update gender ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during update": err.Error()})
			return
		}

		logger.Log.Infof("Successfully updated gender with ID %d", id)
		c.JSON(http.StatusOK, gender)
	}
}

// DeleteGenderHandler godoc
// @Summary Delete a gender
// @Description Delete a gender by its ID
// @Tags genders
// @Accept json
// @Produce json
// @Param id path integer true "Gender ID"
// @Success 200 {object} models.Gender "Successfully deleted gender"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genders/{id} [delete]
func DeleteGenderHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing delete gender request for ID: %s", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		logger.Log.Debugf("Deleting gender with ID: %d", id)
		gender, err := models.DeleteGender(db, c.Request.Context(), id)
		if err != nil {
			logger.Log.Errorf("Failed to delete gender ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during delete": err.Error()})
			return
		}

		logger.Log.Infof("Successfully deleted gender with ID %d", id)
		c.JSON(http.StatusOK, gender)
	}
}
