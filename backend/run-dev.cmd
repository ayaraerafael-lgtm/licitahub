@echo off
setlocal
cd /d "%~dp0"
title LicitaHub API - Backend local

echo.
echo ==========================================
echo   LicitaHub API - Backend local
echo ==========================================
echo.
echo Esta janela precisa ficar aberta.
echo Se voce fechar esta janela, o backend para.
echo.

set "PSQL_PATH=C:\Program Files\PostgreSQL\17\bin\psql.exe"
set "PGHOST=localhost"
set "PGPORT=5432"
set "PGUSER=postgres"
set "PGDATABASE=licitahub_dev"
set "APP_PORT=8080"
set "GOCACHE=C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-cache"
set "GOMODCACHE=C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-mod-cache"

if "%PGPASSWORD%"=="" (
  set /p PGPASSWORD=Senha do usuario postgres: 
)

if not exist "%~dp0..\dist\index.html" (
  echo.
  echo Frontend compilado nao encontrado. Gerando versao Vite...
  pushd "%~dp0.."
  if not exist "node_modules" (
    call npm.cmd install
    if errorlevel 1 goto frontend_error
  )
  call npm.cmd run build
  if errorlevel 1 goto frontend_error
  popd
)

echo.
echo Ligando backend em http://127.0.0.1:8080 ...
echo Aguarde aparecer: LicitaHub API listening on :8080
echo.

"%~dp0licitahub-v29.exe"

echo.
echo O backend foi encerrado.
echo Se isso aconteceu sem voce querer, confira se a senha do PostgreSQL esta correta
echo e tente abrir este arquivo novamente.
echo.
pause
exit /b

:frontend_error
popd
echo.
echo Nao foi possivel gerar o frontend Vite.
echo Verifique se o Node.js esta instalado e tente novamente.
echo.
pause
