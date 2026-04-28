package services

import (
	"errors"
	"parking-system-go/models"
	"time"

	"gorm.io/gorm"
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
	result := s.db.Model(&models.ParkingSpot{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("parking spot not found")
	}
	return nil
}

func (s *ParkingService) GetAvailableSpots() ([]models.ParkingSpot, error) {
	var spots []models.ParkingSpot
	if err := s.db.Where("status = ?", "available").Find(&spots).Error; err != nil {
		return nil, err
	}
	return spots, nil
}

type ReservationService struct {
	db *gorm.DB
}

func NewReservationService(db *gorm.DB) *ReservationService {
	return &ReservationService{db: db}
}

func (s *ReservationService) CreateReservation(reservation *models.Reservation) error {
	var spot models.ParkingSpot
	if err := s.db.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err != nil {
		return errors.New("parking spot not found")
	}

	if spot.Status != "available" {
		return errors.New("parking spot is not available")
	}

	spot.Status = "reserved"
	if err := s.db.Save(&spot).Error; err != nil {
		return err
	}

	reservation.Status = "active"
	return s.db.Create(reservation).Error
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

func (s *ReservationService) CancelReservation(id string) error {
	var reservation models.Reservation
	if err := s.db.First(&reservation, "id = ?", id).Error; err != nil {
		return err
	}

	if reservation.Status == "completed" || reservation.Status == "cancelled" {
		return errors.New("reservation cannot be cancelled")
	}

	reservation.Status = "cancelled"
	if err := s.db.Save(&reservation).Error; err != nil {
		return err
	}

	var spot models.ParkingSpot
	if err := s.db.First(&spot, "id = ?", reservation.ParkingSpotID).Error; err == nil {
		spot.Status = "available"
		s.db.Save(&spot)
	}

	return nil
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
	var order models.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
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

	return s.db.Save(&order).Error
}

type AccessControlService struct {
	db *gorm.DB
}

func NewAccessControlService(db *gorm.DB) *AccessControlService {
	return &AccessControlService{db: db}
}

func (s *AccessControlService) VehicleEntry(licensePlate string, deviceID string, confidence float64) (*models.AccessLog, error) {
	spot, err := s.findAvailableSpot()
	if err != nil {
		return nil, err
	}

	accessLog := &models.AccessLog{
		Type:         "entry",
		LicensePlate: licensePlate,
		DeviceID:     deviceID,
		EntryTime:    time.Now(),
		Confidence:   confidence,
		Status:       "success",
	}

	if err := s.db.Create(accessLog).Error; err != nil {
		return nil, err
	}

	spot.Status = "occupied"
	if err := s.db.Save(spot).Error; err != nil {
		return nil, err
	}

	order := &models.Order{
		LicensePlate:  licensePlate,
		ParkingSpotID: spot.ID,
		EntryTime:     time.Now(),
		Status:        "unpaid",
		TotalAmount:   0,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	return accessLog, nil
}

func (s *AccessControlService) VehicleExit(licensePlate string, deviceID string, confidence float64) (*models.AccessLog, error) {
	var order models.Order
	if err := s.db.Where("license_plate = ? AND status = ?", licensePlate, "unpaid").First(&order).Error; err != nil {
		return nil, errors.New("no active parking record found")
	}

	now := time.Now()
	duration := now.Sub(order.EntryTime)
	minutes := int64(duration.Minutes())

	billingService := NewBillingService(s.db)
	totalAmount, err := billingService.CalculateFee(minutes, "standard")
	if err != nil {
		return nil, err
	}

	order.ExitTime = &now
	order.ParkingDuration = minutes
	order.TotalAmount = totalAmount
	order.Status = "unpaid"

	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	accessLog := &models.AccessLog{
		Type:         "exit",
		LicensePlate: licensePlate,
		DeviceID:     deviceID,
		EntryTime:    order.EntryTime,
		ExitTime:     &now,
		Confidence:   confidence,
		Status:       "success",
	}

	if err := s.db.Create(accessLog).Error; err != nil {
		return nil, err
	}

	var spot models.ParkingSpot
	if err := s.db.First(&spot, "id = ?", order.ParkingSpotID).Error; err == nil {
		spot.Status = "available"
		s.db.Save(&spot)
	}

	return accessLog, nil
}

func (s *AccessControlService) findAvailableSpot() (*models.ParkingSpot, error) {
	var spot models.ParkingSpot
	if err := s.db.Where("status = ?", "available").First(&spot).Error; err != nil {
		return nil, errors.New("no available parking spots")
	}
	return &spot, nil
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

func (s *BillingService) CalculateFee(minutes int64, spotType string) (float64, error) {
	var rule models.BillingRule
	if err := s.db.Where("spot_type = ? AND status = ?", spotType, "active").First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rule = models.BillingRule{
				FirstHour:   10.0,
				HourlyRate:  8.0,
				DailyMax:    80.0,
				MinCharge:   5.0,
				GracePeriod: 15,
			}
		} else {
			return 0, err
		}
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

	return fee, nil
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
		"device_id":     device.ID,
		"device_no":     device.DeviceNo,
		"status":        device.Status,
		"last_active_at": device.LastActiveAt,
	}, nil
}
