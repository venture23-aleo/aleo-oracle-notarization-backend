# Makefile Guide

A `Makefile` is provided to simplify the most common development, build, and deployment tasks. The Makefile will automatically use variables from your `.env` file if present, so you can configure your environment in one place.

## Common Make Targets

### Development Targets
- `make build`         – Build the Go binary for your application.
- `make run`           – Build and run the application locally.
- `make test`          – Run unit tests.
- `make fmt`           – Format code using go fmt.
- `make vet`           – Static vetting of code.
- `make lint`          – Static analysis using staticcheck.
- `make clean`         – Remove built binaries.

### Native Targets
- `make native-build`      – Build the native binary for Gramine.
- `make native-run`        – Build and run the application natively with Gramine.
- `make native-clean`      – Clean native build artifacts.

### Docker Targets
- `make docker-build`  – Build the Docker image (including manifest generation).
- `make docker-run`    – Build the Docker image and run the container using Docker Compose (detached mode).
- `make docker-run-fg` – Build the Docker image and run the container in foreground mode.
- `make docker-run-rebuild` – Build the Docker image and run the container with force rebuild.

### SGX/Enclave Targets
- `make generate-enclave-signing-key` – Generate the private key for SGX signing (OpenSSL).
- `make generate-manifest-template` – Generate manifest template.
- `make extract-enclave-artifacts` – Extract enclave artifacts to verify the enclave properties

### Utility
- `make help`          – Show a summary of available make targets.

## Docker Variables

The Makefile supports several variables that can be overridden from the command line:

- `DOCKER_FLAGS` - Docker compose flags (default: `-d` for detached mode)
- `DOCKER_SERVICES` - Services to run (default: `aleo-oracle-notarization-backend`)
- `APP` - Application name (default: `aleo-oracle-notarization-backend`)

## Example Usage

### Basic Usage
```sh
make build         # Compile the Go binary
make run           # Build and run locally
make test          # Run tests
make native-build  # Build native binary for Gramine
make native-run    # Build and run natively with Gramine
make docker-build  # Build Docker image (with manifest)
make docker-run    # Build and run with Docker Compose (detached)
make generate-enclave-signing-key # Generate the private key for SGX signing (OpenSSL)
make clean         # Clean up binaries
make help          # Show help
```

### Advanced Native Usage
```sh
# Clean native build artifacts
make native-clean

# Build only (without running)
make native-build

# Run native application with Gramine
make native-run
```

### Advanced Docker Usage
```sh
# Run in foreground mode
make docker-run DOCKER_FLAGS=""

# Run with additional flags
make docker-run DOCKER_FLAGS="-d --force-recreate"

# Run specific services
make docker-run DOCKER_SERVICES="app db"

# Run with environment variables
make docker-run DOCKER_FLAGS="-d -e DEBUG=true"

# Convenience targets
make docker-run-fg        # Run in foreground
make docker-run-rebuild   # Force rebuild and recreate
```

> **Tip:** The Makefile will use variables from your `.env` file (such as `PORT` and `WHITELISTED_DOMAINS`) if present, so you can easily configure your build and run environment.

> **Note:** The Makefile will automatically configure Docker Compose with the correct application name, private key, and port mapping based on its targets, so you do not need to manually edit `docker-compose.yml` for these values. 