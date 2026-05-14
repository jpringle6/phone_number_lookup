package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jpringle6/phone_number_lookup/types"
)

// PhoneService handles phone number lookup operations
type PhoneService struct {
	// In production, could inject database, cache, external API client, etc.
}

// NewPhoneService creates a new PhoneService instance
func NewPhoneService() *PhoneService {
	return &PhoneService{}
}

// LookupPhoneNumber performs a phone number lookup and returns formatted result
func (s *PhoneService) LookupPhoneNumber(req *types.PhoneLookupRequest) (*types.PhoneLookupResponse, error) {

	// Validate request - make sure number is present and either has country code
	// or country code is provided
	errResp, err := req.Validate()
	if err != nil {
		return nil, err
	}
	if errResp != nil {
		return nil, fmt.Errorf("validation error")
	}

	// Normalize the phone number (remove spaces and +)
	normalized := types.NormalizePhoneNumber(req.PhoneNumber)

	// Determine country code
	var countryCode string
	var localNumber string
	var fullLocalNumber string

	if strings.HasPrefix(req.PhoneNumber, "+") {
		// Phone number has country code embedded
		countryCode = types.ExtractCountryCodeFromNumber(req.PhoneNumber)
		if countryCode == "" {
			return nil, fmt.Errorf("Unable to extract country code from phone number")
		}
		fullLocalNumber = strings.TrimPrefix(normalized, countryCode)

	} else if req.CountryCode != "" {
		// Use provided country code
		countryCode = req.CountryCode
		fullLocalNumber = normalized
	} else {
		return nil, fmt.Errorf("unable to determine country code")
	}

	// Validate that we have both
	if countryCode == "" || fullLocalNumber == "" {
		return nil, fmt.Errorf("invalid phone number")
	}

	// Extract area code from local number (first 3 digits of local number)
	areaCode := ""
	if len(fullLocalNumber) >= 3 {
		areaCode = fullLocalNumber[:3]
		localNumber = fullLocalNumber[3:] // Remove area code from local number
	}

	// Build response
	response := &types.PhoneLookupResponse{
		PhoneNumber:      req.PhoneNumber,
		CountryCode:      countryCode,
		AreaCode:         areaCode,
		LocalPhoneNumber: localNumber,
	}

	return response, nil
}

// NewPhoneLookupHandler creates a handler for phone lookup
func (s *PhoneService) NewPhoneLookupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Parse query parameters
		phoneNumber := r.URL.Query().Get("phoneNumber")
		countryCode := r.URL.Query().Get("countryCode")

		req := &types.PhoneLookupRequest{
			PhoneNumber: phoneNumber,
			CountryCode: countryCode,
		}

		// Validate request
		errResp, _ := req.Validate()
		if errResp != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		// Perform lookup
		result, err := s.LookupPhoneNumber(req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(types.ErrorResponse{
				Error: map[string]interface{}{
					"message": err.Error(),
				},
			})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}
