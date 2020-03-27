package api

var (
	errorMessageMap = map[int64]string{
		999:  "internal server error",
		1000: "invalid signature",
		1001: "invalid authorization format",
		1002: "authorization expired",
		1003: "invalid token",
		1004: "invalid parameters",
		1005: "cannot parse request",
		1006: "API for this client version has been discontinued",

		1100: "his account has been registered or has been taken",
		1101: "account not found",
		1102: "the account is under deletion",
	}

	errorInternalServer             = errorJSON(999)
	errorInvalidSignature           = errorJSON(1000)
	errorInvalidAuthorizationFormat = errorJSON(1001)
	errorAuthorizationExpired       = errorJSON(1002)
	errorInvalidToken               = errorJSON(1003)

	errorInvalidParameters        = errorJSON(1004)
	errorCannotParseRequest       = errorJSON(1005)
	errorUnsupportedClientVersion = errorJSON(1006)

	errorAccountTaken    = errorJSON(1100)
	errorAccountNotFound = errorJSON(1101)
	errorAccountDeleting = errorJSON(1102)
)

type ErrorResponse struct {
	Code    int64
	Message string
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
