package repositories

import "gorm.io/gorm"

type ClaimRepository struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) *ClaimRepository {
	return &ClaimRepository{db: db}
}

func (r *ClaimRepository) CreateClaim(claim interface{}) error {
	return r.db.Create(claim).Error
}

func (r *ClaimRepository) GetClaimByID(id string, out interface{}) error {
	return r.db.First(out, "id = ?", id).Error
}
