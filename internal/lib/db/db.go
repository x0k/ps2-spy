// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.deleteChannelCharacterStmt, err = db.PrepareContext(ctx, deleteChannelCharacter); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteChannelCharacter: %w", err)
	}
	if q.deleteChannelOutfitStmt, err = db.PrepareContext(ctx, deleteChannelOutfit); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteChannelOutfit: %w", err)
	}
	if q.deleteOutfitMemberStmt, err = db.PrepareContext(ctx, deleteOutfitMember); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteOutfitMember: %w", err)
	}
	if q.getChannelStmt, err = db.PrepareContext(ctx, getChannel); err != nil {
		return nil, fmt.Errorf("error preparing query GetChannel: %w", err)
	}
	if q.getFacilityStmt, err = db.PrepareContext(ctx, getFacility); err != nil {
		return nil, fmt.Errorf("error preparing query GetFacility: %w", err)
	}
	if q.getPlatformOutfitStmt, err = db.PrepareContext(ctx, getPlatformOutfit); err != nil {
		return nil, fmt.Errorf("error preparing query GetPlatformOutfit: %w", err)
	}
	if q.getPlatformOutfitSynchronizedAtStmt, err = db.PrepareContext(ctx, getPlatformOutfitSynchronizedAt); err != nil {
		return nil, fmt.Errorf("error preparing query GetPlatformOutfitSynchronizedAt: %w", err)
	}
	if q.insertChannelCharacterStmt, err = db.PrepareContext(ctx, insertChannelCharacter); err != nil {
		return nil, fmt.Errorf("error preparing query InsertChannelCharacter: %w", err)
	}
	if q.insertChannelOutfitStmt, err = db.PrepareContext(ctx, insertChannelOutfit); err != nil {
		return nil, fmt.Errorf("error preparing query InsertChannelOutfit: %w", err)
	}
	if q.insertChannelStatsTrackerTaskStmt, err = db.PrepareContext(ctx, insertChannelStatsTrackerTask); err != nil {
		return nil, fmt.Errorf("error preparing query InsertChannelStatsTrackerTask: %w", err)
	}
	if q.insertFacilityStmt, err = db.PrepareContext(ctx, insertFacility); err != nil {
		return nil, fmt.Errorf("error preparing query InsertFacility: %w", err)
	}
	if q.insertOutfitStmt, err = db.PrepareContext(ctx, insertOutfit); err != nil {
		return nil, fmt.Errorf("error preparing query InsertOutfit: %w", err)
	}
	if q.insertOutfitMemberStmt, err = db.PrepareContext(ctx, insertOutfitMember); err != nil {
		return nil, fmt.Errorf("error preparing query InsertOutfitMember: %w", err)
	}
	if q.listChannelCharacterIdsForPlatformStmt, err = db.PrepareContext(ctx, listChannelCharacterIdsForPlatform); err != nil {
		return nil, fmt.Errorf("error preparing query ListChannelCharacterIdsForPlatform: %w", err)
	}
	if q.listChannelOutfitIdsForPlatformStmt, err = db.PrepareContext(ctx, listChannelOutfitIdsForPlatform); err != nil {
		return nil, fmt.Errorf("error preparing query ListChannelOutfitIdsForPlatform: %w", err)
	}
	if q.listChannelOverlappingStatsTrackerTasksStmt, err = db.PrepareContext(ctx, listChannelOverlappingStatsTrackerTasks); err != nil {
		return nil, fmt.Errorf("error preparing query ListChannelOverlappingStatsTrackerTasks: %w", err)
	}
	if q.listChannelStatsTrackerTasksStmt, err = db.PrepareContext(ctx, listChannelStatsTrackerTasks); err != nil {
		return nil, fmt.Errorf("error preparing query ListChannelStatsTrackerTasks: %w", err)
	}
	if q.listChannelTrackablePlatformsStmt, err = db.PrepareContext(ctx, listChannelTrackablePlatforms); err != nil {
		return nil, fmt.Errorf("error preparing query ListChannelTrackablePlatforms: %w", err)
	}
	if q.listPlatformOutfitMembersStmt, err = db.PrepareContext(ctx, listPlatformOutfitMembers); err != nil {
		return nil, fmt.Errorf("error preparing query ListPlatformOutfitMembers: %w", err)
	}
	if q.listPlatformOutfitsStmt, err = db.PrepareContext(ctx, listPlatformOutfits); err != nil {
		return nil, fmt.Errorf("error preparing query ListPlatformOutfits: %w", err)
	}
	if q.listPlatformTrackingChannelsForCharacterStmt, err = db.PrepareContext(ctx, listPlatformTrackingChannelsForCharacter); err != nil {
		return nil, fmt.Errorf("error preparing query ListPlatformTrackingChannelsForCharacter: %w", err)
	}
	if q.listPlatformTrackingChannelsForOutfitStmt, err = db.PrepareContext(ctx, listPlatformTrackingChannelsForOutfit); err != nil {
		return nil, fmt.Errorf("error preparing query ListPlatformTrackingChannelsForOutfit: %w", err)
	}
	if q.listTrackableCharacterIdsWithDuplicationForPlatformStmt, err = db.PrepareContext(ctx, listTrackableCharacterIdsWithDuplicationForPlatform); err != nil {
		return nil, fmt.Errorf("error preparing query ListTrackableCharacterIdsWithDuplicationForPlatform: %w", err)
	}
	if q.listTrackableOutfitIdsWithDuplicationForPlatformStmt, err = db.PrepareContext(ctx, listTrackableOutfitIdsWithDuplicationForPlatform); err != nil {
		return nil, fmt.Errorf("error preparing query ListTrackableOutfitIdsWithDuplicationForPlatform: %w", err)
	}
	if q.listUniqueTrackableOutfitIdsForPlatformStmt, err = db.PrepareContext(ctx, listUniqueTrackableOutfitIdsForPlatform); err != nil {
		return nil, fmt.Errorf("error preparing query ListUniqueTrackableOutfitIdsForPlatform: %w", err)
	}
	if q.upsertChannelCharacterNotificationsStmt, err = db.PrepareContext(ctx, upsertChannelCharacterNotifications); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertChannelCharacterNotifications: %w", err)
	}
	if q.upsertChannelLanguageStmt, err = db.PrepareContext(ctx, upsertChannelLanguage); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertChannelLanguage: %w", err)
	}
	if q.upsertChannelOutfitNotificationsStmt, err = db.PrepareContext(ctx, upsertChannelOutfitNotifications); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertChannelOutfitNotifications: %w", err)
	}
	if q.upsertChannelTitleUpdatesStmt, err = db.PrepareContext(ctx, upsertChannelTitleUpdates); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertChannelTitleUpdates: %w", err)
	}
	if q.upsertPlatformOutfitSynchronizedAtStmt, err = db.PrepareContext(ctx, upsertPlatformOutfitSynchronizedAt); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertPlatformOutfitSynchronizedAt: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.deleteChannelCharacterStmt != nil {
		if cerr := q.deleteChannelCharacterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteChannelCharacterStmt: %w", cerr)
		}
	}
	if q.deleteChannelOutfitStmt != nil {
		if cerr := q.deleteChannelOutfitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteChannelOutfitStmt: %w", cerr)
		}
	}
	if q.deleteOutfitMemberStmt != nil {
		if cerr := q.deleteOutfitMemberStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteOutfitMemberStmt: %w", cerr)
		}
	}
	if q.getChannelStmt != nil {
		if cerr := q.getChannelStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getChannelStmt: %w", cerr)
		}
	}
	if q.getFacilityStmt != nil {
		if cerr := q.getFacilityStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getFacilityStmt: %w", cerr)
		}
	}
	if q.getPlatformOutfitStmt != nil {
		if cerr := q.getPlatformOutfitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPlatformOutfitStmt: %w", cerr)
		}
	}
	if q.getPlatformOutfitSynchronizedAtStmt != nil {
		if cerr := q.getPlatformOutfitSynchronizedAtStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPlatformOutfitSynchronizedAtStmt: %w", cerr)
		}
	}
	if q.insertChannelCharacterStmt != nil {
		if cerr := q.insertChannelCharacterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertChannelCharacterStmt: %w", cerr)
		}
	}
	if q.insertChannelOutfitStmt != nil {
		if cerr := q.insertChannelOutfitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertChannelOutfitStmt: %w", cerr)
		}
	}
	if q.insertChannelStatsTrackerTaskStmt != nil {
		if cerr := q.insertChannelStatsTrackerTaskStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertChannelStatsTrackerTaskStmt: %w", cerr)
		}
	}
	if q.insertFacilityStmt != nil {
		if cerr := q.insertFacilityStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertFacilityStmt: %w", cerr)
		}
	}
	if q.insertOutfitStmt != nil {
		if cerr := q.insertOutfitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertOutfitStmt: %w", cerr)
		}
	}
	if q.insertOutfitMemberStmt != nil {
		if cerr := q.insertOutfitMemberStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertOutfitMemberStmt: %w", cerr)
		}
	}
	if q.listChannelCharacterIdsForPlatformStmt != nil {
		if cerr := q.listChannelCharacterIdsForPlatformStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listChannelCharacterIdsForPlatformStmt: %w", cerr)
		}
	}
	if q.listChannelOutfitIdsForPlatformStmt != nil {
		if cerr := q.listChannelOutfitIdsForPlatformStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listChannelOutfitIdsForPlatformStmt: %w", cerr)
		}
	}
	if q.listChannelOverlappingStatsTrackerTasksStmt != nil {
		if cerr := q.listChannelOverlappingStatsTrackerTasksStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listChannelOverlappingStatsTrackerTasksStmt: %w", cerr)
		}
	}
	if q.listChannelStatsTrackerTasksStmt != nil {
		if cerr := q.listChannelStatsTrackerTasksStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listChannelStatsTrackerTasksStmt: %w", cerr)
		}
	}
	if q.listChannelTrackablePlatformsStmt != nil {
		if cerr := q.listChannelTrackablePlatformsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listChannelTrackablePlatformsStmt: %w", cerr)
		}
	}
	if q.listPlatformOutfitMembersStmt != nil {
		if cerr := q.listPlatformOutfitMembersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listPlatformOutfitMembersStmt: %w", cerr)
		}
	}
	if q.listPlatformOutfitsStmt != nil {
		if cerr := q.listPlatformOutfitsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listPlatformOutfitsStmt: %w", cerr)
		}
	}
	if q.listPlatformTrackingChannelsForCharacterStmt != nil {
		if cerr := q.listPlatformTrackingChannelsForCharacterStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listPlatformTrackingChannelsForCharacterStmt: %w", cerr)
		}
	}
	if q.listPlatformTrackingChannelsForOutfitStmt != nil {
		if cerr := q.listPlatformTrackingChannelsForOutfitStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listPlatformTrackingChannelsForOutfitStmt: %w", cerr)
		}
	}
	if q.listTrackableCharacterIdsWithDuplicationForPlatformStmt != nil {
		if cerr := q.listTrackableCharacterIdsWithDuplicationForPlatformStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listTrackableCharacterIdsWithDuplicationForPlatformStmt: %w", cerr)
		}
	}
	if q.listTrackableOutfitIdsWithDuplicationForPlatformStmt != nil {
		if cerr := q.listTrackableOutfitIdsWithDuplicationForPlatformStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listTrackableOutfitIdsWithDuplicationForPlatformStmt: %w", cerr)
		}
	}
	if q.listUniqueTrackableOutfitIdsForPlatformStmt != nil {
		if cerr := q.listUniqueTrackableOutfitIdsForPlatformStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing listUniqueTrackableOutfitIdsForPlatformStmt: %w", cerr)
		}
	}
	if q.upsertChannelCharacterNotificationsStmt != nil {
		if cerr := q.upsertChannelCharacterNotificationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertChannelCharacterNotificationsStmt: %w", cerr)
		}
	}
	if q.upsertChannelLanguageStmt != nil {
		if cerr := q.upsertChannelLanguageStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertChannelLanguageStmt: %w", cerr)
		}
	}
	if q.upsertChannelOutfitNotificationsStmt != nil {
		if cerr := q.upsertChannelOutfitNotificationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertChannelOutfitNotificationsStmt: %w", cerr)
		}
	}
	if q.upsertChannelTitleUpdatesStmt != nil {
		if cerr := q.upsertChannelTitleUpdatesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertChannelTitleUpdatesStmt: %w", cerr)
		}
	}
	if q.upsertPlatformOutfitSynchronizedAtStmt != nil {
		if cerr := q.upsertPlatformOutfitSynchronizedAtStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertPlatformOutfitSynchronizedAtStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                                                      DBTX
	tx                                                      *sql.Tx
	deleteChannelCharacterStmt                              *sql.Stmt
	deleteChannelOutfitStmt                                 *sql.Stmt
	deleteOutfitMemberStmt                                  *sql.Stmt
	getChannelStmt                                          *sql.Stmt
	getFacilityStmt                                         *sql.Stmt
	getPlatformOutfitStmt                                   *sql.Stmt
	getPlatformOutfitSynchronizedAtStmt                     *sql.Stmt
	insertChannelCharacterStmt                              *sql.Stmt
	insertChannelOutfitStmt                                 *sql.Stmt
	insertChannelStatsTrackerTaskStmt                       *sql.Stmt
	insertFacilityStmt                                      *sql.Stmt
	insertOutfitStmt                                        *sql.Stmt
	insertOutfitMemberStmt                                  *sql.Stmt
	listChannelCharacterIdsForPlatformStmt                  *sql.Stmt
	listChannelOutfitIdsForPlatformStmt                     *sql.Stmt
	listChannelOverlappingStatsTrackerTasksStmt             *sql.Stmt
	listChannelStatsTrackerTasksStmt                        *sql.Stmt
	listChannelTrackablePlatformsStmt                       *sql.Stmt
	listPlatformOutfitMembersStmt                           *sql.Stmt
	listPlatformOutfitsStmt                                 *sql.Stmt
	listPlatformTrackingChannelsForCharacterStmt            *sql.Stmt
	listPlatformTrackingChannelsForOutfitStmt               *sql.Stmt
	listTrackableCharacterIdsWithDuplicationForPlatformStmt *sql.Stmt
	listTrackableOutfitIdsWithDuplicationForPlatformStmt    *sql.Stmt
	listUniqueTrackableOutfitIdsForPlatformStmt             *sql.Stmt
	upsertChannelCharacterNotificationsStmt                 *sql.Stmt
	upsertChannelLanguageStmt                               *sql.Stmt
	upsertChannelOutfitNotificationsStmt                    *sql.Stmt
	upsertChannelTitleUpdatesStmt                           *sql.Stmt
	upsertPlatformOutfitSynchronizedAtStmt                  *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                                           tx,
		tx:                                           tx,
		deleteChannelCharacterStmt:                   q.deleteChannelCharacterStmt,
		deleteChannelOutfitStmt:                      q.deleteChannelOutfitStmt,
		deleteOutfitMemberStmt:                       q.deleteOutfitMemberStmt,
		getChannelStmt:                               q.getChannelStmt,
		getFacilityStmt:                              q.getFacilityStmt,
		getPlatformOutfitStmt:                        q.getPlatformOutfitStmt,
		getPlatformOutfitSynchronizedAtStmt:          q.getPlatformOutfitSynchronizedAtStmt,
		insertChannelCharacterStmt:                   q.insertChannelCharacterStmt,
		insertChannelOutfitStmt:                      q.insertChannelOutfitStmt,
		insertChannelStatsTrackerTaskStmt:            q.insertChannelStatsTrackerTaskStmt,
		insertFacilityStmt:                           q.insertFacilityStmt,
		insertOutfitStmt:                             q.insertOutfitStmt,
		insertOutfitMemberStmt:                       q.insertOutfitMemberStmt,
		listChannelCharacterIdsForPlatformStmt:       q.listChannelCharacterIdsForPlatformStmt,
		listChannelOutfitIdsForPlatformStmt:          q.listChannelOutfitIdsForPlatformStmt,
		listChannelOverlappingStatsTrackerTasksStmt:  q.listChannelOverlappingStatsTrackerTasksStmt,
		listChannelStatsTrackerTasksStmt:             q.listChannelStatsTrackerTasksStmt,
		listChannelTrackablePlatformsStmt:            q.listChannelTrackablePlatformsStmt,
		listPlatformOutfitMembersStmt:                q.listPlatformOutfitMembersStmt,
		listPlatformOutfitsStmt:                      q.listPlatformOutfitsStmt,
		listPlatformTrackingChannelsForCharacterStmt: q.listPlatformTrackingChannelsForCharacterStmt,
		listPlatformTrackingChannelsForOutfitStmt:    q.listPlatformTrackingChannelsForOutfitStmt,
		listTrackableCharacterIdsWithDuplicationForPlatformStmt: q.listTrackableCharacterIdsWithDuplicationForPlatformStmt,
		listTrackableOutfitIdsWithDuplicationForPlatformStmt:    q.listTrackableOutfitIdsWithDuplicationForPlatformStmt,
		listUniqueTrackableOutfitIdsForPlatformStmt:             q.listUniqueTrackableOutfitIdsForPlatformStmt,
		upsertChannelCharacterNotificationsStmt:                 q.upsertChannelCharacterNotificationsStmt,
		upsertChannelLanguageStmt:                               q.upsertChannelLanguageStmt,
		upsertChannelOutfitNotificationsStmt:                    q.upsertChannelOutfitNotificationsStmt,
		upsertChannelTitleUpdatesStmt:                           q.upsertChannelTitleUpdatesStmt,
		upsertPlatformOutfitSynchronizedAtStmt:                  q.upsertPlatformOutfitSynchronizedAtStmt,
	}
}
