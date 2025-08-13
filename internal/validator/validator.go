package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// Validator provides validation methods
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 {
		return fmt.Errorf("email is too long (max 254 characters)")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password is too long (max 128 characters)")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateName validates user name
func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}

	if len(name) > 100 {
		return fmt.Errorf("name is too long (max 100 characters)")
	}

	// Allow letters, spaces, hyphens, and apostrophes
	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("name can only contain letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

// ValidateUUID validates UUID format
func (v *Validator) ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("ID is required")
	}

	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

// ValidatePagination validates pagination parameters
func (v *Validator) ValidatePagination(limit, offset int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	return nil
}

// ValidateCreateUserRequest validates user creation request
func (v *Validator) ValidateCreateUserRequest(name, email, password string) ValidationErrors {
	var errors ValidationErrors

	if err := v.ValidateName(name); err != nil {
		errors = append(errors, ValidationError{Field: "name", Message: err.Error()})
	}

	if err := v.ValidateEmail(email); err != nil {
		errors = append(errors, ValidationError{Field: "email", Message: err.Error()})
	}

	if err := v.ValidatePassword(password); err != nil {
		errors = append(errors, ValidationError{Field: "password", Message: err.Error()})
	}

	return errors
}

// ValidateUpdateUserRequest validates user update request
func (v *Validator) ValidateUpdateUserRequest(name, email, password string) ValidationErrors {
	var errors ValidationErrors

	if err := v.ValidateName(name); err != nil {
		errors = append(errors, ValidationError{Field: "name", Message: err.Error()})
	}

	if err := v.ValidateEmail(email); err != nil {
		errors = append(errors, ValidationError{Field: "email", Message: err.Error()})
	}

	// Password is optional for updates, but if provided, must be valid
	if password != "" {
		if err := v.ValidatePassword(password); err != nil {
			errors = append(errors, ValidationError{Field: "password", Message: err.Error()})
		}
	}

	return errors
}

// ValidateLoginRequest validates login request
func (v *Validator) ValidateLoginRequest(email, password string) ValidationErrors {
	var errors ValidationErrors

	if err := v.ValidateEmail(email); err != nil {
		errors = append(errors, ValidationError{Field: "email", Message: err.Error()})
	}

	if password == "" {
		errors = append(errors, ValidationError{Field: "password", Message: "password is required"})
	}

	return errors
}

// ValidateChangePasswordRequest validates change password request
func (v *Validator) ValidateChangePasswordRequest(currentPassword, newPassword string) ValidationErrors {
	var errors ValidationErrors

	if currentPassword == "" {
		errors = append(errors, ValidationError{Field: "current_password", Message: "current password is required"})
	}

	if err := v.ValidatePassword(newPassword); err != nil {
		errors = append(errors, ValidationError{Field: "new_password", Message: err.Error()})
	}

	if currentPassword == newPassword {
		errors = append(errors, ValidationError{Field: "new_password", Message: "new password must be different from current password"})
	}

	return errors
}
