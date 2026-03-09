package scanner

import "testing"

func TestHybridScanner_NilNER(t *testing.T) {
	sc := HybridScanner(nil, nil)
	if sc == nil {
		t.Fatal("HybridScanner with nil NER should return a valid scanner")
	}

	entities := sc.Scan("test@example.com")
	found := false
	for _, e := range entities {
		if e.Type == "EMAIL" {
			found = true
		}
	}
	if !found {
		t.Error("regex detection should still work with nil NER")
	}
}

func TestHybridScanner_FallsBackToRegex(t *testing.T) {
	// Without NLP build tag, HybridScanner should behave like DefaultScanner.
	sc := HybridScanner(nil, nil)
	defaultSc := DefaultScanner(nil)

	text := "Hans hat IBAN DE89370400440532013000 und Telefon +49 170 1234567"

	hybridEntities := sc.Scan(text)
	defaultEntities := defaultSc.Scan(text)

	if len(hybridEntities) != len(defaultEntities) {
		t.Errorf("HybridScanner(nil NER) found %d entities, DefaultScanner found %d",
			len(hybridEntities), len(defaultEntities))
	}

	for i := range hybridEntities {
		if i >= len(defaultEntities) {
			break
		}
		if hybridEntities[i].Type != defaultEntities[i].Type {
			t.Errorf("entity %d: hybrid=%s, default=%s", i, hybridEntities[i].Type, defaultEntities[i].Type)
		}
	}
}
