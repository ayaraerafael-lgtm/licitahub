$ErrorActionPreference = "Stop"

$postgresBin = "C:\Program Files\PostgreSQL\17\bin"
$sourceDatabase = "licitahub_dev"
$targetDatabase = "licitahub_homologacao"
$psql = Join-Path $postgresBin "psql.exe"
$pgDump = Join-Path $postgresBin "pg_dump.exe"
$createdb = Join-Path $postgresBin "createdb.exe"
$pgRestore = Join-Path $postgresBin "pg_restore.exe"

foreach ($tool in @($psql, $pgDump, $createdb, $pgRestore)) {
  if (-not (Test-Path $tool)) {
    throw "PostgreSQL nao encontrado em: $tool"
  }
}

$securePassword = Read-Host "Senha do usuario postgres" -AsSecureString
$plainPassword = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
  [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
)
$env:PGPASSWORD = $plainPassword
$env:PGCONNECT_TIMEOUT = "10"

$exists = & $psql -h localhost -p 5432 -U postgres -d postgres -Atqc "SELECT 1 FROM pg_database WHERE datname = '$targetDatabase';"
if ($exists -eq "1") {
  throw "A base $targetDatabase ja existe. Ela foi preservada e nada foi alterado."
}

$dumpFile = Join-Path $env:TEMP ("$targetDatabase-" + (Get-Date -Format "yyyyMMddHHmmss") + ".dump")
try {
  Write-Host "Criando uma copia protegida da base atual..."
  & $pgDump -h localhost -p 5432 -U postgres -Fc -d $sourceDatabase -f $dumpFile
  & $createdb -h localhost -p 5432 -U postgres $targetDatabase
  & $pgRestore -h localhost -p 5432 -U postgres -d $targetDatabase $dumpFile
  Write-Host ""
  Write-Host "Base de homologacao criada: $targetDatabase"
  Write-Host "A base original $sourceDatabase nao foi alterada."
  Write-Host "Agora abra INICIAR-HOMOLOGACAO.cmd para ligar a versao de teste em http://127.0.0.1:8081"
} finally {
  if (Test-Path $dumpFile) { Remove-Item -LiteralPath $dumpFile -Force }
  Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
