package ps2_storage_facility_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) Facility(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, error) {
	facility, err := r.storage.Queries().GetFacility(ctx, string(facilityId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ps2.Facility{}, fmt.Errorf("facility %s: %w", facilityId, shared.ErrNotFound)
		}
		return ps2.Facility{}, err
	}
	return ps2.Facility{
		Id:     ps2.FacilityId(facility.FacilityID),
		Name:   facility.FacilityName,
		Type:   facility.FacilityType,
		ZoneId: ps2.ZoneId(facility.ZoneID),
	}, nil
}

func (r *Repository) SaveFacility(ctx context.Context, facility ps2.Facility) error {
	return r.storage.Queries().InsertFacility(ctx, db.InsertFacilityParams{
		FacilityID:   string(facility.Id),
		FacilityName: facility.Name,
		FacilityType: facility.Type,
		ZoneID:       string(facility.ZoneId),
	})
}
