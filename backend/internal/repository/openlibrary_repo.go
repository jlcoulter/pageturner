package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jlcoulter/pageturner/internal/db/generated"
)

type OpenLibraryRepo struct {
	q *generated.Queries
}

func NewOpenLibraryRepo(q *generated.Queries) *OpenLibraryRepo {
	return &OpenLibraryRepo{q: q}
}

// buildTsquery takes a user search term and builds a tsquery string
// that combines exact word matching with prefix matching on the last word.
// "The hobbit tol" → "hobbit & tol:*" so "tol" matches "tolkien".
// Single word: "hobbit" → "hobbit:*"
func buildTsquery(term string) string {
	words := strings.Fields(term)
	if len(words) == 0 {
		return ""
	}

	// Build prefix-match tokens: last word gets :* for prefix matching,
	// all other words match exactly.
	tokens := make([]string, len(words))
	for i, w := range words {
		if i == len(words)-1 {
			// Last word: exact OR prefix match
			tokens[i] = fmt.Sprintf("(%s | %s:*)", w, w)
		} else {
			tokens[i] = w
		}
	}

	return strings.Join(tokens, " & ")
}

// Search searches both title and author using full-text search with prefix matching
func (r *OpenLibraryRepo) Search(
	ctx context.Context,
	term string,
) ([]generated.SearchOpenLibraryRow, error) {
	if term == "" {
		return nil, nil
	}

	tsquery := buildTsquery(term)
	if tsquery == "" {
		return nil, nil
	}

	return r.q.SearchOpenLibrary(ctx, generated.SearchOpenLibraryParams{
		ToTsquery: tsquery,
		Column2:   term,
	})
}

// SearchPrefix searches by trigram similarity for short queries where FTS returns nothing
func (r *OpenLibraryRepo) SearchPrefix(
	ctx context.Context,
	term string,
) ([]generated.SearchOpenLibraryPrefixRow, error) {
	if term == "" {
		return nil, nil
	}
	return r.q.SearchOpenLibraryPrefix(ctx, term)
}