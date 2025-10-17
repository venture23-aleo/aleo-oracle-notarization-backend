# syntax=docker/dockerfile:1.4
# ========================
# Stage 1: Build Go Binary
# ========================
FROM golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

WORKDIR /app

# Copy Go modules files first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

ARG APP

# Copy source code
COPY . .


# Build the statically linked Go binary
RUN CGO_ENABLED=0 go build -gcflags=all=-spectre=all -asmflags=all=-spectre=all -trimpath -ldflags="-s -w" -buildvcs=false -o ${APP} ./cmd/server

# =============================
# Stage 2: Gramine SGX Runtime
# =============================
FROM gramineproject/gramine:stable-jammy@sha256:84b3d222e0bd9ab941f0078a462af0dbc5518156b99b147c10a7b83722ac0c38

ARG APP

# Set workdir
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/${APP} .

# Copy the manifest template, static files, and the entrypoint script
COPY deployment/docker/inputs/ ./

# Install SGX DCAP QPL
RUN DEBIAN_FRONTEND=noninteractive dpkg -i /app/libsgx-dcap-default-qpl_1.23.100.0-jammy1_amd64.deb

# Copy the SGX default QCNL configuration file
COPY deployment/docker/inputs/sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf

# Make the entrypoint script executable
RUN chmod +x entrypoint.sh

RUN --mount=type=secret,id=gramine-private-key,target=/tmp/private-key.pem \
    gramine-manifest ${APP}.manifest.template ${APP}.manifest && \
    gramine-sgx-sign --manifest ${APP}.manifest --key /tmp/private-key.pem --output ${APP}.manifest.sgx && \
    gramine-sgx-sigstruct-view ${APP}.sig -v --output-format json > ${APP}.metadata.json

ENV APP=${APP}
ENTRYPOINT ["./entrypoint.sh"]