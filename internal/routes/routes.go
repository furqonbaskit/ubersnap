package routes

import (
	"ubersnap/internal/handlers"
	"ubersnap/internal/repositories"
	"ubersnap/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	couponRepo := repositories.NewCouponRepository(db)
	couponService := services.NewCouponService(couponRepo)
	couponHandler := handlers.NewCouponHandler(couponService)
	v1 := router.Group("/api")
	{
		v1.POST("/coupons", couponHandler.CreateCouponHandler)
		v1.POST("/coupons/claim", couponHandler.ClaimCouponHandler)
		v1.GET("/coupons/:name", couponHandler.GetCouponByNameHandler)

	}
}
