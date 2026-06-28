-- name: SearchOpenLibrary :many
-- Full-text search with title similarity boost for relevance.
-- FTS finds candidates fast (GIN index), then we boost exact/prefix
-- title matches so "The Hobbit" ranks above "Hobbit Cookery".
SELECT
    work_id,
    title,
    author_names,
    (ts_rank(search_vector, websearch_to_tsquery($1))
     + similarity(title, $1) * 5) AS rank
FROM
    openlibrary.search_documents
WHERE
    search_vector @@ websearch_to_tsquery($1)
ORDER BY
    rank DESC
LIMIT 20;

-- name: SearchOpenLibraryPrefix :many
-- Trigram similarity search for short queries where FTS returns nothing.
-- Only used as fallback for queries <= 3 characters.
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