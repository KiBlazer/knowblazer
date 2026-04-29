# Knowblazer 产品需求文档

日期：2026-04-29

## 1. 背景与机会

AI Coding 工具正在从单一 IDE 插件变成多工具、多终端、多模型协作的日常工作流。开发者可能同时使用 Claude Code、Codex、Gemini、Cursor、SpecStory 等工具，但这些工具之间的长期记忆并不互通。

真实问题不是“AI 能不能记住东西”，而是：

> 开发者如何拥有一套自己可控制、可迁移、可审计、可被不同 AI Coding 工具使用的工程记忆。

现在是做 Knowblazer 的合适时机，原因有四个：

- AI Coding 工具多样化，个人工程经验开始被分散到不同工具和会话中。
- SpecStory 等项目让 AI 对话捕获逐渐成熟，但原始对话不等于长期工程记忆。
- 开发者已经熟悉 Git、Markdown、dotfiles、private repo 这类本地优先工作流。
- memory server 和 agent memory backend 赛道拥挤，但“私有 Git 工程记忆库”仍然是一个更窄、更清晰的机会。

Knowblazer 要抓住的是这个窄机会：让开发者把 AI Coding 过程中产生的经验、偏好、项目背景和踩坑记录沉淀到自己的私有 Git 仓库中，并能在需要时召回给不同 AI Coding 工具使用。

## 2. 产品定位

Knowblazer 的定位是：

> Private Git-backed engineering memory for AI coding.

中文表达：

> Knowblazer 是一个开源工具，用来维护你自己的私有 Git 工程记忆库，让 AI Coding 工具可以安全复用你的工程经验。

产品边界必须清楚：

- Knowblazer 是工具、规范、模板和 adapter。
- 用户记忆库默认保存在用户自己的本地目录或私有 Git remote。
- Knowblazer 第一阶段不提供官方云服务，不托管用户个人工程记忆。
- Markdown/Git 是主存储；索引、检索、MCP、向量库都只能是后续增强层。

一句话区别：

```text
SpecStory remembers AI coding conversations.
Knowblazer turns selected conversations, notes, and lessons into a private engineering memory repo.
```

## 3. 目标用户

第一阶段只服务个人开发者，不服务团队知识库和企业协作平台。

### 3.1 首批用户画像

**多 AI 工具切换的独立开发者**

这类用户同时使用 Claude Code、Codex、Cursor、Gemini。每个工具都知道一点上下文，但没有一个地方能沉淀稳定的个人工程偏好和项目经验。他们需要一个工具无关的私有记忆库。

**多项目维护者**

这类用户维护多个项目，经常需要在不同仓库、服务器、部署流程和技术栈之间切换。他们的问题不是没有文档，而是项目上下文、部署坑、操作禁忌、决策原因散落在不同地方。

**隐私敏感的工程师**

这类用户可能处理客户项目、内部系统、服务器信息或数据库连接。相比云端 memory 服务，他们更愿意把工程记忆放在自己的 private Git repo、自建 Git 或纯本地目录中。

### 3.2 非目标用户

第一阶段不面向：

- 非技术个人知识管理用户。
- 团队知识库管理员。
- 企业合规知识库平台。
- 想要完整 AI 对话云同步和分享的人。
- 想要无 Git 基础、纯图形化知识库的人。

## 4. 核心问题

Knowblazer 第一阶段要解决四个问题。

### 4.1 经验没有稳定归宿

开发经验可能散落在 AI 对话、本机分析目录、项目 README、Cursor 规则、Claude 记忆、Codex 会话里。换机器、换工具、换项目后，这些经验难以继续使用。

### 4.2 AI 工具不能共享长期记忆

Claude Code 知道的偏好，Codex 不一定知道。Cursor 项目规则中的上下文，Gemini CLI 不一定能读取。开发者需要反复解释自己的工程偏好、项目背景和注意事项。

### 4.3 自动保存存在隐私风险

