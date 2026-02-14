package scanner

import (
	"regexp"
	"strings"
	"testing"
)

// --- PERSON tests ---

func TestPerson_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Herr Thomas Schmidt ist hier.", "Thomas Schmidt"},
		{"Frau Maria Müller kommt morgen.", "Maria Müller"},
		{"Patient Jean-Pierre Dubois wird behandelt.", "Jean-Pierre Dubois"},
		{"Mr. John Smith arrived.", "John Smith"},
		{"Mevrouw Anna Bakker belde.", "Anna Bakker"},
		{"Monsieur Pierre Dupont est arrivé.", "Pierre Dupont"},
		{"Madame Sophie Laurent est là.", "Sophie Laurent"},
		{"Mevrouw Anna Bakker woont hier.", "Anna Bakker"},
		{"my colleague Thomas Weber said hello.", "Thomas Weber"},
		// German role triggers
		{"Leiter: Franz Zaman", "Franz Zaman"},
		{"Leiterin Maria Berger ist zuständig.", "Maria Berger"},
		{"Geschäftsführer Thomas Weber leitet.", "Thomas Weber"},
		{"Inhaber: Hans Gruber", "Hans Gruber"},
		{"Direktor Karl Müller berichtet.", "Karl Müller"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "PERSON" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PERSON not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestPerson_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The weather is nice today.",
		"Berlin is a city in Germany.",
		"I went to the store yesterday.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "PERSON" {
				t.Errorf("PERSON false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- EMAIL tests ---

func TestEmail_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Contact thomas.schmidt@example.com for info.", "thomas.schmidt@example.com"},
		{"Send to müller@example.de please.", "müller@example.de"},
		{"Email anna.devries@voorbeeld.nl now.", "anna.devries@voorbeeld.nl"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "EMAIL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("EMAIL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestEmail_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"This is not an email address.",
		"Use the @ symbol for emails.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "EMAIL" {
				t.Errorf("EMAIL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- PHONE tests ---

func TestPhone_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Call +49 170 4839201 now.", "+49 170 4839201"},
		{"Ring +44 20 7946 0958 please.", "+44 20 7946 0958"},
		{"Appelez le +33 6 12 34 56 78.", "+33 6 12 34 56 78"},
		{"Chiami +39 06 1234 5678.", "+39 06 1234 5678"},
		// Trunk prefix (0) notation
		{"Tel: +43-(0)6212 2368", "+43-(0)6212 2368"},
		{"Call +49 (0)30 1234567", "+49 (0)30 1234567"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "PHONE" {
				// Allow partial match (phone patterns may vary in boundaries).
				if e.Text == tc.want || containsDigits(e.Text, tc.want) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("PHONE not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func containsDigits(a, b string) bool {
	da := digitsOnly(a)
	db := digitsOnly(b)
	return da == db
}

func digitsOnly(s string) string {
	var out []byte
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out = append(out, byte(r))
		}
	}
	return string(out)
}

func TestPhone_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The version is 2.0.1 and the build number is 12345.",
		"The year 2024 was great.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "PHONE" {
				t.Errorf("PHONE false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- IBAN tests ---

func TestIBAN_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"IBAN: DE89 3704 0044 0532 0130 00", "DE89 3704 0044 0532 0130 00"},
		{"Transfer to GB29 NWBK 6016 1331 9268 19 please.", "GB29 NWBK 6016 1331 9268 19"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "IBAN" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("IBAN not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestIBAN_ChecksumValidation(t *testing.T) {
	// Valid IBAN
	if !validateIBAN("DE89370400440532013000") {
		t.Error("valid IBAN DE89370400440532013000 rejected")
	}
	if !validateIBAN("GB29 NWBK 6016 1331 9268 19") {
		t.Error("valid IBAN GB29 NWBK 6016 1331 9268 19 rejected")
	}

	// Invalid checksum
	if validateIBAN("DE00370400440532013000") {
		t.Error("invalid IBAN DE00370400440532013000 accepted")
	}
	if validateIBAN("XX123456") {
		t.Error("invalid IBAN XX123456 accepted")
	}
}

func TestIBAN_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The code is ABCD1234.",
		"Product ID: XX99 1234 5678 9999",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "IBAN" {
				t.Errorf("IBAN false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- CREDIT_CARD tests ---

func TestCreditCard_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
	}{
		{"Card: 4111 1111 1111 1111"},      // Visa
		{"Card: 5500 0000 0000 0004"},      // Mastercard
		{"Card: 3782 822463 10005"},        // Amex
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "CREDIT_CARD" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CREDIT_CARD not found in %q, got %v", tc.input, entities)
		}
	}
}

func TestCreditCard_LuhnValidation(t *testing.T) {
	if !validateLuhn("4111111111111111") {
		t.Error("valid Visa card rejected")
	}
	if !validateLuhn("5500 0000 0000 0004") {
		t.Error("valid Mastercard rejected")
	}
	if validateLuhn("4111111111111112") {
		t.Error("invalid card accepted")
	}
}

func TestCreditCard_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Order #1234567890123456",
		"Tracking: 9999888877776666",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "CREDIT_CARD" {
				t.Errorf("CREDIT_CARD false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- DATE tests ---

func TestDate_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// Numeric formats
		{"DE dot", "geboren am 15.03.1990 in Berlin.", "15.03.1990"},
		{"slash", "Date: 14/03/2024 confirmed.", "14/03/2024"},
		{"dash", "Termin am 01-12-2023.", "01-12-2023"},
		// Written English dates
		{"EN full month", "Date of issue February 12, 2026", "February 12, 2026"},
		{"EN abbr month", "Filed on Jan 5, 2025", "Jan 5, 2025"},
		{"EN no comma", "Due March 1 2024", "March 1 2024"},
		// Written German dates
		{"DE written", "geboren am 15. März 1990", "15. März 1990"},
		{"DE written long", "am 1. Januar 2026 eingereicht", "1. Januar 2026"},
		// ISO format
		{"ISO", "Created: 2024-03-15", "2024-03-15"},
		// Written French dates
		{"FR written", "le 12 février 2026", "12 février 2026"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "DATE" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("DATE not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestDate_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Version 2.0.1 released.",
		"The ratio is 3/4 of the total.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "DATE" {
				t.Errorf("DATE false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- URL tests ---

func TestURL_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Visit https://example.com for info.", "https://example.com"},
		{"Link: http://test.org/path?q=1 here.", "http://test.org/path?q=1"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "URL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("URL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestURL_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Visit example.com for info.", // no scheme
		"ftp is old protocol.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "URL" {
				t.Errorf("URL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- IP_ADDRESS tests ---

func TestIP_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Server at 192.168.1.1 is down.", "192.168.1.1"},
		{"Localhost is ::1 always.", "::1"},
		{"IP: 10.0.0.1 assigned.", "10.0.0.1"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "IP_ADDRESS" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("IP_ADDRESS not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestIP_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The value is 999.999.999.999 which is invalid.",
		"Version 1.2.3 is stable.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "IP_ADDRESS" {
				t.Errorf("IP_ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- FINANCIAL tests ---

func TestFinancial_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Rechnung über €1.250,00 gesendet.", "€1.250,00"},
		{"Total: $2,500.00 charged.", "$2,500.00"},
		{"Amount: £1,000.00 paid.", "£1,000.00"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "FINANCIAL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FINANCIAL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestFinancial_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The temperature is 25 degrees.",
		"Page 42 of the report.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "FINANCIAL" {
				t.Errorf("FINANCIAL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- ADDRESS tests ---

func TestAddress_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		// German
		{"DE suffix", "Wohnt in Hauptstraße 42, 10115 Berlin."},
		// Italian
		{"IT via", "Via Roma 42, 00184 Roma."},
		// Dutch
		{"NL gracht", "Keizersgracht 100 in Amsterdam."},
		// US street
		{"US street", "Address: 440 N Barranca Ave #4133"},
		{"US street simple", "Lives at 123 Main Street now."},
		{"US street blvd", "Office at 500 Broadway Blvd"},
		{"US street directional", "Located at 1600 Pennsylvania Ave NW"},
		// US city/state/ZIP
		{"US city state zip abbr", "Covina, CA 91723"},
		{"US city state zip full", "Covina, California 91723"},
		{"US zip+4", "New York, NY 10001-1234"},
		{"US multi-word city", "San Francisco, CA 94102"},
		// Austrian apartment notation
		{"AT apartment", "Quellenstraße 42/3/12"},
		{"AT apartment short", "Mariahilferstraße 5/2"},
		{"AT 4-digit postcode", "Hauptstraße 42, 1010 Wien"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q, got %v", tc.input, entities)
			}
		})
	}
}

func TestAddress_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"plain text", "The weather is sunny today."},
		{"project text", "We discussed the project plan."},
		{"bare number", "The 440 report was filed."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, e := range entities {
				if e.Type == "ADDRESS" {
					t.Errorf("ADDRESS false positive in %q: got %v", tc.input, e)
				}
			}
		})
	}
}

// --- SECRET tests ---

func TestSecret_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
	}{
		{"Key: sk-proj-abc123def456ghi789jkl012mno345pqr678"},
		{"AWS: AKIA1234567890ABCDEF"},
		{"Token: ghp_abcdefghijklmnopqrstuvwxyz0123456789"},
		{"API: sk-ant-abcdefghij1234567890abcd"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "SECRET" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SECRET not found in %q, got %v", tc.input, entities)
		}
	}
}

func TestSecret_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The sky is blue and the grass is green.",
		"Use ssh-keygen to generate keys.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "SECRET" {
				t.Errorf("SECRET false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- SSN tests ---

func TestSSN_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{"US SSN", "My SSN is 123-45-6789.", "SSN", "123-45-6789"},
		{"Swiss AHV", "AHV-Nr: 756.1234.5678.97", "SSN", "756.1234.5678.97"},
		{"UK NINO", "National insurance AB123456C", "SSN", "AB123456C"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == tc.wantType && e.Text == tc.wantText {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.wantText, entities)
			}
		})
	}
}

func TestSSN_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"US SSN rejected 000", "SSN 000-12-3456"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, e := range entities {
				if e.Type == "SSN" {
					t.Errorf("SSN false positive in %q: got %v", tc.input, e)
				}
			}
		})
	}
}

