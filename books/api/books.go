package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/books/books"
	"github.com/books/validate"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// BookController handles book API requests.
type BookController struct {
	service *books.BookService
}

// newBookController returns a new BookController.
func newBookController(service *books.BookService) *BookController {
	return &BookController{service: service}
}

// Routes sets up the routes for the book controller.
func (c *BookController) Routes(g *echo.Group) {
	api := g.Group("/books", ErrorHandler)

	api.GET("", c.GetAll)
	api.GET("/:id", c.GetByID)
	api.POST("", c.Create)
	api.PUT("/:id", c.Update)
	api.DELETE("/:id", c.Delete)
}

// CreateBookRequest represents the request body for creating a book.
type CreateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
	PublishedAt string `json:"publishedAt"` // Format: "2006-01-02"
}

// UpdateBookRequest represents the request body for updating a book.
type UpdateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
	PublishedAt string `json:"publishedAt"` // Format: "2006-01-02"
}

// GetAllBooksResponse represents the response body for getting all books.
type GetAllBooksResponse struct {
	Books []books.Book `json:"books"`
	Page  int          `json:"page,omitempty"`
	Limit int          `json:"limit,omitempty"`
}

// Create creates a new book.
func (c *BookController) Create(ctx echo.Context) error {
	var req CreateBookRequest
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	v := validate.New()
	v.Required("title", req.Title)
	v.Required("author", req.Author)
	v.Required("publishedAt", req.PublishedAt)
	if v.HasErrors() {
		return v
	}

	// Parse published date (required)
	publishedAt, err := time.Parse("2006-01-02", req.PublishedAt)
	if err != nil {
		return errors.Wrap(books.ErrInvalidBookData, "invalid publishedAt format, expected YYYY-MM-DD")
	}

	book := books.Book{
		Title:       req.Title,
		Author:      req.Author,
		ISBN:        req.ISBN,
		Description: req.Description,
		PublishedAt: publishedAt,
	}

	createdBook, err := c.service.Create(ctx.Request().Context(), book)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, createdBook)
}

// GetByID retrieves a book by ID.
func (c *BookController) GetByID(ctx echo.Context) error {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return errors.Wrap(books.ErrInvalidBookData, "invalid book ID")
	}

	book, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, book)
}

// GetAll retrieves all books.
// Query parameters:
//   - author: filter by author name (optional)
//   - page: page number (1-indexed, optional)
//   - limit: number of items per page (optional)
func (c *BookController) GetAll(ctx echo.Context) error {
	author := ctx.QueryParam("author")
	var authorPtr *string
	if author != "" {
		authorPtr = &author
	}

	// Parse pagination parameters (both are optional)
	var page, limit int
	hasPage := false
	hasLimit := false

	if pageStr := ctx.QueryParam("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
			hasPage = true
		}
	}

	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			hasLimit = true
		}
	}

	// Calculate offset
	offset := 0
	if hasLimit {
		if hasPage {
			// Both page and limit provided: calculate offset for the page
			offset = (page - 1) * limit
		}
		// If only limit is provided, offset stays 0 (return first N items)
	}

	bookList, err := c.service.GetAll(ctx.Request().Context(), authorPtr, limit, offset)
	if err != nil {
		return err
	}

	response := GetAllBooksResponse{
		Books: bookList,
	}
	
	// Include pagination metadata only if provided
	// Page only makes sense if limit is also provided
	if hasPage && hasLimit {
		response.Page = page
	}
	if hasLimit {
		response.Limit = limit
	}

	return ctx.JSON(http.StatusOK, response)
}

// Update updates an existing book.
func (c *BookController) Update(ctx echo.Context) error {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return errors.Wrap(books.ErrInvalidBookData, "invalid book ID")
	}

	var req UpdateBookRequest
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	v := validate.New()
	v.Required("title", req.Title)
	v.Required("author", req.Author)
	if v.HasErrors() {
		return v
	}

	// Parse published date if provided
	var publishedAt time.Time
	if req.PublishedAt != "" {
		parsed, err := time.Parse("2006-01-02", req.PublishedAt)
		if err != nil {
			return errors.Wrap(books.ErrInvalidBookData, "invalid publishedAt format, expected YYYY-MM-DD")
		}
		publishedAt = parsed
	}

	book := books.Book{
		Title:       req.Title,
		Author:      req.Author,
		ISBN:        req.ISBN,
		Description: req.Description,
		PublishedAt: publishedAt,
	}

	updatedBook, err := c.service.Update(ctx.Request().Context(), id, book)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, updatedBook)
}

// Delete deletes a book.
func (c *BookController) Delete(ctx echo.Context) error {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return errors.Wrap(books.ErrInvalidBookData, "invalid book ID")
	}

	err = c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}

