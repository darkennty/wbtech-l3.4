package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"WBTech_L3.4/internal/model"
	"github.com/wb-go/wbf/dbpg"
)

var ErrNotFound = errors.New("image not found")

type ImageRepository struct {
	db *dbpg.DB
}

func NewImageRepository(db *dbpg.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, img *model.Image) error {
	query := `INSERT INTO images (id, status, original_path, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5)`
	now := time.Now()
	img.CreatedAt = now
	img.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, query, img.ID, img.Status, img.OriginalPath, img.CreatedAt, img.UpdatedAt)
	return err
}

func (r *ImageRepository) Update(ctx context.Context, img *model.Image) error {
	query := `UPDATE images SET status=$1, watermark_path=$2, thumb_path=$3, error=$4, updated_at=$5 WHERE id=$6`
	img.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, img.Status, img.WatermarkPath, img.ThumbPath, img.Error, img.UpdatedAt, img.ID)
	return err
}

func (r *ImageRepository) GetByID(ctx context.Context, id string) (*model.Image, error) {
	query := `SELECT id, status, original_path, watermark_path, thumb_path, error, created_at, updated_at FROM images WHERE id=$1`
	var img model.Image
	var watermarkPath sql.NullString
	var thumbPath sql.NullString
	var errMsg sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&img.ID, &img.Status, &img.OriginalPath, &watermarkPath, &thumbPath, &errMsg, &img.CreatedAt, &img.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if watermarkPath.Valid {
		img.WatermarkPath = watermarkPath.String
	}
	if thumbPath.Valid {
		img.ThumbPath = thumbPath.String
	}
	if errMsg.Valid {
		img.Error = errMsg.String
	}
	return &img, nil
}

func (r *ImageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM images WHERE id=$1", id)
	return err
}
