-- name: UpsertVote :one
INSERT INTO vote (voter_id, target_id, cycle_day, is_elimination, weight, context, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (voter_id, cycle_day, is_elimination)
DO UPDATE SET 
    target_id = EXCLUDED.target_id,
    weight = EXCLUDED.weight,
    context = EXCLUDED.context,
    updated_at = NOW()
RETURNING *;

-- name: GetVote :one
SELECT * FROM vote WHERE id = $1;

-- name: GetVoteByVoterAndCycle :one
SELECT * FROM vote 
WHERE voter_id = $1 AND cycle_day = $2 AND is_elimination = $3;

-- name: ListVotesByCycle :many
SELECT * FROM vote 
WHERE cycle_day = $1 AND is_elimination = $2
ORDER BY updated_at DESC;

-- name: ListVotesByTarget :many
SELECT * FROM vote 
WHERE target_id = $1
ORDER BY cycle_day DESC, is_elimination DESC;

-- name: ListVotesByVoter :many
SELECT * FROM vote 
WHERE voter_id = $1
ORDER BY cycle_day DESC, is_elimination DESC;

-- name: ListAllVotes :many
SELECT * FROM vote
ORDER BY cycle_day DESC, is_elimination DESC, updated_at DESC;

-- name: DeleteVote :exec
DELETE FROM vote WHERE id = $1;

-- name: DeleteVotesByCycle :exec
DELETE FROM vote WHERE cycle_day = $1 AND is_elimination = $2;

-- name: WipeAllVotes :exec
DELETE FROM vote;

-- name: CountVotesForTarget :one
SELECT COALESCE(SUM(weight), 0)::INTEGER as total_votes
FROM vote
WHERE target_id = $1 AND cycle_day = $2 AND is_elimination = $3;

-- name: GetVoteTalliesByCycle :many
SELECT 
    target_id,
    COALESCE(SUM(weight), 0)::INTEGER as total_votes,
    COUNT(*) as vote_count
FROM vote
WHERE cycle_day = $1 AND is_elimination = $2
GROUP BY target_id
ORDER BY total_votes DESC;

-- name: GetMostVotedPlayer :one
SELECT 
    target_id,
    COALESCE(SUM(weight), 0)::INTEGER as total_votes
FROM vote
GROUP BY target_id
ORDER BY total_votes DESC
LIMIT 1;

-- name: GetVoteStatsByPlayer :many
SELECT 
    target_id,
    COALESCE(SUM(weight), 0)::INTEGER as total_votes_received,
    COUNT(*) as times_voted_for
FROM vote
GROUP BY target_id
ORDER BY total_votes_received DESC;

-- name: GetDistinctCyclesWithVotes :many
SELECT DISTINCT cycle_day, is_elimination
FROM vote
ORDER BY cycle_day DESC, is_elimination DESC;

-- name: GetVoterParticipation :many
SELECT 
    voter_id,
    COUNT(*) as votes_cast
FROM vote
GROUP BY voter_id
ORDER BY votes_cast DESC;
