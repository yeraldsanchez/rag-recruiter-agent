package ai

import (
	"AnalizadorCVs/internal/service"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type GenkitFlowClient struct {
	g             *genkit.Genkit
	service       *service.ResumeService
	candidateFlow *core.Flow[RoleRequirementsInput, *SummarizeCandidatesOutput, *SummarizeCandidatesOutput]
}

func NewGenkitClient(g *genkit.Genkit, service *service.ResumeService) *GenkitFlowClient {
	client := &GenkitFlowClient{
		g:       g,
		service: service,
	}
	client.registerFlows()
	return client
}
