-- ---------------------------------------------------------
-- 1. Session tuning for bulk operations
-- ---------------------------------------------------------
SET maintenance_work_mem = '4GB';
SET work_mem = '512MB';
SET synchronous_commit = OFF;
SET temp_buffers = '512MB';

-- ---------------------------------------------------------
-- 2. Drop previous temp tables
-- ---------------------------------------------------------

DROP TABLE IF EXISTS openlibrary.search_documents_new;
DROP TABLE IF EXISTS openlibrary.work_authors_new;
DROP TABLE IF EXISTS openlibrary.authors_new;
DROP TABLE IF EXISTS openlibrary.works_new;

-- ---------------------------------------------------------
-- 3. Create UNLOGGED build tables
-- ---------------------------------------------------------

CREATE UNLOGGED TABLE openlibrary.authors_new (
    id TEXT PRIMARY KEY,
    author_name TEXT NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', author_name)
    ) STORED
);

CREATE UNLOGGED TABLE openlibrary.works_new (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title)
    ) STORED
);

CREATE UNLOGGED TABLE openlibrary.work_authors_new (
    work_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    PRIMARY KEY (work_id, author_id)
);

-- ---------------------------------------------------------
-- 4. Bulk load authors
-- ---------------------------------------------------------

INSERT INTO openlibrary.authors_new (id, author_name)
SELECT id, author_name
FROM openlibrary.authors_stage
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------
-- 5. Bulk load works
-- ---------------------------------------------------------

INSERT INTO openlibrary.works_new (id, title)
SELECT id, title
FROM openlibrary.works_stage
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------
-- 6. Bulk load relationships
-- ---------------------------------------------------------

INSERT INTO openlibrary.work_authors_new (work_id, author_id)
SELECT wa.work_id, wa.author_id
FROM openlibrary.work_authors_stage wa
JOIN openlibrary.works_new w ON w.id = wa.work_id
JOIN openlibrary.authors_new a ON a.id = wa.author_id
ON CONFLICT DO NOTHING;


-- ---------------------------------------------------------
-- 7. Create search document table (denormalised search)
-- ---------------------------------------------------------

CREATE UNLOGGED TABLE openlibrary.search_documents_new (
    work_id TEXT PRIMARY KEY,
    title TEXT,
    author_names TEXT,
    search_vector tsvector
);

INSERT INTO openlibrary.search_documents_new
SELECT
    w.id,
    w.title,
    string_agg(a.author_name, ' ') AS author_names,
    setweight(to_tsvector('english', w.title), 'A') ||
    setweight(to_tsvector('english', string_agg(a.author_name, ' ')), 'B')
FROM openlibrary.works_new w
JOIN openlibrary.work_authors_new wa
    ON w.id = wa.work_id
JOIN openlibrary.authors_new a
    ON wa.author_id = a.id
GROUP BY w.id, w.title;

-- ---------------------------------------------------------
-- 8. Convert tables to logged
-- ---------------------------------------------------------

ALTER TABLE openlibrary.authors_new SET LOGGED;
ALTER TABLE openlibrary.works_new SET LOGGED;
ALTER TABLE openlibrary.work_authors_new SET LOGGED;
ALTER TABLE openlibrary.search_documents_new SET LOGGED;

-- ---------------------------------------------------------
-- 9. Create indexes AFTER bulk load
-- ---------------------------------------------------------

CREATE INDEX idx_authors_name_new
ON openlibrary.authors_new (author_name);

CREATE INDEX idx_works_title_new
ON openlibrary.works_new (title);

CREATE INDEX idx_work_authors_work_new
ON openlibrary.work_authors_new (work_id);

CREATE INDEX idx_work_authors_author_new
ON openlibrary.work_authors_new (author_id);

CREATE INDEX idx_search_documents_vector_new
ON openlibrary.search_documents_new
USING GIN(search_vector);

-- ---------------------------------------------------------
-- 10. Add foreign keys without validation
-- ---------------------------------------------------------

ALTER TABLE openlibrary.work_authors_new
ADD CONSTRAINT fk_work
FOREIGN KEY (work_id)
REFERENCES openlibrary.works_new(id)
ON DELETE CASCADE
NOT VALID;

ALTER TABLE openlibrary.work_authors_new
ADD CONSTRAINT fk_author
FOREIGN KEY (author_id)
REFERENCES openlibrary.authors_new(id)
ON DELETE CASCADE
NOT VALID;

-- ---------------------------------------------------------
-- 11. Validate constraints
-- ---------------------------------------------------------

ALTER TABLE openlibrary.work_authors_new VALIDATE CONSTRAINT fk_work;
ALTER TABLE openlibrary.work_authors_new VALIDATE CONSTRAINT fk_author;

-- ---------------------------------------------------------
-- 12. Atomic table swap
-- ---------------------------------------------------------

BEGIN;

ALTER TABLE openlibrary.search_documents RENAME TO search_documents_old;
ALTER TABLE openlibrary.work_authors RENAME TO work_authors_old;
ALTER TABLE openlibrary.authors RENAME TO authors_old;
ALTER TABLE openlibrary.works RENAME TO works_old;

ALTER TABLE openlibrary.authors_new RENAME TO authors;
ALTER TABLE openlibrary.works_new RENAME TO works;
ALTER TABLE openlibrary.work_authors_new RENAME TO work_authors;
ALTER TABLE openlibrary.search_documents_new RENAME TO search_documents;

COMMIT;

-- ---------------------------------------------------------
-- 13. Cleanup old tables
-- ---------------------------------------------------------

DROP TABLE openlibrary.search_documents_old;
DROP TABLE openlibrary.work_authors_old;
DROP TABLE openlibrary.authors_old;
DROP TABLE openlibrary.works_old;

-- ---------------------------------------------------------
-- 14. Reset session settings
-- ---------------------------------------------------------

RESET maintenance_work_mem;
RESET work_mem;
RESET synchronous_commit;
RESET temp_buffers;

-- =========================================================
-- SEARCH QUERY
-- =========================================================

SELECT
    work_id,
    title,
    author_names,
    ts_rank(search_vector, plainto_tsquery($1)) AS rank
FROM openlibrary.search_documents
WHERE search_vector @@ plainto_tsquery($1)
ORDER BY rank DESC
LIMIT 20;
