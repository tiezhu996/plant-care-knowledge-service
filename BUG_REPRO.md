# BUG_REPRO: 问答详情首访 panic（nil map 写入）

## Bug 是什么
- 首次访问问题详情时向未初始化的 `viewStats` map 写计数，直接 `panic: assignment to entry in nil map`；
- id=0 走 typed-nil 旁路返回 200 null；空问题列表返回 null 而非 []。

## 如何触发
1. 启动服务并登录；
2. 请求 `GET /api/v1/questions/1`（首次）；
3. 请求 `GET /api/v1/questions/0`、`GET /api/v1/questions`。

## 真实错误信息
- `panic: assignment to entry in nil map`，goroutine 栈指向 `QuestionService.Get`；
- `GET /api/v1/questions/0` 返回 200 且 data 为 null（应 404）。
