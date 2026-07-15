# Documentacao funcional do LicitaHub

O plano de preparacao para hospedagem e seguranca esta em `docs/SEGURANCA.md`.

## Objetivo

O LicitaHub e uma aplicacao web para empresas de engenharia consultiva. A plataforma combina rede social empresarial, divulgacao institucional, noticias da plataforma, editais publicos e formacao de parcerias/consorcios.

## Atualizacao desta entrega

Esta versao consolida os recursos mais recentes do produto:

- Participacao individual ou busca de parceiros em cada edital.
- Central de Montagem para participacoes individuais e consorciais.
- Calendario mensal de montagens, com progresso e acesso direto ao trabalho relacionado.
- Menu lateral recolhido por modulos e sino de alertas com destaque apenas quando houver novidade.
- Regra de lideranca para consorcios com duas ou tres empresas.
- Aviso automatico quando um edital suspenso volta a ser publicado, destinado as empresas que ja possuem relacao com ele.
- Roteiro manual completo de validacao em `docs/ROTEIRO-DE-TESTES.md`.

Antes de uma liberacao para usuarios externos, os cenarios de ciclo de vida do edital, acessos por perfil, match, consorcio e montagem devem ser executados conforme esse roteiro.

## Papeis de acesso

- Administrador da plataforma: administra convites, empresas, noticias e editais.
- Administrador da empresa: administra perfil da propria empresa e usuarios vinculados.
- Comercial: atua na comunidade, publicacoes, editais, interesses e matches.
- Tecnico: consulta informacoes, apoia analises tecnicas e acompanha editais conforme permissao.
- Leitor: visualiza conteudos permitidos, sem administracao.

As telas administrativas sao protegidas por perfil. O menu esconde o que o usuario nao pode acessar, e o backend tambem bloqueia chamadas administrativas indevidas.

## Modulo Acesso e Administracao

Uso principal: entrada controlada de empresas por convite.

Fluxo:

1. Administrador da plataforma cria convite para empresa.
2. Sistema gera link de convite.
3. Empresa acessa o link e preenche dados da empresa e do administrador principal.
4. Empresa fica aguardando analise.
5. Administrador da plataforma revisa a empresa.
6. Administrador pode aprovar, recusar ou solicitar ajuste.
7. Empresa aprovada passa a acessar o sistema.

Regras principais:

- CNPJ e nome da empresa devem ser unicos.
- Telefone e CNPJ sao obrigatorios no convite.
- Email pode repetir.
- CNPJ, email e telefone possuem validacao no formulario.
- Contato principal deve ter pelo menos nome e sobrenome.
- Ao concluir o aceite do convite com sucesso, a empresa retorna para o login com mensagem de cadastro concluido.
- Somente administrador da plataforma gerencia convites e analise de empresas.
- Telas que dependem de um registro especifico, como analise de empresa, nao ficam no menu lateral.
- Empresas cadastradas: o administrador da plataforma pode filtrar empresas ativas ou bloqueadas e bloquear/desbloquear uma empresa inteira.
- Ao bloquear uma empresa, todas as sessoes dos usuarios vinculados sao encerradas e nenhum deles volta a acessar enquanto a empresa permanecer bloqueada. Usuarios e historicos nao sao apagados.
- Cada bloqueio ou desbloqueio registra o administrador responsavel, a data, o status aplicado e o motivo opcional na auditoria interna.

## Modulo Empresa

Uso principal: manter a presenca institucional da empresa.

### Painel da empresa

Uso principal: concentrar o acompanhamento diario sem criar ou alterar registros.

- Indicadores de editais acompanhados, anuncios ativos, consorcios ativos e tarefas abertas.
- Prioridades: tarefas atrasadas, proximas do prazo, aguardando informacao, devolvidas para ajuste, editais com sessao proxima e novidades ainda nao lidas.
- Listas clicaveis de editais com interesse, consorcios ativos, tarefas sob responsabilidade da empresa e atividade recente.
- Filtros fixos por area do painel e periodo das novidades: 7, 30 ou 90 dias.
- Cada item leva para sua tela de origem; o painel funciona como consulta e navegacao, sem alterar dados operacionais.

Telas principais:

