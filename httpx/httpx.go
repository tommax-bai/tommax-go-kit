// Package httpx 提供统一响应包与标准中间件链（docs/03 §1.2、docs/04 §1.2）。
// 响应契约: { "code": 0, "message": "OK", "data": ..., "requestId": "..." }
package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/tommax-bai/tommax-go-kit/errs"
	"github.com/tommax-bai/tommax-go-kit/logx"
)

type envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// OK 写出成功响应。
func OK(w http.ResponseWriter, r *http.Request, data any) {
	writeJSON(w, http.StatusOK, envelope{Code: 0, Message: "OK", Data: data, RequestID: logx.RequestID(r.Context())})
}

// Fail 写出错误响应：任意 error 归一为 errs.Error 后映射 HTTP 状态与业务码。
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	e := errs.From(err)
	// 500 级错误在此统一打日志（业务层只 wrap 不打，见 docs/04 §2.2）。
	if e.HTTPStatus() >= 500 {
		logx.L(r.Context()).Error("request failed", "err", err.Error(), "code", e.Code)
	} else {
		logx.L(r.Context()).Warn("request rejected", "reason", e.Reason, "code", e.Code)
	}
	writeJSON(w, e.HTTPStatus(), envelope{Code: e.Code, Message: e.Message, RequestID: logx.RequestID(r.Context())})
}

// Bind 反序列化 JSON 请求体。
func Bind(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return errs.ErrInvalidParam.WithCause(err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
