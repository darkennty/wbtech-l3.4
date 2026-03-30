package service

import (
	"context"
	"io"
	"path/filepath"

	"WBTech_L3.4/internal/kafka"
	"WBTech_L3.4/internal/model"
	"WBTech_L3.4/internal/repository"
	"WBTech_L3.4/internal/storage"
	"github.com/google/uuid"
)

type ImageService struct {
	repo     repository.Image
	storage  storage.Storage
	producer *kafka.Producer
}

func NewImageService(repo repository.Image, storage storage.Storage, producer *kafka.Producer) *ImageService {
	return &ImageService{
		repo:     repo,
		storage:  storage,
		producer: producer,
	}
}

func (s *ImageService) Upload(ctx context.Context, filename string, data io.Reader) (string, error) {
	id := uuid.New().String()
	originalPath, err := s.storage.Save(id+"_original"+filepath.Ext(filename), data)
	if err != nil {
		return "", err
	}

	img := &model.Image{
		ID:           id,
		Status:       "pending",
		OriginalPath: originalPath,
	}
	if err = s.repo.Create(ctx, img); err != nil {
		deleteErr := s.storage.Delete(originalPath)
		if deleteErr != nil {
			return "", deleteErr
		}
		return "", err
	}

	task := kafka.Task{
		ImageID:      img.ID,
		OriginalPath: img.OriginalPath,
	}

	if err = s.producer.Send(ctx, task); err != nil {
		img.Status = "failed"
		img.Error = "failed to send to queue"
		updateErr := s.repo.Update(ctx, img)
		if updateErr != nil {
			return "", updateErr
		}
		return "", err
	}

	return id, nil
}

func (s *ImageService) GetImage(ctx context.Context, id string, imgType string) (*model.Image, []byte, error) {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if img.Status != "completed" || imgType == "status_only" {
		return img, nil, nil
	}

	var filename string
	switch imgType {
	case "original":
		filename = filepath.Base(img.OriginalPath)
	case "thumb":
		filename = filepath.Base(img.ThumbPath)
	default:
		filename = filepath.Base(img.WatermarkPath)
	}

	data, err := s.storage.Load(filename)
	if err != nil {
		return nil, nil, err
	}
	return img, data, nil
}

func (s *ImageService) Delete(ctx context.Context, id string) error {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	originalBase := filepath.Base(img.OriginalPath)
	deleteErr := s.storage.Delete(originalBase)
	if deleteErr != nil {
		return deleteErr
	}

	if img.WatermarkPath != "" {
		deleteErr = s.storage.Delete(filepath.Base(img.WatermarkPath))
		if deleteErr != nil {
			return deleteErr
		}
	}

	if img.ThumbPath != "" {
		deleteErr = s.storage.Delete(filepath.Base(img.ThumbPath))
		if deleteErr != nil {
			return deleteErr
		}
	}

	return s.repo.Delete(ctx, id)
}
