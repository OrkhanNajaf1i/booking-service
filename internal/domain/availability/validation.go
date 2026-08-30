// File: internal/domain/availability/validation.go
package availability

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// maxAvailabilityDays – bir sorguda hesablana bilecek maksimum gun sayi.
const maxAvailabilityDays = 62

// ParseClock – "09:00" -> (9, 0). Yalniz HH:MM qebul edir.
func ParseClock(value string) (hour int, minute int, err error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidTime
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, ErrInvalidTime
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, ErrInvalidTime
	}

	return hour, minute, nil
}

// clockMinutes – "09:30" -> 570 (gunun evvelinden deqiqe).
func clockMinutes(value string) (int, error) {
	hour, minute, err := ParseClock(value)
	if err != nil {
		return 0, err
	}
	return hour*60 + minute, nil
}

// FormatClock – 570 -> "09:30".
func FormatClock(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

// ValidateWorkingHours – bir gunun qrafikini yoxlayir.
// Nahar fasilesi yalniz aktiv olanda yoxlanilir: admin onu sondurende
// break_start/break_end qiymetleri sax͏lanilsa da tetbiq edilmir.
func ValidateWorkingHours(req *SetWorkingHoursRequest) error {
	if req == nil {
		return NewError("INVALID_REQUEST", "request bosdur")
	}
	if req.StaffID == uuid.Nil {
		return ErrStaffRequired
	}
	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		return ErrInvalidDayOfWeek
	}

	start, err := clockMinutes(req.StartTime)
	if err != nil {
		return err
	}
	end, err := clockMinutes(req.EndTime)
	if err != nil {
		return err
	}
	if end <= start {
		return ErrInvalidRange
	}

	if !req.BreakEnabled {
		return nil
	}

	if req.BreakStart == nil || req.BreakEnd == nil {
		return NewError("BREAK_INCOMPLETE", "break_enabled true olanda break_start ve break_end teleb olunur")
	}

	breakStart, err := clockMinutes(*req.BreakStart)
	if err != nil {
		return err
	}
	breakEnd, err := clockMinutes(*req.BreakEnd)
	if err != nil {
		return err
	}
	if breakEnd <= breakStart {
		return ErrInvalidRange
	}
	if breakStart < start || breakEnd > end {
		return ErrBreakOutOfRange
	}

	return nil
}

// ValidateSettings – ayar deyerlerinin serhedlerini yoxlayir.
func ValidateSettings(settings *ScheduleSettings) error {
	if settings.SlotStepMins < 5 || settings.SlotStepMins > 480 {
		return ErrInvalidStep
	}
	if settings.DefaultDurationMins < 5 || settings.DefaultDurationMins > 1440 {
		return ErrInvalidDuration
	}
	if settings.BufferBeforeMins < 0 || settings.BufferBeforeMins > 240 {
		return NewError("INVALID_BUFFER", "buffer_before_mins 0-240 araliginda olmalidir")
	}
	if settings.BufferAfterMins < 0 || settings.BufferAfterMins > 240 {
		return NewError("INVALID_BUFFER", "buffer_after_mins 0-240 araliginda olmalidir")
	}
	if settings.MinNoticeMins < 0 || settings.MinNoticeMins > 43200 {
		return NewError("INVALID_MIN_NOTICE", "min_notice_mins 0-43200 araliginda olmalidir")
	}
	if settings.MaxAdvanceDays < 1 || settings.MaxAdvanceDays > 365 {
		return NewError("INVALID_MAX_ADVANCE", "max_advance_days 1-365 araliginda olmalidir")
	}
	for _, window := range []struct {
		name  string
		value int
	}{
		{"pending_expires_mins", settings.PendingExpiresMins},
		{"cancellation_window_mins", settings.CancellationWindowMins},
		{"reschedule_window_mins", settings.RescheduleWindowMins},
	} {
		// 20160 deqiqe = 14 gun. Bundan uzun pencere praktiki deyil.
		if window.value < 0 || window.value > 20160 {
			return NewError("INVALID_WINDOW", window.name+" 0-20160 deqiqe araliginda olmalidir")
		}
	}

	if strings.TrimSpace(settings.Timezone) != "" {
		if _, err := time.LoadLocation(settings.Timezone); err != nil {
			return NewError("INVALID_TIMEZONE", "timezone taninmadi: "+settings.Timezone)
		}
	}
	return nil
}

// ValidateAvailabilityQuery – tarix araligini yoxlayir.
func ValidateAvailabilityQuery(query *AvailabilityQuery) error {
	if query == nil {
		return NewError("INVALID_REQUEST", "request bosdur")
	}
	if query.StaffID == uuid.Nil {
		return ErrStaffRequired
	}
	if query.ToDate.Before(query.FromDate) {
		return ErrInvalidRange
	}
	if query.ToDate.Sub(query.FromDate) > time.Duration(maxAvailabilityDays)*24*time.Hour {
		return ErrRangeTooWide
	}
	return nil
}

// ValidateTimeOff – bloklanmis intervali yoxlayir.
func ValidateTimeOff(req *CreateTimeOffRequest) error {
	if req == nil {
		return NewError("INVALID_REQUEST", "request bosdur")
	}
	if req.StaffID == uuid.Nil {
		return ErrStaffRequired
	}
	if !req.EndAt.After(req.StartAt) {
		return ErrInvalidRange
	}
	return nil
}
