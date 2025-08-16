-- +goose Up
-- Create tables for comments and status history

CREATE TABLE "comments" (
	"id" BIGSERIAL PRIMARY KEY,
	"item_id" BIGINT NOT NULL REFERENCES "items"("id") ON DELETE CASCADE,
	"comment" TEXT NOT NULL,
	"user_id" BIGINT NOT NULL REFERENCES "users"("id"),
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "status_history" (
	"id" BIGSERIAL PRIMARY KEY,
	"item_id" BIGINT NOT NULL REFERENCES "items"("id") ON DELETE CASCADE,
	"status" item_status NOT NULL,
	"notes" TEXT,
	"user_id" BIGINT NOT NULL REFERENCES "users"("id"),
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "items_events" (
	"id" BIGSERIAL PRIMARY KEY,
	"item_id" BIGINT NOT NULL REFERENCES "items"("id") ON DELETE RESTRICT,
	"event_type" VARCHAR(100) NOT NULL,
	"event_data" JSONB NOT NULL, 
	"created_by" BIGINT NOT NULL REFERENCES "users"("id"),
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for quickly finding all events for  specific use
CREATE INDEX idx_item_events_item_id ON "items_events" (item_id);

-- Index for ordering events chronologically
CREATE INDEX idx_item_events_created_at ON "items_events" (created_at DESC);


-- +goose Down
DROP TABLE IF EXISTS "items_events";
DROP TABLE IF EXISTS "status_history";
DROP TABLE IF EXISTS "comments";
