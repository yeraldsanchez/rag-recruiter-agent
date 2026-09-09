package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"

	"AnalizadorCVs/internal/model"
	"AnalizadorCVs/internal/repository"
)

var ErrCandidateAlreadyExists = errors.New("candidate already exists")

type ResumeRepository interface {
	CreateCandidateWithResume(ctx context.Context, candidate model.Candidate, resume model.Resume) (model.Candidate, model.Resume, error)
	FindCandidatesByEmbedding(ctx context.Context, embedding []float32, limit int) ([]model.CandidateMatch, error)
}

type ResumeMessaging interface {
	PublishUploadResumeEvent(resume model.Resume) error
}

type TextEmbedder interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

type ResumeStorage interface {
	Save(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (path string, err error)
	Get(ctx context.Context, objectPath string) (io.ReadCloser, error)
}

type ResumeService struct {
	repo      ResumeRepository
	messaging ResumeMessaging
	embedder  TextEmbedder
	storage   ResumeStorage
}

func NewResumeService(repo ResumeRepository, messaging ResumeMessaging, embedder TextEmbedder, storage ResumeStorage) *ResumeService {
	return &ResumeService{repo: repo, messaging: messaging, embedder: embedder, storage: storage}
}

func (s *ResumeService) UploadResume(ctx context.Context, name, email string, file io.Reader, fileHeader *multipart.FileHeader) (model.Candidate, model.Resume, error) {
	candidate := model.Candidate{Name: name, Email: email}
	filename := filepath.Base(fileHeader.Filename)

	path, err := s.storage.Save(ctx, filename, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return model.Candidate{}, model.Resume{}, err
	}

	resume := model.Resume{
		FileName: fileHeader.Filename,
		FilePath: path,
		FileType: filepath.Ext(fileHeader.Filename),
		FileSize: fileHeader.Size,
	}

	candidate, resume, err = s.repo.CreateCandidateWithResume(ctx, candidate, resume)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return model.Candidate{}, model.Resume{}, ErrCandidateAlreadyExists
		}
		return model.Candidate{}, model.Resume{}, err
	}
	err = s.messaging.PublishUploadResumeEvent(resume)
	if err != nil {
		return model.Candidate{}, model.Resume{}, err
	}
	return candidate, resume, nil
}

func (s *ResumeService) GetCandidates(ctx context.Context, query string) ([]model.CandidateMatch, error) {
	queryEmbedding, err := s.embedder.EmbedTexts(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	candidates, err := s.repo.FindCandidatesByEmbedding(ctx, queryEmbedding[0], 5)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}
