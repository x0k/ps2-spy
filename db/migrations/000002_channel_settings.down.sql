CREATE TABLE
  channel_locale (
    channel_id TEXT PRIMARY KEY NOT NULL,
    locale TEXT NOT NULL
  );

INSERT INTO
  channel_locale (channel_id, locale)
SELECT
  channel_id,
  locale
FROM
  channel;

DROP TABLE channel;
