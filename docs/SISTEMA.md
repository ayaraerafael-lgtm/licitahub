# Documentacao funcional do LicitaHub

O plano de preparacao para hospedagem e seguranca esta em `docs/SEGURANCA.md`.

Documentacao revisada em 18/07/2026.

## Objetivo

O LicitaHub e uma aplicacao web para empresas de engenharia consultiva. A plataforma combina rede social empresarial, divulgacao institucional, noticias da plataforma, editais publicos e formacao de parcerias/consorcios.

## Estado funcional consolidado

Esta versao consolida os recursos mais recentes do produto:

- Participacao individual ou busca de parceiros em cada edital.
- Central de Montagem para participacoes individuais e consorciais.
- Calendario mensal de montagens, com progresso e acesso direto ao trabalho relacionado.
- Menu lateral recolhido por modulos e sino de alertas com destaque apenas quando houver novidade.
- Regra de lideranca para consorcios com duas ou tres empresas.
- Aviso automatico quando um edital suspenso volta a ser publicado, destinado as empresas que ja possuem relacao com ele.
- Comunidade reorganizada com filtros laterais fixos, pesquisa imediata por empresa e cards de publicacao com imagem proporcional.
- Publicacoes arquivadas deixam a comunidade, mas continuam em **Minhas publicacoes** para consulta, edicao e eventual reativacao.
- Modulo de Capacidade Tecnica para profissionais, atestados, leitura de PDF, OCR e analises assistidas por IA.
- Academia LicitaHub com cursos, videoaulas externas, progresso individual, provas de liberacao e certificados verificaveis.
- Avaliacao anonima das associadas por rodadas, com distribuicao integral de estrelas, resultado por sessao, historico proporcional e grafico de tendencia.
- Regra de exclusividade operacional no consorcio: membro ativo nao busca nova parceria no mesmo edital, exceto a lider quando abre anuncio de consorcio para complementar a composicao.
- Roteiro manual completo de validacao em `docs/ROTEIRO-DE-TESTES.md`.
- Mapa comercial e operacional para demonstracao, apostila e slides em `docs/MAPA-DE-APRESENTACAO-E-TREINAMENTO.md`.

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
- A redefinicao de senha nao reativa automaticamente usuario inativo ou bloqueado. A tela confirma a troca da senha e informa claramente que o administrador da empresa precisa liberar o acesso.
- O login diferencia credencial incorreta, usuario bloqueado, usuario inativo e empresa sem acesso liberado.

## Modulo Empresa

Uso principal: manter a presenca institucional da empresa.

Telas principais:

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
- Filtros laterais fixos por empresa, categoria e UF, com busca por nome aplicada enquanto o usuario digita.

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
- Publicacoes arquivadas nao aparecem na comunidade publica, mas permanecem em Minhas publicacoes para que a empresa possa reativa-las, edita-las ou exclui-las.
- Minhas publicacoes mostra o conteudo da empresa logada, incluindo itens arquivados.
- Imagens devem se ajustar ao card sem deformar a interface, respeitando altura maxima e preservando a proporcao original.
- Curtidas e comentarios geram notificacoes no sino.

### Avaliacao de parcerias entre associadas

Uso principal: medir como o mercado associado percebe a abertura de cada empresa para realizar negocios, parcerias e composicoes empresariais.

Fluxo:

1. O administrador da plataforma abre uma rodada e define obrigatoriamente seu prazo de encerramento.
2. O sistema registra uma fotografia das empresas ativas que participarao daquela sessao e envia um alerta no sino a todos os usuarios dessas empresas, informando o prazo.
3. Cada empresa recebe um saldo equivalente a 30% da quantidade de participantes, arredondado para cima.
4. O administrador da empresa distribui livremente o saldo entre outras associadas. Uma unica empresa pode receber qualquer quantidade, limitada apenas pelo saldo ainda disponivel.
5. A distribuicao somente pode ser enviada quando todo o saldo tiver sido utilizado. O comando fixo de conclusao mostra as estrelas restantes e permanece inativo ate o saldo chegar a zero.
6. Antes do envio, o sistema apresenta uma revisao com cada empresa escolhida e sua quantidade de estrelas. O envio confirmado e unico, definitivo e anonimo.
7. A rodada e encerrada quando todas as empresas participantes concluirem ou automaticamente quando o prazo terminar. Empresas pendentes perdem o direito de distribuir estrelas naquela sessao.
8. Depois do envio, a empresa consulta seu resultado parcial da sessao. Rodadas encerradas passam a compor o historico.

