package validation

import (
	"strings"
	"unicode"

	appErrors "github.com/uiansol/zentube/internal/errors"
)

// SearchInput represents validated search input
type SearchInput struct {
	Query      string
	MaxResults int64
}

// ValidateSearchQuery validates and sanitizes search query input
// Performs multiple checks:
// 1. Query length validation (prevent abuse)
// 2. Character sanitization (remove control characters)
// 3. Whitespace normalization
// 4. Empty query detection
func ValidateSearchQuery(query string, maxResults int64) (*SearchInput, error) {
	// Trim leading/trailing whitespace
	query = strings.TrimSpace(query)

	// Check if query is empty after trimming
	if query == "" {
		return nil, appErrors.NewValidationError("search query cannot be empty", nil)
	}

	// Validate query length
	// Min: 1 character (already checked above)
	// Max: 200 characters (prevent abuse, YouTube API limit is ~500)
	const maxQueryLength = 200
	if len(query) > maxQueryLength {
		return nil, appErrors.NewValidationError(
			"search query too long (maximum 200 characters)",
			nil,
		)
	}

	// Sanitize query - remove control characters
	// Control characters can cause issues with:
	// - Logging systems
	// - Database storage
	// - Terminal output
	// - API requests
	query = sanitizeString(query)

	// Normalize whitespace (replace multiple spaces with single space)
	query = normalizeWhitespace(query)

	// Validate maxResults
	if maxResults < 1 {
		return nil, appErrors.NewValidationError(
			"max_results must be at least 1",
			nil,
		)
	}

	// YouTube API allows max 50 results per request
	const maxAllowedResults = 50
	if maxResults > maxAllowedResults {
		return nil, appErrors.NewValidationError(
			"max_results cannot exceed 50",
			nil,
		)
	}

	return &SearchInput{
		Query:      query,
		MaxResults: maxResults,
	}, nil
}

// sanitizeString removes control characters and other potentially harmful characters
// Keeps: letters, numbers, spaces, and common punctuation
// Removes: control characters (0x00-0x1F), null bytes, etc.
func sanitizeString(s string) string {
	return strings.Map(func(r rune) rune {
		// Allow printable characters and common whitespace
		if unicode.IsPrint(r) || r == ' ' || r == '\t' || r == '\n' {
			return r
		}
		// Remove control characters
		return -1
	}, s)
}

// normalizeWhitespace replaces multiple consecutive spaces with a single space
// Example: "hello    world" -> "hello world"
func normalizeWhitespace(s string) string {
	// Split by whitespace
	fields := strings.Fields(s)
	// Join with single space
	return strings.Join(fields, " ")
}

// ValidatePageSize validates pagination page size
// This is a common pattern for paginated APIs
func ValidatePageSize(pageSize int, defaultSize, maxSize int) (int, error) {
	// Use default if not specified
	if pageSize == 0 {
		return defaultSize, nil
	}

	// Check minimum
	if pageSize < 1 {
		return 0, appErrors.NewValidationError(
			"page_size must be at least 1",
			nil,
		)
	}

	// Check maximum
	if pageSize > maxSize {
		return 0, appErrors.NewValidationError(
			"page_size too large",
			nil,
		)
	}

	return pageSize, nil
}

// SanitizeFilename removes dangerous characters from filenames
// Prevents directory traversal attacks (../, ..\, etc.)
func SanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Trim whitespace
	filename = strings.TrimSpace(filename)

	return filename
}

// ValidateEmail validates email format (basic validation)
// For production, consider using a proper email validation library
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return appErrors.NewValidationError("email cannot be empty", nil)
	}

	// Basic email format check
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return appErrors.NewValidationError("invalid email format", nil)
	}

	// Check length
	const maxEmailLength = 254 // RFC 5321
	if len(email) > maxEmailLength {
		return appErrors.NewValidationError("email too long", nil)
	}

	return nil
}
