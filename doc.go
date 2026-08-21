// Package gitmirror 将上游 Git 仓库镜像到本地裸仓目录。
//
// 通过 Transport 拉取 refs/pack，经本地 ref 事务与镜像锁写入；
// 可用文件系统模拟 refs（不必调用 git 二进制），但对外 API 均为 Git 镜像语义：
// Mirror、RemoteSpec、RefUpdate、SyncReport。
package gitmirror
