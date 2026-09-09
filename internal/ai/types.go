package ai

type RoleRequirementsInput struct {
	Requirements string `json:"requirements" jsonschema:"description=Free-text description of the role requirements provided by the recruiter"`
}

type CandidateSearchQuery struct {
	Query string `json:"query" jsonschema:"description=Clean, embedding-optimized natural-language description of the role/skills/experience to search for. Must contain ONLY the semantic content needed to match candidates..."`
}

type CandidateAnalysis struct {
	CandidateID   string `json:"candidateId" jsonschema:"description=Unique identifier of the candidate"`
	Name          string `json:"name" jsonschema:"description=Full name of the candidate"`
	MatchScore    int    `json:"matchScore" jsonschema:"description=Fit score from 1 to 100 measuring how well the candidate matches the requirements"`
	Justification string `json:"justification" jsonschema:"description=Brief justification for the score, grounded in the candidate's resume data and the stated requirements"`
}

type SummarizeCandidatesOutput struct {
	Summary string              `json:"summary" jsonschema:"description=Executive summary of the search results, highlighting the strongest matches and any notable gaps"`
	Ranked  []CandidateAnalysis `json:"ranked" jsonschema:"description=Candidates ordered from most to least relevant"`
}
