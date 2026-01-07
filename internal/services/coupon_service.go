package services

import (
	"ubersnap/internal/models"
	"ubersnap/internal/repositories"
)

type couponService struct {
	repo repositories.CouponRepository
}

type CouponService interface {
	CreateCoupon(req *models.CouponRequest) (*models.CouponResponse, error)
	GetCouponByName(name string) (*models.CouponResponse, error)
	ClaimCoupon(req *models.ClaimRequest) (*models.ClaimResponse, error)
}

func NewCouponService(repo *repositories.CouponRepository) CouponService {
	return &couponService{repo: *repo}
}

func (s *couponService) CreateCoupon(req *models.CouponRequest) (*models.CouponResponse, error) {
	couponPayload := &models.Coupon{
		Name:   req.Name,
		Amount: req.Amount,
	}
	if err := s.repo.CreateCoupon(couponPayload); err != nil {
		return nil, err
	}
	coupon := &models.CouponResponse{
		Name:   req.Name,
		Amount: req.Amount,
	}
	return coupon, nil
}

func (s *couponService) GetCouponByName(name string) (*models.CouponResponse, error) {
	coupon, err := s.repo.GetCouponByName(name)
	if err != nil {
		return nil, err
	}

	userClaims, err := s.repo.GetUserClaim(coupon.ID)
	if err != nil {
		return nil, err
	}
	couponResp := coupon.ToResponse()
	couponResp.ClaimedBy = userClaims
	return couponResp, nil
}

func (s *couponService) ClaimCoupon(req *models.ClaimRequest) (*models.ClaimResponse, error) {
	if err := s.repo.ClaimCoupon(req); err != nil {
		return nil, err
	}

	claimResp := &models.ClaimResponse{
		CouponName: req.CouponName,
		UserID:     req.UserID,
	}
	return claimResp, nil
}
