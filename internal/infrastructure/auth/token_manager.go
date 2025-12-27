package auth

import (
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTTokenManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTTokenManager(secretKey string) *JWTTokenManager {
	return &JWTTokenManager{
		secretKey:     secretKey,
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (m *JWTTokenManager) GenerateAccessToken(claims *auth.JWTClaims) (string, error) {
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     claims.UserID.String(),
		"email":       claims.Email,
		"role":        string(claims.Role),
		"business_id": claims.BusinessID.String(),
		"is_owner":    claims.IsOwner,
		"exp":         time.Now().Add(m.accessExpiry).Unix(),
		"iat":         time.Now().Unix(),
	})
	return tk.SignedString([]byte(m.secretKey))
}

func (m *JWTTokenManager) ValidateAccessToken(tokenString string) (*auth.JWTClaims, error) {
	tk, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(m.secretKey), nil
	})
	if err != nil || !tk.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claimsMap, ok := tk.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	userID, _ := uuid.Parse((*claimsMap)["user_id"].(string))
	businessID, _ := uuid.Parse((*claimsMap)["business_id"].(string))
	return &auth.JWTClaims{
		UserID:     userID,
		Email:      (*claimsMap)["email"].(string),
		Role:       auth.UserRole((*claimsMap)["role"].(string)),
		BusinessID: businessID,
		IsOwner:    (*claimsMap)["is_owner"].(bool),
		ExpiresAt:  int64((*claimsMap)["exp"].(float64)),
	}, nil
}
