# LicitaHub

Plataforma web para rede empresarial de engenharia consultiva, com comunidade, Radar de noticias, central de editais, vitrine de parceiros e fluxo de match/consorcio.

## Estado atual

O projeto deixou de ser apenas prototipo visual. Ele ja possui frontend React compilado, backend local em Go e banco PostgreSQL.

Rodada de testes manuais encerrada em 11/07/2026, com ajustes finais aplicados nos modulos de acesso, empresa, Radar, editais, interesse, vitrine, match e consorcios.

Dashboards analiticos e auditorias completas ficam para uma fase posterior.

## Stack

- Frontend: React compilado com Vite, HTML e CSS
- Backend: Go
- Banco de dados: PostgreSQL
- Execucao local: backend Go servindo API e frontend compilado

## Estrutura

- `src/`: codigo-fonte do frontend React.
- `dist/`: frontend compilado pelo Vite.
- `backend/`: API Go e executaveis locais.
- `database/`: schema PostgreSQL e base das migracoes.
- `docs/`: documentacao funcional do sistema.
- `preview-react-cdn.html`: arquivo legado que redireciona para a versao Vite.

## Como rodar localmente

1. Instale Node.js, Go e PostgreSQL.
2. Rode `npm.cmd install` quando precisar instalar dependencias.
3. Rode `npm.cmd run build` para gerar a pasta `dist/`.
4. Abra `backend/run-dev.cmd`.
5. Acesse `http://127.0.0.1:8080/`.

O `run-dev.cmd` liga o backend e, se a pasta `dist/` nao existir, tenta gerar a versao Vite automaticamente.

## Acesso local

- Usuario administrador local: `admin@licitahub.local`
- A senha nao fica documentada no repositorio. Use a senha definida no seu banco local ou gere uma nova pelo fluxo de recuperacao/configuracao de senha.

Nao publique senhas, tokens ou credenciais no GitHub.

## Modulos implementados

- Acesso e administracao.
- Empresa e usuarios vinculados.
- Meu perfil.
- Comunidade.
- Perfil publico da empresa.
- Radar LicitaHub.
- Editais.
- Interesse em edital.
- Vitrine de parceiros.
- Empresas interessadas.
- Avaliacao de candidata.
- Match/consorcio.
- Chat de parceria em tempo real.
- Notificacoes no sino.

## Ajustes recentes validados

- Convite com mascara/validacao de CNPJ, email, telefone e contato com nome completo.
- Estados brasileiros em selects e apoio para cidades principais por UF.
- Aceite de convite redirecionando para login apos cadastro concluido.
- Menu lateral sem telas dependentes de contexto, como detalhe de noticia e analise de empresa.
- Cadastro/edicao de edital com modalidade, valor monetario, estado, cidade e status impugnado.
- Editais com data de abertura passada tratados como ocorridos.
- Lista de editais indica quando a empresa ja registrou interesse e remove o botao de registrar novamente.
- Empresas interessadas e vitrine mostram o proprio anuncio como "Meu anuncio", sem permitir match consigo mesmo.
- Vitrine separa meus anuncios dos anuncios de outras empresas, com edicao de resumo e encerramento seguro do anuncio proprio.
- Consorcios permitem inclusao de terceira empresa pela lider, com candidatura, aceite e registro dos membros.
- Desistencia de consorcio registra a retirada e exige sucessora quando a lider deixa uma composicao que permanece ativa.
- Perfil publico permite abrir detalhes profissionais de usuarios ativos sem expor dados internos.
- Registrar interesse usa linguagem de requisitos e pontuacao.
- Radar com filtro fixo, paginacao e limite de titulo/resumo.
- Usuarios vinculados e admin de editais com filtros fixos.

## Documentacao funcional

Consulte [docs/SISTEMA.md](docs/SISTEMA.md) para ver os fluxos, regras de acesso, telas principais e itens deixados para fase posterior.

## Cuidados antes de testar

- Abrir sempre `http://127.0.0.1:8080/`, nao o arquivo HTML antigo.
- Manter a janela do backend aberta.
- Se alterar o frontend, rodar `npm.cmd run build` antes de testar pela URL do backend.
- Testar com perfis diferentes: administrador da plataforma, administrador da empresa, comercial, tecnico e leitor.
