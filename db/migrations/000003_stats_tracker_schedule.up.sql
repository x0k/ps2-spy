CREATE TABLE
  stats_tracker_task (
    task_id INTEGER PRIMARY KEY NOT NULL,
    channel_id TEXT NOT NULL,
    weekday INTEGER NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    utc_start_time INTEGER NOT NULL,
    utc_end_time INTEGER NOT NULL,
    CHECK (utc_start_time < utc_end_time)
  );

CREATE INDEX idx_stats_tracker_task ON stats_tracker_task (channel_id, weekday, utc_start_time, utc_end_time);