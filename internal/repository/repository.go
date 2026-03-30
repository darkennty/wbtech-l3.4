package repository

import (
	"context"

	"WBTech_L3.4/internal/model"
	"github.com/wb-go/wbf/dbpg"
)

type Image interface {
	Create(ctx context.Context, img *model.Image) error
	Update(ctx context.Context, img *model.Image) error
	GetByID(ctx context.Context, id string) (*model.Image, error)
	Delete(ctx context.Context, id string) error
}

type Repository struct {
	Image
}

func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{
		Image: NewImageRepository(db),
	}
}
