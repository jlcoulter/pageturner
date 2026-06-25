-- name: SearchOpenLibrary :many
-- Search for works by title or author name
SELECT
    work_id,
    title,
    author_names,
    ts_rank(
        search_vector,
        plainto_tsquery($1)
    ) AS RANK
FROM
    openlibrary.search_documents
WHERE
    search_vector @@plainto_tsquery($1)
ORDER BY
    RANK DESC
LIMIT 20;
