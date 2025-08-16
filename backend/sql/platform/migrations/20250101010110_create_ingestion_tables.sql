-- +goose Up

-- The "ingestion_jobs" table tracks the status and metadata of each job
CREATE TABLE "ingestion_jobs" (
	"id" UUID PRIMARY KEY,
	"source_type" VARCHAR(50) NOT NULL,
	"source_details" JSONB,
	"report_type" TEXT NOT NULL,
	"status" VARCHAR(50) NOT NULL,
	"started_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	"completed_at" TIMESTAMPTZ,
	"error_details" TEXT,
	"user_id" BIGINT REFERENCES "users"("id"),
	"source_uri" TEXT,
	"rows_upserted" INTEGER,
	"rows_triaged" INTEGER
);


-- The ingestion_errors table stores details about rows that fail validation
CREATE TABLE "ingestion_errors" (
	"id" UUID PRIMARY KEY,
	"job_id" UUID NOT NULL REFERENCES "ingestion_jobs"("id") ON DELETE CASCADE,
	"timestamp" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	"original_row_data" JSONB NOT NULL,
	"reason_for_failure" TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS "ingestion_errors";
DROP TABLE IF EXISTS "ingestion_jobs";


