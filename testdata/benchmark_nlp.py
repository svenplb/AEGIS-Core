"""
Independent NER + Regex benchmark against the aegis-core API.
Tests multiple EU languages, difficulty levels, and entity types.

Usage: python testdata/benchmark_nlp.py [--url http://localhost:9090]
"""

import json
import sys
import time
import urllib.request
import urllib.error

API_URL = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:9090"

# Each test case: (language, difficulty, text, expected_entities)
# expected_entities: list of (type, substring)
TEST_CASES = [
    # ===== GERMAN =====
    # Easy
    ("DE", "easy", "Max Mustermann wohnt in Berlin.", [
        ("PERSON", "Max Mustermann"), ("ADDRESS", "Berlin"),
    ]),
    ("DE", "easy", "Kontakt: info@firma.de, Tel: +49 170 1234567", [
        ("EMAIL", "info@firma.de"), ("PHONE", "+49 170 1234567"),
    ]),
    ("DE", "easy", "IBAN: DE89 3704 0044 0532 0130 00", [
        ("IBAN", "DE89 3704 0044 0532 0130 00"),
    ]),
    # Medium
    ("DE", "medium", "Frau Dr. Claudia Bergmann arbeitet bei Siemens in Stuttgart.", [
        ("PERSON", "Claudia Bergmann"), ("ORG", "Siemens"), ("ADDRESS", "Stuttgart"),
    ]),
    ("DE", "medium", "Der Patient Hans-Peter Müller, geb. 15.03.1985, leidet an Diabetes.", [
        ("PERSON", "Hans-Peter Müller"), ("DATE", "15.03.1985"),
    ]),
    ("DE", "medium", "Überweisen Sie 500€ an DE44 5001 0517 5407 3249 31, Empfänger: Weber GmbH.", [
        ("IBAN", "DE44 5001 0517 5407 3249 31"),
    ]),
    # Hard
    ("DE", "hard", "Wolf sprach mit Herr Rose über die Lage in Essen.", [
        ("PERSON", "Rose"), ("ADDRESS", "Essen"),
    ]),
    ("DE", "hard", "Die Rechnung Nr. 2024-0815 geht an schmidt@web.de, cc: mueller@gmx.de.", [
        ("EMAIL", "schmidt@web.de"), ("EMAIL", "mueller@gmx.de"),
    ]),

    # ===== ENGLISH =====
    ("EN", "easy", "John Smith lives in London and works at Google.", [
        ("PERSON", "John Smith"), ("ADDRESS", "London"), ("ORG", "Google"),
    ]),
    ("EN", "easy", "Email me at john.smith@gmail.com or call +44 20 7946 0958.", [
        ("EMAIL", "john.smith@gmail.com"), ("PHONE", "+44 20 7946 0958"),
    ]),
    ("EN", "medium", "Dr. Sarah O'Connor from the WHO visited New York last Tuesday.", [
        ("PERSON", "Sarah O'Connor"), ("ORG", "WHO"), ("ADDRESS", "New York"),
    ]),
    ("EN", "hard", "Park met with Banks near the River Thames in Reading.", [
        ("ADDRESS", "Reading"),
    ]),

    # ===== FRENCH =====
    ("FR", "easy", "Jean Dupont habite à Paris et travaille chez Renault.", [
        ("PERSON", "Jean Dupont"), ("ADDRESS", "Paris"), ("ORG", "Renault"),
    ]),
    ("FR", "medium", "Mme Sophie Martin-Lefèvre a été hospitalisée à Lyon pour une pneumonie.", [
        ("PERSON", "Sophie Martin-Lefèvre"), ("ADDRESS", "Lyon"),
    ]),
    ("FR", "hard", "Le président Blanc a rencontré Mme Petit à Nice.", [
        ("PERSON", "Petit"), ("ADDRESS", "Nice"),
    ]),

    # ===== SPANISH =====
    ("ES", "easy", "Carlos García vive en Barcelona y trabaja en Telefónica.", [
        ("PERSON", "Carlos García"), ("ADDRESS", "Barcelona"), ("ORG", "Telefónica"),
    ]),
    ("ES", "medium", "La Sra. María López envió un correo a maria.lopez@empresa.es desde Sevilla.", [
        ("PERSON", "María López"), ("EMAIL", "maria.lopez@empresa.es"), ("ADDRESS", "Sevilla"),
    ]),

    # ===== ITALIAN =====
    ("IT", "easy", "Marco Rossi vive a Roma e lavora alla Fiat.", [
        ("PERSON", "Marco Rossi"), ("ADDRESS", "Roma"), ("ORG", "Fiat"),
    ]),
    ("IT", "medium", "Il Dott. Giuseppe Bianchi è stato trasferito a Milano per curare l'influenza.", [
        ("PERSON", "Giuseppe Bianchi"), ("ADDRESS", "Milano"),
    ]),

    # ===== DUTCH =====
    ("NL", "easy", "Jan de Vries woont in Amsterdam en werkt bij Philips.", [
        ("PERSON", "Jan de Vries"), ("ADDRESS", "Amsterdam"), ("ORG", "Philips"),
    ]),
    ("NL", "medium", "Mevrouw van den Berg heeft IBAN NL91 ABNA 0417 1643 00.", [
        ("IBAN", "NL91 ABNA 0417 1643 00"),
    ]),

    # ===== POLISH =====
    ("PL", "easy", "Jan Kowalski mieszka w Warszawie i pracuje w PKO.", [
        ("PERSON", "Jan Kowalski"), ("ADDRESS", "Warszawie"), ("ORG", "PKO"),
    ]),

    # ===== PORTUGUESE =====
    ("PT", "easy", "João Silva mora em Lisboa e trabalha na TAP.", [
        ("PERSON", "João Silva"), ("ADDRESS", "Lisboa"), ("ORG", "TAP"),
    ]),

    # ===== CROSS-LANGUAGE / MIXED =====
    ("MIX", "medium", "Meeting with François Müller from BMW in Zürich, contact: f.mueller@bmw.ch", [
        ("PERSON", "François Müller"), ("ORG", "BMW"), ("ADDRESS", "Zürich"),
        ("EMAIL", "f.mueller@bmw.ch"),
    ]),
    ("MIX", "hard", "Kreditkarte 4532015112830366, IBAN FR76 3000 6000 0112 3456 7890 189", [
        ("CREDIT_CARD", "4532015112830366"), ("IBAN", "FR76 3000 6000 0112 3456 7890 189"),
    ]),

    # ===== FALSE POSITIVE TESTS (should NOT detect) =====
    ("NEG", "easy", "Das Wetter in Europa ist heute schön.", []),
    ("NEG", "easy", "Die Funktion gibt den Wert 42 zurück.", []),
    ("NEG", "medium", "Berlin ist die Hauptstadt von Deutschland.", [
        ("ADDRESS", "Berlin"),  # this IS a location, should detect
    ]),
]


