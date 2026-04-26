# humanize 使用教程

`humanize` 命令把 AI 生成的文章改写成更像真人写的中文。本教程从最简单的用法开始，逐步介绍每个选项。

---

## 目录

- [你需要什么时候用它](#你需要什么时候用它)
- [最简单的用法](#最简单的用法)
- [四种强度模式](#四种强度模式)
- [看看改了什么](#看看改了什么)
- [输出到文件](#输出到文件)
- [配合 write 命令使用](#配合-write-命令使用)
- [JSON 输出（Agent 用）](#json-输出agent-用)
- [从管道读取输入](#从管道读取输入)
- [常见问题](#常见问题)

---

## 你需要什么时候用它

用 AI 写完一篇公众号文章后，文章经常有这些问题：

- 句式过于工整，像翻译腔："不仅……而且……"、"值得注意的是……"
- 用词空泛：深刻、本质上、具有重要意义、标志着
- 每段都有总结句，读起来像报告
- 连续排比，五连、六连
- 开头写"当然！"或"希望这对您有帮助"

`humanize` 会识别这些模式，按你选的力度重写。

---

## 最简单的用法

准备好一个 Markdown 文件（比如 `article.md`），执行：

```bash
md2wechat humanize article.md
```

处理结果直接打印到终端。默认使用 `medium` 强度。

---

## 四种强度模式

用 `--intensity`（或 `-i`）指定强度：

```bash
md2wechat humanize article.md --intensity gentle
md2wechat humanize article.md --intensity medium      # 默认，可省略
md2wechat humanize article.md --intensity aggressive
md2wechat humanize article.md --intensity authentic
```

### gentle — 温和

只处理最明显的问题。原文结构基本不动，只去掉填充短语、过度强调的连接词等。

**适合**：文章本身已经比较自然，只想做轻微清洁。

```bash
md2wechat humanize article.md -i gentle
```

### medium — 中等（默认）

平衡处理。去除明显 AI 痕迹，同时保留合理的表达方式。

**适合**：大多数 AI 生成文章的日常处理。

```bash
md2wechat humanize article.md
```

### aggressive — 激进

深度审查。大幅改写句式结构，最大化去除 AI 痕迹，注入更强的个性。

**适合**：AI 味很重、需要大改的文本。运行后建议人工检查一遍，确认没有改过头。

```bash
md2wechat humanize article.md -i aggressive
```

### authentic — 真实写作

**这个模式和其他三个不同。**

`gentle`、`medium`、`aggressive` 的思路是：找出 24 种 AI 写作痕迹模式，然后减掉它们。`authentic` 不走这条路，它用六个维度的规则重新引导写作质量：用词、句式、语气、内容表达、结构、整体原则。

目标不是"去掉 AI 感"，而是"写得像一个真正会写字的人"：表达具体，语气稳，不装深刻，不刻意煽动。

**适合**：对文字质量要求高，希望从根本上改变腔调，而不只是修补的场景。

```bash
md2wechat humanize article.md -i authentic
```

> `authentic` 的质量评分维度和其他三个略有不同：输出包含六维评分（直接性、节奏、信任度、真实性、精炼度），总分 /50。

---

## 看看改了什么

加上 `--show-changes`（或 `-c`）可以看到：
- 具体改了哪些地方
- 每处修改的类型和原因
- 五维质量评分

```bash
md2wechat humanize article.md --show-changes
```

输出示例（片段）：

```
# 人性化后的文本

[改写后的文章内容]

# 修改说明

共处理 12 处：
- [填充短语] "值得注意的是" → 删除
- [AI词汇] "标志着" → "是"
- [公式化结构] "不仅……而且……" → 改写为直接陈述

# 质量评分

| 维度   | 得分  |
|--------|-------|
| 直接性 | 8/10  |
| 节奏   | 7/10  |
| 信任度 | 8/10  |
| 真实性 | 7/10  |
| 精炼度 | 9/10  |
| 总分   | 39/50 |
```

评分等级：45+ 优秀，35–44 良好，25–34 一般，25 以下建议重新处理。

---

## 输出到文件

加 `-o` 把结果写到文件，方便后续编辑：

```bash
md2wechat humanize article.md -o article-humanized.md
```

同时看修改、输出到文件：

```bash
md2wechat humanize article.md --show-changes -o article-humanized.md
```

---

## 配合 write 命令使用

`write` 命令生成文章时，可以直接加 `--humanize` 标志，在生成后自动做一轮去痕处理，省掉手动再跑一次的步骤：

```bash
# 生成文章，默认 medium 强度
md2wechat write --style dan-koe --humanize article.md

# 指定强度
md2wechat write --style dan-koe --humanize=aggressive article.md
md2wechat write --style dan-koe --humanize=authentic article.md
```

---

## JSON 输出（Agent 用）

加 `--json` 可以得到结构化 JSON，适合在 Claude Code 或其他 Agent 流程中解析：

```bash
md2wechat humanize article.md --json
```

返回结构：

```json
{
  "success": true,
  "status": "action_required",
  "action": "humanize",
  "content": "",
  "prompt": "[完整的提示词，交给 AI 执行]"
}
```

`status` 为 `action_required` 时，说明 CLI 已准备好提示词，需要 Agent 把 `prompt` 字段发给 Claude 执行，然后把 Claude 的响应写回文件。

> Agent 集成时的标准流程：`md2wechat humanize --json` → 解析 `prompt` → 调用 Claude → 把结果写到 `-o` 指定的文件。

---

## 从管道读取输入

`humanize` 支持从 stdin 读取，适合接在其他命令后面：

```bash
cat article.md | md2wechat humanize -
pbpaste | md2wechat humanize - -o cleaned.md
```

---

## 常见问题

**Q：`gentle` 处理完感觉没什么变化？**

A：`gentle` 就是这么设计的——只动最明显的问题。如果文章 AI 味比较重，换 `medium` 或 `aggressive`。

**Q：`aggressive` 改完读起来感觉有点奇怪？**

A：激进模式会大幅改写句式，有时候会偏离原意或改过头。建议跑完之后过一遍，把不合适的部分手动改回来。

**Q：`authentic` 和 `aggressive` 有什么区别？**

A：`aggressive` 是减法：找出 24 种 AI 模式，尽量去掉。`authentic` 是另一套思路：用六维写作规则重写，目标是"像真人"而不是"不像 AI"。两者方向不同，`authentic` 更适合对文章腔调有要求的场景。

**Q：质量评分低是什么意思？**

A：评分基于直接性、节奏、信任度、真实性、精炼度五个维度，总分 50 分。分数低说明文章仍然有明显的 AI 写作特征。可以换更高强度再跑一次，或手动修改。

**Q：可以只处理某类 AI 痕迹吗？**

A：CLI 暂时不暴露 `--focus` 参数，当前版本默认处理全部 24 种模式。如果只想处理特定类别，用 `gentle` 强度可以减少误改。

**Q：文章已经是人写的，还需要 `humanize` 吗？**

A：不需要。`humanize` 是为 AI 生成文章设计的。如果文章本身就是人写的，跑 `humanize` 有可能改出问题。

---

更多命令参考 [USAGE.md](USAGE.md) 或执行 `md2wechat humanize --help`。
