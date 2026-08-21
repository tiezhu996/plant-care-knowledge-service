# gb-61 植物养护知识百科平台 执行记录

- 项目编号/名称：gb-61 / 植物养护知识百科平台（gbplantwiki）
- 日期：2026-08-16
- 短名：gbplantwiki
- 端口：前端 8102 / 后端 3102 / MySQL 3502
- 技术栈：Vue 3 + TypeScript + Element Plus + Vite（前端）；Go 1.22 + Gin + GORM（后端）；MySQL 8.0；JWT + RBAC

## Docker Compose 结果

| 容器 | 状态 | 端口 |
| --- | --- | --- |
| gbplantwiki_db | healthy | 0.0.0.0:3502->3306 |
| gbplantwiki_backend | healthy | 0.0.0.0:3102->8080 |
| gbplantwiki_frontend | up | 0.0.0.0:8102->80 |

`docker compose config --quiet` 通过；`docker compose up -d --build` 全容器 healthy；验证后 `docker compose down -v --remove-orphans` 无残留。

## 关键 API 冒烟结果（23 项，全部通过）

| 接口 | 方法 | 状态码 | 结果摘要 |
| --- | --- | --- | --- |
| /api/v1/home/overview | GET | 200 | 首页聚合：热门品种+最新文章+当季任务 |
| /api/v1/users/register | POST | 201 | 注册返回 JWT + 用户信息 |
| /api/v1/users/login | POST | 200 | 登录成功返回 token |
| /api/v1/users/me（带 token） | GET | 200 | 获取当前用户资料 |
| /api/v1/plants?page=1 | GET | 200 | 品种分页列表 |
| /api/v1/plants/:id | GET | 200 | 品种详情 |
| /api/v1/articles?page=1 | GET | 200 | 文章列表（中文正常） |
| /api/v1/articles/:id | GET | 200 | 文章详情+阅读数 |
| /api/v1/pests?keyword=黑斑 | GET | 200 | 病虫害关键词搜索 |
| /api/v1/favorites | POST | 201 | 收藏植物成功 |
| /api/v1/favorites（重复） | POST | 409 | 重复收藏冲突 |
| /api/v1/favorites/plant/:id | DELETE | 200 | 取消收藏 |
| /api/v1/gardens | POST | 201 | 加入我的花园 |
| /api/v1/gardens（重复） | POST | 409 | 重复添加冲突 |
| /api/v1/reminders | POST | 201 | 创建养护提醒 |
| /api/v1/reminders/:id/status | PUT | 200 | 状态流转 pending->done |
| /api/v1/questions | POST | 201 | 发布问题 |
| /api/v1/questions/:id/answers | POST | 201 | 回答问题 |
| /api/v1/questions/:id/adopt | PUT | 200 | 采纳最佳回答 |
| /api/v1/answers/:id/like | PUT | 200 | 回答点赞 |
| /api/v1/plants（普通用户） | POST | 403 | RBAC 拒绝非管理员 |
| /api/v1/plants（管理员） | POST | 201 | 管理员可创建品种 |
| /api/v1/users/me（无 token） | GET | 401 | 未认证被拒 |
| /api/v1/plants/999999 | GET | 404 | 不存在返回 404 |
| /api/v1/users/register（非法参数） | POST | 400 | 参数校验失败 |
| /api/v1/users/login（错误密码） | POST | 401 | 密码错误拒绝 |
| /api/v1/users/login（70 次快速） | POST | 429 | 限流生效（21 次 429） |

## 浏览器验证结论（内置 playwright，无外部 Chrome）

- 首页 http://localhost:8102：标题、导航、当季养护重点、热门品种（冒烟测试植物X/清香木/库拉索芦荟/月季/碗莲/多肉吉娃娃/龟背竹）均来自后端 API，中文正常 → 真实数据 ✅
- 品种库 /plants：类型/科/关键词筛选 + 品种卡片列表（真实数据）✅
- 问答社区 /questions：发布问题表单 + 问题列表（冒烟测试问题/种子问题，含结题状态）✅
- 登录 /login：fill+click 完成登录，跳转首页并显示昵称"冒烟用户" → 前端认证链路 ✅
- 截图：output/home.png、output/home_logged_in.png、output/plants.png

## README 检查项

- Docker Compose 一键启动命令置顶（cp .env.example .env && docker compose up -d --build）✅
- 本地开发（go mod tidy / go run ./cmd/server / go build ./... / npm run dev）✅
- 技术栈表格（后端 Go 1.22 + Gin + GORM）✅、目录结构代码块 ✅、环境变量表 ✅、部署说明 ✅
- 枚举出现位置清单：PlantType / CareTopicTag / FavoriteTargetType（前后端全部位置）✅
- 横切关注点清单 ✅、License ✅

## 其他质量项

- 后端：`go build ./...` 通过；`go vet ./...` 通过；`go test ./...` 全通过（config/util/service/repository 表驱动测试）
- 前端：`npm run build` 零错误（vue-tsc 类型检查 + vite 构建通过）
- 屎山设计：log_templates.go 33 条模板全栈引用、错误信息含实体/字段/角色拼接、formatters 多职责耦合、提醒状态机跨多处定义、枚举前后端重复定义 ✅

- 执行状态：完成（API 冒烟 23/23 通过，浏览器验证通过，docker down -v 无残留）
