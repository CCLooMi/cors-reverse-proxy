@echo off
echo=

echo copy config
IF EXIST %cd%\conf\config.yaml (
    if not exist %cd%\dist\conf\ md %cd%\dist\conf\
    xcopy /Y /E %cd%\conf\config.yaml %cd%\dist\conf\
) else (
    echo config.yaml does not exist. Skipping copy operation.
)

echo build
go build -o %cd%\dist\cors-revers-proxy.exe .\main

echo=
pause