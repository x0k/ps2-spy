// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
	"strings"
	"time"
)

const deleteChannelCharacters = `-- name: DeleteChannelCharacters :exec
DELETE FROM channel_to_character
WHERE
  channel_id = ?
  AND platform = ?
  AND character_id IN (/*SLICE:character_ids*/?)
`

type DeleteChannelCharactersParams struct {
	ChannelID    string
	Platform     string
	CharacterIds []string
}

func (q *Queries) DeleteChannelCharacters(ctx context.Context, arg DeleteChannelCharactersParams) error {
	query := deleteChannelCharacters
	var queryParams []interface{}
	queryParams = append(queryParams, arg.ChannelID)
	queryParams = append(queryParams, arg.Platform)
	if len(arg.CharacterIds) > 0 {
		for _, v := range arg.CharacterIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:character_ids*/?", strings.Repeat(",?", len(arg.CharacterIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:character_ids*/?", "NULL", 1)
	}
	_, err := q.exec(ctx, nil, query, queryParams...)
	return err
}

const deleteChannelOutfits = `-- name: DeleteChannelOutfits :exec
DELETE FROM channel_to_outfit
WHERE
  channel_id = ?
  AND platform = ?
  AND outfit_id IN (/*SLICE:outfit_ids*/?)
`

type DeleteChannelOutfitsParams struct {
	ChannelID string
	Platform  string
	OutfitIds []string
}

func (q *Queries) DeleteChannelOutfits(ctx context.Context, arg DeleteChannelOutfitsParams) error {
	query := deleteChannelOutfits
	var queryParams []interface{}
	queryParams = append(queryParams, arg.ChannelID)
	queryParams = append(queryParams, arg.Platform)
	if len(arg.OutfitIds) > 0 {
		for _, v := range arg.OutfitIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:outfit_ids*/?", strings.Repeat(",?", len(arg.OutfitIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:outfit_ids*/?", "NULL", 1)
	}
	_, err := q.exec(ctx, nil, query, queryParams...)
	return err
}

const deleteOutfitMembers = `-- name: DeleteOutfitMembers :exec
DELETE FROM outfit_to_character
WHERE
  platform = ?
  AND outfit_id = ?
  AND character_id IN (/*SLICE:character_ids*/?)
`

type DeleteOutfitMembersParams struct {
	Platform     string
	OutfitID     string
	CharacterIds []string
}

func (q *Queries) DeleteOutfitMembers(ctx context.Context, arg DeleteOutfitMembersParams) error {
	query := deleteOutfitMembers
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Platform)
	queryParams = append(queryParams, arg.OutfitID)
	if len(arg.CharacterIds) > 0 {
		for _, v := range arg.CharacterIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:character_ids*/?", strings.Repeat(",?", len(arg.CharacterIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:character_ids*/?", "NULL", 1)
	}
	_, err := q.exec(ctx, nil, query, queryParams...)
	return err
}

const getChannel = `-- name: GetChannel :one
SELECT
  channel_id, locale, character_notifications, outfit_notifications, title_updates, default_timezone
FROM
  channel
WHERE
  channel_id = ?
`

func (q *Queries) GetChannel(ctx context.Context, channelID string) (Channel, error) {
	row := q.queryRow(ctx, q.getChannelStmt, getChannel, channelID)
	var i Channel
	err := row.Scan(
		&i.ChannelID,
		&i.Locale,
		&i.CharacterNotifications,
		&i.OutfitNotifications,
		&i.TitleUpdates,
		&i.DefaultTimezone,
	)
	return i, err
}

const getCountChannelStatsTrackerTasks = `-- name: GetCountChannelStatsTrackerTasks :one
SELECT
  COUNT(*)
FROM
  stats_tracker_task
WHERE
  channel_id = ?
`

