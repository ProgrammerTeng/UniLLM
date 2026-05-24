# 02 — 主题切换（明/暗）与多语言（中/英）

**状态：** Phase 1 + Phase 2 已实现（2026-05-24）  
**日期：** 2026-05-24  
**范围：** `unillm/web/` 前端  
**不包含：** 后端 API 多语言、邮件/通知国际化、RTL 语言

---

## 1. 背景与目标

### 1.1 现状

| 项 | 当前实现 |
|----|----------|
| **主题** | 仅暗色。`globals.css` 中 `:root` 写死深色变量（背景 `#0a0a0a`、前景 `#ededed` 等） |
| **语言** | 仅英文。`layout.tsx` 中 `lang="en"`；各页面文案为硬编码英文字符串 |
| **顶栏** | `Sign In` 按钮出现在首页、`/models`、`/calculator` 等公开页的 header 右侧；各页面 header 代码重复，未抽公共组件 |

用户感知为「纯黑界面 + 全英文」，无切换入口。

### 1.2 目标

1. 支持 **浅色 / 深色** 两种主题，用户可手动切换，选择持久化到本地。
2. 支持 **英文（en）** 与 **简体中文（zh-CN）** 两种界面语言，用户可手动切换，选择持久化到本地。
3. 在 **Sign In 按钮左侧** 放置「主题切换」与「语言选择」两个控件（见 §3.1 布局）。
4. 抽离公共顶栏组件，避免 4+ 处 header 复制粘贴。

### 1.3 非目标

- 不跟随系统主题自动切换（可作为后续增强，本需求默认 **手动切换**）
- 不支持繁体中文、日语等第三语言
- 不翻译 API 错误码、上游模型名、Docs 中的代码示例与 HTTP 原文
- 不要求 SSR 阶段按 `Accept-Language` 猜语言（首屏以 localStorage / 默认 `en` 为准）

---

## 2. 交互与视觉设计

### 2.1 顶栏布局（Sign In 左侧）

适用于 **未登录** 且 header 右侧为 **Sign In** 的公开页（至少：`/`、`/models`、`/calculator`）。

```
┌─────────────────────────────────────────────────────────────────┐
│ UniLLM  Models  Playground  Docs          [🌙/☀️] [EN ▾] [Sign In] │
└─────────────────────────────────────────────────────────────────┘
                              ↑ 主题      ↑ 语言    ↑ 现有按钮
```

**顺序（从左到右，紧靠 Sign In 左侧）：**

1. **主题切换按钮** — 图标按钮（暗色模式下显示「太阳」示意切到浅色；浅色模式下显示「月亮」示意切到深色）
2. **语言选择** — 下拉或分段控件，显示当前语言缩写：`EN` / `中文`
3. **Sign In**（或已登录时的 Dashboard / Logout）— 保持现有样式与行为

**间距：** 主题与语言控件作为一组，`gap` 与 Sign In 之间 `gap-3`～`gap-4`，与现有 header `gap-4` 一致。

### 2.2 已登录顶栏

在 **Dashboard / Logout** 左侧同样放置主题 + 语言控件（顺序不变），保证登录态全站一致。

涉及页面（header 需统一）：`/`、`/models`、`/calculator`、`/dashboard`、`/playground`、`/docs`、`/admin`（若 header 结构不同，至少在含全局导航的 header 右侧操作区左侧加入）。

### 2.3 登录页 `/login`

登录页 **无 Sign In 按钮**，但应在页面 **右上角** 或 **Logo 上方右侧** 提供相同的主题 + 语言控件（不强制 Sign In 锚点，视觉与公开页一致）。

### 2.4 主题色板

**深色（默认，与现网一致）**

| 变量 | 值 |
|------|-----|
| `--background` | `#0a0a0a` |
| `--foreground` | `#ededed` |
| `--card` | `#141414` |
| `--border` | `#262626` |
| `--muted` | `#737373` |

**浅色（新增）**

| 变量 | 建议值 |
|------|--------|
| `--background` | `#ffffff` |
| `--foreground` | `#171717` |
| `--card` | `#f5f5f5` |
| `--border` | `#e5e5e5` |
| `--muted` | `#737373` |

