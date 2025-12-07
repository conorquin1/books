package api

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/books/books"
	"github.com/books/testdata"
	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Create(t *testing.T) {
	suite := testdata.NewSuite(t).
		WithDB().
		WithCache()
	defer suite.Close()

	Convey("POST /api/v1/books", t, func() {
		e, _, service := suite.SetupAPI()
		apiGroup := e.Group("/api/v1")
		controller := &BookController{service: service}
		controller.Routes(apiGroup)

		Convey("Return 400 when title is missing", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "POST",
				Path:   "/api/v1/books",
				Body: CreateBookRequest{
					Author:      "Test Author",
					ISBN:        "1234567890",
					Description: "Test Description",
					PublishedAt: "2024-01-01",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when author is missing", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "POST",
				Path:   "/api/v1/books",
				Body: CreateBookRequest{
					Title:       "Test Book",
					ISBN:        "1234567890",
					Description: "Test Description",
					PublishedAt: "2024-01-01",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when publishedAt is missing", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "POST",
				Path:   "/api/v1/books",
				Body: CreateBookRequest{
					Title:       "Test Book",
					Author:      "Test Author",
					ISBN:        "1234567890",
					Description: "Test Description",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when publishedAt is in wrong format", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "POST",
				Path:   "/api/v1/books",
				Body: CreateBookRequest{
					Title:       "Test Book",
					Author:      "Test Author",
					ISBN:        "1234567890",
					Description: "Test Description",
					PublishedAt: "01-01-2024", // Wrong format
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 201 when book is created successfully", func() {
			suite.ClearBooks()

			var book books.Book
			res := suite.Request(e, &testdata.Request{
				Method: "POST",
				Path:   "/api/v1/books",
				Body: CreateBookRequest{
					Title:       "Test Book",
					Author:      "Test Author",
					ISBN:        "1234567890",
					Description: "Test Description",
					PublishedAt: "2024-01-01",
				},
			}, &book)

			So(res.StatusCode, ShouldEqual, http.StatusCreated)
			So(book.Title, ShouldEqual, "Test Book")
			So(book.Author, ShouldEqual, "Test Author")
			So(book.ID, ShouldBeGreaterThan, 0)
			So(book.PublishedAt.Format("2006-01-02"), ShouldEqual, "2024-01-01")

			// Verify the book was actually saved to the database
			savedBook := suite.GetBook(book.ID)
			So(savedBook, ShouldNotBeNil)
			So(savedBook.Title, ShouldEqual, "Test Book")
			So(savedBook.Author, ShouldEqual, "Test Author")
		})
	})
}

func Test_GetByID(t *testing.T) {
	suite := testdata.NewSuite(t).
		WithDB().
		WithCache()
	defer suite.Close()

	Convey("GET /api/v1/books/:id", t, func() {
		e, _, service := suite.SetupAPI()
		apiGroup := e.Group("/api/v1")
		controller := &BookController{service: service}
		controller.Routes(apiGroup)

		Convey("Return 400 when ID is invalid", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books/invalid",
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 404 when book is not found", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books/999",
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Return 200 when book is found", func() {
			suite.ClearBooks()

			// Insert a test book
			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var book books.Book
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
			}, &book)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(book.ID, ShouldEqual, testBook.ID)
			So(book.Title, ShouldEqual, "Test Book")
			So(book.Author, ShouldEqual, "Test Author")
		})
	})
}

