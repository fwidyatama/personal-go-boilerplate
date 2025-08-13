package validator_test

import (
	"go-boilerplate/internal/validator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var v *validator.Validator

	BeforeEach(func() {
		v = validator.NewValidator()
	})

	Describe("ValidateEmail", func() {
		Context("with valid emails", func() {
			It("should pass validation", func() {
				validEmails := []string{
					"test@example.com",
					"user.name@domain.co.uk",
					"user+tag@example.org",
				}

				for _, email := range validEmails {
					err := v.ValidateEmail(email)
					Expect(err).ToNot(HaveOccurred(), "Email: %s", email)
				}
			})
		})

		Context("with invalid emails", func() {
			It("should fail validation", func() {
				invalidEmails := []string{
					"",
					"invalid-email",
					"@example.com",
					"user@",
					"user..name@example.com",
				}

				for _, email := range invalidEmails {
					err := v.ValidateEmail(email)
					Expect(err).To(HaveOccurred(), "Email: %s", email)
				}
			})
		})
	})

	Describe("ValidatePassword", func() {
		Context("with valid passwords", func() {
			It("should pass validation", func() {
				validPasswords := []string{
					"Password123!",
					"MySecure@Pass1",
					"Complex#Pass99",
				}

				for _, password := range validPasswords {
					err := v.ValidatePassword(password)
					Expect(err).ToNot(HaveOccurred(), "Password: %s", password)
				}
			})
		})

		Context("with invalid passwords", func() {
			It("should fail validation", func() {
				invalidPasswords := map[string]string{
					"":                "password is required",
					"short":           "password must be at least 8 characters long",
					"nouppercase123!": "password must contain at least one uppercase letter",
					"NOLOWERCASE123!": "password must contain at least one lowercase letter",
					"NoNumbers!":      "password must contain at least one number",
					"NoSpecial123":    "password must contain at least one special character",
				}

				for password, expectedError := range invalidPasswords {
					err := v.ValidatePassword(password)
					Expect(err).To(HaveOccurred(), "Password: %s", password)
					Expect(err.Error()).To(ContainSubstring(expectedError))
				}
			})
		})
	})

	Describe("ValidateName", func() {
		Context("with valid names", func() {
			It("should pass validation", func() {
				validNames := []string{
					"John Doe",
					"Mary-Jane",
					"O'Connor",
					"Jean-Pierre",
				}

				for _, name := range validNames {
					err := v.ValidateName(name)
					Expect(err).ToNot(HaveOccurred(), "Name: %s", name)
				}
			})
		})

		Context("with invalid names", func() {
			It("should fail validation", func() {
				invalidNames := []string{
					"",
					"A",
					"John123",
					"User@Name",
				}

				for _, name := range invalidNames {
					err := v.ValidateName(name)
					Expect(err).To(HaveOccurred(), "Name: %s", name)
				}
			})
		})
	})

	Describe("ValidateCreateUserRequest", func() {
		It("should validate all fields", func() {
			errors := v.ValidateCreateUserRequest("John Doe", "john@example.com", "Password123!")
			Expect(errors).To(BeEmpty())
		})

		It("should return multiple errors for invalid fields", func() {
			errors := v.ValidateCreateUserRequest("", "invalid-email", "weak")
			Expect(errors).To(HaveLen(3))
		})
	})
})
