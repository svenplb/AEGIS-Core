.PHONY: build build-scan build-tui build-server build-server-nlp test test-race test-nlp lint lint-nlp clean docker-server docker-nlp docker-nlp-run benchmark-accuracy bench download-model

BINARY_SCAN   := bin/aegis-scan
BINARY_TUI    := bin/aegis
BINARY_SERVER := bin/aegis-server

build: build-scan build-tui build-server

build-scan:
	go build -o $(BINARY_SCAN) ./cmd/aegis-scan

build-tui:
	go build -o $(BINARY_TUI) ./cmd/aegis

build-server:
	go build -o $(BINARY_SERVER) ./cmd/aegis-server

build-server-nlp:
	go build -tags nlp -o $(BINARY_SERVER) ./cmd/aegis-server

docker-server:
	docker build -f Dockerfile.server -t aegis-server .

docker-nlp:
	docker build -f Dockerfile.nlp -t aegis-nlp .

docker-nlp-run:
	docker run --rm -p 9090:9090 -v ./models:/app/models aegis-nlp

test:
	go test ./... -v

test-race:
	go test ./... -race -v

test-nlp:
	go test -tags nlp ./... -v

bench:
	go test ./internal/scanner/ -bench=. -benchmem

benchmark-accuracy:
	go test ./internal/scanner/ -run TestBenchmark -v -count=1

lint:
	go vet ./...

lint-nlp:
	go vet -tags nlp ./...

MODEL_REPO  := svenplb/aegis-core
MODEL_TAG   := v0.2.0

download-model:
	@mkdir -p models
	@echo "Downloading NER model from GitHub Releases ($(MODEL_TAG))..."
	curl -L -o models/ner.onnx "https://github.com/$(MODEL_REPO)/releases/download/$(MODEL_TAG)/ner.onnx"
	curl -L -o models/tokenizer.json "https://github.com/$(MODEL_REPO)/releases/download/$(MODEL_TAG)/tokenizer.json"
	@echo "Done. Models saved to models/"

clean:
	rm -rf bin/
