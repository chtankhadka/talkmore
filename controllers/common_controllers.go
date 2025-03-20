package controllers

import (
	"golang.org/x/crypto/bcrypt"
)

func Generate_OTP() *int {

	return nil

}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain-text password with a stored hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SendMail(userMail string, message string) bool {
	return true
}
