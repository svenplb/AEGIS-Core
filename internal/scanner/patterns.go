package scanner

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// BuiltinScanners returns all built-in regex-based scanners.
func BuiltinScanners() []Scanner {
	var scanners []Scanner

	// Order matters for overlap: more specific patterns first.
	scanners = append(scanners, secretScanners()...)
	scanners = append(scanners, emailScanners()...)
	scanners = append(scanners, urlScanners()...)
	scanners = append(scanners, ibanScanners()...)
	scanners = append(scanners, creditCardScanners()...)
	scanners = append(scanners, ssnScanners()...)
	scanners = append(scanners, macAddressScanners()...)
	scanners = append(scanners, phoneScanners()...)
	scanners = append(scanners, dateScanners()...)
	scanners = append(scanners, ipScanners()...)
	scanners = append(scanners, medicalScanners()...)
	scanners = append(scanners, ageScanners()...)
	scanners = append(scanners, idNumberScanners()...)
	scanners = append(scanners, taxNumberScanners()...)
	scanners = append(scanners, businessIDScanners()...)
	scanners = append(scanners, orgScanners()...)
	scanners = append(scanners, financialScanners()...)
	scanners = append(scanners, addressScanners()...)
	scanners = append(scanners, personScanners()...)

	return scanners
}

// --- SSN ---

func ssnScanners() []Scanner {
	return []Scanner{
		// US SSN: 123-45-6789
		NewRegexScanner(
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"SSN", 0.95,
			WithValidator(func(s string) bool {
				area := s[:3]
				return area != "000" && area != "666" && area[0] != '9'
			}),
		),
		// German Sozialversicherungsnummer (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Sozialversicherungsnummer|SVN|SV-Nummer|Versicherungsnummer)[:\s]+(\d{2}\s?\d{6}\s?[A-Z]\s?\d{3})`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Swiss AHV: 756.1234.5678.97
		NewRegexScanner(
			regexp.MustCompile(`\b756\.\d{4}\.\d{4}\.\d{2}\b`),
			"SSN", 0.95,
		),
		// UK NINO: AB 12 34 56 C
		NewRegexScanner(
			regexp.MustCompile(`\b[A-CEGHJ-PR-TW-Z][A-CEGHJ-NPR-TW-Z]\s?\d{2}\s?\d{2}\s?\d{2}\s?[A-D]\b`),
			"SSN", 0.90,
		),
		// French INSEE: 1 85 12 75 108 042 36
		NewRegexScanner(
			regexp.MustCompile(`\b[12]\s?\d{2}\s?\d{2}\s?\d{2}\s?\d{3}\s?\d{3}\s?\d{2}\b`),
			"SSN", 0.85,
		),

		// --- New European national IDs ---

		// Polish PESEL: 11 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:PESEL|numer\s+PESEL)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Czech/Slovak Rodné číslo: XXXXXX/XXXX (context-triggered to avoid matching fractions/references)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:rodné\s+číslo|r\.?\s?č\.?|birth\s+number)[:\s]+(\d{6}/\d{3,4})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Swedish Personnummer: YYYYMMDD-XXXX
		NewRegexScanner(
			regexp.MustCompile(`\b(?:19|20)\d{6}[-+]\d{4}\b`),
			"SSN", 0.90,
		),
		// Danish CPR: DDMMYY-XXXX
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:CPR|CPR-nr\.?)[:\s]+(\d{6}-\d{4})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Finnish Henkilötunnus: DDMMYY-XXXC (separator: - + A)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:henkilötunnus|hetu)[:\s]+(\d{6}[-+A]\d{3}[A-Z0-9])\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Norwegian Fødselsnummer: DDMMYYXXXXX (11 digits, context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:fødselsnummer|f(?:ø|oe)dselsnr\.?|personnummer)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Italian Codice Fiscale: 16 alphanumeric (RSSMRA80A01H501U)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:codice\s+fiscale|C\.?F\.?)[:\s]+([A-Z]{6}\d{2}[A-Z]\d{2}[A-Z]\d{3}[A-Z])\b`),
			"SSN", 0.95,
			WithExtractGroup(1),
		),
		// Italian CF standalone (strict uppercase, 16 chars)
		NewRegexScanner(
			regexp.MustCompile(`\b[A-Z]{6}\d{2}[A-Z]\d{2}[A-Z]\d{3}[A-Z]\b`),
			"SSN", 0.80,
		),
		// Spanish DNI: 8 digits + letter
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:DNI|D\.?N\.?I\.?)[:\s]+(\d{8}[A-Z])\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Spanish NIE: X/Y/Z + 7 digits + letter
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:NIE|N\.?I\.?E\.?)[:\s]+([XYZ]\d{7}[A-Z])\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Portuguese NIF: 9 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:NIF|contribuinte)[:\s]+(\d{9})\b`),
			"SSN", 0.85,
			WithExtractGroup(1),
		),
		// Belgian Rijksregisternummer: XX.XX.XX-XXX.XX
		NewRegexScanner(
			regexp.MustCompile(`\b\d{2}\.\d{2}\.\d{2}-\d{3}\.\d{2}\b`),
			"SSN", 0.90,
		),
		// Dutch BSN: 9 digits (context-triggered, elfproef validation)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:BSN|burgerservicenummer)[:\s]+(\d{9})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
			WithValidator(validateBSN),
		),
		// Irish PPS: 7 digits + 1-2 letters
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:PPS|PPSN)[:\s]+(\d{7}[A-Z]{1,2})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Croatian OIB: 11 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:OIB)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Romanian CNP: 13 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:CNP|cod\s+numeric\s+personal)[:\s]+([1-8]\d{12})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Bulgarian EGN: 10 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:ЕГН|EGN)[:\s]+(\d{10})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Estonian Isikukood: 11 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:isikukood)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Latvian Personas kods: DDMMYY-XXXXX (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:personas\s+kods)[:\s]+(\d{6}-\d{5})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Lithuanian Asmens kodas: 11 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:asmens\s+kodas)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Greek AMKA: 11 digits (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:ΑΜΚΑ|AMKA)[:\s]+(\d{11})\b`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
	}
}

// validateBSN performs the Dutch elfproef (11-check) validation.
func validateBSN(s string) bool {
	if len(s) != 9 {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		d := int(s[i] - '0')
		if i < 8 {
			sum += d * (9 - i)
		} else {
			sum -= d // last digit is subtracted
		}
	}
	return sum > 0 && sum%11 == 0
}

// --- MEDICAL ---

func medicalScanners() []Scanner {
	return []Scanner{
		// ICD-10 codes (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Diagnose|ICD|diagnosis|diagnostic)[:\s]+([A-Z]\d{2}(?:\.\d{1,4})?)`),
			"MEDICAL", 0.90,
			WithExtractGroup(1),
		),
		// Blood pressure: 120/80 mmHg
		NewRegexScanner(
			regexp.MustCompile(`\b\d{2,3}/\d{2,3}\s?(?:mmHg|mm\s?Hg)\b`),
			"MEDICAL", 0.90,
		),
		// Lab values with units
		NewRegexScanner(
			regexp.MustCompile(`\b\d{1,4}(?:[.,]\d{1,2})?\s?(?:mg/dL|mmol/L|g/dL|mL/min|ng/mL|ng/L|µg/L|U/L|IU/L|pg/mL|µmol/L)\b`),
			"MEDICAL", 0.85,
		),
		// BMI values (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:BMI|Body Mass Index)[:\s]+(\d{2}(?:[.,]\d{1,2})?)`),
			"MEDICAL", 0.85,
			WithExtractGroup(1),
		),
		// ICD-10 codes standalone in parentheses: (I21.0), (E11.65)
		NewRegexScanner(
			regexp.MustCompile(`\(([A-Z]\d{2}(?:\.\d{1,4})?)\)`),
			"MEDICAL", 0.85,
			WithExtractGroup(1),
		),
	}
}

// --- AGE ---

