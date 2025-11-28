package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uiansol/zentube/internal/entities"
)

type SQLiteRepository struct {
	db          *sql.DB
	saveStmt    *sql.Stmt
	getLastStmt *sql.Stmt
}

// NewSQLiteRepository creates a new SQLite repository with optimized settings
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Open with SQLite-specific optimizations
	// WAL mode enables concurrent reads and better performance
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_timeout=5000&_fk=true")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for optimal performance
	// SQLite benefits from a single writer, but multiple readers
	db.SetMaxOpenConns(25)           // Limit concurrent connections
	db.SetMaxIdleConns(5)            // Keep some connections ready
	db.SetConnMaxLifetime(time.Hour) // Recycle connections periodically

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &SQLiteRepository{db: db}

	// Initialize schema
	if err := repo.initSchema(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Prepare frequently used statements for better performance
	if err := repo.prepareStatements(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) initSchema(ctx context.Context) error {
	// Apply performance pragmas
	pragmas := `
	PRAGMA journal_mode=WAL;
	PRAGMA synchronous=NORMAL;
	PRAGMA cache_size=-64000;
	PRAGMA temp_store=MEMORY;
	PRAGMA busy_timeout=5000;
	`

	if _, err := r.db.ExecContext(ctx, pragmas); err != nil {
		return fmt.Errorf("failed to set pragmas: %w", err)
	}

	// Create schema with proper constraints
	schema := `
	CREATE TABLE IF NOT EXISTS search_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		query TEXT NOT NULL CHECK(length(query) > 0),
		results INTEGER NOT NULL CHECK(results >= 0),
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_created_at ON search_history(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_query ON search_history(query);
	`

	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) prepareStatements() error {
	var err error

	// Prepare INSERT statement
	r.saveStmt, err = r.db.Prepare(
		`INSERT INTO search_history (query, results, created_at) VALUES (?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare save statement: %w", err)
	}

	// Prepare SELECT statement
	r.getLastStmt, err = r.db.Prepare(
		`SELECT id, query, results, created_at FROM search_history ORDER BY created_at DESC LIMIT ?`,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare getLastStmt: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Save(ctx context.Context, history *entities.SearchHistory) error {
	result, err := r.saveStmt.ExecContext(ctx, history.Query, history.Results, history.CreatedAt)
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

func (r *SQLiteRepository) GetLast(ctx context.Context, limit int) ([]entities.SearchHistory, error) {
	rows, err := r.getLastStmt.QueryContext(ctx, limit)
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

// Close gracefully closes all prepared statements and the database connection
func (r *SQLiteRepository) Close() error {
	var errs []error

	// Close prepared statements first
	if r.saveStmt != nil {
		if err := r.saveStmt.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close save statement: %w", err))
		}
	}

	if r.getLastStmt != nil {
		if err := r.getLastStmt.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close getLastStmt: %w", err))
		}
	}

	// Close database connection
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// DB returns the underlying database connection for health checks
func (r *SQLiteRepository) DB() *sql.DB {
	return r.db
}
