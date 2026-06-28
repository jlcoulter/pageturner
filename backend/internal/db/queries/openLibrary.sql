-- name: SearchOpenLibrary :many
-- Combined full-text and trigram search with ranked results
(
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
    LIMIT 20
)
UNION ALL
(
    SELECT
        work_id,
        title,
        author_names,
        (similarity(title, $1) + similarity(author_names, $1)) * 10 AS rank
    FROM
        openlibrary.search_documents
    WHERE
        title % $1
        OR author_names % $1
        AND NOT (
            search_vector @@ websearch_to_tsquery($1)
        )
    ORDER BY
        rank DESC
    LIMIT 20
)
ORDER BY rank DESC
LIMIT 20;