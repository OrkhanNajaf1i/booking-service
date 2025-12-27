package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost ...int) *BcryptPasswordHasher {
	c := bcrypt.DefaultCost
	if len(cost) > 0 {
		c = cost[0]
		if c < bcrypt.MinCost {
			c = bcrypt.MinCost
		} else if c > bcrypt.MaxCost {
			c = bcrypt.MaxCost
		}
	}
	return &BcryptPasswordHasher{
		cost: c,
	}
}

func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hash), nil
}

func (h *BcryptPasswordHasher) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
