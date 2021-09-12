package errs

import "encoding/json"

type CustomError struct {
	Code int    `json:"Code"`
	Msg  string `json:"msg"`
}

func (e CustomError) Error() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func Code(err error) int {
	if customError, is := err.(*CustomError); is {
		return customError.Code
	}
	return 0
}

func Msg(err error) string {
	if customError, is := err.(CustomError); is {
		return customError.Msg
	}
	return ""
}
