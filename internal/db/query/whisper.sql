-- name: ListWhisperGroupMembers :many
SELECT g.id AS group_id, g.name, gm.player_id, pc.channel_id
FROM whisper_group g
JOIN whisper_group_member gm ON gm.group_id = g.id
LEFT JOIN player_confessional pc ON pc.player_id = gm.player_id
ORDER BY g.name, gm.player_id;

-- name: ListWhisperGroups :many
SELECT g.id, g.name, gm.player_id
FROM whisper_group g
LEFT JOIN whisper_group_member gm ON gm.group_id = g.id
ORDER BY g.name, gm.player_id;

-- name: CreateWhisperGroup :one
INSERT INTO whisper_group (name) VALUES ($1) RETURNING *;

-- name: DeleteWhisperGroup :exec
DELETE FROM whisper_group WHERE id = $1;

-- name: AddWhisperGroupMember :exec
INSERT INTO whisper_group_member (group_id, player_id) VALUES ($1, $2);

-- name: RemoveWhisperGroupMember :exec
DELETE FROM whisper_group_member WHERE group_id = $1 AND player_id = $2;

-- name: ListEnabledWhisperDoubtMessages :many
SELECT * FROM whisper_doubt_message
WHERE enabled AND deleted_at IS NULL
ORDER BY id;

-- name: ListWhisperDoubtMessages :many
SELECT * FROM whisper_doubt_message
WHERE deleted_at IS NULL
ORDER BY id;

-- name: CreateWhisperDoubtMessage :one
INSERT INTO whisper_doubt_message (message) VALUES ($1) RETURNING *;

-- name: UpdateWhisperDoubtMessage :one
UPDATE whisper_doubt_message
SET message = $2, enabled = $3, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWhisperDoubtMessage :exec
UPDATE whisper_doubt_message SET deleted_at = NOW(), enabled = FALSE, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
