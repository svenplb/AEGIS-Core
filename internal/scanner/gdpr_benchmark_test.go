package scanner

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// gdprCase is a single benchmark test case.
type gdprCase struct {
	Category string
	Language string
	Text     string
	// Expected entities: map of type → list of substrings that must be detected.
	Expected map[string][]string
}

// gdprBenchmarkCases returns all DSGVO/GDPR benchmark test cases.
// Each case contains real-world text with known PII that must be detected.
func gdprBenchmarkCases() []gdprCase {
	return []gdprCase{
		// ============================================================
		// CATEGORY: HR / Employee Data (Art. 6, 9, 88 DSGVO)
		// ============================================================
		{
			Category: "HR",
			Language: "DE",
			Text: `Personalakte: Herr Thomas Schmidt, geboren am 15.03.1985, wohnhaft in Hauptstraße 42, 10115 Berlin.
Steuer-ID: DE123456789. Kontakt: thomas.schmidt@example.com, +49 170 1234567.
Gehalt: €4.500,00 brutto. IBAN: DE89 3704 0044 0532 0130 00. Arbeitgeber: Siemens AG.
Krankheitstage 2025: 12 (Diagnose: J06.9 akute Infektion der oberen Atemwege).
Notfallkontakt: Maria Schmidt, +49 171 9876543.`,
			Expected: map[string][]string{
				"PERSON":   {"Thomas Schmidt", "Maria Schmidt"},
				"DATE":     {"15.03.1985"},
				"EMAIL":    {"thomas.schmidt@example.com"},
				"PHONE":    {"+49 170 1234567", "+49 171 9876543"},
				"IBAN":     {"DE89 3704 0044 0532 0130 00"},
				"ORG":      {"Siemens AG"},
				"MEDICAL":  {"J06.9"},
			},
		},
		{
			Category: "HR",
			Language: "EN",
			Text: `Employee Record — Sarah Johnson (DOB: 1990-07-22)
Address: 15 Baker Street, London W1U 3BW, United Kingdom
Email: sarah.johnson@company.co.uk | Phone: +44 20 7946 0958
National Insurance: AB 12 34 56 C
Annual Salary: £65,000.00 | IBAN: GB29 NWBK 6016 1331 9268 19
Emergency Contact: David Johnson, +44 7700 900123
Medical Note: Employee diagnosed with F32.1 moderate depressive episode.`,
			Expected: map[string][]string{
				"DATE":    {"1990-07-22"},
				"EMAIL":   {"sarah.johnson@company.co.uk"},
				"PHONE":   {"+44 20 7946 0958", "+44 7700 900123"},
				"IBAN":    {"GB29 NWBK 6016 1331 9268 19"},
				"MEDICAL": {"F32.1"},
			},
		},
		{
			Category: "HR",
			Language: "FR",
			Text: `Fiche employé: Monsieur Jean-Pierre Dubois
Date de naissance: 28/04/1978
Adresse: 42 rue de Rivoli, 75001 Paris, France
Téléphone: +33 6 12 34 56 78 | Email: jp.dubois@entreprise.fr
Numéro de sécurité sociale: 1 78 04 75 042 023 45
IBAN: FR76 3000 6000 0112 3456 7890 189
Employeur: BNP Paribas SA`,
			Expected: map[string][]string{
				"DATE":  {"28/04/1978"},
				"PHONE": {"+33 6 12 34 56 78"},
				"EMAIL": {"jp.dubois@entreprise.fr"},
				"IBAN":  {"FR76 3000 6000 0112 3456 7890 189"},
				"ORG":   {"BNP Paribas SA"},
			},
		},

		// ============================================================
		// CATEGORY: Medical Records (Art. 9 DSGVO — special categories)
		// ============================================================
		{
			Category: "Medical",
			Language: "DE",
			Text: `Arztbrief — Universitätsklinikum München
Patient: Frau Anna Weber, geb. 03.11.1972
Versichertennummer: A123456789
Diagnosen: E11.65 (Diabetes mellitus Typ 2), I10 (Hypertonie), M54.5 (Kreuzschmerzen)
Blutdruck: 145/92 mmHg, HbA1c: 7.8%, BMI: 31.4
Medikation: Metformin 1000mg, Ramipril 5mg
Nächster Termin: 15.04.2026
Kontakt Hausarzt: Dr. Heinrich Müller, praxis@dr-mueller.de, +49 89 12345678
Abrechnung: IBAN DE44 5001 0517 0648 4898 90`,
			Expected: map[string][]string{
				"DATE":    {"03.11.1972", "15.04.2026"},
				"MEDICAL": {"E11.65", "I10", "M54.5", "145/92 mmHg"},
				"EMAIL":   {"praxis@dr-mueller.de"},
				"PHONE":   {"+49 89 12345678"},
				"IBAN":    {"DE44 5001 0517 0648 4898 90"},
			},
		},
		{
			Category: "Medical",
			Language: "IT",
			Text: `Lettera medica — Ospedale San Raffaele, Milano
Paziente: Signor Marco Rossi, nato il 22.06.1965
Codice fiscale: RSSMRC65H22H501Z
Diagnosi: C34.1 (carcinoma polmonare), J44.1 (BPCO con riacutizzazione)
Pressione arteriosa: 130/85 mmHg
Prossimo appuntamento: 01.03.2026
Contatto: marco.rossi@email.it, +39 02 1234567
Pagamento: IBAN IT60 X054 2811 1010 0000 0123 456`,
			Expected: map[string][]string{
				"DATE":    {"22.06.1965", "01.03.2026"},
				"MEDICAL": {"C34.1", "J44.1", "130/85 mmHg"},
				"EMAIL":   {"marco.rossi@email.it"},
				"PHONE":   {"+39 02 1234567"},
				"IBAN":    {"IT60 X054 2811 1010 0000 0123 456"},
			},
		},
		{
			Category: "Medical",
			Language: "ES",
			Text: `Informe médico — Hospital Universitario La Paz, Madrid
Paciente: Señora Carmen López García, nacida el 14.09.1988
DNI: 12345678Z
Diagnóstico: K21.0 (enfermedad por reflujo gastroesofágico), F41.1 (trastorno de ansiedad generalizada)
Tensión arterial: 120/80 mmHg
Contacto: carmen.lopez@correo.es, +34 612 345 678
IBAN: ES91 2100 0418 4502 0005 1332`,
			Expected: map[string][]string{
				"DATE":    {"14.09.1988"},
				"MEDICAL": {"K21.0", "F41.1", "120/80 mmHg"},
				"EMAIL":   {"carmen.lopez@correo.es"},
				"PHONE":   {"+34 612 345 678"},
				"IBAN":    {"ES91 2100 0418 4502 0005 1332"},
			},
		},

		// ============================================================
		// CATEGORY: Financial / Banking (Art. 6 DSGVO)
		// ============================================================
		{
			Category: "Financial",
			Language: "DE",
			Text: `Kontoauszug — Deutsche Bank AG
Kontoinhaber: Maximilian Fischer
IBAN: DE89 3704 0044 0532 0130 00 | BIC: COBADEFFXXX
Kreditkarte: 4111 1111 1111 1111 (Visa, gültig bis 12/2027)
Umsätze Januar 2026:
  05.01.2026  Amazon EU S.a.r.l.        -€129,99
  12.01.2026  Gehalt Siemens AG       +€4.500,00
  15.01.2026  Miete Hauptstr. 42       -€1.200,00
  20.01.2026  SEPA-Lastschrift AOK     -€285,00
Kontostand: €8.456,78
Kundenservice: +49 69 910-10000, service@deutsche-bank.de`,
			Expected: map[string][]string{
				"ORG":         {"Deutsche Bank AG", "Siemens AG"},
				"IBAN":        {"DE89 3704 0044 0532 0130 00"},
				"CREDIT_CARD": {"4111 1111 1111 1111"},
				"EMAIL":       {"service@deutsche-bank.de"},
				"PHONE":       {"+49 69 910-10000"},
				"DATE":        {"05.01.2026", "12.01.2026", "15.01.2026", "20.01.2026"},
			},
		},
		{
			Category: "Financial",
			Language: "NL",
			Text: `Bankafschrift — ING Bank N.V.
Rekeninghouder: Pieter de Vries
IBAN: NL91 ABNA 0417 1643 00 | BIC: ABNANL2A
Creditcard: 5500 0000 0000 0004 (Mastercard)
Transacties:
  03.01.2026  Salaris Philips N.V.    +€3.800,00
  10.01.2026  Huur Prinsengracht 42   -€1.450,00
  15.01.2026  Zorgverzekering CZ      -€135,00
Saldo: €5.234,56
Contact: pieter.devries@email.nl, +31 20 123 4567`,
			Expected: map[string][]string{
				"ORG":         {"ING Bank N.V."},
				"IBAN":        {"NL91 ABNA 0417 1643 00"},
				"CREDIT_CARD": {"5500 0000 0000 0004"},
				"EMAIL":       {"pieter.devries@email.nl"},
				"PHONE":       {"+31 20 123 4567"},
				"DATE":        {"03.01.2026", "10.01.2026", "15.01.2026"},
			},
		},
		{
			Category: "Financial",
			Language: "AT",
			Text: `Bankbeleg — Erste Bank der oesterreichischen Sparkassen AG
Kontoinhaber: Mag. Leopold Gruber
IBAN: AT61 1904 3002 3457 3201 | BIC: GIBAATWWXXX
Kreditkarte: 3782 822463 10005 (Amex)
Steuer-Nr: ATU12345678
Letzte Buchung: 28.02.2026, Gehalt Voestalpine AG: +€5.200,00
Kontakt: leopold.gruber@email.at, +43 1 234 5678`,
			Expected: map[string][]string{
				"IBAN":        {"AT61 1904 3002 3457 3201"},
				"CREDIT_CARD": {"3782 822463 10005"},
				"EMAIL":       {"leopold.gruber@email.at"},
				"PHONE":       {"+43 1 234 5678"},
				"DATE":        {"28.02.2026"},
			},
		},

		// ============================================================
		// CATEGORY: Invoices / Contracts (Art. 6 DSGVO)
		// ============================================================
		{
			Category: "Invoice",
			Language: "DE",
			Text: `Rechnung Nr. RE-2026-0042
Müller & Partner GmbH, Friedrichstraße 100, 10117 Berlin
USt-IdNr: DE234567890 | Tel: +49 30 12345678

Rechnungsempfänger:
Herr Dr. Stefan Wagner
Mozartstraße 15, 80336 München
stefan.wagner@example.de

Leistungszeitraum: 01.01.2026 – 31.01.2026
Nettobetrag: €8.500,00 | MwSt 19%: €1.615,00 | Brutto: €10.115,00

Zahlung bitte auf: IBAN DE02 1203 0000 0000 2020 51
Fällig bis: 28.02.2026`,
			Expected: map[string][]string{
				"ORG":   {"Müller & Partner GmbH"},
				"PHONE": {"+49 30 12345678"},
				"EMAIL": {"stefan.wagner@example.de"},
				"DATE":  {"01.01.2026", "31.01.2026", "28.02.2026"},
				"IBAN":  {"DE02 1203 0000 0000 2020 51"},
			},
		},
		{
			Category: "Invoice",
			Language: "PL",
			Text: `Faktura VAT Nr FV/2026/01/0042
Wystawca: Kowalski Sp. z o.o., ul. Marszałkowska 42, 00-624 Warszawa
NIP: PL1234567890 | Tel: +48 22 123 45 67

Nabywca:
Pan Andrzej Nowak
ul. Krakowska 15, 31-066 Kraków
andrzej.nowak@poczta.pl

Data wystawienia: 15.01.2026
Kwota netto: 8 500,00 zł | VAT 23%: 1 955,00 zł | Brutto: 10 455,00 zł
Płatność na: IBAN PL61 1090 1014 0000 0712 1981 2874
Termin płatności: 15.02.2026`,
			Expected: map[string][]string{
				"PHONE": {"+48 22 123 45 67"},
				"EMAIL": {"andrzej.nowak@poczta.pl"},
				"DATE":  {"15.01.2026", "15.02.2026"},
				"IBAN":  {"PL61 1090 1014 0000 0712 1981 2874"},
			},
		},
		{
			Category: "Invoice",
			Language: "SE",
			Text: `Faktura Nr 2026-0088
Eriksson & Söner AB, Sveavägen 42, 111 34 Stockholm
Organisationsnummer: 556123-4567 | Tel: +46 8 123 45 67

Mottagare:
Herr Lars Andersson
Drottninggatan 15, 411 14 Göteborg
lars.andersson@email.se

Fakturadatum: 2026-01-15
Belopp exkl. moms: 85 000,00 kr | Moms 25%: 21 250,00 kr | Totalt: 106 250,00 kr
Betalning till: IBAN SE45 5000 0000 0583 9825 7466
Förfallodatum: 2026-02-15`,
			Expected: map[string][]string{
				"PHONE": {"+46 8 123 45 67"},
				"EMAIL": {"lars.andersson@email.se"},
				"DATE":  {"2026-01-15", "2026-02-15"},
				"IBAN":  {"SE45 5000 0000 0583 9825 7466"},
			},
		},

		// ============================================================
		// CATEGORY: Customer Data / CRM (Art. 6, 7 DSGVO)
		// ============================================================
		{
			Category: "CRM",
			Language: "DE",
			Text: `Kundendatenbank-Export (CRM)
---
Kunde #1042: Hans-Jürgen Becker, geb. 12.05.1968
  E-Mail: hj.becker@t-online.de
  Telefon: +49 221 9876543
  Adresse: Am Kölner Dom 5, 50667 Köln
  Kreditkarte: 4532 1234 5678 9012 (gültig: 09/2027)
  Kundenseit: 01.03.2018
---
Kunde #1043: Elisabeth Schwarz, geb. 23.09.1991
  E-Mail: e.schwarz@gmail.com
  Telefon: +49 89 1111 2222
  Adresse: Leopoldstraße 77, 80802 München
  IBAN: DE75 5121 0800 1245 1261 99
  Kundenseit: 15.11.2020`,
			Expected: map[string][]string{
				"DATE":        {"12.05.1968", "01.03.2018", "23.09.1991", "15.11.2020"},
				"EMAIL":       {"hj.becker@t-online.de", "e.schwarz@gmail.com"},
				"PHONE":       {"+49 221 9876543", "+49 89 1111 2222"},
				"CREDIT_CARD": {"4532 1234 5678 9012"},
				"IBAN":        {"DE75 5121 0800 1245 1261 99"},
			},
		},
		{
			Category: "CRM",
			Language: "Mixed",
			Text: `International Customer Records:

Customer: Pierre Martin, pierre.martin@orange.fr, +33 1 42 68 53 00
  Address: 8 Avenue des Champs-Élysées, 75008 Paris
  IBAN: FR76 1234 5678 9012 3456 7890 123

Customer: Giulia Romano, giulia.romano@libero.it, +39 06 1234 5678
  Address: Via del Corso 42, 00186 Roma
  IBAN: IT60 X054 2811 1010 0000 0123 456

Customer: Maria Fernandez, maria.fernandez@gmail.es, +34 91 555 1234
  Address: Calle Gran Vía 28, 28013 Madrid
  IBAN: ES91 2100 0418 4502 0005 1332

Customer: Jan Kowalski, jan.kowalski@wp.pl, +48 22 555 6789
  Address: ul. Nowy Świat 42, 00-363 Warszawa
  IBAN: PL61 1090 1014 0000 0712 1981 2874`,
			Expected: map[string][]string{
				"EMAIL": {
					"pierre.martin@orange.fr", "giulia.romano@libero.it",
					"maria.fernandez@gmail.es", "jan.kowalski@wp.pl",
				},
				"PHONE": {
					"+33 1 42 68 53 00", "+39 06 1234 5678",
					"+34 91 555 1234", "+48 22 555 6789",
				},
				"IBAN": {
					"FR76 1234 5678 9012 3456 7890 123",
					"IT60 X054 2811 1010 0000 0123 456",
					"ES91 2100 0418 4502 0005 1332",
					"PL61 1090 1014 0000 0712 1981 2874",
				},
			},
		},

		// ============================================================
		// CATEGORY: Legal / Contracts (Art. 6 DSGVO)
		// ============================================================
		{
			Category: "Legal",
			Language: "DE",
			Text: `Mietvertrag
zwischen Vermieter: Schmidt Immobilien GmbH, vertreten durch Herrn Klaus Schmidt
und Mieter: Frau Dr. Katharina Bauer, geboren am 08.12.1983

Mietobjekt: Schillerstraße 22, 60313 Frankfurt am Main
Mietbeginn: 01.04.2026 | Kaltmiete: €1.350,00 | Nebenkosten: €250,00

Bankverbindung Vermieter: IBAN DE89 3704 0044 0532 0130 00
Bankverbindung Mieter: IBAN DE44 5001 0517 0648 4898 90
Kontakt Vermieter: info@schmidt-immobilien.de, +49 69 12345678
Kontakt Mieter: k.bauer@posteo.de, +49 170 8765432

Personalausweis-Nr. Mieter: T220001293`,
			Expected: map[string][]string{
				"ORG":   {"Schmidt Immobilien GmbH"},
				"DATE":  {"08.12.1983", "01.04.2026"},
				"IBAN":  {"DE89 3704 0044 0532 0130 00", "DE44 5001 0517 0648 4898 90"},
				"EMAIL": {"info@schmidt-immobilien.de", "k.bauer@posteo.de"},
				"PHONE": {"+49 69 12345678", "+49 170 8765432"},
			},
		},

		// ============================================================
		// CATEGORY: Insurance Claims (Art. 9 DSGVO — health data)
		// ============================================================
		{
			Category: "Insurance",
			Language: "DE",
			Text: `Schadenmeldung Nr. SM-2026-14823 — Allianz Versicherungs-AG
Versicherungsnehmer: Peter Hoffmann, geb. 19.07.1975
Versicherungsnr.: AZ-4711-2345-6789
Telefon: +49 151 12345678 | E-Mail: p.hoffmann@web.de
Schadensdatum: 14.02.2026
Art des Schadens: Verkehrsunfall — Diagnose: S72.0 (Schenkelhalsfraktur)
Krankenhausaufenthalt: 14.02.2026 – 28.02.2026
Behandelnder Arzt: Dr. med. Friedrich Braun, +49 89 5555 4444
Unfallgegner: Maria Schulz, Kennzeichen M-AB 1234
Schadenhöhe geschätzt: €18.500,00
Überweisung auf: IBAN DE21 3702 0500 0008 0681 00`,
			Expected: map[string][]string{
				"ORG":     {"Allianz Versicherungs-AG"},
				"DATE":    {"19.07.1975", "14.02.2026", "28.02.2026"},
				"PHONE":   {"+49 151 12345678", "+49 89 5555 4444"},
				"EMAIL":   {"p.hoffmann@web.de"},
				"MEDICAL": {"S72.0"},
				"IBAN":    {"DE21 3702 0500 0008 0681 00"},
			},
		},

		// ============================================================
		// CATEGORY: IT / Security Data (Art. 32 DSGVO)
		// ============================================================
		{
			Category: "IT-Security",
			Language: "EN",
			Text: `SECURITY INCIDENT REPORT — 2026-02-15
Affected systems: 192.168.1.100, 10.0.0.42, 203.0.113.50
Compromised accounts: admin@company.com, root@server.internal.net
Exposed API keys:
  sk-proj-abc123def456ghi789jkl012mno345pqr678stu901vwx234
  AKIA1234567890ABCDEF
  ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12
MAC addresses of affected devices:
  00:1B:44:11:3A:B7, AA:BB:CC:DD:EE:FF
Incident reported by: John Miller, john.miller@security.com, +1 555 123 4567
Affected server URL: https://api.internal.company.com/v2/users`,
			Expected: map[string][]string{
				"IP_ADDRESS":  {"192.168.1.100", "10.0.0.42", "203.0.113.50"},
				"EMAIL":       {"admin@company.com", "root@server.internal.net", "john.miller@security.com"},
				"SECRET":      {"sk-proj-abc123def456ghi789jkl012mno345pqr678stu901vwx234", "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef12"},
				"MAC_ADDRESS": {"00:1B:44:11:3A:B7", "AA:BB:CC:DD:EE:FF"},
				"PHONE":       {"+1 555 123 4567"},
				"URL":         {"https://api.internal.company.com/v2/users"},
				"DATE":        {"2026-02-15"},
			},
		},
		{
			Category: "IT-Security",
			Language: "EN",
			Text: `Server Access Log Review
Suspicious login from IP 198.51.100.23 at 2026-01-20 14:23:51
User agent indicates automated scanning tool.
Private key found in repository:
  PRIVATE KEY: MIIEvQIBADANBgkqhkiG9w0BAQEFAASC
Database connection string leaked: postgres://admin:P@ssw0rd!@db.example.com:5432/production
Slack webhook: https://hooks.example.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
AWS access key: AKIAIOSFODNN7EXAMPLE
Contact security team: security@company.org, +44 20 7946 0958`,
			Expected: map[string][]string{
				"IP_ADDRESS": {"198.51.100.23"},
				"EMAIL":      {"security@company.org"},
				"PHONE":      {"+44 20 7946 0958"},
				"SECRET":     {"AKIAIOSFODNN7EXAMPLE"},
				"URL":        {"https://hooks.example.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"},
				"DATE":       {"2026-01-20"},
			},
		},

		// ============================================================
		// CATEGORY: Education (Art. 6 DSGVO — student data)
		// ============================================================
		{
			Category: "Education",
			Language: "DE",
			Text: `Immatrikulationsbescheinigung — Technische Universität München
Student: Lukas Zimmermann, Matrikelnr. 03712345
Geburtsdatum: 02.08.2000 | Geburtsort: Stuttgart
Adresse: Arcisstraße 21, 80333 München
E-Mail: lukas.zimmermann@tum.de | Tel: +49 176 2345 6789
Studiengang: Informatik (B.Sc.), 5. Fachsemester
Immatrikuliert seit: 01.10.2021
Semesterbeitrag überwiesen auf: IBAN DE02 7001 0080 0940 8038 01
Krankenkasse: Techniker Krankenkasse, Vers.-Nr.: T012345678`,
			Expected: map[string][]string{
				"ORG":   {"Technische Universität München", "Techniker Krankenkasse"},
				"DATE":  {"02.08.2000", "01.10.2021"},
				"EMAIL": {"lukas.zimmermann@tum.de"},
				"PHONE": {"+49 176 2345 6789"},
				"IBAN":  {"DE02 7001 0080 0940 8038 01"},
			},
		},

		// ============================================================
		// CATEGORY: Chat / Messaging (Art. 6 DSGVO — informal PII)
		// ============================================================
		{
			Category: "Chat",
			Language: "DE",
			Text: `[14:32] Lisa: Hey, kannst du mir die IBAN von Klaus schicken?
[14:33] Tom: Klar, ist DE89 3704 0044 0532 0130 00
[14:33] Lisa: Danke! Und seine neue Nummer?
[14:34] Tom: +49 176 9988 7766, die alte +49 170 1234567 geht nicht mehr
[14:35] Lisa: Alles klar. Sein Geburtstag war doch am 15.03.1985 oder?
[14:35] Tom: Ja genau. Und seine Mail ist jetzt klaus.meyer@gmail.com
[14:36] Lisa: Perfekt. Ich überweise ihm die 500€ für den Schrank
[14:37] Tom: Mach das. Seine Kreditkarte 4111 1111 1111 1111 nimmt er auch`,
			Expected: map[string][]string{
				"IBAN":        {"DE89 3704 0044 0532 0130 00"},
				"PHONE":       {"+49 176 9988 7766", "+49 170 1234567"},
				"DATE":        {"15.03.1985"},
				"EMAIL":       {"klaus.meyer@gmail.com"},
				"CREDIT_CARD": {"4111 1111 1111 1111"},
			},
		},
		{
			Category: "Chat",
			Language: "EN",
			Text: `From: Support Agent
To: Customer

Hi James, I found your account. Let me verify:
- Email on file: james.wilson@outlook.com
- Phone: +1 (212) 555-0147
- Last 4 of SSN: ... actually your full SSN 078-05-1120 was exposed in our logs.
- Card on file: 5500 0000 0000 0004
- Your address: 742 Evergreen Terrace, Springfield, IL 62701

Please change your password immediately. The breach was detected on 2026-02-10.
Incident ID: INC-2026-00523
Contact our security team at security@company.com or +1 800 555 0199.`,
			Expected: map[string][]string{
				"EMAIL":       {"james.wilson@outlook.com", "security@company.com"},
				"PHONE":       {"+1 (212) 555-0147", "+1 800 555 0199"},
				"SSN":         {"078-05-1120"},
				"CREDIT_CARD": {"5500 0000 0000 0004"},
				"DATE":        {"2026-02-10"},
			},
		},

		// ============================================================
		// CATEGORY: DSGVO Article 15 — Data Subject Access Request
		// ============================================================
		{
			Category: "DSAR",
			Language: "DE",
			Text: `Auskunftsersuchen gemäß Art. 15 DSGVO

Sehr geehrte Damen und Herren,

ich, Franziska Neumann (geb. 30.06.1989), bitte um Auskunft über alle bei Ihnen
gespeicherten personenbezogenen Daten zu meiner Person.

Meine Kontaktdaten: franziska.neumann@protonmail.com, +49 160 1234567
Anschrift: Goethestraße 12, 70174 Stuttgart
Personalausweis-Nr.: T220001293
Steuer-ID: DE987654321

Bitte senden Sie die Auskunft per verschlüsselter E-Mail an obige Adresse.
Falls Kosten anfallen, überweisen Sie bitte auf: IBAN DE89 3704 0044 0532 0130 00

Mit freundlichen Grüßen
Franziska Neumann`,
			Expected: map[string][]string{
				"DATE":  {"30.06.1989"},
				"EMAIL": {"franziska.neumann@protonmail.com"},
				"PHONE": {"+49 160 1234567"},
				"IBAN":  {"DE89 3704 0044 0532 0130 00"},
			},
		},

		// ============================================================
		// CATEGORY: Multi-language EU documents
		// ============================================================
		{
			Category: "EU-Multi",
			Language: "PT",
			Text: `Contrato de Trabalho — Portugal
Empregador: TAP Air Portugal, S.A.
Empregado: João Manuel da Silva Santos
NIF: PT123456789
Data de nascimento: 15.03.1990
Morada: Rua Augusta 42, 1100-053 Lisboa
Contacto: joao.santos@tap.pt, +351 21 841 5000
IBAN: PT50 0035 0726 0000 3587 7850 7
Salário bruto mensal: €2.800,00
Data de início: 01.03.2026`,
			Expected: map[string][]string{
				"ORG":   {"TAP Air Portugal"},
				"DATE":  {"15.03.1990", "01.03.2026"},
				"EMAIL": {"joao.santos@tap.pt"},
				"PHONE": {"+351 21 841 5000"},
				"IBAN":  {"PT50 0035 0726 0000 3587 7850 7"},
			},
		},
		{
			Category: "EU-Multi",
			Language: "FI",
			Text: `Työsopimus — Nokia Oyj
Työntekijä: Matti Virtanen
Henkilötunnus: 150390-1234
Osoite: Mannerheimintie 42, 00100 Helsinki
Puhelin: +358 40 123 4567 | Sähköposti: matti.virtanen@nokia.com
IBAN: FI21 1234 5600 0007 85
Bruttopalkka: €4.200,00/kk
Työsuhteen alkamispäivä: 01.03.2026`,
			Expected: map[string][]string{
				"ORG":   {"Nokia Oyj"},
				"DATE":  {"01.03.2026"},
				"EMAIL": {"matti.virtanen@nokia.com"},
				"PHONE": {"+358 40 123 4567"},
				"IBAN":  {"FI21 1234 5600 0007 85"},
			},
		},
		{
			Category: "EU-Multi",
			Language: "DK",
			Text: `Ansættelseskontrakt — Novo Nordisk A/S
Medarbejder: Lars Nielsen
CPR-nummer: 150390-1234
Adresse: Strøget 42, 1154 København K
Telefon: +45 44 44 88 88 | E-mail: lars.nielsen@novonordisk.com
IBAN: DK50 0040 0440 1162 43
Bruttoløn: 42.000,00 kr/md
Ansættelsesdato: 01.03.2026`,
			Expected: map[string][]string{
				"ORG":   {"Novo Nordisk A/S"},
				"DATE":  {"01.03.2026"},
				"EMAIL": {"lars.nielsen@novonordisk.com"},
				"PHONE": {"+45 44 44 88 88"},
				"IBAN":  {"DK50 0040 0440 1162 43"},
			},
		},

		// ============================================================
		// CATEGORY: Edge Cases / Adversarial
		// ============================================================
		{
			Category: "Edge",
			Language: "Mixed",
			Text: `PII at text boundaries:
test@example.com
+49 170 1234567
DE89 3704 0044 0532 0130 00
4111 1111 1111 1111
192.168.1.1
00:1A:2B:3C:4D:5E`,
			Expected: map[string][]string{
				"EMAIL":       {"test@example.com"},
				"PHONE":       {"+49 170 1234567"},
				"IBAN":        {"DE89 3704 0044 0532 0130 00"},
				"CREDIT_CARD": {"4111 1111 1111 1111"},
				"IP_ADDRESS":  {"192.168.1.1"},
				"MAC_ADDRESS": {"00:1A:2B:3C:4D:5E"},
			},
		},
		{
			Category: "Edge",
			Language: "Mixed",
			Text: `Consecutive PII without separators:
Email:test@example.com Phone:+49 170 1234567 IBAN:DE89370400440532013000

Multiple IBANs in one line:
DE89 3704 0044 0532 0130 00 and FR76 3000 6000 0112 3456 7890 189 and NL91 ABNA 0417 1643 00

PII in parentheses: (kontakt@firma.de) (+49 30 12345678) (DE89 3704 0044 0532 0130 00)`,
			Expected: map[string][]string{
				"EMAIL": {"test@example.com", "kontakt@firma.de"},
				"PHONE": {"+49 170 1234567", "+49 30 12345678"},
				"IBAN":  {"DE89 3704 0044 0532 0130 00", "FR76 3000 6000 0112 3456 7890 189", "NL91 ABNA 0417 1643 00"},
			},
		},
		{
			Category: "Edge",
			Language: "Mixed",
			Text: `Unicode edge cases:
Patient Müller-Lüdenscheidt, E-Mail: müller@büro.de
Adresse: Königstraße 42, 70173 Stuttgart
Téléphone: +33 6 98 76 54 32
名前: 田中太郎 (tanaka@email.co.jp)
Straße with ß: Große Eschenheimer Straße 42`,
			Expected: map[string][]string{
				"EMAIL": {"tanaka@email.co.jp"},
				"PHONE": {"+33 6 98 76 54 32"},
			},
		},

		// ============================================================
		// CATEGORY: False Positive Tests (should NOT detect PII)
		// ============================================================
		{
			Category: "FalsePositive",
			Language: "DE",
			Text: `Die Temperatur beträgt heute 25 Grad bei 1013 hPa Luftdruck.
Version 2.0.1 wurde am Montag veröffentlicht.
Die Produktionszahlen stiegen um 12% im Vergleich zum Vorquartal.
Der Umsatz betrug 1,5 Millionen Euro. Artikelnummer: 4711-0815.
Kapitel 3, Absatz 2, Satz 1 des BGB regelt dies.
HTTP Status 404 bedeutet "Not Found". Port 8080 ist belegt.
Die Konferenz findet in Raum 42 statt. Teilnehmer: 150 Personen.`,
			Expected: map[string][]string{},
		},
		{
			Category: "FalsePositive",
			Language: "EN",
			Text: `The server processed 1,234,567 requests yesterday.
Build number 20260215.1 was deployed to production.
Error code E-1234 occurred in module auth.handler.
The fibonacci sequence: 1, 1, 2, 3, 5, 8, 13, 21, 34, 55.
Reference number: ORD-2026-00042. Ticket ID: JIRA-1234.
Temperature: 98.6°F / 37°C. Distance: 42.195 km (marathon).`,
			Expected: map[string][]string{},
		},

		// ============================================================
		// CATEGORY: Bulk data export (realistic DSGVO Art. 20 portability)
		// ============================================================
		{
			Category: "DataExport",
			Language: "DE",
			Text: `Datenexport gemäß Art. 20 DSGVO — Datenportabilität
Exportdatum: 01.03.2026

Stammdaten:
  Name: Prof. Dr. Michael Braun
  Geburtsdatum: 22.11.1970
  E-Mail (privat): m.braun@gmx.de
  E-Mail (beruflich): michael.braun@uni-heidelberg.de
  Telefon (mobil): +49 172 3456789
  Telefon (Büro): +49 6221 54-0
  Adresse: Neuenheimer Feld 288, 69120 Heidelberg

Zahlungsinformationen:
  IBAN primär: DE89 3704 0044 0532 0130 00
  IBAN sekundär: DE44 5001 0517 0648 4898 90
  Kreditkarte: 4532 0123 4567 8901 (Visa, gültig 03/2028)

Gesundheitsdaten:
  Blutgruppe: AB positiv
  Allergien: Penicillin, Latex
  Diagnosen: E78.0 (Hypercholesterinämie), G43.0 (Migräne ohne Aura)
  Letzte Untersuchung: 15.01.2026

Arbeitgeber: Universität Heidelberg
Steuer-ID: DE123456789
Sozialversicherungsnr.: 12 150870 A 123`,
			Expected: map[string][]string{
				"DATE":        {"01.03.2026", "22.11.1970", "15.01.2026"},
				"EMAIL":       {"m.braun@gmx.de", "michael.braun@uni-heidelberg.de"},
				"PHONE":       {"+49 172 3456789"},
				"IBAN":        {"DE89 3704 0044 0532 0130 00", "DE44 5001 0517 0648 4898 90"},
				"CREDIT_CARD": {"4532 0123 4567 8901"},
				"MEDICAL":     {"E78.0", "G43.0"},
				"ORG":         {"Universität Heidelberg"},
			},
		},
	}
}

