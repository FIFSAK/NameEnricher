package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"NameEnricher/internal/models"
	"github.com/gin-gonic/gin"
)

func GetNationalitiesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := models.NationalityFilter{}

		if idStr := c.Query("id"); idStr != "" {
			if idVal, err := strconv.Atoi(idStr); err == nil && idVal > 0 {
				filter.ID = idVal
			}
		}

		if name := c.Query("name"); name != "" {
			filter.Name = name
		}

		if pageStr := c.Query("page"); pageStr != "" {
			if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
				filter.Page = pageVal
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
				filter.Limit = limitVal
			}
		}

		nationalities, err := models.GetNationalities(db, c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		c.JSON(http.StatusOK, nationalities)
	}
}

func GetNationalityByIDHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		nationalities, err := models.GetNationalities(db, c.Request.Context(), models.NationalityFilter{ID: id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during getting": err.Error()})
			return
		}

		if len(nationalities) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "nationality not found"})
			return
		}

		c.JSON(http.StatusOK, nationalities[0])
	}
}

func CreateNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		nationality, err := models.CreateNationality(db, c.Request.Context(), request.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during creation": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, nationality)
	}
}

func UpdateNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		var patch models.PatchNationality
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error during handling request": err.Error()})
			return
		}

		nationality, err := models.UpdateNationality(db, c.Request.Context(), id, patch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during update": err.Error()})
			return
		}

		c.JSON(http.StatusOK, nationality)
	}
}

func DeleteNationalityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error Wrong ID format": err.Error()})
			return
		}

		nationality, err := models.DeleteNationality(db, c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error during delete": err.Error()})
			return
		}

		c.JSON(http.StatusOK, nationality)
	}
}
