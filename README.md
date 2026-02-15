# AEGIS Core
  <p align="center">
   <img width="703" height="778" alt="image" src="https://github.com/user-attachments/assets/a34b84c7-3684-4d17-bb19-738a6835e647" />
  </p>

PII detection and redaction engine written in Go. Scans text for personally identifiable information using regex-based pattern matching, replaces matches with deterministic tokens, and can restore the original text from those tokens.

## Detected entity types

`PERSON` `EMAIL` `PHONE` `ADDRESS` `DATE` `IBAN` `CREDIT_CARD` `IP_ADDRESS` `URL` `SECRET` `FINANCIAL` `SSN` `MEDICAL` `AGE` `ID_NUMBER` `ORG` `MAC_ADDRESS`

## Install

```
go install github.com/svenplb/aegis-core/cmd/aegis-scan@latest
```

Or build from source:

```
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

Paste text, press `Ctrl+D` to scan. `Tab` opens settings (score threshold, allowlist). `q` to quit from results.

### HTTP server

```bash
aegis-server                    # default port 9090
aegis-server --port 8080        # custom port
aegis-server --config config.yaml
```

Port can also be set via `AEGIS_SERVER_PORT`. CORS origin via `AEGIS_CORS_ORIGINS` (default `*`).

#### Endpoints

**GET /health**

```json
{"status": "ok", "version": "0.1.0"}
```

**POST /api/scan** — detect entities

```bash
curl -X POST localhost:9090/api/scan \
  -H "Content-Type: application/json" \
  -d '{"text": "Call Dr. Schmidt at +49 170 1234567"}'
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

## Configuration

Optional YAML config file (see `config.example.yaml`):

```yaml
scanner:
  allowlist:
    - "example\\.com"
logging:
  level: info
```

Pass with `--config config.yaml` to `aegis-scan` or `aegis-server`.

## Docker

```bash
make docker-server
docker run -p 9090:9090 aegis-server
```

## Test

```bash
make test        # all tests
make test-race   # with race detector
make bench       # scanner benchmarks
make lint        # go vet
```

## License

MIT
