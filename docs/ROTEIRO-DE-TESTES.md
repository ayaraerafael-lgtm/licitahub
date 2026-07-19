# Roteiro Completo de Testes - LicitaHub

Este roteiro serve para validar o sistema como se ele estivesse em operacao real. Cada teste deve registrar: data, usuario utilizado, resultado, evidencia (imagem quando necessario) e observacao.

Documentacao revisada em 18/07/2026.

## 1. Regra de ouro antes dos testes

Nao usar exclusao como forma normal de interromper um edital que ja foi publicado. Para manter historico e evitar perda de rastreabilidade, a regra recomendada e:

| Situacao | Acao recomendada | Resultado esperado |
|---|---|---|
| Edital em rascunho, sem empresa interessada | Excluir | Registro deixa de aparecer; historico administrativo permanece. |
| Edital publicado com erro temporario | Suspender | Continua consultavel com aviso; bloqueia novas manifestacoes, anuncios, avaliacoes e matches. Dados existentes permanecem preservados. |
| Edital definitivamente cancelado | Cancelar | Fecha anuncios, bloqueia conversas daquele edital, encerra montagens e deixa tarefas bloqueadas, sem apagar historico. |
| Sessao realizada | Marcar como ocorrido | Bloqueia novas manifestacoes e novos matches; preserva consulta, consorcios e montagem em modo de consulta. |
| Edital com impugnacao | Marcar como impugnado | Mantem consulta e fluxo da impugnacao; a decisao sobre suspender ou manter participacoes deve ser tomada pelo administrador. |

**Regra a implementar/validar:** nenhum edital publicado com interesses, anuncios, match, consorcio ou montagem deve ser apagado fisicamente. Nesses casos, a acao correta e `suspender` ou `cancelar`.

## 2. Massa de teste

Preparar antes de iniciar:

- Um administrador da plataforma.
- Tres empresas ativas: Empresa Alfa, Empresa Beta e Empresa Gama.
- Em cada empresa: um administrador, um comercial, um tecnico e um leitor.
- Um edital futuro para participacao individual.
- Um edital futuro para consorcio entre duas empresas.
- Um edital futuro para consorcio entre tres empresas.
- Um edital suspenso, um cancelado, um ocorrido e um em analise.
- Noticias publicadas, uma noticia expirada e uma noticia em rascunho.

## 3. Acesso, sessoes e perfis

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| AC-01 | Login valido | Entrar com administrador da plataforma | Painel administrativo abre. |
| AC-02 | Login invalido | Informar senha incorreta | Acesso negado sem revelar detalhes internos. |
| AC-03 | Empresa bloqueada | Bloquear empresa e tentar login de seus usuarios | Nenhum usuario da empresa consegue entrar. |
| AC-04 | Usuario bloqueado | Bloquear somente um usuario e tentar login | Apenas esse usuario perde acesso. |
| AC-05 | Usuario removido | Remover vinculo e tentar login | Usuario nao consegue entrar e nao aparece na lista ativa. |
| AC-06 | Sessao encerrada | Sair da conta e atualizar pagina | Sistema retorna ao login. |
| AC-07 | URL administrativa | Logar como tecnico e abrir URL de administracao | Tela de acesso negado; nenhum dado administrativo e exposto. |
| AC-08 | Perfil de leitor | Navegar pelo sistema como leitor | Consulta permitida; botoes de criar, editar, bloquear, excluir ou dar match nao aparecem. |
| AC-09 | Recuperacao de senha | Gerar e usar link de recuperacao | Link funciona uma vez; nova senha permite login. |

## 4. Convites, empresas e usuarios

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| EM-01 | Novo convite | Criar convite com CNPJ, telefone, email e contato valido | Sistema confirma gravacao e gera link de aceite. |
| EM-02 | Validacoes | Tentar CNPJ, telefone, email ou nome de contato invalidos | Formulario impede envio e explica o campo. |
| EM-03 | Aceite | Abrir link, completar empresa, usuario, foto e senha | Cadastro fica aguardando analise; tela segue para login somente apos sucesso. |
| EM-04 | Aprovacao | Administrador aprova empresa aceita | Empresa e usuario passam a ter acesso. |
| EM-05 | Recusa | Administrador recusa empresa | Empresa nao recebe acesso. |
| EM-06 | Perfil da empresa | Administrador da empresa altera logo, descricao, porte, site e area de atuacao | Perfil publico mostra dados atualizados. |
| EM-07 | Usuario vinculado | Administrador cria usuario com foto, email, telefone, cargo, perfil e senha | Novo usuario consegue login; foto aparece no topo e no perfil publico. |
| EM-08 | Edicao propria | Usuario altera propria foto, nome, email e telefone | Alteracoes aparecem na proxima sessao. Cargo e perfil nao podem ser alterados pelo proprio usuario. |
| EM-09 | Gestao de usuarios | Administrador troca perfil, bloqueia, desbloqueia e remove usuario | Cada operacao respeita permissoes e gera efeito imediato. |

