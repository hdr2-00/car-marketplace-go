package usecase

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"sync"
	"time"
)

// referenceSnapshot is an immutable copy of all reference data, indexed by ID.
type referenceSnapshot struct {
	data   domain.ReferenceData
	makes  []domain.CarMake
	models []domain.CarModel

	makeByID         map[int64]*domain.CarMake
	modelByID        map[int64]*domain.CarModel
	bodyTypeByID     map[int64]*domain.BodyType
	fuelTypeByID     map[int64]*domain.FuelType
	transmissionByID map[int64]*domain.TransmissionType
	colorByID        map[int64]*domain.Color
	conditionByID    map[int64]*domain.CarCondition
	featureByID      map[int64]*domain.Feature
}

// referenceCache keeps reference data in memory. It only changes through
// migrations, so it is reloaded after ttl rather than queried per request.
type referenceCache struct {
	repo repository.CarRepository
	ttl  time.Duration

	mu       sync.Mutex
	snapshot *referenceSnapshot
	loadedAt time.Time
}

func newReferenceCache(repo repository.CarRepository, ttl time.Duration) *referenceCache {
	return &referenceCache{repo: repo, ttl: ttl}
}

func (c *referenceCache) get(ctx context.Context) (*referenceSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.snapshot != nil && time.Since(c.loadedAt) < c.ttl {
		return c.snapshot, nil
	}

	snap, err := c.load(ctx)
	if err != nil {
		return nil, err
	}
	c.snapshot, c.loadedAt = snap, time.Now()
	return snap, nil
}

func (c *referenceCache) load(ctx context.Context) (*referenceSnapshot, error) {
	s := &referenceSnapshot{}
	var err error

	if s.makes, err = c.repo.GetAllMakes(ctx); err != nil {
		return nil, err
	}
	if s.models, err = c.repo.GetAllModels(ctx); err != nil {
		return nil, err
	}
	if s.data.BodyTypes, err = c.repo.GetAllBodyTypes(ctx); err != nil {
		return nil, err
	}
	if s.data.FuelTypes, err = c.repo.GetAllFuelTypes(ctx); err != nil {
		return nil, err
	}
	if s.data.Transmissions, err = c.repo.GetAllTransmissions(ctx); err != nil {
		return nil, err
	}
	if s.data.Colors, err = c.repo.GetAllColors(ctx); err != nil {
		return nil, err
	}
	if s.data.Conditions, err = c.repo.GetAllConditions(ctx); err != nil {
		return nil, err
	}
	if s.data.Features, err = c.repo.GetAllFeatures(ctx); err != nil {
		return nil, err
	}

	s.makeByID = indexByID(s.makes, func(m *domain.CarMake) int64 { return m.ID })
	s.modelByID = indexByID(s.models, func(m *domain.CarModel) int64 { return m.ID })
	s.bodyTypeByID = indexByID(s.data.BodyTypes, func(b *domain.BodyType) int64 { return b.ID })
	s.fuelTypeByID = indexByID(s.data.FuelTypes, func(f *domain.FuelType) int64 { return f.ID })
	s.transmissionByID = indexByID(s.data.Transmissions, func(t *domain.TransmissionType) int64 { return t.ID })
	s.colorByID = indexByID(s.data.Colors, func(c *domain.Color) int64 { return c.ID })
	s.conditionByID = indexByID(s.data.Conditions, func(c *domain.CarCondition) int64 { return c.ID })
	s.featureByID = indexByID(s.data.Features, func(f *domain.Feature) int64 { return f.ID })

	return s, nil
}

func indexByID[T any](items []T, id func(*T) int64) map[int64]*T {
	m := make(map[int64]*T, len(items))
	for i := range items {
		m[id(&items[i])] = &items[i]
	}
	return m
}

// validate checks that every reference a listing uses exists and that the
// model belongs to the make.
func (s *referenceSnapshot) validate(l *domain.CarListing, featureIDs []int64) error {
	if s.makeByID[l.MakeID] == nil {
		return domain.ErrInvalidMake
	}
	if model := s.modelByID[l.ModelID]; model == nil || model.MakeID != l.MakeID {
		return domain.ErrInvalidModel
	}
	if s.bodyTypeByID[l.BodyTypeID] == nil {
		return domain.ErrInvalidBodyType
	}
	if s.fuelTypeByID[l.FuelTypeID] == nil {
		return domain.ErrInvalidFuelType
	}
	if s.transmissionByID[l.TransmissionID] == nil {
		return domain.ErrInvalidTransmission
	}
	if s.colorByID[l.ColorID] == nil {
		return domain.ErrInvalidColor
	}
	if s.conditionByID[l.ConditionID] == nil {
		return domain.ErrInvalidCondition
	}
	for _, id := range featureIDs {
		if s.featureByID[id] == nil {
			return domain.ErrInvalidFeature
		}
	}
	return nil
}

// attach fills in a listing's reference entities.
func (s *referenceSnapshot) attach(l *domain.CarListing) {
	l.Make = s.makeByID[l.MakeID]
	l.Model = s.modelByID[l.ModelID]
	l.BodyType = s.bodyTypeByID[l.BodyTypeID]
	l.FuelType = s.fuelTypeByID[l.FuelTypeID]
	l.Transmission = s.transmissionByID[l.TransmissionID]
	l.Color = s.colorByID[l.ColorID]
	l.Condition = s.conditionByID[l.ConditionID]
}