Regras:

- Uma empresa nao pode avaliar a si mesma.
- A ordem das logos e sorteada novamente sempre que a tela e aberta.
- O painel usa uma grade compacta e responsiva para comportar um numero elevado de associadas sem aumentar desnecessariamente os cartoes.
- Nao existe limite de tres estrelas por empresa avaliada. Todo o saldo disponivel pode ser concentrado em uma unica associada, se essa for a decisao da avaliadora.
- A confirmacao na tela de revisao conclui e grava a distribuicao; nao existe uma segunda confirmacao posterior.
- O envio e idempotente: repetir tecnicamente a mesma confirmacao nao duplica estrelas nem cria uma segunda participacao.
- Somente o administrador da empresa envia a distribuicao. Os demais usuarios consultam resultados.
- Empresas aprovadas depois da abertura nao entram na rodada em andamento e passam a participar somente da proxima fotografia de associadas.
- O administrador da plataforma acompanha empresas concluidas e pendentes, mas nao ve o ranking da rodada aberta.
- O painel administrativo permite filtrar rodadas por nome, situacao e ano. Rodadas encerradas ficam organizadas em uma secao retratil; o ranking da sessao e a media geral tambem possuem secoes retrateis independentes.
- O administrador da plataforma pode excluir definitivamente uma rodada encerrada. A exclusao remove as distribuicoes e os resultados daquela sessao, elimina seus alertas vinculados e recalcula automaticamente a media historica. Rodadas abertas nao podem ser excluidas.
- O alerta de nova rodada leva diretamente para a tela Avaliacao de parcerias. Todos os usuarios da empresa recebem o aviso, mas somente o administrador da empresa envia a distribuicao.
- Nao existe encerramento administrativo manual. O backend verifica os prazos periodicamente e tambem a cada acesso ao modulo.
- O backend registra a empresa avaliadora apenas para impedir duplicidade e aplicar as regras. Essa identidade nao e devolvida pelas APIs de resultado.
- O resultado da sessao usa indice relativo: 100 representa a media de estrelas disponiveis naquela rodada.
- O resultado geral e a media aritmetica dos indices obtidos nas rodadas encerradas. Cada rodada possui o mesmo peso, mesmo quando a quantidade de associadas muda.
- O grafico apresenta indice real, linha de media do mercado e tendencia linear a partir de tres rodadas encerradas.
- A consulta historica pode ser filtrada por sessao, ano, ultimas cinco, ultimas dez ou todas as rodadas.

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

## Modulo Capacidade Tecnica

Uso principal: organizar profissionais e atestados tecnicos como insumos para avaliar a capacidade da empresa em futuras licitacoes.

### Profissionais tecnicos

- Cadastro de profissionais com foto, contatos, formacao e dados profissionais relevantes.
- Um profissional pode ter mais de uma formacao e mais de uma formacao complementar, como especializacao, mestrado ou doutorado.
- Os profissionais cadastrados ficam disponiveis para vinculacao aos atestados tecnicos e para analises futuras de equipe.

### Atestados tecnicos

- Cadastro de contratante, contratado, objeto, UF, inicio e fim da execucao, valor do contrato, numero da CAT, profissional, cargo e funcao.
- Registro da situacao do atestado: utilizavel pela empresa, pelo profissional ou por ambos.
- Um atestado pode ser vinculado a profissional de outra empresa, pois a CAT pode acompanhar a disponibilidade futura desse profissional.
- Cada atestado aceita varios quantitativos livres, com descricao, unidade e valor conforme a natureza do documento.
- O registro guarda o texto completo capturado do documento, permitindo consulta posterior e revisao manual.
- A data de inicio deve ser anterior a data final da execucao.

### Leitura de documentos e OCR