## 5. Radar LicitaHub e noticias

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| RA-01 | Cadastro | Administrador cria noticia com titulo, resumo, imagem, status e periodo | Noticia grava e aparece conforme status e periodo. |
| RA-02 | Limites | Tentar ultrapassar limite de titulo ou resumo | Formulario bloqueia excesso. |
| RA-03 | Imagem | Enviar imagem horizontal, vertical e quadrada | Card e detalhe preservam imagem sem distorcer ou aumentar o layout. |
| RA-04 | Principal | Marcar noticia como principal | Ela aparece como destaque no Radar. |
| RA-05 | Expiracao | Definir data anterior a hoje | Noticia nao aparece como publicada. |
| RA-06 | Paginacao | Criar mais de 12 noticias secundarias | Lista pagina corretamente; noticias antigas ficam depois das recentes. |
| RA-07 | Detalhe | Clicar em ler noticia | Abre a noticia correta com imagem e texto. |

## 6. Comunidade e perfil publico

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| CO-01 | Publicacao | Criar publicacao de cada categoria com texto e imagem | Post aparece no feed e no perfil conforme visibilidade. |
| CO-02 | Curtida e favorito | Curtir e favoritar; desfazer ambos | Contadores e cores atualizam sem duplicar registro. |
| CO-03 | Comentario | Abrir comentarios, comentar e recolher | Campo abre, comentario grava e recolhimento funciona. |
| CO-04 | Gestao propria | Editar, arquivar e excluir publicacao propria | Alteracao afeta somente a publicacao escolhida. |
| CO-05 | Filtros | Filtrar por categoria, regiao e empresa | Feed mostra somente resultados coerentes; filtro permanece visivel ao rolar. |
| CO-06 | Perfil externo | Abrir perfil publico de outra empresa pelo feed | Dados institucionais, profissionais e publicacoes da empresa correta aparecem. |
| CO-07 | Profissionais | Abrir profissionais vinculados e detalhe de profissional | Dados permitidos aparecem; icone de chat abre conversa direta com o profissional. |

## 7. Ciclo de vida do edital

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| ED-01 | Cadastro | Administrador cria edital futuro com todos os campos, documentos e analise HTML | Edital aparece na lista e no detalhe. |
| ED-02 | Sem analise | Criar edital sem HTML | Detalhe exibe aviso de edital ainda nao analisado; administrador pode anexar depois. |
| ED-03 | Arquivos | Anexar varios documentos e baixar como empresa | Todos os documentos aparecem e baixam corretamente. |
| ED-04 | Edicao | Alterar modalidade, valor, UF, cidade, status e data | Lista e detalhe refletem dados novos. |
| ED-05 | Data invalida | Tentar cadastrar ou editar sessao para data passada | Sistema impede a gravacao. |
| ED-06 | Ocorrido | Deixar data passar e atualizar sistema | Edital muda para ocorrido; nao aceita novo interesse. |
| ED-07 | Suspensao | Suspender edital sem interesse | Sai de oportunidades ativas; nao aceita novo interesse. |
| ED-08 | Suspensao com interesse | Suspender edital com anuncios, conversa e montagem | Anuncios saem da vitrine; a montagem deixa de aparecer e nao aceita alteracoes ou alertas de prazo; historico e tarefas ficam preservados. |
| ED-08A | Retomada | Alterar edital suspenso para publicado | Anuncios e montagem pausados pelo sistema sao restaurados no estado anterior; empresas relacionadas recebem alerta; nenhuma notificacao duplicada e criada por edicao sem troca de status. |
| ED-09 | Cancelamento com consorcio | Cancelar edital com consorcio e montagem | Anuncios devem sair da vitrine; chat do edital deve encerrar; consorcio e montagem devem ficar encerrados; tarefas devem ficar bloqueadas; historico deve permanecer consultavel. |
| ED-10 | Exclusao indevida | Tentar excluir edital publicado com relacionamentos | Sistema deve impedir exclusao e orientar a usar cancelamento. |
| ED-11 | Exclusao de rascunho | Excluir edital sem interesse ou montagem | Edital desaparece das listas sem deixar registros operacionais ativos. |
| ED-12 | Impugnacao | Protocolar com e sem prazo legal, anexar documentos e alterar status | Alerta de intempestividade aparece quando aplicavel; administrador acompanha na central de impugnacoes. |

