CREATE SCHEMA openlibrary;
CREATE TABLE openlibrary.authors (
    id TEXT PRIMARY KEY,
    author_name TEXT NOT NULL
);

CREATE TABLE openlibrary.works (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL
);

CREATE TABLE openlibrary.work_authors (
    work_id   TEXT NOT NULL,
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

CREATE INDEX idx_authors_name
ON openlibrary.authors(author_name);

CREATE INDEX idx_works_title
ON openlibrary.works(title);

CREATE INDEX idx_work_authors_author
ON openlibrary.work_authors(author_id);

CREATE UNLOGGED TABLE openlibrary.authors_stage(
    id TEXT,
    author_name TEXT
);

CREATE UNLOGGED TABLE openlibrary.works_stage(
    id TEXT,
    title TEXT
);

CREATE UNLOGGED TABLE openlibrary.work_authors_stage(
    work_id TEXT,
    author_id TEXT
);

CREATE UNLOGGED TABLE openlibrary.search_documents (
    work_id TEXT PRIMARY KEY,
    title TEXT,
    author_names TEXT,
    search_vector tsvector
);
