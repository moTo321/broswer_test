@echo off
REM Browser 测试框架运行脚本 (Windows)

setlocal

REM 默认配置
set CONFIG_FILE=config.yaml
set TEST_FILE=testcase\login_example.json

REM 解析命令行参数
:parse_args
if "%~1"=="" goto run
if /i "%~1"=="-c" (
    set CONFIG_FILE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--config" (
    set CONFIG_FILE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="-f" (
    set TEST_FILE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--file" (
    set TEST_FILE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="-h" goto help
if /i "%~1"=="--help" goto help
shift
goto parse_args

:run
browser_test.exe -c "%CONFIG_FILE%" -f "%TEST_FILE%"
goto end

:help
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   -c, --config FILE    配置文件路径 (默认: config.yaml)
echo   -f, --file FILE      测试用例文件路径 (默认: testcase\login_example.json)
echo   -h, --help           显示帮助信息
goto end

:end
endlocal
