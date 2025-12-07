package testdata

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/books/books"
)

// InsertBook inserts a book into the database
func (s *Suite) InsertBook(book books.Book) *books.Book {
	if s.db == nil {
		s.t.Fatal("Database not initialized. Call WithDB() first.")
	}

	now := time.Now().UTC()
	if book.CreatedAt.IsZero() {
		book.CreatedAt = now
	}
	if book.UpdatedAt.IsZero() {
		book.UpdatedAt = now
	}

	result, err := s.db.NamedExec(`
		INSERT INTO books (title, author, isbn, description, publishedAt, createdAt, updatedAt, deletedAt)
		VALUES (:title, :author, :isbn, :description, :publishedAt, :createdAt, :updatedAt, :deletedAt)
	`, book)
	if err != nil {
		s.t.Fatalf("Failed to insert book: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		s.t.Fatalf("Failed to get last insert id: %v", err)
	}

	book.ID = id
	return &book
}

// GetBook retrieves a book by ID from the database
func (s *Suite) GetBook(id int64) *books.Book {
	if s.db == nil {
		s.t.Fatal("Database not initialized. Call WithDB() first.")
	}

	var book books.Book
	err := s.db.Get(&book, "SELECT * FROM books WHERE id = ? AND deletedAt IS NULL", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		s.t.Fatalf("Failed to get book: %v", err)
	}

	return &book
}

// ClearBooks clears all books from the database and clears related cache
func (s *Suite) ClearBooks() {
	if s.db == nil {
		return
	}

	if _, err := s.db.Exec("DELETE FROM books"); err != nil {
		s.t.Fatalf("Failed to clear books table: %v", err)
	}

	// Reset auto increment
	if _, err := s.db.Exec("ALTER TABLE books AUTO_INCREMENT = 1"); err != nil {
		// Ignore error if table doesn't exist or doesn't have auto increment
		s.t.Logf("Warning: Failed to reset auto increment: %v", err)
	}

	// Flush Redis database
	if s.cache != nil {
		if err := s.cache.FlushDB(context.Background()); err != nil {
			s.t.Logf("Warning: Failed to flush Redis database: %v", err)
		}
	}
}
