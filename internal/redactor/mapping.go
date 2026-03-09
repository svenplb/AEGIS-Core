package redactor

// Mapping links a placeholder token to its original text.
type Mapping struct {
	Token    string `json:"token"`    // e.g. "[PERSON_1]"
	Original string `json:"original"` // e.g. "Thomas Schmidt"
	Type     string `json:"type"`     // e.g. "PERSON"
}
