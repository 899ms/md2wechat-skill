# 真实烟雾测试记录

本文档记录最近一次带真实配置、真实外部服务的最小闭环验证结果。

最近更新：

- 日期：2026-08-28
- 环境：本地配置文件 `~/.config/md2wechat/config.yaml`
- 目标：验证安装、配置、确认层、排版、图片、微信上传、草稿创建是否真实可用

> 说明：本文档只记录验证结论和关键观察，不记录敏感凭证、草稿 ID、素材 ID。

> v3.3.0 release boundary: Gate B passed against `https://www.md2wechat.cn/api/convert`; the run was observed at 2026-08-28T03:41:50Z and its report completed at 2026-08-28T03:41:51Z. Evidence `/tmp/md2wechat-layout-conformance-v3.3.0-a9e0a08-main-deployed.jsonl` binds CLI commit `a9e0a08870aa1f55f995b32f2fc554da48ca87dd` to upstream main `0e7027616dd1654802cf11615f6ba8bd23e539ae`, with all 84 individual witnesses and all six compact-theme probes green. The endpoint exposed no recognized remote build identity, so this is target-and-time production evidence, not remote commit evidence; local `layout validate` alone remains insufficient. Historical smoke evidence is not rewritten by this note.

---

## 测试前提

已具备：

- 可读取的 `~/.config/md2wechat/config.yaml`
- 有效的 `WECHAT_APPID` / `WECHAT_SECRET`
- 有效的 `MD2WECHAT_API_KEY`
- 可用的图片服务配置
- 微信接口白名单已放通当前执行机 IP

### 高级排版 API 一致性

高级排版 conformance 默认离线跳过，不属于 `go test ./...` 的外部网络前提。显式执行：

```bash
make e2e-layout
```

该目标与 API 模式 `convert` 共用同一端点解析：默认完整 URL 是 `https://www.md2wechat.cn/api/convert`，并对 `MD2WECHAT_BASE_URL` / 配置文件值统一补全 `/api/convert`。发布模式（默认）必须同时提供已通过的上游字段契约证据：`MD2WECHAT_UPSTREAM_FIELD_CONTRACT_SHA=0e7027616dd1654802cf11615f6ba8bd23e539ae` 与 `MD2WECHAT_UPSTREAM_FIELD_CONTRACT_RESULT=passed`；缺任一项或 SHA 不匹配即失败。仅本地或 staging 冒烟可显式设置 `MD2WECHAT_LAYOUT_CONFORMANCE_MODE=smoke`，并可用 `MD2WECHAT_BASE_URL` 覆盖目标。凭证从 `~/.config/md2wechat/config.yaml` 的 `api.md2wechat_key` 读取，`MD2WECHAT_API_KEY` 可覆盖配置文件。API key 只进入 `X-API-Key` 请求头，不会出现在命令行、JSONL 报告或测试日志中。

报告默认写入 `/tmp/md2wechat-layout-conformance.jsonl`，可通过 `LAYOUT_CONFORMANCE_OUTPUT` 改路径。脚本为 Go 测试设置六分钟总时限；请求串行执行，且仅网络/5xx 短暂失败会最多重试两次。它会运行 84 个 witness conformance，以及一个覆盖 `default`、`apple`、`cyber`、`bytedance`、`sports`、`chinese` 的紧凑边界/组合探针。84 个 witness 由 56 个 canonical、25 个结构不同的 non-default branch 和 3 个 compatibility witness 组成；任一 module marker、稳定正文、精确 variant 分支属性、语义 DOM 约束或原始 fence 不一致都会失败。

可选设置 `MD2WECHAT_API_BUILD_ID` 锁定预期部署版本；设置后每个响应都必须携带 recognized build identity header 且精确匹配。未设置时仍要求本次 84-witness run 内所有响应 identity 一致；若均无 header，只记录 target 与 UTC 观测时间，并明确标记为非 commit 证据。失败类别区分 authentication、API drift 与 network failure。JSONL 测试日志中的 `conformance_target_normalized` 是实际请求的规范化 API URL；脚本会把该字段名作为 target evidence label 输出。

发布时保留 JSONL 报告和测试日志作为目标 API 证据；不要把本地 `layout validate` 或常青文档中的历史运行结果当作远端部署证明。

执行过的基础自检：

```bash
md2wechat config validate --json
md2wechat config show --format json
md2wechat skills read md2wechat --json
```

## v3.6.0 发布前检查（2026-09-12）

