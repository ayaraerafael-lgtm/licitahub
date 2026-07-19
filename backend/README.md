# Backend LicitaHub

API local em Go para o LicitaHub. Documentacao revisada em 18/07/2026.

## Responsabilidades

- Servir API HTTP e o frontend compilado em `../dist`.
- Autenticar sessoes e validar perfis de acesso.
- Executar consultas e migracoes no PostgreSQL por meio do `psql`.
- Servir arquivos locais em `/uploads/`.
- Processar OCR, certificados, captacao oficial e integracoes opcionais.
- Registrar acessos, erros e eventos operacionais no terminal e em arquivo.
- Executar rotinas periodicas, como encerramento de rodadas de avaliacao.

## Inicio local no Windows

Abra:

```powershell
.\backend\run-dev.cmd
```

O script:

1. Solicita `PGPASSWORD` quando ela nao estiver definida.
2. Configura o banco local `licitahub_dev`.
3. Verifica se `../dist/index.html` existe.
4. Compila o frontend quando necessario.
5. encerra executaveis locais antigos conhecidos.
6. Compila o backend atual como `licitahub-v98.exe`.
7. Inicia a API em `http://127.0.0.1:8080`.

Mantenha a janela aberta. Fechar a janela encerra o backend.

## Dependencias

- Go.
- PostgreSQL 17 e `psql.exe`.
- Node.js, quando for necessario gerar `dist/`.
- Tesseract OCR, para PDFs digitalizados ou imagens.
- Acesso de rede, apenas para integracoes externas.

## Variaveis de ambiente

Banco e servidor:

- `PSQL_PATH`: caminho do `psql.exe`.
- `PGHOST`: host do PostgreSQL.
- `PGPORT`: porta do PostgreSQL.
- `PGUSER`: usuario do PostgreSQL.
- `PGPASSWORD`: senha do PostgreSQL.
- `PGDATABASE`: banco da aplicacao.
- `APP_PORT`: porta HTTP.
- `PUBLIC_BASE_URL`: endereco usado em links e certificados.
- `LICITAHUB_LOG_PATH`: arquivo de logs.
- `LICITAHUB_BOOTSTRAP_ADMIN_PASSWORD`: senha inicial opcional do administrador local.

Integracoes:

- `OPENAI_API_KEY`: chave opcional da OpenAI.
- `OPENAI_TECHNICAL_ANALYSIS_MODEL`: modelo da OpenAI usado na analise de atestados.
- `GEMINI_API_KEY`: chave opcional do Google Gemini.
- `GEMINI_TECHNICAL_ANALYSIS_MODEL`: modelo do Gemini usado na analise de atestados.
- `GEMINI_TENDER_ANALYSIS_MODEL`: modelo de reserva usado na pre-analise dos editais.
- `OPENAI_CAPTURE_CLASSIFICATION_MODEL`: modelo principal usado na triagem da fila captada.
- `GEMINI_CAPTURE_CLASSIFICATION_MODEL`: modelo de reserva usado na triagem da fila captada.
- `GROQ_API_KEY`: chave opcional da Groq.
- `GROQ_MODEL`: modelo padrao da Groq.
- `GROQ_TECHNICAL_ANALYSIS_MODEL`: modelo da Groq usado na analise de atestados.
- `GROQ_CAPTURE_CLASSIFICATION_MODEL`: segundo modelo de reserva usado na triagem da fila captada.
- Variaveis de fontes oficiais e limites podem ser adicionadas por ambiente conforme a implantacao.

`backend/.env.example` nao contem segredos. Chaves e senhas reais devem ficar no ambiente do processo ou em gerenciador de segredos.

## Arquivos e diretorios

- `main.go`: servidor, autenticacao e recursos centrais.
- `assemblies.go`: Central de Montagem.
- `company_ratings.go`: rodadas anonimas de avaliacao.
- `academy.go`: cursos, aulas, progresso e provas.
- `academy_certificate.go`: certificado PDF e validacao.
- `technical_professionals.go`: profissionais tecnicos.
- `technical_certificates.go`: atestados e OCR.
- `technical_certificate_ai.go`: analise opcional de atestados por IA.
- `pncp.go`: captacao PNCP e Compras.gov.br.
- `pncp_ai.go`: classificacao em lotes da fila captada por OpenAI, com fallback para Gemini e Groq.
- `prompts/`: roteiros de IA mantidos pelo produto.
- `uploads/`: arquivos locais enviados.
- `logs/`: logs locais quando configurados.

## Grupos de rotas

Autenticacao e usuarios:

