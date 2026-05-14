package types

import (
	"regexp"
	"strings"
)

// PhoneLookupRequest represents the incoming HTTP request
type PhoneLookupRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	CountryCode string `json:"countryCode,omitempty"`
}

// PhoneLookupResponse represents the successful response
type PhoneLookupResponse struct {
	PhoneNumber      string `json:"phoneNumber"`
	CountryCode      string `json:"countryCode"`
	AreaCode         string `json:"areaCode,omitempty"`
	LocalPhoneNumber string `json:"localPhoneNumber"`
}

// ErrorDetail represents a single validation error
type ErrorDetail struct {
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error map[string]interface{} `json:"error"`
}

// ValidatePhoneLookupRequest validates the phone lookup request
func (r *PhoneLookupRequest) Validate() (*ErrorResponse, error) {
	if r.PhoneNumber == "" {
		return &ErrorResponse{
			Error: map[string]interface{}{
				"phoneNumber": "required value is missing",
			},
		}, nil
	}

	// Check if phone number has country code (starts with +)
	hasCountryCode := strings.HasPrefix(r.PhoneNumber, "+")

	if !hasCountryCode && r.CountryCode == "" {
		return &ErrorResponse{
			Error: map[string]interface{}{
				"countryCode": "required value is missing",
			},
		}, nil
	}

	if r.CountryCode != "" && !isValidISO3166Alpha2(r.CountryCode) {
		return &ErrorResponse{
			Error: map[string]interface{}{
				"countryCode": "invalid ISO 3166-1 alpha-2 format",
			},
		}, nil
	}

	if !isValidPhoneFormat(r.PhoneNumber) {
		return &ErrorResponse{
			Error: map[string]interface{}{
				"phoneNumber": "invalid phone number format",
			},
		}, nil
	}

	return nil, nil
}

func isValidPhoneFormat(phoneNumber string) bool {
	// Remove leading + if present
	cleaned := strings.TrimPrefix(phoneNumber, "+")

	// Check if only contains digits and spaces
	for _, ch := range cleaned {
		if !((ch >= '0' && ch <= '9') || ch == ' ') {
			return false
		}
	}

	// Check for consecutive spaces (invalid)
	if strings.Contains(cleaned, "  ") {
		return false // Multiple spaces in a row are invalid
	}

	// Remove all spaces to check total digit count
	digitsOnly := strings.ReplaceAll(cleaned, " ", "")
	if len(digitsOnly) < 10 {
		return false // Too short - less than 7 digits total
	}

	// Validate space placement - each space-separated part should be digits only
	parts := strings.Fields(cleaned) // Split by whitespace

	// Each part should be digits only and non-empty
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}

	return true
}

// isValidISO3166Alpha2 checks if country code is valid ISO 3166-1 alpha-2
func isValidISO3166Alpha2(code string) bool {
	if len(code) != 2 {
		return false
	}

	matched, _ := regexp.MatchString(`^[A-Z]{2}$`, code)
	return matched
}

// NormalizePhoneNumber removes spaces and leading +
func NormalizePhoneNumber(phoneNumber string) string {
	normalized := strings.ReplaceAll(phoneNumber, " ", "")
	return strings.TrimPrefix(normalized, "+")
}

// ExtractCountryCodeFromNumber extracts country code from a phone number
// Special case: +1 is automatically mapped to US (North America)
// If the number has spaces: country code is everything before the first space
// If no spaces and starts with 1: assume US country code
func ExtractCountryCodeFromNumber(phoneNumber string) string {
	// Remove leading +
	normalized := strings.TrimPrefix(phoneNumber, "+")

	// Find first space - country code is before it
	spaceIndex := strings.Index(normalized, " ")

	if spaceIndex != -1 {
		// Space found - country code is everything before first space
		countryCodeDigits := normalized[:spaceIndex]
		// Special case: +1 is US
		if countryCodeDigits == "1" {
			return "US"
		}
		return countryCodeDigits
	}

	// No space found
	// Special case: +1 (country code 1) is North America (US, Canada, etc.)
	if strings.HasPrefix(normalized, "1") && (len(normalized) > 1 && normalized[1] >= '0' && normalized[1] <= '9') {
		return "US"
	}

	// Ambiguous without a lookup library - return empty string
	// This will require the countryCode parameter
	return ""
}
