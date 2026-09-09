package ai

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type GenkitEmbedder struct {
	g *genkit.Genkit
}

func NewEmbedder(g *genkit.Genkit) *GenkitEmbedder {
	return &GenkitEmbedder{g: g}
}

func (c *GenkitEmbedder) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	resp, err := genkit.Embed(ctx, c.g,
		ai.WithEmbedderName("googleai/gemini-embedding-001"),
		ai.WithTextDocs(texts...),
	)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Embedding
	}
	return embeddings, nil
}

func (c *GenkitEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, nil
	}
	embeddings, err := c.EmbedTexts(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no genkit vector was received")
	}

	return embeddings[0], nil
}
