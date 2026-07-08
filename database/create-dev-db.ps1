param(
  [string]$DatabaseName = "licitahub_dev",
  [string]$User = "postgres",
  [string]$HostName = "localhost",
  [string]$Port = "5432"
)

$ErrorActionPreference = "Stop"

$schemaPath = Join-Path $PSScriptRoot "schema.sql"

if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
  throw "psql nao foi encontrado. Instale o PostgreSQL ou adicione a pasta bin do PostgreSQL ao PATH."
}

if (-not (Test-Path $schemaPath)) {
  throw "schema.sql nao encontrado em $schemaPath"
}

$env:PGHOST = $HostName
$env:PGPORT = $Port
$env:PGUSER = $User

$exists = & psql -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$DatabaseName';"

if ($exists.Trim() -ne "1") {
  & psql -d postgres -c "CREATE DATABASE $DatabaseName;"
}

& psql -d $DatabaseName -f $schemaPath

Write-Host "Banco $DatabaseName pronto."