// --- MEDICAL tests ---

func TestMedical_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{"ICD-10 code", "Diagnose: J45.0", "MEDICAL", "J45.0"},
		{"Blood pressure", "BP: 120/80 mmHg", "MEDICAL", "120/80 mmHg"},
		{"Lab value mg/dL", "Glucose: 126 mg/dL", "MEDICAL", "126 mg/dL"},
		{"BMI value", "BMI: 28.5", "MEDICAL", "28.5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == tc.wantType && e.Text == tc.wantText {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.wantText, entities)
			}
		})
	}
}

func TestMedical_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The score is 5/10 overall.",
		"The temperature is 25 degrees.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "MEDICAL" {
				t.Errorf("MEDICAL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- AGE tests ---

func TestAge_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{"years old", "Patient is 45 years old", "AGE", "45"},
		{"Jahre alt", "Patient ist 67 Jahre alt", "AGE", "67"},
		{"age context", "age: 32", "AGE", "32"},
		{"born in year", "born in 1990", "AGE", "1990"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == tc.wantType && e.Text == tc.wantText {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.wantText, entities)
			}
		})
	}
}

func TestAge_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The building is 200 meters tall.",
		"We need 50 items in stock.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "AGE" {
				t.Errorf("AGE false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- ID_NUMBER tests ---

func TestIDNumber_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{"German Steuer-ID", "Steuer-ID: 12345678901", "ID_NUMBER", "12345678901"},
		{"Passport number", "Reisepass: C01X00T47", "ID_NUMBER", "C01X00T47"},
		{"EU VAT", "VAT DE123456789", "ID_NUMBER", "DE123456789"},
		// Invoice numbers
		{"Invoice number", "Invoice number A49E63AA-0002", "ID_NUMBER", "A49E63AA-0002"},
		{"Invoice no.", "Invoice no. INV-2024-001", "ID_NUMBER", "INV-2024-001"},
		{"Rechnungsnummer", "Rechnungsnummer RE-2024/0042", "ID_NUMBER", "RE-2024/0042"},
		{"Order number", "Order number ORD-99887", "ID_NUMBER", "ORD-99887"},
		{"Reference colon", "Reference: REF-ABC-123", "ID_NUMBER", "REF-ABC-123"},
		{"Invoice colon", "Invoice: 2024-0042", "ID_NUMBER", "2024-0042"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == tc.wantType && e.Text == tc.wantText {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.wantText, entities)
			}
		})
	}
}

