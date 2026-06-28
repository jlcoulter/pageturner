package scripts

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorRow struct {
	ID   string
	Name string
}

type WorkRow struct {
	ID    string
	Title string
}

type WorkAuthorRow struct {
	WorkID   string
	AuthorID string
}

type Author struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Work struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Authors []struct {
		Author struct {
			Key string `json:"key"`
		} `json:"author"`
	} `json:"authors"`
}

func ImportOpenLibraryData(ctx context.Context) {

	connStr := "postgres://username:password@10.1.1.50:5432/pageturner?sslmode=disable"

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("pool config error: %v", err)
	}

	cfg.MaxConns = 4
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("pool creation error: %v", err)
	}
	defer pool.Close()

	log.Println("starting import")

	log.Println("importing authors")
	if err := ImportFile(ctx, pool, "../../openlibrarydata/authors.txt", "authors"); err != nil {
		log.Fatal(err)
	}

	log.Println("importing works")
	if err := ImportFile(ctx, pool, "../../openlibrarydata/works.txt", "works"); err != nil {
		log.Fatal(err)
	}

	log.Println("import complete")
}

func ImportFile(
	ctx context.Context,
	pool *pgxpool.Pool,
	filePath string,
	mode string,
) error {

	log.Printf("starting import mode=%s file=%s", mode, filePath)

	lines := make(chan string, 50000)
	authorCh := make(chan AuthorRow, 50000)
	workCh := make(chan WorkRow, 50000)
	workAuthorCh := make(chan WorkAuthorRow, 50000)

	parseWorkers := runtime.NumCPU() * 2

	log.Printf("starting %d parse workers", parseWorkers)

	var parsedCount atomic.Int64
	var parseWG sync.WaitGroup

	for i := 0; i < parseWorkers; i++ {

		parseWG.Add(1)

		go func(worker int) {

			defer parseWG.Done()

			for line := range lines {

				switch mode {

				case "authors":
					if parseAuthor(line, authorCh) {
						parsedCount.Add(1)
					}

				case "works":
					if parseWork(line, workCh, workAuthorCh) {
						parsedCount.Add(1)
					}
				}

			}

			log.Printf("parse worker %d finished", worker)

		}(i)
	}

	var insertWG sync.WaitGroup

	insertWG.Add(1)
	go func() {
		defer insertWG.Done()
		authorInsertWorker(ctx, pool, authorCh)
	}()

	insertWG.Add(1)
	go func() {
		defer insertWG.Done()
		workInsertWorker(ctx, pool, workCh)
	}()

	insertWG.Add(1)
	go func() {
		defer insertWG.Done()
		workAuthorInsertWorker(ctx, pool, workAuthorCh)
	}()

	log.Println("starting file scan")

	err := scanFile(filePath, lines)
	if err != nil {
		return err
	}

	close(lines)

	log.Println("waiting for parse workers")

	parseWG.Wait()

	log.Printf("total parsed rows: %d", parsedCount.Load())

	close(authorCh)
	close(workCh)
	close(workAuthorCh)

	insertWG.Wait()

	log.Println("insert workers finished")

	return nil
}

func scanFile(filePath string, lines chan<- string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 4*1024*1024)

	var count int64

	for {

		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		lines <- string(line)

		count++

		if count%1_000_000 == 0 {
			log.Printf("scanned %d lines", count)
		}
	}

	log.Printf("file scan complete, lines=%d", count)

	return nil
}

func parseAuthor(line string, authorCh chan<- AuthorRow) bool {

	i := strings.LastIndexByte(line, '\t')
	if i < 0 {
		log.Printf("author line missing tab")
		return false
	}

	jsonPart := strings.TrimSpace(line[i+1:])

	var a Author

	if err := json.Unmarshal([]byte(jsonPart), &a); err != nil {
		log.Printf("author json parse error: %v", err)
		return false
	}

	authorID := strings.TrimPrefix(a.Key, "/authors/")

	authorCh <- AuthorRow{
		ID:   authorID,
		Name: a.Name,
	}

	return true
}

func parseWork(
	line string,
	workCh chan<- WorkRow,
	workAuthorCh chan<- WorkAuthorRow,
) bool {

	i := strings.LastIndexByte(line, '\t')
	if i < 0 {
		log.Printf("work line missing tab")
		return false
	}

	jsonPart := strings.TrimSpace(line[i+1:])

	var w Work

	if err := json.Unmarshal([]byte(jsonPart), &w); err != nil {
		log.Printf("work json parse error: %v", err)
		return false
	}

	workID := strings.TrimPrefix(w.Key, "/works/")

	workCh <- WorkRow{
		ID:    workID,
		Title: w.Title,
	}

	for _, a := range w.Authors {

		authorID := strings.TrimPrefix(a.Author.Key, "/authors/")

		workAuthorCh <- WorkAuthorRow{
			WorkID:   workID,
			AuthorID: authorID,
		}
	}

	return true
}

