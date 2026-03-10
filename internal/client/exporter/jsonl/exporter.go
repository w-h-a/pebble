package jsonl

import (
	"encoding/json"
	"io"

	"github.com/w-h-a/bees/internal/client/exporter"
	"github.com/w-h-a/bees/internal/domain"
)

type jsonlExporter struct {
	options exporter.Options
}

func (e *jsonlExporter) Write(w io.Writer, issues []domain.Issue) error {
	enc := json.NewEncoder(w)

	for _, issue := range issues {
		if err := enc.Encode(mapFromIssue(issue)); err != nil {
			return err
		}
	}

	return nil
}

func NewExporter(opts ...exporter.Option) (exporter.Exporter, error) {
	options := exporter.NewOptions(opts...)

	e := &jsonlExporter{
		options: options,
	}

	return e, nil
}
