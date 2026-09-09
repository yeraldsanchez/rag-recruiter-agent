package messaging

import (
	"AnalizadorCVs/internal/model"
	"encoding/json"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsProducer struct {
	js            nats.JetStreamContext
	streamSubject string
}

type UploadResumeEvent struct {
	FilePath    string `json:"file_path"`
	CandidateID string `json:"candidate_id"`
	CreatedAt   string `json:"created_at"`
}

func NewNatsProducer(js nats.JetStreamContext, subject string) *NatsProducer {
	return &NatsProducer{js: js, streamSubject: subject}
}

func (p *NatsProducer) PublishUploadResumeEvent(resume model.Resume) error {
	event := UploadResumeEvent{
		FilePath:    resume.FilePath,
		CandidateID: strconv.FormatInt(resume.CandidateID, 10),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(p.streamSubject, data)
	if err != nil {
		return err
	}
	return nil
}
