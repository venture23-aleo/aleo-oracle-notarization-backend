package constants

type GraminePathsStruct struct {
	MY_TARGET_INFO_PATH   string
	TARGET_INFO_PATH      string
	USER_REPORT_DATA_PATH string
	REPORT_PATH           string
	ATTESTATION_TYPE_PATH string
	QUOTE_PATH            string
}

var GRAMINE_PATHS = GraminePathsStruct{
	MY_TARGET_INFO_PATH:   "/dev/attestation/my_target_info",
	TARGET_INFO_PATH:      "/dev/attestation/target_info",
	USER_REPORT_DATA_PATH: "/dev/attestation/user_report_data",
	REPORT_PATH:           "/dev/attestation/report",
	ATTESTATION_TYPE_PATH: "/dev/attestation/attestation_type",
	QUOTE_PATH:            "/dev/attestation/quote",
}

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
