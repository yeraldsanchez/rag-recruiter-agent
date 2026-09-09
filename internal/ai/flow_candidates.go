package ai

import (
	"AnalizadorCVs/internal/model"
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"google.golang.org/genai"
)

func (c *GenkitFlowClient) CandidateFlow() *core.Flow[RoleRequirementsInput, *SummarizeCandidatesOutput, *SummarizeCandidatesOutput] {
	return c.candidateFlow
}

func (c *GenkitFlowClient) registerFlows() {
	getCandidatesTool := genkit.DefineTool(c.g, "getCandidatesTool",
		"Searches the resume database for candidates that semantically match a role description, and returns each matching candidate with the resume text fragments that matched plus a similarity score. This is the ONLY way to retrieve candidate data — call it with a clean query built from the recruiter's request, containing only the actual hiring criteria (role, skills, seniority, technologies, domain, etc). Never pass the recruiter's raw message verbatim: first strip conversational text, greetings, meta-requests, and formatting/language instructions (e.g. 'please', 'I need a...', 'respond in Spanish') before calling this tool.",
		func(toolCtx *ai.ToolContext, input CandidateSearchQuery) ([]model.CandidateMatch, error) {
			candidates, err := c.service.GetCandidates(toolCtx, input.Query)
			if err != nil {
				return nil, err
			}
			return candidates, nil
		},
	)

	c.candidateFlow = genkit.DefineStreamingFlow(c.g, "getCandidatesFlow",
		func(ctx context.Context, input RoleRequirementsInput, sendChunk func(context.Context, *SummarizeCandidatesOutput) error) (*SummarizeCandidatesOutput, error) {
			prompt := fmt.Sprintf(`You are a technical recruiting assistant that evaluates candidates against a role's requirements.

Recruiter's raw request:
%s

Instructions:
1. From the recruiter's raw request above, extract the actual hiring criteria (role, seniority, required skills, technologies, domain, years of experience, etc). Ignore and discard any greetings, conversational filler, meta-requests, or output-formatting/language instructions (e.g. "respond in Spanish", "please", "I need...").
2. Call getCandidatesTool with that clean, embedding-optimized query to retrieve candidate data. You MUST call this tool before producing your answer — do not answer from memory or assumptions.
3. Use ONLY the candidate data returned by the tool, which was extracted from resumes stored in the system. Do not invent skills, experience, or facts that are not present in the data. If the data is insufficient to judge a requirement, say so explicitly instead of guessing.
4. Evaluate every candidate returned against the stated requirements.
5. Assign each candidate a matchScore from 1 to 100, where 100 means the candidate fully satisfies the requirements and 1 means almost no overlap.
6. Write a short, specific justification for each score, referencing concrete evidence from the candidate's data (skills, experience, tools) rather than generic statements.
7. Order the "ranked" list from the highest matchScore to the lowest.
8. Write a concise executive summary (2-4 sentences) highlighting the top candidates and any notable gaps across the pool.
9. If no candidates match the requirements, or the tool returns none, state that clearly in the summary instead of forcing a positive assessment.
`, input.Requirements,
			)
			stream := genkit.GenerateDataStream[*SummarizeCandidatesOutput](ctx, c.g,
				ai.WithModelName("googleai/gemini-3.5-flash"),
				ai.WithConfig(&genai.GenerateContentConfig{
					ThinkingConfig: &genai.ThinkingConfig{
						ThinkingLevel: genai.ThinkingLevelLow,
					},
				}),
				ai.WithPrompt(prompt), ai.WithTools(getCandidatesTool),
			)
			for result, err := range stream {
				if err != nil {
					return nil, fmt.Errorf("failed to generate candidate summary: %w", err)
				}
				if result.Done {
					return result.Output, nil
				}
				if result.Chunk != nil {
					if err = sendChunk(ctx, result.Chunk); err != nil {
						return nil, err
					}
				}
			}
			return nil, fmt.Errorf("stream ended without a final candidate summary")
		})
}
