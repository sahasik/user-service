package utils

import (
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordRequirements defines password validation rules
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordRequirements returns default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}
}

// ValidatePassword validates password against requirements
func ValidatePassword(password string) error {
	req := DefaultPasswordRequirements()

	// Check minimum length
	if len(password) < req.MinLength {
		return fmt.Errorf("password must be at least %d characters long", req.MinLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	// Check character types
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Validate requirements
	var errors []string

	if req.RequireUpper && !hasUpper {
		errors = append(errors, "at least one uppercase letter")
	}
	if req.RequireLower && !hasLower {
		errors = append(errors, "at least one lowercase letter")
	}
	if req.RequireDigit && !hasDigit {
		errors = append(errors, "at least one number")
	}
	if req.RequireSpecial && !hasSpecial {
		errors = append(errors, "at least one special character (!@#$%^&*)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("password must contain: %v", errors)
	}

	return nil
}

// HashPassword hashes a password using bcrypt with validation
func HashPassword(password string) (string, error) {
	// Validate password first
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength returns password strength score (1-5)
func ValidatePasswordStrength(password string) (int, string) {
	score := 0
	feedback := []string{}

	// Length check
	if len(password) >= 8 {
		score++
	} else {
		feedback = append(feedback, "increase length to 8+ characters")
	}

	// Character variety checks
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasLower {
		score++
	} else {
		feedback = append(feedback, "add lowercase letters")
	}

	if hasUpper {
		score++
	} else {
		feedback = append(feedback, "add uppercase letters")
	}

	if hasDigit {
		score++
	} else {
		feedback = append(feedback, "add numbers")
	}

	if hasSpecial {
		score++
	} else {
		feedback = append(feedback, "add special characters")
	}

	// Additional length bonus
	if len(password) >= 12 {
		score++
	}

	// Common patterns check
	commonPatterns := []string{
		"123456", "password", "admin", "qwerty", "abc123",
		"000000", "111111", "password123", "admin123",
	}

	for _, pattern := range commonPatterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, password)
		if matched {
			score = max(score-2, 0)
			feedback = append(feedback, "avoid common patterns")
			break
		}
	}

	strength := "Very Weak"
	switch {
	case score >= 5:
		strength = "Very Strong"
	case score >= 4:
		strength = "Strong"
	case score >= 3:
		strength = "Medium"
	case score >= 2:
		strength = "Weak"
	}

	return score, strength
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