func TestIDNumber_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The reference number is ABC123.",
		"Order ID 9876 confirmed.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "ID_NUMBER" {
				t.Errorf("ID_NUMBER false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- MAC_ADDRESS tests ---

func TestMACAddress_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{"MAC colon", "Device MAC: 00:1A:2B:3C:4D:5E", "MAC_ADDRESS", "00:1A:2B:3C:4D:5E"},
		{"MAC dash", "MAC 00-1A-2B-3C-4D-5E", "MAC_ADDRESS", "00-1A-2B-3C-4D-5E"},
		{"MAC Cisco", "Interface aabb.ccdd.eeff", "MAC_ADDRESS", "aabb.ccdd.eeff"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == tc.wantType && e.Text == tc.wantText {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.wantText, entities)
			}
		})
	}
}

func TestMACAddress_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The hex value is 0x1A2B3C.",
		"Color code #FF00FF is magenta.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "MAC_ADDRESS" {
				t.Errorf("MAC_ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- PHONE/IBAN overlap tests ---

func TestPhone_NotInsideIBAN(t *testing.T) {
	s := DefaultScanner(nil)
	// Fake IBANs with invalid checksums — should NOT produce PHONE entities.
	cases := []string{
		"IBAN: DE75 5001 0517 0123 4567 89",
		"Transfer to DE27 5135 0025 0205 1340 64 please.",
		"FR00 2004 1010 0505 0001 3M02 606",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "PHONE" {
				t.Errorf("PHONE false positive inside fake IBAN in %q: got %v", input, e)
			}
		}
	}
}

