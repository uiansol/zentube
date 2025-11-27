# State-of-the-Art Database Integration in Go

This document outlines the production-ready database patterns implemented in zentube, suitable for a reference article on modern Go application architecture.

## Architecture Overview

The implementation follows **Hexagonal Architecture** (Ports & Adapters pattern):

```
Domain Layer (Entities)
    ↓
Port Layer (Interfaces)
    ↓
Use Case Layer (Business Logic)
    ↓
Adapter Layer (Infrastructure)
```

### Benefits:
- **Testability**: Business logic is isolated and easily testable
- **Flexibility**: Database can be swapped without changing business logic
- **Maintainability**: Clear separation of concerns

## Key Implementation Patterns

### 1. **Context-Aware Operations**

All database operations accept `context.Context` for:
- **Cancellation**: Stop long-running queries when client disconnects
- **Timeouts**: Prevent resource exhaustion from stuck operations
- **Tracing**: Enable distributed tracing in production

```go
func (r *SQLiteRepository) Save(ctx context.Context, history *entities.SearchHistory) error {
    result, err := r.saveStmt.ExecContext(ctx, history.Query, history.Results, history.CreatedAt)
    // ...
}
```

### 2. **Connection Pool Configuration**

Optimized for SQLite's single-writer, multiple-reader model:

```go
db.SetMaxOpenConns(25)           // Limit concurrent connections
db.SetMaxIdleConns(5)            // Keep connections ready
db.SetConnMaxLifetime(time.Hour) // Recycle periodically
```

**Why these values?**
- SQLite handles ~25 concurrent readers efficiently
- 5 idle connections balance startup time vs memory
- 1-hour lifetime prevents connection staleness

### 3. **Prepared Statements**

Pre-compiled queries for frequently executed operations:

```go
type SQLiteRepository struct {
    db          *sql.DB
    saveStmt    *sql.Stmt  // Prepared INSERT
    getLastStmt *sql.Stmt  // Prepared SELECT
}
```

**Benefits:**
- **Performance**: Query is parsed once, executed many times
- **Security**: Better protection against SQL injection
- **Efficiency**: Reduced CPU and memory usage

### 4. **SQLite Optimization Pragmas**

Production-ready configuration:

```go
PRAGMA journal_mode=WAL;      // Write-Ahead Logging for concurrency
PRAGMA synchronous=NORMAL;    // Balance safety vs performance
PRAGMA cache_size=-64000;     // 64MB cache (in KB)
PRAGMA temp_store=MEMORY;     // Temporary tables in RAM
PRAGMA busy_timeout=5000;     // Wait up to 5s for locks
```

**WAL Mode** is critical:
- Enables concurrent reads during writes
- Better crash recovery
- Improved performance for write-heavy workloads

### 5. **Schema with Constraints**

Database-level validation:

```sql
CREATE TABLE IF NOT EXISTS search_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT NOT NULL CHECK(length(query) > 0),
    results INTEGER NOT NULL CHECK(results >= 0),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Benefits:**
- Data integrity enforced at database level
- Prevents invalid data even if application bugs exist
- Self-documenting schema

### 6. **Strategic Indexing**

Indexes for common query patterns:

```sql
CREATE INDEX IF NOT EXISTS idx_created_at ON search_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_query ON search_history(query);
```

**Guideline**: Index columns used in:
- WHERE clauses
- ORDER BY clauses
- JOIN conditions

### 7. **Graceful Shutdown**

Proper resource cleanup order:

```go
// 1. Stop accepting new requests
srv.Shutdown(ctx)

// 2. Close prepared statements
r.saveStmt.Close()
r.getLastStmt.Close()

