# BUG_REPRO: 普通用户可自提权为管理员

## Bug 是什么
- 资料更新接口接受 `role` 字段，普通用户可把自己改成 admin；
- AuthRequired 只信任 JWT 里的旧角色，不查库，提权立即生效且旧 token 长期有效；
- JWT 解析不校验签发方 issuer。

## 如何触发
1. 启动服务并登录普通用户；
2. `PUT /api/v1/users/me` 带 `{"role":"admin"}`；
3. 用提权后的 token 访问管理接口。

## 真实错误信息
- 普通用户资料更新返回 200 且 role 变为 admin；
- 带 admin 角色的 token 可直接调用管理接口（应 403）。
