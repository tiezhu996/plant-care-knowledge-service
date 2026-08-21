# BUG_REPRO: 已删除品种的详情/删除/更新返回 500

## Bug 是什么
- 访问一个不存在的植物品种（已被删除）的详情页 `GET /api/v1/plants/:id` 返回 500；
- 对不存在的品种执行删除 `DELETE /api/v1/plants/:id` 也返回 500；
- 对不存在的品种执行更新 `PUT /api/v1/plants/:id` 同样返回 500（应全部为 404）。

## 如何触发
1. 启动服务（`go run ./cmd/server`）；
2. 请求一个不存在的品种 id，例如 `GET /api/v1/plants/999999`；
3. 用管理员 token 请求 `DELETE /api/v1/plants/999999` 或 `PUT /api/v1/plants/999999`。

## 真实错误信息
- `GET /api/v1/plants/999999` → HTTP 500，响应体 `{"code":50000,"message":"internal server error"}`
- 后端日志：`unhandled error ... plant get failed: ...`
- 根因：repository 层用 `%v` 包装 `ErrNotFound` 导致 `errors.Is` 无法识别，service/handler 因此把“不存在”误判成系统错误返回 500。