func Test_GetAll(t *testing.T) {
	suite := testdata.NewSuite(t).
		WithDB().
		WithCache()
	defer suite.Close()

	Convey("GET /api/v1/books", t, func() {
		e, _, service := suite.SetupAPI()
		apiGroup := e.Group("/api/v1")
		controller := &BookController{service: service}
		controller.Routes(apiGroup)

		Convey("Return 200 with empty list when no books exist", func() {
			suite.ClearBooks()

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 0)
		})

		Convey("Return 200 with all books when no author filter is provided", func() {
			suite.ClearBooks()

			// Insert multiple books
			suite.InsertBook(books.Book{
				Title:       "Book 1",
				Author:      "Author A",
				ISBN:        "1111111111",
				Description: "Description 1",
				PublishedAt: parseTime("2024-01-01"),
			})
			suite.InsertBook(books.Book{
				Title:       "Book 2",
				Author:      "Author B",
				ISBN:        "2222222222",
				Description: "Description 2",
				PublishedAt: parseTime("2024-02-01"),
			})

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 2)
		})

		Convey("Return 200 with filtered books when author filter is provided", func() {
			suite.ClearBooks()

			// Insert books with different authors
			suite.InsertBook(books.Book{
				Title:       "Book 1",
				Author:      "Author A",
				ISBN:        "1111111111",
				Description: "Description 1",
				PublishedAt: parseTime("2024-01-01"),
			})
			suite.InsertBook(books.Book{
				Title:       "Book 2",
				Author:      "Author B",
				ISBN:        "2222222222",
				Description: "Description 2",
				PublishedAt: parseTime("2024-02-01"),
			})
			suite.InsertBook(books.Book{
				Title:       "Book 3",
				Author:      "Author A",
				ISBN:        "3333333333",
				Description: "Description 3",
				PublishedAt: parseTime("2024-03-01"),
			})

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books?author=Author+A",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 2)
		})

		Convey("Return 200 with empty list when author filter matches no books", func() {
			suite.ClearBooks()

			suite.InsertBook(books.Book{
				Title:       "Book 1",
				Author:      "Author A",
				ISBN:        "1111111111",
				Description: "Description 1",
				PublishedAt: parseTime("2024-01-01"),
			})

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books?author=Nonexistent+Author",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 0)
		})

		Convey("Return 200 with limited books when limit parameter is provided", func() {
			suite.ClearBooks()

			// Insert 5 books
			for i := 1; i <= 5; i++ {
				suite.InsertBook(books.Book{
					Title:       fmt.Sprintf("Book %d", i),
					Author:      "Author A",
					ISBN:        fmt.Sprintf("111111111%d", i),
					Description: fmt.Sprintf("Description %d", i),
					PublishedAt: parseTime("2024-01-01"),
				})
			}

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books?limit=3",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 3)
			So(response.Limit, ShouldEqual, 3)
			So(response.Page, ShouldEqual, 0) // Page should not be set when only limit is provided
			So(response.Books[0].Title, ShouldEqual, "Book 1")
			So(response.Books[1].Title, ShouldEqual, "Book 2")
			So(response.Books[2].Title, ShouldEqual, "Book 3")
		})

		Convey("Return 200 with paginated books when page and limit are provided", func() {
			suite.ClearBooks()

			// Insert 10 books
			for i := 1; i <= 10; i++ {
				suite.InsertBook(books.Book{
					Title:       fmt.Sprintf("Book %d", i),
					Author:      "Author A",
					ISBN:        fmt.Sprintf("111111111%d", i),
					Description: fmt.Sprintf("Description %d", i),
					PublishedAt: parseTime("2024-01-01"),
				})
			}

			Convey("Return first page", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=1&limit=3",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 3)
				So(response.Page, ShouldEqual, 1)
				So(response.Limit, ShouldEqual, 3)
				So(response.Books[0].Title, ShouldEqual, "Book 1")
				So(response.Books[1].Title, ShouldEqual, "Book 2")
				So(response.Books[2].Title, ShouldEqual, "Book 3")
			})

			Convey("Return second page", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=2&limit=3",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 3)
				So(response.Page, ShouldEqual, 2)
				So(response.Limit, ShouldEqual, 3)
				So(response.Books[0].Title, ShouldEqual, "Book 4")
				So(response.Books[1].Title, ShouldEqual, "Book 5")
				So(response.Books[2].Title, ShouldEqual, "Book 6")
			})

			Convey("Return last page with partial results", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=4&limit=3",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 1) // Only 1 book left (10 total, 3 per page = 4 pages, last page has 1)
				So(response.Page, ShouldEqual, 4)
				So(response.Limit, ShouldEqual, 3)
				So(response.Books[0].Title, ShouldEqual, "Book 10")
			})

			Convey("Return empty list when page is beyond available data", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=10&limit=3",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 0)
				So(response.Page, ShouldEqual, 10)
				So(response.Limit, ShouldEqual, 3)
			})
		})

		Convey("Return 200 ignoring page when only page is provided without limit", func() {
			suite.ClearBooks()

			// Insert 3 books
			for i := 1; i <= 3; i++ {
				suite.InsertBook(books.Book{
					Title:       fmt.Sprintf("Book %d", i),
					Author:      "Author A",
					ISBN:        fmt.Sprintf("111111111%d", i),
					Description: fmt.Sprintf("Description %d", i),
					PublishedAt: parseTime("2024-01-01"),
				})
			}

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books?page=2",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 3) // All books returned (page ignored)
			So(response.Page, ShouldEqual, 0)        // Page not included in response
			So(response.Limit, ShouldEqual, 0)      // Limit not included in response
		})

		Convey("Return 200 with paginated and filtered books when author, page and limit are provided", func() {
			suite.ClearBooks()

			// Insert books with different authors
			for i := 1; i <= 5; i++ {
				suite.InsertBook(books.Book{
					Title:       fmt.Sprintf("Book A %d", i),
					Author:      "Author A",
					ISBN:        fmt.Sprintf("111111111%d", i),
					Description: fmt.Sprintf("Description %d", i),
					PublishedAt: parseTime("2024-01-01"),
				})
			}
			for i := 1; i <= 3; i++ {
				suite.InsertBook(books.Book{
					Title:       fmt.Sprintf("Book B %d", i),
					Author:      "Author B",
					ISBN:        fmt.Sprintf("222222222%d", i),
					Description: fmt.Sprintf("Description %d", i),
					PublishedAt: parseTime("2024-01-01"),
				})
			}

			var response GetAllBooksResponse
			res := suite.Request(e, &testdata.Request{
				Method: "GET",
				Path:   "/api/v1/books?author=Author+A&page=2&limit=2",
			}, &response)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(response.Books, ShouldNotBeNil)
			So(len(response.Books), ShouldEqual, 2)
			So(response.Page, ShouldEqual, 2)
			So(response.Limit, ShouldEqual, 2)
			So(response.Books[0].Title, ShouldEqual, "Book A 3")
			So(response.Books[1].Title, ShouldEqual, "Book A 4")
			So(response.Books[0].Author, ShouldEqual, "Author A")
			So(response.Books[1].Author, ShouldEqual, "Author A")
		})

		Convey("Return 200 with invalid pagination parameters ignored", func() {
			suite.ClearBooks()

			// Insert 2 books
			suite.InsertBook(books.Book{
				Title:       "Book 1",
				Author:      "Author A",
				ISBN:        "1111111111",
				Description: "Description 1",
				PublishedAt: parseTime("2024-01-01"),
			})
			suite.InsertBook(books.Book{
				Title:       "Book 2",
				Author:      "Author A",
				ISBN:        "2222222222",
				Description: "Description 2",
				PublishedAt: parseTime("2024-02-01"),
			})

			Convey("Invalid page (negative)", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=-1&limit=10",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 2) // All books returned (invalid page ignored)
			})

			Convey("Invalid limit", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=1&limit=-5",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 2) // All books returned (invalid limit ignored)
			})

			Convey("Invalid page (non-numeric)", func() {
				var response GetAllBooksResponse
				res := suite.Request(e, &testdata.Request{
					Method: "GET",
					Path:   "/api/v1/books?page=abc&limit=10",
				}, &response)

				So(res.StatusCode, ShouldEqual, http.StatusOK)
				So(response.Books, ShouldNotBeNil)
				So(len(response.Books), ShouldEqual, 2) // All books returned (invalid page ignored)
			})
		})
	})
}

