# UniLLM 优化计划

**创建日期：** 2026-03-25
**基于：** Token Switch vs Geneasy 公平对比测评数据
**当前上游：** Geneasy.AI（临时，后续切换直连）

---

## 完成状态追踪

### P0: 性能与可靠性（Week 1-3）— 追平竞品

- [x] **1.1 HTTP 连接池 + HTTP/2** — `transport.go` 新建，共享 Transport
- [x] **1.2 所有 Provider 共享 Transport** — openai/anthropic/google 三个 provider 已重构
- [x] **1.3 Anthropic 流式读取重写** — bufio.Scanner 替换逐字节读取，吞吐量 10-50x
- [x] **1.4 空输出检测 + 自动重试** — resilience.go 增加 isEmptyOutput 校验
- [x] **1.5 Per-provider 并发限制器** — concurrency.go 新建，信号量模式
- [x] **1.6 Rate limiter bug 修复** — `string(rune(i))` → `strconv.FormatInt`
- [x] **1.7 日志统一** — proxy.go 的 log.Printf → zerolog
- [x] **1.8 类型扩展** — types.go 增加 reasoning_tokens、stream_options、max_completion_tokens
- [x] **1.9 Geneasy 种子数据** — seed_geneasy.sql（12 个模型）
- [x] **1.10 直连迁移脚本** — migrate_to_direct.sql（一键切换）
- [ ] **1.11 Per-model 超时配置** — ModelConfig 增加 TimeoutSeconds 字段
- [ ] **1.12 Nginx upstream 优化** — keepalive、proxy_buffering off
- [ ] **1.13 US-West 部署** — 待购买服务器

### P1: 差异化功能（Week 4-7）— 超越竞品

- [x] **2.1 Embedding API** — `/v1/embeddings` 端点
  - `internal/handler/embedding.go` 新建
  - 支持 text-embedding-3-small/large
  - 复用 auth、billing、rate limiting middleware
  - 已注册到 main.go 路由
- [x] **2.2 reasoning_tokens 透传** — Provider 层解析并传递到 response
  - Google: thoughtsTokenCount → CompletionTokensDetails.ReasoningTokens
  - OpenAI: 直接透传 completion_tokens_details（types.go 已支持）
  - Anthropic: cache usage 提取
- [ ] **2.3 max_reasoning_tokens 参数** — 用户可控制 reasoning 预算
  - Anthropic: 映射到 thinking.budget_tokens
  - Google: 映射到 generationConfig.thinkingConfig.thinkingBudget
- [x] **2.4 Status Page — 主动健康探测**
  - 每 60s 发送轻量探测请求（max_tokens=1，~$0.00001/次）
  - 7 天历史数据环形缓冲区（10080 条记录）
  - `GET /status` 实时状态 + 延迟
  - `GET /status/history` 按小时统计 uptime%
- [ ] **2.5 Function Calling 兼容层**
  - Gemini 多工具检测与补发
  - 自动化 FC 测试套件
- [ ] **2.6 新增 Provider 支持**
  - Grok (xAI) — OpenAI 兼容协议
  - 国产模型：Qwen (DashScope)、Doubao (火山引擎)、Kimi (Moonshot)

### P2: 产品体验（Week 8-10）— 商业化就绪

- [ ] **3.1 单元测试** — 80%+ 覆盖率
- [ ] **3.2 Usage API 增强** — reasoning_tokens 分账
- [ ] **3.3 实时 Dashboard** — WebSocket 消费图表
- [x] **3.4 Fallback 路由链** — `internal/provider/fallback.go` 新建
  - FallbackProvider: 按序尝试多个 provider
  - 每个 provider 独立 API key
  - 自动日志记录 fallback 行为
- [x] **3.5 Redis rate limiter** — `internal/middleware/redis_ratelimit.go` 新建
  - Redis sorted set 滑动窗口
  - 支持多实例部署
  - Redis 故障时 fail-open
- [ ] **3.6 模型列表真实性** — /v1/models 只返回实际可用模型

---

## 关键文件索引

| 文件 | 职责 |
|------|------|
| `internal/provider/transport.go` | 共享 HTTP Transport (连接池+HTTP/2) |
| `internal/provider/concurrency.go` | Per-provider 并发信号量 |
| `internal/provider/resilience.go` | 重试 + 熔断 + 空输出检测 |
| `internal/provider/openai_provider.go` | OpenAI/DeepSeek/Geneasy 上游 |
| `internal/provider/anthropic_provider.go` | Anthropic 原生协议转换 |
| `internal/provider/google_provider.go` | Google Gemini 协议转换 |
| `internal/handler/proxy.go` | 核心代理处理器 |
| `internal/middleware/ratelimit.go` | 请求限速 |
| `pkg/openai/types.go` | OpenAI 兼容类型定义 |
| `scripts/seed_geneasy.sql` | Geneasy 上游种子数据 |
| `scripts/migrate_to_direct.sql` | 切换到直连 API 的迁移脚本 |

---

## 成本估算

| 项目 | 月费 |
|------|------|
| 服务器 (US-West) | $60-80 |
| PostgreSQL | $30-70 |
| Redis | $15-25 |
| 健康探测 API 调用 | ~$0.25 |
| **总计** | **$105-175/月** |

盈亏平衡：15-20 个付费用户 ($29/月)

---

## 启动命令

```bash
# 开发环境
docker-compose up -d postgres redis
psql $DATABASE_URL < scripts/seed_geneasy.sql
go run cmd/server/main.go

# 切换直连（以后）
vim scripts/migrate_to_direct.sql  # 填入 API keys
psql $DATABASE_URL < scripts/migrate_to_direct.sql
```
