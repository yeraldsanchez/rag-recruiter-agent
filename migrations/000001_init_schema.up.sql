CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE candidates (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE resumes (
    id BIGSERIAL PRIMARY KEY,
    candidate_id BIGINT NOT NULL REFERENCES candidates (id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_size BIGINT NOT NULL
);

CREATE TABLE resume_chunks (
    id BIGSERIAL PRIMARY KEY,
    candidate_id BIGINT NOT NULL REFERENCES candidates (id) ON DELETE CASCADE,
    chunk_text TEXT NOT NULL,
    embedding VECTOR(3072) NOT NULL
);

CREATE INDEX idx_resumes_candidate_id ON resumes (candidate_id);
CREATE INDEX idx_resume_chunks_candidate_id ON resume_chunks (candidate_id);
