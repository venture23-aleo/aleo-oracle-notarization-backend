# Examples

This folder contains small runnable clients that call the mTLS-protected reverse proxy.

Before running, ensure certificates exist (from repo root):

```bash
cd deployment/secrets
./generate-mtls-certs.sh init --with-default-client
```

The examples assume default paths:

- CA: `deployment/secrets/mtls/ca.crt`
- Client: `deployment/secrets/mtls/clients/AleoOracleClient/client.crt`, `client.key`
- Server URL: `https://localhost:8443/health`

## Go client

See `go-mtls-client`.

Run (from repo root):

```bash
cd examples/go-mtls-client
go run .
```

## TypeScript client

See `ts-mtls-client`.

Run (from repo root):

```bash
cd examples/ts-mtls-client
npm install
npm run start
```

---
If you generated a different client CN, update paths in the example sources accordingly.
