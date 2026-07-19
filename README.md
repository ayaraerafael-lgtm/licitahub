# LicitaHub

Plataforma web para empresas de engenharia consultiva. O produto reune rede social empresarial, noticias, editais publicos, formacao de consorcios, montagem colaborativa de propostas, capacidade tecnica e capacitacao profissional.

Documentacao revisada em 18/07/2026.

## Estado do projeto

O LicitaHub possui frontend React compilado, backend em Go, banco PostgreSQL e armazenamento local de arquivos. Os principais fluxos funcionam no ambiente local de desenvolvimento e homologacao.

O sistema ainda nao deve ser tratado como producao publica. Antes da hospedagem externa, devem ser concluidos os pontos de seguranca, infraestrutura, backup, e-mail, HTTPS, armazenamento de arquivos, monitoramento e testes descritos em `docs/SEGURANCA.md`.

## Tecnologias

- Frontend: React, HTML e CSS, compilados com Vite.
- Backend: Go, API HTTP e servidor do frontend compilado.
- Banco: PostgreSQL 17.
- Arquivos: diretorio local `backend/uploads/`.
- OCR: Tesseract instalado no servidor.
- IA opcional: APIs da OpenAI, Google Gemini e Groq configuradas apenas no backend.
- Fontes oficiais: PNCP e Compras.gov.br/Dados Abertos.

## Estrutura

- `src/`: codigo-fonte do frontend.
- `dist/`: frontend compilado pelo Vite.
- `backend/`: API Go, migracoes de inicializacao, prompts, uploads e scripts locais.
- `database/schema.sql`: referencia consolidada do banco PostgreSQL.
- `docs/SISTEMA.md`: fluxos, telas e regras de negocio.
- `docs/SEGURANCA.md`: plano e situacao de seguranca para producao.
- `docs/ROTEIRO-DE-TESTES.md`: cenarios manuais de validacao.
- `preview-react-cdn.html`: redirecionador legado; nao deve ser usado como aplicacao principal.

## Como executar localmente

Pre-requisitos:

- Node.js.
- Go.
- PostgreSQL 17.
- Banco `licitahub_dev`.
- Tesseract, somente para OCR de documentos digitalizados.

Passos:

1. Na raiz do projeto, execute `npm.cmd install` se as dependencias ainda nao estiverem instaladas.
2. Execute `npm.cmd run build` para gerar `dist/`.
3. Abra `backend/run-dev.cmd`.
4. Informe a senha do PostgreSQL quando solicitada.
5. Aguarde a mensagem `LicitaHub API listening on :8080`.
6. Acesse `http://127.0.0.1:8080/`.

O script recompila o backend antes de iniciar. Se `dist/index.html` nao existir, ele tambem tenta compilar o frontend.

## Papeis de acesso

- Administrador da plataforma: empresas, convites, noticias, editais, impugnacoes, cursos, captacao oficial e rodadas de avaliacao.
- Administrador da empresa: perfil empresarial, usuarios vinculados e distribuicao institucional de estrelas.
- Comercial: comunidade, editais, anuncios, matches, consorcios e operacao comercial permitida.
- Tecnico: capacidade tecnica, tarefas, documentos e consultas autorizadas.
- Leitor: consulta de conteudos sem comandos administrativos.

As permissoes sao verificadas no menu e novamente no backend. Esconder um comando na tela nao substitui a validacao da API.

## Modulos implementados

### Acesso e administracao

- Convites, aceite, analise e aprovacao de empresas.
- Bloqueio de empresa e encerramento das sessoes vinculadas.
- Perfis de acesso, login, limite de tentativas e recuperacao de senha.
- Consulta administrativa dos pedidos de recuperacao.
- Logs de acesso, erros e acoes sensiveis selecionadas.

### Radar LicitaHub

- Noticias com imagem, destaque, periodo de publicacao e situacao.
- Gerenciamento, filtros, paginacao e detalhe da noticia.

### Comunidade

- Publicacoes empresariais, imagens, categorias, curtidas, favoritos e comentarios.
- Gestao de publicacoes proprias, inclusive arquivamento e reativacao.
- Perfil publico de empresas e profissionais.

### Avaliacao de parcerias

- Rodadas anonimas abertas pelo administrador da plataforma.
- Fotografia das empresas ativas no momento da abertura.
- Prazo obrigatorio e encerramento automatico.
- Saldo de estrelas igual a 30% das participantes, arredondado para cima.
- Distribuicao livre do saldo, inclusive todas as estrelas para uma unica empresa.
- Revisao final, envio unico, ranking da rodada, media historica e tendencia.
- Exclusao administrativa somente de rodada encerrada.