AI Coding 会话里可能出现 token、password、private key、数据库连接串、内部域名、客户信息等内容。不能简单把所有对话自动保存并推送到远端。

### 4.4 原始对话不等于长期记忆

完整 transcript 太长、太杂、噪声太多。真正有长期价值的是经过筛选的偏好、决策、项目背景、踩坑经验、排障方法和操作约束。

## 5. MVP 用户路径

MVP 的主线不是“支持很多功能”，而是让一个开发者在 10 分钟内跑通第一条私有工程记忆路径。

目标路径：

```text
install Knowblazer
  ↓
init a private memory repo
  ↓
capture one Markdown lesson
  ↓
scan for secrets
  ↓
store in inbox or quarantine
  ↓
promote one reviewed lesson
  ↓
generate one recall pack for an AI coding task
```

这个路径成功，才说明 Knowblazer 的核心假设成立。

## 6. 功能范围与优先级

第一版必须克制。下面的优先级用于指导实现取舍。

### 6.1 P0：必须完成

**init**

初始化一个本地 Knowblazer 记忆库，生成最小目录结构、入口说明和基础策略文件。

**capture**

把一个本地 Markdown 文件保存到记忆库。默认进入 `inbox/YYYY-MM-DD/`，并保留来源、时间、标题等基础元信息。

**scan**

对待保存、待提交或待推送的内容进行敏感信息扫描。发现高风险内容时，默认写入或移动到 `quarantine/`，并阻止自动 commit/push。

**promote**

把 review 后的候选记忆从 `inbox/` 或 `daily/` 提升到长期目录，例如 `experience/`、`projects/` 或 `profile/`。第一版可以是显式命令或清晰的手动流程。

**recall**

根据任务描述生成一份短上下文包。第一版可以基于文件名、目录和关键词匹配，不要求语义向量检索。

### 6.2 P1：应该完成

**sync helper**

提供基于用户自有 Git remote 的同步辅助。Knowblazer 可以提示或封装 `git status`、`git commit`、`git push`，但不提供官方 remote。

**daily**

支持每日短期记录。第一版 recall 可以默认读取今天和昨天的 daily 文件。

**project mapping**

支持把当前工作目录映射到 `projects/` 中的项目记忆文件，让 recall 能优先使用当前项目上下文。

### 6.3 P2：暂缓

以下功能不进入第一版核心交付：

- Claude Code hook 自动安装。
- Codex、Gemini、Cursor adapter。
- SpecStory `.specstory/history/` 导入。
- MCP server。
- 向量检索、SQLite 索引、hybrid search。
- 自动整理或类似 OpenClaw dreaming 的后台 promotion。
- Web UI。
- 团队共享。
- 官方云同步。

这些功能可以进入后续路线图，但不能阻塞 P0。

## 7. 记忆模型

Knowblazer 需要分层管理记忆。分层的目的不是增加复杂度，而是防止长期记忆变成杂乱的聊天记录归档。

### 7.1 第一版最小目录

第一版默认创建以下目录：

```text
knowblazer-notes/
├── AI-SETUP.md
├── inbox/
│   └── YYYY-MM-DD/
├── daily/
│   └── YYYY-MM-DD.md
├── profile/
│   ├── preferences.md
│   └── decision-principles.md
├── projects/
├── experience/
│   ├── deployment/
│   ├── frontend/
│   ├── backend/
│   ├── ai-tools/
│   └── operations/
├── system/
│   ├── memory-policy.md
│   └── privacy-policy.md
├── quarantine/
└── recall/
```

第一版不默认创建 `relationships/`、`templates/`、`bin/`。这些可以后续增加，避免初始结构过重。

### 7.2 分层规则

`inbox/` 是原始候选层。AI 会话摘要、SpecStory 导入、本地分析文档、人工临时笔记都先进入这里。默认不参与 recall。

