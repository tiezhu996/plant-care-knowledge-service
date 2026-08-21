# BUG_REPRO: 文章浏览量冲刷 worker 挂死、丢数据

## Bug 是什么
- 浏览量先写入内存缓冲，停止 worker 时直接丢弃 pending 数据不落库；
- 冲刷失败时写无缓冲 errCh 阻塞，StopViewFlusher 永久挂死；
- worker 停止后 RecordView 阻塞请求。

## 如何触发
1. 启动服务；
2. 浏览文章产生浏览量；
3. 触发 worker 停止或冲刷失败。

## 真实错误信息
- 停止后浏览量未落库；冲刷失败时 `StopViewFlusher` 卡死；
- worker 停止后再浏览文章请求阻塞无响应。
