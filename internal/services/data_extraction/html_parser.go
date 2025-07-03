package data_extraction

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
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
//   request := services.AttestationRequest{
//       Url: "https://example.com/price-page",
//       Selector: "//span[@class='price']/text()",
//       ResponseFormat: "html",
//       HTMLResultType: &[]string{"value"}[0],
//       EncodingOptions: encoding.EncodingOptions{Value: "float", Precision: 2}
//   }
//   result, err := ExtractDataFromHTML(request)
func ExtractDataFromHTML(attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {
	// Make the HTTP request
	resp, err := makeHTTPRequest(attestationRequest)
	if err != nil {
		log.Print("[ERROR] [ExtractDataFromHTML] Error making HTTP request:", err)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, err
	}
	defer resp.Body.Close()

	// Read the full HTML content.
	htmlBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Print("[ERROR] [ExtractDataFromHTML] Error reading HTML content:", readErr)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrReadingHTMLContent)
	}

	// Convert the HTML bytes to a string.
	htmlContent := string(htmlBytes)

	// Parse the HTML content.
	htmlDoc, parseErr := htmlquery.Parse(strings.NewReader(htmlContent))
	if parseErr != nil {
		log.Print("[ERROR] [ExtractDataFromHTML] Error parsing HTML content:", parseErr)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrParsingHTMLContent)
	}

	// Query the HTML content using XPath selector.
	result, queryErr := htmlquery.Query(htmlDoc, attestationRequest.Selector)

	// Check if the error is not nil or the result is nil.
	if queryErr != nil || result == nil {
		log.Print("[ERROR] [ExtractDataFromHTML] Error querying HTML content:", queryErr)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrSelectorNotFound)
	}

	var valueStr string

	// Handle different HTML result types
	if attestationRequest.HTMLResultType != nil && *attestationRequest.HTMLResultType == "element" {
		// For "element" type, return the entire HTML element (including tags)
		valueStr = htmlquery.OutputHTML(result, true)
	} else {
		// For "value" type (default), return just the text content
		if result.FirstChild != nil {
			valueStr = result.FirstChild.Data
		} else {
			// If no text child, get the text content of the element itself
			valueStr = htmlquery.InnerText(result)
		}
	}

	// Apply float precision if needed (only for "value" type)
	if attestationRequest.HTMLResultType != nil && *attestationRequest.HTMLResultType == "value" && 
		attestationRequest.EncodingOptions.Value == "float" {
		_, floatErr := strconv.ParseFloat(valueStr, 64)
		if floatErr != nil {
			log.Print("[ERROR] [ExtractDataFromHTML] Error parsing float value:", floatErr)
			return ExtractDataResult{
				StatusCode: resp.StatusCode,
			}, appErrors.NewAppError(appErrors.ErrParsingHTMLContent)
		}
		valueStr = applyFloatPrecision(valueStr, attestationRequest.EncodingOptions.Precision)
	}

	// Return the HTML content, data, status code, and error.
	return ExtractDataResult{
		ResponseBody:    htmlContent,
		AttestationData: valueStr,
		StatusCode:      resp.StatusCode,
	}, nil
}
