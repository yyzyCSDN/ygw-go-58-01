# ReconcileSvc

ReconcileSvc 是一个通用数据对账与差异检测引擎。它按窗口从源侧读取数据，
与目标侧快照逐键比对，检测缺失、不一致与多余三类差异，将差异写入差异库，
并维护对账位点以支持失败重试与全量/增量两阶段对账。

## 构建

```bash
go build -mod=vendor ./...
```

依赖已 vendor 进仓库，构建过程不访问网络。

## 运行

```bash
go run ./cmd/reconcilesvc -addr :8080
```

启动后打开 <http://localhost:8080/> 查看对账监控页面，或直接调用 REST 接口：

- `GET /api/health` 健康检查
- `POST /api/reconcile/run` 触发一次全量对账
- `GET /api/reconcile/status` 对账任务状态
- `GET /api/diffs` 差异列表
- `GET /api/offsets` 对账位点
- `GET /api/offsets/switch-incremental` 切换到增量对账
- `GET /api/notify/summary` 通知统计

## Docker

```bash
docker build -f benzhi.Dockerfile -t reconcilesvc .
docker run --rm -p 8080:8080 reconcilesvc
```
