import axios from 'axios';
import fs from 'fs';
import https from 'https';

function env(key: string, def: string): string {
  return process.env[key] && process.env[key]!.length > 0 ? process.env[key]! : def;
}

const caPath = env('CA_CERT', '../../deployment/secrets/mtls/ca.crt');
const certPath = env('CLIENT_CERT', '../../deployment/secrets/mtls/clients/AleoOracleClient/client.crt');
const keyPath = env('CLIENT_KEY', '../../deployment/secrets/mtls/clients/AleoOracleClient/client.key');
const url = env('SERVER_URL', 'https://localhost:8443/health');

const ca = fs.readFileSync(caPath);
const cert = fs.readFileSync(certPath);
const key = fs.readFileSync(keyPath);

const agent = new https.Agent({
  ca,
  cert,
  key,
  minVersion: 'TLSv1.2'
});

async function main() {
  const res = await axios.get(url, { httpsAgent: agent });
  console.log(res.status, res.data);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
