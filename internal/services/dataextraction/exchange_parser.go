package data_extraction

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

type BinanceResponse struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume"`
	Symbol string `json:"symbol"`
	Timestamp int64 `json:"closeTime"`
}

type BybitListItem struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume24h"`
	Symbol string `json:"symbol"`
}

type BybitResponse struct {
	Timestamp int64 `json:"time"`
	Result struct {
		List []BybitListItem `json:"list"`
	} `json:"result"`
}

type CoinbaseResponse struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Timestamp string `json:"time"`
}

type CryptoListItem struct {
	Price  string `json:"k"`
	Volume string `json:"v"`
	Symbol string `json:"i"`
	Timestamp int64 `json:"t"`
}

type CryptoResponse struct {
	Result struct {
		Data []CryptoListItem `json:"data"`
	} `json:"result"`
}

type XTResponseItem struct {
	Price  string `json:"c"`
	Volume string `json:"q"`
	Timestamp int64 `json:"t"`
	Symbol string `json:"s"`
}

type XTResponse struct {
	Result []XTResponseItem `json:"result"`
}

type GateResponseItem struct {
	Price  string `json:"last"`
	Volume string `json:"base_volume"`
	Symbol string `json:"currency_pair"`
}

type GateResponse []GateResponseItem

type MEXCResponse struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume"`
	Symbol string `json:"symbol"`
	Timestamp int64 `json:"closeTime"`
}

type KrakenResponseItem struct {
	Price  [2]string `json:"c"`
	Volume [2]string `json:"v"`
}

type KrakenResponse struct {
	Error  []string `json:"error"`
	Result map[string]KrakenResponseItem `json:"result"`
}