func authorInsertWorker(
	ctx context.Context,
	pool *pgxpool.Pool,
	authorCh <-chan AuthorRow,
) {

	log.Println("author insert worker started")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("author worker acquire error: %v", err)
		return
	}
	defer conn.Release()

	batchSize := 10000

	rows := make([][]any, 0, batchSize)

	var inserted int64

	for a := range authorCh {

		rows = append(rows, []any{a.ID, a.Name})

		if len(rows) >= batchSize {

			n, err := conn.Conn().CopyFrom(
				ctx,
				pgx.Identifier{"openlibrary", "authors_stage"},
				[]string{"id", "author_name"},
				pgx.CopyFromRows(rows),
			)

			if err != nil {
				log.Printf("author copy error: %v", err)
			}

			inserted += n

			log.Printf("author batch inserted rows=%d total=%d", n, inserted)

			rows = rows[:0]
		}
	}

	if len(rows) > 0 {

		n, err := conn.Conn().CopyFrom(
			ctx,
			pgx.Identifier{"openlibrary", "authors_stage"},
			[]string{"id", "author_name"},
			pgx.CopyFromRows(rows),
		)

		if err != nil {
			log.Printf("author final copy error: %v", err)
		}

		inserted += n
	}

	log.Printf("author worker complete total_inserted=%d", inserted)
}

func workInsertWorker(
	ctx context.Context,
	pool *pgxpool.Pool,
	workCh <-chan WorkRow,
) {

	log.Println("work insert worker started")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("work worker acquire error: %v", err)
		return
	}
	defer conn.Release()

	batchSize := 10000

	rows := make([][]any, 0, batchSize)

	var inserted int64

	for w := range workCh {

		rows = append(rows, []any{w.ID, w.Title})

		if len(rows) >= batchSize {

			n, err := conn.Conn().CopyFrom(
				ctx,
				pgx.Identifier{"openlibrary", "works_stage"},
				[]string{"id", "title"},
				pgx.CopyFromRows(rows),
			)

			if err != nil {
				log.Printf("work copy error: %v", err)
			}

			inserted += n

			log.Printf("work batch inserted rows=%d total=%d", n, inserted)

			rows = rows[:0]
		}
	}

	if len(rows) > 0 {

		n, err := conn.Conn().CopyFrom(
			ctx,
			pgx.Identifier{"openlibrary", "works_stage"},
			[]string{"id", "title"},
			pgx.CopyFromRows(rows),
		)

		if err != nil {
			log.Printf("work final copy error: %v", err)
		}

		inserted += n
	}

	log.Printf("work worker complete total_inserted=%d", inserted)
}

func workAuthorInsertWorker(
	ctx context.Context,
	pool *pgxpool.Pool,
	workAuthorCh <-chan WorkAuthorRow,
) {

	log.Println("work-author insert worker started")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("work-author worker acquire error: %v", err)
		return
	}
	defer conn.Release()

	batchSize := 10000

	rows := make([][]any, 0, batchSize)

	var inserted int64

	for wa := range workAuthorCh {

		rows = append(rows, []any{wa.WorkID, wa.AuthorID})

		if len(rows) >= batchSize {

			n, err := conn.Conn().CopyFrom(
				ctx,
				pgx.Identifier{"openlibrary", "work_authors_stage"},
				[]string{"work_id", "author_id"},
				pgx.CopyFromRows(rows),
			)

			if err != nil {
				log.Printf("work-author copy error: %v", err)
			}

			inserted += n

			log.Printf("work-author batch inserted rows=%d total=%d", n, inserted)

			rows = rows[:0]
		}
	}

	if len(rows) > 0 {

		n, err := conn.Conn().CopyFrom(
			ctx,
			pgx.Identifier{"openlibrary", "work_authors_stage"},
			[]string{"work_id", "author_id"},
			pgx.CopyFromRows(rows),
		)

		if err != nil {
			log.Printf("work-author final copy error: %v", err)
		}

		inserted += n
	}

	log.Printf("work-author worker complete total_inserted=%d", inserted)
}
