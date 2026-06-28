-- +goose Up
-- Enable pg_trgm extension for trigram similarity matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add trigram indexes on title and author_names for prefix/partial matching
CREATE INDEX idx_search_documents_title_trgm
ON openlibrary.search_documents
USING GIN (title gin_trgm_ops);

CREATE INDEX idx_search_documents_authors_trgm
ON openlibrary.search_documents
USING GIN (author_names gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS openlibrary.idx_search_documents_title_trgm;
DROP INDEX IF EXISTS openlibrary.idx_search_documents_authors_trgm;
DROP EXTENSION IF EXISTS pg_trgm;