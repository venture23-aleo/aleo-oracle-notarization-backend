package constants

// AllowedHeaders defines which HTTP headers are permitted. Non-whitelisted header values are masked with asterisks for security.
var AllowedHeaders = []string{
	"Accept",
	"Accept-Charset",
	"Accept-Datetime",
	"Accept-Encoding",
	"Accept-Language",
	"Access-Control-Request-Method",
	"Access-Control-Request-Headers",
	"Cache-Control",
	"Connection",
	"Content-Encoding",
	"Content-Length",
	"Content-MD5",
	"Content-Type",
	"Date",
	"Expect",
	"Forwarded",
	"Host",
	"HTTP2-Settings",
	"If-Match",
	"If-Modified-Since",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
	"Max-Forwards",
	"Origin",
	"Pragma",
	"Prefer",
	"Range",
	"Referer",
	"TE",
	"Trailer",
	"Transfer-Encoding",
	"User-Agent",
	"Upgrade",
	"Via",
	"Warning",
	"Upgrade-Insecure-Requests",
	"X-Requested-With",
	"DNT",
	"X-Forwarded-For",
	"X-Forwarded-Host",
	"X-Forwarded-Proto",
	"Front-End-Https",
	"X-Http-Method-Override",
	"X-ATT-DeviceId",
	"X-Wap-Profile",
	"Proxy-Connection",
	"Save-Data",
	"Sec-GPC",
}

// PriceFeedBTCURL, PriceFeedETHURL, and PriceFeedAleoURL are the URLs for the price feeds.
// BTCTokenID, ETHTokenID, and AleoTokenID are the token IDs for the price feeds.
// AttestationDataSizeLimit is the size limit for the string attestation data.
// PriceFeedSelector is the selector for the price feed.
// RequestMethodGET, RequestMethodPOST, ResponseFormatHTML, ResponseFormatJSON, HTMLResultTypeValue, HTMLResultTypeElement, EncodingOptionString, EncodingOptionFloat, and EncodingOptionInt are the constants for the attestation.
const (
	SGXReportType string = "sgx"

	// Price feed Constants
	PriceFeedBTCURL          string = "price_feed: btc"
	PriceFeedETHURL          string = "price_feed: eth"
	PriceFeedAleoURL         string = "price_feed: aleo"
	PriceFeedUSDTURL         string = "price_feed: usdt"
	PriceFeedUSDCURL         string = "price_feed: usdc"
	AleoTokenID              int    = 8
	USDTTokenID              int    = 9
	USDCTokenID              int    = 10
	ETHTokenID               int    = 11
	BTCTokenID               int    = 12
	AttestationDataSizeLimit int    = 1024 * 3
	PriceFeedSelector        string = "weightedAvgPrice"
	MaxAllowedTimeDiff       int64  = 600 // 10 minutes in seconds

	// Attestation Constants
	RequestMethodGET      string = "GET"
	RequestMethodPOST     string = "POST"
	ResponseFormatHTML    string = "html"
	ResponseFormatJSON    string = "json"
	HTMLResultTypeValue   string = "value"
	HTMLResultTypeElement string = "element"
	EncodingOptionString  string = "string"
	EncodingOptionFloat   string = "float"
	EncodingOptionInt     string = "int"

	// Max request and response body sizes
	MaxRequestBodySize  = 10 * 1024   // 10 KB
	MaxResponseBodySize = 1024 * 1024 // 1 MB

	// Oracle Report Constants
	OracleReportChunkSize   = 10
	OracleUserDataChunkSize = 8
	ChunkSizeInBytes   = 512
)
