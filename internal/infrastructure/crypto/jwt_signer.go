// File: internal/infrastructure/crypto/jwt_signer.go
package crypto

import (
	"errors"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTSigner struct {
	secretKey []byte
}

func NewJWTSigner(secret string) *JWTSigner {
	return &JWTSigner{
		secretKey: []byte(secret),
	}
}

func (j *JWTSigner) GenerateAccessToken(claims *auth.JWTClaims) (string, error) {
	var businessIDStr string
	if claims.BusinessID != nil {
		businessIDStr = claims.BusinessID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     claims.UserID.String(),
		"email":       claims.Email,
		"role":        string(claims.Role),
		"business_id": businessIDStr,
		"is_owner":    claims.IsOwner,
		"exp":         claims.ExpiresAt,
		"iat":         time.Now().Unix(),
	})

	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}

func (j *JWTSigner) GenerateRefreshToken() (string, error) {
	return uuid.NewString(), nil
}

func (j *JWTSigner) ValidateAccessToken(tokenString string) (*auth.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claimsMap, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, _ := claimsMap["user_id"].(string)
		roleStr, _ := claimsMap["role"].(string)
		businessIDStr, _ := claimsMap["business_id"].(string)
		emailStr, _ := claimsMap["email"].(string)
		isOwner, _ := claimsMap["is_owner"].(bool)
		exp, _ := claimsMap["exp"].(float64)

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user id")
		}

		var bIDPtr *uuid.UUID
		if businessIDStr != "" {
			parsedBID, err := uuid.Parse(businessIDStr)
			if err == nil && parsedBID != uuid.Nil {
				bIDPtr = &parsedBID
			}
		}

		return &auth.JWTClaims{
			UserID:     userID,
			Email:      emailStr,
			Role:       auth.UserRole(roleStr),
			BusinessID: bIDPtr,
			IsOwner:    isOwner,
			ExpiresAt:  int64(exp),
		}, nil
	}

	return nil, errors.New("invalid token claims")
}
