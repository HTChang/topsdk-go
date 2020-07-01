package top

import "fmt"

// ErrorResponse defines the open taobao platform error response.
type ErrorResponse struct {
	Code    int32  `json:"code"`
	Msg     string `json:"msg"`
	SubMsg  string `json:"sub_msg"`
	SubCode string `json:"sub_code"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s, %s: %s", e.Code, e.Msg, e.SubCode, e.SubMsg)
}
