# task151-depresolve

依赖版本兼容解析与发布计划引擎。服务把组件、版本、语义化版本约束、平台和能力声明解析成确定性的依赖图与 SQLite 发布快照，并在重启后恢复活动快照和审计记录。

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/depresolve --smoke-test
go run ./cmd/depresolve --addr=:8080 --db=depresolve.db
```

页面入口是 `/`，核心 API 位于 `/api/components`、`/api/resolve` 和 `/api/snapshots`。
