// File: internal/domain/availability/service.go
package availability

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// availabilityService – bosluq hesablama muherriki.
type availabilityService struct {
	repo      Repository
	durations ServiceDurationProvider
	log       logger.Logger
	now       func() time.Time // testde sabitlemek ucun
}

// NewService – hesablama servisi yaradir.
// durations nil ola biler; bu halda hemise default mueddet islenir.
func NewService(repo Repository, durations ServiceDurationProvider, log logger.Logger) Service {
	return &availabilityService{
		repo:      repo,
		durations: durations,
		log:       log,
		now:       time.Now,
	}
}

// ============================================================
// SETTINGS RESOLUTION
// ============================================================

// ResolveSettings – uc pilleli zencir:
//
//  1. hemin isciye mexsus override
//  2. biznesin default ayari
//  3. sistem default-u
func (s *availabilityService) ResolveSettings(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
) (*ScheduleSettings, error) {
	if staffID != uuid.Nil {
		staffSettings, err := s.repo.GetSettings(ctx, businessID, &staffID)
		if err != nil {
			return nil, fmt.Errorf("staff ayarlari oxunmadi: %w", err)
		}
		if staffSettings != nil {
			return staffSettings, nil
		}
	}

	businessSettings, err := s.repo.GetSettings(ctx, businessID, nil)
	if err != nil {
		return nil, fmt.Errorf("biznes ayarlari oxunmadi: %w", err)
	}
	if businessSettings != nil {
		return businessSettings, nil
	}

	return DefaultSettings(businessID), nil
}

// ============================================================
// AVAILABILITY ENGINE
// ============================================================

// GetAvailability – verilmis tarix araliginda iscinin bosluqlarini hesablayir.
func (s *availabilityService) GetAvailability(
	ctx context.Context,
	businessID uuid.UUID,
	query *AvailabilityQuery,
) (*AvailabilityResult, error) {
	if err := ValidateAvailabilityQuery(query); err != nil {
		return nil, err
	}

	settings, err := s.ResolveSettings(ctx, businessID, query.StaffID)
	if err != nil {
		return nil, err
	}
	loc := settings.Location()

	duration, err := s.resolveDuration(ctx, businessID, query.ServiceID, settings)
	if err != nil {
		return nil, err
	}

	// Heftelik qrafiki gun -> qayda seklinde yigiriq.
	weekly, err := s.repo.ListWorkingHours(ctx, businessID, query.StaffID)
	if err != nil {
		return nil, fmt.Errorf("is saatlari oxunmadi: %w", err)
	}
	byDay := make(map[int]*WorkingHours, len(weekly))
	for _, wh := range weekly {
		byDay[wh.DayOfWeek] = wh
	}

	// Hesablama penceresi: ilk gunun 00:00-i ile son gunun 24:00-i arasi.
	firstDay := startOfDay(query.FromDate, loc)
	lastDay := startOfDay(query.ToDate, loc)
	rangeStart := firstDay
	rangeEnd := lastDay.AddDate(0, 0, 1)

	busy, err := s.collectBusy(ctx, businessID, query.StaffID, rangeStart, rangeEnd, settings)
	if err != nil {
		return nil, err
	}

	now := s.now().In(loc)
	earliest := now.Add(time.Duration(settings.MinNoticeMins) * time.Minute)
	horizon := startOfDay(now, loc).AddDate(0, 0, settings.MaxAdvanceDays+1)

	result := &AvailabilityResult{
		StaffID:      query.StaffID,
		ServiceID:    query.ServiceID,
		Timezone:     settings.Timezone,
		DurationMins: duration,
		SlotStepMins: settings.SlotStepMins,
		Days:         make([]DayAvailability, 0, 8),
	}

	for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
		result.Days = append(result.Days, s.buildDay(
			day, loc, byDay[int(day.Weekday())],
			settings, duration, busy, now, earliest, horizon,
		))
	}

	return result, nil
}

