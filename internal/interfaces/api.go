package interfaces

import (
	"backend/internal/transport/api/dto"
	"context"
)

type SearchAPI interface {
	Search(ctx context.Context, query string) ([]dto.Track, error)
}
