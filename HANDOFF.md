# UniLLM 技术交接

日期：2026-05-24

这是当前 UniLLM 平台代码和产品文档包。请先帮忙判断这套代码适合继续迭代，还是需要先重构关键模块。

## 包内容

- `unillm/`：平台代码快照，包含本地未提交的最新改动
- `product-docs/prd.md`：早期 PRD
- `product-docs/product-plan-multi-model-platform.md`：产品方案、竞品差异和路线

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

前端已实现：

- 首页、登录注册、Dashboard、模型页、费用计算器、API 文档、Playground、Status、Admin

## 重点代码入口

- `unillm/cmd/server/main.go`：路由注册、依赖组装、Provider 初始化
- `unillm/internal/handler/proxy.go`：核心 chat completions 代理、计费记录
- `unillm/internal/provider/`：各上游 Provider 适配、fallback、重试、并发控制
- `unillm/internal/service/billing.go`：用量和成本记录
- `unillm/internal/middleware/`：鉴权、余额、限流、metrics
- `unillm/internal/handler/admin.go`：后台管理 API
- `unillm/web/src/app/`：Next.js 页面
- `unillm/migrations/001_init.sql`：数据库结构
- `unillm/scripts/seed_geneasy.sql`：Geneasy 上游初始化脚本，真实 key 已替换成占位符

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

我本地验证结果：

- `go test ./...` 通过，但基本没有单元测试，主要代表能编译
- `npm run build` 通过
- 生产地址 `http://45.76.67.69:3000/` 本地探测超时，线上部署需要重新检查

## 已知风险

- 当前代码是工作区快照，包含未提交改动，协作前建议先整理成正式 Git 分支
- 测试覆盖不足，尤其是流式计费、余额扣减、Provider fallback、Admin 权限
- 上游 Provider key 管理和模型可用性还需要生产级验证
- `seed_geneasy.sql` 适合初始化演示环境，不应直接带真实 key 入库
- 线上服务目前不可从本地访问，需要检查服务器、Docker/Nginx、防火墙和端口

## 希望你重点判断

1. 这套架构能否继续演进，还是 Provider、billing、auth/admin 需要先重构。
2. 如果继续做，第一阶段应该补哪些测试和生产化能力。
3. 上游策略是继续走 Geneasy 代理，还是尽快切 OpenAI/Anthropic/Google/DeepSeek 直连。
4. 前端现在是够用型页面，是否需要重做设计系统和组件结构。
5. 如果你接手，建议先从哪几个模块下刀，预计工作量多大。
