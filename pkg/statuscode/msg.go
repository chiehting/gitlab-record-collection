package statuscode

// MsgFlags is
var MsgFlags = map[int]string{
	Success:  "Success",
	Error:    "Failed",
	AuthFail: "Authorization failed",
	AuthMiss: "Authorization header is missing",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[Error]
}
