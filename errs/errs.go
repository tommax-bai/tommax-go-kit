// Package errs 实现 04 规范定义的错误码基建：
// 8 位数字码 AABBCCC（AA 映射 HTTP 大类，BB 域编号，CCC 序号）+ SCREAMING_SNAKE reason。
// 业务代码只创建/包装 Error，日志与响应转换发生在最外层（httpx / grpc 拦截器）。
package errs

import (
	"errors"
	"fmt"
)

// 域编号注册表（见 docs/04 §1.6，禁止私设）。
const (
	DomainCommon       = 1
	DomainUser         = 11
	DomainBilling      = 12
	DomainGeneration   = 13
	DomainModelAdapter = 14
	DomainCanvas       = 15
	DomainCollab       = 16
	DomainAsset        = 17
	DomainMedia        = 18
)

type Error struct {
	Code    int    `json:"code"`    // 8 位：AA(HTTP 大类) BB(域) CCC(序号)
	Reason  string `json:"reason"`  // 稳定的机器可读标识，客户端据此分支
	Message string `json:"message"` // 展示文案，允许随时调整
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s(%d): %s: %v", e.Reason, e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s(%d): %s", e.Reason, e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithCause 返回携带底层错误的副本，保持原码与 reason。
func (e *Error) WithCause(cause error) *Error {
	clone := *e
	clone.cause = cause
	return &clone
}

// WithMessagef 返回覆盖展示文案的副本。
func (e *Error) WithMessagef(format string, args ...any) *Error {
	clone := *e
	clone.Message = fmt.Sprintf(format, args...)
	return &clone
}

// HTTPStatus 从码的前两位推导 HTTP 状态。
func (e *Error) HTTPStatus() int {
	switch e.Code / 1000000 {
	case 40:
		switch e.Code / 100000 % 10 { // 第三位区分 400/401/402/403/404/409/429
		default:
			return 400
		}
	case 41:
		return 401
	case 42:
		return 402
	case 43:
		return 403
	case 44:
		return 404
	case 49:
		return 429
	case 50:
		return 500
	default:
		return 500
	}
}

// New 注册一个错误码。aa 取 40/41/42/43/44/49/50，domain 取上方注册表，seq 三位序号。
func New(aa, domain, seq int, reason, message string) *Error {
	return &Error{Code: aa*1000000 + domain*1000 + seq, Reason: reason, Message: message}
}

// From 从任意 error 提取 *Error；非本包错误归一为 ErrInternal。
func From(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternal.WithCause(err)
}

// 通用域预置错误。
var (
	ErrInternal     = New(50, DomainCommon, 1, "INTERNAL", "服务开小差了，请稍后重试")
	ErrInvalidParam = New(40, DomainCommon, 2, "INVALID_PARAM", "参数不合法")
	ErrUnauthorized = New(41, DomainCommon, 3, "UNAUTHORIZED", "请先登录")
	ErrForbidden    = New(43, DomainCommon, 4, "FORBIDDEN", "无权访问该资源")
	ErrNotFound     = New(44, DomainCommon, 5, "NOT_FOUND", "资源不存在")
	ErrRateLimited  = New(49, DomainCommon, 6, "RATE_LIMITED", "请求过于频繁，请稍后再试")
)
