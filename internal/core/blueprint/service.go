package blueprint

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("blueprint not found")
	ErrAlreadyExists = errors.New("blueprint already exists")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, teamID uuid.UUID, req *CreateBlueprintRequest) (*Blueprint, error) {
	// Check if blueprint already exists
	exists, err := s.repo.Exists(ctx, teamID, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	bp := &Blueprint{
		ID:          req.ID,
		TeamID:      teamID,
		Title:       req.Title,
		Description: req.Description,
		Icon:        req.Icon,
		Schema:      req.Schema,
	}

	if err := s.repo.Create(ctx, bp); err != nil {
		return nil, err
	}

	return bp, nil
}

func (s *Service) Get(ctx context.Context, teamID uuid.UUID, id string) (*Blueprint, error) {
	bp, err := s.repo.GetByID(ctx, teamID, id)
	if err != nil {
		return nil, err
	}
	if bp == nil {
		return nil, ErrNotFound
	}
	return bp, nil
}

func (s *Service) List(ctx context.Context, teamID uuid.UUID) (*ListBlueprintsResponse, error) {
	blueprints, err := s.repo.List(ctx, teamID)
	if err != nil {
		return nil, err
	}

	if blueprints == nil {
		blueprints = []*Blueprint{}
	}

	return &ListBlueprintsResponse{
		Blueprints: blueprints,
		Total:      len(blueprints),
	}, nil
}

func (s *Service) Update(ctx context.Context, teamID uuid.UUID, id string, req *UpdateBlueprintRequest) (*Blueprint, error) {
	bp, err := s.repo.GetByID(ctx, teamID, id)
	if err != nil {
		return nil, err
	}
	if bp == nil {
		return nil, ErrNotFound
	}

	if req.Title != "" {
		bp.Title = req.Title
	}
	if req.Description != "" {
		bp.Description = req.Description
	}
	if req.Icon != "" {
		bp.Icon = req.Icon
	}
	if req.Schema != nil {
		bp.Schema = req.Schema
	}

	if err := s.repo.Update(ctx, bp); err != nil {
		return nil, err
	}

	return bp, nil
}

func (s *Service) Delete(ctx context.Context, teamID uuid.UUID, id string) error {
	exists, err := s.repo.Exists(ctx, teamID, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	return s.repo.Delete(ctx, teamID, id)
}

func (s *Service) GetSchema(ctx context.Context, teamID uuid.UUID, id string) (map[string]interface{}, error) {
	bp, err := s.Get(ctx, teamID, id)
	if err != nil {
		return nil, err
	}
	return bp.Schema, nil
}