## 8. Interesse, anuncios e vitrine

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| IN-01 | Participacao individual | Registrar interesse como individual | Nenhum anuncio publico e criado; botao abre montagem individual. |
| IN-02 | Busca de parceiros | Registrar interesse buscando parceiros | Anuncio aparece em Empresas interessadas e Vitrine. |
| IN-03 | Acompanhamento | Registrar como ainda avaliando | Registro fica privado; nao cria anuncio. |
| IN-04 | Requisitos | Preencher todas as opcoes de requisito | Detalhe do anuncio mostra requisito, situacao, o que possui e o que busca sem misturar dados. |
| IN-05 | Alterar estrategia | Mudar individual para busca de parceiros | Anuncio e criado; montagem individual e tarefas sao bloqueadas e encerradas, preservando historico. |
| IN-06 | Desistir | Desistir antes de consorcio fechado | Anuncio some e interesse deixa de ser ativo. |
| IN-07 | Repeticao | Tentar registrar interesse duas vezes | Sistema atualiza o mesmo registro, sem criar duplicidade. |
| IN-08 | Proprio anuncio | Abrir Empresas interessadas como anunciante | Empresa nao pode avaliar a si mesma. |
| IN-09 | Sem interesse | Tentar avaliar anuncio sem ter interesse no mesmo edital | Sistema direciona para registro de interesse, sem permitir avaliacao invalida. |

## 9. Match, consorcio e lideranca

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| MA-01 | Match entre duas | Alfa e Beta se avaliam reciprocamente | Match e consorcio aparecem para ambas. |
| MA-02 | Anuncios apos match | Consultar vitrine apos fechamento | Anuncios das participantes deixam de aparecer para terceiros. |
| MA-03 | Terceira empresa | Lider publica busca complementar e Gama se candidata | Gama entra no consorcio somente apos aceite da lider. |
| MA-04 | Lideranca inicial | Administrador/comercial de empresa ativa escolhe lider | Lider aparece em Meus consorcios e nas telas relacionadas. |
| MA-05 | Troca de lider | Cada empresa ativa, inclusive a terceira, troca a lideranca | Troca funciona independentemente da ordem de entrada; todos recebem notificacao. |
| MA-05A | Montagem automatica | Definir lider de consorcio e sair da tela | A Central de Montagem e criada automaticamente com prazo geral igual a data da sessao. |
| MA-06 | Perfil sem permissao | Tecnico ou leitor tenta trocar lider | Campo e acao nao ficam disponiveis. |
| MA-07 | Desistencia de membro | Empresa nao lider desiste | Ela deixa o consorcio; demais membros permanecem consistentes. |
| MA-08 | Desistencia da lider | Lider tenta sair com duas ou mais empresas restantes | Sistema exige nova lideranca antes da saida. |
| MA-09 | Menos de duas | Uma empresa sai deixando somente uma | Consorcio e anuncios associados sao encerrados de forma coerente. |