// --- ORG tests ---

func TestOrganization_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"German GmbH", "Kontakt mit Müller GmbH aufgenommen.", "Müller GmbH"},
		{"German SE", "SAP SE hat berichtet.", "SAP SE"},
		{"German AG", "von Siemens AG erhalten.", "Siemens AG"},
		{"International Ltd", "Filed by Acme Ltd today.", "Acme Ltd"},
		{"Universitätsklinikum", "Behandlung im Universitätsklinikum Frankfurt.", "Universitätsklinikum Frankfurt"},
		{"Klinik am", "Aufnahme in die Klinik am See.", "Klinik am See"},
		{"French hospital", "Admis à l'Hôpital Européen.", "Hôpital Européen"},
		{"Italian hospital", "Ricoverato all'Ospedale San Raffaele.", "Ospedale San Raffaele"},
		{"Spanish hospital", "Ingresado en Hospital Universitario.", "Hospital Universitario"},
		{"AOK", "Versichert bei AOK Hessen.", "AOK Hessen"},
		{"Deutsche Rentenversicherung", "Antrag bei Deutsche Rentenversicherung.", "Deutsche Rentenversicherung"},
		{"UMC suffix", "Behandlung im Amsterdam UMC.", "Amsterdam UMC"},
		{"UMC prefix", "Referred to UMC Utrecht.", "UMC Utrecht"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ORG" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ORG not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestOrganization_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"plain text", "The weather is nice today."},
		{"bare Klinik", "Die Klinik ist geschlossen."},
		{"lowercase gmbh", "Das ist keine gmbh firma."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, e := range entities {
				if e.Type == "ORG" {
					t.Errorf("ORG false positive in %q: got %v", tc.input, e)
				}
			}
		})
	}
}

