package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/books/books"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// BookRepository contains all methods to access the books table.
type BookRepository struct {
	db *sqlx.DB
}

// NewBookRepository returns a new BookRepository.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

// Create creates a new book in the database.
func (r *BookRepository) Create(ctx context.Context, book books.Book) (*books.Book, error) {
	result, err := r.db.NamedExecContext(ctx, `
		INSERT INTO books (
			title,
			author,
			isbn,
			description,
			publishedAt,
			createdAt,
			updatedAt
		) VALUES (
			:title,
			:author,
			:isbn,
			:description,
			:publishedAt,
			:createdAt,
			:updatedAt
		)
	`, map[string]interface{}{
		"title":       book.Title,
		"author":      book.Author,
		"isbn":        book.ISBN,
		"description": book.Description,
		"publishedAt": book.PublishedAt,
		"createdAt":   book.CreatedAt,
		"updatedAt":   book.UpdatedAt,
	})
	if err != nil {
		// Check for duplicate key error (MySQL error code 1062)
		if mysqlErr, ok := errors.Cause(err).(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, books.ErrBookAlreadyExists
		}
		return nil, errors.WithStack(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	book.ID = id
	return &book, nil
}

// GetByID retrieves a book by its ID.
func (r *BookRepository) GetByID(ctx context.Context, id int64) (*books.Book, error) {
	var book books.Book
	err := r.db.GetContext(ctx, &book, `
		SELECT 
			id,
			title,
			author,
			isbn,
			description,
			publishedAt,
			createdAt,
			updatedAt,
			deletedAt
		FROM books
		WHERE id = ?
		AND deletedAt IS NULL
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, books.ErrBookNotFound
		}
		return nil, errors.WithStack(err)
	}
	return &book, nil
}

// GetAll retrieves all books (excluding deleted ones).
// If author is provided, filters books by that author.
// limit and offset are used for pagination. If limit is 0, no limit is applied.
func (r *BookRepository) GetAll(ctx context.Context, author *string, limit, offset int) ([]books.Book, error) {
	bookList := []books.Book{}
	query := `
		SELECT 
			id,
			title,
			author,
			isbn,
			description,
			publishedAt,
			createdAt,
			updatedAt,
			deletedAt
		FROM books
		WHERE deletedAt IS NULL
	`
	args := []interface{}{}
	
	if author != nil && *author != "" {
		query += ` AND author = ?`
		args = append(args, *author)
	}
	
	query += ` ORDER BY id ASC`
	
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
		
		if offset > 0 {
			query += ` OFFSET ?`
			args = append(args, offset)
		}
	}
		
	err := r.db.SelectContext(ctx, &bookList, query, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	
	// Ensure we always return a non-nil slice (empty slice instead of nil)
	// This ensures JSON serialization produces [] instead of null
	if bookList == nil {
		bookList = []books.Book{}
	}
	
	return bookList, nil
}

// Update updates an existing book.
func (r *BookRepository) Update(ctx context.Context, id int64, book books.Book) (*books.Book, error) {
	// First check if book exists
	existingBook, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the book
	_, err = r.db.NamedExecContext(ctx, `
		UPDATE books
		SET 
			title = :title,
			author = :author,
			isbn = :isbn,
			description = :description,
			publishedAt = :publishedAt,
			updatedAt = :updatedAt
		WHERE id = :id
		AND deletedAt IS NULL
	`, map[string]interface{}{
		"id":          id,
		"title":       book.Title,
		"author":      book.Author,
		"isbn":        book.ISBN,
		"description": book.Description,
		"publishedAt": book.PublishedAt,
		"updatedAt":   book.UpdatedAt,
	})
	if err != nil {
		// Check for duplicate key error (MySQL error code 1062)
		if mysqlErr, ok := errors.Cause(err).(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, books.ErrBookAlreadyExists
		}
		return nil, errors.WithStack(err)
	}

	// Return updated book
	updatedBook := *existingBook
	updatedBook.Title = book.Title
	updatedBook.Author = book.Author
	updatedBook.ISBN = book.ISBN
	updatedBook.Description = book.Description
	updatedBook.PublishedAt = book.PublishedAt
	updatedBook.UpdatedAt = book.UpdatedAt

	return &updatedBook, nil
}

// Delete soft deletes a book by setting deletedAt.
func (r *BookRepository) Delete(ctx context.Context, id int64) error {
	// First check if book exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.db.NamedExecContext(ctx, `
		UPDATE books
		SET 
			deletedAt = :deletedAt
		WHERE id = :id
		AND deletedAt IS NULL
	`, map[string]interface{}{
		"id":        id,
		"deletedAt": now,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
