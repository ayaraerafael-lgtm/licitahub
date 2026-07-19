# Mapa de Seguranca do LicitaHub

Documentacao revisada em 18/07/2026.

## Objetivo

Preparar o LicitaHub para uso online sem alterar os fluxos funcionais ja aprovados: convite, login, empresa, comunidade, noticias, editais, interesse, consorcios, montagem, chat e impugnacoes.

O trabalho deve ser feito por etapas pequenas, com testes antes e depois de cada mudanca. Nenhuma tela precisa ser redesenhada para a seguranca ser aplicada.

## Situacao atual

Controles ja presentes no ambiente local:

- Perfis de acesso e verificacoes de autorizacao no backend para rotas administrativas.
- Isolamento por empresa aplicado nos principais fluxos privados.
- Sessoes por cookie `HttpOnly` e `SameSite=Lax`.
- Bloqueio temporario do login depois de 15 tentativas em 15 minutos, com contador de tentativas restantes exibido a partir do terceiro erro.
- Encerramento de sessoes quando usuario ou empresa e bloqueado e nos fluxos sensiveis de redefinicao.
- Tokens de convite e recuperacao com validade e controle de uso.
- Logs estruturados de acesso HTTP, falhas da API, banco, autenticacao e eventos sensiveis selecionados.
- Segredos de banco e de integracoes lidos por variaveis locais, sem incorporacao deliberada ao frontend.
- Validacoes funcionais de tipo, tamanho e permissao nos fluxos de upload mais recentes.

Pendencias que bloqueiam uma publicacao profissional:

- As senhas ainda precisam ser migradas para hash forte, preferencialmente `Argon2id`.
- O cookie deve receber `Secure` quando o sistema estiver sob HTTPS.
- A camada de banco baseada em comandos `psql` e SQL montado deve migrar gradualmente para conexao nativa e consultas parametrizadas.
- Os uploads locais precisam de armazenamento persistente privado, verificacao antimalware e downloads autorizados.
- Backup automatico, restauracao testada, centralizacao de logs e monitoramento ainda precisam ser implantados no ambiente de hospedagem.
- Recuperacao de senha ainda depende de atendimento administrativo; o envio seguro por e-mail precisa ser configurado.
- A bateria automatizada de seguranca e regressao ainda nao cobre integralmente todos os modulos.

Portanto, o LicitaHub esta em condicao de desenvolvimento e homologacao local, mas esta documentacao nao classifica a versao atual como pronta para exposicao publica.

## Principios

1. O navegador nunca define permissao: toda permissao e confirmada no backend.
2. Um identificador na URL localiza um registro, mas nunca autoriza acesso a ele.
3. Dados e comandos SQL devem ser separados por consultas parametrizadas.
4. Arquivos, senhas, tokens e chaves nunca ficam expostos em URL, Git ou logs.
5. Cada empresa so acessa seus dados privados; conteudo publico segue regras proprias.
6. Toda mudanca sensivel deixa registro de auditoria.

## Fase 0 - Linha de base e protecao da mudanca

Antes de alterar o backend:

- Fazer backup do banco e testar a restauracao em banco separado.
- Registrar uma bateria de testes manuais dos fluxos atuais.
- Criar testes automatizados para login, convite, empresa, usuarios, editais, interesse, anuncio, consorcio, chat, montagem e impugnacao.
- Criar ambiente de homologacao separado do banco de desenvolvimento e do banco futuro de producao.
- Adotar migracoes versionadas de banco. Nenhuma alteracao deve exigir apagar dados existentes.

Regra de seguranca da mudanca: cada entrega precisa compilar, executar testes, manter os contratos das APIs e ter plano de retorno.

## Fase 1 - Banco de dados e prevencao de SQL Injection

Situacao a evoluir: o backend atual monta diversas consultas SQL como texto e as executa pelo `psql`. Isso funciona localmente, mas nao e a arquitetura indicada para ambiente online.

Acao profissional:

- Substituir a execucao por `psql` por conexao nativa do Go ao PostgreSQL, usando `pgxpool`.
- Trocar todas as interpolacoes de dados por consultas parametrizadas: `WHERE id = $1`, com valores enviados separadamente.
- Usar transacoes para operacoes que gravam varias tabelas, como aceite de convite, match, consorcio, montagem e impugnacao.
- Validar por lista permitida os poucos campos que nao podem ser parametros, como ordenacao e status predefinidos.
- Criar usuario de banco exclusivo da aplicacao, sem permissao de superusuario, criacao de banco ou acesso administrativo.

Resultado: caracteres maliciosos enviados em campos passam a ser tratados somente como texto, nunca como comando SQL.

Referencia: [OWASP - SQL Injection Prevention](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html).

## Fase 2 - Autenticacao, senha e sessao

- Migrar senhas para `Argon2id` com salt individual e parametros adequados; manter compatibilidade temporaria para que usuarios existentes atualizem a senha no proximo login.
- Exigir senha minima forte, bloquear senhas muito comuns e aplicar limite de tentativas por IP e por conta. O ambiente atual usa bloqueio temporario apos 15 tentativas em 15 minutos e informa tentativas restantes a partir do terceiro erro.
- Usar cookies de sessao com `HttpOnly`, `Secure` e `SameSite` adequados quando houver HTTPS.
- Renovar o identificador de sessao apos login e troca de senha; encerrar todas as sessoes ao bloquear uma empresa, bloquear/remover usuario ou trocar senha. O bloqueio de empresa ja encerra as sessoes de seus usuarios no ambiente local.
- Recuperacao de senha com token aleatorio, uso unico, expiracao curta e resposta generica para nao revelar se um email existe.
- Convites com token aleatorio, expiracao, uso unico e invalidacao apos aceite, recusa ou cancelamento.