`--primary` / `--primary-hover` 两种主题共用，无需改动。

实现方式：在 `html` 或 `body` 上挂 `data-theme="dark" | "light"`（或 class `dark` / `light`），在 `globals.css` 用选择器覆盖变量；**禁止** 在组件内散落硬编码 `#0a0a0a`。

### 2.5 语言切换行为

| 语言 | `html[lang]` | 存储 key | 显示名 |
|------|--------------|----------|--------|
| 英文 | `en` | `unillm-locale` = `en` | `EN` |
| 简体中文 | `zh-CN` | `unillm-locale` = `zh-CN` | `中文` |

切换后立即更新界面文案，无需刷新页面（Client Component + Context）。

---

## 3. 技术方案

### 3.1 推荐依赖

| 能力 | 方案 | 说明 |
|------|------|------|
| 主题 | `next-themes` | 与 Next.js App Router 兼容；`attribute="data-theme"` |
| 文案 | 轻量自建 `lib/i18n/` | 字典 JSON/TS，无强制引入 `next-intl`（减少配置量） |

若团队更熟 `next-intl`，可在实现阶段替换，但本 spec 以 **Context + 字典** 为验收基准。

### 3.2 目录结构（新增/调整）

```
web/src/
├── components/
│   ├── SiteHeader.tsx          # 统一顶栏（含 nav + 主题 + 语言 + 登录区）
│   ├── ThemeToggle.tsx         # 图标按钮
│   └── LocaleSwitcher.tsx      # EN / 中文 切换
├── lib/
│   ├── i18n/
│   │   ├── index.ts            # useI18n、I18nProvider
│   │   ├── locales.ts          # 类型、默认语言
│   │   └── messages/
│   │       ├── en.ts
│   │       └── zh-CN.ts
│   └── theme/                  # 可选：主题相关常量
├── app/
│   ├── layout.tsx              # 包裹 Providers（Theme + I18n）
│   ├── providers.tsx           # Client providers 组合
│   └── globals.css             # dark/light CSS 变量
```

### 3.3 Providers 挂载

在 `app/layout.tsx` 中：

```tsx
<html lang={locale} suppressHydrationWarning>
  <body>
    <Providers>{children}</Providers>
  </body>
</html>
```

- `ThemeProvider`（`next-themes`）：`defaultTheme="dark"`，`enableSystem={false}`（与本需求非目标一致）
- `I18nProvider`：读取 `localStorage` 初始语言，避免闪烁可加 `suppressHydrationWarning` 于 `html`

### 3.4 文案覆盖范围（Phase 1 必须）

| 区域 | 翻译 |
|------|------|
| 公开顶栏导航 | Models、Playground、Docs、Sign In、Dashboard、Logout |
| 首页 Hero / Features / Stats / CTA | 是 |
| `/login` 表单与按钮 | 是 |
| `/models`、`/calculator` 页标题与表头 | 是 |
| `/dashboard` 侧栏/标题（若有） | 是 |

**Phase 1 可不翻译（保留英文）：**

- `/docs` 长文档正文（体量大，可 Phase 2）
- Admin 后台细节文案（可 Phase 2）
- Playground 内部分提示（可 Phase 2）

### 3.5 无障碍与细节

- 主题按钮：`aria-label` 随语言变化（如「切换到浅色模式」/ "Switch to light mode"）
- 语言控件：`aria-label="Language"` / `语言`
- 切换主题时避免布局跳动（仅颜色变量变化）
- 浅色模式下将 `hover:text-white` 等硬编码改为 `hover:text-[var(--foreground)]`（实现时全局排查）

---

## 4. 实现阶段

### Phase 1 — 基础设施 + 公开页（MVP）

- [ ] `globals.css` 增加 `[data-theme="light"]` 变量集
- [ ] 安装并配置 `next-themes`
- [ ] 实现 `I18nProvider` + `en` / `zh-CN` 字典骨架
- [ ] 实现 `ThemeToggle`、`LocaleSwitcher`、`SiteHeader`
- [ ] 替换 `page.tsx`、`models/page.tsx`、`calculator/page.tsx` 的重复 header
- [ ] `/login` 页加入主题 + 语言控件
- [ ] Sign In **左侧** 放置两个控件（验收 §5.1）