- `make quality-gates` 通过，版本资料统一为 3.6.0。无效 `sync` 子命令已验证返回失败；正常准备仍返回 `action_required`。
- 独立代码审查发现仅检查图片头会误收损坏图片，已改为完整解码并验证：只有有效文件头的截断 PNG 返回失败，且不创建输出目录。修复后独立复审无剩余问题，完整质量检查通过。
- 同一候选批次完成五个平台编译。macOS ARM64 程序、压缩包、npm 离线安装和 shell 安装均实际运行；版本、内置 Skill 与四份跨平台说明、本地准备及无效命令输出符合预期。
- Homebrew 配方的版本化下载地址和四份归档校验值与同批文件一致；Linux 两份程序确认静态链接。Linux、Windows 和 macOS Intel 未在对应系统运行，须由发布 CI 和对应环境继续验证。
- 本轮未变更排版渲染或转换逻辑，未重复执行远端排版验证。三平台浏览器验证范围沿用下方样本记录；agent-browser 仍未完成三平台验证。
- 变更文件扫描命中的账号密码形式均为测试用例中的占位输入，未发现真实凭证或已知私人草稿地址；本地计划继续被忽略。

## 腾讯云开发者社区验证（2026-09-19）

- 环境：macOS、Chrome 已登录会话，新版富文本编辑器；仅保存未发布测试草稿。
- 样本：短正文、H2/H3、单项列表、代码 `hello world`、两列两行表格和结束标记。正常 HTML 粘贴后点击“存草稿”，出现保存提示并取得带 `draftId` 的地址；重新加载同一地址后逐项内容与层级保留。页面字数曾显示 0，不能作为正文丢失或完整的依据。
- 图片复测：浏览器直接选文件返回 `Not allowed`，不能据此认定腾讯云拒绝上传。在用户要求测试系统窗口后，通过 Chrome 原生“插入 → 图片”打开 macOS 文件窗口，以 `Command+Shift+G` 定位仓库展示图并点“打开”，上传成功；未修改扩展权限。保存后重新打开同稿，单张图片为 860×2320、加载完成，位置仍在 H2 与首段之间，全文结构保留。
- 路径差异：读取已上传图片地址并正常整体粘贴图文后，图片虽加载但组件持续显示“图片载入中”，保存提示“文章中有载入失败或上传中的文件，无法操作”。撤回该次粘贴，恢复原生上传图片后保存重开通过。因此腾讯云使用先正文后原生插图流程，不能照搬其他平台的图片地址回填方式。多图及重复图尚未独立验证。
- 约一万字长文、从草稿箱找回未记录地址的稿件尚未验证；不宣称腾讯云已通过原三平台全部样本。
- 私人账号、草稿编号与图片地址不写入仓库。

## 三平台草稿真实门槛

本地测试通过不能替代编辑器验证。经授权在知乎、CSDN、头条的指定账号上只保存测试草稿，复用已登录会话。样本位于 `examples/sync/`：短图文、代码/表格/列表/多级标题、约一万字长正文。图片上传使用正常编辑器流程，不能调用内部接口。

```bash
md2wechat sync prepare examples/sync/editor-smoke.md --output ./smoke-prepared --json
md2wechat skills read md2wechat references/sync/workflow.md --json
```

每个平台须验证标题、逐段全文、图片顺序、代码与表格，保存后重开同一唯一地址，并覆盖一次中断后继续同稿。只看首尾或保存提示不够；有丢失就修复流程或明确缩小支持范围，不能登记为成功。本文件记录实际日期、样本与结果；平台说明只保留执行限制，不提交私人账号、草稿地址或临时图片链接。

头条结构样本发现两个限制：H2/H3 统一为 H1，🧪 丢失。混合标题级别在写入前停止；单一级别允许已知归一化差异。🧪 丢失只证明此样本未完整保留，不据此排除全部补充平面字符；逐字符核验失败时不能报告成功。

完整执行与恢复规则见 [SYNC.md](SYNC.md)。

2026-09-12 在 macOS、Chrome 已登录会话中的实测结果：

| 样本 | 知乎 | CSDN | 头条 |
| --- | --- | --- | --- |
| 短正文与本地图片 | 上传、保存、重开通过 | 上传、保存、重开通过 | 上传、保存、重开通过；标题归一化 |
| 代码、表格、列表、引用、多级标题 | 内容与层级保留 | 内容与层级保留；同稿修正标题后通过 | 代码、表格、列表保留；层级与 🧪 丢失，超出支持范围 |
| 约一万字、60 段、6 个同级标题 | 重开后全文比较一致 | 重开后全文比较一致 | 重开后全文比较一致；标题归一化 |

已有地址后中断并恢复到同稿已验证。不同宿主仍需能力检查；独立 Agent 的浏览器连接复验因宿主连接/审批超时未完成，不能据此声称所有 Agent 均已验证。agent-browser 尚未完成三平台富文本、图片和重开闭环验证。操作说明的内置读取另由本地测试和实际安装验证保护。

---

## 通过的真实链路

### 0. 确认层：inspect / preview

