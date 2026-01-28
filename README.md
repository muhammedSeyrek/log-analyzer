Option 1: Running with Docker (Recommended)
This project is containerized. To run it without installing Go on your machine:

Build the Image:

Bash
docker build -t log-analyzer .
Run the Container: We need to map the current directory to the container to access logs and save reports.

Windows (PowerShell):

PowerShell
docker run -it --rm -v ${PWD}:/root/ log-analyzer
Linux / macOS:

Bash
docker run -it --rm -v $(pwd):/root/ log-analyzer
Option 2: Running Locally (Go Installed)
Initialize Modules:

Bash
go mod tidy
Run the Application:

Bash
go run cmd/main.go