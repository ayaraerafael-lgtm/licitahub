@echo off
title LicitaHub - Enviar ao GitHub
cd /d "%~dp0"

echo.
echo Enviando o LicitaHub ao GitHub...
echo.

"C:\Program Files\Git\cmd\git.exe" push origin main

echo.
if errorlevel 1 (
  echo O envio nao foi concluido. Verifique sua conexao ou login do GitHub.
) else (
  echo Envio concluido com sucesso.
)
echo.
pause
