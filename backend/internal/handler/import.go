package handler

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImportHandler handles OpenLibrary data file uploads
type ImportHandler struct {
	pool *pgxpool.Pool
}

// NewImportHandler creates a new ImportHandler
func NewImportHandler(pool *pgxpool.Pool) *ImportHandler {
	return &ImportHandler{pool: pool}
}

type authorRow struct {
	ID   string
	Name string
}

type workRow struct {
	ID    string
	Title string
}

type workAuthorRow struct {
	WorkID   string
	AuthorID string
}

type olAuthor struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type olWork struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Authors []struct {
		Author struct {
			Key string `json:"key"`
		} `json:"author"`
	} `json:"authors"`
}

// ServeHTTP handles the import upload request
func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload to 5GB total, buffer up to 256MB in memory for multipart parsing
	r.Body = http.MaxBytesReader(w, r.Body, 5<<30)
	if err := r.ParseMultipartForm(256 << 20); err != nil {
		slog.Error("failed to parse multipart form", "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	fileType := r.FormValue("type")
	if fileType != "authors" && fileType != "works" {
		http.Error(w, "type must be 'authors' or 'works'", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("failed to get file from form", "error", err)
		http.Error(w, "failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decompress gzip files on the fly
	var reader io.Reader = file
	if strings.HasSuffix(strings.ToLower(header.Filename), ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			slog.Error("failed to open gzip reader", "error", err)
			http.Error(w, "failed to decompress gzip file", http.StatusBadRequest)
			return
		}
		defer gzReader.Close()
		reader = gzReader
	}

	slog.Info("starting import", "type", fileType, "filename", header.Filename)

	if err := h.importStream(r, reader, fileType); err != nil {
		slog.Error("import failed", "type", fileType, "error", err)
		http.Error(w, "import failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("import complete", "type", fileType)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ImportHandler) importStream(r *http.Request, reader io.Reader, mode string) error {
	ctx := r.Context()

	lines := make(chan string, 50000)
	authorCh := make(chan authorRow, 50000)
	workCh := make(chan workRow, 50000)
	workAuthorCh := make(chan workAuthorRow, 50000)

	var parsedCount atomic.Int64
	var parseWG sync.WaitGroup

	for i := 0; i < 4; i++ {
		parseWG.Add(1)
		go func() {
			defer parseWG.Done()
			for line := range lines {
				switch mode {
				case "authors":
					if a, ok := parseAuthorLine(line); ok {
						authorCh <- a
						parsedCount.Add(1)
					}
				case "works":
					if w, was, ok := parseWorkLine(line); ok {
						workCh <- w
						for _, wa := range was {
							workAuthorCh <- wa
						}
						parsedCount.Add(1)
					}
				}
			}
		}()
	}

	var insertWG sync.WaitGroup

	if mode == "authors" {
		insertWG.Add(1)
		go func() {
			defer insertWG.Done()
			conn, err := h.pool.Acquire(ctx)
			if err != nil {
				slog.Error("failed to acquire connection", "error", err)
				return
			}
			defer conn.Release()
			batchInsert(ctx, conn, "openlibrary", "authors_stage",
				[]string{"id", "author_name"}, authorCh, 10000)
		}()
	} else {
		insertWG.Add(1)
		go func() {
			defer insertWG.Done()
			conn, err := h.pool.Acquire(ctx)
			if err != nil {
				slog.Error("failed to acquire connection for works", "error", err)
				return
			}
			defer conn.Release()
			batchInsert(ctx, conn, "openlibrary", "works_stage",
				[]string{"id", "title"}, workCh, 10000)
		}()
		insertWG.Add(1)
		go func() {
			defer insertWG.Done()
			conn, err := h.pool.Acquire(ctx)
			if err != nil {
				slog.Error("failed to acquire connection for work_authors", "error", err)
				return
			}
			defer conn.Release()
			batchInsert(ctx, conn, "openlibrary", "work_authors_stage",
				[]string{"work_id", "author_id"}, workAuthorCh, 10000)
		}()
	}

	scanner := bufio.NewReaderSize(reader, 4*1024*1024)
	var lineCount int64
	for {
		line, err := scanner.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("error reading upload", "error", err)
			break
		}
		lines <- string(line)
		lineCount++
		if lineCount%100000 == 0 {
			slog.Info("upload progress", "lines", lineCount, "type", mode)
		}
	}

	close(lines)
	parseWG.Wait()

	slog.Info("parsing complete", "type", mode, "lines_read", lineCount, "parsed", parsedCount.Load())

	close(authorCh)
	close(workCh)
	close(workAuthorCh)
	insertWG.Wait()

	slog.Info("staging complete", "type", mode)
	return nil
}

func batchInsert(ctx context.Context, conn *pgxpool.Conn, schema, table string, columns []string, ch interface{}, batchSize int) {
	var rows [][]any

	switch c := ch.(type) {
	case chan authorRow:
		for a := range c {
			rows = append(rows, []any{a.ID, a.Name})
			if len(rows) >= batchSize {
				n, err := conn.Conn().CopyFrom(ctx,
					pgx.Identifier{schema, table},
					columns,
					pgx.CopyFromRows(rows),
				)
				if err != nil {
					slog.Error("copy error", "table", table, "error", err)
				} else {
					slog.Info("batch inserted", "table", table, "rows", n)
				}
				rows = rows[:0]
			}
		}
	case chan workRow:
		for w := range c {
			rows = append(rows, []any{w.ID, w.Title})
			if len(rows) >= batchSize {
				n, err := conn.Conn().CopyFrom(ctx,
					pgx.Identifier{schema, table},
					columns,
					pgx.CopyFromRows(rows),
				)
				if err != nil {
					slog.Error("copy error", "table", table, "error", err)
				} else {
					slog.Info("batch inserted", "table", table, "rows", n)
				}
				rows = rows[:0]
			}
		}
	case chan workAuthorRow:
		for wa := range c {
			rows = append(rows, []any{wa.WorkID, wa.AuthorID})
			if len(rows) >= batchSize {
				n, err := conn.Conn().CopyFrom(ctx,
					pgx.Identifier{schema, table},
					columns,
					pgx.CopyFromRows(rows),
				)
				if err != nil {
					slog.Error("copy error", "table", table, "error", err)
				} else {
					slog.Info("batch inserted", "table", table, "rows", n)
				}
				rows = rows[:0]
			}
		}
	}

	if len(rows) > 0 {
		n, err := conn.Conn().CopyFrom(ctx,
			pgx.Identifier{schema, table},
			columns,
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			slog.Error("final copy error", "table", table, "error", err)
		} else {
			slog.Info("final batch inserted", "table", table, "rows", n)
		}
	}
}

func parseAuthorLine(line string) (authorRow, bool) {
	i := strings.LastIndexByte(line, '\t')
	if i < 0 {
		return authorRow{}, false
	}

	jsonPart := strings.TrimSpace(line[i+1:])
	var a olAuthor
	if err := json.Unmarshal([]byte(jsonPart), &a); err != nil {
		return authorRow{}, false
	}

	authorID := strings.TrimPrefix(a.Key, "/authors/")
	return authorRow{ID: authorID, Name: a.Name}, true
}

func parseWorkLine(line string) (workRow, []workAuthorRow, bool) {
	i := strings.LastIndexByte(line, '\t')
	if i < 0 {
		return workRow{}, nil, false
	}

	jsonPart := strings.TrimSpace(line[i+1:])
	var w olWork
	if err := json.Unmarshal([]byte(jsonPart), &w); err != nil {
		return workRow{}, nil, false
	}

	workID := strings.TrimPrefix(w.Key, "/works/")
	wr := workRow{ID: workID, Title: w.Title}

	var was []workAuthorRow
	for _, a := range w.Authors {
		authorID := strings.TrimPrefix(a.Author.Key, "/authors/")
		was = append(was, workAuthorRow{WorkID: workID, AuthorID: authorID})
	}

	return wr, was, true
}