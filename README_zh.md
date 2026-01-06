# Certificate Exporter

一个用Go编写的Prometheus Exporter，用于监控本地和远程SSL/TLS证书的过期时间和有效性。

## 功能特性

- 📡 **远程证书监控**：通过TCP连接检查远程服务器的TLS证书
- 📁 **本地证书监控**：检查本地文件系统上的证书文件
- ⏰ **异步定时采集**：后台定时采集，避免Prometheus查询时的性能问题
- 📊 **多维度指标**：提供证书过期天数、有效性状态、检查延迟等指标
- 🛡️ **线程安全缓存**：使用并发安全的缓存机制存储采集结果

## 安装

### 从源代码构建

```bash
# 克隆仓库
git clone https://github.com/HuckOps/cert_exporter.git
cd cert_exporter

# 构建二进制文件
go build -o cert_exporter

# 运行
./cert_exporter --config config.yaml
```

### 直接运行

```bash
# 确保安装了Go 1.18+
go run . --config config.yaml
```

## 配置

### 配置文件示例 (config.yaml)

```yaml
# 日志级别：debug, info, warn, error
log_level: "info"

# 采集间隔（秒）
interval: 60

# 远程证书监控目标
remote:
  - "example.com:443"
  - "github.com:443"
  - "prometheus.io:443"

# 本地证书监控目标
local:
  - public_key_path: "/path/to/local/cert.pem"
  - public_key_path: "/path/to/another/cert.pem"
```

### 命令行参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `--config` | 配置文件路径 | `config.yaml` |
| `--web.listen-address` | 监听地址和端口 | `:9101` |
| `--web.telemetry-path` | 暴露指标的路径 | `/metrics` |

## 暴露的指标

### `certificate_expiry_days`
证书过期剩余天数，Gauge类型。

标签：
- `domain`：证书对应的域名
- `sn`：证书序列号
- `source_type`：证书来源（remote/local）
- `source`：证书来源（配置条目）

### `certificate_valid`
证书是否有效（1=有效，0=无效），Gauge类型。

标签：
- `domain`：证书对应的域名
- `sn`：证书序列号
- `source_type`：证书来源（remote/local）
- `source`：证书来源（配置条目）

### `certificate_subject`
证书主题信息，Gauge类型。

标签：
- `domain`：证书对应的域名
- `sn`：证书序列号
- `subject`：证书主题信息
- `source_type`：证书来源（remote/local）
- `source`：证书来源（配置条目）

### `certificate_check_status`
证书检查状态（1=成功，0=失败），Gauge类型。

标签：
- `domain`：证书对应的域名
- `source_type`：证书来源（remote/local）
- `source`：证书来源（配置条目）

### `certificate_check_latency_milliseconds`
检查证书所花费的时间（毫秒），Gauge类型。

标签：
- `domain`：证书对应的域名
- `source_type`：证书来源（remote/local）
- `source`：证书来源（配置条目）

## Prometheus配置示例

将以下内容添加到您的Prometheus配置文件中：

```yaml
scrape_configs:
  - job_name: 'certificate_exporter'
    static_configs:
      - targets: ['localhost:9101']
    scrape_interval: 60s
```

## 技术栈

- **Go**：1.18+
- **Prometheus客户端库**：github.com/prometheus/client_golang
- **Zap**：go.uber.org/zap（日志框架）
- **YAML**：go.yaml.in/yaml/v2（配置解析）

## 许可证

[MIT License](LICENSE)