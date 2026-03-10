package noop

import (
	"fmt"
	"io"

	"github.com/w-h-a/bees/internal/client/exporter"
	"github.com/w-h-a/bees/internal/domain"
)

type noopExporter struct {
	options exporter.Options
}

func (e *noopExporter) Write(w io.Writer, issues []domain.Issue) error {
	return fmt.Errorf("no exporter configured")
}

func NewExporter(opts ...exporter.Option) (exporter.Exporter, error) {
	options := exporter.NewOptions(opts...)
	return &noopExporter{
		options: options,
	}, nil
}