命令形态：

```bash
md2wechat inspect article.md --json
md2wechat preview article.md --json
```

结果：

- `inspect` 能返回最终 title / author / digest 来源、`data.readiness`、`data.checks`；`data.readiness.targets/blockers` 能表达目标是否 blocked 以及对应原因
- 真实验证到 `TITLE_BODY_MISMATCH`、`DIGEST_METADATA_ONLY`、`IMAGE_REPLACEMENT_REQUIRES_UPLOAD_OR_DRAFT`
- `preview` 在转换成功时写入字节一致的 converter HTML；AI handoff、API 失败和空结果在本次调用中均不新建或覆盖 HTML，既有显式输出不得冒充本次结果
- `convert` 只有在显式请求 `--upload` / `--draft` 时进入远程副作用，并在任何远程调用前拒绝已知无效的 draft intent、cover 或本地资产

结论：

- confirm-first 路径真实可用
- 确认层没有用 dashboard、错误页或 Markdown fallback 伪造最终视觉结果

### 1. 图片生成

命令：

```bash
md2wechat generate_image --json "minimal smoke test banner..."
```

结果：

- 外部图片服务返回生成结果
- 生成图下载成功
- 微信素材上传成功

结论：

- `generate_image` 真实闭环通过

### 2. 上传本地图片

命令：

```bash
md2wechat upload_image --json /tmp/md2wechat-real-smoke.png
```

结果：

- 微信素材上传成功

结论：

- `upload_image` 真实闭环通过

### 3. 创建图片帖子

命令：

```bash
md2wechat create_image_post --json \
  -t "Smoke Test" \
  --images /tmp/md2wechat-real-smoke.png
```

结果：

- 图片上传成功
- 微信图片消息 / newspic 草稿创建成功

结论：

- `create_image_post` 真实闭环通过

### 4. 从 Markdown 创建图片帖子

验证了两种输入：

- 本地图片
- 远程图片 URL

结果：

- 两种输入都能进入统一资产链
- 远程图片会先下载，再上传微信，再创建图片帖子

结论：

- `create_image_post --from-markdown` 本地图 / 远程图均通过

### 5. 创建普通草稿

命令：

```bash
md2wechat test-draft --json draft.html cover.png
```

结果：

- 封面上传成功
- 微信普通草稿创建成功

结论：

- `test-draft` 真实闭环通过

### 5.1 草稿错误码提示

命令形态：

```bash
md2wechat create_draft oversize-digest.json --json
```

结果：

- 真实触发微信 `errcode=45004`
- 返回 `description size out of limit`
- CLI 额外补充字段级 hint，明确先检查 `--digest` / frontmatter `digest` / `summary` / `description`

结论：

- `45004` 不再被误导成“先缩正文”
- draft 错误提示已能指导用户和 Agent 走正确排障路径

### 6. API 模式转换并创建草稿

命令形态：

```bash
md2wechat convert \
  --json \
  --mode api \
  --upload \
  --draft \
  --cover cover.png \
  --output output.html \
  article.md
```

结果：

- `md2wechat.cn` API 排版成功
- 文内本地图片上传成功并回填
- 封面上传成功
- 草稿创建成功
- HTML 文件成功写出

结论：

- `convert --mode api --upload --draft --output` 真实闭环通过

### 7. AI 模式动作输出

命令形态：

```bash
md2wechat convert \
  --json \
  --mode ai \
  --theme autumn-warm \
  --output result.html \
  article.md
```

结果：

- 返回 `status=action_required`
- 返回 `code=CONVERT_AI_REQUEST_READY`
- 实际写出的是 `result.prompt.txt`
- 不再把 prompt 错写成 `.html`

结论：

- AI 模式当前是“生成 AI request / prompt”的半闭环
- 输出契约已经修正，不会伪造 HTML 成功

---

## 测试中发现的重要现象

### 1. 微信对白名单敏感

如果当前执行机 IP 不在微信公众号接口白名单中，上传图片和建草稿都会失败。  
这不是代码问题，是微信接口前置条件。

### 2. 极小测试图可能被微信拒绝

1x1 PNG 测试图曾被微信返回：

```text
unsupported file type
```

换成正常尺寸 PNG 后上传成功。  
结论是：真实 smoke 不要使用异常小图作为上传样本。

### 2.1 `--json` 契约已收口

当前 `inspect --json`、`preview --json` 等命令已验证：

- stdout 只输出 JSON
- JSON 是单行紧凑对象并以最终换行结束；人工检查时用 `jq` 格式化
- 结构化输出不再混入配置 banner
- `inspect --json` 的 Agent 决策字段位于 `data.readiness.targets/blockers`
- `skills read md2wechat --json` 能读取当前二进制内置 SOP，避免 smoke Agent 依赖旧 README 或旧外部 skill

