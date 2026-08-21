# go-gitmirror

将上游 Git 仓库镜像到本地裸仓：Fetch 更新 refs、维护 mirror 锁、pack 完整性校验、可选 prune。
可用文件系统模拟 refs（不必调用 git 二进制）；API 为 Mirror / RefUpdate / SyncReport。

## Build / Test

```bash
go build ./...
go test ./... -count=1
```

## Daemon

```bash
go run ./cmd/gitd -addr :8114 -root ./data -web web
```

管理页：http://127.0.0.1:8114/
