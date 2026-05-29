# Vercel 部署

Vercel 是最简单的部署方式之一，支持从 Git 仓库自动部署，无需服务器配置。

## 部署步骤

### 1. 推送代码到 Git 仓库

确保项目已推送到 GitHub、GitLab 或 Bitbucket：

```bash
git init
git add .
git commit -m "init: MDOnline docs site"
git remote add origin https://github.com/yourname/mdonline.git
git push -u origin main
```

### 2. 导入项目

1. 访问 [vercel.com](https://vercel.com) 并登录
2. 点击 **Add New → Project**
3. 选择你的 Git 仓库
4. 点击 **Import**

### 3. 配置项目

| 配置项 | 推荐值 | 说明 |
|--------|-------|------|
| Framework Preset | **Other** | MDOnline 是纯静态站点 |
| Root Directory | `.` | 保持默认，无需修改 |
| Build Command | 留空 | 零构建项目，无需构建命令 |
| Output Directory | 留空 | 静态文件就在根目录 |

### 4. 部署

点击 **Deploy**，等待部署完成（通常 10-30 秒）。

## 自定义域名

部署成功后，可以绑定自定义域名：

1. 进入项目 **Settings → Domains**
2. 输入你的域名（如 `docs.example.com`）
3. 按提示在域名注册商处添加 CNAME 记录
4. Vercel 自动配置 HTTPS 证书

## 自动部署

Vercel 会监听 Git 仓库的变更：

- 推送到 `main` 分支 → 自动部署到生产环境
- 推送到其他分支 → 自动创建预览部署
- Pull Request → 自动生成预览链接

## 环境变量

MDOnline 不需要环境变量。如果你的项目需要区分环境，可以在 **Settings → Environment Variables** 中配置。

## 常见问题

### `.md` 文件返回 404

Vercel 默认能正确处理 `.md` 文件。如果遇到问题，在项目根目录创建 `vercel.json`：

```json
{
  "headers": [
    {
      "source": "(.*)\\.md",
      "headers": [
        {
          "key": "Content-Type",
          "value": "text/plain"
        }
      ]
    }
  ]
}
```

### SPA 路由回退

Vercel 默认支持 SPA 路由回退，无需额外配置。如果需要自定义：

```json
{
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

### 部署后侧边栏为空

确保在推送前运行了 `python gen_sidebar.py`，或在 CI 中添加生成步骤。
