package books

import (
	"context"
	"time"
)

// Book represents a book in the catalog.
type Book struct {
	ID          int64     `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Author      string    `db:"author" json:"author"`
	ISBN        string    `db:"isbn" json:"isbn"`
	Description string    `db:"description" json:"description"`
	PublishedAt time.Time `db:"publishedAt" json:"publishedAt"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deletedAt" json:"deletedAt,omitempty"`
}

// BookRepository contains all methods to access the books table.
type BookRepository interface {
	// Create creates a new book in the database.
	Create(ctx context.Context, book Book) (*Book, error)

	// GetByID retrieves a book by its ID.
	GetByID(ctx context.Context, id int64) (*Book, error)

	// GetAll retrieves all books (excluding deleted ones).
	// If author is provided, filters books by that author.
	// limit and offset are used for pagination. If limit is 0, no limit is applied.
	GetAll(ctx context.Context, author *string, limit, offset int) ([]Book, error)

	// Update updates an existing book.
	Update(ctx context.Context, id int64, book Book) (*Book, error)

	// Delete soft deletes a book by setting deletedAt.
	Delete(ctx context.Context, id int64) error
}

