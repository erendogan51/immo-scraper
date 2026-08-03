package util

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
)

func DeterministicUUID(input string) uuid.UUID {
	hash := [16]byte(sha3.New224().Sum([]byte(input)))
	id, _ := uuid.FromBytes(hash[:])
	return id
}