type GeminiVolumeInfo struct {
	USDT      string `json:"USDT,omitempty"`
	USDC      string `json:"USDC,omitempty"`
	USD       string `json:"USD,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type GeminiResponse struct {
	Price  string           `json:"last"`
	VolumeInfo GeminiVolumeInfo `json:"volume"`
}

type BitstampResponse struct {
	Price  string `json:"last"`
	Volume string `json:"volume"`
	Timestamp string `json:"timestamp"`
}

func validateTimestamp(exchange string, timestamp int64, attestationTimestamp int64) *appErrors.AppError {
	timestampInUnix := timestamp / 1000
	timeDiff := timestampInUnix - attestationTimestamp
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	if timeDiff > constants.MaxAllowedTimeDiff {
		logger.Error("Timestamp difference too large: ", "exchange", exchange, "expected", attestationTimestamp, "got", timestamp, "diff_seconds", timeDiff)
		return appErrors.ErrTimestampTooOld
	}
	return nil
}


func validateSymbol(exchange, parsedSymbol, symbol string) *appErrors.AppError {
	if parsedSymbol == "" {
		logger.Error("Symbol is empty: ", "exchange", exchange)
		return appErrors.ErrSymbolMismatch
	}

	if !strings.EqualFold(parsedSymbol, symbol) {
		logger.Error("Symbol mismatch: ", "exchange", exchange, "expected", symbol, "got", parsedSymbol)
		return appErrors.ErrSymbolMismatch
	}
	return nil
}


func validatePrice(price string) *appErrors.AppError {
	priceRat, ok := new(big.Rat).SetString(price)
	if !ok || priceRat.Sign() <= 0 {
		logger.Error("Error parsing price: ", "exchange", "binance", "price", price)
		return appErrors.ErrParsingPrice
	}
	return nil
}

func validateVolume(volume string) *appErrors.AppError {
	volumeRat, ok := new(big.Rat).SetString(volume)
	if !ok || volumeRat.Sign() <= 0 {
		logger.Error("Error parsing volume: ", "exchange", "binance", "volume", volume)
		return appErrors.ErrParsingVolume
	}
	return nil
}

func validatePriceAndVolume(price, volume string) *appErrors.AppError {
	err := validatePrice(price)
	if err != nil {
		return err
	}
	err = validateVolume(volume)
	if err != nil {
		return err
	}
	return nil
}

// parseBinanceResponse parses the response from Binance
func parseBinanceResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	var binanceResponse BinanceResponse
	if err := json.Unmarshal(data, &binanceResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", "binance", "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = binanceResponse.Price
	volume = binanceResponse.Volume

	err = validateSymbol("binance", binanceResponse.Symbol, symbol)
	if err != nil {
		return "", "", err
	}

	err = validateTimestamp("binance", binanceResponse.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

// parseBybitResponse parses the response from Bybit
func parseBybitResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	var bybitResponse BybitResponse
	if err := json.Unmarshal(data, &bybitResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", "bybit", "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}
	
	list := bybitResponse.Result.List
	if len(list) == 0 {
		logger.Error("No data in response", "exchange", "bybit")
		return "", "", appErrors.ErrMissingDataInResponse
	}

	item := list[0]
	
	price = item.Price
	volume = item.Volume

	err = validateSymbol("bybit", item.Symbol, symbol)
	if err != nil {
		return "", "", err
	}
	
	err = validateTimestamp("bybit", bybitResponse.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}


	return price, volume, nil
}

// parseCoinbaseResponse parses the response from Coinbase
func parseCoinbaseResponse(data []byte, _ string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	exchange := "coinbase"
	var coinbaseResponse CoinbaseResponse
	if err := json.Unmarshal(data, &coinbaseResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = coinbaseResponse.Price
	volume = coinbaseResponse.Volume

	t, parseErr := time.Parse(time.RFC3339Nano, coinbaseResponse.Timestamp)
	if parseErr != nil {
		logger.Error("Error parsing timestamp: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrParsingTimestamp
	}

	err = validateTimestamp("coinbase", t.UnixMilli(), timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

// parseCryptoResponse parses the response from Crypto.com
func parseCryptoResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	exchange := "crypto"
	var cryptoResponse CryptoResponse
	if err := json.Unmarshal(data, &cryptoResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	dataArray := cryptoResponse.Result.Data
	if len(dataArray) == 0 {
		logger.Error("No data in response", "exchange", exchange)
		return "", "", appErrors.ErrMissingDataInResponse
	}

	item := dataArray[0]
	price = item.Price
	volume = item.Volume

	err = validateSymbol("crypto", item.Symbol, symbol)
	if err != nil {
		return "", "", err
	}

	err = validateTimestamp("crypto", item.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

// parseXTResponse parses the response from XT
func parseXTResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	exchange := "xt"
	var xtResponse XTResponse
	if err := json.Unmarshal(data, &xtResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	result := xtResponse.Result
	if len(result) == 0 {
		logger.Error("No data in response", "exchange", exchange)
		return "", "", appErrors.ErrMissingDataInResponse
	}

	item := result[0]

	price = item.Price
	volume = item.Volume

	err = validateSymbol("xt", item.Symbol, symbol)
	if err != nil {
		return "", "", err
	}

	err = validateTimestamp("xt", item.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil

}

// parseGateIOResponse parses the response from Gate.io
func parseGateResponse(data []byte, symbol string, _ int64) (price, volume string, err *appErrors.AppError) {
	exchange := "gate"
	var gateResponse GateResponse
	if err := json.Unmarshal(data, &gateResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	list := gateResponse
	if len(list) == 0 {
		logger.Error("No data in response", "exchange", exchange)
		return "", "", appErrors.ErrMissingDataInResponse
	}

	item := list[0]

	price = item.Price
	volume = item.Volume

	err = validateSymbol("gate", item.Symbol, symbol)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

// parseMEXCResponse parses the response from MEXC
func parseMEXCResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	exchange := "mexc"
	var mexcResponse MEXCResponse
	if err := json.Unmarshal(data, &mexcResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = mexcResponse.Price
	volume = mexcResponse.Volume

	err = validateSymbol("mexc", mexcResponse.Symbol, symbol)
	if err != nil {
		return "", "", err
	}

	err = validateTimestamp("mexc", mexcResponse.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

func parseKrakenResponse(data []byte, symbol string) (price, volume string, err *appErrors.AppError) {
	exchange := "kraken"
	var krakenResponse KrakenResponse
	if err := json.Unmarshal(data, &krakenResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	result, ok := krakenResponse.Result[symbol]
	if !ok {
		logger.Error("No data in response", "exchange", exchange, "symbol", symbol)
		return "", "", appErrors.ErrMissingDataInResponse
	}

	price = result.Price[0]
	volume = result.Volume[0]

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

func parseGeminiResponse(data []byte, symbol string, timestamp int64, token string) (price, volume string, err *appErrors.AppError) {
	exchange := "gemini"
	var geminiResponse GeminiResponse
	if err := json.Unmarshal(data, &geminiResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "symbol", symbol, "token", token, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = geminiResponse.Price

	switch token {
	case "USDT":
		volume = geminiResponse.VolumeInfo.USDT
	case "USDC":
		volume = geminiResponse.VolumeInfo.USDC
	}

	if volume == "" {
		logger.Error("No volume in response", "exchange", exchange, "symbol", symbol, "token", token)
		return "", "", appErrors.ErrMissingDataInResponse
	}

	err = validateTimestamp(exchange,geminiResponse.VolumeInfo.Timestamp, timestamp)
	if err != nil {
		return "", "", err
	}
	
	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}

	return price, volume, nil
}

func parseBitstampResponse(data []byte, symbol string, timestamp int64) (price, volume string, err *appErrors.AppError) {
	exchange := "bitstamp"
	var bitstampResponse BitstampResponse
	if err := json.Unmarshal(data, &bitstampResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "symbol", symbol, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = bitstampResponse.Price
	volume = bitstampResponse.Volume

	parsedTimestamp, parseErr := strconv.ParseInt(bitstampResponse.Timestamp, 10, 64)
	if parseErr != nil {
		logger.Error("Error parsing timestamp: ", "exchange", exchange, "symbol", symbol, "error", parseErr)
		return "", "", appErrors.ErrParsingTimestamp
	}

	err = validateTimestamp(exchange, parsedTimestamp * 1000, timestamp)
	if err != nil {
		return "", "", err
	}

	err = validatePriceAndVolume(price, volume)
	if err != nil {
		return "", "", err
	}
	
	return price, volume, nil
}

// parseExchangeResponse parses the response from different exchanges
func (c *PriceFeedClient) parseExchangeResponse(exchange string, data []byte, symbol string, timestamp int64, token string) (price, volume string, err *appErrors.AppError) {
	switch exchange {
	case "binance", "binance-us":
		return parseBinanceResponse(data, symbol, timestamp)
	case "bybit":
		return parseBybitResponse(data, symbol, timestamp)
	case "coinbase":
		return parseCoinbaseResponse(data, symbol, timestamp)
	case "crypto":
		return parseCryptoResponse(data, symbol, timestamp)
	case "xt":
		return parseXTResponse(data, symbol, timestamp)
	case "gate":
		return parseGateResponse(data, symbol, timestamp)
	case "mexc":
		return parseMEXCResponse(data, symbol, timestamp)
	case "kraken":
		return parseKrakenResponse(data, symbol)
	case "gemini":
		return parseGeminiResponse(data, symbol, timestamp, token)
	case "bitstamp":
		return parseBitstampResponse(data, symbol, timestamp)
	default:
		logger.Error("Unsupported exchange: ", "exchange", exchange)
		return "", "", appErrors.ErrExchangeNotSupported
	}
}
