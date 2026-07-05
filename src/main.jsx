import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const roles = {
  platformAdmin: "Administrador da plataforma",
  companyAdmin: "Administrador da empresa",
  commercial: "Comercial / Relacionamento",
  technical: "Técnico",
  reader: "Leitor"
};

const modules = [
  {
    id: "access",
    label: "Acesso e administração",
    roles: ["platformAdmin"],
    items: [
      { id: "admin-dashboard", label: "Painel administrativo" },
      { id: "invite-new", label: "Novo convite" },
      { id: "invite-list", label: "Lista de convites" },
      { id: "company-review", label: "Análise de empresas" }
    ]
  },
  {
    id: "company",
    label: "Empresa",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "company-dashboard", label: "Dashboard" },
      { id: "company-profile-edit", label: "Editar perfil" },
      { id: "company-users", label: "Usuários vinculados", roles: ["companyAdmin"] },
      { id: "company-user-profile", label: "Cadastro de usuário", roles: ["companyAdmin"] }
    ]
  },
  {
    id: "community",
    label: "Comunidade",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "community-home", label: "Comunidade" },
      { id: "company-public-profile", label: "Perfil público" },
      { id: "publication-new", label: "Criar publicação", roles: ["companyAdmin", "commercial"] },
      { id: "publication-list", label: "Minhas publicações", roles: ["companyAdmin", "commercial"] }
    ]
  },
  {
    id: "tenders",
    label: "Editais",
    roles: ["platformAdmin", "companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "tender-admin", label: "Admin editais", roles: ["platformAdmin"] },
      { id: "tender-new", label: "Cadastro de edital", roles: ["platformAdmin"] },
      { id: "tender-list", label: "Lista de editais", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "tender-detail", label: "Detalhe do edital", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "tender-interest", label: "Interesse no edital", roles: ["companyAdmin", "commercial"] },
      { id: "tender-interest-list", label: "Empresas interessadas", roles: ["companyAdmin", "commercial"] }
    ]
  },
  {
    id: "radar",
    label: "Radar LicitaHub",
    roles: ["platformAdmin", "companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "radar-home", label: "Notícias" },
      { id: "radar-detail", label: "Detalhe da notícia" },
      { id: "radar-new", label: "Cadastrar notícia", roles: ["platformAdmin"] }
    ]
  },
  {
    id: "match",
    label: "Match e consórcios",
    roles: ["companyAdmin", "commercial"],
    items: [
      { id: "match-partners", label: "Match de parceiros" },
      { id: "match-tinder", label: "Avaliar candidata" },
      { id: "match-profile", label: "Perfil no match" },
      { id: "match-success", label: "Match realizado" },
      { id: "matches-by-tender", label: "Matches por edital" }
    ]
  }
];

const stats = [
  ["Editais compatíveis", "9"],
  ["Interesses ativos", "3"],
  ["Matches abertos", "2"],
  ["Publicações", "14"]
];

const tenders = [
  {
    agency: "Prefeitura Municipal",
    number: "CP 004/2026",
    object: "Projetos de saneamento e drenagem urbana",
    location: "MG",
    value: "R$ 2.400.000,00",
    status: "Publicado"
  },
  {
    agency: "Departamento de Estradas",
    number: "TP 012/2026",
    object: "Supervisão de obras rodoviárias",
    location: "PR",
    value: "R$ 5.800.000,00",
    status: "Em avaliação"
  }
];

const partners = [
  {
    name: "GeoArq Projetos",
    location: "Belo Horizonte - MG",
    offers: "Arqueologia, estudos socioambientais e equipe de campo.",
    seeks: "Coordenação técnica em saneamento e proposta técnica.",
    fit: "Alta"
  },
  {
    name: "SocialTec Consultoria",
    location: "Salvador - BA",
    offers: "Projetos sociais, comunicação comunitária e reassentamento.",
    seeks: "Empresa líder com experiência em infraestrutura urbana.",
    fit: "Média"
  }
];

function canSee(item, role) {
  return !item.roles || item.roles.includes(role);
}

function firstScreenFor(role) {
  const module = modules.find((group) => group.roles.includes(role));
  const item = module?.items.find((entry) => canSee(entry, role));
  return item?.id || "company-dashboard";
}

