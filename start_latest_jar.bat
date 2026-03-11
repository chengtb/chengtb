@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

:: ============================================================
:: start_latest_jar.bat
:: 功能：
::   1. 扫描当前目录下所有 *-exec.jar 文件，找出版本最大的那个
::      文件名格式：PREFIX-MAJOR.MINOR.PATCH-exec.jar
::      例如：ac-aftersale-aicmos-1.0.10-exec.jar
::   2. 检查是否有相同 JAR 的 java 进程正在运行
::      - 若正在运行且已是最大版本 → 跳过
::      - 若正在运行但不是最大版本 → 关闭旧版本，启动最大版本
::      - 若未运行              → 用 java -jar 在当前窗口启动最大版本
:: ============================================================

:: -----------------------------------------------------------
:: 1. 扫描目录，找出版本最大的 *-exec.jar
:: -----------------------------------------------------------
set "MAX_VER_FILE="
set "MAX_MAJOR=0"
set "MAX_MINOR=0"
set "MAX_PATCH=0"

for %%F in (*-exec.jar) do (
    set "FNAME=%%~nF"
    :: 去掉 "-exec" 后缀，得到 "PREFIX-MAJOR.MINOR.PATCH"
    set "BASENAME=!FNAME:-exec=!"

    :: 将所有 "-" 替换为空格后逐 token 迭代，最后一个即版本号 "MAJOR.MINOR.PATCH"
    set "VER="
    for %%T in (!BASENAME:-= !) do set "VER=%%T"

    :: 拆分版本号各段
    set "CUR_MAJOR=0"
    set "CUR_MINOR=0"
    set "CUR_PATCH=0"
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
    echo [ERROR] 当前目录下未找到 *-exec.jar 文件，请确认路径后重试。
    goto :EOF
)

echo [INFO] 最大版本文件: !MAX_VER_FILE!

:: -----------------------------------------------------------
:: 2. 将所有 java.exe 进程信息写入临时文件（只查询一次，后续复用）
::    wmic CSV 格式：Node,CommandLine,ProcessId
::    注：若命令行中含英文逗号，tokens=3 可能取到错误值（极少见于 java -jar 场景）
:: -----------------------------------------------------------
set "JAVA_PROCS=%TEMP%\_javaprocs_!RANDOM!!RANDOM!.tmp"
wmic process where "name='java.exe'" get commandline,processid /format:csv 2>nul > "!JAVA_PROCS!"

:: -----------------------------------------------------------
:: 3. 判断各 JAR 的运行状态
::    - MAX_RUNNING: 最大版本是否正在运行（yes/no）
::    - OLD_RUNNING: 是否有非最大版本正在运行（yes/no）
:: -----------------------------------------------------------
set "MAX_RUNNING=no"
set "OLD_RUNNING=no"

for %%F in (*-exec.jar) do (
    find /i "%%F" "!JAVA_PROCS!" >nul 2>&1
    if not errorlevel 1 (
        if /i "%%F"=="!MAX_VER_FILE!" (
            set "MAX_RUNNING=yes"
        ) else (
            set "OLD_RUNNING=yes"
        )
    )
)

:: -----------------------------------------------------------
:: 4. 决策逻辑
:: -----------------------------------------------------------
:: 情形 A：最大版本已在运行 → 跳过（无论是否同时有旧版本）
if "!MAX_RUNNING!"=="yes" (
    del "!JAVA_PROCS!" >nul 2>&1
    echo [INFO] 最大版本 !MAX_VER_FILE! 已在运行，无需操作。
    goto :EOF
)

:: 情形 B：有旧版本在运行 → 按 PID 逐一关闭，再启动最大版本
if "!OLD_RUNNING!"=="yes" (
    echo [INFO] 检测到旧版本进程正在运行，准备关闭...
    set "KILL_FAILED=no"

    for %%F in (*-exec.jar) do (
        if /i not "%%F"=="!MAX_VER_FILE!" (
            find /i "%%F" "!JAVA_PROCS!" >nul 2>&1
            if not errorlevel 1 (
                echo [INFO]   正在关闭 %%F 对应的 java 进程...
                :: 用 PowerShell 按命令行匹配取 PID，可靠处理命令行中含逗号的情况
                set "OLD_PID="
                for /f %%P in ('powershell -NoProfile -Command "(Get-CimInstance Win32_Process | Where-Object {$_.Name -eq 'java.exe' -and $_.CommandLine -like '*%%F*'}).ProcessId"') do (
                    set "OLD_PID=%%P"
                )

                if "!OLD_PID!"=="" (
                    echo [WARNING]   未能获取 %%F 进程的 PID，请手动关闭后重试。
                    set "KILL_FAILED=yes"
                ) else (
                    echo [INFO]     PID: !OLD_PID!
                    :: 先尝试正常退出
                    taskkill /pid !OLD_PID! >nul 2>&1
                    timeout /t 3 /nobreak >nul
                    :: 若仍在运行则强制终止（find /v "" 检查 tasklist 输出是否非空）
                    tasklist /nh /fi "pid eq !OLD_PID!" 2>nul | find /v "" >nul 2>&1
                    if not errorlevel 1 (
                        taskkill /f /pid !OLD_PID! >nul 2>&1
                        if errorlevel 1 (
                            echo [WARNING]   强制关闭进程 PID=!OLD_PID!（%%F）失败，请手动关闭后重试。
                            set "KILL_FAILED=yes"
                        )
                    )
                )
            )
        )
    )

    del "!JAVA_PROCS!" >nul 2>&1
    if "!KILL_FAILED!"=="yes" (
        echo [WARNING] 部分旧版本进程未能关闭，已取消启动。
        goto :EOF
    )
    :: 再次等待，确保系统资源完全释放
    timeout /t 2 /nobreak >nul
) else (
    del "!JAVA_PROCS!" >nul 2>&1
)

:: 情形 C（或 B 成功关闭后）：在当前窗口用 java -jar 启动最大版本
echo [INFO] 正在启动最大版本 !MAX_VER_FILE! ...
java -jar "!MAX_VER_FILE!"

:EOF
endlocal