// --- Composite scanner tests ---

func TestOverlapDedup(t *testing.T) {
	// Two scanners that both match the same region with different lengths.
	short := NewRegexScanner(regexp.MustCompile(`test`), "X", 0.9)
	long := NewRegexScanner(regexp.MustCompile(`test case`), "Y", 0.9)

	cs := NewCompositeScanner([]Scanner{short, long}, nil)
	entities := cs.Scan("this is a test case here")

	// Should keep only the longer match "test case".
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity after dedup, got %d: %v", len(entities), entities)
	}
	if entities[0].Text != "test case" {
		t.Errorf("expected longer match 'test case', got %q", entities[0].Text)
	}
}

func TestAllowlistFiltering(t *testing.T) {
	emailScanner := NewRegexScanner(
		regexp.MustCompile(`[a-z]+@[a-z]+\.[a-z]+`),
		"EMAIL", 0.99,
	)
	allowlist := []*regexp.Regexp{
		regexp.MustCompile(`example\.com`),
	}

	cs := NewCompositeScanner([]Scanner{emailScanner}, allowlist)
	entities := cs.Scan("Send to test@example.com and real@secret.org")

	// test@example.com should be filtered out by allowlist.
	for _, e := range entities {
		if e.Text == "test@example.com" {
			t.Error("allowlisted email was not filtered out")
		}
	}

	found := false
	for _, e := range entities {
		if e.Text == "real@secret.org" {
			found = true
		}
	}
	if !found {
		t.Error("non-allowlisted email was incorrectly filtered")
	}
}

func TestEmptyText(t *testing.T) {
	s := DefaultScanner(nil)
	entities := s.Scan("")
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for empty text, got %d", len(entities))
	}
}

func TestUTF8MultibyteOffsets(t *testing.T) {
	s := DefaultScanner(nil)
	input := "Herr Müller schrieb an müller@example.com"
	entities := s.Scan(input)

	for _, e := range entities {
		// Verify that byte offsets produce the correct substring.
		extracted := input[e.Start:e.End]
		if extracted != e.Text {
			t.Errorf("offset mismatch for %q: input[%d:%d] = %q, but Text = %q",
				input, e.Start, e.End, extracted, e.Text)
		}
	}
}

// --- Real-world invoice test ---

func TestInvoiceDocument(t *testing.T) {
	s := DefaultScanner(nil)
	// Simulates a real invoice document — should detect multiple entity types.
	input := `Invoice
Invoice number A49E63AA-0002
Date of issue February 12, 2026
Date due February 12, 2026
440 N Barranca Ave #4133
Covina, California 91723
United States
contact@example.com`

	entities := s.Scan(input)

	// Build a set of detected types for easy checking.
	typeFound := make(map[string][]string)
	for _, e := range entities {
		typeFound[e.Type] = append(typeFound[e.Type], e.Text)
	}

	// Invoice number should be detected
	if _, ok := typeFound["ID_NUMBER"]; !ok {
		t.Errorf("Invoice number not detected; entities: %v", entities)
	}
	// Written dates should be detected
	if _, ok := typeFound["DATE"]; !ok {
		t.Errorf("Written dates not detected; entities: %v", entities)
	}
	// US street address should be detected
	if _, ok := typeFound["ADDRESS"]; !ok {
		t.Errorf("US address not detected; entities: %v", entities)
	}
	// Email should be detected
	if _, ok := typeFound["EMAIL"]; !ok {
		t.Errorf("Email not detected; entities: %v", entities)
	}

	t.Logf("Detected entities: %v", entities)
}

