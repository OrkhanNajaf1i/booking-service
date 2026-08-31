// File: internal/domain/availability/entity.go
//
// Bu paket "hansi vaxtlar bosdur" sualinin yegane cavabdehidir.
// Bazada onceden slot setirleri saxlanmir; bosluqlar ucu-uca gelen
// uc qaydadan runtime-da hesablanir:
//
//	WorkingHours     – hansi gun, saat necede-necede, nahar fasilesi harada
//	ScheduleSettings – secim addimi (16/30/60 deq), bufer, min xeberdarliq
//	TimeOff          – mezuniyyet / bloklanmis interval
//
// Admin bu qaydalari deyisen kimi netice deyisir; hec ne yenidien
// generasiya olunmur.
package availability

import (
	"time"

	"github.com/google/uuid"
)

// ============================================
// WORKING HOURS
// ============================================

// WorkingHours – bir iscinin heftenin bir gunundeki is qrafiki.
// StartTime/EndTime/BreakStart/BreakEnd "HH:MM" formatindadir ve
// ScheduleSettings.Timezone-a gore serh olunur.
type WorkingHours struct {
	ID         uuid.UUID `db:"id"           json:"id"`
	BusinessID uuid.UUID `db:"business_id"  json:"business_id"`
	StaffID    uuid.UUID `db:"staff_id"     json:"staff_id"`
	DayOfWeek  int       `db:"day_of_week"  json:"day_of_week"` // 0=Bazar ... 6=Senbe

	StartTime string `db:"start_time" json:"start_time"` // "09:00"
	EndTime   string `db:"end_time"   json:"end_time"`   // "18:00"

	// Nahar fasilesi. BreakEnabled=false olsa fasile tetbiq edilmir
	// (admin "nahar fasilesi" secimini disabled ede biler).
	BreakEnabled bool    `db:"break_enabled" json:"break_enabled"`
	BreakStart   *string `db:"break_start"   json:"break_start,omitempty"` // "13:00"
	BreakEnd     *string `db:"break_end"     json:"break_end,omitempty"`   // "14:00"

	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================
// SCHEDULE SETTINGS
// ============================================

// ScheduleSettings – bron pencerelerinin necə kesileceyini teyin edir.
// StaffID nil olanda bu, biznesin default ayaridir.
type ScheduleSettings struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	BusinessID uuid.UUID  `db:"business_id" json:"business_id"`
	StaffID    *uuid.UUID `db:"staff_id"    json:"staff_id,omitempty"`

	Timezone string `db:"timezone" json:"timezone"`

	// Musteriye gosterilen vaxtlarin addimi.
	// 60 -> 09:00, 10:00, 11:00 ...
	// 16 -> 09:00, 09:16, 09:32 ...
	SlotStepMins int `db:"slot_step_mins" json:"slot_step_mins"`

	// Xidmet secilmeyibse randevunun uzunlugu.
	DefaultDurationMins int `db:"default_duration_mins" json:"default_duration_mins"`

	// Randevudan evvel/sonra saxlanilan bos vaxt.
	BufferBeforeMins int `db:"buffer_before_mins" json:"buffer_before_mins"`
	BufferAfterMins  int `db:"buffer_after_mins"  json:"buffer_after_mins"`

	// Bron en azi bu qeder deqiqe evvelceden edile biler.
	MinNoticeMins int `db:"min_notice_mins" json:"min_notice_mins"`
	// Ne qeder ireli tarixe bron acilsin.
	MaxAdvanceDays int `db:"max_advance_days" json:"max_advance_days"`

	// TRUE olsa booking "pending" yox, birbasa "confirmed" yaranir.
	AutoConfirm bool `db:"auto_confirm" json:"auto_confirm"`
	// Provider musteriye alternativ vaxt teklif ede bilsin?
	AllowRescheduleProposal bool `db:"allow_reschedule_proposal" json:"allow_reschedule_proposal"`

	// ---------- RANDEVU SIYASETI ----------

	// PendingExpiresMins – provider bu muddet erzinde cavab vermese bron
	// avtomatik legv olunur ve slot azad olur. 0 = sondurulub.
	PendingExpiresMins int `db:"pending_expires_mins" json:"pending_expires_mins"`

	// CancellationWindowMins – randevuya bu qeder vaxt qalanda musteri
	// artiq legv ede bilmez. 0 = istenilen vaxt legv etmek olar.
	CancellationWindowMins int `db:"cancellation_window_mins" json:"cancellation_window_mins"`

	// AllowCustomerReschedule – musteri ozu vaxti deyise bilsin?
	AllowCustomerReschedule bool `db:"allow_customer_reschedule" json:"allow_customer_reschedule"`

	// RescheduleWindowMins – vaxt deyismek ucun son muddet.
	RescheduleWindowMins int `db:"reschedule_window_mins" json:"reschedule_window_mins"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// DefaultSettings – ne staff, ne de biznes ayari yoxdursa istifade olunur.
func DefaultSettings(businessID uuid.UUID) *ScheduleSettings {
	return &ScheduleSettings{
		BusinessID:              businessID,
		Timezone:                "Asia/Baku",
		SlotStepMins:            30,
		DefaultDurationMins:     30,
		BufferBeforeMins:        0,
		BufferAfterMins:         0,
		MinNoticeMins:           60,
		MaxAdvanceDays:          30,
		AutoConfirm:             false,
		AllowRescheduleProposal: true,

		// Sahe standarti (Booksy/Fresha): 24 saat.
		PendingExpiresMins:      1440,
		CancellationWindowMins:  1440,
		AllowCustomerReschedule: true,
		RescheduleWindowMins:    1440,
	}
}

// Location – ayarlardaki timezone-u time.Location-a cevirir.
// Tanınmayan timezone UTC-ye dusur, cunki hesablama dayanmamalidir.
func (s *ScheduleSettings) Location() *time.Location {
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil || loc == nil {
		return time.UTC
	}
	return loc
}

// ============================================
// TIME OFF
// ============================================

// TimeOff – iscinin tam bagli oldugu interval (mezuniyyet, xestelik).
type TimeOff struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
	StaffID    uuid.UUID `db:"staff_id"    json:"staff_id"`
	StartAt    time.Time `db:"start_at"    json:"start_at"`
	EndAt      time.Time `db:"end_at"      json:"end_at"`
	Reason     string    `db:"reason"      json:"reason"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"  json:"updated_at"`
}

