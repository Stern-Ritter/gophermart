package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestGetPasswordHash(t *testing.T) {
	t.Run("should correct hash password", func(t *testing.T) {
		password := "secretPassword"
		hashedPassword, err := GetPasswordHash(password)

		require.NoError(t, err, "Error hashing password")
		require.NotEmpty(t, hashedPassword, "Hashed password should not be empty")

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		assert.NoError(t, err, "Expected hashed password to match origin password")
	})

	t.Run("should return error when hashing password is empty", func(t *testing.T) {
		invalidPassword := ""
		_, err := GetPasswordHash(invalidPassword)

		assert.Error(t, err, "Expected error when hashing password is empty")
	})
}

func TestCheckPasswordHash(t *testing.T) {
	password := "secretPassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Error hashing password")

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "should return true when password and hash are valid",
			password: password,
			hash:     string(hashedPassword),
			want:     true},
		{
			name:     "should return false when password is invalid",
			password: "wrongPassword",
			hash:     string(hashedPassword),
			want:     false,
		},
		{
			name:     "should return false when hash is invalid",
			password: password,
			hash:     "$2a$10$",
			want:     false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.want, got, "CheckPasswordHash(%s, %s) = %v, want %v", tt.password, tt.hash, got, tt.want)
		})
	}
}
