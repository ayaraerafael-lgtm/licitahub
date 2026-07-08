# Banco de dados LicitaHub

O schema principal do MVP esta em `schema.sql`.

## Criar banco local

1. Instale o PostgreSQL.
2. Garanta que `psql` esteja no PATH do Windows.
3. Rode:

```powershell
.\database\create-dev-db.ps1
```

Por padrao, o script cria/aplica o schema no banco `licitahub_dev`.

Tambem pode informar parametros:

```powershell
.\database\create-dev-db.ps1 -DatabaseName licitahub_dev -User postgres -HostName localhost -Port 5432
```

O backend Go deve usar a variavel `DATABASE_URL`, conforme `.env.example`.