// ============================================
// BUSY INTERVAL – hesablama ucun daxili tip
// ============================================

// BusyInterval – artiq tutulmus vaxt. Hem booking-lerden,
// hem de time-off-dan gelir.
type BusyInterval struct {
	Start  time.Time
	End    time.Time
	Reason string // "booking" | "time_off"
}

// Overlaps – iki interval kesisirmi? Sərhəd toxunması kesisme sayilmir:
// [09:00,09:30) ile [09:30,10:00) kesismir.
func (b BusyInterval) Overlaps(start, end time.Time) bool {
	return start.Before(b.End) && end.After(b.Start)
}

// ============================================
// COMPUTED OUTPUT
// ============================================

// SlotState – hesablanmis vaxtin veziyyeti.
type SlotState string

const (
	SlotStateAvailable SlotState = "available" // secile biler
	SlotStateBooked    SlotState = "booked"    // bron olunub
	SlotStateBlocked   SlotState = "blocked"   // mezuniyyet / bagli
	SlotStatePast      SlotState = "past"      // vaxti kecib
	SlotStateTooSoon   SlotState = "too_soon"  // min_notice penceresine dusur
	SlotStateTooFar    SlotState = "too_far"   // max_advance_days xaricindedir
)

// TimeSlot – musteriye gosterilen bir secim.
type TimeSlot struct {
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	DurationMins int       `json:"duration_mins"`
	State        SlotState `json:"state"`
	Available    bool      `json:"available"`
}

// DayAvailability – bir gunun tam menzeresi.
type DayAvailability struct {
	Date      string     `json:"date"` // "2026-09-01"
	DayOfWeek int        `json:"day_of_week"`
	IsWorkday bool       `json:"is_workday"`
	OpensAt   string     `json:"opens_at,omitempty"`  // "09:00"
	ClosesAt  string     `json:"closes_at,omitempty"` // "18:00"
	Break     *BreakInfo `json:"break,omitempty"`
	Slots     []TimeSlot `json:"slots"`
}