func ageScanners() []Scanner {
	return []Scanner{
		// "X years old" / "X-year-old"
		NewRegexScanner(
			regexp.MustCompile(`\b(\d{1,3})\s?(?:-\s?)?(?:years?\s?(?:old)?|year-old)\b`),
			"AGE", 0.85,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// "X Jahre alt"
		NewRegexScanner(
			regexp.MustCompile(`\b(\d{1,3})\s?(?:Jahre?\s?(?:alt)?)\b`),
			"AGE", 0.85,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// Context-triggered: "age: X" / "Alter: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:age|Alter)[:\s]+(\d{1,3})\b`),
			"AGE", 0.80,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// Birth year: "born in 1990", "geboren 1985"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:born\s+(?:in\s+)?|geboren\s+(?:im\s+)?(?:Jahr\s+)?)((?:19|20)\d{2})\b`),
			"AGE", 0.80,
			WithExtractGroup(1),
		),
	}
}

// --- ID_NUMBER ---

func idNumberScanners() []Scanner {
	return []Scanner{
		// German Steuer-ID (context-triggered): 11 digits
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Steuer-?ID|Steueridentifikationsnummer|Tax\s?ID|TIN)[:\s]+(\d{11})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// German Personalausweis (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Personalausweis|Ausweis(?:nummer)?|ID\s?card)[:\s]+([A-Z0-9]{9,10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// German Reisepass (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Reisepass|Passport)[:\s]+([A-Z0-9]{9,10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// EU VAT numbers: 2-letter country code + 8-12 alphanumeric (must contain at least one digit)
		NewRegexScanner(
			regexp.MustCompile(`\b(AT|BE|BG|CY|CZ|DE|DK|EE|EL|ES|FI|FR|HR|HU|IE|IT|LT|LU|LV|MT|NL|PL|PT|RO|SE|SI|SK)[A-Z0-9]{8,12}\b`),
			"ID_NUMBER", 0.85,
			WithValidator(func(s string) bool {
				// Must contain at least one digit after country code to avoid matching words like ITALIENISCHES.
				for _, r := range s[2:] {
					if r >= '0' && r <= '9' {
						return true
					}
				}
				return false
			}),
		),
		// German Versichertennummer (insurance number, context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Versichertennummer|Versicherten-?Nr\.?|Versicherungsnr\.?)[:\s]+([A-Z]?\d{6,12})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// German Rentenversicherungsnummer (pension number, context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Rentenversicherungsnr\.?|Rentenversicherungsnummer|RVNR)[:\s]+(\d{2}\s?\d{6}\s?[A-Z]\s?\d{3})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Invoice/order/receipt with qualifier: "Invoice number X", "Order no. X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
			// DE
			`Invoice|Rechnung|Bill|Receipt|Order|Reference|Bestell|Auftrags`+
			`|Angebot|Auftragsbestätigung|Lieferschein|Gutschrift|Mahnung|Kostenvoranschlag`+
			// EN
			`|Quotation|Estimate|Delivery[ \t]+Note|Credit[ \t]+Note|Proforma|Purchase[ \t]+Order|Debit[ \t]+Note`+
			// FR
			`|Facture|Devis|Offre|Avoir|Bon[ \t]+de[ \t]+livraison|Note[ \t]+de[ \t]+crédit|Bon[ \t]+de[ \t]+commande`+
			// IT
			`|Fattura|Preventivo|Offerta|DDT|Nota[ \t]+di[ \t]+credito|Bolla[ \t]+di[ \t]+consegna|Sollecito`+
			// ES
			`|Presupuesto|Albarán|Nota[ \t]+de[ \t]+crédito|Nota[ \t]+de[ \t]+entrega|Orden[ \t]+de[ \t]+compra|Abono`+
			// NL
			`|Offerte|Creditnota|Pakbon|Afleverbon|Aanmaning|Orderbevestiging`+
			// PL
			`|Oferta|Zamówienie|Wycena|Nota[ \t]+kredytowa|Faktura[ \t]+korygująca`+
			// SE
			`|Offert|Följesedel|Kreditnota|Kreditfaktura|Påminnelse`+
			// PT
			`|Orçamento|Fatura|Nota[ \t]+de[ \t]+crédito`+
			// DK/NO
			`|Tilbud|Faktura|Kreditnota|Følgeseddel|Purring|Inkassokrav`+
			// FI
			`|Tarjous|Lasku|Hyvityslasku|Lähetysluettelo|Maksumuistutus`+
			// CZ/SK
			`|Nabídka|Dodací[ \t]+list|Dobropis|Upomínka|Ponuka|Dodací[ \t]+list|Ťarchopis`+
			// HU
			`|Ajánlat|Számla|Jóváírás|Szállítólevél|Fizetési[ \t]+felszólítás`+
			// RO
			`|Ofertă|Factură|Notă[ \t]+de[ \t]+credit|Aviz[ \t]+de[ \t]+expediție|Somație`+
			// BG
			`|Оферта|Фактура|Кредитно[ \t]+известие|Товарителница`+
			// HR
			`|Ponuda|Račun|Otpremnica|Odobrenje`+
			// SI
			`|Ponudba|Račun|Dobavnica|Dobropis`+
			// EL
			`|Προσφορά|Τιμολόγιο|Πιστωτικό[ \t]+σημείωμα|Δελτίο[ \t]+αποστολής`+
			// EE
			`|Pakkumine|Arve|Kreeditarve|Saateleht`+
			// LV
			`|Piedāvājums|Rēķins|Kredītrēķins|Pavadzīme`+
			// LT
			`|Pasiūlymas|Sąskaita|Kreditinė[ \t]+sąskaita|Važtaraštis`+
			`)[ \t]*(?:number|no\.?|num\.?|nr\.?|nummer|nº\.?|n°\.?|číslo|szám|#)[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Invoice/order/receipt compound forms: "Rechnungsnummer X", "Beleg-Nr. X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
			// DE compound forms
			`Rechnungsnummer|Rechnungs-?Nr\.?|Bestellnummer|Bestell-?Nr\.?`+
			`|Auftragsnummer|Auftrags-?Nr\.?|Referenz-?Nr\.?|Beleg-?Nr\.?`+
			`|Angebots-?Nr\.?|Angebotsnummer|Lieferschein-?Nr\.?|Lieferscheinnummer`+
			`|Gutschrift-?Nr\.?|Gutschrifts?nummer|Mahnungs?-?Nr\.?|Auftragsbestätigungs?-?Nr\.?`+
			// NL compound forms
			`|Factuurnummer|Factuur-?nr\.?|Offertenummer|Offerte-?nr\.?|Bestelnummer|Bestel-?nr\.?`+
			`|Creditnotanummer|Creditnota-?nr\.?|Ordernummer|Order-?nr\.?`+
			// SE/DK/NO compound forms
			`|Fakturanummer|Faktura-?nr\.?|Offertnummer|Offert-?nr\.?|Ordernummer|Order-?nr\.?`+
			`|Kreditnotanummer|Kreditnota-?nr\.?|Följesedelsnummer|Följesedels-?nr\.?`+
			`)[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Invoice/order with colon separator: "Invoice: X", "Reference: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
			`Invoice|Rechnung|Bill|Receipt|Order|Reference|Beleg`+
			`|Angebot|Auftragsbestätigung|Lieferschein|Gutschrift|Mahnung`+
			`|Quotation|Estimate|Credit[ \t]+Note|Debit[ \t]+Note`+
			`|Facture|Devis|Offre|Avoir|Sollecito|Preventivo|Offerta|DDT`+
			`|Presupuesto|Albarán|Abono|Offerte|Creditnota|Pakbon|Aanmaning`+
			`|Oferta|Zamówienie|Offert|Kreditnota|Følgeseddel|Följesedel`+
			`|Orçamento|Fatura|Tilbud|Faktura|Påminnelse|Purring|Inkassokrav`+
			// FI
			`|Tarjous|Lasku|Hyvityslasku|Maksumuistutus`+
			// CZ/SK
			`|Nabídka|Dobropis|Upomínka|Ponuka`+
			// HU
			`|Ajánlat|Számla|Jóváírás|Szállítólevél`+
			// RO
			`|Ofertă|Factură|Somație`+
			// BG
			`|Оферта|Фактура`+
			// HR/SI
			`|Ponuda|Račun|Otpremnica|Odobrenje|Ponudba|Dobavnica|Dobropis`+
			// EL
			`|Προσφορά|Τιμολόγιο`+
			// EE/LV/LT
			`|Pakkumine|Arve|Kreeditarve|Piedāvājums|Rēķins|Pasiūlymas|Sąskaita`+
			`)[ \t]*:[ \t]*([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// French invoice: "Facture N°: FA-2026-1234", "Numéro de facture: X"
		// Use [ \t] to prevent matching across newlines.
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Facture|Numéro[ \t]+de[ \t]+facture|N°[ \t]*(?:de[ \t]+)?facture)[ \t]*(?:N°|no\.?|nr\.?)?[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Italian invoice: "Fattura n.: FT-2026-0099", "Numero fattura: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Fattura|Numero[ \t]+fattura|N\.[ \t]*fattura)[ \t]*(?:n\.?|nr\.?|no\.?)?[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Dutch reference: "Referentienummer: NL-2026-5678", "Factuurnummer: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Referentienummer|Referentie-?nr\.?|Factuurnummer|Factuur-?nr\.?|Kenmerk)[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// German insurance/policy: "Versicherungsschein: WS-2026-887654", "Polizzennummer: X", "Aktenzeichen: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Versicherungsschein|Polizzen?-?(?:nummer|nr\.?)|Aktenzeichen|Vertrags?-?(?:nummer|nr\.?)|Schadens?-?(?:nummer|nr\.?))[: \t]+([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Invoice No: / Invoice No.: (with period in No.)
		NewRegexScanner(
			regexp.MustCompile(`(?i)Invoice\s+No\.?\s*:?\s*([A-Za-z0-9][\w.\-/]{2,})`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// "Invoice XXX-YYYY-ZZZZ" or "Rechnung 2026-001" (keyword + structured ID directly)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
				`Invoice|Rechnung|Facture|Fattura|Factuur|Faktura|Fatura`+
				`|Angebot|Lieferschein|Gutschrift|Mahnung|Auftragsbestätigung`+
				`|Quotation|Estimate|Offerte|Offert|Devis|Preventivo|Presupuesto|Tilbud|Orçamento`+
				`)[ \t]+([A-Za-z0-9]{2,4}[\-/]\d{2,6}(?:[\-/]\d{2,6})?)\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// "ref. NL-2026-5678", "Ref: XXX", "Referenz: XXX"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:ref\.?|referenz|referentie|référence|riferimento)[ \t:]+([A-Z]{2,4}-\d{4}-\d{4,6})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
	}
}

// --- ORG ---

