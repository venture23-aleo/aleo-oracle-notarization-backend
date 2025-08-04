# Price Feed Integration with Attestation System

## Overview

The Aleo Oracle Notarization Backend now supports price feed attestation requests that fetch volume-weighted average prices from multiple cryptocurrency exchanges. This integration allows users to get attested price data for BTC, ETH, and ALEO tokens.

## Supported Price Feeds

- **BTC**: `price_feed: btc`
- **ETH**: `price_feed: eth` 
- **ALEO**: `price_feed: aleo`

## Attestation Request Format

```json
{
  "url": "price_feed: btc",
  "requestMethod": "GET",
  "selector": "weightedAvgPrice",
  "responseFormat": "json",
  "encodingOptions": {
    "value": "float",
    "precision": 6
  }
}
```

### Parameters

- **url**: The price feed URL (e.g., `price_feed: btc`)
- **requestMethod**: Always `GET` for price feeds
- **selector**: Field to extract from the response (e.g., `weightedAvgPrice`)
- **responseFormat**: Always `json` for price feeds
- **encodingOptions.value**: Always `float` for price data
- **encodingOptions.precision**: Number of decimal places (1-12, max 12)

## Response Format

The system returns a JSON response with the following structure:

```json
{
  token: 'BTC',
  volumeWeightedAvg: '114655.40268302493',
  totalVolume: '21301.003131079997',
  exchangeCount: 4,
  timestamp: 1754294883,
  exchangePrices: [
    {
      exchange: 'Binance',
      price: 114641.51,
      volume: 7306.44218,
      token: 'BTC',
      symbol: 'BTCUSDT'
    },
    {
      exchange: 'Binance',
      price: 114666.75,
      volume: 1748.21702,
      token: 'BTC',
      symbol: 'BTCUSDC'
    },
    {
      exchange: 'Coinbase',
      price: 114659.93,
      volume: 2282.30629082,
      token: 'BTC',
      symbol: 'BTC-USD'
    },
    {
      exchange: 'Coinbase',
      price: 114637.18,
      volume: 40.83332326,
      token: 'BTC',
      symbol: 'BTC-USDT'
    },
    {
      exchange: 'Crypto',
      price: 114644,
      volume: 2455.9773,
      token: 'BTC',
      symbol: 'BTC_USDT'
    },
    {
      exchange: 'Bybit',
      price: 114669.3,
      volume: 3393.085573,
      token: 'BTC',
      symbol: 'BTCUSDT'
    },
    {
      exchange: 'Bybit',
      price: 114671.6,
      volume: 446.270844,
      token: 'BTC',
      symbol: 'BTCUSDC'
    },
    {
      exchange: 'Crypto',
      price: 114668,
      volume: 3627.8706,
      token: 'BTC',
      symbol: 'BTC_USD'
    }
  ],
  success: true
}
```

## Precision Handling

The system respects the `precision` parameter in the `encodingOptions`:

- **Input**: `precision: 6`
- **Output**: Price truncated to 6 decimal places
- **Example**: `43250.123456789` becomes `43250.123456`

The maximum allowed precision is 12 decimal places (defined by `ENCODING_OPTION_FLOAT_MAX_PRECISION`).

## Available Selectors

- `weightedAvgPrice`: Volume-weighted average price (recommended)
- `volumeWeightedAvg`: Same as weightedAvgPrice
- `totalVolume`: Total trading volume across all exchanges
- `exchangeCount`: Number of exchanges that provided data
- `timestamp`: Unix timestamp of when the data was fetched
- `symbol`: The cryptocurrency symbol

## Exchange Integration

The system aggregates price data from multiple exchanges:

1. **Binance**: `https://api.binance.com/api/v3/ticker/24hr?symbol={symbol}USDT`
2. **Bybit**: `https://api.bybit.com/v5/market/tickers?category=spot&symbol={symbol}USDT`
3. **Coinbase**: `https://api.exchange.coinbase.com/products/{symbol}-USD/ticker`
4. **Crypto.com**: `https://api.crypto.com/v2/public/get-ticker?instrument_name={symbol}_USD`
5. **Gate.io**: `https://api.gateio.ws/api/v4/spot/tickers?currency_pair={symbol}_USDT`
6. **MEXC**: `https://api.mexc.com/api/v3/ticker/24hr?symbol={symbol}USDT`
7. **XT.com**: `https://xt.com/sapi/v4/market/public/ticker/24h?symbol={symbol}_USDT`

## Volume-Weighted Average Calculation

The system calculates the volume-weighted average price using the formula:

```
VWAP = Σ(price × volume) / Σ(volume)
```

This ensures that exchanges with higher trading volumes have more influence on the final price.

## Error Handling

- **Insufficient Data**: Requires at least 2 exchanges to respond
- **API Failures**: Individual exchange failures are logged but don't stop the process
- **Invalid Responses**: Malformed data from exchanges is skipped
- **Network Issues**: Timeouts and connection errors are handled gracefully

## Security

- Price data is fetched within the SGX enclave
- Attestation reports are cryptographically signed
- No sensitive data is exposed outside the enclave

## Usage Example

```go
attestationRequest := AttestationRequest{
    Url:            "price_feed: btc",
    RequestMethod:  "GET",
    Selector:       "weightedAvgPrice",
    ResponseFormat: "json",
    EncodingOptions: encoding.EncodingOptions{
        Value:     "float",
        Precision: 6,
    },
}

extractedData, err := ExtractPriceFeedData(attestationRequest)
```

## Integration with Existing Attestation Flow

The price feed functionality is seamlessly integrated into the existing attestation system:

1. **Request Validation**: Uses the same validation logic as other attestation requests
2. **Data Fetching**: Handled by `ExtractPriceFeedData` when a price feed URL is detected
3. **Precision Processing**: Uses the same precision handling as other float values
4. **Attestation Generation**: Generates standard attestation reports with price data
5. **Oracle Data**: Creates signed attestation reports for blockchain verification

This integration allows the Aleo Oracle to provide trusted, volume-weighted cryptocurrency prices that can be used in smart contracts and DeFi applications. 