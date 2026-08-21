# BUG_REPRO: 已删除病虫害条目的详情/删除/更新返回 500

## Bug 是什么
- 访问不存在的病虫害条目详情 `GET /api/v1/pests/:id` 返回 500；
- 删除/更新不存在的条目也返回 500（应全部为 404）。

## 如何触发
1. 启动服务；
2. 请求不存在的条目 id，如 `GET /api/v1/pests/999999`；
3. 用管理员 token 请求 `DELETE /api/v1/pests/999999`、`PUT /api/v1/pests/999999`。

## 真实错误信息
- 上述请求返回 HTTP 500，响应体 `{"code":50000,"message":"internal server error"}`；
- 根因：仓储层 `%v` 断链导致 `errors.Is` 失效，404 被误判成 500。