func orgScanners() []Scanner {
	// CamelCase pattern: two or more capitalized segments joined without space (SynergyTech, MediaMarkt).
	camelCase := `[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+(?:[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+)+`

	// Name part for corporate names: allow CamelCase, abbreviations (2-6 uppercase), and normal names.
	corpNamePart := `(?:` + camelCase + `|[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ]{2,6}|` + nameComponent + `)(?:-` + nameComponent + `)?`

	// Use [ \t]+ instead of \s+ to prevent matching across newlines.
	sp := `[ \t]+`

	// Corporate suffixes (German) — allow & connector between name parts
	corpDE := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `(?:GmbH|AG|SE|KG|OHG|KGaA|UG|e\.G\.|e\.V\.)\b`
	// Corporate suffixes (International) — allow & connector between name parts
	corpIntl := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `(?:Ltd|Inc|Corp|LLC|PLC|Plc|SA|SAS|SARL|SpA|SRL|BV|NV|ULC|DAC|LLP)\.?\b`
	// German institutions: Universitätsklinikum/Uniklinik/Universität/Klinikum + Name
	deInstitution := `(?:Universitätsklinikum|Uniklinik|Universität|Klinikum)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// Klinik + preposition + Name
	klinikPrep := `Klinik` + sp + `(?:am|für|an` + sp + `der|im)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// French hospitals
	frHospital := `(?:Hôpital|CHU)` + sp + namePattern + `(?:[ \t\-]+` + namePattern + `)*`
	// Italian hospitals
	itHospital := `(?:Ospedale|Policlinico)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// Spanish hospitals
	esHospital := `Hospital` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// German insurance: AOK + Name
	aok := `AOK` + sp + namePattern
	// German government: Deutsche Rentenversicherung (+ optional Name)
	drv := `Deutsche` + sp + `Rentenversicherung(?:` + sp + namePattern + `)?`
	// University medical centers: Name UMC | UMC Name
	umcSuffix := namePattern + sp + `UMC`
	umcPrefix := `UMC` + sp + namePattern

	// Nordic corporate suffixes: AB (SE), AS (NO), A/S (DK), Oy/Oyj (FI), HF (IS)
	corpNordic := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `(?:AB|AS|A/S|Oy|Oyj|HF)\b`

	// Eastern European + Portuguese/Brazilian corporate suffixes (no \b — suffixes end with ".")
	// Kft. (HU), s.r.o. (CZ/SK), d.o.o. (HR/SI), Lda. (PT), Ltda. (BR)
	corpEasternDot := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `(?:Kft\.|s\.r\.o\.|d\.o\.o\.|Lda\.|Ltda\.)`
	// EOOD (BG) — word boundary OK since no trailing dot
	corpEasternWord := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `EOOD\b`

	// German banking names: compound words ending in -bank (Kantonalbank, Volksbank, Raiffeisenbank)
	bankCompound := `[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+bank\b`
	corpBank := `(?:` + namePattern + sp + `)*` + bankCompound

	// Polish sp. z o.o. (spaces in suffix need special handling)
	corpPolish := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `sp\.\s?z\s?o\.o\.`

	// Asian "Co., Ltd." / "Co. Ltd" suffix (JP, KR, TW, HK)
	corpCoLtd := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `Co\.,?` + sp + `Ltd\.?`
	// Indian "Pvt. Ltd." / "Pvt Ltd" suffix
	corpPvtLtd := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `Pvt\.?` + sp + `Ltd\.?`
	// Mexican "S.A. de C.V." suffix
	corpMexican := corpNamePart + `(?:` + sp + `(?:&` + sp + `)?` + corpNamePart + `)*` + sp + `S\.A\.\s?de\s?C\.V\.`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(corpDE), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpIntl), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpNordic), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpEasternDot), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpEasternWord), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpBank), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(corpPolish), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpCoLtd), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpPvtLtd), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpMexican), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(deInstitution), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(klinikPrep), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(frHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(itHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(esHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(aok), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(drv), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(umcSuffix), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(umcPrefix), "ORG", 0.85),
	}
}

// --- MAC_ADDRESS ---

func macAddressScanners() []Scanner {
	return []Scanner{
		// Standard: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
		NewRegexScanner(
			regexp.MustCompile(`\b([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`),
			"MAC_ADDRESS", 0.95,
		),
		// Cisco format: XXXX.XXXX.XXXX
		NewRegexScanner(
			regexp.MustCompile(`\b[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\b`),
			"MAC_ADDRESS", 0.90,
		),
	}
}

// --- PERSON ---

// Unicode-aware name component: uppercase letter followed by lowercase letters,
// including diacritics (Müller, Ñoño, Ólafsson, etc.).
// Hyphenated names supported: Jean-Pierre, Müller-Schmidt.
const nameComponent = `[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+`
const namePattern = nameComponent + `(?:-` + nameComponent + `)?`

// Name particles for multi-part surnames (de Groot, van der Berg, von Stein, etc.)
const nameParticle = `(?:de|van|der|von|di|del|della|le|la|da|dos|das|du|ten|ter|het)`

// Full name: 2-4 name components with optional particles between them.
const fullName = namePattern + `(?:[ \t]+(?:` + nameParticle + `[ \t]+)*` + namePattern + `){1,3}`

func personScanners() []Scanner {
	// Context-triggered: keyword + CapFirst CapLast
	// Longer/more specific patterns first to avoid partial matches.
	triggers := []string{
		// Multi-word triggers first
		`Dr\.\s?med\.`, `de\s+heer`,
		`mein Freund`, `meine Freundin`,
		`meinen Patienten`, `meiner Patientin`,
		`my friend`, `my colleague`, `my patient`,
		`mon ami`, `mon amie`,
		// Family triggers (EN, DE, FR, ES)
		`his wife`, `her husband`, `his spouse`, `her spouse`,
		`seine Frau`, `ihr Mann`, `Ehefrau`, `Ehemann`,
		`son épouse`, `sa femme`, `son mari`,
		`su esposa`, `su esposo`,
		// German role triggers
		`Antragsteller(?:in)?`, `Sachbearbeiter(?:in)?`, `Bearbeiter(?:in)?`,
		`Konsiliarius`,
		`Leiter(?:in)?`, `Geschäftsführer(?:in)?`, `Inhaber(?:in)?`,
		`Direktor(?:in)?`, `Vorstand`, `Vorsitzende[r]?`,
		`Mitarbeiter(?:in)?`, `Angestellte[r]?`,
		`Vorgesetzte[r]?`,
		// German medical role triggers
		`Oberarzt`, `Oberärztin`, `Chefarzt`, `Chefärztin`,
		`Arzt`, `Ärztin`, `Krankenschwester`, `Pfleger(?:in)?`,
		// German legal/professional triggers
		`Rechtsanwalt`, `Rechtsanwältin`, `Notar(?:in)?`,
		`Steuerberater(?:in)?`, `Buchhalter(?:in)?`,
		// English role triggers
		`Employee`, `Supervisor`, `Manager`,
		`Attorney`, `Solicitor`, `Barrister`,
		`Accountant`, `Nurse`,
		// Legal triggers (EN, DE)
		`plaintiff`, `defendant`, `witness`,
		`Kläger(?:in)?`, `Beklagte[r]?`,
		`Zeuge`, `Zeugin`,
		// Titles (more specific first)
		`Dott\.?\s?ssa`, `Dott\.?`, `Dra\.?`,
		`Prof\.?`, `Dr\.?`,
		// German
		`Herr`, `Frau`, `Patient(?:in)?`, `Kollege`, `Kollegin`,
		// French
		`Monsieur`, `Madame`, `Mademoiselle`,
		// English
		`Mr\.?`, `Mrs\.?`, `Ms\.?`, `colleague`,
		// Dutch
		`Meneer`, `Mevrouw`,
		// Italian
		`Signor(?:a)?`,
		// Spanish
		`Señor(?:a)?`,
		// Polish
		`Pan`, `Pani`,
		// Czech
		`Pán`, `Paní`,
		// Finnish
		`Herra`, `Rouva`,
		// Romanian
		`Domnul`, `Doamna`,
		// Croatian
		`Gospodin`, `Gospođa`,
		// Portuguese
		`Senhor(?:a)?`,
		// Greek (Latin transliteration)
		`Kyrios`, `Kyria`,
	}

	triggerGroup := `(?:` + strings.Join(triggers, `|`) + `)`
	// Use (?i:...) only for the trigger group, keep name pattern case-sensitive.
	// Allow optional colon/comma between trigger and name (e.g. "Antragsteller: Thomas Schmidt").
	contextPattern := `(?i:` + triggerGroup + `)[: \t]+(` + fullName + `)`

	// Verb-triggered: "told/asked/called/emailed Name Name"
	verbs := `(?i:told|asked|called|emailed|contacted|met|visited|informed)`
	verbPattern := verbs + `[ \t]+(` + fullName + `)`

	// Maiden name: "geb. Müller", "geboren Weber"
	maidenPattern := `(?i:geb(?:oren(?:e)?)?\.)[ \t]+(` + namePattern + `)`

	// Billing/invoice label → name (allows newline between label and name)
	billingTrigger := `(?i:Bill\s+to|Billed\s+to|Invoice\s+to|Sold\s+to|Ship\s+to|Deliver\s+to|Attn\.?|Attention)`
	billingPattern := billingTrigger + `[\s:]+(` + fullName + `)`

	return []Scanner{
		NewRegexScanner(
			regexp.MustCompile(contextPattern),
			"PERSON", 0.95,
			WithExtractGroup(1),
		),
		NewRegexScanner(
			regexp.MustCompile(verbPattern),
			"PERSON", 0.85,
			WithExtractGroup(1),
		),
		NewRegexScanner(
			regexp.MustCompile(maidenPattern),
			"PERSON", 0.85,
			WithExtractGroup(1),
		),
		NewRegexScanner(
			regexp.MustCompile(billingPattern),
			"PERSON", 0.90,
			WithExtractGroup(1),
		),
	}
}

// --- EMAIL ---

func emailScanners() []Scanner {
	// RFC 5322 simplified with unicode support for DACH region.
	pattern := `[a-zA-Z0-9._%+\-àáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`
	return []Scanner{
		NewRegexScanner(regexp.MustCompile(pattern), "EMAIL", 0.99),
	}
}

// --- PHONE ---

// ibanPrefixRe matches the leading portion of an IBAN (country code + check digits +
// space-separated groups) that may appear before a phone-like digit sequence.
var ibanPrefixRe = regexp.MustCompile(`[A-Z]{2}\d{2}(?:[\s\-][\dA-Z]{4})*[\s\-]?[\dA-Z]{0,4}$`)

// phoneNotInIBAN rejects a phone match if it sits inside an IBAN-like structure.
// It looks back up to 40 bytes from the match start for an IBAN prefix.
func phoneNotInIBAN(fullText string, start, end int) bool {
	lookback := 40
	from := start - lookback
	if from < 0 {
		from = 0
	}
	prefix := fullText[from:start]
	if ibanPrefixRe.MatchString(prefix) {
		return false
	}
	return true
}

func phoneScanners() []Scanner {
	// International format with + prefix.
	// EU27 + UK, CH, NO, IS + US/CA, AU.
	// Supports separators: space, dash, dot, or none.
	// Use [ \t] instead of \s to prevent matching across newlines.
	intlCodes := `(?:` +
		`49|43|41|33|39|34|31|32|351|48|46|358|45|47|353|44` + // existing (DE,AT,CH,FR,IT,ES,NL,BE,PT,PL,SE,FI,DK,NO,IE,UK)
		`|1|61` + // US/CA, AU
		`|30|36|40|356|357|359` + // GR, HU, RO, MT, CY, BG
		`|370|371|372|385|386` + // LT, LV, EE, HR, SI
		`|420|421|352|354` + // CZ, SK, LU, IS
		`|55|52|81|82|91|65|90|27` + // BR, MX, JP, KR, IN, SG, TR, ZA
		`)`
	intl := `\+` + intlCodes + `[\- \t]?(?:\(0\))?[\- \t]?[\d][\d \t.\-]{6,14}\d`

	// Generic 00-prefix international
	generic00 := `00\d{1,3}[ \t.\-]?\d[\d \t.\-]{6,14}\d`

	// German local: 0XXX XXXXXXX (word boundary prevents matching inside IDs like WS-2026-887654)
	deLocal := `\b0[1-9]\d{1,4}[ \t.\-/]?\d[\d \t.\-]{4,10}\d`

	// UK local: 020 XXXX XXXX, 07XXX XXXXXX
	ukLocal := `0[1-9]\d{2,4}[ \t]?\d{3,4}[ \t]?\d{3,4}`

	// French local: 0X XX XX XX XX
	frLocal := `0[1-9](?:[ \t.]?\d{2}){4}`

	// US/CA: (XXX) XXX-XXXX or XXX-XXX-XXXX
	usPhone := `(?:\(\d{3}\)[ \t]?|\d{3}[\-.])\d{3}[\-.]\d{4}`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(intl), "PHONE", 0.95, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(generic00), "PHONE", 0.90, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(usPhone), "PHONE", 0.90, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(frLocal), "PHONE", 0.85, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(ukLocal), "PHONE", 0.85, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(deLocal), "PHONE", 0.85, WithContextValidator(phoneNotInIBAN)),
	}
}

// --- IBAN ---

func ibanScanners() []Scanner {
	// Generic IBAN: 2 letters + 2 digits + 8-30 alphanumeric (with optional spaces/dashes).
	// Use [ \t] instead of \s to prevent matching across newlines.
	pattern := `\b[A-Z]{2}\d{2}[ \t\-]?[\dA-Z]{4}[ \t\-]?[\dA-Z]{4}(?:[ \t\-]?[\dA-Z]{4}){1,7}(?:[ \t\-]?[\dA-Z]{1,4})?\b`

	// Context-triggered: "IBAN: AT61 1904 ..." or "IBAN AT61..."
	// Allows newline between "IBAN:" and the number, but not within the number itself.
	contextPattern := `(?i)IBAN[:\s]+([A-Z]{2}\d{2}[ \t\-]?[\dA-Z]{4}[ \t\-]?[\dA-Z]{4}(?:[ \t\-]?[\dA-Z]{4}){1,7}(?:[ \t\-]?[\dA-Z]{1,4})?)`

	return []Scanner{
		NewRegexScanner(
			regexp.MustCompile(pattern),
			"IBAN", 0.99,
			WithValidator(validateIBAN),
		),
		NewRegexScanner(
			regexp.MustCompile(contextPattern),
			"IBAN", 0.95,
			WithExtractGroup(1),
		),
	}
}

// validateIBAN performs MOD-97 checksum validation.
func validateIBAN(s string) bool {
	// Remove spaces and dashes.
	clean := strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return -1
		}
		return r
	}, s)

	if len(clean) < 5 || len(clean) > 34 {
		return false
	}

	// Check format: 2 letters + 2 digits + rest alphanumeric.
	for i, r := range clean {
		if i < 2 {
			if !unicode.IsUpper(r) {
				return false
			}
		} else if i < 4 {
			if !unicode.IsDigit(r) {
				return false
			}
		} else {
			if !unicode.IsUpper(r) && !unicode.IsDigit(r) {
				return false
			}
		}
	}

	// Move first 4 chars to end.
	rearranged := clean[4:] + clean[:4]

	// Convert letters to numbers: A=10, B=11, ..., Z=35.
	var numStr strings.Builder
	for _, r := range rearranged {
		if unicode.IsUpper(r) {
			numStr.WriteString(big.NewInt(int64(r - 'A' + 10)).String())
		} else {
			numStr.WriteRune(r)
		}
	}

	n := new(big.Int)
	n.SetString(numStr.String(), 10)
	mod := new(big.Int)
	mod.Mod(n, big.NewInt(97))

	return mod.Int64() == 1
}

// --- CREDIT CARD ---

func creditCardScanners() []Scanner {
	// Visa (16 digits): 4xxx xxxx xxxx xxxx
	// Mastercard (16 digits): 5[1-5]xx or 2[2-7]xx
	visa := `\b4\d{3}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`
	mc := `\b(?:5[1-5]\d{2}|2[2-7]\d{2})[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`

	// Amex (15 digits): 3[47]xx xxxxxx xxxxx
	amex := `\b3[47]\d{2}[\s\-]?\d{6}[\s\-]?\d{5}\b`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(visa), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
		NewRegexScanner(regexp.MustCompile(mc), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
		NewRegexScanner(regexp.MustCompile(amex), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
	}
}

// validateLuhn performs the Luhn algorithm check.
func validateLuhn(s string) bool {
	// Extract digits only.
	var digits []int
	for _, r := range s {
		if unicode.IsDigit(r) {
			digits = append(digits, int(r-'0'))
		}
	}

	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}

	return sum%10 == 0
}

// --- DATE ---

func dateScanners() []Scanner {
	// DD.MM.YYYY, DD/MM/YYYY, DD-MM-YYYY
	dateCore := `\b(?:0[1-9]|[12]\d|3[01])[./\-](?:0[1-9]|1[0-2])[./\-](?:19|20)\d{2}\b`

	// Written English dates: "February 12, 2026" or "Feb 12, 2026"
	enMonths := `(?:January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Sept|Oct|Nov|Dec)\.?`
	// US format: January 15, 2026
	enDateWritten := enMonths + `[ \t]+\d{1,2},?[ \t]+(?:19|20)\d{2}`
	// International English format: 15 January 2026
	enDateDayFirst := `\d{1,2}[ \t]+` + enMonths + `[ \t]+(?:19|20)\d{2}`

	// Written German dates: "12. Februar 2026", "1. März 1990"
	deMonths := `(?:Januar|Februar|März|April|Mai|Juni|Juli|August|September|Oktober|November|Dezember)`
	deDateWritten := `\d{1,2}\.[ \t]+` + deMonths + `[ \t]+(?:19|20)\d{2}`

	// Written French dates: "12 février 2026"
	frMonths := `(?:janvier|février|mars|avril|mai|juin|juillet|août|septembre|octobre|novembre|décembre)`
	frDateWritten := `\d{1,2}[ \t]+` + frMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Spanish dates: "12 de febrero de 2026"
	esMonths := `(?:enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)`
	esDateWritten := `\d{1,2}[ \t]+(?:de[ \t]+)?` + esMonths + `[ \t]+(?:de[ \t]+)?(?:19|20)\d{2}`

	// Written Italian dates: "12 febbraio 2026"
	itMonths := `(?:gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)`
	itDateWritten := `\d{1,2}[ \t]+` + itMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Dutch dates: "12 februari 2026"
	nlMonths := `(?:januari|februari|maart|april|mei|juni|juli|augustus|september|oktober|november|december)`
	nlDateWritten := `\d{1,2}[ \t]+` + nlMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Polish dates: "12 lutego 2026"
	plMonths := `(?:stycznia|lutego|marca|kwietnia|maja|czerwca|lipca|sierpnia|września|października|listopada|grudnia)`
	plDateWritten := `\d{1,2}[ \t]+` + plMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Swedish dates: "12 februari 2026"
	seMonths := `(?:januari|februari|mars|april|maj|juni|juli|augusti|september|oktober|november|december)`
	seDateWritten := `\d{1,2}[ \t]+` + seMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Portuguese dates: "12 de fevereiro de 2026"
	ptMonths := `(?:janeiro|fevereiro|março|abril|maio|junho|julho|agosto|setembro|outubro|novembro|dezembro)`
	ptDateWritten := `\d{1,2}[ \t]+(?:de[ \t]+)?` + ptMonths + `[ \t]+(?:de[ \t]+)?(?:19|20)\d{2}`

	// Written Czech dates: "12. února 2026"
	czMonths := `(?:ledna|února|března|dubna|května|června|července|srpna|září|října|listopadu|prosince|leden|únor|březen|duben|květen|červen|červenec|srpen|říjen|listopad|prosinec)`
	czDateWritten := `\d{1,2}\.?[ \t]+` + czMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Slovak dates: "12. februára 2026"
	skMonths := `(?:januára?|februára?|marca?|apríla?|mája?|júna?|júla?|augusta?|septembra?|októbra?|novembra?|decembra?|január|február|marec|apríl|máj|jún|júl|august|september|október|november|december)`
	skDateWritten := `\d{1,2}\.?[ \t]+` + skMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Hungarian dates: "2026. február 12." (year first!)
	huMonths := `(?:január|február|március|április|május|június|július|augusztus|szeptember|október|november|december)`
	huDateWritten := `(?:19|20)\d{2}\.?[ \t]+` + huMonths + `[ \t]+\d{1,2}\.?`

	// Written Romanian dates: "12 februarie 2026"
	roMonths := `(?:ianuarie|februarie|martie|aprilie|mai|iunie|iulie|august|septembrie|octombrie|noiembrie|decembrie)`
	roDateWritten := `\d{1,2}[ \t]+` + roMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Bulgarian dates: "12 февруари 2026"
	bgMonths := `(?:януари|февруари|март|април|май|юни|юли|август|септември|октомври|ноември|декември)`
	bgDateWritten := `\d{1,2}[ \t]+` + bgMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Croatian dates: "12. veljače 2026"
	hrMonths := `(?:siječnja|veljače|ožujka|travnja|svibnja|lipnja|srpnja|kolovoza|rujna|listopada|prosinca|siječanj|veljača|ožujak|travanj|svibanj|lipanj|srpanj|kolovoz|rujan|listopad|studeni|prosinac)`
	hrDateWritten := `\d{1,2}\.?[ \t]+` + hrMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Slovenian dates: "12. februar 2026"
	siMonths := `(?:januar|februar|marec|april|maj|junij|julij|avgust|september|oktober|november|december)`
	siDateWritten := `\d{1,2}\.?[ \t]+` + siMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Greek dates: "12 Φεβρουαρίου 2026"
	elMonths := `(?:Ιανουαρίου|Φεβρουαρίου|Μαρτίου|Απριλίου|Μαΐου|Ιουνίου|Ιουλίου|Αυγούστου|Σεπτεμβρίου|Οκτωβρίου|Νοεμβρίου|Δεκεμβρίου|Ιανουάριος|Φεβρουάριος|Μάρτιος|Απρίλιος|Μάιος|Ιούνιος|Ιούλιος|Αύγουστος|Σεπτέμβριος|Οκτώβριος|Νοέμβριος|Δεκέμβριος)`
	elDateWritten := `\d{1,2}[ \t]+` + elMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Finnish dates: "12. helmikuuta 2026"
	fiMonths := `(?:tammikuuta?|helmikuuta?|maaliskuuta?|huhtikuuta?|toukokuuta?|kesäkuuta?|heinäkuuta?|elokuuta?|syyskuuta?|lokakuuta?|marraskuuta?|joulukuuta?|tammikuu|helmikuu|maaliskuu|huhtikuu|toukokuu|kesäkuu|heinäkuu|elokuu|syyskuu|lokakuu|marraskuu|joulukuu)`
	fiDateWritten := `\d{1,2}\.?[ \t]+` + fiMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Estonian dates: "12. veebruar 2026"
	eeMonths := `(?:jaanuar|veebruar|märts|aprill|mai|juuni|juuli|august|september|oktoober|november|detsember)`
	eeDateWritten := `\d{1,2}\.?[ \t]+` + eeMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Latvian dates: "12. februāris 2026"
	lvMonths := `(?:janvāris|februāris|marts|aprīlis|maijs|jūnijs|jūlijs|augusts|septembris|oktobris|novembris|decembris)`
	lvDateWritten := `\d{1,2}\.?[ \t]+` + lvMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Lithuanian dates: "12 vasario 2026"
	ltMonths := `(?:sausio|vasario|kovo|balandžio|gegužės|birželio|liepos|rugpjūčio|rugsėjo|spalio|lapkričio|gruodžio|sausis|vasaris|kovas|balandis|gegužė|birželis|liepa|rugpjūtis|rugsėjis|spalis|lapkritis|gruodis)`
	ltDateWritten := `\d{1,2}\.?[ \t]+` + ltMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Danish dates: "12. februar 2026"
	dkMonths := `(?:januar|februar|marts|april|maj|juni|juli|august|september|oktober|november|december)`
	dkDateWritten := `\d{1,2}\.?[ \t]+` + dkMonths + `[ \t]+(?:19|20)\d{2}`

	// Written Norwegian dates: "12. februar 2026"
	noMonths := `(?:januar|februar|mars|april|mai|juni|juli|august|september|oktober|november|desember)`
	noDateWritten := `\d{1,2}\.?[ \t]+` + noMonths + `[ \t]+(?:19|20)\d{2}`

	// All month names across ALL supported languages.
	allMonths := `(?:` + enMonths + `|` + deMonths + `|` + frMonths + `|` + esMonths + `|` + itMonths +
		`|` + nlMonths + `|` + plMonths + `|` + seMonths + `|` + ptMonths +
		`|` + czMonths + `|` + skMonths + `|` + huMonths + `|` + roMonths + `|` + bgMonths +
		`|` + hrMonths + `|` + siMonths + `|` + elMonths + `|` + fiMonths +
		`|` + eeMonths + `|` + lvMonths + `|` + ltMonths + `|` + dkMonths + `|` + noMonths + `)`

	// Bare month + 4-digit year: "Februar 2026", "March 2025", "janvier 2024"
	bareMonthYear := `(?i)\b` + allMonths + `[ \t]+(?:19|20)\d{2}\b`

	// Month + short/full year (context-triggered): "Leistungszeitraum: November 25"
	monthYear := `(?i)(?:Leistungszeitraum|Abrechnungszeitraum|Zeitraum|Abrechnungsmonat|Billing[ \t]+period|Period|Mois)[: \t]+(` + allMonths + `[ \t]+\d{2,4})`

	// ISO format: YYYY-MM-DD
	dateISO := `\b(?:19|20)\d{2}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])\b`

	// US format: MM/DD/YYYY — only matches when day > 12 to avoid ambiguity with EU DD/MM/YYYY
	dateUS := `\b(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])/((?:19|20)\d{2})\b`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(dateCore), "DATE", 0.90),
		NewRegexScanner(regexp.MustCompile(enDateWritten), "DATE", 0.90),
		NewRegexScanner(regexp.MustCompile(enDateDayFirst), "DATE", 0.90),
		NewRegexScanner(regexp.MustCompile(deDateWritten), "DATE", 0.90),
		NewRegexScanner(regexp.MustCompile(frDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(esDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(itDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(nlDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(plDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(seDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(ptDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(czDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(skDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(huDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(roDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(bgDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(hrDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(siDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(elDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(fiDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(eeDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(lvDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(ltDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(dkDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(noDateWritten), "DATE", 0.85),
		NewRegexScanner(regexp.MustCompile(bareMonthYear), "DATE", 0.80),
		NewRegexScanner(
			regexp.MustCompile(monthYear),
			"DATE", 0.85,
			WithExtractGroup(1),
		),
		NewRegexScanner(regexp.MustCompile(dateISO), "DATE", 0.90),
		NewRegexScanner(
			regexp.MustCompile(dateUS),
			"DATE", 0.85,
			WithValidator(func(s string) bool {
				parts := strings.SplitN(s, "/", 3)
				if len(parts) != 3 {
					return false
				}
				day, err := strconv.Atoi(parts[1])
				if err != nil {
					return false
				}
				return day > 12
			}),
		),
	}
}

// --- URL ---

func urlScanners() []Scanner {
	pattern := `https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`
	return []Scanner{
		NewRegexScanner(regexp.MustCompile(pattern), "URL", 0.95),
	}
}

// --- IP_ADDRESS ---

func ipScanners() []Scanner {
	// IPv4
	ipv4 := `\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`

	// IPv6 (simplified: full form and common abbreviations)
	ipv6 := `(?:` +
		`(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}` + // full
		`|(?:[0-9a-fA-F]{1,4}:){1,7}:` + // trailing ::
		`|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}` + // :: with one group
		`|::1` + // loopback
		`|::` + // unspecified
		`)`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(ipv4), "IP_ADDRESS", 0.90, WithValidator(validateIPv4)),
		NewRegexScanner(regexp.MustCompile(ipv6), "IP_ADDRESS", 0.90),
	}
}

func validateIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) > 1 && p[0] == '0' {
			return false // reject leading zeros like 01.02.03.04
		}
	}
	return true
}

// --- FINANCIAL ---

func financialScanners() []Scanner {
	// EUR format: €1.500,00 or 1.500,00 € or 1.500,00€
	eurPrefix := `€\s?\d{1,3}(?:\.\d{3})*,\d{2}`
	eurSuffix := `\d{1,3}(?:\.\d{3})*,\d{2}\s?€`

	// USD/GBP format: $2,500.00 or £2,500.00
	usdGbp := `[$£]\s?\d{1,3}(?:,\d{3})*\.\d{2}`

	// CHF: CHF 1'500.00, CHF 1,500.00, or CHF 1500.00
	chf := `CHF\s?\d{1,3}(?:[,'\x{2019}]\d{3})*\.\d{2}`

	// Common currency codes with dot-decimal format: AUD 4,500.00
	currencyCodeDot := `(?:AUD|CAD|SGD|NZD|HKD)\s\d{1,3}(?:,\d{3})*\.\d{2}`

	// EUR international format (dot decimal): €8.00, €1,000.00 (used in Ireland, English contexts)
	eurDotPrefix := `€\s?\d{1,3}(?:,\d{3})*\.\d{2}`
	eurDotSuffix := `\d{1,3}(?:,\d{3})*\.\d{2}\s?€`

	// EUR with thousand separator but no decimals: €9.500, 9.500 €
	// Requires at least one dot-separated group to distinguish from bare amounts.
	eurThousandNodecPrefix := `€\s?\d{1,3}(?:\.\d{3})+\b`
	eurThousandNodecSuffix := `\b\d{1,3}(?:\.\d{3})+\s?€`

	// European amounts WITH thousand separator but no symbol: 2.544,70, 1.250,00
	// Distinctive enough to not need context (dot-thousand + comma-decimal + exactly 2 decimals).
	eurBareThousands := `\b\d{1,3}(?:\.\d{3})+,\d{2}\b`

	// European amounts WITHOUT symbol and without thousand separator: 65,00, 94,70, 2229,00
	// Requires financial context nearby to avoid false positives.
	eurBare := `\b\d{2,6},\d{2}\b`

	// BIC/SWIFT codes (context-triggered): BKAUATWW, GIBAATWWXXX
	bicContext := `(?i)(?:BIC|SWIFT|BIC/SWIFT)[:\s/]+([A-Z]{6}[A-Z0-9]{2}(?:[A-Z0-9]{3})?)`

	// BIC/SWIFT standalone with known EU country codes
	bicStandalone := `\b[A-Z]{4}(?:AT|DE|CH|FR|IT|ES|NL|BE|IE|GB|LU|PT|PL|CZ|HU|SK|SI|HR|BG|RO|LT|LV|EE|FI|SE|DK|NO|LI|MT|CY|GR)[A-Z0-9]{2}(?:[A-Z0-9]{3})?\b`

	// Polish Złoty: 1 500,00 zł or 1500.00 PLN
	plnSuffix := `\d{1,3}(?:[\s.]\d{3})*,\d{2}\s?(?:zł|PLN)\b`
	plnCode := `\bPLN\s?\d{1,3}(?:[\s.]\d{3})*,\d{2}`

	// Czech Koruna: 15 000,00 Kč or 15000 CZK
	czkSuffix := `\d{1,3}(?:[\s.]\d{3})*,?\d{0,2}\s?(?:Kč|CZK)\b`

	// Hungarian Forint: 1 500 000 Ft or HUF
	hufSuffix := `\d{1,3}(?:[\s.]\d{3})*\s?(?:Ft|HUF)\b`

	// Romanian Leu: 15.000,00 lei or RON
	ronSuffix := `\d{1,3}(?:\.\d{3})*,\d{2}\s?(?:lei|RON)\b`

	// Swedish/Norwegian/Danish Krona/Krone: 15 000,00 kr
	sekSuffix := `\d{1,3}(?:[\s.]\d{3})*,\d{2}\s?(?:kr\.?|SEK|NOK|DKK)\b`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(eurPrefix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(eurSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(eurDotPrefix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(eurDotSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(usdGbp), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(chf), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(currencyCodeDot), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(plnSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(plnCode), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(czkSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(hufSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(ronSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(sekSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(eurThousandNodecPrefix), "FINANCIAL", 0.85),
		NewRegexScanner(regexp.MustCompile(eurThousandNodecSuffix), "FINANCIAL", 0.85),
		NewRegexScanner(regexp.MustCompile(eurBareThousands), "FINANCIAL", 0.85),
		NewRegexScanner(
			regexp.MustCompile(eurBare),
			"FINANCIAL", 0.75,
			WithContextValidator(financialContext),
		),
		NewRegexScanner(
			regexp.MustCompile(bicContext),
			"FINANCIAL", 0.95,
			WithExtractGroup(1),
		),
		NewRegexScanner(regexp.MustCompile(bicStandalone), "FINANCIAL", 0.85),
	}
}

// --- ADDRESS ---

func addressScanners() []Scanner {
	// Use [ \t] instead of \s to prevent matching across newlines.

	// House number with optional letter and Austrian/Swiss apartment notation (5/2/3)
	houseNum := `\d{1,4}[a-zA-Z]?(?:/\d{1,4})*`

	// German/Austrian street suffixes (compound form: Gartenstraße, Margaretengürtel, Fleischmarkt)
	deSuffixes := `(?:straße|strasse|str\.|weg|platz|allee|gasse|ring|damm|ufer|kai|quai|gürtel|markt|graben|steig|steg|berg|promenade|zeile|hof|siedlung|anger)`

	// German: suffix form (Gartenstraße 27, Margaretengürtel 5)
	deStreetSuffix := `(?:[A-ZÄÖÜ][a-zäöüß]+` + deSuffixes + `)[ \t]+` + houseNum

	// German: separate-word street name (Berliner Straße 15, Hoher Markt 3)
	deSepWords := `(?:Straße|Strasse|Str\.|Weg|Platz|Allee|Gasse|Ring|Damm|Ufer|Kai|Quai|Gürtel|Markt|Graben|Steig|Steg|Berg|Promenade|Zeile|Hof|Siedlung|Anger)`
	deStreetSep := namePattern + `(?:[ \t]+` + namePattern + `)?[ \t]+` + deSepWords + `[ \t]+` + houseNum

	// German: hyphenated street names ending in suffix (Theodor-Stern-Kai 7)
	deStreetHyphen := `(?:[A-ZÄÖÜ][a-zäöüß]+-)+(?:Straße|Strasse|Str|Weg|Platz|Allee|Gasse|Ring|Damm|Ufer|Kai|Quai|Gürtel|Markt|Graben|Steig|Berg|Promenade|Zeile|Hof)[ \t]+` + houseNum

	// City pattern: "Frankfurt", "Bad Homburg", "Frankfurt am Main"
	cityWord := `[A-ZÄÖÜ][a-zäöüß]+`
	cityPattern := cityWord + `(?:[ \t]+` + cityWord + `|[ \t]+[a-z]+[ \t]+` + cityWord + `)?`

	// German/Austrian/Swiss with postcode + city (\d{4,5} supports AT 4-digit and DE 5-digit)
	deWithCitySuffix := deStreetSuffix + `(?:,[ \t]*\d{4,5}[ \t]+` + cityPattern + `)?`
	deWithCitySep := deStreetSep + `(?:,[ \t]*\d{4,5}[ \t]+` + cityPattern + `)?`
	deWithCityHyphen := deStreetHyphen + `(?:,[ \t]*\d{4,5}[ \t]+` + cityPattern + `)?`

	// French: number + rue/avenue/boulevard (42, rue de la Loi)
	frStreet := `\d{1,4},?[ \t]+(?:rue|avenue|boulevard|place|chemin|impasse)[ \t]+(?:de[ \t]+(?:la[ \t]+)?|du[ \t]+|des[ \t]+|l')?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*`

	// French reversed: rue de la Loi 42 (street name first, number at end — Belgian/informal)
	frStreetReversed := `(?:[Rr]ue|[Aa]venue|[Bb]oulevard|[Pp]lace|[Cc]hemin|[Ii]mpasse)[ \t]+(?:de[ \t]+(?:la[ \t]+)?|du[ \t]+|des[ \t]+|l')?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Italian: via/piazza/corso + name + number (with articles: del, della, etc.)
	itStreet := `(?:[Vv]ia|[Pp]iazza|[Cc]orso|[Vv]iale)[ \t]+(?:(?:del|della|dello|dei|degli|delle|di)[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Spanish: calle/avenida/plaza/paseo
	esStreet := `(?:[Cc]alle|[Aa]venida|[Pp]laza|[Pp]aseo)[ \t]+(?:de[ \t]+(?:la[ \t]+)?|del[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Dutch: straat/laan/weg/plein/gracht/dreef + number
	nlStreet := `[A-ZÄÖÜ][a-zäöüß]+(?:straat|laan|weg|plein|gracht|kade|singel|dreef)[ \t]+\d{1,4}`

	// Swedish: vägen/gatan/stigen + number
	seStreet := `[A-ZÅÄÖ][a-zåäö]+(?:vägen|väg|gatan|gata|stigen|stig)[ \t]+\d{1,4}`

	// Danish: vej/gade/allé/stræde + number
	dkStreet := `[A-ZÆØÅ][a-zæøå]+(?:vej|gade|allé|stræde)[ \t]+\d{1,4}`

	// Norwegian: veien/gata/gate + number
	noStreet := `[A-ZÆØÅ][a-zæøå]+(?:veien|vei|gata|gate)[ \t]+\d{1,4}`

	// Finnish: katu/tie/polku + number
	fiStreet := `[A-ZÄÖÅ][a-zäöå]+(?:katu|tie|polku|puistikko)[ \t]+\d{1,4}`

	// Polish: ul./ulica + name + number
	plStreet := `(?:ul\.|ulica|al\.|aleja)[ \t]+[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+(?:[ \t]+[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+)*[ \t]+\d{1,4}`

	// Czech: ulice/třída/náměstí + name + number
	czStreet := `(?:ulice|třída|tř\.|náměstí|nám\.)[ \t]+[A-ZÁČĎÉĚÍŇÓŘŠŤÚŮÝŽ][a-záčďéěíňóřšťúůýž]+(?:[ \t]+[A-ZÁČĎÉĚÍŇÓŘŠŤÚŮÝŽ][a-záčďéěíňóřšťúůýž]+)*[ \t]+\d{1,4}`

	// Czech reversed: "Václavské náměstí 25" (adjective + náměstí/třída + number)
	czStreetReversed := `[A-ZÁČĎÉĚÍŇÓŘŠŤÚŮÝŽ][a-záčďéěíňóřšťúůýž]+(?:[ \t]+[A-Za-záčďéěíňóřšťúůýž]+)*[ \t]+(?:náměstí|nám\.|třída|tř\.)[ \t]+\d{1,4}`

	// Hungarian: utca/út/tér/körút + name + number
	huStreet := `[A-ZÁÉÍÓÖŐÚÜŰ][a-záéíóöőúüű]+(?:[ \t]+[A-Za-záéíóöőúüű]+)*[ \t]+(?:utca|út|tér|körút)[ \t]+\d{1,4}`

	// Romanian: strada/bulevardul/bd. + name + number
	// Use full words to avoid collision with German "Str." abbreviation for Straße.
	roStreet := `(?:strada|bd\.|bulevardul)[ \t]+[A-ZĂÂÎȘȚ][a-zăâîșț]+(?:[ \t]+[A-ZĂÂÎȘȚ][a-zăâîșț]+)*[ \t]+(?:nr\.\s?)?\d{1,4}`
	// Romanian "str." variant — require "nr." nearby to disambiguate from German Str.
	roStreetAbbr := `str\.[ \t]+[A-ZĂÂÎȘȚ][a-zăâîșț]+(?:[ \t]+[A-ZĂÂÎȘȚ][a-zăâîșț]+)*[ \t]+nr\.\s?\d{1,4}`

	// Croatian: ulica/trg + name + number
	hrStreet := `(?:ulica|trg|ul\.)[ \t]+[A-ZČĆĐŠŽ][a-zčćđšž]+(?:[ \t]+[A-ZČĆĐŠŽ][a-zčćđšž]+)*[ \t]+\d{1,4}`

	// Portuguese: rua/avenida/praça/travessa + name + number
	ptStreet := `(?:[Rr]ua|[Aa]venida|[Pp]raça|[Tt]ravessa)[ \t]+(?:(?:de|da|do|dos|das)[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Greek: transliterated (odos/leoforos/plateia + name + number)
	grStreet := `(?:odos|leoforos|plateia|Odos|Leoforos|Plateia)[ \t]+[A-Z][a-z]+(?:[ \t]+[A-Z][a-z]+)*[ \t]+\d{1,4}`

	// Belgian bilingual: rue/straat patterns (already covered by FR + NL, but add explicit BE patterns)
	beStreet := `(?:rue|straat|laan|avenue)[ \t]+(?:(?:de|du|des|van|het)[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// UK postcode: SW1A 2AA format
	ukPostcode := `\b[A-Z]{1,2}[0-9][0-9A-Z]?\s?[0-9][A-Z]{2}\b`

	// Canadian postcode: A1A 1A1
	caPostcode := `\b[A-Z][0-9][A-Z]\s?[0-9][A-Z][0-9]\b`

	// Dutch postcode: 1234 AB
	nlPostcode := `\b[0-9]{4}\s?[A-Z]{2}\b`

	// Polish postcode + city: XX-XXX Warszawa (require city to reduce false positives)
	plPostcodeCity := `\b\d{2}-\d{3}[ \t]+[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+`

	// --- US/English address patterns ---

	// US street type suffixes
	usStreetType := `(?:Ave(?:nue)?|Blvd|Boulevard|Cir(?:cle)?|Ct|Court|Dr(?:ive)?|Expy|Expressway|Hwy|Highway|Ln|Lane|Pkwy|Parkway|Pl(?:ace)?|Rd|Road|St(?:reet)?|Ter(?:r(?:ace)?)?|Trl|Trail|Way)\.?`

	// Optional directional prefix/suffix (N, S, E, W, NE, NW, SE, SW)
	usDir := `(?:[NESW]\.?|NE|NW|SE|SW)`

	// US street: 440 N Barranca Ave #4133
	usStreet := `\d{1,5}[ \t]+(?:` + usDir + `[ \t]+)?[A-Z][a-z]+(?:[ \t]+[A-Z][a-z]+)*[ \t]+` + usStreetType + `(?:[ \t]+` + usDir + `)?(?:[ \t]+(?:#|Apt\.?|Suite|Ste\.?|Unit|Fl\.?)[ \t]*[A-Za-z0-9]+)?`

	// US state abbreviations
	usStateAbbr := `(?:AL|AK|AZ|AR|CA|CO|CT|DE|FL|GA|HI|ID|IL|IN|IA|KS|KY|LA|ME|MD|MA|MI|MN|MS|MO|MT|NE|NV|NH|NJ|NM|NY|NC|ND|OH|OK|OR|PA|RI|SC|SD|TN|TX|UT|VT|VA|WA|WV|WI|WY|DC)`

	// US state full names
	usStateNames := `(?:Alabama|Alaska|Arizona|Arkansas|California|Colorado|Connecticut|Delaware|Florida|Georgia|Hawaii|Idaho|Illinois|Indiana|Iowa|Kansas|Kentucky|Louisiana|Maine|Maryland|Massachusetts|Michigan|Minnesota|Mississippi|Missouri|Montana|Nebraska|Nevada|New[ \t]+Hampshire|New[ \t]+Jersey|New[ \t]+Mexico|New[ \t]+York|North[ \t]+Carolina|North[ \t]+Dakota|Ohio|Oklahoma|Oregon|Pennsylvania|Rhode[ \t]+Island|South[ \t]+Carolina|South[ \t]+Dakota|Tennessee|Texas|Utah|Vermont|Virginia|Washington|West[ \t]+Virginia|Wisconsin|Wyoming|District[ \t]+of[ \t]+Columbia)`

	// US city + state + ZIP: Covina, California 91723 or Covina, CA 91723-1234
	usCityStateZip := `[A-Z][a-z]+(?:[ \t]+[A-Z][a-z]+)*,[ \t]+(?:` + usStateAbbr + `|` + usStateNames + `)[ \t]+\d{5}(?:-\d{4})?`

	// Irish Eircode: D02 AX07, A65 F4E2, T12 AB34
	// Routing key: specific letter + digit + (digit|W), unique ID: 4 alphanumeric
	eircode := `\b[ACDEFHKNPRTVWXY]\d[0-9W][ \t]+[A-Z0-9]{4}\b`

	// Dublin postal district: "Dublin 2", "Dublin 24", "Dublin 6W"
	dublinDistrict := `Dublin[ \t]+(?:\d{1,2}|6W)\b`

	// English/Irish street name without house number (context-validated, line-anchored)
	// Catches "Fenian St", "Baker Street" near other address components
	enStreetNoNum := `(?m)^([A-Z][a-z]+(?:[ \t]+[A-Z][a-z]+){0,2}[ \t]+` + usStreetType + `)[ \t]*$`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(deWithCitySuffix), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(deWithCitySep), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(deWithCityHyphen), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(frStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(frStreetReversed), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(itStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(esStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(nlStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(seStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(dkStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(noStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(fiStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(plStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(czStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(czStreetReversed), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(huStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(roStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(roStreetAbbr), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(hrStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(ptStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(grStreet), "ADDRESS", 0.80),
		NewRegexScanner(regexp.MustCompile(beStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(usStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(usCityStateZip), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(ukPostcode), "ADDRESS", 0.85),
		NewRegexScanner(
			regexp.MustCompile(caPostcode),
			"ADDRESS", 0.80,
			WithContextValidator(postcodeNearCountry),
		),
		NewRegexScanner(
			regexp.MustCompile(nlPostcode),
			"ADDRESS", 0.75,
			WithContextValidator(postcodeNearCountry),
		),
		NewRegexScanner(
			regexp.MustCompile(plPostcodeCity),
			"ADDRESS", 0.80,
			WithContextValidator(postcodeNearCountry),
		),
		NewRegexScanner(regexp.MustCompile(eircode), "ADDRESS", 0.90),
		NewRegexScanner(regexp.MustCompile(dublinDistrict), "ADDRESS", 0.85),
		// Standalone European postcode + city: "1100 Wien", "10115 Berlin", "8001 Zürich"
		// AT/CH: 4 digits (1xxx-9xxx), DE: 5 digits
		NewRegexScanner(
			regexp.MustCompile(`\b\d{4,5}[ \t]+`+cityPattern),
			"ADDRESS", 0.80,
			WithContextValidator(postcodeNearCountry),
		),
		// Generic street: CapWord(s) + house number on its own line.
		// Uses (?m) so ^ and $ match line boundaries.
		// Only matches when a postcode, country, or known street suffix appears nearby.
		// Catches streets without standard suffixes (e.g. "Am Tabor 5", "Spittelau 3").
		NewRegexScanner(
			regexp.MustCompile(`(?m)^([A-ZÄÖÜ][A-Za-zäöüßÀ-ÿ]+(?:[ \t]+[A-Za-zäöüßÀ-ÿ]+){0,3}[ \t]+`+houseNum+`)[ \t]*$`),
			"ADDRESS", 0.75,
			WithExtractGroup(1),
			WithContextValidator(postcodeNearCountry),
		),
		// English/Irish street name without number, context-validated
		NewRegexScanner(
			regexp.MustCompile(enStreetNoNum),
			"ADDRESS", 0.75,
			WithExtractGroup(1),
			WithContextValidator(postcodeNearCountry),
		),
	}
}

// postcodeNearCountry boosts confidence by checking if a country name appears
// within ~200 bytes of the postcode match (common in structured addresses).
// If no country is found, the match is still valid but the base score applies.
func postcodeNearCountry(fullText string, start, end int) bool {
	// Look within 200 bytes around the match for country/address context.
	from := start - 200
	if from < 0 {
		from = 0
	}
	to := end + 200
	if to > len(fullText) {
		to = len(fullText)
	}
	window := strings.ToLower(fullText[from:to])

	// Country names that confirm this is an address
	countries := []string{
		"austria", "österreich", "germany", "deutschland",
		"switzerland", "schweiz", "suisse", "svizzera",
		"netherlands", "niederlande", "nederland",
		"belgium", "belgien", "belgique", "belgië",
		"france", "frankreich",
		"italy", "italien", "italia",
		"spain", "spanien", "españa",
		"portugal", "poland", "polen", "polska",
		"czech", "tschechien", "česko",
		"hungary", "ungarn", "magyarország",
		"romania", "rumänien", "românia",
		"croatia", "kroatien", "hrvatska",
		"bulgaria", "bulgarien", "българия",
		"greece", "griechenland", "ελλάδα",
		"sweden", "schweden", "sverige",
		"denmark", "dänemark", "danmark",
		"norway", "norwegen", "norge",
		"finland", "finnland", "suomi",
		"iceland", "island", "ísland",
		"ireland", "éire", "united kingdom",
		"estonia", "estland", "eesti",
		"latvia", "lettland", "latvija",
		"lithuania", "litauen", "lietuva",
		"slovenia", "slowenien", "slovenija",
		"slovakia", "slowakei", "slovensko",
		"luxembourg", "luxemburg",
		"malta", "cyprus", "zypern",
		"canada", "kanada",
		"australia", "australien",
		"dublin", "london", "edinburgh",
		"toronto", "vancouver", "montreal", "ottawa", "calgary",
		"warsaw", "warschau", "praha", "budapest",
		"bucharest", "bukarest", "zagreb", "sofia",
		"athens", "athen", "stockholm", "copenhagen",
		"oslo", "helsinki", "reykjavik",
		"lisbon", "lissabon", "lisboa",
	}
	for _, c := range countries {
		if strings.Contains(window, c) {
			return true
		}
	}

	// Also match if there's a street-like line nearby (address block context)
	streetIndicators := []string{
		"straße", "str.", "gasse", "weg ", "platz",
		"allee", "ring ", "damm", "gürtel",
		"ave ", "avenue", "street", "road", "blvd",
		"rue ", "via ", "calle", "avenida",
		"ulica", "ul.", "ulice",  // PL, CZ
		"utca", "út ",            // HU
		"strada", "str.",         // RO
		"vägen", "gatan",         // SE
		"vej ", "gade",           // DK
		"veien", "gata ", "gate", // NO
		"katu", "tie ",           // FI
		"rua ", "praça",          // PT
		"odos",                   // GR
	}
	for _, s := range streetIndicators {
		if strings.Contains(window, s) {
			return true
		}
	}

	return false
}

// financialContext checks if a bare numeric amount (e.g. "65,00") appears
// near financial keywords, confirming it's likely a price/amount.
func financialContext(fullText string, start, end int) bool {
	from := start - 300
	if from < 0 {
		from = 0
	}
	to := end + 300
	if to > len(fullText) {
		to = len(fullText)
	}
	window := strings.ToLower(fullText[from:to])

	keywords := []string{
		// German
		"preis", "e-preis", "g-preis", "betrag", "summe", "gesamt",
		"netto", "brutto", "mwst", "ust", "rechnung", "zahlung",
		"rabatt", "skonto", "gebühr", "kosten", "honorar", "entgelt",
		"leistung", "rechnungsbetrag", "gesamtbetrag", "endbetrag",
		// English
		"price", "amount", "total", "subtotal", "tax", "payment",
		"invoice", "receipt", "fee", "charge", "cost", "balance",
		// French
		"prix", "montant", "facture", "paiement", "solde",
		// Italian
		"prezzo", "importo", "fattura", "pagamento",
		// Spanish
		"precio", "importe", "factura", "pago",
		// Dutch
		"prijs", "bedrag", "factuur", "betaling",
		// Polish
		"cena", "kwota", "faktura", "płatność",
		// Symbols/codes
		"€", "eur", "$", "£", "chf",
		"zł", "pln", "kč", "czk", "ft", "huf",
		"lei", "ron", "kr", "sek", "nok", "dkk",
	}
	for _, k := range keywords {
		if strings.Contains(window, k) {
			return true
		}
	}
	return false
}

// --- SECRET ---

func secretScanners() []Scanner {
	patterns := []struct {
		pattern string
		score   float64
	}{
		// OpenAI
		{`sk-proj-[A-Za-z0-9_\-]{20,}`, 0.99},
		{`sk-[A-Za-z0-9]{20,}`, 0.99},
		// Anthropic
		{`sk-ant-[A-Za-z0-9_\-]{20,}`, 0.99},
		// AWS access key
		{`AKIA[0-9A-Z]{16}`, 0.99},
		// GitHub
		{`gh[patos]_[A-Za-z0-9]{30,}`, 0.99},
		// Slack
		{`xox[bp]-[0-9]{10,}-[A-Za-z0-9\-]+`, 0.99},
		// Bearer token
		{`Bearer\s+[A-Za-z0-9._~+/=\-]{20,}`, 0.95},
		// PEM private key (just the header line)
		{`-----BEGIN (?:RSA |EC |DSA )?PRIVATE KEY-----`, 0.99},

		// Google Cloud API Key
		{`AIza[0-9A-Za-z_\-]{35}`, 0.99},
		// Firebase server key
		{`AAAA[A-Za-z0-9_\-]{7}:[A-Za-z0-9_\-]{140}`, 0.99},

		// Stripe keys
		{`sk_live_[0-9a-zA-Z]{24,}`, 0.99},
		{`pk_live_[0-9a-zA-Z]{24,}`, 0.99},
		{`sk_test_[0-9a-zA-Z]{24,}`, 0.95},
		{`rk_live_[0-9a-zA-Z]{24,}`, 0.99},

		// Twilio Account SID
		{`AC[0-9a-f]{32}`, 0.95},
		// SendGrid
		{`SG\.[A-Za-z0-9_\-]{22,}\.[A-Za-z0-9_\-]{43,}`, 0.99},
		// Discord Bot Token
		{`[MN][A-Za-z\d]{23,}\.\w{6}\.[\w\-]{27,}`, 0.95},

		// GitLab Personal Access Token
		{`glpat-[0-9a-zA-Z_\-]{20,}`, 0.99},
		// npm token
		{`npm_[A-Za-z0-9]{36}`, 0.99},
		// PyPI token
		{`pypi-[A-Za-z0-9_]{50,}`, 0.99},

		// JWT Token
		{`eyJ[A-Za-z0-9_\-]*\.eyJ[A-Za-z0-9_\-]*\.[A-Za-z0-9_\-]+`, 0.95},

		// Connection string with credentials: known protocols only
		{`(?:mysql|postgres|postgresql|mongodb|redis|amqp|mqtt|ftp|sftp|ssh|ldap|smtp|nats)://[^\s:]+:[^\s@]+@[^\s]+`, 0.95},
	}

	scanners := make([]Scanner, 0, len(patterns))
	for _, p := range patterns {
		scanners = append(scanners, NewRegexScanner(
			regexp.MustCompile(p.pattern), "SECRET", p.score,
		))
	}
	return scanners
}

// --- TAX NUMBERS ---

// taxNumberScanners returns context-triggered scanners for EU and European tax/business numbers.
// All patterns require a keyword prefix to avoid false positives on bare digit sequences.
func taxNumberScanners() []Scanner {
	return []Scanner{
		// DE: Steuernummer (143/262/10560)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Steuernummer|Steuer-Nr\.?|St\.?-?Nr\.?)[:\s]+(\d{2,3}/\d{3}/\d{4,5})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// AT: Steuernummer (12-345/6789 or 123456789)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Steuernummer|Steuer-Nr\.?|Abgabenkontonr\.?)[:\s]+(\d{2}-?\d{3}/?-?\d{4})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// FR: Numéro fiscal (13 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:numéro\s+fiscal|num[ée]ro\s+fiscal|SPI|n°\s*fiscal)[:\s]+(\d{13})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// IT: Partita IVA (11 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Partita\s+IVA|P\.?\s*IVA)[:\s]+(\d{11})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// ES: NIF/CIF (letter + 7 digits + alphanumeric)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:NIF|CIF|N\.I\.F\.)[:\s]+([A-Z]\d{7}[A-Z0-9])\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// PL: NIP (XXX-XXX-XX-XX or 10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:NIP|N\.I\.P\.)[:\s]+(\d{3}-?\d{3}-?\d{2}-?\d{2})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// HU: Adószám (XXXXXXXX-X-XX)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:adószám|adóazonosító\s+jel)[:\s]+(\d{8}-?\d-?\d{2})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// BE: Ondernemingsnummer (XXXX.XXX.XXX or 10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:ondernemingsnummer|numéro\s+d'entreprise|KBO|BCE)[:\s]+(\d{4}\.?\d{3}\.?\d{3})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// SK: DIČ / IČ DPH (10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:DIČ|IČ\s+DPH)[:\s]+(\d{10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// SI: Davčna številka (8 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:davčna\s+številka|ID\s+za\s+DDV)[:\s]+(\d{8})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// SE: Organisationsnummer (XXXXXX-XXXX or 10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:organisationsnummer|org\.?\s*nr\.?)[:\s]+(\d{6}-?\d{4})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// DK: CVR / SE-nummer (8 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:CVR|SE-nummer)[:\s]+(\d{8})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// FI: Y-tunnus (XXXXXXX-X)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Y-tunnus|FO-nummer)[:\s]+(\d{7}-?\d)\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// NO: Organisasjonsnummer (9 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:organisasjonsnummer|org\.?\s*nr\.?)[:\s]+(\d{9})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// RO: CUI / CIF / Cod fiscal (2-10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:CUI|CIF|cod\s+fiscal)[:\s]+(\d{2,10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// BG: BULSTAT / ЕИК / ИН (9-13 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:BULSTAT|ЕИК|ИН)[:\s]+(\d{9,13})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// GR: ΑΦΜ / AFM / TIN (9 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:ΑΦΜ|AFM)[:\s]+(\d{9})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// LU: Matricule national (11-13 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:matricule\s+national|numéro\s+d'identification)[:\s]+(\d{11,13})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// CY: TIC / tax identification (8 digits + letter)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:TIC|tax\s+identification)[:\s]+(\d{8}[A-Z])\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// MT: TIN (7-9 digits, keyword-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Malta\s+TIN|MT\s+TIN)[:\s]+(\d{7,9})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// EE: Registrikood / KMKR (8 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:registrikood|KMKR)[:\s]+(\d{8})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// LV: Reģistrācijas numurs / PVN (11 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:reģistrācijas\s+numurs|PVN)[:\s]+(\d{11})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// LT: Įmonės kodas / PVM (7-12 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:įmonės\s+kodas|PVM)[:\s]+(\d{7,12})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// CH: UID (CHE-XXX.XXX.XXX)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:UID|Unternehmens-ID)[:\s]+(CHE-?\d{3}\.?\d{3}\.?\d{3})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// GB: UTR / Unique Taxpayer Reference (10 digits)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:UTR|Unique\s+Taxpayer\s+Reference|tax\s+reference)[:\s]+(\d{10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
	}
}

// --- BUSINESS IDs (Customer, Employee, Contract, Salary, SEPA) ---

// businessIDScanners returns context-triggered scanners for common business identifiers
// across all EU languages. All patterns require a keyword to avoid false positives.
func businessIDScanners() []Scanner {
	// --- Customer Number ---
	customerKW := `(?:` +
		// DE
		`Kundennummer|Kunden-Nr\.?|Kundenkonto|Kd\.?-?Nr\.?|Kundenkennung` +
		// EN
		`|Customer\s?Number|Customer\s?No\.?|Customer\s?ID|Client\s?ID|Client\s?No\.?|Client\s?Number|Account\s?Number|Account\s?No\.?` +
		// FR
		`|Numéro\s+(?:de\s+)?client|N°\s*client|Code\s+client` +
		// IT
		`|Numero\s+cliente|Cod(?:ice)?\s+cliente` +
		// ES/PT/GL
		`|Número\s+(?:de\s+)?cliente|Cód(?:igo)?\s+(?:de\s+)?cliente` +
		// NL
		`|Klantnummer|Klant-Nr\.?|Relatienummer` +
		// PL
		`|Numer\s+klienta|Nr\s+klienta` +
		// CZ
		`|Číslo\s+zákazníka|Číslo\s+klienta` +
		// SK
		`|Zákaznícke\s+číslo` +
		// HU
		`|Ügyfélszám|Ügyfél-?azonosító` +
		// RO
		`|Număr\s+client|Cod\s+client` +
		// BG
		`|Клиентски\s+номер` +
		// HR
		`|Broj\s+kupca|Šifra\s+kupca` +
		// SI
		`|Številka\s+stranke|Šifra\s+stranke` +
		// EL
		`|Αριθμός\s+πελάτη|Κωδικός\s+πελάτη` +
		// SE/NO/DK
		`|Kundnummer|Kund-Nr\.?|Kundenummer|Kunde-Nr\.?` +
		// FI
		`|Asiakasnumero|Asiakastunnus` +
		// EE
		`|Kliendikood` +
		// LV
		`|Klienta\s+numurs|Klienta\s+Nr\.?` +
		// LT
		`|Kliento\s+numeris|Kliento\s+Nr\.?` +
		// GA
		`|Uimhir\s+chustaiméara` +
		// MT
		`|Numru\s+tal-klijent` +
		`)`

	// --- Employee Number ---
	employeeKW := `(?:` +
		// DE
		`Mitarbeiternummer|Mitarbeiter-Nr\.?|Personalnummer|Personal-Nr\.?|Pers\.?-?Nr\.?|Dienstnummer` +
		// EN
		`|Employee\s?Number|Employee\s?No\.?|Employee\s?ID|Staff\s?ID|Staff\s?No\.?|Staff\s?Number|Personnel\s?Number|Personnel\s?No\.?|Personnel\s?ID` +
		// FR
		`|Numéro\s+(?:de\s+)?salarié|Numéro\s+d['']employé|Matricule|N°\s*(?:de\s+)?salarié|N°\s*matricule` +
		// IT
		`|Numero\s+(?:di\s+)?dipendente|Matricola` +
		// ES
		`|Número\s+(?:de\s+)?empleado|Número\s+(?:de\s+)?personal` +
		// PT
		`|Número\s+(?:de\s+)?funcionário` +
		// NL
		`|Personeelsnummer|Werknemer-?Nr\.?|Medewerkersnummer` +
		// PL
		`|Numer\s+pracownika|Nr\s+pracownika` +
		// CZ
		`|Číslo\s+zaměstnance|Osobní\s+číslo` +
		// SK
		`|Číslo\s+zamestnanca` +
		// HU
		`|Alkalmazotti\s+szám|Dolgozói\s+szám` +
		// RO
		`|Număr\s+(?:de\s+)?angajat|Număr\s+(?:de\s+)?personal` +
		// BG
		`|Служебен\s+номер` +
		// HR
		`|Matični\s+broj|Broj\s+zaposlenika` +
		// SI
		`|Matična\s+številka|Številka\s+zaposlenega` +
		// EL
		`|Αριθμός\s+(?:μητρώου\s+)?υπαλλήλου|Αριθμός\s+μητρώου` +
		// SE
		`|Anställningsnummer|Anställnings-?Nr\.?` +
		// DA
		`|Medarbejdernummer|Medarbejder-Nr\.?` +
		// NO
		`|Ansattnummer|Ansatt-?Nr\.?` +
		// FI
		`|Henkilönumero` +
		// EE
		`|Töötaja\s+kood` +
		// LV
		`|Darbinieka\s+numurs|Darbinieka\s+Nr\.?` +
		// LT
		`|Darbuotojo\s+numeris|Darbuotojo\s+Nr\.?` +
		// GA
		`|Uimhir\s+fhostaí` +
		// MT
		`|Numru\s+tal-impjegat` +
		// LB
		`|Matrikelnummer` +
		`)`

	// --- Contract Number ---
	contractKW := `(?:` +
		// DE
		`Vertragsnummer|Vertrags-?Nr\.?|Vertragsnr\.?` +
		// EN
		`|Contract\s?Number|Contract\s?No\.?|Contract\s?ID|Agreement\s?Number|Agreement\s?No\.?` +
		// FR
		`|Numéro\s+(?:de\s+)?contrat|N°\s*(?:de\s+)?contrat|Réf(?:érence)?\s+contrat` +
		// IT
		`|Numero\s+(?:di\s+)?contratto|Rif\.?\s+contratto` +
		// ES/PT/GL
		`|Número\s+(?:de\s+)?contrato|Ref\.?\s+contrato` +
		// NL
		`|Contractnummer|Contract-?Nr\.?|Overeenkomstnummer` +
		// PL
		`|Numer\s+umowy|Nr\s+umowy` +
		// CZ
		`|Číslo\s+smlouvy` +
		// SK
		`|Číslo\s+zmluvy` +
		// HU
		`|Szerződésszám` +
		// RO
		`|Număr\s+(?:de\s+)?contract` +
		// BG
		`|Номер\s+(?:на\s+)?договор` +
		// HR
		`|Broj\s+ugovora` +
		// SI
		`|Številka\s+pogodbe` +
		// EL
		`|Αριθμός\s+σύμβασης` +
		// SE
		`|Avtalsnummer|Avtals-?Nr\.?` +
		// DA
		`|Aftalenummer|Aftale-?Nr\.?` +
		// NO
		`|Kontraktnummer|Kontrakt-?Nr\.?` +
		// FI
		`|Sopimusnumero|Sopimus-?Nr\.?` +
		// EE
		`|Lepingu\s+number|Lepingu\s+Nr\.?` +
		// LV
		`|Līguma\s+numurs|Līguma\s+Nr\.?` +
		// LT
		`|Sutarties\s+numeris|Sutarties\s+Nr\.?` +
		// GA
		`|Uimhir\s+chonartha` +
		// MT
		`|Numru\s+tal-kuntratt` +
		`)`

	// --- Salary keywords (before an amount) ---
	salaryKW := `(?:` +
		// DE
		`Bruttolohn|Bruttogehalt|Bruttoverdienst|Nettolohn|Nettogehalt|Nettoverdienst|Grundgehalt|Grundlohn|Stundenlohn|Stundensatz|Monatslohn|Monatsgehalt|Jahresgehalt|Jahresbrutto|Gesamtvergütung` +
		// EN
		`|Gross\s?salary|Gross\s?pay|Gross\s?wage|Net\s?salary|Net\s?pay|Net\s?wage|Base\s?salary|Base\s?pay|Hourly\s?rate|Hourly\s?wage|Annual\s?salary|Monthly\s?salary|Monthly\s?wage|Monthly\s?pay|Total\s?compensation` +
		// FR
		`|Salaire\s+brut|Salaire\s+net|Salaire\s+(?:de\s+)?base|Salaire\s+mensuel|Salaire\s+annuel|Taux\s+horaire|Rémunération\s+brute|Rémunération\s+nette` +
		// IT
		`|Stipendio\s+lordo|Stipendio\s+netto|Stipendio\s+base|Retribuzione\s+lorda|Retribuzione\s+netta|Paga\s+oraria|Paga\s+mensile` +
		// ES
		`|Salario\s+bruto|Salario\s+neto|Sueldo\s+bruto|Sueldo\s+neto|Sueldo\s+base|Salario\s+base|Salario\s+mensual|Salario\s+anual|Tarifa\s+horaria` +
		// PT
		`|Salário\s+bruto|Salário\s+líquido|Salário\s+base|Salário\s+mensal|Salário\s+anual` +
		// NL
		`|Brutoloon|Nettoloon|Bruto\s?salaris|Netto\s?salaris|Basissalaris|Basisloon|Uurloon|Maandloon|Maandsalaris|Jaarsalaris` +
		// PL
		`|Wynagrodzenie\s+brutto|Wynagrodzenie\s+netto|Płaca\s+zasadnicza|Stawka\s+godzinowa|Płaca\s+miesięczna` +
		// CZ/SK
		`|Hrubá\s+mzda|Čistá\s+mzda|Základní\s+mzda|Hodinová\s+sazba|Měsíční\s+mzda|Základná\s+mzda|Mesačná\s+mzda` +
		// HU
		`|Bruttó\s+bér|Nettó\s+bér|Alapbér|Órabér|Havi\s+bér|Éves\s+fizetés` +
		// RO
		`|Salariu\s+brut|Salariu\s+net|Salariu\s+(?:de\s+)?bază|Salariu\s+lunar|Tarif\s+orar` +
		// BG
		`|Брутна\s+заплата|Нетна\s+заплата|Основна\s+заплата|Месечна\s+заплата` +
		// HR
		`|Bruto\s+plaća|Neto\s+plaća|Osnovna\s+plaća|Satnica|Mjesečna\s+plaća` +
		// SI
		`|Bruto\s+plača|Neto\s+plača|Osnovna\s+plača|Urna\s+postavka|Mesečna\s+plača` +
		// EL
		`|Μικτός\s+μισθός|Καθαρός\s+μισθός|Βασικός\s+μισθός|Ωριαία\s+αμοιβή|Μηνιαίος\s+μισθός` +
		// SE
		`|Bruttolön|Nettolön|Grundlön|Timlön|Månadslön|Årslön` +
		// DA
		`|Bruttoløn|Nettoløn|Grundløn|Timeløn|Månedsløn|Årsløn` +
		// NO
		`|Bruttolønn|Nettolønn|Grunnlønn|Timelønn|Månedslønn|Årslønn` +
		// FI
		`|Bruttopalkka|Nettopalkka|Peruspalkka|Tuntipalkka|Kuukausipalkka|Vuosipalkka` +
		// EE
		`|Brutopalk|Netopalk|Põhipalk|Tunnitasu|Kuupalk` +
		// LV
		`|Bruto\s+alga|Neto\s+alga|Pamatalga|Stundas\s+likme|Mēneša\s+alga` +
		// LT
		`|Bruto\s+atlyginimas|Neto\s+atlyginimas|Bazinis\s+atlyginimas|Valandinis\s+atlygis|Mėnesinis\s+atlyginimas` +
		`)`

	// Amount pattern for salary: handles €1.500, 1.500,00 €, $2,500.00, bare 4500,00, 45000 SEK etc.
	salaryAmount := `[€$£]?\s?\d{1,3}(?:[.,]\d{3})*(?:[.,]\d{2})?\s?(?:€|EUR|USD|GBP|CHF|PLN|CZK|HUF|RON|SEK|NOK|DKK|kr\.?|zł|Kč|Ft|lei)?`
	// Bare salary amount: handles numbers without thousand separators like "45000 SEK", "8500,00 PLN"
	salaryBareAmount := `\d{3,7}(?:[.,]\d{2})?\s?(?:€|EUR|USD|GBP|CHF|PLN|CZK|HUF|RON|SEK|NOK|DKK|kr\.?|zł|Kč|Ft|lei)`

	return []Scanner{
		// Customer number: keyword + alphanumeric ID
		NewRegexScanner(
			regexp.MustCompile(`(?i)`+customerKW+`[:\s]+([A-Za-z0-9][\w.\-/]{2,20})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Employee number: keyword + alphanumeric ID
		NewRegexScanner(
			regexp.MustCompile(`(?i)`+employeeKW+`[:\s]+([A-Za-z0-9][\w.\-/]{2,20})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Contract number: keyword + alphanumeric ID
		NewRegexScanner(
			regexp.MustCompile(`(?i)`+contractKW+`[:\s]+([A-Za-z0-9][\w.\-/]{2,20})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// Salary: keyword + amount with currency symbol (tagged as FINANCIAL)
		NewRegexScanner(
			regexp.MustCompile(`(?i)`+salaryKW+`[:\s]+(` + salaryAmount + `)`),
			"FINANCIAL", 0.90,
			WithExtractGroup(1),
		),
		// Salary: keyword + bare amount with currency suffix (45000 SEK, 8500,00 PLN)
		NewRegexScanner(
			regexp.MustCompile(`(?i)`+salaryKW+`[:\s]+(` + salaryBareAmount + `)`),
			"FINANCIAL", 0.90,
			WithExtractGroup(1),
		),
		// SEPA Creditor Reference (RF + check digits + reference, standalone)
		NewRegexScanner(
			regexp.MustCompile(`\bRF\d{2}[A-Z0-9]{1,21}\b`),
			"FINANCIAL", 0.95,
		),
		// SEPA Creditor ID: keyword + CC + 2 check digits + alphanumeric creditor reference
		// e.g. DE98ZZZ09999999999, FR98ZZZ123456, GB98ZZZSDDBARC0000000000001
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
				// DE
				`Gläubiger-?(?:ID|Identifikation(?:snummer)?)`+
				// EN
				`|Creditor[ \t]+(?:ID|Identifier)|SEPA[ \t]+Creditor[ \t]+ID`+
				// FR
				`|Identifiant[ \t]+créancier|ID[ \t]+créancier|ICS`+
				// IT
				`|ID[ \t]+creditore|Identificativo[ \t]+creditore|Codice[ \t]+creditore`+
				// ES
				`|Identificador[ \t]+del[ \t]+acreedor|ID[ \t]+acreedor`+
				// NL
				`|Incassant-?ID|Crediteur-?ID`+
				// PT
				`|ID[ \t]+credor|Identificação[ \t]+do[ \t]+credor`+
				// PL
				`|Identyfikator[ \t]+wierzyciela`+
				// SE/DK/NO
				`|Borgenärs-?ID|Kreditor-?ID|Kreditors-?ID`+
				// HU
				`|Hitelező[ \t]+azonosító`+
				// CZ/SK
				`|Identifikátor[ \t]+věřitele|Identifikátor[ \t]+veriteľa`+
				// RO
				`|ID[ \t]+creditor`+
				// EL
				`|Αναγνωριστικό[ \t]+πιστωτή`+
				`)[: \t]+([A-Z]{2}\d{2}[A-Za-z0-9]{3,30})\b`),
			"FINANCIAL", 0.95,
			WithExtractGroup(1),
		),
		// SEPA Mandate Reference: keyword + alphanumeric reference (up to 35 chars)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:`+
				// DE
				`Mandatsreferenz(?:nummer)?|Mandat-?Ref\.?`+
				// EN
				`|Mandate[ \t]+(?:reference|ref\.?)|SEPA[ \t]+mandate`+
				// FR
				`|Référence[ \t]+(?:de[ \t]+)?mandat|Réf\.?[ \t]+mandat|RUM`+
				// IT
				`|Riferimento[ \t]+mandato|Rif\.?[ \t]+mandato`+
				// ES
				`|Referencia[ \t]+(?:de[ \t]+)?mandato|Ref\.?[ \t]+mandato`+
				// NL
				`|Mandaatsreferentie|Mandaat-?ref\.?|Machtigingsreferentie`+
				// PT
				`|Referência[ \t]+(?:de[ \t]+)?mandato`+
				// PL
				`|Referencja[ \t]+mandatu`+
				// SE/DK/NO
				`|Mandatreferens|Mandatreference|Mandatreferanse`+
				// HU
				`|Mandátum[ \t]+hivatkozás|Felhatalmazás[ \t]+hivatkozás`+
				// CZ/SK
				`|Reference[ \t]+mandátu|Referencia[ \t]+mandátu`+
				// RO
				`|Referință[ \t]+mandat`+
				// EL
				`|Αναφορά[ \t]+εντολής`+
				`)[: \t]+([A-Za-z0-9][\w.\-/]{2,35})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
	}
}