- PDFs com texto nativo passam por extracao direta.
- PDFs escaneados ou compostos por imagem podem ser lidos por OCR em portugues, com apoio do Tesseract instalado no computador do servidor.
- O sistema informa o tipo e a situacao da leitura: texto extraido, OCR extraido, pendente, manual ou falhou.
- O usuario pode abrir, corrigir o texto capturado e solicitar nova leitura do documento.

### Analise com IA

- A lista de atestados permite selecionar ate dez registros para analise conjunta.
- Antes do envio, a tela apresenta dois paineis retrateis de apoio interno:
  - quadro documental com dados de cada atestado e seus quantitativos preservados por descricao, valor e unidade;
  - cronograma mensal por profissional, com experiencia bruta, periodos sobrepostos e experiencia sem sobreposicao.
- Quantitativos de servicos ou unidades diferentes nao sao somados automaticamente.
- O calculo cronologico agrupa somente atestados do mesmo profissional, conta cada dia uma unica vez na experiencia liquida e sinaliza registros sem datas validas.
- Esses paineis sao calculados localmente com os dados ja carregados e nao alteram o roteiro, o JSON, os provedores nem o fluxo atual da IA.
- A tela de analise recebe um roteiro escrito pelo usuario e envia para a IA os dados estruturados e o texto capturado dos atestados selecionados.
- A resposta fica gravada no historico da analise, com situacao de fila, processamento, conclusao ou falha.
- O usuario pode escolher selecao automatica, somente OpenAI, somente Google Gemini ou somente Groq.
- No modo automatico, a ordem e OpenAI, Google Gemini e Groq. Se ocorrer falha tecnica, indisponibilidade, limite ou falta de credito, o backend tenta o provedor seguinte que estiver configurado.
- Uma resposta valida, mesmo desfavoravel ao conteudo analisado, nao aciona outro provedor. O fallback existe apenas para falha de processamento.
- O resultado identifica de forma visivel o provedor e o modelo que realmente responderam. O banco preserva esses dados no historico.
- Essa funcao depende de ao menos uma chave de API configurada somente no backend. Os documentos originais nao sao enviados nessa etapa; segue apenas o texto ja extraido e os dados cadastrais selecionados.
- A camada gratuita do Gemini deve ser usada somente com dados adequados as suas condicoes de privacidade. Antes da producao, o tratamento de documentos reais deve ser revisado conforme LGPD e termos do provedor.

## Modulo Academia LicitaHub

Uso principal: oferecer formacao propria para os usuarios da rede, sem depender de uma plataforma externa de cursos.

Funcionalidades:

- O catalogo mostra somente cursos publicados que o usuario ainda nao iniciou.
- A tela **Gerenciar cursos** lista e filtra os cursos; cada curso abre uma tela administrativa exclusiva.
- Na gestao do curso, o administrador define categoria, descricao, carga horaria, imagem de capa e situacao de publicacao.
- A mesma tela permite incluir e editar quantas aulas forem necessarias, escolhendo exclusivamente entre link do YouTube ou arquivo MP4/WebM enviado ao LicitaHub.
- Videos enviados ao LicitaHub aceitam ate 500 MB e ficam armazenados na area de uploads da Academia.
- O questionario de cada aula e configurado em uma area retratil, com questoes, alternativas, resposta correta e limite de tentativas.
- A gestao administrativa permite arquivar, reativar e excluir cursos sem matriculas; cursos com historico de alunos devem ser arquivados.
- O catalogo apresenta um resumo antes da matricula; depois de iniciado, o curso permanece disponivel em **Meus cursos**.
- Cada usuario retoma a videoaula do ponto salvo.
- A conclusao de todas as aulas emite um certificado individual em PDF, proprio para download e impressao, com aluno, empresa, curso, categoria, carga horaria, data e codigo unico.
- A primeira pagina usa a imagem de capa escolhida pelo administrador em uma area fixa; as paginas seguintes detalham cada aula, sua duracao e descricao, sem cortar cursos com muitos conteudos.
- O certificado informa o endereco publico de validacao; o conferente consulta o codigo sem precisar entrar no LicitaHub.
- A area **Meus cursos** concentra os cursos iniciados, permite pesquisar por titulo ou categoria, filtrar por andamento, retomar aulas e baixar certificados emitidos.

