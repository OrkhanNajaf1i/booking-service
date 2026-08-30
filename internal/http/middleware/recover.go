// File: internal/http/middleware/recover.go
package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

// RecoverMiddleware – handler panic edende serveri qorumaq ucun.
//
// Panic bas verende net/http baglantini sadece qirir: client "failed"
// gorur, amma sebebi bilmir ve xeta cavabi almir. Bu middleware panic-i
// tutur, tam stack trace-i loglayir ve client-e duzgun 500 qaytarir.
//
// Bu, handler-lerdeki sehvleri gizletmek ucun deyil – log-da hemise
// tam trace qalir; meqsed bir handler-in butun sorgunu ucurmamasidir.
func RecoverMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				// http.ErrAbortHandler qesdli dayandirmadir – onu udmuruq.
				if err, ok := recovered.(error); ok && err == http.ErrAbortHandler {
					panic(recovered)
				}

				log.Error("Handler panic",
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "method", Value: r.Method},
					logger.Field{Key: "panic", Value: recovered},
					logger.Field{Key: "stack", Value: string(debug.Stack())},
				)

				// Cavab artiq baslamisdirsa basliq yaza bilmerik.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"code":    "INTERNAL_ERROR",
					"message": "Daxili xeta bas verdi",
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
