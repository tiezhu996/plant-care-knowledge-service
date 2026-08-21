# BUG_REPRO: 回答点赞并发时 data race，点赞数丢更新

## Bug 是什么
- 多个请求并发点赞同一回答时触发 `WARNING: DATA RACE`；
- 点赞数先写内存热点 map 再批量落库，map 无锁并发写导致竞态，落库时用覆盖而非累加，点赞数丢失。

## 如何触发
1. 启动服务并登录；
2. 对同一回答并发发起 `PUT /api/v1/answers/:id/like`（如 20 个并发请求）。

## 真实错误信息
- `-race` 运行输出 `WARNING: DATA RACE`，读写栈指向 `service.(*AnswerService).Like` 的 `s.hotLikes[answerID]++`；
- 点赞数落库后低于实际点赞次数（如 20 次并发只落库 1 次）。
