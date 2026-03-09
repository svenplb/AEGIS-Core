# AEGIS Core

<p align="center">
  <img width="1508" height="527" alt="image" src="https://github.com/user-attachments/assets/a0a1c806-303f-43fd-b123-3f55773f3ef4" />
</p>

PII detection and redaction engine written in Go. Scans text for personally identifiable information using regex patterns and optional NER (Named Entity Recognition), replaces matches with deterministic tokens, and can restore the original text.

## Detection

### Regex (built-in, all languages)

`EMAIL` `PHONE` `IBAN` `CREDIT_CARD` `SSN` `IP_ADDRESS` `URL` `SECRET` `FINANCIAL` `MEDICAL` `AGE` `ID_NUMBER` `MAC_ADDRESS` `DATE` `TAX_NUMBER` `SEPA`

Covers EU-wide formats: DE, AT, CH, FR, IT, ES, NL, BE, PL, PT, SE, DK, FI, NO, CZ, SK, IE, UK, and more.

### NER (optional, requires model)

`PERSON` `ADDRESS` `ORG` `MEDICAL` `DATE`

Detects free-text entities (names, locations, organizations) using a fine-tuned XLM-RoBERTa model. Supports DE, EN, FR, ES, IT, NL, PL, PT.

### Benchmark (Regex + NER combined)

| Metric | Score |
|---|---|
| **Overall Recall** | **91.5%** |
| Easy cases | 93% |
| Medium cases | 95% |
| Hard cases | 78% |
| False positives | 1 (across 30 test cases) |

Tested across 8 EU languages. Full benchmark: `python testdata/benchmark_nlp.py`

## Install

```bash
go install github.com/svenplb/aegis-core/cmd/aegis-scan@latest
```

Or build from source:

```bash
make build
```

This produces three binaries in `bin/`:

- **aegis-scan** — CLI scanner
- **aegis** — interactive TUI
- **aegis-server** — HTTP API server

## Usage

### CLI

```bash
# inline text
aegis-scan --text "Call me at +49 170 1234567"

# from file
aegis-scan --file document.txt

# from stdin
cat document.txt | aegis-scan

# JSON output
aegis-scan --text "john@example.com" --json
```

Exit codes: `0` = no PII found, `1` = PII found, `2` = error.

### TUI

```bash
aegis
```

Paste text, press `Ctrl+D` to scan. `Tab` opens settings. `q` to quit.

### HTTP Server

```bash
aegis-server                    # default port 9090
aegis-server --port 8080        # custom port
aegis-server --config config.yaml
```

Env vars: `AEGIS_SERVER_PORT`, `AEGIS_CORS_ORIGINS`.

Open `http://localhost:9090` for the built-in web UI.

#### Endpoints

**GET /health**

```json
{"status": "ok", "version": "0.1.0"}
```

**POST /api/scan** — detect entities

```bash
curl -X POST localhost:9090/api/scan \
  -H "Content-Type: application/json" \
  -d '{"text": "Max Mustermann, IBAN DE89 3704 0044 0532 0130 00"}'
```

**POST /api/redact** — detect and replace with tokens

```bash
curl -X POST localhost:9090/api/redact \
  -H "Content-Type: application/json" \
  -d '{"text": "Email me at hans@example.com"}'
```

Returns `sanitized_text`, `entities`, and `mappings`.

**POST /api/restore** — restore tokens to original text

```bash
curl -X POST localhost:9090/api/restore \
  -H "Content-Type: application/json" \
  -d '{"text": "Email me at [EMAIL_1]", "mappings": [{"token": "[EMAIL_1]", "original": "hans@example.com", "type": "EMAIL"}]}'
```

## NLP/NER (optional)

By default, aegis-core uses regex-only detection. For free-text entity recognition (names, locations, organizations), enable the NER model:

### 1. Download the model

```bash
make download-model
```

Downloads the pre-trained NER model (~280 MB) from GitHub Releases into `models/`.

### 2. Run with Docker

```bash
make docker-nlp
docker run --rm -p 9090:9090 \
  -v ./models:/app/models \
  -v ./config.nlp.yaml:/app/config.yaml \
  aegis-nlp
```

Create a `config.nlp.yaml` (or copy from `config.example.yaml`) with:

```yaml
nlp:
  enabled: true
  model_dir: "models"
```

### Training your own model

You can train a better NER model yourself using a free GPU on [Kaggle](https://kaggle.com) or [Google Colab](https://colab.research.google.com). The model is based on XLM-RoBERTa and trained on the [MultiNERD](https://huggingface.co/datasets/Babelscape/multinerd) dataset (164K sentences, 10 languages).

**Step 1: Set up the environment**

Open a Kaggle Notebook (or Colab), select a **GPU runtime**, and install dependencies:

```bash
pip install torch transformers datasets seqeval accelerate optimum[onnxruntime]
```

**Step 2: Train the model**

Upload `training/ner/train_ner.py` and run it. Training takes ~2-4 hours on a T4 GPU:

```bash
python train_ner.py --output_dir ./ner-model --epochs 5 --batch_size 16
```

Alternatively, use `training/ner/train_kaggle.py` which is a self-contained script you can paste directly into a Kaggle notebook cell.

**Step 3: Export to ONNX**

Convert the trained PyTorch model to ONNX format with INT8 quantization (~280 MB instead of ~1.1 GB):

```bash
python training/ner/export_onnx.py --model_dir ./ner-model --output_dir ./output
```

This produces two files:
- `ner.onnx` — the quantized ONNX model (~280 MB)
- `tokenizer.json` — the tokenizer config (~16 MB)

**Step 4: Use it**

Download both files and place them in the `models/` directory. Start the server with NER enabled and your new model is active.

**Contributing a better model**: Open an issue with your benchmark results (`python testdata/benchmark_nlp.py`) and share the two files (`ner.onnx` + `tokenizer.json`). If it improves recall, it will be included in a new release.

## Configuration

Optional YAML config file (see `config.example.yaml`):

```yaml
scanner:
  allowlist:
    - "example\\.com"
nlp:
  enabled: true
  model_dir: "models"
context:
  boost_factor: 0.15
  window_size: 50
logging:
  level: info
```

## Docker

```bash
# Regex-only (lightweight, no model needed)
make docker-server
docker run -p 9090:9090 aegis-server

# With NLP/NER support (requires model in models/)
make docker-nlp
docker run -p 9090:9090 -v ./models:/app/models aegis-nlp
```

Both variants serve the web UI at `http://localhost:9090`.

## Test

```bash
make test        # all tests
make test-race   # with race detector
make bench       # scanner benchmarks
make lint        # go vet
```

## License

PolyForm Noncommercial 1.0.0 — free for personal, academic, and non-profit use. Commercial use is not permitted. See [LICENSE](LICENSE) for details.