`daily/` 是短期工作层。记录今天做了什么、当前任务状态、临时观察和未整理上下文。recall 可以读取今天和昨天的 daily。

`profile/` 是个人稳定偏好层。保存工程偏好、决策原则、写作风格等高稳定内容。

`projects/` 是项目长期上下文层。保存项目背景、架构、部署方式、关键目录、操作约束和常见坑。

`experience/` 是可复用工程经验层。保存部署、前端、后端、运维、AI tools 等经验。

`quarantine/` 是敏感信息隔离层。这里的内容不参与 recall，不自动 commit，不自动 push。

`recall/` 是输出层。保存或临时生成面向具体任务的上下文包。默认可以不纳入 Git。

### 7.3 提升机制

长期记忆不能由 AI 自动无条件写入。

基础流程：

```text
capture / import
  ↓
scan
  ↓
inbox / daily / quarantine
  ↓
review
  ↓
promote
  ↓
profile / projects / experience
  ↓
recall
```

第一版只要求轻量 promotion：移动文件、生成目标路径建议、或修改文档头部状态。后续再考虑自动评分、定期整理和 OpenClaw dreaming 类机制。

## 8. 数据与隐私边界

Knowblazer 必须把数据所有权写成产品原则，而不是实现细节。

### 8.1 用户拥有记忆库

用户的工程记忆属于用户。Knowblazer 不拥有、不托管、不默认收集用户的个人工程记忆。

### 8.2 默认本地保存

所有 P0 功能都必须在无网络环境下可运行。没有 Git remote、没有官方账号、没有 LLM API，也应该能 init、capture、scan、promote 和 recall。

### 8.3 远端由用户选择

跨设备同步只能通过用户自己配置的 Git remote 完成。可选远端包括 GitHub private repo、GitLab private repo、自建 Gitea/Forgejo、公司内部 Git、NAS 或个人服务器。

### 8.4 不做官方托管

第一阶段不提供 Knowblazer Cloud，不提供官方记忆托管服务，也不要求用户登录官方账号。

未来如果出现云服务，也必须是可选增强，不能改变本地优先、用户自有 Git 仓库优先的默认模式。

### 8.5 推送前默认扫描

任何自动提交、自动同步或自动推送流程，都必须先执行敏感信息扫描。扫描发现高风险内容时，默认行为是阻止推送，并提示用户查看和处理。

## 9. Recall 策略

recall 的目标不是搜索整个知识库，而是为当前 AI Coding 任务生成一份短、准、可读的上下文包。

第一版 recall 的输入：

- 任务描述，例如 `deploy new frontend`。
- 当前工作目录。
- 可选项目名。

第一版 recall 的候选来源：

- `profile/` 中稳定且短小的偏好。
- 当前项目对应的 `projects/` 文件。
- 与任务关键词匹配的 `experience/`。
- 今天和昨天的 `daily/`。

默认排除：

- `inbox/`，除非用户显式指定。
- `quarantine/`，永远不参与默认 recall。
- `recall/` 历史输出，除非用户显式指定。

第一版 recall 输出应满足：

- 人类可读。
- 可以直接贴给 Claude Code、Codex、Gemini 或 Cursor。
- 不包含扫描命中的高风险敏感内容。
- 默认长度受控，避免把整个记忆库塞进上下文。

## 10. 关键用户流程

### 10.1 首次使用

开发者安装 Knowblazer 后，在一个空目录运行 init，得到一个可读、可 Git 管理的记忆库。即使不配置远端仓库，也能继续使用本地功能。

### 10.2 保存一次真实经验

开发者完成一次部署排障后，把总结写成 Markdown。Knowblazer capture 后先 scan。扫描通过则进入 `inbox/`；命中敏感内容则进入 `quarantine/` 并阻止自动提交。

### 10.3 提升一条长期经验

开发者 review `inbox/` 中的候选内容，把已确认有价值的经验提升到 `experience/deployment/` 或对应项目文件中。

