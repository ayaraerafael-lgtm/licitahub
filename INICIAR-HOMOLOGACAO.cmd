@echo off
setlocal
cd /d "%~dp0backend"
title LicitaHub - Homologacao local

set /p PGPASSWORD=Senha do usuario postgres: 
set "PSQL_PATH=C:\Program Files\PostgreSQL\17\bin\psql.exe"
set "PGHOST=localhost"
set "PGPORT=5432"
set "PGUSER=postgres"
set "PGDATABASE=licitahub_homologacao"
set "APP_PORT=8081"
set "GOCACHE=C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-cache"
set "GOMODCACHE=C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-mod-cache"

if exist "%CD%\.env.openai" (
  for /f "usebackq tokens=1,* delims==" %%A in ("%CD%\.env.openai") do set "%%A=%%B"
)

echo.
echo Ligando homologacao em http://127.0.0.1:8081 ...
echo Esta janela deve ficar aberta durante os testes.
echo.
"%CD%\licitahub-v97.exe"
pause
