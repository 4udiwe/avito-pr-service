package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{}

func New() *BcryptHasher {
	return &BcryptHasher{}
}

func (BcryptHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (BcryptHasher) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
