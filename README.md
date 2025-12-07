# Book Catalog API

A REST API for managing a catalog of books with full CRUD capabilities.

## Architecture

- **Controllers** (`api/`) - Handle HTTP requests and responses
- **Services** (`books/`) - Contain business logic
- **Repositories** (`books/mysql/`) - Handle database operations
- **Models** (`books/`) - Domain models and interfaces

## API Endpoints

All endpoints are under `/api/v1/books`:

- `GET /api/v1/books` - Get all books (supports filtering and pagination)
- `GET /api/v1/books/:id` - Get a book by ID
- `POST /api/v1/books` - Create a new book
- `PUT /api/v1/books/:id` - Update a book
- `DELETE /api/v1/books/:id` - Delete a book (soft delete)

### Query Parameters

**GET /api/v1/books** supports the following optional query parameters:

- `author` - Filter books by author name (e.g., `?author=Author+Name`)
- `page` - Page number for pagination (1-indexed, requires `limit` to be effective)
- `limit` - Number of items per page (can be used alone or with `page`)

**Examples:**

- `GET /api/v1/books` - Returns all books
- `GET /api/v1/books?author=John+Doe` - Returns all books by John Doe
- `GET /api/v1/books?limit=10` - Returns first 10 books
- `GET /api/v1/books?page=2&limit=10` - Returns page 2 with 10 items per page
- `GET /api/v1/books?author=John+Doe&page=1&limit=5` - Returns first page of John Doe's books (5 per page)

## Request/Response Examples

### Get All Books

```bash
GET /api/v1/books
```

**Response (without pagination):**

```json
{
  "books": [
    {
      "id": 1,
      "title": "The Great Gatsby",
      "author": "F. Scott Fitzgerald",
      "isbn": "978-0-7432-7356-5",
      "description": "A classic American novel",
      "publishedAt": "1925-04-10T00:00:00Z",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Response (with pagination):**

```json
{
  "books": [
    {
      "id": 1,
      "title": "The Great Gatsby",
      "author": "F. Scott Fitzgerald",
      "isbn": "978-0-7432-7356-5",
      "description": "A classic American novel",
      "publishedAt": "1925-04-10T00:00:00Z",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 1,
  "limit": 10
}
```

### Get Book by ID

```bash
GET /api/v1/books/1
```

**Response:**

```json
{
  "id": 1,
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "isbn": "978-0-7432-7356-5",
  "description": "A classic American novel",
  "publishedAt": "1925-04-10T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

### Create Book

```bash
POST /api/v1/books
Content-Type: application/json

{
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "isbn": "978-0-7432-7356-5",
  "description": "A classic American novel",
  "publishedAt": "1925-04-10"
}
```

**Response (201 Created):**

```json
{
  "id": 1,
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "isbn": "978-0-7432-7356-5",
  "description": "A classic American novel",
  "publishedAt": "1925-04-10T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

### Update Book

```bash
PUT /api/v1/books/1
Content-Type: application/json

{
  "title": "The Great Gatsby (Updated)",
  "author": "F. Scott Fitzgerald",
  "isbn": "978-0-7432-7356-5",
  "description": "A classic American novel - updated description",
  "publishedAt": "1925-04-10"
}
```

**Response (200 OK):**

```json
{
  "id": 1,
  "title": "The Great Gatsby (Updated)",
  "author": "F. Scott Fitzgerald",
  "isbn": "978-0-7432-7356-5",
  "description": "A classic American novel - updated description",
  "publishedAt": "1925-04-10T00:00:00Z",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-02T00:00:00Z"
}
```

### Delete Book

```bash
DELETE /api/v1/books/1
```

**Response (204 No Content)**

## Database Schema

The books table should have the following structure:

```
CREATE TABLE IF NOT EXISTS books (
  id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  author VARCHAR(255) NOT NULL,
  isbn VARCHAR(13),
  description TEXT,
  publishedAt DATETIME(6) NOT NULL,
  createdAt DATETIME(6) NOT NULL,
  updatedAt DATETIME(6) NOT NULL,
  deletedAt DATETIME(6) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `UQ_books_title_author_isbn` (`title`, `author`, `isbn`),
  INDEX idx_deletedAt (deletedAt),
  INDEX idx_author (author),
  INDEX idx_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Configuration

This application uses [Viper](https://github.com/spf13/viper) for configuration management

### Environment Variables

All environment variables use the `BOOKS_` prefix. Dots in configuration keys are automatically converted to underscores.

**Database:**

- `BOOKS_DB_HOST` - Database host
- `BOOKS_DB_USER` - Database username
- `BOOKS_DB_PASSWORD` - Database password
- `BOOKS_DB_DATABASE` - Database name
- `BOOKS_DB_PORT` - Database port

**Redis:**

- `BOOKS_REDIS_DSN` - Redis connection string (format: `host:port` or `redis://host:port`)

**Server:**

- `BOOKS_SERVER_PORT` - Server port

**Test Database (for tests):**

- `BOOKS_TEST_DB_HOST` - Test database host
- `BOOKS_TEST_DB_USER` - Test database user
- `BOOKS_TEST_DB_PASSWORD` - Test database password
- `BOOKS_TEST_DB_DATABASE` - Test database name
- `BOOKS_TEST_DB_PORT` - Test database port

### Config File (Optional)

You can also create a `config.yaml` file in the project root or `/opt/books/conf/`:

```yaml
db:
  host: "localhost"
  user: "root"
  password: "mysecretpassword"
  database: "books"
  port: 3306

redis:
  dsn: "127.0.0.1:6379"

server:
  port: "8080"

test:
  db:
    host: "127.0.0.1"
    user: "root"
    password: "mysecretpassword"
    database: "mysql"
    port: 3306
```

**Priority:** Environment variables override config file values.

## Running the Application

### Local Development

```bash
# Set environment variables
export BOOKS_DB_PASSWORD=mysecretpassword
export BOOKS_DB_DATABASE=books

# Run the application
go run main.go
```

The server will start on the port specified by `BOOKS_SERVER_PORT` (or default from config file).

### Docker

#### Building the Docker Image

```bash
# Build the Docker image
docker-compose build books

# Or build without cache
docker-compose build --no-cache books
```

#### Running with Docker Compose

The easiest way to run the application is using Docker Compose, which will start all required services (MySQL, Redis, and the books API):

```bash
# Start all services in the background
docker-compose up -d

# View logs
docker-compose logs -f books

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

The API will be available at `http://localhost:8080` once the containers are running.

#### Running the Docker Image Directly

If you want to run just the Docker image without docker-compose:

```bash
# Build the image
docker build -t books:latest .

# Run the container (requires MySQL and Redis to be running separately)
docker run -p 8080:8080 \
  -e BOOKS_DB_HOST=host.docker.internal \
  -e BOOKS_DB_USER=root \
  -e BOOKS_DB_PASSWORD=mysecretpassword \
  -e BOOKS_DB_DATABASE=books \
  -e BOOKS_DB_PORT=3306 \
  -e BOOKS_REDIS_DSN=host.docker.internal:6379 \
  -e BOOKS_SERVER_PORT=8080 \
  books:latest
```

TEST
