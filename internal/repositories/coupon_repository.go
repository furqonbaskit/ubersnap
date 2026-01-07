package repositories

import (
	"ubersnap/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) CreateCoupon(coupon *models.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *CouponRepository) GetCouponByName(name string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := r.db.Where("name = ?", name).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *CouponRepository) ClaimCoupon(claim *models.ClaimRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		coupon := &models.Coupon{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("name = ?", claim.CouponName).
			First(coupon).Error; err != nil {
			return err
		}

		if coupon.RemainingAmount <= 0 {
			return gorm.ErrRecordNotFound
		}

		var existingClaim models.Claim
		err := tx.Where("user_id = ? AND coupon_id = ?", claim.UserID, coupon.ID).
			First(&existingClaim).Error
		if err == nil {
			return gorm.ErrRegistered
		}

		claimRecord := &models.Claim{
			UserID:   claim.UserID,
			CouponID: coupon.ID,
		}
		if err := tx.Create(claimRecord).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Coupon{}).
			Where("id = ? AND remaining_amount > 0", coupon.ID).
			UpdateColumn("remaining_amount", gorm.Expr("remaining_amount - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *CouponRepository) GetUserClaim(couponId string) ([]string, error) {
	var claims []models.Claim
	if err := r.db.Where("coupon_id = ?", couponId).Find(&claims).Error; err != nil {
		return nil, err
	}

	userNames := make([]string, len(claims))
	for i, claim := range claims {
		userNames[i] = claim.UserID
	}

	return userNames, nil
}
