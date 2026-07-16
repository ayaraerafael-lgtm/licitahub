@echo off
setlocal
title LicitaHub - Commit e envio ao GitHub
cd /d "%~dp0"

set "GIT=C:\Program Files\Git\cmd\git.exe"

echo.
echo ================================================
echo   LicitaHub - Commit e envio ao GitHub
echo ================================================
echo.

if not exist "%GIT%" (
  echo Git nao foi encontrado neste computador.
  echo Instale o Git e tente novamente.
  goto end
)

echo Preparando os arquivos alterados...
"%GIT%" config --global --add safe.directory "%~dp0"
"%GIT%" add -A
if errorlevel 1 goto error

"%GIT%" diff --cached --quiet
if not errorlevel 1 goto push_existing

echo.
echo Criando o commit...
"%GIT%" commit -m "Atualiza captacao e documentacao"
if errorlevel 1 goto error

:push_existing
echo.
echo Enviando ao GitHub...
"%GIT%" push origin main
if errorlevel 1 goto error

echo.
echo ================================================
echo   Commit e envio concluidos com sucesso.
echo ================================================
"%GIT%" log -1 --oneline
goto end

:error
echo.
echo O processo nao foi concluido.
echo Leia a mensagem acima para identificar o motivo.

:end
echo.
pause
endlocal
