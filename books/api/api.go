package api

import (
	"github.com/books/books"
	"github.com/books/books/cache"
	"github.com/books/books/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// InitRoutes initializes all API routes.
func InitRoutes(g *echo.Group, db *sqlx.DB, c *cache.Cache) {
	// Create repository provider
	repoProvider := mysql.NewRepositoryProvider(db)

	// Create service
	bookService := books.NewBookService(repoProvider, c)

	// Create and register controllers
	bookController := newBookController(bookService)
	bookController.Routes(g)
}

