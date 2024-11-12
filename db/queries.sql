-- name: InsertChannelOutfit :exec
INSERT INTO
  channel_to_outfit
VALUES
  (?, ?, ?);

-- name: DeleteChannelOutfit :exec
DELETE FROM channel_to_outfit
WHERE
  channel_id = ?
  AND platform = ?
  AND outfit_id = ?;

-- name: InsertChannelCharacter :exec
INSERT INTO
  channel_to_character
VALUES
  (?, ?, ?);

-- name: DeleteChannelCharacter :exec
DELETE FROM channel_to_character
WHERE
  channel_id = ?
  AND platform = ?
  AND character_id = ?;

-- name: InsertOutfitMember :exec
INSERT INTO
  outfit_to_character
VALUES
  (?, ?, ?);

-- name: DeleteOutfitMember :exec
DELETE FROM outfit_to_character
WHERE
  platform = ?
  AND outfit_id = ?
  AND character_id = ?;

-- name: UpsertPlatformOutfitSynchronizedAt :exec
INSERT INTO
  outfit_synchronization
VALUES
  (?, ?, ?) ON CONFLICT (platform, outfit_id) DO
UPDATE
SET
  synchronized_at = EXCLUDED.synchronized_at;

-- name: GetPlatformOutfitSynchronizedAt :one
SELECT
  synchronized_at
FROM
  outfit_synchronization
WHERE
  platform = ?
  AND outfit_id = ?;

-- name: ListPlatformTrackingChannelIdsForCharacter :many
SELECT
  channel_id
FROM
  channel_to_character
WHERE
  platform = ?
  AND character_id = ?
UNION
SELECT
  channel_id
FROM
  channel_to_outfit
WHERE
  platform = ?
  AND outfit_id = ?;

-- name: ListPlatformTrackingChannelIdsForOutfit :many
SELECT
  channel_id
FROM
  channel_to_outfit
WHERE
  platform = ?
  AND outfit_id = ?;

-- name: ListChannelOutfitIdsForPlatform :many
SELECT
  outfit_id
FROM
  channel_to_outfit
WHERE
  channel_id = ?
  AND platform = ?;

-- name: ListChannelCharacterIdsForPlatform :many
SELECT
  character_id
FROM
  channel_to_character
WHERE
  channel_id = ?
  AND platform = ?;

-- name: ListTrackableCharacterIdsWithDuplicationForPlatform :many
SELECT
  character_id
FROM
  channel_to_character
WHERE
  platform = ?
UNION ALL
SELECT
  character_id
FROM
  channel_to_outfit
  JOIN outfit_to_character ON channel_to_outfit.outfit_id = outfit_to_character.outfit_id
  AND channel_to_outfit.platform = outfit_to_character.platform
WHERE
  channel_to_outfit.platform = ?;

-- name: ListTrackableOutfitIdsWithDuplicationForPlatform :many
SELECT
  outfit_id
FROM
  channel_to_outfit
WHERE
  platform = ?;

-- name: ListUniqueTrackableOutfitIdsForPlatform :many
SELECT DISTINCT
  outfit_id
FROM
  channel_to_outfit
WHERE
  platform = ?;

-- name: ListPlatformOutfitMembers :many
SELECT
  character_id
FROM
  outfit_to_character
WHERE
  platform = ?
  AND outfit_id = ?;

-- name: GetPlatformOutfit :one
SELECT
  *
FROM
  outfit
WHERE
  platform = ?
  AND outfit_id = ?;

-- name: InsertOutfit :exec
INSERT INTO
  outfit
VALUES
  (?, ?, ?, ?);

-- name: GetFacility :one
SELECT
  *
FROM
  facility
WHERE
  facility_id = ?;

-- name: InsertFacility :exec
INSERT INTO
  facility
VALUES
  (?, ?, ?, ?);

-- name: ListPlatformOutfits :many
SELECT
  *
FROM
  outfit
WHERE
  platform = ?
  AND outfit_id IN (sqlc.slice("outfitIds"));
