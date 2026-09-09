package handler

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"

	"AnalizadorCVs/internal/model"
	"AnalizadorCVs/internal/service"
)

type ResumeService interface {
	UploadResume(ctx context.Context, name, email string, file io.Reader, fileHeader *multipart.FileHeader) (model.Candidate, model.Resume, error)
	GetCandidates(ctx context.Context, query string) ([]model.CandidateMatch, error)
}

type ResumeHandler struct {
	service ResumeService
}

func NewResumeHandler(service ResumeService) *ResumeHandler {
	return &ResumeHandler{service: service}
}

func (h *ResumeHandler) UploadResume(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	if name == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and email are required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	candidate, resume, err := h.service.UploadResume(c.Request.Context(), name, email, file, fileHeader)
	if err != nil {
		if errors.Is(err, service.ErrCandidateAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"candidate": candidate,
		"resume":    resume,
	})
}

func (h *ResumeHandler) GetCandidates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	candidates, err := h.service.GetCandidates(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}