- Dashboard da empresa: mantido como area resumida, com evolucao posterior.
- Editar perfil da empresa.
- Usuarios vinculados.
- Cadastro/edicao/bloqueio/desbloqueio/remocao de usuario.
- Meu perfil.
- Perfil publico da empresa.

Perfil da empresa:

- Site.
- Descricao institucional.
- Porte.
- Cidade e UF.
- Atuacao nacional.
- Logomarca.

Regras principais:

- Somente administrador da empresa edita o perfil da empresa.
- Usuario comum edita apenas seus proprios dados permitidos.
- Cargo e perfil de acesso de usuario vinculado sao alterados pelo administrador da empresa.
- Usuarios vinculados aparecem apenas dentro da propria empresa.
- Lista de usuarios vinculados possui filtros fixos por busca, perfil e status.
- O perfil publico apresenta profissionais ativos em area retratil. Ao selecionar um profissional, a comunidade ve foto, nome, cargo, e-mail e telefone institucionais; senha, perfil de acesso e dados internos nao sao exibidos.
- Na comunidade, o nome e a logomarca de cada empresa nas publicacoes abrem o perfil publico daquela empresa. A vitrine de parceiros e a lista de empresas interessadas tambem oferecem acesso ao perfil da anunciante. Os comandos de criar e administrar publicacoes aparecem apenas no perfil da propria empresa logada.

## Modulo Comunidade

Uso principal: rede social empresarial entre empresas participantes.

Funcionalidades:

- Criar publicacao com categoria, texto e imagem.
- Ver publicacoes de todas as empresas.
- Filtrar por tipo, regiao e nome da empresa.
- Curtir publicacoes.
- Favoritar publicacoes.
- Comentar publicacoes.
- Ver contador de curtidas.
- Abrir quem curtiu.
- Expandir/recolher comentarios.
- Editar/excluir comentarios quando permitido.
- Gerenciar minhas publicacoes.
- Editar, arquivar ou excluir publicacoes da propria empresa.

Categorias previstas:

- Equipe comercial.
- Noticias.
- Atividades.
- Eventos.
- Conquistas.
- Conteudo tecnico.
- Destaque.

Regras principais:

- A comunidade mostra publicacoes visiveis para a rede.
- Minhas publicacoes mostra o conteudo da empresa logada.
- Imagens devem se ajustar ao card sem deformar a interface.
- Curtidas e comentarios geram notificacoes no sino.

## Perfil publico da empresa

Uso principal: pagina publica interna da empresa dentro da comunidade.

Conteudos:

- Logo.
- Nome da empresa.
- Descricao institucional.
- Porte.
- Site.
- Cidade/UF.
- Atuacao nacional.
- Publicacoes da empresa.
- Profissionais cadastrados em area retratil.

## Modulo Radar LicitaHub

Uso principal: noticias e comunicados publicados pela plataforma.

Funcionalidades:

- Administrador da plataforma cadastra noticia.
- Noticia pode ser rascunho, disponivel, destaque, arquivada ou expirada.
- Noticia possui imagem, resumo, texto e data final de publicacao.
- Radar mostra noticia principal e cards de outras noticias.
- Clique no card abre detalhe da noticia.
- Gerenciamento permite mudar status e filtrar noticias.
- Radar possui filtro fixo e paginacao para evitar lista longa demais.
- Cadastro de noticia possui limite de caracteres para titulo e resumo.

Regras principais:

- Somente administrador da plataforma cadastra e gerencia noticias.
- Noticia com data vencida nao deve aparecer no Radar publico.
- Apenas uma noticia deve atuar como destaque principal quando essa regra for aplicada.
- Publicacao de noticia gera notificacao.

## Modulo Editais

- O administrador da plataforma pode anexar quantidade livre de arquivos a cada edital, inclusive no cadastro inicial ou posteriormente no detalhe. Os documentos ficam disponíveis para download aos usuários da plataforma; o limite operacional é de 25 MB por arquivo.

Uso principal: centralizar editais captados no mercado de licitacao publica.

Administrador da plataforma cadastra:

- Orgao.
- Numero.
- Objeto.
- Modalidade.
- Criterio de julgamento quando aplicavel.
- Valor estimado.
- Cidade/UF.
- Data da sessao.
- Status.
- Link de pasta/arquivos em nuvem.
- Arquivo HTML de pre-analise.