func Test_Update(t *testing.T) {
	suite := testdata.NewSuite(t).
		WithDB().
		WithCache()
	defer suite.Close()

	Convey("PUT /api/v1/books/:id", t, func() {
		e, _, service := suite.SetupAPI()
		apiGroup := e.Group("/api/v1")
		controller := &BookController{service: service}
		controller.Routes(apiGroup)

		Convey("Return 400 when ID is invalid", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/invalid",
				Body: UpdateBookRequest{
					Title:  "Updated Book",
					Author: "Updated Author",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when title is missing", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
				Body: UpdateBookRequest{
					Author: "Updated Author",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when author is missing", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
				Body: UpdateBookRequest{
					Title: "Updated Book",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 400 when publishedAt is in wrong format", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
				Body: UpdateBookRequest{
					Title:       "Updated Book",
					Author:      "Updated Author",
					PublishedAt: "01-01-2024", // Wrong format
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 404 when book is not found", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/999",
				Body: UpdateBookRequest{
					Title:  "Updated Book",
					Author: "Updated Author",
				},
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Return 200 when book is updated successfully", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var book books.Book
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
				Body: UpdateBookRequest{
					Title:       "Updated Book",
					Author:      "Updated Author",
					ISBN:        "9876543210",
					Description: "Updated Description",
					PublishedAt: "2024-06-01",
				},
			}, &book)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(book.ID, ShouldEqual, testBook.ID)
			So(book.Title, ShouldEqual, "Updated Book")
			So(book.Author, ShouldEqual, "Updated Author")
			So(book.ISBN, ShouldEqual, "9876543210")
			So(book.Description, ShouldEqual, "Updated Description")
			So(book.PublishedAt.Format("2006-01-02"), ShouldEqual, "2024-06-01")

			// Verify the book was actually updated in the database
			savedBook := suite.GetBook(testBook.ID)
			So(savedBook, ShouldNotBeNil)
			So(savedBook.Title, ShouldEqual, "Updated Book")
			So(savedBook.Author, ShouldEqual, "Updated Author")
		})

		Convey("Return 200 when book is updated without publishedAt", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			var book books.Book
			res := suite.Request(e, &testdata.Request{
				Method: "PUT",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
				Body: UpdateBookRequest{
					Title:  "Updated Book",
					Author: "Updated Author",
				},
			}, &book)

			So(res.StatusCode, ShouldEqual, http.StatusOK)
			So(book.ID, ShouldEqual, testBook.ID)
			So(book.Title, ShouldEqual, "Updated Book")
			So(book.Author, ShouldEqual, "Updated Author")
		})
	})
}

