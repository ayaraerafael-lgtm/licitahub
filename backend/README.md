# LicitaHub API

Backend em Go para ligar as telas do LicitaHub ao PostgreSQL.

## Modulo Acesso e Administracao

Rotas criadas para o fluxo de entrada de empresas:

- `GET /api/access-profiles`
  - Lista perfis de acesso: administrador da empresa, comercial, tecnico e leitor.

- `GET /api/company-invitations`
  - Lista convites de empresas.

- `POST /api/company-invitations`
  - Cria um convite de empresa.
  - Obrigatorios: nome da empresa, CNPJ, responsavel, email e telefone.
  - CNPJ nao repete.
  - Nome da empresa nao repete.
  - Email pode repetir.

Exemplo:

```json
{
  "tradeName": "Engenharia Teste",
  "cnpj": "12.345.678/0001-90",
  "contactName": "Maria Silva",
  "email": "contato@engenhariateste.com.br",
  "phone": "(11) 99999-0000",
  "state": "SP",
  "internalNote": "Empresa indicada pelo administrador."
}
```

- `POST /api/company-invitations/{id}/accept`
  - Aceita o convite, cria a empresa em analise e cria o primeiro usuario administrador da empresa.

Exemplo:

```json
{
  "adminFullName": "Maria Silva",
  "adminEmail": "maria@engenhariateste.com.br",
  "adminPhone": "(11) 99999-0000",
  "adminJobTitle": "Diretora comercial",
  "website": "https://www.engenhariateste.com.br",
  "institutionalDescription": "Empresa de engenharia consultiva.",
  "city": "Sao Paulo",
  "state": "SP"
}
```

- `PATCH /api/company-invitations/{id}/cancel`
  - Cancela um convite ainda nao usado.

- `GET /api/companies`
  - Lista empresas cadastradas.

## Modulo Usuarios Vinculados

- `GET /api/company-users?companyId={idDaEmpresa}`
  - Lista usuarios vinculados a uma empresa.

- `POST /api/company-users`
  - Cadastra um usuario vinculado a uma empresa.

Exemplo:

```json
{
  "companyId": "id-da-empresa",
  "fullName": "Joao Almeida",
  "email": "joao@empresa.com.br",
  "phone": "(31) 98888-0000",
  "jobTitle": "Coordenador tecnico",
  "accessProfileKey": "technical"
}
```

- `PUT /api/company-users/{id}`
  - Edita dados do usuario vinculado.

- `PATCH /api/company-users/{id}/block`
  - Bloqueia acesso do usuario.

- `PATCH /api/company-users/{id}/unblock`
  - Desbloqueia acesso do usuario.

- `PATCH /api/company-users/{id}/remove`
  - Remove acesso do usuario, preservando historico.

## Modulo Radar LicitaHub

Rotas iniciais:

- `GET /health`
  - Confere se a API esta ligada.

- `GET /api/news/categories`
  - Lista categorias ativas de noticias.

- `GET /api/news`
  - Lista noticias cadastradas.

- `POST /api/news`
  - Cadastra uma noticia.

Exemplo de corpo para `POST /api/news`:

```json
{
  "title": "Nova rodada de oportunidades em infraestrutura",
  "categorySlug": "licitacoes",
  "status": "published",
  "summary": "Resumo que aparece nos cards do Radar LicitaHub.",
  "content": "Texto completo da noticia.",
  "mainImageUrl": "https://exemplo.com/imagem.jpg",
  "mainImageFileName": "imagem.jpg",
  "mainImageMimeType": "image/jpeg"
}
```

Status aceitos:

- `draft` ou `rascunho`
- `published` ou `publicado`
- `featured` ou `destaque principal`

Categorias iniciais:

- `licitacoes`
- `mercado`
- `legislacao`
- `eventos`
- `comunicados`

## Como iniciar no Windows

Depois de instalar o Go e o PostgreSQL, inicie com:

```powershell
.\start-dev.ps1
```

Esse script pede a senha do PostgreSQL na hora e nao grava a senha em arquivo. A API usa o `psql` instalado junto com o PostgreSQL para conversar com o banco local de desenvolvimento.
