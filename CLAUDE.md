# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build                  # Build all 3 binaries (aegis-scan, aegis, aegis-server)
make build-server-nlp       # Build server with NLP/NER support (-tags nlp)
make test                   # Run all tests
make test-race              # Run tests with race detector
make test-nlp               # Run tests with NLP build tag
make bench                  # Benchmark scanner performance
make lint                   # go vet
make lint-nlp               # go vet with nlp tag
```

Run a single test:
```bash
go test ./internal/scanner/ -run TestName -v
```

## Build Tags

The `nlp` build tag controls NLP/NER support. Without it, only regex scanning is available (faster builds, smaller binaries). Files guarded by build tags come in pairs:
- `foo.go` (`//go:build nlp`) — real implementation
- `foo_stub.go` (`//go:build !nlp`) — no-op stub returning nil

Affected files: `internal/nlp/runtime*.go`, `internal/nlp/session*.go`, `internal/nlp/tokenizer.go`, `internal/scanner/ner_scanner*.go`, `internal/scanner/hybrid*.go`, `cmd/aegis-server/nlp_init*.go`

**NLP build requires**: CGO enabled + C compiler (gcc/MinGW) + ONNX Runtime shared library. Typically used in Docker/CI, not local dev.

## Architecture

**Scanner interface** (`internal/scanner/scanner.go`): All detectors implement `Scanner.Scan(text string) []Entity`. Entities carry byte offsets, type, score, and detector origin.

**RegexScanner**: Wraps a compiled regex for one entity type. Options: `WithValidator()` for post-match validation (e.g. Luhn check), `WithExtractGroup()` for capture groups, `WithContextValidator()` for surrounding-text checks.

**CompositeScanner**: Merges results from multiple scanners. Deduplication: sort by (start asc, length desc), longest match wins overlapping spans.

**Scanner chain**:
- Without NLP: `DefaultScanner(allowlist)` → `NewCompositeScanner(BuiltinScanners(), allowlist)`
- With NLP: `HybridScanner(allowlist, nerScanner)` → `NewCompositeScanner(BuiltinScanners() + NERScanner, allowlist)`

**NERScanner** (`internal/scanner/ner_scanner.go`, nlp tag): Wraps `nlp.NERSession` ONNX pipeline. Converts character offsets to byte offsets via `nlp.DecodeTokenOffsets()`. Maps NER labels (PER→PERSON, LOC→ADDRESS, ORG→ORG, DIS→MEDICAL, TIME→DATE).

**NLP Package** (`internal/nlp/`):
- `runtime.go` — ONNX Runtime init/shutdown via `yalue/onnxruntime_go`
- `tokenizer.go` — Pure Go Unigram (SentencePiece) tokenizer, loads HuggingFace `tokenizer.json`
- `session.go` — `NERSession`: tokenize → ONNX inference → BIO decode → `[]NERResult`
- `types.go` — Shared types (`NERResult`, `bioLabels`, MultiNERD label scheme with 31 BIO labels)

**Context enhancer** (`internal/scanner/context_enhancer.go`): `EnhanceScores()` boosts entity scores when multilingual context keywords appear within a configurable window around the match.

**Redaction** (`internal/redactor/`): Deterministic token assignment (e.g. `[PERSON_1]`), reverse-order replacement to preserve byte offsets.

**Restoration** (`internal/restorer/`): Replaces tokens back to original text, longest-first to avoid prefix collisions. Includes `StreamRestorer` for chunked processing.

## Three Binaries

- **aegis-scan** (`cmd/aegis-scan/`): CLI scanner. Reads from `--text`, `--file`, or stdin. Exit code 1 = PII found.
- **aegis** (`cmd/aegis/`): Interactive TUI (Bubble Tea). States: input → results → settings.
- **aegis-server** (`cmd/aegis-server/`): HTTP API on port 9090. Endpoints: `/health`, `/api/scan`, `/api/redact`, `/api/restore`. NLP init via build-tagged `nlp_init.go` / `nlp_init_stub.go`.

## Configuration

YAML config loaded via `config.Load(path)`. Sections: `scanner` (custom_patterns, allowlist), `nlp` (enabled, model_dir, ner_model, ner_tokenizer, max_length, num_threads), `context` (boost_factor, window_size), `logging` (level). See `config.example.yaml`.

Env vars: `AEGIS_SERVER_PORT`, `AEGIS_CORS_ORIGINS`.

## Training Pipeline

`training/` contains Python scripts for NER model training:
- `ner/train_base.py` — Fine-tune xlm-roberta-base on MultiNERD (Colab/local GPU)
- `ner/train_large.py` — Fine-tune xlm-roberta-large on MultiNERD (RunPod A100)
- `ner/export_onnx.py` — ONNX export + INT8 quantization
- `tokenizer/export_tokenizer.py` — Export tokenizer.json for Go loader

Output files go into `models/` directory (gitignored).

## Colab Training Gotchas

- **MultiNERD dataset**: Must use `verification_mode="no_checks"` in `load_dataset()` — shard verification fails otherwise.
- **transformers API changes**: Use `eval_strategy` (NOT `evaluation_strategy`) — renamed in transformers v4.46+.
- **Colab resets**: `/content/` is wiped on reconnect — uploaded scripts and pip packages must be re-uploaded/reinstalled. Always install deps (`!pip install seqeval accelerate optimum[onnxruntime]`) before running scripts.

## Key Gotchas

- **Typed nil**: `NewNERScanner()` returns `nil` if session is nil. Check before passing to `HybridScanner`.
- **Scanner order**: `BuiltinScanners()` ordering matters for overlap resolution — higher-priority scanners come first.
- **Byte vs character offsets**: Entity offsets are byte-based. NERScanner converts from character offsets via `nlp.DecodeTokenOffsets()`.
- **NFC normalization**: All input text is Unicode-normalized before scanning.
- **CGO requirement**: The `nlp` build tag requires CGO for `onnxruntime_go`. Without CGO, only the stub files compile.
