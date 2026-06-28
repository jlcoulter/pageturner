-- name: SearchOpenLibrary :many
-- Full-text search with title similarity boost.
-- $1 = combined tsquery string (built in Go from websearch + prefix)
-- $2 = original search term for similarity ranking
SELECT
    work_id,
    title,
    author_names,
    (ts_rank(search_vector, to_tsquery('english', $1)) + similarity(title, $2::text) * 5)::float8 AS rank
FROM
    openlibrary.search_documents
WHERE
    search_vector @@ to_tsquery('english', $1)
ORDER BY
    rank DESC
LIMIT 20;

-- name: SearchOpenLibraryPrefix :many
-- Trigram similarity search for short queries where FTS returns nothing.
SELECT
    work_id,
    title,
    author_names,
    (similarity(title, $1) * 10)::float8 AS rank
FROM
    openlibrary.search_documents
WHERE
    title % $1
    AND NOT EXISTS (
        SELECT 1 FROM openlibrary.search_documents sd
        WHERE sd.search_vector @@ websearch_to_tsquery($1)
        LIMIT 1
    )
ORDER BY
    rank DESC
LIMIT 20;