// BreakInfo – gunun nahar fasilesi (aktivdirsə).
type BreakInfo struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// AvailabilityResult – endpoint-in tam cavabı.
type AvailabilityResult struct {
	StaffID      uuid.UUID         `json:"staff_id"`
	ServiceID    *uuid.UUID        `json:"service_id,omitempty"`
	Timezone     string            `json:"timezone"`
	DurationMins int               `json:"duration_mins"`
	SlotStepMins int               `json:"slot_step_mins"`
	Days         []DayAvailability `json:"days"`
}

// ============================================
// REQUEST DTOS
// ============================================

// SetWorkingHoursRequest – bir gunun qrafikini yazir/yenileyir.
type SetWorkingHoursRequest struct {
	StaffID      uuid.UUID `json:"staff_id"`
	DayOfWeek    int       `json:"day_of_week"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	BreakEnabled bool      `json:"break_enabled"`
	BreakStart   *string   `json:"break_start,omitempty"`
	BreakEnd     *string   `json:"break_end,omitempty"`
	IsActive     bool      `json:"is_active"`
}

// BulkWorkingHoursRequest – butun hefteni bir sorguda yazmaq ucun.
type BulkWorkingHoursRequest struct {
	StaffID uuid.UUID                `json:"staff_id"`
	Days    []SetWorkingHoursRequest `json:"days"`
}

// UpdateScheduleSettingsRequest – gonderilmeyen sahe deyismir.
type UpdateScheduleSettingsRequest struct {
	StaffID *uuid.UUID `json:"staff_id,omitempty"`

	Timezone                *string `json:"timezone,omitempty"`
	SlotStepMins            *int    `json:"slot_step_mins,omitempty"`
	DefaultDurationMins     *int    `json:"default_duration_mins,omitempty"`
	BufferBeforeMins        *int    `json:"buffer_before_mins,omitempty"`
	BufferAfterMins         *int    `json:"buffer_after_mins,omitempty"`
	MinNoticeMins           *int    `json:"min_notice_mins,omitempty"`
	MaxAdvanceDays          *int    `json:"max_advance_days,omitempty"`
	AutoConfirm             *bool   `json:"auto_confirm,omitempty"`
	AllowRescheduleProposal *bool   `json:"allow_reschedule_proposal,omitempty"`

	PendingExpiresMins      *int  `json:"pending_expires_mins,omitempty"`
	CancellationWindowMins  *int  `json:"cancellation_window_mins,omitempty"`
	AllowCustomerReschedule *bool `json:"allow_customer_reschedule,omitempty"`
	RescheduleWindowMins    *int  `json:"reschedule_window_mins,omitempty"`
}

// CreateTimeOffRequest – bloklanmis interval yaradir.
type CreateTimeOffRequest struct {
	StaffID uuid.UUID `json:"staff_id"`
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
	Reason  string    `json:"reason"`
}

// AvailabilityQuery – bosluq hesablamasinin girisi.
type AvailabilityQuery struct {
	StaffID   uuid.UUID
	ServiceID *uuid.UUID
	FromDate  time.Time // gun deqiqliyinde
	ToDate    time.Time // daxil
}

// ============================================
// ERRORS
// ============================================

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

func NewError(code, message string) *Error { return &Error{Code: code, Message: message} }

var (
	ErrStaffRequired    = NewError("STAFF_REQUIRED", "staff_id teleb olunur")
	ErrInvalidDayOfWeek = NewError("INVALID_DAY_OF_WEEK", "day_of_week 0-6 araliginda olmalidir")
	ErrInvalidTime      = NewError("INVALID_TIME", "vaxt HH:MM formatinda olmalidir")
	ErrInvalidRange     = NewError("INVALID_RANGE", "bitis vaxti baslangicdan sonra olmalidir")
	ErrBreakOutOfRange  = NewError("BREAK_OUT_OF_RANGE", "nahar fasilesi is saatlari icinde olmalidir")
	ErrInvalidStep      = NewError("INVALID_STEP", "slot_step_mins 5-480 araliginda olmalidir")
	ErrInvalidDuration  = NewError("INVALID_DURATION", "muddet 5-1440 deqiqe araliginda olmalidir")
	ErrRangeTooWide     = NewError("RANGE_TOO_WIDE", "tarix araligi cox genisdir (maksimum 62 gun)")
	ErrNotFound         = NewError("NOT_FOUND", "tapilmadi")
)
