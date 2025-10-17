# mTLS client examples (Go and TypeScript)

This guide shows how to call the Nginx mTLS reverse proxy of Aleo Oracle Notarization Backend from:

- curl (quick sanity check)
- Go (net/http)
- TypeScript/Node.js (https/axios)

Assumptions

- You’ve already run the deployment and generated certs with the helper:
  - CA: `deployment/secrets/mtls/ca.crt`
  - Server: `deployment/secrets/mtls/server.crt` + `server.key`
  - Client: `deployment/secrets/mtls/clients/<CN>/client.crt` + `client.key`
- Nginx is listening on <https://localhost:8443> by default.

If not, initialize mTLS material:

```bash
# From repo root
cd deployment/secrets
./generate-mtls-certs.sh init --with-default-client
```

The default client will be created at:

- `deployment/secrets/mtls/clients/AleoOracleClient/client.crt`
- `deployment/secrets/mtls/clients/AleoOracleClient/client.key`

## Quick check with curl

```bash
curl --cacert deployment/secrets/mtls/ca.crt \
  --cert   deployment/secrets/mtls/clients/AleoOracleClient/client.crt \
  --key    deployment/secrets/mtls/clients/AleoOracleClient/client.key \
  https://localhost:8443/health -vk
```

Expected: HTTP 200 with health JSON. If you get certificate errors, see Troubleshooting.

## Go example (net/http)

Create `mtls_client.go` (example snippet):

```go
package main

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "log"
    "net/http"
)

func main() {
    caCertPath := "deployment/secrets/mtls/ca.crt"
    clientCertPath := "deployment/secrets/mtls/clients/AleoOracleClient/client.crt"
    clientKeyPath := "deployment/secrets/mtls/clients/AleoOracleClient/client.key"

    // Load CA
    caCert, err := ioutil.ReadFile(caCertPath)
    if err != nil { log.Fatal(err) }
    caPool := x509.NewCertPool()
    if !caPool.AppendCertsFromPEM(caCert) { log.Fatal("failed to append CA") }

    // Load client certificate
    cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
    if err != nil { log.Fatal(err) }

    // TLS config
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caPool,
        MinVersion:   tls.VersionTLS12,
    }
    transport := &http.Transport{TLSClientConfig: tlsConfig}
    client := &http.Client{Transport: transport}

    // Request
    resp, err := client.Get("https://localhost:8443/health")
    if err != nil { log.Fatal(err) }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("status=%d body=%s", resp.StatusCode, string(body))
}
```

Run:

```bash
go run mtls_client.go
```

## TypeScript/Node.js (https/axios)

Install deps:

```bash
npm install axios
```

Create `mtls-client.ts`:

```ts
import fs from 'fs';
import https from 'https';
import axios from 'axios';

const ca = fs.readFileSync('deployment/secrets/mtls/ca.crt');
const cert = fs.readFileSync('deployment/secrets/mtls/clients/AleoOracleClient/client.crt');
const key = fs.readFileSync('deployment/secrets/mtls/clients/AleoOracleClient/client.key');

const agent = new https.Agent({
  ca,
  cert,
  key,
  minVersion: 'TLSv1.2',
  // rejectUnauthorized: true, // keep true in production
});

async function main() {
  const res = await axios.get('https://localhost:8443/health', { httpsAgent: agent });
  console.log(res.status, res.data);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
```

Run:

```bash
npx ts-node mtls-client.ts
```

## Using a non-default client CN

If you generated a different client (e.g., `--cn my-app`), update the cert/key paths accordingly in the examples.

## Certificate rotation and revocation

- Renew client:

  ```bash
  cd deployment/secrets
  ./generate-mtls-certs.sh renew-client --cn AleoOracleClient --days 180
  ```

- Revoke client (prevents access if server enforces revocations):

  ```bash
  cd deployment/secrets
  ./generate-mtls-certs.sh revoke-client --cn AleoOracleClient
  ```

If revocation is configured in Nginx, the revoked client will receive HTTP 403.

## Troubleshooting

- x509: certificate signed by unknown authority
  - Ensure you pass the correct CA file (ca.crt). Do not use server.crt as CA.
- tls: bad certificate
  - Ensure client.crt/key are paired and not expired; CN must match a generated client.
- ECONNREFUSED / connection errors
  - Ensure Nginx is running and listening on the expected port (default 8443).
- Self-signed server warning in some tools
  - That’s expected in dev; we validate the server cert via our CA (ca.crt).

## Useful paths (relative to repo root)

- CA: `deployment/secrets/mtls/ca.crt`
- Client: `deployment/secrets/mtls/clients/<CN>/client.crt`, `client.key`
- Server: `deployment/secrets/mtls/server.crt`, `server.key`
- Script: `deployment/secrets/generate-mtls-certs.sh`
