package dto

import "pr-manage-service/pkg/codes"

type ErrorResponseBody struct {
	Code codes.CODE `json:"code"`
	Msg  string     `json:"message"`
}

type ErrorResponse struct {
	Err ErrorResponseBody `json:"error"`
}