Regras principais:

- Apenas o administrador da plataforma cria cursos e aulas.
- Usuarios visualizam somente cursos publicados.
- O questionario permanece bloqueado ate que pelo menos 98% do tempo do video tenha sido efetivamente assistido.
- A proxima aula permanece bloqueada ate a conclusao da anterior.
- Aula sem questionario e concluida pelo video; aula com questionario exige pelo menos 75% de acerto.
- As respostas corretas ficam exclusivamente no backend e nao sao enviadas ao navegador do aluno.
- Depois da aprovacao, o questionario e recolhido e nao pode ser reaberto pelo aluno; a videoaula continua disponivel para revisao.
- O backend recusa o download de certificado quando o usuario ainda nao concluiu integralmente o curso.
- Em producao, `PUBLIC_BASE_URL` deve conter o endereco HTTPS oficial do LicitaHub para que a validacao impressa no PDF aponte ao ambiente correto.
- A origem do video fica registrada em cada aula. Cursos antigos permanecem como YouTube e os novos podem combinar aulas do YouTube e aulas com video proprio, mantendo as mesmas regras de progresso, prova e certificado.

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
8. A participacao individual permanece privada e cria automaticamente a Central de Montagem com os profissionais da propria empresa no momento em que o interesse e salvo. O prazo geral da montagem assume a data da sessao do edital.
9. Ao definir a empresa lider de um consorcio, a Central de Montagem consorcial tambem e criada automaticamente. O prazo geral assume a data da sessao do edital e todos os membros ativos podem acompanhar o plano.
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

### Central de montagens

Uso principal: recuperar e acompanhar qualquer montagem ativa sem depender do caminho pelo edital ou pelo consorcio.

- Disponivel no menu lateral em **Editais > Central de montagens**.
- Lista participacoes individuais e consorciais nas quais a empresa possui acesso.
- Permite buscar por edital, orgao, objeto ou empresa lider e filtrar por tipo e andamento.
- Cada card mostra a sessao, empresas participantes, tarefas abertas e percentual concluido.
- Cada item abre diretamente a Central de Montagem da licitacao correspondente.

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
- Empresa que ja integra consorcio ativo no edital nao visualiza anuncios de terceiros nem recebe comandos para avaliar, conversar ou abrir nova busca de parceiros nesse mesmo edital.
- A excecao e a empresa lider, depois de abrir em **Meus consorcios** um anuncio complementar para buscar nova consorciada. Nesse caso, ela visualiza e avalia somente as candidaturas dessa busca.

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
- Empresa ja integrante de consorcio ativo fica fora da vitrine daquele edital, salvo a lider quando existir anuncio complementar publicado pelo proprio consorcio.

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
- A protecao de consorcio e aplicada tambem no backend: nao basta esconder botoes. Membro ativo nao pode abrir por URL direta detalhe, chat ou avaliacao de anuncio externo no mesmo edital, exceto a lider em busca complementar.

## Pausa por status do edital

Anuncios de parceria e montagens existem operacionalmente apenas enquanto o edital esta com status `Publicado`.

1. Ao mudar para qualquer status nao publicado, o sistema pausa anuncios que estavam publicados e a Central de Montagem, preservando tarefas, documentos e historico.
2. Enquanto estiver pausado, o edital nao aparece na vitrine nem nas centrais ou calendario de montagem e nao gera alertas de prazo.
3. Ao retornar para `Publicado`, anuncios e montagens pausados pelo sistema sao restaurados no estado anterior; anuncios fechados manualmente nao sao reabertos.

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
- Empresa participante de consorcio ativo ve, no detalhe do edital, o atalho para **Meus consorcios** em vez dos comandos de registrar interesse, editar estrategia ou desistir isoladamente.

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
- Alertas pessoais sao visiveis apenas para o usuario destinatario. Um alerta geral da empresa so aparece para todos quando for gravado sem destinatario individual.
- Ao abrir o sino, o contador zera.
- Alertas lidos nao voltam como novos para aquele usuario.
- O botao **Historico** no sino abre o arquivo de alertas, com busca por texto, filtros por tipo, situacao de leitura e periodo.
- O historico permanece gravado no banco e cada registro leva de volta para a origem quando houver uma tela relacionada.

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
- Capacidade tecnica: profissionais tecnicos, formacoes, formacoes complementares, atestados, quantitativos, documentos, texto extraido, resultados de OCR e historico de analises por IA.
- Academia: cursos, aulas, fontes de video, questoes, alternativas, tentativas, progresso, matriculas e certificados verificaveis.
- Avaliacao de parcerias: rodadas, fotografia de participantes, distribuicoes anonimas, conclusoes por empresa e resultados historicos.
- Notificacoes: alertas por usuario/empresa.
- Recuperacao de senha: o administrador da plataforma consulta pedidos, filtra por usuario, empresa e situacao e copia links ainda validos enquanto o envio por e-mail nao estiver configurado. O link mostra se esta disponivel, usado ou expirado.
- Auditoria: estrutura registra acoes sensiveis e eventos de acesso; a cobertura completa de todos os fluxos fica para uma etapa posterior.