function useHashScreen(role) {
  const [screen, setScreen] = useState(() => window.location.hash.replace("#", "") || firstScreenFor(role));

  useEffect(() => {
    const onHashChange = () => setScreen(window.location.hash.replace("#", "") || firstScreenFor(role));
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, [role]);

  useEffect(() => {
    const allowed = modules.some((group) =>
      group.roles.includes(role) && group.items.some((item) => item.id === screen && canSee(item, role))
    );
    if (!allowed) {
      const next = firstScreenFor(role);
      window.location.hash = next;
      setScreen(next);
    }
  }, [role, screen]);

  return screen;
}

function App() {
  const [role, setRole] = useState("companyAdmin");
  const [menuCollapsed, setMenuCollapsed] = useState(false);
  const screen = useHashScreen(role);
  const visibleModules = useMemo(() => modules.filter((group) => group.roles.includes(role)), [role]);

  return (
    <div className={`app ${menuCollapsed ? "menuCollapsed" : ""}`}>
      <aside className="sidebar">
        <div className="brand">
          <span className="brandMark">LH</span>
          <div>
            <strong>LicitaHub</strong>
            <small>Rede consultiva</small>
          </div>
        </div>

        <button className="menuToggle" onClick={() => setMenuCollapsed(!menuCollapsed)}>
          {menuCollapsed ? "Expandir menu" : "Recolher menu"}
        </button>

        <label className="roleSwitch">
          Visualizar como
          <select value={role} onChange={(event) => setRole(event.target.value)}>
            {Object.entries(roles).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>

        <nav>
          {visibleModules.map((group) => (
            <div className="navGroup" key={group.id}>
              <span>{group.label}</span>
              {group.items.filter((item) => canSee(item, role)).map((item) => (
                <a className={screen === item.id ? "active" : ""} href={`#${item.id}`} key={item.id}>
                  {item.label}
                </a>
              ))}
            </div>
          ))}
        </nav>
      </aside>

      <main className="main">
        <Topbar role={roles[role]} />
        <Screen screen={screen} navigate={(id) => { window.location.hash = id; }} />
      </main>
    </div>
  );
}

function Topbar({ role }) {
  return (
    <header className="topbar">
      <div>
        <span className="eyebrow">Ambiente de produto</span>
        <h1>LicitaHub</h1>
      </div>
      <div className="userPill">
        <span>Perfil ativo</span>
        <strong>{role}</strong>
      </div>
    </header>
  );
}

function Screen({ screen, navigate }) {
  const screens = {
    "admin-dashboard": <AdminDashboard />,
    "invite-new": <InviteNew />,
    "invite-list": <InviteList />,
    "company-review": <CompanyReview />,
    "company-dashboard": <CompanyDashboard />,
    "company-profile-edit": <CompanyProfileEdit />,
    "company-users": <CompanyUsers />,
    "company-user-profile": <CompanyUserProfile />,
    "community-home": <CommunityHome />,
    "company-public-profile": <CompanyPublicProfile />,
    "publication-new": <PublicationNew />,
    "publication-list": <PublicationList />,
    "radar-home": <RadarHome navigate={navigate} />,
    "radar-detail": <RadarDetail />,
    "radar-new": <RadarNew />,
    "tender-admin": <TenderAdmin />,
    "tender-new": <TenderNew />,
    "tender-list": <TenderList />,
    "tender-detail": <TenderDetail />,
    "tender-interest": <TenderInterest navigate={navigate} />,
    "tender-interest-list": <TenderInterestList navigate={navigate} />,
    "match-partners": <MatchPartners />,
    "match-tinder": <MatchTinder navigate={navigate} />,
    "match-profile": <MatchProfile />,
    "match-success": <MatchSuccess />,
    "matches-by-tender": <MatchesByTender />
  };

  return screens[screen] || <CompanyDashboard />;
}

function Page({ label, title, children, actions }) {
  return (
    <section className="page">
      <div className="pageHeader">
        <div>
          <span className="eyebrow">{label}</span>
          <h2>{title}</h2>
          <p className="pageHelp">{getPageHelp(title)}</p>
        </div>
        {actions && <div className="actions">{actions}</div>}
      </div>
      {children}
    </section>
  );
}

function getPageHelp(title) {
  const help = {
    "Painel administrativo": "Acompanhe convites, empresas pendentes, editais e pontos que exigem ação da plataforma.",
    "Novo convite de empresa": "Cadastre a empresa que será convidada. CNPJ e nome fantasia identificam a empresa de forma única.",
    "Lista de convites": "Veja o andamento dos convites enviados, identifique pendências e acompanhe quem já iniciou cadastro.",
    "Análise e aprovação da empresa": "Revise os dados enviados pela empresa e decida se ela entra, ajusta informações ou será recusada.",
    "Dashboard da empresa": "Resumo operacional da empresa: oportunidades, matches, comunidade e próximos passos em um só lugar.",
    "Editar perfil da empresa": "Mantenha a vitrine institucional atualizada. Essas informações aparecem na comunidade e no match.",
    "Usuários vinculados": "Gerencie quem opera pela empresa. O perfil de acesso define as permissões de cada pessoa.",
    "Cadastro do usuário vinculado": "Inclua ou edite uma pessoa da empresa, escolhendo o perfil adequado para sua função.",
    "Rede de empresas": "Acompanhe publicações, encontre empresas por tema e fortaleça a presença institucional da sua empresa.",
    "Perfil público da empresa": "Veja como a empresa aparece para a comunidade: identidade, destaques, categorias e publicações.",
    "Criar publicação": "Publique fotos, notícias, eventos, conquistas ou conteúdo técnico no perfil e na comunidade.",
    "Minhas publicações": "Gerencie o conteúdo publicado pela empresa e acompanhe o que está em rascunho ou visível.",
    "Notícias e inteligência de mercado": "Conteúdo publicado pela LicitaHub para orientar empresas sobre mercado, editais e tendências.",
    "Detalhe da notícia": "Leitura completa da notícia, com contexto e orientações úteis para a comunidade.",
    "Cadastrar notícia": "Área do administrador para publicar comunicados, notícias e análises da plataforma.",
    "Painel administrativo de editais": "Controle os editais cadastrados, seus status e oportunidades que já geraram interesse.",
    "Cadastro de edital": "Registre a oportunidade e organize os dados principais para que empresas possam avaliar com clareza.",
    "Lista de editais": "Encontre oportunidades, marque interesse e abra o detalhe completo de cada edital.",
    "Detalhe do edital": "Entenda a oportunidade, veja exigências e acesse a ficha técnica antes de decidir participar.",
    "Manifestação de interesse": "Registre a posição da empresa e indique se deseja buscar parceiros para esta licitação.",
    "Empresas interessadas no edital": "Veja empresas que também demonstraram interesse e escolha quais avaliar como possíveis parceiras.",
    "Possíveis parceiros": "Compare empresas candidatas e entenda rapidamente quem pode complementar sua participação.",
    "Avaliar candidata da licitação": "Avalie uma empresa por vez: veja o que ela oferece, o que falta e decida recusar ou dar match.",
    "Perfil da empresa no match": "Visão detalhada da candidata dentro da licitação, com oferta, necessidades e aderência.",
    "Match realizado": "Confirmação de interesse recíproco entre empresas na mesma licitação.",
    "Matches por edital": "Acompanhe matches já realizados e inicie contato pelo WhatsApp do responsável."
  };
  return help[title] || "Tela da LicitaHub para apoiar decisões empresariais com clareza e contexto.";
}

function Button({ children, variant = "primary", onClick }) {
  return <button className={`btn ${variant}`} onClick={onClick}>{children}</button>;
}

function Field({ label, children, hint }) {
  return (
    <label className="field">
      <span>{label}</span>
      {children}
      {hint && <small>{hint}</small>}
    </label>
  );
}

function Card({ children, className = "" }) {
  return <div className={`card ${className}`}>{children}</div>;
}

function AdminDashboard() {
  return (
    <Page label="Administração" title="Painel administrativo">
      <Stats items={[["Convites enviados", "12"], ["Empresas pendentes", "4"], ["Empresas ativas", "28"], ["Editais abertos", "9"]]} />
      <div className="grid two">
        <Card><h3>Entrada de empresas</h3><p>Convites, aprovações e bloqueios ficam concentrados aqui.</p><Button>Novo convite</Button></Card>
        <Card><h3>Editais</h3><p>Cadastre oportunidades e acompanhe manifestações de interesse.</p><Button variant="secondary">Cadastrar edital</Button></Card>
      </div>
    </Page>
  );
}

function InviteNew() {
  return (
    <Page label="Convites" title="Novo convite de empresa" actions={<Button>Enviar convite</Button>}>
      <FormGrid>
        <Field label="Nome fantasia" hint="Obrigatório. Único."><input placeholder="Engenvale Consultoria" /></Field>
        <Field label="CNPJ" hint="Obrigatório. Único."><input placeholder="00.000.000/0000-00" /></Field>
        <Field label="Contato principal" hint="Obrigatório. Pode repetir."><input placeholder="Nome completo" /></Field>
        <Field label="E-mail" hint="Obrigatório. Pode repetir."><input type="email" placeholder="contato@empresa.com.br" /></Field>
        <Field label="Telefone" hint="Obrigatório. Pode repetir."><input placeholder="(00) 00000-0000" /></Field>
        <Field label="Estado"><select><option>Selecione</option><option>SP</option><option>MG</option></select></Field>
      </FormGrid>
      <Field label="Observação interna"><textarea placeholder="Visível apenas para administradores" /></Field>
    </Page>
  );
}

function InviteList() {
  return (
    <Page label="Convites" title="Lista de convites" actions={<Button>Novo convite</Button>}>
      <Table columns={["Empresa", "CNPJ", "Contato", "E-mail", "Telefone", "Status"]} rows={[
        ["Engenvale Consultoria", "12.345.678/0001-90", "Marina Costa", "marina@engenvale.com.br", "(11) 98888-1000", "Enviado"],
        ["Plano Sul Engenharia", "22.222.222/0001-22", "Carlos Lima", "contato@planosul.com.br", "(41) 3333-2020", "Aguardando aprovação"]
      ]} />
    </Page>
  );
}

function CompanyReview() {
  return (
    <Page label="Administração" title="Análise e aprovação da empresa">
      <FormGrid>
        <Field label="Empresa"><input value="Plano Sul Engenharia" readOnly /></Field>
        <Field label="CNPJ"><input value="22.222.222/0001-22" readOnly /></Field>
        <Field label="Contato"><input value="Carlos Lima" readOnly /></Field>
        <Field label="Status"><input value="Aguardando aprovação" readOnly /></Field>
      </FormGrid>
      <Field label="Solicitação de ajuste"><textarea placeholder="Informe o que precisa ser corrigido" /></Field>
      <div className="actions"><Button>Aprovar</Button><Button variant="secondary">Solicitar ajuste</Button><Button variant="danger">Recusar</Button></div>
    </Page>
  );
}

function CompanyDashboard() {
  return (
    <Page label="Empresa" title="Dashboard da empresa">
      <Stats items={stats} />
      <div className="grid three">
        <Card><h3>Próximos editais</h3><p>3 oportunidades têm aderência alta com o perfil da empresa.</p></Card>
        <Card><h3>Comunidade</h3><p>Novas publicações de empresas de saneamento e supervisão ambiental.</p></Card>
        <Card><h3>Matches</h3><p>2 conversas aguardam atualização de status.</p></Card>
      </div>
    </Page>
  );
}

function CompanyProfileEdit() {
  return (
    <Page label="Empresa" title="Editar perfil da empresa" actions={<Button>Salvar perfil</Button>}>
      <div className="profileEditGrid">
        <Card>
          <h3>Identidade visual</h3>
          <div className="logoPreview">EC</div>
          <Field label="Logomarca da empresa" hint="PNG, JPG ou SVG. Recomenda-se imagem quadrada. Usada no perfil público, comunidade e avaliação candidata."><input type="file" accept="image/png,image/jpeg,image/svg+xml" /></Field>
        </Card>
        <div>
          <FormGrid>
            <Field label="Nome fantasia" hint="Único."><input value="Engenvale Consultoria" readOnly /></Field>
            <Field label="CNPJ" hint="Único."><input value="12.345.678/0001-90" readOnly /></Field>
            <Field label="Site"><input placeholder="https://www.empresa.com.br" /></Field>
            <Field label="Porte"><select><option>Média</option><option>Pequena</option><option>Grande</option></select></Field>
          </FormGrid>
          <Field label="Descrição institucional"><textarea placeholder="Resumo profissional da atuação da empresa" /></Field>
        </div>
      </div>
      <h3>Áreas técnicas</h3>
      <AreaChips />
    </Page>
  );
}

function CompanyUsers() {
  return (
    <Page label="Empresa" title="Usuários vinculados" actions={<Button>Adicionar usuário</Button>}>
      <Card className="notice">
        <strong>Permissões por perfil</strong>
        <p>As permissões não são marcadas individualmente. Cada usuário recebe um perfil de acesso, e o perfil define o que ele pode fazer.</p>
      </Card>
      <Table columns={["Nome", "E-mail", "Perfil", "Status"]} rows={[
        ["Marina Costa", "marina@engenvale.com.br", "Administrador", "Ativo"],
        ["Renato Alves", "renato@engenvale.com.br", "Técnico", "Ativo"],
        ["Paula Martins", "paula@engenvale.com.br", "Comercial", "Convite pendente"]
      ]} />
    </Page>
  );
}

function CompanyUserProfile() {
  return (
    <Page label="Empresa" title="Cadastro do usuário vinculado" actions={<Button>Salvar usuário</Button>}>
      <FormGrid>
        <Field label="Nome completo"><input placeholder="Nome do usuário" /></Field>
        <Field label="E-mail"><input type="email" placeholder="usuario@empresa.com.br" /></Field>
        <Field label="Telefone"><input placeholder="(00) 00000-0000" /></Field>
        <Field label="Cargo ou função"><input placeholder="Ex.: Coordenador técnico" /></Field>
        <Field label="Perfil de acesso"><select><option>Administrador da empresa</option><option>Comercial / Relacionamento</option><option>Técnico</option><option>Leitor</option></select></Field>
        <Field label="Status"><select><option>Convite pendente</option><option>Ativo</option><option>Bloqueado</option><option>Inativo</option></select></Field>
      </FormGrid>
      <Card>
        <h3>Permissões aplicadas pelo perfil</h3>
        <p>Administrador da empresa: pode editar perfil, gerenciar usuários, criar publicações, manifestar interesse em editais, participar do match e responder mensagens.</p>
      </Card>
      <Field label="Observação interna"><textarea placeholder="Anotação visível apenas para administradores da empresa" /></Field>
    </Page>
  );
}

function CommunityHome() {
  return (
    <Page label="Comunidade" title="Rede de empresas" actions={<Button variant="secondary">Criar publicação</Button>}>
      <div className="communityToolbar">
        <div className="communitySearch">
          <span>Buscar</span>
          <input placeholder="Empresa, especialidade, notícia ou região" />
        </div>
        <select><option>Brasil inteiro</option><option>SP</option><option>MG</option><option>PR</option></select>
      </div>

      <div className="communityLayout">
        <aside className="communityRail">
          <Card className="companyMiniProfile">
            <span className="avatar">E</span>
            <strong>Engenvale Consultoria</strong>
            <p>Perfil 82% completo</p>
            <div className="progress"><i style={{width: "82%"}}></i></div>
          </Card>
          <Card>
            <h3>Categorias</h3>
            <div className="compactTags">
              {["Equipe", "Notícias", "Atividades", "Eventos", "Conquistas", "Técnico", "Destaques"].map((item) => <button key={item}>{item}</button>)}
            </div>
          </Card>
        </aside>

        <section className="communityFeed">
          <div className="composer modern">
            <span className="avatar">E</span>
            <div>
              <button className="composerInput">Compartilhar notícia, foto, evento ou conquista...</button>
              <div className="composerActions"><button>Imagem</button><button>Evento</button><button>Conteúdo técnico</button></div>
            </div>
          </div>
          <div className="categoryTabs compact">
            {["Todos", "Equipe comercial", "Notícias", "Atividades", "Eventos", "Conquistas", "Conteúdo técnico", "Destaques"].map((item) => <button key={item}>{item}</button>)}
          </div>
          <div className="socialFeed">
            <PostCard category="Notícias" company="GeoArq Projetos" title="Estudos arqueológicos preventivos em nova frente de infraestrutura" />
            <PostCard category="Conquistas" company="Plano Sul Engenharia" title="Certificação técnica da equipe de supervisão ambiental" />
            <PostCard category="Equipe comercial" company="Engenvale Consultoria" title="Equipe comercial reunida para planejamento de oportunidades públicas" />
          </div>
        </section>

        <aside className="communityAside">
          <Card>
            <h3>Empresas sugeridas</h3>
            <div className="suggestList">
              <CompanySuggestion name="SocialTec" area="Projetos sociais" />
              <CompanySuggestion name="GeoArq" area="Arqueologia e meio ambiente" />
              <CompanySuggestion name="Plano Sul" area="Supervisão rodoviária" />
            </div>
          </Card>
          <Card><h3>Destaque LicitaHub</h3><p>Empresas com perfil completo tendem a aparecer melhor em matches e recomendações.</p></Card>
        </aside>
      </div>
    </Page>
  );
}

function CompanySuggestion({ name, area }) {
  return (
    <div className="suggestItem">
      <span className="avatar">{name[0]}</span>
      <div><strong>{name}</strong><small>{area}</small></div>
      <button>Ver</button>
    </div>
  );
}

function CompanyPublicProfile() {
  return (
    <Page label="Comunidade" title="Perfil público da empresa">
      <div className="profileHero">
        <div>
          <span className="avatar">E</span>
          <h3>Engenvale Consultoria</h3>
          <p>São Paulo - SP | Atuação nacional</p>
        </div>
        <Button variant="secondary">Salvar empresa</Button>
      </div>
      <Card><h3>Áreas técnicas</h3><p>Saneamento, supervisão ambiental, projetos de engenharia e gerenciamento.</p></Card>
      <Card><h3>Experiências públicas</h3><p>Projetos de saneamento e drenagem urbana para municípios de médio porte.</p></Card>
    </Page>
  );
}

function PublicationNew() {
  return (
    <Page label="Comunidade" title="Criar publicação" actions={<Button>Publicar</Button>}>
      <FormGrid>
        <Field label="Tipo"><select><option>Notícia institucional</option><option>Atividade recente</option><option>Conteúdo técnico</option></select></Field>
        <Field label="Visibilidade"><select><option>Publicado na comunidade</option><option>Apenas no perfil</option><option>Rascunho</option></select></Field>
        <Field label="Título"><input placeholder="Título da publicação" /></Field>
        <Field label="Área relacionada"><select><option>Saneamento</option><option>Supervisão ambiental</option></select></Field>
      </FormGrid>
      <Field label="Texto"><textarea placeholder="Escreva a atualização institucional" /></Field>
    </Page>
  );
}

function PublicationList() {
  return <Page label="Comunidade" title="Minhas publicações"><Table columns={["Título", "Tipo", "Visibilidade", "Status"]} rows={[["Nova equipe de saneamento", "Notícia", "Comunidade", "Publicado"], ["Seminário técnico", "Evento", "Perfil", "Rascunho"]]} /></Page>;
}

function RadarHome({ navigate }) {
  return (
    <Page label="Radar LicitaHub" title="Notícias e inteligência de mercado">
      <div className="newsHero">
        <div className="newsHeroImage">Radar</div>
        <div className="newsHeroText">
          <span className="badge">Licitações</span>
          <h3>Nova rodada de oportunidades em infraestrutura deve movimentar consórcios técnicos</h3>
          <p>Levantamento da LicitaHub aponta aumento de editais com exigências multidisciplinares em saneamento, meio ambiente e supervisão.</p>
          <Button onClick={() => navigate("radar-detail")}>Ler notícia</Button>
        </div>
      </div>
      <div className="categoryTabs">{["Todos", "Licitações", "Mercado", "Legislação", "Eventos", "Comunicados"].map((item) => <button key={item}>{item}</button>)}</div>
      <div className="newsGrid">
        <NewsCard category="Mercado" title="Empresas consultivas ampliam busca por parceiros regionais" navigate={navigate} />
        <NewsCard category="Legislação" title="Mudanças em critérios técnicos exigem atenção na proposta" navigate={navigate} />
        <NewsCard category="Eventos" title="Agenda de eventos técnicos para engenharia consultiva" navigate={navigate} />
        <NewsCard category="Comunicados" title="LicitaHub prepara nova área de recomendações por perfil" navigate={navigate} />
      </div>
    </Page>
  );
}

function NewsCard({ category, title, navigate }) {
  return (
    <Card className="newsCard">
      <div className="newsThumb">{category}</div>
      <span className="badge">{category}</span>
      <h3>{title}</h3>
      <p>Resumo curto da notícia para orientar empresas associadas e usuários vinculados.</p>
      <button className="textLink" onClick={() => navigate("radar-detail")}>Ler notícia</button>
    </Card>
  );
}

function RadarDetail() {
  return (
    <Page label="Radar LicitaHub" title="Detalhe da notícia">
      <article className="article">
        <div className="articleImage">LicitaHub Radar</div>
        <span className="badge">Licitações</span>
        <h3>Nova rodada de oportunidades em infraestrutura deve movimentar consórcios técnicos</h3>
        <p className="articleMeta">Publicado pela LicitaHub em 04/07/2026</p>
        <p>O mercado de engenharia consultiva segue observando editais com escopos mais amplos e exigências multidisciplinares.</p>
        <p>Empresas que antes avaliavam oportunidades de forma isolada passam a buscar composição técnica mais cedo.</p>
        <p>A LicitaHub recomenda que empresas mantenham seus perfis atualizados e registrem com clareza o que podem oferecer.</p>
      </article>
    </Page>
  );
}

function RadarNew() {
  return (
    <Page label="Radar LicitaHub" title="Cadastrar notícia" actions={<Button>Publicar notícia</Button>}>
      <FormGrid>
        <Field label="Título"><input placeholder="Título da notícia" /></Field>
        <Field label="Categoria"><select><option>Licitações</option><option>Mercado</option><option>Legislação</option><option>Eventos</option><option>Comunicados</option></select></Field>
        <Field label="Status"><select><option>Rascunho</option><option>Publicado</option><option>Destaque principal</option></select></Field>
      </FormGrid>
      <Field label="Imagem"><input type="file" accept="image/*" /></Field>
      <Field label="Resumo"><textarea placeholder="Resumo que aparece nos cards" /></Field>
      <Field label="Texto completo"><textarea placeholder="Texto completo da notícia" /></Field>
    </Page>
  );
}

function TenderAdmin() {
  return <Page label="Editais" title="Painel administrativo de editais" actions={<Button>Cadastrar edital</Button>}><Stats items={[["Publicados", "9"], ["Rascunhos", "3"], ["Suspensos", "2"], ["Com interesse", "14"]]} /><TenderTable admin /></Page>;
}

function TenderNew() {
  return (
    <Page label="Editais" title="Cadastro de edital" actions={<Button>Publicar edital</Button>}>
      <FormGrid>
        <Field label="Órgão público"><input placeholder="Prefeitura Municipal" /></Field>
        <Field label="Número do edital"><input placeholder="CP 004/2026" /></Field>
        <Field label="Modalidade"><select><option>Concorrência</option><option>Pregão</option></select></Field>
        <Field label="Valor estimado"><input placeholder="R$ 0,00" /></Field>
        <Field label="Estado"><select><option>MG</option><option>SP</option><option>PR</option></select></Field>
        <Field label="Data de abertura"><input type="date" /></Field>
      </FormGrid>
      <Field label="Objeto"><textarea placeholder="Descreva o objeto do edital" /></Field>
      <Field label="Link do diretório em nuvem"><input placeholder="https://drive.google.com/..." /></Field>
    </Page>
  );
}

function TenderList() {
  return <Page label="Editais" title="Lista de editais"><TenderTable /></Page>;
}

function TenderDetail() {
  return (
    <Page label="Editais" title="Detalhe do edital" actions={<><Button>Registrar interesse</Button><a className="downloadButton" href="./ficha-tecnica-edital-dnit-0226-2026.html" download>Baixar ficha técnica</a></>}>
      <Card><h3>CP 004/2026 - Projetos de saneamento e drenagem</h3><p>Contratação de empresa especializada para projetos de saneamento, drenagem urbana e apoio técnico municipal.</p></Card>
      <div className="grid three">
        <Card><strong>Equipe técnica</strong><p>Coordenação, engenheiro sanitarista e especialista em drenagem.</p></Card>
        <Card><strong>Experiência</strong><p>Atestados em saneamento, drenagem ou infraestrutura municipal.</p></Card>
        <Card><strong>Proposta técnica</strong><p>Metodologia, cronograma, equipe e controle de qualidade.</p></Card>
      </div>
      <Card>
        <div className="cardHeader"><h3>Ficha técnica do edital</h3><a className="downloadButton" href="./ficha-tecnica-edital-dnit-0226-2026.html" download>Baixar ficha técnica</a></div>
        <iframe className="technicalSheetFrame" src="./ficha-tecnica-edital-dnit-0226-2026.html" title="Ficha técnica do edital DNIT 0226/2026"></iframe>
      </Card>
    </Page>
  );
}

function TenderInterest({ navigate }) {
  return (
    <Page label="Editais" title="Manifestação de interesse" actions={<Button onClick={() => navigate("tender-interest-list")}>Salvar e ver empresas interessadas</Button>}>
      <div className="choiceGrid">
        {["Tenho interesse", "Em avaliação", "Não tenho interesse"].map((item) => <label key={item}><input name="interest" type="radio" />{item}</label>)}
      </div>
      <div className="choiceGrid">
        {["Quero buscar parceiros", "Posso liderar consórcio", "Quero participar como parceira", "Preciso complementar equipe", "Preciso complementar experiência"].map((item) => <label key={item}><input type="checkbox" />{item}</label>)}
      </div>
      <Field label="Observação interna"><textarea placeholder="Privado da empresa" /></Field>
    </Page>
  );
}

function TenderInterestList({ navigate }) {
  return (
    <Page label="Editais" title="Empresas interessadas no edital" actions={<Button onClick={() => navigate("match-tinder")}>Avaliar empresas</Button>}>
      <Card className="notice">
        <strong>CP 004/2026 - Projetos de saneamento e drenagem</strong>
        <p>Empresas abaixo também manifestaram interesse e aceitaram aparecer para avaliação de possíveis parceiros nesta licitação.</p>
      </Card>
      <Table columns={["Empresa", "Local", "Oferece", "Procura", "Status", "Ações"]} rows={[
        ["GeoArq Projetos", "Belo Horizonte - MG", "Arqueologia e supervisão ambiental", "Liderança em saneamento", "Disponível para match", <Button key="geo" onClick={() => navigate("match-tinder")}>Avaliar</Button>],
        ["SocialTec Consultoria", "Salvador - BA", "Projetos sociais e comunicação", "Empresa líder de proposta", "Disponível para match", <Button key="social" onClick={() => navigate("match-tinder")}>Avaliar</Button>],
        ["Plano Sul Engenharia", "Curitiba - PR", "Supervisão e gerenciamento", "Complemento ambiental", "Avaliar depois", <Button key="plano" variant="secondary" onClick={() => navigate("match-tinder")}>Abrir</Button>]
      ]} />
    </Page>
  );
}

function MatchPartners() {
  return (
    <Page label="Match e consórcios" title="Possíveis parceiros">
      <Card className="notice"><strong>Sua exposição nesta licitação</strong><p>Buscando parceiros para complementar equipe e experiência técnica em saneamento e drenagem.</p></Card>
      <div className="grid two">{partners.map((partner) => <PartnerCard key={partner.name} partner={partner} />)}</div>
    </Page>
  );
}

function MatchTinder({ navigate }) {
  return (
    <Page label="Match e consórcios" title="Avaliar candidata da licitação">
      <div className="tinderStage">
        <div className="tinderPhone">
          <div className="tinderPhoto">
            <div className="companyLogo"><span>GA</span><small>logo cadastrada</small></div>
            <span className="tenderChip">CP 004/2026</span>
            <div className="tinderIdentity">
              <h3>GeoArq Projetos</h3>
              <p>Belo Horizonte - MG</p>
            </div>
          </div>

          <div className="tinderInfo">
            <strong>Aderência alta para complemento socioambiental</strong>
            <div className="matchColumns">
              <div><strong>Tem</strong><span>Arqueologia, supervisão ambiental, estudos socioambientais e equipe de campo.</span></div>
              <div><strong>Falta</strong><span>Coordenação em saneamento, liderança da proposta e atestados principais.</span></div>
            </div>
          </div>

          <div className="tinderBottomActions">
            <button className="circleButton info">i</button>
            <button className="circleButton reject">×</button>
            <button className="circleButton like" onClick={() => navigate("match-success")}>♥</button>
            <button className="circleButton save">★</button>
          </div>
        </div>
      </div>
    </Page>
  );
}

function MatchProfile() {
  return (
    <Page label="Match e consórcios" title="Perfil da empresa no match" actions={<Button>Curtir empresa</Button>}>
      <Card><h3>GeoArq Projetos</h3><p>Belo Horizonte - MG | Aderência alta</p></Card>
      <Card><h3>Oferece</h3><p>Equipe de arqueologia, supervisão ambiental, estudos socioambientais e relatórios de campo.</p></Card>
      <Card><h3>Procura</h3><p>Empresa com experiência em saneamento, coordenação técnica e liderança de proposta.</p></Card>
    </Page>
  );
}

function MatchSuccess() {
  return (
    <Page label="Match e consórcios" title="Match realizado" actions={<Button>Ver matches por edital</Button>}>
      <Card className="success"><h3>Engenvale Consultoria + GeoArq Projetos</h3><p>As duas empresas demonstraram interesse recíproco na licitação CP 004/2026. A conversa será iniciada fora da plataforma, pelo WhatsApp do responsável informado.</p></Card>
    </Page>
  );
}

function MatchesByTender() {
  return (
    <Page label="Match e consórcios" title="Matches por edital">
      <Card className="notice">
        <strong>Conversas fora da plataforma</strong>
        <p>No MVP, a LicitaHub registra o match e encaminha a conversa pelo WhatsApp do responsável indicado pela empresa.</p>
      </Card>
      <Table columns={["Edital", "Empresa", "Responsável", "WhatsApp", "Status", "Ações"]} rows={[
        ["CP 004/2026", "GeoArq Projetos", "Ana Ribeiro", "(31) 97777-3030", <span className="statusPill open" key="geo-status">Match realizado</span>, <div className="rowActions" key="geo-whats"><a className="whatsappButton" href="https://wa.me/5531977773030?text=Olá%2C%20somos%20da%20Engenvale%20Consultoria.%20Demos%20match%20na%20LicitaHub%20para%20a%20licitação%20CP%20004%2F2026%20e%20gostaríamos%20de%20conversar%20sobre%20possível%20parceria." target="_blank">Abrir WhatsApp</a><Button variant="secondary">Ver perfil</Button></div>],
        ["TP 012/2026", "Plano Sul Engenharia", "Carlos Lima", "(41) 3333-2020", <span className="statusPill review" key="plano-status">Em avaliação</span>, <div className="rowActions" key="plano-whats"><a className="whatsappButton" href="https://wa.me/554133332020" target="_blank">Abrir WhatsApp</a><Button variant="secondary">Ver perfil</Button></div>]
      ]} />
    </Page>
  );
}

function TenderTable() {
  return <Table columns={["Órgão", "Número", "Objeto", "Local", "Valor", "Status", "Ações"]} rows={tenders.map((tender) => [
    tender.agency,
    tender.number,
    tender.object,
    tender.location,
    tender.value,
    <span className={`statusPill ${tender.status === "Publicado" ? "open" : "review"}`} key={`${tender.number}-status`}>{tender.status}</span>,
    <div className="rowActions" key={`${tender.number}-actions`}>
      <Button>Tenho interesse</Button>
      <Button variant="secondary">Ver edital</Button>
    </div>
  ])} />;
}

function PartnerCard({ partner }) {
  return (
    <Card>
      <h3>{partner.name}</h3>
      <p>{partner.location}</p>
      <p><strong>Oferece:</strong> {partner.offers}</p>
      <p><strong>Procura:</strong> {partner.seeks}</p>
      <span className="badge">Aderência {partner.fit}</span>
      <div className="actions"><Button>Curtir</Button><Button variant="secondary">Avaliar depois</Button><Button variant="danger">Ignorar</Button></div>
    </Card>
  );
}

function Stats({ items }) {
  return <div className="stats">{items.map(([label, value]) => <Card key={label} className="stat"><strong>{value}</strong><span>{label}</span></Card>)}</div>;
}

function FormGrid({ children }) {
  return <div className="formGrid">{children}</div>;
}

function Table({ columns, rows }) {
  return (
    <div className="tableWrap">
      <table>
        <thead><tr>{columns.map((column) => <th key={column}>{column}</th>)}</tr></thead>
        <tbody>{rows.map((row, index) => <tr key={index}>{row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}</tr>)}</tbody>
      </table>
    </div>
  );
}

function AreaChips() {
  return (
    <div className="chips">
      {["Saneamento", "Projetos de engenharia", "Supervisão ambiental", "Geotecnia", "BIM"].map((area) => <span key={area}>{area}</span>)}
    </div>
  );
}

createRoot(document.getElementById("root")).render(<App />);
