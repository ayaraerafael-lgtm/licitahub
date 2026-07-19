@echo off
setlocal
cd /d "%~dp0"
title Configurar Google Gemini no LicitaHub

echo.
echo =============================================
echo   Configurar Google Gemini no LicitaHub
echo =============================================
echo.
echo Cole a chave da API do Google Gemini nesta janela.
echo Ela sera guardada somente neste computador e nao ira para o GitHub.
echo.
set /p GEMINI_API_KEY=Chave da API:
if "%GEMINI_API_KEY%"=="" goto missing_key

set /p GEMINI_MODEL=Modelo [gemini-3.5-flash]:
if "%GEMINI_MODEL%"=="" set "GEMINI_MODEL=gemini-3.5-flash"

> ".env.gemini" echo GEMINI_API_KEY=%GEMINI_API_KEY%
>> ".env.gemini" echo GEMINI_TECHNICAL_ANALYSIS_MODEL=%GEMINI_MODEL%
>> ".env.gemini" echo GEMINI_TENDER_ANALYSIS_MODEL=%GEMINI_MODEL%
>> ".env.gemini" echo GEMINI_CAPTURE_CLASSIFICATION_MODEL=%GEMINI_MODEL%

echo.
echo Configuracao salva.
echo Feche e abra novamente o backend pelo arquivo run-dev.cmd.
echo.
pause
exit /b 0

:missing_key
echo.
echo Nenhuma chave foi informada. Nada foi salvo.
pause
exit /b 1