// buildDay – bir gunun bosluqlarini qurur.
func (s *availabilityService) buildDay(
	day time.Time,
	loc *time.Location,
	workingHours *WorkingHours,
	settings *ScheduleSettings,
	durationMins int,
	busy []BusyInterval,
	now, earliest, horizon time.Time,
) DayAvailability {
	dayResult := DayAvailability{
		Date:      day.Format("2006-01-02"),
		DayOfWeek: int(day.Weekday()),
		IsWorkday: false,
		Slots:     []TimeSlot{},
	}

	if workingHours == nil || !workingHours.IsActive {
		return dayResult
	}

	openMins, err := clockMinutes(workingHours.StartTime)
	if err != nil {
		s.log.Warn("Yanlis is saati formati",
			logger.Field{Key: "staff_id", Value: workingHours.StaffID.String()},
			logger.Field{Key: "start_time", Value: workingHours.StartTime},
		)
		return dayResult
	}
	closeMins, err := clockMinutes(workingHours.EndTime)
	if err != nil || closeMins <= openMins {
		return dayResult
	}

	dayResult.IsWorkday = true
	dayResult.OpensAt = FormatClock(openMins)
	dayResult.ClosesAt = FormatClock(closeMins)

	// Is penceresi. Nahar fasilesi aktivdirsa pencere ikiye bolunur.
	windows := [][2]int{{openMins, closeMins}}

	if workingHours.BreakEnabled && workingHours.BreakStart != nil && workingHours.BreakEnd != nil {
		breakStart, errStart := clockMinutes(*workingHours.BreakStart)
		breakEnd, errEnd := clockMinutes(*workingHours.BreakEnd)

		if errStart == nil && errEnd == nil && breakEnd > breakStart &&
			breakStart >= openMins && breakEnd <= closeMins {

			dayResult.Break = &BreakInfo{
				Start: FormatClock(breakStart),
				End:   FormatClock(breakEnd),
			}

			windows = windows[:0]
			if breakStart > openMins {
				windows = append(windows, [2]int{openMins, breakStart})
			}
			if closeMins > breakEnd {
				windows = append(windows, [2]int{breakEnd, closeMins})
			}
		}
	}

	duration := time.Duration(durationMins) * time.Minute
	step := time.Duration(settings.SlotStepMins) * time.Minute

	for _, window := range windows {
		// time.Date ile qururuq (day.Add deyil): saat deyisimi olan gunlerde
		// "09:00" hemise yerli 09:00 qalmalidir.
		windowStart := atClock(day, window[0], loc)
		windowEnd := atClock(day, window[1], loc)

		// Randevu pencerenin icinde tam yerlesmelidir; yarimciq slot verilmir.
		for cursor := windowStart; !cursor.Add(duration).After(windowEnd); cursor = cursor.Add(step) {
			slotEnd := cursor.Add(duration)
			state := classify(cursor, slotEnd, busy, now, earliest, horizon)

			dayResult.Slots = append(dayResult.Slots, TimeSlot{
				Start:        cursor,
				End:          slotEnd,
				DurationMins: durationMins,
				State:        state,
				Available:    state == SlotStateAvailable,
			})
		}
	}

	return dayResult
}

// classify – bir namized intervalin veziyyetini teyin edir.
// Siralanma vacibdir: kecmis > horizon > min-notice > tutulmus.
func classify(
	start, end time.Time,
	busy []BusyInterval,
	now, earliest, horizon time.Time,
) SlotState {
	// Basi kecmis slot bron edile bilmez – hetta hele bitmeyibse de.
	if !start.After(now) {
		return SlotStatePast
	}
	if !start.Before(horizon) {
		return SlotStateTooFar
	}

	for _, interval := range busy {
		if interval.Overlaps(start, end) {
			if interval.Reason == "time_off" {
				return SlotStateBlocked
			}
			return SlotStateBooked
		}
	}

	if start.Before(earliest) {
		return SlotStateTooSoon
	}

	return SlotStateAvailable
}

