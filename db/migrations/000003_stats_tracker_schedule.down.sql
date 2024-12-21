DROP INDEX idx_stats_tracker_task;

DROP TABLE stats_tracker_task;

CREATE TABLE
  channel_tmp (
    channel_id TEXT PRIMARY KEY NOT NULL,
    locale TEXT NOT NULL DEFAULT 'en',
    character_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    outfit_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    title_updates BOOLEAN NOT NULL DEFAULT TRUE
  );

INSERT INTO
  channel_tmp (
    channel_id,
    locale,
    character_notifications,
    outfit_notifications,
    title_updates
  )
SELECT
  channel_id,
  locale,
  character_notifications,
  outfit_notifications,
  title_updates
FROM
  channel;

DROP TABLE channel;

ALTER TABLE channel_tmp
RENAME TO channel;