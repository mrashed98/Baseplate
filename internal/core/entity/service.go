package entity

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/core/blueprint"
	"github.com/baseplate/baseplate/internal/core/validation"
)

var (
	ErrNotFound       = errors.New("entity not found")
	ErrAlreadyExists  = errors.New("entity already exists")
	ErrValidation     = errors.New("validation failed")
	ErrBlueprintNotFound = errors.New("blueprint not found")
)

type Service struct {
	repo            *Repository
	blueprintSvc    *blueprint.Service
	validator       *validation.Validator
}

func NewService(repo *Repository, blueprintSvc *blueprint.Service, validator *validation.Validator) *Service {
	return &Service{
		repo:         repo,
		blueprintSvc: blueprintSvc,
		validator:    validator,
	}
}

func (s *Service) Create(ctx context.Context, teamID uuid.UUID, blueprintID string, req *CreateEntityRequest) (*Entity, error) {
	// Get blueprint schema
	bp, err := s.blueprintSvc.Get(ctx, teamID, blueprintID)
	if err != nil {
		if errors.Is(err, blueprint.ErrNotFound) {
			return nil, ErrBlueprintNotFound
		}
		return nil, err
	}

	// Validate data against schema
	if err := s.validator.Validate(req.Data, bp.Schema); err != nil {
		return nil, err
	}

	// Check if entity already exists
	existing, err := s.repo.GetByIdentifier(ctx, teamID, blueprintID, req.Identifier)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	entity := &Entity{
		ID:          uuid.New(),
		TeamID:      teamID,
		BlueprintID: blueprintID,
		Identifier:  req.Identifier,
		Title:       req.Title,
		Data:        req.Data,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Entity, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrNotFound
	}
	return entity, nil
}

func (s *Service) GetByIdentifier(ctx context.Context, teamID uuid.UUID, blueprintID, identifier string) (*Entity, error) {
	entity, err := s.repo.GetByIdentifier(ctx, teamID, blueprintID, identifier)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrNotFound
	}
	return entity, nil
}

func (s *Service) List(ctx context.Context, teamID uuid.UUID, blueprintID string, limit, offset int) (*ListEntitiesResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entities, total, err := s.repo.List(ctx, teamID, blueprintID, limit, offset)
	if err != nil {
		return nil, err
	}

	if entities == nil {
		entities = []*Entity{}
	}

	return &ListEntitiesResponse{
		Entities: entities,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *Service) Search(ctx context.Context, teamID uuid.UUID, blueprintID string, req *SearchRequest) (*ListEntitiesResponse, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	entities, total, err := s.repo.Search(ctx, teamID, blueprintID, req)
	if err != nil {
		return nil, err
	}

	if entities == nil {
		entities = []*Entity{}
	}

	return &ListEntitiesResponse{
		Entities: entities,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdateEntityRequest) (*Entity, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrNotFound
	}

	// Get blueprint for validation
	bp, err := s.blueprintSvc.Get(ctx, entity.TeamID, entity.BlueprintID)
	if err != nil {
		return nil, err
	}

	// Merge and validate data
	if req.Data != nil {
		// Merge existing data with new data
		for k, v := range req.Data {
			entity.Data[k] = v
		}

		if err := s.validator.Validate(entity.Data, bp.Schema); err != nil {
			return nil, err
		}
	}

	if req.Title != "" {
		entity.Title = req.Title
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if entity == nil {
		return ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) DeleteByBlueprint(ctx context.Context, teamID uuid.UUID, blueprintID string) error {
	return s.repo.DeleteByBlueprint(ctx, teamID, blueprintID)
}
