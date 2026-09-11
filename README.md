# GoDrop

[![CI](https://github.com/1347032239-prog/GoDrop/actions/workflows/ci.yml/badge.svg)](https://github.com/1347032239-prog/GoDrop/actions/workflows/ci.yml)

GoDrop 是一个只使用 Go 标准库实现的文件索引与传输工具。服务端扫描共享目录并发布文件索引，客户端可以查看索引、下载单个文件，或并发同步索引中的全部文件。

## 功能

- 递归扫描目录，记录相对路径、大小和 SHA-256
- 将扫描结果保存为 JSON 索引，并可重新读取展示
- 通过 HTTP 提供文件索引和文件流
- 限制下载范围，阻止目录逃逸并只允许下载索引中的文件
- 客户端流式下载、临时文件写入和原子替换
- 下载后校验 SHA-256，发现扫描后被修改的文件
- 使用 goroutine、channel 和 `sync.WaitGroup` 有界并发同步
- 支持超时、信号取消、HTTP 服务超时和优雅关闭
- 提供 JSON 结构化日志、健康探针、就绪探针和轻量指标
- 支持 Docker 镜像和 GitHub Actions 自动检查

## 工作流程

```text
共享目录
   │
   │ scan：扫描路径、大小和 SHA-256
   ▼
index.json
   │
   │ serve：发布 /files 和 /download
   ▼
HTTP 服务
   │
   │ fetch / download / sync
   ▼
客户端目录
```

## 环境要求

- Go 1.27 或兼容版本
- Docker（仅在容器运行方式中需要）

## 构建

```bash
git clone https://github.com/1347032239-prog/GoDrop.git
cd GoDrop
go build -o godrop .
```

查看支持的命令：

```bash
./godrop
```

预期输出：

```text
支持用法:scan , show, serve, fetch, download, sync
```

## 快速开始

下面的示例在一台机器的两个终端中完成一次完整同步。

### 1. 准备共享文件并生成索引

服务端终端：

```bash
cd ~/Go/GoDrop

mkdir -p /tmp/godrop-demo/share/nested
printf 'hello' > /tmp/godrop-demo/share/a.txt
printf 'world' > /tmp/godrop-demo/share/nested/b.txt

go run . scan \
  -dir /tmp/godrop-demo/share \
  -out /tmp/godrop-demo/index.json
```

关键输出：

```text
文件数量: 2
总大小: 10 bytes
索引已写入:/tmp/godrop-demo/index.json
```

### 2. 启动服务

继续在服务端终端运行：

```bash
go run . serve \
  -index /tmp/godrop-demo/index.json \
  -dir /tmp/godrop-demo/share \
  -addr 0.0.0.0:18080
```

该命令在前台持续运行。按 `Ctrl+C` 会触发优雅关闭。

### 3. 同步全部文件

打开第二个终端作为客户端：

```bash
cd ~/Go/GoDrop

go run . sync \
  -index-url http://127.0.0.1:18080/files \
  -download-url http://127.0.0.1:18080/download \
  -out /tmp/godrop-demo/client \
  -workers 2 \
  -timeout 30s
```

预期输出：

```text
同步完成: 2 files, 10 bytes
```

验证客户端文件与服务端文件完全一致：

```bash
cmp /tmp/godrop-demo/share/a.txt /tmp/godrop-demo/client/a.txt
cmp /tmp/godrop-demo/share/nested/b.txt /tmp/godrop-demo/client/nested/b.txt
echo $?
```

两个 `cmp` 都应当没有输出，最后的退出码应当是：

```text
0
```

## 命令

### scan

扫描目录并可选地写入 JSON 索引：

```bash
go run . scan -dir <共享目录> [-out <索引文件>]
```

### show

读取并展示本地索引：

```bash
go run . show -index <索引文件>
```

### serve

启动 HTTP 服务：

```bash
go run . serve \
  -index <索引文件> \
  -dir <共享目录> \
  [-addr 127.0.0.1:8080]
```

### fetch

获取并展示远程索引：

```bash
go run . fetch -url http://127.0.0.1:8080/files
```

### download

下载一个文件：

```bash
go run . download \
  -url http://127.0.0.1:8080/download \
  -path nested/b.txt \
  -out /tmp/downloaded-b.txt \
  -timeout 30s
```

### sync

并发同步索引中的全部文件：

```bash
go run . sync \
  -index-url http://127.0.0.1:8080/files \
  -download-url http://127.0.0.1:8080/download \
  -out /tmp/godrop-client \
  -workers 4 \
  -timeout 2m
```

`workers` 必须在 1 到 32 之间。第一次下载失败会取消本批次剩余工作。

## HTTP 接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/files` | 返回启动时加载的 JSON 文件索引 |
| `GET` | `/download?path=<相对路径>` | 流式返回索引中的文件 |
| `GET` | `/healthz` | 进程存活检查，正常返回 `ok` |
| `GET` | `/readyz` | 服务就绪检查，正常返回 `ready` |
| `GET` | `/metrics` | 返回 Prometheus 文本格式的轻量指标 |

服务启动后可以手动检查：

```bash
curl -fsS http://127.0.0.1:18080/healthz
curl -fsS http://127.0.0.1:18080/readyz
curl -fsS http://127.0.0.1:18080/metrics
```

## 服务端环境变量

命令行参数优先于环境变量。

| 环境变量 | 对应参数 | 默认值 |
| --- | --- | --- |
| `GODROP_INDEX` | `-index` | 无，必须提供 |
| `GODROP_DIR` | `-dir` | 无，必须提供 |
| `GODROP_ADDR` | `-addr` | `127.0.0.1:8080` |

示例：

```bash
GODROP_INDEX=/tmp/godrop-demo/index.json \
GODROP_DIR=/tmp/godrop-demo/share \
GODROP_ADDR=0.0.0.0:18080 \
go run . serve
```

## Docker

构建镜像：

```bash
sudo docker build -t godrop:local .
```

使用前面生成的共享目录和索引启动容器：

```bash
sudo docker run --rm \
  --name godrop-server \
  -p 18080:8080 \
  -e GODROP_INDEX=/config/index.json \
  -e GODROP_DIR=/data \
  -v /tmp/godrop-demo/index.json:/config/index.json:ro \
  -v /tmp/godrop-demo/share:/data:ro \
  godrop:local
```

这里的 `:ro` 表示只读挂载；容器可以读取索引和共享文件，但不能修改宿主机内容。客户端仍然访问：

```text
http://127.0.0.1:18080
```

## 网络地址

- `127.0.0.1` 只允许当前机器访问。
- 服务端监听 `0.0.0.0:18080` 时，会监听所有网卡。
- 局域网中的另一台机器应使用服务端实际 IP，例如 `http://192.168.147.128:18080/files`。
- 如果跨机器访问失败，请检查服务端防火墙、虚拟机网络模式和端口映射。

## 一致性与安全边界

- 下载路径必须是索引中存在的本地相对路径。
- 服务端通过受限文件根目录打开文件，避免路径逃逸到共享目录之外。
- 索引保存扫描时的 SHA-256；如果文件在扫描后被修改，客户端会发现哈希不一致。
- 客户端先写入同目录临时文件，下载和校验成功后再原子替换目标文件。
- `sync` 使用固定数量的 worker，不会为每个文件无限制创建 goroutine。

## 当前不包含

- 用户账号、认证和权限系统
- TLS/HTTPS 配置
- 断点续传和 HTTP Range
- 数据库、消息队列和服务发现
- 分布式存储、复制和一致性协议
- Kubernetes 部署清单

这些能力适合在 GoDrop 完成后，通过后续的云原生和多服务项目继续学习。

## 自动检查

GitHub Actions 会在推送代码以及创建或更新面向 `master` 的 Pull Request 时自动执行：

```text
gofmt 检查
go test -count=1 ./...
go vet ./...
Docker 镜像构建
```
