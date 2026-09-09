package model

type CandidateMatch struct {
	CandidateID int64   `json:"candidateId" jsonschema:"description=Unique identifier of the candidate"`
	Name        string  `json:"name" jsonschema:"description=Full name of the candidate"`
	Email       string  `json:"email" jsonschema:"description=Contact email address of the candidate"`
	ChunkText   string  `json:"chunkText" jsonschema:"description=Excerpt of the resume text that matched the search query"`
	Similarity  float64 `json:"similarity" jsonschema:"description=Semantic similarity score between the query and the resume chunk, ranging from 0 to 1"`
}
