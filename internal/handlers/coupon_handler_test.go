package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"ubersnap/internal/models"
	"ubersnap/internal/repositories"
	"ubersnap/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getTestDSN() string {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dbName := os.Getenv("DB_TEST_NAME")
	if dbName == "" {
		dbName = "ubersnap_test"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
}

func setupIntegrationTestDB() *gorm.DB {
	dsn := getTestDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Drop and recreate tables for clean state
	db.Migrator().DropTable(&models.Claim{})
	db.Migrator().DropTable(&models.Coupon{})

	err = db.AutoMigrate(&models.Coupon{}, &models.Claim{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate tables: %v", err))
	}

	return db
}

// Integration Test: Flash Sale Attack - 50 concurrent requests for 5 items
func TestFlashSaleAttackIntegration(t *testing.T) {
	db := setupIntegrationTestDB()
	repo := repositories.NewCouponRepository(db)
	svc := services.NewCouponService(repo)
	handler := NewCouponHandler(svc)

	// Create coupon with 5 items
	coupon := &models.Coupon{
		Name:            "FLASH5",
		Amount:          5,
		RemainingAmount: 5,
	}
	db.Create(coupon)

	// Simulate 50 concurrent requests
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			req := models.ClaimRequest{
				UserID:     fmt.Sprintf("user_%d", userID),
				CouponName: "FLASH5",
			}

			reqBody, _ := json.Marshal(req)
			httpReq, _ := http.NewRequest("POST", "/api/coupons/claim",
				bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq

			handler.ClaimCouponHandler(c)

			if w.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify results
	assert.Equal(t, 5, successCount, "Expected exactly 5 successful claims")

	var remaining models.Coupon
	db.Where("name = ?", "FLASH5").First(&remaining)
	assert.Equal(t, int8(0), remaining.RemainingAmount, "Expected 0 remaining coupons")

	var claimCount int64
	db.Model(&models.Claim{}).Where("coupon_id = ?", coupon.ID).Count(&claimCount)
	assert.Equal(t, int64(5), claimCount, "Expected exactly 5 claims in database")
}

// Integration Test: Double Dip Attack - 10 concurrent requests from SAME user
func TestDoubleDipAttackIntegration(t *testing.T) {
	db := setupIntegrationTestDB()
	repo := repositories.NewCouponRepository(db)
	svc := services.NewCouponService(repo)
	handler := NewCouponHandler(svc)

	// Create coupon
	coupon := &models.Coupon{
		Name:            "DOUBLEDIP",
		Amount:          100,
		RemainingAmount: 100,
	}
	db.Create(coupon)

	// Simulate 10 concurrent requests from SAME user
	var wg sync.WaitGroup
	successCount := 0
	conflictCount := 0
	otherCount := 0
	statusCodes := make(map[int]int)
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := models.ClaimRequest{
				UserID:     "user_same",
				CouponName: "DOUBLEDIP",
			}

			reqBody, _ := json.Marshal(req)
			httpReq, _ := http.NewRequest("POST", "/api/coupons/claim",
				bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httpReq

			handler.ClaimCouponHandler(c)

			mu.Lock()
			statusCodes[w.Code]++
			switch w.Code {
			case http.StatusOK:
				successCount++
			case http.StatusConflict:
				conflictCount++
			default:
				otherCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Log all status codes
	t.Logf("Status codes: %v", statusCodes)
	t.Logf("Success: %d, Conflict: %d, Other: %d", successCount, conflictCount, otherCount)

	// Verify results
	assert.Equal(t, 1, successCount, "Expected exactly 1 successful claim")
	assert.Equal(t, 9, conflictCount, "Expected exactly 9 conflict responses")

	var claimCount int64
	db.Model(&models.Claim{}).Where("coupon_id = ?", coupon.ID).Count(&claimCount)
	assert.Equal(t, int64(1), claimCount, "Expected exactly 1 claim in database for same user")
}

// Integration Test: Create Coupon
func TestCreateCouponIntegration(t *testing.T) {
	db := setupIntegrationTestDB()
	repo := repositories.NewCouponRepository(db)
	svc := services.NewCouponService(repo)
	handler := NewCouponHandler(svc)

	req := models.CouponRequest{
		Name:   "INTTEST",
		Amount: 20,
	}

	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/api/coupons",
		bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	handler.CreateCouponHandler(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.CouponResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "INTTEST", resp.Name)
	assert.Equal(t, int8(20), resp.Amount)
}

// Integration Test: Get Coupon by Name
func TestGetCouponByNameIntegration(t *testing.T) {
	db := setupIntegrationTestDB()
	repo := repositories.NewCouponRepository(db)
	svc := services.NewCouponService(repo)
	handler := NewCouponHandler(svc)

	// Create coupon first
	coupon := &models.Coupon{
		Name:            "GETTEST",
		Amount:          15,
		RemainingAmount: 15,
	}
	db.Create(coupon)

	httpReq, _ := http.NewRequest("GET", "/api/coupons/GETTEST", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = append(c.Params, gin.Param{Key: "name", Value: "GETTEST"})

	handler.GetCouponByNameHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CouponResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "GETTEST", resp.Name)
	assert.Equal(t, int8(15), resp.Amount)
}
