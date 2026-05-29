@echo off
REM ============================================================
REM  MDOnline - Docs Server Launcher
REM
REM  If double-clicking this file doesn't work, open CMD in
REM  this folder and run the commands manually:
REM
REM    cd /d "D:\project\MDOline"
REM    py -m http.server 8080
REM
REM  Then open http://localhost:8080 in your browser.
REM ============================================================
setlocal
cd /d "%~dp0"

echo Generating sidebar...
py gen_sidebar.py 2>nul

echo Starting docs server on http://localhost:8080 ...
echo Press Ctrl+C to stop.
py -m http.server 8080
endlocal