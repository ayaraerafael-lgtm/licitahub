@echo off
setlocal EnableExtensions
title LicitaHub - Commit e envio ao GitHub
cd /d "%~dp0"

set "GIT=C:\Program Files\Git\cmd\git.exe"
set "COMMIT_MESSAGE=Atualiza modulos, regras de negocio e documentacao"

echo.
echo ================================================
echo   LicitaHub - Commit e envio ao GitHub
echo ================================================
echo.

if not exist "%GIT%" (
  where git.exe >nul 2>&1
  if errorlevel 1 (
    echo Git nao foi encontrado neste computador.
    echo Instale o Git e tente novamente.
    goto error
  )
  set "GIT=git.exe"
)

echo Conferindo o repositorio...
"%GIT%" config --global --add safe.directory "%~dp0"
if errorlevel 1 goto error

"%GIT%" rev-parse --is-inside-work-tree >nul 2>&1
if errorlevel 1 (
  echo Esta pasta nao e um repositorio Git.
  goto error
)

for /f "delims=" %%B in ('"%GIT%" branch --show-current') do set "BRANCH=%%B"
if not defined BRANCH (
  echo Nao foi possivel identificar a branch atual.
  goto error
)

"%GIT%" remote get-url origin >nul 2>&1
if errorlevel 1 (
  echo O repositorio remoto origin nao esta configurado.
  goto error
)

echo Validando a protecao de arquivos locais...
for %%P in (
  "backend/.env.openai"
  "backend/uploads"
  "backend/private_uploads"
  "backend/.tmp-go-build"
  "backend/.tmp-go-cache"
  "backend/tools"
  "tmp"
) do (
  "%GIT%" check-ignore -q "%%~P"
  if errorlevel 1 (
    echo O caminho %%~P nao esta protegido pelo .gitignore.
    echo O envio foi interrompido para proteger dados locais.
    goto error
  )
)

echo Conferindo a formatacao das alteracoes...
"%GIT%" diff --check
if errorlevel 1 goto error

echo Preparando os arquivos alterados...
"%GIT%" add -A
if errorlevel 1 goto error

echo Validando o conteudo preparado...
"%GIT%" diff --cached --check
if errorlevel 1 goto error

echo Conferindo se algum segredo foi preparado para envio...
"%GIT%" diff --cached --name-only | findstr /I /R /C:"^backend/\.env\.openai$" /C:"^\.env$" /C:"^backend/uploads/" /C:"^backend/private_uploads/" /C:"^backend/tools/" >nul
if not errorlevel 1 (
  echo Um arquivo privado foi encontrado no commit.
  echo O envio foi interrompido.
  goto error
)

"%GIT%" diff --cached --quiet
if not errorlevel 1 goto push_existing

echo.
echo Criando o commit...
"%GIT%" commit -m "%COMMIT_MESSAGE%"
if errorlevel 1 goto error

:push_existing
echo.
echo Enviando a branch %BRANCH% ao GitHub...
"%GIT%" push origin "%BRANCH%"
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
