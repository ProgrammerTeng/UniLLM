# UniLLM 技术交接

日期：2026-05-24

这是当前 UniLLM 平台代码和产品文档包。请先帮忙判断这套代码适合继续迭代，还是需要先重构关键模块。

## 包内容

- `unillm/`：平台代码快照，包含本地未提交的最新改动
- `product-docs/prd.md`：早期 PRD
- `product-docs/product-plan-multi-model-platform.md`：产品方案、竞品差异和路线
- `specs/`：需求与重构方案文档

## 当前功能状态

后端已实现：

- 用户注册/登录、JWT、API Key 管理
- OpenAI-compatible `/v1/chat/completions`
- 流式/非流式模型代理
- `/v1/models`
- `/v1/embeddings` 初版
- 余额校验、用量记录、成本计算
- Dashboard Usage API
- Admin API：用户、余额、Provider、模型、Provider Key
- Provider 适配：OpenAI-compatible、Anthropic、Google Gemini
- 共享 HTTP transport、fallback、并发限制、空输出重试
- Status Page 主动健康探测
- Redis 滑动窗口限流
- Provider Key AES 加密（`ENCRYPTION_KEY` 配置）

前端已实现：

- 首页、登录注册、Dashboard、模型页、费用计算器、API 文档、Playground、Status、Admin

## 架构（Phase 0–2 重构后）

```
api/          → HTTP 入口（v1 / dashboard / admin / middleware）
core/         → 业务逻辑（billing / catalog / inference）
infra/        → 基础设施（persistence / provider / jwt / crypto / billing）
internal/     → 配置、日志、GORM 实体、遗留 service（auth）
```

## 重点代码入口

- `unillm/cmd/server/main.go`：组合根，组装 core / infra / api
- `unillm/api/v1/proxy.go`：Chat HTTP 绑定层
- `unillm/core/inference/`：Chat / Embedding 编排
- `unillm/core/billing/`：余额校验与用量记录
- `unillm/core/catalog/`：模型路由与 Provider Key 池
- `unillm/infra/provider/`：上游 Provider 适配、Fallback、重试
- `unillm/infra/persistence/`：GORM 数据访问
- `unillm/api/admin/`：管理后台 API
- `unillm/web/src/app/`：Next.js 页面

## 本地启动

```bash
cd unillm
docker-compose up -d postgres redis
psql "$DATABASE_URL" < scripts/seed_geneasy.sql
go run cmd/server/main.go
```

```bash
cd unillm/web
npm install
npm run dev
```

验证：

```bash
cd unillm && go test ./...
cd unillm/web && npm run build
```

## 环境变量（新增）

| 变量 | 说明 |
|------|------|
| `ENCRYPTION_KEY` | 64 位 hex，AES-256 加密 Provider Key（生产必填） |
| `FALLBACK_CHAIN` | 可选，格式 `groupName:provider1,provider2` 启用 Fallback |

## 已知风险

- 测试覆盖不足，尤其是流式计费、余额扣减、Provider fallback、Admin 权限
- `internal/service/auth.go` 尚未迁入 `core/auth`（Phase 3 可选）
- 线上服务目前不可从本地访问，需要检查服务器、Docker/Nginx、防火墙和端口