// --- Austrian address block test ---

func TestAustrianAddressBlock(t *testing.T) {
	s := DefaultScanner(nil)
	// Typical Austrian invoice address: street, city, postcode+city, country
	input := `Musterstraße 5/2/3
Musterort
1100 Wien
Austria`

	entities := s.Scan(input)

	streetFound := false
	postcodeFound := false
	for _, e := range entities {
		if e.Type == "ADDRESS" {
			if strings.Contains(e.Text, "Musterstraße") {
				streetFound = true
			}
			if strings.Contains(e.Text, "1100") && strings.Contains(e.Text, "Wien") {
				postcodeFound = true
			}
		}
	}
	if !streetFound {
		t.Errorf("Austrian street not detected; entities: %v", entities)
	}
	if !postcodeFound {
		t.Errorf("Austrian postcode+city not detected; entities: %v", entities)
	}

	t.Logf("Detected entities: %v", entities)
}

func TestAustrianAddressNoSuffix(t *testing.T) {
	s := DefaultScanner(nil)
	// Street name without standard suffix — should still be detected
	// when part of an address block with postcode/country nearby.
	cases := []struct {
		name  string
		input string
	}{
		{
			"no suffix with country",
			"Spittelau 12\n1090 Wien\nAustria",
		},
		{
			"no suffix with postcode",
			"Am Tabor 5/2\n1020 Wien\nÖsterreich",
		},
		{
			"no suffix multi-word",
			"Untere Donaulände 7\n4020 Linz\nAustria",
		},
		{
			"no suffix with street nearby",
			"Hauptstraße 1\nSomeplace 42\n8010 Graz",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q, got %v", tc.input, entities)
			}
		})
	}
}

