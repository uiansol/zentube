package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uiansol/zentube/internal/entities"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &SQLiteRepository{db: db}

	// Initialize schema
	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS search_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		query TEXT NOT NULL,
		results INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_created_at ON search_history(created_at DESC);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *SQLiteRepository) Save(history *entities.SearchHistory) error {
	query := `INSERT INTO search_history (query, results, created_at) VALUES (?, ?, ?)`

	result, err := r.db.Exec(query, history.Query, history.Results, history.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save search history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	history.ID = id
	return nil
}

func (r *SQLiteRepository) GetLast(limit int) ([]entities.SearchHistory, error) {
	query := `SELECT id, query, results, created_at FROM search_history ORDER BY created_at DESC LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query search history: %w", err)
	}
	defer rows.Close()

	var histories []entities.SearchHistory
	for rows.Next() {
		var h entities.SearchHistory
		if err := rows.Scan(&h.ID, &h.Query, &h.Results, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return histories, nil
}

func (r *SQLiteRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
