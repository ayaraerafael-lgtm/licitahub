# LicitaHub API

Backend em Go para ligar as telas do LicitaHub ao PostgreSQL.

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
