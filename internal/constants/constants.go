package constants

// GraminePathsStruct is the structure for the Gramine paths.
type GraminePathsStruct struct {
	MY_TARGET_INFO_PATH   string
	TARGET_INFO_PATH      string
	USER_REPORT_DATA_PATH string
	REPORT_PATH           string
	ATTESTATION_TYPE_PATH string
	QUOTE_PATH            string
}

// GRAMINE_PATHS is the structure for the Gramine paths.
var GRAMINE_PATHS = GraminePathsStruct{
	MY_TARGET_INFO_PATH:   "/dev/attestation/my_target_info",
	TARGET_INFO_PATH:      "/dev/attestation/target_info",
	USER_REPORT_DATA_PATH: "/dev/attestation/user_report_data",
	REPORT_PATH:           "/dev/attestation/report",
	ATTESTATION_TYPE_PATH: "/dev/attestation/attestation_type",
	QUOTE_PATH:            "/dev/attestation/quote",
}

// ALLOWED_HEADERS defines which HTTP headers are permitted. Non-whitelisted header values are masked with asterisks for security.
var ALLOWED_HEADERS = []string{
	"Accept",
	"Accept-Charset",
	"Accept-Datetime",
	"Accept-Encoding",
	"Accept-Language",
	"Access-Control-Request-Method",
	"Access-Control-Request-Header",
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

// PRICE_FEED_BTC_URL, PRICE_FEED_ETH_URL, and PRICE_FEED_ALEO_URL are the URLs for the price feeds.
// BTC_TOKEN_ID, ETH_TOKEN_ID, and ALEO_TOKEN_ID are the token IDs for the price feeds.
// ATTESTATION_DATA_SIZE_LIMIT is the size limit for the string attestation data.
const (
	PRICE_FEED_BTC_URL  = "price_feed: btc"
	PRICE_FEED_ETH_URL  = "price_feed: eth"
	PRICE_FEED_ALEO_URL = "price_feed: aleo"
	BTC_TOKEN_ID        = 12
	ETH_TOKEN_ID        = 11
	ALEO_TOKEN_ID       = 8
	// ATTESTATION_DATA_SIZE_LIMIT is the size limit for the attestation data.
	ATTESTATION_DATA_SIZE_LIMIT = 1024 * 3
)
