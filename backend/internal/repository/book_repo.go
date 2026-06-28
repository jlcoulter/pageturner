package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/jlcoulter/pageturner/internal/db/generated"
	"github.com/jlcoulter/pageturner/internal/types"
)

type BookRepo struct {
	q *generated.Queries
}

func NewBookRepo(q *generated.Queries) *BookRepo {
	return &BookRepo{q: q}
}

func (r *BookRepo) SearchBooks(ctx context.Context, term string) ([]generated.Book, error) {
	// If term provided, filter by it (implementation depends on your query)
	// For now, return all books - the filtering should be done at DB level
	return r.q.SearchBooksByTerm(ctx)
}

func (r *BookRepo) GetAllBooks(ctx context.Context) ([]generated.Book, error) {
	return r.q.GetAllBooks(ctx)
}

func (r *BookRepo) SaveBook(ctx context.Context, entry types.BookEntry) error {
	// Convert StartDate string to sql.NullTime
	var start sql.NullTime
	if entry.StartDate != "" {
		t, err := time.Parse("2006-01-02", entry.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
		start = sql.NullTime{Time: t, Valid: true}
	}

	// Convert FinishDate string to sql.NullTime
	var finish sql.NullTime
	if entry.FinishDate != "" {
		t, err := time.Parse("2006-01-02", entry.FinishDate)
		if err != nil {
			return fmt.Errorf("invalid finish date: %w", err)
		}
		finish = sql.NullTime{Time: t, Valid: true}
	}

	// Convert Pages int to sql.NullInt32
	var pages sql.NullInt32
	if entry.Pages != "" {
		p, err := strconv.Atoi(entry.Pages)
		if err != nil {
			return fmt.Errorf("invalid pages number: %w", err)
		}
		pages = sql.NullInt32{Int32: int32(p), Valid: true}
	}

	// Convert Thoughts string to sql.NullString
	var thoughts sql.NullString
	if entry.Thoughts != "" {
		thoughts = sql.NullString{String: entry.Thoughts, Valid: true}
	}

	params := generated.InsertBookParams{
		Book:       entry.Book,
		Rating:     int32(entry.Rating),
		StartDate:  start,
		FinishDate: finish,
		Pages:      pages,
		Thoughts:   thoughts,
	}

	return r.q.InsertBook(ctx, params)
}