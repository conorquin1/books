-- Create books table
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

