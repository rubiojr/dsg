package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var defaultDataDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "dsg")

// Response represents a stored OpenAI response
type Response struct {
	ID          int64
	Prompt      string
	Response    string
	SchemaName  string
	SchemaURN   string
	CreatedAt   time.Time
	DatasetName string
}

// SQLiteStorage handles storing responses in SQLite
type SQLiteStorage struct {
	db      *sql.DB
	dataDir string
	dbPath  string
}

// Option defines a functional option for configuring SQLiteStorage
type Option func(*SQLiteStorage)

// WithDataDir sets a custom data directory for the SQLite database
func WithDataDir(path string) Option {
	return func(s *SQLiteStorage) {
		s.dataDir = path
		s.dbPath = filepath.Join(path, "history.db")
	}
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(opts ...Option) (*SQLiteStorage, error) {
	s := &SQLiteStorage{
		dataDir: defaultDataDir,
	}
	s.dbPath = filepath.Join(s.dataDir, "history.db")

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s.db = db

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prompt TEXT NOT NULL,
			response TEXT NOT NULL,
			schema_name TEXT,
			schema_urn TEXT,
			dataset_name TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// SaveResponse stores a response in the database
func (s *SQLiteStorage) SaveResponse(prompt, response, schemaName, schemaURN, datasetName string) (int64, error) {
	stmt, err := s.db.Prepare(`
		INSERT INTO responses (prompt, response, schema_name, schema_urn, dataset_name)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(prompt, response, schemaName, schemaURN, datasetName)
	if err != nil {
		return 0, fmt.Errorf("failed to insert response: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetResponse retrieves a response by ID
func (s *SQLiteStorage) GetResponse(id int64) (*Response, error) {
	row := s.db.QueryRow(`
		SELECT id, prompt, response, schema_name, schema_urn, dataset_name, created_at
		FROM responses WHERE id = ?
	`, id)

	var resp Response
	var createdAt time.Time
	err := row.Scan(&resp.ID, &resp.Prompt, &resp.Response, &resp.SchemaName, &resp.SchemaURN, &resp.DatasetName, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no response found with ID %d", id)
		}
		return nil, fmt.Errorf("failed to scan response: %w", err)
	}

	return &resp, nil
}

// ListResponses retrieves all responses, with optional limit and offset
func (s *SQLiteStorage) ListResponses(limit, offset int) ([]*Response, error) {
	rows, err := s.db.Query(`
		SELECT id, prompt, response, schema_name, schema_urn, dataset_name, created_at
		FROM responses ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query responses: %w", err)
	}
	defer rows.Close()

	var responses []*Response
	for rows.Next() {
		var resp Response
		var createdAt time.Time
		err := rows.Scan(&resp.ID, &resp.Prompt, &resp.Response, &resp.SchemaName, &resp.SchemaURN, &resp.DatasetName, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}

		responses = append(responses, &resp)
	}

	return responses, nil
}

// DeleteResponse deletes a response by ID
func (s *SQLiteStorage) DeleteResponse(id int64) error {
	_, err := s.db.Exec("DELETE FROM responses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete response: %w", err)
	}
	return nil
}

// ClearHistory deletes all response history
func (s *SQLiteStorage) ClearHistory() error {
	_, err := s.db.Exec("DELETE FROM responses")
	if err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}
	return nil
}
