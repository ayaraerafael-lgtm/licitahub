$ErrorActionPreference = "Stop"

$go = "go"
if (Test-Path "C:\Program Files\Go\bin\go.exe") {
  $go = "C:\Program Files\Go\bin\go.exe"
}

$password = Read-Host "Senha do usuario postgres" -AsSecureString
$plainPassword = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
  [Runtime.InteropServices.Marshal]::SecureStringToBSTR($password)
)

$env:PSQL_PATH = "C:\Program Files\PostgreSQL\17\bin\psql.exe"
$env:PGHOST = "localhost"
$env:PGPORT = "5432"
$env:PGUSER = "postgres"
$env:PGPASSWORD = $plainPassword
$env:PGDATABASE = "licitahub_dev"
$env:APP_PORT = "8080"
$env:GOCACHE = "C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-cache"
$env:GOMODCACHE = "C:\Users\Financeiro Hollus\Documents\Codex\2026-07-04\pre\outputs\licitahub-app\.go-mod-cache"

& $go run .
