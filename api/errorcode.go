package api

import "github.com/bitmark-inc/autonomy-api/store"

var (
	errorMessageMap = map[int64]string{
		999:  "internal server error",
		1000: "invalid signature",
		1001: "invalid authorization format",
		1002: "authorization expired",
		1003: "invalid token",
		1004: "invalid parameters",
		1005: "cannot parse request",
		1006: "invalid value of client version",
		1007: "API for this client version has been discontinued",

		1100: "his account has been registered or has been taken",
		1101: "account not found",
		1102: "the account is under deletion",

		1200: store.ErrRequestNotExist.Error(),
	}

	errorInternalServer             = errorJSON(999)
	errorInvalidSignature           = errorJSON(1000)
	errorInvalidAuthorizationFormat = errorJSON(1001)
	errorAuthorizationExpired       = errorJSON(1002)
	errorInvalidToken               = errorJSON(1003)

	errorInvalidParameters        = errorJSON(1004)
	errorCannotParseRequest       = errorJSON(1005)
	errorInvalidClientVersion     = errorJSON(1006)
	errorUnsupportedClientVersion = errorJSON(1007)

	errorAccountTaken    = errorJSON(1100)
	errorAccountNotFound = errorJSON(1101)
	errorAccountDeleting = errorJSON(1102)

	errorRequestNotExist = errorJSON(1200)
)

type ErrorResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// errorJSON converts an error code to a standardized error object
func errorJSON(code int64) ErrorResponse {
	var message string
	if msg, ok := errorMessageMap[code]; ok {
		message = msg
	} else {
		message = "unknown"
	}

	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}
