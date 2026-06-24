# Mermaid 图表示例

下面的代码块会自动渲染为 Mermaid 图表。

```mermaid
flowchart LR
  A[编写 Markdown] --> B{代码块语言是 mermaid?}
  B -- 是 --> C[渲染为图表]
  B -- 否 --> D[保持普通代码块]
  C --> E[Docsify 预览]
```

