package main

import (
	_ "NameEnricher/docs"
	"NameEnricher/internal/handlers"
	"NameEnricher/pkg/logger"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
)

// @title           Name Enricher API
// @version         1.0
// @description     API for enriching names
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := godotenv.Load(); err != nil {
		logger.Log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		logger.Log.Fatal("DATABASE_DSN is not set in .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_DSN"))
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}
	logger.Log.Info("Database connected")

	if err = runMigrations(db); err != nil {
		logger.Log.WithError(err).Fatal("Failed to migrate database")
	}
	logger.Log.Info("Database migrated")

	router := gin.New()
	router.Use(gin.LoggerWithWriter(logger.Log.Writer()), gin.Recovery())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	gendersRouter := router.Group("/genders")
	gendersRouter.GET("", handlers.GetGendersHandler(db))
	gendersRouter.GET("/:id", handlers.GetGenderByIDHandler(db))
	gendersRouter.POST("", handlers.CreateGenderHandler(db))
	gendersRouter.PATCH("/:id", handlers.UpdateGenderHandler(db))
	gendersRouter.DELETE("/:id", handlers.DeleteGenderHandler(db))

	nationalitiesRouter := router.Group("/nationalities")
	nationalitiesRouter.GET("", handlers.GetNationalitiesHandler(db))
	nationalitiesRouter.GET("/:id", handlers.GetNationalityByIDHandler(db))
	nationalitiesRouter.POST("", handlers.CreateNationalityHandler(db))
	nationalitiesRouter.PATCH("/:id", handlers.UpdateNationalityHandler(db))
	nationalitiesRouter.DELETE("/:id", handlers.DeleteNationalityHandler(db))

	personsRouter := router.Group("/persons")
	personsRouter.GET("", handlers.GetPersonsHandler(db))
	personsRouter.POST("", handlers.CreatePersonHandler(db))
	personsRouter.PATCH("/:id", handlers.UpdatePersonHandler(db))
	personsRouter.DELETE("/:id", handlers.DeletePersonHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Log.Infof("Server running on port %s", port)
	router.Run(":" + port)
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
