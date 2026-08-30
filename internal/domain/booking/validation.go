// File: internal/domain/booking/validation.go
package booking

import (
	"github.com/google/uuid"
)

// maxNotesLength – qeyd sahesinin limiti.
const maxNotesLength = 1000

// validateCreateRequest – bron sorgusunun ilkin yoxlanisi.
// Vaxtin qrafike uygunlugu burada yox, availability muherrikinde yoxlanilir.
func validateCreateRequest(req *CreateBookingRequest) error {
	if req == nil {
		return NewBookingError("INVALID_REQUEST", "request bosdur")
	}
	if req.CustomerID == uuid.Nil {
		return ErrCustomerRequired
	}
	if req.StaffID == uuid.Nil {
		return ErrStaffRequired
	}
	if req.StartTime.IsZero() {
		return ErrStartRequired
	}
	if len(req.Notes) > maxNotesLength {
		return NewBookingError("NOTES_TOO_LONG", "qeyd 1000 simvoldan uzun ola bilmez")
	}
	return nil
}
