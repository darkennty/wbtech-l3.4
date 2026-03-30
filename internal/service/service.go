package service

import (
	"context"
	"io"

	"WBTech_L3.4/internal/kafka"
	"WBTech_L3.4/internal/model"
	"WBTech_L3.4/internal/repository"
	"WBTech_L3.4/internal/storage"
)

type Image interface {
	Upload(ctx context.Context, filename string, data io.Reader) (string, error)
	GetImage(ctx context.Context, id string, imgType string) (*model.Image, []byte, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	Image
}

func NewService(repo *repository.Repository, storage storage.Storage, producer *kafka.Producer) *Service {
	return &Service{
		Image: NewImageService(repo, storage, producer),
	}
}
