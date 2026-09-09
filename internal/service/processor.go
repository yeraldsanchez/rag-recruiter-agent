package service

import (
	"AnalizadorCVs/internal/model"
	"context"
	"io"
	"strconv"
)

type TextExtractor interface {
	ExtractText(r io.Reader) (string, error)
}

type VectorStorer interface {
	SaveEmbedding(ctx context.Context, candidateID int64, chunks []model.TextChunk) error
}
type ResumeProcessorService struct {
	storage      ResumeStorage
	extractor    TextExtractor
	embedder     TextEmbedder
	vectorStorer VectorStorer
}

func NewResumeProcessorService(storage ResumeStorage, extractor TextExtractor, embedder TextEmbedder, vectorStorer VectorStorer) ResumeProcessorService {
	return ResumeProcessorService{
		storage:      storage,
		extractor:    extractor,
		embedder:     embedder,
		vectorStorer: vectorStorer,
	}
}

func (s *ResumeProcessorService) ProcessResume(ctx context.Context, candidateId, filepath string) error {
	file, err := s.storage.Get(ctx, filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	text, err := s.extractor.ExtractText(file)
	if err != nil {
		return err
	}
	vector, err := s.embedder.EmbedText(ctx, text)
	if err != nil {
		return err
	}
	chunks := []model.TextChunk{
		model.NewTextChunk(text, vector),
	}

	intCandidateID, err := strconv.ParseInt(candidateId, 10, 64)
	if err != nil {
		return err
	}
	err = s.vectorStorer.SaveEmbedding(ctx, intCandidateID, chunks)
	if err != nil {
		return err
	}
	return nil
}
