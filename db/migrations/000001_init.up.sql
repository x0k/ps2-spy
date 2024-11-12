CREATE TABLE
  IF NOT EXISTS outfit_to_character (
    platform TEXT NOT NULL,
    outfit_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    PRIMARY KEY (platform, outfit_id, character_id)
  );

CREATE TABLE
  IF NOT EXISTS outfit_synchronization (
    platform TEXT NOT NULL,
    outfit_id TEXT NOT NULL,
    synchronized_at TIMESTAMP NOT NULL,
    PRIMARY KEY (platform, outfit_id)
  );

CREATE TABLE
  IF NOT EXISTS channel_to_outfit (
    channel_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    outfit_id TEXT NOT NULL,
    PRIMARY KEY (channel_id, platform, outfit_id)
  );

CREATE TABLE
  IF NOT EXISTS channel_to_character (
    channel_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    character_id TEXT NOT NULL,
    PRIMARY KEY (channel_id, platform, character_id)
  );

CREATE TABLE
  IF NOT EXISTS outfit (
    platform TEXT NOT NULL,
    outfit_id TEXT NOT NULL,
    outfit_name TEXT NOT NULL,
    outfit_tag TEXT NOT NULL,
    PRIMARY KEY (platform, outfit_id)
  );

CREATE TABLE
  IF NOT EXISTS facility (
    facility_id TEXT PRIMARY KEY NOT NULL,
    facility_name TEXT NOT NULL,
    facility_type TEXT NOT NULL,
    zone_id TEXT NOT NULL
  );