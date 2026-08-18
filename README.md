# tommax-go-kit

Go 公共库：errs（错误码基建，docs/04 §1.6）、logx（结构化日志）、configx（yaml+env）、idgen（雪花）、httpx（响应包+中间件链）、objstore（S3 兼容存储）。

规则（docs/03 §3）：不放业务逻辑与领域类型；只加新不改旧，禁止破坏性变更。