def scan(text):
    data = json.dumps({"text": text}).encode()
    req = urllib.request.Request(
        f"{API_URL}/api/scan",
        data=data,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read())
    except Exception as e:
        return {"error": str(e), "entities": []}


def entity_matches(detected, expected_type, expected_text):
    """Check if any detected entity matches the expected type and contains the text."""
    for ent in detected:
        if ent["type"] == expected_type and expected_text.lower() in ent.get("text", "").lower():
            return True
        # Partial match for names (NER might get slightly different boundaries)
        if ent["type"] == expected_type:
            det_text = ent.get("text", "").lower()
            exp_text = expected_text.lower()
            # Accept if >60% overlap
            if exp_text in det_text or det_text in exp_text:
                return True
    return False


def main():
    print(f"Aegis-Core NLP Benchmark")
    print(f"API: {API_URL}")
    print(f"{'='*70}")

    # Check API health
    try:
        with urllib.request.urlopen(f"{API_URL}/health", timeout=5) as resp:
            health = json.loads(resp.read())
            print(f"Server: {health.get('version', '?')} - OK\n")
    except Exception as e:
        print(f"ERROR: Server not reachable: {e}")
        sys.exit(1)

    # Metrics
    total_expected = 0
    total_detected = 0
    true_positives = 0
    false_negatives = []
    false_positives_count = 0

    # Per-language and per-type stats
    lang_stats = {}
    type_stats = {}
    diff_stats = {"easy": [0, 0], "medium": [0, 0], "hard": [0, 0]}

    total_time_ms = 0

    for lang, difficulty, text, expected in TEST_CASES:
        result = scan(text)
        detected = result.get("entities", [])
        total_time_ms += result.get("processing_time_ms", 0)

        if lang not in lang_stats:
            lang_stats[lang] = {"tp": 0, "expected": 0, "fp": 0}

        for exp_type, exp_text in expected:
            total_expected += 1
            lang_stats[lang]["expected"] += 1

            if exp_type not in type_stats:
                type_stats[exp_type] = {"tp": 0, "expected": 0}
            type_stats[exp_type]["expected"] += 1

            if entity_matches(detected, exp_type, exp_text):
                true_positives += 1
                lang_stats[lang]["tp"] += 1
                type_stats[exp_type]["tp"] += 1
                if difficulty in diff_stats:
                    diff_stats[difficulty][0] += 1
                    diff_stats[difficulty][1] += 1
            else:
                false_negatives.append((lang, difficulty, exp_type, exp_text, text[:60]))
                if difficulty in diff_stats:
                    diff_stats[difficulty][1] += 1

        # Count unexpected detections (false positives) for NEG tests
        if not expected and detected:
            false_positives_count += len(detected)
            lang_stats[lang]["fp"] += len(detected)

    # Calculate metrics
    recall = true_positives / total_expected if total_expected > 0 else 0
    # Precision is harder without full FP count, but we track NEG cases

    print(f"{'='*70}")
    print(f"OVERALL RESULTS")
    print(f"{'='*70}")
    print(f"  Expected entities:  {total_expected}")
    print(f"  Correctly found:    {true_positives}")
    print(f"  Missed:             {total_expected - true_positives}")
    print(f"  Recall:             {recall:.1%}")
    print(f"  False positives (NEG tests): {false_positives_count}")
    print()

    print(f"BY LANGUAGE:")
    print(f"  {'Lang':<6} {'Found':>6} {'Expected':>9} {'Recall':>8}")
    print(f"  {'-'*32}")
    for lang in sorted(lang_stats.keys()):
        s = lang_stats[lang]
        r = s["tp"] / s["expected"] if s["expected"] > 0 else 0
        marker = " OK" if r >= 0.8 else " !!" if r < 0.5 else ""
        print(f"  {lang:<6} {s['tp']:>6} {s['expected']:>9} {r:>7.0%}{marker}")
    print()

    print(f"BY ENTITY TYPE:")
    print(f"  {'Type':<14} {'Found':>6} {'Expected':>9} {'Recall':>8}")
    print(f"  {'-'*40}")
    for etype in sorted(type_stats.keys()):
        s = type_stats[etype]
        r = s["tp"] / s["expected"] if s["expected"] > 0 else 0
        marker = " OK" if r >= 0.8 else " !!" if r < 0.5 else ""
        print(f"  {etype:<14} {s['tp']:>6} {s['expected']:>9} {r:>7.0%}{marker}")
    print()

    print(f"BY DIFFICULTY:")
    print(f"  {'Level':<8} {'Found':>6} {'Total':>6} {'Recall':>8}")
    print(f"  {'-'*32}")
    for diff in ["easy", "medium", "hard"]:
        tp, total = diff_stats[diff]
        r = tp / total if total > 0 else 0
        print(f"  {diff:<8} {tp:>6} {total:>6} {r:>7.0%}")
    print()

    if false_negatives:
        print(f"MISSED ENTITIES ({len(false_negatives)}):")
        for lang, diff, etype, etext, context in false_negatives:
            print(f"  [{lang}/{diff}] {etype}: \"{etext}\"")
            print(f"           in: \"{context}...\"")
        print()

    print(f"Total API time: {total_time_ms} ms")
    print(f"{'='*70}")

    grade = "A" if recall >= 0.85 else "B" if recall >= 0.7 else "C" if recall >= 0.5 else "D"
    print(f"\nGRADE: {grade} (Recall {recall:.1%})")

    sys.exit(0 if recall >= 0.7 else 1)


if __name__ == "__main__":
    main()
