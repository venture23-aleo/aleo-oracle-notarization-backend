
# ========================
# Stage 1: Build Go Binary
# ========================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go modules files first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

ARG VERSION
ARG COMMIT
ARG APP

# Copy source code
COPY . .

# Build the statically linked Go binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -buildvcs=false -o ${APP} ./cmd/server

# =============================
# Stage 2: Gramine SGX Runtime
# =============================
FROM gramineproject/gramine:stable-jammy

ARG APP

# Set workdir
WORKDIR /app

COPY --from=builder /app/${APP} .

RUN DEBIAN_FRONTEND=noninteractive apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates libssl-dev libprotobuf-dev protobuf-compiler python3 python3-pip libsgx-dcap-default-qpl && rm -rf /var/lib/apt/lists/*

# Copy configuration files

COPY build/inputs/sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf

COPY build/inputs/entrypoint.sh build/inputs/${APP}.manifest.template ./

RUN chmod +x entrypoint.sh

RUN gramine-sgx-gen-private-key && \
    gramine-manifest -Dlog_level=debug ${APP}.manifest.template ${APP}.manifest && \
    gramine-sgx-sign --manifest ${APP}.manifest --output ${APP}.manifest.sgx

ENV APP=${APP}

ENTRYPOINT ["./entrypoint.sh"]

