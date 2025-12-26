package idgen

import (
	"github.com/google/uuid"
)

type SecureIDGenerator struct{}

func NewSecureIDGenerator() *SecureIDGenerator {
	return &SecureIDGenerator{}
}

func (g *SecureIDGenerator) Generate() uuid.UUID {
	return uuid.New()
}