这让 Agent / 脚本可以直接解析 stdout，而不用先清洗杂音。

### 3. AI 模式不是“自动完成 HTML 排版”

当前 CLI 下的 AI 模式语义是：

- 准备 prompt / request
- 交给外部 AI 继续完成

所以它应被视为 `action_required`，不是 `completed`。

---

## 当前建议的真实验证顺序

如果你要在新环境复现这组 smoke，按这个顺序最稳：

1. `md2wechat config validate --json`
2. `md2wechat upload_image --json <normal-image.png>`
3. `md2wechat create_image_post --json ...`
4. `md2wechat test-draft --json ...`
5. `md2wechat convert --mode api --upload --draft --cover ... --output ...`
6. 最后再测 `md2wechat convert --mode ai --json`

---

## 相关文档

- [配置指南](CONFIG.md)
- [安装指南](INSTALL.md)
- [OpenClaw 指南](OPENCLAW.md)

## v3.7.0 发布准备检查（2026-09-23）

- 版本来源、npm 包、插件资料和更新记录统一为 3.7.0；当前安装文档中的固定版本链接已同步，历史 3.6.0 记录保持不变。尚未打标签或发布，远端 3.7.0 下载链接尚不可作为已发布产物使用。
- `GOCACHE=/tmp/md2wechat-go-build make quality-gates` 通过，最终输出 `release-check: OK (version 3.7.0)` 和 `quality-gates: OK`。
- 使用同一个本机构建包验证原始程序、压缩包解压、shell 安装器和 npm 离线安装；四条路径均返回 3.7.0，六份写作指引与源码一致。安装仅发生在临时目录，未替换现有安装。
- 此次安装验证限于本机 macOS，不代表 Windows、Linux 或真实 OpenClaw 宿主已经验证；全平台发布流程仍须执行各平台检查。本轮没有渲染能力变更，未运行 API 渲染实测。

## 2026-09-23 定向写作开发版本验证

本次验证基于 `342cc66` 的未提交工作区，版本号仍为 3.6.0；不是新版本已经发布的记录。新增内容是宿主写作指引，未增加命令、模型接入或自动发布能力。

- 资料：`references/writing/` 下的 `workflow.md`、`forms.md`、`platforms.md`、`search-strategies.md`、`baidu-baike.md`、`toutiao-baike.md` 均可由当前构建读取。两份技能入口指向同一份内嵌资料；在脱离源码的临时目录运行，返回路径及全文与源文件一致。缺失路径明确失败。这是分发模拟，不是完整安装器测试。
- 回归：`GOCACHE=/tmp/md2wechat-go-build go test ./cmd/md2wechat -run '^TestSkills' -count=1`、`GOCACHE=/tmp/md2wechat-go-build go test ./internal/skillcontent -count=1` 通过。最终 `GOCACHE=/tmp/md2wechat-go-build make quality-gates` 通过，包含格式、vet、固定版本 lint、全量测试、npm 内容检查及 release-check。初次受缓存权限及工具下载连接错误影响，获准重试后通过，未跳过检查。
- 实际写作：原生 Codex 宿主执行 C01–C20 的二十类场景，含七个搜索目标、五个文章平台和三类体裁子例；另完成项目自身介绍。三组内部共享上下文，不视为逐例独立盲测。正文已逐份核对；未把通用写法包装成模型专属效果。
- 问题及修复：旧入口会把含糊的“头条百科式”直接写成文章；新增指引后，以完全相同请求在新上下文运行，先辨明文章或词条。组合稿首次标题把作者的担忧强化为反复发生的经历，补充标题事实核对要求后，重新完成组合稿及公众号、头条号两稿，未再出现该问题。
- 保护边界：三份原始材料的前后摘要一致；实际写作未上传、提交百科、保存远端草稿或发布。当前构建的 `capabilities --json` 与变更前相同。旧安装程序读取写作路径实际返回 `CONFIG_INVALID`。另以新上下文宿主配合真实 3.5.0 安装运行，宿主报告版本不匹配及尚未生成文章，未改读源码资料、切换程序或执行升级，没有静默假装支持。
- 未验证：没有可用 OpenClaw 程序，未运行其真实宿主；只核验该入口与资料读取。仅排版场景保留原文并完成本地示意及语法检查，未调用付费 API；本地页面视觉检查被浏览器地址策略拒绝，不能称为最终公众号渲染已验证。本次不涉及渲染能力变更。
- 效果范围：未开展真实客户试用、目标搜索自然发现/引用观察或付费转化实验。头条百科完整当前官方编写细则仍未核实；该路径交付有明确限制的词条草稿，不承诺审核通过。

使用方法见 [定向产品写作](WRITING.md)。
