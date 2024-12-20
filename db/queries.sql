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
  *
FROM
  channel
WHERE
  channel.channel_id IN (
    SELECT
      channel_id
    FROM
      channel_to_character
    WHERE
      channel_to_character.platform = sqlc.arg (platform)
      AND character_id = sqlc.arg (character_id)
    UNION
    SELECT
      channel_id
    FROM
      channel_to_outfit
    WHERE
      channel_to_outfit.platform = sqlc.arg (platform)
      AND outfit_id = sqlc.arg (outfit_id)
  );

-- name: ListPlatformTrackingChannelsForOutfit :many
SELECT
  *
FROM
  channel
WHERE
  channel_id IN (
    SELECT
      channel_id
    FROM
      channel_to_outfit
    WHERE
      platform = ?
      AND outfit_id = ?
  );

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

-- name: UpsertChannelLanguage :exec
INSERT INTO
  channel (channel_id, locale)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  locale = EXCLUDED.locale;

-- name: UpsertChannelCharacterNotifications :exec
INSERT INTO
  channel (channel_id, character_notifications)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  character_notifications = EXCLUDED.character_notifications;

-- name: UpsertChannelOutfitNotifications :exec
INSERT INTO
  channel (channel_id, outfit_notifications)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  outfit_notifications = EXCLUDED.outfit_notifications;

-- name: UpsertChannelTitleUpdates :exec
INSERT INTO
  channel (channel_id, title_updates)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  title_updates = EXCLUDED.title_updates;

-- name: UpsertChannelDefaultTimezone :exec
INSERT INTO
  channel (channel_id, default_timezone)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  default_timezone = EXCLUDED.default_timezone;

-- name: GetChannel :one
SELECT
  *
FROM
  channel
WHERE
  channel_id = ?;

-- name: ListChannelTrackablePlatforms :many
SELECT DISTINCT
  platform
FROM
  channel_to_character
WHERE
  channel_to_character.channel_id = sqlc.arg (channel_id)
UNION
SELECT DISTINCT
  platform
FROM
  channel_to_outfit
WHERE
  channel_to_outfit.channel_id = sqlc.arg (channel_id);

-- name: ListActiveStatsTrackerTasks :many
SELECT
  channel_id
FROM
  stats_tracker_task
WHERE
  (
    (
      sqlc.arg (utc_weekday) = utc_start_weekday
      AND sqlc.arg (utc_time) >= utc_start_time
    )
    OR (sqlc.arg (utc_weekday) > utc_start_weekday)
    OR (
      sqlc.arg (utc_weekday) = 0
      AND utc_start_weekday = 6
    )
  )
  AND (
    (
      sqlc.arg (utc_weekday) = utc_end_weekday
      AND sqlc.arg (utc_time) < utc_end_time
    )
    OR (sqlc.arg (utc_weekday) < utc_end_weekday)
    OR (
      sqlc.arg (utc_weekday) = 6
      AND utc_end_weekday = 0
    )
  );

-- name: ListChannelStatsTrackerTasks :many
SELECT
  *
FROM
  stats_tracker_task
WHERE
  channel_id = ?;

-- name: ListChannelIntersectingStatsTrackerTasks :many
SELECT
  *
FROM
  stats_tracker_task
WHERE
  channel_id = ?
  AND (
    (
      ?2 - utc_start_weekday IN (1, -6)
    )
    OR (
      sqlc.arg (end_weekday) = utc_start_weekday
      AND sqlc.arg (end_time) > utc_start_time
    )
  )
  AND (
    (
      utc_end_weekday - ?4 IN (1, -6)
    )
    OR (
      sqlc.arg (start_weekday) = utc_end_weekday
      AND sqlc.arg (start_time) < utc_end_time
    )
  );

-- name: InsertChannelStatsTrackerTask :exec
INSERT INTO
  stats_tracker_task (
    channel_id,
    utc_start_weekday,
    utc_start_time,
    utc_end_weekday,
    utc_end_time
  )
VALUES
  (?, ?, ?, ?, ?);