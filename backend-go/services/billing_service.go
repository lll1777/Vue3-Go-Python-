package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"parking-system-go/models"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidBillingRule   = errors.New("invalid billing rule")
	ErrPeriodOverlap        = errors.New("time periods overlap in billing rule")
	ErrNegativeRate         = errors.New("hourly rate cannot be negative")
	ErrInvalidTimeRange     = errors.New("invalid time range")
	ErrGracePeriodNegative  = errors.New("grace period cannot be negative")
	ErrDailyMaxLessThanMin  = errors.New("daily max cannot be less than min charge")
)

const (
	PeriodTypeNormal   = "normal"
	PeriodTypePeak     = "peak"
	PeriodTypeNight    = "night"
	PeriodTypeHoliday  = "holiday"
)

type DayBoundary struct {
	Date  time.Time
	Start time.Time
	End   time.Time
}

type BillingPeriodDetail struct {
	PeriodType   string        `json:"period_type"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	DurationMin  int64         `json:"duration_minutes"`
	HourlyRate   float64       `json:"hourly_rate"`
	BaseRate     float64       `json:"base_rate"`
	IsFirstHour  bool          `json:"is_first_hour"`
	PeriodAmount float64       `json:"period_amount"`
}

type DailyBilling struct {
	Date       string                 `json:"date"`
	DayOfWeek  int                    `json:"day_of_week"`
	IsHoliday  bool                   `json:"is_holiday"`
	Periods    []BillingPeriodDetail  `json:"periods"`
	SubTotal   float64                `json:"sub_total"`
	DailyMax   float64                `json:"daily_max"`
	Discount   float64                `json:"discount"`
	DailyTotal float64                `json:"daily_total"`
}

type BillingCalculationResult struct {
	EntryTime          time.Time       `json:"entry_time"`
	ExitTime           time.Time       `json:"exit_time"`
	TotalDuration      time.Duration   `json:"total_duration"`
	TotalDurationMin   int64           `json:"total_duration_minutes"`
	
	WithinGracePeriod  bool            `json:"within_grace_period"`
	GracePeriodMinutes int             `json:"grace_period_minutes"`
	
	DailyBillings      []DailyBilling  `json:"daily_billings"`
	
	FirstHourUsed      float64         `json:"first_hour_rate"`
	FirstHourApplied   bool            `json:"first_hour_applied"`
	
	TotalBeforeRules   float64         `json:"total_before_rules"`
	TotalDiscount      float64         `json:"total_discount"`
	MinChargeApplied   bool            `json:"min_charge_applied"`
	
	FinalAmount        float64         `json:"final_amount"`
	
	RuleSummary map[string]interface{} `json:"rule_summary,omitempty"`
}

type EnhancedBillingService struct {
	db *gorm.DB
}

func NewEnhancedBillingService(db *gorm.DB) *EnhancedBillingService {
	return &EnhancedBillingService{db: db}
}

func (s *EnhancedBillingService) ValidateBillingRule(rule *models.BillingRule) error {
	if rule.HourlyRate < 0 {
		return ErrNegativeRate
	}
	if rule.PeakRate < 0 {
		return ErrNegativeRate
	}
	if rule.NightRate < 0 {
		return ErrNegativeRate
	}
	if rule.HolidayRate < 0 {
		return ErrNegativeRate
	}
	if rule.FirstHour < 0 {
		return ErrNegativeRate
	}
	
	if rule.GracePeriod < 0 {
		return ErrGracePeriodNegative
	}
	
	if rule.DailyMax > 0 && rule.MinCharge > rule.DailyMax {
		return ErrDailyMaxLessThanMin
	}
	
	if rule.PeakStart != "" && rule.PeakEnd != "" {
		peakStart, err := parseTimeStrToMinutes(rule.PeakStart)
		if err != nil {
			return fmt.Errorf("invalid peak_start: %w", err)
		}
		peakEnd, err := parseTimeStrToMinutes(rule.PeakEnd)
		if err != nil {
			return fmt.Errorf("invalid peak_end: %w", err)
		}
		if peakStart >= peakEnd {
			return fmt.Errorf("peak period: %w: start must be before end", ErrInvalidTimeRange)
		}
	}
	
	if rule.PeakStart != "" && rule.NightStart != "" {
		peakStart, _ := parseTimeStrToMinutes(rule.PeakStart)
		peakEnd, _ := parseTimeStrToMinutes(rule.PeakEnd)
		nightStart, _ := parseTimeStrToMinutes(rule.NightStart)
		nightEnd, _ := parseTimeStrToMinutes(rule.NightEnd)
		
		if nightStart < nightEnd {
			if !(peakEnd <= nightStart || peakStart >= nightEnd) {
				return fmt.Errorf("peak period overlaps with night period: %w", ErrPeriodOverlap)
			}
		}
	}
	
	return nil
}

func (s *EnhancedBillingService) GetActiveBillingRule(spotType string) (*models.BillingRule, error) {
	var rule models.BillingRule
	
	if err := s.db.Where("spot_type = ? AND status = ?", spotType, "active").First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultRule := &models.BillingRule{
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
			}
			return defaultRule, nil
		}
		return nil, err
	}
	
	return &rule, nil
}

func (s *EnhancedBillingService) CalculateParkingFee(entryTime, exitTime time.Time, spotType string) (*BillingCalculationResult, error) {
	rule, err := s.GetActiveBillingRule(spotType)
	if err != nil {
		return nil, err
	}
	
	return s.CalculateParkingFeeWithRule(entryTime, exitTime, rule)
}

func (s *EnhancedBillingService) CalculateParkingFeeWithRule(entryTime, exitTime time.Time, rule *models.BillingRule) (*BillingCalculationResult, error) {
	entryTime = entryTime.Truncate(time.Second)
	exitTime = exitTime.Truncate(time.Second)
	
	if exitTime.Before(entryTime) {
		return nil, errors.New("exit time cannot be before entry time")
	}
	
	totalDuration := exitTime.Sub(entryTime)
	totalMinutes := int64(totalDuration.Minutes())
	
	result := &BillingCalculationResult{
		EntryTime:          entryTime,
		ExitTime:           exitTime,
		TotalDuration:      totalDuration,
		TotalDurationMin:   totalMinutes,
		GracePeriodMinutes: rule.GracePeriod,
		FirstHourUsed:      rule.FirstHour,
	}
	
	if totalMinutes <= int64(rule.GracePeriod) {
		result.WithinGracePeriod = true
		result.FinalAmount = 0
		return result, nil
	}
	
	dayBoundaries := s.splitIntoDayBoundaries(entryTime, exitTime)
	
	var dailyBillings []DailyBilling
	var firstHourRemaining = true
	var totalBeforeRules float64
	var totalDiscount float64
	
	for _, dayBoundary := range dayBoundaries {
		dailyBilling, remainingFirstHour := s.calculateDayBilling(
			dayBoundary,
			entryTime,
			exitTime,
			rule,
			firstHourRemaining,
		)
		
		firstHourRemaining = remainingFirstHour
		
		if dailyBilling.SubTotal > 0 {
			firstHourRemaining = false
		}
		
		if rule.DailyMax > 0 && dailyBilling.SubTotal > rule.DailyMax {
			dailyBilling.DailyMax = rule.DailyMax
			dailyBilling.Discount = dailyBilling.SubTotal - rule.DailyMax
			dailyBilling.DailyTotal = rule.DailyMax
		} else {
			dailyBilling.DailyMax = rule.DailyMax
			dailyBilling.Discount = 0
			dailyBilling.DailyTotal = dailyBilling.SubTotal
		}
		
		totalBeforeRules += dailyBilling.SubTotal
		totalDiscount += dailyBilling.Discount
		dailyBillings = append(dailyBillings, dailyBilling)
	}
	
	result.DailyBillings = dailyBillings
	result.FirstHourApplied = !firstHourRemaining
	result.TotalBeforeRules = totalBeforeRules
	result.TotalDiscount = totalDiscount
	
	finalAmount := totalBeforeRules - totalDiscount
	
	if finalAmount < rule.MinCharge {
		result.MinChargeApplied = true
		finalAmount = rule.MinCharge
	}
	
	result.FinalAmount = math.Round(finalAmount*100) / 100
	
	result.RuleSummary = map[string]interface{}{
		"first_hour":    rule.FirstHour,
		"hourly_rate":   rule.HourlyRate,
		"daily_max":     rule.DailyMax,
		"min_charge":    rule.MinCharge,
		"grace_period":  rule.GracePeriod,
		"peak_rate":     rule.PeakRate,
		"night_rate":    rule.NightRate,
		"holiday_rate":  rule.HolidayRate,
	}
	
	return result, nil
}

func (s *EnhancedBillingService) splitIntoDayBoundaries(entryTime, exitTime time.Time) []DayBoundary {
	var boundaries []DayBoundary
	
	currentDate := time.Date(entryTime.Year(), entryTime.Month(), entryTime.Day(), 0, 0, 0, 0, entryTime.Location())
	exitDate := time.Date(exitTime.Year(), exitTime.Month(), exitTime.Day(), 0, 0, 0, 0, exitTime.Location())
	
	for !currentDate.After(exitDate) {
		dayStart := currentDate
		dayEnd := currentDate.Add(24 * time.Hour)
		
		effectiveStart := entryTime
		if dayStart.After(entryTime) {
			effectiveStart = dayStart
		}
		
		effectiveEnd := exitTime
		if dayEnd.Before(exitTime) {
			effectiveEnd = dayEnd
		}
		
		if effectiveStart.Before(effectiveEnd) {
			boundaries = append(boundaries, DayBoundary{
				Date:  currentDate,
				Start: effectiveStart,
				End:   effectiveEnd,
			})
		}
		
		currentDate = currentDate.Add(24 * time.Hour)
	}
	
	return boundaries
}

func (s *EnhancedBillingService) calculateDayBilling(
	dayBoundary DayBoundary,
	globalEntry time.Time,
	globalExit time.Time,
	rule *models.BillingRule,
	applyFirstHour bool,
) (DailyBilling, bool) {
	
	isHoliday := s.isHoliday(dayBoundary.Date)
	
	periods := s.splitDayIntoTimePeriods(dayBoundary.Start, dayBoundary.End, rule, isHoliday)
	
	var billingPeriods []BillingPeriodDetail
	var subTotal float64
	var firstHourRemaining = applyFirstHour
	
	for _, period := range periods {
		detail := BillingPeriodDetail{
			PeriodType:   period.periodType,
			StartTime:    period.start,
			EndTime:      period.end,
			Duration:     period.end.Sub(period.start),
			DurationMin:  int64(period.end.Sub(period.start).Minutes()),
			HourlyRate:   period.hourlyRate,
			BaseRate:     period.hourlyRate,
			IsFirstHour:  false,
			PeriodAmount: 0,
		}
		
		var periodAmount float64
		periodAmount, firstHourRemaining = s.calculatePeriodAmount(
			detail.Duration,
			rule,
			period.hourlyRate,
			firstHourRemaining,
		)
		
		detail.PeriodAmount = periodAmount
		
		if applyFirstHour && !firstHourRemaining {
			detail.IsFirstHour = true
		}
		
		subTotal += detail.PeriodAmount
		billingPeriods = append(billingPeriods, detail)
	}
	
	return DailyBilling{
		Date:      dayBoundary.Date.Format("2006-01-02"),
		DayOfWeek: int(dayBoundary.Date.Weekday()),
		IsHoliday: isHoliday,
		Periods:   billingPeriods,
		SubTotal:  subTotal,
	}, firstHourRemaining
}

type internalPeriod struct {
	periodType string
	start      time.Time
	end        time.Time
	hourlyRate float64
}

func (s *EnhancedBillingService) splitDayIntoTimePeriods(
	dayStart, dayEnd time.Time,
	rule *models.BillingRule,
	isHoliday bool,
) []internalPeriod {
	
	var periods []internalPeriod
	
	current := dayStart
	for current.Before(dayEnd) {
		periodType, hourlyRate := s.determinePeriodTypeAndRate(current, rule, isHoliday)
		
		nextBoundary := s.findNextPeriodBoundary(current, dayEnd, rule, isHoliday)
		
		periods = append(periods, internalPeriod{
			periodType: periodType,
			start:      current,
			end:        nextBoundary,
			hourlyRate: hourlyRate,
		})
		
		current = nextBoundary
	}
	
	return periods
}

func (s *EnhancedBillingService) determinePeriodTypeAndRate(
	t time.Time,
	rule *models.BillingRule,
	isHoliday bool,
) (string, float64) {
	
	if s.isInNightPeriod(t, rule) && rule.NightRate > 0 {
		if isHoliday && rule.HolidayRate > 0 && rule.HolidayRate < rule.NightRate {
			return PeriodTypeHoliday, rule.HolidayRate
		}
		return PeriodTypeNight, rule.NightRate
	}
	
	if isHoliday && rule.HolidayRate > 0 {
		return PeriodTypeHoliday, rule.HolidayRate
	}
	
	if s.isInPeakPeriod(t, rule) && rule.PeakRate > 0 {
		return PeriodTypePeak, rule.PeakRate
	}
	
	return PeriodTypeNormal, rule.HourlyRate
}

func (s *EnhancedBillingService) isInNightPeriod(t time.Time, rule *models.BillingRule) bool {
	if rule.NightStart == "" || rule.NightEnd == "" {
		return false
	}
	
	nightStartMin, _ := parseTimeStrToMinutes(rule.NightStart)
	nightEndMin, _ := parseTimeStrToMinutes(rule.NightEnd)
	currentMin := t.Hour()*60 + t.Minute()
	
	if nightStartMin < nightEndMin {
		return currentMin >= nightStartMin && currentMin < nightEndMin
	} else {
		return currentMin >= nightStartMin || currentMin < nightEndMin
	}
}

func (s *EnhancedBillingService) isInPeakPeriod(t time.Time, rule *models.BillingRule) bool {
	if rule.PeakStart == "" || rule.PeakEnd == "" {
		return false
	}
	
	peakStartMin, _ := parseTimeStrToMinutes(rule.PeakStart)
	peakEndMin, _ := parseTimeStrToMinutes(rule.PeakEnd)
	currentMin := t.Hour()*60 + t.Minute()
	
	return currentMin >= peakStartMin && currentMin < peakEndMin
}

func (s *EnhancedBillingService) findNextPeriodBoundary(
	current, dayEnd time.Time,
	rule *models.BillingRule,
	isHoliday bool,
) time.Time {
	
	var boundaries []time.Time
	
	nextHour := time.Date(current.Year(), current.Month(), current.Day(), current.Hour()+1, 0, 0, 0, current.Location())
	if nextHour.Before(dayEnd) {
		boundaries = append(boundaries, nextHour)
	}
	
	if rule.NightStart != "" && rule.NightEnd != "" {
		nightStartMin, _ := parseTimeStrToMinutes(rule.NightStart)
		nightEndMin, _ := parseTimeStrToMinutes(rule.NightEnd)
		currentMin := current.Hour()*60 + current.Minute()
		
		if nightStartMin < nightEndMin {
			if currentMin < nightStartMin {
				nightStart := time.Date(current.Year(), current.Month(), current.Day(),
					nightStartMin/60, nightStartMin%60, 0, 0, current.Location())
				if nightStart.After(current) && nightStart.Before(dayEnd) {
					boundaries = append(boundaries, nightStart)
				}
			}
			if currentMin < nightEndMin {
				nightEnd := time.Date(current.Year(), current.Month(), current.Day(),
					nightEndMin/60, nightEndMin%60, 0, 0, current.Location())
				if nightEnd.After(current) && nightEnd.Before(dayEnd) {
					boundaries = append(boundaries, nightEnd)
				}
			}
		} else {
			if currentMin < nightEndMin {
				nightEnd := time.Date(current.Year(), current.Month(), current.Day(),
					nightEndMin/60, nightEndMin%60, 0, 0, current.Location())
				if nightEnd.After(current) && nightEnd.Before(dayEnd) {
					boundaries = append(boundaries, nightEnd)
				}
			}
			if currentMin < nightStartMin {
				nightStart := time.Date(current.Year(), current.Month(), current.Day(),
					nightStartMin/60, nightStartMin%60, 0, 0, current.Location())
				if nightStart.After(current) && nightStart.Before(dayEnd) {
					boundaries = append(boundaries, nightStart)
				}
			}
		}
	}
	
	if rule.PeakStart != "" && rule.PeakEnd != "" {
		peakStartMin, _ := parseTimeStrToMinutes(rule.PeakStart)
		peakEndMin, _ := parseTimeStrToMinutes(rule.PeakEnd)
		currentMin := current.Hour()*60 + current.Minute()
		
		if currentMin < peakStartMin {
			peakStart := time.Date(current.Year(), current.Month(), current.Day(),
				peakStartMin/60, peakStartMin%60, 0, 0, current.Location())
			if peakStart.After(current) && peakStart.Before(dayEnd) {
				boundaries = append(boundaries, peakStart)
			}
		}
		if currentMin < peakEndMin {
			peakEnd := time.Date(current.Year(), current.Month(), current.Day(),
				peakEndMin/60, peakEndMin%60, 0, 0, current.Location())
			if peakEnd.After(current) && peakEnd.Before(dayEnd) {
				boundaries = append(boundaries, peakEnd)
			}
		}
	}
	
	if len(boundaries) == 0 {
		return dayEnd
	}
	
	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i].Before(boundaries[j])
	})
	
	nextBoundary := dayEnd
	for _, b := range boundaries {
		if b.After(current) {
			nextBoundary = b
			break
		}
	}
	
	return nextBoundary
}

func (s *EnhancedBillingService) calculatePeriodAmount(
	duration time.Duration,
	rule *models.BillingRule,
	hourlyRate float64,
	applyFirstHour bool,
) (float64, bool) {
	
	hours := duration.Hours()
	minutes := int64(duration.Minutes())
	
	if minutes <= 0 {
		return 0, applyFirstHour
	}
	
	if applyFirstHour {
		if hours <= 1 {
			return rule.FirstHour, false
		} else {
			remainingHours := hours - 1
			additionalAmount := math.Ceil(remainingHours) * hourlyRate
			return rule.FirstHour + additionalAmount, false
		}
	} else {
		return math.Ceil(hours) * hourlyRate, false
	}
}

func (s *EnhancedBillingService) isHoliday(t time.Time) bool {
	month := int(t.Month())
	day := t.Day()
	
	fixedHolidays := map[[2]int]bool{
		{1, 1}:   true,
		{10, 1}:  true,
		{5, 1}:   true,
		{12, 25}: true,
	}
	
	if fixedHolidays[[2]int{month, day}] {
		return true
	}
	
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return true
	}
	
	return false
}

func parseTimeStrToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return 0, errors.New("invalid time format")
	}
	
	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)
	
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, errors.New("invalid time values")
	}
	
	return hour*60 + minute, nil
}

type jsonSerializableResult struct {
	EntryTime          string                 `json:"entry_time"`
	ExitTime           string                 `json:"exit_time"`
	TotalDuration      string                 `json:"total_duration"`
	TotalDurationMin   int64                  `json:"total_duration_minutes"`
	WithinGracePeriod  bool                   `json:"within_grace_period"`
	GracePeriodMinutes int                    `json:"grace_period_minutes"`
	DailyBillings      []DailyBilling         `json:"daily_billings"`
	FirstHourUsed      float64                `json:"first_hour_rate"`
	FirstHourApplied   bool                   `json:"first_hour_applied"`
	TotalBeforeRules   float64                `json:"total_before_rules"`
	TotalDiscount      float64                `json:"total_discount"`
	MinChargeApplied   bool                   `json:"min_charge_applied"`
	FinalAmount        float64                `json:"final_amount"`
	RuleSummary        map[string]interface{} `json:"rule_summary"`
}

func (s *EnhancedBillingService) BillingResultToJSON(result *BillingCalculationResult) (string, error) {
	if result == nil {
		return "{}", nil
	}
	
	jsonResult := jsonSerializableResult{
		EntryTime:          result.EntryTime.Format(time.RFC3339),
		ExitTime:           result.ExitTime.Format(time.RFC3339),
		TotalDuration:      result.TotalDuration.String(),
		TotalDurationMin:   result.TotalDurationMin,
		WithinGracePeriod:  result.WithinGracePeriod,
		GracePeriodMinutes: result.GracePeriodMinutes,
		DailyBillings:      result.DailyBillings,
		FirstHourUsed:      result.FirstHourUsed,
		FirstHourApplied:   result.FirstHourApplied,
		TotalBeforeRules:   result.TotalBeforeRules,
		TotalDiscount:      result.TotalDiscount,
		MinChargeApplied:   result.MinChargeApplied,
		FinalAmount:        result.FinalAmount,
		RuleSummary:        result.RuleSummary,
	}
	
	bytes, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}

func (s *EnhancedBillingService) CalculateFee(minutes int64, spotType string) (float64, error) {
	now := time.Now()
	entryTime := now.Add(-time.Duration(minutes) * time.Minute)
	exitTime := now
	
	result, err := s.CalculateParkingFee(entryTime, exitTime, spotType)
	if err != nil {
		return 0, err
	}
	
	return result.FinalAmount, nil
}