## 10. Central de Montagem e tarefas

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| MO-01 | Inicio consorcial | Lider inicia montagem de consorcio | Oito fases e tarefas padrao sao criadas uma unica vez. |
| MO-02 | Inicio individual | Empresa inicia montagem individual | Apenas profissionais da propria empresa podem receber tarefas. |
| MO-03 | Atribuicao | Lider atribui tarefa a empresa e profissional valido | Responsavel recebe alerta e ve tarefa em Minhas tarefas. |
| MO-04 | Acesso de colaboradora | Profissional da consorciada atualiza tarefa atribuida | Pode alterar andamento, comentario e evidencia; nao altera estrutura. |
| MO-05 | Estrutura | Lider cria e exclui fase/tarefa | Somente lider pode alterar estrutura; exclusao remove dependencias previstas. |
| MO-06 | Prazos | Definir prazo antes de hoje ou depois da sessao | Sistema impede. |
| MO-07 | Status | Testar pendente, andamento, aguardando informacao, bloqueada, revisao, devolvida, concluida e nao se aplica | Cor, percentual e permissoes correspondem ao status. |
| MO-08 | Percentual | Concluir e marcar nao se aplica em tarefas | Percentuais da fase, montagem e calendario recalculam corretamente. |
| MO-09 | Evidencias | Incluir arquivo, link e anotacao | Itens aparecem na tarefa e no dossie. |
| MO-10 | Encerramento | Cancelar edital ou sair da participacao individual | Montagem deixa de aceitar alteracoes; tarefas ficam bloqueadas; historico permanece. |
| MO-10A | Pausa por edital | Suspender edital com montagem ativa e tentar abrir link direto da central | A montagem permanece preservada, mas fica indisponivel enquanto o edital nao estiver publicado; nenhum aviso de prazo e emitido. |
| MO-11 | Central de montagens | Salvar participacao individual, sair da tela e abrir Editais > Central de montagens | A montagem individual e criada no salvamento, aparece na lista e abre diretamente o mesmo plano. |

## 11. Calendario de montagens

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| CA-01 | Exibicao | Abrir mes com montagem individual e consorcial | Cards aparecem no dia da sessao correta. |
| CA-02 | Navegacao | Ir para mes anterior, proximo e Hoje | Grade e dados correspondem ao mes selecionado. |
| CA-03 | Card | Clicar em cada card | Abre a Central de Montagem correta. |
| CA-04 | Percentual | Comparar card com central | Percentual e cor da barra sao coerentes: vermelho ate 30%, amarelo 31-70%, verde acima de 70%. |
| CA-05 | Filtros | Alternar todas, individual e consorcio | Nenhuma montagem indevida aparece. |
| CA-06 | Cancelada | Cancelar edital ou montagem | Card deixa de aparecer como montagem ativa. |

## 12. Chat e notificacoes

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| CH-01 | Chat de anuncio | Abrir chat por anuncio e enviar entre duas empresas | Mensagens chegam em tempo real e ficam no historico. |
| CH-02 | Multiplos chats | Abrir duas conversas e minimizar uma | Conversas coexistem sem perda de mensagens. |
| CH-03 | Chat de tarefa | Responsavel e lider conversam sobre uma tarefa | Apenas participantes autorizados acessam o historico. |
| CH-04 | Chat encerrado | Fechar consorcio ou cancelar edital | Chat relacionado nao permite nova mensagem. |
| CH-05 | Sino | Gerar curtida, comentario, convite, match, tarefa e chat | Sino fica amarelo com pendencias, cinza sem novidades e direciona para a origem correta. |
| CH-06 | Som | Receber mensagem nova com tela aberta | Som toca em volume perceptivel sem repetir indevidamente. |
| CH-07 | Histórico | Abrir Histórico no sino e filtrar alertas lidos, não lidos, tipo e período | Registros antigos aparecem paginados e cada item abre sua origem correta. |
| CH-08 | Isolamento por usuário | Gerar alerta para uma pessoa da Empresa Alfa e abrir o sino de outra pessoa da mesma empresa | A segunda pessoa não vê o alerta pessoal da primeira; apenas avisos gerais da empresa aparecem para ambas. |

## 13. Integracoes, seguranca e modulos complementares