// collectBusy – booking-leri (bufer ile genislendirilmis) ve time-off-lari birlesdirir.
//
// Bufer mentiqi: movcud randevu E ile yeni randevu N arasinda en azi
// max(buffer_before, buffer_after) deqiqe bosluq qalmalidir. Buna gore
// E-ni her iki terefden hemin qeder genislendirmek kifayetdir.
func (s *availabilityService) collectBusy(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	from, to time.Time,
	settings *ScheduleSettings,
) ([]BusyInterval, error) {
	pad := settings.BufferBeforeMins
	if settings.BufferAfterMins > pad {
		pad = settings.BufferAfterMins
	}
	padding := time.Duration(pad) * time.Minute

	// Serhedde duran randevular da tesir edir deye araligi bufer qeder genisledirik.
	booked, err := s.repo.ListBookedIntervals(ctx, businessID, staffID, from.Add(-padding), to.Add(padding))
	if err != nil {
		return nil, fmt.Errorf("bron edilmis vaxtlar oxunmadi: %w", err)
	}

	busy := make([]BusyInterval, 0, len(booked)+4)
	for _, interval := range booked {
		busy = append(busy, BusyInterval{
			Start:  interval.Start.Add(-padding),
			End:    interval.End.Add(padding),
			Reason: "booking",
		})
	}

	timeOff, err := s.repo.ListTimeOff(ctx, businessID, staffID, from, to)
	if err != nil {
		return nil, fmt.Errorf("time-off oxunmadi: %w", err)
	}
	for _, off := range timeOff {
		busy = append(busy, BusyInterval{
			Start:  off.StartAt,
			End:    off.EndAt,
			Reason: "time_off",
		})
	}

	return busy, nil
}

// resolveDuration – xidmet secilibse onun mueddeti, yoxsa default.
func (s *availabilityService) resolveDuration(
	ctx context.Context,
	businessID uuid.UUID,
	serviceID *uuid.UUID,
	settings *ScheduleSettings,
) (int, error) {
	if serviceID == nil || *serviceID == uuid.Nil || s.durations == nil {
		return settings.DefaultDurationMins, nil
	}

	duration, err := s.durations.GetDurationMins(ctx, businessID, *serviceID)
	if err != nil {
		return 0, NewError("SERVICE_NOT_FOUND", "xidmet tapilmadi")
	}
	if duration <= 0 {
		return settings.DefaultDurationMins, nil
	}
	return duration, nil
}

// ============================================================
// SLOT CHECK (booking yaratmadan evvel)
// ============================================================

// CheckSlot – musterinin sectiyi vaxtin hele de etibarli oldugunu tesdiqleyir.
//
// Hesablamani GetAvailability-nin ozune tapsiririq ki, ekranda gorunen
// ile qebul edilen arasinda ferq olmasin. Bu hem de ixtiyari vaxt
// gonderilmesinin qarsisini alir: yalniz grid uzerindeki vaxtlar tapilir.
func (s *availabilityService) CheckSlot(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	serviceID *uuid.UUID,
	start time.Time,
) (*TimeSlot, error) {
	settings, err := s.ResolveSettings(ctx, businessID, staffID)
	if err != nil {
		return nil, err
	}
	loc := settings.Location()
	localStart := start.In(loc)

	result, err := s.GetAvailability(ctx, businessID, &AvailabilityQuery{
		StaffID:   staffID,
		ServiceID: serviceID,
		FromDate:  localStart,
		ToDate:    localStart,
	})
	if err != nil {
		return nil, err
	}

	for _, day := range result.Days {
		for _, slot := range day.Slots {
			if !slot.Start.Equal(start) {
				continue
			}

			switch slot.State {
			case SlotStateAvailable:
				matched := slot
				return &matched, nil
			case SlotStateBooked:
				return nil, NewError("SLOT_TAKEN", "bu vaxt artiq bron olunub")
			case SlotStateBlocked:
				return nil, NewError("SLOT_BLOCKED", "bu vaxt bagli elan olunub")
			case SlotStatePast:
				return nil, NewError("SLOT_PAST", "bu vaxt artiq kecib")
			case SlotStateTooSoon:
				return nil, NewError("SLOT_TOO_SOON", fmt.Sprintf(
					"bron en azi %d deqiqe evvelceden edilmelidir", settings.MinNoticeMins))
			case SlotStateTooFar:
				return nil, NewError("SLOT_TOO_FAR", fmt.Sprintf(
					"bron maksimum %d gun ireli acila biler", settings.MaxAdvanceDays))
			}
		}
	}

	return nil, NewError("SLOT_NOT_ON_GRID", "secilen vaxt is qrafikine uygun deyil")
}