Resultado: reduz roubo de sessao, tentativa automatizada de senha e uso indevido de links antigos.

Referencias: [OWASP - Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html) e [OWASP - Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).

## Fase 3 - Permissoes e isolamento entre empresas

Para cada rota, revisar duas perguntas antes de devolver ou alterar qualquer dado:

1. O usuario possui o perfil funcional necessario?
2. O registro pertence a empresa dele ou e realmente publico para aquele perfil?

Aplicacoes praticas:

- Administrador da plataforma: convites, empresas, noticias, editais e impugnacoes.
- Administrador da empresa: perfil empresarial e usuarios da propria empresa.
- Comercial e tecnico: somente as operacoes previstas para o perfil e empresa vinculada.
- Leitor: apenas consulta autorizada.
- Chat: somente participantes da conversa, membros autorizados do consorcio ou profissionais especificamente envolvidos em tarefa.
- Arquivos: permissao conferida antes de gerar download.
- Consorcio: empresa membro ativa nao pode abrir, por URL direta ou API, anuncios, chat ou avaliacao de terceiros no mesmo edital. A unica excecao e a empresa lider apos publicar anuncio complementar de busca por nova consorciada.

Testes obrigatorios:

- Alterar IDs manualmente na URL e na requisicao para tentar acessar empresa, arquivo, chat, anuncio ou tarefa de terceiros.
- Tentar abrir rota administrativa com usuario comum.
- Tentar alterar lider, membro, status ou documento de consorcio de outra empresa.

Resultado: mesmo que uma pessoa conheca um ID, o backend nega o acesso sem permissao.

## Fase 4 - Arquivos, HTML e conteudo enviado

- Guardar anexos em armazenamento privado, fora da pasta publica do servidor.
- Liberar downloads por rota autenticada ou URL assinada de curta duracao.
- Gerar nome interno aleatorio para cada arquivo; nunca usar o nome original como caminho.
- Aceitar somente extensoes e tipos permitidos, confirmar a assinatura real do arquivo e limitar tamanho e quantidade.
- Colocar arquivos novos em quarentena e integrar verificacao antimalware antes de disponibilizar download.
- Tratar o HTML de pre-analise como conteudo nao confiavel: sanitizar, impedir scripts e servir em origem separada quando houver hospedagem.
- Manter limites especificos para upload de edital, impugnacao, comunidade e montagem.

Resultado: reduz envio de arquivo malicioso, exposicao direta de documentos e execucao de HTML enviado por usuarios.

Referencia: [OWASP - File Upload Security](https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html).

## Fase 5 - Protecao da aplicacao online

- Hospedar atras de HTTPS com certificado valido e redirecionar todo acesso HTTP para HTTPS.
- Configurar cabecalhos de seguranca: Content-Security-Policy, HSTS, X-Frame-Options, X-Content-Type-Options e Referrer-Policy.
- Restringir CORS somente ao dominio oficial do LicitaHub.
- Proteger requisicoes que alteram dados contra CSRF quando a sessao usar cookie.
- Limitar tamanho de requisicoes e aplicar rate limit para login, convite, upload, chat, IA e endpoints de consulta intensiva.
- Configurar paginas de erro sem expor SQL, caminho interno, senha, token ou stack trace.
- Usar protecao contra ataques de forca bruta e abuso automatizado no login e no chat.

## Fase 6 - Segredos, operacao e privacidade

- Manter senha do banco, chaves da OpenAI, Google Gemini e Groq e demais segredos somente em gerenciador de segredos ou variaveis do ambiente de hospedagem.
- Nunca enviar segredos ao GitHub, frontend, URL, mensagens de erro ou log.
- Usar banco gerenciado com backup automatico, criptografia, restauracao testada e acesso restrito por rede.
- Logs estruturados foram iniciados no backend: acessos HTTP, erros da API, falhas de banco, login, logout e tentativas de login recusadas.
- A auditoria funcional existente registra acoes sensiveis com usuario, empresa, data, registro afetado e metadados; a cobertura deve continuar sendo ampliada para novos fluxos.
- No ambiente local, o link temporario de recuperacao fica armazenado para atendimento autorizado. Em producao, essa armazenagem deve ser removida e substituida por envio seguro por e-mail, sem exibir o token no painel.
- Definir politica de retencao de documentos e dados pessoais, termos de uso e politica de privacidade alinhados a LGPD.
- Atualizar dependencias e verificar vulnerabilidades antes de cada entrega.

## Fase 7 - Validacao antes do piloto online

1. Rodar testes automatizados e roteiro manual completo.
2. Fazer revisao de codigo focada em autorizacao, SQL, upload, chat e tokens.
3. Executar teste de seguranca controlado em homologacao, sem dados reais.
4. Corrigir achados de gravidade alta e critica antes de liberar acesso externo.
5. Publicar inicialmente para um grupo pequeno de empresas convidadas.
6. Monitorar erros, tentativas de acesso negadas, uploads e consumo da IA.
7. Ter plano de resposta: backup, restauracao, contato tecnico e comunicacao aos usuarios.

## Ordem recomendada

1. Fase 0 e Fase 1.
2. Fase 2 e Fase 3.
3. Fase 4 e Fase 5.
4. Fase 6 e Fase 7.

Somente apos essas etapas o LicitaHub deve seguir para um piloto online com usuarios reais.
