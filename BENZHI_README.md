# task151-depresolve 评测说明

本服务面向组件发布工程师，把组件版本、语义化版本约束、平台和能力声明解析成确定性的依赖图与发布快照；冲突、循环、预发布和锁定版本问题均保留可复核证据，SQLite 保存权威目录与解析结果，重启后可恢复活动快照。

## 标准命令

```bash
go build ./...
go run ./cmd/depresolve --addr=:8080 --db=depresolve.db
go test ./...
go vet ./...
go run ./cmd/depresolve --smoke-test
```

启动后页面为 `http://localhost:8080/`，页面会访问 `/api/stats` 与 `/api/resolve`；smoke-test 会实际请求页面和业务 API，创建组件/版本/依赖，解析并激活快照，关闭数据库后重新打开并验证恢复。

## Docker

```bash
bash ./build_benzhi_docker.sh task151-depresolve:amd64 linux/amd64
docker run --rm task151-depresolve:amd64 go version
docker run --rm task151-depresolve:amd64 /app/depresolve --smoke-test

bash ./build_benzhi_docker.sh task151-depresolve:arm64 linux/arm64
docker run --rm task151-depresolve:arm64 go version
docker run --rm task151-depresolve:arm64 /app/depresolve --smoke-test
```

进入容器：`docker run -it task151-depresolve:amd64`。构建脚本参数依次为镜像名和目标平台；镜像 builder 使用 Go 1.26.3 bookworm，SQLite 依赖在 module mode 下下载。