Fluxo da empresa:

1. Empresa abre lista de editais.
2. Empresa abre detalhe do edital.
3. Empresa le a pre-analise HTML, se existir.
4. Empresa acessa link da pasta do edital.
5. Empresa registra interesse e define sua estrategia: participar sozinha, buscar parceiros ou apenas acompanhar.
6. Empresa informa o que tem, o que atende parcialmente, o que nao se aplica e o que busca em parceiros.
7. Apenas a estrategia de buscar parceiros gera anuncio para empresas interessadas e vitrine.
8. A participacao individual permanece privada e pode iniciar a Central de Montagem com os profissionais da propria empresa.
9. Enquanto nao houver consorcio fechado, a empresa pode trocar a estrategia para buscar parceiros ou desistir da participacao. Ao sair da participacao individual, a montagem e as tarefas ficam bloqueadas, encerradas e fora da operacao; o historico permanece guardado. A desistência encerra o anuncio e a montagem individual associada.

Regras principais:

- Somente administrador da plataforma cadastra, edita ou exclui editais.
- Status do edital inclui impugnado.
- Valor estimado usa mascara monetaria.
- UF usa lista de estados brasileiros.
- Cidade pode ser preenchida com apoio por UF.
- Se nao houver analise HTML, detalhe do edital mostra aviso de edital ainda nao analisado.
- Editais com data de sessao passada podem mudar para status ocorrido.
- Lista de editais indica quando a empresa ja registrou interesse e deixa de exibir o botao de registrar novamente.
- Tela de detalhe/interesse precisa receber um edital especifico; por isso nao deve ser usada como item direto de menu.
- Quando edital suspenso volta para publicado, usuarios ativos das empresas com interesse, anuncio, consorcio ou montagem relacionada recebem notificacao com atalho para o detalhe do edital.

### Calendario de montagens

Uso principal: acompanhar, por mes, todas as licitacoes que a empresa esta montando sozinha ou em consorcio.

- Cada card fica no dia da sessao do edital e abre a Central de Montagem correspondente.
- O card mostra numero, orgao, objeto resumido, tipo de participacao e percentual concluido.
- A barra de percentual usa vermelho ate 30%, amarelo entre 31% e 70% e verde acima de 70%.
- O filtro permite ver todas as montagens, somente individuais ou somente consorcios.

## Empresas interessadas

Uso principal: listar empresas interessadas em um edital especifico.

Conteudos:

- Empresa interessada.
- Logo.
- Resumo do anuncio.
- O que possui.
- O que busca.
- Botao para detalhe do anuncio.
- Botao para avaliacao.

Regras principais:

- Deve mostrar apenas empresas interessadas naquele edital.
- O proprio anuncio da empresa logada aparece marcado como "Meu anuncio".
- A empresa pode revisar seu proprio anuncio, mas nao pode avaliar/gerar consorcio consigo mesma.

## Vitrine de parceiros

Uso principal: pagina geral de classificados de parceria.

Diferenca para empresas interessadas:

- Empresas interessadas: mostra anuncios de um edital especifico.
- Vitrine de parceiros: mostra anuncios de todos os editais.

Filtros previstos:

- Licitacao.
- Orgao.
- Empresa.
- Regiao/UF.
- Status.

Organizacao da tela:

- A vitrine principal apresenta anuncios de outras empresas.
- A area retratil **Meus anuncios** separa os anuncios da empresa logada.
- Anuncio proprio de empresa pode ter resumo de oferta e busca editados, ou ser encerrado/excluido da vitrine sem apagar o historico.
- Anuncio de consorcio e administrado pela empresa lider em **Meus consorcios**.

Regras principais:

- O proprio anuncio nao se mistura aos anuncios externos da vitrine.
- O botao de avaliacao nao aparece no proprio anuncio.
- Para avaliar outra empresa no mesmo edital, a empresa precisa ter manifestacao/anuncio ativo para aquela licitacao.
- Depois de registrar uma avaliacao positiva, a acao passa a aparecer como avaliacao registrada, evitando repeticao de likes.

## Detalhe do anuncio

Uso principal: apresentar a empresa candidata em profundidade.

Conteudos:

- Empresa.
- Edital.
- Objeto.
- O que atende.
- O que atende parcialmente.
- O que nao se aplica.
- O que busca.
- Observacoes.