### Empresa

- Perfil institucional, logo, site, porte, cidade, UF e atuacao nacional.
- Usuarios vinculados, foto, cargo, perfil, bloqueio, desbloqueio e remocao.
- Meu perfil e perfil publico empresarial.

### Editais e parcerias

- Cadastro, edicao, documentos, HTML de pre-analise e ciclo de vida do edital, com OpenAI e fallback automatico para Google Gemini.
- Lista, detalhe, linha do tempo e registro de interesse.
- Participacao individual, busca de parceiros ou acompanhamento.
- Empresas interessadas, vitrine, avaliacao reciproca, match e consorcio.
- Consorcios com lideranca, terceira empresa, desistencias e anuncios complementares.
- Chat em tempo real relacionado a anuncios, tarefas e profissionais.

### Area de trabalho

- Central de montagens individuais e consorciais.
- Fases, tarefas, responsaveis, prazos, comentarios e evidencias.
- Kanban pessoal de tarefas.
- Calendario mensal de montagens.

### Impugnacoes

- Pedido de impugnacao por empresa, fundamentacao e anexos.
- Alerta de intempestividade conforme prazo informado.
- Kanban administrativo, situacoes, responsavel e contato do solicitante.

### Captacao oficial

- Consulta PNCP e Compras.gov.br.
- Consulta paginada e comandada pelo administrador, sem varredura automatica ilimitada.
- No PNCP, cada comando solicita uma pagina de ate 50 registros por modalidade selecionada.
- Fila administrativa sem publicacao automatica.
- Aderencia por termos de engenharia consultiva.
- Preparacao de rascunho, descarte e saneamento entre fontes exclusivamente pelo numero de controle PNCP.
- Classificacao opcional da fila por OpenAI, Google Gemini e Groq, mantendo a decisao final com o administrador.

### Capacidade tecnica

- Profissionais, formacoes e contatos.
- Atestados, CAT, quantitativos, documentos e texto completo.
- Extracao de PDF, OCR e correcao manual.
- Selecao de atestados e analise opcional por OpenAI, Google Gemini ou Groq, com fallback automatico e identificacao do provedor no resultado.

### Academia LicitaHub

- Cursos, aulas por YouTube ou video enviado, progresso e retomada.
- Questionarios, nota minima, bloqueio de aulas e historico individual.
- Certificado PDF com codigo publico de validacao.

### Notificacoes

- Sino com contador, leitura, destino relacionado e historico pesquisavel.
- Alertas para noticias, editais, comunidade, convites, matches, tarefas, chat e rodadas de avaliacao.

## Dados locais e credenciais

- O usuario administrativo local padrao e `admin@licitahub.local`.
- Senhas, tokens e chaves reais nao devem ser documentados nem enviados ao GitHub.
- O arquivo `backend/.env.example` contem somente nomes de variaveis.
- As chaves da OpenAI, Gemini e Groq, quando utilizadas, ficam em arquivos locais ignorados pelo Git.

## Comandos de verificacao

Frontend:

```powershell
npm.cmd run build
```

Backend:

```powershell
cd backend
go test ./...
```

Testes automatizados locais:

```powershell
npm.cmd run test:smoke
npm.cmd run test:homologacao
```

O teste de homologacao exige o backend ativo e uma base preparada exclusivamente para validacao. Ele nao deve ser executado contra dados de producao.

Banco de referencia:

```powershell
psql -U postgres -d licitahub_dev -f database/schema.sql
```

O ultimo comando deve ser usado com cuidado e somente em banco apropriado. O backend tambem aplica migracoes incrementais ao iniciar.

## Limites do ambiente atual

- Senhas locais ainda exigem migracao para algoritmo de hash adequado antes da producao.
- Cookies locais usam configuracao sem `Secure`, apropriada apenas para HTTP local.
- Uploads ficam no disco do servidor e precisam migrar para armazenamento persistente em producao.
- Recuperacao de senha ainda depende de atendimento administrativo quando o e-mail nao esta configurado.
- IA depende de chave e creditos independentes da assinatura do ChatGPT.
- WhatsApp permanece planejado e depende da plataforma oficial da Meta.
- Backup, HTTPS, monitoramento externo, rotacao de logs e testes automatizados amplos ainda precisam ser implantados.

## Documentacao

- [Sistema e regras de negocio](docs/SISTEMA.md)
- [Seguranca e preparacao para producao](docs/SEGURANCA.md)
- [Roteiro completo de testes](docs/ROTEIRO-DE-TESTES.md)