// ============================================================
// ADMIN: WORKING HOURS
// ============================================================

func (s *availabilityService) SetWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	req *SetWorkingHoursRequest,
) (*WorkingHours, error) {
	if err := ValidateWorkingHours(req); err != nil {
		return nil, err
	}

	now := s.now()
	workingHours := &WorkingHours{
		ID:           uuid.New(),
		BusinessID:   businessID,
		StaffID:      req.StaffID,
		DayOfWeek:    req.DayOfWeek,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		BreakEnabled: req.BreakEnabled,
		BreakStart:   req.BreakStart,
		BreakEnd:     req.BreakEnd,
		IsActive:     req.IsActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.UpsertWorkingHours(ctx, workingHours); err != nil {
		return nil, fmt.Errorf("is saatlari yazilmadi: %w", err)
	}

	s.log.Info("Is saatlari yenilendi",
		logger.Field{Key: "staff_id", Value: req.StaffID.String()},
		logger.Field{Key: "day_of_week", Value: req.DayOfWeek},
		logger.Field{Key: "break_enabled", Value: req.BreakEnabled},
	)

	return workingHours, nil
}

// BulkSetWorkingHours – butun hefteni bir sorguda yazir.
// Evvelce hamisi validasiyadan kecir; bir gun sehvdirse hec biri yazilmir.
func (s *availabilityService) BulkSetWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	req *BulkWorkingHoursRequest,
) ([]*WorkingHours, error) {
	if req == nil || len(req.Days) == 0 {
		return nil, NewError("INVALID_REQUEST", "days bosdur")
	}

	seen := make(map[int]bool, len(req.Days))
	for i := range req.Days {
		if req.Days[i].StaffID == uuid.Nil {
			req.Days[i].StaffID = req.StaffID
		}
		if err := ValidateWorkingHours(&req.Days[i]); err != nil {
			return nil, err
		}
		if seen[req.Days[i].DayOfWeek] {
			return nil, NewError("DUPLICATE_DAY", fmt.Sprintf(
				"day_of_week %d tekrarlanib", req.Days[i].DayOfWeek))
		}
		seen[req.Days[i].DayOfWeek] = true
	}

	saved := make([]*WorkingHours, 0, len(req.Days))
	for i := range req.Days {
		workingHours, err := s.SetWorkingHours(ctx, businessID, &req.Days[i])
		if err != nil {
			return nil, err
		}
		saved = append(saved, workingHours)
	}

	return saved, nil
}

func (s *availabilityService) ListWorkingHours(
	ctx context.Context,
	businessID, staffID uuid.UUID,
) ([]*WorkingHours, error) {
	if staffID == uuid.Nil {
		return nil, ErrStaffRequired
	}
	return s.repo.ListWorkingHours(ctx, businessID, staffID)
}

func (s *availabilityService) DeleteWorkingHours(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	dayOfWeek int,
) error {
	if staffID == uuid.Nil {
		return ErrStaffRequired
	}
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return ErrInvalidDayOfWeek
	}
	return s.repo.DeleteWorkingHours(ctx, businessID, staffID, dayOfWeek)
}

// ============================================================
// ADMIN: SETTINGS
// ============================================================

func (s *availabilityService) GetSettings(
	ctx context.Context,
	businessID uuid.UUID,
	staffID *uuid.UUID,
) (*ScheduleSettings, error) {
	settings, err := s.repo.GetSettings(ctx, businessID, staffID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		// Hele yazilmayib – effektiv deyerleri qaytaririq ki, UI bos qalmasin.
		if staffID != nil && *staffID != uuid.Nil {
			return s.ResolveSettings(ctx, businessID, *staffID)
		}
		return DefaultSettings(businessID), nil
	}
	return settings, nil
}