## Itens deixados para fase posterior

- Dashboards analiticos completos.
- Ampliacao da cobertura da auditoria para todos os fluxos administrativos.
- Backend de producao com autenticacao mais robusta.
- Hash de senha adequado para producao.
- Rotacao, centralizacao e retencao dos logs em ambiente de producao.
- Configuracao por ambiente.
- Deploy externo.
- Evolucao para aplicativo mobile.
- Integracao futura com outras IAs e automacoes avancadas.

## Arquitetura e operacao local

- Frontend: React compilado pelo Vite e servido a partir de `dist/`.
- Backend: API HTTP em Go, que tambem entrega o frontend compilado.
- Banco: PostgreSQL, com estrutura versionada em `database/schema.sql`.
- Arquivos: armazenados localmente em `uploads/` no ambiente de desenvolvimento.
- OCR: Tesseract instalado na maquina do backend, usado quando o PDF nao possui texto pesquisavel.
- IA: integracao opcional pelo backend, habilitada apenas quando a chave e o modelo estiverem configurados em variaveis de ambiente.
- Inicializacao local: `backend/run-dev.cmd`, mantendo a janela aberta durante o uso.
- Endereco padrao: `http://127.0.0.1:8080`.

O ambiente atual e adequado para desenvolvimento e validacao local. Antes de uso publico, devem ser concluidos os itens de producao descritos em `docs/SEGURANCA.md`, especialmente senha com hash forte, HTTPS, cookies seguros, consultas parametrizadas, armazenamento persistente de arquivos, backup e monitoramento.

## Pre-analise de Editais com IA

O administrador da plataforma pode anexar quantos documentos forem necessarios ao edital e, no detalhe do edital, solicitar a geracao da pre-analise tecnica em HTML.

- O roteiro oficial esta em `backend/prompts/analise-primaria-edital.txt` e foi fornecido pela LicitaHub.
- O HTML gerado aparece no proprio detalhe do edital, pode ser baixado e continua podendo ser substituido por um HTML manual.
- Cada tentativa fica registrada com situacao: aguardando, em processamento, concluida ou falhou.
- O processamento usa somente os documentos diretamente anexados ao edital. Arquivos compactados e arquivos tecnicos sem conteudo textual nao entram na analise automatica.
- A analise deve receber arquivos de ate 22 MB cada e 50 MB somados em uma unica solicitacao. Para maior fidelidade de tabelas, imagens e diagramas, prefira documentos em PDF.
- A OpenAI e consultada primeiro. Se houver falta de credito, limite, indisponibilidade ou falha tecnica, o backend tenta automaticamente o Google Gemini.
- O Gemini recebe os mesmos documentos e o mesmo roteiro. Os arquivos temporarios enviados ao servico sao removidos ao final da tentativa.
- O provedor e o modelo que efetivamente geraram o HTML ficam registrados no banco e visiveis no detalhe do edital.
- As chaves ficam apenas nos arquivos locais `backend/.env.openai`, `backend/.env.gemini` e `backend/.env.groq`, ignorados pelo Git. Elas nunca devem ser adicionadas ao frontend ou enviadas ao repositorio.
- Para configurar no computador local, abra o arquivo `CONFIGURAR` correspondente ao provedor; depois reinicie `backend/run-dev.cmd`.
- A Groq e utilizada somente nos fluxos baseados em texto e JSON. A pre-analise dos documentos de editais continua usando OpenAI e Gemini.

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

