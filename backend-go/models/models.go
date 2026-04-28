package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ParkingLot struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	TotalSpots  int       `json:"total_spots"`
	FreeSpots   int       `json:"free_spots"`
	Status      string    `json:"status" gorm:"default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ParkingSpot struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	SpotNumber       string    `json:"spot_number" gorm:"unique;not null"`
	ParkingLotID     string    `json:"parking_lot_id"`
	Zone             string    `json:"zone"`
	Floor            int       `json:"floor"`
	Type             string    `json:"type"`
	Status           string    `json:"status" gorm:"default:available"`
	CurrentVehicleID string    `json:"current_vehicle_id,omitempty"`
	DeviceID         string    `json:"device_id,omitempty"`
	Version          int       `json:"version" gorm:"default:0"`
	LockedBy         string    `json:"locked_by,omitempty"`
	LockedAt         *time.Time `json:"locked_at,omitempty"`
	LockExpiresAt    *time.Time `json:"lock_expires_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Reservation struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ReservationNo string    `json:"reservation_no" gorm:"unique;not null"`
	ParkingSpotID string    `json:"parking_spot_id" gorm:"not null"`
	LicensePlate  string    `json:"license_plate" gorm:"not null"`
	Phone         string    `json:"phone"`
	UserID        string    `json:"user_id,omitempty"`
	StartTime     time.Time `json:"start_time" gorm:"not null"`
	EndTime       time.Time `json:"end_time" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:pending"`
	TotalFee      float64   `json:"total_fee"`
	Notes         string    `json:"notes"`
	CheckInTime   *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime  *time.Time `json:"check_out_time,omitempty"`
	GraceMinutes  int       `json:"grace_minutes" gorm:"default:15"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Order struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrderNo        string    `json:"order_no" gorm:"unique;not null"`
	Type           string    `json:"type" gorm:"default:parking"`
	LicensePlate   string    `json:"license_plate" gorm:"not null"`
	ParkingSpotID  string    `json:"parking_spot_id"`
	ReservationID  string    `json:"reservation_id,omitempty"`
	EntryTime      time.Time `json:"entry_time" gorm:"not null"`
	ExitTime       *time.Time `json:"exit_time"`
	ParkingDuration int64     `json:"parking_duration"`
	BillingRuleID  string    `json:"billing_rule_id"`
	TotalAmount    float64   `json:"total_amount"`
	PaidAmount     float64   `json:"paid_amount"`
	Status         string    `json:"status" gorm:"default:unpaid"`
	PaymentMethod  string    `json:"payment_method,omitempty"`
	PaidTime       *time.Time `json:"paid_time"`
	BillingDetails string    `json:"billing_details,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BillingRule struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	SpotType    string    `json:"spot_type" gorm:"default:standard"`
	FirstHour   float64   `json:"first_hour"`
	HourlyRate  float64   `json:"hourly_rate"`
	DailyMax    float64   `json:"daily_max"`
	MinCharge   float64   `json:"min_charge"`
	GracePeriod int       `json:"grace_period"`
	PeakStart   string    `json:"peak_start,omitempty"`
	PeakEnd     string    `json:"peak_end,omitempty"`
	PeakRate    float64   `json:"peak_rate"`
	NightStart  string    `json:"night_start,omitempty"`
	NightEnd    string    `json:"night_end,omitempty"`
	NightRate   float64   `json:"night_rate"`
	HolidayRate float64   `json:"holiday_rate"`
	Status      string    `json:"status" gorm:"default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BillingDetail struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	OrderID     string    `json:"order_id" gorm:"not null"`
	Date        time.Time `json:"date"`
	DayOfWeek   int       `json:"day_of_week"`
	IsHoliday   bool      `json:"is_holiday"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int       `json:"duration_minutes"`
	PeriodType  string    `json:"period_type"`
	HourlyRate  float64   `json:"hourly_rate"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

type Device struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	DeviceNo     string    `json:"device_no" gorm:"unique;not null"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Status       string    `json:"status" gorm:"default:online"`
	Location     string    `json:"location"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AccessLog struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Type         string    `json:"type" gorm:"not null"`
	LicensePlate string    `json:"license_plate" gorm:"not null"`
	DeviceID     string    `json:"device_id"`
	EntryTime    time.Time `json:"entry_time"`
	ExitTime     *time.Time `json:"exit_time"`
	ImageURL     string    `json:"image_url"`
	Confidence   float64   `json:"confidence"`
	Status       string    `json:"status" gorm:"default:success"`
	ReservationID string   `json:"reservation_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p *ParkingLot) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (p *ParkingSpot) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	p.Version = 0
	return nil
}

func (r *Reservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.ReservationNo == "" {
		r.ReservationNo = "RES" + time.Now().Format("20060102150405")
	}
	if r.GraceMinutes == 0 {
		r.GraceMinutes = 15
	}
	return nil
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.OrderNo == "" {
		o.OrderNo = "ORD" + time.Now().Format("20060102150405")
	}
	return nil
}

func (b *BillingRule) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

func (bd *BillingDetail) BeforeCreate(tx *gorm.DB) error {
	if bd.ID == "" {
		bd.ID = uuid.New().String()
	}
	return nil
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

func (a *AccessLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (p *ParkingSpot) IsAvailable() bool {
	return p.Status == "available"
}

func (p *ParkingSpot) IsLocked() bool {
	if p.LockExpiresAt == nil {
		return p.LockedBy != ""
	}
	return p.LockedBy != "" && time.Now().Before(*p.LockExpiresAt)
}

func (p *ParkingSpot) CanBeReserved() bool {
	return p.IsAvailable() && !p.IsLocked()
}

func (r *Reservation) IsActive() bool {
	return r.Status == "active" || r.Status == "pending"
}

func (r *Reservation) HasStarted() bool {
	return r.Status == "active" || (r.Status == "pending" && time.Now().After(r.StartTime))
}

func (r *Reservation) HasEnded() bool {
	return time.Now().After(r.EndTime)
}

func (r *Reservation) IsExpired() bool {
	if r.Status == "cancelled" || r.Status == "completed" {
		return false
	}
	graceEnd := r.StartTime.Add(time.Duration(r.GraceMinutes) * time.Minute)
	return time.Now().After(graceEnd) && r.CheckInTime == nil
}
