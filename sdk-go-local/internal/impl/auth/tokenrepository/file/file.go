package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

// TokenRepository is a struct that implements the auth.TokenRepository interface
// using the local file system for token persistence.
type TokenRepository struct {
	baseDir string // The root directory where token files are stored.
	path    string // The full, absolute path to the specific token file.
}

var _ auth.TokenRepository = (*TokenRepository)(nil)

// NewFileTokenRepository is the constructor for TokenRepository.
// It constructs the full file path where the token will be stored.
func NewFileTokenRepository(clientID, baseDir string) *TokenRepository {
	// Construct the file name using the client ID.
	name := fmt.Sprintf("%s.token.json", clientID)
	return &TokenRepository{
		baseDir: baseDir,
		// Join the base directory and the file name using the system's path separator.
		path: filepath.Join(baseDir, name),
	}
}

// FetchToken retrieves the token from the file system.
// It reads the file, decodes the JSON content into an *auth.Token, and returns it.
func (tr *TokenRepository) FetchToken(ctx context.Context) (*auth.Token, error) {
	// 1. Read the entire file content into a byte slice.
	data, err := os.ReadFile(tr.path)
	if err != nil {
		// Check if the error is due to the file not existing.
		if errors.Is(err, os.ErrNotExist) {
			return nil, auth.ErrTokenNotFound
		}
		// Wrap other file access errors (permission denied, etc.)
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token auth.Token
	if err := json.Unmarshal(data, &token); err != nil {
		// Return an error if JSON unmarshalling fails (file corruption).
		return nil, err
	}

	return &token, nil
}

// SaveToken serializes and persists the given *auth.Token to the file system.
// It ensures the directory structure exists before writing the file.
func (tr *TokenRepository) SaveToken(ctx context.Context, token *auth.Token) error {
	// Basic validation: ensure the input token is not nil.
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}
	// Marshal the token struct into a JSON byte array.
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(tr.baseDir, 0o700); err != nil {
		return err
	}
	// Write the JSON data to the file path.
	return os.WriteFile(tr.path, tokenJSON, 0o600)
}
