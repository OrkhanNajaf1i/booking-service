// File: internal/http/middleware/context.go
//
// Context oxuma helper-leri. Her handler paketinin eyni kodu
// tekrarlamasinin qarsisini alir – bu tekrar artiq bir defe
// customer handler-inde acar uyusmazligina sebeb olmusdu.
package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrNoBusinessID = errors.New("business id konteksti tapilmadi")
	ErrNoUserID     = errors.New("user id konteksti tapilmadi")
)

// BusinessIDFromContext – context-den business_id.
// Handler-lerin hamisi bu bir yerden oxumalidir: tip yalniz burada
// ve AuthMiddleware-de teyin olunur.
func BusinessIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(BusinessKey)
	if value == nil {
		return uuid.Nil, ErrNoBusinessID
	}

	businessID, ok := value.(uuid.UUID)
	if !ok || businessID == uuid.Nil {
		return uuid.Nil, ErrNoBusinessID
	}
	return businessID, nil
}

// BusinessIDFrom – sorgudan business_id.
// Business-i olmayan istifadeci ucun ErrNoBusinessID qaytarir.
func BusinessIDFrom(r *http.Request) (uuid.UUID, error) {
	return BusinessIDFromContext(r.Context())
}

// OptionalBusinessIDFrom – business yoxdursa uuid.Nil qaytarir, xeta vermir.
// Musteri rolundaki istifadeciler ucun.
func OptionalBusinessIDFrom(r *http.Request) uuid.UUID {
	businessID, err := BusinessIDFrom(r)
	if err != nil {
		return uuid.Nil
	}
	return businessID
}

// UserIDFromContext – context-den user_id.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return uuid.Nil, ErrNoUserID
	}

	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, ErrNoUserID
	}
	return userID, nil
}

// UserIDFrom – sorgudan user_id.
func UserIDFrom(r *http.Request) (uuid.UUID, error) {
	return UserIDFromContext(r.Context())
}

// RoleFrom – JWT-den gelen rol. Yoxdursa bos string.
func RoleFrom(r *http.Request) string {
	value := r.Context().Value(RoleKey)
	if value == nil {
		return ""
	}
	role, _ := value.(string)
	return role
}
