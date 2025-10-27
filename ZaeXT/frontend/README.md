# AI Conversation Frontend

生产可用的 Vue 3 + TypeScript 单页应用，面向现有的 Go + Gin 后端 (`/api/v1`)。项目聚焦高可维护性与稳健用户体验，实现 AI 问答场景所需的鉴权、会话管理、SSE 流式聊天、模型权限约束、分类树、回收站、用户长期记忆维护以及中英双语界面等核心能力。

## 技术栈

- **Vue 3 + `<script setup>`**：组合式 API，配合 **Pinia** 管理全局状态
- **Vue Router 4**：路由与守卫，登陆态校验 & 进度条提示
- **TypeScript 严格模式**：DTO / 领域模型与 store 类型全覆盖
- **Tailwind CSS + 自定义tokens**：响应式、深浅主题与无障碍对比
- **Markdown-it + highlight.js**：渲染助手消息、代码高亮与复制按钮
- **Axios**：统一 HTTP 客户端，封装错误处理、鉴权头与 Toast
- **SSE 工具**：基于 Fetch ReadableStream 的流解析、重试、AbortController
- 工程化：Vite、Vitest、ESLint、Stylelint、Prettier、vue-tsc

## 功能亮点

- **身份认证**：注册、登录、JWT 本地持久化；路由守卫自动拉取 `/profile`
- **聊天体验**：
  - SSE 打字机式回复，支持中断、错误提示与复制代码
  - 消息列表采用常规列表布局，用户消息右对齐、助手消息左对齐，保证刷新/跳转稳定
  - 输入区模型选择、启用思考链开关、停止生成按钮
- **会话管理**：新建、重命名（支持空字符串触发自动命名）、搜索、分类归档、自动分类触发、回收站软删除/恢复/永久删除
- **模型目录**：根据用户 tier 过滤展示，不可用模型显示禁用提示
- **分类树**：树形 CRUD、级联删除，与聊天信息面板联动
- **用户资料**：查看 tier / 创建时间，编辑 `memory_info` 长期记忆字段
- **多语言 + 主题**：内置 zh-CN / en-US 文案切换、暗黑/浅色模式按钮
- **错误与空态**：统一加载指示、错误提示、空态组件

## 目录结构

```
frontend/
├─ public/
├─ src/
│  ├─ api/              # axios 实例、REST 封装、SSE 工具
│  ├─ components/       # 全局组件（Markdown 渲染、空态、设置面板等）
│  ├─ features/         # 领域组件（auth / chat / models / categories / recycle-bin）
│  ├─ views/            # 页面级视图，对应路由
│  ├─ stores/           # Pinia stores（auth / conversations / models / categories / recycleBin）
│  ├─ router/           # 路由定义、守卫、nprogress
│  ├─ utils/            # jwt、fetch、通知、时间格式化、i18n 助手
│  ├─ i18n/             # 语言资源与配置
│  ├─ styles/           # Tailwind 入口与全局样式
│  └─ types/            # DTO、领域模型、响应类型
├─ .env.example         # 前端环境变量模版
├─ vitest.config.ts     # 测试配置
└─ 各类工程配置         # Vite / Tailwind / ESLint / Stylelint / Prettier / tsconfig
```

## 环境准备

1. **安装依赖**

   ```bash
   npm install
   # 或 pnpm install / yarn install
   ```

2. **配置环境变量**

   ```bash
   cp .env.example .env
   # 视情况调整 VITE_API_BASE（默认 http://localhost:8080/api/v1）
   ```

3. **启动开发服务器**

   ```bash
   npm run dev
   ```

   Vite 默认监听 `http://localhost:5173`。请确保后端已按 `backend_README.md` 启动，并可通过 `VITE_API_BASE` 访问。

## 常用脚本

| 命令 | 说明 |
| --- | --- |
| `npm run dev` | 启动 Vite 开发服务器（含 HMR） |
| `npm run build` | 执行 `vue-tsc -b` 类型检查并打包生产构建 |
| `npm run preview` | 本地预览生产构建 |
| `npm run type-check` | TypeScript 单独类型检查（无输出） |
| `npm run lint` | ESLint（整合 Prettier）检查 `.ts/.vue` |
| `npm run lint:styles` | Stylelint 检查全局 & 组件样式 |
| `npm run test` | Vitest 单测（覆盖 HTTP 客户端与 SSE 解析） |
| `npm run coverage` | 生成 Vitest 覆盖率报告 |

建议在提交前运行质量门禁：

```bash
npm run lint && npm run lint:styles && npm run type-check && npm run test
```

## 与后端联调

1. 后端需按照 `backend_README.md` 启动并开放以下关键端点：
   - `POST /register` / `POST /login`
   - `GET /profile`, `PUT /profile/memory`
   - `GET /models`
   - `POST /conversations`、`GET /conversations`、`GET /conversations/:id/messages`
   - `POST /conversations/:id/messages`（SSE 流，`delta.content` 增量 + `[DONE]` 终止）
   - `PUT /conversations/:id/title|category`、`POST /conversations/:id/auto-classify`、`DELETE /conversations/:id`
   - 分类与回收站相关端点
2. 确保跨域设置允许前端地址，并配置 `Authorization: Bearer <token>` 头。
3. 前端 `.env` 的 `VITE_API_BASE` 指向上述 `/api/v1` 根路径。

## 设计与架构要点

- **KISS / YAGNI**：界面与逻辑保持简洁；消息列表采用常规实现，避免过度抽象导致刷新异常。
- **SOLID / DRY**：请求封装、类型定义与 store 分层，复用通用组件（空态、加载、Markdown 渲染）。
- **可观测性**：HTTP 客户端统一抛错并触发 Toast；SSE 客户端暴露状态回调，方便调试与 UI 反馈。
- **响应式体验**：移动端显示会话入口/回收站入口；消息列表、输入区在不同视口下保持可用性。
- **国际化 / 主题**：内置中英双语与深浅模式按钮，局部文案存放在 `src/i18n/locales/`。

## 测试现状

- 最新执行：
  - `npm run test` → 2 个测试文件 / 8 个用例全部通过
  - `npm run lint` → 无 lint 错误（Prettier 已修复 `AppMarkdownRenderer.vue` 尾部格式）
- 可按需扩展 Playwright E2E 测试覆盖登录、发消息、回收站等关键流程。

## 维护建议

- 若新增接口，请先在 `src/types` 编写 DTO，再在 `src/api` 封装请求，最后在对应 store/feature 中使用。
- 扩展国际化时同步更新 `src/i18n/index.ts` 的语言列表，并补充 UI 切换入口。
- Safari 等浏览器下的 `vue3-resize-observer` 白块问题，已通过 `src/styles/main.css` 中的全局样式隐藏，修改布局时注意保留。

欢迎继续迭代，祝开发顺利！