## Captação PNCP

O administrador da plataforma possui, em `Editais > Captação PNCP`, uma fila separada para oportunidades originadas no Portal Nacional de Contratações Públicas.

1. Define o período de publicação e, opcionalmente, o estado.
2. Consulta o PNCP e registra os resultados apenas como capturas internas.
3. O LicitaHub atribui uma aderência preliminar a partir de termos ligados à engenharia consultiva, como projetos, supervisão, meio ambiente, arqueologia, saneamento e infraestrutura.
4. O administrador pode consultar o registro original no PNCP, preparar seu cadastro ou descartar.
5. Ao preparar, o LicitaHub cria um rascunho já preenchido e abre o cadastro de edital para revisão, complementação de dados e inclusão de documentos.
6. O edital só fica visível para as empresas associadas e gera a notificação de novo edital quando o administrador conclui o cadastro e o salva com status `Publicado`.

O filtro de aderência é apenas apoio de triagem. A decisão de publicar é sempre humana e manual.

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

## Situacao dos testes

Uma rodada manual dos fluxos centrais foi encerrada em 11/07/2026. Ela validou:

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

Desde essa rodada foram incluidos ou ampliados Academia, Capacidade Tecnica, captacao oficial, avaliacoes de parcerias, seguranca, notificacoes e outros ajustes. Esses recursos possuem cenarios no roteiro atualizado, mas a documentacao nao presume que uma nova rodada integral ja tenha sido executada. Antes de qualquer piloto externo, todo o arquivo `docs/ROTEIRO-DE-TESTES.md` deve ser refeito e as evidencias devem ser registradas.

## Atualizacao: captacao por fontes oficiais

Em `Editais > Captacao de editais`, o administrador pode consultar duas fontes: `PNCP` e `Compras.gov.br`. As oportunidades entram primeiro em uma fila de revisao e nunca sao publicadas automaticamente.

Na fila, e possivel filtrar por texto, situacao, aderencia e valor estimado minimo ou maximo. Cada registro pode exibir fonte, orgao, numero, objeto, modalidade, criterio de julgamento, local, sessao e valor estimado, conforme os dados retornados pela fonte.

O administrador pode abrir a fonte oficial, preparar o cadastro ou descartar a oportunidade. Descartar remove a captura pendente da fila e do banco, reduzindo o contador e evitando acumulo de registros sem utilidade. Ao preparar, a origem e o criterio de julgamento sao levados para o rascunho do edital.

### Paginacao das fontes

- A consulta e controlada pelo administrador e processa somente a pagina indicada em cada comando.
- No PNCP, o LicitaHub solicita ate 50 registros por pagina para cada modalidade selecionada.
- Com as quatro modalidades padrao selecionadas, um comando pode receber ate 200 registros brutos: 50 de cada modalidade.
- O total efetivamente salvo pode ser menor por causa da ultima pagina, registros incompletos ou repetidos.
- Depois de receber uma pagina, o sistema prepara o comando para a pagina seguinte. Nao existe varredura automatica ilimitada.
- Alterar periodo, estado, fonte ou modalidades reinicia a consulta na pagina 1.

### Triagem da fila com IA

O administrador pode solicitar a classificacao das oportunidades pendentes por inteligencia artificial, analisando toda a fila ou somente registros marcados.