// UpdateSettings – yalniz gonderilen saheler deyisir (partial update).
func (s *availabilityService) UpdateSettings(
	ctx context.Context,
	businessID uuid.UUID,
	req *UpdateScheduleSettingsRequest,
) (*ScheduleSettings, error) {
	if req == nil {
		return nil, NewError("INVALID_REQUEST", "request bosdur")
	}

	existing, err := s.repo.GetSettings(ctx, businessID, req.StaffID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		existing = DefaultSettings(businessID)
		existing.ID = uuid.New()
		existing.StaffID = req.StaffID
		existing.CreatedAt = s.now()
	}

	if req.Timezone != nil {
		existing.Timezone = *req.Timezone
	}
	if req.SlotStepMins != nil {
		existing.SlotStepMins = *req.SlotStepMins
	}
	if req.DefaultDurationMins != nil {
		existing.DefaultDurationMins = *req.DefaultDurationMins
	}
	if req.BufferBeforeMins != nil {
		existing.BufferBeforeMins = *req.BufferBeforeMins
	}
	if req.BufferAfterMins != nil {
		existing.BufferAfterMins = *req.BufferAfterMins
	}
	if req.MinNoticeMins != nil {
		existing.MinNoticeMins = *req.MinNoticeMins
	}
	if req.MaxAdvanceDays != nil {
		existing.MaxAdvanceDays = *req.MaxAdvanceDays
	}
	if req.AutoConfirm != nil {
		existing.AutoConfirm = *req.AutoConfirm
	}
	if req.AllowRescheduleProposal != nil {
		existing.AllowRescheduleProposal = *req.AllowRescheduleProposal
	}
	if req.PendingExpiresMins != nil {
		existing.PendingExpiresMins = *req.PendingExpiresMins
	}
	if req.CancellationWindowMins != nil {
		existing.CancellationWindowMins = *req.CancellationWindowMins
	}
	if req.AllowCustomerReschedule != nil {
		existing.AllowCustomerReschedule = *req.AllowCustomerReschedule
	}
	if req.RescheduleWindowMins != nil {
		existing.RescheduleWindowMins = *req.RescheduleWindowMins
	}

	if err := ValidateSettings(existing); err != nil {
		return nil, err
	}

	existing.BusinessID = businessID
	existing.UpdatedAt = s.now()
	if existing.ID == uuid.Nil {
		existing.ID = uuid.New()
	}

	if err := s.repo.UpsertSettings(ctx, existing); err != nil {
		return nil, fmt.Errorf("ayarlar yazilmadi: %w", err)
	}

	s.log.Info("Qrafik ayarlari yenilendi",
		logger.Field{Key: "business_id", Value: businessID.String()},
		logger.Field{Key: "slot_step_mins", Value: existing.SlotStepMins},
		logger.Field{Key: "auto_confirm", Value: existing.AutoConfirm},
	)

	return existing, nil
}

// ============================================================
// ADMIN: TIME OFF
// ============================================================

func (s *availabilityService) CreateTimeOff(
	ctx context.Context,
	businessID uuid.UUID,
	req *CreateTimeOffRequest,
) (*TimeOff, error) {
	if err := ValidateTimeOff(req); err != nil {
		return nil, err
	}

	now := s.now()
	timeOff := &TimeOff{
		ID:         uuid.New(),
		BusinessID: businessID,
		StaffID:    req.StaffID,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Reason:     req.Reason,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.CreateTimeOff(ctx, timeOff); err != nil {
		return nil, fmt.Errorf("time-off yaradilmadi: %w", err)
	}
	return timeOff, nil
}

func (s *availabilityService) ListTimeOff(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	from, to time.Time,
) ([]*TimeOff, error) {
	if staffID == uuid.Nil {
		return nil, ErrStaffRequired
	}
	return s.repo.ListTimeOff(ctx, businessID, staffID, from, to)
}

func (s *availabilityService) DeleteTimeOff(ctx context.Context, businessID, timeOffID uuid.UUID) error {
	return s.repo.DeleteTimeOff(ctx, businessID, timeOffID)
}

// ============================================================
// HELPERS
// ============================================================

// startOfDay – verilmis vaxtin hemin timezone-dakı 00:00-i.
func startOfDay(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

// atClock – gunun tarixi + gun evvelinden verilmis deqiqe = yerli vaxt.
func atClock(day time.Time, minutesFromMidnight int, loc *time.Location) time.Time {
	return time.Date(
		day.Year(), day.Month(), day.Day(),
		minutesFromMidnight/60, minutesFromMidnight%60, 0, 0,
		loc,
	)
}
