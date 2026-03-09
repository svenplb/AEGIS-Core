package scanner

import "strings"

// contextKeywords maps entity types to multilingual context keywords.
// When a keyword appears within the configured window around an entity,
// the entity's score is boosted by the boost factor.
var contextKeywords = map[string][]string{
	"PERSON": {
		// German
		"name", "vorname", "nachname", "herr", "frau", "person",
		"mitarbeiter", "kunde", "patient", "ansprechpartner",
		// English
		"name", "first name", "last name", "mr", "mrs", "ms",
		"employee", "customer", "patient", "contact",
		// French
		"nom", "prénom", "monsieur", "madame", "employé", "client",
		// Spanish
		"nombre", "apellido", "señor", "señora", "empleado", "cliente",
		// Italian
		"nome", "cognome", "signor", "signora", "impiegato",
		// Dutch
		"naam", "voornaam", "achternaam", "meneer", "mevrouw",
		// Portuguese
		"nome", "sobrenome", "senhor", "senhora",
		// Polish
		"imię", "nazwisko", "pan", "pani",
		// Swedish
		"namn", "förnamn", "efternamn",
	},
	"ADDRESS": {
		// German
		"adresse", "anschrift", "straße", "strasse", "plz", "ort", "wohnort",
		// English
		"address", "street", "city", "zip", "postal", "residence",
		// French
		"adresse", "rue", "ville", "code postal",
		// Spanish
		"dirección", "calle", "ciudad", "código postal",
		// Italian
		"indirizzo", "via", "città", "cap",
		// Dutch
		"adres", "straat", "stad", "postcode",
	},
	"ORG": {
		// German
		"firma", "unternehmen", "organisation", "arbeitgeber", "gesellschaft",
		// English
		"company", "organization", "organisation", "employer", "corporation",
		// French
		"entreprise", "société", "organisation",
		// Spanish
		"empresa", "organización", "compañía",
		// Italian
		"azienda", "società", "organizzazione",
	},
}

// EnhanceScores boosts entity scores when contextual keywords appear
// within windowSize characters around the entity in the source text.
func EnhanceScores(entities []Entity, text string, boostFactor float64, windowSize int) []Entity {
	if boostFactor <= 0 || windowSize <= 0 || len(entities) == 0 {
		return entities
	}

	lowerText := strings.ToLower(text)

	for i := range entities {
		keywords, ok := contextKeywords[entities[i].Type]
		if !ok {
			continue
		}

		// windowSize is in bytes (entity offsets are byte-based).
		windowStart := entities[i].Start - windowSize
		if windowStart < 0 {
			windowStart = 0
		}
		windowEnd := entities[i].End + windowSize
		if windowEnd > len(lowerText) {
			windowEnd = len(lowerText)
		}

		// Align to rune boundaries to avoid slicing mid-character.
		windowStart = alignToRuneStart(lowerText, windowStart)
		windowEnd = alignToRuneEnd(lowerText, windowEnd)

		window := lowerText[windowStart:windowEnd]

		for _, kw := range keywords {
			if strings.Contains(window, kw) {
				entities[i].Score += boostFactor
				if entities[i].Score > 1.0 {
					entities[i].Score = 1.0
				}
				break // one boost per entity
			}
		}
	}

	return entities
}

// alignToRuneStart finds the start of the rune at or before byteIdx.
func alignToRuneStart(s string, byteIdx int) int {
	if byteIdx <= 0 {
		return 0
	}
	if byteIdx >= len(s) {
		return len(s)
	}
	for byteIdx > 0 && !isRuneStart(s[byteIdx]) {
		byteIdx--
	}
	return byteIdx
}

// alignToRuneEnd finds the end of the rune at or after byteIdx.
func alignToRuneEnd(s string, byteIdx int) int {
	if byteIdx >= len(s) {
		return len(s)
	}
	for byteIdx < len(s) && !isRuneStart(s[byteIdx]) {
		byteIdx++
	}
	return byteIdx
}

func isRuneStart(b byte) bool {
	// In UTF-8, continuation bytes have the form 10xxxxxx.
	return b&0xC0 != 0x80
}
