-- +goose Up
-- Initial OpenLibrary schema including staging tables

CREATE SCHEMA IF NOT EXISTS openlibrary;

-- =====================================================
-- STAGING TABLES (raw ingestion layer)
-- =====================================================

CREATE TABLE openlibrary.authors_stage (
id TEXT,
author_name TEXT
);

CREATE TABLE openlibrary.works_stage (
id TEXT,
title TEXT
);

CREATE TABLE openlibrary.work_authors_stage (
work_id TEXT,
author_id TEXT
);

-- No indexes or constraints on staging tables.
-- They exist only for high-speed bulk ingestion.

-- =====================================================
-- PRODUCTION TABLES
-- =====================================================

CREATE TABLE openlibrary.authors (
id TEXT PRIMARY KEY,
author_name TEXT NOT NULL,
search_vector tsvector GENERATED ALWAYS AS (
to_tsvector('english', author_name)
) STORED
);

CREATE INDEX idx_authors_name
ON openlibrary.authors(author_name);

CREATE INDEX idx_authors_search_vector
ON openlibrary.authors
USING GIN(search_vector);

CREATE TABLE openlibrary.works (
id TEXT PRIMARY KEY,
title TEXT NOT NULL,
search_vector tsvector GENERATED ALWAYS AS (
to_tsvector('english', title)
) STORED
);

CREATE INDEX idx_works_title
ON openlibrary.works(title);

CREATE INDEX idx_works_search_vector
ON openlibrary.works
USING GIN(search_vector);

CREATE TABLE openlibrary.work_authors (
work_id TEXT NOT NULL,
author_id TEXT NOT NULL,
PRIMARY KEY (work_id, author_id),

CONSTRAINT fk_work
    FOREIGN KEY (work_id)
    REFERENCES openlibrary.works(id)
    ON DELETE CASCADE,

CONSTRAINT fk_author
    FOREIGN KEY (author_id)
    REFERENCES openlibrary.authors(id)
    ON DELETE CASCADE

);

CREATE INDEX idx_work_authors_work
ON openlibrary.work_authors(work_id);

CREATE INDEX idx_work_authors_author
ON openlibrary.work_authors(author_id);

-- =====================================================
-- SEARCH TABLE (denormalized index for fast queries)
-- =====================================================

CREATE TABLE openlibrary.search_documents (
work_id TEXT PRIMARY KEY,
title TEXT,
author_names TEXT,
search_vector tsvector
);

CREATE INDEX idx_search_documents_vector
ON openlibrary.search_documents
USING GIN(search_vector);

-- +goose Down

DROP TABLE IF EXISTS openlibrary.search_documents;
DROP TABLE IF EXISTS openlibrary.work_authors;
DROP TABLE IF EXISTS openlibrary.authors;
DROP TABLE IF EXISTS openlibrary.works;

DROP TABLE IF EXISTS openlibrary.work_authors_stage;
DROP TABLE IF EXISTS openlibrary.authors_stage;
DROP TABLE IF EXISTS openlibrary.works_stage;

DROP SCHEMA IF EXISTS openlibrary;

