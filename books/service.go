package books

import (
	"context"
	"fmt"
	"time"

	"github.com/books/books/cache"
	"github.com/pkg/errors"
)

// RepositoryProvider manages all repositories.
type RepositoryProvider interface {
	Book() BookRepository
}

// BookService manages book operations.
type BookService struct {
	repo  RepositoryProvider
	cache *cache.Cache
}

// NewBookService returns a new BookService.
func NewBookService(repo RepositoryProvider, c *cache.Cache) *BookService {
	return &BookService{
		repo:  repo,
		cache: c,
	}
}

// Create creates a new book.
func (s *BookService) Create(ctx context.Context, book Book) (*Book, error) {
	// Set timestamps
	now := time.Now().UTC()
	book.CreatedAt = now
	book.UpdatedAt = now

	createdBook, err := s.repo.Book().Create(ctx, book)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Invalidate the "all books" cache after creating a new book
	if s.cache != nil {
		cacheKey := getAllCacheKey(nil) // nil means "all books"
		go func() {
			_ = s.cache.Delete(context.Background(), cacheKey)
		}()
	}

	return createdBook, nil
}

// GetByID retrieves a book by ID.
func (s *BookService) GetByID(ctx context.Context, id int64) (*Book, error) {
	// Try to get from cache first (if cache is available)
	if s.cache != nil {
		cacheKey := getByIDCacheKey(id)
		var book Book
		err := s.cache.Get(ctx, cacheKey, &book)
		if err == nil {
			// Cache hit, return cached value
			return &book, nil
		}
	}

	// Cache miss or error, fetch from database
	book, err := s.repo.Book().GetByID(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Write to cache asynchronously
	if s.cache != nil && book != nil {
		cacheKey := getByIDCacheKey(id)
		go func() {
			_ = s.cache.Set(
				context.Background(),
				cacheKey,
				book,
				time.Hour*1, // Cache for 1 hour
			)
		}()
	}

	return book, nil
}

// GetAll retrieves all books.
// If author is provided, filters books by that author.
// limit and offset are used for pagination. If limit is 0, no limit is applied.
func (s *BookService) GetAll(ctx context.Context, author *string, limit, offset int) ([]Book, error) {
	// Try to get from cache first (if cache is available and not paginated)
	// Note: We don't cache paginated results
	if s.cache != nil && limit == 0 && offset == 0 {
		cacheKey := getAllCacheKey(author)
		var books []Book
		err := s.cache.Get(ctx, cacheKey, &books)
		if err == nil {
			// Cache hit, return cached value
			return books, nil
		}
	}

	// Cache miss or error, fetch from database
	books, err := s.repo.Book().GetAll(ctx, author, limit, offset)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Write to cache asynchronously (only for non-paginated requests)
	if s.cache != nil && limit == 0 && offset == 0 {
		cacheKey := getAllCacheKey(author)
		go func() {
			_ = s.cache.Set(
				context.Background(),
				cacheKey,
				books,
				time.Hour*1, // Cache for 1 hour
			)
		}()
	}

	return books, nil
}

// getAllCacheKey generates a cache key for GetAll based on author filter.
func getAllCacheKey(author *string) string {
	if author != nil && *author != "" {
		return fmt.Sprintf("books:books:getall:author:%s", *author)
	}
	return "books:books:getall:all"
}

// getByIDCacheKey generates a cache key for GetByID based on book ID.
func getByIDCacheKey(id int64) string {
	return fmt.Sprintf("books:books:getbyid:%d", id)
}

// Update updates an existing book.
func (s *BookService) Update(ctx context.Context, id int64, book Book) (*Book, error) {
	// Validate required fields
	if book.Title == "" {
		return nil, errors.Wrap(ErrInvalidBookData, "title is required")
	}
	if book.Author == "" {
		return nil, errors.Wrap(ErrInvalidBookData, "author is required")
	}

	// Set updated timestamp
	book.UpdatedAt = time.Now().UTC()

	// Ensure publishedAt is set (required field in schema)
	if book.PublishedAt.IsZero() {
		book.PublishedAt = time.Now().UTC()
	}

	updatedBook, err := s.repo.Book().Update(ctx, id, book)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Invalidate the cache for this book and the "all books" cache after updating
	if s.cache != nil {
		bookCacheKey := getByIDCacheKey(id)
		allBooksCacheKey := getAllCacheKey(nil)
		go func() {
			_ = s.cache.Delete(context.Background(), bookCacheKey)
			_ = s.cache.Delete(context.Background(), allBooksCacheKey)
		}()
	}

	return updatedBook, nil
}

// Delete soft deletes a book.
func (s *BookService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Book().Delete(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}

	// Invalidate the cache for this book and the "all books" cache after deleting
	if s.cache != nil {
		bookCacheKey := getByIDCacheKey(id)
		allBooksCacheKey := getAllCacheKey(nil)
		go func() {
			_ = s.cache.Delete(context.Background(), bookCacheKey)
			_ = s.cache.Delete(context.Background(), allBooksCacheKey)
		}()
	}

	return nil
}

