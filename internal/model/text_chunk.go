package model

type TextChunk struct {
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
}

func NewTextChunk(text string, embedding []float32) TextChunk {
	return TextChunk{
		Text:      text,
		Embedding: embedding,
	}
}

func ZipChunks(texts []string, embeddings [][]float32) []TextChunk {
	chunks := make([]TextChunk, len(texts))
	for i := range texts {
		chunks[i] = TextChunk{
			Text:      texts[i],
			Embedding: embeddings[i],
		}
	}
	return chunks
}
