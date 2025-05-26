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

// GetNationalitiesHandler godoc
// @Summary List nationalities
// @Description Get a list of nationalities with optional filtering
// @Tags nationalities
// @Accept json
// @Produce json
// @Param id query integer false "Nationality ID"
// @Param name query string false "Nationality name"
// @Param page query integer false "Page number for pagination"
// @Param limit query integer false "Number of items per page"
// @Success 200 {array} models.Nationality "Successfully retrieved nationality list"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /nationalities [get]
func GetNationalitiesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing get nationalities request")
		filter := models.NationalityFilter{}

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

		logger.Log.Debugf("Executing GetNationalities with filter: %+v", filter)
		nationalities, err := models.GetNationalities(db, c.Request.Context(), filter)
		if err != nil {
			logger.Log.Errorf("Failed to get nationalities: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		logger.Log.Infof("Successfully retrieved %d nationalities", len(nationalities))
		c.JSON(http.StatusOK, nationalities)
	}
}

// CreateNationalityHandler godoc
// @Summary Create a new nationality
// @Description Create a new nationality entry in the database
// @Tags nationalities
// @Accept json
// @Produce json
// @Param nationality body models.NationalityCreateRequest true "Nationality data (name is required for enrichment)"
// @Success 201 {object} models.Nationality "Successfully created nationality"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /nationalities [post]
func CreateNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Processing create nationality request")

		var request struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Log.Errorf("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Creating nationality with name: %s", request.Name)

		nationality, err := models.CreateNationality(db, c.Request.Context(), request.Name)
		if err != nil {
			logger.Log.Errorf("Failed to create nationality '%s': %v", request.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during creation": err.Error()})
			return
		}

		logger.Log.Infof("Successfully created nationality '%s' with ID %d", request.Name, nationality.ID)
		c.JSON(http.StatusCreated, nationality)
	}
}

// UpdateNationalityHandler godoc
// @Summary Update a nationality
// @Description Update an existing nationality by ID
// @Tags nationalities
// @Accept json
// @Produce json
// @Param id path integer true "Nationality ID"
// @Param nationality body models.PatchNationality true "Nationality update data"
// @Success 200 {object} models.Nationality "Successfully updated nationality"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /nationalities/{id} [put]
func UpdateNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing update nationality request for ID: %s", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		var patch models.PatchNationality
		if err := c.ShouldBindJSON(&patch); err != nil {
			logger.Log.Errorf("Failed to bind JSON for nationality update: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		logger.Log.Debugf("Updating nationality ID %d with patch: %+v", id, patch)

		nationality, err := models.UpdateNationality(db, c.Request.Context(), id, patch)
		if err != nil {
			logger.Log.Errorf("Failed to update nationality ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during update": err.Error()})
			return
		}

		logger.Log.Infof("Successfully updated nationality with ID %d", id)
		c.JSON(http.StatusOK, nationality)
	}
}

// DeleteNationalityHandler godoc
// @Summary Delete a nationality
// @Description Delete a nationality by its ID
// @Tags nationalities
// @Accept json
// @Produce json
// @Param id path integer true "Nationality ID"
// @Success 200 {object} models.Nationality "Successfully deleted nationality"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /nationalities/{id} [delete]
func DeleteNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		logger.Log.Infof("Processing delete nationality request for ID: %s", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Errorf("Invalid ID format: %s - %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		logger.Log.Debugf("Deleting nationality with ID: %d", id)
		nationality, err := models.DeleteNationality(db, c.Request.Context(), id)
		if err != nil {
			logger.Log.Errorf("Failed to delete nationality ID %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error during delete": err.Error()})
			return
		}

		logger.Log.Infof("Successfully deleted nationality with ID %d", id)
		c.JSON(http.StatusOK, nationality)
	}
}
