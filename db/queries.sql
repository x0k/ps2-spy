-- name: InsertChannelOutfit :exec
INSERT INTO
  channel_to_outfit (channel_id, platform, outfit_id)
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
  channel_to_character (channel_id, platform, character_id)
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
  outfit_to_character (platform, outfit_id, character_id)
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
  outfit_synchronization (platform, outfit_id, synchronized_at)
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

-- name: ListPlatformTrackingChannelsForCharacter :many
SELECT
  channel.channel_id,
  channel_locale.locale
FROM
  (
    SELECT
      channel_id
    FROM
      channel_to_character
    WHERE
      channel_to_character.platform = sqlc.arg (platform)
      AND character_id = ?
    UNION
    SELECT
      channel_id
    FROM
      channel_to_outfit
    WHERE
      channel_to_outfit.platform = sqlc.arg (platform)
      AND outfit_id = ?
  ) AS channel
  LEFT JOIN channel_locale ON channel_locale.channel_id = subquery.channel_id;

-- name: ListPlatformTrackingChannelsForOutfit :many
SELECT
  channel_to_outfit.channel_id,
  channel_locale.locale
FROM
  channel_to_outfit
  LEFT JOIN channel_locale ON channel_locale.channel_id = channel_to_outfit.channel_id
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
  channel_to_character.platform = sqlc.arg (platform)
UNION ALL
SELECT
  character_id
FROM
  channel_to_outfit
  JOIN outfit_to_character ON channel_to_outfit.outfit_id = outfit_to_character.outfit_id
  AND channel_to_outfit.platform = outfit_to_character.platform
WHERE
  channel_to_outfit.platform = sqlc.arg (platform);

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
  outfit (platform, outfit_id, outfit_name, outfit_tag)
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
  facility (
    facility_id,
    facility_name,
    facility_type,
    zone_id
  )
VALUES
  (?, ?, ?, ?);

-- name: ListPlatformOutfits :many
SELECT
  *
FROM
  outfit
WHERE
  platform = ?
  AND outfit_id IN (sqlc.slice (outfit_ids));