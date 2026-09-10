# RAG Recruiter Agent

An AI-powered candidate screening system that processes resumes, runs semantic vector searches, and leverages GenAI to match candidate profiles against job requirements.

It uses an event-driven setup to handle file storage, background processing, and LLM-based candidate ranking asynchronously.

## Quickstart with Docker

1. Environment setup

   ```bash
   cp .env.example .env # Edit .env and set your GEMINI_API_KEY
   ```

2. Start core services & run migrations

   ```bash
   docker compose up -d db nats minio
   docker compose --profile tools run --rm migrate up
   ```

3. Build & launch the application

   ```bash
   docker compose up --build -d api worker
   ```

4. Verify it's running

   ```bash
   curl http://localhost:8080/ping
   ```

## Architecture

The system uses an event-driven, decoupled architecture to separate HTTP handling from heavy background tasks (PDF text extraction, embedding generation, and vector indexing).

![System Architecture](docs/rag-recruiter-agent---system.svg)

- **API:** Serves client requests, saves raw uploads, queries embeddings, and triggers GenAI evaluations.

- **NATS JetStream:** Message bus that decouples resume uploads from processing workers.

- **Worker:** Consumes events from NATS, reads raw PDFs, extracts text, generates embeddings, and indexes them.

- **PostgreSQL (pgvector):** Stores candidate metadata and vector embeddings for fast semantic searching.

- **MinIO:** S3-compatible object storage for raw resume files.

- **Gemini (Genkit):** Generates embeddings for chunks and executes RAG flows to rank/evaluate candidates.

## Key Endpoints & Workflow

1. **Upload Resume**: `POST /resumes` with multipart form data (`name`, `email`, `file`). 

   Saves the PDF to MinIO, records metadata, and publishes a processing event to NATS.
2. **Semantic Search**: `GET /candidates?q=Senior backend Go engineer`

   Converts the search query into an embedding and fetches top matches from PostgreSQL via pgvector.
3. **AI Candidate Evaluation**: `POST /candidates/genkit`

   Executes a Genkit flow that matches, ranks, and provides a structured evaluation of fitting candidates.

   ```json
   {
    "data": {
      "requirements": "Senior backend Go engineer with distributed systems and NATS experience"
    }
   }
   ```

## Configuration

Set these environment variables in your `.env` file:

| variable            | description                                                       |
|---------------------|-------------------------------------------------------------------|
| DATABASE_URL        | PostgreSQL connection string for local app execution              |
| DOCKER_DATABASE_URL | PostgreSQL connection string used by containers in Docker network |
| POSTGRES_USER       | PostgreSQL username (used by Docker services)                     |
| POSTGRES_PASSWORD   | PostgreSQL password (used by Docker services)                     |
| POSTGRES_DB         | PostgreSQL database name (used by Docker services)                |
| NATS_URL            | NATS server URL                                                   |
| NATS_STREAM_NAME    | JetStream stream name used for resume events                      |
| NATS_STREAM_SUBJECT | JetStream subject used when publishing upload events              |
| MINIO_ENDPOINT      | MinIO endpoint (`host:port`)                                      |
| MINIO_ACCESS_KEY    | MinIO access key                                                  |
| MINIO_SECRET_KEY    | MinIO secret key                                                  |
| MINIO_BUCKET        | Bucket name where resumes are stored                              |
| MINIO_USE_SSL       | Enables HTTPS for MinIO (`true`/`false`)                          |
| GEMINI_API_KEY      | Gemini API key used by Genkit/AI flows                            |
