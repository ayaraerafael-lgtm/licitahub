# Backend LicitaHub

Backend local em Go para a API do LicitaHub.

## Responsabilidades

- Servir a API HTTP.
- Validar sessao e perfil de acesso.
- Gravar e consultar dados no PostgreSQL.
- Servir o frontend compilado em `../dist`.
- Servir arquivos enviados em `/uploads/`.
- Aplicar migracoes seguras ao iniciar.

## Como iniciar no Windows

Abra:

```powershell
.\backend\run-dev.cmd
```

O script:

1. Pede a senha do PostgreSQL se ela nao estiver em variavel de ambiente.
2. Verifica se existe `../dist/index.html`.
3. Se necessario, tenta rodar `npm.cmd install` e `npm.cmd run build`.
4. Liga o backend em `http://127.0.0.1:8080`.

## Executavel atual

O `run-dev.cmd` aponta para:

```text
licitahub-v44.exe
```

Se o backend for recompilado, atualize o `run-dev.cmd` para apontar para a versao nova.

## Variaveis usadas

- `PSQL_PATH`: caminho do `psql.exe`.
- `PGHOST`: host do PostgreSQL.
- `PGPORT`: porta do PostgreSQL.
- `PGUSER`: usuario do PostgreSQL.
- `PGPASSWORD`: senha do PostgreSQL.
- `PGDATABASE`: banco usado pela aplicacao.
- `APP_PORT`: porta HTTP do backend.
- `PUBLIC_BASE_URL`: base usada em links gerados pelo sistema.

## Rotas principais

- `/health`
- `/api/auth/login`
- `/api/auth/session`
- `/api/auth/logout`
- `/api/auth/forgot-password`
- `/api/auth/reset-password`
- `/api/users/me`
- `/api/notifications`
- `/api/company-invitations`
- `/api/companies`
- `/api/company-users`
- `/api/news`
- `/api/community/posts`
- `/api/tenders`
- `/api/partnership-ads`
- `/api/matches`
- `/api/chat/threads`
- `/api/chat/events`

## Permissoes

As rotas administrativas exigem perfil adequado no backend.

Exemplos:

- Convites e analise de empresas: administrador da plataforma.
- Noticias administrativas: administrador da plataforma.
- Cadastro/edicao/exclusao de edital: administrador da plataforma.
- Edicao do perfil da empresa: administrador da empresa.
- Gestao de usuarios vinculados: administrador da empresa.
- Desistencia, lideranca e anuncio complementar do consorcio: administrador da empresa.

## Pendente para fase posterior

- Autenticacao de producao mais robusta.
- Hash de senha proprio para producao.
- Logs estruturados.
- Auditoria completa de acoes sensiveis.
- Configuracao formal por ambiente.