// 3. Close database connection
r.db.Close()
```

**Critical**: Database must close AFTER all HTTP handlers complete.

### 8. **Asynchronous Non-Critical Operations**

History saving doesn't block the user response:

```go
go func() {
    saveCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    _ = s.historyRepo.Save(saveCtx, history)
}()
```

**Pattern**: Fire-and-forget for analytics/logging that shouldn't impact UX.

### 9. **Error Handling Best Practices**

```go
if err != nil {
    return fmt.Errorf("failed to save search history: %w", err)
}
```

**Using `%w`**:
- Preserves error chain for debugging
- Enables `errors.Is()` and `errors.As()` checks
- Better error context in logs

### 10. **Comprehensive Resource Cleanup**

Handle multiple potential failures during shutdown:

```go
func (r *SQLiteRepository) Close() error {
    var errs []error
    
    if r.saveStmt != nil {
        if err := r.saveStmt.Close(); err != nil {
            errs = append(errs, err)
        }
    }
    // ... close other resources
    
    if len(errs) > 0 {
        return fmt.Errorf("errors during close: %v", errs)
    }
    return nil
}
```

## Testing Strategy

### Mock-Based Unit Tests

```go
type MockSearchHistoryRepository struct {
    mock.Mock
}

func (m *MockSearchHistoryRepository) Save(ctx context.Context, h *entities.SearchHistory) error {
    args := m.Called(ctx, h)
    return args.Error(0)
}
```

**Benefits:**
- Fast execution (no real database)
- Predictable outcomes
- Easy to test error scenarios

### Testing Async Operations

```go
// Execute async operation
videos, err := uc.Execute(ctx, "golang", 10)

// Wait for goroutine to complete
time.Sleep(100 * time.Millisecond)

// Verify side effects
mockRepo.AssertExpectations(t)
```

## Configuration Management

**YAML + Environment Variables**:

```yaml
database:
  path: ./data/zentube.db
```

**Benefits:**
- Secrets via environment variables
- Defaults in version control
- Easy per-environment configuration

## Performance Considerations

### When to Use SQLite

**Good for:**
- ✅ Low to medium write volume (<100 writes/sec)
- ✅ High read volume
- ✅ Embedded applications
- ✅ Single-server deployments
- ✅ Development/testing

**Consider PostgreSQL/MySQL when:**
- ❌ High concurrent writes
- ❌ Multi-server deployments
- ❌ Need advanced features (full-text search, JSON queries)
- ❌ Very large datasets (>100GB)

## Security Best Practices

1. **Parameterized Queries**: Always use `?` placeholders
2. **Input Validation**: CHECK constraints in schema
3. **Least Privilege**: Application uses dedicated user (for multi-user DBs)
4. **Regular Backups**: WAL mode makes hot backups easier

## Monitoring & Observability

**Metrics to track:**
- Query duration (via context deadlines)
- Connection pool utilization
- Failed query rate
- Database file size

**Implementation tip:**
```go
start := time.Now()
err := r.saveStmt.ExecContext(ctx, ...)
metrics.RecordDuration("db.save", time.Since(start))
```

## Migration Strategy

For production, consider:
- **golang-migrate/migrate**: Version-controlled schema changes
- **goose**: Alternative with Go migrations
- **sql-migrate**: Embedded migrations

Example structure:
```
migrations/
  001_initial_schema.up.sql
  001_initial_schema.down.sql
  002_add_user_id.up.sql
  002_add_user_id.down.sql
```

## Common Pitfalls to Avoid

1. ❌ **Not closing rows**: Always `defer rows.Close()`
2. ❌ **Ignoring `rows.Err()`**: Check after iteration
3. ❌ **Missing context**: Use `*Context` methods
4. ❌ **Hardcoded queries**: Use prepared statements
5. ❌ **No connection limits**: Always set pool size
6. ❌ **Blocking on analytics**: Use goroutines for non-critical ops

## Conclusion

This implementation demonstrates:
- ✅ Clean architecture principles
- ✅ Production-ready database configuration
- ✅ Comprehensive error handling
- ✅ Graceful shutdown
- ✅ Performance optimizations
- ✅ Testable design

Perfect foundation for a reference article on modern Go database patterns.
