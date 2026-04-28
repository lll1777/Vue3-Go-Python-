package controllers

import (
	"net/http"
	"parking-system-go/models"
	"parking-system-go/services"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParkingController struct {
	parkingService *services.ParkingService
}

func NewParkingController(db *gorm.DB) *ParkingController {
	return &ParkingController{
		parkingService: services.NewParkingService(db),
	}
}

func (c *ParkingController) GetParkingLots(ctx *gin.Context) {
	lots, err := c.parkingService.GetAllParkingLots()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get parking lots",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lots,
	})
}

func (c *ParkingController) GetParkingLotByID(ctx *gin.Context) {
	id := ctx.Param("id")
	lot, err := c.parkingService.GetParkingLotByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Parking lot not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lot,
	})
}

func (c *ParkingController) GetParkingSpots(ctx *gin.Context) {
	lotID := ctx.Query("lot_id")
	
	var spots []models.ParkingSpot
	var err error
	
	if lotID != "" {
		spots, err = c.parkingService.GetParkingSpotsByLot(lotID)
	} else {
		spots, err = c.parkingService.GetAllParkingSpots()
	}
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get parking spots",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spots,
	})
}

func (c *ParkingController) GetParkingSpotByID(ctx *gin.Context) {
	id := ctx.Param("id")
	spot, err := c.parkingService.GetParkingSpotByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Parking spot not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spot,
	})
}

func (c *ParkingController) UpdateParkingSpotStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := c.parkingService.UpdateParkingSpotStatus(id, req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update parking spot status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking spot status updated",
	})
}

type ReservationController struct {
	reservationService *services.ReservationService
}

func NewReservationController(db *gorm.DB) *ReservationController {
	return &ReservationController{
		reservationService: services.NewReservationService(db),
	}
}

func (c *ReservationController) CreateReservation(ctx *gin.Context) {
	var req struct {
		ParkingSpotID string  `json:"parking_spot_id" binding:"required"`
		LicensePlate  string  `json:"license_plate" binding:"required"`
		Phone         string  `json:"phone"`
		StartTime     string  `json:"start_time" binding:"required"`
		EndTime       string  `json:"end_time" binding:"required"`
		Notes         string  `json:"notes"`
		TotalFee      float64 `json:"total_fee"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		startTime, _ = time.Parse("2006-01-02 15:04:05", req.StartTime)
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		endTime, _ = time.Parse("2006-01-02 15:04:05", req.EndTime)
	}

	reservation := &models.Reservation{
		ParkingSpotID: req.ParkingSpotID,
		LicensePlate:  req.LicensePlate,
		Phone:         req.Phone,
		StartTime:     startTime,
		EndTime:       endTime,
		Notes:         req.Notes,
		TotalFee:      req.TotalFee,
	}

	if err := c.reservationService.CreateReservation(reservation); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create reservation",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Reservation created successfully",
		"data":    reservation,
	})
}

func (c *ReservationController) GetReservations(ctx *gin.Context) {
	reservations, err := c.reservationService.GetAllReservations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get reservations",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reservations,
	})
}

func (c *ReservationController) GetReservationByID(ctx *gin.Context) {
	id := ctx.Param("id")
	reservation, err := c.reservationService.GetReservationByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Reservation not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reservation,
	})
}

func (c *ReservationController) CancelReservation(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.reservationService.CancelReservation(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to cancel reservation",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reservation cancelled successfully",
	})
}

func (c *ReservationController) CheckInReservation(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.reservationService.CheckInReservation(id); err != nil {
		if err == services.ErrReservationExpired {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Reservation has expired",
				"error":   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check in reservation",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reservation checked in successfully",
	})
}

func (c *ReservationController) CheckOutReservation(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.reservationService.CheckOutReservation(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check out reservation",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reservation checked out successfully",
	})
}

func (c *ReservationController) CleanupExpiredReservations(ctx *gin.Context) {
	count, err := c.reservationService.CleanupExpiredReservations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to cleanup expired reservations",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Expired reservations cleaned up",
		"data": gin.H{
			"cleaned_count": count,
		},
	})
}

func (c *ReservationController) GetActiveReservationByPlate(ctx *gin.Context) {
	licensePlate := ctx.Query("license_plate")
	if licensePlate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "License plate is required",
		})
		return
	}

	reservation, err := c.reservationService.GetActiveReservationByLicensePlate(licensePlate)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "No active reservation found",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reservation,
	})
}

type OrderController struct {
	orderService *services.OrderService
}

func NewOrderController(db *gorm.DB) *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(db),
	}
}

func (c *OrderController) GetOrders(ctx *gin.Context) {
	orders, err := c.orderService.GetAllOrders()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get orders",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}

func (c *OrderController) GetOrderByID(ctx *gin.Context) {
	id := ctx.Param("id")
	order, err := c.orderService.GetOrderByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Order not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var order models.Order

	if err := ctx.ShouldBindJSON(&order); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if err := c.orderService.CreateOrder(&order); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create order",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Order created successfully",
		"data":    order,
	})
}

func (c *OrderController) PayOrder(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		PaymentMethod string  `json:"payment_method" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := c.orderService.PayOrder(id, req.PaymentMethod, req.Amount); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to pay order",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment successful",
	})
}

type AccessControlController struct {
	accessService *services.AccessControlService
}

func NewAccessControlController(db *gorm.DB) *AccessControlController {
	return &AccessControlController{
		accessService: services.NewAccessControlService(db),
	}
}

func (c *AccessControlController) VehicleEntry(ctx *gin.Context) {
	var req struct {
		LicensePlate string  `json:"license_plate" binding:"required"`
		DeviceID      string  `json:"device_id"`
		Confidence    float64 `json:"confidence"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if req.Confidence == 0 {
		req.Confidence = 0.95
	}

	accessLog, err := c.accessService.VehicleEntry(req.LicensePlate, req.DeviceID, req.Confidence)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process vehicle entry",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Vehicle entry processed",
		"data":    accessLog,
	})
}

func (c *AccessControlController) VehicleExit(ctx *gin.Context) {
	var req struct {
		LicensePlate string  `json:"license_plate" binding:"required"`
		DeviceID      string  `json:"device_id"`
		Confidence    float64 `json:"confidence"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if req.Confidence == 0 {
		req.Confidence = 0.95
	}

	accessLog, err := c.accessService.VehicleExit(req.LicensePlate, req.DeviceID, req.Confidence)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process vehicle exit",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle exit processed",
		"data":    accessLog,
	})
}

func (c *AccessControlController) GetAccessLogs(ctx *gin.Context) {
	logs, err := c.accessService.GetAccessLogs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get access logs",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

type BillingController struct {
	billingService        *services.BillingService
	enhancedBillingService *services.EnhancedBillingService
}

func NewBillingController(db *gorm.DB) *BillingController {
	return &BillingController{
		billingService:        services.NewBillingService(db),
		enhancedBillingService: services.NewEnhancedBillingService(db),
	}
}

func (c *BillingController) CalculateFee(ctx *gin.Context) {
	var req struct {
		Minutes  int64  `json:"minutes" binding:"required"`
		SpotType string `json:"spot_type"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if req.SpotType == "" {
		req.SpotType = "standard"
	}

	fee, err := c.enhancedBillingService.CalculateFee(req.Minutes, req.SpotType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to calculate fee",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"minutes": req.Minutes,
			"fee":     fee,
		},
	})
}

func (c *BillingController) CalculateDetailedFee(ctx *gin.Context) {
	var req struct {
		EntryTime string `json:"entry_time" binding:"required"`
		ExitTime  string `json:"exit_time" binding:"required"`
		SpotType  string `json:"spot_type"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	entryTime, err := time.Parse(time.RFC3339, req.EntryTime)
	if err != nil {
		entryTime, err = time.Parse("2006-01-02 15:04:05", req.EntryTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid entry_time format",
			})
			return
		}
	}

	exitTime, err := time.Parse(time.RFC3339, req.ExitTime)
	if err != nil {
		exitTime, err = time.Parse("2006-01-02 15:04:05", req.ExitTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid exit_time format",
			})
			return
		}
	}

	if req.SpotType == "" {
		req.SpotType = "standard"
	}

	result, err := c.enhancedBillingService.CalculateParkingFee(entryTime, exitTime, req.SpotType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to calculate detailed fee",
			"error":   err.Error(),
		})
		return
	}

	dailyBillingsJSON := make([]gin.H, len(result.DailyBillings))
	for i, db := range result.DailyBillings {
		periodsJSON := make([]gin.H, len(db.Periods))
		for j, p := range db.Periods {
			periodsJSON[j] = gin.H{
				"start_time":     p.StartTime.Format(time.RFC3339),
				"end_time":       p.EndTime.Format(time.RFC3339),
				"duration":       p.Duration.String(),
				"duration_min":   p.DurationMin,
				"period_type":    p.PeriodType,
				"hourly_rate":    p.HourlyRate,
				"base_rate":      p.BaseRate,
				"is_first_hour":  p.IsFirstHour,
				"period_amount":  p.PeriodAmount,
			}
		}
		dailyBillingsJSON[i] = gin.H{
			"date":        db.Date,
			"day_of_week": db.DayOfWeek,
			"is_holiday":  db.IsHoliday,
			"periods":     periodsJSON,
			"sub_total":   db.SubTotal,
			"daily_max":   db.DailyMax,
			"discount":    db.Discount,
			"daily_total": db.DailyTotal,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"entry_time":           result.EntryTime.Format(time.RFC3339),
			"exit_time":            result.ExitTime.Format(time.RFC3339),
			"total_duration":       result.TotalDuration.String(),
			"total_duration_min":   result.TotalDurationMin,
			"within_grace_period":  result.WithinGracePeriod,
			"grace_period_minutes": result.GracePeriodMinutes,
			"daily_billings":       dailyBillingsJSON,
			"first_hour_used":      result.FirstHourUsed,
			"first_hour_applied":   result.FirstHourApplied,
			"total_before_rules":   result.TotalBeforeRules,
			"total_discount":       result.TotalDiscount,
			"min_charge_applied":   result.MinChargeApplied,
			"final_amount":         result.FinalAmount,
			"rule_summary":         result.RuleSummary,
		},
	})
}

func (c *BillingController) GetBillingRules(ctx *gin.Context) {
	rules, err := c.billingService.GetAllBillingRules()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get billing rules",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rules,
	})
}

func (c *BillingController) UpdateBillingRule(ctx *gin.Context) {
	id := ctx.Param("id")

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := c.billingService.UpdateBillingRule(id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update billing rule",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Billing rule updated",
	})
}

type DeviceController struct {
	deviceService *services.DeviceService
}

func NewDeviceController(db *gorm.DB) *DeviceController {
	return &DeviceController{
		deviceService: services.NewDeviceService(db),
	}
}

func (c *DeviceController) GetDevices(ctx *gin.Context) {
	devices, err := c.deviceService.GetAllDevices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get devices",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    devices,
	})
}

func (c *DeviceController) GetDeviceByID(ctx *gin.Context) {
	id := ctx.Param("id")
	device, err := c.deviceService.GetDeviceByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Device not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    device,
	})
}

func (c *DeviceController) ControlDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Action string `json:"action" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := c.deviceService.ControlDevice(id, req.Action); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to control device",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Device control successful",
	})
}

func (c *DeviceController) GetDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	status, err := c.deviceService.GetDeviceStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Device not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}
