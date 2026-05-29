# MDOnline

> 零构建的 Markdown 文档站点解决方案，让你专注写作。

## 特性一览

- **零构建**：纯静态文件，无需 Webpack / Vite / Node.js
- **双主题**：暗黑 / 浅色一键切换，偏好自动持久化
- **智能侧边栏**：Python 脚本自动生成，支持分组折叠与顺序控制
- **全文搜索**：内置中文搜索插件，即时定位内容
- **子路径兼容**：自动适配根目录与 Nginx 子路径部署

## 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/yourname/mdonline.git
cd mdonline

# 2. 启动预览
python -m http.server 8080

# 3. 打开浏览器访问
# http://localhost:8080
```

> Windows 用户可直接双击 `start-server.bat` 一键启动。

## 目录结构

```
MDOnline/
├── index.html          # 入口文件（Docsify 配置）
├── style.css           # 双主题样式
├── vue.css             # Docsify 基础主题
├── favicon.svg         # 站点图标
├── docsify.min.js      # Docsify 核心
├── search.min.js       # 全文搜索插件
├── gen_sidebar.py      # 侧边栏自动生成脚本
├── start-server.bat    # Windows 一键启动
└── docs/               # 文档内容目录
    ├── 快速开始/        # 入门指南 & 安装部署
    ├── 功能特性/        # 主题切换 & 侧边栏 & 搜索
    ├── 写作指南/        # Markdown 语法 & 最佳实践
    └── 部署运维/        # Nginx & Vercel & FAQ
```

## 文档导航

| 分类 | 文档 | 说明 |
|------|------|------|
| 快速开始 | [入门指南](/docs/快速开始/入门指南) | 3 分钟搭建文档站点 |
| 快速开始 | [安装与部署](/docs/快速开始/安装与部署) | 多平台部署指南 |
| 功能特性 | [主题切换](/docs/功能特性/主题切换) | 暗黑/浅色双主题 |
| 功能特性 | [智能侧边栏](/docs/功能特性/智能侧边栏) | 自动生成导航 |
| 功能特性 | [全文搜索](/docs/功能特性/全文搜索) | 中文全文检索 |
| 写作指南 | [Markdown 语法速查](/docs/写作指南/Markdown语法速查) | 常用语法汇总 |
| 写作指南 | [写作最佳实践](/docs/写作指南/写作最佳实践) | 文档撰写规范 |
| 部署运维 | [Nginx 部署](/docs/部署运维/Nginx部署) | Nginx 配置详解 |
| 部署运维 | [Vercel 部署](/docs/部署运维/Vercel部署) | 一键部署到 Vercel |
| 部署运维 | [常见问题](/docs/部署运维/常见问题) | FAQ 故障排查 |

## 技术栈

| 技术 | 用途 |
|------|------|
| Docsify | Markdown 文档站点引擎 |
| CSS Variables | 双主题切换实现 |
| localStorage | 主题偏好持久化 |
| Python | 侧边栏自动生成脚本 |

## 许可证

MIT License