- `/health`
- `/api/auth/login`
- `/api/auth/session`
- `/api/auth/logout`
- `/api/auth/forgot-password`
- `/api/auth/reset-password`
- `/api/users/me`
- `/api/access-profiles`

Administracao e empresas:

- `/api/company-invitations`
- `/api/companies`
- `/api/company-users`
- `/api/admin/password-reset-requests`

Conteudo e comunidade:

- `/api/news`
- `/api/news/admin`
- `/api/community/posts`
- `/api/notifications`
- `/api/notification-history`

Editais e parcerias:

- `/api/tenders`
- `/api/pncp/captures`
- `/api/tender-challenges`
- `/api/partnership-ads`
- `/api/matches`
- `/api/chats`
- `/api/chats/stream`
- `/api/task-chats`
- `/api/direct-chats`

Area de trabalho:

- `/api/assemblies`
- `/api/assembly-list`
- `/api/assembly-calendar`
- `/api/my-assembly-tasks`

Capacidade tecnica:

- `/api/technical-professionals`
- `/api/technical-certificates`
- `/api/technical-certificate-ai-analyses`

Academia:

- `/api/academy/courses`
- `/api/academy/videos`
- `/api/academy/lessons`
- `/api/academy/my-learning`
- `/certificates/verify`

Avaliacao das associadas:

- `/api/company-ratings`
- `/api/company-ratings/allocations/{companyId}`
- `/api/company-ratings/submit`
- `/api/company-rating-results`
- `/api/admin/company-rating-rounds`
- `/api/admin/company-rating-results`

## Regras tecnicas importantes

- Toda rota administrativa valida o perfil no backend.
- Consultas privadas devem filtrar a empresa ou o usuario da sessao.
- Arquivos HTML enviados ou gerados devem ser exibidos de forma isolada e controlada.
- Uploads devem validar tamanho, extensao, tipo e permissao.
- Rodadas de avaliacao usam fotografia das empresas ativas na abertura.
- O saldo de estrelas e 30% das empresas participantes, arredondado para cima.
- Cada avaliadora pode concentrar qualquer quantidade em uma empresa, limitada pelo saldo total.
- O envio da distribuicao e idempotente: repetir uma confirmacao ja gravada nao duplica votos.
- Somente rodadas encerradas podem ser excluidas pelo administrador da plataforma.
- O fechamento de uma rodada ocorre quando todas concluem ou quando o prazo termina.
- A analise de atestados aceita selecao automatica, OpenAI, Google Gemini ou Groq.
- No modo automatico, a ordem e OpenAI, Google Gemini e Groq. O provedor seguinte assume quando ocorre falha tecnica, limite ou indisponibilidade.
- Cada resultado de IA registra e devolve explicitamente o provedor e o modelo que produziram a resposta.
- A Groq atua nos fluxos baseados em texto e JSON. A pre-analise de documentos de editais permanece em OpenAI e Gemini.

Captacao oficial:

- Cada comando do PNCP consulta uma pagina por modalidade selecionada, com ate 50 registros por modalidade.
- A pagina seguinte somente e consultada quando o administrador aciona o comando novamente.
- Resultados repetidos da mesma fonte sao atualizados pela chave de origem.
- PNCP e Compras.gov.br somente sao unificados entre si quando apresentam exatamente o mesmo numero de controle PNCP.
- A ordem das consultas nao altera o saneamento. O PNCP permanece como referencia e o Compras.gov.br complementa campos ausentes.

## Testes e compilacao

```powershell
cd backend
go test ./...
go build -o licitahub-v98.exe .
```

O frontend deve ser validado separadamente:

```powershell
cd ..
npm.cmd run build
```

## Logs

O backend registra:

- metodo e caminho HTTP;
- status;
- duracao;
- quantidade de bytes;
- endereco remoto;
- erros de banco e API;
- bloqueios de login;
- acoes sensiveis selecionadas, como exclusao de rodada.

Nao grave senhas, tokens, conteudo integral de documentos ou chaves nos logs.

## Limitacoes antes da producao

- Migrar senhas para hash forte com estrategia de transicao.
- Servir somente por HTTPS e habilitar cookie `Secure`.
- Substituir segredos locais por cofre de segredos.
- Definir armazenamento persistente para uploads.
- Configurar backup e restauracao do PostgreSQL.
- Implementar envio real de e-mail.
- Centralizar, rotacionar e monitorar logs.
- Ampliar testes automatizados e auditoria de acoes.
- Avaliar substituicao gradual das consultas montadas para uma camada de acesso parametrizada ao banco.

Consulte `../docs/SEGURANCA.md` antes de qualquer disponibilizacao publica.
