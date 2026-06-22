package generate

import (
	"context"
	"errors"

	"github.com/wahyusiddarta/oasgo/internal/openapi"
	"github.com/wahyusiddarta/oasgo/internal/scan"
)

// Config controls generator behavior.
type Config struct {
	Dir     string
	Title   string
	Version string
}

// Generate coordinates source scanning, operation parsing, schema generation,
// and OpenAPI document construction.
func Generate(ctx context.Context, cfg Config) (*openapi.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cfg.Dir == "" {
		return nil, errors.New("oasgo: Dir is required")
	}

	if _, err := scan.ScanDir(cfg.Dir); err != nil {
		return nil, err
	}

	doc := openapi.NewDocument(cfg.Title, cfg.Version)
	return doc, nil
}
