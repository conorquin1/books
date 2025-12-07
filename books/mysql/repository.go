package mysql

import (
	"github.com/books/books"
	"github.com/jmoiron/sqlx"
)

// RepositoryProvider manages all repositories.
type RepositoryProvider struct {
	db *sqlx.DB
}

// NewRepositoryProvider returns a new RepositoryProvider.
func NewRepositoryProvider(db *sqlx.DB) *RepositoryProvider {
	return &RepositoryProvider{db: db}
}

// Book returns a new BookRepository.
func (rp *RepositoryProvider) Book() books.BookRepository {
	return NewBookRepository(rp.db)
}

