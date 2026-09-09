package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"AnalizadorCVs/internal/model"
)

var ErrDuplicateEmail = errors.New("candidate email already exists")

const uniqueViolationCode = "23505"

type ResumeRepository struct {
	db *pgxpool.Pool
}

func NewResumeRepository(db *pgxpool.Pool) *ResumeRepository {
	return &ResumeRepository{db: db}
}

func (r *ResumeRepository) CreateCandidateWithResume(ctx context.Context, candidate model.Candidate, resume model.Resume) (model.Candidate, model.Resume, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Candidate{}, model.Resume{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		"INSERT INTO candidates (name, email) VALUES ($1, $2) RETURNING id",
		candidate.Name, candidate.Email,
	).Scan(&candidate.ID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolationCode {
			return model.Candidate{}, model.Resume{}, ErrDuplicateEmail
		}
		return model.Candidate{}, model.Resume{}, err
	}

	resume.CandidateID = candidate.ID
	err = tx.QueryRow(
		ctx,
		"INSERT INTO resumes (candidate_id, file_name, file_path, file_type, file_size) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		resume.CandidateID, resume.FileName, resume.FilePath, resume.FileType, resume.FileSize,
	).Scan(&resume.ID)
	if err != nil {
		return model.Candidate{}, model.Resume{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Candidate{}, model.Resume{}, err
	}

	return candidate, resume, nil
}

func (r *ResumeRepository) FindCandidatesByEmbedding(ctx context.Context, embedding []float32, limit int) ([]model.CandidateMatch, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			c.id AS candidate_id,
			c.name,
			c.email,
			rc.chunk_text,
			1 - (rc.embedding <=> $1) AS similarity
		FROM resume_chunks rc
		JOIN candidates c ON c.id = rc.candidate_id
		ORDER BY rc.embedding <=> $1
		LIMIT $2`,
		pgvector.NewVector(embedding), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []model.CandidateMatch
	for rows.Next() {
		var m model.CandidateMatch
		if err := rows.Scan(&m.CandidateID, &m.Name, &m.Email, &m.ChunkText, &m.Similarity); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}
