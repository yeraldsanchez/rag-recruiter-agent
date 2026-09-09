package model

type Resume struct {
	ID          int64
	CandidateID int64
	FileName    string
	FilePath    string
	FileType    string
	FileSize    int64
}
