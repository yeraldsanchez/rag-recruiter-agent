package main

import (
	"AnalizadorCVs/internal/ai"
	"AnalizadorCVs/internal/config"
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"

	"AnalizadorCVs/internal/handler"
	"AnalizadorCVs/internal/messaging"
	"AnalizadorCVs/internal/repository"
	"AnalizadorCVs/internal/service"
	"AnalizadorCVs/internal/storage"
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
		Storage:  nats.FileStorage,
	})
	if err != nil {
		log.Printf("Stream info/status: %v", err)
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

	g := genkit.Init(context.Background(),
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
	)

	embeddingClient := ai.NewEmbedder(g)
	natsProducer := messaging.NewNatsProducer(js, natsCfg.StreamSubject)
	resumeRepo := repository.NewResumeRepository(pool)
	resumeService := service.NewResumeService(resumeRepo, natsProducer, embeddingClient, resumeStorage)
	resumeHandler := handler.NewResumeHandler(resumeService)
	aiClient := ai.NewGenkitClient(g, resumeService)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "pong")
	})

	r.POST("/resumes", resumeHandler.UploadResume)
	r.GET("/candidates", resumeHandler.GetCandidates)
	r.POST("/candidates/genkit", gin.WrapH(genkit.Handler(aiClient.CandidateFlow())))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	stop()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
