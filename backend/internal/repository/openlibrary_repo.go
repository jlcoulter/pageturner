package repository

import (
	"context"

	"github.com/jlcoulter/pageturner/internal/db/generated"
)

type OpenLibraryRepo struct {
	q *generated.Queries
}

func NewOpenLibraryRepo(q *generated.Queries) *OpenLibraryRepo {
	return &OpenLibraryRepo{q: q}
}

// Search searches both title and author using full-text search
func (r *OpenLibraryRepo) Search(
	ctx context.Context,
	term string,
) ([]generated.SearchOpenLibraryRow, error) {
	if term == "" {
		return nil, nil
	}
	return r.q.SearchOpenLibrary(ctx, term)
}

// SearchPrefix searches by trigram similarity for short queries
func (r *OpenLibraryRepo) SearchPrefix(
	ctx context.Context,
	term string,
) ([]generated.SearchOpenLibraryPrefixRow, error) {
	if term == "" {
		return nil, nil
	}
	return r.q.SearchOpenLibraryPrefix(ctx, term)
}

