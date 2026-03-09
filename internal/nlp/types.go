package nlp

// NERResult represents a named entity recognized by the NER model.
type NERResult struct {
	Text  string  // matched text from the original input
	Label string  // NER label: "PER", "LOC", "ORG", etc.
	Start int     // character offset in original text
	End   int     // character offset in original text (exclusive)
	Score float64 // average token confidence
}

// BIO label definitions for MultiNERD.
// The model outputs logits with indices corresponding to these labels.
var bioLabels = []string{
	"O",
	"B-PER", "I-PER",
	"B-ORG", "I-ORG",
	"B-LOC", "I-LOC",
	"B-ANIM", "I-ANIM",
	"B-BIO", "I-BIO",
	"B-CEL", "I-CEL",
	"B-DIS", "I-DIS",
	"B-EVE", "I-EVE",
	"B-FOOD", "I-FOOD",
	"B-INST", "I-INST",
	"B-MEDIA", "I-MEDIA",
	"B-MYTH", "I-MYTH",
	"B-PLANT", "I-PLANT",
	"B-TIME", "I-TIME",
	"B-VEHI", "I-VEHI",
}

// NumLabels is the total number of BIO labels.
const NumLabels = 31
