package exporter

import (
	"io"

	"github.com/w-h-a/bees/internal/domain"
)

type Exporter interface {
	Write(w io.Writer, issues []domain.Issue) error
}
