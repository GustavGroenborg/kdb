package auth

import (
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Errorf("CheckPassword failed on correct password")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Errorf("CheckPassword succeeded on wrong password")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := []byte("super-secret-jwt-key")
	userID := "usr-12345"
	username := "testuser"
	email := "test@example.com"

	token, err := GenerateToken(secret, userID, username, email, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(secret, token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID || claims.Username != username || claims.Email != email {
		t.Errorf("Claims mismatch: got %+v", claims)
	}

	// Test invalid secret
	_, err = ValidateToken([]byte("wrong-secret"), token)
	if err == nil {
		t.Errorf("Expected token validation failure with wrong secret")
	}

	// Test expired token
	expiredToken, err := GenerateToken(secret, userID, username, email, -time.Minute)
	if err != nil {
		t.Fatalf("Generate expired token failed: %v", err)
	}
	_, err = ValidateToken(secret, expiredToken)
	if err == nil {
		t.Errorf("Expected token validation failure with expired token")
	}
}
