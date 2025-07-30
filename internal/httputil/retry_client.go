package http

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

func GetRetryableHTTPClient(maxRetries int) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	retryClient.Logger = logger.Logger
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 3 * time.Second
	retryClient.RetryMax = maxRetries
	return retryClient
}