Regra principal:

- Quando o anuncio pertence a empresa logada, o detalhe nao mostra acao para avaliar candidata.
- Quando o anuncio pertence a outra empresa, o nome da anunciante ou da lider do consorcio exibe um icone de chat para abrir a conversa geral ja existente entre as empresas.

## Avaliacao, match e consorcio

Fluxo:

1. Empresa A avalia anuncio da empresa B.
2. Empresa A aprova, recusa ou deixa para depois.
3. Empresa B avalia anuncio da empresa A.
4. Quando ha aprovacao reciproca, o sistema gera match.
5. Match aparece em meus consorcios.
6. Sistema disponibiliza contato/WhatsApp.
7. Empresas podem definir lider do consorcio.
8. Ao fechar match/consorcio, anuncios das empresas consorciadas saem da vitrine publica daquele edital.
9. A lider pode publicar anuncio de consorcio buscando empresa complementar.
10. Uma terceira empresa interessada registra candidatura e o aceite reciproco com a lider a inclui no consorcio.

Regras principais:

- Uma empresa nao pode avaliar o proprio anuncio.
- Uma empresa nao pode gerar consorcio consigo mesma.
- A composicao e registrada por membros ativos, permitindo mais de duas empresas na estrutura de dados.
- Na regra atual do produto, a busca complementar e encerrada quando o consorcio alcanca tres empresas. A expansao para quarta ou mais empresas podera ser liberada futuramente pela lider.
- Administrador ou comercial de qualquer empresa consorciada ativa pode definir a lideranca. A criacao de anuncio complementar e o aceite de candidatura permanecem centralizados na empresa lider; a desistencia do consorcio e exclusiva do administrador da empresa.
- Ao desistir, a empresa fica marcada como retirada, com data e usuario responsavel. Ela deixa de ver o consorcio ativo.
- Se a lider desistir e permanecerem pelo menos duas empresas, outra lider deve ser definida antes da saida. Se restar menos de duas empresas, o consorcio e encerrado.

## Chat de parceria

Uso principal: permitir alinhamento entre empresas antes da avaliacao e do match.

Funcionalidades:

- Botao **Conversar** em anuncios da vitrine e empresas interessadas.
- Icone de chat no detalhe do anuncio para conversar diretamente com a empresa anunciante ou com a lider do anuncio de consorcio.
- Janela flutuante que pode ser minimizada.
- Mais de uma conversa simultanea.
- Mensagens gravadas no banco e entregues em tempo real enquanto o backend estiver conectado.
- Alerta visual e sonoro para nova mensagem, quando permitido pelo navegador.

Regras principais:

- Administrador da empresa e comercial podem usar o chat em nome da empresa.
- Conversas ficam vinculadas ao edital e aos participantes do anuncio.
- Quando um consorcio e fechado, a conversa relacionada passa a ser usada apenas para alinhamento da composicao e novas conversas sobre aquele anuncio podem ser bloqueadas conforme o status.

## Notificacoes no sino

Eventos previstos:

- Nova noticia publicada.
- Novo edital publicado.
- Pre-analise de edital adicionada.
- Empresa interessada no mesmo edital.
- Curtida em publicacao.
- Comentario em publicacao.
- Convite aceito/empresa aguardando analise.
- Empresa aprovada ou recusada.
- Match realizado.
- Lider de consorcio definido.
- Candidatura ao consorcio, entrada aprovada e desistencia de empresa.
- Nova mensagem de chat.

Regra principal:

- O alerta aparece para o usuario destinatario.
- Ao abrir o sino, o contador zera.
- Alertas lidos nao voltam como novos para aquele usuario.

## Banco de dados

Principais grupos:

- Acesso e administracao: empresas, convites, usuarios, sessoes e tokens.
- Empresa: perfil da empresa, areas tecnicas e midias.
- Radar: categorias e noticias.
- Comunidade: categorias, publicacoes, imagens, curtidas, comentarios e favoritos.
- Editais: editais, arquivos, requisitos, interesses e requisitos do interesse.
- Impugnacoes: pedido da empresa licitante, fundamentacao, anexos, protocolo e notificacao ao administrador da plataforma.
- IA de editais: historico de pre-analises, documentos de origem, modelo utilizado, resposta e falhas de processamento.
- Match/consorcio: anuncios, avaliacoes, matches, contatos, intencoes, membros, candidaturas e chat.
- Central de Montagem: modelos, fases fixas, tarefas, responsaveis, comentarios, evidencias, prazos e historico operacional.
- Notificacoes: alertas por usuario/empresa.
- Auditoria: estrutura existe, mas uso completo fica para fase posterior.