## 13A. Captação PNCP

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| PN-01 | Consulta | Como administrador da plataforma, abrir Editais > Captação PNCP e pesquisar período válido | A fila recebe as oportunidades retornadas pela fonte oficial, sem publicar automaticamente. |
| PN-02 | Triagem | Alternar aderência e situação da fila | Editais com termos técnicos ficam priorizados; administrador ainda pode exibir todas as capturas. |
| PN-03 | Preparar cadastro | Preparar uma captura ainda pendente | Um único edital em rascunho é criado, com os dados do PNCP preenchidos, e o sistema abre seu cadastro para complementação. |
| PN-03A | Publicação final | Completar o rascunho, anexar documentos se necessário e salvar com status Publicado | O edital fica visível na lista das associadas e gera alerta somente neste momento. |
| PN-04 | Descarte | Descartar uma captura pendente | Registro e removido da fila e do banco, o contador diminui e nenhum edital e criado. |
| PN-05 | Duplicidade | Repetir uma consulta que já retornou o mesmo registro do PNCP | A fila atualiza a mesma captura sem duplicar o edital. |
| PN-06 | Fonte indisponível | Consultar quando a fonte externa não responder | Sistema mostra mensagem clara e não altera a fila existente. |
| PN-07 | Página seguinte | Consultar a página 1 e depois acionar a próxima página | Cada comando consulta apenas a página indicada; os resultados anteriores permanecem na fila. |
| PN-08 | Multimodalidade | Consultar PNCP com quatro modalidades selecionadas | O backend solicita a mesma página de cada modalidade, com até 50 registros por modalidade, e unifica repetições. |
| PN-09 | Saneamento por ordem | Captar um processo no Compras.gov.br e depois no PNCP; repetir em ordem inversa com outro processo | Quando o número de controle PNCP é idêntico, cada par resulta em um único item saneado, independentemente da ordem. |
| PN-10 | Sem controle comum | Captar registros parecidos das duas fontes sem número de controle PNCP comum | Os itens permanecem separados para revisão; objeto, órgão, número e valor não provocam unificação aproximada. |

## 13B. Seguranca e regressao

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| SE-01 | URL direta | Tentar abrir telas administrativas por URL com usuario comum | Acesso e negado pelo frontend e backend. |
| SE-02 | Empresa isolada | Logar em Alfa e tentar abrir identificador de Beta | Dados privados de Beta nao sao expostos. |
| SE-03 | Campos maliciosos | Informar texto com SQL, HTML ou script em formularios | Dados sao tratados como texto; nenhuma consulta ou script e executado. |
| SE-04 | Anexos | Enviar extensao invalida ou arquivo acima do limite | Sistema rejeita com mensagem clara. |
| SE-05 | Reinicio | Reiniciar backend e navegador durante cada fluxo critico | Dados gravados permanecem e telas recarregam sem erro. |
| SE-06 | Duplicidade | Clicar duas vezes em salvar, dar match ou criar tarefa | Nao cria registros duplicados. |
| SE-07 | Auditoria | Bloquear empresa, trocar lider, desistir e cancelar edital | Operacoes ficam registradas para consulta futura. |

## 13C. Avaliacao de parcerias

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| AV-01 | Abertura | Administrador abre rodada com prazo futuro | Empresas ativas sao fotografadas, recebem alerta e a rodada aparece como aberta. |
| AV-02 | Empresa posterior | Aprovar nova empresa depois da abertura | Nova empresa nao entra na rodada atual e participara somente da proxima. |
| AV-03 | Saldo | Abrir rodada com dez participantes | Cada empresa recebe tres estrelas, equivalentes a 30% arredondados para cima. |
| AV-04 | Concentracao | Entregar todo o saldo a uma unica empresa | Sistema aceita qualquer quantidade para a mesma empresa, limitada apenas pelo saldo total. |
| AV-05 | Autoavaliacao | Tentar selecionar a propria empresa | Logo nao permite receber estrelas da propria avaliadora. |
| AV-06 | Saldo incompleto | Distribuir menos que o total | Botao de conclusao permanece inativo e informa quantas estrelas faltam. |
| AV-07 | Revisao | Distribuir todo o saldo e clicar em Concluir distribuicao | Sistema lista empresas e quantidades antes da gravacao. |
| AV-08 | Confirmacao unica | Clicar uma vez em Confirmar envio na revisao | Distribuicao e gravada e a tela passa ao resultado, sem pedir nova confirmacao. |
| AV-09 | Repeticao tecnica | Repetir a requisicao de confirmacao | Backend devolve o estado ja concluido sem duplicar estrelas. |
| AV-10 | Anonimato | Consultar resultado como empresa avaliada | Pontuacao aparece sem identificar quem enviou cada estrela. |
| AV-11 | Prazo | Deixar a data limite passar com empresas pendentes | Rodada fecha automaticamente e pendentes perdem o direito de votar. |
| AV-12 | Fechamento completo | Todas as participantes concluem antes do prazo | Rodada fecha automaticamente quando a ultima distribuicao e gravada. |
| AV-13 | Resultado da sessao | Abrir ranking de rodada encerrada | Indice relativo da sessao e exibido de forma retratil. |
| AV-14 | Media geral | Comparar empresas em varias rodadas encerradas | Resultado geral corresponde a media aritmetica dos indices das rodadas, nao a soma dos pontos. |
| AV-15 | Historico | Filtrar por sessao, ano e quantidade de rodadas | Grafico, media de mercado e tendencia usam apenas as sessoes filtradas. |
| AV-16 | Exclusao fechada | Administrador exclui rodada encerrada | Rodada, distribuicoes, resultados e alertas vinculados somem; media historica e recalculada. |
| AV-17 | Protecao da aberta | Tentar excluir rodada em andamento | Backend impede a exclusao. |

## 13D. Capacidade tecnica e IA

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| CT-01 | Profissional | Cadastrar profissional com mais de uma formacao e formacao complementar | Todas as formacoes ficam vinculadas e disponiveis na consulta. |
| CT-02 | Profissional externo | Registrar atestado de profissional que nao pertence a empresa | Cadastro e permitido e a regra de uso do acervo define empresa, profissional ou ambos. |
| CT-03 | Datas | Informar inicio posterior ao fim da execucao | Sistema impede a gravacao e explica a inconsistencia. |
| CT-04 | Quantitativos | Incluir varios quantitativos com unidades diferentes | Todos ficam vinculados ao mesmo atestado sem impor um tipo unico. |
| CT-05 | PDF pesquisavel | Enviar PDF com camada de texto | Leitura comum e registrada e identificada na listagem. |
| CT-06 | PDF digitalizado | Enviar PDF composto por imagem | OCR e executado, o status informa processamento/conclusao e o texto fica pesquisavel. |
| CT-07 | Correcao | Abrir texto extraido, corrigir e salvar | Versao corrigida substitui o texto de trabalho sem perder os dados do atestado. |
| CT-08 | Pesquisa | Buscar termo existente em varios atestados | Sistema retorna somente os registros cujo texto ou dados contenham o termo. |
| CT-09 | Selecao IA | Marcar varios atestados e abrir Analise com IA | Tela recebe somente os atestados selecionados e monta o conjunto estruturado para analise. |
| CT-10 | IA indisponivel | Solicitar analise sem chave ou credito | Sistema preserva os dados e mostra erro compreensivel sem travar o modulo. |
| CT-11 | OpenAI | Escolher somente OpenAI e concluir uma analise | Resultado identifica OpenAI e o modelo utilizado. |
| CT-12 | Gemini | Escolher somente Google Gemini e concluir uma analise | Resultado identifica Google Gemini e o modelo utilizado. |
| CT-13 | Fallback | Usar modo automatico com OpenAI indisponivel e Gemini configurado | Gemini assume a solicitacao; resultado identifica o provedor real sem duplicar a analise. |
| CT-14 | Groq | Escolher somente Groq e concluir uma analise | Resultado identifica Groq e o modelo utilizado. |
| CT-15 | Fallback completo | Usar modo automatico com OpenAI e Gemini indisponiveis e Groq configurada | Groq assume a solicitacao e o historico identifica o provedor real. |
| CT-16 | Quadro documental | Selecionar atestados com varios quantitativos e abrir a analise | Painel retratil apresenta todos os registros; cada quantitativo conserva descricao, valor e unidade sem soma indevida. |
| CT-17 | Cronograma sem sobreposicao | Selecionar dois periodos consecutivos do mesmo profissional | Cronograma apresenta os dois periodos e a experiencia liquida corresponde ao total cronologico. |
| CT-18 | Cronograma com sobreposicao | Selecionar dois periodos sobrepostos do mesmo profissional | Experiencia bruta soma os periodos, sobreposicao e destacada e experiencia liquida conta cada dia uma unica vez. |
| CT-19 | Profissionais diferentes | Selecionar periodos simultaneos de profissionais diferentes | Cada profissional recebe calculo e cronograma proprios; os periodos nao sao descontados entre pessoas. |
| CT-20 | Data incompleta | Selecionar atestado sem inicio ou fim | Registro aparece no quadro, fica fora do calculo cronologico e a tela informa a pendencia. |
| CT-21 | Fluxo da IA preservado | Abrir e recolher os dois paineis e solicitar uma analise | Roteiro, provedores, JSON e processamento continuam seguindo o fluxo anterior. |

### Pre-analise de editais com IA

| ID | Cenario | Acao | Resultado esperado |
|---|---|---|---|
| IA-ED-01 | OpenAI disponivel | Gerar a pre-analise de um edital com documentos anexados | HTML e salvo no edital e o resultado identifica OpenAI e o modelo. |
| IA-ED-02 | Fallback Gemini | Simular falta de credito ou indisponibilidade da OpenAI com Gemini configurado | Os mesmos documentos e roteiro seguem para o Gemini; o HTML e salvo e identifica Google Gemini e o modelo. |
| IA-ED-03 | Falha dos dois provedores | Tornar OpenAI e Gemini indisponiveis | Job termina como falhou e apresenta as duas causas sem perder os documentos do edital. |
| IA-ED-04 | Substituicao manual | Anexar um HTML depois de uma analise automatica | O novo HTML passa a ser exibido e continua disponivel para download. |

### Triagem da captacao com IA

| ID | Cenario | Acao | Resultado esperado |
|---|---|---|---|
| IA-CAP-01 | Toda a fila | Acionar `Analisar pendentes` | Job percorre todos os registros pendentes em lotes e mostra o progresso. |
| IA-CAP-02 | Selecao | Marcar oportunidades e acionar `Analisar selecionadas` | Somente os identificadores marcados recebem nova classificacao. |
| IA-CAP-03 | Fallback | Simular falha da OpenAI com Gemini configurado | Gemini classifica o lote e seu nome/modelo ficam registrados em cada resultado. |
| IA-CAP-04 | JSON invalido | Simular resposta sem todos os identificadores ou com identificador externo | Resposta e recusada e nenhum resultado inconsistente e gravado. |
| IA-CAP-05 | Filtros | Filtrar por consultiva, relacionada, duvidosa, nao consultiva e pendente | A fila apresenta somente a classificacao escolhida. |
| IA-CAP-06 | Fallback Groq | Simular falha da OpenAI e do Gemini com Groq configurada | Groq classifica o lote em JSON estrito e seu nome/modelo ficam registrados. |
| IA-CAP-06 | Decisao humana | Classificar uma oportunidade como nao consultiva | O sistema nao descarta nem publica automaticamente; os botoes administrativos continuam disponiveis. |
| CT-14 | Resposta desfavoravel | Receber conclusao valida que nao atende ao roteiro | Sistema preserva a resposta e nao chama outro provedor apenas por discordancia do conteudo. |

## 13E. Academia LicitaHub

| ID | Cenario | Passos | Resultado esperado |
|---|---|---|---|
| AC-01 | Curso | Administrador cria curso com capa, carga horaria e descricao | Curso aparece no gerenciamento e, quando publicado, no catalogo. |
| AC-02 | Fonte de video | Cadastrar aula por YouTube e outra por upload | Cada aula aceita somente uma fonte e reproduz o video correspondente. |
| AC-03 | Ordem | Tentar abrir aula seguinte sem concluir a anterior | Sistema impede o avanco. |
| AC-04 | Progresso | Assistir parte do video, sair e retornar | Reproducao continua proxima ao ponto salvo. |
| AC-05 | Conclusao do video | Assistir ate o fim | Aula libera questionario quando ele existir. |
| AC-06 | Reprovacao | Obter menos de 75% no questionario | Sistema mostra quais questoes foram erradas, sem revelar a alternativa correta, e nao libera a proxima aula. |
| AC-07 | Aprovacao | Obter pelo menos 75% | Proxima aula e liberada e o questionario concluido nao pode ser respondido novamente. |
| AC-08 | Catalogo | Iniciar um curso do catalogo | Curso deixa o catalogo de nao iniciados e aparece em Meus cursos. |
| AC-09 | Certificado | Concluir ultima aula e avaliacao | Um unico botao libera PDF formal com dados do usuario, curso, capa, aulas, duracoes e codigo verificavel. |
| AC-10 | Validacao publica | Consultar codigo do certificado | Pagina informa autenticidade e dados essenciais sem exigir acesso administrativo. |
| AC-11 | Administracao | Editar, arquivar e excluir curso conforme suas dependencias | Lista administrativa reflete o novo estado sem afetar cursos alheios. |

## 14. Criterio de aceite

O sistema so esta pronto para um piloto com usuarios reais quando:

1. Todos os testes criticos de acesso, edital, match, consorcio, montagem e cancelamento forem aprovados.
2. Nenhuma acao deixar anuncios, chats, consorcios, calendario ou tarefas em estado contraditorio.
3. Perfis sem permissao nao conseguirem executar acoes por tela nem por URL direta.
4. Cada problema encontrado tiver registro, causa, correcao aplicada e novo teste de confirmacao.