func TestGenericStreet_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	// Should NOT match generic word+number without address context
	cases := []string{
		"Chapter 42 was interesting.",
		"Room 101 is occupied.",
		"Version 3 released today.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "ADDRESS" {
				t.Errorf("ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- Billing label PERSON tests ---

func TestPerson_BillingTriggers(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"Bill to", "Bill to John Smith\nAustria", "John Smith"},
		{"Billed to", "Billed to Maria Müller", "Maria Müller"},
		{"Invoice to", "Invoice to Thomas Weber", "Thomas Weber"},
		{"Sold to", "Sold to Pierre Dupont", "Pierre Dupont"},
		{"Ship to", "Ship to Anna Bakker", "Anna Bakker"},
		{"Attn", "Attn: Jean Dupont", "Jean Dupont"},
		{"Attention", "Attention Sophie Laurent", "Sophie Laurent"},
		{"Bill to newline", "Bill to\nJohn Smith", "John Smith"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "PERSON" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("PERSON not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

// --- EUR dot-decimal FINANCIAL tests ---

func TestFinancial_EURDotDecimal(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"EUR prefix simple", "Total: €8.00 due", "€8.00"},
		{"EUR prefix thousands", "Amount: €1,000.00 charged", "€1,000.00"},
		{"EUR prefix space", "Cost € 250.00 billed", "€ 250.00"},
		{"EUR suffix", "Total 8.00€ due", "8.00€"},
		{"EUR suffix space", "Total 1,500.00 € billed", "1,500.00 €"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "FINANCIAL" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("FINANCIAL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

// --- ORG ULC/DAC/LLP tests ---

func TestOrganization_IrishCorporate(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"ULC", "Registered as Twitter International ULC in Ireland.", "Twitter International ULC"},
		{"DAC", "Filed by Acme Holdings DAC today.", "Acme Holdings DAC"},
		{"LLP", "Partners at Smith Jones LLP agreed.", "Smith Jones LLP"},
		{"Plc", "Shares in Royal Bank Plc.", "Royal Bank Plc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ORG" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ORG not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

// --- Irish Eircode tests ---

func TestAddress_IrishEircode(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"Dublin Eircode", "Address: Dublin 2, D02 AX07, Ireland", "D02 AX07"},
		{"Cork Eircode", "Located at T12 AB34 Cork", "T12 AB34"},
		{"Galway Eircode", "H91 E2F3 is the code", "H91 E2F3"},
		{"Dublin 6W Eircode", "Postal code D6W YF40", "D6W YF40"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestAddress_DublinDistrict(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"Dublin 2", "Located in Dublin 2 area", "Dublin 2"},
		{"Dublin 24", "Dublin 24\nIreland", "Dublin 24"},
		{"Dublin 6W", "Dublin 6W\nIreland", "Dublin 6W"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

// --- English street without number tests ---

func TestAddress_EnglishStreetNoNumber(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"Fenian St", "Fenian St\nDublin 2, D02 AX07\nIreland", "Fenian St"},
		{"Baker Street", "Baker Street\nLondon\nUnited Kingdom", "Baker Street"},
		{"Oak Lane", "Oak Lane\nDublin 4\nIreland", "Oak Lane"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestAddress_EnglishStreetNoNumber_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	// Should NOT match street-like patterns without address context
	cases := []string{
		"Main St is a popular name.",
		"Wall Street is a movie.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "ADDRESS" && (e.Text == "Main St" || e.Text == "Wall Street") {
				t.Errorf("ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- Real Twitter/X invoice test ---

func TestTwitterInvoice(t *testing.T) {
	s := DefaultScanner(nil)
	input := `Invoice
Invoice number A49E63AA-0002
Date of issue February 12, 2026
Date due February 12, 2026
Twitter International ULC
Fenian St
Dublin 2, D02 AX07
Ireland
VAT ID IE9834041A
Bill to John Smith
Austria
contact@example.com
€8.00 due February 12, 2026`

	entities := s.Scan(input)
	typeFound := make(map[string][]string)
	for _, e := range entities {
		typeFound[e.Type] = append(typeFound[e.Type], e.Text)
	}

	checks := []struct {
		typ  string
		desc string
	}{
		{"ID_NUMBER", "Invoice number / VAT ID"},
		{"DATE", "Written dates"},
		{"ORG", "Twitter International ULC"},
		{"ADDRESS", "Eircode / Dublin district / Fenian St"},
		{"PERSON", "Bill to name"},
		{"EMAIL", "Email address"},
		{"FINANCIAL", "EUR amount"},
	}
	for _, c := range checks {
		if _, ok := typeFound[c.typ]; !ok {
			t.Errorf("%s not detected (%s); entities: %v", c.typ, c.desc, entities)
		}
	}

	t.Logf("Detected entities: %v", entities)
}

func TestAustrianGuertel(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"Gürtel suffix", "Wohnt in Margaretengürtel 12"},
		{"Markt suffix", "Treffen am Fleischmarkt 7"},
		{"Graben compound", "Büro am Petersgraben 31"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ADDRESS" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ADDRESS not found in %q, got %v", tc.input, entities)
			}
		})
	}
}

// --- BIC/SWIFT tests ---

func TestBIC_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"BIC context", "BIC: BKAUATWW", "BKAUATWW"},
		{"SWIFT context", "SWIFT: GIBAATWWXXX", "GIBAATWWXXX"},
		{"BIC/SWIFT", "BIC/SWIFT COBADEFFXXX", "COBADEFFXXX"},
		{"BIC standalone AT", "Transfer via BKAUATWW bitte.", "BKAUATWW"},
		{"BIC standalone DE", "Code: COBADEFF", "COBADEFF"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "FINANCIAL" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("BIC not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestBIC_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The word HELLO is common.",
		"File FORMAT is important.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "FINANCIAL" && len(e.Text) >= 8 && len(e.Text) <= 11 {
				// Check if it looks like a BIC false positive
				t.Errorf("BIC false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- European bare amount tests ---

func TestFinancial_BareEuropeanAmounts(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// With thousand separator (no context needed)
		{"thousands comma", "Summe 2.544,70 bezahlt", "2.544,70"},
		{"thousands large", "Betrag: 1.250,00", "1.250,00"},
		{"thousands 2229", "Zahlung 2.229,00 erhalten", "2.229,00"},
		// Without separator but with financial context
		{"bare small", "Rechnung: Gebühr 65,00 fällig", "65,00"},
		{"bare medium", "E-Preis 94,70 pro Stück", "94,70"},
		{"bare large", "Leistung 2229,00 gesamt", "2229,00"},
		{"bare tax", "USt 20,00 Prozent", "20,00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "FINANCIAL" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("FINANCIAL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestFinancial_BareAmounts_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	// Bare amounts without financial context should NOT be detected
	cases := []string{
		"The score is 65,00 points.",
		"Temperature was 20,00 degrees.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "FINANCIAL" {
				t.Errorf("FINANCIAL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- IBAN context-triggered test ---

func TestIBAN_ContextTriggered(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"IBAN prefix", "IBAN: DE89 3704 0044 0532 0130 00"},
		{"IBAN no colon", "IBAN DE89 3704 0044 0532 0130 00"},
		{"iban lowercase prefix", "iban: AT611904300234573201"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "IBAN" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("IBAN not found in %q, got %v", tc.input, entities)
			}
		})
	}
}

// --- Driving school invoice test ---

func TestDrivingSchoolInvoice(t *testing.T) {
	s := DefaultScanner(nil)
	input := `Rechnung
Datum Menge Leistung USt% E-Preis G-Preis
24.05.24 1 B - LÖWE 20,00 2229,00 2229,00
27.11.24 1 Nichtantritt Theorie 20,00 65,00 65,00
08.01.25 1 Verwaltungsabgabe 20,00 94,70 94,70
Zahlungen
2.544,70
Datum Zahlung USt% Betrag
20.02.25 von cremul-Datei 20,00 2.229,00
IBAN: AT611904300234573201
BIC: BKAUATWW`

	entities := s.Scan(input)
	typeFound := make(map[string][]string)
	for _, e := range entities {
		typeFound[e.Type] = append(typeFound[e.Type], e.Text)
	}

	checks := []struct {
		typ  string
		desc string
	}{
		{"FINANCIAL", "amounts (bare European and BIC)"},
		{"IBAN", "IBAN"},
	}
	for _, c := range checks {
		if _, ok := typeFound[c.typ]; !ok {
			t.Errorf("%s not detected (%s); entities: %v", c.typ, c.desc, entities)
		}
	}

	// Verify specific amounts
	financials := typeFound["FINANCIAL"]
	wantAmounts := []string{"2.544,70", "2.229,00"}
	for _, w := range wantAmounts {
		found := false
		for _, f := range financials {
			if f == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Amount %q not in FINANCIAL entities: %v", w, financials)
		}
	}

	t.Logf("Detected entities: %v", entities)
}

// --- Benchmark ---

func BenchmarkScan100KB(b *testing.B) {
	// Build a ~100KB text with some PII scattered.
	base := "Der Patient Thomas Schmidt, geboren am 15.03.1990, wohnt in Hauptstraße 42, 10115 Berlin. " +
		"Seine E-Mail ist thomas.schmidt@example.com und seine Telefonnummer ist +49 170 4839201. " +
		"Die Rechnung über €1.250,00 wurde per IBAN DE89 3704 0044 0532 0130 00 bezahlt. " +
		"Das Wetter ist heute schön und die Temperatur beträgt 25 Grad. " +
		"Wir erwarten moderate Winde aus dem Nordwesten mit 15 km/h. "
	var sb []byte
	for len(sb) < 100*1024 {
		sb = append(sb, base...)
	}
	text := string(sb)

	s := DefaultScanner(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Scan(text)
	}
}
