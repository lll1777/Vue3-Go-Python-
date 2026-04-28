package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"parking-system-go/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSpotNotAvailable = errors.New("parking spot is not available")
	ErrSpotLocked       = errors.New("parking spot is locked")
	ErrTimeConflict     = errors.New("parking spot has conflicting reservation")
	ErrVersionMismatch  = errors.New("parking spot version mismatch, concurrent update detected")
	ErrReservationExpired = errors.New("reservation has expired")
	ErrNoActiveParking  = errors.New("no active parking record found")
)

type ParkingService struct {
	db *gorm.DB
}

func NewParkingService(db *gorm.DB) *ParkingService {
	return &ParkingService{db: db}
}

func (s *ParkingService) GetAllParkingLots() ([]models.ParkingLot, error) {
	var lots []models.ParkingLot
	if err := s.db.Find(&lots).Error; err != nil {
		return nil, err
	}
	return lots, nil
}

func (s *ParkingService) GetParkingLotByID(id string) (*models.ParkingLot, error) {
	var lot models.ParkingLot
	if err := s.db.First(&lot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &lot, nil
}

func (s *ParkingService) GetAllParkingSpots() ([]models.ParkingSpot, error) {
	var spots []models.ParkingSpot
	if err := s.db.Find(&spots).Error; err != nil {
		return nil, err
	}
	return spots, nil
}

func (s *ParkingService) GetParkingSpotsByLot(lotID string) ([]models.ParkingSpot, error) {
	var spots []models.ParkingSpot
	if err := s.db.Where("parking_lot_id = ?", lotID).Find(&spots).Error; err != nil {
		return nil, err
	}
	return spots, nil
}

func (s *ParkingService) GetParkingSpotByID(id string) (*models.ParkingSpot, error) {
	var spot models.ParkingSpot
	if err := s.db.First(&spot, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &spot, nil
}

func (s *ParkingService) UpdateParkingSpotStatus(id string, status string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", id).Error; err != nil {
			return err
		}

		oldVersion := spot.Version
		spot.Status = status
		spot.Version++

		result := tx.Model(&spot).Where("id = ? AND version = ?", id, oldVersion).Updates(map[string]interface{}{
			"status":   status,
			"version":  spot.Version,
			"updated_at": time.Now(),
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ParkingService) GetAvailableSpots() ([]models.ParkingSpot, error) {
	var spots []models.ParkingSpot
	if err := s.db.Where("status = ?", "available").Find(&spots).Error; err != nil {
		return nil, err
	}
	return spots, nil
}

func (s *ParkingService) LockSpot(spotID string, reservationID string, lockDuration time.Duration) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", spotID).Error; err != nil {
			return err
		}

		if spot.IsLocked() {
			return ErrSpotLocked
		}

		if spot.Status != "available" {
			return ErrSpotNotAvailable
		}

		oldVersion := spot.Version
		now := time.Now()
		expiresAt := now.Add(lockDuration)

		result := tx.Model(&spot).Where("id = ? AND version = ?", spotID, oldVersion).Updates(map[string]interface{}{
			"locked_by":       reservationID,
			"locked_at":       now,
			"lock_expires_at": expiresAt,
			"version":         oldVersion + 1,
			"updated_at":      now,
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ParkingService) UnlockSpot(spotID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", spotID).Error; err != nil {
			return err
		}

		oldVersion := spot.Version

		result := tx.Model(&spot).Where("id = ? AND version = ?", spotID, oldVersion).Updates(map[string]interface{}{
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
			"version":         oldVersion + 1,
			"updated_at":      time.Now(),
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

type ReservationService struct {
	db *gorm.DB
}

func NewReservationService(db *gorm.DB) *ReservationService {
	return &ReservationService{db: db}
}

func (s *ReservationService) CreateReservation(reservation *models.Reservation) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
			return errors.New("parking spot not found")
		}

		if spot.Status != "available" {
			return ErrSpotNotAvailable
		}

		if spot.IsLocked() {
			return ErrSpotLocked
		}

		if err := s.checkTimeConflict(tx, reservation.ParkingSpotID, reservation.StartTime, reservation.EndTime, ""); err != nil {
			return err
		}

		oldVersion := spot.Version
		lockDuration := reservation.EndTime.Sub(reservation.StartTime) + 30*time.Minute

		reservation.Status = "pending"
		if err := tx.Create(reservation).Error; err != nil {
			return err
		}

		result := tx.Model(&spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":          "reserved",
			"locked_by":       reservation.ID,
			"locked_at":       time.Now(),
			"lock_expires_at": reservation.EndTime.Add(30 * time.Minute),
			"version":         oldVersion + 1,
			"updated_at":      time.Now(),
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ReservationService) checkTimeConflict(tx *gorm.DB, spotID string, startTime, endTime time.Time, excludeReservationID string) error {
	var conflictingReservations []models.Reservation
	
	query := tx.Where("parking_spot_id = ? AND status IN ? AND start_time < ? AND end_time > ?",
		spotID,
		[]string{"pending", "active"},
		endTime,
		startTime,
	)

	if excludeReservationID != "" {
		query = query.Where("id != ?", excludeReservationID)
	}

	if err := query.Find(&conflictingReservations).Error; err != nil {
		return err
	}

	if len(conflictingReservations) > 0 {
		return fmt.Errorf("%w: spot %s has conflicting reservations", ErrTimeConflict, spotID)
	}

	return nil
}

func (s *ReservationService) GetAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	if err := s.db.Order("created_at DESC").Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

func (s *ReservationService) GetReservationByID(id string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := s.db.First(&reservation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (s *ReservationService) GetActiveReservationByLicensePlate(licensePlate string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := s.db.Where("license_plate = ? AND status IN ?", 
		licensePlate, []string{"pending", "active"}).First(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (s *ReservationService) CheckInReservation(reservationID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var reservation models.Reservation
		if err := tx.First(&reservation, "id = ?", reservationID).Error; err != nil {
			return err
		}

		if reservation.Status != "pending" {
			return errors.New("reservation is not pending")
		}

		now := time.Now()
		graceEnd := reservation.StartTime.Add(time.Duration(reservation.GraceMinutes) * time.Minute)

		if now.After(graceEnd) {
			return ErrReservationExpired
		}

		reservation.Status = "active"
		reservation.CheckInTime = &now

		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
			return err
		}

		oldVersion := spot.Version
		spot.Status = "occupied"

		result := tx.Model(&spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":     "occupied",
			"version":    oldVersion + 1,
			"updated_at": now,
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ReservationService) CheckOutReservation(reservationID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var reservation models.Reservation
		if err := tx.First(&reservation, "id = ?", reservationID).Error; err != nil {
			return err
		}

		if reservation.Status != "active" {
			return errors.New("reservation is not active")
		}

		now := time.Now()
		reservation.Status = "completed"
		reservation.CheckOutTime = &now

		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
			return err
		}

		oldVersion := spot.Version

		result := tx.Model(&spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":          "available",
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
			"version":         oldVersion + 1,
			"updated_at":      now,
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ReservationService) CancelReservation(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var reservation models.Reservation
		if err := tx.First(&reservation, "id = ?", id).Error; err != nil {
			return err
		}

		if reservation.Status == "completed" || reservation.Status == "cancelled" {
			return errors.New("reservation cannot be cancelled")
		}

		reservation.Status = "cancelled"
		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
			return err
		}

		oldVersion := spot.Version

		result := tx.Model(&spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":          "available",
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
			"version":         oldVersion + 1,
			"updated_at":      time.Now(),
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		return nil
	})
}

func (s *ReservationService) CleanupExpiredReservations() (int64, error) {
	var expiredReservations []models.Reservation
	
	if err := s.db.Where("status IN ? AND check_in_time IS NULL", 
		[]string{"pending", "active"}).Find(&expiredReservations).Error; err != nil {
		return 0, err
	}

	var cleanedCount int64

	for _, reservation := range expiredReservations {
		graceEnd := reservation.StartTime.Add(time.Duration(reservation.GraceMinutes) * time.Minute)
		if time.Now().After(graceEnd) {
			if err := s.CancelReservation(reservation.ID); err == nil {
				cleanedCount++
			}
		}
	}

	return cleanedCount, nil
}

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) GetOrderByID(id string) (*models.Order, error) {
	var order models.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	return s.db.Create(order).Error
}

func (s *OrderService) PayOrder(id string, paymentMethod string, paidAmount float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.First(&order, "id = ?", id).Error; err != nil {
			return err
		}

		if order.Status == "paid" || order.Status == "completed" {
			return errors.New("order already paid")
		}

		now := time.Now()
		order.Status = "paid"
		order.PaymentMethod = paymentMethod
		order.PaidAmount = paidAmount
		order.PaidTime = &now

		return tx.Save(&order).Error
	})
}

type AccessControlService struct {
	db *gorm.DB
}

func NewAccessControlService(db *gorm.DB) *AccessControlService {
	return &AccessControlService{db: db}
}

func (s *AccessControlService) VehicleEntry(licensePlate string, deviceID string, confidence float64) (*models.AccessLog, error) {
	return s.vehicleEntryWithReservation(licensePlate, deviceID, confidence, "")
}

func (s *AccessControlService) vehicleEntryWithReservation(licensePlate string, deviceID string, confidence float64, reservationID string) (*models.AccessLog, error) {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var spot *models.ParkingSpot
		var reservation *models.Reservation
		var err error

		if reservationID == "" {
			reservation, err = s.findActiveReservationByPlate(tx, licensePlate)
			if err == nil && reservation != nil {
				reservationID = reservation.ID
			}
		}

		if reservationID != "" {
			spot, err = s.getAndValidateReservedSpot(tx, reservationID, licensePlate)
			if err != nil {
				return err
			}
		} else {
			spot, err = s.findAvailableSpot(tx)
			if err != nil {
				return err
			}
		}

		now := time.Now()
		accessLog := &models.AccessLog{
			Type:          "entry",
			LicensePlate:  licensePlate,
			DeviceID:      deviceID,
			EntryTime:     now,
			Confidence:    confidence,
			Status:        "success",
			ReservationID: reservationID,
		}

		if err := tx.Create(accessLog).Error; err != nil {
			return err
		}

		oldVersion := spot.Version
		spot.Status = "occupied"

		result := tx.Model(spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":     "occupied",
			"version":    oldVersion + 1,
			"updated_at": now,
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		order := &models.Order{
			LicensePlate:  licensePlate,
			ParkingSpotID: spot.ID,
			ReservationID: reservationID,
			EntryTime:     now,
			Status:        "unpaid",
			TotalAmount:   0,
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		if reservationID != "" {
			var res models.Reservation
			if err := tx.First(&res, "id = ?", reservationID).Error; err == nil {
				if res.Status == "pending" {
					res.Status = "active"
					res.CheckInTime = &now
					tx.Save(&res)
				}
			}
		}

		return nil
	}).(*models.AccessLog), nil
}

func (s *AccessControlService) findActiveReservationByPlate(tx *gorm.DB, licensePlate string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := tx.Where("license_plate = ? AND status IN ?", 
		licensePlate, []string{"pending", "active"}).First(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (s *AccessControlService) getAndValidateReservedSpot(tx *gorm.DB, reservationID string, licensePlate string) (*models.ParkingSpot, error) {
	var reservation models.Reservation
	if err := tx.First(&reservation, "id = ?", reservationID).Error; err != nil {
		return nil, errors.New("reservation not found")
	}

	if reservation.LicensePlate != licensePlate {
		return nil, errors.New("license plate does not match reservation")
	}

	graceEnd := reservation.StartTime.Add(time.Duration(reservation.GraceMinutes) * time.Minute)
	if time.Now().After(graceEnd) && reservation.Status == "pending" {
		return nil, ErrReservationExpired
	}

	var spot models.ParkingSpot
	if err := tx.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
		return nil, errors.New("parking spot not found")
	}

	if spot.Status != "reserved" && spot.Status != "available" {
		return nil, ErrSpotNotAvailable
	}

	return &spot, nil
}

func (s *AccessControlService) findAvailableSpot(tx *gorm.DB) (*models.ParkingSpot, error) {
	var spot models.ParkingSpot
	
	err := tx.Transaction(func(tx2 *gorm.DB) error {
		result := tx2.Raw(`
			SELECT * FROM parking_spots 
			WHERE status = 'available' AND (locked_by IS NULL OR lock_expires_at < ?)
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`, time.Now()).Scan(&spot)

		if result.Error != nil {
			return result.Error
		}

		if spot.ID == "" {
			return errors.New("no available parking spots")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &spot, nil
}

func (s *AccessControlService) VehicleExit(licensePlate string, deviceID string, confidence float64) (*models.AccessLog, error) {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Where("license_plate = ? AND status = ?", licensePlate, "unpaid").First(&order).Error; err != nil {
			return ErrNoActiveParking
		}

		now := time.Now()
		duration := now.Sub(order.EntryTime)
		minutes := int64(duration.Minutes())

		billingService := NewBillingService(tx)
		totalAmount, billingDetails, err := billingService.CalculateDetailedFee(order.EntryTime, now, "standard")
		if err != nil {
			return err
		}

		billingDetailsJSON, _ := json.Marshal(billingDetails)

		order.ExitTime = &now
		order.ParkingDuration = minutes
		order.TotalAmount = totalAmount
		order.Status = "unpaid"
		order.BillingDetails = string(billingDetailsJSON)

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		accessLog := &models.AccessLog{
			Type:          "exit",
			LicensePlate:  licensePlate,
			DeviceID:      deviceID,
			EntryTime:     order.EntryTime,
			ExitTime:      &now,
			Confidence:    confidence,
			Status:        "success",
			ReservationID: order.ReservationID,
		}

		if err := tx.Create(accessLog).Error; err != nil {
			return err
		}

		var spot models.ParkingSpot
		if err := tx.First(&spot, "id = ?", order.ParkingSpotID).Error; err != nil {
			return err
		}

		oldVersion := spot.Version

		result := tx.Model(&spot).Where("id = ? AND version = ?", spot.ID, oldVersion).Updates(map[string]interface{}{
			"status":          "available",
			"locked_by":       nil,
			"locked_at":       nil,
			"lock_expires_at": nil,
			"version":         oldVersion + 1,
			"updated_at":      now,
		})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionMismatch
		}

		if order.ReservationID != "" {
			var res models.Reservation
			if err := tx.First(&res, "id = ?", order.ReservationID).Error; err == nil {
				if res.Status == "active" {
					res.Status = "completed"
					res.CheckOutTime = &now
					tx.Save(&res)
				}
			}
		}

		return nil
	}).(*models.AccessLog), nil
}

func (s *AccessControlService) GetAccessLogs() ([]models.AccessLog, error) {
	var logs []models.AccessLog
	if err := s.db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

type BillingService struct {
	db *gorm.DB
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

type BillingPeriod struct {
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	PeriodType string
	HourlyRate float64
}

type DetailedBillingResult struct {
	Periods       []BillingPeriod
	DailyTotals   map[string]float64
	TotalAmount   float64
	TotalDuration time.Duration
}

func (s *BillingService) GetAllBillingRules() ([]models.BillingRule, error) {
	var rules []models.BillingRule
	if err := s.db.Where("status = ?", "active").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *BillingService) GetBillingRuleByID(id string) (*models.BillingRule, error) {
	var rule models.BillingRule
	if err := s.db.First(&rule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *BillingService) getActiveRule(spotType string) (*models.BillingRule, error) {
	var rule models.BillingRule
	if err := s.db.Where("spot_type = ? AND status = ?", spotType, "active").First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.BillingRule{
				FirstHour:   10.0,
				HourlyRate:  8.0,
				DailyMax:    80.0,
				MinCharge:   5.0,
				GracePeriod: 15,
				PeakStart:   "07:00",
				PeakEnd:     "09:00",
				PeakRate:    12.0,
				NightStart:  "22:00",
				NightEnd:    "06:00",
				NightRate:   5.0,
				HolidayRate: 10.0,
			}, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (s *BillingService) CalculateFee(minutes int64, spotType string) (float64, error) {
	rule, err := s.getActiveRule(spotType)
	if err != nil {
		return 0, err
	}

	if minutes <= int64(rule.GracePeriod) {
		return 0, nil
	}

	hours := float64(minutes) / 60.0
	var fee float64

	if hours <= 1 {
		fee = rule.FirstHour
	} else {
		additionalHours := hours - 1
		fee = rule.FirstHour + float64(int(additionalHours+0.999))*rule.HourlyRate
	}

	if fee < rule.MinCharge {
		fee = rule.MinCharge
	}

	if fee > rule.DailyMax {
		fee = rule.DailyMax
	}

	return math.Round(fee*100) / 100, nil
}

func (s *BillingService) CalculateDetailedFee(entryTime, exitTime time.Time, spotType string) (float64, []BillingPeriod, error) {
	rule, err := s.getActiveRule(spotType)
	if err != nil {
		return 0, nil, err
	}

	duration := exitTime.Sub(entryTime)
	totalMinutes := int64(duration.Minutes())

	if totalMinutes <= int64(rule.GracePeriod) {
		return 0, []BillingPeriod{}, nil
	}

	periods := s.splitIntoPeriods(entryTime, exitTime, rule)

	dailyTotals := make(map[string]float64)
	var totalAmount float64

	for _, period := range periods {
		periodMinutes := int64(period.Duration.Minutes())
		periodHours := float64(periodMinutes) / 60.0
		
		var periodFee float64
		if periodHours <= 1 {
			periodFee = rule.FirstHour
		} else {
			periodFee = rule.FirstHour + float64(int(periodHours-1+0.999))*period.HourlyRate
		}

		dateKey := period.StartTime.Format("2006-01-02")
		dailyTotals[dateKey] += periodFee
		totalAmount += periodFee
	}

	for dateKey, dailyTotal := range dailyTotals {
		if dailyTotal > rule.DailyMax {
			discount := dailyTotal - rule.DailyMax
			totalAmount -= discount
			dailyTotals[dateKey] = rule.DailyMax
		}
	}

	if totalAmount < rule.MinCharge {
		totalAmount = rule.MinCharge
	}

	return math.Round(totalAmount*100) / 100, periods, nil
}

func (s *BillingService) splitIntoPeriods(entryTime, exitTime time.Time, rule *models.BillingRule) []BillingPeriod {
	var periods []BillingPeriod

	current := entryTime
	for current.Before(exitTime) {
		periodEnd := s.getNextPeriodBoundary(current, exitTime, rule)
		
		periodType, hourlyRate := s.getPeriodInfo(current, rule)
		
		duration := periodEnd.Sub(current)
		if duration > 0 {
			periods = append(periods, BillingPeriod{
				StartTime:  current,
				EndTime:    periodEnd,
				Duration:   duration,
				PeriodType: periodType,
				HourlyRate: hourlyRate,
			})
		}

		current = periodEnd
	}

	return periods
}

func (s *BillingService) getNextPeriodBoundary(current, exitTime time.Time, rule *models.BillingRule) time.Time {
	boundaries := []time.Time{}

	dayEnd := time.Date(current.Year(), current.Month(), current.Day()+1, 0, 0, 0, 0, current.Location())
	boundaries = append(boundaries, dayEnd)

	if rule.PeakStart != "" && rule.PeakEnd != "" {
		peakStart, _ := parseTimeOfDay(current, rule.PeakStart)
		peakEnd, _ := parseTimeOfDay(current, rule.PeakEnd)
		
		if current.Before(peakStart) {
			boundaries = append(boundaries, peakStart)
		} else if current.Before(peakEnd) {
			boundaries = append(boundaries, peakEnd)
		}
	}

	if rule.NightStart != "" && rule.NightEnd != "" {
		nightStart, _ := parseTimeOfDay(current, rule.NightStart)
		nightEnd, _ := parseTimeOfDay(current.Add(24*time.Hour), rule.NightEnd)
		
		if current.Before(nightStart) {
			boundaries = append(boundaries, nightStart)
		} else if current.Before(nightEnd) {
			boundaries = append(boundaries, nightEnd)
		}
	}

	nextBoundary := exitTime
	for _, b := range boundaries {
		if b.After(current) && b.Before(nextBoundary) {
			nextBoundary = b
		}
	}

	return nextBoundary
}

func (s *BillingService) getPeriodInfo(t time.Time, rule *models.BillingRule) (string, float64) {
	if s.isHoliday(t) && rule.HolidayRate > 0 {
		return "holiday", rule.HolidayRate
	}

	if rule.NightStart != "" && rule.NightEnd != "" {
		nightStart, _ := parseTimeOfDay(t, rule.NightStart)
		nightEnd, _ := parseTimeOfDay(t.Add(24*time.Hour), rule.NightEnd)
		
		if (t.After(nightStart) || t.Equal(nightStart)) && t.Before(nightEnd) {
			return "night", rule.NightRate
		}
	}

	if rule.PeakStart != "" && rule.PeakEnd != "" {
		peakStart, _ := parseTimeOfDay(t, rule.PeakStart)
		peakEnd, _ := parseTimeOfDay(t, rule.PeakEnd)
		
		if t.After(peakStart) && t.Before(peakEnd) {
			return "peak", rule.PeakRate
		}
	}

	return "normal", rule.HourlyRate
}

func (s *BillingService) isHoliday(t time.Time) bool {
	month := int(t.Month())
	day := t.Day()

	holidays := [][2]int{
		{1, 1},
		{10, 1},
		{5, 1},
	}

	for _, holiday := range holidays {
		if month == holiday[0] && day == holiday[1] {
			return true
		}
	}

	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return true
	}

	return false
}

func parseTimeOfDay(base time.Time, timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return base, errors.New("invalid time format")
	}

	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	return time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, base.Location()), nil
}

func (s *BillingService) UpdateBillingRule(id string, updates map[string]interface{}) error {
	result := s.db.Model(&models.BillingRule{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("billing rule not found")
	}
	return nil
}

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

func (s *DeviceService) GetAllDevices() ([]models.Device, error) {
	var devices []models.Device
	if err := s.db.Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *DeviceService) GetDeviceByID(id string) (*models.Device, error) {
	var device models.Device
	if err := s.db.First(&device, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *DeviceService) ControlDevice(id string, action string) error {
	var device models.Device
	if err := s.db.First(&device, "id = ?", id).Error; err != nil {
		return err
	}

	switch action {
	case "lock":
		device.Status = "locked"
	case "unlock":
		device.Status = "online"
	case "reset":
		device.Status = "online"
	default:
		return errors.New("invalid action")
	}

	device.LastActiveAt = time.Now()
	return s.db.Save(&device).Error
}

func (s *DeviceService) GetDeviceStatus(id string) (map[string]interface{}, error) {
	var device models.Device
	if err := s.db.First(&device, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"device_id":      device.ID,
		"device_no":      device.DeviceNo,
		"status":         device.Status,
		"last_active_at": device.LastActiveAt,
	}, nil
}
