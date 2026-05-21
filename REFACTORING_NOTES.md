# Refactoring Notes

This branch (`subdomains`) is extracting PS2, tracking, stats tracking, and Discord scheduling concerns out of the older storage/app-centered wiring. Use this as the working checklist for finishing the refactor and reviewing it before merge.

## Diff Summary

Compared with `main`, the branch changes 61 files with about 2,060 insertions and 1,630 deletions.

Main moves:

- `internal/characters_tracker` is now platform-aware at the top level. `Tracker` owns per-platform `platformTracker` instances instead of `app/root.go` owning a map of trackers.
- `internal/tracking.Manager` is now an aggregate over per-platform managers. Platform-specific filter logic moved into `internal/tracking/platform_manager.go`.
- Stats tracker task domain types moved from `internal/discord` into `internal/stats_tracker/model.go`.
- Stats tracker task creation and validation moved into `internal/stats_tracker/tasks_creator`.
- Stats tracker task persistence moved into `internal/stats_tracker/storage_tasks_repo`.
- Outfit member synchronization moved from `internal/outfit_members_synchronizer` into `internal/ps2/outfit_members_synchronizer`.
- Outfit member storage access moved into `internal/ps2/storage_outfits_repo`.
- Tracking settings storage repo moved from `internal/tracking/settings_repo` into `internal/tracking/storage_settings_repo`.
- Tracking read helpers moved into `internal/tracking/storage_tracking_repo`.
- PS2 outfit-member events moved out of `internal/storage` into `internal/ps2/events.go`.
- Shared time helpers are being relocated from `internal/shared/timezones.go` to `internal/lib/timex`.
- `internal/app/root.go` now wires fewer per-platform service maps directly and delegates more behavior to aggregate subdomain services.

## Target Shape

The intended end state looks like this:

- `internal/storage` remains a persistence and channel-settings infrastructure boundary.
- `internal/ps2` owns PS2 domain models, outfit-member events, and outfit-member synchronization.
- `internal/tracking` owns tracking settings, tracking filters, and channel lookup behavior.
- `internal/stats_tracker` owns stats task models, validation, scheduling, and runtime tracking.
- `internal/characters_tracker` owns online character state and population state across platforms.
- `internal/discord` owns Discord presentation, interaction routing, and Discord-specific form state only.
- `internal/app/root.go` wires concrete implementations, but no longer contains domain rules.

## Completed Follow-up

- Added focused tests around the new domain boundaries:
  - `internal/tracking/platform_manager.go`: filter rebuilds, duplicated trackers, outfit member add/remove deltas, and tracking settings updates.
  - `internal/characters_tracker/platform_tracker.go`: platform-specific login/logout event publication.
  - `internal/stats_tracker/tasks_creator/creator.go`: timezone conversion, max task count, update semantics, and overlap rejection.
  - `internal/ps2/outfit_members_synchronizer/outfit_members_synchronizer.go`: diff publication, no-op syncs, sync timestamp behavior, and transaction behavior.
- Added guard errors for unknown platform lookups in aggregate `tracking.Manager` and `characters_tracker.Tracker` read paths.
- Fixed no-op outfit member syncs so they still save the synchronized-at timestamp without publishing empty update events.
- Moved the outfit sync "synchronizing" log after successful trackable outfit loading.
- Renamed channel lookup helpers that return `[]discord.Channel`, not IDs.
- Renamed `ErrMaxTooManyTasksPerChannel` to `ErrTooManyTasksPerChannel`.
- Kept the stats schedule custom ID separator local to `internal/discord/messages` instead of exporting it from core `internal/discord`.
- Replaced `timex.NormalizeDate` loop normalization with modulo arithmetic.
- Removed the unused `internal/discord/storage_channels_settings_repo` scaffold.

## Remaining Work

- Review goroutine lifecycle semantics in aggregate `Start` methods. `tracking.Manager.Start` and `characters_tracker.Tracker.Start` wait on `ctx.Done()` after launching platform goroutines; this matches service-style usage, but tests should cover cancellation.
- Review event fan-out in `internal/modules/discord/module.go`. Character tracker events are now published through one shared pubsub and filtered per platform by event payload.
- Confirm whether first outfit-member syncs should remain silent for Discord. Current behavior preserves the old storage behavior: initial population does not produce `OutfitMembersUpdate`, while later diffs do.
- Remove or resolve stale TODOs introduced or exposed by the refactor:
  - `internal/characters_tracker/platform_tracker.go`: zone open state is still inferred from population.
  - `internal/discord/messages/messages.go`: existing "Fix this" TODO.

## Compatibility Checks

- No imports remain for deleted packages:
  - `internal/outfit_members_synchronizer`
  - `internal/ps2/characters_tracker_characters_repo`
  - `internal/ps2/characters_tracker_outfits_repo`
  - `internal/tracking/settings_repo`
  - `internal/storage/sql/outfit_members.go`
  - `internal/storage/sql/stats_tracker_tasks.go`
- `go test ./...` compiled and ran all packages, but the command ended with a sandbox filesystem error while trimming the Go build cache:
  - `go: failed to trim cache: open /home/roman/.cache/go-build/trim.txt: read-only file system`
  - Re-run outside the sandbox or with a writable `GOCACHE` to get a fully clean exit code.

## Suggested Validation Commands

Run these before merging:

```sh
GOCACHE=/tmp/ps2-spy-go-build go test ./...
go test -race ./internal/tracking ./internal/characters_tracker ./internal/stats_tracker ./internal/ps2/outfit_members_synchronizer
go test ./internal/discord/commands ./internal/discord/messages ./internal/modules/discord
```

If SQL query generation is part of the normal workflow, also regenerate and verify `internal/lib/db/queries.sql.go` after changes to `db/queries.sql`.

## Review Focus

When reviewing the final branch, prioritize behavior that crosses subdomain boundaries:

- Streaming events should carry the correct platform through `app -> characters_tracker -> discord/events`.
- Outfit member sync should update tracking filters and Discord notifications through PS2 events, not storage events.
- Stats tracker schedules should behave the same in Discord while using `stats_tracker.Task` and `CreateOrUpdateTask`.
- Tracking settings changes should still update character/outfit filters and trigger outfit member sync for newly tracked outfits.
- Storage should expose data through subdomain repositories without republishing domain events from low-level persistence methods.
