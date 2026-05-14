#!/bin/bash

# Comprehensive test script for phone number lookup API
# Tests all permutations and combinations of valid/invalid inputs

BASE_URL="http://localhost:8080/v1/phone-numbers"

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run a test
run_test() {
    local test_name=$1
    local phone_number=$2
    local country_code=$3
    local expected_contains=$4
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Build URL with query parameters
    local url="${BASE_URL}?phoneNumber=$(echo -n "$phone_number" | jq -sRr @uri)"
    if [ -n "$country_code" ]; then
        url="${url}&countryCode=${country_code}"
    fi
    
    echo -e "\n${YELLOW}Test $TESTS_RUN: $test_name${NC}"
    echo "URL: $url"
    
    # Make the request and capture output
    response=$(curl -s "$url")
    echo "Response: $response"
    
    # Check if expected content is in response
    if echo "$response" | grep -q "$expected_contains"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

echo "========================================"
echo "Phone Number Lookup API Test Suite"
echo "========================================"
echo "Testing endpoint: $BASE_URL"
echo ""

# ========== VALID CASES ==========
echo -e "\n${GREEN}========== VALID CASES ==========${NC}"

# US numbers with +1
run_test "US number with + and no spaces" "+12125690123" "" "212"
run_test "US number with + and spaces" "+1 212 569 0123" "" "212"
run_test "US number with + and different spacing" "+1 212 5690123" "" "212"

# Mexico numbers
run_test "Mexico number with + and spaces" "+52 631 3118150" "" "631"

# Spain numbers
run_test "Spain number with + and spaces" "+34 915 872200" "" "915"

# Portugal numbers (but without spaces - ambiguous, needs countryCode)
run_test "Portugal number with countryCode param" "2109420000" "PT" "210"

# Numbers without + but with countryCode
run_test "Number without + with US countryCode" "2125690123" "US" "212"
run_test "Number without + with MX countryCode" "6313118150" "MX" "631"

# Edge case: single digit country code with spaces
run_test "Japan number with + and spaces" "+81 90 1234 5678" "" "901"

# ========== INVALID CASES - FORMAT ISSUES ==========
echo -e "\n${YELLOW}========== INVALID CASES - FORMAT ISSUES ==========${NC}"

# Multiple space-separated parts is VALID (e.g., country code, area code, local parts)
run_test "Multiple space-separated parts (4 parts) is VALID" "351 21 094 2000" "PT" "351"

# Consecutive spaces
run_test "Consecutive spaces" "+1  212  5690123" "" "invalid"

# Invalid characters (hyphens)
run_test "Hyphenated number (invalid)" "+1-212-569-0123" "" "invalid"

# Invalid characters (parentheses)
run_test "Parenthesized number (invalid)" "+1 (212) 569-0123" "" "invalid"

# Too short (less than 7 digits total)
run_test "Too short - 6 digits" "+1 234567" "" "invalid"

# Empty search term is handled by checking missing phoneNumber
run_test "Missing phoneNumber parameter" "" "" "required"

# ========== MISSING COUNTRY CODE CASES ==========
echo -e "\n${YELLOW}========== MISSING COUNTRY CODE CASES ==========${NC}"

# Ambiguous format (no + and no spaces, can't determine CC)
run_test "Ambiguous format - no + no spaces, no CC param" "2125690123" "" "required value is missing"

# Format with spaces but no +, no CC param
run_test "Format with spaces no +, no CC param" "212 569 0123" "" "required value is missing"

# ========== INVALID COUNTRY CODE CASES ==========
echo -e "\n${YELLOW}========== INVALID COUNTRY CODE CASES ==========${NC}"

# Country code not in ISO 3166-1 alpha-2 format
run_test "Invalid CC format - 3 letters" "2125690123" "USA" "invalid ISO 3166-1 alpha-2 format"

# Country code lowercase
run_test "Invalid CC format - lowercase" "2125690123" "us" "invalid ISO 3166-1 alpha-2 format"

# Country code with numbers
run_test "Invalid CC format - with numbers" "2125690123" "U1" "invalid ISO 3166-1 alpha-2 format"

# Single letter country code
run_test "Invalid CC format - single letter" "2125690123" "U" "invalid ISO 3166-1 alpha-2 format"



# ========== EDGE CASES ==========
echo -e "\n${YELLOW}========== EDGE CASES ==========${NC}"

# Very long number (15 digits total max for E.164)
run_test "Maximum length E.164 (15 digits)" "+1 201 555 0123" "" "201"

# Country code at boundary (1 digit)
run_test "1-digit CC (US)" "+1 555 555 5555" "" "555"

# Country code at boundary (3 digits)
run_test "3-digit CC (assumed valid)" "+886 2 1234 5678" "" "212"

# Single space format
run_test "Single part after CC" "+1 2125690123" "" "212"

# ========== SPECIAL +1 CASES ==========
echo -e "\n${YELLOW}========== SPECIAL +1 (US) CASES ==========${NC}"

# +1 without spaces should map to US
run_test "US number +1 no spaces maps to US" "+12125690123" "" "212"

# +1 with spaces should also map to US
run_test "US number +1 with spaces maps to US" "+1 212 569 0123" "" "212"

# +1 with CC param US (should work)
run_test "+1 with explicit US countryCode" "+1 212 569 0123" "US" "212"
# ========== AMBIGUOUS COUNTRY CODE (with +) ==========
echo -e "\n${YELLOW}========== AMBIGUOUS COUNTRY CODE (with +) ==========${NC}"

# Number with + but ambiguous CC (could be +52 898 or +528 98, etc.) - should fail
run_test "Ambiguous CC with + (no spaces), no CC param - should fail" "+528981231111" "" "Unable to extract country code"

# Same number but WITH countryCode param - still fails because we can't extract from the + number
run_test "Ambiguous CC with + (no spaces), WITH CC param - should fail" "+528981231111" "MX" "Unable to extract country code"

# Number with + and spaces - should work with extracted CC
run_test "Valid CC with + and spaces" "+52 898 123 1111" "" "898"

# Number with + and spaces but conflicting CC param - uses extracted CC from +
run_test "Valid CC with + and spaces, CC param provided" "+52 898 123 1111" "US" "898"
# ========== SUMMARY ==========
echo -e "\n========================================${NC}"
echo -e "Test Summary:"
echo -e "Total Tests Run: ${YELLOW}$TESTS_RUN${NC}"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi