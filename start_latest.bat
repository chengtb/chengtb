@echo off
setlocal EnableDelayedExpansion

:: ============================================================
:: start_latest.bat
:: 功能：
::   1. 扫描当前目录下所有 program-v*.exe 文件，找出版本最大的那个
::   2. 检查是否有 program 相关进程正在运行
::      - 若正在运行且已是最大版本 → 跳过
::      - 若正在运行但不是最大版本 → 关闭旧版本，启动最大版本
::      - 若未运行              → 启动最大版本
:: ============================================================

:: -----------------------------------------------------------
:: 1. 扫描目录，找出版本最大的 program-v*.exe
:: -----------------------------------------------------------
set "PROGRAM_NAME=program"
set "MAX_VER_FILE="
set "MAX_MAJOR=0"
set "MAX_MINOR=0"
set "MAX_PATCH=0"

for %%F in (%PROGRAM_NAME%-v*.exe) do (
    set "FNAME=%%~nF"
    :: 去掉前缀 "program-v" 得到 "MAJOR.MINOR.PATCH"
    set "VER=!FNAME:*-v=!"

    :: 拆分版本号各段
    for /f "tokens=1,2,3 delims=." %%A in ("!VER!") do (
        set /a "CUR_MAJOR=%%A" 2>nul
        set /a "CUR_MINOR=%%B" 2>nul
        set /a "CUR_PATCH=%%C" 2>nul
    )

    :: 与当前最大版本比较（先比 MAJOR，再 MINOR，再 PATCH）
    set "IS_GREATER=0"
    if !CUR_MAJOR! gtr !MAX_MAJOR! set "IS_GREATER=1"
    if !CUR_MAJOR! equ !MAX_MAJOR! (
        if !CUR_MINOR! gtr !MAX_MINOR! set "IS_GREATER=1"
        if !CUR_MINOR! equ !MAX_MINOR! (
            if !CUR_PATCH! gtr !MAX_PATCH! set "IS_GREATER=1"
        )
    )

    if "!IS_GREATER!"=="1" (
        set "MAX_MAJOR=!CUR_MAJOR!"
        set "MAX_MINOR=!CUR_MINOR!"
        set "MAX_PATCH=!CUR_PATCH!"
        set "MAX_VER_FILE=%%F"
    )
)

:: 若未找到任何匹配文件则退出
if "!MAX_VER_FILE!"=="" (
    echo [ERROR] 当前目录下未找到 %PROGRAM_NAME%-v*.exe 文件，请确认路径后重试。
    goto :EOF
)

echo [INFO] 最大版本文件: !MAX_VER_FILE!

:: -----------------------------------------------------------
:: 2. 检查 program 相关进程是否正在运行
::    - MAX_RUNNING: 最大版本是否正在运行（yes/no）
::    - OLD_RUNNING:  是否有非最大版本正在运行（yes/no）
:: -----------------------------------------------------------
set "MAX_RUNNING=no"
set "OLD_RUNNING=no"

for %%F in (%PROGRAM_NAME%-v*.exe) do (
    tasklist /fi "IMAGENAME eq %%F" 2>nul | find /i "%%F" >nul 2>&1
    if not errorlevel 1 (
        if /i "%%F"=="!MAX_VER_FILE!" (
            set "MAX_RUNNING=yes"
        ) else (
            set "OLD_RUNNING=yes"
        )
    )
)

:: -----------------------------------------------------------
:: 3. 决策逻辑
:: -----------------------------------------------------------
:: 情形 A：最大版本已在运行 → 跳过（无论是否同时有旧版本）
if "!MAX_RUNNING!"=="yes" (
    echo [INFO] 最大版本 !MAX_VER_FILE! 已在运行，无需操作。
    goto :EOF
)

:: 情形 B：有旧版本在运行 → 逐一关闭所有旧版本，再启动最大版本
if "!OLD_RUNNING!"=="yes" (
    echo [INFO] 检测到旧版本进程正在运行，准备关闭...
    set "KILL_FAILED=no"
    for %%F in (%PROGRAM_NAME%-v*.exe) do (
        if /i not "%%F"=="!MAX_VER_FILE!" (
            tasklist /fi "IMAGENAME eq %%F" 2>nul | find /i "%%F" >nul 2>&1
            if not errorlevel 1 (
                echo [INFO]   正在关闭 %%F ...
                :: 先尝试正常退出
                taskkill /im "%%F" >nul 2>&1
                :: 等待进程退出（最多 3 秒）
                timeout /t 3 /nobreak >nul
                :: 若仍在运行则强制终止
                tasklist /fi "IMAGENAME eq %%F" 2>nul | find /i "%%F" >nul 2>&1
                if not errorlevel 1 (
                    taskkill /f /im "%%F" >nul 2>&1
                    if errorlevel 1 (
                        echo [WARNING]   强制关闭 %%F 失败，请手动关闭后重试。
                        set "KILL_FAILED=yes"
                    )
                )
            )
        )
    )
    if "!KILL_FAILED!"=="yes" (
        echo [WARNING] 部分旧版本进程未能关闭，已取消启动。
        goto :EOF
    )
    :: 再次等待，确保系统资源完全释放
    timeout /t 2 /nobreak >nul
)

:: 情形 C（或 B 成功关闭后）：启动最大版本
echo [INFO] 正在启动最大版本 !MAX_VER_FILE! ...
start "" "!MAX_VER_FILE!"

:EOF
endlocal