func Test_Delete(t *testing.T) {
	suite := testdata.NewSuite(t).
		WithDB().
		WithCache()
	defer suite.Close()

	Convey("DELETE /api/v1/books/:id", t, func() {
		e, _, service := suite.SetupAPI()
		apiGroup := e.Group("/api/v1")
		controller := &BookController{service: service}
		controller.Routes(apiGroup)

		Convey("Return 400 when ID is invalid", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "DELETE",
				Path:   "/api/v1/books/invalid",
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Return 404 when book is not found", func() {
			suite.ClearBooks()

			var resp echo.HTTPError
			res := suite.Request(e, &testdata.Request{
				Method: "DELETE",
				Path:   "/api/v1/books/999",
			}, &resp)

			So(res.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Return 204 when book is deleted successfully", func() {
			suite.ClearBooks()

			testBook := suite.InsertBook(books.Book{
				Title:       "Test Book",
				Author:      "Test Author",
				ISBN:        "1234567890",
				Description: "Test Description",
				PublishedAt: parseTime("2024-01-01"),
			})

			res := suite.Request(e, &testdata.Request{
				Method: "DELETE",
				Path:   "/api/v1/books/" + int64ToString(testBook.ID),
			})

			So(res.StatusCode, ShouldEqual, http.StatusNoContent)
			So(res.BodyString, ShouldEqual, "")

			// Verify the book was actually deleted (soft delete)
			deletedBook := suite.GetBook(testBook.ID)
			So(deletedBook, ShouldBeNil)
		})
	})
}

// Helper functions
func parseTime(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
