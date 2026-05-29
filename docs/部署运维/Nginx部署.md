# Nginx 部署

本文详细介绍如何将 MDOnline 部署到 Nginx 服务器。

## 前提条件

- 已安装 Nginx（≥ 1.18）
- 具有 Nginx 配置文件的编辑权限
- 已构建好 MDOnline 的静态文件

## 根目录部署

最简单的部署方式，站点直接运行在域名根路径下。

### 1. 上传文件

将项目所有文件上传到服务器目录：

```bash
scp -r ./  user@server:/var/www/mdonline/
```

### 2. 配置 Nginx

```nginx
server {
    listen 80;
    server_name docs.example.com;
    root /var/www/mdonline;
    index index.html;

    # SPA 路由回退
    location / {
        try_files $uri $uri/ /index.html;
    }

    # .md 文件 MIME 类型
    types {
        text/plain  md;
    }
}
```

### 3. 重载 Nginx

```bash
sudo nginx -t          # 检查配置语法
sudo nginx -s reload   # 重载配置
```

## 子路径部署

将站点部署在域名的某个子路径下（如 `/apidocs/`）。

### 配置示例

```nginx
server {
    listen 80;
    server_name example.com;

    location /apidocs/ {
        alias /var/www/mdonline/;
        index index.html;
        try_files $uri $uri/ /apidocs/index.html;

        types {
            text/plain  md;
        }
    }
}
```

> MDOnline 的 `basePath` 检测机制会自动适配子路径，无需手动修改 `index.html`。

## HTTPS 配置（推荐）

生产环境建议启用 HTTPS：

```nginx
server {
    listen 443 ssl http2;
    server_name docs.example.com;

    ssl_certificate     /etc/ssl/certs/docs.example.com.crt;
    ssl_certificate_key /etc/ssl/private/docs.example.com.key;

    root /var/www/mdonline;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    types {
        text/plain  md;
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name docs.example.com;
    return 301 https://$host$request_uri;
}
```

## 缓存优化

对静态资源配置缓存，提升访问速度：

```nginx
location ~* \.(js|css|svg|png|jpg|ico)$ {
    expires 7d;
    add_header Cache-Control "public, immutable";
}

location ~* \.md$ {
    expires 1h;
    add_header Cache-Control "public";
}
```

## 安全头

推荐添加以下安全响应头：

```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
```

## 故障排查

| 症状 | 可能原因 | 解决方案 |
|------|---------|---------|
| 页面 404 | SPA 路由未回退 | 检查 `try_files` 配置 |
| `.md` 文件下载而非渲染 | MIME 类型错误 | 添加 `text/plain md` 类型映射 |
| 样式/脚本加载失败 | 文件路径错误 | 检查 `root` / `alias` 配置 |
| 子路径下资源 404 | `basePath` 检测失败 | 检查 `alias` 尾部是否有 `/` |
