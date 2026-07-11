# Documentacao funcional do LicitaHub

## Objetivo

O LicitaHub e uma aplicacao web para empresas de engenharia consultiva. A plataforma combina rede social empresarial, divulgacao institucional, noticias da plataforma, editais publicos e formacao de parcerias/consorcios.

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

## Modulo Empresa

Uso principal: manter a presenca institucional da empresa.

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
5. Empresa registra interesse.
6. Empresa informa o que tem, o que atende parcialmente, o que nao se aplica e o que busca em parceiros.
7. Interesse gera anuncio para empresas interessadas e vitrine.

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

Regras principais:

- O proprio anuncio da empresa logada aparece marcado como "Meu anuncio".
- O botao de avaliacao nao aparece no proprio anuncio.

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

## Avaliacao, match e consorcio

Fluxo:

1. Empresa A avalia anuncio da empresa B.
2. Empresa A aprova, recusa ou deixa para depois.
3. Empresa B avalia anuncio da empresa A.
4. Quando ha aprovacao reciproca, o sistema gera match.
5. Match aparece em meus consorcios.
6. Sistema disponibiliza contato/WhatsApp.
7. Empresas podem definir lider do consorcio.
8. Ao fechar match/consorcio, anuncios das empresas consorciadas devem sair da vitrine publica daquele edital.

Regras principais:

- Uma empresa nao pode avaliar o proprio anuncio.
- Uma empresa nao pode gerar consorcio consigo mesma.

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
- Match/consorcio: anuncios, avaliacoes, matches, contatos, intencoes e membros.
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
- Integracoes com IA/API.

## Roteiro de teste manual

O teste manual deve passar por:

1. Login com todos os perfis.
2. Convite e aprovacao de empresa.
3. Edicao de perfil da empresa.
4. Criacao e gestao de usuarios.
5. Criacao e interacao na comunidade.
6. Perfil publico da empresa.
7. Cadastro e gerenciamento de noticias.
8. Cadastro e detalhe de editais.
9. Registro de interesse.
10. Empresas interessadas por edital.
11. Vitrine geral.
12. Detalhe de anuncio.
13. Avaliacao reciproca.
14. Match e definicao de lider.
15. Notificacoes no sino.
16. Tentativas de acesso indevido por usuario comum.

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
