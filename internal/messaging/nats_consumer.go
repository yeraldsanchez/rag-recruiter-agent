package messaging

import (
	"AnalizadorCVs/internal/service"
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type NatsConsumer struct {
	js               nats.JetStreamContext
	streamSubject    string
	durableName      string
	processorService service.ResumeProcessorService
}

func NewNatsConsumer(js nats.JetStreamContext, subject, durableName string, procSvc service.ResumeProcessorService) *NatsConsumer {
	return &NatsConsumer{
		js:               js,
		streamSubject:    subject,
		durableName:      durableName,
		processorService: procSvc,
	}
}

func (c *NatsConsumer) StartWorker() (*nats.Subscription, error) {
	opts := []nats.SubOpt{
		nats.Durable(c.durableName),
		nats.ManualAck(),
	}

	return c.js.Subscribe(c.streamSubject, func(msg *nats.Msg) {
		var event UploadResumeEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			msg.Nak()
			return
		}

		ctx := context.Background()
		if err := c.processorService.ProcessResume(ctx, event.CandidateID, event.FilePath); err != nil {
			log.Printf("Error procesando CV del candidato %s: %v", event.CandidateID, err)
			msg.Nak() // Pide reintento a NATS
			return
		}

		msg.Ack()
	}, opts...)
}
