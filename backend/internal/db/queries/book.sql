-- name: InsertBook :exec
INSERT INTO books(book,rating,start_date,finish_date,pages,thoughts)
VALUES($1,$2,$3,$4,$5,$6) 
RETURNING *;

-- name: GetBookByID :one
SELECT
    id,
    book,
    rating,
    start_date,
    finish_date,
    pages,
    thoughts,
    created_at,
    updated_at
FROM
    books
WHERE
    id = $1;

-- name: SearchBooksByTerm :many
SELECT id, book, rating, start_date, finish_date, pages, thoughts, created_at, updated_at
FROM books
WHERE
    book ILIKE '%' || @term || '%'
   OR thoughts ILIKE '%' || @term || '%'
ORDER BY created_at DESC;

-- name: GetAllBooks :many
SELECT id, book, rating, start_date, finish_date, pages, thoughts, created_at, updated_at
FROM books
ORDER BY created_at DESC;
