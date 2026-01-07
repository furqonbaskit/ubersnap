package models

import "time"

type Claim struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"uuid"`
	UserID    string    `gorm:"uniqueIndex:idx_user_coupon" json:"user_id"`
	CouponID  string    `gorm:"uniqueIndex:idx_user_coupon;type:uuid" json:"coupon_id"`
	Coupon    *Coupon   `gorm:"foreignKey:CouponID;references:ID" json:"coupon"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
}

func (Claim) TableName() string {
	return "claims"
}

type ClaimRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	CouponName string `json:"coupon_name" binding:"required"`
}

type ClaimResponse struct {
	UserID     string `json:"user_id"`
	CouponName string `json:"coupon_name"`
}

func (c *Claim) ToResponse() *ClaimResponse {
	return &ClaimResponse{
		UserID:     c.UserID,
		CouponName: c.Coupon.Name,
	}
}