### Phase 2 — 全站补齐

- [x] `dashboard`、`playground`、`docs`、`admin`、`status` 接入 `SiteHeader` 或统一右侧控件区
- [x] 扩展字典至 Dashboard / Playground / Admin / Status
- [x] 文档页 `/docs` 导航与章节标题中文化（正文可仍英文）

### Phase 3 — 打磨（可选）

- [ ] 跟随系统主题（`enableSystem`）
- [ ] 首屏无闪烁（`next-themes` 内联 script）
- [ ] 浏览器 `Accept-Language` 作为无 localStorage 时的初始语言

---

## 5. 验收标准

### 5.1 布局

- [ ] 在 `/` 未登录状态下，从左到右顺序为：**主题按钮 → 语言控件 → Sign In**
- [ ] 已登录时，**主题按钮 → 语言控件 → Dashboard / Logout**，位置与未登录一致
- [ ] `/login` 页可见主题与语言控件（无 Sign In 亦可）

### 5.2 主题

- [ ] 默认进入为深色，与当前视觉一致
- [ ] 点击主题按钮可切换到浅色，全页背景/卡片/边框随之变化
- [ ] 刷新页面后主题选择保持不变（`localStorage`）

### 5.3 语言

- [ ] 默认语言为英文（与现网一致）
- [ ] 切换到「中文」后，Phase 1 范围内文案显示为简体中文
- [ ] `document.documentElement.lang` 为 `en` 或 `zh-CN`
- [ ] 刷新页面后语言选择保持不变

### 5.4 质量

- [ ] `npm run build` 通过
- [ ] 无控制台 hydration 报错（或仅可接受的 `suppressHydrationWarning`）
- [ ] 主题/语言切换不触发登出、不影响 API 请求

---

## 6. 风险与缓解

| 风险 | 缓解 |
|------|------|
| 多处 header 漏改 | 强制通过 `SiteHeader` 单点维护 |
| 浅色模式对比度不足 | 用 §2.4 色板做一轮对比度检查（正文/-muted/边框） |
| Hydration 主题/语言闪烁 | `suppressHydrationWarning` + 尽早读 storage |
| 硬编码 `hover:text-white` 在浅色下不可见 | Phase 1 收尾时 grep 替换为 CSS 变量 |
| 文案遗漏 | 字典 key 按页面分组；PR 检查清单对照 §3.4 |

---

## 7. 完成定义（Definition of Done）

- [ ] Phase 1 全部 checkbox 完成
- [ ] §5 验收标准满足
- [ ] `HANDOFF.md` 或 `web/README` 补充一句：主题/语言偏好存于浏览器本地

---

## 8. 预估工作量

| 阶段 | 人天（1 名熟悉 React/Next 开发者） |
|------|-----------------------------------|
| Phase 1 | 1.5–2 |
| Phase 2 | 1–1.5 |
| Phase 3（可选） | 0.5 |
| **合计** | **约 2.5–4 人天** |

---

## 附录 A — 字典 key 示例（节选）

```ts
// messages/en.ts
export const en = {
  nav: {
    models: "Models",
    playground: "Playground",
    docs: "Docs",
    signIn: "Sign In",
    dashboard: "Dashboard",
    logout: "Logout",
  },
  theme: {
    switchToLight: "Switch to light mode",
    switchToDark: "Switch to dark mode",
  },
  locale: {
    en: "EN",
    zh: "中文",
  },
  home: {
    heroTitle: "One API for All Leading AI Models",
    // ...
  },
};

// messages/zh-CN.ts
export const zhCN = {
  nav: {
    models: "模型",
    playground: "Playground", // 或保留英文专名
    docs: "文档",
    signIn: "登录",
    dashboard: "控制台",
    logout: "退出",
  },
  // ...
};
```

---

## 附录 B — 与 01 后端重构的关系

无依赖。可并行开发；不要求后端改动。
