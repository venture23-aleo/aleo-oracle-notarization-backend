# syntax=docker/dockerfile:1.4
# ========================
# Stage 1: Build Go Binary
# ========================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go modules files first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

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

RUN DEBIAN_FRONTEND=noninteractive apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libsgx-dcap-default-qpl && rm -rf /var/lib/apt/lists/*

# Copy configuration files
COPY build/inputs/sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf
COPY build/inputs/static_hosts build/inputs/static_resolv.conf build/inputs/entrypoint.sh build/inputs/${APP}.manifest.template ./

RUN chmod +x entrypoint.sh

RUN --mount=type=secret,id=gramine-private-key,target=/tmp/private-key.pem \
    gramine-manifest -Dlog_level=debug ${APP}.manifest.template ${APP}.manifest && \
    gramine-sgx-sign --manifest ${APP}.manifest --key /tmp/private-key.pem --date 0000-00-00 --output ${APP}.manifest.sgx

ENV APP=${APP}

ENTRYPOINT ["./entrypoint.sh"]