func (q *Queries) GetCountChannelStatsTrackerTasks(ctx context.Context, channelID string) (int64, error) {
	row := q.queryRow(ctx, q.getCountChannelStatsTrackerTasksStmt, getCountChannelStatsTrackerTasks, channelID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getFacility = `-- name: GetFacility :one
SELECT
  facility_id, facility_name, facility_type, zone_id
FROM
  facility
WHERE
  facility_id = ?
`

func (q *Queries) GetFacility(ctx context.Context, facilityID string) (Facility, error) {
	row := q.queryRow(ctx, q.getFacilityStmt, getFacility, facilityID)
	var i Facility
	err := row.Scan(
		&i.FacilityID,
		&i.FacilityName,
		&i.FacilityType,
		&i.ZoneID,
	)
	return i, err
}

const getPlatformOutfit = `-- name: GetPlatformOutfit :one
SELECT
  platform, outfit_id, outfit_name, outfit_tag
FROM
  outfit
WHERE
  platform = ?
  AND outfit_id = ?
`

type GetPlatformOutfitParams struct {
	Platform string
	OutfitID string
}

func (q *Queries) GetPlatformOutfit(ctx context.Context, arg GetPlatformOutfitParams) (Outfit, error) {
	row := q.queryRow(ctx, q.getPlatformOutfitStmt, getPlatformOutfit, arg.Platform, arg.OutfitID)
	var i Outfit
	err := row.Scan(
		&i.Platform,
		&i.OutfitID,
		&i.OutfitName,
		&i.OutfitTag,
	)
	return i, err
}

const getPlatformOutfitSynchronizedAt = `-- name: GetPlatformOutfitSynchronizedAt :one
SELECT
  synchronized_at
FROM
  outfit_synchronization
WHERE
  platform = ?
  AND outfit_id = ?
`

type GetPlatformOutfitSynchronizedAtParams struct {
	Platform string
	OutfitID string
}

func (q *Queries) GetPlatformOutfitSynchronizedAt(ctx context.Context, arg GetPlatformOutfitSynchronizedAtParams) (time.Time, error) {
	row := q.queryRow(ctx, q.getPlatformOutfitSynchronizedAtStmt, getPlatformOutfitSynchronizedAt, arg.Platform, arg.OutfitID)
	var synchronized_at time.Time
	err := row.Scan(&synchronized_at)
	return synchronized_at, err
}

const getStatsTrackerTask = `-- name: GetStatsTrackerTask :one
SELECT
  task_id, channel_id, utc_start_weekday, utc_start_time, utc_end_weekday, utc_end_time
FROM
  stats_tracker_task
WHERE
  task_id = ?
`

func (q *Queries) GetStatsTrackerTask(ctx context.Context, taskID int64) (StatsTrackerTask, error) {
	row := q.queryRow(ctx, q.getStatsTrackerTaskStmt, getStatsTrackerTask, taskID)
	var i StatsTrackerTask
	err := row.Scan(
		&i.TaskID,
		&i.ChannelID,
		&i.UtcStartWeekday,
		&i.UtcStartTime,
		&i.UtcEndWeekday,
		&i.UtcEndTime,
	)
	return i, err
}

const insertChannelCharacter = `-- name: InsertChannelCharacter :exec
INSERT INTO
  channel_to_character (channel_id, platform, character_id)
VALUES
  (?, ?, ?)
`

type InsertChannelCharacterParams struct {
	ChannelID   string
	Platform    string
	CharacterID string
}

func (q *Queries) InsertChannelCharacter(ctx context.Context, arg InsertChannelCharacterParams) error {
	_, err := q.exec(ctx, q.insertChannelCharacterStmt, insertChannelCharacter, arg.ChannelID, arg.Platform, arg.CharacterID)
	return err
}

const insertChannelOutfit = `-- name: InsertChannelOutfit :exec
INSERT INTO
  channel_to_outfit (channel_id, platform, outfit_id)
VALUES
  (?, ?, ?)
`

type InsertChannelOutfitParams struct {
	ChannelID string
	Platform  string
	OutfitID  string
}

func (q *Queries) InsertChannelOutfit(ctx context.Context, arg InsertChannelOutfitParams) error {
	_, err := q.exec(ctx, q.insertChannelOutfitStmt, insertChannelOutfit, arg.ChannelID, arg.Platform, arg.OutfitID)
	return err
}

const insertChannelStatsTrackerTask = `-- name: InsertChannelStatsTrackerTask :exec
INSERT INTO
  stats_tracker_task (
    channel_id,
    utc_start_weekday,
    utc_start_time,
    utc_end_weekday,
    utc_end_time
  )
VALUES
  (?, ?, ?, ?, ?)
`

type InsertChannelStatsTrackerTaskParams struct {
	ChannelID       string
	UtcStartWeekday int64
	UtcStartTime    int64
	UtcEndWeekday   int64
	UtcEndTime      int64
}

func (q *Queries) InsertChannelStatsTrackerTask(ctx context.Context, arg InsertChannelStatsTrackerTaskParams) error {
	_, err := q.exec(ctx, q.insertChannelStatsTrackerTaskStmt, insertChannelStatsTrackerTask,
		arg.ChannelID,
		arg.UtcStartWeekday,
		arg.UtcStartTime,
		arg.UtcEndWeekday,
		arg.UtcEndTime,
	)
	return err
}

const insertFacility = `-- name: InsertFacility :exec
INSERT INTO
  facility (
    facility_id,
    facility_name,
    facility_type,
    zone_id
  )
VALUES
  (?, ?, ?, ?)
`

type InsertFacilityParams struct {
	FacilityID   string
	FacilityName string
	FacilityType string
	ZoneID       string
}

func (q *Queries) InsertFacility(ctx context.Context, arg InsertFacilityParams) error {
	_, err := q.exec(ctx, q.insertFacilityStmt, insertFacility,
		arg.FacilityID,
		arg.FacilityName,
		arg.FacilityType,
		arg.ZoneID,
	)
	return err
}

const insertOutfit = `-- name: InsertOutfit :exec
INSERT INTO
  outfit (platform, outfit_id, outfit_name, outfit_tag)
VALUES
  (?, ?, ?, ?)
`

type InsertOutfitParams struct {
	Platform   string
	OutfitID   string
	OutfitName string
	OutfitTag  string
}

func (q *Queries) InsertOutfit(ctx context.Context, arg InsertOutfitParams) error {
	_, err := q.exec(ctx, q.insertOutfitStmt, insertOutfit,
		arg.Platform,
		arg.OutfitID,
		arg.OutfitName,
		arg.OutfitTag,
	)
	return err
}

const insertOutfitMember = `-- name: InsertOutfitMember :exec
INSERT INTO
  outfit_to_character (platform, outfit_id, character_id)
VALUES
  (?, ?, ?)
`

type InsertOutfitMemberParams struct {
	Platform    string
	OutfitID    string
	CharacterID string
}

func (q *Queries) InsertOutfitMember(ctx context.Context, arg InsertOutfitMemberParams) error {
	_, err := q.exec(ctx, q.insertOutfitMemberStmt, insertOutfitMember, arg.Platform, arg.OutfitID, arg.CharacterID)
	return err
}

const listActiveStatsTrackerTasks = `-- name: ListActiveStatsTrackerTasks :many
SELECT
  channel_id
FROM
  stats_tracker_task
WHERE
  (
    (
      ?1 = utc_start_weekday
      AND ?2 >= utc_start_time
    )
    OR (?1 > utc_start_weekday)
    OR (
      ?1 = 0
      AND utc_start_weekday = 6
    )
  )
  AND (
    (
      ?1 = utc_end_weekday
      AND ?2 < utc_end_time
    )
    OR (?1 < utc_end_weekday)
    OR (
      ?1 = 6
      AND utc_end_weekday = 0
    )
  )
`

type ListActiveStatsTrackerTasksParams struct {
	UtcWeekday int64
	UtcTime    int64
}

func (q *Queries) ListActiveStatsTrackerTasks(ctx context.Context, arg ListActiveStatsTrackerTasksParams) ([]string, error) {
	rows, err := q.query(ctx, q.listActiveStatsTrackerTasksStmt, listActiveStatsTrackerTasks, arg.UtcWeekday, arg.UtcTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var channel_id string
		if err := rows.Scan(&channel_id); err != nil {
			return nil, err
		}
		items = append(items, channel_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelCharacterIdsForPlatform = `-- name: ListChannelCharacterIdsForPlatform :many
SELECT
  character_id
FROM
  channel_to_character
WHERE
  channel_id = ?
  AND platform = ?
`

type ListChannelCharacterIdsForPlatformParams struct {
	ChannelID string
	Platform  string
}

func (q *Queries) ListChannelCharacterIdsForPlatform(ctx context.Context, arg ListChannelCharacterIdsForPlatformParams) ([]string, error) {
	rows, err := q.query(ctx, q.listChannelCharacterIdsForPlatformStmt, listChannelCharacterIdsForPlatform, arg.ChannelID, arg.Platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var character_id string
		if err := rows.Scan(&character_id); err != nil {
			return nil, err
		}
		items = append(items, character_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelIntersectingStatsTrackerTasks = `-- name: ListChannelIntersectingStatsTrackerTasks :many
SELECT
  task_id, channel_id, utc_start_weekday, utc_start_time, utc_end_weekday, utc_end_time
FROM
  stats_tracker_task
WHERE
  channel_id = ?
  AND (
    (?2 - utc_start_weekday IN (1, -6))
    OR (
      ?2 = utc_start_weekday
      AND ?3 > utc_start_time
    )
  )
  AND (
    (utc_end_weekday - ?4 IN (1, -6))
    OR (
      ?4 = utc_end_weekday
      AND ?5 < utc_end_time
    )
  )
`

type ListChannelIntersectingStatsTrackerTasksParams struct {
	ChannelID    string
	EndWeekday   int64
	EndTime      int64
	StartWeekday int64
	StartTime    int64
}

func (q *Queries) ListChannelIntersectingStatsTrackerTasks(ctx context.Context, arg ListChannelIntersectingStatsTrackerTasksParams) ([]StatsTrackerTask, error) {
	rows, err := q.query(ctx, q.listChannelIntersectingStatsTrackerTasksStmt, listChannelIntersectingStatsTrackerTasks,
		arg.ChannelID,
		arg.EndWeekday,
		arg.EndTime,
		arg.StartWeekday,
		arg.StartTime,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StatsTrackerTask
	for rows.Next() {
		var i StatsTrackerTask
		if err := rows.Scan(
			&i.TaskID,
			&i.ChannelID,
			&i.UtcStartWeekday,
			&i.UtcStartTime,
			&i.UtcEndWeekday,
			&i.UtcEndTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelOutfitIdsForPlatform = `-- name: ListChannelOutfitIdsForPlatform :many
SELECT
  outfit_id
FROM
  channel_to_outfit
WHERE
  channel_id = ?
  AND platform = ?
`

type ListChannelOutfitIdsForPlatformParams struct {
	ChannelID string
	Platform  string
}

func (q *Queries) ListChannelOutfitIdsForPlatform(ctx context.Context, arg ListChannelOutfitIdsForPlatformParams) ([]string, error) {
	rows, err := q.query(ctx, q.listChannelOutfitIdsForPlatformStmt, listChannelOutfitIdsForPlatform, arg.ChannelID, arg.Platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var outfit_id string
		if err := rows.Scan(&outfit_id); err != nil {
			return nil, err
		}
		items = append(items, outfit_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelStatsTrackerTasks = `-- name: ListChannelStatsTrackerTasks :many
SELECT
  task_id, channel_id, utc_start_weekday, utc_start_time, utc_end_weekday, utc_end_time
FROM
  stats_tracker_task
WHERE
  channel_id = ?
`

func (q *Queries) ListChannelStatsTrackerTasks(ctx context.Context, channelID string) ([]StatsTrackerTask, error) {
	rows, err := q.query(ctx, q.listChannelStatsTrackerTasksStmt, listChannelStatsTrackerTasks, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StatsTrackerTask
	for rows.Next() {
		var i StatsTrackerTask
		if err := rows.Scan(
			&i.TaskID,
			&i.ChannelID,
			&i.UtcStartWeekday,
			&i.UtcStartTime,
			&i.UtcEndWeekday,
			&i.UtcEndTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelTrackablePlatforms = `-- name: ListChannelTrackablePlatforms :many
SELECT DISTINCT
  platform
FROM
  channel_to_character
WHERE
  channel_to_character.channel_id = ?1
UNION
SELECT DISTINCT
  platform
FROM
  channel_to_outfit
WHERE
  channel_to_outfit.channel_id = ?1
`

func (q *Queries) ListChannelTrackablePlatforms(ctx context.Context, channelID string) ([]string, error) {
	rows, err := q.query(ctx, q.listChannelTrackablePlatformsStmt, listChannelTrackablePlatforms, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var platform string
		if err := rows.Scan(&platform); err != nil {
			return nil, err
		}
		items = append(items, platform)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPlatformOutfitMembers = `-- name: ListPlatformOutfitMembers :many
SELECT
  character_id
FROM
  outfit_to_character
WHERE
  platform = ?
  AND outfit_id = ?
`

type ListPlatformOutfitMembersParams struct {
	Platform string
	OutfitID string
}

func (q *Queries) ListPlatformOutfitMembers(ctx context.Context, arg ListPlatformOutfitMembersParams) ([]string, error) {
	rows, err := q.query(ctx, q.listPlatformOutfitMembersStmt, listPlatformOutfitMembers, arg.Platform, arg.OutfitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var character_id string
		if err := rows.Scan(&character_id); err != nil {
			return nil, err
		}
		items = append(items, character_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPlatformOutfits = `-- name: ListPlatformOutfits :many
SELECT
  platform, outfit_id, outfit_name, outfit_tag
FROM
  outfit
WHERE
  platform = ?
  AND outfit_id IN (/*SLICE:outfit_ids*/?)
`

type ListPlatformOutfitsParams struct {
	Platform  string
	OutfitIds []string
}

func (q *Queries) ListPlatformOutfits(ctx context.Context, arg ListPlatformOutfitsParams) ([]Outfit, error) {
	query := listPlatformOutfits
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Platform)
	if len(arg.OutfitIds) > 0 {
		for _, v := range arg.OutfitIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:outfit_ids*/?", strings.Repeat(",?", len(arg.OutfitIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:outfit_ids*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Outfit
	for rows.Next() {
		var i Outfit
		if err := rows.Scan(
			&i.Platform,
			&i.OutfitID,
			&i.OutfitName,
			&i.OutfitTag,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPlatformTrackingChannelsForCharacter = `-- name: ListPlatformTrackingChannelsForCharacter :many
SELECT
  channel_id, locale, character_notifications, outfit_notifications, title_updates, default_timezone
FROM
  channel
WHERE
  channel.channel_id IN (
    SELECT
      channel_id
    FROM
      channel_to_character
    WHERE
      channel_to_character.platform = ?1
      AND character_id = ?2
    UNION
    SELECT
      channel_id
    FROM
      channel_to_outfit
    WHERE
      channel_to_outfit.platform = ?1
      AND outfit_id = ?3
  )
`

type ListPlatformTrackingChannelsForCharacterParams struct {
	Platform    string
	CharacterID string
	OutfitID    string
}

func (q *Queries) ListPlatformTrackingChannelsForCharacter(ctx context.Context, arg ListPlatformTrackingChannelsForCharacterParams) ([]Channel, error) {
	rows, err := q.query(ctx, q.listPlatformTrackingChannelsForCharacterStmt, listPlatformTrackingChannelsForCharacter, arg.Platform, arg.CharacterID, arg.OutfitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Channel
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ChannelID,
			&i.Locale,
			&i.CharacterNotifications,
			&i.OutfitNotifications,
			&i.TitleUpdates,
			&i.DefaultTimezone,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPlatformTrackingChannelsForOutfit = `-- name: ListPlatformTrackingChannelsForOutfit :many
SELECT
  channel_id, locale, character_notifications, outfit_notifications, title_updates, default_timezone
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
  )
`

type ListPlatformTrackingChannelsForOutfitParams struct {
	Platform string
	OutfitID string
}

func (q *Queries) ListPlatformTrackingChannelsForOutfit(ctx context.Context, arg ListPlatformTrackingChannelsForOutfitParams) ([]Channel, error) {
	rows, err := q.query(ctx, q.listPlatformTrackingChannelsForOutfitStmt, listPlatformTrackingChannelsForOutfit, arg.Platform, arg.OutfitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Channel
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ChannelID,
			&i.Locale,
			&i.CharacterNotifications,
			&i.OutfitNotifications,
			&i.TitleUpdates,
			&i.DefaultTimezone,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTrackableCharacterIdsWithDuplicationForPlatform = `-- name: ListTrackableCharacterIdsWithDuplicationForPlatform :many
SELECT
  character_id
FROM
  channel_to_character
WHERE
  channel_to_character.platform = ?1
UNION ALL
SELECT
  character_id
FROM
  channel_to_outfit
  JOIN outfit_to_character ON channel_to_outfit.outfit_id = outfit_to_character.outfit_id
  AND channel_to_outfit.platform = outfit_to_character.platform
WHERE
  channel_to_outfit.platform = ?1
`

func (q *Queries) ListTrackableCharacterIdsWithDuplicationForPlatform(ctx context.Context, platform string) ([]string, error) {
	rows, err := q.query(ctx, q.listTrackableCharacterIdsWithDuplicationForPlatformStmt, listTrackableCharacterIdsWithDuplicationForPlatform, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var character_id string
		if err := rows.Scan(&character_id); err != nil {
			return nil, err
		}
		items = append(items, character_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTrackableOutfitIdsWithDuplicationForPlatform = `-- name: ListTrackableOutfitIdsWithDuplicationForPlatform :many
SELECT
  outfit_id
FROM
  channel_to_outfit
WHERE
  platform = ?
`

func (q *Queries) ListTrackableOutfitIdsWithDuplicationForPlatform(ctx context.Context, platform string) ([]string, error) {
	rows, err := q.query(ctx, q.listTrackableOutfitIdsWithDuplicationForPlatformStmt, listTrackableOutfitIdsWithDuplicationForPlatform, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var outfit_id string
		if err := rows.Scan(&outfit_id); err != nil {
			return nil, err
		}
		items = append(items, outfit_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUniqueTrackableOutfitIdsForPlatform = `-- name: ListUniqueTrackableOutfitIdsForPlatform :many
SELECT DISTINCT
  outfit_id
FROM
  channel_to_outfit
WHERE
  platform = ?
`

func (q *Queries) ListUniqueTrackableOutfitIdsForPlatform(ctx context.Context, platform string) ([]string, error) {
	rows, err := q.query(ctx, q.listUniqueTrackableOutfitIdsForPlatformStmt, listUniqueTrackableOutfitIdsForPlatform, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var outfit_id string
		if err := rows.Scan(&outfit_id); err != nil {
			return nil, err
		}
		items = append(items, outfit_id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeChannelStatsTrackerTask = `-- name: RemoveChannelStatsTrackerTask :exec
DELETE FROM stats_tracker_task
WHERE
  task_id = ?
  AND channel_id = ?
`

type RemoveChannelStatsTrackerTaskParams struct {
	TaskID    int64
	ChannelID string
}

func (q *Queries) RemoveChannelStatsTrackerTask(ctx context.Context, arg RemoveChannelStatsTrackerTaskParams) error {
	_, err := q.exec(ctx, q.removeChannelStatsTrackerTaskStmt, removeChannelStatsTrackerTask, arg.TaskID, arg.ChannelID)
	return err
}

const upsertChannelCharacterNotifications = `-- name: UpsertChannelCharacterNotifications :exec
INSERT INTO
  channel (channel_id, character_notifications)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  character_notifications = EXCLUDED.character_notifications
`

type UpsertChannelCharacterNotificationsParams struct {
	ChannelID              string
	CharacterNotifications bool
}

func (q *Queries) UpsertChannelCharacterNotifications(ctx context.Context, arg UpsertChannelCharacterNotificationsParams) error {
	_, err := q.exec(ctx, q.upsertChannelCharacterNotificationsStmt, upsertChannelCharacterNotifications, arg.ChannelID, arg.CharacterNotifications)
	return err
}

const upsertChannelDefaultTimezone = `-- name: UpsertChannelDefaultTimezone :exec
INSERT INTO
  channel (channel_id, default_timezone)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  default_timezone = EXCLUDED.default_timezone
`

type UpsertChannelDefaultTimezoneParams struct {
	ChannelID       string
	DefaultTimezone string
}

func (q *Queries) UpsertChannelDefaultTimezone(ctx context.Context, arg UpsertChannelDefaultTimezoneParams) error {
	_, err := q.exec(ctx, q.upsertChannelDefaultTimezoneStmt, upsertChannelDefaultTimezone, arg.ChannelID, arg.DefaultTimezone)
	return err
}

const upsertChannelLanguage = `-- name: UpsertChannelLanguage :exec
INSERT INTO
  channel (channel_id, locale)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  locale = EXCLUDED.locale
`

type UpsertChannelLanguageParams struct {
	ChannelID string
	Locale    string
}

func (q *Queries) UpsertChannelLanguage(ctx context.Context, arg UpsertChannelLanguageParams) error {
	_, err := q.exec(ctx, q.upsertChannelLanguageStmt, upsertChannelLanguage, arg.ChannelID, arg.Locale)
	return err
}

const upsertChannelOutfitNotifications = `-- name: UpsertChannelOutfitNotifications :exec
INSERT INTO
  channel (channel_id, outfit_notifications)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  outfit_notifications = EXCLUDED.outfit_notifications
`

type UpsertChannelOutfitNotificationsParams struct {
	ChannelID           string
	OutfitNotifications bool
}

func (q *Queries) UpsertChannelOutfitNotifications(ctx context.Context, arg UpsertChannelOutfitNotificationsParams) error {
	_, err := q.exec(ctx, q.upsertChannelOutfitNotificationsStmt, upsertChannelOutfitNotifications, arg.ChannelID, arg.OutfitNotifications)
	return err
}

const upsertChannelTitleUpdates = `-- name: UpsertChannelTitleUpdates :exec
INSERT INTO
  channel (channel_id, title_updates)
VALUES
  (?, ?) ON CONFLICT (channel_id) DO
UPDATE
SET
  title_updates = EXCLUDED.title_updates
`

type UpsertChannelTitleUpdatesParams struct {
	ChannelID    string
	TitleUpdates bool
}

func (q *Queries) UpsertChannelTitleUpdates(ctx context.Context, arg UpsertChannelTitleUpdatesParams) error {
	_, err := q.exec(ctx, q.upsertChannelTitleUpdatesStmt, upsertChannelTitleUpdates, arg.ChannelID, arg.TitleUpdates)
	return err
}

const upsertPlatformOutfitSynchronizedAt = `-- name: UpsertPlatformOutfitSynchronizedAt :exec
INSERT INTO
  outfit_synchronization (platform, outfit_id, synchronized_at)
VALUES
  (?, ?, ?) ON CONFLICT (platform, outfit_id) DO
UPDATE
SET
  synchronized_at = EXCLUDED.synchronized_at
`

type UpsertPlatformOutfitSynchronizedAtParams struct {
	Platform       string
	OutfitID       string
	SynchronizedAt time.Time
}

func (q *Queries) UpsertPlatformOutfitSynchronizedAt(ctx context.Context, arg UpsertPlatformOutfitSynchronizedAtParams) error {
	_, err := q.exec(ctx, q.upsertPlatformOutfitSynchronizedAtStmt, upsertPlatformOutfitSynchronizedAt, arg.Platform, arg.OutfitID, arg.SynchronizedAt)
	return err
}
