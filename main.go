package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/books/books/api"
	"github.com/books/books/cache"
	"github.com/books/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func main() {
	flag.Parse()

	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Setup database
	host := viper.GetString("db.host")
	user := viper.GetString("db.user")
	password := viper.GetString("db.password")
	databaseName := viper.GetString("db.database")
	port := viper.GetInt64("db.port")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, databaseName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewCache()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v (continuing without cache)", err)
		redisCache = nil
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// API routes
	v1 := e.Group("/api/v1")
	api.InitRoutes(v1, db, redisCache)

	// Start server
	serverPort := viper.GetString("server.port")
	log.Printf("Starting server on port %s", serverPort)
	if err := e.Start(":" + serverPort); err != nil {
		log.Fatal(err)
	}
}

