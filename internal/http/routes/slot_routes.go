// File: internal/http/routes/slot_routes.go
package routes

import (
	"net/http"

	slotHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/slot"
)

func RegisterSlotRoutes(
	mux *http.ServeMux,
	handler *slotHandler.SlotHandler,
) {
	mux.HandleFunc("POST /api/v1/slots", handler.CreateSlot)
	mux.HandleFunc("GET /api/v1/slots", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "" {
			handler.GetSlot(w, r)
		} else {
			handler.ListSlots(w, r)
		}
	})
	mux.HandleFunc("PATCH /api/v1/slots", handler.UpdateSlot)
	mux.HandleFunc("DELETE /api/v1/slots", handler.DeleteSlot)
	mux.HandleFunc("GET /api/v1/available-slots", handler.GetAvailableSlots)
	mux.HandleFunc("POST /api/v1/staff/working-hours", handler.SetWorkingHours)
	mux.HandleFunc("GET /api/v1/staff/working-hours", handler.GetStaffWorkingHours)
	mux.HandleFunc("POST /api/v1/admin/slots/generate", handler.GenerateSlots)
}
