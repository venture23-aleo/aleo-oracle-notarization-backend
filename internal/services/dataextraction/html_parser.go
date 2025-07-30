package data_extraction

import (
	"context"
	"io"

	"github.com/antchfx/htmlquery"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

// Package data_extraction provides data extraction capabilities for the Aleo Oracle Notarization Backend.
// This file contains HTML-specific data extraction functionality using XPath selectors.

// ExtractDataFromHTML scrapes and extracts data from HTML responses for attestation purposes.
//
// This function:
// 1. Makes an HTTP request to the specified URL
// 2. Reads and parses the HTML content using htmlquery
// 3. Extracts data using XPath selectors
// 4. Handles different result types (element vs value)
// 5. Applies encoding options for numeric values
// 6. Returns the full HTML content and extracted data
//
// The selector uses XPath syntax for HTML element selection:
// - "//div[@class='price']" - finds div with class 'price'
// - "//span[@id='total']/text()" - gets text content of span with id 'total'
// - "//table//tr[1]/td[2]" - gets second cell of first table row
//
// HTMLResultType controls what is extracted:
// - "value" (default): extracts text content only
// - "element": extracts the complete HTML element including tags
//
// Example usage:
//
//	request := services.AttestationRequest{
//	    Url: "https://example.com/price-page",
//	    Selector: "//span[@class='price']/text()",
//	    ResponseFormat: "html",
//	    HTMLResultType: &[]string{"value"}[0],
//	    EncodingOptions: encoding.EncodingOptions{Value: "float", Precision: 2}
//	}
//	result, err := ExtractDataFromHTML(request)
func ExtractDataFromHTML(ctx context.Context, attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {

	// Make the HTTP request
	reqLogger := logger.FromContext(ctx)
	resp, err := makeHTTPRequestToTarget(ctx, attestationRequest)
	if err != nil {
		reqLogger.Error("Error making HTTP request: ", "error", err)
		return ExtractDataResult{}, err
	}
	defer resp.Body.Close()

	limitReader := io.LimitReader(resp.Body, constants.MaxResponseBodySize)

	// Parse the HTML content.
	htmlDoc, parseErr := htmlquery.Parse(limitReader)
	if parseErr != nil {
		reqLogger.Error("Error parsing HTML content: ", "error", parseErr)
		return ExtractDataResult{}, appErrors.ErrParsingHTMLContent
	}

	// Query the HTML content using XPath selector.
	result, queryErr := htmlquery.Query(htmlDoc, attestationRequest.Selector)

	// Check if the error is not nil or the result is nil.
	if queryErr != nil || result == nil {
		reqLogger.Error("Error querying HTML content: ", "error", queryErr)
		return ExtractDataResult{}, appErrors.ErrSelectorNotFound
	}

	var valueStr string

	// Handle different HTML result types
	if *attestationRequest.HTMLResultType == constants.HTMLResultTypeElement {
		// For "element" type, return the entire HTML element (including tags)
		valueStr = htmlquery.OutputHTML(result, true)
	} else {
		// For "value" type (default), return just the text content
		valueStr = htmlquery.InnerText(result)
	}

	if valueStr == "" {
		return ExtractDataResult{}, appErrors.ErrEmptyAttestationData
	}

	formattedAttestationData, appErr := formatAttestationData(ctx, valueStr, attestationRequest.EncodingOptions.Value, attestationRequest.EncodingOptions.Precision)
	if appErr != nil {
		return ExtractDataResult{}, appErr
	}

	// Return the HTML content, data, status code, and error.
	return ExtractDataResult{
		ResponseBody:    htmlquery.OutputHTML(htmlDoc, true),
		AttestationData: formattedAttestationData,
		StatusCode:      resp.StatusCode,
	}, nil
}
