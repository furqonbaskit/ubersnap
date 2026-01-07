package models

import "time"

type Coupon struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"uuid"`
	Name            string    `gorm:"unique;not null"`
	Amount          int8      `json:"amount"`
	RemainingAmount int8      `json:"remaining_amount" gorm:"default:0"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy       string    `json:"created_by"`
	UpdatedBy       string    `json:"updated_by"`
}

func (Coupon) TableName() string {
	return "coupons"
}

type CouponRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int8   `json:"amount" binding:"required,gte=1"`
}

type CouponResponse struct {
	Name            string   `json:"name"`
	Amount          int8     `json:"amount"`
	RemainingAmount int8     `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}

func (c *Coupon) ToResponse() *CouponResponse {
	return &CouponResponse{
		Name:            c.Name,
		Amount:          c.Amount,
		RemainingAmount: c.RemainingAmount,
		ClaimedBy:       []string{},
	}
}
