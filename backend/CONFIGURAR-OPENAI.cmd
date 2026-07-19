@echo off
setlocal
cd /d "%~dp0"
title Configurar IA do LicitaHub

echo.
echo =============================================
echo   Configurar analise de editais com IA
echo =============================================
echo.
echo Cole a chave da API da OpenAI nesta janela.
echo Ela sera guardada somente neste computador e nao ira para o GitHub.
echo.
set /p OPENAI_API_KEY=Chave da API: 
if "%OPENAI_API_KEY%"=="" goto missing_key

set /p OPENAI_ANALYSIS_MODEL=Modelo [gpt-5.6]: 
if "%OPENAI_ANALYSIS_MODEL%"=="" set "OPENAI_ANALYSIS_MODEL=gpt-5.6"

> ".env.openai" echo OPENAI_API_KEY=%OPENAI_API_KEY%
>> ".env.openai" echo OPENAI_ANALYSIS_MODEL=%OPENAI_ANALYSIS_MODEL%
>> ".env.openai" echo OPENAI_CAPTURE_CLASSIFICATION_MODEL=%OPENAI_ANALYSIS_MODEL%

echo.
echo Configuracao salva. Feche e abra novamente o arquivo run-dev.cmd.
echo.
pause
exit /b 0

:missing_key
echo.
echo Nenhuma chave foi informada. Nada foi salvo.
pause
exit /b 1
