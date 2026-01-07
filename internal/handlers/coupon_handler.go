package handlers

import (
	"net/http"
	"ubersnap/internal/models"
	"ubersnap/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CouponHandler struct {
	service services.CouponService
}

func NewCouponHandler(service services.CouponService) *CouponHandler {
	return &CouponHandler{service: service}
}

func (h *CouponHandler) CreateCouponHandler(c *gin.Context) {
	var req models.CouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, err := h.service.CreateCoupon(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

func (h *CouponHandler) GetCouponByNameHandler(c *gin.Context) {
	name := c.Param("name")
	coupon, err := h.service.GetCouponByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	c.JSON(http.StatusOK, coupon)
}

func (h *CouponHandler) ClaimCouponHandler(c *gin.Context) {
	var req models.ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	claimResp, err := h.service.ClaimCoupon(&req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No Stock"})
			return
		}
		if err == gorm.ErrRegistered {
			c.JSON(http.StatusConflict, gin.H{"error": "Already Claimed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to claim coupon"})
		return
	}

	c.JSON(http.StatusOK, claimResp)
}
