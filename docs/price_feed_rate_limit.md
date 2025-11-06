| Exchange   | Public API Rate Limit                                                | Private API Rate Limit                | Rate Limit Scope     | Notes                            |
| ---------- | -------------------------------------------------------------------- | ------------------------------------- | -------------------- | -------------------------------- |
| Binance    | 6,000 weight/minute (weighted)                                       | Weighted, varies by endpoint          | API key weighted     | Market data endpoints low weight |
| Bybit      | \~600 req/5 sec per IP (public)                                      | Separate per API key limits           | Per IP, per UID      | API keys recommended             |
| Coinbase   | 10 req/sec per IP (burst 15)                                         | 15 req/sec per profile (burst 30)     | Per IP / per profile | Clear burst allowance            |
| Crypto.com | 100 req/sec per IP (public)                                          | Varies, e.g., 15 req/100ms per method | Per IP / per API key | High public data limits          |
| XT.com     | 10 req/sec per ip, 1000 req/min per IP | Same as public per user               | Per user, per IP     | Harsh penalty: 10 min lockout    |
| Gate.io    | 200 req/10 sec per endpoint per IP                                   | 10 req/sec per user for private       | Per IP / per UID     | Per-endpoint public limits       |
| MEXC       | 500 req/10 sec per endpoint (IP)                                     | 500 req/10 sec per endpoint (UID)     | Per IP / per UID     | Independent limits               |
| Binance US    | 6,000 weight/minute (weighted)                                       | Weighted, varies by endpoint          | API key weighted     | Market data endpoints low weight |
| Kraken       | Recommends 1 req/sec per IP address                                         | By arrangement                           | Per IP address       | Higher limits via agreement     |
| Gemini     | 120 req/min per IP (public), burst: +5 delayed                     | 600 req/min per API key (private), burst: +5 delayed | Per IP / per API key | Recommends ≤1 req/sec (public), ≤5 req/sec (private) |
| Bitstamp   | 400 req/sec per client (default: 10,000 req / 10 min)                | By arrangement                           | Per client           | Higher limits via agreement     |

## Exchange Rate Limits

### Binance API

Documentation:
- https://developers.binance.com/docs/binance-spot-api-docs/rest-api/limits
- https://developers.binance.com/docs/binance-spot-api-docs/rest-api/market-data-endpoints
- https://developers.binance.com/docs/binance-spot-api-docs/faqs/market_data_only
- https://api.binance.com/api/v3/exchangeInfo

Rate Limit:
- 6,000 request weight per minute 

We can check this header
```bash
x-mbx-uuid	0441b692-39d6-4315-81d2-dc8f194d3d1e
x-mbx-used-weight	2
x-mbx-used-weight-1m	2
```

### Binance US API

Documentation:
- https://docs.binance.us/#general-rest-api-information
- https://docs.binance.us/#rate-limits-rest
- https://api.binance.us/api/v3/exchangeInfo

Rate Limit:
- 6,000 request weight per minute 

We can check this header
```bash
x-mbx-uuid	0441b692-39d6-4315-81d2-dc8f194d3d1e
x-mbx-used-weight	2
x-mbx-used-weight-1m	2
```

### Bybit API

Documentation:
- https://bybit-exchange.github.io/docs/v5/rate-limit
- https://bybit-exchange.github.io/docs/v5/rate-limit

Rate Limit:
- 600 requests per 5 seconds per IP address

### Coinbase

Documentation:
- https://docs.cloud.coinbase.com/exchange/rest-api/rate-limits

Rate Limit:
- Public Endpoints
  - Requests per second per IP: 10
  - Requests per second per IP in bursts: Up to 15

- Private Endpoints
  - Private endpoints are authenticated.
  - Requests per second per profile: 15
  - Requests per second per profile in bursts: Up to 30


### Crypto.com

Documentation:
- https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#rest-api-root-endpoint
- https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#rate-limits

Rate Limit
- 100 requests per second per IP address

### XT.com

Documentation:
- https://doc.xt.com/docs/spot/Market/Get24hStatisticsTicker
- https://futuresee.github.io/github.io/#documentationrestApi
- https://futuresee.github.io/github.io/#documentationlimitRules

Rate Limit: 
  - 10 requests per second for each single user (API key)
  - 1000 requests per minute for each IP address

### Gate.io

Documentation:
- https://www.gate.com/announcements/article/33995

Rate Limit
- 200 requests per 10 seconds per endpoint per IP address

### MEXC

Documentation:
- https://mexcdevelop.github.io/apidocs/spot_v3_en/#limits

Rate Limit: 
  - Each endpoint with IP/UID limits has an independent 500 every 10 second limit.

### Kraken API

Documentation:
- https://docs.kraken.com/rest/#rate-limits
- https://docs.kraken.com/api/docs/guides/spot-rest-ratelimits/

Rate Limit:
- 1 req per second per IP address

### Bitstamp API

Documentation:
- https://www.bitstamp.net/api/#section/What-is-API
- https://www.bitstamp.net/api/#tag/Tickers/operation/GetMarketTicker
- https://www.bitstamp.net/api/#rate-limits

Rate Limit:
- 400 requests per second per client (default: 10,000 requests per 10 minutes)

### Gemini API

Documentation:
- https://docs.gemini.com/get-started/intro
- https://docs.gemini.com/rate-limit

Rate Limit:
- 1 req per second per IP address
- 120 requests per minute per IP address