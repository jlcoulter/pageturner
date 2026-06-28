-- name: SearchOpenLibrary :many
-- Full-text search using websearch_to_tsquery for smart query parsing
-- (handles quotes, OR, negation). Fast GIN index scan.
SELECT
    work_id,
    title,
    author_names,
    ts_rank(search_vector, websearch_to_tsquery($1)) AS rank
FROM
    openlibrary.search_documents
WHERE
    search_vector @@ websearch_to_tsquery($1)
ORDER BY
    rank DESC
LIMIT 20;

-- name: SearchOpenLibraryPrefix :many
-- Trigram similarity search for short queries / prefix matching
-- where full-text search can't help (e.g. "Pott" matching "Potter").
-- Only used as a fallback for queries <= 3 characters.
SELECT
    work_id,
    title,
    author_names,
    similarity(title, $1) * 10 AS rank
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