// entityMatch describes whether an expected entity was found.
type entityMatch struct {
	entityType string
	text       string
	found      bool
}

// TestGDPRBenchmark runs all DSGVO benchmark cases and reports comprehensive metrics.
func TestGDPRBenchmark(t *testing.T) {
	cases := gdprBenchmarkCases()
	s := DefaultScanner(nil)

	type categoryMetrics struct {
		overall metrics
		perType map[string]*metrics
	}
	categories := make(map[string]*categoryMetrics)

	var totalOverall metrics
	totalPerType := make(map[string]*metrics)

	for i, tc := range cases {
		detected := s.Scan(tc.Text)

		// Build expected list.
		var expectedList []entityMatch
		for typ, texts := range tc.Expected {
			for _, txt := range texts {
				expectedList = append(expectedList, entityMatch{entityType: typ, text: txt})
			}
		}

		// Match: detected entity matches expected if same type and detected text contains expected (or vice versa).
		matchedExpected := make([]bool, len(expectedList))
		matchedDetected := make([]bool, len(detected))

		for di, det := range detected {
			for ei := range expectedList {
				if matchedExpected[ei] {
					continue
				}
				exp := expectedList[ei]
				if det.Type != exp.entityType {
					continue
				}
				// Text overlap: one contains the other, or significant overlap.
				if strings.Contains(det.Text, exp.text) || strings.Contains(exp.text, det.Text) || textOverlap(det.Text, exp.text) > 0.6 {
					matchedExpected[ei] = true
					matchedDetected[di] = true
					break
				}
			}
		}

		// Compute metrics for this case.
		var caseOverall metrics
		casePerType := make(map[string]*metrics)

		for ei, exp := range expectedList {
			if _, ok := casePerType[exp.entityType]; !ok {
				casePerType[exp.entityType] = &metrics{}
			}
			if matchedExpected[ei] {
				casePerType[exp.entityType].TP++
				caseOverall.TP++
			} else {
				casePerType[exp.entityType].FN++
				caseOverall.FN++
			}
		}

		// Count FP: detected entities not matched to any expected.
		for di, det := range detected {
			if _, ok := casePerType[det.Type]; !ok {
				casePerType[det.Type] = &metrics{}
			}
			if !matchedDetected[di] {
				casePerType[det.Type].FP++
				caseOverall.FP++
			}
		}

		// Log case results.
		label := fmt.Sprintf("[%d] %s/%s", i+1, tc.Category, tc.Language)
		t.Logf("%-35s TP=%d FP=%d FN=%d  P=%.0f%% R=%.0f%% F1=%.0f%%",
			label, caseOverall.TP, caseOverall.FP, caseOverall.FN,
			caseOverall.Precision()*100, caseOverall.Recall()*100, caseOverall.F1()*100)

		// Log misses.
		for ei, exp := range expectedList {
			if !matchedExpected[ei] {
				t.Logf("    MISS: %s %q", exp.entityType, exp.text)
			}
		}

		// Log false positives (limit to avoid noise).
		fpCount := 0
		for di, det := range detected {
			if !matchedDetected[di] {
				fpCount++
				if fpCount <= 5 {
					t.Logf("    FP:   %s %q", det.Type, det.Text)
				}
			}
		}
		if fpCount > 5 {
			t.Logf("    ... and %d more false positives", fpCount-5)
		}

		// Accumulate totals.
		totalOverall.TP += caseOverall.TP
		totalOverall.FP += caseOverall.FP
		totalOverall.FN += caseOverall.FN
		mergeMetrics(totalPerType, casePerType)

		// Accumulate per category.
		cm, ok := categories[tc.Category]
		if !ok {
			cm = &categoryMetrics{perType: make(map[string]*metrics)}
			categories[tc.Category] = cm
		}
		cm.overall.TP += caseOverall.TP
		cm.overall.FP += caseOverall.FP
		cm.overall.FN += caseOverall.FN
		mergeMetrics(cm.perType, casePerType)
	}

	// Print per-type summary.
	printReport(t, fmt.Sprintf("GDPR Benchmark — Per Type (%d cases)", len(cases)), totalOverall, totalPerType)

	// Print per-category summary.
	catNames := make([]string, 0, len(categories))
	for c := range categories {
		catNames = append(catNames, c)
	}
	sort.Strings(catNames)

	t.Logf("")
	t.Logf("=== GDPR Benchmark — Per Category ===")
	t.Logf("%-16s %5s %5s %5s %9s %9s %9s",
		"Category", "TP", "FP", "FN", "Precision", "Recall", "F1")
	t.Logf("%-16s %5s %5s %5s %9s %9s %9s",
		"--------", "--", "--", "--", "---------", "------", "--")
	for _, cat := range catNames {
		cm := categories[cat]
		t.Logf("%-16s %5d %5d %5d %8.1f%% %8.1f%% %8.1f%%",
			cat, cm.overall.TP, cm.overall.FP, cm.overall.FN,
			cm.overall.Precision()*100, cm.overall.Recall()*100, cm.overall.F1()*100)
	}
	t.Logf("%-16s %5s %5s %5s %9s %9s %9s",
		"--------", "--", "--", "--", "---------", "------", "--")
	t.Logf("%-16s %5d %5d %5d %8.1f%% %8.1f%% %8.1f%%",
		"TOTAL", totalOverall.TP, totalOverall.FP, totalOverall.FN,
		totalOverall.Precision()*100, totalOverall.Recall()*100, totalOverall.F1()*100)
	t.Logf("")

	// Grade.
	recall := totalOverall.Recall()
	f1 := totalOverall.F1()
	grade := "D"
	switch {
	case f1 >= 0.90:
		grade = "A+"
	case f1 >= 0.85:
		grade = "A"
	case f1 >= 0.80:
		grade = "B+"
	case f1 >= 0.75:
		grade = "B"
	case f1 >= 0.70:
		grade = "C+"
	case f1 >= 0.60:
		grade = "C"
	case f1 >= 0.50:
		grade = "D"
	}
	t.Logf("=== FINAL GRADE: %s (Recall=%.1f%%, Precision=%.1f%%, F1=%.1f%%) ===",
		grade, recall*100, totalOverall.Precision()*100, f1*100)
	t.Logf("")
}

// textOverlap computes the fraction of characters in a that appear in b.
func textOverlap(a, b string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	// Normalize spaces for comparison.
	a = strings.ReplaceAll(a, " ", "")
	b = strings.ReplaceAll(b, " ", "")

	shorter, longer := a, b
	if len(a) > len(b) {
		shorter, longer = b, a
	}

	if strings.Contains(longer, shorter) {
		return float64(len(shorter)) / float64(len(longer))
	}
	return 0
}
