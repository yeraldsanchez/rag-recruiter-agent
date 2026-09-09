package repository

import (
	"AnalizadorCVs/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type PgVectorStorer struct {
	pool *pgxpool.Pool
}

func NewPgVectorStorer(pool *pgxpool.Pool) *PgVectorStorer {
	return &PgVectorStorer{pool: pool}
}

func (r *PgVectorStorer) SaveEmbedding(ctx context.Context, candidateID int64, chunks []model.TextChunk) error {
	rows := make([][]any, len(chunks))
	for i, c := range chunks {
		rows[i] = []any{
			candidateID,
			c.Text,
			pgvector.NewVector(c.Embedding),
		}
	}

	copyCount, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"resume_chunks"},
		[]string{"candidate_id", "chunk_text", "embedding"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error inserting chunks: %w", err)
	}

	fmt.Printf("Successfully inserted %d chunks.\n", copyCount)
	return nil
}
