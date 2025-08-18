-- name: GetMissionFacts :one
-- Fetches the structured facts for a single Apollo mission by its name.
SELECT *
FROM vw_apollo_mission_facts
WHERE mission_name = $1;

-- name: FindSimilarMissionKnowledge :many
-- Finds knowledge chunks from mission reports that are semantically similar
-- to a given input embedding, ordered by similarity (vector cosine distance).
SELECT
    id,
    COALESCE(chunk_text, '')::TEXT AS chunk_text,
    embedding <=> $1 AS similarity_score
FROM
    vw_apollo_mission_knowledge
ORDER BY
    similarity_score
LIMIT 5;
