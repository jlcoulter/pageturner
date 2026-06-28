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

// Search searches both title and author
func (r *OpenLibraryRepo) Search(
	ctx context.Context,
	term string,
) ([]generated.SearchOpenLibraryRow, error) {
	if term == "" {
		return nil, nil
	}
	var query string
	query = term
	return r.q.SearchOpenLibrary(ctx, query)
}

