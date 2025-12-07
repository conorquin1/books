package testdata

import (
	"fmt"
	"testing"

	"github.com/books/books"
	"github.com/books/books/cache"
	"github.com/books/books/mysql"
	"github.com/books/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// Suite holds integration test utilities
type Suite struct {
	t            *testing.T
	db           *sqlx.DB
	cache        *cache.Cache
	echo         *echo.Echo
	repoProvider *mysql.RepositoryProvider
	service      *books.BookService
}

// NewSuite returns a new test suite
func NewSuite(t *testing.T) *Suite {
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return &Suite{
		t: t,
	}
}

// WithDB initialises the test database connection
func (s *Suite) WithDB() *Suite {
	// Get config values with defaults
	host := viper.GetString("test.db.host")
	if host == "" {
		host = "127.0.0.1"
	}
	
	user := viper.GetString("test.db.user")
	if user == "" {
		user = "root"
	}
	
	password := viper.GetString("test.db.password")
	if password == "" {
		password = "mysecretpassword"
	}
	
	databaseName := viper.GetString("test.db.database")
	if databaseName == "" {
		databaseName = "mysql"
	}
	
	port := viper.GetInt64("test.db.port")
	if port == 0 {
		port = 3306
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, databaseName)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		s.t.Fatalf("Failed to connect to test database: %v (DSN: %s)", err, dsn)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		s.t.Fatalf("Failed to ping test database: %v", err)
	}

	s.db = db
	return s
}

// WithCache initialises Redis cache
func (s *Suite) WithCache() *Suite {
	redisCache, err := cache.NewCache()
	if err != nil {
		// Cache is optional for tests, so we just log but don't fail
		s.t.Logf("Warning: Failed to initialize cache: %v (tests will run without caching)", err)
		redisCache = nil
	}
	s.cache = redisCache
	return s
}

// Close closes all connections
func (s *Suite) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// DB returns the database connection
func (s *Suite) DB() *sqlx.DB {
	return s.db
}

// Cache returns the cache instance
func (s *Suite) Cache() *cache.Cache {
	return s.cache
}

// SetupAPI initializes the repository provider, service, and echo instance
// Returns the echo instance, repository provider, and service for controller setup
func (s *Suite) SetupAPI() (*echo.Echo, *mysql.RepositoryProvider, *books.BookService) {
	if s.db == nil {
		s.t.Fatal("Database not initialized. Call WithDB() first.")
	}

	s.repoProvider = mysql.NewRepositoryProvider(s.db)
	s.service = books.NewBookService(s.repoProvider, s.cache)
	s.echo = echo.New()

	return s.echo, s.repoProvider, s.service
}

// Echo returns the echo instance (must call SetupAPI first)
func (s *Suite) Echo() *echo.Echo {
	if s.echo == nil {
		s.t.Fatal("Echo not initialized. Call SetupAPI() first.")
	}
	return s.echo
}

// Clear empties test data from the database
func (s *Suite) Clear() {
	if s.db == nil {
		return
	}

	// Clear books table
	if _, err := s.db.Exec("DELETE FROM books"); err != nil {
		s.t.Fatalf("Failed to clear books table: %v", err)
	}
	
	// Reset auto increment
	if _, err := s.db.Exec("ALTER TABLE books AUTO_INCREMENT = 1"); err != nil {
		// Ignore error if table doesn't exist or doesn't have auto increment
		s.t.Logf("Warning: Failed to reset auto increment: %v", err)
	}
}

