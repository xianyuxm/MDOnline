@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

echo 正在自动生成侧边栏...
python gen_sidebar.py

echo 正在启动文档服务器...
python -m http.server 8080
endlocal
