CREATE TABLE
  stats_tracker_task (
    task_id INTEGER PRIMARY KEY NOT NULL,
    channel_id TEXT NOT NULL,
    utc_start_weekday INTEGER NOT NULL CHECK (utc_start_weekday BETWEEN 0 AND 6),
    utc_start_time INTEGER NOT NULL,
    utc_end_weekday INTEGER NOT NULL CHECK (utc_end_weekday BETWEEN 0 AND 6),
    utc_end_time INTEGER NOT NULL
  );

CREATE INDEX idx_stats_tracker_task ON stats_tracker_task (
  channel_id,
  utc_start_weekday,
  utc_end_weekday,
  utc_start_time,
  utc_end_time
);

ALTER TABLE channel
ADD COLUMN default_timezone TEXT NOT NULL DEFAULT 'UTC';