### 10.4 为当前任务召回上下文

开发者准备让 AI 工具处理一个任务。运行 recall 后，Knowblazer 根据 profile、project、experience、daily 生成一份短上下文包。

### 10.5 新机器恢复

开发者在新电脑上 clone 自己的私有 Knowblazer 记忆库。运行 init 或 doctor 后，AI Coding 工具可以继续读取同一套工程偏好和项目经验。

## 11. 验收标准

P0 版本必须满足以下可测试标准。

### 11.1 init

- 在空目录运行 init 后，生成第一版最小目录结构。
- 生成 `AI-SETUP.md`、`system/memory-policy.md`、`system/privacy-policy.md`。
- 重复运行 init 不应覆盖用户已有记忆内容。

### 11.2 capture

- 输入一个 Markdown 文件后，Knowblazer 将其保存到 `inbox/YYYY-MM-DD/`。
- 保存后的文件保留原始标题或生成可读标题。
- 保存后的文件包含来源路径和 capture 时间。

### 11.3 scan

- 包含 mock token、mock private key、mock password、mock database URL 的文件会被识别为高风险。
- 高风险文件不会被自动 commit 或 push。
- 高风险内容进入 `quarantine/` 或保持本地待处理状态，并给出明确提示。

### 11.4 promote

- 用户可以把一条 inbox 记忆提升到 `experience/`、`projects/` 或 `profile/`。
- promote 后原始文件状态清晰，不会让用户分不清哪些内容已经进入长期记忆。

### 11.5 recall

- `recall --task "<task>"` 能生成一份人类可读上下文。
- 输出包含 profile、当前 project、匹配 experience、近期 daily 中的相关内容。
- 输出默认不包含 `inbox/` 和 `quarantine/`。
- 输出长度受控，不应简单拼接整个仓库。

### 11.6 本地优先

- 在无网络环境下，P0 功能可运行。
- 不配置 Knowblazer 官方账号也能完整使用 P0。
- 不配置 Git remote 也能 init、capture、scan、promote 和 recall。

## 12. 风险与反指标

下面这些现象说明产品方向或设计需要调整。

- 用户只把 Knowblazer 当备份脚本，从不使用 recall。
- `inbox/` 很快堆积，但用户没有 review 和 promote。
- 目录结构让用户困惑，不知道内容应该放在哪里。
- scan 误报太多，导致用户关闭或绕过扫描。
- recall 输出太长，AI 工具仍然抓不到重点。
- recall 输出太泛，不能减少用户重复解释项目背景。
- 用户担心隐私，误以为 Knowblazer 会上传数据到官方服务。
- 用户认为 SpecStory 已经完全覆盖 Knowblazer 的价值。

## 13. 明确不做

第一阶段明确不做：

- 官方云同步。
- 官方记忆托管。
- 团队知识库协作。
- Web UI。
- 完整 AI 对话记录器。
- IDE 插件。
- 通用 memory server。
- 默认 MCP server。
- 向量数据库。
- 自动理解并整理所有历史聊天。
- 自动生成高质量长期记忆。

## 14. 后续路线图

P0 跑通后，可以考虑：

- Codex、Claude Code、Gemini、Cursor adapter。
- SpecStory `.specstory/history/` 导入。
- Git sync helper。
- 更智能的 project mapping。
- 可选 SQLite 或向量索引。
- MCP server。
- 类似 OpenClaw dreaming 的定期整理机制，但必须可 review、可回滚。
- 自托管同步服务或加密备份。

这些扩展不能改变第一原则：用户自己的私有 Git 工程记忆库是主存储，Knowblazer 官方不托管用户记忆。

## 15. 一句话总结

Knowblazer 要做的不是保存所有 AI 对话，也不是托管用户数据，而是帮助开发者把重要的工程经验维护在自己的私有 Git 仓库中，形成可迁移、可审计、可召回的长期工程记忆。
