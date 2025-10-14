package data_extraction

import (
	"encoding/json"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

type BinanceResponse struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume"`
}

type BybitListItem struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume24h"`
}

type BybitResponse struct {
	Result struct {
		List []BybitListItem `json:"list"`
	} `json:"result"`
}

type CoinbaseResponse struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
}

type CryptoListItem struct {
	Price  string `json:"k"`
	Volume string `json:"v"`
}

type CryptoResponse struct {
	Result struct {
		Data []CryptoListItem `json:"data"`
	} `json:"result"`
}

type XTResponseItem struct {
	Price  string `json:"c"`
	Volume string `json:"q"`
}

type XTResponse struct {
	Result []XTResponseItem `json:"result"`
}

type GateResponseItem struct {
	Price  string `json:"last"`
	Volume string `json:"base_volume"`
}

type GateResponse []GateResponseItem

type MEXCResponse struct {
	Price  string `json:"lastPrice"`
	Volume string `json:"volume"`
}


// parseBinanceResponse parses the response from Binance
func parseBinanceResponse(data []byte) (price, volume string, err *appErrors.AppError) {
	var binanceResponse BinanceResponse
	if err := json.Unmarshal(data, &binanceResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", "binance", "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = binanceResponse.Price
	volume = binanceResponse.Volume
	return price, volume, nil
}

// parseBybitResponse parses the response from Bybit
func parseBybitResponse(data []byte) (price, volume string, err *appErrors.AppError) {
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
	return price, volume, nil
}

// parseCoinbaseResponse parses the response from Coinbase
func parseCoinbaseResponse(data []byte) (price, volume string, err *appErrors.AppError) {
	exchange := "coinbase"
	var coinbaseResponse CoinbaseResponse
	if err := json.Unmarshal(data, &coinbaseResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = coinbaseResponse.Price
	volume = coinbaseResponse.Volume
	return price, volume, nil
}

// parseCryptoResponse parses the response from Crypto.com
func parseCryptoResponse(data []byte) (price, volume string, err *appErrors.AppError) {
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
	return price, volume, nil
}

// parseXTResponse parses the response from XT
func parseXTResponse(data []byte) (price, volume string, err *appErrors.AppError) {
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
	return price, volume, nil

}

// parseGateIOResponse parses the response from Gate.io
func parseGateResponse(data []byte) (price, volume string, err *appErrors.AppError) {
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
	return price, volume, nil
}

// parseMEXCResponse parses the response from MEXC
func parseMEXCResponse(data []byte) (price, volume string, err *appErrors.AppError) {
	exchange := "mexc"
	var mexcResponse MEXCResponse
	if err := json.Unmarshal(data, &mexcResponse); err != nil {
		logger.Error("Error unmarshalling data: ", "exchange", exchange, "error", err)
		return "", "", appErrors.ErrDecodingExchangeResponse
	}

	price = mexcResponse.Price
	volume = mexcResponse.Volume
	return price, volume, nil
}

// parseExchangeResponse parses the response from different exchanges
func (c *PriceFeedClient) parseExchangeResponse(exchange string, data []byte) (price, volume string, err *appErrors.AppError) {
	switch exchange {
	case "binance":
		return parseBinanceResponse(data)
	case "bybit":
		return parseBybitResponse(data)
	case "coinbase":
		return parseCoinbaseResponse(data)
	case "crypto":
		return parseCryptoResponse(data)
	case "xt":
		return parseXTResponse(data)
	case "gate":
		return parseGateResponse(data)
	case "mexc":
		return parseMEXCResponse(data)
	default:
		logger.Error("Unsupported exchange: ", "exchange", exchange)
		return "", "", appErrors.ErrExchangeNotSupported
	}
}
