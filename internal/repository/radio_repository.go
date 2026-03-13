package repository

import (
	"context"

	"github.com/ifinu/radio-api/internal/models"
	"gorm.io/gorm"
)

type RadioRepository interface {
	Create(ctx context.Context, radio *models.Radio) error
	Update(ctx context.Context, radio *models.Radio) error
	FindByID(ctx context.Context, id uint) (*models.Radio, error)
	FindByUUID(ctx context.Context, uuid string) (*models.Radio, error)
	List(ctx context.Context, limit, offset int) ([]*models.Radio, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Radio, error)
	Count(ctx context.Context) (int64, error)
	UpsertBatch(ctx context.Context, radios []*models.Radio) error
}

type radioRepository struct {
	db *gorm.DB
}

func NewRadioRepository(db *gorm.DB) RadioRepository {
	return &radioRepository{db: db}
}

func (r *radioRepository) Create(ctx context.Context, radio *models.Radio) error {
	return r.db.WithContext(ctx).Create(radio).Error
}

func (r *radioRepository) Update(ctx context.Context, radio *models.Radio) error {
	return r.db.WithContext(ctx).Save(radio).Error
}

func (r *radioRepository) FindByID(ctx context.Context, id uint) (*models.Radio, error) {
	var radio models.Radio
	err := r.db.WithContext(ctx).First(&radio, id).Error
	if err != nil {
		return nil, err
	}
	return &radio, nil
}

func (r *radioRepository) FindByUUID(ctx context.Context, uuid string) (*models.Radio, error) {
	var radio models.Radio
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&radio).Error
	if err != nil {
		return nil, err
	}
	return &radio, nil
}

func (r *radioRepository) List(ctx context.Context, limit, offset int) ([]*models.Radio, error) {
	var radios []*models.Radio
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&radios).Error
	return radios, err
}

func (r *radioRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Radio, error) {
	var radios []*models.Radio
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR country ILIKE ? OR tags ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&radios).Error
	return radios, err
}

func (r *radioRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Radio{}).Count(&count).Error
	return count, err
}

func (r *radioRepository) UpsertBatch(ctx context.Context, radios []*models.Radio) error {
	if len(radios) == 0 {
		return nil
	}

	// Use ON CONFLICT to update existing records
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, radio := range radios {
			var existing models.Radio
			err := tx.Where("uuid = ?", radio.UUID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// Create new
				if err := tx.Create(radio).Error; err != nil {
					return err
				}
			} else if err == nil {
				// Update existing
				radio.ID = existing.ID
				radio.CreatedAt = existing.CreatedAt
				if err := tx.Save(radio).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}
