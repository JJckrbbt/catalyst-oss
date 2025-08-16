-- name: ListParkVisitationRecords :many
-- Fetches all park visitation records, ordered by the highest visitor count first.
SELECT *
FROM vw_nps_visitation
ORDER BY visitor_count DESC;

-- name: GetTotalVisitorsForPark :one
-- Calculates the sum of all visitors for a single park across all recorded years.
SELECT park_name, SUM(visitor_count) AS total_visitors
FROM vw_nps_visitation
WHERE park_name = $1
GROUP BY park_name;

-- name: ListParksByState :many
-- Fetches all park visitation records for a specific state.
SELECT *
FROM vw_nps_visitation
WHERE state_code = $1
ORDER BY year DESC, visitor_count DESC;
