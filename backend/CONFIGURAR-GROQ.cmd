@echo off
setlocal
cd /d "%~dp0"
title Configurar Groq no LicitaHub

echo.
echo ==========================================
echo   Configurar Groq no LicitaHub
echo ==========================================
echo.
echo Cole a chave da API da Groq nesta janela.
echo Ela sera guardada somente neste computador
echo e o arquivo nao sera enviado ao GitHub.
echo.

set /p GROQ_KEY=Chave da API:
if "%GROQ_KEY%"=="" (
  echo.
  echo Nenhuma chave foi informada.
  pause
  exit /b 1
)

set /p GROQ_MODEL=Modelo [openai/gpt-oss-20b]:
if "%GROQ_MODEL%"=="" set "GROQ_MODEL=openai/gpt-oss-20b"

(
  echo GROQ_API_KEY=%GROQ_KEY%
  echo GROQ_MODEL=%GROQ_MODEL%
  echo GROQ_TECHNICAL_ANALYSIS_MODEL=%GROQ_MODEL%
  echo GROQ_CAPTURE_CLASSIFICATION_MODEL=%GROQ_MODEL%
) > "%~dp0.env.groq"

echo.
echo Groq configurada com sucesso.
echo Feche e abra novamente o backend para aplicar.
echo.
pause
