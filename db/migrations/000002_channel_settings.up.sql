CREATE TABLE
  channel (
    channel_id TEXT PRIMARY KEY NOT NULL,
    locale TEXT NOT NULL DEFAULT 'en',
    character_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    outfit_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    title_updates BOOLEAN NOT NULL DEFAULT TRUE
  );

INSERT INTO channel (channel_id, locale)
SELECT channel_id, locale FROM channel_locale;

DROP TABLE channel_locale;
