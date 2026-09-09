package main

import (
	"AnalizadorCVs/internal/ai"
	"AnalizadorCVs/internal/config"
	"AnalizadorCVs/internal/messaging"
	"AnalizadorCVs/internal/repository"
	"AnalizadorCVs/internal/service"
	"AnalizadorCVs/internal/storage"
	extractor "AnalizadorCVs/internal/text_extractor"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbCfg, err := config.LoadDBConfig()
	if err != nil {
		log.Fatal("Error loading database config")
	}

	poolConfig, err := pgxpool.ParseConfig(dbCfg.DSN)
	if err != nil {
		log.Fatalf("Invalid database URL: %v", err)
	}
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	natsCfg, err := config.LoadNatsConfig()
	if err != nil {
		log.Fatalf("Unable to load nats config: %v", err)
	}

	nc, err := nats.Connect(natsCfg.URL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error activating JetStream: %v", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     natsCfg.StreamName,
		Subjects: []string{natsCfg.StreamSubject},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Fatalf("Error creating/verifying NATS stream: %v", err)
	}

	minioCfg, err := config.LoadMinioConfig()
	if err != nil {
		log.Fatalf("Error loading minio config: %v", err)
	}

	resumeStorage, err := storage.NewMinioResumeStorage(context.Background(),
		minioCfg.Endpoint,
		minioCfg.AccessKey,
		minioCfg.SecretKey,
		minioCfg.Bucket,
		minioCfg.UseSSL,
	)
	if err != nil {
		log.Fatalf("Unable to connect to MinIO: %v", err)
	}
	textExtractor := extractor.NewPDFExtractor()

	g := genkit.Init(context.Background(),
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
	)

	embeddingClient := ai.NewEmbedder(g)
	vectorStorer := repository.NewPgVectorStorer(pool)
	processorService := service.NewResumeProcessorService(resumeStorage, textExtractor, embeddingClient, vectorStorer)

	durableName := "resume-processor-worker"
	consumer := messaging.NewNatsConsumer(js, natsCfg.StreamSubject, durableName, processorService)

	sub, err := consumer.StartWorker()
	if err != nil {
		log.Fatalf("Error starting the NATS consumer worker: %v", err)
	}
	defer sub.Drain()

	log.Printf("Worker started, listening on subject: %s (Durable: %s)", natsCfg.StreamSubject, durableName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker gracefully...")
}