- O backend divide a selecao em lotes de 10 oportunidades e acompanha o progresso em segundo plano, reduzindo risco de resposta truncada.
- A ordem de tentativa e OpenAI, Google Gemini e Groq. Se houver falha tecnica, falta de credito, limite ou resposta JSON inconsistente, o mesmo lote e encaminhado ao proximo provedor configurado.
- O Gemini recebe um esquema de resposta `application/json` obrigatorio; se a primeira resposta ainda for inconsistente, o backend realiza uma segunda tentativa estruturada antes de encerrar o lote com falha.
- A Groq recebe um esquema JSON estrito e atua como terceira alternativa, preservando a mesma validacao de identificadores e categorias antes da gravacao.
- Cada oportunidade recebe uma classificacao: `consultiva`, `relacionada`, `nao_consultiva` ou `duvidosa`.
- O resultado registra confianca de 0 a 100, justificativa, areas tecnicas identificadas, provedor, modelo, versao do roteiro e data.
- A pontuacao por palavras do LicitaHub permanece separada da classificacao feita pela IA.
- A fila permite filtrar somente engenharia consultiva, relacionadas, duvidosas, nao consultivas ou ainda nao analisadas.
- A IA nunca publica nem descarta uma oportunidade. A decisao final continua exclusiva do administrador.
- Os textos vindos das fontes oficiais sao tratados como dados nao confiaveis; eventuais instrucoes contidas no objeto devem ser ignoradas pelos provedores.
- Novas analises criam historico e nao apagam resultados anteriores.

A API publica do Compras.gov.br fornece dados estruturados disponibilizados, mas esta integracao nao le nem envia mensagens privadas, diligencias ou chat da sala de disputa em tempo real. Para essas operacoes, o usuario deve acessar o processo oficial.

## Atualizacao: WhatsApp planejado

A integracao oficial de alertas por WhatsApp ainda nao esta ativa. O desenho previsto usa a WhatsApp Business Platform/Cloud API como canal opcional, mediante telefone confirmado e consentimento do usuario. Tokens e credenciais deverao ficar somente em variaveis protegidas do backend, nunca no frontend ou no GitHub.

## Atualizacao: linha do tempo do edital

O detalhe do edital apresenta uma linha do tempo real com cadastro ou captacao, publicacao, retomada, suspensao, sessao ocorrida, encerramento, cancelamento, analise e impugnacao. Os eventos sao armazenados em `tender_timeline_events`, e as alteracoes de status sao registradas automaticamente pelo banco. Editais antigos recebem um evento inicial e o status atual na primeira ativacao da funcionalidade.

Movimentacoes de interesse, match, consorcio e montagem sao filtradas pela empresa logada. A linha do tempo fica recolhida por padrao para preservar espaco na tela.

## Atualizacao: saneamento entre PNCP e Compras.gov.br

As captacoes do PNCP e do Compras.gov.br sao unificadas na fila pelo numero de controle PNCP (`numeroControlePNCP`). Esse e o identificador principal e evita que a mesma contratacao apareca duas vezes.

O saneamento independe da ordem de consulta. Se o Compras.gov.br for consultado primeiro e o PNCP depois, ou vice-versa, o segundo registro encontra o primeiro pelo mesmo numero de controle e atualiza o item existente.

Quando uma das fontes nao fornecer o numero de controle PNCP, o sistema nao unifica os registros entre fontes por semelhanca de objeto, orgao, numero ou valor. Eles permanecem separados para revisao humana, evitando unir indevidamente licitacoes parecidas de trechos ou escopos diferentes.

No saneamento, os dados do PNCP sao a referencia principal. O Compras.gov.br complementa apenas campos ausentes, como modalidade, criterio de julgamento, valor, sessao, local ou outros dados estruturados. O item passa a indicar `PNCP + Compras.gov.br / saneado`, e o payload bruto das duas fontes permanece armazenado para rastreabilidade.

## Atualizacao: outros portais de licitacao

O LicitaHub nao presume que todos os portais oferecam uma API publica. As fontes sao tratadas em tres grupos:

1. APIs publicas ou dados abertos, que podem ser integrados diretamente apos validacao tecnica e juridica.
2. APIs privadas para compradores, clientes ou parceiros, que exigem credenciais, contrato ou autorizacao do portal.
3. Portais apenas com consulta web ou area autenticada, que nao devem ser automatizados sem permissao e verificacao dos termos de uso.

PNCP e Compras.gov.br permanecem como fontes automaticas ativas. Licitacoes-e, Portal de Compras Publicas, BNC, BLL, BEC/SP, Licitacoes CAIXA e Petronect sao fontes potenciais para estudo futuro. Uma integracao so deve entrar no produto quando houver meio oficial, estavel e autorizado de consulta. Captura por raspagem de pagina nao faz parte da arquitetura planejada.