## Itens deixados para fase posterior

- Dashboards analiticos completos.
- Auditorias completas de acoes sensiveis.
- Backend de producao com autenticacao mais robusta.
- Hash de senha adequado para producao.
- Logs estruturados.
- Configuracao por ambiente.
- Deploy externo.
- Evolucao para aplicativo mobile.
- Integracao futura com outras IAs e automacoes avancadas.

## Pre-analise de Editais com IA

O administrador da plataforma pode anexar quantos documentos forem necessarios ao edital e, no detalhe do edital, solicitar a geracao da pre-analise tecnica em HTML.

- O roteiro oficial esta em `backend/prompts/analise-primaria-edital.txt` e foi fornecido pela LicitaHub.
- O HTML gerado aparece no proprio detalhe do edital, pode ser baixado e continua podendo ser substituido por um HTML manual.
- Cada tentativa fica registrada com situacao: aguardando, em processamento, concluida ou falhou.
- O processamento usa somente os documentos diretamente anexados ao edital. Arquivos compactados e arquivos tecnicos sem conteudo textual nao entram na analise automatica.
- A analise deve receber arquivos de ate 22 MB cada e 50 MB somados em uma unica solicitacao. Para maior fidelidade de tabelas, imagens e diagramas, prefira documentos em PDF.
- A chave da API fica apenas no arquivo local `backend/.env.openai`, que e ignorado pelo Git. Ela nunca deve ser adicionada ao frontend ou enviada ao repositorio.
- Para configurar no computador local, abra `backend/CONFIGURAR-OPENAI.cmd` e depois reinicie `backend/run-dev.cmd`.

## Pedido de Impugnacao de Edital

No detalhe de um edital publicado, em analise ou ja marcado como impugnado, usuarios administrador, comercial ou tecnico da empresa podem abrir um pedido de impugnacao.

- O pedido recebe assunto, fundamentacao e anexos de apoio.
- Cada empresa possui um pedido por edital, que pode ser atualizado enquanto estiver protocolado.
- O protocolo gera aviso para os administradores da plataforma.
- Protocolar o pedido **nao** muda automaticamente o status do edital para `Impugnado`; essa decisao continua sendo administrativa.
- O pedido e seus anexos ficam registrados no banco para consulta posterior.
- Se o protocolo for feito com menos de tres dias uteis antes da sessao, o sistema mostra o alerta do Art. 164 e registra a marca de pedido intempestivo, sem bloquear a empresa.
- O administrador da plataforma possui a tela `Impugnacoes`, organizada como Kanban em `Protocolados`, `Em analise` e `Concluidos`.
- Cada cartao mostra empresa, edital, sessao, prazo interno calculado para seis dias antes da sessao, fundamentacao e anexos para download.
- O administrador pode mudar o andamento para protocolado, em analise, procedente, improcedente ou retirado. A empresa solicitante recebe uma notificacao da atualizacao.

## Central de Montagem da Licitacao

A Central de Montagem nasce em `Meus consorcios`, depois que a empresa lider foi definida, ou pela participacao individual da empresa no detalhe do edital. Nao e um Kanban convencional: as tarefas nao mudam de fase. Cada tarefa permanece na fase a que pertence e evolui por status.

O Modelo LicitaHub inicia cada montagem com oito fases:

1. Planejamento da montagem.
2. Concepcao consorcial.
3. Montagem da peca qualitativa.
4. Montagem do orcamento.
5. Montagem da equipe tecnica.
6. Montagem das declaracoes.
7. Certificacoes e quesitos de pontuacao.
8. Revisao e consolidacao.

Regras principais:

- O administrador ou comercial da empresa lider inicia a montagem do consorcio. Na participacao individual, o administrador ou comercial da propria empresa inicia e coordena a montagem.
- Todas as empresas ativas do consorcio podem acompanhar o painel.
- A lider cria fases complementares, tarefas, prazos e atribuicoes.
- O responsavel pode atualizar a tarefa e envia-la para revisao; a conclusao definitiva fica sob validacao da lider.
- Profissionais ativos de qualquer empresa consorciada podem ser responsaveis. Na montagem individual, somente profissionais ativos da propria empresa aparecem para atribuicao.
- `Nao se aplica` retira o peso da tarefa do percentual da fase.
- Todo prazo da montagem deve estar entre a data atual e a data de abertura do edital; sem data de abertura, o prazo fica indisponivel.
- `Minhas tarefas` reune, em um Kanban pessoal, somente as tarefas atribuidas diretamente ao profissional em todas as montagens de que sua empresa participa. Cada cartao identifica o edital e abre a tarefa na Central de Montagem correspondente.
- Ao clicar em um cartao de `Minhas tarefas`, o profissional abre uma divisao lateral na propria tela para atualizar o status, comentar e incluir documentos, links ou anotacoes. Ele nao altera a estrutura da tarefa nessa tela.
- O chat flutuante atende conversas de anuncios e tarefas. A conversa de tarefa tem historico proprio e e acessivel apenas ao profissional responsavel e aos usuarios de coordenacao da empresa lider.
- No perfil publico da empresa, um usuario pode abrir uma conversa direta com um profissional vinculado. O historico e privado e acessivel somente aos dois usuarios participantes.
- Somente o administrador ou comercial da empresa lider pode criar, alterar a estrutura ou excluir tarefas. A exclusao remove tambem os comentarios e documentos vinculados a tarefa.
- Comentarios, arquivos, links e anotacoes ficam vinculados a tarefa.
- O dossie consolida automaticamente as evidencias por fase e tarefa.
- Prazos vencidos ou proximos do vencimento geram alertas internos sem duplicacao diaria.
- A saida de uma empresa do consorcio preserva o historico e permite redistribuir suas tarefas.

## Roteiro de teste manual

O teste manual deve passar por:

1. Login com todos os perfis.
2. Convite e aprovacao de empresa.
3. Edicao de perfil da empresa.
4. Criacao e gestao de usuarios, incluindo bloqueio, desbloqueio e desativacao de vinculo.
5. Bloqueio e desbloqueio de empresa pelo administrador da plataforma, confirmando a perda e a restauracao de acesso para todos os usuarios vinculados.
6. Criacao e interacao na comunidade.
7. Perfil publico da empresa.
8. Cadastro e gerenciamento de noticias.
9. Cadastro e detalhe de editais.
10. Registro de interesse.
11. Empresas interessadas por edital.
12. Vitrine geral.
13. Detalhe de anuncio e abertura de conversa pelo icone de chat.
14. Avaliacao reciproca.
15. Match, definicao de lider e anuncio para busca complementar.
16. Inclusao de terceira empresa e encerramento dos anuncios envolvidos.
17. Desistencia de empresa e troca de lider, quando aplicavel.
18. Chat entre empresas antes do match.
19. Notificacoes no sino.
20. Tentativas de acesso indevido por usuario comum.
21. Inicio da Central de Montagem pela empresa lider.
22. Atribuicao de tarefas a empresas e profissionais consorciados.
23. Atualizacao de status, revisao, comentarios, evidencias e dossie final.
24. Protocolo de pedido de impugnacao com assunto, fundamentacao e anexos.
25. Protocolo com menos de tres dias uteis antes da sessao, confirmando o alerta de intempestividade e o registro do pedido.
26. Consulta, download de anexos e atualizacao de andamento na Central de Impugnacoes pelo administrador da plataforma.

## Situacao dos testes manuais

Rodada de testes manuais encerrada em 11/07/2026.

Fluxos principais validados em ambiente local:

- Login e perfis.
- Convite, aceite e aprovacao de empresa.
- Gestao de usuarios vinculados.
- Radar LicitaHub.
- Comunidade e perfil publico.
- Cadastro e administracao de editais.
- Registro de interesse.
- Empresas interessadas.
- Vitrine de parceiros.
- Detalhe de anuncio.
- Avaliacao, match e meus consorcios.

Novas funcionalidades e ajustes finos serao tratados em rodadas posteriores.
