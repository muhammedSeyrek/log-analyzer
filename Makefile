# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

# Binary name. Used for both Windows .exe and Linux output.
BINARY      := log-analyzer
PKG         := ./cmd

# Docker image name and tag.
IMAGE       := log-analyzer
TAG         := latest

# Detect the host OS so we can name the binary correctly. Windows users
# get .exe, others get a plain executable.
ifeq ($(OS),Windows_NT)
	BIN_EXT := .exe
else
	BIN_EXT :=
endif

# -----------------------------------------------------------------------------
# Help — printed by default when `make` is run with no arguments.
# -----------------------------------------------------------------------------

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "log-analyzer Makefile"
	@echo ""
	@echo "Local development:"
	@echo "  make build          Generate plugin list and compile the binary"
	@echo "  make run            Build and show --help"
	@echo "  make list           Build and run the 'list' command"
	@echo "  make demo           Build and run a static analysis on dummy.log"
	@echo "  make clean          Remove generated artifacts"
	@echo ""
	@echo "Code quality:"
	@echo "  make generate       Regenerate plugins/plugins_gen.go"
	@echo "  make fmt            Run gofmt on the whole tree"
	@echo "  make vet            Run go vet on the whole tree"
	@echo ""
	@echo "Docker workflow:"
	@echo "  make docker         Build the runtime Docker image"
	@echo "  make docker-run     Run the container with --help"
	@echo "  make docker-list    Run 'list' inside the container"
	@echo "  make extract        Build and extract the Linux binary to ./out"
	@echo "  make docker-clean   Remove the Docker image"
	@echo ""
	@echo "Release packaging:"
	@echo "  make zip            Create a deliverable zip (excludes build artifacts)"

# -----------------------------------------------------------------------------
# Local development
# -----------------------------------------------------------------------------

.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build: generate
	go build -o $(BINARY)$(BIN_EXT) $(PKG)
	@echo "Built $(BINARY)$(BIN_EXT)"

.PHONY: run
run: build
	./$(BINARY)$(BIN_EXT) --help

.PHONY: list
list: build
	./$(BINARY)$(BIN_EXT) list

.PHONY: demo
demo: build
	./$(BINARY)$(BIN_EXT) loganalyzer static --file dummy.log --report

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: live
live: build
	./$(BINARY)$(BIN_EXT) loganalyzer live

.PHONY: advanced-demo
advanced-demo: build
	./$(BINARY)$(BIN_EXT) loganalyzer static --file advanced_dummy.log --report

.PHONY: clean
ifeq ($(OS),Windows_NT)
clean:
	-del /Q $(BINARY).exe 2>nul
	-del /Q log-analyzer-linux 2>nul
	-del /Q log-analyzer 2>nul
	-rmdir /S /Q out 2>nul
	-rmdir /S /Q reports 2>nul
	@echo Cleaned build artifacts
else
clean:
	-rm -f $(BINARY) $(BINARY).exe log-analyzer-linux log-analyzer
	-rm -rf out reports
	@echo "Cleaned build artifacts"
endif
# -----------------------------------------------------------------------------
# Docker workflow
# -----------------------------------------------------------------------------

.PHONY: docker
docker:
	docker build --target runner -t $(IMAGE):$(TAG) .

.PHONY: docker-run
docker-run: docker
	docker run --rm $(IMAGE):$(TAG)

.PHONY: docker-list
docker-list: docker
	docker run --rm $(IMAGE):$(TAG) list

.PHONY: extract
extract:
	docker build --target bin --output type=local,dest=./out .
	@echo "Linux binary extracted to ./out/log-analyzer"

.PHONY: docker-clean
docker-clean:
	-docker rmi $(IMAGE):$(TAG)
	-docker rmi $(IMAGE)-bin:latest

.PHONY: zip
zip: clean
	@echo "Creating delivery archive..."
	@powershell -Command "Compress-Archive -Path '*.go','*.md','Dockerfile','docker-compose.yml','Makefile','.dockerignore','go.mod','go.sum','cmd','config','internal','plugins','tools','dummy.log' -DestinationPath 'log-analyzer.zip' -Force"
	@echo "Created log-analyzer.zip"

	