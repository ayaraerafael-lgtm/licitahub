import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import { buildTechnicalExperience, formatExperienceDays } from "./technicalExperience.js";
import "./styles.css";

const NavigationContext = React.createContext({
  canGoBack: false,
  goBack: () => {}
});

class ScreenErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  componentDidCatch(error, info) {
    console.error("Falha ao abrir tela", this.props.screen, error, info);
  }

  render() {
    if (!this.state.error) return this.props.children;
    return (
      <section className="page">
        <div className="pageHeader">
          <div>
            <span className="eyebrow">Navegação</span>
            <h2>Não foi possível abrir esta tela</h2>
            <p className="pageHelp">O restante do sistema continua disponível. Volte ao Radar e tente novamente.</p>
          </div>
        </div>
        <div className="card dangerNotice">
          <p>Ocorreu uma falha ao carregar esta funcionalidade.</p>
          <button type="button" className="btn primary" onClick={() => { window.location.hash = "radar-home"; }}>Voltar ao Radar</button>
        </div>
      </section>
    );
  }
}

const API_BASE_URL = window.location.protocol === "file:" ? "http://127.0.0.1:8080" : window.location.origin;

const newsCategoryOptions = [
  { label: "Licitações", slug: "licitacoes" },
  { label: "Mercado", slug: "mercado" },
  { label: "Legislação", slug: "legislacao" },
  { label: "Eventos", slug: "eventos" },
  { label: "Comunicados", slug: "comunicados" }
];

const newsStatusOptions = [
  { label: "Rascunho", value: "draft" },
  { label: "Disponível", value: "published" },
  { label: "Destaque principal", value: "featured" },
  { label: "Antiga / arquivada", value: "archived" },
  { label: "Expirada", value: "expired" }
];

const communityCategoryOptions = [
  { label: "Equipe comercial", slug: "equipe-comercial" },
  { label: "Notícias", slug: "noticias" },
  { label: "Atividades", slug: "atividades" },
  { label: "Eventos", slug: "eventos" },
  { label: "Conquistas", slug: "conquistas" },
  { label: "Conteúdo técnico", slug: "conteudo-tecnico" },
  { label: "Destaque", slug: "destaque" }
];

const brazilStates = [
  { uf: "AC", name: "Acre" },
  { uf: "AL", name: "Alagoas" },
  { uf: "AP", name: "Amapá" },
  { uf: "AM", name: "Amazonas" },
  { uf: "BA", name: "Bahia" },
  { uf: "CE", name: "Ceará" },
  { uf: "DF", name: "Distrito Federal" },
  { uf: "ES", name: "Espírito Santo" },
  { uf: "GO", name: "Goiás" },
  { uf: "MA", name: "Maranhão" },
  { uf: "MT", name: "Mato Grosso" },
  { uf: "MS", name: "Mato Grosso do Sul" },
  { uf: "MG", name: "Minas Gerais" },
  { uf: "PA", name: "Pará" },
  { uf: "PB", name: "Paraíba" },
  { uf: "PR", name: "Paraná" },
  { uf: "PE", name: "Pernambuco" },
  { uf: "PI", name: "Piauí" },
  { uf: "RJ", name: "Rio de Janeiro" },
  { uf: "RN", name: "Rio Grande do Norte" },
  { uf: "RS", name: "Rio Grande do Sul" },
  { uf: "RO", name: "Rondônia" },
  { uf: "RR", name: "Roraima" },
  { uf: "SC", name: "Santa Catarina" },
  { uf: "SP", name: "São Paulo" },
  { uf: "SE", name: "Sergipe" },
  { uf: "TO", name: "Tocantins" }
];

const majorCitiesByState = {
  AC: ["Rio Branco", "Cruzeiro do Sul"],
  AL: ["Maceió", "Arapiraca"],
  AP: ["Macapá", "Santana"],
  AM: ["Manaus", "Parintins"],
  BA: ["Salvador", "Feira de Santana", "Vitória da Conquista"],
  CE: ["Fortaleza", "Juazeiro do Norte", "Sobral"],
  DF: ["Brasília"],
  ES: ["Vitória", "Vila Velha", "Serra"],
  GO: ["Goiânia", "Aparecida de Goiânia", "Anápolis"],
  MA: ["São Luís", "Imperatriz"],
  MT: ["Cuiabá", "Várzea Grande", "Rondonópolis"],
  MS: ["Campo Grande", "Dourados"],
  MG: ["Belo Horizonte", "Uberlândia", "Contagem", "Juiz de Fora"],
  PA: ["Belém", "Ananindeua", "Santarém"],
  PB: ["João Pessoa", "Campina Grande"],
  PR: ["Curitiba", "Londrina", "Maringá"],
  PE: ["Recife", "Jaboatão dos Guararapes", "Petrolina"],
  PI: ["Teresina", "Parnaíba"],
  RJ: ["Rio de Janeiro", "Niterói", "Campos dos Goytacazes"],
  RN: ["Natal", "Mossoró"],
  RS: ["Porto Alegre", "Caxias do Sul", "Pelotas"],
  RO: ["Porto Velho", "Ji-Paraná"],
  RR: ["Boa Vista"],
  SC: ["Florianópolis", "Joinville", "Blumenau"],
  SP: ["São Paulo", "Campinas", "Santos", "São José dos Campos"],
  SE: ["Aracaju", "Nossa Senhora do Socorro"],
  TO: ["Palmas", "Araguaína"]
};

const tenderModalityOptions = [
  "Concorrência",
  "Pregão eletrônico",
  "Pregão presencial",
  "Tomada de preços",
  "Convite",
  "Concurso",
  "Leilão",
  "Diálogo competitivo",
  "Dispensa",
  "Inexigibilidade",
  "Credenciamento"
];

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
      { id: "my-profile", label: "Meu perfil", hidden: true },
      { id: "invite-new", label: "Novo convite" },
      { id: "invite-list", label: "Lista de convites" },
      { id: "company-manage", label: "Empresas cadastradas" },
	  { id: "password-reset-management", label: "Recuperação de senha" },
      { id: "company-rating-admin", label: "Avaliação das associadas" },
      { id: "invite-accept", label: "Aceite do convite", hidden: true },
      { id: "company-review", label: "Análise de empresas", hidden: true }
    ]
  },
  {
    id: "company",
    label: "Empresa",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "notification-history", label: "Histórico de alertas", hidden: true },
      { id: "my-profile", label: "Meu perfil", hidden: true },
      { id: "company-profile-edit", label: "Editar perfil", roles: ["companyAdmin"] },
      { id: "company-users", label: "Usuários vinculados", roles: ["companyAdmin"] },
      { id: "company-user-profile", label: "Cadastro de usuário", roles: ["companyAdmin"] },
      { id: "company-user-block", label: "Confirmar bloqueio", roles: ["companyAdmin"], hidden: true },
      { id: "company-user-unblock", label: "Confirmar desbloqueio", roles: ["companyAdmin"], hidden: true },
      { id: "company-user-delete", label: "Desativar vínculo", roles: ["companyAdmin"], hidden: true }
    ]
  },
  {
    id: "technical-capability",
    label: "Capacidade técnica",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "technical-professionals", label: "Profissionais técnicos" },
      { id: "technical-certificates", label: "Atestados técnicos" },
      { id: "technical-certificate-ai", label: "Análise de atestados por IA", hidden: true }
    ]
  },
  {
    id: "community",
    label: "Comunidade",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "community-home", label: "Comunidade" },
      { id: "company-rating", label: "Avaliação de parcerias" },
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
      { id: "pncp-capture", label: "Captação de editais", roles: ["platformAdmin"] },
      { id: "tender-new", label: "Cadastro de edital", roles: ["platformAdmin"] },
      { id: "tender-challenge-board", label: "Impugnações", roles: ["platformAdmin"] },
      { id: "tender-list", label: "Lista de editais", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "match-partners", label: "Vitrine de parceiros", roles: ["companyAdmin", "commercial"] },
      { id: "match-list", label: "Meus consórcios", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "tender-detail", label: "Detalhe do edital", roles: ["platformAdmin", "companyAdmin", "commercial", "technical", "reader"], hidden: true },
      { id: "tender-challenge", label: "Pedido de impugnação", roles: ["companyAdmin", "commercial", "technical"], hidden: true },
      { id: "tender-interest", label: "Interesse no edital", roles: ["companyAdmin", "commercial"], hidden: true },
      { id: "tender-interest-list", label: "Empresas interessadas", roles: ["companyAdmin", "commercial"], hidden: true }
    ]
  },
  {
    id: "workspace",
    label: "Área de trabalho",
    roles: ["companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "assembly-center", label: "Central de montagens" },
      { id: "my-assembly-tasks", label: "Minhas tarefas" },
      { id: "tender-calendar", label: "Calendário" },
      { id: "assembly-board", label: "Central de Montagem", hidden: true }
    ]
  },
  {
    id: "radar",
    label: "Radar LicitaHub",
    roles: ["platformAdmin", "companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "radar-home", label: "Notícias" },
      { id: "radar-detail", label: "Detalhe da notícia", hidden: true },
      { id: "radar-new", label: "Cadastrar notícia", roles: ["platformAdmin"] },
      { id: "radar-manage", label: "Gerenciar notícias", roles: ["platformAdmin"] }
    ]
  },
  {
    id: "academy",
    label: "Academia LicitaHub",
    roles: ["platformAdmin", "companyAdmin", "commercial", "technical", "reader"],
    items: [
      { id: "academy-home", label: "Catálogo de cursos", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "academy-my-learning", label: "Meus cursos", roles: ["companyAdmin", "commercial", "technical", "reader"] },
      { id: "academy-admin", label: "Gerenciar cursos", roles: ["platformAdmin"] },
      { id: "academy-course-admin", label: "Gestão do curso", roles: ["platformAdmin"], hidden: true },
      { id: "academy-course", label: "Curso", roles: ["companyAdmin", "commercial", "technical", "reader"], hidden: true }
    ]
  },
  {
    id: "match",
    label: "Match e consórcios",
    roles: ["companyAdmin", "commercial"],
    items: [
      { id: "match-tinder", label: "Avaliar candidata", hidden: true },
      { id: "match-profile", label: "Detalhe do anúncio", hidden: true },
      { id: "match-success", label: "Match realizado", hidden: true }
    ]
  }
];

function ModuleIcon({ moduleId }) {
  const props = { viewBox: "0 0 24 24", fill: "none", stroke: "currentColor", strokeWidth: "1.8", strokeLinecap: "round", strokeLinejoin: "round", "aria-hidden": true };
  if (moduleId === "access") return <svg {...props}><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" /><path d="m9 12 2 2 4-4" /></svg>;
  if (moduleId === "company") return <svg {...props}><path d="M3 21h18" /><path d="M5 21V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16" /><path d="M9 7h1M14 7h1M9 11h1M14 11h1M9 15h1M14 15h1" /></svg>;
  if (moduleId === "technical-capability") return <svg {...props}><path d="M2 18a1 1 0 0 0 1 1h18a1 1 0 0 0 1-1v-1a7 7 0 0 0-14 0v1" /><path d="M10 15v-3.5a1.5 1.5 0 0 1 3 0V15" /><path d="M12 4v3" /></svg>;
  if (moduleId === "community") return <svg {...props}><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" /></svg>;
  if (moduleId === "tenders") return <svg {...props}><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z" /><path d="M14 2v6h6M8 13h8M8 17h6" /></svg>;
  if (moduleId === "workspace") return <svg {...props}><circle cx="6.5" cy="5.5" r="2" /><path d="M3 20v-2a3.5 3.5 0 0 1 7 0v2" /><rect x="12" y="8" width="9" height="7" rx="1" /><path d="M16.5 15v3M13.5 20h6" /></svg>;
  if (moduleId === "radar") return <svg {...props}><path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9" /><path d="M7.8 16.2a6 6 0 0 1 0-8.4" /><circle cx="12" cy="12" r="2" /><path d="M16.2 7.8a6 6 0 0 1 0 8.4M19.1 4.9c3.9 3.9 3.9 10.3 0 14.2" /></svg>;
  if (moduleId === "academy") return <svg {...props}><path d="m3 10 9-5 9 5-9 5-9-5Z" /><path d="M7 12.2v4.2c2.8 2.1 7.2 2.1 10 0v-4.2" /><path d="M21 10v6" /></svg>;
  return <svg {...props}><path d="m7 7 5-5 5 5" /><path d="M12 2v15" /><path d="m17 17-5 5-5-5" /><path d="M2 12h20" /></svg>;
}

const stats = [
  ["Editais compatíveis", "9"],
  ["Interesses ativos", "3"],
  ["Matches abertos", "2"],
  ["Publicações", "14"]
];

const tenders = [
  {
    id: "cp-004-2026",
    agency: "Prefeitura Municipal",
    number: "CP 004/2026",
    modality: "Concorrência",
    object: "Projetos de saneamento e drenagem urbana",
    location: "MG",
    opening: "18/08/2026",
    value: "R$ 2.400.000,00",
    criterion: "Técnica e preço",
    status: "Publicado"
  },
  {
    id: "tp-012-2026",
    agency: "Departamento de Estradas",
    number: "TP 012/2026",
    modality: "Tomada de preços",
    object: "Supervisão de obras rodoviárias",
    location: "PR",
    opening: "02/09/2026",
    value: "R$ 5.800.000,00",
    criterion: "Menor preço",
    status: "Em avaliação"
  }
];

const partners = [
  {
    name: "GeoArq Projetos",
    location: "Belo Horizonte - MG",
    offers: "Arqueologia, estudos socioambientais e equipe de campo.",
    seeks: "Coordenação técnica em saneamento e proposta técnica."
  },
  {
    name: "SocialTec Consultoria",
    location: "Salvador - BA",
    offers: "Projetos sociais, comunicação comunitária e reassentamento.",
    seeks: "Empresa líder com experiência em infraestrutura urbana."
  }
];

function canSee(item, role) {
  return !item.roles || item.roles.includes(role);
}

function findScreenConfig(screenId) {
  for (const group of modules) {
    const item = group.items.find((entry) => entry.id === screenId);
    if (item) return { group, item };
  }
  return null;
}

function canAccessScreen(screenId, role) {
  if (screenId === "invite-accept" || screenId === "reset-password") return true;
  return modules.some((group) =>
    group.roles.includes(role) &&
    group.items.some((item) => item.id === screenId && canSee(item, role))
  );
}

function firstScreenFor(role) {
  const radar = modules.find((group) => group.id === "radar");
  const news = radar?.items.find((entry) => entry.id === "radar-home" && canSee(entry, role));
  if (radar?.roles.includes(role) && news) return news.id;
  const module = modules.find((group) => group.roles.includes(role));
  const item = module?.items.find((entry) => canSee(entry, role) && !entry.hidden);
  return item?.id || "radar-home";
}

function normalizeScreenFromHash(hash, role) {
  const raw = String(hash || "").replace("#", "");
  const [screenId] = raw.split("?");
  if (screenId === "aceite-convite") return "invite-accept";
  return screenId || firstScreenFor(role);
}

function currentHashParams() {
  const raw = String(window.location.hash || "").replace("#", "");
  const query = raw.includes("?") ? raw.slice(raw.indexOf("?") + 1) : "";
  return new URLSearchParams(query);
}

function frontendRole(roleKey) {
  return {
    platform_admin: "platformAdmin",
    company_admin: "companyAdmin",
    commercial: "commercial",
    technical: "technical",
    reader: "reader"
  }[roleKey] || "reader";
}

function useHashScreen(role) {
  const [screen, setScreen] = useState(() => normalizeScreenFromHash(window.location.hash, role));

  useEffect(() => {
    const onHashChange = () => setScreen(normalizeScreenFromHash(window.location.hash, role));
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, [role]);

  useEffect(() => {
    if (!canAccessScreen(screen, role)) {
      const next = firstScreenFor(role);
      window.location.hash = next;
      setScreen(next);
    }
  }, [role, screen]);

  return screen;
}

function App() {
  const [role, setRole] = useState("reader");
  const [sessionUser, setSessionUser] = useState(null);
  const [checkingSession, setCheckingSession] = useState(true);
  const [menuCollapsed, setMenuCollapsed] = useState(false);
  const [openModuleId, setOpenModuleId] = useState("radar");
  const sidebarRef = React.useRef(null);
  const [userStatuses, setUserStatuses] = useState({ marina: "Ativo", renato: "Ativo", paula: "Convite pendente" });
  const [selectedUserAction, setSelectedUserAction] = useState(null);
  const [selectedUserProfile, setSelectedUserProfile] = useState({ mode: "create", user: null });
  const [selectedPublicationId, setSelectedPublicationId] = useState(null);
  const [selectedTenderId, setSelectedTenderId] = useState("cp-004-2026");
  const [selectedNews, setSelectedNews] = useState(null);
  const [navigationStack, setNavigationStack] = useState([]);
  const [chatSeedAd, setChatSeedAd] = useState(null);
  const [chatSeedTask, setChatSeedTask] = useState(null);
  const [chatSeedUser, setChatSeedUser] = useState(null);
  const screen = useHashScreen(role);
  const visibleModules = useMemo(() => {
    const menuOrder = { radar: 10, community: 20, company: 30, tenders: 40, workspace: 50, "technical-capability": 60, academy: 65, access: 70, match: 80 };
    return modules.filter((group) => group.roles.includes(role) && group.items.some((item) => canSee(item, role) && !item.hidden)).sort((left, right) => (menuOrder[left.id] || 99) - (menuOrder[right.id] || 99));
  }, [role]);
  const isPublicInvitation = screen === "invite-accept";
  const isPasswordReset = screen === "reset-password";

  useEffect(() => {
    const activeModule = modules.find((group) => group.items.some((item) => item.id === screen));
    if (activeModule) setOpenModuleId(activeModule.id);
  }, [screen]);

  const refreshSession = async () => {
    const response = await fetch(`${API_BASE_URL}/api/auth/session`, { credentials: "include" });
    if (!response.ok) return null;
    const user = await response.json();
    setSessionUser(user);
    setRole(frontendRole(user.roleKey));
    return user;
  };

  useEffect(() => {
    if (isPublicInvitation || isPasswordReset) {
      setCheckingSession(false);
      return;
    }
    refreshSession()
      .finally(() => setCheckingSession(false));
  }, [isPublicInvitation, isPasswordReset]);

  useEffect(() => {
    const closeMenuOnOutsideClick = (event) => {
      if (menuCollapsed || sidebarRef.current?.contains(event.target)) return;
      setMenuCollapsed(true);
    };

    document.addEventListener("pointerdown", closeMenuOnOutsideClick);
    return () => document.removeEventListener("pointerdown", closeMenuOnOutsideClick);
  }, [menuCollapsed]);

  const navigateTo = (id) => {
    if (id === screen) return;
    setNavigationStack((current) => [...current, screen]);
    window.location.hash = id;
  };

  const goBack = () => {
    setNavigationStack((current) => {
      const previous = current[current.length - 1];
      if (previous) {
        window.location.hash = previous;
        return current.slice(0, -1);
      }
      window.location.hash = firstScreenFor(role);
      return [];
    });
  };

  const openUserProfile = (user = null, mode = "create") => {
    setSelectedUserProfile({ mode, user });
    navigateTo("company-user-profile");
  };

  const openUserAction = (id, name, action) => {
    setSelectedUserAction({ id, name, action });
    navigateTo(action === "block" ? "company-user-block" : action === "unblock" ? "company-user-unblock" : "company-user-delete");
  };

  const updateUserStatus = (id, status) => {
    setUserStatuses((current) => ({ ...current, [id]: status }));
    navigateTo("company-users");
  };

  const openPublicationManager = (publicationId) => {
    setSelectedPublicationId(publicationId);
    navigateTo("publication-list");
  };

  const openTenderInterestCompanies = (tenderId) => {
    setSelectedTenderId(tenderId);
    navigateTo(`tender-interest-list?id=${tenderId}`);
  };

  const openNewsDetail = (news) => {
    setSelectedNews(news);
    navigateTo("radar-detail");
  };

  const openChatForAd = (ad) => {
    setChatSeedAd(ad);
  };
  const openChatForTask = (task) => setChatSeedTask(task);
  const openChatForUser = (user) => setChatSeedUser(user);

  const handleLogin = (user) => {
    const nextRole = frontendRole(user.roleKey);
    setSessionUser(user);
    setRole(nextRole);
    window.location.hash = firstScreenFor(nextRole);
  };

  const handleLogout = async () => {
    await fetch(`${API_BASE_URL}/api/auth/logout`, { method: "POST", credentials: "include" }).catch(() => {});
    setSessionUser(null);
    setRole("reader");
    window.location.hash = "";
  };

  if (isPublicInvitation) return <main className="publicFlow"><InviteAccept /></main>;
  if (isPasswordReset) return <ResetPasswordScreen />;
  if (checkingSession) return <div className="authLoading">Carregando LicitaHub...</div>;
  if (!sessionUser) return <LoginScreen onLogin={handleLogin} />;

  return (
    <div className={`app ${menuCollapsed ? "menuCollapsed" : ""}`}>
      <aside className="sidebar" ref={sidebarRef}>
        <div className="brand">
          <span className="brandMark">LH</span>
          <div>
            <strong>LicitaHub</strong>
            <small>Rede consultiva</small>
          </div>
        </div>

        <nav>
          {visibleModules.map((group) => (
            <div className={`navGroup ${openModuleId === group.id ? "isOpen" : ""}`} key={group.id}>
              <button className="navGroupHeader" type="button" title={group.label} aria-label={group.label} onClick={() => { setOpenModuleId(group.id); if (menuCollapsed) setMenuCollapsed(false); }}>
                <ModuleIcon moduleId={group.id} />
                <span>{group.label}</span>
                <b aria-hidden="true">⌄</b>
              </button>
              <div className="navGroupLinks">
                {group.items.filter((item) => canSee(item, role) && !item.hidden).map((item) => (
                  <a className={screen === item.id ? "active" : ""} href={`#${item.id}`} key={item.id} onClick={(event) => { event.preventDefault(); navigateTo(item.id); }}>
                    {item.label}
                  </a>
                ))}
              </div>
            </div>
          ))}
        </nav>
      </aside>

      <main className="main">
        <NavigationContext.Provider value={{ canGoBack: navigationStack.length > 0, goBack }}>
          <Topbar navigate={navigateTo} openPublicationManager={openPublicationManager} openTenderInterestCompanies={openTenderInterestCompanies} sessionUser={sessionUser} onLogout={handleLogout} />
          <ScreenErrorBoundary key={screen} screen={screen}>
            <Screen screen={screen} navigate={navigateTo} userStatuses={userStatuses} openUserAction={openUserAction} selectedUserAction={selectedUserAction} updateUserStatus={updateUserStatus} selectedUserProfile={selectedUserProfile} openUserProfile={openUserProfile} selectedPublicationId={selectedPublicationId} openPublicationManager={openPublicationManager} selectedTenderId={selectedTenderId} openTenderInterestCompanies={openTenderInterestCompanies} selectedNews={selectedNews} openNewsDetail={openNewsDetail} refreshSession={refreshSession} sessionUser={sessionUser} openChatForAd={openChatForAd} openChatForTask={openChatForTask} openChatForUser={openChatForUser} />
          </ScreenErrorBoundary>
          <FloatingChat sessionUser={sessionUser} seedAd={chatSeedAd} seedTask={chatSeedTask} seedUser={chatSeedUser} onSeedConsumed={() => { setChatSeedAd(null); setChatSeedTask(null); setChatSeedUser(null); }} navigate={navigateTo} />
          <ScrollControls />
        </NavigationContext.Provider>
      </main>
    </div>
  );
}

function LoginScreen({ onLogin }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [forgotMode, setForgotMode] = useState(false);

  useEffect(() => {
    const notice = window.sessionStorage.getItem("licitahubLoginNotice") || "";
    if (notice) {
      setSuccessMessage(notice);
      window.sessionStorage.removeItem("licitahubLoginNotice");
    }
  }, []);

  const submit = async (event) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    setSuccessMessage("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível entrar.");
      onLogin(data.user);
    } catch (error) {
      setMessage(error.message);
    } finally {
      setSaving(false);
    }
  };

  const requestReset = async (event) => {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/forgot-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível gerar o link.");
      setMessage(data.message);
    } catch (error) {
      setMessage(error.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <main className="loginPage">
      <section className="loginPanel">
        <div className="loginBrand"><span className="brandMark">LH</span><div><strong>LicitaHub</strong><small>Rede consultiva</small></div></div>
        <div className="loginIntro"><span className="eyebrow">Acesso seguro</span><h1>{forgotMode ? "Recuperar senha" : "Entre na sua conta"}</h1><p>{forgotMode ? "Informe seu e-mail para gerar um link temporário." : "Use o e-mail e a senha cadastrados para acessar a plataforma."}</p></div>
        {successMessage && <div className="loginSuccess">{successMessage}</div>}
        {message && <div className="loginError">{message}</div>}
        <form onSubmit={forgotMode ? requestReset : submit}>
          <Field label="E-mail"><input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="username" required /></Field>
          {!forgotMode && <Field label="Senha"><input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" required /></Field>}
          <Button type="submit" disabled={saving}>{saving ? "Aguarde..." : forgotMode ? "Gerar link de recuperação" : "Entrar"}</Button>
        </form>
        <button type="button" className="loginTextButton" onClick={() => { setForgotMode((current) => !current); setMessage(""); }}>{forgotMode ? "Voltar para o login" : "Esqueci minha senha"}</button>
      </section>
      <section className="loginContext"><span>Engenharia consultiva</span><h2>Empresas, oportunidades e parcerias em um só ambiente.</h2><p>Acesso exclusivo para empresas convidadas e aprovadas pela LicitaHub.</p></section>
    </main>
  );
}

function ResetPasswordScreen() {
  const token = currentHashParams().get("token") || "";
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);

  const submit = async (event) => {
    event.preventDefault();
    if (password.length < 8) {
      setMessage({ type: "error", text: "A senha deve ter pelo menos 8 caracteres." });
      return;
    }
    if (password !== confirmPassword) {
      setMessage({ type: "error", text: "As senhas não conferem." });
      return;
    }
    setSaving(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/reset-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, password })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível redefinir a senha.");
      if (data.accountActive) {
        setMessage({ type: "success", text: "Senha alterada. Você já pode voltar ao login." });
      } else if (data.accountStatus === "blocked") {
        setMessage({ type: "error", text: "A senha foi alterada, mas este usuário está bloqueado. O administrador da empresa precisa desbloquear o acesso." });
      } else {
        setMessage({ type: "error", text: "A senha foi alterada, mas este usuário está inativo. O administrador da empresa precisa reativar o acesso." });
      }
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving(false);
    }
  };

  return <main className="loginPage"><section className="loginPanel"><div className="loginBrand"><span className="brandMark">LH</span><div><strong>LicitaHub</strong><small>Rede consultiva</small></div></div><div className="loginIntro"><span className="eyebrow">Acesso</span><h1>Defina sua senha</h1><p>Use este link para criar ou trocar sua senha. Ele só pode ser usado uma vez.</p></div>{message && <div className={message.type === "success" ? "loginSuccess" : "loginError"}>{message.text}</div>}<form onSubmit={submit}><Field label="Nova senha"><input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required /></Field><Field label="Confirmar nova senha"><input type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} required /></Field><Button type="submit" disabled={saving || !token}>{saving ? "Salvando..." : "Salvar senha"}</Button></form><a className="loginTextButton" href={window.location.pathname}>Voltar para o login</a></section><section className="loginContext"><span>Segurança</span><h2>Sua senha libera o acesso pessoal à LicitaHub.</h2></section></main>;
}

function Topbar({ navigate, openPublicationManager, openTenderInterestCompanies, sessionUser, onLogout }) {
  const [alertsOpen, setAlertsOpen] = useState(false);
  const [alerts, setAlerts] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = React.useRef(null);
  const userInitials = String(sessionUser?.fullName || "Usuário").split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase();

  const openAlert = (alert) => {
    const destination = alert.destinationScreen || "";
    const relatedId = alert.relatedEntityId || "";
    if (destination === "publication-list" && relatedId) {
      openPublicationManager(relatedId);
    } else if (destination === "tender-interest-list" && relatedId) {
      openTenderInterestCompanies(relatedId);
    } else if (destination === "tender-detail" && relatedId) {
      navigate(`tender-detail?id=${relatedId}`);
    } else if (destination === "match-profile" && relatedId) {
      navigate(`match-profile?id=${relatedId}`);
    } else if (destination === "match-success" && relatedId) {
      navigate(`match-success?id=${relatedId}`);
    } else if (destination === "company-dashboard") {
      navigate("radar-home");
    } else if (destination === "match-list" || destination === "company-review" || destination === "tender-list" || destination === "radar-home") {
      navigate(destination);
    } else {
      navigate(destination || "community-home");
    }
    setAlertsOpen(false);
  };

  const loadAlerts = () => {
    fetch(`${API_BASE_URL}/api/notifications`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => []);
        if (!response.ok) throw new Error(data.error || "Falha ao carregar notificações.");
        return data;
      })
      .then((data) => {
        const items = Array.isArray(data) ? data : [];
        setAlerts(items);
        setUnreadCount(items.length);
      })
      .catch(() => {
        setAlerts([]);
        setUnreadCount(0);
      });
  };

  useEffect(() => {
    if (!sessionUser?.id) return;
    loadAlerts();
    const timer = window.setInterval(loadAlerts, 45000);
    return () => window.clearInterval(timer);
  }, [sessionUser?.id]);

  const toggleAlerts = () => {
    setAlertsOpen((open) => !open);
    if (!alertsOpen && unreadCount > 0) {
      setUnreadCount(0);
      fetch(`${API_BASE_URL}/api/notifications/read-all`, { method: "PATCH", credentials: "include" }).catch(() => {});
    }
  };

  useEffect(() => {
    const closeUserMenu = (event) => {
      if (!userMenuOpen || userMenuRef.current?.contains(event.target)) return;
      setUserMenuOpen(false);
    };

    document.addEventListener("pointerdown", closeUserMenu);
    return () => document.removeEventListener("pointerdown", closeUserMenu);
  }, [userMenuOpen]);

  return (
    <header className="topbar">
      <div className="companyTopIdentity">
        <LogoSlot src={sessionUser?.companyLogoUrl} initials={sessionUser?.companyName === "LicitaHub" ? "LH" : String(sessionUser?.companyName || "Empresa").slice(0, 2).toUpperCase()} size="sm" label={`Logo de ${sessionUser?.companyName || "LicitaHub"}`} />
        <div>
          <span className="eyebrow">{sessionUser?.companyName || "LicitaHub"}</span>
          <h1>LicitaHub</h1>
        </div>
      </div>
      <div className="alertCenter">
        <button className={`alertBell ${unreadCount > 0 ? "hasUnread" : ""}`} type="button" title="Ver alertas importantes" aria-label={unreadCount > 0 ? `${unreadCount} alertas novos` : "Ver alertas importantes"} onClick={toggleAlerts}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9" /><path d="M13.73 21a2 2 0 0 1-3.46 0" /></svg>
          {unreadCount > 0 && <strong>{unreadCount}</strong>}
        </button>
        {alertsOpen && (
          <div className="alertDropdown">
            <div className="alertDropdownHeader">
              <strong>Alertas importantes</strong>
              <div><span>{unreadCount} novos</span><button type="button" className="alertHistoryButton" onClick={() => { setAlertsOpen(false); navigate("notification-history"); }}>Histórico</button></div>
            </div>
            <div className="alertDropdownList">
              {alerts.length === 0 && <div className="alertEmpty">Nenhum alerta novo.</div>}
              {alerts.map((alert) => (
                <button type="button" className="alertItem" key={alert.id} onClick={() => openAlert(alert)}>
                  <strong>{alert.title}</strong>
                  <span>{alert.message}</span>
                </button>
              ))}
            </div>
          </div>
        )}
        <div className="userIdentityMenu" ref={userMenuRef}>
          <button type="button" className="userIdentityButton" aria-label="Ver usuário conectado" onClick={() => setUserMenuOpen((open) => !open)}>
            <span className="userAvatar">{sessionUser?.profilePhotoUrl ? <img src={sessionUser.profilePhotoUrl} alt="" /> : userInitials}</span>
            <span className="userIdentityText"><strong>{sessionUser?.fullName}</strong><small>{sessionUser?.roleName}</small></span>
            <span aria-hidden="true">{"\u2304"}</span>
          </button>
          {userMenuOpen && <div className="userDropdown"><div className="userDropdownHeader"><span className="userAvatar large">{sessionUser?.profilePhotoUrl ? <img src={sessionUser.profilePhotoUrl} alt="" /> : userInitials}</span><div><strong>{sessionUser?.fullName}</strong><span>{sessionUser?.email}</span></div></div><dl><div><dt>Empresa</dt><dd>{sessionUser?.companyName}</dd></div><div><dt>Perfil</dt><dd>{sessionUser?.roleName}</dd></div></dl><button type="button" className="logoutButton" onClick={() => { setUserMenuOpen(false); navigate("my-profile"); }}>Meu perfil</button><button type="button" className="logoutButton" onClick={onLogout}>Sair da conta</button></div>}
        </div>
      </div>
    </header>
  );
}

function ScrollControls() {
  const { canGoBack, goBack } = React.useContext(NavigationContext);
  const scrollToTop = () => window.scrollTo({ top: 0, behavior: "smooth" });
  const scrollToBottom = () => window.scrollTo({ top: document.documentElement.scrollHeight, behavior: "smooth" });
  return (
    <div className="scrollControls" aria-label="Navegação rápida da página">
      <button type="button" title="Subir para o topo" aria-label="Subir para o topo" onClick={scrollToTop}>{"\u2191"}</button>
      <button type="button" title="Voltar para tela anterior" aria-label="Voltar para tela anterior" disabled={!canGoBack} onClick={goBack}>{"\u2190"}</button>
      <button type="button" title="Descer para o final" aria-label="Descer para o final" onClick={scrollToBottom}>{"\u2193"}</button>
    </div>
  );
}

function NotificationHistory({ navigate, openPublicationManager, openTenderInterestCompanies }) {
  const [filters, setFilters] = useState({ search: "", type: "", status: "", period: "30" });
  const [page, setPage] = useState(1);
  const [data, setData] = useState({ items: [], total: 0, page: 1, pageSize: 20 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const typeLabels = { post_like: "Curtida", post_comment: "Comentário", company_post: "Publicação da empresa", match: "Match e consórcio", company_interested: "Interesse em edital", news: "Notícia", system: "Sistema" };

  useEffect(() => {
    setLoading(true);
    setError("");
    const query = new URLSearchParams({ page: String(page), period: filters.period });
    if (filters.search.trim()) query.set("search", filters.search.trim());
    if (filters.type) query.set("type", filters.type);
    if (filters.status) query.set("status", filters.status);
    fetch(`${API_BASE_URL}/api/notification-history?${query.toString()}`, { credentials: "include" })
      .then(async (response) => {
        const result = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(result.error || "Não foi possível carregar o histórico de alertas.");
        return result;
      })
      .then((result) => setData({ items: Array.isArray(result.items) ? result.items : [], total: Number(result.total || 0), page: Number(result.page || page), pageSize: Number(result.pageSize || 20) }))
      .catch((loadError) => setError(loadError.message))
      .finally(() => setLoading(false));
  }, [filters, page]);

  const updateFilter = (key, value) => {
    setPage(1);
    setFilters((current) => ({ ...current, [key]: value }));
  };
  const openNotification = (notice) => {
    const destination = notice.destinationScreen || "";
    const relatedId = notice.relatedEntityId || "";
    if (destination === "publication-list" && relatedId) openPublicationManager(relatedId);
    else if (destination === "tender-interest-list" && relatedId) openTenderInterestCompanies(relatedId);
    else if (destination === "tender-detail" && relatedId) navigate(`tender-detail?id=${encodeURIComponent(relatedId)}`);
    else if (destination === "match-profile" && relatedId) navigate(`match-profile?id=${encodeURIComponent(relatedId)}`);
    else if (destination === "match-success" && relatedId) navigate(`match-success?id=${encodeURIComponent(relatedId)}`);
    else navigate(destination === "company-dashboard" ? "radar-home" : destination || "radar-home");
  };
  const totalPages = Math.max(1, Math.ceil(data.total / data.pageSize));

  return <Page label="Acompanhamento" title="Histórico de alertas">
    <section className="notificationHistoryHero"><div><span className="eyebrow">Arquivo de acontecimentos</span><h3>Alertas recebidos por você</h3><p>Consulte movimentações anteriores e retorne diretamente ao conteúdo, edital, consórcio ou publicação de origem.</p></div><div className="notificationHistoryTotal"><strong>{data.total}</strong><span>registro{data.total === 1 ? "" : "s"}</span></div></section>
    <Card className="compactFilters notificationHistoryFilters"><FormGrid>
      <Field label="Buscar"><input value={filters.search} onChange={(event) => updateFilter("search", event.target.value)} placeholder="Título ou mensagem do alerta" /></Field>
      <Field label="Tipo"><select value={filters.type} onChange={(event) => updateFilter("type", event.target.value)}><option value="">Todos</option>{Object.entries(typeLabels).map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
      <Field label="Situação"><select value={filters.status} onChange={(event) => updateFilter("status", event.target.value)}><option value="">Todos</option><option value="unread">Não lidos</option><option value="read">Já lidos</option></select></Field>
      <Field label="Período"><select value={filters.period} onChange={(event) => updateFilter("period", event.target.value)}><option value="7">Últimos 7 dias</option><option value="30">Últimos 30 dias</option><option value="90">Últimos 90 dias</option><option value="365">Último ano</option><option value="">Todo o histórico</option></select></Field>
    </FormGrid></Card>
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {loading ? <Card><p>Carregando o histórico de alertas...</p></Card> : data.items.length === 0 ? <Card className="notificationHistoryEmpty"><p>Nenhum alerta encontrado para os filtros informados.</p></Card> : <section className="notificationHistoryList">{data.items.map((notice) => <button type="button" className={`notificationHistoryItem ${notice.isRead ? "isRead" : "isUnread"}`} key={notice.id} onClick={() => openNotification(notice)}><span className="notificationHistoryType">{typeLabels[notice.type] || "Atualização"}</span><div><strong>{notice.title}</strong><p>{notice.message || "Há uma atualização para consultar."}</p><small>{notice.createdAt ? new Date(notice.createdAt).toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" }) : ""}</small></div><em>{notice.isRead ? "Lido" : "Novo"}</em></button>)}</section>}
    {!loading && data.total > data.pageSize && <div className="notificationPagination"><Button variant="secondary" onClick={() => setPage((current) => Math.max(1, current - 1))} disabled={page <= 1}>Anterior</Button><span>Página {page} de {totalPages}</span><Button variant="secondary" onClick={() => setPage((current) => Math.min(totalPages, current + 1))} disabled={page >= totalPages}>Próxima</Button></div>}
  </Page>;
}

function FloatingChat({ sessionUser, seedAd, seedTask, seedUser, onSeedConsumed, navigate }) {
  const canUseChat = sessionUser?.companyId && sessionUser?.roleKey !== "reader";
  const canUsePartnershipChat = ["company_admin", "commercial"].includes(sessionUser?.roleKey);
  const [open, setOpen] = useState(false);
  const [minimized, setMinimized] = useState(true);
  const [threads, setThreads] = useState([]);
  const [activeThreadId, setActiveThreadId] = useState("");
  const [messages, setMessages] = useState([]);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [threadFilters, setThreadFilters] = useState({ search: "", contextType: "all", date: "" });
  const [messageFilters, setMessageFilters] = useState({ search: "", sender: "all", date: "" });
  const [soundEnabled, setSoundEnabled] = useState(() => window.localStorage.getItem("licitahubChatSound") !== "off");
  const activeThreadRef = React.useRef("");
  const chatMessagesRef = React.useRef(null);
  const lastSoundAtRef = React.useRef(0);
  const audioContextRef = React.useRef(null);
  const audioUnlockedRef = React.useRef(false);
  const previousUnreadTotalRef = React.useRef(0);
  const unreadInitializedRef = React.useRef(false);

  useEffect(() => {
    activeThreadRef.current = activeThreadId;
  }, [activeThreadId]);

  const unreadTotal = threads.reduce((total, thread) => total + Number(thread.unreadCount || 0), 0);
  const activeThread = threads.find((thread) => thread.id === activeThreadId);
  const canReply = !activeThread || activeThread.canReply !== false;
  const latestMessageId = messages[messages.length - 1]?.id || "";
  const dateKey = (value) => {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "";
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  };
  const normalized = (value) => String(value || "").toLocaleLowerCase("pt-BR");
  const contextLabel = (thread) => thread?.contextType === "assembly_task" ? "Tarefa de montagem" : thread?.contextType === "direct_user" ? "Conversa direta" : "Parceria de licitação";
  const filteredThreads = threads.filter((thread) => {
    const search = normalized(threadFilters.search);
    const searchable = normalized([thread.otherCompanyName, thread.contextTitle, thread.tenderNumber, thread.agency, thread.tenderObject, thread.lastMessage].join(" "));
    return (!search || searchable.includes(search))
      && (threadFilters.contextType === "all" || thread.contextType === threadFilters.contextType)
      && (!threadFilters.date || dateKey(thread.lastActivityAt) === threadFilters.date);
  });
  const messageSenders = Array.from(new Map(messages.map((message) => [message.mine ? "me" : message.senderUserId || message.senderName, message.mine ? "Você" : message.senderName || "Usuário"])).entries());
  const filteredMessages = messages.filter((message) => {
    const search = normalized(messageFilters.search);
    const searchable = normalized([message.content, message.senderName, message.senderJobTitle].join(" "));
    const sender = message.mine ? "me" : message.senderUserId || message.senderName;
    return (!search || searchable.includes(search))
      && (messageFilters.sender === "all" || messageFilters.sender === sender)
      && (!messageFilters.date || dateKey(message.createdAt) === messageFilters.date);
  });
  const clearChatFilters = () => {
    setThreadFilters({ search: "", contextType: "all", date: "" });
    setMessageFilters({ search: "", sender: "all", date: "" });
  };

  useEffect(() => {
    const container = chatMessagesRef.current;
    if (!container || !activeThreadId) return;
    container.scrollTo({ top: container.scrollHeight, behavior: "smooth" });
  }, [activeThreadId, latestMessageId]);

  const unlockChatAudio = () => {
    if (audioUnlockedRef.current) return;
    const now = Date.now();
    try {
      const AudioContext = window.AudioContext || window.webkitAudioContext;
      if (!AudioContext) return;
      const context = audioContextRef.current || new AudioContext();
      audioContextRef.current = context;
      if (context.state === "suspended") {
        context.resume?.();
      }
      const oscillator = context.createOscillator();
      const gain = context.createGain();
      gain.gain.setValueAtTime(0.0001, context.currentTime);
      oscillator.connect(gain);
      gain.connect(context.destination);
      oscillator.start();
      oscillator.stop(context.currentTime + 0.01);
      audioUnlockedRef.current = true;
    } catch (_err) {
      audioUnlockedRef.current = false;
    }
  };

  const playMessageSound = (force = false) => {
    if (!soundEnabled && !force) return;
    const now = Date.now();
    if (now - lastSoundAtRef.current < 1200) return;
    lastSoundAtRef.current = now;
    try {
      const AudioContext = window.AudioContext || window.webkitAudioContext;
      if (!AudioContext) return;
      const context = audioContextRef.current || new AudioContext();
      audioContextRef.current = context;
      if (context.state === "suspended") {
        context.resume?.();
      }
      const oscillator = context.createOscillator();
      const gain = context.createGain();
      oscillator.type = "triangle";
      oscillator.frequency.setValueAtTime(880, context.currentTime);
      oscillator.frequency.exponentialRampToValueAtTime(620, context.currentTime + 0.22);
      gain.gain.setValueAtTime(0.0001, context.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.18, context.currentTime + 0.025);
      gain.gain.exponentialRampToValueAtTime(0.0001, context.currentTime + 0.25);
      oscillator.connect(gain);
      gain.connect(context.destination);
      oscillator.start();
      oscillator.stop(context.currentTime + 0.27);
    } catch (_err) {
      // Browsers may block audio until the first user interaction.
    }
  };

  const toggleSound = () => {
    unlockChatAudio();
    setSoundEnabled((current) => {
      const next = !current;
      window.localStorage.setItem("licitahubChatSound", next ? "on" : "off");
      if (next) {
        window.setTimeout(() => playMessageSound(true), 60);
      }
      return next;
    });
  };

  const loadThreads = async () => {
    if (!canUseChat) return;
    const response = await fetch(`${API_BASE_URL}/api/chats`, { credentials: "include" });
    const data = await response.json().catch(() => []);
    if (!response.ok) throw new Error(data.error || "Não foi possível carregar conversas.");
    const nextThreads = Array.isArray(data) ? data : [];
    const nextUnreadTotal = nextThreads.reduce((total, thread) => total + Number(thread.unreadCount || 0), 0);
    if (unreadInitializedRef.current && nextUnreadTotal > previousUnreadTotalRef.current) {
      playMessageSound();
    }
    unreadInitializedRef.current = true;
    previousUnreadTotalRef.current = nextUnreadTotal;
    setThreads(nextThreads);
  };

  const loadMessages = async (threadId) => {
    if (!threadId) return;
    const response = await fetch(`${API_BASE_URL}/api/chats/${encodeURIComponent(threadId)}/messages`, { credentials: "include" });
    const data = await response.json().catch(() => []);
    if (!response.ok) throw new Error(data.error || "Não foi possível carregar mensagens.");
    setMessages(Array.isArray(data) ? data : []);
    setThreads((current) => current.map((thread) => thread.id === threadId ? { ...thread, unreadCount: 0 } : thread));
  };

  const openThread = async (threadId) => {
    setActiveThreadId(threadId);
    setError("");
    setLoading(true);
    try {
      await loadMessages(threadId);
      await loadThreads();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const startChat = async (ad) => {
    if (!ad?.id) return;
    setOpen(true);
    setMinimized(false);
    setError("");
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/chats`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ adId: ad.id })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível abrir a conversa.");
      setActiveThreadId(data.id);
      await loadThreads();
      await loadMessages(data.id);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
      onSeedConsumed?.();
    }
  };

  const startTaskChat = async (task) => {
    if (!task?.taskId || !task?.assemblyId) return;
    setOpen(true); setMinimized(false); setError(""); setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/task-chats`, { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify(task) });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível abrir a conversa da tarefa.");
      setActiveThreadId(data.id); await loadThreads(); await loadMessages(data.id);
    } catch (err) { setError(err.message); } finally { setLoading(false); onSeedConsumed?.(); }
  };

  const startDirectChat = async (user) => {
    if (!user?.id) return;
    setOpen(true); setMinimized(false); setError(""); setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/direct-chats`, { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ userId: user.id }) });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível abrir a conversa direta.");
      setActiveThreadId(data.id); await loadThreads(); await loadMessages(data.id);
    } catch (err) { setError(err.message); } finally { setLoading(false); onSeedConsumed?.(); }
  };

  useEffect(() => {
    if (canUseChat) {
      loadThreads().catch(() => {});
    }
  }, [canUseChat]);

  useEffect(() => {
    const unlock = () => unlockChatAudio();
    window.addEventListener("pointerdown", unlock, { once: true });
    window.addEventListener("keydown", unlock, { once: true });
    return () => {
      window.removeEventListener("pointerdown", unlock);
      window.removeEventListener("keydown", unlock);
    };
  }, []);

  useEffect(() => {
    if (canUseChat && seedAd?.id) {
      startChat(seedAd);
    }
  }, [canUseChat, seedAd?.id]);

  useEffect(() => {
    if (canUseChat && seedTask?.taskId) startTaskChat(seedTask);
  }, [canUseChat, seedTask?.taskId]);

  useEffect(() => {
    if (canUseChat && seedUser?.id) startDirectChat(seedUser);
  }, [canUseChat, seedUser?.id]);

  useEffect(() => {
    if (!canUseChat) return undefined;
    const refreshFallback = () => {
      if (document.visibilityState !== "visible") return;
      loadThreads().catch(() => {});
      if (open && !minimized && activeThreadRef.current) {
        loadMessages(activeThreadRef.current).catch(() => {});
      }
    };
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible") refreshFallback();
    };
    const interval = window.setInterval(refreshFallback, 45000);
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [canUseChat, open, minimized]);

  useEffect(() => {
    if (!canUseChat) return undefined;
    if (!canUsePartnershipChat) return undefined;
    const source = new EventSource(`${API_BASE_URL}/api/chats/stream`, { withCredentials: true });
    source.addEventListener("chat-message", (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.eventType === "chat-message") {
          if (data.senderUserId !== sessionUser.userId) {
            playMessageSound();
          }
          if (data.threadId === activeThreadRef.current) {
            setMessages((current) => current.some((message) => message.id === data.id) ? current : [...current, { ...data, mine: data.senderUserId === sessionUser.userId }]);
            loadMessages(data.threadId).catch(() => {});
          }
          loadThreads().catch(() => {});
        }
        if (data.eventType === "chat-thread") {
          loadThreads().catch(() => {});
        }
      } catch (_err) {
        loadThreads().catch(() => {});
      }
    });
    return () => source.close();
  }, [canUseChat, canUsePartnershipChat, sessionUser?.userId, soundEnabled]);

  const sendMessage = async (event) => {
    event.preventDefault();
    if (!activeThreadId || !draft.trim()) return;
    const content = draft.trim();
    setDraft("");
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/chats/${encodeURIComponent(activeThreadId)}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível enviar mensagem.");
      setMessages((current) => current.some((message) => message.id === data.id) ? current : [...current, { ...data, mine: true }]);
      await loadThreads();
    } catch (err) {
      setDraft(content);
      setError(err.message);
    }
  };

  if (!canUseChat) return null;

  if (!open || minimized) {
    return (
      <button type="button" className={`chatLauncher ${unreadTotal > 0 ? "hasUnread" : "isQuiet"}`} onClick={() => { unlockChatAudio(); setOpen(true); setMinimized(false); loadThreads().catch(() => {}); }} title="Abrir conversas de parceria">
        {unreadTotal > 0 ? <span>Chat</span> : <span className="chatLauncherSymbol" aria-hidden="true">{"\u2709"}</span>}
        {unreadTotal > 0 && <strong>{unreadTotal}</strong>}
      </button>
    );
  }

  return (
    <section className="floatingChat" aria-label="Conversas de parceria" onPointerDown={unlockChatAudio}>
      <header className="floatingChatHeader">
        <div>
          <strong>{activeThread ? (activeThread.contextType === "assembly_task" ? activeThread.contextTitle : activeThread.otherCompanyName) : "Conversas"}</strong>
          <small>{activeThread ? (activeThread.contextType === "direct_user" ? "Conversa direta" : `${activeThread.contextType === "assembly_task" ? "Tarefa" : activeThread.tenderNumber} | ${activeThread.agency}`) : "Anúncios, tarefas e contatos"}</small>
        </div>
        <div className="floatingChatHeaderActions">
          {activeThread?.contextType === "partnership_ad" && activeThread?.evaluationAdId && activeThread.status === "open" && <button type="button" title="Avaliar candidata" onClick={() => navigate(`match-tinder?id=${activeThread.evaluationAdId}`)}>{"\u2713"}</button>}
          <button type="button" title={soundEnabled ? "Desligar som" : "Ligar som"} onClick={toggleSound}>{soundEnabled ? "\u266B" : "\u266A"}</button>
          <button type="button" title={filtersOpen ? "Recolher filtros" : "Filtrar conversas e mensagens"} aria-label={filtersOpen ? "Recolher filtros" : "Filtrar conversas e mensagens"} onClick={() => setFiltersOpen((current) => !current)}>{"\u2630"}</button>
          {activeThreadId && <button type="button" title="Voltar para lista" onClick={() => { setActiveThreadId(""); setMessages([]); }}>{"\u2190"}</button>}
          <button type="button" title="Minimizar chat" onClick={() => setMinimized(true)}>{"\u2212"}</button>
        </div>
      </header>
      {error && <p className="chatError">{error}</p>}
      {filtersOpen && (
        <div className="chatFilterPanel">
          {!activeThreadId ? <>
            <label><span>Pesquisar</span><input value={threadFilters.search} onChange={(event) => setThreadFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Pessoa, empresa, tema ou edital" /></label>
            <label><span>Tipo</span><select value={threadFilters.contextType} onChange={(event) => setThreadFilters((current) => ({ ...current, contextType: event.target.value }))}><option value="all">Todas as conversas</option><option value="partnership_ad">Parcerias</option><option value="assembly_task">Tarefas</option><option value="direct_user">Diretas</option></select></label>
            <label><span>Data</span><input type="date" value={threadFilters.date} onChange={(event) => setThreadFilters((current) => ({ ...current, date: event.target.value }))} /></label>
          </> : <>
            <label><span>Pesquisar mensagens</span><input value={messageFilters.search} onChange={(event) => setMessageFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Texto ou assunto da conversa" /></label>
            <label><span>Enviada por</span><select value={messageFilters.sender} onChange={(event) => setMessageFilters((current) => ({ ...current, sender: event.target.value }))}><option value="all">Todos os participantes</option>{messageSenders.map(([value, name]) => <option value={value} key={value}>{name}</option>)}</select></label>
            <label><span>Data</span><input type="date" value={messageFilters.date} onChange={(event) => setMessageFilters((current) => ({ ...current, date: event.target.value }))} /></label>
          </>}
          <button type="button" className="chatFilterClear" onClick={clearChatFilters}>Limpar filtros</button>
        </div>
      )}
      {!activeThreadId ? (
        <div className="chatThreadList">
          {threads.length === 0 && <p className="emptyChat">Nenhuma conversa iniciada ainda.</p>}
          {threads.length > 0 && filteredThreads.length === 0 && <p className="emptyChat">Nenhuma conversa encontrada com estes filtros.</p>}
          {filteredThreads.map((thread) => (
            <button type="button" className="chatThreadItem" key={thread.id} onClick={() => openThread(thread.id)}>
              <LogoSlot initials={thread.otherCompanyName?.split(" ").map((word) => word[0]).join("").slice(0, 2) || "CH"} src={thread.otherCompanyLogoUrl} size="xs" label={`Logo da ${thread.otherCompanyName}`} />
              <span>
                <strong>{thread.contextType === "assembly_task" ? thread.contextTitle : thread.otherCompanyName}</strong>
                <small>{thread.contextType === "direct_user" ? `Conversa direta | ${thread.lastMessage || "Sem mensagens"}` : thread.contextType === "assembly_task" ? `Tarefa | ${thread.lastMessage || thread.tenderObject}` : `${thread.tenderNumber} | ${thread.lastMessage || thread.tenderObject}`}</small>
                <small className="chatThreadMeta">{contextLabel(thread)} · {thread.lastActivityAt ? new Date(thread.lastActivityAt).toLocaleDateString("pt-BR") : "Sem data"}</small>
                {thread.isClosed && <small className="closedChatLabel">Encerrada</small>}
              </span>
              {thread.unreadCount > 0 && <em>{thread.unreadCount}</em>}
            </button>
          ))}
        </div>
      ) : (
        <>
          {activeThread?.closedReason && <div className={`chatNotice ${activeThread.isClosed ? "closed" : ""}`}>{activeThread.closedReason}</div>}
          <div className="chatMessages" ref={chatMessagesRef}>
            {loading && <p className="emptyChat">Carregando conversa...</p>}
            {!loading && messages.length === 0 && <p className="emptyChat">Conversa aberta. Envie a primeira mensagem.</p>}
            {!loading && messages.length > 0 && filteredMessages.length === 0 && <p className="emptyChat">Nenhuma mensagem encontrada com estes filtros.</p>}
            {filteredMessages.map((message) => (
              <div className={`chatBubble ${message.mine ? "mine" : ""}`} key={message.id}>
                <div className="chatSender">
                  {!message.mine && <LogoSlot initials={message.senderName?.split(" ").map((word) => word[0]).join("").slice(0, 2) || "US"} src={message.senderPhotoUrl} size="xs" fit="cover" label={`Foto de ${message.senderName}`} />}
                  <small>{message.mine ? "Você" : `${message.senderName}${message.senderJobTitle ? ` | ${message.senderJobTitle}` : ""}`}</small>
                </div>
                <p>{message.content}</p>
                <time>{message.createdAt ? new Date(message.createdAt).toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", hour: "2-digit", minute: "2-digit" }) : ""}</time>
              </div>
            ))}
          </div>
          <form className="chatComposer" onSubmit={sendMessage}>
            <input value={draft} onChange={(event) => setDraft(event.target.value)} placeholder={canReply ? "Escreva uma mensagem" : "Conversa encerrada"} maxLength={2000} disabled={!canReply} />
            <button type="submit" disabled={!draft.trim() || !canReply}>Enviar</button>
          </form>
        </>
      )}
    </section>
  );
}

function academyRequest(path, options = {}) {
  return fetch(`${API_BASE_URL}${path}`, { credentials: "include", ...options, headers: { "Content-Type": "application/json", ...(options.headers || {}) } })
    .then(async (response) => {
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível concluir esta ação.");
      return data;
    });
}

function academyCertificateUrl(courseId) {
  return `${API_BASE_URL}/api/academy/courses/${encodeURIComponent(courseId)}/certificate`;
}

function academyCertificateVerificationUrl(code) {
  return `${API_BASE_URL}/certificates/verify?code=${encodeURIComponent(code)}`;
}

function AcademyHome({ navigate }) {
  const [courses, setCourses] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const load = () => academyRequest("/api/academy/courses").then((data) => setCourses(Array.isArray(data) ? data : [])).catch((err) => setError(err.message)).finally(() => setLoading(false));
  useEffect(() => { load(); }, []);
  const availableCourses = courses.filter((course) => !course.enrolled);
  return <Page label="Academia LicitaHub" title="Academia LicitaHub" actions={<Button variant="secondary" onClick={() => navigate("academy-my-learning")}>Meus cursos</Button>}>
    <Card className="academyIntro"><div><span className="eyebrow">Desenvolvimento da equipe</span><h3>Aprendizado organizado para a prática da engenharia consultiva</h3><p>Assista às aulas, retome de onde parou, avance pelas avaliações e emita seu certificado ao finalizar.</p></div></Card>
    {loading && <Card><p>Carregando cursos disponíveis...</p></Card>}
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {!loading && !error && availableCourses.length === 0 && <Card><h3>Nenhum novo curso disponível</h3><p>{courses.length ? "Todos os cursos disponíveis já foram iniciados e podem ser encontrados em Meus cursos." : "Ainda não há cursos publicados na Academia LicitaHub."}</p>{courses.length > 0 && <Button onClick={() => navigate("academy-my-learning")}>Abrir meus cursos</Button>}</Card>}
    <div className="academyCourseGrid">{availableCourses.map((course) => <article className="academyCourseCard" key={course.id}>
      <div className="academyCourseCover">{course.coverImageUrl ? <img src={course.coverImageUrl} alt="" /> : <strong>{course.title}</strong>}</div>
      <div className="academyCourseBody"><span className="academyTag">{course.category}</span><h3>{course.title}</h3><p>{course.description || "Curso da Academia LicitaHub."}</p><div className="academyMeta"><span>{course.lessonCount} aula(s)</span><span>{course.workloadHours || 0}h de carga</span></div>{course.enrolled && <div className="academyProgress"><span style={{ width: `${course.progressPercent}%` }} /></div>}<Button onClick={() => navigate(`academy-course?id=${course.id}`)}>{course.enrolled ? "Continuar curso" : "Ver curso"}</Button></div>
    </article>)}</div>
  </Page>;
}

function AcademyMyLearning({ navigate }) {
  const [items, setItems] = useState([]); const [loading, setLoading] = useState(true); const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", status: "all" });
  useEffect(() => { academyRequest("/api/academy/my-learning").then((data) => setItems(Array.isArray(data) ? data : [])).catch((err) => setError(err.message)).finally(() => setLoading(false)); }, []);
  const filteredItems = items.filter((item) => {
    const term = filters.search.trim().toLowerCase();
    const matchesText = !term || [item.title, item.category].some((value) => String(value || "").toLowerCase().includes(term));
    const matchesStatus = filters.status === "all" || (filters.status === "completed" ? item.progressPercent === 100 : item.progressPercent < 100);
    return matchesText && matchesStatus;
  });
  return <Page label="Academia LicitaHub" title="Meus cursos" actions={<Button variant="secondary" onClick={() => navigate("academy-home")}>Ver catálogo</Button>}>
    {loading && <Card><p>Carregando seus cursos...</p></Card>}{error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {!loading && !error && !items.length && <Card><h3>Nenhum curso iniciado</h3><p>Escolha um curso no catálogo para começar sua jornada de aprendizagem.</p><Button onClick={() => navigate("academy-home")}>Abrir catálogo</Button></Card>}
    {!loading && items.length > 0 && <Card className="compactFilters academyLearningFilters"><FormGrid><Field label="Localizar curso"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Título ou categoria" /></Field><Field label="Andamento"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="all">Todos</option><option value="inProgress">Em andamento</option><option value="completed">Concluídos</option></select></Field></FormGrid></Card>}
    {!loading && items.length > 0 && filteredItems.length === 0 && <Card><p>Nenhum curso corresponde aos filtros selecionados.</p></Card>}
    <div className="academyLearningList">{filteredItems.map((item) => <Card key={item.id} className="academyLearningRow"><div className="academySmallCover">{item.coverImageUrl ? <img src={item.coverImageUrl} alt="" /> : <span>LH</span>}</div><div className="academyLearningInfo"><span className="academyTag">{item.category}</span><h3>{item.title}</h3><div className="academyProgress"><span style={{ width: `${item.progressPercent}%` }} /></div><small>{item.progressPercent}% concluído</small></div><div className="academyLearningActions"><Button onClick={() => navigate(`academy-course?id=${item.id}`)}>{item.progressPercent === 100 ? "Revisar curso" : "Continuar"}</Button>{item.certificateCode && <a className="linkButton" href={academyCertificateUrl(item.id)}>Baixar certificado em PDF</a>}</div></Card>)}</div>
  </Page>;
}

function AcademyAdminLegacy({ navigate }) {
  const emptyCourse = { title: "", description: "", category: "Engenharia consultiva", coverImageUrl: "", workloadHours: 1, status: "draft" };
  const emptyLesson = { title: "", description: "", videoUrl: "", durationSeconds: 0, requiresQuiz: false, passingScore: 70, maxAttempts: 0, questions: [{ question: "", options: ["", ""], correct: 0 }] };
  const [courses, setCourses] = useState([]); const [course, setCourse] = useState(emptyCourse); const [lesson, setLesson] = useState(emptyLesson); const [selectedCourse, setSelectedCourse] = useState(""); const [editingCourse, setEditingCourse] = useState(null); const [courseFilter, setCourseFilter] = useState("all"); const [message, setMessage] = useState(null); const [saving, setSaving] = useState(false);
  const load = () => academyRequest("/api/academy/courses").then((data) => { setCourses(Array.isArray(data) ? data : []); if (!selectedCourse && data?.[0]?.id) setSelectedCourse(data[0].id); }).catch((err) => setMessage({ type: "error", text: err.message }));
  useEffect(() => { load(); }, []);
  const saveCourse = async (event) => { event.preventDefault(); setSaving(true); setMessage(null); try { const created = await academyRequest("/api/academy/courses", { method: "POST", body: JSON.stringify({ ...course, workloadHours: Number(course.workloadHours) || 0 }) }); setCourse(emptyCourse); setSelectedCourse(created.id); setMessage({ type: "success", text: "Curso criado. Agora você pode incluir as aulas." }); load(); } catch (err) { setMessage({ type: "error", text: err.message }); } finally { setSaving(false); } };
  const saveLesson = async (event) => { event.preventDefault(); if (!selectedCourse) return; setSaving(true); setMessage(null); try { const questions = lesson.requiresQuiz ? lesson.questions.filter((q) => q.question.trim() && q.options.filter((option) => option.trim()).length >= 2).map((q) => ({ ...q, options: q.options.filter((option) => option.trim()) })) : []; await academyRequest(`/api/academy/courses/${selectedCourse}/lessons`, { method: "POST", body: JSON.stringify({ ...lesson, durationSeconds: Number(lesson.durationSeconds) || 0, passingScore: Number(lesson.passingScore) || 70, maxAttempts: Number(lesson.maxAttempts) || 0, questions }) }); setLesson(emptyLesson); setMessage({ type: "success", text: "Aula incluída no curso." }); } catch (err) { setMessage({ type: "error", text: err.message }); } finally { setSaving(false); } };
  const beginCourseEdit = (item) => { setEditingCourse({ ...item, workloadHours: Number(item.workloadHours) || 0 }); window.scrollTo({ top: 0, behavior: "smooth" }); };
  const saveCourseEdit = async (event) => { event.preventDefault(); if (!editingCourse) return; setSaving(true); setMessage(null); try { await academyRequest(`/api/academy/courses/${editingCourse.id}`, { method: "PUT", body: JSON.stringify(editingCourse) }); setEditingCourse(null); setMessage({ type: "success", text: "Curso atualizado com sucesso." }); await load(); } catch (err) { setMessage({ type: "error", text: err.message }); } finally { setSaving(false); } };
  const changeCourseStatus = async (item, status) => { const action = status === "archived" ? "arquivar" : "reativar"; if (!window.confirm(`Deseja ${action} o curso ${item.title}?`)) return; setMessage(null); try { await academyRequest(`/api/academy/courses/${item.id}`, { method: "PUT", body: JSON.stringify({ ...item, status }) }); setMessage({ type: "success", text: status === "archived" ? "Curso arquivado." : "Curso reativado como rascunho." }); await load(); } catch (err) { setMessage({ type: "error", text: err.message }); } };
  const deleteCourse = async (item) => { if (!window.confirm(`Excluir definitivamente o curso ${item.title}? Esta ação não pode ser desfeita.`)) return; setMessage(null); try { await academyRequest(`/api/academy/courses/${item.id}`, { method: "DELETE" }); if (selectedCourse === item.id) setSelectedCourse(""); if (editingCourse?.id === item.id) setEditingCourse(null); setMessage({ type: "success", text: "Curso excluído definitivamente." }); await load(); } catch (err) { setMessage({ type: "error", text: err.message }); } };
  const filteredCourses = courses.filter((item) => courseFilter === "all" || item.status === courseFilter);
  const changeQuestion = (index, field, value) => setLesson((current) => ({ ...current, questions: current.questions.map((question, questionIndex) => questionIndex === index ? { ...question, [field]: value } : question) }));
  const changeOption = (questionIndex, optionIndex, value) => setLesson((current) => ({ ...current, questions: current.questions.map((question, index) => index === questionIndex ? { ...question, options: question.options.map((option, i) => i === optionIndex ? value : option) } : question) }));
  return <Page label="Academia LicitaHub" title="Gerenciar cursos" actions={<Button variant="secondary" onClick={() => navigate("academy-home")}>Ver catálogo</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <div className="academyAdminGrid"><Card><h3>Novo curso</h3><form onSubmit={saveCourse}><FormGrid><Field label="Título do curso"><input value={course.title} onChange={(e) => setCourse({ ...course, title: e.target.value })} maxLength="220" required /></Field><Field label="Categoria"><input value={course.category} onChange={(e) => setCourse({ ...course, category: e.target.value })} /></Field><Field label="Carga horária estimada"><input type="number" min="0" value={course.workloadHours} onChange={(e) => setCourse({ ...course, workloadHours: e.target.value })} /></Field><Field label="Situação"><select value={course.status} onChange={(e) => setCourse({ ...course, status: e.target.value })}><option value="draft">Rascunho</option><option value="published">Publicado</option></select></Field><Field label="Imagem de capa (link)"><input type="url" value={course.coverImageUrl} onChange={(e) => setCourse({ ...course, coverImageUrl: e.target.value })} placeholder="https://..." /></Field><Field label="Descrição"><textarea value={course.description} onChange={(e) => setCourse({ ...course, description: e.target.value })} rows="4" /></Field></FormGrid><Button type="submit" disabled={saving}>{saving ? "Salvando..." : "Criar curso"}</Button></form></Card>
      <Card><h3>Incluir aula</h3><form onSubmit={saveLesson}><FormGrid><Field label="Curso"><select value={selectedCourse} onChange={(e) => setSelectedCourse(e.target.value)} required><option value="">Selecione</option>{courses.map((item) => <option value={item.id} key={item.id}>{item.title}</option>)}</select></Field><Field label="Título da aula"><input value={lesson.title} onChange={(e) => setLesson({ ...lesson, title: e.target.value })} required /></Field><Field label="Link do vídeo"><input type="url" value={lesson.videoUrl} onChange={(e) => setLesson({ ...lesson, videoUrl: e.target.value })} placeholder="YouTube ou outro vídeo incorporável" required /></Field><Field label="Duração em segundos"><input type="number" min="0" value={lesson.durationSeconds} onChange={(e) => setLesson({ ...lesson, durationSeconds: e.target.value })} /></Field><Field label="Descrição"><textarea value={lesson.description} onChange={(e) => setLesson({ ...lesson, description: e.target.value })} rows="3" /></Field></FormGrid><label className="checkOption"><input type="checkbox" checked={lesson.requiresQuiz} onChange={(e) => setLesson({ ...lesson, requiresQuiz: e.target.checked })} /> Esta aula exige prova para liberar a próxima.</label>{lesson.requiresQuiz && <div className="academyQuizBuilder"><FormGrid><Field label="Aproveitamento mínimo (%)"><input type="number" min="1" max="100" value={lesson.passingScore} onChange={(e) => setLesson({ ...lesson, passingScore: e.target.value })} /></Field><Field label="Tentativas (0 = sem limite)"><input type="number" min="0" value={lesson.maxAttempts} onChange={(e) => setLesson({ ...lesson, maxAttempts: e.target.value })} /></Field></FormGrid>{lesson.questions.map((question, questionIndex) => <div className="academyQuestionEditor" key={questionIndex}><strong>Questão {questionIndex + 1}</strong><input value={question.question} onChange={(e) => changeQuestion(questionIndex, "question", e.target.value)} placeholder="Enunciado" />{question.options.map((option, optionIndex) => <label className="academyOptionEditor" key={optionIndex}><input type="radio" name={`correct-${questionIndex}`} checked={question.correct === optionIndex} onChange={() => changeQuestion(questionIndex, "correct", optionIndex)} /><input value={option} onChange={(e) => changeOption(questionIndex, optionIndex, e.target.value)} placeholder={`Alternativa ${optionIndex + 1}`} /></label>)}<button type="button" className="linkButton" onClick={() => setLesson((current) => ({ ...current, questions: current.questions.map((q, i) => i === questionIndex ? { ...q, options: [...q.options, ""] } : q) }))}>Adicionar alternativa</button></div>)}<button type="button" className="linkButton" onClick={() => setLesson((current) => ({ ...current, questions: [...current.questions, { question: "", options: ["", ""], correct: 0 }] }))}>Adicionar questão</button></div>}<Button type="submit" disabled={saving || !selectedCourse}>{saving ? "Salvando..." : "Adicionar aula"}</Button></form></Card></div>
    {editingCourse && <Card className="academyCourseEditor"><div className="sectionTitle"><div><span className="eyebrow">Edição do curso</span><h3>{editingCourse.title}</h3></div><Button variant="secondary" onClick={() => setEditingCourse(null)}>Fechar</Button></div><form onSubmit={saveCourseEdit}><FormGrid><Field label="Título do curso"><input value={editingCourse.title} onChange={(e) => setEditingCourse({ ...editingCourse, title: e.target.value })} maxLength="220" required /></Field><Field label="Categoria"><input value={editingCourse.category} onChange={(e) => setEditingCourse({ ...editingCourse, category: e.target.value })} /></Field><Field label="Carga horária estimada"><input type="number" min="0" value={editingCourse.workloadHours} onChange={(e) => setEditingCourse({ ...editingCourse, workloadHours: e.target.value })} /></Field><Field label="Situação"><select value={editingCourse.status} onChange={(e) => setEditingCourse({ ...editingCourse, status: e.target.value })}><option value="draft">Rascunho</option><option value="published">Publicado</option><option value="archived">Arquivado</option></select></Field><Field label="Imagem de capa (link)"><input type="url" value={editingCourse.coverImageUrl || ""} onChange={(e) => setEditingCourse({ ...editingCourse, coverImageUrl: e.target.value })} placeholder="https://..." /></Field><Field label="Descrição"><textarea value={editingCourse.description || ""} onChange={(e) => setEditingCourse({ ...editingCourse, description: e.target.value })} rows="4" /></Field></FormGrid><div className="formActions"><Button type="submit" disabled={saving}>{saving ? "Salvando..." : "Salvar alterações"}</Button><Button variant="secondary" onClick={() => setEditingCourse(null)}>Cancelar</Button></div></form></Card>}
    <Card className="academyCourseManagement"><div className="sectionTitle"><div><span className="eyebrow">Administração</span><h3>Cursos cadastrados</h3><p>Edite os dados, altere a situação ou exclua cursos que ainda não possuem alunos.</p></div><Field label="Filtrar por situação"><select value={courseFilter} onChange={(e) => setCourseFilter(e.target.value)}><option value="all">Todos</option><option value="published">Publicados</option><option value="draft">Rascunhos</option><option value="archived">Arquivados</option></select></Field></div>{filteredCourses.length === 0 ? <p>Nenhum curso encontrado neste filtro.</p> : <div className="academyAdminCourseList">{filteredCourses.map((item) => <article className="academyAdminCourseRow" key={item.id}><div className="academyAdminCourseCover">{item.coverImageUrl ? <img src={item.coverImageUrl} alt="" /> : <span>LH</span>}</div><div className="academyAdminCourseInfo"><div><span className={`statusPill ${item.status === "published" ? "open" : item.status === "archived" ? "review" : ""}`}>{item.status === "published" ? "Publicado" : item.status === "archived" ? "Arquivado" : "Rascunho"}</span><span>{item.category}</span></div><h4>{item.title}</h4><small>{item.lessonCount || 0} aula(s) · {item.enrollmentCount || 0} aluno(s) · {item.workloadHours || 0}h</small></div><div className="academyAdminCourseActions"><button className="iconButton secondaryIcon" title="Editar curso" aria-label="Editar curso" onClick={() => beginCourseEdit(item)}>✎</button><button className="iconButton partnerIcon" title="Abrir curso" aria-label="Abrir curso" onClick={() => navigate(`academy-course?id=${item.id}`)}>◉</button>{item.status === "archived" ? <button className="iconButton successIcon" title="Reativar como rascunho" aria-label="Reativar como rascunho" onClick={() => changeCourseStatus(item, "draft")}>↻</button> : <button className="iconButton secondaryIcon" title="Arquivar curso" aria-label="Arquivar curso" onClick={() => changeCourseStatus(item, "archived")}>▣</button>}<button className="iconButton dangerIcon" title={item.enrollmentCount ? "Curso com alunos: use arquivar" : "Excluir curso"} aria-label="Excluir curso" disabled={Boolean(item.enrollmentCount)} onClick={() => deleteCourse(item)}>×</button></div></article>)}</div>}</Card>
  </Page>;
}

function academyStatusLabel(status) {
  return { published: "Publicado", draft: "Rascunho", archived: "Arquivado" }[status] || status;
}

function AcademyAdmin({ navigate }) {
  const [courses, setCourses] = useState([]);
  const [filters, setFilters] = useState({ search: "", status: "all" });
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [courseEditor, setCourseEditor] = useState(null);
  const load = () => { setLoading(true); return academyRequest("/api/academy/courses").then((data) => setCourses(Array.isArray(data) ? data : [])).catch((error) => setMessage({ type: "error", text: error.message })).finally(() => setLoading(false)); };
  useEffect(() => { load(); }, []);
  const changeStatus = async (item, status) => { const verb = status === "archived" ? "arquivar" : "reativar"; if (!window.confirm(`Deseja ${verb} o curso ${item.title}?`)) return; try { await academyRequest(`/api/academy/courses/${item.id}`, { method: "PUT", body: JSON.stringify({ ...item, status }) }); setMessage({ type: "success", text: status === "archived" ? "Curso arquivado." : "Curso reativado como rascunho." }); load(); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  const remove = async (item) => { if (!window.confirm(`Excluir definitivamente o curso ${item.title}?`)) return; try { await academyRequest(`/api/academy/courses/${item.id}`, { method: "DELETE" }); setMessage({ type: "success", text: "Curso excluído." }); load(); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  const filtered = courses.filter((item) => {
    const term = filters.search.trim().toLowerCase();
    return (filters.status === "all" || item.status === filters.status) && (!term || [item.title, item.category, item.description].some((value) => String(value || "").toLowerCase().includes(term)));
  });
  const openCourseEditor = (courseId = "") => {
    setCourseEditor({ courseId });
    window.scrollTo({ top: 0, behavior: "auto" });
  };
  if (courseEditor) {
    return <AcademyCourseAdmin
      navigate={navigate}
      initialCourseId={courseEditor.courseId}
      onBack={() => {
        setCourseEditor(null);
        load();
      }}
    />;
  }
  return <Page label="Academia LicitaHub" title="Gerenciar cursos" actions={<Button onClick={() => openCourseEditor()}>Novo curso</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <Card className="compactFilters academyAdminFilters"><FormGrid><Field label="Buscar curso"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Título, categoria ou descrição" /></Field><Field label="Situação"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="all">Todos</option><option value="published">Publicados</option><option value="draft">Rascunhos</option><option value="archived">Arquivados</option></select></Field></FormGrid></Card>
    {loading ? <Card><p>Carregando cursos...</p></Card> : filtered.length === 0 ? <Card><p>Nenhum curso encontrado.</p></Card> : <section className="academyAdminCourseListPanel" aria-label="Cursos cadastrados"><div className="academyAdminCourseListHeader"><span>Capa</span><span>Curso e situação</span><span>Ações</span></div><div className="academyAdminCourseList">{filtered.map((item) => <article className="academyAdminCourseRow" key={item.id}><div className="academyAdminCourseCover">{item.coverImageUrl ? <img src={item.coverImageUrl} alt="" /> : <span>LH</span>}</div><div className="academyAdminCourseInfo"><div><span className={`statusPill ${item.status === "published" ? "open" : item.status === "archived" ? "review" : ""}`}>{academyStatusLabel(item.status)}</span><span>{item.category}</span></div><h4>{item.title}</h4><small>{item.lessonCount || 0} aula(s) · {item.enrollmentCount || 0} aluno(s) · {item.workloadHours || 0}h</small></div><div className="academyAdminCourseActions"><button className="iconButton secondaryIcon" title="Gerenciar curso" aria-label={`Gerenciar ${item.title}`} onClick={() => openCourseEditor(item.id)}>✎</button>{item.status === "archived" ? <button className="iconButton successIcon" title="Reativar como rascunho" aria-label={`Reativar ${item.title}`} onClick={() => changeStatus(item, "draft")}>↻</button> : <button className="iconButton secondaryIcon" title="Arquivar curso" aria-label={`Arquivar ${item.title}`} onClick={() => changeStatus(item, "archived")}>▣</button>}<button className="iconButton dangerIcon" title={item.enrollmentCount ? "Curso com alunos não pode ser excluído" : "Excluir curso"} aria-label={`Excluir ${item.title}`} disabled={Boolean(item.enrollmentCount)} onClick={() => remove(item)}>×</button></div></article>)}</div></section>}
  </Page>;
}

function academyBlankLesson() {
  return { title: "", description: "", videoUrl: "", videoSource: "youtube", videoFileName: "", durationMinutes: "", requiresQuiz: false, maxAttempts: 0, isPublished: true, questions: [{ question: "", options: ["", ""], correct: 0 }] };
}

function youtubeThumbnail(videoUrl) {
  const match = String(videoUrl || "").match(/(?:youtu\.be\/|v=|embed\/)([^?&/]+)/i);
  return match ? `https://img.youtube.com/vi/${match[1]}/hqdefault.jpg` : "";
}

function AcademyLessonVideoFields({ lesson, setLesson, changeVideoSource, uploadLessonVideo, uploadingVideo }) {
  return (
    <div className="academyVideoSourceEditor">
      <div className="academyVideoSourceChoice" role="radiogroup" aria-label="Origem do vídeo">
        <label className={lesson.videoSource === "youtube" ? "active" : ""}>
          <input type="radio" name="academy-video-source" checked={lesson.videoSource === "youtube"} onChange={() => changeVideoSource("youtube")} />
          <span>Link do YouTube</span>
        </label>
        <label className={lesson.videoSource === "upload" ? "active" : ""}>
          <input type="radio" name="academy-video-source" checked={lesson.videoSource === "upload"} onChange={() => changeVideoSource("upload")} />
          <span>Vídeo no LicitaHub</span>
        </label>
      </div>
      {lesson.videoSource === "youtube" ? (
        <Field label="Link do vídeo no YouTube">
          <input type="url" value={lesson.videoUrl} onChange={(event) => setLesson({ ...lesson, videoUrl: event.target.value })} placeholder="https://www.youtube.com/watch?v=..." required />
        </Field>
      ) : (
        <Field label="Arquivo do vídeo">
          <input type="file" accept="video/mp4,video/webm" onChange={uploadLessonVideo} disabled={uploadingVideo} />
          <small>{uploadingVideo ? "Enviando vídeo..." : lesson.videoFileName || "Formatos aceitos: MP4 ou WebM, até 500 MB."}</small>
        </Field>
      )}
      <div className="academyLessonPreview">
        {lesson.videoSource === "youtube" && youtubeThumbnail(lesson.videoUrl) ? (
          <img src={youtubeThumbnail(lesson.videoUrl)} alt="Miniatura obtida do YouTube" />
        ) : lesson.videoSource === "upload" && lesson.videoUrl ? (
          <video src={lesson.videoUrl} controls preload="metadata" />
        ) : (
          <span>{lesson.videoSource === "youtube" ? "A miniatura do YouTube aparecerá aqui" : "A prévia do vídeo enviado aparecerá aqui"}</span>
        )}
      </div>
    </div>
  );
}

function AcademyCourseAdmin({ navigate, initialCourseId = "", onBack = null }) {
  const initialId = initialCourseId || currentHashParams().get("id") || "";
  const [courseId, setCourseId] = useState(initialId);
  const [course, setCourse] = useState({ title: "", description: "", category: "Engenharia consultiva", coverImageUrl: "", workloadHours: 1, status: "draft" });
  const [lessons, setLessons] = useState([]);
  const [lesson, setLesson] = useState(academyBlankLesson());
  const [lessonEditorOpen, setLessonEditorOpen] = useState(false);
  const [editingLessonId, setEditingLessonId] = useState("");
  const [quizOpen, setQuizOpen] = useState(false);
  const [loading, setLoading] = useState(Boolean(initialId));
  const [saving, setSaving] = useState(false);
  const [uploadingVideo, setUploadingVideo] = useState(false);
  const [message, setMessage] = useState(null);
  const loadCourse = () => { if (!courseId) return Promise.resolve(); setLoading(true); return academyRequest(`/api/academy/courses/${courseId}`).then((data) => { setCourse({ title: data.title || "", description: data.description || "", category: data.category || "Geral", coverImageUrl: data.coverImageUrl || "", workloadHours: Number(data.workloadHours) || 0, status: data.status || "draft" }); setLessons(Array.isArray(data.lessons) ? data.lessons : []); }).catch((error) => setMessage({ type: "error", text: error.message })).finally(() => setLoading(false)); };
  useEffect(() => { loadCourse(); }, [courseId]);
  const selectCover = (event) => { const file = event.target.files?.[0]; if (!file) return; if (!file.type.startsWith("image/")) { setMessage({ type: "error", text: "Selecione um arquivo de imagem." }); return; } if (file.size > 3 * 1024 * 1024) { setMessage({ type: "error", text: "A imagem deve ter no máximo 3 MB." }); return; } const reader = new FileReader(); reader.onload = () => setCourse((current) => ({ ...current, coverImageUrl: String(reader.result || "") })); reader.readAsDataURL(file); };
  const saveCourse = async (event) => { event.preventDefault(); if (course.status === "published" && lessons.length === 0) { setMessage({ type: "error", text: "Inclua ao menos uma aula antes de publicar o curso." }); return; } setSaving(true); setMessage(null); try { if (courseId) { await academyRequest(`/api/academy/courses/${courseId}`, { method: "PUT", body: JSON.stringify({ ...course, workloadHours: Number(course.workloadHours) || 0 }) }); setMessage({ type: "success", text: "Informações do curso atualizadas." }); await loadCourse(); } else { const created = await academyRequest("/api/academy/courses", { method: "POST", body: JSON.stringify({ ...course, status: "draft", workloadHours: Number(course.workloadHours) || 0 }) }); setCourseId(created.id); if (!onBack) navigate(`academy-course-admin?id=${created.id}`); setMessage({ type: "success", text: "Curso criado como rascunho. Agora inclua as aulas e publique quando estiver pronto." }); } } catch (error) { setMessage({ type: "error", text: error.message }); } finally { setSaving(false); } };
  const openNewLesson = () => { setEditingLessonId(""); setLesson(academyBlankLesson()); setQuizOpen(false); setLessonEditorOpen(true); };
  const openLessonEdit = (item) => { setEditingLessonId(item.id); setLesson({ title: item.title || "", description: item.description || "", videoUrl: item.videoUrl || "", videoSource: item.videoSource || "youtube", videoFileName: item.videoSource === "upload" ? "Vídeo já enviado" : "", durationMinutes: item.durationSeconds ? String(Math.ceil(item.durationSeconds / 60)) : "", requiresQuiz: Boolean(item.requiresQuiz), maxAttempts: Number(item.maxAttempts) || 0, isPublished: item.isPublished !== false, questions: Array.isArray(item.quizQuestions) && item.quizQuestions.length ? item.quizQuestions.map((question) => ({ question: question.question || "", options: Array.isArray(question.options) ? question.options : ["", ""], correct: Number(question.correct) || 0 })) : [{ question: "", options: ["", ""], correct: 0 }] }); setQuizOpen(Boolean(item.requiresQuiz)); setLessonEditorOpen(true); };
  const closeLessonEditor = () => { setLessonEditorOpen(false); setEditingLessonId(""); setLesson(academyBlankLesson()); setQuizOpen(false); };
  const updateQuestion = (index, field, value) => setLesson((current) => ({ ...current, questions: current.questions.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item) }));
  const updateOption = (questionIndex, optionIndex, value) => setLesson((current) => ({ ...current, questions: current.questions.map((item, itemIndex) => itemIndex === questionIndex ? { ...item, options: item.options.map((option, index) => index === optionIndex ? value : option) } : item) }));
  const changeVideoSource = (videoSource) => setLesson((current) => ({ ...current, videoSource, videoUrl: "", videoFileName: "" }));
  const uploadLessonVideo = async (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!["video/mp4", "video/webm"].includes(file.type)) {
      setMessage({ type: "error", text: "Selecione um vídeo MP4 ou WebM." });
      event.target.value = "";
      return;
    }
    if (file.size > 500 * 1024 * 1024) {
      setMessage({ type: "error", text: "O vídeo deve ter no máximo 500 MB." });
      event.target.value = "";
      return;
    }
    setUploadingVideo(true);
    setMessage({ type: "success", text: "Enviando o vídeo. Aguarde a conclusão antes de salvar a aula." });
    try {
      const body = new FormData();
      body.append("video", file);
      const response = await fetch(`${API_BASE_URL}/api/academy/videos`, { method: "POST", credentials: "include", body });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível enviar o vídeo.");
      setLesson((current) => ({ ...current, videoSource: "upload", videoUrl: data.url, videoFileName: data.fileName || file.name }));
      setMessage({ type: "success", text: "Vídeo enviado. Agora você pode salvar a aula." });
    } catch (error) {
      setMessage({ type: "error", text: error.message });
      event.target.value = "";
    } finally {
      setUploadingVideo(false);
    }
  };
  const saveLesson = async (event) => {
    event.preventDefault();
    if (!courseId || uploadingVideo) return;
    if (!lesson.videoUrl) {
      setMessage({ type: "error", text: lesson.videoSource === "upload" ? "Envie o arquivo do vídeo antes de salvar." : "Informe o link do vídeo no YouTube." });
      return;
    }
    const questions = lesson.requiresQuiz ? lesson.questions.map((item) => ({ question: item.question.trim(), options: item.options.map((option) => option.trim()).filter(Boolean), correct: item.correct })) : [];
    if (lesson.requiresQuiz && questions.some((item) => !item.question || item.options.length < 2 || item.correct >= item.options.length)) {
      setMessage({ type: "error", text: "Revise as perguntas, alternativas e respostas corretas." });
      return;
    }
    setSaving(true);
    setMessage(null);
    try {
      const payload = { ...lesson, durationSeconds: Math.max(0, Number(lesson.durationMinutes) || 0) * 60, passingScore: 75, maxAttempts: Number(lesson.maxAttempts) || 0, questions };
      if (editingLessonId) await academyRequest(`/api/academy/lessons/${editingLessonId}`, { method: "PUT", body: JSON.stringify(payload) });
      else await academyRequest(`/api/academy/courses/${courseId}/lessons`, { method: "POST", body: JSON.stringify(payload) });
      setMessage({ type: "success", text: editingLessonId ? "Aula atualizada." : "Aula incluída no curso." });
      closeLessonEditor();
      await loadCourse();
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving(false);
    }
  };
  const removeLesson = async (item) => { if (!window.confirm(`Remover a aula ${item.title}?`)) return; try { const result = await academyRequest(`/api/academy/lessons/${item.id}`, { method: "DELETE" }); setMessage({ type: "success", text: result.action === "archived" ? "A aula possui histórico e foi arquivada." : "Aula excluída." }); loadCourse(); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  if (loading) return <Page label="Academia LicitaHub" title="Gestão do curso"><Card><p>Carregando curso...</p></Card></Page>;
  return <Page label="Academia LicitaHub" title={courseId ? "Gestão do curso" : "Novo curso"} actions={<Button variant="secondary" onClick={() => onBack ? onBack() : navigate("academy-admin")}>Voltar aos cursos</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <Card className="academyCourseIdentity"><div className="academyCourseIdentityCover">{course.coverImageUrl ? <img src={course.coverImageUrl} alt="Prévia da capa do curso" /> : <span>Imagem do curso</span>}</div><form onSubmit={saveCourse}><div className="sectionTitle"><div><span className="eyebrow">Informações do curso</span><h3>{courseId ? course.title || "Curso sem título" : "Cadastre a identidade do curso"}</h3></div></div><FormGrid><Field label="Título"><input value={course.title} onChange={(event) => setCourse({ ...course, title: event.target.value })} maxLength="220" required /></Field><Field label="Categoria"><input value={course.category} onChange={(event) => setCourse({ ...course, category: event.target.value })} maxLength="120" /></Field><Field label="Carga horária"><input type="number" min="0" value={course.workloadHours} onChange={(event) => setCourse({ ...course, workloadHours: event.target.value })} /></Field><Field label="Situação"><select value={course.status} onChange={(event) => setCourse({ ...course, status: event.target.value })}><option value="draft">Rascunho</option><option value="published" disabled={!lessons.length}>Publicado</option><option value="archived">Arquivado</option></select></Field><Field label="Imagem de capa"><input type="file" accept="image/*" onChange={selectCover} /></Field><Field label="Descrição breve"><textarea rows="4" value={course.description} onChange={(event) => setCourse({ ...course, description: event.target.value })} maxLength="1500" /></Field></FormGrid><div className="formActions"><Button type="submit" disabled={saving}>{saving ? "Salvando..." : courseId ? "Salvar informações" : "Criar curso"}</Button>{course.coverImageUrl && <Button variant="secondary" onClick={() => setCourse({ ...course, coverImageUrl: "" })}>Remover imagem</Button>}</div></form></Card>
    {courseId && <Card className="academyLessonManagement"><div className="sectionTitle"><div><span className="eyebrow">Conteúdo do curso</span><h3>Aulas cadastradas</h3><p>Inclua quantas aulas forem necessárias. A ordem abaixo será a ordem de avanço do aluno.</p></div><Button onClick={openNewLesson}>Adicionar aula</Button></div>{lessons.length === 0 ? <div className="emptyState"><p>Nenhuma aula cadastrada. Inclua a primeira aula para poder publicar o curso.</p></div> : <div className="academyAdminLessonList">{lessons.map((item, index) => <article className={`academyAdminLesson ${item.isPublished === false ? "isArchived" : ""}`} key={item.id}><div className="academyAdminLessonThumb">{youtubeThumbnail(item.videoUrl) ? <img src={youtubeThumbnail(item.videoUrl)} alt="" /> : <span>{index + 1}</span>}</div><div><span className="eyebrow">Aula {index + 1}</span><h4>{item.title}</h4><p>{item.description || "Sem descrição."}</p><small>{item.requiresQuiz ? "Questionário obrigatório · aprovação mínima de 75%" : "Sem questionário"}{item.isPublished === false ? " · Arquivada" : ""}</small></div><div className="academyAdminLessonActions"><button className="iconButton secondaryIcon" title="Editar aula" onClick={() => openLessonEdit(item)}>✎</button><button className="iconButton dangerIcon" title="Remover aula" onClick={() => removeLesson(item)}>×</button></div></article>)}</div>}</Card>}
    {courseId && lessonEditorOpen && (
      <Card className="academyLessonEditor">
        <div className="sectionTitle">
          <div><span className="eyebrow">{editingLessonId ? "Edição" : "Nova aula"}</span><h3>{editingLessonId ? "Editar aula" : "Incluir aula"}</h3></div>
          <Button variant="secondary" onClick={closeLessonEditor}>Fechar</Button>
        </div>
        <form onSubmit={saveLesson}>
          <AcademyLessonVideoFields lesson={lesson} setLesson={setLesson} changeVideoSource={changeVideoSource} uploadLessonVideo={uploadLessonVideo} uploadingVideo={uploadingVideo} />
          <FormGrid>
            <Field label="Título da aula"><input value={lesson.title} onChange={(event) => setLesson({ ...lesson, title: event.target.value })} required maxLength="220" /></Field>
            <Field label="Duração aproximada em minutos"><input type="number" min="0" value={lesson.durationMinutes} onChange={(event) => setLesson({ ...lesson, durationMinutes: event.target.value })} /></Field>
            <Field label="Descrição da aula"><textarea rows="3" value={lesson.description} onChange={(event) => setLesson({ ...lesson, description: event.target.value })} maxLength="2000" /></Field>
          </FormGrid>
          <div className="academyQuizToggle">
            <div><strong>Questionário validante</strong><p>Quando ativo, o aluno precisa assistir ao vídeo e acertar pelo menos 75% para liberar a próxima aula.</p></div>
            <Button type="button" variant={quizOpen ? "secondary" : "primary"} onClick={() => { const next = !quizOpen; setQuizOpen(next); setLesson((current) => ({ ...current, requiresQuiz: next })); }}>{quizOpen ? "Retirar questionário" : "Adicionar questionário"}</Button>
          </div>
          {quizOpen && (
            <div className="academyQuizBuilder">
              <Field label="Limite de tentativas"><input type="number" min="0" value={lesson.maxAttempts} onChange={(event) => setLesson({ ...lesson, maxAttempts: event.target.value })} /><small>Use zero para tentativas ilimitadas.</small></Field>
              {lesson.questions.map((question, questionIndex) => (
                <section className="academyQuestionEditor" key={questionIndex}>
                  <div className="sectionTitle"><strong>Questão {questionIndex + 1}</strong>{lesson.questions.length > 1 && <button type="button" className="iconButton dangerIcon" title="Remover questão" onClick={() => setLesson((current) => ({ ...current, questions: current.questions.filter((_, index) => index !== questionIndex) }))}>×</button>}</div>
                  <input value={question.question} onChange={(event) => updateQuestion(questionIndex, "question", event.target.value)} placeholder="Escreva a pergunta" />
                  {question.options.map((option, optionIndex) => (
                    <label className="academyOptionEditor" key={optionIndex}>
                      <input type="radio" name={`correct-answer-${questionIndex}`} checked={question.correct === optionIndex} onChange={() => updateQuestion(questionIndex, "correct", optionIndex)} title="Marcar como resposta correta" />
                      <input value={option} onChange={(event) => updateOption(questionIndex, optionIndex, event.target.value)} placeholder={`Alternativa ${optionIndex + 1}`} />
                      {question.options.length > 2 && <button type="button" className="iconButton dangerIcon" title="Remover alternativa" onClick={() => setLesson((current) => ({ ...current, questions: current.questions.map((item, index) => index === questionIndex ? { ...item, options: item.options.filter((_, itemIndex) => itemIndex !== optionIndex), correct: item.correct === optionIndex ? 0 : item.correct > optionIndex ? item.correct - 1 : item.correct } : item) }))}>×</button>}
                    </label>
                  ))}
                  <button type="button" className="linkButton" onClick={() => setLesson((current) => ({ ...current, questions: current.questions.map((item, index) => index === questionIndex ? { ...item, options: [...item.options, ""] } : item) }))}>Adicionar alternativa</button>
                </section>
              ))}
              <Button type="button" variant="secondary" onClick={() => setLesson((current) => ({ ...current, questions: [...current.questions, { question: "", options: ["", ""], correct: 0 }] }))}>Adicionar questão</Button>
            </div>
          )}
          <div className="formActions">
            <Button type="submit" disabled={saving || uploadingVideo}>{uploadingVideo ? "Enviando vídeo..." : saving ? "Salvando..." : editingLessonId ? "Salvar aula" : "Incluir aula"}</Button>
            <Button variant="secondary" onClick={closeLessonEditor}>Cancelar</Button>
          </div>
        </form>
      </Card>
    )}
  </Page>;
}

function youtubeEmbed(videoUrl, start) {
  const match = String(videoUrl || "").match(/(?:youtu\.be\/|v=|embed\/)([^?&/]+)/i);
  if (!match) return videoUrl;
  const origin = encodeURIComponent(window.location.origin);
  return `https://www.youtube.com/embed/${match[1]}?enablejsapi=1&rel=0&playsinline=1&origin=${origin}&start=${Math.max(0, Number(start) || 0)}`;
}

function AcademyCourseLegacy({ navigate }) {
  const courseId = currentHashParams().get("id"); const [course, setCourse] = useState(null); const [selectedId, setSelectedId] = useState(""); const [answers, setAnswers] = useState({}); const [message, setMessage] = useState(null); const [loading, setLoading] = useState(true); const iframeRef = React.useRef(null); const lastPosition = React.useRef(0);
  const load = () => { if (!courseId) { setLoading(false); return; } setLoading(true); academyRequest(`/api/academy/courses/${courseId}`).then((data) => { setCourse(data); setSelectedId((current) => current || data.lessons?.find((lesson) => lesson.unlocked)?.id || data.lessons?.[0]?.id || ""); }).catch((err) => setMessage({ type: "error", text: err.message })).finally(() => setLoading(false)); };
  useEffect(() => { load(); }, [courseId]);
  const selected = course?.lessons?.find((lesson) => lesson.id === selectedId);
  useEffect(() => { if (!selected) return undefined; lastPosition.current = selected.lastPositionSeconds || 0; const timer = setInterval(() => iframeRef.current?.contentWindow?.postMessage(JSON.stringify({ event: "command", func: "getCurrentTime", args: [] }), "*"), 15000); const onMessage = (event) => { try { const data = typeof event.data === "string" ? JSON.parse(event.data) : event.data; if (typeof data?.info?.currentTime === "number") lastPosition.current = Math.floor(data.info.currentTime); } catch (_) {} }; window.addEventListener("message", onMessage); return () => { clearInterval(timer); window.removeEventListener("message", onMessage); }; }, [selected?.id]);
  const enroll = async () => { try { await academyRequest(`/api/academy/courses/${courseId}/enroll`, { method: "POST" }); setMessage({ type: "success", text: "Curso iniciado. Seu progresso será salvo." }); load(); } catch (err) { setMessage({ type: "error", text: err.message }); } };
  const saveProgress = async () => { if (!selected) return; try { await academyRequest(`/api/academy/lessons/${selected.id}/progress`, { method: "PUT", body: JSON.stringify({ positionSeconds: lastPosition.current }) }); setMessage({ type: "success", text: "Ponto do vídeo salvo." }); load(); } catch (err) { setMessage({ type: "error", text: err.message }); } };
  const submitQuiz = async () => { try { const selectedAnswers = selected.quizQuestions.map((_, index) => Number(answers[index])); const result = await academyRequest(`/api/academy/lessons/${selected.id}/assessment`, { method: "POST", body: JSON.stringify({ answers: selectedAnswers }) }); setMessage({ type: result.approved ? "success" : "error", text: result.approved ? "Prova aprovada. A próxima aula foi liberada." : `Aproveitamento de ${Math.round(result.score || 0)}%. Revise a aula e tente novamente.` }); setAnswers({}); load(); } catch (err) { setMessage({ type: "error", text: err.message }); } };
  if (loading) return <Page label="Academia LicitaHub" title="Curso"><Card><p>Carregando curso...</p></Card></Page>;
  if (!course) return <Page label="Academia LicitaHub" title="Curso"><Card className="dangerNotice"><p>{message?.text || "Curso não encontrado."}</p></Card></Page>;
  return <Page label="Academia LicitaHub" title="Curso" actions={<Button variant="secondary" onClick={() => navigate("academy-my-learning")}>Meus cursos</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <Card className="academyCourseHeader"><div><span className="academyTag">{course.category}</span><h3>{course.title}</h3><p>{course.description}</p></div>{!course.enrolled && <Button onClick={enroll}>Iniciar curso</Button>}{course.certificateCode && <button className="certificateButton" onClick={() => window.print()}>Certificado emitido: {course.certificateCode}</button>}</Card>
    {!course.enrolled ? <Card><p>Ao iniciar o curso, o progresso das aulas será registrado no seu perfil.</p></Card> : <div className="academyPlayerLayout"><Card className="academyLessonRail"><h3>Conteúdo</h3>{course.lessons.map((lesson, index) => <button className={`academyLessonItem ${selectedId === lesson.id ? "active" : ""}`} key={lesson.id} disabled={!lesson.unlocked} onClick={() => setSelectedId(lesson.id)}><span>{lesson.completed ? "✓" : String(index + 1).padStart(2, "0")}</span><div><strong>{lesson.title}</strong><small>{lesson.unlocked ? (lesson.completed ? "Concluída" : "Disponível") : "Bloqueada até concluir a anterior"}</small></div></button>)}</Card><div className="academyLessonContent">{selected && <><Card><h3>{selected.title}</h3>{selected.description && <p>{selected.description}</p>}<div className="academyVideo"><iframe ref={iframeRef} title={selected.title} src={youtubeEmbed(selected.videoUrl, selected.lastPositionSeconds)} allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen /></div><div className="academyVideoActions"><small>O ponto do vídeo é salvo periodicamente. Você também pode salvar manualmente.</small><Button variant="secondary" onClick={saveProgress}>Salvar ponto</Button></div></Card>{selected.requiresQuiz && <Card className="academyQuiz"><h3>Prova da aula</h3>{selected.quizQuestions.map((question, questionIndex) => <fieldset key={questionIndex}><legend>{questionIndex + 1}. {question.question}</legend>{question.options.map((option, optionIndex) => <label className="checkOption" key={optionIndex}><input type="radio" name={`answer-${selected.id}-${questionIndex}`} checked={Number(answers[questionIndex]) === optionIndex} onChange={() => setAnswers({ ...answers, [questionIndex]: optionIndex })} /> {option}</label>)}</fieldset>)}<Button onClick={submitQuiz}>Enviar prova</Button></Card>}</>}</div></div>}
    {course.certificateCode && <section className="academyCertificate"><span>Academia LicitaHub certifica a conclusão de</span><h2>{course.title}</h2><p>Certificado verificável: <strong>{course.certificateCode}</strong></p></section>}
  </Page>;
}

function formatAcademyTime(seconds) {
  const total = Math.max(0, Math.floor(Number(seconds) || 0));
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

function AcademyAdaptivePlayer({ lesson, playerRef, initializePlayer }) {
  if (lesson.videoSource === "upload") {
    return (
      <video
        key={lesson.id}
        ref={playerRef}
        title={lesson.title}
        src={lesson.videoUrl}
        controls
        preload="metadata"
        onLoadedMetadata={(event) => {
          const savedPosition = Math.max(0, Number(lesson.lastPositionSeconds) || 0);
          if (savedPosition > 0 && savedPosition < event.currentTarget.duration) event.currentTarget.currentTime = savedPosition;
        }}
      />
    );
  }
  return <iframe key={lesson.id} ref={playerRef} onLoad={initializePlayer} title={lesson.title} src={youtubeEmbed(lesson.videoUrl, lesson.lastPositionSeconds)} allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen />;
}

function AcademyQuizPanel({ lesson, answers, wrongQuestionIndexes, setAnswers, setWrongQuestionIndexes, submitQuiz }) {
  if (lesson.completed) {
    return <Card className="academyQuiz academyQuizCompleted"><span className="statusPill open">Questionário aprovado</span><h3>Avaliação concluída</h3><p>As respostas não ficam disponíveis para nova consulta. O vídeo desta aula continua liberado para revisão.</p></Card>;
  }
  return <Card className="academyQuiz"><div className="sectionTitle"><div><span className="eyebrow">Etapa validante</span><h3>Questionário da aula</h3><p>É necessário acertar pelo menos 75% para avançar.</p></div></div>{lesson.quizQuestions.map((question, questionIndex) => <fieldset className={wrongQuestionIndexes.includes(questionIndex) ? "academyQuestionWrong" : ""} key={questionIndex}><legend>{questionIndex + 1}. {question.question}</legend>{question.options.map((option, optionIndex) => <label className="checkOption" key={optionIndex}><input type="radio" name={`academy-answer-${lesson.id}-${questionIndex}`} checked={answers[questionIndex] === optionIndex} onChange={() => { setAnswers((current) => ({ ...current, [questionIndex]: optionIndex })); setWrongQuestionIndexes((current) => current.filter((index) => index !== questionIndex)); }} /> {option}</label>)}{wrongQuestionIndexes.includes(questionIndex) && <small className="academyWrongAnswerNotice">Esta questão foi respondida incorretamente. Revise sua resposta.</small>}</fieldset>)}<Button onClick={submitQuiz} disabled={lesson.quizQuestions.some((_, index) => answers[index] === undefined)}>Enviar respostas</Button></Card>;
}

function AcademyCourse({ navigate }) {
  const courseId = currentHashParams().get("id") || "";
  const [course, setCourse] = useState(null);
  const [selectedId, setSelectedId] = useState("");
  const [answers, setAnswers] = useState({});
  const [wrongQuestionIndexes, setWrongQuestionIndexes] = useState([]);
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [videoState, setVideoState] = useState({ position: 0, duration: 0, playing: false });
  const iframeRef = React.useRef(null);
  const playbackRef = React.useRef({ position: 0, duration: 0, playing: false, playbackRate: 1, watchedDelta: 0, samplePosition: 0 });
  const load = () => { if (!courseId) { setLoading(false); return Promise.resolve(); } setLoading(true); return academyRequest(`/api/academy/courses/${courseId}`).then((data) => { setCourse(data); setSelectedId((current) => { const currentLesson = data.lessons?.find((item) => item.id === current); if (currentLesson?.unlocked) return current; return data.lessons?.find((item) => item.unlocked && !item.completed)?.id || data.lessons?.find((item) => item.unlocked)?.id || ""; }); }).catch((error) => setMessage({ type: "error", text: error.message })).finally(() => setLoading(false)); };
  useEffect(() => { load(); }, [courseId]);
  const lessons = course?.lessons || [];
  const selectedIndex = lessons.findIndex((item) => item.id === selectedId);
  const selected = selectedIndex >= 0 ? lessons[selectedIndex] : null;
  const nextLesson = selectedIndex >= 0 ? lessons[selectedIndex + 1] : null;
  useEffect(() => {
    if (!selected || !course?.enrolled) return undefined;
    playbackRef.current = { position: Number(selected.lastPositionSeconds) || 0, duration: Number(selected.durationSeconds) || 0, playing: false, playbackRate: 1, watchedDelta: 0, samplePosition: Number(selected.lastPositionSeconds) || 0 };
    setVideoState({ position: playbackRef.current.position, duration: playbackRef.current.duration, playing: false });
    const postCommand = (func) => iframeRef.current?.contentWindow?.postMessage(JSON.stringify({ event: "command", func, args: [] }), "*");
    const persistProgress = async (ended = false) => {
      const snapshot = { ...playbackRef.current };
      if (!ended && snapshot.watchedDelta <= 0) return;
      if (ended && snapshot.duration > 0) snapshot.position = snapshot.duration;
      playbackRef.current.watchedDelta = 0;
      try {
        const result = await academyRequest(`/api/academy/lessons/${selected.id}/progress`, {
          method: "PUT",
          body: JSON.stringify({
            positionSeconds: Math.floor(snapshot.position),
            durationSeconds: Math.floor(snapshot.duration),
            watchedSecondsDelta: Math.floor(Math.min(30, snapshot.watchedDelta)),
            ended
          })
        });
        setCourse((current) => current ? { ...current, lessons: current.lessons.map((item) => item.id === selected.id ? { ...item, videoCompleted: result.videoCompleted, completed: result.completed } : item) } : current);
        if (result.videoCompleted && !selected.videoCompleted) setMessage({ type: "success", text: selected.requiresQuiz ? "Vídeo concluído. O questionário foi liberado." : "Aula concluída. A próxima aula foi liberada." });
        if (result.completed) await load();
      } catch (_) {
        playbackRef.current.watchedDelta += snapshot.watchedDelta;
      }
    };
    let nativeEndedSent = false;
    const sampleNativeVideo = () => {
      if (selected.videoSource !== "upload") return;
      const player = iframeRef.current;
      if (!player || typeof player.currentTime !== "number") return;
      const currentTime = Math.max(0, player.currentTime || 0);
      const duration = Math.max(0, player.duration || playbackRef.current.duration || 0);
      const elapsed = currentTime - playbackRef.current.samplePosition;
      const playing = !player.paused && !player.ended;
      if (playing && elapsed > 0 && elapsed <= 4) playbackRef.current.watchedDelta += elapsed;
      playbackRef.current.position = Math.floor(currentTime);
      playbackRef.current.samplePosition = currentTime;
      playbackRef.current.duration = Math.floor(duration);
      playbackRef.current.playing = playing;
      setVideoState({ position: playbackRef.current.position, duration: playbackRef.current.duration, playing });
      if (player.ended && !nativeEndedSent && duration > 0) {
        nativeEndedSent = true;
        playbackRef.current.position = Math.floor(duration);
        persistProgress(true);
      }
    };
    const onMessage = (event) => {
      if (event.source !== iframeRef.current?.contentWindow || !String(event.origin || "").includes("youtube.com")) return;
      let data = event.data;
      try { if (typeof data === "string") data = JSON.parse(data); } catch (_) { return; }
      const info = data?.info;
      if (info && typeof info === "object") {
        if (typeof info.currentTime === "number") {
          const currentTime = Math.max(0, info.currentTime);
          const elapsed = currentTime - playbackRef.current.samplePosition;
          const maximumExpected = Math.max(4, (playbackRef.current.playbackRate || 1) * 4);
          if (playbackRef.current.playing && elapsed > 0 && elapsed <= maximumExpected) playbackRef.current.watchedDelta += elapsed;
          playbackRef.current.samplePosition = currentTime;
          playbackRef.current.position = Math.floor(currentTime);
        }
        if (typeof info.duration === "number" && info.duration > 0) playbackRef.current.duration = Math.floor(info.duration);
        if (typeof info.playerState === "number") playbackRef.current.playing = info.playerState === 1;
        if (typeof info.playbackRate === "number" && info.playbackRate > 0) playbackRef.current.playbackRate = info.playbackRate;
        setVideoState({ position: playbackRef.current.position, duration: playbackRef.current.duration, playing: playbackRef.current.playing });
      }
      if (data?.event === "onStateChange" && typeof info === "number") {
        playbackRef.current.playing = info === 1;
        if (info === 0 && playbackRef.current.duration > 0) {
          playbackRef.current.position = playbackRef.current.duration;
          playbackRef.current.samplePosition = playbackRef.current.duration;
          persistProgress(true);
        }
        setVideoState({ position: playbackRef.current.position, duration: playbackRef.current.duration, playing: playbackRef.current.playing });
      }
    };
    const tick = window.setInterval(() => {
      if (selected.videoSource === "upload") sampleNativeVideo();
      else {
        postCommand("getCurrentTime");
        postCommand("getDuration");
        postCommand("getPlayerState");
        postCommand("getPlaybackRate");
      }
    }, 1000);
    const save = window.setInterval(() => { persistProgress(false); }, 10000);
    window.addEventListener("message", onMessage);
    return () => { window.clearInterval(tick); window.clearInterval(save); window.removeEventListener("message", onMessage); const snapshot = { ...playbackRef.current }; if (snapshot.watchedDelta > 0) fetch(`${API_BASE_URL}/api/academy/lessons/${selected.id}/progress`, { method: "PUT", credentials: "include", keepalive: true, headers: { "Content-Type": "application/json" }, body: JSON.stringify({ positionSeconds: snapshot.position, durationSeconds: snapshot.duration, watchedSecondsDelta: Math.floor(Math.min(30, snapshot.watchedDelta)) }) }).catch(() => {}); };
  }, [selected?.id, course?.enrolled]);
  const initializePlayer = () => {
    const connect = () => {
      const frame = iframeRef.current?.contentWindow;
      if (!frame || !selected) return;
      frame.postMessage(JSON.stringify({ event: "listening", id: selected.id }), "*");
      frame.postMessage(JSON.stringify({ event: "command", func: "addEventListener", args: ["onStateChange"] }), "*");
      frame.postMessage(JSON.stringify({ event: "command", func: "getDuration", args: [] }), "*");
    };
    connect();
    window.setTimeout(connect, 500);
    window.setTimeout(connect, 1500);
  };
  const enroll = async () => { try { await academyRequest(`/api/academy/courses/${courseId}/enroll`, { method: "POST" }); setMessage({ type: "success", text: "Curso iniciado. Você poderá continuar de onde parar." }); await load(); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  const submitQuiz = async () => { if (!selected) return; if (selected.quizQuestions.some((_, index) => answers[index] === undefined)) { setMessage({ type: "error", text: "Responda todas as questões antes de enviar." }); return; } try { const result = await academyRequest(`/api/academy/lessons/${selected.id}/assessment`, { method: "POST", body: JSON.stringify({ answers: selected.quizQuestions.map((_, index) => Number(answers[index])) }) }); if (result.approved) { setAnswers({}); setWrongQuestionIndexes([]); setMessage({ type: "success", text: "Aprovado com pelo menos 75%. A próxima aula foi liberada." }); await load(); } else { setWrongQuestionIndexes(Array.isArray(result.wrongQuestionIndexes) ? result.wrongQuestionIndexes : []); setMessage({ type: "error", text: `Você obteve ${Math.round(result.score || 0)}%. As questões destacadas foram respondidas incorretamente. Revise-as e tente novamente.` }); } } catch (error) { setMessage({ type: "error", text: error.message }); } };
  const goNext = () => { if (!nextLesson) return; setAnswers({}); setWrongQuestionIndexes([]); setMessage(null); setSelectedId(nextLesson.id); window.scrollTo({ top: 0, behavior: "smooth" }); };
  if (loading) return <Page label="Academia LicitaHub" title="Curso"><Card><p>Carregando curso...</p></Card></Page>;
  if (!course) return <Page label="Academia LicitaHub" title="Curso"><Card className="dangerNotice"><p>{message?.text || "Curso não encontrado."}</p></Card></Page>;
  if (!course.enrolled) return <Page label="Academia LicitaHub" title={course.title} actions={<Button variant="secondary" onClick={() => navigate("academy-home")}>Voltar ao catálogo</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <section className="academyCourseSummary"><div className="academyCourseSummaryCover">{course.coverImageUrl ? <img src={course.coverImageUrl} alt={`Capa de ${course.title}`} /> : <strong>{course.title}</strong>}</div><div className="academyCourseSummaryInfo"><span className="academyTag">{course.category}</span><h2>{course.title}</h2><p>{course.description || "Curso da Academia LicitaHub."}</p><div className="academyCourseFacts"><span>{lessons.length} aula(s)</span><span>{course.workloadHours || 0}h de carga horária</span><span>Certificado ao concluir</span></div><Button onClick={enroll}>Iniciar curso</Button></div></section>
    <Card><h3>Conteúdo do curso</h3><div className="academyCourseOutline">{lessons.map((item, index) => <div key={item.id}><span>{String(index + 1).padStart(2, "0")}</span><div><strong>{item.title}</strong><small>{item.requiresQuiz ? "Vídeo e questionário validante" : "Videoaula"}</small></div></div>)}</div></Card>
  </Page>;
  const percent = videoState.duration > 0 ? Math.min(100, Math.round(videoState.position * 100 / videoState.duration)) : 0;
  return <Page label="Academia LicitaHub" title={course.title} actions={<Button variant="secondary" onClick={() => navigate("academy-my-learning")}>Meus cursos</Button>}>
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <div className="academyClassroom"><aside className="academyClassroomRail"><div className="academyClassroomCourse"><span className="academyTag">{course.category}</span><strong>{course.title}</strong></div>{lessons.map((item, index) => <button key={item.id} className={`academyClassroomLesson ${selectedId === item.id ? "active" : ""} ${item.completed ? "completed" : ""}`} disabled={!item.unlocked} onClick={() => { setAnswers({}); setWrongQuestionIndexes([]); setMessage(null); setSelectedId(item.id); }}><div className="academyClassroomThumb">{youtubeThumbnail(item.videoUrl) ? <img src={youtubeThumbnail(item.videoUrl)} alt="" /> : <span>{index + 1}</span>}</div><div><strong>Aula {index + 1}: {item.title}</strong><small>{item.completed ? "Concluída" : item.videoCompleted && item.requiresQuiz ? "Questionário liberado" : item.unlocked ? "Disponível" : "Bloqueada"}</small></div></button>)}</aside><main className="academyClassroomMain">{selected ? <><Card className="academyPlayerCard"><div className="academyPlayerHeading"><div><span className="eyebrow">Aula {selectedIndex + 1} de {lessons.length}</span><h3>{selected.title}</h3><p>{selected.description}</p></div>{selected.completed && <span className="statusPill open">Concluída</span>}</div><div className="academyVideo"><AcademyAdaptivePlayer lesson={selected} playerRef={iframeRef} initializePlayer={initializePlayer} /></div><div className="academyWatchingProgress"><div><span style={{ width: `${selected.videoCompleted ? 100 : percent}%` }} /></div><small>{selected.videoCompleted ? "Vídeo concluído" : `${formatAcademyTime(videoState.position)} de ${formatAcademyTime(videoState.duration)} · assista até o final`}</small></div></Card>{selected.requiresQuiz ? selected.videoCompleted ? <AcademyQuizPanel lesson={selected} answers={answers} wrongQuestionIndexes={wrongQuestionIndexes} setAnswers={setAnswers} setWrongQuestionIndexes={setWrongQuestionIndexes} submitQuiz={submitQuiz} /> : <Card className="academyLockedQuiz"><strong>Questionário bloqueado</strong><p>Assista à aula até o final para liberar as questões.</p></Card> : null}{selected.completed && nextLesson && <div className="academyNextLesson"><div><span className="eyebrow">Próxima etapa</span><strong>{nextLesson.title}</strong></div><Button onClick={goNext}>Próxima aula</Button></div>}{selected.completed && !nextLesson && <Card className="academyCourseCompleted"><h3>Curso concluído</h3><p>Parabéns. Todas as aulas e avaliações foram concluídas. O certificado oficial está disponível abaixo.</p></Card>}</> : <Card><p>Nenhuma aula disponível.</p></Card>}</main></div>
    {course.certificateCode && <section className="academyCertificate"><span>Certificado oficial liberado</span><h2>{course.title}</h2><p>Código verificável: <strong>{course.certificateCode}</strong></p><div className="formActions"><a className="certificateButton" href={academyCertificateUrl(course.id)}>Baixar PDF</a><a className="linkButton" href={academyCertificateVerificationUrl(course.certificateCode)} target="_blank" rel="noreferrer">Validar autenticidade</a></div></section>}
  </Page>;
}

function shuffleCompanyRatings(items) {
  const next = [...items];
  for (let index = next.length - 1; index > 0; index -= 1) {
    const target = Math.floor(Math.random() * (index + 1));
    [next[index], next[target]] = [next[target], next[index]];
  }
  return next;
}

function companyRatingRequest(path, options = {}) {
  return fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    ...options,
    headers: { "Content-Type": "application/json", ...(options.headers || {}) }
  }).then(async (response) => {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Não foi possível concluir esta ação.");
    return data;
  });
}

function formatRatingDeadline(value) {
  if (!value) return "Prazo não informado";
  return new Date(value).toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" });
}

function minimumRatingDeadline() {
  const date = new Date(Date.now() + 5 * 60000);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function CompanyRatingModal({ company, maxStars, onClose, onSave }) {
  const [stars, setStars] = useState(company?.allocatedStars || 0);
  if (!company) return null;
  const maximum = Math.max(0, Number(maxStars || 0));
  const changeStars = (next) => setStars(Math.min(maximum, Math.max(0, Number(next) || 0)));
  return <div className="ratingModalBackdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
    <section className="ratingModal" role="dialog" aria-modal="true" aria-labelledby="rating-company-title">
      <button type="button" className="ratingModalClose" title="Fechar" aria-label="Fechar" onClick={onClose}>×</button>
      <LogoSlot src={company.logoUrl} initials={company.name?.slice(0, 2).toUpperCase()} size="xl" label={`Logo de ${company.name}`} />
      <span className="eyebrow">Percepção de parceria</span>
      <h3 id="rating-company-title">{company.name}</h3>
      <p>Avalie a abertura desta empresa para realizar negócios, parcerias e composições empresariais.</p>
      <div className="ratingAmount">
        <span>Quantidade de estrelas</span>
        <div>
          <button type="button" onClick={() => changeStars(stars - 1)} disabled={stars <= 0} title="Retirar uma estrela" aria-label="Retirar uma estrela">−</button>
          <input type="number" min="0" max={maximum} value={stars} onChange={(event) => changeStars(event.target.value)} aria-label="Quantidade de estrelas para esta empresa" />
          <button type="button" onClick={() => changeStars(stars + 1)} disabled={stars >= maximum} title="Adicionar uma estrela" aria-label="Adicionar uma estrela">+</button>
        </div>
        <small>Você pode destinar até {maximum} estrela(s) para esta empresa neste momento.</small>
      </div>
      <div className="formActions">
        {company.allocatedStars > 0 && <Button variant="secondary" onClick={() => onSave(0)}>Retirar estrelas</Button>}
        <Button onClick={() => onSave(stars)} disabled={!stars}>Confirmar avaliação</Button>
      </div>
      <small className="ratingPrivacy">Sua identidade não será exibida à empresa avaliada nem nos resultados.</small>
    </section>
  </div>;
}

function CompanyRatingConfirmation({ allocations, total, onClose, onConfirm, submitting }) {
  return <div className="ratingModalBackdrop" onMouseDown={(event) => { if (event.target === event.currentTarget && !submitting) onClose(); }}>
    <section className="ratingConfirmModal" role="dialog" aria-modal="true" aria-labelledby="rating-confirm-title">
      <button type="button" className="ratingModalClose" title="Fechar" aria-label="Fechar" onClick={onClose} disabled={submitting}>×</button>
      <span className="eyebrow">Revisão final</span>
      <h3 id="rating-confirm-title">Confirme sua distribuição</h3>
      <p>Confira as empresas e as quantidades antes do envio. Depois de confirmar, a avaliação não poderá ser alterada.</p>
      <div className="ratingConfirmationList">{allocations.map((company) => <div key={company.id}>
        <LogoSlot src={company.logoUrl} initials={company.name?.slice(0, 2).toUpperCase()} size="sm" label={`Logo de ${company.name}`} />
        <strong>{company.name}</strong>
        <span>★ {company.allocatedStars}</span>
      </div>)}</div>
      <div className="ratingConfirmationTotal"><span>Total distribuído</span><strong>★ {total}</strong></div>
      <div className="formActions"><Button variant="secondary" onClick={onClose} disabled={submitting}>Voltar e ajustar</Button><Button onClick={onConfirm} disabled={submitting}>{submitting ? "Enviando..." : "Confirmar envio"}</Button></div>
    </section>
  </div>;
}

function ratingTrend(points) {
  if (points.length < 3) return [];
  const values = points.map((point) => Number(point.relativeIndex || 0));
  const xMean = (values.length - 1) / 2;
  const yMean = values.reduce((sum, value) => sum + value, 0) / values.length;
  const numerator = values.reduce((sum, value, index) => sum + (index - xMean) * (value - yMean), 0);
  const denominator = values.reduce((sum, _, index) => sum + (index - xMean) ** 2, 0);
  const slope = denominator ? numerator / denominator : 0;
  return values.map((_, index) => yMean + slope * (index - xMean));
}

function RatingEvolutionChart({ series }) {
  const closed = series.filter((item) => item.status === "closed");
  if (!closed.length) return <div className="ratingChartEmpty"><strong>A evolução começará após o fechamento da primeira rodada.</strong><p>A rodada atual não entra na tendência enquanto ainda recebe avaliações.</p></div>;
  const width = 900; const height = 320; const left = 56; const right = 24; const top = 24; const bottom = 58;
  const trend = ratingTrend(closed);
  const maxValue = Math.max(120, ...closed.map((item) => Number(item.relativeIndex || 0)), ...trend);
  const roundedMax = Math.ceil(maxValue / 20) * 20;
  const x = (index) => closed.length === 1 ? width / 2 : left + index * (width - left - right) / (closed.length - 1);
  const y = (value) => top + (roundedMax - value) * (height - top - bottom) / roundedMax;
  const points = closed.map((item, index) => `${x(index)},${y(Number(item.relativeIndex || 0))}`).join(" ");
  const trendPoints = trend.map((value, index) => `${x(index)},${y(value)}`).join(" ");
  const guides = [0, 50, 100, roundedMax].filter((value, index, values) => values.indexOf(value) === index && value <= roundedMax);
  return <div className="ratingChartWrap">
    <svg className="ratingChart" viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Evolução da percepção de parceria">
      {guides.map((value) => <g key={value}><line x1={left} y1={y(value)} x2={width - right} y2={y(value)} className={value === 100 ? "marketGuide" : "chartGuide"} /><text x={left - 10} y={y(value) + 4} textAnchor="end">{value}</text></g>)}
      {closed.length > 1 && <polyline points={points} className="ratingActualLine" />}
      {trend.length > 0 && <polyline points={trendPoints} className="ratingTrendLine" />}
      {closed.map((item, index) => <g key={item.id}>
        <circle cx={x(index)} cy={y(Number(item.relativeIndex || 0))} r="6" className="ratingPoint"><title>{`${item.title}: ${Number(item.relativeIndex || 0).toLocaleString("pt-BR")} pontos`}</title></circle>
        <text x={x(index)} y={height - 24} textAnchor="middle">{item.title.length > 16 ? `${item.title.slice(0, 14)}…` : item.title}</text>
      </g>)}
    </svg>
    <div className="ratingChartLegend"><span className="actual">Sua empresa</span><span className="market">Média do mercado (100)</span>{trend.length > 0 && <span className="trend">Tendência</span>}</div>
  </div>;
}

function CompanyRatingResults({ results }) {
  const [scope, setScope] = useState("all");
  const [year, setYear] = useState("all");
  const [sessionId, setSessionId] = useState("all");
  const allSeries = Array.isArray(results?.series) ? results.series : [];
  const years = [...new Set(allSeries.map((item) => new Date(item.openedAt).getFullYear()).filter(Boolean))].sort((a, b) => b - a);
  let filtered = allSeries.filter((item) => (year === "all" || String(new Date(item.openedAt).getFullYear()) === year) && (sessionId === "all" || item.id === sessionId));
  if (scope === "last5") filtered = filtered.slice(-5);
  if (scope === "last10") filtered = filtered.slice(-10);
  const latest = results?.latest;
  const history = results?.history || {};
  const relativeText = (value) => Number(value || 0) >= 100 ? "acima ou na média do mercado" : "abaixo da média do mercado";
  return <div className="ratingResults">
    <section className="ratingResultBand">
      <div><span className="eyebrow">{latest?.status === "open" ? "Resultado parcial da rodada" : "Última rodada concluída"}</span><h3>{latest?.title || "Ainda não há resultado"}</h3><strong>{latest ? `${latest.receivedStars} estrela(s)` : "Aguardando avaliações"}</strong><p>{latest ? `Posição ${latest.position} · média atual ${Number(latest.marketAverage || 0).toLocaleString("pt-BR")} · índice ${Number(latest.relativeIndex || 0).toLocaleString("pt-BR")} · ${relativeText(latest.relativeIndex)}` : "Os resultados aparecerão depois do primeiro envio da empresa."}</p></div>
      <div><span className="eyebrow">Média geral das rodadas</span><h3>Desempenho médio histórico</h3><strong>Índice {Number(history.relativeIndex || 0).toLocaleString("pt-BR")}</strong><p>Média dos índices de {history.sessionCount || 0} rodada(s) encerrada(s). Em cada uma, 100 representa a média do mercado.</p></div>
    </section>
    <Card className="ratingHistoryCard">
      <div className="ratingHistoryHeading"><div><span className="eyebrow">Evolução e tendência</span><h3>Percepção da sua empresa ao longo do tempo</h3></div><div className="ratingResultFilters">
        <select value={scope} onChange={(event) => setScope(event.target.value)} aria-label="Período"><option value="all">Todas as rodadas</option><option value="last5">Últimas 5</option><option value="last10">Últimas 10</option></select>
        <select value={year} onChange={(event) => setYear(event.target.value)} aria-label="Ano"><option value="all">Todos os anos</option>{years.map((item) => <option key={item} value={String(item)}>{item}</option>)}</select>
        <select value={sessionId} onChange={(event) => setSessionId(event.target.value)} aria-label="Rodada"><option value="all">Todas as sessões</option>{allSeries.map((item) => <option key={item.id} value={item.id}>{item.title}</option>)}</select>
      </div></div>
      <RatingEvolutionChart series={filtered} />
    </Card>
  </div>;
}

function CompanyRating() {
  const [state, setState] = useState(null);
  const [results, setResults] = useState(null);
  const [companies, setCompanies] = useState([]);
  const [selected, setSelected] = useState(null);
  const [confirming, setConfirming] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const load = async () => {
    setLoading(true);
    try {
      const [nextState, nextResults] = await Promise.all([
        companyRatingRequest("/api/company-ratings"),
        companyRatingRequest("/api/company-rating-results")
      ]);
      setState(nextState);
      setResults(nextResults);
      setCompanies((current) => {
        const incoming = Array.isArray(nextState.companies) ? nextState.companies : [];
        if (!current.length || current.map((item) => item.id).sort().join(",") !== incoming.map((item) => item.id).sort().join(",")) return shuffleCompanyRatings(incoming);
        return current.map((item) => ({ ...item, allocatedStars: incoming.find((entry) => entry.id === item.id)?.allocatedStars || 0 }));
      });
      return nextState;
    } catch (error) {
      setMessage({ type: "error", text: error.message });
      return null;
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);
  const budget = Number(state?.round?.starBudget || 0);
  const used = Number(state?.usedStars || 0);
  const remaining = Math.max(0, budget - used);
  const allocations = companies.filter((company) => !company.isOwn && Number(company.allocatedStars || 0) > 0).sort((left, right) => left.name.localeCompare(right.name, "pt-BR"));
  const saveAllocation = async (stars) => {
    if (!selected) return;
    try {
      await companyRatingRequest(`/api/company-ratings/allocations/${selected.id}`, { method: "PUT", body: JSON.stringify({ stars }) });
      setSelected(null);
      setMessage({ type: "success", text: stars ? "Avaliação reservada. Continue até distribuir todas as estrelas." : "Estrelas retiradas desta empresa." });
      await load();
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    }
  };
  const submit = async () => {
    if (remaining) return;
    setConfirming(true);
  };
  const confirmSubmission = async () => {
    setSubmitting(true);
    try {
      await companyRatingRequest("/api/company-ratings/submit", { method: "POST", body: "{}" });
      setConfirming(false);
      setMessage({ type: "success", text: "Avaliação enviada. Sua distribuição ficou anônima e não poderá ser alterada." });
      await load();
    } catch (error) {
      const refreshedState = await load();
      if (refreshedState?.submitted) {
        setConfirming(false);
        setMessage({ type: "success", text: "Avaliação enviada. Sua distribuição ficou anônima e não poderá ser alterada." });
      } else {
        setMessage({ type: "error", text: error.message });
      }
    } finally {
      setSubmitting(false);
    }
  };
  if (loading && !state) return <Page label="Comunidade" title="Avaliação de parcerias"><Card><p>Preparando a rodada...</p></Card></Page>;
  const hasOpenVote = state?.round && state.included && !state.submitted;
  return <Page label="Comunidade" title="Avaliação de parcerias">
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <section className="ratingPurpose"><span className="eyebrow">Critério desta avaliação</span><h3>Abertura para negócios e parcerias</h3><p>A percepção sobre a abertura da empresa avaliada para realizar negócios, parcerias e composições empresariais.</p><small>As avaliações são anônimas. Nenhuma empresa verá quem lhe atribuiu estrelas.</small></section>
    {!state?.round && <><Card><h3>Nenhuma rodada aberta</h3><p>Aguarde a próxima liberação do administrador da LicitaHub. Seus resultados históricos continuam disponíveis abaixo.</p></Card><CompanyRatingResults results={results} /></>}
    {state?.round && !state.included && <Card><h3>Empresa fora desta rodada</h3><p>Esta sessão considera o conjunto de associadas ativo no momento em que foi aberta.</p></Card>}
    {hasOpenVote && !state.canManage && <><Card><h3>Aguardando o administrador da sua empresa</h3><p>A rodada <strong>{state.round.title}</strong> está aberta. A distribuição é única por empresa e deve ser concluída pelo administrador da empresa.</p></Card>{results?.latest && <CompanyRatingResults results={results} />}</>}
    {hasOpenVote && state.canManage && <section className="companyRatingStage">
      <header className="ratingRoundStrip"><div><span>Rodada atual</span><strong>{state.round.title}</strong></div><small>{state.round.companyCount} empresas · prazo {formatRatingDeadline(state.round.closesAt)}</small></header>
      <div className="companyLogoBackdrop">{companies.map((company) => <button type="button" key={company.id} className={`companyRatingTile ${company.isOwn ? "isOwn" : ""} ${company.allocatedStars ? "hasRating" : ""}`} disabled={company.isOwn} onClick={() => setSelected(company)}>
        <LogoSlot src={company.logoUrl} initials={company.name?.slice(0, 2).toUpperCase()} size="md" label={`Logo de ${company.name}`} />
        <strong>{company.name}</strong>
        {company.isOwn ? <small>Sua empresa</small> : company.allocatedStars > 0 ? <span className="ratingAllocatedBadge">★ {company.allocatedStars}</span> : <small>Avaliar</small>}
      </button>)}</div>
      <footer className="ratingStageFooter"><strong>Distribua as {budget} estrelas integralmente.</strong><span>Você decide livremente quantas estrelas cada empresa receberá.</span></footer>
      <aside className={`ratingSubmissionDock ${remaining === 0 ? "ready" : ""}`}>
        <button type="button" onClick={submit} disabled={remaining !== 0}>{remaining > 0 ? `${remaining} estrela${remaining === 1 ? "" : "s"} restante${remaining === 1 ? "" : "s"}` : "Concluir distribuição"}</button>
        <div aria-label={`${used} de ${budget} estrelas distribuídas`}><span style={{ width: `${budget ? used * 100 / budget : 0}%` }} /></div>
        <small>{used} de {budget} distribuídas</small>
      </aside>
    </section>}
    {state?.submitted && <><Card className="successNotice"><h3>Avaliação concluída pela sua empresa</h3><p>O painel de distribuição foi fechado e o resultado parcial já está disponível. A rodada encerra automaticamente quando todas as empresas participantes enviarem suas avaliações.</p></Card><CompanyRatingResults results={results} /></>}
    <CompanyRatingModal company={selected} maxStars={remaining + Number(selected?.allocatedStars || 0)} onClose={() => setSelected(null)} onSave={saveAllocation} />
    {confirming && <CompanyRatingConfirmation allocations={allocations} total={used} onClose={() => setConfirming(false)} onConfirm={confirmSubmission} submitting={submitting} />}
  </Page>;
}

function RatingRankingTable({ items, emptyText, average = false }) {
  if (!items?.length) return <p>{emptyText}</p>;
  return <div className="ratingRanking">{items.map((item, index) => <div key={item.id}><span className="ratingRank">{index + 1}</span><strong>{item.name}</strong><span>{average ? `${Number(item.stars || 0).toLocaleString("pt-BR")} média de estrelas` : `${item.stars} estrela(s)`}</span><b>Índice {Number(item.relativeIndex || 0).toLocaleString("pt-BR")}</b></div>)}</div>;
}

function CompanyRatingAdmin() {
  const [rounds, setRounds] = useState([]);
  const [selectedId, setSelectedId] = useState("");
  const [details, setDetails] = useState(null);
  const [title, setTitle] = useState("");
  const [deadline, setDeadline] = useState("");
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState("");
  const [filters, setFilters] = useState({ search: "", status: "all", year: "all" });
  const [closedExpanded, setClosedExpanded] = useState(false);
  const load = async (preferredId = "", useCurrentSelection = true) => {
    setLoading(true);
    try {
      const list = await companyRatingRequest("/api/admin/company-rating-rounds");
      const items = Array.isArray(list) ? list : [];
      setRounds(items);
      const nextId = preferredId || (useCurrentSelection ? selectedId : "") || items[0]?.id || "";
      setSelectedId(nextId);
      setClosedExpanded(items.find((item) => item.id === nextId)?.status === "closed");
      const result = await companyRatingRequest(`/api/admin/company-rating-results${nextId ? `?roundId=${encodeURIComponent(nextId)}` : ""}`);
      setDetails(result);
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);
  const selectRound = async (id) => { setSelectedId(id); setClosedExpanded(rounds.find((item) => item.id === id)?.status === "closed"); setLoading(true); try { setDetails(await companyRatingRequest(`/api/admin/company-rating-results?roundId=${encodeURIComponent(id)}`)); } catch (error) { setMessage({ type: "error", text: error.message }); } finally { setLoading(false); } };
  const openRound = async (event) => {
    event.preventDefault();
    if (!deadline) {
      setMessage({ type: "error", text: "Informe o prazo de encerramento da rodada." });
      return;
    }
    try {
      const created = await companyRatingRequest("/api/admin/company-rating-rounds", { method: "POST", body: JSON.stringify({ title, closesAt: new Date(deadline).toISOString() }) });
      setTitle("");
      setDeadline("");
      setMessage({ type: "success", text: `Rodada aberta para ${created.companyCount} empresas, com ${created.starBudget} estrelas por empresa e encerramento em ${formatRatingDeadline(created.closesAt)}.` });
      await load(created.id);
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    }
  };
  const deleteRound = async () => {
    if (!current || current.status !== "closed") return;
    const confirmed = window.confirm(`Excluir definitivamente a rodada "${current.title}"?\n\nTodas as avaliações e resultados desta rodada serão removidos, e a média histórica será recalculada. Esta ação não pode ser desfeita.`);
    if (!confirmed) return;
    setDeletingId(current.id);
    try {
      await companyRatingRequest(`/api/admin/company-rating-rounds?roundId=${encodeURIComponent(current.id)}`, { method: "DELETE" });
      setMessage({ type: "success", text: `A rodada "${current.title}" foi excluída e o histórico foi recalculado.` });
      setSelectedId("");
      setDetails(null);
      await load("", false);
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setDeletingId("");
    }
  };
  const openRoundItem = rounds.find((item) => item.status === "open");
  const availableYears = [...new Set(rounds.map((item) => new Date(item.openedAt).getFullYear()).filter(Boolean))].sort((left, right) => right - left);
  const filteredRounds = rounds.filter((item) => {
    const term = filters.search.trim().toLowerCase();
    const matchesSearch = !term || String(item.title || "").toLowerCase().includes(term);
    const matchesStatus = filters.status === "all" || item.status === filters.status;
    const matchesYear = filters.year === "all" || String(new Date(item.openedAt).getFullYear()) === filters.year;
    return matchesSearch && matchesStatus && matchesYear;
  });
  const openRounds = filteredRounds.filter((item) => item.status === "open");
  const closedRounds = filteredRounds.filter((item) => item.status === "closed");
  const current = details?.round;
  const submitted = current ? details.participation?.filter((item) => item.submitted).length || 0 : 0;
  const progress = current?.companyCount ? submitted * 100 / current.companyCount : 0;
  return <Page label="Acesso e administração" title="Avaliação das associadas">
    {message && <Card className={message.type === "error" ? "dangerNotice" : "successNotice"}><p>{message.text}</p></Card>}
    <section className="ratingAdminIntro"><div><span className="eyebrow">Rodadas anônimas</span><h3>Percepção de abertura para negócios e parcerias</h3><p>Abra uma sessão para todas as associadas ativas. Ela encerra quando todas concluírem ou automaticamente no prazo definido, mesmo que existam empresas pendentes.</p></div>{!openRoundItem && <form onSubmit={openRound}><Field label="Nome da nova rodada"><input value={title} onChange={(event) => setTitle(event.target.value)} maxLength={180} placeholder="Ex.: Percepção de parcerias · Julho 2026" required /></Field><Field label="Prazo de encerramento"><input type="datetime-local" value={deadline} min={minimumRatingDeadline()} onChange={(event) => setDeadline(event.target.value)} required /></Field><Button type="submit">Abrir rodada</Button></form>}</section>
    <Card className="compactFilters ratingAdminFilters"><FormGrid>
      <Field label="Buscar rodada"><input value={filters.search} onChange={(event) => setFilters((currentFilters) => ({ ...currentFilters, search: event.target.value }))} placeholder="Digite o nome da rodada" /></Field>
      <Field label="Situação"><select value={filters.status} onChange={(event) => { const status = event.target.value; setFilters((currentFilters) => ({ ...currentFilters, status })); if (status === "closed") setClosedExpanded(true); }}><option value="all">Todas</option><option value="open">Abertas</option><option value="closed">Encerradas</option></select></Field>
      <Field label="Ano"><select value={filters.year} onChange={(event) => setFilters((currentFilters) => ({ ...currentFilters, year: event.target.value }))}><option value="all">Todos os anos</option>{availableYears.map((year) => <option key={year} value={String(year)}>{year}</option>)}</select></Field>
    </FormGrid></Card>
    <div className="ratingAdminLayout">
      <aside className="ratingRoundList">
        <span className="eyebrow">Sessões</span>
        {filters.status !== "closed" && openRounds.length > 0 && <div className="ratingOpenRounds"><small>Em andamento</small>{openRounds.map((item) => <button type="button" key={item.id} className={selectedId === item.id ? "active" : ""} onClick={() => selectRound(item.id)}><strong>{item.title}</strong><small>{item.submittedCount}/{item.companyCount} concluíram · até {formatRatingDeadline(item.closesAt)}</small></button>)}</div>}
        {filters.status !== "open" && <details className="ratingClosedRounds" open={closedExpanded} onToggle={(event) => setClosedExpanded(event.currentTarget.open)}>
          <summary><span>Rodadas encerradas</span><strong>{closedRounds.length}</strong></summary>
          <div>{closedRounds.map((item) => <button type="button" key={item.id} className={selectedId === item.id ? "active" : ""} onClick={() => selectRound(item.id)}><strong>{item.title}</strong><small>{new Date(item.openedAt).toLocaleDateString("pt-BR")} · {item.companyCount} empresas</small></button>)}</div>
          {closedRounds.length === 0 && <p>Nenhuma rodada encerrada corresponde aos filtros.</p>}
        </details>}
        {filteredRounds.length === 0 && <p>Nenhuma rodada encontrada.</p>}
      </aside>
      <main className="ratingAdminMain">
        {loading && <Card><p>Carregando acompanhamento...</p></Card>}
        {!loading && current && <><Card className="ratingAdminStatus"><div><span className={`statusPill ${current.status === "open" ? "open" : ""}`}>{current.status === "open" ? "Rodada aberta" : "Rodada encerrada"}</span><h3>{current.title}</h3><p>{current.companyCount} empresas · {current.starBudget} estrelas por distribuição · prazo {formatRatingDeadline(current.closesAt)}</p></div><div className="ratingAdminStatusActions"><div className="ratingAdminProgress"><strong>{submitted}/{current.companyCount}</strong><span>empresas concluíram</span><div><span style={{ width: `${progress}%` }} /></div></div>{current.status === "closed" && <Button variant="danger" onClick={deleteRound} disabled={deletingId === current.id}>{deletingId === current.id ? "Excluindo..." : "Excluir rodada"}</Button>}</div></Card>
        {current.status === "open" ? <Card><div className="ratingParticipationHeading"><div><h3>Acompanhamento da participação</h3><p>O ranking fica oculto até o fechamento automático da rodada.</p></div><strong>{Math.round(progress)}%</strong></div><div className="ratingParticipationList">{details.participation?.map((item) => <div key={item.id}><span className={`ratingParticipationDot ${item.submitted ? "done" : ""}`} /><strong>{item.name}</strong><small>{item.submitted ? "Distribuição concluída" : "Pendente"}</small></div>)}</div></Card> : <div className="ratingRankingSections">
          <details className="ratingRankingDetails">
            <summary><div><span className="eyebrow">Resultado da sessão</span><strong>Ranking desta rodada</strong></div><span>{details.sessionRanking?.length || 0} empresas</span></summary>
            <div className="ratingRankingDetailsBody"><p>Classificação calculada apenas com as distribuições confirmadas até o encerramento.</p><RatingRankingTable items={details.sessionRanking} emptyText="Sem resultados nesta sessão." /></div>
          </details>
          <details className="ratingRankingDetails">
            <summary><div><span className="eyebrow">Histórico</span><strong>Média geral das rodadas</strong></div><span>{details.historicalRanking?.length || 0} empresas</span></summary>
            <div className="ratingRankingDetailsBody"><p>Média dos índices obtidos por cada empresa nas rodadas encerradas. Cada rodada possui o mesmo peso.</p><RatingRankingTable items={details.historicalRanking} average emptyText="O histórico ainda não possui resultados." /></div>
          </details>
        </div>}</>}
      </main>
    </div>
  </Page>;
}

function Screen({ screen, navigate, userStatuses, openUserAction, selectedUserAction, updateUserStatus, selectedUserProfile, openUserProfile, selectedPublicationId, openPublicationManager, selectedTenderId, openTenderInterestCompanies, selectedNews, openNewsDetail, refreshSession, sessionUser, openChatForAd, openChatForTask, openChatForUser }) {
  const currentRole = frontendRole(sessionUser?.roleKey);
  if (!canAccessScreen(screen, currentRole)) {
    return <AccessDenied navigate={navigate} role={currentRole} />;
  }

  const screens = {
    "admin-dashboard": <AdminDashboard navigate={navigate} />,
    "invite-new": <InviteNew />,
    "invite-list": <InviteList navigate={navigate} />,
    "company-manage": <CompanyManage />,
	"password-reset-management": <PasswordResetManagement />,
    "company-rating-admin": <CompanyRatingAdmin />,
    "invite-accept": <InviteAccept navigate={navigate} />,
    "company-review": <CompanyReview />,
    "my-profile": <MyProfile refreshSession={refreshSession} sessionUser={sessionUser} />,
    "my-assembly-tasks": <MyAssemblyTasks navigate={navigate} openChatForTask={openChatForTask} />,
    "notification-history": <NotificationHistory navigate={navigate} openPublicationManager={openPublicationManager} openTenderInterestCompanies={openTenderInterestCompanies} />,
    "company-profile-edit": <CompanyProfileEdit refreshSession={refreshSession} />,
    "company-users": <CompanyUsers navigate={navigate} openUserAction={openUserAction} openUserProfile={openUserProfile} sessionUser={sessionUser} />,
	"technical-professionals": <TechnicalProfessionals sessionUser={sessionUser} />,
	"technical-certificates": <TechnicalCertificates sessionUser={sessionUser} navigate={navigate} />,
	"technical-certificate-ai": <TechnicalCertificateAI navigate={navigate} />,
    "company-user-profile": <CompanyUserProfile selectedUserProfile={selectedUserProfile} navigate={navigate} />,
    "company-user-block": <CompanyUserAccessConfirm navigate={navigate} selectedUserAction={selectedUserAction} updateUserStatus={updateUserStatus} mode="block" />,
    "company-user-unblock": <CompanyUserAccessConfirm navigate={navigate} selectedUserAction={selectedUserAction} updateUserStatus={updateUserStatus} mode="unblock" />,
    "company-user-delete": <CompanyUserDelete navigate={navigate} selectedUserAction={selectedUserAction} />,
    "community-home": <CommunityHome sessionUser={sessionUser} navigate={navigate} />,
    "company-rating": <CompanyRating />,
    "company-public-profile": <CompanyPublicProfile navigate={navigate} openPublicationManager={openPublicationManager} sessionUser={sessionUser} openChatForUser={openChatForUser} />,
    "publication-new": <PublicationNew openPublicationManager={openPublicationManager} navigate={navigate} />,
    "publication-list": <PublicationList selectedPublicationId={selectedPublicationId} />,
    "radar-home": <RadarHomeConnected navigate={navigate} openNewsDetail={openNewsDetail} />,
    "radar-detail": <RadarDetailConnected selectedNews={selectedNews} navigate={navigate} />,
    "radar-new": <RadarNewConnected navigate={navigate} />,
    "radar-manage": <RadarManage />,
    "academy-home": <AcademyHome navigate={navigate} />,
    "academy-my-learning": <AcademyMyLearning navigate={navigate} />,
    "academy-admin": <AcademyAdmin navigate={navigate} />,
    "academy-course-admin": <AcademyCourseAdmin navigate={navigate} />,
    "academy-course": <AcademyCourse navigate={navigate} />,
    "tender-admin": <TenderAdmin navigate={navigate} />,
    "pncp-capture": <PNCPCapture navigate={navigate} />,
    "tender-new": <TenderNew navigate={navigate} />,
    "tender-list": <TenderList navigate={navigate} openTenderInterestCompanies={openTenderInterestCompanies} />,
	"assembly-center": <AssemblyCenter navigate={navigate} />,
	"tender-calendar": <AssemblyCalendar navigate={navigate} />,
    "tender-detail": <TenderDetail navigate={navigate} sessionUser={sessionUser} openTenderInterestCompanies={openTenderInterestCompanies} />,
    "tender-challenge": <TenderChallenge navigate={navigate} />,
	"tender-challenge-board": <TenderChallengeBoard navigate={navigate} />,
    "tender-interest": <TenderInterest navigate={navigate} />,
    "tender-interest-list": <TenderInterestList navigate={navigate} selectedTenderId={selectedTenderId} sessionUser={sessionUser} openChatForAd={openChatForAd} />,
    "match-partners": <MatchPartners navigate={navigate} sessionUser={sessionUser} openChatForAd={openChatForAd} />,
    "match-tinder": <MatchTinder navigate={navigate} sessionUser={sessionUser} />,
    "match-profile": <MatchProfile navigate={navigate} sessionUser={sessionUser} openChatForAd={openChatForAd} />,
    "match-success": <MatchSuccess />,
    "match-list": <MatchList sessionUser={sessionUser} navigate={navigate} />,
    "assembly-board": <AssemblyBoard sessionUser={sessionUser} navigate={navigate} openChatForTask={openChatForTask} />
  };

  return screens[screen] || <AccessDenied navigate={navigate} role={currentRole} />;
}

function AccessDenied({ navigate, role }) {
  return (
    <Page label="Acesso restrito" title="Tela não disponível para seu perfil">
      <Card>
        <h3>Acesso não permitido</h3>
        <p>Esta área exige outro perfil de acesso. Você pode voltar para a tela inicial do seu perfil.</p>
        <Button onClick={() => navigate(firstScreenFor(role))}>Ir para minha tela inicial</Button>
      </Card>
    </Page>
  );
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
    "Central de Montagem da Licitação": "Organize entregas, responsáveis, prazos, revisões e documentos do consórcio em fases permanentes.",
    "Central de montagens": "Acesse e acompanhe todas as montagens individuais e consorciais da sua empresa em um só lugar.",
    "Histórico de alertas": "Consulte acontecimentos anteriores e retorne diretamente ao conteúdo relacionado.",
    "Análise e aprovação da empresa": "Revise os dados enviados pela empresa e decida se ela entra, ajusta informações ou será recusada.",
    "Editar perfil da empresa": "Mantenha a vitrine institucional atualizada. Essas informações aparecem na comunidade e no match.",
	"Acervo técnico": "Organize atestados técnicos da empresa e encontre experiências por conteúdo, serviços, dados contratuais e palavras-chave.",
    "Usuários vinculados": "Gerencie quem opera pela empresa. O perfil de acesso define as permissões de cada pessoa.",
    "Cadastro do usuário vinculado": "Inclua ou edite uma pessoa da empresa, escolhendo o perfil adequado para sua função.",
    "Cadastrar usuário vinculado": "Inclua uma nova pessoa da empresa e envie o convite de acesso.",
    "Editar usuário vinculado": "Atualize os dados, cargo e perfil de acesso de uma pessoa já vinculada à empresa.",
    "Confirmar bloqueio de usuário": "Suspenda temporariamente o acesso da pessoa sem remover seu vínculo ou histórico.",
    "Confirmar desbloqueio de usuário": "Reative o acesso de uma pessoa bloqueada para que ela volte a operar pela empresa.",
    "Desativar vínculo do usuário": "Confirme a desativação definitiva do vínculo deste usuário com a empresa.",
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
    "Vitrine de parceiros": "Veja anúncios de empresas interessadas em diferentes licitações e filtre oportunidades de consórcio.",
    "Avaliar candidata da licitação": "Avalie uma empresa por vez: veja o que ela oferece, o que falta e decida recusar ou dar match.",
    "Detalhe do anúncio": "Visão detalhada da empresa dentro da licitação, com oferta, necessidades e aderência.",
    "Match realizado": "Confirmação de interesse recíproco entre empresas na mesma licitação, com contato direto pelo WhatsApp.",
    "Academia LicitaHub": "Encontre cursos publicados para desenvolver a equipe e fortalecer a atuação em engenharia consultiva.",
    "Meus cursos": "Retome cursos iniciados, acompanhe o progresso das aulas e acesse certificados concluídos.",
    "Gerenciar cursos": "Consulte os cursos cadastrados e abra a gestão exclusiva de cada conteúdo.",
    "Gestão do curso": "Edite a apresentação do curso e organize todas as aulas, vídeos e questionários.",
    "Novo curso": "Cadastre a apresentação e a imagem do novo curso antes de organizar as aulas."
  };
  return help[title] || "Tela da LicitaHub para apoiar decisões empresariais com clareza e contexto.";
}

function Button({ children, variant = "primary", onClick, type = "button", disabled = false }) {
  return <button type={type} className={`btn ${variant}`} onClick={onClick} disabled={disabled}>{children}</button>;
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

function LogoSlot({ initials = "EC", size = "md", label = "Logo da empresa", src = "", fit = "contain" }) {
  return (
    <span className={`logoSlot logoSlot-${size} logoSlot-${fit}`} title={label} aria-label={label}>
      {src ? <img src={src} alt={label} /> : <span>{initials}</span>}
    </span>
  );
}

function ImageUploadField({ label, hint, accept = "image/*", initials = "IMG", variant = "mini", previewLabel }) {
  const [preview, setPreview] = useState("");
  const [fileName, setFileName] = useState("");
  const handleChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setPreview("");
      setFileName("");
      return;
    }
    setFileName(file.name);
    setPreview(URL.createObjectURL(file));
  };

  return (
    <Field label={label} hint={hint}>
      <div className={`imageUploadField ${variant === "hero" ? "imageUploadHero" : ""}`}>
        <div className={variant === "hero" ? "imageHeroPreview" : "imageMiniPreview"}>
          {preview ? <img src={preview} alt="Prévia da imagem selecionada" /> : <span>{initials}</span>}
          {variant === "hero" && <small>{previewLabel || "Identidade visual"}</small>}
        </div>
        <div>
          <input type="file" accept={accept} onChange={handleChange} />
          <small>{fileName || "Nenhuma imagem selecionada"}</small>
        </div>
      </div>
    </Field>
  );
}

function Card({ children, className = "", onClick, id }) {
  return <div className={`card ${className}`} onClick={onClick} id={id}>{children}</div>;
}

function PasswordResetManagement() {
  const [filters, setFilters] = useState({ user: "", company: "", status: "" });
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [copiedId, setCopiedId] = useState("");

  const load = async () => {
    setLoading(true);
    setError("");
    try {
      const params = new URLSearchParams();
      if (filters.user.trim()) params.set("user", filters.user.trim());
      if (filters.company.trim()) params.set("company", filters.company.trim());
      if (filters.status) params.set("status", filters.status);
      const response = await fetch(`${API_BASE_URL}/api/admin/password-reset-requests?${params.toString()}`, { credentials: "include" });
      const data = await response.json().catch(() => []);
      if (!response.ok) throw new Error(data.error || "Não foi possível carregar os pedidos.");
      setItems(Array.isArray(data) ? data : []);
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [filters.user, filters.company, filters.status]);

  const copyLink = async (item) => {
    if (!item.resetUrl) return;
    await navigator.clipboard.writeText(item.resetUrl);
    setCopiedId(item.id);
    window.setTimeout(() => setCopiedId(""), 1800);
  };

  const statusLabels = { available: "Disponível", used: "Usado", expired: "Expirado" };
  return <Page label="Acesso e administração" title="Recuperações de senha" actions={<Button variant="secondary" onClick={load}>Atualizar</Button>}>
    <section className="infoBanner"><strong>Controle temporário de links</strong><span>Como o e-mail ainda não está conectado, os links ficam disponíveis aqui para atendimento autorizado. Confirme a identidade do usuário antes de compartilhar.</span></section>
    <Card className="compactFilters resetRequestsFilters"><FormGrid>
      <Field label="Usuário"><input value={filters.user} onChange={(event) => setFilters((current) => ({ ...current, user: event.target.value }))} placeholder="Nome ou e-mail" /></Field>
      <Field label="Empresa"><input value={filters.company} onChange={(event) => setFilters((current) => ({ ...current, company: event.target.value }))} placeholder="Nome da empresa" /></Field>
      <Field label="Situação"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todas</option><option value="available">Disponível</option><option value="used">Usado</option><option value="expired">Expirado</option></select></Field>
    </FormGrid></Card>
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {loading ? <Card><p>Carregando pedidos de recuperação...</p></Card> : items.length === 0 ? <Card className="emptyState"><p>Nenhum pedido encontrado.</p></Card> : <section className="resetRequestList">{items.map((item) => <Card className={`resetRequestCard status-${item.status}`} key={item.id}>
      <div className="resetRequestHeader"><div><span className="eyebrow">{item.companyName}</span><h3>{item.fullName}</h3><p>{item.email}</p></div><span className={`statusPill ${item.status === "available" ? "open" : item.status === "used" ? "success" : "review"}`}>{statusLabels[item.status] || item.status}</span></div>
      <div className="resetRequestMeta"><span>Solicitado em <strong>{item.createdAt ? new Date(item.createdAt).toLocaleString("pt-BR") : "-"}</strong></span><span>Expira em <strong>{item.expiresAt ? new Date(item.expiresAt).toLocaleString("pt-BR") : "-"}</strong></span>{item.usedAt && <span>Usado em <strong>{new Date(item.usedAt).toLocaleString("pt-BR")}</strong></span>}</div>
      {item.resetUrl && <div className="resetRequestLink"><input value={item.resetUrl} readOnly /><Button onClick={() => copyLink(item)}>{copiedId === item.id ? "Copiado" : "Copiar link"}</Button></div>}
    </Card>)}</section>}
  </Page>;
}

function AdminDashboard({ navigate }) {
  return (
    <Page label="Administração" title="Painel administrativo">
      <Stats items={[["Convites enviados", "12"], ["Empresas pendentes", "4"], ["Empresas ativas", "28"], ["Editais abertos", "9"]]} />
      <div className="grid two">
        <Card><h3>Entrada de empresas</h3><p>Convites, aprovações e bloqueios ficam concentrados aqui.</p><Button onClick={() => navigate("invite-new")}>Novo convite</Button></Card>
        <Card><h3>Editais</h3><p>Cadastre oportunidades e acompanhe manifestações de interesse.</p><Button variant="secondary">Cadastrar edital</Button></Card>
      </div>
    </Page>
  );
}

function buildInvitationLink(invitation) {
  const token = invitation?.token || invitation?.invitation_token || "";
  const id = invitation?.id || "";
  const value = token ? `token=${encodeURIComponent(token)}` : id ? `id=${encodeURIComponent(id)}` : "";
  return `${window.location.origin}${window.location.pathname}#invite-accept${value ? `?${value}` : ""}`;
}

function parseAPIResponseText(text) {
  const trimmed = String(text || "").trim();
  if (!trimmed) return null;
  const jsonStart = trimmed.search(/[\[{]/);
  if (jsonStart === -1) return { error: trimmed };
  try {
    return JSON.parse(trimmed.slice(jsonStart));
  } catch {
    return { error: trimmed };
  }
}

function formatCNPJ(value = "") {
  const digits = String(value).replace(/\D/g, "").slice(0, 14);
  return digits
    .replace(/^(\d{2})(\d)/, "$1.$2")
    .replace(/^(\d{2})\.(\d{3})(\d)/, "$1.$2.$3")
    .replace(/\.(\d{3})(\d)/, ".$1/$2")
    .replace(/(\d{4})(\d)/, "$1-$2");
}

function isValidCNPJ(value = "") {
  const cnpj = String(value).replace(/\D/g, "");
  if (cnpj.length !== 14 || /^(\d)\1+$/.test(cnpj)) return false;
  const calcDigit = (base) => {
    const weights = base.length === 12 ? [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2] : [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2];
    const sum = base.split("").reduce((acc, digit, index) => acc + Number(digit) * weights[index], 0);
    const result = sum % 11;
    return result < 2 ? "0" : String(11 - result);
  };
  const first = calcDigit(cnpj.slice(0, 12));
  const second = calcDigit(cnpj.slice(0, 12) + first);
  return cnpj.endsWith(first + second);
}

function formatPhoneBR(value = "") {
  const digits = String(value).replace(/\D/g, "").slice(0, 11);
  if (digits.length <= 2) return digits;
  if (digits.length <= 6) return digits.replace(/^(\d{2})(\d+)/, "($1) $2");
  if (digits.length <= 10) return digits.replace(/^(\d{2})(\d{4})(\d+)/, "($1) $2-$3");
  return digits.replace(/^(\d{2})(\d{5})(\d+)/, "($1) $2-$3");
}

function isValidEmail(value = "") {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(String(value).trim());
}

function hasFirstAndLastName(value = "") {
  return String(value).trim().split(/\s+/).filter((part) => part.length >= 2).length >= 2;
}

function formatCurrencyBR(value = "") {
  const digits = String(value).replace(/\D/g, "");
  if (!digits) return "";
  const number = Number(digits) / 100;
  return number.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatCurrencyBRFromNumber(value = "") {
  if (value === "" || value === null || value === undefined) return "";
  const number = Number(value);
  return Number.isFinite(number) ? number.toLocaleString("pt-BR", { style: "currency", currency: "BRL" }) : "";
}

function parseCurrencyBR(value = "") {
  const normalized = String(value).replace(/[^\d,.-]/g, "").replace(/\./g, "").replace(",", ".");
  const number = Number(normalized);
  return Number.isFinite(number) ? number.toFixed(2) : "";
}

function isPastDateISO(value = "") {
  if (!value) return false;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const date = new Date(`${value}T00:00:00`);
  return date < today;
}

function StateSelect({ value, onChange, required = false }) {
  return (
    <select value={value} onChange={onChange} required={required}>
      <option value="">Selecione</option>
      {brazilStates.map((state) => <option value={state.uf} key={state.uf}>{state.uf} - {state.name}</option>)}
    </select>
  );
}

function CityField({ state, value, onChange, required = false }) {
  const options = majorCitiesByState[state] || [];
  return (
    <>
      <input list={state ? `cities-${state}` : undefined} value={value} onChange={onChange} placeholder={state ? "Cidade" : "Selecione o estado primeiro"} required={required} />
      {state && <datalist id={`cities-${state}`}>{options.map((city) => <option value={city} key={city} />)}</datalist>}
    </>
  );
}

function initialsFromName(value = "") {
  const parts = String(value || "Usuário").trim().split(/\s+/).filter(Boolean);
  return parts.map((part) => part[0]).join("").slice(0, 2).toUpperCase() || "US";
}

function userStatusLabel(status = "") {
  const labels = {
    active: "Ativo",
    blocked: "Bloqueado",
    inactive: "Inativo",
    pending_invite: "Convite pendente",
    removed: "Vínculo desativado"
  };
  return labels[status] || status || "Não informado";
}

function InviteNew() {
  const [form, setForm] = useState({
    tradeName: "",
    cnpj: "",
    contactName: "",
    email: "",
    phone: "",
    state: "",
    internalNote: ""
  });
  const [message, setMessage] = useState(null);
  const [createdInvitation, setCreatedInvitation] = useState(null);
  const [saving, setSaving] = useState(false);

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const invitationLink = createdInvitation ? buildInvitationLink(createdInvitation) : "";

  const copyInvitationLink = async () => {
    if (!invitationLink) return;
    try {
      await navigator.clipboard.writeText(invitationLink);
      setMessage({ type: "success", text: "Link do convite copiado. Agora voce pode enviar ao convidado." });
    } catch {
      setMessage({ type: "success", text: `Copie o link manualmente: ${invitationLink}` });
    }
  };

  const submitInvitation = async (event) => {
    event.preventDefault();
    if (!isValidCNPJ(form.cnpj)) {
      setMessage({ type: "error", text: "Informe um CNPJ válido." });
      return;
    }
    if (!hasFirstAndLastName(form.contactName)) {
      setMessage({ type: "error", text: "Informe o nome e sobrenome do contato principal." });
      return;
    }
    if (!isValidEmail(form.email)) {
      setMessage({ type: "error", text: "Informe um e-mail válido." });
      return;
    }
    if (String(form.phone).replace(/\D/g, "").length < 10) {
      setMessage({ type: "error", text: "Informe um telefone válido com DDD." });
      return;
    }
    setSaving(true);
    setMessage(null);
    setCreatedInvitation(null);

    try {
      const response = await fetch(`${API_BASE_URL}/api/company-invitations`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form)
      });

      const text = await response.text();
      const data = parseAPIResponseText(text);

      if (!response.ok || data?.error) {
        throw new Error(data?.error || "Não foi possível enviar o convite.");
      }

      setCreatedInvitation(data);
      setMessage({ type: "success", text: "Convite gravado com sucesso. Copie o link abaixo e envie ao convidado." });
      setForm({
        tradeName: "",
        cnpj: "",
        contactName: "",
        email: "",
        phone: "",
        state: "",
        internalNote: ""
      });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível gravar o convite. Confirme se o backend está ligado." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Page label="Convites" title="Novo convite de empresa">
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {createdInvitation && (
        <Card className="inviteLinkBox">
          <h3>Link do convite</h3>
          <p>Envie este link para a empresa convidada completar o cadastro no sistema.</p>
          <div className="copyLine">
            <input readOnly value={invitationLink} />
            <Button type="button" onClick={copyInvitationLink}>Copiar link</Button>
          </div>
        </Card>
      )}
      <form onSubmit={submitInvitation}>
        <FormGrid>
          <Field label="Nome fantasia" hint="Obrigatório. Único."><input value={form.tradeName} onChange={(event) => updateField("tradeName", event.target.value)} placeholder="Engenvale Consultoria" required /></Field>
          <Field label="CNPJ" hint="Obrigatório. Único."><input value={form.cnpj} onChange={(event) => updateField("cnpj", formatCNPJ(event.target.value))} placeholder="00.000.000/0000-00" maxLength="18" required /></Field>
          <Field label="Contato principal" hint="Obrigatório. Pode repetir."><input value={form.contactName} onChange={(event) => updateField("contactName", event.target.value)} placeholder="Nome completo" required /></Field>
          <Field label="E-mail" hint="Obrigatório. Pode repetir."><input type="email" value={form.email} onChange={(event) => updateField("email", event.target.value)} placeholder="contato@empresa.com.br" required /></Field>
          <Field label="Telefone" hint="Obrigatório. Pode repetir."><input value={form.phone} onChange={(event) => updateField("phone", formatPhoneBR(event.target.value))} placeholder="(00) 00000-0000" maxLength="15" required /></Field>
          <Field label="Estado"><StateSelect value={form.state} onChange={(event) => updateField("state", event.target.value)} /></Field>
        </FormGrid>
        <Field label="Observação interna"><textarea value={form.internalNote} onChange={(event) => updateField("internalNote", event.target.value)} placeholder="Visível apenas para administradores" /></Field>
        <div className="formActionBar">
          <Button type="submit" disabled={saving}>{saving ? "Enviando..." : "Enviar convite"}</Button>
        </div>
      </form>
    </Page>
  );
}

function InviteList({ navigate }) {
  const [invitations, setInvitations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ search: "", status: "", sort: "created_desc" });

  const loadInvitations = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/company-invitations`)
      .then((response) => {
        if (!response.ok) throw new Error("Não foi possível carregar os convites.");
        return response.json();
      })
      .then((data) => {
        setInvitations(Array.isArray(data) ? data : []);
        setMessage(null);
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar os convites. Confirme se o backend está ligado." }))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadInvitations();
  }, []);

  const copyInvitationLink = async (invitation) => {
    const link = buildInvitationLink(invitation);
    try {
      await navigator.clipboard.writeText(link);
      setMessage({ type: "success", text: "Link do convite copiado." });
    } catch {
      setMessage({ type: "success", text: `Copie o link manualmente: ${link}` });
    }
  };

  const statusLabel = (status) => ({
    sent: "Enviado",
    pending_review: "Aguardando análise",
    accepted: "Aceito",
    expired: "Expirado",
    cancelled: "Cancelado",
    rejected: "Recusado"
  }[status] || status || "-");

  const filteredInvitations = invitations.filter((invitation) => {
    const search = filters.search.trim().toLowerCase();
    const status = filters.status;
    const matchesSearch = !search || [
      invitation.tradeName,
      invitation.cnpj,
      invitation.contactName,
      invitation.email,
      invitation.phone
    ].some((value) => String(value || "").toLowerCase().includes(search));
    const matchesStatus = !status || invitation.status === status;
    return matchesSearch && matchesStatus;
  }).sort((a, b) => {
    if (filters.sort === "name_asc") return String(a.tradeName || "").localeCompare(String(b.tradeName || ""), "pt-BR");
    if (filters.sort === "expires_asc") return new Date(a.expiresAt || "9999-12-31").getTime() - new Date(b.expiresAt || "9999-12-31").getTime();
    if (filters.sort === "created_asc") return new Date(a.createdAt || 0).getTime() - new Date(b.createdAt || 0).getTime();
    return new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime();
  });
  const pendingReviews = invitations.filter((invitation) => invitation.status === "pending_review");
  const pendingReviewCount = pendingReviews.length;

  return (
    <Page label="Convites" title="Lista de convites" actions={<Button onClick={() => navigate("invite-new")}>Novo convite</Button>}>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {pendingReviewCount > 0 && <Card className="notice"><h3>{pendingReviewCount} {pendingReviewCount === 1 ? "empresa aguarda" : "empresas aguardam"} análise</h3><p>O cadastro foi preenchido, mas o acesso somente será liberado após sua aprovação.</p><div className="reviewQueue">{pendingReviews.map((invitation) => <div className="reviewQueueItem" key={invitation.id}><div><strong>{invitation.tradeName}</strong><span>CNPJ {invitation.cnpj}</span></div><Button onClick={() => navigate(`company-review?id=${encodeURIComponent(invitation.id)}`)}>Analisar empresa</Button></div>)}</div></Card>}
      <Card className="compactFilters inviteFiltersSticky">
        <FormGrid>
          <Field label="Buscar convite"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Empresa, CNPJ, e-mail ou telefone" /></Field>
          <Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option><option value="sent">Enviado</option><option value="pending_review">Aguardando análise</option><option value="accepted">Aceito</option><option value="cancelled">Cancelado</option><option value="rejected">Recusado</option><option value="expired">Expirado</option></select></Field>
          <Field label="Ordenar por"><select value={filters.sort} onChange={(event) => setFilters((current) => ({ ...current, sort: event.target.value }))}><option value="created_desc">Mais recentes</option><option value="created_asc">Mais antigos</option><option value="expires_asc">Validade mais próxima</option><option value="name_asc">Empresa A-Z</option></select></Field>
        </FormGrid>
      </Card>
      {loading && <Card className="formFeedback"><p>Carregando convites do banco...</p></Card>}
      {!loading && invitations.length === 0 && <Card className="formFeedback"><p>Nenhum convite gravado ainda.</p></Card>}
      {!loading && invitations.length > 0 && filteredInvitations.length === 0 && <Card className="formFeedback"><p>Nenhum convite encontrado com esses filtros.</p></Card>}
      {!loading && filteredInvitations.length > 0 && (
        <Table columns={["Empresa", "CNPJ", "Contato", "E-mail", "Telefone", "Status", "Ações"]} rows={filteredInvitations.map((invitation) => [
          invitation.tradeName,
          invitation.cnpj,
          invitation.contactName,
          invitation.email,
          invitation.phone,
          statusLabel(invitation.status),
          <div className="rowActions compactActions" key={`${invitation.id}-actions`}>
            <button className="iconButton secondaryIcon" title="Copiar link do convite" aria-label="Copiar link do convite" onClick={() => copyInvitationLink(invitation)}>{"\u2197"}</button>
            {invitation.status === "pending_review" && <button className="iconButton successIcon" title="Analisar empresa" aria-label="Analisar empresa" onClick={() => navigate(`company-review?id=${encodeURIComponent(invitation.id)}`)}>{"\u2713"}</button>}
          </div>
        ])} />
      )}
    </Page>
  );
}

function InviteAccept({ navigate }) {
  const params = currentHashParams();
  const token = params.get("token") || "";
  const invitationId = params.get("id") || "";
  const [invitation, setInvitation] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [form, setForm] = useState({
    website: "",
    institutionalDescription: "",
    city: "",
    state: "",
    adminFullName: "",
    adminEmail: "",
    adminPhone: "",
    adminJobTitle: "",
    password: "",
    confirmPassword: "",
    profilePhotoDataUrl: "",
    profilePhotoFileName: "",
    profilePhotoMimeType: ""
  });
  const [photoPreview, setPhotoPreview] = useState("");

  const updateField = (field, value) => setForm((current) => ({ ...current, [field]: value }));
  const handlePhotoChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!["image/png", "image/jpeg", "image/webp"].includes(file.type)) {
      setMessage({ type: "error", text: "Use uma foto PNG, JPG ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A foto deve ter no máximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setPhotoPreview(dataUrl);
      setForm((current) => ({ ...current, profilePhotoDataUrl: dataUrl, profilePhotoFileName: file.name, profilePhotoMimeType: file.type }));
    };
    reader.readAsDataURL(file);
  };

  useEffect(() => {
    const endpoint = token
      ? `${API_BASE_URL}/api/company-invitations/by-token?token=${encodeURIComponent(token)}`
      : invitationId
        ? `${API_BASE_URL}/api/company-invitations/${encodeURIComponent(invitationId)}`
        : "";

    if (!endpoint) {
      setLoading(false);
      setMessage({ type: "error", text: "Link do convite invalido. Solicite um novo link ao administrador." });
      return;
    }

    fetch(endpoint)
      .then((response) => {
        if (!response.ok) throw new Error("Não foi possível carregar o convite.");
        return response.text();
      })
      .then((text) => {
        const data = parseAPIResponseText(text);
        if (!data || data.error) throw new Error(data?.error || "Convite não encontrado.");
        setInvitation(data);
        setForm((current) => ({
          ...current,
          website: data.website || "",
          institutionalDescription: data.institutionalDescription || "",
          city: data.city || "",
          state: data.state || "",
          adminFullName: data.adminFullName || data.contactName || "",
          adminEmail: data.adminEmail || data.email || "",
          adminPhone: data.adminPhone || data.phone || "",
          adminJobTitle: data.adminJobTitle || ""
        }));
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar o convite." }))
      .finally(() => setLoading(false));
  }, [token, invitationId]);

  const submitAccept = async (event) => {
    event.preventDefault();
    setMessage(null);
    const isCorrection = invitation?.reviewStatus === "adjustment_requested";

    if (!isCorrection && form.password.length < 8) {
      setMessage({ type: "error", text: "A senha deve ter pelo menos 8 caracteres." });
      return;
    }
    if (!isCorrection && form.password !== form.confirmPassword) {
      setMessage({ type: "error", text: "A confirmação da senha não confere." });
      return;
    }
    if (!hasFirstAndLastName(form.adminFullName)) {
      setMessage({ type: "error", text: "Informe nome e sobrenome do administrador da empresa." });
      return;
    }
    if (!isValidEmail(form.adminEmail)) {
      setMessage({ type: "error", text: "Informe um e-mail válido para o administrador." });
      return;
    }
    if (String(form.adminPhone).replace(/\D/g, "").length < 10) {
      setMessage({ type: "error", text: "Informe um telefone válido com DDD para o administrador." });
      return;
    }

    setSaving(true);
    try {
      const endpoint = isCorrection
        ? `${API_BASE_URL}/api/company-invitations/${encodeURIComponent(invitation?.id)}/resubmit`
        : token ? `${API_BASE_URL}/api/company-invitations/by-token/accept` : `${API_BASE_URL}/api/company-invitations/${encodeURIComponent(invitationId || invitation?.id)}/accept`;

      const response = await fetch(endpoint, {
        method: isCorrection ? "PATCH" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token,
          invitationId: invitationId || invitation?.id || "",
          website: form.website,
          institutionalDescription: form.institutionalDescription,
          city: form.city,
          state: form.state,
          adminFullName: form.adminFullName,
          adminEmail: form.adminEmail,
          adminPhone: form.adminPhone,
          adminJobTitle: form.adminJobTitle,
          password: form.password,
          profilePhotoDataUrl: form.profilePhotoDataUrl,
          profilePhotoFileName: form.profilePhotoFileName,
          profilePhotoMimeType: form.profilePhotoMimeType
        })
      });

      const data = parseAPIResponseText(await response.text());
      if (!response.ok || data?.error) {
        throw new Error(data?.error || "Não foi possível aceitar o convite.");
      }

      setMessage({ type: "success", text: isCorrection ? "Cadastro corrigido e reenviado para nova análise." : "Cadastro enviado com sucesso. A empresa agora ficou aguardando análise da LicitaHub." });
      if (!isCorrection) {
        window.sessionStorage.setItem("licitahubLoginNotice", "Cadastro concluído com sucesso. Agora entre com o e-mail e a senha cadastrados.");
        window.location.hash = "";
      }
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível concluir o aceite do convite." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Page label="Convite" title="Aceite do convite">
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading && <Card className="formFeedback"><p>Carregando convite...</p></Card>}
      {invitation && (
        <form onSubmit={submitAccept}>
          {invitation.reviewStatus === "adjustment_requested" && <Card className="notice"><h3>Ajustes solicitados pela LicitaHub</h3><p>{invitation.adjustmentRequest}</p></Card>}
          <Card className="inviteLinkBox">
            <h3>{invitation.tradeName}</h3>
            <p>CNPJ {invitation.cnpj} | convite para {invitation.email}</p>
          </Card>
          <FormGrid>
            <Field label="Nome fantasia"><input value={invitation.tradeName || ""} readOnly /></Field>
            <Field label="CNPJ"><input value={invitation.cnpj || ""} readOnly /></Field>
            <Field label="Site da empresa"><input value={form.website} onChange={(event) => updateField("website", event.target.value)} placeholder="https://www.empresa.com.br" /></Field>
            <Field label="Estado"><StateSelect value={form.state} onChange={(event) => updateField("state", event.target.value)} /></Field>
            <Field label="Cidade"><CityField state={form.state} value={form.city} onChange={(event) => updateField("city", event.target.value)} /></Field>
            <Field label="Nome do administrador"><input value={form.adminFullName} onChange={(event) => updateField("adminFullName", event.target.value)} required /></Field>
            <Field label="E-mail do administrador"><input type="email" value={form.adminEmail} onChange={(event) => updateField("adminEmail", event.target.value)} required /></Field>
            <Field label="Telefone do administrador"><input value={form.adminPhone} onChange={(event) => updateField("adminPhone", formatPhoneBR(event.target.value))} maxLength="15" required /></Field>
            <Field label="Cargo ou funcao"><input value={form.adminJobTitle} onChange={(event) => updateField("adminJobTitle", event.target.value)} placeholder="Ex.: Diretor comercial" /></Field>
            <Field label="Perfil de acesso"><input value="Administrador da empresa" readOnly /></Field>
            {invitation.reviewStatus !== "adjustment_requested" && <Field label="Senha de acesso" hint="Minimo de 8 caracteres."><input type="password" value={form.password} onChange={(event) => updateField("password", event.target.value)} required /></Field>}
            {invitation.reviewStatus !== "adjustment_requested" && <Field label="Confirmar senha"><input type="password" value={form.confirmPassword} onChange={(event) => updateField("confirmPassword", event.target.value)} required /></Field>}
          </FormGrid>
          <Field label="Foto do administrador" hint="Será usada no perfil, comentários e identificação do usuário.">
            <div className="imageUploadField">
              <div className="imageMiniPreview profilePhotoPreview">{photoPreview ? <img src={photoPreview} alt="Prévia da foto do administrador" /> : <span>Foto</span>}</div>
              <div><input type="file" accept="image/png,image/jpeg,image/webp" onChange={handlePhotoChange} /><small>{form.profilePhotoFileName || "Nenhuma foto selecionada"}</small></div>
            </div>
          </Field>
          <Field label="Descrição institucional"><textarea value={form.institutionalDescription} onChange={(event) => updateField("institutionalDescription", event.target.value)} placeholder="Resumo da atuação da empresa" /></Field>
          <div className="formActionBar">
            <Button type="submit" disabled={saving}>{saving ? "Enviando..." : invitation.reviewStatus === "adjustment_requested" ? "Reenviar cadastro corrigido" : "Enviar cadastro para análise"}</Button>
          </div>
        </form>
      )}
    </Page>
  );
}

function CompanyReview() {
  const invitationId = currentHashParams().get("id") || "";
  const [invitation, setInvitation] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState("");
  const [adjustmentRequest, setAdjustmentRequest] = useState("");
  const [reviewNote, setReviewNote] = useState("");
  const [message, setMessage] = useState(null);

  useEffect(() => {
    if (!invitationId) {
      setLoading(false);
      setMessage({ type: "error", text: "Abra esta tela pelo botão Analisar empresa na lista de convites." });
      return;
    }
    fetch(`${API_BASE_URL}/api/company-invitations/${encodeURIComponent(invitationId)}`)
      .then(async (response) => {
        const data = parseAPIResponseText(await response.text());
        if (!response.ok || !data) throw new Error(data?.error || "Não foi possível carregar a empresa.");
        return data;
      })
      .then((data) => setInvitation(data))
      .catch((error) => setMessage({ type: "error", text: error.message }))
      .finally(() => setLoading(false));
  }, [invitationId]);

  const submitReview = async (decision) => {
    if (decision === "adjustment_requested" && !adjustmentRequest.trim()) {
      setMessage({ type: "error", text: "Descreva o ajuste que a empresa precisa realizar." });
      return;
    }
    setSaving(decision);
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/company-invitations/${encodeURIComponent(invitationId)}/review`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ decision, adjustmentRequest, reviewNote })
      });
      const data = parseAPIResponseText(await response.text());
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível concluir a análise.");
      const successText = decision === "approved"
        ? "Empresa aprovada. O login do administrador foi liberado."
        : decision === "rejected"
          ? "Empresa recusada e acesso não liberado."
          : "Solicitação de ajuste registrada.";
      setMessage({ type: "success", text: successText });
      setInvitation((current) => ({ ...current, status: data.invitationStatus }));
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving("");
    }
  };

  return (
    <Page label="Administração" title="Análise e aprovação da empresa">
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading && <Card><p>Carregando cadastro da empresa...</p></Card>}
      {invitation && (
        <>
          <FormGrid>
            <Field label="Empresa"><input value={invitation.tradeName || ""} readOnly /></Field>
            <Field label="CNPJ"><input value={invitation.cnpj || ""} readOnly /></Field>
            <Field label="Responsável"><input value={invitation.adminFullName || invitation.contactName || ""} readOnly /></Field>
            <Field label="E-mail"><input value={invitation.adminEmail || invitation.email || ""} readOnly /></Field>
            <Field label="Telefone"><input value={invitation.adminPhone || invitation.phone || ""} readOnly /></Field>
            <Field label="Cargo"><input value={invitation.adminJobTitle || ""} readOnly /></Field>
            <Field label="Cidade / Estado"><input value={[invitation.city, invitation.state].filter(Boolean).join(" / ")} readOnly /></Field>
            <Field label="Status"><input value={invitation.status === "pending_review" ? "Aguardando aprovação" : invitation.status} readOnly /></Field>
          </FormGrid>
          <Field label="Site"><input value={invitation.website || "Não informado"} readOnly /></Field>
          <Field label="Descrição institucional"><textarea value={invitation.institutionalDescription || "Não informada"} readOnly /></Field>
          <Field label="Observação interna da análise"><textarea value={reviewNote} onChange={(event) => setReviewNote(event.target.value)} placeholder="Observação visível apenas para a administração" /></Field>
          {invitation.status === "pending_review" && <div className="actions"><Button onClick={() => submitReview("approved")} disabled={Boolean(saving)}>{saving === "approved" ? "Aprovando..." : "Aprovar e liberar acesso"}</Button><Button variant="danger" onClick={() => submitReview("rejected")} disabled={Boolean(saving)}>Recusar</Button></div>}
        </>
      )}
    </Page>
  );
}

function CompanyManage() {
  const [companies, setCompanies] = useState([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ search: "", status: "" });
  const [selectedCompany, setSelectedCompany] = useState(null);
  const [reason, setReason] = useState("");
  const [saving, setSaving] = useState(false);

  const loadCompanies = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/companies`, { credentials: "include" })
      .then(async (response) => {
        const data = parseAPIResponseText(await response.text());
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível carregar as empresas.");
        return Array.isArray(data) ? data : [];
      })
      .then(setCompanies)
      .catch((error) => setMessage({ type: "error", text: error.message }))
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadCompanies(); }, []);

  const filteredCompanies = useMemo(() => companies.filter((company) => {
    const searchable = [company.tradeName, company.cnpj, company.mainContactName, company.mainContactEmail, company.city, company.state].join(" ").toLowerCase();
    return (!filters.search || searchable.includes(filters.search.toLowerCase())) && (!filters.status || company.status === filters.status);
  }), [companies, filters]);

  const statusLabel = (status) => ({
    active: "Ativa",
    blocked: "Bloqueada",
    pending_review: "Em análise",
    invited: "Convidada",
    inactive: "Inativa",
    rejected: "Recusada"
  }[status] || status);

  const confirmStatus = async () => {
    if (!selectedCompany) return;
    const nextStatus = selectedCompany.status === "blocked" ? "active" : "blocked";
    setSaving(true);
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/companies/${encodeURIComponent(selectedCompany.id)}/status`, {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: nextStatus, reason })
      });
      const data = parseAPIResponseText(await response.text());
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível alterar o acesso da empresa.");
      setCompanies((current) => current.map((company) => company.id === data.id ? { ...company, status: data.status } : company));
      setMessage({ type: "success", text: nextStatus === "blocked" ? "Empresa bloqueada. Todos os usuários vinculados tiveram as sessões encerradas." : "Empresa desbloqueada. Seus usuários podem entrar novamente." });
      setSelectedCompany(null);
      setReason("");
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving(false);
    }
  };

  const activeCount = companies.filter((company) => company.status === "active").length;
  const blockedCount = companies.filter((company) => company.status === "blocked").length;
  const isBlocking = selectedCompany?.status !== "blocked";

  return (
    <Page label="Administração" title="Empresas cadastradas">
      <Stats items={[["Ativas", String(activeCount)], ["Bloqueadas", String(blockedCount)], ["Total", String(companies.length)]]} />
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      <Card className="compactFilters companyManageFiltersSticky">
        <FormGrid>
          <Field label="Buscar empresa"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Empresa, CNPJ, contato ou local" /></Field>
          <Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option><option value="active">Ativa</option><option value="blocked">Bloqueada</option><option value="pending_review">Em análise</option><option value="inactive">Inativa</option><option value="rejected">Recusada</option></select></Field>
        </FormGrid>
      </Card>
      {loading && <Card><p>Carregando empresas...</p></Card>}
      {!loading && filteredCompanies.length === 0 && <Card><p>Nenhuma empresa encontrada com estes filtros.</p></Card>}
      {!loading && filteredCompanies.length > 0 && <Table columns={["Empresa", "CNPJ", "Contato", "Local", "Status", "Ação"]} rows={filteredCompanies.map((company) => [
        <div key={`${company.id}-name`}><strong>{company.tradeName}</strong><small>{company.mainContactEmail || "Sem e-mail informado"}</small></div>,
        company.cnpj || "-",
        <div key={`${company.id}-contact`}><strong>{company.mainContactName || "-"}</strong><small>{company.mainContactPhone || "Sem telefone informado"}</small></div>,
        [company.city, company.state].filter(Boolean).join(" / ") || "-",
        <span className={`statusPill ${company.status === "blocked" ? "closed" : company.status === "active" ? "open" : "review"}`}>{statusLabel(company.status)}</span>,
        ["active", "blocked"].includes(company.status) ? <button className={`iconButton ${company.status === "blocked" ? "successIcon" : "dangerIcon"}`} title={company.status === "blocked" ? "Desbloquear empresa" : "Bloquear empresa e todos os usuários"} aria-label={company.status === "blocked" ? "Desbloquear empresa" : "Bloquear empresa e todos os usuários"} onClick={() => { setSelectedCompany(company); setReason(""); }}>{company.status === "blocked" ? "✓" : "!"}</button> : "-"
      ])} />}
      {selectedCompany && <div className="modalBackdrop" role="presentation"><section className="modalCard" role="dialog" aria-modal="true" aria-labelledby="company-access-title">
        <div className="modalHeader"><div><span className="eyebrow">Confirmação</span><h2 id="company-access-title">{isBlocking ? "Bloquear empresa" : "Desbloquear empresa"}</h2></div><button className="iconButton" title="Fechar" aria-label="Fechar" onClick={() => setSelectedCompany(null)}>×</button></div>
        <p>{isBlocking ? <>Ao bloquear <strong>{selectedCompany.tradeName}</strong>, todos os usuários vinculados perderão o acesso imediatamente. Nenhum cadastro será apagado.</> : <>Ao desbloquear <strong>{selectedCompany.tradeName}</strong>, os usuários ativos poderão entrar novamente conforme seus perfis de acesso.</>}</p>
        <Field label="Motivo da ação" hint="Opcional. Fica guardado apenas para auditoria administrativa."><textarea value={reason} onChange={(event) => setReason(event.target.value)} placeholder={isBlocking ? "Ex.: pendência cadastral em análise" : "Ex.: pendência regularizada"} /></Field>
        <div className="actions"><Button variant={isBlocking ? "danger" : "primary"} onClick={confirmStatus} disabled={saving}>{saving ? "Salvando..." : isBlocking ? "Confirmar bloqueio" : "Confirmar desbloqueio"}</Button><Button variant="secondary" onClick={() => setSelectedCompany(null)} disabled={saving}>Cancelar</Button></div>
      </section></div>}
    </Page>
  );
}

function CompanyDashboard({ navigate, sessionUser }) {
  const isCompanyAdmin = sessionUser?.roleKey === "company_admin";
  const [dashboard, setDashboard] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ focus: "all", period: "30" });

  const loadDashboard = () => {
    setLoading(true);
    setError("");
    fetch(`${API_BASE_URL}/api/company-dashboard`, { credentials: "include" })
      .then(async (response) => {
        const data = parseAPIResponseText(await response.text());
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível carregar o painel da empresa.");
        return data;
      })
      .then(setDashboard)
      .catch((loadError) => setError(loadError.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadDashboard(); }, []);

  const statsData = dashboard?.stats || {};
  const tenders = Array.isArray(dashboard?.tenders) ? dashboard.tenders : [];
  const tasks = Array.isArray(dashboard?.tasks) ? dashboard.tasks : [];
  const consortia = Array.isArray(dashboard?.consortia) ? dashboard.consortia : [];
  const notifications = Array.isArray(dashboard?.notifications) ? dashboard.notifications : [];
  const formatDate = (value) => value ? new Date(value).toLocaleDateString("pt-BR") : "Sem data definida";
  const deadlineState = (task) => {
    if (!task.dueAt) return { label: "Sem prazo", tone: "neutral" };
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const due = new Date(`${String(task.dueAt).slice(0, 10)}T00:00:00`);
    const days = Math.ceil((due - today) / 86400000);
    if (days < 0) return { label: `${Math.abs(days)} dia${Math.abs(days) === 1 ? "" : "s"} em atraso`, tone: "danger" };
    if (days <= 1) return { label: days === 0 ? "Vence hoje" : "Vence amanhã", tone: "attention" };
    return { label: `Prazo ${due.toLocaleDateString("pt-BR")}`, tone: "neutral" };
  };
  const priorityItems = useMemo(() => {
    const taskItems = tasks.filter((task) => {
      const state = deadlineState(task);
      return state.tone === "danger" || state.tone === "attention" || ["waiting_information", "blocked", "returned_for_adjustment"].includes(task.status);
    }).map((task) => {
      const state = deadlineState(task);
      return { id: `task-${task.id}`, kind: "Tarefa", title: task.title, description: `${task.tenderNumber || "Edital"} · ${task.stageTitle || "Montagem"}`, badge: state.label, tone: state.tone, destination: `assembly-board?assemblyId=${encodeURIComponent(task.assemblyId)}&taskId=${encodeURIComponent(task.id)}` };
    });
    const tenderItems = tenders.filter((tender) => {
      if (!tender.openingDate) return false;
      const days = Math.ceil((new Date(tender.openingDate).getTime() - Date.now()) / 86400000);
      return days >= 0 && days <= 7;
    }).map((tender) => ({ id: `tender-${tender.id}`, kind: "Edital", title: `${tender.agency} · ${tender.number}`, description: "Sessão próxima. Revise a participação e os parceiros.", badge: `Abertura ${formatDate(tender.openingDate)}`, tone: "attention", destination: `tender-detail?id=${encodeURIComponent(tender.id)}` }));
    const maxDays = Number(filters.period || 30);
    const notificationItems = notifications.filter((notification) => !notification.createdAt || (Date.now() - new Date(notification.createdAt).getTime()) <= maxDays * 86400000).map((notification) => ({ id: `notice-${notification.id}`, kind: "Novidade", title: notification.title, description: notification.message || "Há uma atualização para acompanhar.", badge: notification.createdAt ? new Date(notification.createdAt).toLocaleDateString("pt-BR") : "Agora", tone: "neutral", destination: notification.destinationScreen || "company-dashboard", relatedId: notification.relatedEntityId }));
    return [...taskItems, ...tenderItems, ...notificationItems].slice(0, 8);
  }, [tasks, tenders, notifications, filters.period]);

  const openNotification = (item) => {
    if (item.destination === "tender-detail" && item.relatedId) navigate(`tender-detail?id=${encodeURIComponent(item.relatedId)}`);
    else if (item.destination === "tender-interest-list" && item.relatedId) navigate(`tender-interest-list?id=${encodeURIComponent(item.relatedId)}`);
    else navigate(item.destination || "company-dashboard");
  };
  const show = (section) => filters.focus === "all" || filters.focus === section;

  return (
    <Page label="Empresa" title="Painel da empresa" actions={<Button variant="secondary" onClick={loadDashboard}>Atualizar dados</Button>}>
      <section className="dashboardHero companyPanelHero">
        <div><span className="eyebrow">Visão de operação</span><h3>{sessionUser?.companyName || "Sua empresa"}</h3><p>Prioridades, oportunidades e andamento dos consórcios reunidos em um único lugar.</p></div>
        <div className="profileScore"><strong>{statsData.unreadNotifications || 0}</strong><span>novidades para olhar</span></div>
      </section>
      <Card className="compactFilters dashboardFiltersSticky"><FormGrid><Field label="Exibir"><select value={filters.focus} onChange={(event) => setFilters((current) => ({ ...current, focus: event.target.value }))}><option value="all">Visão completa</option><option value="priorities">Prioridades</option><option value="tenders">Editais</option><option value="consortia">Consórcios</option><option value="tasks">Tarefas</option></select></Field><Field label="Novidades dos últimos"><select value={filters.period} onChange={(event) => setFilters((current) => ({ ...current, period: event.target.value }))}><option value="7">7 dias</option><option value="30">30 dias</option><option value="90">90 dias</option></select></Field></FormGrid></Card>
      {loading && <Card><p>Carregando o painel da empresa...</p></Card>}
      {error && <Card className="dangerNotice"><p>{error}</p><Button variant="secondary" onClick={loadDashboard}>Tentar novamente</Button></Card>}
      {!loading && !error && <>
        <Stats items={[["Editais acompanhados", String(statsData.interestedTenders || 0)], ["Anúncios ativos", String(statsData.activeAds || 0)], ["Consórcios ativos", String(statsData.activeConsortia || 0)], ["Tarefas abertas", String(statsData.openTasks || 0)]]} />
        {show("priorities") && <div className="dashboardGrid"><Card><div className="cardHeader"><h3>Prioridades</h3><span className="dashboardCounter">{priorityItems.length}</span></div><div className="priorityList">{priorityItems.length === 0 && <p className="dashboardEmpty">Nenhuma prioridade imediata. Bom momento para acompanhar os editais e a comunidade.</p>}{priorityItems.map((item) => <button className={`priorityItem dashboardPriority ${item.tone}`} key={item.id} onClick={() => item.relatedId ? openNotification(item) : navigate(item.destination)}><span><small>{item.kind}</small><strong>{item.title}</strong><span>{item.description}</span></span><em>{item.badge}</em></button>)}</div></Card><Card><div className="cardHeader"><h3>Atividade recente</h3><button className="textAction" onClick={() => navigate("community-home")}>Comunidade</button></div><div className="activityList">{notifications.length === 0 && <p className="dashboardEmpty">Nenhuma novidade pendente no momento.</p>}{notifications.slice(0, 5).map((notification) => <button className="dashboardActivity" key={notification.id} onClick={() => openNotification({ destination: notification.destinationScreen, relatedId: notification.relatedEntityId })}><strong>{notification.title}</strong><span>{notification.message || "Atualização disponível"}</span><small>{notification.createdAt ? new Date(notification.createdAt).toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", hour: "2-digit", minute: "2-digit" }) : "Agora"}</small></button>)}</div></Card></div>}
        {show("tenders") && <Card className="dashboardSection"><div className="cardHeader"><div><span className="eyebrow">Editais</span><h3>Participações em acompanhamento</h3></div><Button variant="secondary" onClick={() => navigate("tender-list")}>Ver lista</Button></div><div className="dashboardRows">{tenders.length === 0 && <p className="dashboardEmpty">Sua empresa ainda não possui interesses registrados em editais abertos.</p>}{tenders.map((tender) => <button className="dashboardRow" key={tender.id} onClick={() => navigate(`tender-detail?id=${encodeURIComponent(tender.id)}`)}><span><strong>{tender.agency} · {tender.number}</strong><small>{tender.object}</small></span><em>{formatDate(tender.openingDate)}</em></button>)}</div></Card>}
        {show("consortia") && <Card className="dashboardSection"><div className="cardHeader"><div><span className="eyebrow">Consórcios</span><h3>Composições em andamento</h3></div><Button variant="secondary" onClick={() => navigate("match-list")}>Meus consórcios</Button></div><div className="dashboardRows">{consortia.length === 0 && <p className="dashboardEmpty">Nenhum consórcio ativo para acompanhar.</p>}{consortia.map((consortium) => <button className="dashboardRow" key={consortium.id} onClick={() => navigate("match-list")}><span><strong>{consortium.tenderNumber} · {consortium.agency}</strong><small>{consortium.tenderObject}</small></span><em>{consortium.memberCount} empresa{Number(consortium.memberCount) === 1 ? "" : "s"}</em></button>)}</div></Card>}
        {show("tasks") && <Card className="dashboardSection"><div className="cardHeader"><div><span className="eyebrow">Montagem</span><h3>Tarefas sob responsabilidade da empresa</h3></div><Button variant="secondary" onClick={() => navigate("my-assembly-tasks")}>Minhas tarefas</Button></div><div className="dashboardRows">{tasks.length === 0 && <p className="dashboardEmpty">Não há tarefas abertas atribuídas à sua empresa.</p>}{tasks.map((task) => { const deadline = deadlineState(task); return <button className={`dashboardRow taskTone ${deadline.tone}`} key={task.id} onClick={() => navigate(`assembly-board?assemblyId=${encodeURIComponent(task.assemblyId)}&taskId=${encodeURIComponent(task.id)}`)}><span><strong>{task.title}</strong><small>{task.tenderNumber} · {task.stageTitle}</small></span><em>{deadline.label}</em></button>; })}</div></Card>}
      </>}
    </Page>
  );
}

function MyProfile({ refreshSession, sessionUser }) {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [photoPreview, setPhotoPreview] = useState("");
  const [form, setForm] = useState({
    fullName: "",
    email: "",
    phone: "",
    jobTitle: "",
    roleName: "",
    companyName: "",
    profilePhotoUrl: "",
    profilePhotoDataUrl: "",
    profilePhotoFileName: "",
    profilePhotoMimeType: ""
  });

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/users/me`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => ({}));
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível carregar seu perfil.");
        return data;
      })
      .then((data) => {
        setForm((current) => ({ ...current, ...data, profilePhotoDataUrl: "", profilePhotoFileName: "", profilePhotoMimeType: "" }));
        setPhotoPreview(data.profilePhotoUrl || "");
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar seu perfil." }))
      .finally(() => setLoading(false));
  }, []);

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handlePhotoChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setPhotoPreview(form.profilePhotoUrl || "");
      setForm((current) => ({ ...current, profilePhotoDataUrl: "", profilePhotoFileName: "", profilePhotoMimeType: "" }));
      return;
    }
    const allowedTypes = ["image/png", "image/jpeg", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      setMessage({ type: "error", text: "Use uma foto PNG, JPG ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A foto deve ter no máximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setPhotoPreview(dataUrl);
      setForm((current) => ({ ...current, profilePhotoDataUrl: dataUrl, profilePhotoFileName: file.name, profilePhotoMimeType: file.type }));
    };
    reader.readAsDataURL(file);
  };

  const save = async () => {
    setMessage(null);
    if (!form.fullName.trim()) {
      setMessage({ type: "error", text: "Informe seu nome." });
      return;
    }
    if (!form.email.trim()) {
      setMessage({ type: "error", text: "Informe seu e-mail." });
      return;
    }
    if (!form.phone.trim()) {
      setMessage({ type: "error", text: "Informe seu telefone." });
      return;
    }
    setSaving(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/me`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form)
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível salvar seu perfil.");
      setForm((current) => ({ ...current, ...data, profilePhotoDataUrl: "", profilePhotoFileName: "", profilePhotoMimeType: "" }));
      setPhotoPreview(data.profilePhotoUrl || photoPreview);
      await refreshSession?.();
      setMessage({ type: "success", text: "Seu perfil foi salvo com sucesso." });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível salvar seu perfil." });
    } finally {
      setSaving(false);
    }
  };

  const initials = initialsFromName(form.fullName);
  const canEditJobTitle = sessionUser?.roleKey === "company_admin";

  return (
    <Page label="Conta" title="Meu perfil" actions={<Button onClick={save} disabled={saving || loading}>{saving ? "Salvando..." : "Salvar meu perfil"}</Button>}>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading ? <Card><p>Carregando seu perfil...</p></Card> : (
        <div className="profileEditGrid">
          <Card>
            <h3>Foto do usuário</h3>
            <Field label="Foto de perfil" hint="Essa foto aparece no topo do sistema e na lista de usuários.">
              <div className="imageUploadField imageUploadHero">
                <div className="imageHeroPreview profilePhotoPreview">
                  {photoPreview ? <img src={photoPreview} alt="Prévia da sua foto" /> : <span>{initials}</span>}
                  <small>Seu perfil</small>
                </div>
                <div>
                  <input type="file" accept="image/png,image/jpeg,image/webp" onChange={handlePhotoChange} />
                  <small>{form.profilePhotoFileName || "Nenhuma nova foto selecionada"}</small>
                </div>
              </div>
            </Field>
          </Card>
          <div>
            <FormGrid>
              <Field label="Nome"><input value={form.fullName || ""} onChange={(event) => updateField("fullName", event.target.value)} /></Field>
              <Field label="E-mail"><input type="email" value={form.email || ""} onChange={(event) => updateField("email", event.target.value)} /></Field>
              <Field label="Telefone"><input value={form.phone || ""} onChange={(event) => updateField("phone", event.target.value)} /></Field>
              <Field label="Cargo ou função" hint={canEditJobTitle ? "Você pode atualizar este dado como administrador da empresa." : "Somente o administrador da empresa pode alterar o cargo."}><input value={form.jobTitle || ""} onChange={(event) => updateField("jobTitle", event.target.value)} readOnly={!canEditJobTitle} /></Field>
              <Field label="Empresa"><input value={form.companyName || ""} readOnly /></Field>
              <Field label="Perfil de acesso"><input value={form.roleName || ""} readOnly /></Field>
            </FormGrid>
          </div>
        </div>
      )}
    </Page>
  );
}

function CompanyProfileEdit({ refreshSession }) {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [logoPreview, setLogoPreview] = useState("");
  const [form, setForm] = useState({
    tradeName: "",
    cnpj: "",
    website: "",
    companySize: "",
    institutionalDescription: "",
    state: "",
    city: "",
    nationalCoverage: false,
    logoUrl: "",
    logoDataUrl: "",
    logoFileName: "",
    logoMimeType: ""
  });

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/companies/me`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => ({}));
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível carregar o perfil.");
        return data;
      })
      .then((data) => {
        setForm((current) => ({ ...current, ...data, logoDataUrl: "", logoFileName: "", logoMimeType: "" }));
        setLogoPreview(data.logoUrl || "");
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar o perfil da empresa." }))
      .finally(() => setLoading(false));
  }, []);

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleLogoChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      updateField("logoDataUrl", "");
      updateField("logoFileName", "");
      updateField("logoMimeType", "");
      setLogoPreview(form.logoUrl || "");
      return;
    }

    const allowedTypes = ["image/png", "image/jpeg", "image/webp", "image/svg+xml"];
    if (!allowedTypes.includes(file.type)) {
      setMessage({ type: "error", text: "Use uma logomarca PNG, JPG, WebP ou SVG." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A logomarca deve ter no maximo 5MB." });
      event.target.value = "";
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setLogoPreview(dataUrl);
      setForm((current) => ({
        ...current,
        logoDataUrl: dataUrl,
        logoFileName: file.name,
        logoMimeType: file.type
      }));
    };
    reader.readAsDataURL(file);
  };

  const handleSave = async () => {
    setMessage(null);
    setSaving(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/companies/me`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form)
      });
      const data = await response.json().catch(() => ({}));
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível salvar o perfil.");
      setForm((current) => ({ ...current, ...data, logoDataUrl: "", logoFileName: "", logoMimeType: "" }));
      setLogoPreview(data.logoUrl || logoPreview);
      await refreshSession?.();
      setMessage({ type: "success", text: "Perfil da empresa salvo com sucesso." });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível salvar o perfil." });
    } finally {
      setSaving(false);
    }
  };

  const initials = (form.tradeName || "Empresa").split(" ").map((word) => word[0]).join("").slice(0, 2).toUpperCase();

  return (
    <Page label="Empresa" title="Editar perfil da empresa" actions={<Button onClick={handleSave} disabled={saving || loading}>{saving ? "Salvando..." : "Salvar perfil"}</Button>}>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading ? <Card><p>Carregando perfil da empresa...</p></Card> : (
      <div className="profileEditGrid">
        <Card>
          <h3>Identidade visual</h3>
          <Field label="Logomarca da empresa" hint="PNG, JPG, WebP ou SVG. Usada no perfil público, comunidade e avaliação candidata.">
            <div className="imageUploadField imageUploadHero">
              <div className="imageHeroPreview">
                {logoPreview ? <img src={logoPreview} alt="Prévia da logomarca" /> : <span>{initials}</span>}
                <small>Identidade visual</small>
              </div>
              <div>
                <input type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" onChange={handleLogoChange} />
                <small>{form.logoFileName || "Nenhuma nova logomarca selecionada"}</small>
              </div>
            </div>
          </Field>
        </Card>
        <div>
          <FormGrid>
            <Field label="Nome fantasia" hint="Único."><input value={form.tradeName} readOnly /></Field>
            <Field label="CNPJ" hint="Único."><input value={formatCNPJ(form.cnpj)} readOnly /></Field>
            <Field label="Site"><input value={form.website || ""} onChange={(event) => updateField("website", event.target.value)} placeholder="https://www.empresa.com.br" /></Field>
            <Field label="Porte"><select value={form.companySize || ""} onChange={(event) => updateField("companySize", event.target.value)}><option value="">Não informado</option><option value="small">Pequena</option><option value="medium">Média</option><option value="large">Grande</option></select></Field>
            <Field label="UF"><StateSelect value={form.state || ""} onChange={(event) => updateField("state", event.target.value)} /></Field>
            <Field label="Cidade"><input value={form.city || ""} onChange={(event) => updateField("city", event.target.value)} /></Field>
          </FormGrid>
          <Field label="Descrição institucional"><textarea value={form.institutionalDescription || ""} onChange={(event) => updateField("institutionalDescription", event.target.value)} placeholder="Resumo profissional da atuação da empresa" /></Field>
          <label className="toggleLine"><input type="checkbox" checked={Boolean(form.nationalCoverage)} onChange={(event) => updateField("nationalCoverage", event.target.checked)} /> Atuação em todo o Brasil</label>
        </div>
      </div>
      )}
    </Page>
  );
}

function CompanyUsers({ navigate, openUserAction, openUserProfile, sessionUser }) {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ search: "", status: "", profile: "" });

  const loadUsers = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/company-users`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => ({}));
        if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível carregar os usuários.");
        return data;
      })
      .then((data) => {
        const safeUsers = Array.isArray(data) ? data : [];
        const companyUsers = sessionUser?.roleKey === "platform_admin" || !sessionUser?.companyId
          ? safeUsers
          : safeUsers.filter((user) => user.companyId === sessionUser.companyId);
        setUsers(companyUsers);
        if (safeUsers.length > 0 && companyUsers.length === 0 && sessionUser?.companyId) {
          setMessage({ type: "error", text: "Nenhum usuário da empresa logada foi encontrado nesta lista. Atualize o backend e tente novamente." });
        }
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar os usuários vinculados." }))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadUsers();
  }, []);

  const filteredUsers = users.filter((user) => {
    const term = filters.search.trim().toLowerCase();
    const matchesSearch = !term || [
      user.fullName,
      user.email,
      user.jobTitle,
      user.companyTradeName,
      user.accessProfileName
    ].some((value) => String(value || "").toLowerCase().includes(term));
    const matchesStatus = !filters.status || user.status === filters.status;
    const matchesProfile = !filters.profile || user.accessProfileKey === filters.profile;
    return matchesSearch && matchesStatus && matchesProfile;
  });

  return (
    <Page label="Empresa" title="Usuários vinculados" actions={<Button onClick={() => openUserProfile(null, "create")}>Adicionar usuário</Button>}>
      <Card className="notice">
        <strong>Permissões por perfil</strong>
        <p>As permissões não são marcadas individualmente. Cada usuário recebe um perfil de acesso, e o perfil define o que ele pode fazer.</p>
      </Card>
      <Card className="compactFilters userFiltersSticky">
        <FormGrid>
          <Field label="Buscar usuário"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Nome, e-mail, cargo ou empresa" /></Field>
          <Field label="Perfil"><select value={filters.profile} onChange={(event) => setFilters((current) => ({ ...current, profile: event.target.value }))}><option value="">Todos</option><option value="company_admin">Administrador da empresa</option><option value="commercial">Comercial / Relacionamento</option><option value="technical">Técnico</option><option value="reader">Leitor</option></select></Field>
          <Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option><option value="active">Ativo</option><option value="pending_invite">Convite pendente</option><option value="blocked">Bloqueado</option><option value="inactive">Inativo</option><option value="removed">Vínculo desativado</option></select></Field>
        </FormGrid>
      </Card>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading && <Card><p>Carregando usuários da empresa...</p></Card>}
      {!loading && users.length === 0 && <Card><p>Nenhum usuário vinculado encontrado para esta empresa.</p></Card>}
      {!loading && users.length > 0 && filteredUsers.length === 0 && <Card><p>Nenhum usuário encontrado com esses filtros.</p></Card>}
      {!loading && filteredUsers.length > 0 && <Table columns={["Nome", "E-mail", "Perfil", "Cargo", "Status", "Ações"]} rows={filteredUsers.map((user) => [
        <div className="userCell" key={`${user.id}-name`}>{user.profilePhotoUrl ? <span className="userAvatar"><img src={user.profilePhotoUrl} alt="" /></span> : <span className="userAvatar">{initialsFromName(user.fullName)}</span>}<div><strong>{user.fullName}</strong><small>{user.companyTradeName}</small></div></div>,
        user.email,
        user.accessProfileName,
        user.jobTitle || "-",
        <StatusBadge key={`${user.id}-status`} status={user.status} />,
        <UserRowActions key={`${user.id}-actions`} navigate={navigate} user={user} status={user.status} openUserAction={openUserAction} openUserProfile={openUserProfile} />
      ])} />}
    </Page>
  );
}

function StatusBadge({ status }) {
  const label = userStatusLabel(status);
  const className = status === "blocked" ? "closed" : status === "pending_invite" ? "review" : "open";
  return <span className={`statusPill ${className}`}>{label}</span>;
}

function UserRowActions({ navigate, user, status, openUserAction, openUserProfile }) {
  const blocked = status === "blocked";
  return (
    <div className="rowActions compactActions">
      <button className="iconButton secondaryIcon" title="Editar usuário" aria-label="Editar usuário" onClick={() => openUserProfile(user, "edit")}>{"\u270E"}</button>
      <button className={blocked ? "iconButton successIcon" : "iconButton warningIcon"} title={blocked ? "Desbloquear usuário" : "Bloquear usuário"} aria-label={blocked ? "Desbloquear usuário" : "Bloquear usuário"} onClick={() => openUserAction(user.id, user.fullName, blocked ? "unblock" : "block")}>{blocked ? "\u2713" : "!"}</button>
      <button className="iconButton dangerIcon" title="Desativar vínculo" aria-label="Desativar vínculo" onClick={() => openUserAction(user.id, user.fullName, "remove")}>{"\u00D7"}</button>
    </div>
  );
}

function CompanyUserProfile({ selectedUserProfile, navigate }) {
  const isEdit = selectedUserProfile?.mode === "edit";
  const user = selectedUserProfile?.user || {};
  const [profiles, setProfiles] = useState([]);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [setupUrl, setSetupUrl] = useState("");
  const [form, setForm] = useState({
    fullName: user.fullName || "",
    email: user.email || "",
    phone: user.phone || "",
    jobTitle: user.jobTitle || "",
    accessProfileKey: user.accessProfileKey || "commercial",
    status: user.status || "pending_invite",
    welcomeDraft: true,
    internalNote: "",
    profilePhotoUrl: user.profilePhotoUrl || "",
    profilePhotoDataUrl: "",
    profilePhotoFileName: "",
    profilePhotoMimeType: ""
  });
  const [photoPreview, setPhotoPreview] = useState(user.profilePhotoUrl || "");

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/access-profiles`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => ([]));
        if (!response.ok) throw new Error("Não foi possível carregar os perfis.");
        return Array.isArray(data) ? data : [];
      })
      .then((items) => setProfiles(items.filter((profile) => profile.key !== "platform_admin")))
      .catch(() => setProfiles([
        { key: "company_admin", name: "Administrador da empresa" },
        { key: "commercial", name: "Comercial / Relacionamento" },
        { key: "technical", name: "Técnico" },
        { key: "reader", name: "Leitor" }
      ]));
  }, []);

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handlePhotoChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setPhotoPreview(form.profilePhotoUrl || "");
      setForm((current) => ({ ...current, profilePhotoDataUrl: "", profilePhotoFileName: "", profilePhotoMimeType: "" }));
      return;
    }
    const allowedTypes = ["image/png", "image/jpeg", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      setMessage({ type: "error", text: "Use uma foto PNG, JPG ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A foto deve ter no máximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setPhotoPreview(dataUrl);
      setForm((current) => ({
        ...current,
        profilePhotoDataUrl: dataUrl,
        profilePhotoFileName: file.name,
        profilePhotoMimeType: file.type
      }));
    };
    reader.readAsDataURL(file);
  };

  const submit = async () => {
    setMessage(null);
    setSetupUrl("");
    if (!form.fullName.trim()) {
      setMessage({ type: "error", text: "Informe o nome completo do usuário." });
      return;
    }
    if (!form.email.trim()) {
      setMessage({ type: "error", text: "Informe o e-mail do usuário." });
      return;
    }
    if (!form.phone.trim()) {
      setMessage({ type: "error", text: "Informe o telefone do usuário." });
      return;
    }
    if (!isEdit && !form.profilePhotoDataUrl) {
      setMessage({ type: "error", text: "Inclua a foto do profissional antes de cadastrar." });
      return;
    }
    if (isEdit && !form.profilePhotoUrl && !form.profilePhotoDataUrl) {
      setMessage({ type: "error", text: "Inclua a foto do profissional antes de salvar." });
      return;
    }

    setSaving(true);
    try {
      const url = isEdit
        ? `${API_BASE_URL}/api/company-users/${encodeURIComponent(user.id)}`
        : `${API_BASE_URL}/api/company-users`;
      const response = await fetch(url, {
        method: isEdit ? "PUT" : "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          fullName: form.fullName,
          email: form.email,
          phone: form.phone,
          jobTitle: form.jobTitle,
          accessProfileKey: form.accessProfileKey,
          status: form.status,
          profilePhotoDataUrl: form.profilePhotoDataUrl,
          profilePhotoFileName: form.profilePhotoFileName,
          profilePhotoMimeType: form.profilePhotoMimeType
        })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível salvar o usuário.");
      if (!isEdit && data.setupUrl) {
        setSetupUrl(data.setupUrl);
        setMessage({ type: "success", text: data.draftPublicationId ? "Usuário cadastrado. Um rascunho de apresentação foi criado em Minhas publicações. Copie o link abaixo e envie para ele definir a senha." : "Usuário cadastrado. Copie o link abaixo e envie para ele definir a senha." });
      } else {
        setMessage({ type: "success", text: "Usuário atualizado com sucesso." });
        setTimeout(() => navigate?.("company-users"), 700);
      }
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível salvar o usuário." });
    } finally {
      setSaving(false);
    }
  };

  const previewName = form.fullName || "Novo profissional";
  const previewJob = form.jobTitle || "cargo a definir";

  return (
    <Page label="Empresa" title={isEdit ? "Editar usuário vinculado" : "Cadastrar usuário vinculado"} actions={<Button onClick={submit} disabled={saving}>{saving ? "Salvando..." : isEdit ? "Salvar alterações" : "Cadastrar usuário"}</Button>}>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {setupUrl && (
        <Card className="inviteLinkBox">
          <h3>Link de primeiro acesso</h3>
          <p>Envie este link para o usuário definir a senha e liberar o login.</p>
          <input value={setupUrl} readOnly />
          <div className="actions">
            <Button onClick={() => navigator.clipboard.writeText(setupUrl)}>Copiar link</Button>
            <a className="btn secondary" href={setupUrl}>Abrir link</a>
            <Button variant="secondary" onClick={() => navigate?.("company-users")}>Voltar para lista</Button>
          </div>
        </Card>
      )}
      <FormGrid>
        <Field label="Nome completo"><input value={form.fullName} onChange={(event) => updateField("fullName", event.target.value)} placeholder="Nome do usuário" /></Field>
        <Field label="E-mail"><input type="email" value={form.email} onChange={(event) => updateField("email", event.target.value)} placeholder="usuario@empresa.com.br" /></Field>
        <Field label="Telefone"><input value={form.phone} onChange={(event) => updateField("phone", event.target.value)} placeholder="(00) 00000-0000" /></Field>
        <Field label="Cargo ou função"><input value={form.jobTitle} onChange={(event) => updateField("jobTitle", event.target.value)} placeholder="Ex.: Coordenador técnico" /></Field>
        <Field label="Foto do profissional" hint="Usada no perfil interno, topo do sistema e rascunho de boas-vindas.">
          <div className="imageUploadField">
            <div className="imageMiniPreview profilePhotoPreview">
              {photoPreview ? <img src={photoPreview} alt="Prévia da foto do profissional" /> : <span>{initialsFromName(previewName)}</span>}
            </div>
            <div>
              <input type="file" accept="image/png,image/jpeg,image/webp" onChange={handlePhotoChange} />
              <small>{form.profilePhotoFileName || "Nenhuma nova foto selecionada"}</small>
            </div>
          </div>
        </Field>
        <Field label="Perfil de acesso"><select value={form.accessProfileKey} onChange={(event) => updateField("accessProfileKey", event.target.value)}>{profiles.map((profile) => <option key={profile.key} value={profile.key}>{profile.name}</option>)}</select></Field>
        <Field label="Status"><select value={form.status} onChange={(event) => updateField("status", event.target.value)}><option value="pending_invite">Convite pendente</option><option value="active">Ativo</option><option value="blocked">Bloqueado</option><option value="inactive">Inativo</option></select></Field>
      </FormGrid>
      {!isEdit && (
        <Card className="notice userWelcomePost">
          <div>
            <h3>Publicação automática de novo profissional</h3>
            <p>Ao enviar o convite, a LicitaHub pode criar um rascunho na categoria Equipe comercial para anunciar o novo membro com foto, nome e cargo.</p>
          </div>
          <label className="inlineCheck"><input type="checkbox" checked={form.welcomeDraft} onChange={(event) => updateField("welcomeDraft", event.target.checked)} /> Criar rascunho de boas-vindas</label>
          <div className="welcomePostPreview">
            <span className="avatar">{photoPreview ? <img src={photoPreview} alt="" /> : initialsFromName(previewName)}</span>
            <div>
              <strong>Prévia do rascunho</strong>
              <p>{previewName} passa a integrar a equipe da empresa como {previewJob}, fortalecendo nossa presença na comunidade LicitaHub.</p>
            </div>
          </div>
        </Card>
      )}
      <Field label="Observação interna"><textarea value={form.internalNote} onChange={(event) => updateField("internalNote", event.target.value)} placeholder="Anotação visível apenas para administradores da empresa" /></Field>
    </Page>
  );
}

function CompanyUserAccessConfirm({ navigate, selectedUserAction, updateUserStatus, mode }) {
  const isUnlock = mode === "unblock";
  const userName = selectedUserAction?.name || "usuário selecionado";
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const confirm = async () => {
    if (!selectedUserAction?.id) return;
    setSaving(true);
    setMessage(null);
    try {
      const action = isUnlock ? "unblock" : "block";
      const response = await fetch(`${API_BASE_URL}/api/company-users/${encodeURIComponent(selectedUserAction.id)}/${action}`, {
        method: "PATCH",
        credentials: "include"
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível atualizar o usuário.");
      updateUserStatus(selectedUserAction.id, isUnlock ? "Ativo" : "Bloqueado");
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível concluir a ação." });
    } finally {
      setSaving(false);
    }
  };
  return (
    <Page label="Empresa" title={isUnlock ? "Confirmar desbloqueio de usuário" : "Confirmar bloqueio de usuário"} actions={<Button variant={isUnlock ? "primary" : "danger"} onClick={confirm} disabled={saving}>{saving ? "Salvando..." : isUnlock ? "Confirmar desbloqueio" : "Confirmar bloqueio"}</Button>}>
      {message && <Card className="formFeedback dangerNotice"><p>{message.text}</p></Card>}
      <Card className={isUnlock ? "notice" : "dangerCard"}>
        <h3>{isUnlock ? `Desbloquear acesso de ${userName}?` : `Bloquear acesso de ${userName}?`}</h3>
        <p>{isUnlock ? "Ao confirmar, a pessoa volta a acessar a conta da empresa conforme seu perfil de acesso." : "Ao confirmar, a pessoa deixa de acessar a conta da empresa até que um administrador faça o desbloqueio."}</p>
        <div className="impactList">
          <span>{isUnlock ? "O usuário volta a entrar na plataforma." : "O usuário não conseguirá entrar na área da empresa."}</span>
          <span>Histórico, publicações e registros anteriores continuam preservados.</span>
          <span>{isUnlock ? "O botão voltará a aparecer como Bloquear na lista." : "O botão passará a aparecer como Desbloquear na lista."}</span>
        </div>
        <div className="actions"><Button variant={isUnlock ? "primary" : "danger"} onClick={confirm} disabled={saving}>{saving ? "Salvando..." : isUnlock ? "Confirmar desbloqueio" : "Confirmar bloqueio"}</Button><Button variant="secondary" onClick={() => navigate("company-users")}>Cancelar</Button></div>
      </Card>
    </Page>
  );
}

function CompanyUserDelete({ navigate, selectedUserAction }) {
  const userName = selectedUserAction?.name || "usuário selecionado";
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const confirm = async () => {
    if (!selectedUserAction?.id) return;
    setSaving(true);
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/company-users/${encodeURIComponent(selectedUserAction.id)}/remove`, {
        method: "PATCH",
        credentials: "include"
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok || data?.error) throw new Error(data?.error || "Não foi possível desativar o vínculo.");
      navigate("company-users");
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível desativar o vínculo." });
    } finally {
      setSaving(false);
    }
  };
  return (
    <Page label="Empresa" title="Desativar vínculo do usuário" actions={<Button variant="danger" onClick={confirm} disabled={saving}>{saving ? "Desativando..." : "Desativar vínculo"}</Button>}>
      {message && <Card className="formFeedback dangerNotice"><p>{message.text}</p></Card>}
      <Card className="dangerCard">
        <h3>Desativar vínculo de {userName}?</h3>
        <p>Esta ação remove o usuário da operação da empresa. Para voltar ao sistema depois, será necessário um novo cadastro.</p>
        <div className="impactList">
          <span>Ela não poderá mais acessar a empresa.</span>
          <span>Publicações, interesses e ações antigas continuam registrados.</span>
          <span>O usuário deixa de aparecer na lista de usuários vinculados.</span>
        </div>
        <div className="actions"><Button variant="danger" onClick={confirm} disabled={saving}>{saving ? "Desativando..." : "Desativar vínculo"}</Button><Button variant="secondary" onClick={() => navigate("company-users")}>Cancelar</Button></div>
      </Card>
    </Page>
  );
}

function CommunityHome({ sessionUser, navigate }) {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ company: "", categorySlug: "all", state: "BR" });
  const [form, setForm] = useState({ categorySlug: "noticias", content: "", mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" });
  const [imagePreview, setImagePreview] = useState("");
  const [publishing, setPublishing] = useState(false);

  const loadPosts = () => {
    setLoading(true);
    const params = new URLSearchParams();
    if (filters.company.trim()) params.set("company", filters.company.trim());
    if (filters.categorySlug !== "all") params.set("categorySlug", filters.categorySlug);
    if (filters.state !== "BR") params.set("state", filters.state);

    fetch(`${API_BASE_URL}/api/community/posts?${params.toString()}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => []);
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar a comunidade.");
        return data;
      })
      .then((data) => {
        setPosts(Array.isArray(data) ? data : []);
        setMessage(null);
      })
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar a comunidade." }))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    const delay = window.setTimeout(loadPosts, filters.company.trim() ? 280 : 0);
    return () => window.clearTimeout(delay);
  }, [filters.company, filters.categorySlug, filters.state]);

  const updateFilter = (field, value) => setFilters((current) => ({ ...current, [field]: value }));
  const updateForm = (field, value) => setForm((current) => ({ ...current, [field]: value }));

  const handlePostImage = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setImagePreview("");
      setForm((current) => ({ ...current, mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" }));
      return;
    }
    if (!["image/png", "image/jpeg", "image/gif", "image/webp"].includes(file.type)) {
      setMessage({ type: "error", text: "Use uma imagem PNG, JPG, GIF ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A imagem deve ter no maximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setImagePreview(dataUrl);
      setForm((current) => ({ ...current, mainImageDataUrl: dataUrl, mainImageFileName: file.name, mainImageMimeType: file.type }));
    };
    reader.readAsDataURL(file);
  };

  const publishPost = async () => {
    setMessage(null);
    if (!form.content.trim()) {
      setMessage({ type: "error", text: "Escreva o texto da publicação antes de publicar." });
      return;
    }
    setPublishing(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ ...form, visibility: "both" })
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível publicar.");
      setForm({ categorySlug: "noticias", content: "", mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" });
      setImagePreview("");
      setMessage({ type: "success", text: "Publicação enviada para a comunidade." });
      loadPosts();
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível publicar agora." });
    } finally {
      setPublishing(false);
    }
  };

  const replacePost = (updatedPost) => {
    setPosts((current) => current.map((post) => post.id === updatedPost.id ? updatedPost : post));
  };

  return (
    <Page label="Comunidade" title="Rede de empresas">
      <div className="communityWorkspace">
        <div className="communityFeedLayout">
          <aside className="communityFilterRail" aria-label="Filtros da comunidade">
            <div className="communityToolbar socialFilters communityFiltersSticky">
              <div className="communityFilterHeading"><span className="eyebrow">Comunidade</span><strong>Filtrar publicações</strong></div>
              <div className="communitySearch">
                <span>Empresa</span>
                <input value={filters.company} onChange={(event) => updateFilter("company", event.target.value)} placeholder="Nome da empresa" />
              </div>
              <label className="communityFilterField"><span>Tipo de publicação</span><select value={filters.categorySlug} onChange={(event) => updateFilter("categorySlug", event.target.value)}><option value="all">Todos os tipos</option>{communityCategoryOptions.map((option) => <option value={option.slug} key={option.slug}>{option.label}</option>)}</select></label>
              <label className="communityFilterField"><span>Região</span><select value={filters.state} onChange={(event) => updateFilter("state", event.target.value)}><option value="BR">Brasil inteiro</option><option value="SP">SP</option><option value="MG">MG</option><option value="PR">PR</option><option value="RJ">RJ</option><option value="BA">BA</option><option value="GO">GO</option><option value="DF">DF</option></select></label>
            </div>
          </aside>

          <main className="communityMainColumn">
            <div className="communityTopRow">
              <Card className="composer modern socialComposer">
                <LogoSlot src={sessionUser?.companyLogoUrl} initials={String(sessionUser?.companyName || "EC").slice(0, 2).toUpperCase()} size="sm" label={`Logo da ${sessionUser?.companyName || "empresa"}`} />
                <div>
                  <textarea value={form.content} onChange={(event) => updateForm("content", event.target.value)} placeholder="O que sua empresa quer publicar na comunidade?" />
                  <div className="composerFields">
                    <select value={form.categorySlug} onChange={(event) => updateForm("categorySlug", event.target.value)}>{communityCategoryOptions.map((option) => <option value={option.slug} key={option.slug}>{option.label}</option>)}</select>
                    <div className="inlineImagePicker">
                      <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={handlePostImage} />
                      <span>{form.mainImageFileName || "Imagem"}</span>
                    </div>
                    <button className="iconButton publishIcon" title="Publicar" aria-label="Publicar" disabled={publishing} onClick={publishPost}>{"\u2191"}</button>
                  </div>
                  {imagePreview && <div className="composerImagePreview"><img src={imagePreview} alt="Prévia da imagem da publicação" /></div>}
                </div>
              </Card>
            </div>

            {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}

            <section className="communityFeed socialFeedWide">
              <div className="socialFeed">
                {loading && <Card><p>Carregando publicações...</p></Card>}
                {!loading && posts.length === 0 && <Card><p>Nenhuma publicação encontrada para estes filtros.</p></Card>}
                {!loading && posts.map((post) => <PostCard key={post.id} {...post} onPostUpdated={replacePost} onOpenCompany={(companyId) => navigate(`company-public-profile?companyId=${encodeURIComponent(companyId)}`)} />)}
              </div>
            </section>
          </main>
        </div>

      </div>
    </Page>
  );
}

function PostCard({ id, companyId = "", category, company, region = "BR", companyLogoUrl = "", title, text = "Texto de exemplo da publicação com contexto empresarial, técnico e institucional.", likes = [], comments = [], imageLabel, imageUrl = "", liked = false, saved = false, likeCount = 0, commentCount = 0, onOpenPublication, onOpenCompany, onPostUpdated }) {
  const [showLikes, setShowLikes] = useState(false);
  const [isLiked, setIsLiked] = useState(Boolean(liked));
  const [isSaved, setIsSaved] = useState(Boolean(saved));
  const [showComments, setShowComments] = useState(false);
  const [likeNames, setLikeNames] = useState(Array.isArray(likes) ? likes : []);
  const [postComments, setPostComments] = useState(Array.isArray(comments) ? comments : []);
  const [likesTotal, setLikesTotal] = useState(Number(likeCount || 0));
  const [commentsTotal, setCommentsTotal] = useState(Number(commentCount || 0));
  const [commentText, setCommentText] = useState("");
  const [editingCommentId, setEditingCommentId] = useState("");
  const [editingCommentText, setEditingCommentText] = useState("");
  const [savingAction, setSavingAction] = useState("");
  const displayedLikeCount = Math.max(likesTotal, likeNames.length);
  const displayedCommentCount = Math.max(commentsTotal, postComments.length);
  const stopPostClick = (event) => event.stopPropagation();

  useEffect(() => {
    setIsLiked(Boolean(liked));
    setIsSaved(Boolean(saved));
    setLikeNames(Array.isArray(likes) ? likes : []);
    setPostComments(Array.isArray(comments) ? comments : []);
    setLikesTotal(Number(likeCount || 0));
    setCommentsTotal(Number(commentCount || 0));
  }, [liked, saved, likes, comments, likeCount, commentCount]);

  const toggleAction = async (action) => {
    if (!id || savingAction) return;
    setSavingAction(action);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(id)}/${action}`, { method: "POST", credentials: "include" });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Falha na ação.");
      if (action === "like") {
        setIsLiked(Boolean(result.liked));
        if (typeof result.likeCount === "number") setLikesTotal(result.likeCount);
      }
      if (action === "favorite") setIsSaved(Boolean(result.favorited));
    } finally {
      setSavingAction("");
    }
  };

  const sendComment = async (event) => {
    stopPostClick(event);
    if (!commentText.trim() || !id) return;
    setSavingAction("comment");
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(id)}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content: commentText.trim() })
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível comentar.");
      setPostComments((current) => [...current, result]);
      setCommentsTotal((current) => current + 1);
      setCommentText("");
      setShowComments(true);
    } finally {
      setSavingAction("");
    }
  };

  const startEditComment = (event, comment) => {
    stopPostClick(event);
    setEditingCommentId(comment.id);
    setEditingCommentText(comment.text || "");
  };

  const cancelEditComment = (event) => {
    stopPostClick(event);
    setEditingCommentId("");
    setEditingCommentText("");
  };

  const saveCommentEdit = async (event, comment) => {
    stopPostClick(event);
    if (!editingCommentText.trim() || !id || !comment.id) return;
    setSavingAction(`comment-edit-${comment.id}`);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(id)}/comments/${encodeURIComponent(comment.id)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content: editingCommentText.trim() })
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível editar o comentário.");
      setPostComments((current) => current.map((item) => item.id === comment.id ? { ...item, ...result } : item));
      setEditingCommentId("");
      setEditingCommentText("");
    } finally {
      setSavingAction("");
    }
  };

  const deleteComment = async (event, comment) => {
    stopPostClick(event);
    if (!id || !comment.id) return;
    if (!window.confirm("Excluir este comentário?")) return;
    setSavingAction(`comment-delete-${comment.id}`);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(id)}/comments/${encodeURIComponent(comment.id)}`, {
        method: "DELETE",
        credentials: "include"
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível excluir o comentário.");
      setPostComments((current) => current.filter((item) => item.id !== comment.id));
      setCommentsTotal((current) => Math.max(0, current - 1));
    } finally {
      setSavingAction("");
    }
  };

  return (
    <Card className={`postCard ${onOpenPublication ? "clickablePostCard" : ""}`} onClick={onOpenPublication ? () => onOpenPublication(id) : undefined}>
      <div className="postHead">
        <button type="button" className="postCompanyIdentity" disabled={!companyId || !onOpenCompany} title={companyId && onOpenCompany ? `Ver perfil público da ${company}` : undefined} onClick={(event) => { stopPostClick(event); if (companyId && onOpenCompany) onOpenCompany(companyId); }}>
          <LogoSlot src={companyLogoUrl} initials={String(company || "EC").split(" ").map((word) => word[0]).join("").slice(0, 2)} size="sm" label={`Logo da ${company}`} />
          <span>
            <strong>{company}</strong>
            <small>{category} | {region}</small>
          </span>
        </button>
      </div>
      <div className={`postImage ${imageUrl ? "postImageWithMedia" : "postImageTitle"}`}>
        {imageUrl ? <img src={imageUrl} alt={title || category} loading="lazy" decoding="async" /> : <span>{title || imageLabel || category}</span>}
      </div>
      {imageUrl && title && <h3>{title}</h3>}
      <p>{text}</p>
      <div className="engagementSummary">
        <button onClick={(event) => { stopPostClick(event); setShowLikes(!showLikes); }}>{displayedLikeCount} curtidas</button>
        <button type="button" onClick={(event) => { stopPostClick(event); setShowComments((open) => !open); }}>{displayedCommentCount} comentários</button>
      </div>
      {showLikes && <div className="likedBy">{likeNames.length === 0 ? <span>Nenhuma curtida ainda</span> : likeNames.map((name) => <span key={name}>{name}</span>)}</div>}
      {showComments && (
        <div className="commentPanel">
          <div className="commentPanelHeader">
            <strong>Comentários</strong>
            <button type="button" className="commentCloseButton" title="Recolher comentários" aria-label="Recolher comentários" onClick={(event) => { stopPostClick(event); setShowComments(false); }}>{"\u00D7"}</button>
          </div>
          <div className="commentComposer">
            <input value={commentText} onChange={(event) => setCommentText(event.target.value)} placeholder="Escreva um comentário..." onClick={stopPostClick} onKeyDown={(event) => { if (event.key === "Enter") sendComment(event); }} />
            <button type="button" className="iconButton publishIcon" title="Enviar comentário" aria-label="Enviar comentário" disabled={savingAction === "comment"} onClick={sendComment}>{"\u2191"}</button>
          </div>
          <div className="commentList">
            {postComments.length === 0 && <p className="emptyComments">Nenhum comentário ainda.</p>}
            {postComments.map((comment) => (
              <div className="commentItem" key={comment.id || `${comment.company}-${comment.text}`}>
                <div className="commentItemHeader">
                  <strong>{comment.userName ? `${comment.userName}${comment.company ? ` · ${comment.company}` : ""}` : comment.company}</strong>
                  <div className="commentItemActions">
                    {comment.canEdit && <button type="button" className="iconButton secondaryIcon" title="Editar comentário" aria-label="Editar comentário" onClick={(event) => startEditComment(event, comment)}>{"\u270E"}</button>}
                    {comment.canDelete && <button type="button" className="iconButton dangerIcon" title="Excluir comentário" aria-label="Excluir comentário" onClick={(event) => deleteComment(event, comment)}>{"\u00D7"}</button>}
                  </div>
                </div>
                {editingCommentId === comment.id ? (
                  <div className="commentEditBox">
                    <input value={editingCommentText} onChange={(event) => setEditingCommentText(event.target.value)} onClick={stopPostClick} />
                    <button type="button" className="iconButton publishIcon" title="Salvar comentário" aria-label="Salvar comentário" disabled={savingAction === `comment-edit-${comment.id}`} onClick={(event) => saveCommentEdit(event, comment)}>{"\u2713"}</button>
                    <button type="button" className="iconButton secondaryIcon" title="Cancelar edição" aria-label="Cancelar edição" onClick={cancelEditComment}>{"\u00D7"}</button>
                  </div>
                ) : (
                  <p>{comment.text}</p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
      <div className="postActions compactPostActions">
        <button className={`iconButton likeIcon ${isLiked ? "active" : ""}`} title={isLiked ? "Descurtir publicação" : "Curtir publicação"} aria-label={isLiked ? "Descurtir publicação" : "Curtir publicação"} disabled={savingAction === "like"} onClick={(event) => { stopPostClick(event); toggleAction("like"); }}>{isLiked ? "\u2665" : "\u2661"}</button>
        <button type="button" className={`iconButton commentIcon ${showComments ? "active" : ""}`} title={showComments ? "Ocultar comentários" : "Comentar publicação"} aria-label={showComments ? "Ocultar comentários" : "Comentar publicação"} onClick={(event) => { stopPostClick(event); setShowComments((open) => !open); }}>{"\u21B5"}</button>
        <button className={`iconButton saveIcon ${isSaved ? "active" : ""}`} title={isSaved ? "Remover dos favoritos" : "Favoritar publicação"} aria-label={isSaved ? "Remover dos favoritos" : "Favoritar publicação"} disabled={savingAction === "favorite"} onClick={(event) => { stopPostClick(event); toggleAction("favorite"); }}>{isSaved ? "\u2605" : "\u2606"}</button>
      </div>
    </Card>
  );
}

function CompanySuggestion({ name, area }) {
  return (
    <div className="suggestItem">
      <LogoSlot initials={name.split(" ").map((word) => word[0]).join("").slice(0, 2)} size="sm" label={`Logo da ${name}`} />
      <div><strong>{name}</strong><small>{area}</small></div>
      <button>Ver</button>
    </div>
  );
}

function CompanyPublicProfile({ navigate, openPublicationManager, sessionUser, openChatForUser }) {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [professionalsOpen, setProfessionalsOpen] = useState(false);
  const [selectedProfessionalId, setSelectedProfessionalId] = useState("");
  const requestedCompanyId = currentHashParams().get("companyId") || "";

  useEffect(() => {
    setLoading(true);
    setError("");
    setProfile(null);
    setProfessionalsOpen(false);
    setSelectedProfessionalId("");
    const profilePath = requestedCompanyId
      ? `/api/companies/${encodeURIComponent(requestedCompanyId)}/public-profile`
      : "/api/companies/me/public-profile";
    fetch(`${API_BASE_URL}${profilePath}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o perfil público.");
        return data;
      })
      .then((data) => {
        setProfile(data);
        setError("");
      })
      .catch((err) => setError(err.message || "Não foi possível carregar o perfil público."))
      .finally(() => setLoading(false));
  }, [requestedCompanyId]);

  const companySizeLabel = {
    small: "Pequena empresa",
    medium: "Média empresa",
    large: "Grande empresa"
  }[profile?.companySize] || "Porte não informado";
  const locationLabel = profile?.nationalCoverage ? "Atuação em todo o Brasil" : [profile?.city, profile?.state].filter(Boolean).join(" - ") || "Atuação não informada";
  const professionals = Array.isArray(profile?.professionals) ? profile.professionals : [];
  const posts = Array.isArray(profile?.posts) ? profile.posts : [];
  const isOwnProfile = Boolean(profile?.id && profile.id === sessionUser?.companyId);

  return (
    <Page label="Comunidade" title="Perfil público da empresa" actions={isOwnProfile ? <Button onClick={() => navigate("publication-new")}>Criar publicação</Button> : null}>
      {loading && <Card className="formFeedback"><p>Carregando perfil público...</p></Card>}
      {error && <Card className="formFeedback dangerNotice"><p>{error}</p></Card>}
      {profile && (
        <>
          <div className="profileHero publicCompanyHero">
            <div className="publicCompanyIdentity">
              <LogoSlot src={profile.logoUrl} initials={String(profile.tradeName || "EC").slice(0, 2).toUpperCase()} size="lg" label={`Logo da ${profile.tradeName}`} />
              <div>
                <h3>{profile.tradeName}</h3>
                <p>{locationLabel}</p>
              </div>
            </div>
            <div className="publicProfileActions">
              {profile.website ? <a className="btn secondary" href={profile.website.startsWith("http") ? profile.website : `https://${profile.website}`} target="_blank" rel="noreferrer">Acessar site</a> : <span className="statusPill review">Site não informado</span>}
            </div>
          </div>

          <div className="publicCompanyGrid">
            <Card>
              <h3>Descrição institucional</h3>
              <p>{profile.institutionalDescription || "A empresa ainda não cadastrou sua descrição institucional."}</p>
            </Card>
            <Card>
              <h3>Dados de atuação</h3>
              <div className="profileFactList">
                <span><strong>Porte</strong>{companySizeLabel}</span>
                <span><strong>Atuação</strong>{locationLabel}</span>
              </div>
            </Card>
          </div>

          <Card className="professionalsPanel">
            <button type="button" className="drawerHeaderButton" onClick={() => setProfessionalsOpen((open) => !open)}>
              <span><strong>Profissionais vinculados</strong><small>{professionals.length} profissionais ativos</small></span>
              <span>{professionalsOpen ? "\u2212" : "+"}</span>
            </button>
            {professionalsOpen && (
              <div className="professionalsGrid">
                {professionals.length === 0 && <p className="emptyComments">Nenhum profissional ativo para exibir.</p>}
                {professionals.map((person) => (
                  <button type="button" className={`professionalCard professionalCardButton ${selectedProfessionalId === person.id ? "selected" : ""}`} key={person.id} onClick={() => setSelectedProfessionalId((current) => current === person.id ? "" : person.id)}>
                    <span className="userAvatar large">{person.profilePhotoUrl ? <img src={person.profilePhotoUrl} alt="" /> : initialsFromName(person.fullName)}</span>
                    <div>
                      <strong>{person.fullName}</strong>
                      <small>{person.jobTitle || "Cargo não informado"}</small>
                      <span>Ver informações profissionais</span>
                    </div>
                  </button>
                ))}
                {selectedProfessionalId && (() => {
                  const person = professionals.find((item) => item.id === selectedProfessionalId);
                  if (!person) return null;
                  return <div className="professionalDetail" key={`detail-${person.id}`}>
                    <span className="userAvatar large">{person.profilePhotoUrl ? <img src={person.profilePhotoUrl} alt="" /> : initialsFromName(person.fullName)}</span>
                    <div className="professionalDetailContent"><span className="eyebrow">Profissional vinculado</span><h3>{person.fullName}</h3><p>{person.jobTitle || "Cargo não informado"}</p><div className="professionalContactLinks">{person.email ? <a href={`mailto:${person.email}`}>{person.email}</a> : <span>E-mail não informado</span>}{person.phone ? <a href={`tel:${String(person.phone).replace(/\D/g, "")}`}>{person.phone}</a> : <span>Telefone não informado</span>}</div></div>
                    <div className="professionalDetailActions">{person.id !== sessionUser?.userId && <button type="button" className="iconButton secondaryIcon" title={`Conversar com ${person.fullName}`} aria-label={`Conversar com ${person.fullName}`} onClick={() => openChatForUser(person)}><i className="taskChatIndicator" aria-hidden="true"><b /><b /><b /></i></button>}<button type="button" className="iconButton secondaryIcon" title="Fechar informações do profissional" aria-label="Fechar informações do profissional" onClick={() => setSelectedProfessionalId("")}>×</button></div>
                  </div>;
                })()}
              </div>
            )}
          </Card>

          <div className="sectionTitleRow">
            <h3>Publicações da empresa</h3>
            {isOwnProfile && <Button variant="secondary" onClick={() => navigate("publication-list")}>Minhas publicações</Button>}
          </div>
          {posts.length === 0 && <Card><p>Esta empresa ainda não possui publicações no perfil público.</p></Card>}
          <div className="grid two">
            {posts.map((post) => <PostCard key={post.id} {...post} onOpenPublication={isOwnProfile ? openPublicationManager : undefined} />)}
          </div>
        </>
      )}
    </Page>
  );
}

function PublicationNew({ openPublicationManager, navigate }) {
  const [form, setForm] = useState({ title: "", categorySlug: "noticias", visibility: "both", content: "", mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" });
  const [imagePreview, setImagePreview] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const updateField = (field, value) => setForm((current) => ({ ...current, [field]: value }));

  const handleImageChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setImagePreview("");
      setForm((current) => ({ ...current, mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" }));
      return;
    }
    if (!["image/png", "image/jpeg", "image/gif", "image/webp"].includes(file.type)) {
      setMessage({ type: "error", text: "Use uma imagem PNG, JPG, GIF ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A imagem deve ter no maximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setImagePreview(dataUrl);
      setForm((current) => ({ ...current, mainImageDataUrl: dataUrl, mainImageFileName: file.name, mainImageMimeType: file.type }));
    };
    reader.readAsDataURL(file);
  };

  const submit = async () => {
    setMessage(null);
    if (!form.content.trim()) {
      setMessage({ type: "error", text: "Informe o texto da publicação." });
      return;
    }
    setSaving(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form)
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível criar a publicação.");
      if (result.id) {
        openPublicationManager(result.id);
      } else {
        navigate("publication-list");
      }
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível publicar agora." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Page label="Comunidade" title="Criar publicação">
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      <FormGrid>
        <Field label="Tipo"><select value={form.categorySlug} onChange={(event) => updateField("categorySlug", event.target.value)}>{communityCategoryOptions.map((option) => <option value={option.slug} key={option.slug}>{option.label}</option>)}</select></Field>
        <Field label="Visibilidade"><select value={form.visibility} onChange={(event) => updateField("visibility", event.target.value)}><option value="both">Comunidade e perfil</option><option value="community">Apenas comunidade</option><option value="profile">Apenas perfil público</option></select></Field>
        <Field label="Título" hint={`${form.title.length}/100 caracteres`}><input value={form.title} onChange={(event) => updateField("title", event.target.value)} placeholder="Título da publicação" maxLength="100" /></Field>
      </FormGrid>
      <Field label="Imagem da publicação" hint="Essa imagem aparece no feed da comunidade, no perfil público e em minhas publicações.">
        <div className="imageUploadField">
          <div className="imageMiniPreview">{imagePreview ? <img src={imagePreview} alt="Prévia da imagem selecionada" /> : <span>IMG</span>}</div>
          <div>
            <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={handleImageChange} />
            <small>{form.mainImageFileName || "Nenhuma imagem selecionada"}</small>
          </div>
        </div>
      </Field>
      <Field label="Texto"><textarea value={form.content} onChange={(event) => updateField("content", event.target.value)} placeholder="Escreva a atualização institucional" /></Field>
      <div className="formActionBar">
        <Button onClick={submit} disabled={saving}>{saving ? "Publicando..." : "Publicar"}</Button>
      </div>
    </Page>
  );
}

function PublicationList({ selectedPublicationId }) {
  const [publications, setPublications] = useState([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ search: "", category: "all", status: "all", sort: "published_desc" });

  const removePublication = (id) => setPublications((current) => current.filter((item) => item.id !== id));
  const updatePublication = (updated) => {
    setPublications((current) => current.map((item) => item.id === updated.id ? { ...item, ...updated } : item));
  };

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/community/posts?scope=mine`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => []);
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar suas publicações.");
        return data;
      })
      .then((data) => setPublications(Array.isArray(data) ? data : []))
      .catch((error) => setMessage({ type: "error", text: error.message || "Não foi possível carregar suas publicações." }))
      .finally(() => setLoading(false));
  }, []);

  const categories = Array.from(new Set(publications.map((publication) => publication.type || publication.category).filter(Boolean))).sort((a, b) => String(a).localeCompare(String(b), "pt-BR"));
  const filteredPublications = publications.filter((publication) => {
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    const category = publication.type || publication.category || "";
    const matchesSearch = !term || [publication.title, publication.content, publication.text, category].some((value) => String(value || "").toLocaleLowerCase("pt-BR").includes(term));
    return matchesSearch && (filters.category === "all" || category === filters.category) && (filters.status === "all" || publication.status === filters.status);
  }).sort((a, b) => {
    if (filters.sort === "likes_desc") return Number(b.likeCount || 0) - Number(a.likeCount || 0);
    if (filters.sort === "comments_desc") return Number(b.commentCount || 0) - Number(a.commentCount || 0);
    return new Date(b.publishedAt || b.createdAt || 0).getTime() - new Date(a.publishedAt || a.createdAt || 0).getTime();
  });

  return (
    <Page label="Comunidade" title="Minhas publicações">
      {message && <Card className="formFeedback dangerNotice"><p>{message.text}</p></Card>}
      <Card className="compactFilters publicationFiltersSticky"><FormGrid><Field label="Buscar publicação"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Título, texto ou categoria" /></Field><Field label="Categoria"><select value={filters.category} onChange={(event) => setFilters((current) => ({ ...current, category: event.target.value }))}><option value="all">Todas</option>{categories.map((category) => <option value={category} key={category}>{category}</option>)}</select></Field><Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="all">Todos</option><option value="published">Publicada</option><option value="archived">Arquivada</option><option value="draft">Rascunho</option></select></Field><Field label="Ordenar por"><select value={filters.sort} onChange={(event) => setFilters((current) => ({ ...current, sort: event.target.value }))}><option value="published_desc">Mais recentes</option><option value="likes_desc">Mais curtidas</option><option value="comments_desc">Mais comentários</option></select></Field></FormGrid></Card>
      {loading && <Card><p>Carregando suas publicações...</p></Card>}
      {!loading && publications.length === 0 && <Card><p>Sua empresa ainda não possui publicações.</p></Card>}
      {!loading && publications.length > 0 && filteredPublications.length === 0 && <Card><p>Nenhuma publicação encontrada com esses filtros.</p></Card>}
      <div className="publicationManager">
        {filteredPublications.map((publication) => <PublicationManagerCard key={publication.id} publication={publication} initiallyOpen={publication.id === selectedPublicationId} highlighted={publication.id === selectedPublicationId} onRemoved={removePublication} onUpdated={updatePublication} />)}
      </div>
    </Page>
  );
}

function PublicationManagerCard({ publication, initiallyOpen = false, highlighted = false, onRemoved, onUpdated }) {
  const [open, setOpen] = useState(initiallyOpen);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState("");
  const [message, setMessage] = useState(null);
  const publicationType = publication.type || publication.category || "Publicação";
  const publicationVisibility = publication.visibility === "both" ? "Comunidade e perfil" : publication.visibility === "profile" ? "Perfil" : "Comunidade";
  const publicationStatus = publication.status === "published" ? "Publicado" : publication.status === "archived" ? "Arquivado" : publication.status === "draft" ? "Rascunho" : publication.status || "Publicado";
  const publishedAt = publication.publishedAt ? new Date(publication.publishedAt).toLocaleDateString("pt-BR") : "Ainda não publicado";
  const publicationLikes = Array.isArray(publication.likes) ? publication.likes : [];
  const publicationComments = Array.isArray(publication.comments) ? publication.comments : [];
  const [editForm, setEditForm] = useState({
    title: publication.title || "",
    categorySlug: publication.categorySlug || "noticias",
    visibility: publication.visibility || "both",
    content: publication.text || "",
    mainImageDataUrl: "",
    mainImageFileName: "",
    mainImageMimeType: ""
  });
  const [editImagePreview, setEditImagePreview] = useState("");

  useEffect(() => {
    setEditForm({
      title: publication.title || "",
      categorySlug: publication.categorySlug || "noticias",
      visibility: publication.visibility || "both",
      content: publication.text || "",
      mainImageDataUrl: "",
      mainImageFileName: "",
      mainImageMimeType: ""
    });
    setEditImagePreview("");
  }, [publication.id, publication.title, publication.categorySlug, publication.visibility, publication.text]);

  const updateEditField = (field, value) => setEditForm((current) => ({ ...current, [field]: value }));

  const handleEditImageChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setEditImagePreview("");
      setEditForm((current) => ({ ...current, mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" }));
      return;
    }
    if (!["image/png", "image/jpeg", "image/gif", "image/webp"].includes(file.type)) {
      setMessage({ type: "error", text: "Use uma imagem PNG, JPG, GIF ou WebP." });
      event.target.value = "";
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A imagem deve ter no maximo 5MB." });
      event.target.value = "";
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setEditImagePreview(dataUrl);
      setEditForm((current) => ({ ...current, mainImageDataUrl: dataUrl, mainImageFileName: file.name, mainImageMimeType: file.type }));
    };
    reader.readAsDataURL(file);
  };

  const saveEdit = async () => {
    setMessage(null);
    if (!editForm.content.trim()) {
      setMessage({ type: "error", text: "O texto da publicação não pode ficar vazio." });
      return;
    }
    setSaving("edit");
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(publication.id)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(editForm)
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível editar a publicação.");
      onUpdated?.({ ...publication, ...result, title: editForm.title, categorySlug: editForm.categorySlug, visibility: editForm.visibility, text: editForm.content, imageUrl: result.imageUrl || publication.imageUrl });
      setEditing(false);
      setMessage({ type: "success", text: "Publicação atualizada." });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível editar agora." });
    } finally {
      setSaving("");
    }
  };

  const archivePost = async () => {
    if (!window.confirm("Arquivar esta publicação? Ela deixará de aparecer na comunidade.")) return;
    setSaving("archive");
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(publication.id)}/archive`, { method: "PATCH", credentials: "include" });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível arquivar.");
      onUpdated?.({ ...publication, status: "archived" });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível arquivar agora." });
    } finally {
      setSaving("");
    }
  };

  const publishPost = async () => {
    setSaving("publish");
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(publication.id)}/publish`, { method: "PATCH", credentials: "include" });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível publicar.");
      onUpdated?.({ ...publication, ...result, status: "published" });
      setMessage({ type: "success", text: "Publicação publicada na comunidade." });
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível publicar agora." });
    } finally {
      setSaving("");
    }
  };

  const deletePost = async () => {
    if (!window.confirm("Excluir esta publicação? Esta ação remove a publicação da sua listagem.")) return;
    setSaving("delete");
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/community/posts/${encodeURIComponent(publication.id)}`, { method: "DELETE", credentials: "include" });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível excluir.");
      onRemoved?.(publication.id);
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível excluir agora." });
    } finally {
      setSaving("");
    }
  };

  useEffect(() => {
    if (initiallyOpen) {
      setOpen(true);
      window.setTimeout(() => {
        document.getElementById(publication.id)?.scrollIntoView({ behavior: "smooth", block: "center" });
      }, 80);
    }
  }, [initiallyOpen, publication.id]);
  return (
    <Card className={`publicationCard ${highlighted ? "publicationCardHighlighted" : ""}`} id={publication.id}>
      <div className="publicationSummary">
        <div className="publicationThumb">
          {publication.imageUrl ? <img src={publication.imageUrl} alt={publication.title || publicationType} loading="lazy" decoding="async" /> : <span>{publication.imageLabel || publicationType}</span>}
        </div>
        <div>
          <strong>{publication.title || publicationType}</strong>
          <span>{publicationType} | {publicationVisibility}</span>
        </div>
        <span className={`statusPill ${publicationStatus === "Publicado" ? "open" : "review"}`}>{publicationStatus}</span>
        <button className="iconButton secondaryIcon" title={open ? "Recolher detalhes" : "Ver detalhes"} aria-label={open ? "Recolher detalhes" : "Ver detalhes"} onClick={() => setOpen((current) => !current)}>{open ? "\u2212" : "+"}</button>
      </div>

      {open && (
        <div className="publicationDetails">
          {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
          <div className="publicationMediaPreview">
            {publication.imageUrl ? <img src={publication.imageUrl} alt={publication.title || publicationType} loading="lazy" decoding="async" /> : <><span>{publication.imageLabel || publicationType}</span><small>Imagem exibida na comunidade e no perfil público</small></>}
          </div>
          {editing ? (
            <div className="publicationEditForm">
              <FormGrid>
                <Field label="Título" hint={`${editForm.title.length}/100 caracteres`}><input value={editForm.title} onChange={(event) => updateEditField("title", event.target.value)} maxLength="100" /></Field>
                <Field label="Tipo"><select value={editForm.categorySlug} onChange={(event) => updateEditField("categorySlug", event.target.value)}>{communityCategoryOptions.map((option) => <option value={option.slug} key={option.slug}>{option.label}</option>)}</select></Field>
                <Field label="Visibilidade"><select value={editForm.visibility} onChange={(event) => updateEditField("visibility", event.target.value)}><option value="both">Comunidade e perfil</option><option value="community">Apenas comunidade</option><option value="profile">Apenas perfil público</option></select></Field>
              </FormGrid>
              <Field label="Trocar imagem">
                <div className="imageUploadField">
                  <div className="imageMiniPreview">{editImagePreview ? <img src={editImagePreview} alt="Prévia da nova imagem" /> : publication.imageUrl ? <img src={publication.imageUrl} alt="Imagem atual da publicação" /> : <span>IMG</span>}</div>
                  <div>
                    <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={handleEditImageChange} />
                    <small>{editForm.mainImageFileName || "Manter imagem atual"}</small>
                  </div>
                </div>
              </Field>
              <Field label="Texto"><textarea value={editForm.content} onChange={(event) => updateEditField("content", event.target.value)} /></Field>
              <div className="actions">
                <Button onClick={saveEdit} disabled={saving === "edit"}>{saving === "edit" ? "Salvando..." : "Salvar alterações"}</Button>
                <Button variant="secondary" onClick={() => setEditing(false)} disabled={Boolean(saving)}>Cancelar</Button>
              </div>
            </div>
          ) : (
            <p>{publication.text}</p>
          )}
          <div className="publicationMetrics">
            <div><strong>{publication.likeCount ?? publicationLikes.length}</strong><span>curtidas</span></div>
            <div><strong>{publication.commentCount ?? publicationComments.length}</strong><span>comentários</span></div>
            <div><strong>{publishedAt}</strong><span>publicação</span></div>
          </div>
          <div className="likedBy">{publicationLikes.length === 0 ? <span>Nenhuma curtida ainda</span> : publicationLikes.map((name) => <span key={name}>{name}</span>)}</div>
          <div className="commentPanel">
            <div className="commentPanelHeader"><strong>Comentários recebidos</strong><button type="button" className="commentCloseButton" title="Recolher detalhes" aria-label="Recolher detalhes" onClick={() => setOpen(false)}>{"\u00D7"}</button></div>
            <div className="commentList">
              {publicationComments.length === 0 && <p className="emptyComments">Nenhum comentário ainda.</p>}
              {publicationComments.map((comment) => <div className="commentItem" key={comment.id || `${comment.company}-${comment.text}`}><strong>{comment.userName ? `${comment.userName}${comment.company ? ` · ${comment.company}` : ""}` : comment.company}</strong><p>{comment.text}</p></div>)}
            </div>
          </div>
          <div className="postActions compactPostActions">
            <button className="iconButton secondaryIcon" title="Editar publicação" aria-label="Editar publicação" disabled={Boolean(saving)} onClick={() => setEditing((current) => !current)}>{"\u270E"}</button>
            {publication.status === "published" ? <button className="iconButton warningIcon" title="Arquivar publicação" aria-label="Arquivar publicação" disabled={Boolean(saving)} onClick={archivePost}>!</button> : <button className="iconButton successIcon" title={publication.status === "draft" ? "Publicar na comunidade" : "Republicar na comunidade"} aria-label={publication.status === "draft" ? "Publicar na comunidade" : "Republicar na comunidade"} disabled={Boolean(saving)} onClick={publishPost}>{"\u21BB"}</button>}
            <button className="iconButton dangerIcon" title="Excluir publicação" aria-label="Excluir publicação" disabled={Boolean(saving)} onClick={deletePost}>{"\u00D7"}</button>
          </div>
        </div>
      )}
    </Card>
  );
}

function RadarHomeConnected({ navigate, openNewsDetail }) {
  const [news, setNews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const pageSize = 12;

  useEffect(() => {
    let active = true;

    fetch(`${API_BASE_URL}/api/news`)
      .then((response) => {
        if (!response.ok) throw new Error("Falha ao carregar noticias");
        return response.text();
      })
      .then((text) => {
        const jsonStart = text.search(/[\\[{]/);
        return JSON.parse(jsonStart >= 0 ? text.slice(jsonStart) : text);
      })
      .then((data) => {
        if (!active) return;
        setNews(Array.isArray(data) ? data : []);
        setError("");
      })
      .catch(() => {
        if (!active) return;
        setError("Não foi possível carregar as notícias do banco. Confirme se o backend está ligado.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  const featured = news.find((item) => item.status === "featured" || item.status === "published") || news[0];
  const remainingNews = featured ? news.filter((item) => item.id !== featured.id) : news;
  const filteredNews = remainingNews.filter((item) => {
    const matchesCategory = categoryFilter === "all" || item.categorySlug === categoryFilter || item.categoryName === categoryFilter;
    const term = search.trim().toLowerCase();
    const matchesSearch = !term || [item.title, item.summary, item.categoryName].some((value) => String(value || "").toLowerCase().includes(term));
    return matchesCategory && matchesSearch;
  });
  const totalPages = Math.max(1, Math.ceil(filteredNews.length / pageSize));
  const currentPage = Math.min(page, totalPages);
  const paginatedNews = filteredNews.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const changeCategory = (value) => {
    setCategoryFilter(value);
    setPage(1);
  };
  const updateSearch = (value) => {
    setSearch(value);
    setPage(1);
  };

  return (
    <Page label="Radar LicitaHub" title="Notícias e inteligência de mercado">
      {loading && <Card className="formFeedback"><p>Carregando notícias do banco...</p></Card>}
      {error && <Card className="formFeedback dangerNotice"><p>{error}</p></Card>}
      {featured && (
        <div className="newsHero">
          <div className="newsHeroImage">{featured.mainImageUrl ? <img src={featured.mainImageUrl} alt={featured.title} loading="lazy" decoding="async" /> : "Radar"}</div>
          <div className="newsHeroText">
            <span className="badge">{featured.categoryName || "Radar"}</span>
            <h3>{featured.title}</h3>
            <p>{featured.summary || "Notícia publicada pela LicitaHub para orientar a comunidade."}</p>
            <Button onClick={() => openNewsDetail(featured)}>Ler notícia</Button>
          </div>
        </div>
      )}
      <div className="radarFilterBar">
        <Field label="Filtrar notícias"><input value={search} onChange={(event) => updateSearch(event.target.value)} placeholder="Buscar por título, resumo ou categoria" /></Field>
        <div className="categoryTabs">
          <button className={categoryFilter === "all" ? "active" : ""} onClick={() => changeCategory("all")}>Todos</button>
          {newsCategoryOptions.map((item) => <button className={categoryFilter === item.slug ? "active" : ""} onClick={() => changeCategory(item.slug)} key={item.slug}>{item.label}</button>)}
        </div>
      </div>
      <div className="newsGrid">
        {!loading && !error && news.length === 0 && <Card className="formFeedback"><p>Nenhuma notícia cadastrada ainda.</p></Card>}
        {!loading && !error && news.length > 0 && filteredNews.length === 0 && <Card className="formFeedback"><p>Nenhuma notícia encontrada com esses filtros.</p></Card>}
        {paginatedNews.map((item) => <NewsCardConnected key={item.id} news={item} openNewsDetail={openNewsDetail} />)}
      </div>
      {filteredNews.length > pageSize && (
        <div className="paginationBar">
          <button disabled={currentPage === 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>Anterior</button>
          <span>Página {currentPage} de {totalPages}</span>
          <button disabled={currentPage === totalPages} onClick={() => setPage((value) => Math.min(totalPages, value + 1))}>Próxima</button>
        </div>
      )}
    </Page>
  );
}

function NewsCardConnected({ news, openNewsDetail }) {
  return (
    <Card className="newsCard">
      <div className="newsThumb">{news.mainImageUrl ? <img src={news.mainImageUrl} alt={news.title} loading="lazy" decoding="async" /> : news.categoryName}</div>
      <span className="badge">{news.categoryName || "Radar"}</span>
      <h3>{news.title}</h3>
      <p>{news.summary || "Notícia publicada pela LicitaHub para orientar empresas associadas e usuários vinculados."}</p>
      <button className="textLink" onClick={() => openNewsDetail(news)}>Ler notícia</button>
    </Card>
  );
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

function RadarDetailConnected({ selectedNews, navigate }) {
  if (!selectedNews) {
    return (
      <Page label="Radar LicitaHub" title="Detalhe da notícia">
        <Card className="formFeedback">
          <p>Selecione uma notícia na tela Radar para abrir o detalhe.</p>
          <button className="textLink" onClick={() => navigate("radar-home")}>Voltar para notícias</button>
        </Card>
      </Page>
    );
  }

  const publishedAt = selectedNews.publishedAt || selectedNews.createdAt;
  const formattedDate = publishedAt ? new Date(publishedAt).toLocaleDateString("pt-BR") : "";
  const paragraphs = String(selectedNews.content || selectedNews.summary || "").split(/\n+/).filter(Boolean);

  return (
    <Page label="Radar LicitaHub" title="Detalhe da notícia">
      <article className="article articleWide newsDetailArticle">
        <div className="articleImage articleCover newsDetailImage">{selectedNews.mainImageUrl ? <img src={selectedNews.mainImageUrl} alt={selectedNews.title} loading="lazy" decoding="async" /> : "LicitaHub Radar"}</div>
        <div className="articleLayout">
          <div className="articleBody">
            <span className="badge">{selectedNews.categoryName || "Radar"}</span>
            <h3>{selectedNews.title}</h3>
            {formattedDate && <p className="articleMeta">Publicado pela LicitaHub em {formattedDate}</p>}
            {selectedNews.summary && <p className="articleLead">{selectedNews.summary}</p>}
            {paragraphs.map((paragraph, index) => <p key={index}>{paragraph}</p>)}
          </div>
          <aside className="articleAside">
            <Card>
              <h3>Resumo</h3>
              <p>{selectedNews.summary || "Notícia publicada no Radar LicitaHub."}</p>
            </Card>
            <Card>
              <h3>Ação</h3>
              <button className="textLink" onClick={() => navigate("radar-home")}>Voltar para notícias</button>
            </Card>
          </aside>
        </div>
      </article>
    </Page>
  );
}

function RadarDetail() {
  return (
    <Page label="Radar LicitaHub" title="Detalhe da notícia">
      <article className="article articleWide newsDetailArticle">
        <div className="articleImage articleCover newsDetailImage">LicitaHub Radar</div>
        <div className="articleLayout">
          <div className="articleBody">
            <span className="badge">Licitações</span>
            <h3>Nova rodada de oportunidades em infraestrutura deve movimentar consórcios técnicos</h3>
            <p className="articleMeta">Publicado pela LicitaHub em 04/07/2026</p>
            <p>O mercado de engenharia consultiva segue observando editais com escopos mais amplos e exigências multidisciplinares.</p>
            <p>Empresas que antes avaliavam oportunidades de forma isolada passam a buscar composição técnica mais cedo.</p>
            <p>A LicitaHub recomenda que empresas mantenham seus perfis atualizados e registrem com clareza o que podem oferecer.</p>
          </div>
          <aside className="articleAside">
            <Card>
              <h3>Resumo</h3>
              <p>Editais com maior complexidade técnica aumentam a necessidade de parcerias bem estruturadas.</p>
            </Card>
            <Card>
              <h3>Relacionadas</h3>
              <button className="textLink">Critérios técnicos em propostas públicas</button>
              <button className="textLink">Como preparar anúncios de parceria</button>
            </Card>
          </aside>
        </div>
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
      <ImageUploadField label="Imagem" />
      <Field label="Resumo"><textarea placeholder="Resumo que aparece nos cards" /></Field>
      <Field label="Texto completo"><textarea placeholder="Texto completo da notícia" /></Field>
    </Page>
  );
}

function RadarNewConnected({ navigate }) {
  const [form, setForm] = useState({
    title: "",
    categorySlug: "licitacoes",
    status: "draft",
    expiresAt: "",
    summary: "",
    content: "",
    mainImageUrl: "",
    mainImageDataUrl: "",
    mainImageFileName: "",
    mainImageMimeType: ""
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [imagePreview, setImagePreview] = useState("");
  const todayISO = new Date().toISOString().slice(0, 10);

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleImageChange = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      setImagePreview("");
      setForm((current) => ({ ...current, mainImageDataUrl: "", mainImageFileName: "", mainImageMimeType: "" }));
      return;
    }

    const allowedTypes = ["image/png", "image/jpeg", "image/gif", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      setMessage({ type: "error", text: "Use uma imagem PNG, JPG, GIF ou WebP." });
      event.target.value = "";
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      setMessage({ type: "error", text: "A imagem deve ter no maximo 5MB." });
      event.target.value = "";
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = String(reader.result || "");
      setImagePreview(dataUrl);
      setForm((current) => ({
        ...current,
        mainImageDataUrl: dataUrl,
        mainImageFileName: file.name,
        mainImageMimeType: file.type
      }));
    };
    reader.readAsDataURL(file);
  };

  const handleSubmit = async () => {
    setMessage(null);

    if (!form.title.trim()) {
      setMessage({ type: "error", text: "Informe o titulo da noticia antes de publicar." });
      return;
    }

    if ((form.status === "published" || form.status === "featured") && !form.expiresAt) {
      setMessage({ type: "error", text: "Informe ate quando a noticia ficara publicada." });
      return;
    }

    if ((form.status === "published" || form.status === "featured") && form.expiresAt < todayISO) {
      setMessage({ type: "error", text: "A data final da publicação não pode ser anterior a hoje." });
      return;
    }

    setSaving(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/news`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form)
      });

      const result = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(result.error || "Não foi possível cadastrar a notícia.");
      }

      setMessage({ type: "success", text: "Noticia cadastrada no banco com sucesso." });
      navigate("radar-home");
      setForm({
        title: "",
        categorySlug: "licitacoes",
        status: "draft",
        expiresAt: "",
        summary: "",
        content: "",
        mainImageUrl: "",
        mainImageDataUrl: "",
        mainImageFileName: "",
        mainImageMimeType: ""
      });
      setImagePreview("");
    } catch (error) {
      setMessage({ type: "error", text: error.message || "Não foi possível gravar agora. Confirme se o backend está ligado." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Page label="Radar LicitaHub" title="Cadastrar notícia">
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      <FormGrid>
        <Field label="Título" hint={`${form.title.length}/90 caracteres`}><input value={form.title} onChange={(event) => updateField("title", event.target.value)} placeholder="Título da notícia" maxLength="90" /></Field>
        <Field label="Categoria"><select value={form.categorySlug} onChange={(event) => updateField("categorySlug", event.target.value)}>{newsCategoryOptions.map((option) => <option value={option.slug} key={option.slug}>{option.label}</option>)}</select></Field>
        <Field label="Status"><select value={form.status} onChange={(event) => updateField("status", event.target.value)}>{newsStatusOptions.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}</select></Field>
        <Field label="Publicar até" hint="Depois desta data, a notícia deixa de aparecer como publicada no Radar."><input type="date" min={todayISO} value={form.expiresAt} onChange={(event) => updateField("expiresAt", event.target.value)} disabled={form.status === "draft" || form.status === "archived"} /></Field>
      </FormGrid>
      <Field label="Imagem">
        <div className="imageUploadField">
          <div className="imageMiniPreview">{imagePreview ? <img src={imagePreview} alt="Previa da imagem selecionada" /> : <span>IMG</span>}</div>
          <div>
            <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={handleImageChange} />
            <small>{form.mainImageFileName || "Nenhuma imagem selecionada"}</small>
          </div>
        </div>
      </Field>
      <Field label="Link da imagem" hint="Opcional. Use apenas se a imagem ja estiver publicada na internet."><input value={form.mainImageUrl} onChange={(event) => updateField("mainImageUrl", event.target.value)} placeholder="https://..." /></Field>
      <Field label="Resumo" hint={`${form.summary.length}/240 caracteres`}><textarea value={form.summary} onChange={(event) => updateField("summary", event.target.value)} placeholder="Resumo que aparece nos cards" maxLength="240" /></Field>
      <Field label="Texto completo"><textarea value={form.content} onChange={(event) => updateField("content", event.target.value)} placeholder="Texto completo da notícia" /></Field>
      <div className="formActionBar">
        <Button onClick={handleSubmit} disabled={saving}>{saving ? "Publicando..." : "Publicar notícia"}</Button>
      </div>
    </Page>
  );
}

function RadarManage() {
  const [news, setNews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [editing, setEditing] = useState({});
  const [savingId, setSavingId] = useState("");
  const [message, setMessage] = useState(null);
  const todayISO = new Date().toISOString().slice(0, 10);

  const loadNews = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/news/admin`);
      const result = await response.json().catch(() => []);
      if (!response.ok) throw new Error(result.error || "Não foi possível carregar as notícias.");
      const items = Array.isArray(result) ? result : [];
      setNews(items);
      setEditing(Object.fromEntries(items.map((item) => [
        item.id,
        {
          status: item.status,
          expiresAt: item.expiresAt ? String(item.expiresAt).slice(0, 10) : ""
        }
      ])));
      setMessage(null);
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadNews();
  }, []);

  const filteredNews = useMemo(() => {
    const term = search.trim().toLocaleLowerCase("pt-BR");
    return news.filter((item) => {
      const matchesStatus = filter === "all" || item.status === filter;
      const matchesSearch = !term || item.title.toLocaleLowerCase("pt-BR").includes(term);
      return matchesStatus && matchesSearch;
    });
  }, [news, filter, search]);

  const updateEditing = (id, field, value) => {
    setEditing((current) => ({
      ...current,
      [id]: { ...current[id], [field]: value }
    }));
  };

  const saveStatus = async (item) => {
    const values = editing[item.id] || {};
    if ((values.status === "published" || values.status === "featured") && !values.expiresAt) {
      setMessage({ type: "error", text: "Informe a data final para disponibilizar ou destacar a notícia." });
      return;
    }

    setSavingId(item.id);
    setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/news/${item.id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(values)
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Não foi possível alterar o status.");
      setMessage({ type: "success", text: `Status de “${item.title}” atualizado.` });
      await loadNews();
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSavingId("");
    }
  };

  const formatDate = (value) => value ? new Date(value).toLocaleDateString("pt-BR") : "Não definida";
  const statusLabel = (value) => newsStatusOptions.find((option) => option.value === value)?.label || value;

  return (
    <Page label="Radar LicitaHub" title="Gerenciar notícias">
      <Card className="newsManageFilters newsManageFiltersSticky">
        <Field label="Buscar notícia">
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Digite parte do título" />
        </Field>
        <Field label="Filtrar por status">
          <select value={filter} onChange={(event) => setFilter(event.target.value)}>
            <option value="all">Todos os status</option>
            {newsStatusOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
          </select>
        </Field>
      </Card>

      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      {loading ? <Card><p>Carregando notícias...</p></Card> : (
        <div className="tableWrap">
          <table className="newsManageTable">
            <thead>
              <tr><th>Notícia</th><th>Categoria</th><th>Status atual</th><th>Publicação</th><th>Novo status</th><th>Publicar até</th><th>Ação</th></tr>
            </thead>
            <tbody>
              {filteredNews.map((item) => {
                const values = editing[item.id] || {};
                const needsExpiration = values.status === "published" || values.status === "featured";
                return (
                  <tr key={item.id}>
                    <td><strong>{item.title}</strong></td>
                    <td>{item.categoryName || "Sem categoria"}</td>
                    <td><span className={`statusBadge status-${item.status}`}>{statusLabel(item.status)}</span></td>
                    <td>{formatDate(item.publishedAt || item.createdAt)}</td>
                    <td>
                      <select value={values.status || item.status} onChange={(event) => updateEditing(item.id, "status", event.target.value)}>
                        {newsStatusOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                      </select>
                    </td>
                    <td>
                      <input type="date" min={todayISO} value={values.expiresAt || ""} disabled={!needsExpiration} onChange={(event) => updateEditing(item.id, "expiresAt", event.target.value)} />
                    </td>
                    <td><Button onClick={() => saveStatus(item)} disabled={savingId === item.id}>{savingId === item.id ? "Salvando..." : "Salvar"}</Button></td>
                  </tr>
                );
              })}
              {filteredNews.length === 0 && <tr><td colSpan="7">Nenhuma notícia encontrada com esses filtros.</td></tr>}
            </tbody>
          </table>
        </div>
      )}
    </Page>
  );
}

function TenderAdmin({ navigate }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", status: "", state: "", modality: "", interest: "" });
  const loadTenders = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/tenders`, { credentials: "include" })
      .then(async (r) => { const d=await r.json(); if(!r.ok) throw new Error(d.error); return d; })
      .then(setItems)
      .catch((e)=>setError(e.message))
      .finally(()=>setLoading(false));
  };
  useEffect(() => { loadTenders(); }, []);
  const deleteTender = async (item) => {
    if (!window.confirm(`Excluir o edital ${item.number}?`)) return;
    try {
      const response = await fetch(`${API_BASE_URL}/api/tenders/${item.id}`, { method: "DELETE", credentials: "include" });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível excluir o edital.");
      loadTenders();
    } catch (err) {
      setError(err.message);
    }
  };
  const count = (status) => items.filter((item) => item.status === status).length;
  const date = (value) => value ? new Date(value).toLocaleDateString("pt-BR") : "-";
  const compactObject = (value) => {
    const fullText = String(value || "").trim() || "Objeto não informado";
    const visibleText = fullText.length > 140 ? `${fullText.slice(0, 137).trimEnd()}...` : fullText;
    return <span className="tenderObjectCell" title={fullText}>{visibleText}</span>;
  };
  const filteredItems = items.filter((item) => {
    const term = filters.search.trim().toLowerCase();
    const matchesSearch = !term || [item.agency, item.number, item.object, item.city, item.judgmentCriterion].some((value) => String(value || "").toLowerCase().includes(term));
    const matchesStatus = !filters.status || item.status === filters.status;
    const matchesState = !filters.state || item.state === filters.state;
    const matchesModality = !filters.modality || item.modality === filters.modality;
    return matchesSearch && matchesStatus && matchesState && matchesModality;
  });
  return <Page label="Editais" title="Painel administrativo de editais" actions={<Button onClick={() => navigate("tender-new")}>Cadastrar edital</Button>}><Stats items={[["Publicados", String(count("published"))], ["Ocorridos", String(count("occurred"))], ["Rascunhos", String(count("draft"))], ["Suspensos", String(count("suspended"))]]} /><Card className="compactFilters tenderAdminFiltersSticky"><FormGrid><Field label="Buscar edital"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Órgão, número, objeto, cidade ou critério" /></Field><Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option><option value="draft">Rascunho</option><option value="published">Publicado</option><option value="under_review">Em análise</option><option value="suspended">Suspenso</option><option value="challenged">Impugnado</option><option value="occurred">Ocorrido</option><option value="closed">Encerrado</option><option value="cancelled">Cancelado</option></select></Field><Field label="Estado"><StateSelect value={filters.state} onChange={(event) => setFilters((current) => ({ ...current, state: event.target.value }))} /></Field><Field label="Modalidade"><select value={filters.modality} onChange={(event) => setFilters((current) => ({ ...current, modality: event.target.value }))}><option value="">Todas</option>{tenderModalityOptions.map((option) => <option value={option} key={option}>{option}</option>)}</select></Field></FormGrid></Card>{loading&&<Card><p>Carregando editais...</p></Card>}{error&&<Card className="dangerNotice"><p>{error}</p></Card>}{!loading&&!error&&filteredItems.length===0&&<Card><p>Nenhum edital encontrado com esses filtros.</p></Card>}{!loading&&!error&&filteredItems.length>0&&<Table columns={["Órgão","Número","Objeto","Abertura","Critério","Status","Ação"]} rows={filteredItems.map((item)=>[item.agency,item.number,item.object,date(item.openingDate),item.judgmentCriterion||"-",item.status,<div className="rowActions compactRowActions" key={item.id}><button className="iconButton secondaryIcon" title="Ver detalhe" onClick={()=>navigate(`tender-detail?id=${item.id}`)}>{"\u25C9"}</button><button className="iconButton partnerIcon" title="Editar edital" onClick={()=>navigate(`tender-new?id=${item.id}`)}>{"\u270E"}</button><button className="iconButton dangerIcon" title="Excluir edital" onClick={()=>deleteTender(item)}>{"\u00D7"}</button></div>])}/>}</Page>;
}

function TechnicalProfessionals({ sessionUser }) {
  const blankEducation = () => ({ level: "graduation", courseName: "", institution: "" });
  const educationLabels = { graduation: "Graduação", specialization: "Especialização / Pós-graduação", mba: "MBA", masters: "Mestrado", doctorate: "Doutorado", postdoctorate: "Pós-doutorado", extension: "Curso de extensão", certification: "Certificação", other: "Outra" };
  const emptyForm = { fullName: "", professionalRegistration: "", roleTitle: "", phone: "", email: "", state: "", availabilityStatus: "available", profileSummary: "", status: "active", educations: [blankEducation()] };
  const [items, setItems] = useState([]); const [search, setSearch] = useState(""); const [statusFilter, setStatusFilter] = useState("active"); const [form, setForm] = useState(emptyForm); const [editingId, setEditingId] = useState(""); const [showForm, setShowForm] = useState(false); const [saving, setSaving] = useState(false); const [message, setMessage] = useState(null);
  const canManage = sessionUser?.roleKey !== "reader";
  const load = async () => { try { const response = await fetch(`${API_BASE_URL}/api/technical-professionals?status=${encodeURIComponent(statusFilter)}&search=${encodeURIComponent(search)}`, { credentials: "include" }); const data = await response.json(); if (!response.ok) throw new Error(data.error || "Não foi possível carregar profissionais."); setItems(Array.isArray(data) ? data : []); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  useEffect(() => { load(); }, [statusFilter]);
  const update = (field, value) => setForm((current) => ({ ...current, [field]: value }));
  const close = () => { setShowForm(false); setEditingId(""); setForm(emptyForm); };
  const edit = (item) => { setEditingId(item.id); setForm({ ...emptyForm, ...item, educations: Array.isArray(item.educations) && item.educations.length ? item.educations.map((education) => ({ level: education.level || "graduation", courseName: education.courseName || "", institution: education.institution || "" })) : [blankEducation()] }); setShowForm(true); window.scrollTo({ top: 0, behavior: "smooth" }); };
  const updateEducation = (index, field, value) => setForm((current) => ({ ...current, educations: current.educations.map((education, currentIndex) => currentIndex === index ? { ...education, [field]: value } : education) }));
  const addEducation = () => setForm((current) => ({ ...current, educations: [...current.educations, blankEducation()] }));
  const removeEducation = (index) => setForm((current) => ({ ...current, educations: current.educations.filter((_, currentIndex) => currentIndex !== index) }));
  const save = async (event) => { event.preventDefault(); if (!form.fullName.trim()) { setMessage({ type: "error", text: "Informe o nome do profissional." }); return; } setSaving(true); try { const response = await fetch(`${API_BASE_URL}/api/technical-professionals${editingId ? `/${editingId}` : ""}`, { method: editingId ? "PUT" : "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify(form) }); const data = await response.json(); if (!response.ok) throw new Error(data.error || "Não foi possível salvar o profissional."); setMessage({ type: "success", text: editingId ? "Profissional atualizado." : "Profissional cadastrado." }); close(); await load(); } catch (error) { setMessage({ type: "error", text: error.message }); } finally { setSaving(false); } };
  const remove = async (item) => { if (!window.confirm(`Remover ${item.fullName} da base de profissionais?`)) return; try { const response = await fetch(`${API_BASE_URL}/api/technical-professionals/${encodeURIComponent(item.id)}`, { method: "DELETE", credentials: "include" }); const data = await response.json(); if (!response.ok) throw new Error(data.error || "Não foi possível remover o profissional."); await load(); } catch (error) { setMessage({ type: "error", text: error.message }); } };
  return <Page label="Capacidade técnica" title="Profissionais técnicos" actions={canManage ? <Button onClick={() => { setShowForm(true); setEditingId(""); setForm(emptyForm); }}>Cadastrar profissional</Button> : null}><Card className="certificateIntro"><span className="eyebrow">Base técnica da empresa</span><h3>Profissionais para CATs e licitações</h3><p>Cadastre as qualificações, contatos e disponibilidade de cada profissional que pode compor uma proposta.</p></Card>{message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}{showForm && <Card className="certificateForm"><div className="sectionTitle"><h3>{editingId ? "Editar profissional" : "Cadastrar profissional"}</h3><Button variant="secondary" onClick={close}>Fechar</Button></div><form onSubmit={save}><FormGrid><Field label="Nome completo"><input value={form.fullName} onChange={(event) => update("fullName", event.target.value)} required maxLength="255" /></Field><Field label="Registro profissional"><input value={form.professionalRegistration} onChange={(event) => update("professionalRegistration", event.target.value)} placeholder="CREA, CAU, CRBio..." maxLength="180" /></Field><Field label="Cargo e função"><input value={form.roleTitle} onChange={(event) => update("roleTitle", event.target.value)} placeholder="Coordenador, engenheiro, especialista..." maxLength="180" /></Field><Field label="Telefone"><input value={form.phone} onChange={(event) => update("phone", formatPhoneBR(event.target.value))} placeholder="(00) 00000-0000" maxLength="15" /></Field><Field label="E-mail"><input type="email" value={form.email} onChange={(event) => update("email", event.target.value)} placeholder="profissional@empresa.com.br" maxLength="255" /></Field><Field label="UF"><StateSelect value={form.state} onChange={(event) => update("state", event.target.value)} /></Field><Field label="Disponibilidade"><select value={form.availabilityStatus} onChange={(event) => update("availabilityStatus", event.target.value)}><option value="available">Disponível</option><option value="limited">Disponibilidade parcial</option><option value="unavailable">Indisponível</option></select></Field><Field label="Situação"><select value={form.status} onChange={(event) => update("status", event.target.value)}><option value="active">Ativo</option><option value="archived">Arquivado</option></select></Field></FormGrid><section className="professionalEducations"><div className="sectionTitle"><div><span className="eyebrow">Qualificação acadêmica</span><h4>Formações</h4><p>Inclua graduação, pós-graduações e demais qualificações em linhas separadas.</p></div><Button type="button" variant="secondary" onClick={addEducation}>Adicionar formação</Button></div>{form.educations.map((education, index) => <div className="professionalEducationRow" key={index}><Field label="Nível"><select value={education.level} onChange={(event) => updateEducation(index, "level", event.target.value)}>{Object.entries(educationLabels).map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field><Field label="Curso ou formação"><input value={education.courseName} onChange={(event) => updateEducation(index, "courseName", event.target.value)} placeholder="Engenharia Civil, Gestão Ambiental..." maxLength="255" /></Field><Field label="Instituição"><input value={education.institution} onChange={(event) => updateEducation(index, "institution", event.target.value)} placeholder="Universidade ou instituição" maxLength="255" /></Field>{form.educations.length > 1 && <Button type="button" variant="secondary" onClick={() => removeEducation(index)}>Remover</Button>}</div>)}</section><Field label="Resumo profissional"><textarea value={form.profileSummary} onChange={(event) => update("profileSummary", event.target.value)} placeholder="Experiências e informações úteis para tomada de decisão em licitações" maxLength="8000" /></Field><div className="formActions"><Button type="submit" disabled={saving}>{saving ? "Salvando..." : "Salvar profissional"}</Button><Button variant="secondary" onClick={close}>Cancelar</Button></div></form></Card>}<Card className="compactFilters certificateFiltersSticky"><FormGrid><Field label="Pesquisar"><input value={search} onChange={(event) => setSearch(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); load(); } }} placeholder="Nome, formação, registro ou função" /></Field><Field label="Situação"><select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}><option value="active">Ativos</option><option value="archived">Arquivados</option><option value="">Todos</option></select></Field><Field label="Consulta"><Button onClick={load}>Pesquisar</Button></Field></FormGrid></Card><div className="certificateList">{items.length === 0 ? <Card><p>Nenhum profissional encontrado.</p></Card> : items.map((item) => <Card className="certificateCard" key={item.id}><div className="certificateCardHeader"><div><span className="eyebrow">{(item.educations || []).map((education) => educationLabels[education.level] || "Formação").join(" · ") || "Formação não informada"}</span><h3>{item.fullName}</h3><small>{[item.roleTitle, item.professionalRegistration].filter(Boolean).join(" | ") || "Sem função ou registro informado"}</small></div><span className={`statusBadge status-${item.status}`}>{item.status === "archived" ? "Arquivado" : item.availabilityStatus === "available" ? "Disponível" : item.availabilityStatus === "limited" ? "Parcial" : "Indisponível"}</span></div><div className="certificateInfoGrid"><div><span>Contato</span><strong>{[item.phone, item.email].filter(Boolean).join(" | ") || "Não informado"}</strong></div><div><span>UF</span><strong>{item.state || "Não informada"}</strong></div><div><span>Atestados vinculados</span><strong>{item.certificateCount || 0}</strong></div><div><span>Formações</span><strong>{(item.educations || []).map((education) => `${educationLabels[education.level] || "Formação"}: ${education.courseName}`).join(" | ") || "Não informadas"}</strong></div></div>{item.profileSummary && <details className="certificateContent"><summary>Ver resumo profissional</summary><p>{item.profileSummary}</p></details>}{canManage && <div className="certificateActions"><Button variant="secondary" onClick={() => edit(item)}>Editar</Button><Button variant="secondary" onClick={() => remove(item)}>Remover</Button></div>}</Card>)}</div></Page>;
}

function TechnicalCertificates({ sessionUser, navigate }) {
  const blankQuantity = () => ({ description: "", value: "", unit: "", note: "" });
  const emptyForm = { certificateNumber: "", issuerName: "", contractedName: "", object: "", state: "", executionStart: "", executionEnd: "", contractValue: "", catNumber: "", technicalProfessionalId: "", catProfessional: "", professionalRole: "", completionStatus: "final", usageScope: "both", quantities: [blankQuantity()], contentText: "", status: "active", fileDataUrl: "", fileName: "", mimeType: "" };
  const [items, setItems] = useState([]);
  const [professionals, setProfessionals] = useState([]);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("active");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editingId, setEditingId] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [message, setMessage] = useState(null);
  const [processingOCRId, setProcessingOCRId] = useState("");
  const [selectedForAI, setSelectedForAI] = useState([]);
  const canManage = sessionUser?.roleKey !== "reader";
  const load = async (nextSearch = search, nextStatus = statusFilter) => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (nextSearch.trim()) params.set("search", nextSearch.trim());
      if (nextStatus) params.set("status", nextStatus);
      const response = await fetch(`${API_BASE_URL}/api/technical-certificates?${params.toString()}`, { credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível carregar o acervo técnico.");
      setItems(Array.isArray(data) ? data : []);
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setLoading(false); }
  };
  const loadProfessionals = async () => { try { const response = await fetch(`${API_BASE_URL}/api/technical-professionals?status=active`, { credentials: "include" }); const data = await response.json(); if (response.ok) setProfessionals(Array.isArray(data) ? data : []); } catch (_error) {} };
  useEffect(() => { load(); loadProfessionals(); }, []);
  const update = (field, value) => setForm((current) => ({ ...current, [field]: value }));
  const addQuantity = () => update("quantities", [...(form.quantities || []), blankQuantity()]);
  const updateQuantity = (index, field, value) => update("quantities", (form.quantities || []).map((quantity, currentIndex) => currentIndex === index ? { ...quantity, [field]: value } : quantity));
  const removeQuantity = (index) => update("quantities", (form.quantities || []).filter((_, currentIndex) => currentIndex !== index));
  const selectProfessional = (id) => { const professional = professionals.find((item) => item.id === id); setForm((current) => ({ ...current, technicalProfessionalId: id, catProfessional: professional?.fullName || "", professionalRole: professional?.roleTitle || current.professionalRole })); };
  const closeForm = () => { setEditingId(""); setForm(emptyForm); setShowForm(false); };
  const selectFile = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) { setMessage({ type: "error", text: "Envie o atestado em PDF." }); return; }
    if (file.size > 25 * 1024 * 1024) { setMessage({ type: "error", text: "Cada atestado pode ter no máximo 25 MB." }); return; }
    const reader = new FileReader();
    reader.onload = () => update("fileDataUrl", String(reader.result || ""));
    reader.readAsDataURL(file);
    setForm((current) => ({ ...current, fileName: file.name, mimeType: file.type || "application/pdf" }));
  };
  const edit = (item) => {
    setEditingId(item.id);
    setForm({ ...emptyForm, ...item, contractValue: item.contractValue ?? "", executionStart: item.executionStart ? String(item.executionStart).slice(0, 10) : "", executionEnd: item.executionEnd ? String(item.executionEnd).slice(0, 10) : "", quantities: Array.isArray(item.quantities) && item.quantities.length ? item.quantities.map((quantity) => ({ description: quantity.description || "", value: quantity.value ?? "", unit: quantity.unit || "", note: quantity.note || "" })) : [blankQuantity()] });
    setShowForm(true);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };
  const save = async (event) => {
    event.preventDefault();
    if (!form.object.trim()) { setMessage({ type: "error", text: "Informe o objeto ou a experiência registrada no atestado." }); return; }
    if (!editingId && !form.fileDataUrl) { setMessage({ type: "error", text: "Anexe o PDF do atestado." }); return; }
    setSaving(true); setMessage(editingId ? { type: "info", text: "Salvando as alterações do atestado..." } : { type: "info", text: "Arquivo salvo. A leitura OCR está sendo realizada e pode levar alguns instantes." });
    try {
      const response = await fetch(`${API_BASE_URL}/api/technical-certificates${editingId ? `/${editingId}` : ""}`, { method: editingId ? "PUT" : "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify(form) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível salvar o atestado.");
      setMessage({ type: "success", text: editingId ? "Atestado atualizado." : data.extractionStatus === "ocr_extracted" ? "Atestado gravado e leitura OCR concluída." : data.extractionStatus === "extracted" ? "Atestado gravado e texto do PDF capturado." : "Atestado gravado. A leitura não encontrou texto automaticamente; use Refazer leitura OCR ou registre o texto manualmente." });
      closeForm();
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setSaving(false); }
  };
  const remove = async (item) => {
    if (!window.confirm(`Remover o atestado ${item.certificateNumber || item.fileName} do acervo?`)) return;
    try {
      const response = await fetch(`${API_BASE_URL}/api/technical-certificates/${encodeURIComponent(item.id)}`, { method: "DELETE", credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível remover o atestado.");
      setMessage({ type: "success", text: "Atestado removido do acervo." });
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
  };
  const reprocessOCR = async (item) => {
    setProcessingOCRId(item.id); setSaving(true); setMessage({ type: "info", text: "Refazendo a leitura OCR do PDF. Aguarde a confirmação de conclusão." });
    try {
      const response = await fetch(`${API_BASE_URL}/api/technical-certificates/${encodeURIComponent(item.id)}/reprocess-ocr`, { method: "POST", credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível refazer a leitura do atestado.");
      setMessage({ type: "success", text: data.extractionStatus === "ocr_extracted" ? "Leitura OCR concluída." : data.extractionStatus === "extracted" ? "Leitura de texto concluída." : "Não foi possível identificar texto automaticamente neste documento." });
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setSaving(false); setProcessingOCRId(""); }
  };
  const toggleAISelection = (id) => setSelectedForAI((current) => current.includes(id) ? current.filter((item) => item !== id) : [...current, id]);
  const openAIAnalysis = () => {
    window.sessionStorage.setItem("licitahubTechnicalCertificateAISelection", JSON.stringify(selectedForAI));
    navigate("technical-certificate-ai");
  };
  const extractionLabel = (status) => ({ extracted: "Leitura de texto do PDF", ocr_extracted: "Leitura OCR de imagem", manual: "Texto revisado manualmente", pending_ocr: "Aguardando OCR", failed: "Leitura não concluída" }[status] || "Aguardando leitura");
  const quantitiesLabel = (quantities) => (quantities || []).filter((quantity) => quantity.description).map((quantity) => [quantity.value, quantity.unit, quantity.description].filter(Boolean).join(" ")).join("; ");
  return <Page label="Capacidade técnica" title="Atestados técnicos" actions={canManage ? <><Button variant="secondary" onClick={openAIAnalysis} disabled={selectedForAI.length === 0}>Analisar com IA ({selectedForAI.length})</Button><Button onClick={() => { setShowForm(true); setEditingId(""); setForm(emptyForm); }}>Adicionar atestado</Button></> : null}>
    <Card className="certificateIntro"><span className="eyebrow">Consulta privada da empresa</span><h3>Atestados e capacidade técnica</h3><p>Registre o PDF, os dados contratuais, a CAT e todos os quantitativos que comprovam a experiência.</p></Card>
    {message && <Card className={`formFeedback ${message.type === "success" ? "success" : message.type === "info" ? "infoNotice" : "dangerNotice"}`}><p>{message.text}</p></Card>}
    {showForm && <Card className="certificateForm"><div className="sectionTitle"><div><span className="eyebrow">{editingId ? "Atualização" : "Novo documento"}</span><h3>{editingId ? "Editar atestado" : "Cadastrar atestado técnico"}</h3></div><Button variant="secondary" onClick={closeForm}>Fechar</Button></div><form onSubmit={save}><FormGrid>
      <Field label="PDF do atestado" hint={editingId ? "O PDF atual permanece protegido. Para trocar o arquivo, remova e cadastre novamente." : "Somente PDF, até 25 MB."}>{!editingId ? <input type="file" accept="application/pdf,.pdf" onChange={selectFile} required={!editingId} /> : <span className="fileReference">{form.fileName || "Arquivo já cadastrado"}</span>}{form.fileName && !editingId && <small>{form.fileName}</small>}</Field>
      <Field label="Identificação do atestado"><input value={form.certificateNumber} onChange={(event) => update("certificateNumber", event.target.value)} placeholder="Número ou identificação interna" maxLength="160" /></Field>
      <Field label="Contratante"><input value={form.issuerName} onChange={(event) => update("issuerName", event.target.value)} placeholder="Órgão ou empresa que contratou" maxLength="255" /></Field>
      <Field label="Contratado"><input value={form.contractedName} onChange={(event) => update("contractedName", event.target.value)} placeholder="Empresa que executou o contrato" maxLength="255" /></Field>
      <Field label="Objeto"><textarea value={form.object} onChange={(event) => update("object", event.target.value)} placeholder="Objeto e experiência comprovada" maxLength="4000" required /></Field>
      <Field label="Estado"><StateSelect value={form.state} onChange={(event) => update("state", event.target.value)} /></Field>
      <Field label="Início da execução"><input type="date" value={form.executionStart} max={form.executionEnd || undefined} onChange={(event) => update("executionStart", event.target.value)} /></Field>
      <Field label="Fim da execução"><input type="date" value={form.executionEnd} min={form.executionStart || undefined} onChange={(event) => update("executionEnd", event.target.value)} /></Field>
      <Field label="Valor do contrato"><input type="number" min="0" step="0.01" value={form.contractValue} onChange={(event) => update("contractValue", event.target.value)} placeholder="0,00" /></Field>
      <Field label="Número da CAT"><input value={form.catNumber} onChange={(event) => update("catNumber", event.target.value)} placeholder="Número da CAT" maxLength="180" /></Field>
      <Field label="Profissional da CAT" hint="Escolha um profissional técnico já cadastrado."><select value={form.technicalProfessionalId} onChange={(event) => selectProfessional(event.target.value)}><option value="">Não vincular agora</option>{professionals.map((professional) => <option value={professional.id} key={professional.id}>{professional.fullName}{professional.roleTitle ? ` | ${professional.roleTitle}` : ""}</option>)}</select></Field>
      <Field label="Cargo e função"><input value={form.professionalRole} onChange={(event) => update("professionalRole", event.target.value)} placeholder="Cargo ou função exercida" maxLength="180" /></Field>
      <Field label="Situação da execução"><select value={form.completionStatus} onChange={(event) => update("completionStatus", event.target.value)}><option value="final">Atestado final</option><option value="partial">Atestado parcial</option></select></Field>
    </FormGrid><div className="certificateProfessionalHelper"><span>Não encontrou o profissional?</span><Button type="button" variant="secondary" onClick={() => navigate("technical-professionals")}>Cadastrar profissional técnico</Button></div><section className="certificateQuantities"><div className="sectionTitle"><div><span className="eyebrow">Experiência mensurável</span><h4>Quantitativos</h4><p>Inclua quantas linhas forem necessárias: extensão, área, volume, unidades ou qualquer outra medida registrada.</p></div><Button type="button" variant="secondary" onClick={addQuantity}>Adicionar quantitativo</Button></div>{(form.quantities || []).map((quantity, index) => <div className="certificateQuantityRow" key={index}><Field label="Serviço ou quantitativo"><input value={quantity.description} onChange={(event) => updateQuantity(index, "description", event.target.value)} placeholder="Pavimento rígido, drenagem, supervisão..." maxLength="1000" /></Field><Field label="Valor"><input type="number" min="0" step="any" value={quantity.value} onChange={(event) => updateQuantity(index, "value", event.target.value)} placeholder="300" /></Field><Field label="Unidade"><input value={quantity.unit} onChange={(event) => updateQuantity(index, "unit", event.target.value)} placeholder="km, m³, un." maxLength="80" /></Field><Field label="Observação"><input value={quantity.note} onChange={(event) => updateQuantity(index, "note", event.target.value)} placeholder="Detalhe complementar" maxLength="2000" /></Field>{form.quantities.length > 1 && <Button type="button" variant="secondary" onClick={() => removeQuantity(index)}>Remover</Button>}</div>)}</section><FormGrid>
      <Field label="Texto capturado do atestado" hint="Leia, corrija e complemente o texto extraído. Esta versão será guardada para pesquisa e futuras análises por IA."><textarea value={form.contentText} onChange={(event) => update("contentText", event.target.value)} placeholder="Texto completo ou trecho relevante do atestado" /></Field>
      <Field label="Uso permitido" hint="Define como esta comprovação pode ser utilizada nas licitações."><select value={form.usageScope} onChange={(event) => update("usageScope", event.target.value)}><option value="company">Somente como atestado da empresa</option><option value="professional">Somente como CAT / experiência profissional</option><option value="both">Como atestado da empresa e experiência profissional</option></select></Field>
    </FormGrid><div className="formActions"><Button type="submit" disabled={saving}>{saving ? (editingId ? "Salvando..." : "Salvando e lendo OCR...") : editingId ? "Salvar alterações" : "Cadastrar atestado"}</Button><Button variant="secondary" onClick={closeForm} disabled={saving}>Cancelar</Button></div></form></Card>}
    <Card className="compactFilters certificateFiltersSticky"><FormGrid>
      <Field label="Pesquisar no acervo"><input value={search} onChange={(event) => setSearch(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") { event.preventDefault(); load(); } }} placeholder="saneamento, supervisão, CAT, contratante..." /></Field>
      <Field label="Situação"><select value={statusFilter} onChange={(event) => { setStatusFilter(event.target.value); load(search, event.target.value); }}><option value="active">Ativos</option><option value="archived">Arquivados</option><option value="">Todos</option></select></Field>
      <Field label="Consulta"><Button onClick={() => load()}>Pesquisar</Button></Field>
    </FormGrid></Card>
    {!loading && <div className="certificateSummary"><strong>{items.length}</strong> atestado(s) encontrado(s).</div>}
    {loading && <Card><p>Carregando acervo técnico...</p></Card>}
    {!loading && items.length === 0 && <Card><p>Nenhum atestado encontrado com os filtros escolhidos.</p></Card>}
    {!loading && items.length > 0 && <div className="certificateList">{items.map((item) => <Card className={`certificateCard ${selectedForAI.includes(item.id) ? "isSelectedForAI" : ""}`} key={item.id}><div className="certificateCardHeader"><div><span className="eyebrow">{item.issuerName || "Contratante não informado"}</span><h3>{item.certificateNumber || item.fileName}</h3><small>{item.fileName}</small></div><div className="certificateHeaderActions"><label className="certificateSelect"><input type="checkbox" checked={selectedForAI.includes(item.id)} onChange={() => toggleAISelection(item.id)} disabled={!canManage} /><span>Selecionar para IA</span></label><span className={`statusBadge status-${item.status}`}>{item.status === "archived" ? "Arquivado" : "Ativo"}</span></div></div><p className="certificateObject">{item.object}</p><div className="certificateInfoGrid"><div><span>Contratado</span><strong>{item.contractedName || "Não informado"}</strong></div><div><span>CAT</span><strong>{item.catNumber || "Não informada"}</strong></div><div><span>Profissional / função</span><strong>{[item.catProfessional, item.professionalRole].filter(Boolean).join(" - ") || "Não informado"}</strong></div><div><span>Uso permitido</span><strong>{item.usageScope === "company" ? "Somente empresa" : item.usageScope === "professional" ? "Somente profissional" : "Empresa e profissional"}</strong></div><div><span>Execução</span><strong>{item.completionStatus === "partial" ? "Parcial" : "Final"}</strong></div><div><span>Estado</span><strong>{item.state || "Não informado"}</strong></div><div><span>Leitura</span><strong>{extractionLabel(item.extractionStatus)}</strong></div></div>{quantitiesLabel(item.quantities) && <div className="certificateQuantitiesSummary"><span>Quantitativos</span><strong>{quantitiesLabel(item.quantities)}</strong></div>}{item.contentText && <details className="certificateContent"><summary>Ver texto capturado</summary><p>{item.contentText}</p></details>}<div className="certificateActions"><a className="textLink" href={item.documentUrl} target="_blank" rel="noreferrer">Abrir PDF</a>{canManage && <><Button variant="secondary" onClick={() => reprocessOCR(item)} disabled={saving}>{processingOCRId === item.id ? "Lendo OCR..." : "Refazer leitura OCR"}</Button><Button variant="secondary" onClick={() => edit(item)}>Editar dados</Button><Button variant="secondary" onClick={() => edit(item)}>Ler e corrigir texto</Button><Button variant="secondary" onClick={() => remove(item)}>Remover</Button></>}</div></Card>)}</div>}
  </Page>;
}

function formatCertificateDate(value) {
  const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})/);
  return match ? `${match[3]}/${match[2]}/${match[1]}` : "Não informada";
}

function AICertificateDocuments({ certificates }) {
  const usageLabel = (value) => value === "company" ? "Somente empresa" : value === "professional" ? "Somente profissional" : "Empresa e profissional";
  const extractionLabel = (status) => ({ extracted: "PDF com texto", ocr_extracted: "OCR", manual: "Revisão manual", pending_ocr: "OCR pendente", failed: "Leitura falhou" }[status] || "Leitura pendente");
  return <details className="aiEvidenceDetails" open>
    <summary><div><span className="eyebrow">Quadro documental</span><strong>Documentos selecionados</strong><small>Dados cadastrais e quantitativos existentes no acervo</small></div><span>{certificates.length} atestado(s)</span></summary>
    <div className="aiEvidenceBody">
      <div className="aiEvidenceTableWrap"><table className="aiEvidenceTable"><thead><tr><th>Atestado</th><th>Profissional</th><th>Contratante / contratado</th><th>Objeto</th><th>Período</th><th>CAT / função</th><th>Uso</th><th>Leitura</th><th>Quantitativos</th></tr></thead><tbody>
        {certificates.map((item) => {
          const quantities = Array.isArray(item.quantities) ? item.quantities.filter((quantity) => quantity.description || quantity.value || quantity.unit) : [];
          return <tr key={item.id}>
            <td><strong>{item.certificateNumber || item.fileName}</strong><small>{item.completionStatus === "partial" ? "Atestado parcial" : "Atestado final"}</small></td>
            <td>{item.catProfessional || "Não informado"}</td>
            <td><strong>{item.issuerName || "Não informado"}</strong><small>{item.contractedName || "Contratado não informado"}</small></td>
            <td className="aiEvidenceObject">{item.object || "Não informado"}</td>
            <td>{formatCertificateDate(item.executionStart)}<small>até {formatCertificateDate(item.executionEnd)}</small></td>
            <td><strong>{item.catNumber || "Não informada"}</strong><small>{item.professionalRole || "Função não informada"}</small></td>
            <td>{usageLabel(item.usageScope)}</td>
            <td>{extractionLabel(item.extractionStatus)}</td>
            <td>{quantities.length ? <details className="aiQuantityDetails"><summary>{quantities.length} item(ns)</summary><div className="aiQuantityTableWrap"><table><thead><tr><th>Serviço ou quantitativo</th><th>Valor</th><th>Unidade</th><th>Observação</th></tr></thead><tbody>{quantities.map((quantity, index) => <tr key={quantity.id || index}><td>{quantity.description || "Não informado"}</td><td>{quantity.value ?? "—"}</td><td>{quantity.unit || "—"}</td><td>{quantity.note || "—"}</td></tr>)}</tbody></table></div></details> : "Nenhum"}</td>
          </tr>;
        })}
      </tbody></table></div>
      <p className="aiEvidenceNote">Os quantitativos são exibidos na unidade original do cadastro. Grandezas ou serviços diferentes não são somados automaticamente.</p>
    </div>
  </details>;
}

function AICertificateExperience({ certificates }) {
  const groups = useMemo(() => buildTechnicalExperience(certificates), [certificates]);
  const monthNames = ["jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"];
  return <details className="aiEvidenceDetails" open>
    <summary><div><span className="eyebrow">Linha do tempo</span><strong>Experiência profissional consolidada</strong><small>Tempo bruto, sobreposição e experiência cronológica líquida</small></div><span>{groups.length} profissional(is)</span></summary>
    <div className="aiEvidenceBody aiExperienceGroups">
      {groups.map((group) => {
        const raw = formatExperienceDays(group.rawDays);
        const overlap = formatExperienceDays(group.overlapDays);
        const unique = formatExperienceDays(group.uniqueDays);
        const columns = `minmax(180px, 220px) repeat(${Math.max(1, group.months.length)}, 34px)`;
        return <section className="aiExperienceGroup" key={group.key}>
          <div className="aiExperienceHeading"><div><span className="eyebrow">Profissional</span><h4>{group.professionalName}</h4></div><span>{group.certificates.length} período(s) calculado(s)</span></div>
          <div className="aiExperienceMetrics">
            <div><span>Experiência bruta</span><strong>{raw.label}</strong><small>{raw.decimalYears.toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} anos · {raw.days} dias</small></div>
            <div><span>Períodos sobrepostos</span><strong>{overlap.label}</strong><small>{overlap.decimalYears.toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} anos · {overlap.days} dias</small></div>
            <div className="primaryMetric"><span>Experiência sem sobreposição</span><strong>{unique.label}</strong><small>{unique.decimalYears.toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 })} anos · {unique.days} dias</small></div>
          </div>
          {group.certificates.length ? <div className="aiGanttScroll">
            <div className="aiGanttHeader" style={{ gridTemplateColumns: columns }}><strong>Atestado</strong>{group.months.map((month, index) => <span key={month.key}><b>{index === 0 || month.month === 0 ? month.year : "\u00a0"}</b><small>{monthNames[month.month]}</small></span>)}</div>
            {group.certificates.map((item) => <div className="aiGanttRow" style={{ gridTemplateColumns: columns }} key={item.id}><div className="aiGanttLabel"><strong>{item.timelineLabel}</strong><small>{formatCertificateDate(item.executionStart)} a {formatCertificateDate(item.executionEnd)}</small></div>{group.months.map((month) => <i className="aiGanttGridCell" key={month.key} />)}<span className="aiGanttBar" style={{ gridColumn: `${item.startIndex + 2} / ${item.endIndex + 3}` }} title={`${item.timelineLabel}: ${formatCertificateDate(item.executionStart)} a ${formatCertificateDate(item.executionEnd)}`} /></div>)}
            {group.overlapDays > 0 && <div className="aiGanttOverlapRow" style={{ gridTemplateColumns: columns }}><div className="aiGanttLabel"><strong>Sobreposição</strong><small>Meses com mais de um período</small></div>{group.months.map((month, index) => <i className={group.occupancy[index] > 1 ? "isOverlap" : ""} key={month.key} title={group.occupancy[index] > 1 ? `${group.occupancy[index]} experiências no mesmo mês` : ""} />)}</div>}
          </div> : <div className="infoNotice"><p>Nenhum atestado deste profissional possui início e fim válidos para o cálculo.</p></div>}
          {group.invalidCertificates.length > 0 && <p className="aiExperienceWarning">{group.invalidCertificates.length} atestado(s) sem período completo não entraram no cálculo.</p>}
        </section>;
      })}
      <p className="aiEvidenceNote">A experiência sem sobreposição conta cada dia uma única vez para o mesmo profissional. Este quadro é apoio cronológico interno; as regras específicas de cada edital continuam prevalecendo.</p>
    </div>
  </details>;
}

function TechnicalCertificateAI({ navigate }) {
  const [selectedIds] = useState(() => {
    try {
      const saved = JSON.parse(window.sessionStorage.getItem("licitahubTechnicalCertificateAISelection") || "[]");
      return Array.isArray(saved) ? [...new Set(saved.filter((item) => typeof item === "string"))] : [];
    } catch (_error) { return []; }
  });
  const [certificates, setCertificates] = useState([]);
  const [prompt, setPrompt] = useState("");
  const [provider, setProvider] = useState("automatic");
  const [analysis, setAnalysis] = useState(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [message, setMessage] = useState(null);
  const selectedKey = selectedIds.join(",");
  useEffect(() => {
    let active = true;
    const load = async () => {
      if (!selectedIds.length) { if (active) { setCertificates([]); setLoading(false); } return; }
      try {
        const response = await fetch(`${API_BASE_URL}/api/technical-certificates?status=active`, { credentials: "include" });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar os atestados selecionados.");
        if (active) setCertificates((Array.isArray(data) ? data : []).filter((item) => selectedIds.includes(item.id)));
      } catch (error) { if (active) setMessage({ type: "error", text: error.message }); }
      finally { if (active) setLoading(false); }
    };
    load();
    return () => { active = false; };
  }, [selectedKey]);
  useEffect(() => {
    if (!analysis?.id || !["queued", "processing"].includes(analysis.status)) return undefined;
    let active = true;
    const refresh = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/technical-certificate-ai-analyses/${encodeURIComponent(analysis.id)}`, { credentials: "include" });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível consultar a análise.");
        if (active) setAnalysis(data);
      } catch (error) { if (active) setMessage({ type: "error", text: error.message }); }
    };
    refresh();
    const timer = window.setInterval(refresh, 2500);
    return () => { active = false; window.clearInterval(timer); };
  }, [analysis?.id, analysis?.status]);
  const submit = async () => {
    if (!prompt.trim()) { setMessage({ type: "error", text: "Descreva no roteiro o que a IA deve analisar." }); return; }
    if (!certificates.length) { setMessage({ type: "error", text: "Nenhum atestado selecionado está disponível para análise." }); return; }
    setSending(true); setMessage({ type: "info", text: "Preparando o JSON e enviando os atestados selecionados para análise." });
    try {
      const response = await fetch(`${API_BASE_URL}/api/technical-certificate-ai-analyses`, { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ certificateIds: certificates.map((item) => item.id), prompt: prompt.trim(), provider }) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível iniciar a análise por IA.");
      window.sessionStorage.removeItem("licitahubTechnicalCertificateAISelection");
      setAnalysis(data);
      setMessage({ type: "info", text: "Análise enviada. O resultado aparecerá nesta tela quando for concluído." });
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setSending(false); }
  };
  const statusLabel = { queued: "Aguardando processamento", processing: "IA analisando os atestados", completed: "Análise concluída", failed: "Análise não concluída" };
  const providerLabel = { openai: "OpenAI", gemini: "Google Gemini", groq: "Groq", automatic: "Seleção automática" };
  return <Page label="Capacidade técnica" title="Análise de atestados por IA" actions={<Button variant="secondary" onClick={() => navigate("technical-certificates")}>Voltar aos atestados</Button>}>
    <Card className="certificateIntro"><span className="eyebrow">Análise privada da empresa</span><h3>Capacidade técnica assistida por IA</h3><p>O sistema envia somente os dados estruturados e o texto capturado dos atestados marcados. O roteiro abaixo orienta a análise.</p></Card>
    {message && <Card className={`formFeedback ${message.type === "success" ? "success" : message.type === "info" ? "infoNotice" : "dangerNotice"}`}><p>{message.text}</p></Card>}
    {loading ? <Card><p>Carregando os atestados selecionados...</p></Card> : certificates.length === 0 ? <Card className="dangerNotice"><p>Selecione um ou mais atestados na tela anterior antes de abrir a análise por IA.</p></Card> : <><div className="aiEvidencePanels"><AICertificateDocuments certificates={certificates} /><AICertificateExperience certificates={certificates} /></div>
      <Card className="certificateForm"><FormGrid><Field label="Provedor de IA" hint="No modo automático, o sistema tenta OpenAI, Google Gemini e Groq, nessa ordem."><select value={provider} onChange={(event) => setProvider(event.target.value)} disabled={sending || !!analysis && ["queued", "processing"].includes(analysis.status)}><option value="automatic">Automático: OpenAI, Gemini e Groq</option><option value="openai">Somente OpenAI</option><option value="gemini">Somente Google Gemini</option><option value="groq">Somente Groq</option></select></Field></FormGrid><Field label="Roteiro da análise" hint={`${prompt.length}/16000 caracteres. Explique o objetivo, os critérios e o formato de resultado desejado.`}><textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} maxLength="16000" placeholder="Ex.: Analise os atestados e identifique quais comprovam supervisão ambiental em rodovias, relacionando os quantitativos, as limitações do texto e os pontos que precisam de conferência humana." /></Field><div className="formActions"><Button onClick={submit} disabled={sending || !!analysis && ["queued", "processing"].includes(analysis.status)}>{sending ? "Enviando para IA..." : analysis && ["queued", "processing"].includes(analysis.status) ? "Análise em andamento..." : "Analisar atestados"}</Button></div></Card></>}
    {analysis && <Card className="aiAnalysisResult"><div className="sectionTitle"><div><span className="eyebrow">Resultado da análise</span><h3>{statusLabel[analysis.status] || analysis.status}</h3></div><span className={`statusPill ${analysis.status === "completed" ? "success" : analysis.status === "failed" ? "review" : "open"}`}>{providerLabel[analysis.provider] || "IA"}{analysis.model ? ` · ${analysis.model}` : ""}</span></div>{["queued", "processing"].includes(analysis.status) && <p>A IA está trabalhando. Esta tela será atualizada automaticamente quando o resultado estiver pronto.</p>}{analysis.status === "failed" && <p className="dangerText">{analysis.errorMessage || "Não foi possível concluir a análise."}</p>}{analysis.status === "completed" && <><p className="aiProviderNotice">Resposta gerada por <strong>{providerLabel[analysis.provider] || "IA"}</strong>{analysis.model ? `, usando o modelo ${analysis.model}.` : "."}</p><pre className="aiAnalysisText">{analysis.resultText}</pre></>}</Card>}
  </Page>;
}

function PNCPCapture({ navigate }) {
  const localISO = () => new Date().toISOString().slice(0, 10);
  const sevenDaysAgo = () => { const date = new Date(); date.setDate(date.getDate() - 7); return date.toISOString().slice(0, 10); };
  const pncpModalities = [{ value: "6", label: "6 - Mão de obra" }, { value: "7", label: "7 - Obras" }, { value: "8", label: "8 - Serviços" }, { value: "9", label: "9 - Serviços de Engenharia" }];
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [capturing, setCapturing] = useState(false);
  const [capturePage, setCapturePage] = useState(1);
  const [captureHasMore, setCaptureHasMore] = useState(true);
  const [discardingAll, setDiscardingAll] = useState(false);
  const [actingId, setActingId] = useState("");
  const [selectedIds, setSelectedIds] = useState([]);
  const [aiJob, setAIJob] = useState(null);
  const [startingAI, setStartingAI] = useState(false);
  const [message, setMessage] = useState(null);
  const [filters, setFilters] = useState({ startDate: sevenDaysAgo(), endDate: localISO(), state: "", source: "pncp", modalities: pncpModalities.map((option) => option.value), status: "captured", relevance: "relevant", aiClassification: "", search: "", minValue: "", maxValue: "" });
  const load = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/captures`, { credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível carregar a fila de captação.");
      const loadedItems = Array.isArray(data) ? data : Array.isArray(data.items) ? data.items : [];
      setItems(loadedItems);
      setSelectedIds((current) => current.filter((id) => loadedItems.some((item) => item.id === id && item.status === "captured")));
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally { setLoading(false); }
  };
  const loadAIJob = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/ai-analysis`, { credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível consultar a análise por IA.");
      setAIJob(data);
      return data;
    } catch (error) {
      setMessage({ type: "error", text: error.message });
      return null;
    }
  };
  useEffect(() => { load(); loadAIJob(); }, []);
  useEffect(() => {
    if (!aiJob || !["queued", "processing"].includes(aiJob.status)) return undefined;
    const timer = window.setInterval(async () => {
      const updated = await loadAIJob();
      if (updated && ["completed", "failed"].includes(updated.status)) await load();
    }, 2500);
    return () => window.clearInterval(timer);
  }, [aiJob?.id, aiJob?.status]);
  const resetCapture = () => { setCapturePage(1); setCaptureHasMore(true); };
  const capture = async () => {
    setCapturing(true); setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/captures`, { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ startDate: filters.startDate, endDate: filters.endDate, state: filters.state, source: filters.source, page: capturePage, modalities: filters.source === "pncp" ? filters.modalities : [] }) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível consultar a fonte escolhida.");
	  const captured = Number(data.captured || 0);
      setCaptureHasMore(captured > 0);
      if (captured > 0) setCapturePage((current) => current + 1);
      setMessage({ type: "success", text: `Página ${capturePage} consultada: ${captured} oportunidade(s) recebida(s).` });
	  setFilters((current) => ({ ...current, status: "captured", relevance: "all", search: "" }));
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setCapturing(false); }
  };
  const decide = async (item, action) => {
    const preparing = action === "prepare";
    const label = preparing ? "preparar o cadastro" : "descartar";
    if (!window.confirm(`Deseja ${label} o edital ${item.number}?`)) return;
    setActingId(item.id); setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/captures/${encodeURIComponent(item.id)}/decision`, { method: "PATCH", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ action }) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível atualizar a captação.");
      if (preparing) {
        navigate(`tender-new?id=${encodeURIComponent(data.tenderId)}`);
        return;
      }
      setMessage({ type: "success", text: "Captação descartada." });
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setActingId(""); }
  };
  const discardAllPending = async () => {
    const pendingCount = items.filter((item) => item.status === "captured").length;
    if (pendingCount === 0) return;
    if (!window.confirm(`Descartar as ${pendingCount} oportunidade(s) pendente(s)? Esta ação não altera editais preparados ou publicados.`)) return;
    setDiscardingAll(true); setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/captures/discard-pending`, { method: "POST", credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível descartar as captações pendentes.");
      setMessage({ type: "success", text: `${data.discarded || 0} oportunidade(s) pendente(s) foram descartadas.` });
      await load();
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setDiscardingAll(false); }
  };
  const startAIAnalysis = async (onlySelected) => {
    const captureIds = onlySelected ? selectedIds : [];
    if (onlySelected && captureIds.length === 0) {
      setMessage({ type: "error", text: "Marque ao menos uma oportunidade pendente para analisar." });
      return;
    }
    const total = onlySelected ? captureIds.length : items.filter((item) => item.status === "captured").length;
    if (total === 0) {
      setMessage({ type: "error", text: "Não existem oportunidades pendentes para analisar." });
      return;
    }
    if (!window.confirm(`Enviar ${total} oportunidade(s) para classificação por IA? O sistema tentará OpenAI, Gemini e Groq, nessa ordem.`)) return;
    setStartingAI(true); setMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/pncp/ai-analysis`, { method: "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify({ captureIds }) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Não foi possível iniciar a análise por IA.");
      setAIJob(data);
      setSelectedIds([]);
      setMessage({ type: "success", text: `Análise iniciada para ${data.totalCount || total} oportunidade(s). Você pode acompanhar o andamento nesta tela.` });
    } catch (error) { setMessage({ type: "error", text: error.message }); }
    finally { setStartingAI(false); }
  };
  const toggleSelection = (id) => setSelectedIds((current) => current.includes(id) ? current.filter((item) => item !== id) : [...current, id]);
  const filtered = items.filter((item) => {
    const term = filters.search.trim().toLowerCase();
    const matchesTerm = !term || [item.agency, item.number, item.object, item.city].some((value) => String(value || "").toLowerCase().includes(term));
    const matchesStatus = !filters.status || item.status === filters.status;
    const matchesRelevance = filters.relevance !== "relevant" || Number(item.relevanceScore || 0) >= 45;
    const value = Number(item.estimatedValue);
    const matchesMinValue = !filters.minValue || (Number.isFinite(value) && value >= Number(filters.minValue));
    const matchesMaxValue = !filters.maxValue || (Number.isFinite(value) && value <= Number(filters.maxValue));
    const matchesAI = !filters.aiClassification || (filters.aiClassification === "pending" ? !item.aiClassification : item.aiClassification === filters.aiClassification);
    return matchesTerm && matchesStatus && matchesRelevance && matchesMinValue && matchesMaxValue && matchesAI;
  });
  const relevanceLabel = (score) => Number(score || 0) >= 70 ? "Alta aderência" : Number(score || 0) >= 45 ? "Possível aderência" : "Baixa aderência";
  const aiClassificationLabel = (classification) => ({ consultiva: "IA: Engenharia consultiva", relacionada: "IA: Atividade relacionada", nao_consultiva: "IA: Não consultiva", duvidosa: "IA: Revisão necessária" }[classification] || "Aguardando análise da IA");
  const providerLabel = (provider) => provider === "gemini" ? "Google Gemini" : provider === "groq" ? "Groq" : provider === "openai" ? "OpenAI" : "IA";
  const statusLabel = (status) => ({ captured: "Aguardando decisão", prepared: "Em edição", approved: "Publicado", discarded: "Descartado" }[status] || status);
  const pendingCount = items.filter((item) => item.status === "captured").length;
  const aiRunning = aiJob && ["queued", "processing"].includes(aiJob.status);
  const aiProgress = aiJob?.totalCount ? Math.min(100, Math.round((Number(aiJob.processedCount || 0) / Number(aiJob.totalCount)) * 100)) : 0;
  const selectableFiltered = filtered.filter((item) => item.status === "captured");
  const allFilteredSelected = selectableFiltered.length > 0 && selectableFiltered.every((item) => selectedIds.includes(item.id));
  return <Page label="Editais" title="Captação de oportunidades" actions={<><Button variant="secondary" onClick={load}>Recarregar fila</Button><Button variant="secondary" onClick={discardAllPending} disabled={discardingAll || pendingCount === 0}>{discardingAll ? "Descartando..." : `Descartar pendentes (${pendingCount})`}</Button><Button variant="secondary" onClick={() => navigate("tender-admin")}>Admin de editais</Button></>}>
    <Card className="pncpIntro"><span className="eyebrow">Fontes oficiais</span><h3>Fila de revisão da engenharia consultiva</h3><p>Consulte o PNCP ou o Compras.gov.br, filtre a aderência técnica e publique somente os editais revisados pelo administrador.</p></Card>
    {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
    <Card className="compactFilters pncpFiltersSticky"><FormGrid>
      <Field label="Publicados de"><input type="date" value={filters.startDate} max={filters.endDate} onChange={(event) => { resetCapture(); setFilters((current) => ({ ...current, startDate: event.target.value })); }} /></Field>
      <Field label="Até"><input type="date" value={filters.endDate} min={filters.startDate} max={localISO()} onChange={(event) => { resetCapture(); setFilters((current) => ({ ...current, endDate: event.target.value })); }} /></Field>
      <Field label="Estado"><StateSelect value={filters.state} onChange={(event) => { resetCapture(); setFilters((current) => ({ ...current, state: event.target.value })); }} /></Field>
      <Field label="Fonte"><select value={filters.source} onChange={(event) => { resetCapture(); setFilters((current) => ({ ...current, source: event.target.value })); }}><option value="pncp">PNCP</option><option value="comprasgov">Compras.gov.br</option></select></Field>
      {filters.source === "pncp" && <Field label="Modalidades PNCP"><select multiple size="4" value={filters.modalities} onChange={(event) => { resetCapture(); setFilters((current) => ({ ...current, modalities: Array.from(event.target.selectedOptions).map((option) => option.value) })); }}>{pncpModalities.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}</select></Field>}
      <Field label="Consulta"><Button onClick={capture} disabled={capturing || !captureHasMore || (filters.source === "pncp" && filters.modalities.length === 0)}>{capturing ? `Consultando página ${capturePage}...` : !captureHasMore ? "Pesquisa concluída" : capturePage === 1 ? "Buscar página 1" : `Buscar próxima página (${capturePage})`}</Button></Field>
    </FormGrid></Card>
    <Card className="pncpProgressCard"><div className="pncpProgressHeader"><div><span className="eyebrow">Andamento da pesquisa</span><strong>{capturing ? `Consultando página ${capturePage}` : captureHasMore ? `Próxima página: ${capturePage}` : "Pesquisa concluída"}</strong></div><span>{capturing ? "Em andamento" : "Pronto"}</span></div><div className={`pncpProgressTrack ${capturing ? "isRunning" : ""}`}><i /></div><small>{capturing ? "Aguarde o retorno da fonte oficial." : "Cada página recebida é salva na fila de captação."}</small></Card>
    <Card className="pncpAIControl">
      <div className="pncpAIControlHeader"><div><span className="eyebrow">Triagem assistida</span><h3>Classificação por inteligência artificial</h3><p>O sistema tenta OpenAI, Google Gemini e Groq, nessa ordem. A decisão de publicar ou descartar continua sendo do administrador.</p></div><span className={`statusPill ${aiRunning ? "open" : aiJob?.status === "completed" ? "success" : aiJob?.status === "failed" ? "review" : ""}`}>{aiRunning ? "Em processamento" : aiJob?.status === "completed" ? "Última análise concluída" : aiJob?.status === "failed" ? "Última análise falhou" : "Pronto para analisar"}</span></div>
      {aiJob && <div className="pncpAIProgress"><div><span>Progresso</span><strong>{Number(aiJob.processedCount || 0)} de {Number(aiJob.totalCount || 0)}</strong></div><div className="pncpAIProgressTrack"><i style={{ width: `${aiProgress}%` }} /></div><small>{aiJob.status === "failed" ? aiJob.errorMessage || "A análise não foi concluída." : `${Number(aiJob.successCount || 0)} classificada(s) e ${Number(aiJob.failureCount || 0)} com falha.`}</small></div>}
      <div className="pncpAIActions"><Button onClick={() => startAIAnalysis(false)} disabled={startingAI || aiRunning || pendingCount === 0}>{startingAI ? "Iniciando análise..." : `Analisar pendentes (${pendingCount})`}</Button><Button variant="secondary" onClick={() => startAIAnalysis(true)} disabled={startingAI || aiRunning || selectedIds.length === 0}>Analisar selecionadas ({selectedIds.length})</Button></div>
    </Card>
    <Card className="compactFilters pncpQueueFiltersSticky"><FormGrid>
      <Field label="Buscar na fila"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Órgão, número, objeto ou cidade" /></Field>
      <Field label="Valor mínimo"><input type="number" min="0" step="0.01" value={filters.minValue} onChange={(event) => setFilters((current) => ({ ...current, minValue: event.target.value }))} placeholder="R$ 0,00" /></Field>
      <Field label="Valor máximo"><input type="number" min="0" step="0.01" value={filters.maxValue} onChange={(event) => setFilters((current) => ({ ...current, maxValue: event.target.value }))} placeholder="Sem limite" /></Field>
      <Field label="Situação"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todas</option><option value="captured">Aguardando decisão</option><option value="prepared">Em edição</option><option value="approved">Publicados</option><option value="discarded">Descartados</option></select></Field>
      <Field label="Aderência"><select value={filters.relevance} onChange={(event) => setFilters((current) => ({ ...current, relevance: event.target.value }))}><option value="relevant">Somente possíveis aderentes</option><option value="all">Todas as capturas</option></select></Field>
      <Field label="Classificação da IA"><select value={filters.aiClassification} onChange={(event) => setFilters((current) => ({ ...current, aiClassification: event.target.value }))}><option value="">Todas</option><option value="consultiva">Somente engenharia consultiva</option><option value="relacionada">Atividade relacionada</option><option value="duvidosa">Revisão necessária</option><option value="nao_consultiva">Não consultiva</option><option value="pending">Ainda não analisada</option></select></Field>
    </FormGrid></Card>
    {!loading && <div className="pncpQueueSummary"><span><strong>{filtered.length}</strong> de <strong>{items.length}</strong> oportunidade(s) exibida(s) na fila.</span>{selectableFiltered.length > 0 && <label className="pncpSelectAll"><input type="checkbox" checked={allFilteredSelected} onChange={() => setSelectedIds((current) => allFilteredSelected ? current.filter((id) => !selectableFiltered.some((item) => item.id === id)) : Array.from(new Set([...current, ...selectableFiltered.map((item) => item.id)])))} /> Selecionar pendentes exibidas</label>}</div>}
    {loading && <Card><p>Carregando fila de captação...</p></Card>}
    {!loading && filtered.length === 0 && <Card><p>Nenhuma oportunidade encontrada. Faça uma consulta na fonte escolhida ou amplie os filtros.</p></Card>}
    {!loading && filtered.length > 0 && <div className="pncpCaptureList">{filtered.map((item) => <Card className={`pncpCaptureCard ${selectedIds.includes(item.id) ? "isSelected" : ""}`} key={item.id}>
      <div className="pncpCaptureHeader"><div className="pncpCaptureIdentity">{item.status === "captured" && <label className="pncpCaptureSelect" title="Selecionar para análise por IA"><input type="checkbox" checked={selectedIds.includes(item.id)} onChange={() => toggleSelection(item.id)} /><span>Selecionar</span></label>}<span className="eyebrow">{item.agency}</span><h3>{item.number}</h3><span className="captureSource">{item.source === "pncp+comprasgov" ? "PNCP + Compras.gov.br / saneado" : item.source === "comprasgov" ? "Compras.gov.br / Dados Abertos" : "PNCP"}</span>{item.pncpControlNumber && <small className="captureControlNumber">Controle PNCP: {item.pncpControlNumber}</small>}</div><div className="pncpClassificationBadges"><span className={`pncpRelevance score-${Number(item.relevanceScore || 0) >= 70 ? "high" : Number(item.relevanceScore || 0) >= 45 ? "medium" : "low"}`}>{relevanceLabel(item.relevanceScore)}</span><span className={`pncpAIClassification ai-${item.aiClassification || "pending"}`}>{aiClassificationLabel(item.aiClassification)}{item.aiClassification ? ` · ${Number(item.aiConfidence || 0)}%` : ""}</span></div></div>
      <p className="pncpObject">{item.object}</p>
      <div className="pncpInfoGrid">
        <div className="pncpInfo"><span>Modalidade</span><strong>{item.modality || "Não informada"}</strong></div>
        <div className="pncpInfo"><span>Julgamento</span><strong>{item.judgmentCriterion || "Não informado"}</strong></div>
        <div className="pncpInfo"><span>Local</span><strong>{[item.city, item.state].filter(Boolean).join(" - ") || "Não informado"}</strong></div>
        <div className="pncpInfo"><span>Sessão</span><strong>{item.openingDate ? new Date(item.openingDate).toLocaleDateString("pt-BR") : "Não informada"}</strong></div>
        <div className="pncpValue"><span>Valor estimado</span><strong>{item.estimatedValue !== null && item.estimatedValue !== undefined ? formatCurrencyBRFromNumber(item.estimatedValue) : "Não informado"}</strong></div>
      </div>
      <div className="pncpReasons">{(item.relevanceReasons || []).map((reason) => <span key={reason}>{reason}</span>)}</div>
      {item.aiClassification && <div className="pncpAIResult"><div><strong>Leitura da IA</strong><span>{providerLabel(item.aiProvider)}{item.aiModel ? ` · ${item.aiModel}` : ""}</span></div><p>{item.aiJustification}</p>{Array.isArray(item.aiAreas) && item.aiAreas.length > 0 && <div>{item.aiAreas.map((area) => <span key={area}>{area}</span>)}</div>}</div>}
      <div className="pncpActions"><a className="textLink" href={item.externalUrl || "#"} target="_blank" rel="noreferrer">Ver fonte oficial</a>{item.status === "captured" && <><Button onClick={() => decide(item, "prepare")} disabled={actingId === item.id}>{actingId === item.id ? "Preparando..." : "Preparar cadastro"}</Button><Button variant="secondary" onClick={() => decide(item, "discard")} disabled={actingId === item.id}>Descartar</Button></>}{item.status !== "captured" && <span className={`statusBadge status-${item.status}`}>{statusLabel(item.status)}</span>}{item.status === "prepared" && item.publishedTenderId && <Button variant="secondary" onClick={() => navigate(`tender-new?id=${encodeURIComponent(item.publishedTenderId)}`)}>Continuar cadastro</Button>}{item.status === "approved" && item.publishedTenderId && <Button variant="secondary" onClick={() => navigate(`tender-detail?id=${encodeURIComponent(item.publishedTenderId)}`)}>Ver edital</Button>}</div>
    </Card>)}</div>}
  </Page>;
}

function TenderNew({ navigate }) {
  const editId = currentHashParams().get("id") || "";
  const isEditing = Boolean(editId);
  const [form, setForm] = useState({ agency: "", number: "", object: "", modality: "", judgmentCriterion: "", estimatedValue: "", state: "", city: "", openingDate: "", status: "published", cloudFolderUrl: "", analysisDataUrl: "", analysisFileName: "", analysisMimeType: "" });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);
  const [documentFiles, setDocumentFiles] = useState([]);
  const [originalOpeningDate, setOriginalOpeningDate] = useState("");
  const update = (field, value) => setForm((current) => {
    const next = { ...current, [field]: value };
    if (field === "openingDate" && !isPastDateISO(value) && current.status === "occurred") next.status = "published";
    return next;
  });
  useEffect(() => {
    if (!editId) return;
    fetch(`${API_BASE_URL}/api/tenders/${editId}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o edital.");
        return data;
      })
      .then((data) => {
        const openingDate = data.openingDate ? String(data.openingDate).slice(0, 10) : "";
        setOriginalOpeningDate(openingDate);
        setForm({
          agency: data.agency || "",
          number: data.number || "",
          object: data.object || "",
          modality: data.modality || "",
          judgmentCriterion: data.judgmentCriterion || "",
          estimatedValue: formatCurrencyBRFromNumber(data.estimatedValue),
          state: data.state || "",
          city: data.city || "",
          openingDate,
          status: data.status || "published",
          cloudFolderUrl: data.cloudFolderUrl || "",
          analysisDataUrl: "",
          analysisFileName: "",
          analysisMimeType: ""
        });
      })
      .catch((err) => setMessage({ type: "error", text: err.message }));
  }, [editId]);
  const selectAnalysis = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith(".html") && !file.name.toLowerCase().endsWith(".htm")) {
      setMessage({ type: "error", text: "Selecione um arquivo HTML." });
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setForm((current) => ({ ...current, analysisDataUrl: String(reader.result || ""), analysisFileName: file.name, analysisMimeType: "text/html" }));
    reader.readAsDataURL(file);
  };
  const selectDocuments = (event) => {
    const selected = Array.from(event.target.files || []);
    const invalid = selected.find((file) => file.size > 25 * 1024 * 1024);
    if (invalid) {
      setMessage({ type: "error", text: `O arquivo ${invalid.name} ultrapassa o limite de 25MB.` });
      event.target.value = "";
      return;
    }
    setDocumentFiles((current) => {
      const next = [...current];
      selected.forEach((file) => {
        const alreadyIncluded = next.some((item) => item.name === file.name && item.size === file.size && item.lastModified === file.lastModified);
        if (!alreadyIncluded) next.push(file);
      });
      return next;
    });
    event.target.value = "";
  };
  const uploadDocuments = async (tenderId) => {
    const failures = [];
    for (const file of documentFiles) {
      try {
        const fileDataUrl = await new Promise((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = () => resolve(String(reader.result || ""));
          reader.onerror = () => reject(new Error("Não foi possível ler o arquivo."));
          reader.readAsDataURL(file);
        });
        const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(tenderId)}/files`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ fileDataUrl, fileName: file.name, mimeType: file.type })
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(data.error || "Falha ao enviar arquivo.");
      } catch (error) {
        failures.push(`${file.name}: ${error.message}`);
      }
    }
    if (failures.length === 0) setDocumentFiles([]);
    return failures;
  };
  const formatDocumentSize = (bytes) => bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / (1024 * 1024)).toFixed(1).replace(".", ",")} MB`;
  const confirmSameValuePublication = async (payload) => {
    if (payload.status !== "published" || !payload.estimatedValue) return true;
    const response = await fetch(`${API_BASE_URL}/api/tenders/value-conflicts`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ estimatedValue: payload.estimatedValue, excludeTenderId: isEditing ? editId : "" })
    });
    const conflicts = await response.json().catch(() => []);
    if (!response.ok) throw new Error(conflicts.error || "Não foi possível verificar editais com valor semelhante.");
    if (!Array.isArray(conflicts) || conflicts.length === 0) return true;
    const details = conflicts.slice(0, 5).map((item) => `• Edital ${item.number}\n  Órgão: ${item.agency}\n  Objeto: ${item.object}`).join("\n\n");
    return window.confirm(`Atenção: há ${conflicts.length} edital(is) publicado(s) com o mesmo valor estimado (${formatCurrencyBRFromNumber(payload.estimatedValue)}).\n\n${details}\n\nDeseja publicar mesmo assim?`);
  };
  const submit = async () => {
    setSaving(true);
    setMessage(null);
    try {
      if (form.openingDate && isPastDateISO(form.openingDate) && (!isEditing || form.openingDate !== originalOpeningDate)) {
        throw new Error("A data de abertura deve ser hoje ou futura. Um edital histórico pode apenas manter a data já registrada.");
      }
      const payload = {
        ...form,
        estimatedValue: parseCurrencyBR(form.estimatedValue),
        status: isPastDateISO(form.openingDate) && !["closed", "cancelled"].includes(form.status) ? "occurred" : form.status
      };
      const confirmedValueConflict = await confirmSameValuePublication(payload);
      if (!confirmedValueConflict) return;
      payload.confirmValueConflict = payload.status === "published";
      const response = await fetch(isEditing ? `${API_BASE_URL}/api/tenders/${editId}` : `${API_BASE_URL}/api/tenders`, { method: isEditing ? "PUT" : "POST", headers: { "Content-Type": "application/json" }, credentials: "include", body: JSON.stringify(payload) });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível salvar o edital.");
      const uploadFailures = documentFiles.length ? await uploadDocuments(data.id) : [];
      if (uploadFailures.length > 0) {
        setMessage({ type: "error", text: `Edital salvo, mas ${uploadFailures.length} arquivo(s) não foram enviados. ${uploadFailures[0]}` });
        return;
      }
      setMessage({ type: "success", text: `Edital ${data.number} ${isEditing ? "atualizado" : "cadastrado"} no banco.${documentFiles.length ? " Documentos anexados." : ""}` });
      if (isEditing) {
        setTimeout(() => navigate("tender-admin"), 500);
      } else {
        setForm({ agency: "", number: "", object: "", modality: "", judgmentCriterion: "", estimatedValue: "", state: "", city: "", openingDate: "", status: "published", cloudFolderUrl: "", analysisDataUrl: "", analysisFileName: "", analysisMimeType: "" });
      }
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving(false);
    }
  };
  return (
    <Page label="Editais" title={isEditing ? "Editar edital" : "Cadastro de edital"}>
      {message && <Card className={`formFeedback ${message.type === "success" ? "success" : "dangerNotice"}`}><p>{message.text}</p></Card>}
      <FormGrid>
        <Field label="Órgão público"><input value={form.agency} onChange={(e) => update("agency", e.target.value)} required /></Field>
        <Field label="Número do edital"><input value={form.number} onChange={(e) => update("number", e.target.value)} required /></Field>
        <Field label="Modalidade"><select value={form.modality} onChange={(e) => update("modality", e.target.value)}><option value="">Selecione</option>{tenderModalityOptions.map((option) => <option value={option} key={option}>{option}</option>)}</select></Field>
        <Field label="Critério de julgamento"><select value={form.judgmentCriterion} onChange={(e) => update("judgmentCriterion", e.target.value)}><option value="">Selecione</option><option>Menor preço</option><option>Técnica e preço</option><option>Melhor técnica</option></select></Field>
        <Field label="Valor estimado"><input value={form.estimatedValue} onChange={(e) => update("estimatedValue", formatCurrencyBR(e.target.value))} placeholder="R$ 0,00" inputMode="numeric" /></Field>
        <Field label="Estado"><StateSelect value={form.state} onChange={(e) => update("state", e.target.value)} /></Field>
        <Field label="Cidade"><CityField state={form.state} value={form.city} onChange={(e) => update("city", e.target.value)} /></Field>
        <Field label="Data de abertura"><input type="date" min={isEditing && form.openingDate === originalOpeningDate && isPastDateISO(form.openingDate) ? undefined : currentLocalISODate()} value={form.openingDate} onChange={(e) => update("openingDate", e.target.value)} /></Field>
        <Field label="Status"><select value={form.status} onChange={(e) => update("status", e.target.value)}><option value="draft">Rascunho</option><option value="published">Publicado</option><option value="under_review">Em análise</option><option value="suspended">Suspenso</option><option value="challenged">Impugnado</option><option value="occurred">Ocorrido</option><option value="closed">Encerrado</option><option value="cancelled">Cancelado</option></select></Field>
      </FormGrid>
      <Field label="Objeto"><textarea value={form.object} onChange={(e) => update("object", e.target.value)} required /></Field>
      <Field label="Link do diretório em nuvem"><input value={form.cloudFolderUrl} onChange={(e) => update("cloudFolderUrl", e.target.value)} placeholder="https://drive.google.com/..." /></Field>
      <Field label="Arquivos do edital" hint="Selecione quantos documentos forem necessários. Cada arquivo pode ter até 25 MB.">
        <input type="file" multiple accept=".pdf,.doc,.docx,.xls,.xlsx,.csv,.ppt,.pptx,.odt,.ods,.txt,.xml,.zip,.rar,.7z,.dwg,.dxf,.jpg,.jpeg,.png,.webp" onChange={selectDocuments} />
        {documentFiles.length === 0 ? <small>Nenhum documento adicional selecionado.</small> : <div className="tenderDocumentQueue">{documentFiles.map((file, index) => <div key={`${file.name}-${file.lastModified}-${index}`}><span><strong>{file.name}</strong><small>{formatDocumentSize(file.size)}</small></span><button type="button" title="Retirar arquivo" aria-label={`Retirar ${file.name}`} onClick={() => setDocumentFiles((current) => current.filter((_, itemIndex) => itemIndex !== index))}>×</button></div>)}</div>}
      </Field>
      {!isEditing && <Field label="Análise do edital em HTML" hint="Será exibida no detalhe e poderá ser baixada."><input type="file" accept=".html,.htm,text/html" onChange={selectAnalysis} /><small>{form.analysisFileName || "Nenhum arquivo selecionado"}</small></Field>}
      {isEditing && <Card className="notice"><p>A análise HTML pode ser anexada ou substituída no detalhe do edital.</p></Card>}
      <div className="formActionBar"><Button onClick={submit} disabled={saving}>{saving ? "Salvando..." : isEditing ? "Salvar alterações" : "Cadastrar edital"}</Button></div>
    </Page>
  );
}

function AssemblyCenter({ navigate }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", type: "", status: "" });
  const [sort, setSort] = useState("opening_asc");

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/assembly-list`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => []);
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar as montagens.");
        return data;
      })
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((loadError) => setError(loadError.message))
      .finally(() => setLoading(false));
  }, []);

  const statusLabels = { preparing: "Planejamento", in_progress: "Em andamento", under_review: "Em revisão", ready_to_submit: "Pronta para envio", submitted: "Enviada" };
  const progressTone = (progress) => Number(progress || 0) > 70 ? "progressGood" : Number(progress || 0) > 30 ? "progressAttention" : "progressCritical";
  const formatDate = (value) => value ? new Date(`${String(value).slice(0, 10)}T12:00:00`).toLocaleDateString("pt-BR") : "Sem sessão informada";
  const visibleItems = useMemo(() => {
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    const filtered = items.filter((item) => {
      const matchesSearch = !term || [item.tenderNumber, item.agency, item.tenderObject, item.leadCompanyName].some((value) => String(value || "").toLocaleLowerCase("pt-BR").includes(term));
      return matchesSearch && (!filters.type || item.assemblyType === filters.type) && (!filters.status || item.status === filters.status);
    });
    return [...filtered].sort((left, right) => {
      if (sort === "progress_desc") return Number(right.progress || 0) - Number(left.progress || 0);
      if (sort === "updated_desc") return new Date(right.updatedAt || 0) - new Date(left.updatedAt || 0);
      return new Date(left.openingDate || "9999-12-31") - new Date(right.openingDate || "9999-12-31");
    });
  }, [items, filters, sort]);

  return <Page label="Editais" title="Central de montagens" actions={<Button variant="secondary" onClick={() => navigate("tender-calendar")}>Ver calendário</Button>}>
    <section className="assemblyCenterHero">
      <div><span className="eyebrow">Acompanhamento operacional</span><h3>Todas as licitações em montagem</h3><p>Consulte rapidamente o andamento de participações individuais e consórcios. Cada item abre o plano completo da licitação.</p></div>
      <div className="assemblyCenterSummary"><strong>{items.length}</strong><span>montagem{items.length === 1 ? "" : "ens"} ativa{items.length === 1 ? "" : "s"}</span></div>
    </section>
    <Card className="compactFilters assemblyCenterFilters">
      <FormGrid>
        <Field label="Buscar"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Edital, órgão, objeto ou empresa líder" /></Field>
        <Field label="Tipo"><select value={filters.type} onChange={(event) => setFilters((current) => ({ ...current, type: event.target.value }))}><option value="">Todos</option><option value="individual">Participação individual</option><option value="consortium">Consórcio</option></select></Field>
        <Field label="Andamento"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option>{Object.entries(statusLabels).map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
        <Field label="Ordenar por"><select value={sort} onChange={(event) => setSort(event.target.value)}><option value="opening_asc">Data da sessão</option><option value="updated_desc">Atualização mais recente</option><option value="progress_desc">Maior percentual concluído</option></select></Field>
      </FormGrid>
    </Card>
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {loading ? <Card><p>Carregando as montagens...</p></Card> : <section className="assemblyCenterGrid">
      {visibleItems.map((item) => { const progress = Math.max(0, Math.min(100, Number(item.progress || 0))); const members = Array.isArray(item.members) ? item.members : []; return <article className="assemblyCenterCard" key={item.id}>
        <div className="assemblyCenterCardHead"><span className={`assemblyTypeBadge ${item.assemblyType}`}>{item.assemblyType === "individual" ? "Individual" : "Consórcio"}</span><span className="assemblyStatusBadge">{statusLabels[item.status] || item.status}</span></div>
        <button type="button" className="assemblyCenterOpen" onClick={() => navigate(`assembly-board?assemblyId=${encodeURIComponent(item.id)}`)} title="Abrir Central de Montagem"><strong>{item.tenderNumber} · {item.agency}</strong><span>{item.tenderObject}</span></button>
        <dl className="assemblyCenterInfo"><div><dt>Sessão</dt><dd>{formatDate(item.openingDate)}</dd></div><div><dt>{item.assemblyType === "individual" ? "Responsável" : "Empresa líder"}</dt><dd>{item.leadCompanyName || "Não definida"}</dd></div><div><dt>Tarefas abertas</dt><dd>{item.openTaskCount || 0}</dd></div></dl>
        {item.assemblyType === "consortium" && members.length > 0 && <div className="assemblyCenterMembers"><span>Participantes</span><p>{members.map((member) => member.companyName).filter(Boolean).join(" · ")}</p></div>}
        <div className="assemblyCenterProgress"><div><span>Concluído</span><strong>{progress}%</strong></div><em className={progressTone(progress)}><i style={{ width: `${progress}%` }} /></em></div>
        <Button variant="secondary" onClick={() => navigate(`assembly-board?assemblyId=${encodeURIComponent(item.id)}`)}>Abrir montagem</Button>
      </article>; })}
    </section>}
    {!loading && !error && visibleItems.length === 0 && <Card className="assemblyCenterEmpty"><p>Nenhuma montagem ativa encontrada para os filtros selecionados.</p><Button variant="secondary" onClick={() => navigate("tender-list")}>Ver lista de editais</Button></Card>}
  </Page>;
}

function AssemblyCalendar({ navigate }) {
  const localMonth = () => {
    const today = new Date();
    return `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}`;
  };
  const [month, setMonth] = useState(localMonth);
  const [items, setItems] = useState([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    setLoading(true);
    setError("");
    fetch(`${API_BASE_URL}/api/assembly-calendar?month=${encodeURIComponent(month)}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json().catch(() => []);
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o calendário.");
        return data;
      })
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((loadError) => setError(loadError.message))
      .finally(() => setLoading(false));
  }, [month]);

  const calendarDate = new Date(`${month}-01T12:00:00`);
  const year = calendarDate.getFullYear();
  const monthIndex = calendarDate.getMonth();
  const monthTitle = calendarDate.toLocaleDateString("pt-BR", { month: "long", year: "numeric" }).replace(/^./, (letter) => letter.toUpperCase());
  const firstWeekday = (new Date(year, monthIndex, 1).getDay() + 6) % 7;
  const daysInMonth = new Date(year, monthIndex + 1, 0).getDate();
  const currentMonth = localMonth();
  const visibleItems = items.filter((item) => !filter || item.assemblyType === filter);
  const itemsByDay = visibleItems.reduce((map, item) => {
    const dateKey = String(item.openingDate || "").slice(0, 10);
    if (!dateKey) return map;
    if (!map[dateKey]) map[dateKey] = [];
    map[dateKey].push(item);
    return map;
  }, {});
  const changeMonth = (offset) => {
    const next = new Date(year, monthIndex + offset, 1);
    setMonth(`${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, "0")}`);
  };
  const progressTone = (value) => {
    const progress = Number(value || 0);
    if (progress > 70) return "progressGood";
    if (progress > 30) return "progressAttention";
    return "progressCritical";
  };
  const weekDays = ["Seg", "Ter", "Qua", "Qui", "Sex", "Sáb", "Dom"];
  const calendarCells = Array.from({ length: firstWeekday + daysInMonth }, (_, index) => index < firstWeekday ? null : index - firstWeekday + 1);

  return <Page label="Editais" title="Calendário de montagens" actions={<div className="calendarNavigation"><button type="button" className="iconButton secondaryIcon" title="Mês anterior" aria-label="Mês anterior" onClick={() => changeMonth(-1)}>‹</button><Button variant="secondary" onClick={() => setMonth(currentMonth)}>Hoje</Button><button type="button" className="iconButton secondaryIcon" title="Próximo mês" aria-label="Próximo mês" onClick={() => changeMonth(1)}>›</button></div>}>
    <section className="assemblyCalendarHero">
      <div><span className="eyebrow">Agenda operacional</span><h3>{monthTitle}</h3><p>Veja as licitações em montagem pela sua empresa, individualmente ou em consórcio, organizadas pela data da sessão.</p></div>
      <div className="assemblyCalendarLegend"><span><i className="progressCritical" />Até 30%</span><span><i className="progressAttention" />31% a 70%</span><span><i className="progressGood" />Acima de 70%</span></div>
    </section>
    <Card className="compactFilters assemblyCalendarFilters"><FormGrid><Field label="Exibir"><select value={filter} onChange={(event) => setFilter(event.target.value)}><option value="">Todas as montagens</option><option value="individual">Participação individual</option><option value="consortium">Consórcios</option></select></Field></FormGrid></Card>
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {loading ? <Card><p>Carregando calendário de montagens...</p></Card> : <section className="assemblyCalendarWrap"><div className="assemblyCalendarGrid"><div className="assemblyCalendarWeekdays">{weekDays.map((day) => <span key={day}>{day}</span>)}</div><div className="assemblyCalendarDays">{calendarCells.map((day, index) => {
      if (!day) return <div className="assemblyCalendarBlank" key={`blank-${index}`} />;
      const dateKey = `${month}-${String(day).padStart(2, "0")}`;
      const dayItems = itemsByDay[dateKey] || [];
      return <section className={`assemblyCalendarDay ${dateKey === currentLocalISODate() ? "isToday" : ""}`} key={dateKey}><header><strong>{day}</strong><span>{dayItems.length ? `${dayItems.length} montagem${dayItems.length === 1 ? "" : "ões"}` : ""}</span></header><div className="assemblyCalendarCards">{dayItems.map((item) => { const progress = Math.max(0, Math.min(100, Number(item.progress || 0))); return <button type="button" className="assemblyCalendarCard" key={item.id} onClick={() => navigate(`assembly-board?assemblyId=${encodeURIComponent(item.id)}`)} title="Abrir Central de Montagem"><small>{item.assemblyType === "individual" ? "Individual" : "Consórcio"}</small><strong>{item.tenderNumber} · {item.agency}</strong><span>{item.tenderObject}</span><div className="calendarProgress"><div><em className={progressTone(progress)}><i style={{ width: `${progress}%` }} /></em><b>{progress}%</b></div></div></button>; })}</div></section>;
    })}</div></div></section>}
    {!loading && !error && visibleItems.length === 0 && <Card className="assemblyCalendarEmpty"><p>Nenhuma licitação em montagem neste mês para os filtros selecionados.</p></Card>}
  </Page>;
}

function TenderList({ navigate, openTenderInterestCompanies }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [sort, setSort] = useState("opening_asc");
  const [filters, setFilters] = useState({ search: "", status: "", state: "", modality: "" });
  useEffect(() => { fetch(`${API_BASE_URL}/api/tenders`, { credentials: "include" }).then(async (r) => { const d = await r.json(); if (!r.ok) throw new Error(d.error); return d; }).then(setItems).catch((e) => setError(e.message)).finally(() => setLoading(false)); }, []);
  const date = (value) => value ? new Date(value).toLocaleDateString("pt-BR") : "-";
  const tenderStatusLabels = { draft: "Rascunho", published: "Publicado", under_review: "Em análise", suspended: "Suspenso", challenged: "Impugnado", occurred: "Ocorrido", closed: "Encerrado", cancelled: "Cancelado" };
  const orderedItems = items.filter((item) => {
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    const matchesSearch = !term || [item.agency, item.number, item.object, item.city, item.judgmentCriterion].some((value) => String(value || "").toLocaleLowerCase("pt-BR").includes(term));
    const matchesInterest = !filters.interest || (filters.interest === "registered" ? item.hasMyInterest : !item.hasMyInterest);
    return matchesSearch && (!filters.status || item.status === filters.status) && (!filters.state || item.state === filters.state) && (!filters.modality || item.modality === filters.modality) && matchesInterest;
  }).sort((a, b) => {
    if (sort === "value_desc") return Number(b.estimatedValue || 0) - Number(a.estimatedValue || 0);
    if (sort === "value_asc") return Number(a.estimatedValue || 0) - Number(b.estimatedValue || 0);
    if (sort === "agency_asc") return String(a.agency || "").localeCompare(String(b.agency || ""), "pt-BR");
    if (sort === "opening_desc") return new Date(b.openingDate || 0).getTime() - new Date(a.openingDate || 0).getTime();
    return new Date(a.openingDate || "9999-12-31").getTime() - new Date(b.openingDate || "9999-12-31").getTime();
  });
  return <Page label="Editais" title="Lista de editais"><Card className="compactFilters tenderListSortSticky"><FormGrid><Field label="Buscar edital"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Órgão, número, objeto ou cidade" /></Field><Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option><option value="published">Publicado</option><option value="under_review">Em análise</option><option value="suspended">Suspenso</option><option value="challenged">Impugnado</option><option value="occurred">Ocorrido</option><option value="closed">Encerrado</option><option value="cancelled">Cancelado</option></select></Field><Field label="Interesse"><select value={filters.interest} onChange={(event) => setFilters((current) => ({ ...current, interest: event.target.value }))}><option value="">Todos</option><option value="registered">Interesse registrado</option><option value="not_registered">Ainda não registrado</option></select></Field><Field label="Estado"><StateSelect value={filters.state} onChange={(event) => setFilters((current) => ({ ...current, state: event.target.value }))} /></Field><Field label="Modalidade"><select value={filters.modality} onChange={(event) => setFilters((current) => ({ ...current, modality: event.target.value }))}><option value="">Todas</option>{tenderModalityOptions.map((option) => <option value={option} key={option}>{option}</option>)}</select></Field><Field label="Ordenar por"><select value={sort} onChange={(event) => setSort(event.target.value)}><option value="opening_asc">Abertura mais próxima</option><option value="opening_desc">Abertura mais distante</option><option value="value_desc">Maior valor estimado</option><option value="value_asc">Menor valor estimado</option><option value="agency_asc">Órgão A-Z</option></select></Field></FormGrid></Card>{loading && <Card><p>Carregando editais...</p></Card>}{error && <Card className="dangerNotice"><p>{error}</p></Card>}{!loading && !error && orderedItems.length === 0 && <Card><p>Nenhum edital encontrado com esses filtros.</p></Card>}{!loading && !error && orderedItems.length > 0 && <Table columns={["Órgão", "Número", "Objeto", "Local", "Abertura", "Valor referencial", "Critério", "Análise", "Status", "Ações"]} rows={orderedItems.map((item) => [item.agency, item.number, compactObject(item.object), [item.city,item.state].filter(Boolean).join(" / ") || "-", date(item.openingDate), formatCurrencyBRFromNumber(item.estimatedValue) || "-", item.judgmentCriterion || "-", <span className={`statusPill ${item.hasAnalysis ? "success" : "review"}`} title={item.hasAnalysis ? "Este edital possui pré-análise HTML" : "Este edital ainda não possui pré-análise HTML"}>{item.hasAnalysis ? "Analisado" : "Pendente"}</span>, tenderStatusLabels[item.status] || item.status, <div className="rowActions compactRowActions" key={item.id}><button className="iconButton secondaryIcon" title="Ver detalhe do edital" onClick={() => navigate(`tender-detail?id=${item.id}`)}>{"\u25C9"}</button>{item.hasMyInterest ? <span className="statusPill open" title="Você já registrou interesse neste edital">Interesse registrado</span> : item.status === "published" ? <button className="iconButton successIcon" title="Registrar interesse" onClick={() => navigate(`tender-interest?id=${item.id}`)}>{"\u2713"}</button> : <span className="statusPill review" title={`Interesse indisponível: edital ${tenderStatusLabels[item.status] || item.status}`}>Indisponível</span>}<button className="iconButton partnerIcon" title="Ver empresas interessadas" onClick={() => openTenderInterestCompanies(item.id)}>{"\u2637"}</button></div>])} />}</Page>;
}

function TenderDetail({ navigate, sessionUser, openTenderInterestCompanies }) {
  const id = currentHashParams().get("id") || "";
  const isPlatformAdmin = sessionUser?.roleKey === "platform_admin";
  const [tender, setTender] = useState(null);
  const [error, setError] = useState("");
  const [analysisMessage, setAnalysisMessage] = useState(null);
  const [uploadingAnalysis, setUploadingAnalysis] = useState(false);
  const [generatingAIAnalysis, setGeneratingAIAnalysis] = useState(false);
  const [documentMessage, setDocumentMessage] = useState(null);
  const [uploadingDocuments, setUploadingDocuments] = useState(false);
  const [timelineOpen, setTimelineOpen] = useState(false);
  const loadTender = () => {
    if (!id) {
      setError("Selecione um edital na lista.");
      return;
    }
    fetch(`${API_BASE_URL}/api/tenders/${id}`, { credentials: "include" })
      .then(async (r) => { const d=await r.json(); if(!r.ok) throw new Error(d.error); return d; })
      .then(setTender)
      .catch((e)=>setError(e.message));
  };
  useEffect(() => { loadTender(); }, [id]);
  useEffect(() => {
    const status = tender?.aiAnalysis?.status;
    if (status !== "queued" && status !== "processing") return undefined;
    const timer = window.setTimeout(loadTender, 3500);
    return () => window.clearTimeout(timer);
  }, [tender?.aiAnalysis?.status, tender?.aiAnalysis?.id]);
  const generateAIAnalysis = async () => {
    setGeneratingAIAnalysis(true);
    setAnalysisMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}/ai-analysis`, {
        method: "POST",
        credentials: "include"
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível iniciar a pré-análise com IA.");
      setAnalysisMessage({ type: "success", text: "Pré-análise iniciada. O sistema continuará atualizando esta tela até o HTML ficar pronto." });
      loadTender();
    } catch (err) {
      setAnalysisMessage({ type: "error", text: err.message });
    } finally {
      setGeneratingAIAnalysis(false);
    }
  };
  const selectAnalysis = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith(".html") && !file.name.toLowerCase().endsWith(".htm")) {
      setAnalysisMessage({ type: "error", text: "Selecione um arquivo HTML." });
      return;
    }
    const reader = new FileReader();
    reader.onload = async () => {
      setUploadingAnalysis(true);
      setAnalysisMessage(null);
      try {
        const response = await fetch(`${API_BASE_URL}/api/tenders/${id}/analysis`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            analysisDataUrl: String(reader.result || ""),
            analysisFileName: file.name,
            analysisMimeType: "text/html"
          })
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(data.error || "Não foi possível anexar a análise.");
        setAnalysisMessage({ type: "success", text: "Análise anexada ao edital." });
        loadTender();
      } catch (err) {
        setAnalysisMessage({ type: "error", text: err.message });
      } finally {
        setUploadingAnalysis(false);
      }
    };
    reader.readAsDataURL(file);
  };
  const uploadDocuments = async (event) => {
    const files = Array.from(event.target.files || []);
    if (files.length === 0) return;
    const oversized = files.find((file) => file.size > 25 * 1024 * 1024);
    if (oversized) {
      setDocumentMessage({ type: "error", text: `O arquivo ${oversized.name} ultrapassa o limite de 25MB.` });
      event.target.value = "";
      return;
    }
    setUploadingDocuments(true);
    setDocumentMessage(null);
    const failures = [];
    for (const file of files) {
      try {
        const fileDataUrl = await new Promise((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = () => resolve(String(reader.result || ""));
          reader.onerror = () => reject(new Error("Não foi possível ler o arquivo."));
          reader.readAsDataURL(file);
        });
        const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}/files`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ fileDataUrl, fileName: file.name, mimeType: file.type })
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(data.error || "Falha ao anexar arquivo.");
      } catch (err) {
        failures.push(`${file.name}: ${err.message}`);
      }
    }
    event.target.value = "";
    setUploadingDocuments(false);
    if (failures.length) {
      setDocumentMessage({ type: "error", text: `${failures.length} arquivo(s) não foram anexados. ${failures[0]}` });
      return;
    }
    setDocumentMessage({ type: "success", text: `${files.length} arquivo(s) anexado(s) ao edital.` });
    loadTender();
  };
  const deleteDocument = async (document) => {
    if (!window.confirm(`Remover o arquivo ${document.title} deste edital?`)) return;
    setDocumentMessage(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}/files?fileId=${encodeURIComponent(document.id)}`, { method: "DELETE", credentials: "include" });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível remover o arquivo.");
      setDocumentMessage({ type: "success", text: "Arquivo removido da lista do edital." });
      loadTender();
    } catch (err) {
      setDocumentMessage({ type: "error", text: err.message });
    }
  };
  const withdrawInterest = async () => {
    if (!window.confirm("Desistir desta participação? O anúncio público será removido e uma montagem individual, se existir, será encerrada.")) return;
    try {
      const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}/withdraw-interest`, { method: "PATCH", credentials: "include" });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível desistir da participação.");
      navigate("tender-list");
    } catch (err) {
      setError(err.message);
    }
  };
  if (error) return <Page label="Editais" title="Detalhe do edital"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  if (!tender) return <Page label="Editais" title="Detalhe do edital"><Card><p>Carregando edital...</p></Card></Page>;
  const documents = Array.isArray(tender.documents) ? tender.documents : [];
  const aiAnalysis = tender.aiAnalysis || null;
  const aiIsWorking = aiAnalysis?.status === "queued" || aiAnalysis?.status === "processing";
  const aiProviderLabel = aiAnalysis?.provider === "gemini" ? "Google Gemini" : aiAnalysis?.provider === "openai" ? "OpenAI" : "IA";
  const aiProviderDetail = aiAnalysis?.model ? `${aiProviderLabel} · ${aiAnalysis.model}` : aiProviderLabel;
  const canChallengeTender = ["company_admin", "commercial", "technical"].includes(sessionUser?.roleKey) && ["published", "under_review", "challenged"].includes(tender.status);
  const formatDocumentSize = (bytes) => !bytes ? "Tamanho não informado" : bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / (1024 * 1024)).toFixed(1).replace(".", ",")} MB`;
  const timelineLabels = { created: "Edital cadastrado", captured: "Edital captado", published: "Edital publicado", resumed: "Edital retomado", suspended: "Edital suspenso", occurred: "Sessão ocorrida", closed: "Edital encerrado", cancelled: "Edital cancelado", under_review: "Edital em análise", challenged: "Edital impugnado", status_changed: "Status atualizado", interest_registered: "Interesse registrado", match: "Match realizado", consortium: "Consórcio formado", assembly: "Montagem iniciada" };
  const timeline = Array.isArray(tender.timeline) ? tender.timeline : [];
  const myInterest = tender.myInterest || null;
  const detailAction = tender.hasActiveConsortium
    ? <><Button onClick={() => navigate("match-list")}>{tender.canSearchNewConsortiumMember ? "Gerenciar busca do consórcio" : "Ver meu consórcio"}</Button><span className="statusPill review">Participação em consórcio ativa</span></>
    : !myInterest
    ? tender.status === "published"
      ? <Button onClick={() => navigate(`tender-interest?id=${tender.id}`)}>Registrar interesse</Button>
      : <span className="statusPill review">Interesse indisponível: edital {tender.status}</span>
    : myInterest.participationMode === "individual"
      ? <><Button onClick={() => navigate(`assembly-board?interestId=${encodeURIComponent(myInterest.id)}`)}>Abrir montagem individual</Button><Button variant="secondary" onClick={() => navigate(`tender-interest?id=${tender.id}`)}>Buscar parceiros</Button></>
      : <><Button onClick={() => openTenderInterestCompanies(tender.id)}>Ver empresas interessadas</Button>{myInterest.individualAssemblyId && <Button variant="secondary" onClick={() => navigate(`assembly-board?interestId=${encodeURIComponent(myInterest.id)}`)}>Abrir montagem individual</Button>}<Button variant="secondary" onClick={() => navigate(`tender-interest?id=${tender.id}`)}>Editar participação</Button></>;
  return (
    <Page label="Editais" title="Detalhe do edital" actions={<>{detailAction}{myInterest && !tender.hasActiveConsortium && <Button variant="secondary" onClick={withdrawInterest}>Desistir da participação</Button>}{canChallengeTender && <Button variant="secondary" onClick={() => navigate(`tender-challenge?id=${tender.id}`)}>{tender.myChallenge ? "Ver pedido de impugnação" : "Protocolar impugnação"}</Button>}</>}>
      <Card><h3>{tender.number} - {tender.agency}</h3><p>{tender.object}</p></Card>
      <div className="grid three">
        <Card><strong>Local</strong><p>{[tender.city,tender.state].filter(Boolean).join(" / ") || "Não informado"}</p></Card>
        <Card><strong>Critério</strong><p>{tender.judgmentCriterion || "Não informado"}</p></Card>
        <Card><strong>Status</strong><p>{tender.status}</p></Card>
      </div>
      <Card className="tenderTimelineCard">
        <div className="cardHeader"><div><h3>Linha do tempo</h3><p>Acontecimentos gerais e movimentações da sua empresa neste edital.</p></div><div className="timelineHeaderActions"><span className="statusPill open">{timeline.length} evento{timeline.length === 1 ? "" : "s"}</span><button type="button" className="iconButton secondaryIcon" title={timelineOpen ? "Recolher linha do tempo" : "Abrir linha do tempo"} aria-label={timelineOpen ? "Recolher linha do tempo" : "Abrir linha do tempo"} onClick={() => setTimelineOpen((open) => !open)}>{timelineOpen ? "⌃" : "⌄"}</button></div></div>
        {timelineOpen && (timeline.length === 0 ? <div className="emptyAnalysisBox"><strong>Histórico começará aqui</strong><p>Os próximos acontecimentos do edital aparecerão nesta linha do tempo.</p></div> : <ol className="tenderTimeline">{timeline.map((event) => <li key={event.id} className={`timelineEvent timeline-${event.eventType}`}><span className="timelineMarker" aria-hidden="true" /><div><div className="timelineEventHead"><strong>{timelineLabels[event.eventType] || event.title}</strong><time>{event.createdAt ? new Date(event.createdAt).toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" }) : ""}</time></div><p>{event.description || event.title}</p>{event.actorName && <small>Registrado por {event.actorName}</small>}</div></li>)}</ol>)}
      </Card>
      {tender.myChallenge && <Card className="challengeSummary"><div><strong>Pedido de impugnação da sua empresa</strong><p>{tender.myChallenge.subject}</p></div><span className={`statusPill ${tender.myChallenge.status}`}>{tender.myChallenge.status === "submitted" ? "Protocolado" : tender.myChallenge.status}</span></Card>}
      {tender.cloudFolderUrl && <Card><h3>Documentos do edital</h3><a className="downloadButton" href={tender.cloudFolderUrl} target="_blank" rel="noreferrer">Abrir diretório em nuvem</a></Card>}
      <Card className="tenderDocumentsCard">
        <div className="cardHeader"><div><h3>Arquivos do edital</h3><p>{documents.length ? `${documents.length} documento(s) disponível(is) para consulta e download.` : "Nenhum documento foi anexado diretamente ao edital."}</p></div></div>
        {documentMessage && <div className={`inlineFeedback ${documentMessage.type === "success" ? "successText" : "dangerText"}`}>{documentMessage.text}</div>}
        {documents.length > 0 && <div className="tenderDocumentList">{documents.map((document) => <div key={document.id}><span className="tenderDocumentIcon">▤</span><span className="tenderDocumentInfo"><strong>{document.title}</strong><small>{formatDocumentSize(Number(document.fileSize || 0))}{document.mimeType ? ` · ${document.mimeType}` : ""}</small></span><a className="iconButton secondaryIcon" href={document.fileUrl} download title="Baixar arquivo" aria-label={`Baixar ${document.title}`}>↓</a>{isPlatformAdmin && <button type="button" className="iconButton dangerIcon" title="Remover arquivo" aria-label={`Remover ${document.title}`} onClick={() => deleteDocument(document)}>×</button>}</div>)}</div>}
        {isPlatformAdmin && <Field label="Anexar documentos" hint="Você pode incluir quantos arquivos forem necessários. Limite de 25 MB por arquivo."><input type="file" multiple accept=".pdf,.doc,.docx,.xls,.xlsx,.csv,.ppt,.pptx,.odt,.ods,.txt,.xml,.zip,.rar,.7z,.dwg,.dxf,.jpg,.jpeg,.png,.webp" onChange={uploadDocuments} disabled={uploadingDocuments} /><small>{uploadingDocuments ? "Enviando documentos..." : "PDF, Office, planilhas, compactados, imagens e arquivos técnicos."}</small></Field>}
      </Card>
      <Card>
        <div className="cardHeader">
          <div><h3>Análise do edital</h3><p>Pré-análise técnica em HTML para leitura comercial.</p></div>
          {tender.analysisUrl && <a className="downloadButton" href={tender.analysisUrl} download>Baixar análise</a>}
        </div>
        {analysisMessage && <div className={`inlineFeedback ${analysisMessage.type === "success" ? "successText" : "dangerText"}`}>{analysisMessage.text}</div>}
        {isPlatformAdmin && aiAnalysis && <div className={`aiAnalysisStatus ${aiAnalysis.status}`}><strong>{aiAnalysis.status === "queued" ? "Pré-análise aguardando início" : aiAnalysis.status === "processing" ? "Pré-análise em processamento" : aiAnalysis.status === "completed" ? `Pré-análise gerada por ${aiProviderDetail}` : "Não foi possível gerar a pré-análise"}</strong><span>{aiAnalysis.status === "failed" ? aiAnalysis.errorMessage || "Tente novamente após conferir os documentos." : aiAnalysis.status === "completed" ? "O HTML gerado está disponível abaixo e pode ser substituído manualmente." : `Analisando ${aiAnalysis.sourceFileCount || documents.length} documento(s) anexado(s) com ${aiProviderLabel}.`}</span></div>}
        {tender.analysisUrl ? (
          <iframe className="technicalSheetFrame" sandbox="allow-scripts" src={tender.analysisUrl} title={`Análise do edital ${tender.number}`}></iframe>
        ) : (
          <div className="emptyAnalysisBox">
            <strong>Edital ainda não analisado</strong>
            <p>A pré-análise em HTML ainda não foi anexada pelo administrador da plataforma. Quando o arquivo for enviado, ele aparecerá aqui aberto para leitura e download.</p>
          </div>
        )}
        {isPlatformAdmin && (
          <div className="analysisAdminActions">
            <Button onClick={generateAIAnalysis} disabled={Boolean(tender.analysisUrl) || !documents.length || generatingAIAnalysis || aiIsWorking} title={tender.analysisUrl ? "Este edital já possui uma pré-análise HTML" : "Gerar pré-análise com IA"}>{tender.analysisUrl ? "Pré-análise já disponível" : generatingAIAnalysis ? "Iniciando IA..." : aiIsWorking ? "IA analisando documentos..." : "Gerar pré-análise com IA"}</Button>
            <small>{documents.length ? "Usa os documentos anexados e o roteiro técnico oficial. A OpenAI é consultada primeiro e o Gemini assume automaticamente se ela falhar. Cada arquivo deve ter até 22 MB." : "Anexe documentos ao edital antes de solicitar a análise."}</small>
            <Field label={tender.analysisUrl ? "Substituir análise HTML manualmente" : "Anexar análise HTML manualmente"}>
              <input type="file" accept=".html,.htm,text/html" onChange={selectAnalysis} disabled={uploadingAnalysis || aiIsWorking} />
              <small>{uploadingAnalysis ? "Enviando análise..." : "A análise manual continua disponível mesmo com a IA."}</small>
            </Field>
          </div>
        )}
      </Card>
    </Page>
  );
}

function TenderChallenge({ navigate }) {
  const id = currentHashParams().get("id") || "";
  const [tender, setTender] = useState(null);
  const [existing, setExisting] = useState(null);
  const [form, setForm] = useState({ subject: "", rationale: "" });
  const [documents, setDocuments] = useState([]);
  const [message, setMessage] = useState(null);
  const [saving, setSaving] = useState(false);
  const load = () => {
    if (!id) return;
    fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}`, { credentials: "include" })
      .then(async (response) => { const data = await response.json(); if (!response.ok) throw new Error(data.error || "Não foi possível carregar o edital."); return data; })
      .then((data) => {
        setTender(data);
        if (data.myChallenge) {
          setExisting(data.myChallenge);
          setForm({ subject: data.myChallenge.subject || "", rationale: data.myChallenge.rationale || "" });
        }
      })
      .catch((error) => setMessage({ type: "error", text: error.message }));
  };
  useEffect(() => { load(); }, [id]);
  const selectDocuments = (event) => {
    const selected = Array.from(event.target.files || []);
    const oversized = selected.find((file) => file.size > 25 * 1024 * 1024);
    if (oversized) {
      setMessage({ type: "error", text: `O arquivo ${oversized.name} ultrapassa o limite de 25 MB.` });
      event.target.value = "";
      return;
    }
    setDocuments((current) => [...current, ...selected.filter((file) => !current.some((item) => item.name === file.name && item.size === file.size && item.lastModified === file.lastModified))]);
    event.target.value = "";
  };
  const submit = async () => {
    setSaving(true);
    setMessage(null);
    try {
      const encodedDocuments = await Promise.all(documents.map((file) => new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve({ fileDataUrl: String(reader.result || ""), fileName: file.name, mimeType: file.type });
        reader.onerror = () => reject(new Error(`Não foi possível ler ${file.name}.`));
        reader.readAsDataURL(file);
      })));
      const response = await fetch(`${API_BASE_URL}/api/tenders/${encodeURIComponent(id)}/challenge`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ ...form, documents: encodedDocuments })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível protocolar a impugnação.");
      setExisting(data);
      setDocuments([]);
      setMessage({ type: "success", text: "Pedido de impugnação protocolado. O administrador da plataforma foi avisado." });
    } catch (error) {
      setMessage({ type: "error", text: error.message });
    } finally {
      setSaving(false);
    }
  };
  const locked = existing && !["draft", "submitted"].includes(existing.status);
  const businessDaysBeforeSession = useMemo(() => {
    if (!tender?.openingDate) return null;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const opening = new Date(`${String(tender.openingDate).slice(0, 10)}T12:00:00`);
    if (Number.isNaN(opening.getTime())) return null;
    let count = 0;
    const cursor = new Date(today);
    cursor.setDate(cursor.getDate() + 1);
    while (cursor < opening) {
      const weekday = cursor.getDay();
      if (weekday !== 0 && weekday !== 6) count += 1;
      cursor.setDate(cursor.getDate() + 1);
    }
    return count;
  }, [tender?.openingDate]);
  const isUntimely = businessDaysBeforeSession !== null && businessDaysBeforeSession < 3;
  const formatFileSize = (value) => !value ? "Tamanho não informado" : value < 1024 * 1024 ? `${Math.max(1, Math.round(value / 1024))} KB` : `${(value / (1024 * 1024)).toFixed(1).replace(".", ",")} MB`;
  if (!id) return <Page label="Editais" title="Pedido de impugnação"><Card className="dangerNotice"><p>Abra esta tela pelo detalhe de um edital.</p></Card></Page>;
  if (!tender) return <Page label="Editais" title="Pedido de impugnação"><Card><p>Carregando edital...</p></Card></Page>;
  return <Page label="Editais" title="Pedido de impugnação" actions={<Button variant="secondary" onClick={() => navigate(`tender-detail?id=${id}`)}>Voltar ao edital</Button>}>
    <Card><h3>{tender.number} - {tender.agency}</h3><p>{tender.object}</p></Card>
    {message && <Card className={message.type === "success" ? "formFeedback success" : "dangerNotice"}><p>{message.text}</p></Card>}
    {isUntimely && <Card className="challengeLegalWarning"><strong>Licitação intempestiva</strong><p>Art. 164. Qualquer pessoa é parte legítima para impugnar edital de licitação por irregularidade na aplicação desta Lei ou para solicitar esclarecimento sobre os seus termos, devendo protocolar o pedido até 3 (três) dias úteis antes da data de abertura do certame.</p><small>Faltam {businessDaysBeforeSession} dia{businessDaysBeforeSession === 1 ? " útil" : "s úteis"} antes da sessão. O protocolo continuará disponível para registro.</small></Card>}
    {existing && <Card className="challengeSummary"><div><strong>Protocolo da sua empresa</strong><p>Status: {existing.status === "submitted" ? "Protocolado e aguardando análise" : existing.status}</p></div></Card>}
    <Card>
      <FormGrid>
        <Field label="Assunto do pedido"><input value={form.subject} maxLength="220" disabled={locked} placeholder="Ex.: Questionamento sobre exigência de habilitação técnica" onChange={(event) => setForm((current) => ({ ...current, subject: event.target.value }))} /></Field>
      </FormGrid>
      <Field label="Fundamentação da impugnação" hint="Explique objetivamente o item questionado, os motivos e a referência ao edital."><textarea rows="10" value={form.rationale} disabled={locked} placeholder="Descreva o entendimento da sua empresa e indique itens, cláusulas, páginas ou anexos aplicáveis." onChange={(event) => setForm((current) => ({ ...current, rationale: event.target.value }))} /></Field>
      {!locked && <Field label="Documentos de apoio" hint="Opcional. Você pode anexar quantos documentos forem necessários, com até 25 MB por arquivo."><input type="file" multiple accept=".pdf,.doc,.docx,.xls,.xlsx,.csv,.ppt,.pptx,.odt,.ods,.txt,.xml,.zip,.rar,.7z,.jpg,.jpeg,.png,.webp" onChange={selectDocuments} /><small>Documentos jurídicos, técnicos, planilhas, imagens e arquivos de apoio.</small></Field>}
      {documents.length > 0 && <div className="tenderDocumentQueue">{documents.map((file, index) => <div key={`${file.name}-${file.lastModified}`}><span className="tenderDocumentIcon">▤</span><span><strong>{file.name}</strong><small>{formatFileSize(file.size)}</small></span><button type="button" onClick={() => setDocuments((current) => current.filter((_, itemIndex) => itemIndex !== index))} title="Remover arquivo">×</button></div>)}</div>}
      {Array.isArray(existing?.documents) && existing.documents.length > 0 && <div className="tenderDocumentList">{existing.documents.map((file) => <div key={file.id}><span className="tenderDocumentIcon">▤</span><span className="tenderDocumentInfo"><strong>{file.title}</strong><small>{formatFileSize(Number(file.fileSize || 0))}</small></span><a className="iconButton secondaryIcon" href={file.fileUrl} download title="Baixar anexo">↓</a></div>)}</div>}
      {!locked && <div className="formActions"><Button onClick={submit} disabled={saving}>{saving ? "Protocolando..." : existing ? "Atualizar pedido" : "Protocolar pedido de impugnação"}</Button></div>}
    </Card>
  </Page>;
}

const tenderChallengeStatuses = [
  ["submitted", "Protocolado"],
  ["under_review", "Em análise"],
  ["accepted", "Procedente"],
  ["rejected", "Improcedente"],
  ["withdrawn", "Retirado"]
];

const tenderChallengeStatusLabel = (value) => tenderChallengeStatuses.find(([key]) => key === value)?.[1] || value;

function TenderChallengeBoard({ navigate }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", status: "", sort: "deadline_asc" });
  const [savingId, setSavingId] = useState("");
  const [selectedStatuses, setSelectedStatuses] = useState({});

  const load = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/tender-challenges`, { credentials: "include" })
      .then(async (response) => { const data = await response.json(); if (!response.ok) throw new Error(data.error || "Não foi possível carregar as impugnações."); return data; })
      .then((data) => { setItems(Array.isArray(data) ? data : []); setError(""); })
      .catch((loadError) => setError(loadError.message))
      .finally(() => setLoading(false));
  };
  useEffect(() => { load(); }, []);

  const updateStatus = async (item, status) => {
    setSavingId(item.id);
    try {
      const response = await fetch(`${API_BASE_URL}/api/tender-challenges/${encodeURIComponent(item.id)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ status })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível atualizar o pedido.");
      setItems((current) => current.map((entry) => entry.id === item.id ? { ...entry, status } : entry));
	  setSelectedStatuses((current) => { const next = { ...current }; delete next[item.id]; return next; });
    } catch (updateError) {
      setError(updateError.message);
    } finally {
      setSavingId("");
    }
  };

  const visible = useMemo(() => {
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    return items.filter((item) => {
      if (filters.status && item.status !== filters.status) return false;
      return !term || [item.tenderNumber, item.tenderAgency, item.tenderObject, item.companyName, item.authorName, item.subject].join(" ").toLocaleLowerCase("pt-BR").includes(term);
    }).sort((a, b) => {
      if (filters.sort === "submitted_desc") return new Date(b.submittedAt || 0).getTime() - new Date(a.submittedAt || 0).getTime();
      if (filters.sort === "opening_asc") return new Date(a.openingDate || 0).getTime() - new Date(b.openingDate || 0).getTime();
      return new Date(a.internalDeadline || 0).getTime() - new Date(b.internalDeadline || 0).getTime();
    });
  }, [items, filters]);
  const columns = [
    ["submitted", "Protocolados", (item) => item.status === "submitted"],
    ["under_review", "Em análise", (item) => item.status === "under_review"],
    ["completed", "Concluídos", (item) => ["accepted", "rejected", "withdrawn"].includes(item.status)]
  ];
  const formatDate = (value) => value ? new Date(value).toLocaleDateString("pt-BR") : "Não informado";
  const whatsappURL = (phone) => {
    const digits = String(phone || "").replace(/\D/g, "");
    if (!digits) return "";
    return `https://wa.me/${digits.startsWith("55") ? digits : `55${digits}`}`;
  };

  return <Page label="Editais" title="Central de impugnações">
    <Card className="challengeBoardIntro"><div><strong>Pedidos protocolados pelas empresas</strong><p>Cada cartão representa uma tarefa de análise. O prazo interno é calculado para seis dias antes da sessão do edital.</p></div><span>{items.length} pedido{items.length === 1 ? "" : "s"}</span></Card>
    <Card className="compactFilters challengeBoardFilters"><FormGrid>
      <Field label="Buscar"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Empresa, edital, órgão ou assunto" /></Field>
      <Field label="Status"><select value={filters.status} onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}><option value="">Todos</option>{tenderChallengeStatuses.map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
      <Field label="Ordenar por"><select value={filters.sort} onChange={(event) => setFilters((current) => ({ ...current, sort: event.target.value }))}><option value="deadline_asc">Prazo interno mais próximo</option><option value="opening_asc">Sessão mais próxima</option><option value="submitted_desc">Protocolo mais recente</option></select></Field>
    </FormGrid></Card>
    {loading && <Card><p>Carregando pedidos de impugnação...</p></Card>}
    {error && <Card className="dangerNotice"><p>{error}</p></Card>}
    {!loading && <div className="challengeKanban">{columns.map(([key, label, matches]) => {
      const columnItems = visible.filter(matches);
      return <section className={`challengeKanbanColumn ${key}`} key={key}><header><h3>{label}</h3><span>{columnItems.length}</span></header><div className="challengeKanbanCards">{columnItems.map((item) => <article className={`challengeKanbanCard ${item.isUntimely ? "is-untimely" : ""}`} key={item.id}>
        <div className="challengeCardTop"><strong>{item.tenderNumber}</strong><span className={`statusPill ${item.status === "submitted" ? "review" : "open"}`}>{tenderChallengeStatusLabel(item.status)}</span></div>
        <p className="challengeTender">{item.tenderAgency}</p><h4>{item.subject}</h4><div className="challengeCompanyLine"><p className="challengeCompany">{item.companyName} {item.authorName ? `· ${item.authorName}` : ""}</p>{whatsappURL(item.requesterPhone) && <a className="iconButton whatsappIcon" href={whatsappURL(item.requesterPhone)} target="_blank" rel="noreferrer" title={`Conversar com ${item.authorName || "o solicitante"} no WhatsApp`} aria-label={`Conversar com ${item.authorName || "o solicitante"} no WhatsApp`}>☎</a>}</div>
        <div className="challengeDates"><span>Sessão: <b>{formatDate(item.openingDate)}</b></span><span>Prazo interno: <b>{formatDate(item.internalDeadline)}</b></span></div>
        {item.isUntimely && <p className="challengeLate">Intempestiva: {item.businessDaysBeforeOpening ?? 0} dia(s) útil(eis) antes da sessão.</p>}
        <details><summary>Ver fundamentação e documentos ({item.documentCount || 0})</summary><p>{item.rationale}</p>{Array.isArray(item.documents) && item.documents.length > 0 && <div className="challengeFiles">{item.documents.map((file) => <a href={file.fileUrl} download key={file.id}>↓ {file.title}</a>)}</div>}<Button variant="secondary" onClick={() => navigate(`tender-detail?id=${item.tenderId}`)}>Abrir edital</Button></details>
        <div className="challengeStatusActions"><Field label="Andamento"><select value={selectedStatuses[item.id] ?? item.status} disabled={savingId === item.id} onChange={(event) => setSelectedStatuses((current) => ({ ...current, [item.id]: event.target.value }))}>{tenderChallengeStatuses.map(([value, statusLabel]) => <option value={value} key={value}>{statusLabel}</option>)}</select></Field><Button onClick={() => updateStatus(item, selectedStatuses[item.id] ?? item.status)} disabled={savingId === item.id || (selectedStatuses[item.id] ?? item.status) === item.status}>{savingId === item.id ? "Atualizando..." : "Atualizar andamento"}</Button></div>
      </article>)}{columnItems.length === 0 && <p className="personalKanbanEmpty">Nenhum pedido nesta fase.</p>}</div></section>;
    })}</div>}
  </Page>;
}

const interestRequirementBlocks = [
  {
    key: "operational_qualification",
    title: "Requisito operacional",
    description: "Acervo, atestados, experiência da empresa e pontuação operacional exigidos pelo edital.",
    options: ["Atendo integralmente", "Atendo parcialmente", "Não atendo", "Atendimento de baixa pontuação", "Não se aplica"],
    offerPlaceholder: "Descreva o que sua empresa possui neste requisito",
    needPlaceholder: "Descreva o que sua empresa busca complementar"
  },
  {
    key: "professional_qualification",
    title: "Requisitos profissionais",
    description: "Equipe, responsáveis técnicos, currículos, CATs, disponibilidade e pontuação profissional.",
    options: ["Tenho equipe completa", "Tenho equipe parcial", "Não tenho equipe", "Possuo equipe com baixa pontuação", "Não se aplica"],
    offerPlaceholder: "Descreva a equipe e os profissionais disponíveis",
    needPlaceholder: "Descreva quais profissionais ou experiências busca"
  },
  {
    key: "technical_proposal",
    title: "Peça técnica qualitativa",
    description: "Metodologia, plano de trabalho, abordagem técnica e proposta qualitativa.",
    options: ["Tenho capacidade interna de montagem", "Tenho capacidade parcial de montagem", "Componho para contratação de apoio especializado", "Não possuo capacidade para essa peça", "Não se aplica"],
    offerPlaceholder: "Descreva sua capacidade para elaborar a peça técnica",
    needPlaceholder: "Descreva qual apoio técnico deseja contratar ou compor"
  },
  {
    key: "certifications",
    title: "Certificações requeridas",
    description: "Certificações, registros ou comprovações formais exigidas.",
    options: ["Possuo todas", "Possuo parcialmente", "Não possuo", "Não se aplica", "Em análise"],
    offerPlaceholder: "Descreva as certificações, registros ou comprovações que possui",
    needPlaceholder: "Descreva certificações ou registros que precisa complementar"
  }
];

const interestStatusValues = {
  "Atendo integralmente": "fully_meets",
  "Atende integralmente": "fully_meets",
  "Tenho equipe completa": "fully_meets",
  "Tenho capacidade interna": "fully_meets",
  "Tenho capacidade interna de montagem": "fully_meets",
  "Possuo todas": "fully_meets",
  "Atendo parcialmente": "partially_meets",
  "Tenho equipe parcial": "partially_meets",
  "Tenho capacidade parcial": "partially_meets",
  "Tenho capacidade parcial de montagem": "partially_meets",
  "Possuo parcialmente": "partially_meets",
  "Não atendo": "does_not_meet",
  "Não tenho equipe": "does_not_meet",
  "Não possuo capacidade para essa peça": "does_not_meet",
  "Não possuo": "does_not_meet",
  "Atendimento de baixa pontuação": "low_score",
  "Possuo equipe com baixa pontuação": "low_score",
  "Posso compor com parceiro": "seeks_partner",
  "Posso montar equipe": "seeks_partner",
  "Componho para contratação de apoio especializado": "seeks_partner",
  "Posso obter/regularizar": "seeks_partner",
  "Não se aplica": "not_applicable",
  "Em análise": "under_review"
};

const interestStatusLabels = {
  fully_meets: "Atende integralmente",
  partially_meets: "Atende parcialmente",
  does_not_meet: "Não atende",
  low_score: "Baixa pontuação",
  seeks_partner: "Busca composição",
  not_applicable: "Não se aplica",
  under_review: "Em análise"
};

function TenderInterest({ navigate }) {
  const id = currentHashParams().get("id") || "";
  const [tender, setTender] = useState(null);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [form, setForm] = useState(() => ({
    participationMode: "seeking_partners",
    generalPosition: "interested",
    desiredRole: "seeks_partner",
    publicSummary: "",
    internalNote: "",
    requirements: interestRequirementBlocks.map((block) => ({
      requirementKey: block.key,
      statusKey: interestStatusValues[block.options[0]],
      whatWeHave: "",
      whatWeSeek: ""
    }))
  }));

  useEffect(() => {
    if (!id) {
      setError("Selecione um edital na lista antes de registrar interesse.");
      return;
    }
    fetch(`${API_BASE_URL}/api/tenders/${id}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o edital.");
        return data;
      })
      .then((data) => {
        setTender(data);
        if (data.myInterest) {
          const savedRequirements = new Map((data.myInterest.requirements || []).map((item) => [item.requirementKey, item]));
          setForm((current) => ({
            ...current,
            participationMode: data.myInterest.participationMode || current.participationMode,
            generalPosition: data.myInterest.generalPosition || current.generalPosition,
            desiredRole: data.myInterest.desiredRole || current.desiredRole,
            publicSummary: data.myInterest.publicSummary || "",
            internalNote: data.myInterest.internalNote || "",
            requirements: current.requirements.map((item) => ({ ...item, ...(savedRequirements.get(item.requirementKey) || {}) }))
          }));
        }
      })
      .catch((err) => setError(err.message));
  }, [id]);

  const updateRequirement = (index, field, value) => {
    setForm((current) => ({
      ...current,
      requirements: current.requirements.map((item, itemIndex) => itemIndex === index ? { ...item, [field]: value } : item)
    }));
  };

  const submit = async () => {
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/tenders/${id}/interests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form)
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível registrar o interesse.");
      if (data.participationMode === "individual" && data.interestId) {
        navigate(`assembly-board?interestId=${encodeURIComponent(data.interestId)}`);
      } else if (data.participationMode === "seeking_partners") {
        navigate(`tender-interest-list?id=${id}`);
      } else {
        navigate(`tender-detail?id=${id}`);
      }
    } catch (err) {
      setMessage(err.message);
    } finally {
      setSaving(false);
    }
  };

  if (error) return <Page label="Editais" title="Manifestação de interesse"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  if (!tender) return <Page label="Editais" title="Manifestação de interesse"><Card><p>Carregando edital...</p></Card></Page>;

  return (
    <Page label="Editais" title="Manifestação de interesse" actions={<Button onClick={submit} disabled={saving}>{saving ? "Salvando..." : form.participationMode === "individual" ? "Salvar e abrir montagem" : form.participationMode === "seeking_partners" ? "Salvar e ver empresas interessadas" : "Salvar participação"}</Button>}>
      {message && <Card className="dangerNotice"><p>{message}</p></Card>}
      <section className="interestHero">
        <div>
          <span className="badge">{tender.number}</span>
          <h3>Defina a estratégia da sua empresa nesta licitação</h3>
          <p>Registre o que sua empresa atende e escolha se trabalhará sozinha, buscará parceiras ou apenas acompanhará a oportunidade.</p>
        </div>
        <div className="interestHeroPanel">
          <strong>Resultado desta tela</strong>
          <span>{form.participationMode === "individual" ? "Montagem privada da sua empresa" : form.participationMode === "seeking_partners" ? "Anúncio visível na vitrine da licitação" : "Registro privado para acompanhamento"}</span>
        </div>
      </section>

      <div className="interestQuickGrid">
        <Field label="Estratégia de participação">
          <select value={form.participationMode} onChange={(event) => setForm((current) => ({ ...current, participationMode: event.target.value }))}>
            <option value="individual">Vou participar sozinha</option>
            <option value="seeking_partners">Quero buscar parceiros</option>
            <option value="undecided">Ainda estou avaliando</option>
          </select>
        </Field>
        <Field label="Posição geral">
          <select value={form.generalPosition} onChange={(event) => setForm((current) => ({ ...current, generalPosition: event.target.value }))}>
            <option value="interested">Tenho interesse em participar</option>
            <option value="under_evaluation">Estou avaliando participação</option>
            <option value="watching">Quero apenas acompanhar por enquanto</option>
            <option value="not_interested">Não tenho interesse nesta licitação</option>
          </select>
        </Field>
        {form.participationMode === "seeking_partners" && <Field label="Papel desejado">
          <select value={form.desiredRole} onChange={(event) => setForm((current) => ({ ...current, desiredRole: event.target.value }))}>
            <option value="seeks_partner">Busco parceiro para complementar requisitos</option>
            <option value="can_lead_consortium">Posso liderar consórcio</option>
            <option value="complementary_partner">Quero participar como parceira complementar</option>
            <option value="seeks_lead_company">Busco empresa líder de proposta</option>
            <option value="evaluating_role">Ainda estou avaliando meu papel</option>
          </select>
        </Field>}
      </div>

      <div className="interestRequirementGrid">
        {interestRequirementBlocks.map((block, blockIndex) => (
          <Card className="interestRequirementCard" key={block.title}>
            <div className="interestRequirementHeader">
              <div>
                <h3>{block.title}</h3>
                <p>{block.description}</p>
              </div>
            </div>
            <div className="interestChoiceRail" aria-label={`Minha situação em ${block.title}`}>
              {block.options.map((option, optionIndex) => (
                <label className="interestChoice" key={option}>
                  <input
                    name={`interest-${blockIndex}`}
                    type="radio"
                    checked={form.requirements[blockIndex]?.statusKey === interestStatusValues[option]}
                    onChange={() => updateRequirement(blockIndex, "statusKey", interestStatusValues[option])}
                  />
                  <span>{option}</span>
                </label>
              ))}
            </div>
            <div className="interestDetailGrid">
              <Field label="O que tenho">
                <textarea value={form.requirements[blockIndex]?.whatWeHave || ""} onChange={(event) => updateRequirement(blockIndex, "whatWeHave", event.target.value)} placeholder={block.offerPlaceholder} />
              </Field>
              <Field label="O que busco">
                <textarea value={form.requirements[blockIndex]?.whatWeSeek || ""} onChange={(event) => updateRequirement(blockIndex, "whatWeSeek", event.target.value)} placeholder={block.needPlaceholder} />
              </Field>
            </div>
          </Card>
        ))}
      </div>

      <div className="interestSummaryGrid">
        {form.participationMode !== "individual" && <Field label={form.participationMode === "seeking_partners" ? "Resumo do anúncio para possíveis parceiros" : "Resumo da participação"}>
          <textarea value={form.publicSummary} onChange={(event) => setForm((current) => ({ ...current, publicSummary: event.target.value }))} placeholder={form.participationMode === "seeking_partners" ? "Escreva um resumo claro do anúncio que será visto por possíveis parceiros." : "Registre um resumo objetivo da estratégia de participação da empresa."} />
        </Field>}
        <Field label="Observação interna">
          <textarea value={form.internalNote} onChange={(event) => setForm((current) => ({ ...current, internalNote: event.target.value }))} placeholder="Anotação privada da sua empresa, não exibida para outros participantes." />
        </Field>
      </div>
    </Page>
  );
}

function TenderInterestList({ navigate, selectedTenderId = "cp-004-2026", sessionUser, openChatForAd }) {
  const id = currentHashParams().get("id") || selectedTenderId;
  const [selectedTender, setSelectedTender] = useState(null);
  const [ads, setAds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [sort, setSort] = useState("published_desc");

  useEffect(() => {
    if (!id) {
      setError("Selecione um edital para ver as empresas interessadas.");
      setLoading(false);
      return;
    }
    setLoading(true);
    Promise.all([
      fetch(`${API_BASE_URL}/api/tenders/${id}`, { credentials: "include" }).then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o edital.");
        return data;
      }),
      fetch(`${API_BASE_URL}/api/tenders/${id}/interests`, { credentials: "include" }).then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar empresas interessadas.");
        return data;
      })
    ])
      .then(([tenderData, adData]) => {
        setSelectedTender(tenderData);
        setAds(adData);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <Page label="Editais" title="Empresas interessadas no edital"><Card><p>Carregando empresas interessadas...</p></Card></Page>;
  if (error) return <Page label="Editais" title="Empresas interessadas no edital"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  const sortedAds = [...ads].sort((a, b) => {
    if (sort === "company_asc") return String(a.companyName || a.leaderCompanyName || "").localeCompare(String(b.companyName || b.leaderCompanyName || ""), "pt-BR");
    return new Date(b.publishedAt || 0).getTime() - new Date(a.publishedAt || 0).getTime();
  });

  return (
    <Page label="Editais" title="Empresas interessadas no edital" actions={<Button onClick={() => navigate(`match-partners`)}>Ver vitrine geral</Button>}>
      <Card className="notice">
        <strong>{selectedTender.number} - {selectedTender.object}</strong>
        <p>Empresas abaixo também manifestaram interesse e aceitaram aparecer para avaliação de possíveis parceiros nesta licitação.</p>
      </Card>
      <Table columns={["Edital selecionado", "Órgão", "Modalidade", "Local", "Abertura", "Critério", "Status"]} rows={[
        [selectedTender.number, selectedTender.agency, selectedTender.modality || "-", [selectedTender.city, selectedTender.state].filter(Boolean).join(" / ") || "-", selectedTender.openingDate ? new Date(selectedTender.openingDate).toLocaleDateString("pt-BR") : "-", selectedTender.judgmentCriterion || "-", <span className={`statusPill ${selectedTender.status === "published" ? "open" : "review"}`} key={`${selectedTender.id}-selected-status`}>{selectedTender.status}</span>]
      ]} />
      <Card className="compactFilters tenderInterestSortSticky"><FormGrid><Field label="Ordenar empresas por"><select value={sort} onChange={(event) => setSort(event.target.value)}><option value="published_desc">Interesse mais recente</option><option value="company_asc">Empresa A-Z</option></select></Field></FormGrid></Card>
      {ads.length === 0 ? (
        <Card><p>Ainda não há anúncios publicados para este edital.</p></Card>
      ) : (
        <Table columns={["Empresa", "Operacional", "Profissional", "Peça técnica", "Certificações", "Busca", "Ações"]} rows={sortedAds.map((ad) => {
          const requirement = (key) => ad.requirements?.find((item) => item.requirementKey === key);
          const isConsortiumMember = (ad.consortiumMembers || []).some((company) => company.companyId === sessionUser?.companyId);
          const isOwnAd = (ad.companyId && sessionUser?.companyId && ad.companyId === sessionUser.companyId) || isConsortiumMember;
          const alreadyEvaluated = Boolean(ad.currentEvaluationDecision || ad.myEvaluationDecision);
          return [
            <div className="companyNameStack" key={`${ad.id}-company`}>
              <strong>{ad.adType === "consortium" ? "Consórcio em formação" : ad.companyName}</strong>
              {ad.adType === "consortium" && <span className="badge">Busca consorciada</span>}
              {isOwnAd && <span className="badge">Meu anúncio</span>}
              {ad.adType === "consortium" && (ad.consortiumMembers || []).length > 0 && (
                <div className="consortiumMembersInline">
                  <small>Empresas já consorciadas</small>
                  <span>{ad.consortiumMembers.map((company) => company.companyName).join(" • ")}</span>
                </div>
              )}
            </div>,
            interestStatusLabels[requirement("operational_qualification")?.statusKey] || "-",
            interestStatusLabels[requirement("professional_qualification")?.statusKey] || "-",
            interestStatusLabels[requirement("technical_proposal")?.statusKey] || "-",
            interestStatusLabels[requirement("certifications")?.statusKey] || "-",
            ad.seekSummary || "-",
            <div className="rowActions compactRowActions" key={ad.id}>
              {ad.companyId && <button className="iconButton secondaryIcon" title="Abrir perfil público da empresa" aria-label="Abrir perfil público da empresa" onClick={() => navigate(`company-public-profile?companyId=${encodeURIComponent(ad.companyId)}`)}>{"\u2197"}</button>}
              <button className="iconButton secondaryIcon" title="Ver detalhe do anúncio" aria-label="Ver detalhe do anúncio" onClick={() => navigate(`match-profile?id=${ad.id}`)}>{"\u25C9"}</button>
              {!isOwnAd && <button className="iconButton chatIcon" title="Conversar com a empresa" aria-label="Conversar com a empresa" onClick={() => openChatForAd(ad)}>{"\u2709"}</button>}
              {!isOwnAd && (alreadyEvaluated
                ? <span className="iconButton successIcon isStatic" title="Avaliação já registrada" aria-label="Avaliação já registrada">{"\u2713"}</span>
                : <button className="iconButton successIcon" title="Avaliar candidata" aria-label="Avaliar candidata" onClick={() => navigate(`match-tinder?id=${ad.id}`)}>{"\u2713"}</button>
              )}
            </div>
          ];
        })} />
      )}
    </Page>
  );
}

function MatchPartners({ navigate, sessionUser, openChatForAd }) {
  const [ads, setAds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", agency: "Todos", criterion: "Todos", type: "Todos", need: "Todas", sort: "published_desc" });
  const [myAdsOpen, setMyAdsOpen] = useState(false);
  const [editingAdId, setEditingAdId] = useState("");
  const [adForm, setAdForm] = useState({ offerSummary: "", seekSummary: "" });
  const [savingAdId, setSavingAdId] = useState("");

  const loadAds = () => {
    setLoading(true);
    setError("");
    return fetch(`${API_BASE_URL}/api/partnership-ads`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar a vitrine.");
        return data;
      })
      .then(setAds)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadAds();
  }, []);

  const agencies = ["Todos", ...Array.from(new Set(ads.map((ad) => ad.agency).filter(Boolean)))];
  const criteria = ["Todos", ...Array.from(new Set(ads.map((ad) => ad.judgmentCriterion).filter(Boolean)))];
  const isMyAd = (ad) => ad.companyId === sessionUser?.companyId || (ad.consortiumMembers || []).some((company) => company.companyId === sessionUser?.companyId);
  const myAds = ads.filter(isMyAd);
  const filteredAds = ads.filter((ad) => !isMyAd(ad)).filter((ad) => {
    const members = (ad.consortiumMembers || []).map((company) => company.companyName).join(" ");
    const haystack = `${ad.companyName} ${ad.leaderCompanyName} ${members} ${ad.tenderNumber} ${ad.tenderObject} ${ad.agency}`.toLocaleLowerCase("pt-BR");
    const matchesSearch = !filters.search || haystack.includes(filters.search.toLocaleLowerCase("pt-BR"));
    const matchesAgency = filters.agency === "Todos" || ad.agency === filters.agency;
    const matchesCriterion = filters.criterion === "Todos" || ad.judgmentCriterion === filters.criterion;
    const matchesType = filters.type === "Todos" || (filters.type === "Consórcio" ? ad.adType === "consortium" : ad.adType !== "consortium");
    const matchesNeed = filters.need === "Todas" || `${ad.seekSummary} ${JSON.stringify(ad.requirements || [])}`.toLocaleLowerCase("pt-BR").includes(filters.need.toLocaleLowerCase("pt-BR"));
    return matchesSearch && matchesAgency && matchesCriterion && matchesType && matchesNeed;
  });
  const orderedAds = [...filteredAds].sort((a, b) => {
    if (filters.sort === "company_asc") return String(a.companyName || a.leaderCompanyName || "").localeCompare(String(b.companyName || b.leaderCompanyName || ""), "pt-BR");
    if (filters.sort === "tender_asc") return String(a.tenderNumber || "").localeCompare(String(b.tenderNumber || ""), "pt-BR", { numeric: true });
    return new Date(b.publishedAt || 0).getTime() - new Date(a.publishedAt || 0).getTime();
  });
  const updateFilter = (field, value) => setFilters((current) => ({ ...current, [field]: value }));
  const clearFilters = () => setFilters({ search: "", agency: "Todos", criterion: "Todos", type: "Todos", need: "Todas", sort: "published_desc" });
  const activeFilterCount = [filters.search, filters.agency !== "Todos", filters.criterion !== "Todos", filters.type !== "Todos", filters.need !== "Todas", filters.sort !== "published_desc"].filter(Boolean).length;
  const hasOwnPublishedPositionForTender = (tenderId) => ads.some((candidate) => (
    candidate.companyId === sessionUser?.companyId
    && candidate.tenderId === tenderId
    && candidate.status === "published"
    && ["company", "consortium"].includes(candidate.adType)
  ));

  const openAdEditor = (ad) => {
    setEditingAdId(ad.id);
    setAdForm({ offerSummary: ad.offerSummary || "", seekSummary: ad.seekSummary || "" });
  };

  const saveAd = async (ad) => {
    setSavingAdId(ad.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/partnership-ads/${ad.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(adForm)
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível editar o anúncio.");
      setEditingAdId("");
      await loadAds();
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingAdId("");
    }
  };

  const deleteAd = async (ad) => {
    if (!window.confirm(`Excluir o anúncio da sua empresa para o edital ${ad.tenderNumber}? Ele deixará de aparecer na vitrine.`)) return;
    setSavingAdId(ad.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/partnership-ads/${ad.id}`, {
        method: "DELETE",
        credentials: "include"
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível excluir o anúncio.");
      if (editingAdId === ad.id) setEditingAdId("");
      await loadAds();
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingAdId("");
    }
  };

  return (
    <Page label="Match e consórcios" title="Vitrine de parceiros">
      <section className="partnerShowcaseHero">
        <div>
          <span className="badge">Classificados de parcerias</span>
          <h3>Anúncios de empresas interessadas em licitações</h3>
          <p>Uma vitrine geral para encontrar oportunidades de consórcio em várias licitações, comparando o edital, o órgão, a empresa interessada, o que ela atende e o que busca complementar.</p>
        </div>
      </section>

      <Card className="partnerFiltersSticky partnerFilterBar">
        <div className="partnerFilterPrimary">
          <Field label="Buscar oportunidades">
            <input value={filters.search} onChange={(event) => updateFilter("search", event.target.value)} placeholder="Empresa, edital, órgão ou objeto" />
          </Field>
          <div className="partnerFilterSummary"><span>{orderedAds.length} anúncio{orderedAds.length === 1 ? "" : "s"} encontrado{orderedAds.length === 1 ? "" : "s"}</span>{activeFilterCount > 0 && <button type="button" className="iconButton secondaryIcon" title="Limpar filtros" aria-label="Limpar filtros" onClick={clearFilters}>×</button>}</div>
        </div>
        <div className="partnerFilterSecondary" aria-label="Filtros da vitrine">
          <Field label="Órgão"><select value={filters.agency} onChange={(event) => updateFilter("agency", event.target.value)}>{agencies.map((agency) => <option key={agency}>{agency}</option>)}</select></Field>
          <Field label="Critério"><select value={filters.criterion} onChange={(event) => updateFilter("criterion", event.target.value)}>{criteria.map((criterion) => <option key={criterion}>{criterion}</option>)}</select></Field>
          <Field label="Tipo de anúncio"><select value={filters.type} onChange={(event) => updateFilter("type", event.target.value)}><option>Todos</option><option>Empresa</option><option>Consórcio</option></select></Field>
          <Field label="Necessidade"><select value={filters.need} onChange={(event) => updateFilter("need", event.target.value)}><option>Todas</option><option>operacional</option><option>equipe</option><option>peça técnica</option><option>certificações</option></select></Field>
          <Field label="Ordenar por"><select value={filters.sort} onChange={(event) => updateFilter("sort", event.target.value)}><option value="published_desc">Mais recentes</option><option value="tender_asc">Número do edital</option><option value="company_asc">Empresa A-Z</option></select></Field>
        </div>
      </Card>

      {!loading && !error && (
        <section className="myAdsSection" aria-label="Meus anúncios">
          <div className="myAdsSectionHead">
            <div><span className="eyebrow">Minha empresa</span><h3>Meus anúncios</h3><p>{myAds.length ? `${myAds.length} anúncio${myAds.length > 1 ? "s" : ""} ativo${myAds.length > 1 ? "s" : ""} na vitrine.` : "Sua empresa ainda não possui anúncios ativos na vitrine."}</p></div>
            <button className="iconButton secondaryIcon" title={myAdsOpen ? "Recolher meus anúncios" : "Expandir meus anúncios"} aria-label={myAdsOpen ? "Recolher meus anúncios" : "Expandir meus anúncios"} onClick={() => setMyAdsOpen((current) => !current)}>{myAdsOpen ? "−" : "+"}</button>
          </div>
          {myAdsOpen && (myAds.length ? <div className="myAdsList">
            {myAds.map((ad) => {
              const isCompanyAd = ad.companyId === sessionUser?.companyId && ad.adType !== "consortium";
              const isEditing = editingAdId === ad.id;
              return <div className="myAdRow" key={ad.id}>
                <div className="myAdInfo"><strong>{ad.tenderNumber} · {ad.agency}</strong><span>{ad.tenderObject}</span><small>{ad.adType === "consortium" ? "Anúncio do consórcio" : "Anúncio da sua empresa"}</small></div>
                <div className="myAdActions">
                  <Button variant="secondary" onClick={() => navigate(`match-profile?id=${ad.id}`)}>Ver detalhe</Button>
                  {isCompanyAd ? <><Button variant="secondary" onClick={() => isEditing ? setEditingAdId("") : openAdEditor(ad)}>Editar</Button><Button variant="danger" onClick={() => deleteAd(ad)} disabled={savingAdId === ad.id}>{savingAdId === ad.id ? "Encerrando..." : "Excluir"}</Button></> : <Button variant="secondary" onClick={() => navigate("match-list")}>Gerenciar consórcio</Button>}
                </div>
                {isEditing && <div className="myAdEditor"><FormGrid><Field label="O que a empresa oferece"><textarea value={adForm.offerSummary} onChange={(event) => setAdForm((current) => ({ ...current, offerSummary: event.target.value }))} /></Field><Field label="O que a empresa busca"><textarea value={adForm.seekSummary} onChange={(event) => setAdForm((current) => ({ ...current, seekSummary: event.target.value }))} /></Field></FormGrid><div className="actions"><Button onClick={() => saveAd(ad)} disabled={savingAdId === ad.id}>{savingAdId === ad.id ? "Salvando..." : "Salvar alterações"}</Button><Button variant="secondary" onClick={() => setEditingAdId("")}>Cancelar</Button></div></div>}
              </div>;
            })}
          </div> : <p className="emptyMyAds">Registre interesse em um edital para publicar o primeiro anúncio da empresa.</p>)}
        </section>
      )}

      {loading && <Card><p>Carregando vitrine de parceiros...</p></Card>}
      {error && <Card className="dangerNotice"><p>{error}</p></Card>}
      {!loading && !error && orderedAds.length === 0 && <Card><p>Nenhum anúncio encontrado com esses filtros.</p></Card>}
      <div className="partnerAdGrid">
        {!loading && !error && orderedAds.map((ad) => (
          (() => {
            const isOwnAd = isMyAd(ad);
            const canEvaluate = hasOwnPublishedPositionForTender(ad.tenderId);
            const alreadyLiked = ad.currentEvaluationDecision === "liked";
            return (
          <Card className={`partnerAdCard ${ad.adType === "consortium" ? "consortiumAdCard" : ""}`} key={ad.id}>
            {isOwnAd && <span className="badge">Meu anúncio</span>}
            {ad.adType === "consortium" && <span className="consortiumRibbon">Consórcio em formação</span>}
            <div className="classifiedTenderHead">
              <span className="badge">{ad.adType === "consortium" ? "Consórcio em formação" : ad.tenderNumber}</span>
              <small>{ad.agency} | {ad.judgmentCriterion || "Critério não informado"}</small>
              <strong>{ad.tenderObject}</strong>
            </div>
            <div className="partnerAdCardHead">
              <LogoSlot initials={ad.companyName.split(" ").map((word) => word[0]).join("").slice(0, 2)} src={ad.companyLogoUrl} size="sm" label={`Logo da ${ad.companyName}`} />
              <div>
                <h3>{ad.adType === "consortium" ? "Consórcio busca nova consorciada" : ad.companyName}</h3>
                <p>{ad.adType === "consortium" ? `Líder: ${ad.leaderCompanyName || ad.companyName}` : [ad.city, ad.state].filter(Boolean).join(" / ") || "Local não informado"}</p>
              </div>
            </div>
            {ad.adType === "consortium" && (ad.consortiumMembers || []).length > 0 && <div className="consortiumMembersGroup"><small>Empresas que já compõem o consórcio</small><div className="consortiumMembersLine">{ad.consortiumMembers.map((company) => <span key={`${ad.id}-${company.companyId}`}>{company.companyName}</span>)}</div></div>}
            <div className="matchColumns compactMatchColumns">
              <div><strong>Oferece</strong><span>{ad.offerSummary || "Não informado"}</span></div>
              <div><strong>Busca</strong><span>{ad.seekSummary || "Não informado"}</span></div>
            </div>
            <div className="adStatusList">
              {(ad.requirements || []).map((item) => <span key={item.requirementKey}>{item.name}: {interestStatusLabels[item.statusKey] || item.statusKey}</span>)}
            </div>
            <div className="actions">
              {ad.companyId && <Button variant="secondary" onClick={() => navigate(`company-public-profile?companyId=${encodeURIComponent(ad.companyId)}`)}>Ver perfil da empresa</Button>}
              <Button onClick={() => navigate(`match-profile?id=${ad.id}`)}>Ver detalhe do anúncio</Button>
              {!isOwnAd && <Button variant="secondary" onClick={() => openChatForAd(ad)}>Conversar</Button>}
              {!isOwnAd && (canEvaluate
                ? (alreadyLiked ? <span className="statusPill open">Avaliação registrada</span> : <Button variant="secondary" onClick={() => navigate(`match-tinder?id=${ad.id}`)}>Avaliar candidata</Button>)
                : <Button variant="secondary" onClick={() => navigate(`tender-interest?id=${ad.tenderId}`)}>Registrar interesse</Button>
              )}
            </div>
          </Card>
            );
          })()
        ))}
      </div>
    </Page>
  );
}

function MatchTinder({ navigate, sessionUser }) {
  const id = currentHashParams().get("id") || "";
  const [ad, setAd] = useState(null);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState("");

  useEffect(() => {
    if (!id) {
      setError("Selecione um anúncio para avaliar.");
      return;
    }
    fetch(`${API_BASE_URL}/api/partnership-ads/${id}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o anúncio.");
        return data;
      })
      .then(setAd)
      .catch((err) => setError(err.message));
  }, [id]);

  const evaluate = async (decision) => {
    setSaving(decision);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/partnership-ads/${id}/evaluate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ decision })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível registrar a avaliação.");
      if (!data || typeof data !== "object") throw new Error("Não foi possível registrar a avaliação deste anúncio. Atualize a página e tente novamente.");
      if (data.matchCreated && data.matchId) {
        navigate(`match-success?id=${data.matchId}`);
      } else if (data.applicationAccepted) {
        navigate("match-list?joined=1");
      } else if (data.applicationCreated) {
        navigate(`match-profile?id=${id}&application=sent`);
      } else {
        navigate(`match-profile?id=${id}`);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving("");
    }
  };

  if (error) return <Page label="Match e consórcios" title="Avaliar candidata da licitação"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  if (!ad) return <Page label="Match e consórcios" title="Avaliar candidata da licitação"><Card><p>Carregando candidata...</p></Card></Page>;
  const isConsortiumAd = ad.adType === "consortium";
  const isOwnAd = (ad.companyId && sessionUser?.companyId && ad.companyId === sessionUser.companyId) || (ad.consortiumMembers || []).some((company) => company.companyId === sessionUser?.companyId);
  const alreadyLiked = ad.myEvaluationDecision === "liked";

  const hasItems = (ad.requirements || []).filter((item) => ["fully_meets", "partially_meets"].includes(item.statusKey)).map((item) => `${item.name}: ${interestStatusLabels[item.statusKey]}`);
  const seekItems = (ad.requirements || []).filter((item) => item.whatWeSeek || ["does_not_meet", "seeks_partner"].includes(item.statusKey)).map((item) => item.whatWeSeek || `${item.name}: ${interestStatusLabels[item.statusKey]}`);

  return (
    <Page label="Match e consórcios" title="Avaliar candidata da licitação">
      <div className="tinderStage">
        <div className="tinderPhone">
          <div className="tinderPhoto">
            <LogoSlot initials={ad.companyName.split(" ").map((word) => word[0]).join("").slice(0, 2)} src={ad.companyLogoUrl} size="xl" label={`Logo da ${ad.companyName}`} />
            <span className="tenderChip">{ad.tenderNumber}</span>
            <div className="tinderIdentity">
              <h3>{isConsortiumAd ? "Consórcio em formação" : ad.companyName}</h3>
              <p>{isConsortiumAd ? `Líder: ${ad.leaderCompanyName || ad.companyName}` : [ad.city, ad.state].filter(Boolean).join(" / ") || ad.agency}</p>
            </div>
          </div>

          <div className="tinderInfo">
            <strong>{ad.title}</strong>
            {isConsortiumAd && <div className="consortiumMembersLine">{(ad.consortiumMembers || []).map((company) => <span key={`${ad.id}-tinder-${company.companyId}`}>{company.companyName}</span>)}</div>}
            <div className="matchColumns">
              <div><strong>{isConsortiumAd ? "Consórcio" : "Tem"}</strong><span>{hasItems.join(" | ") || ad.offerSummary || "Não informado"}</span></div>
              <div><strong>{isConsortiumAd ? "Busca nova consorciada" : "Busca"}</strong><span>{seekItems.join(" | ") || ad.seekSummary || "Não informado"}</span></div>
            </div>
          </div>

          <div className="tinderBottomActions">
            <button className="circleButton info" onClick={() => navigate(`match-profile?id=${ad.id}`)}>i</button>
            {!isOwnAd && <button className="circleButton reject" title="Recusar candidata" aria-label="Recusar candidata" disabled={Boolean(saving)} onClick={() => evaluate("rejected")}>{"\u00D7"}</button>}
            {!isOwnAd && !alreadyLiked && <button className="circleButton like" title="Aprovar candidata" aria-label="Aprovar candidata" disabled={Boolean(saving)} onClick={() => evaluate("liked")}>{"\u2665"}</button>}
            {!isOwnAd && <button className="circleButton save" title="Avaliar depois" aria-label="Avaliar depois" disabled={Boolean(saving)} onClick={() => evaluate("later")}>{"\u2605"}</button>}
          </div>
          {isOwnAd && <p className="ownAdNotice">Este é o anúncio da sua empresa. Você pode revisar o conteúdo, mas não pode gerar consórcio consigo mesmo.</p>}
        </div>
      </div>
    </Page>
  );
}

function MatchProfile({ navigate, sessionUser, openChatForAd }) {
  const id = currentHashParams().get("id") || "";
  const [ad, setAd] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) {
      setError("Selecione um anúncio para ver o detalhe.");
      return;
    }
    fetch(`${API_BASE_URL}/api/partnership-ads/${id}`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o anúncio.");
        return data;
      })
      .then(setAd)
      .catch((err) => setError(err.message));
  }, [id]);

  if (error) return <Page label="Match e consórcios" title="Detalhe do anúncio"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  if (!ad) return <Page label="Match e consórcios" title="Detalhe do anúncio"><Card><p>Carregando anúncio...</p></Card></Page>;
  const isConsortiumAd = ad.adType === "consortium";
  const isOwnAd = (ad.companyId && sessionUser?.companyId && ad.companyId === sessionUser.companyId) || (ad.consortiumMembers || []).some((company) => company.companyId === sessionUser?.companyId);
  const applicationSent = currentHashParams().get("application") === "sent";
  const alreadyEvaluated = Boolean(ad.myEvaluationDecision);

  return (
    <Page label="Match e consórcios" title="Detalhe do anúncio" actions={!isOwnAd ? (alreadyEvaluated ? <Button variant="secondary" disabled>Já avaliada</Button> : <Button onClick={() => navigate(`match-tinder?id=${ad.id}`)}>Ir para avaliar candidata</Button>) : null}>
      {applicationSent && isConsortiumAd && <Card className="successNotice"><strong>Interesse enviado para a líder do consórcio.</strong><p>A empresa líder receberá a candidata em Meus consórcios e poderá dar o match final para incluí-la na composição.</p></Card>}
      <section className="partnerAdHero">
        <div>
          <span className="badge">{isConsortiumAd ? "Consórcio em formação" : isOwnAd ? "Meu anúncio" : ad.tenderNumber}</span>
          {isOwnAd && <span className="badge">{ad.tenderNumber}</span>}
          <h3>{ad.title}</h3>
          <p>{ad.tenderObject}</p>
          {isConsortiumAd && (ad.consortiumMembers || []).length > 0 && <div className="consortiumMembersGroup"><small>Empresas que já compõem o consórcio</small><div className="consortiumMembersLine">{ad.consortiumMembers.map((company) => <span key={`${ad.id}-profile-${company.companyId}`}>{company.companyName}</span>)}</div></div>}
        </div>
        <div className="partnerAdCompany">
          <LogoSlot initials={ad.companyName.split(" ").map((word) => word[0]).join("").slice(0, 2)} src={ad.companyLogoUrl} size="lg" label={`Logo da ${ad.companyName}`} />
          <div className="partnerAdCompanyName"><strong>{isConsortiumAd ? "Líder: " + (ad.leaderCompanyName || ad.companyName) : ad.companyName}</strong>{!isOwnAd && <button className="iconButton chatIcon" title="Abrir conversa geral com o anunciante" aria-label="Abrir conversa geral com o anunciante" onClick={() => openChatForAd(ad)}>{"\u2709"}</button>}</div>
          <small>{[ad.city, ad.state].filter(Boolean).join(" / ") || ad.agency} | {isConsortiumAd ? "Anúncio do consórcio" : "Anúncio de parceria"}</small>
        </div>
      </section>

      <section className="partnerRequirementSection">
        <div className="sectionHeading"><div><span className="eyebrow">Posicionamento na licitação</span><h3>Atendimento por requisito</h3><p>Veja separadamente o nível de atendimento, o que a empresa possui e o que deseja complementar.</p></div></div>
        <div className="partnerAdRequirements">
          {(ad.requirements || []).map((item) => (
            <Card className="partnerAdRequirement" key={item.requirementKey}>
              <div className="partnerRequirementHead"><div><span>Requisito</span><h3>{item.name}</h3></div><span className="statusPill review">{interestStatusLabels[item.statusKey] || item.statusKey}</span></div>
              <div className="partnerRequirementDetail"><strong>O que possui ou atende</strong><p>{item.whatWeHave || "Não informado"}</p></div>
              <div className="partnerRequirementDetail"><strong>O que busca complementar</strong><p>{item.whatWeSeek || "Não informado"}</p></div>
            </Card>
          ))}
        </div>
        {!(ad.requirements || []).length && <Card><p>Este anúncio ainda não possui o detalhamento por requisito.</p></Card>}
      </section>
    </Page>
  );
}

function MatchSuccess() {
  const id = currentHashParams().get("id") || "";
  const [matches, setMatches] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    const endpoint = id ? `${API_BASE_URL}/api/matches/${id}` : `${API_BASE_URL}/api/matches`;
    fetch(endpoint, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar o match.");
        return Array.isArray(data) ? data : [data];
      })
      .then(setMatches)
      .catch((err) => setError(err.message));
  }, [id]);

  if (error) return <Page label="Match e consórcios" title="Match realizado"><Card className="dangerNotice"><p>{error}</p></Card></Page>;
  if (!matches.length) return <Page label="Match e consórcios" title="Match realizado"><Card><p>Nenhum match ativo encontrado.</p></Card></Page>;
  const match = matches[0];
  const contacts = match.contacts || [];

  return (
    <Page label="Match e consórcios" title="Match realizado">
      <Card className="success matchSuccessCard">
        <h3>{match.companyAName} + {match.companyBName}</h3>
        <p>As duas empresas demonstraram interesse recíproco na licitação {match.tenderNumber}. Os anúncios envolvidos foram fechados na vitrine e o vínculo ficou registrado no banco.</p>
        {contacts.map((contact) => (
          <div className="matchContactGrid" key={contact.companyId}>
            <div>
              <strong>{contact.contactName}</strong>
              <span>{contact.companyName}</span>
              <small>{contact.phone}</small>
            </div>
            {contact.whatsappUrl ? <a className="whatsappButton" href={contact.whatsappUrl} target="_blank" rel="noreferrer">Abrir WhatsApp</a> : <span className="statusPill review">WhatsApp não informado</span>}
          </div>
        ))}
      </Card>
    </Page>
  );
}

function MatchList({ sessionUser, navigate }) {
  const [matches, setMatches] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [openMatchId, setOpenMatchId] = useState("");
  const [leaders, setLeaders] = useState({});
  const [notes, setNotes] = useState({});
  const [adNeeds, setAdNeeds] = useState({});
  const [adNotes, setAdNotes] = useState({});
  const [adFormsOpen, setAdFormsOpen] = useState({});
  const [savingId, setSavingId] = useState("");
  const [savingAdId, setSavingAdId] = useState("");
  const [savingApplicationId, setSavingApplicationId] = useState("");
  const [withdrawingMatchId, setWithdrawingMatchId] = useState("");
  const [filters, setFilters] = useState({ search: "", participation: "all", sort: "matched_desc" });
  const joinedByReciprocalLike = currentHashParams().get("joined") === "1";
  const canManageConsortium = sessionUser?.roleKey === "company_admin" || sessionUser?.roleKey === "commercial";

  const loadMatches = () => {
    setLoading(true);
    fetch(`${API_BASE_URL}/api/matches`, { credentials: "include" })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Não foi possível carregar seus consórcios.");
        return data;
      })
      .then((data) => {
        setMatches(data);
        setLeaders(Object.fromEntries(data.map((match) => [match.id, match.leadCompanyId || ""])));
        setNotes(Object.fromEntries(data.map((match) => [match.id, match.consortiumNotes || ""])));
        setAdNeeds(Object.fromEntries(data.map((match) => [match.id, match.consortiumAdSeekSummary || ""])));
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadMatches();
  }, []);

  const compactObject = (text) => {
    const clean = String(text || "Não informado").trim();
    return clean.length > 120 ? `${clean.slice(0, 120)}...` : clean;
  };

  const companiesFor = (match) => {
    const companies = new Map();
    const addCompany = (company) => {
      if (!company?.companyName) return;
      const key = company.companyId || company.companyName;
      const existingEntry = Array.from(companies.entries()).find(([, current]) => current.companyName === company.companyName);
      if (!company.companyId && existingEntry) return;
      if (company.companyId && existingEntry && existingEntry[0] !== key) companies.delete(existingEntry[0]);
      companies.set(key, { ...companies.get(key), ...company });
    };
    const consortiumMembers = match.consortiumMembers || [];
    if (match.consortiumIntentionId && consortiumMembers.length > 0) {
      consortiumMembers.forEach(addCompany);
      return Array.from(companies.values());
    }
    (match.contacts || []).forEach(addCompany);
    consortiumMembers.forEach(addCompany);
    addCompany({ companyName: match.companyAName });
    addCompany({ companyName: match.companyBName });
    return Array.from(companies.values());
  };

  const filteredMatches = matches.filter((match) => {
    const members = companiesFor(match);
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    const searchable = [match.tenderNumber, match.agency, match.tenderObject, match.leadCompanyName, ...members.map((company) => company.companyName)].join(" ").toLocaleLowerCase("pt-BR");
    const matchesSearch = !term || searchable.includes(term);
    const isLeader = match.leadCompanyId === sessionUser?.companyId;
    const matchesParticipation = filters.participation === "all" || (filters.participation === "leader" ? isLeader : !isLeader);
    return matchesSearch && matchesParticipation;
  }).sort((a, b) => {
    if (filters.sort === "tender_asc") return String(a.tenderNumber || "").localeCompare(String(b.tenderNumber || ""), "pt-BR", { numeric: true });
    if (filters.sort === "agency_asc") return String(a.agency || "").localeCompare(String(b.agency || ""), "pt-BR");
    return new Date(b.matchedAt || 0).getTime() - new Date(a.matchedAt || 0).getTime();
  });

  const saveLeader = async (match) => {
    const leadCompanyId = leaders[match.id];
    if (!leadCompanyId) {
      setError("Selecione quem será o líder do consórcio.");
      return;
    }
    setSavingId(match.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/matches/${match.id}/leader`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ leadCompanyId, notes: notes[match.id] || "" })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível definir o líder.");
      navigate(`assembly-board?matchId=${encodeURIComponent(match.id)}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingId("");
    }
  };

  const createConsortiumAd = async (match) => {
    const needSummary = String(adNeeds[match.id] || "").trim();
    if (!needSummary) {
      setError("Descreva o que falta complementar para publicar o anúncio do consórcio.");
      return;
    }
    setSavingAdId(match.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/matches/${match.id}/consortium-ad`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ needSummary, notes: adNotes[match.id] || "" })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível publicar o anúncio do consórcio.");
      setAdFormsOpen((current) => ({ ...current, [match.id]: false }));
      loadMatches();
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingAdId("");
    }
  };

  const acceptApplication = async (match, application) => {
    setSavingApplicationId(application.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/matches/${match.id}/application-accept`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ applicationId: application.id })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível aceitar a candidata.");
      loadMatches();
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingApplicationId("");
    }
  };

  const withdrawFromConsortium = async (match, members) => {
    const isLeader = match.leadCompanyId === sessionUser?.companyId;
    const remainingMembers = members.filter((company) => company.companyId !== sessionUser?.companyId);
    const successorCompanyId = isLeader && remainingMembers.length >= 2 ? leaders[match.id] : "";
    if (isLeader && remainingMembers.length >= 2 && (!successorCompanyId || successorCompanyId === sessionUser?.companyId)) {
      setError("Antes de desistir, selecione outra empresa ativa como líder do consórcio.");
      return;
    }
    const warning = remainingMembers.length < 2
      ? "A desistência deixará menos de duas empresas e encerrará este consórcio. Confirmar?"
      : "Sua empresa deixará o consórcio. O anúncio de busca, se existir, será fechado. Confirmar desistência?";
    if (!window.confirm(warning)) return;
    setWithdrawingMatchId(match.id);
    setError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/matches/${match.id}/withdraw`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ successorCompanyId })
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || "Não foi possível registrar a desistência do consórcio.");
      setOpenMatchId("");
      loadMatches();
    } catch (err) {
      setError(err.message);
    } finally {
      setWithdrawingMatchId("");
    }
  };

  if (loading) return <Page label="Editais" title="Meus consórcios"><Card><p>Carregando consórcios...</p></Card></Page>;

  return (
    <Page label="Editais" title="Meus consórcios">
      {joinedByReciprocalLike && <Card className="successNotice"><strong>Match confirmado.</strong><p>O aceite foi recíproco e a nova empresa já foi incluída neste consórcio.</p></Card>}
      {error && <Card className="dangerNotice"><p>{error}</p></Card>}
      <Card className="compactFilters consortiumFiltersSticky">
        <FormGrid>
          <Field label="Buscar consórcio"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Edital, órgão, objeto ou empresa" /></Field>
          <Field label="Minha participação"><select value={filters.participation} onChange={(event) => setFilters((current) => ({ ...current, participation: event.target.value }))}><option value="all">Todos</option><option value="leader">Sou líder</option><option value="member">Sou consorciada</option></select></Field>
          <Field label="Ordenar por"><select value={filters.sort} onChange={(event) => setFilters((current) => ({ ...current, sort: event.target.value }))}><option value="matched_desc">Match mais recente</option><option value="tender_asc">Número do edital</option><option value="agency_asc">Órgão A-Z</option></select></Field>
        </FormGrid>
      </Card>
      {matches.length === 0 ? (
        <Card><p>Ainda não há consórcios/matches ativos para sua empresa.</p></Card>
      ) : filteredMatches.length === 0 ? (
        <Card><p>Nenhum consórcio corresponde aos filtros informados.</p></Card>
      ) : (
        <div className="consortiumList">
          {filteredMatches.map((match) => {
            const members = companiesFor(match);
            const isOpen = openMatchId === match.id;
            const isLeader = match.leadCompanyId && match.leadCompanyId === sessionUser?.companyId;
            const tenderStatusLabels = { draft: "rascunho", published: "publicado", under_review: "em análise", suspended: "suspenso", challenged: "impugnado", occurred: "ocorrido", closed: "encerrado", cancelled: "cancelado" };
            const tenderStatus = String(match.tenderStatus || match.tender?.status || "").trim().toLowerCase();
            const tenderIsPublished = tenderStatus === "published";
            const tenderStatusLabel = tenderStatusLabels[tenderStatus] || "em atualização";
            const tenderAssemblyMessage = tenderIsPublished
              ? ""
              : `A montagem está pausada porque o edital está ${tenderStatusLabel}.`;
            const adFormOpen = Boolean(adFormsOpen[match.id]);
            const applications = Array.isArray(match.applications) ? match.applications : [];
            const pendingApplications = applications.filter((application) => application.status === "interested");
            return (
              <Card className="consortiumItem" key={match.id}>
                <div className="consortiumRow">
                  <div><small>Edital</small><strong>{match.tenderNumber || "-"}</strong></div>
                  <div><small>Órgão</small><strong>{match.agency || "-"}</strong></div>
                  <div className="wide"><small>Objeto resumido</small><span>{compactObject(match.tenderObject)}</span></div>
                  <div><small>Líder</small><strong>{match.leadCompanyName || "A definir"}</strong></div>
                  <div className="wide consortiumCompanies"><small>Empresas consorciadas</small><strong>{members.map((company) => company.companyName).join(" • ")}</strong></div>
                  <div className="consortiumActions">
                    {isLeader && pendingApplications.length > 0 && <span className="statusPill review">{pendingApplications.length} candidata{pendingApplications.length > 1 ? "s" : ""} aguardando</span>}
                    {tenderIsPublished && <button
                      type="button"
                      className="assemblyEntryButton"
                      disabled={!match.leadCompanyId}
                      title={!match.leadCompanyId ? "Defina primeiro a empresa líder do consórcio" : "Abrir a Central de Montagem desta licitação"}
                      onClick={() => match.leadCompanyId && navigate(`assembly-board?matchId=${match.id}`)}
                    >
                      {!match.leadCompanyId ? "Defina a liderança" : (isLeader && ["company_admin", "commercial"].includes(sessionUser?.roleKey) ? "Seguir para montagem" : "Acompanhar montagem")}
                    </button>}
                    {!tenderIsPublished && <small className="consortiumAssemblyUnavailable">{tenderAssemblyMessage}</small>}
                    {tenderIsPublished && <button className="iconButton secondaryIcon" title="Gerenciar consórcio" aria-label="Gerenciar consórcio" onClick={() => setOpenMatchId(isOpen ? "" : match.id)}>{isOpen ? "\u2212" : "\u002B"}</button>}
                  </div>
                </div>
                {isOpen && tenderIsPublished && (
                  <div className="consortiumDrawer">
                    <div>
                      <h3>Definir líder do consórcio</h3>
                      <p>Escolha qual empresa ficará registrada como líder operacional deste consórcio.</p>
                      {match.leadCompanyName && <span className="statusPill open">Líder atual: {match.leadCompanyName}</span>}
                      <div className="consortiumMembersGroup"><small>Composição atual do consórcio</small><div className="consortiumMembersLine">{members.map((company) => <span key={`${match.id}-current-${company.companyId || company.companyName}`}>{company.companyName}</span>)}</div></div>
                    </div>
                    {canManageConsortium && <FormGrid>
                      <Field label="Empresa líder">
                        <select value={leaders[match.id] || ""} onChange={(event) => setLeaders((current) => ({ ...current, [match.id]: event.target.value }))}>
                          <option value="">Selecione</option>
                          {members.map((company) => <option key={`${match.id}-${company.companyId || company.companyName}`} value={company.companyId}>{company.companyName}</option>)}
                        </select>
                      </Field>
                      <Field label="Observação do consórcio">
                        <textarea value={notes[match.id] || ""} onChange={(event) => setNotes((current) => ({ ...current, [match.id]: event.target.value }))} placeholder="Ex.: empresa líder cuidará da proposta comercial e integração documental." />
                      </Field>
                    </FormGrid>}
                    <div className="actions">
                      {canManageConsortium && <Button onClick={() => saveLeader(match)} disabled={savingId === match.id}>{savingId === match.id ? "Salvando..." : "Salvar liderança"}</Button>}
                      {canManageConsortium && isLeader && <Button variant="secondary" onClick={() => setAdFormsOpen((current) => ({ ...current, [match.id]: !current[match.id] }))}>{match.consortiumAdId ? "Editar anúncio de busca" : "Buscar nova consorciada"}</Button>}
                      {sessionUser?.roleKey === "company_admin" && match.consortiumIntentionId && <Button variant="danger" onClick={() => withdrawFromConsortium(match, members)} disabled={withdrawingMatchId === match.id}>{withdrawingMatchId === match.id ? "Registrando desistência..." : "Desistir do consórcio"}</Button>}
                      <Button variant="secondary" onClick={() => setOpenMatchId("")}>Recolher</Button>
                    </div>
                    {isLeader && adFormOpen && (
                      <Card className="nestedPanel">
                        <h3>{match.consortiumAdId ? "Anúncio do consórcio publicado" : "Criar anúncio do consórcio"}</h3>
                        <p>Este anúncio aparecerá na vitrine e em empresas interessadas como consórcio em formação buscando nova consorciada. A comunicação fica centralizada na empresa líder.</p>
                        <div className="consortiumMembersLine">
                          {members.map((company) => <span key={`${match.id}-member-${company.companyId || company.companyName}`}>{company.companyName}</span>)}
                        </div>
                        <FormGrid>
                          <Field label="O que falta complementar">
                            <textarea value={adNeeds[match.id] || ""} onChange={(event) => setAdNeeds((current) => ({ ...current, [match.id]: event.target.value }))} placeholder="Ex.: precisamos de empresa com acervo em arqueologia e equipe disponível para coordenação ambiental." />
                          </Field>
                          <Field label="Observação interna">
                            <textarea value={adNotes[match.id] || ""} onChange={(event) => setAdNotes((current) => ({ ...current, [match.id]: event.target.value }))} placeholder="Observação da líder sobre a busca, se necessário." />
                          </Field>
                        </FormGrid>
                        <div className="actions">
                          <Button onClick={() => createConsortiumAd(match)} disabled={savingAdId === match.id}>{savingAdId === match.id ? "Publicando..." : "Publicar anúncio do consórcio"}</Button>
                        </div>
                      </Card>
                    )}
                    {canManageConsortium && isLeader && applications.length > 0 && (
                      <Card className="nestedPanel">
                        <h3>Candidatas aguardando seu match</h3>
                        <p>Ao dar match, a empresa será incluída oficialmente no consórcio e passará a vê-lo em Meus consórcios.</p>
                        <div className="applicationList">
                          {applications.map((application) => (
                            <div className="applicationItem" key={application.id}>
                              <strong>{application.companyName}</strong>
                              <span>{application.status === "interested" ? "Demonstrou interesse" : application.status}</span>
                              {application.status === "interested" && <Button variant="secondary" onClick={() => acceptApplication(match, application)} disabled={savingApplicationId === application.id}>{savingApplicationId === application.id ? "Registrando match..." : "Dar match e incluir"}</Button>}
                            </div>
                          ))}
                        </div>
                      </Card>
                    )}
                  </div>
                )}
              </Card>
            );
          })}
        </div>
      )}
    </Page>
  );
}

const assemblyTaskStatuses = [
  ["pending", "Pendente"],
  ["in_progress", "Em andamento"],
  ["waiting_information", "Aguardando informação"],
  ["blocked", "Bloqueada"],
  ["under_review", "Em revisão"],
  ["returned_for_adjustment", "Devolvida para ajuste"],
  ["completed", "Concluída"],
  ["not_applicable", "Não se aplica"]
];

const assemblyPriorities = [
  ["low", "Baixa"],
  ["normal", "Normal"],
  ["high", "Alta"],
  ["urgent", "Urgente"]
];

const assemblyStatusLabel = (value) => assemblyTaskStatuses.find(([key]) => key === value)?.[1] || value;

const personalTaskColumns = [
  ["pending", "Pendentes"],
  ["in_progress", "Em andamento"],
  ["waiting_information", "Aguardando informação"],
  ["blocked", "Bloqueadas"],
  ["under_review", "Em revisão"],
  ["returned_for_adjustment", "Para ajuste"],
  ["completed", "Concluídas"]
];

const currentLocalISODate = () => {
  const date = new Date();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
};

const assemblyDateError = (value, openingDate, fieldLabel) => {
  if (!value) return "";
  const today = currentLocalISODate();
  const opening = openingDate ? String(openingDate).slice(0, 10) : "";
  if (!opening) return "Informe a data de abertura do edital antes de definir prazos da montagem.";
  if (value < today) return `${fieldLabel} não pode ser anterior à data atual.`;
  if (value > opening) return `${fieldLabel} não pode ser posterior à data de abertura do edital.`;
  return "";
};

function MyAssemblyTasks({ navigate, openChatForTask }) {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ search: "", sort: "due_asc" });
  const [selected, setSelected] = useState(null);
  const [panelLoading, setPanelLoading] = useState(false);
  const [taskStatus, setTaskStatus] = useState("");
  const [savingStatus, setSavingStatus] = useState(false);
  const [comment, setComment] = useState("");
  const [savingComment, setSavingComment] = useState(false);
  const [evidence, setEvidence] = useState({ evidenceType: "file", title: "", externalUrl: "", note: "", file: null });
  const [savingEvidence, setSavingEvidence] = useState(false);

  const request = async (url, options = {}) => {
    const response = await fetch(`${API_BASE_URL}${url}`, { credentials: "include", ...options });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Não foi possível concluir esta operação.");
    return data;
  };

  const loadTasks = async () => {
    const data = await request("/api/my-assembly-tasks");
    setTasks(Array.isArray(data) ? data : []);
  };

  useEffect(() => {
    loadTasks()
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const deadline = (task) => {
    if (!task.dueAt || task.status === "completed") return null;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const due = new Date(`${String(task.dueAt).slice(0, 10)}T00:00:00`);
    const days = Math.ceil((due - today) / 86400000);
    if (days < 0) return { className: "overdue", label: `${Math.abs(days)} dia${Math.abs(days) === 1 ? "" : "s"} em atraso` };
    if (days <= 1) return { className: "soon", label: days === 0 ? "Vence hoje" : "Vence amanhã" };
    return { className: "planned", label: due.toLocaleDateString("pt-BR") };
  };

  const visibleTasks = tasks.filter((task) => {
    const term = filters.search.trim().toLocaleLowerCase("pt-BR");
    if (!term) return true;
    return [task.title, task.stageTitle, task.tenderNumber, task.agency, task.tenderObject].join(" ").toLocaleLowerCase("pt-BR").includes(term);
  }).sort((a, b) => {
    if (filters.sort === "updated_desc") return new Date(b.updatedAt || 0).getTime() - new Date(a.updatedAt || 0).getTime();
    if (filters.sort === "tender_asc") return String(a.tenderNumber || "").localeCompare(String(b.tenderNumber || ""), "pt-BR", { numeric: true });
    return new Date(a.dueAt || "9999-12-31").getTime() - new Date(b.dueAt || "9999-12-31").getTime();
  });

  const findTaskInAssembly = (assembly, taskId) => {
    for (const stage of assembly.stages || []) {
      const task = (stage.tasks || []).find((item) => item.id === taskId);
      if (task) return { assembly, task, stageTitle: stage.title };
    }
    return null;
  };

  const openTask = async (task) => {
    setPanelLoading(true);
    setError("");
    try {
      const assembly = await request(`/api/assemblies/${task.assemblyId}`);
      const detail = findTaskInAssembly(assembly, task.id);
      if (!detail) throw new Error("A tarefa não foi encontrada nesta montagem.");
      setSelected(detail);
      setTaskStatus(detail.task.status || "pending");
      setComment("");
      setEvidence({ evidenceType: "file", title: "", externalUrl: "", note: "", file: null });
    } catch (err) {
      setError(err.message);
    } finally {
      setPanelLoading(false);
    }
  };

  const refreshSelected = async () => {
    if (!selected) return;
    const assembly = await request(`/api/assemblies/${selected.assembly.id}`);
    const detail = findTaskInAssembly(assembly, selected.task.id);
    if (!detail) {
      setSelected(null);
      return;
    }
    setSelected(detail);
    setTaskStatus(detail.task.status || "pending");
  };

  const saveStatus = async () => {
    if (!selected) return;
    setSavingStatus(true);
    setError("");
    try {
      await request(`/api/assemblies/${selected.assembly.id}/tasks/${selected.task.id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: taskStatus })
      });
      await Promise.all([refreshSelected(), loadTasks()]);
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingStatus(false);
    }
  };

  const addComment = async () => {
    if (!selected || !comment.trim()) return;
    setSavingComment(true);
    setError("");
    try {
      await request(`/api/assemblies/${selected.assembly.id}/tasks/${selected.task.id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: comment.trim() })
      });
      setComment("");
      await Promise.all([refreshSelected(), loadTasks()]);
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingComment(false);
    }
  };

  const fileToDataURL = (file) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });

  const addEvidence = async () => {
    if (!selected || !evidence.title.trim()) {
      setError("Informe um título para o documento ou evidência.");
      return;
    }
    setSavingEvidence(true);
    setError("");
    try {
      const payload = { evidenceType: evidence.evidenceType, title: evidence.title.trim(), externalUrl: evidence.externalUrl.trim(), note: evidence.note.trim() };
      if (evidence.evidenceType === "file" && evidence.file) {
        payload.fileDataUrl = await fileToDataURL(evidence.file);
        payload.fileName = evidence.file.name;
        payload.mimeType = evidence.file.type;
      }
      await request(`/api/assemblies/${selected.assembly.id}/tasks/${selected.task.id}/evidences`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setEvidence({ evidenceType: "file", title: "", externalUrl: "", note: "", file: null });
      await Promise.all([refreshSelected(), loadTasks()]);
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingEvidence(false);
    }
  };
  const overdueCount = visibleTasks.filter((task) => deadline(task)?.className === "overdue").length;
  const dueSoonCount = visibleTasks.filter((task) => deadline(task)?.className === "soon").length;

  if (loading) return <Page label="Empresa" title="Minhas tarefas"><div className="assemblyLoading"><span />Carregando suas tarefas...</div></Page>;

  return (
    <Page label="Empresa" title="Minhas tarefas">
      {error && <Card className="dangerNotice"><p>{error}</p></Card>}
      <section className="personalTasksSummary">
        <div><strong>{visibleTasks.length}</strong><span>tarefas atribuídas a você</span></div>
        <div className={overdueCount ? "danger" : "ok"}><strong>{overdueCount}</strong><span>em atraso</span></div>
        <div className={dueSoonCount ? "attention" : "ok"}><strong>{dueSoonCount}</strong><span>vencem hoje ou amanhã</span></div>
      </section>
      <Card className="compactFilters personalTasksFilters">
        <FormGrid>
          <Field label="Buscar tarefa"><input value={filters.search} onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))} placeholder="Tarefa, fase, edital ou órgão" /></Field>
          <Field label="Ordenar por"><select value={filters.sort} onChange={(event) => setFilters((current) => ({ ...current, sort: event.target.value }))}><option value="due_asc">Prazo mais próximo</option><option value="updated_desc">Atualização mais recente</option><option value="tender_asc">Número do edital</option></select></Field>
        </FormGrid>
      </Card>
      {visibleTasks.length === 0 ? <Card><p>Você ainda não possui tarefas de montagem atribuídas diretamente ao seu usuário.</p></Card> : (
        <div className="personalKanban" aria-label="Kanban das minhas tarefas">
          {personalTaskColumns.map(([status, label]) => {
            const columnTasks = visibleTasks.filter((task) => task.status === status);
            return (
              <section className={`personalKanbanColumn status-${status}`} key={status}>
                <header><h3>{label}</h3><span>{columnTasks.length}</span></header>
                <div className="personalKanbanCards">
                  {columnTasks.map((task) => {
                    const taskDeadline = deadline(task);
                    return (
                      <button type="button" className={`personalTaskCard status-${task.status}${taskDeadline?.className === "overdue" ? " is-overdue" : ""}${taskDeadline?.className === "soon" ? " is-due-soon" : ""}`} key={task.id} onClick={() => openTask(task)} disabled={panelLoading}>
                        <span className="personalTaskTender">{task.tenderNumber} · {task.agency}</span>
                        <strong>{task.title}</strong>
                        <span className="personalTaskStage">{task.stageTitle}</span>
                        <div><span className={taskDeadline?.className || "planned"}>{taskDeadline?.label || "Sem prazo"}</span><span className="taskChatSummary" role="button" tabIndex="0" title="Abrir conversa da tarefa" onClick={(event) => { event.stopPropagation(); openChatForTask({ assemblyId: task.assemblyId, taskId: task.id }); }}><i className="taskChatIndicator" aria-hidden="true"><b /><b /><b /></i>Chat</span></div>
                      </button>
                    );
                  })}
                  {columnTasks.length === 0 && <p className="personalKanbanEmpty">Nenhuma tarefa.</p>}
                </div>
              </section>
            );
          })}
        </div>
      )}
      {selected && (
        <div className="assemblyPanelBackdrop" onMouseDown={(event) => event.target === event.currentTarget && setSelected(null)}>
          <aside className="assemblyTaskPanel personalTaskPanel" aria-label="Atualizar tarefa">
            <header><div><span>{selected.task.tenderNumber || selected.assembly.tenderNumber} · {selected.stageTitle}</span><h3>{selected.task.title}</h3></div><button type="button" onClick={() => setSelected(null)} aria-label="Fechar tarefa">×</button></header>
            <div className="assemblyTaskPanelBody">
              <section className="assemblyTaskSection">
                <h4>Andamento</h4>
                <p className="personalTaskDescription">{selected.task.description || "Sem descrição complementar."}</p>
                <Field label="Status"><select value={taskStatus} onChange={(event) => setTaskStatus(event.target.value)}>{assemblyTaskStatuses.filter(([value]) => value !== "not_applicable").map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
                {taskStatus === "completed" && <small>Ao concluir, a tarefa será enviada para revisão da empresa líder.</small>}
                <Button onClick={saveStatus} disabled={savingStatus || taskStatus === selected.task.status}>{savingStatus ? "Salvando..." : "Atualizar andamento"}</Button>
              </section>
              <section className="assemblyTaskSection">
                <h4>Comentários</h4>
                <div className="assemblyComments">{(selected.task.comments || []).length === 0 ? <p>Nenhum comentário nesta tarefa.</p> : selected.task.comments.map((item) => <article key={item.id}><strong>{item.userName} · {item.companyName}</strong><p>{item.content}</p><time>{new Date(item.createdAt).toLocaleString("pt-BR")}</time></article>)}</div>
                <div className="assemblyComposer"><textarea value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Registre uma orientação, pendência ou retorno" maxLength={2000} /><Button onClick={addComment} disabled={savingComment || !comment.trim()}>{savingComment ? "Enviando..." : "Comentar"}</Button></div>
              </section>
              <section className="assemblyTaskSection">
                <h4>Documentos e evidências</h4>
                <div className="assemblyEvidenceList">{(selected.task.evidences || []).length === 0 ? <p>Nenhum documento anexado.</p> : selected.task.evidences.map((item) => <article key={item.id}><div><strong>{item.title}</strong><small>Versão {item.versionNumber} · {item.userName || item.companyName}</small></div>{item.url ? <a href={item.url} target="_blank" rel="noreferrer">Abrir</a> : <span>{item.note}</span>}</article>)}</div>
                <div className="assemblyEvidenceForm">
                  <div className="assemblyFieldGrid"><Field label="Tipo"><select value={evidence.evidenceType} onChange={(event) => setEvidence((current) => ({ ...current, evidenceType: event.target.value }))}><option value="file">Arquivo</option><option value="link">Link externo</option><option value="note">Anotação</option></select></Field><Field label="Título"><input value={evidence.title} onChange={(event) => setEvidence((current) => ({ ...current, title: event.target.value }))} placeholder="Ex.: CAT do coordenador" /></Field></div>
                  {evidence.evidenceType === "file" && <Field label="Arquivo"><input type="file" onChange={(event) => setEvidence((current) => ({ ...current, file: event.target.files?.[0] || null }))} /></Field>}
                  {evidence.evidenceType === "link" && <Field label="Endereço do documento"><input type="url" value={evidence.externalUrl} onChange={(event) => setEvidence((current) => ({ ...current, externalUrl: event.target.value }))} placeholder="https://" /></Field>}
                  {evidence.evidenceType === "note" && <Field label="Anotação"><textarea value={evidence.note} onChange={(event) => setEvidence((current) => ({ ...current, note: event.target.value }))} /></Field>}
                  <Button onClick={addEvidence} disabled={savingEvidence}>{savingEvidence ? "Incluindo..." : "Incluir no dossiê"}</Button>
                </div>
              </section>
            </div>
          </aside>
        </div>
      )}
    </Page>
  );
}

function AssemblyBoard({ sessionUser, navigate, openChatForTask }) {
  const matchId = currentHashParams().get("matchId") || "";
  const interestId = currentHashParams().get("interestId") || "";
  const assemblyId = currentHashParams().get("assemblyId") || "";
  const requestedTaskId = currentHashParams().get("taskId") || "";
  const [assembly, setAssembly] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [creating, setCreating] = useState(false);
  const [startForm, setStartForm] = useState({ title: "", dueDate: "" });
  const [selectedTaskId, setSelectedTaskId] = useState("");
  const [taskDraft, setTaskDraft] = useState(null);
  const [savingTask, setSavingTask] = useState(false);
  const [deletingTask, setDeletingTask] = useState(false);
  const [comment, setComment] = useState("");
  const [savingComment, setSavingComment] = useState(false);
  const [evidence, setEvidence] = useState({ evidenceType: "file", title: "", externalUrl: "", note: "", file: null });
  const [savingEvidence, setSavingEvidence] = useState(false);
  const [dossierOpen, setDossierOpen] = useState(false);
  const [newStageOpen, setNewStageOpen] = useState(false);
  const [newStage, setNewStage] = useState({ title: "", description: "" });
  const [newTaskStageId, setNewTaskStageId] = useState("");
  const [newTask, setNewTask] = useState({ title: "", description: "" });
  const [savingStructure, setSavingStructure] = useState(false);
  const todayISO = currentLocalISODate();
  const openingDate = assembly?.openingDate ? String(assembly.openingDate).slice(0, 10) : "";
  const canSchedule = Boolean(openingDate && openingDate >= todayISO);

  const request = async (url, options = {}) => {
    const response = await fetch(`${API_BASE_URL}${url}`, { credentials: "include", ...options });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Não foi possível concluir esta operação.");
    return data;
  };

  const loadAssembly = async ({ quiet = false } = {}) => {
    if (!matchId && !interestId && !assemblyId) {
      setError("Abra a Central de Montagem por uma participação ou consórcio.");
      setLoading(false);
      return;
    }
    if (!quiet) setLoading(true);
    try {
      const endpoint = assemblyId
        ? `/api/assemblies/${encodeURIComponent(assemblyId)}`
        : matchId
          ? `/api/assemblies?matchId=${encodeURIComponent(matchId)}`
          : `/api/assemblies?interestId=${encodeURIComponent(interestId)}`;
      const data = await request(endpoint);
      setAssembly(data);
      setError("");
      if (requestedTaskId && data.exists && Array.isArray(data.stages)) {
        const requestedTaskExists = data.stages.some((stage) =>
          (stage.tasks || []).some((task) => task.id === requestedTaskId)
        );
        if (requestedTaskExists) setSelectedTaskId(requestedTaskId);
      }
      if (data.exists === false && !startForm.title) {
        setStartForm((current) => ({ ...current, title: `Montagem da licitação ${data.tenderNumber || ""}`.trim() }));
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAssembly();
  }, [matchId, interestId, assemblyId, requestedTaskId]);

  const stages = assembly?.stages || [];
  const allTasks = stages.flatMap((stage) => (stage.tasks || []).map((task) => ({ ...task, stageId: stage.id, stageTitle: stage.title })));
  const selectedTask = allTasks.find((task) => task.id === selectedTaskId) || null;

  useEffect(() => {
    if (!selectedTask) {
      setTaskDraft(null);
      return;
    }
    setTaskDraft({
      title: selectedTask.title || "",
      description: selectedTask.description || "",
      status: selectedTask.status || "pending",
      priority: selectedTask.priority || "normal",
      responsibleCompanyId: selectedTask.responsibleCompanyId || "",
      responsibleUserId: selectedTask.responsibleUserId || "",
      dueDate: selectedTask.dueAt ? String(selectedTask.dueAt).slice(0, 10) : ""
    });
  }, [selectedTaskId, selectedTask?.status, selectedTask?.responsibleUserId, selectedTask?.dueAt]);

  const beginAssembly = async () => {
    const dateError = assemblyDateError(startForm.dueDate, openingDate, "Prazo geral");
    if (dateError) {
      setError(dateError);
      return;
    }
    setCreating(true);
    setError("");
    try {
      const data = await request("/api/assemblies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ matchId, interestId, ...startForm })
      });
      setAssembly(data);
      setNotice("Central de Montagem iniciada com as oito fases do Modelo LicitaHub.");
    } catch (err) {
      setError(err.message);
    } finally {
      setCreating(false);
    }
  };

  const saveTask = async () => {
    if (!selectedTask || !taskDraft) return;
    const dateError = assemblyDateError(taskDraft.dueDate, openingDate, "Prazo da tarefa");
    if (dateError) {
      setError(dateError);
      return;
    }
    setSavingTask(true);
    setError("");
    try {
      await request(`/api/assemblies/${assembly.id}/tasks/${selectedTask.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(taskDraft)
      });
      await loadAssembly({ quiet: true });
      setNotice(taskDraft.status === "completed" && !assembly.canManage ? "Tarefa enviada para revisão da empresa líder." : "Tarefa atualizada.");
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingTask(false);
    }
  };

  const deleteTask = async () => {
    if (!selectedTask || !assembly?.canManage) return;
    const confirmed = window.confirm(`Excluir a tarefa "${selectedTask.title}"? Comentários e documentos vinculados a ela também serão removidos.`);
    if (!confirmed) return;
    setDeletingTask(true);
    setError("");
    try {
      await request(`/api/assemblies/${assembly.id}/tasks/${selectedTask.id}`, { method: "DELETE" });
      setSelectedTaskId("");
      setTaskDraft(null);
      await loadAssembly({ quiet: true });
      setNotice("Tarefa excluída da montagem.");
    } catch (err) {
      setError(err.message);
    } finally {
      setDeletingTask(false);
    }
  };

  const addComment = async () => {
    if (!comment.trim() || !selectedTask) return;
    setSavingComment(true);
    try {
      await request(`/api/assemblies/${assembly.id}/tasks/${selectedTask.id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: comment.trim() })
      });
      setComment("");
      await loadAssembly({ quiet: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingComment(false);
    }
  };

  const fileToDataURL = (file) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });

  const addEvidence = async () => {
    if (!selectedTask || !evidence.title.trim()) {
      setError("Informe um título para o documento ou evidência.");
      return;
    }
    setSavingEvidence(true);
    setError("");
    try {
      const payload = {
        evidenceType: evidence.evidenceType,
        title: evidence.title.trim(),
        externalUrl: evidence.externalUrl.trim(),
        note: evidence.note.trim()
      };
      if (evidence.evidenceType === "file" && evidence.file) {
        payload.fileDataUrl = await fileToDataURL(evidence.file);
        payload.fileName = evidence.file.name;
        payload.mimeType = evidence.file.type;
      }
      await request(`/api/assemblies/${assembly.id}/tasks/${selectedTask.id}/evidences`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setEvidence({ evidenceType: "file", title: "", externalUrl: "", note: "", file: null });
      await loadAssembly({ quiet: true });
      setNotice("Documento ou evidência incluído no dossiê.");
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingEvidence(false);
    }
  };

  const createStage = async () => {
    if (!newStage.title.trim()) return;
    setSavingStructure(true);
    try {
      await request(`/api/assemblies/${assembly.id}/stages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newStage)
      });
      setNewStage({ title: "", description: "" });
      setNewStageOpen(false);
      await loadAssembly({ quiet: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingStructure(false);
    }
  };

  const createTask = async () => {
    if (!newTaskStageId || !newTask.title.trim()) return;
    setSavingStructure(true);
    try {
      await request(`/api/assemblies/${assembly.id}/stages/${newTaskStageId}/tasks`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newTask)
      });
      setNewTask({ title: "", description: "" });
      setNewTaskStageId("");
      await loadAssembly({ quiet: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingStructure(false);
    }
  };

  const dueState = (task) => {
    if (!task.dueAt || task.status === "completed" || task.status === "not_applicable") return null;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const due = new Date(`${String(task.dueAt).slice(0, 10)}T00:00:00`);
    const days = Math.ceil((due - today) / 86400000);
    if (days < 0) return { className: "overdue", label: `${Math.abs(days)} dia${Math.abs(days) === 1 ? "" : "s"} em atraso` };
    if (days <= 1) return { className: "soon", label: days === 0 ? "Vence hoje" : "Vence amanhã" };
    return { className: "planned", label: new Date(`${String(task.dueAt).slice(0, 10)}T12:00:00`).toLocaleDateString("pt-BR") };
  };

  if (loading) return <Page label="Licitações" title="Central de Montagem da Licitação"><div className="assemblyLoading"><span />Carregando o plano de montagem...</div></Page>;

  if ((!matchId && !interestId && !assemblyId) || (!assembly && error)) {
    return <Page label="Licitações" title="Central de Montagem da Licitação"><Card className="dangerNotice"><p>{error || "Consórcio não informado."}</p><Button variant="secondary" onClick={() => navigate("match-list")}>Voltar para Meus consórcios</Button></Card></Page>;
  }

  if (assembly?.exists === false) {
    return (
      <Page label="Licitações" title="Central de Montagem da Licitação">
        {error && <Card className="dangerNotice"><p>{error}</p></Card>}
        <section className="assemblyWelcome">
          <div className="assemblyWelcomeCopy">
            <span className="assemblyKicker">{assembly.assemblyType === "individual" ? "Participação individual pronta para organizar" : "Consórcio pronto para organizar"}</span>
            <h3>{assembly.tenderNumber} · {assembly.agency}</h3>
            <p>{assembly.tenderObject}</p>
            <div className="assemblyLeadLine"><span>{assembly.assemblyType === "individual" ? "Empresa responsável" : "Empresa líder"}</span><strong>{assembly.leadCompanyName}</strong></div>
          </div>
          {assembly.canCreate ? (
            <div className="assemblyStartForm">
              <h3>Iniciar a montagem</h3>
              <Field label="Nome do plano"><input value={startForm.title} onChange={(event) => setStartForm((current) => ({ ...current, title: event.target.value }))} maxLength={180} /></Field>
              <Field label="Prazo geral"><input type="date" value={startForm.dueDate} min={todayISO} max={openingDate || undefined} disabled={!canSchedule} onChange={(event) => setStartForm((current) => ({ ...current, dueDate: event.target.value }))} />{!canSchedule && <small>{openingDate ? "A abertura do edital já ocorreu; não há prazo disponível para planejamento." : "Defina a data de abertura do edital para liberar os prazos."}</small>}</Field>
              <Button onClick={beginAssembly} disabled={creating}>{creating ? "Preparando fases..." : "Iniciar com Modelo LicitaHub"}</Button>
            </div>
          ) : (
            <div className="assemblyWaiting"><strong>Aguardando a empresa responsável</strong><p>Assim que a responsável iniciar a montagem, os profissionais vinculados poderão acompanhar as fases e colaborar nas tarefas atribuídas.</p></div>
          )}
        </section>
      </Page>
    );
  }

  const applicableTasks = allTasks.filter((task) => task.status !== "not_applicable");
  const completedTasks = applicableTasks.filter((task) => task.status === "completed").length;
  const overallProgress = applicableTasks.length ? Math.round((completedTasks / applicableTasks.length) * 100) : 100;
  const overdueCount = allTasks.filter((task) => dueState(task)?.className === "overdue").length;
  const reviewCount = allTasks.filter((task) => task.status === "under_review").length;
  const blockedCount = allTasks.filter((task) => task.status === "blocked").length;
  const dossierItems = allTasks.flatMap((task) => (task.evidences || []).map((item) => ({ ...item, taskTitle: task.title, stageTitle: task.stageTitle })));
  const availableProfessionals = (assembly.professionals || []).filter((professional) => !taskDraft?.responsibleCompanyId || professional.companyId === taskDraft.responsibleCompanyId);

  return (
    <Page label="Licitações" title="Central de Montagem da Licitação" actions={<Button variant="secondary" onClick={() => navigate(assembly?.assemblyType === "individual" ? "tender-list" : "match-list")}>{assembly?.assemblyType === "individual" ? "Lista de editais" : "Meus consórcios"}</Button>}>
      {error && <Card className="dangerNotice"><p>{error}</p></Card>}
      {notice && <div className="assemblyNotice"><span>{notice}</span><button type="button" onClick={() => setNotice("")} aria-label="Fechar aviso">×</button></div>}

      <section className="assemblyOverview">
        <div className="assemblyTenderIdentity">
          <span>{assembly.tenderNumber} · {assembly.agency}</span>
          <h3>{assembly.title}</h3>
          <p>{assembly.tenderObject}</p>
          <div className="assemblyMemberList"><strong>Liderança: {assembly.leadCompanyName}</strong>{(assembly.members || []).map((member) => <span key={member.companyId}>{member.companyName}</span>)}</div>
        </div>
        <div className="assemblyProgressHero">
          <div className="assemblyProgressNumber"><strong>{overallProgress}%</strong><span>concluído</span></div>
          <div className="assemblyProgressTrack"><i style={{ width: `${overallProgress}%` }} /></div>
          <small>{completedTasks} de {applicableTasks.length} tarefas aplicáveis concluídas</small>
        </div>
        <div className="assemblyHealth">
          <div className={overdueCount ? "danger" : "ok"}><strong>{overdueCount}</strong><span>em atraso</span></div>
          <div className={reviewCount ? "attention" : "ok"}><strong>{reviewCount}</strong><span>em revisão</span></div>
          <div className={blockedCount ? "danger" : "ok"}><strong>{blockedCount}</strong><span>bloqueadas</span></div>
        </div>
      </section>

      <div className="assemblyToolbar">
        <div><strong>{stages.length} fases</strong><span>{allTasks.length} tarefas no plano</span></div>
        <div className="actions">
          <Button variant="secondary" onClick={() => setDossierOpen(true)}>Abrir dossiê ({dossierItems.length})</Button>
          {assembly.canManage && <Button variant="secondary" onClick={() => setNewStageOpen((current) => !current)}>Nova fase</Button>}
        </div>
      </div>

      {assembly.canManage && newStageOpen && (
        <section className="assemblyInlineForm">
          <div><h3>Criar fase complementar</h3><p>A fase será adicionada ao final deste plano e não altera o Modelo LicitaHub.</p></div>
          <Field label="Nome da fase"><input value={newStage.title} onChange={(event) => setNewStage((current) => ({ ...current, title: event.target.value }))} /></Field>
          <Field label="Objetivo"><input value={newStage.description} onChange={(event) => setNewStage((current) => ({ ...current, description: event.target.value }))} /></Field>
          <Button onClick={createStage} disabled={savingStructure || !newStage.title.trim()}>Criar fase</Button>
        </section>
      )}

      <div className="assemblyStageGrid">
        {stages.map((stage, index) => (
          <section className="assemblyStage" key={stage.id}>
            <header>
              <div className="assemblyStageNumber">{String(index + 1).padStart(2, "0")}</div>
              <div><span>Fase {index + 1}</span><h3>{stage.title}</h3></div>
              <strong>{stage.progress || 0}%</strong>
            </header>
            <p className="assemblyStageDescription">{stage.description}</p>
            <div className="assemblyStageProgress"><i style={{ width: `${stage.progress || 0}%` }} /></div>
            <div className="assemblyTaskList">
              {(stage.tasks || []).map((task) => {
                const deadline = dueState(task);
                return (
                  <button type="button" className={`assemblyTaskCard status-${task.status}${deadline?.className === "overdue" ? " is-overdue" : ""}${deadline?.className === "soon" ? " is-due-soon" : ""}`} key={task.id} onClick={() => setSelectedTaskId(task.id)}>
                    <div className="assemblyTaskTop"><span className={`assemblyStatusDot ${task.status}`} /><strong>{task.title}</strong><em className={`priority-${task.priority}`}>{assemblyPriorities.find(([key]) => key === task.priority)?.[1]}</em></div>
                    <div className="assemblyTaskMeta"><span>{assemblyStatusLabel(task.status)}</span>{task.responsibleCompanyName && <span>{task.responsibleCompanyName}</span>}</div>
                    <div className="assemblyTaskFoot">{deadline ? <span className={deadline.className}>{deadline.label}</span> : <span>Sem prazo</span>}{task.responsibleUserId ? <span className="taskChatSummary" role="button" tabIndex="0" title="Conversar com o responsável" onClick={(event) => { event.stopPropagation(); openChatForTask({ assemblyId: assembly.id, taskId: task.id }); }}><i className="taskChatIndicator" aria-hidden="true"><b /><b /><b /></i>Chat</span> : <span>{(task.comments || []).length} comentários · {(task.evidences || []).length} arquivos</span>}</div>
                  </button>
                );
              })}
            </div>
            {assembly.canManage && (
              newTaskStageId === stage.id ? (
                <div className="assemblyNewTaskForm">
                  <input value={newTask.title} onChange={(event) => setNewTask((current) => ({ ...current, title: event.target.value }))} placeholder="Nome da nova tarefa" />
                  <textarea value={newTask.description} onChange={(event) => setNewTask((current) => ({ ...current, description: event.target.value }))} placeholder="Descrição e critério de conclusão" />
                  <div className="actions"><Button onClick={createTask} disabled={savingStructure || !newTask.title.trim()}>Adicionar</Button><Button variant="secondary" onClick={() => setNewTaskStageId("")}>Cancelar</Button></div>
                </div>
              ) : <button type="button" className="assemblyAddTask" onClick={() => setNewTaskStageId(stage.id)}>+ Adicionar tarefa nesta fase</button>
            )}
          </section>
        ))}
      </div>

      {selectedTask && taskDraft && (
        <div className="assemblyPanelBackdrop" onMouseDown={(event) => event.target === event.currentTarget && setSelectedTaskId("")}>
          <aside className="assemblyTaskPanel" aria-label="Detalhes da tarefa">
            <header><div><span>{selectedTask.stageTitle}</span><h3>{selectedTask.title}</h3></div><button type="button" onClick={() => setSelectedTaskId("")} aria-label="Fechar detalhes">×</button></header>
            <div className="assemblyTaskPanelBody">
              <section className="assemblyTaskSection">
                <h4>Execução da tarefa</h4>
                <Field label="Título"><input value={taskDraft.title} disabled={!assembly.canManage} onChange={(event) => setTaskDraft((current) => ({ ...current, title: event.target.value }))} /></Field>
                <Field label="Descrição e critério de conclusão"><textarea value={taskDraft.description} disabled={!assembly.canManage} onChange={(event) => setTaskDraft((current) => ({ ...current, description: event.target.value }))} /></Field>
                <div className="assemblyFieldGrid">
                  <Field label="Status"><select value={taskDraft.status} disabled={!assembly.canWork} onChange={(event) => setTaskDraft((current) => ({ ...current, status: event.target.value }))}>{assemblyTaskStatuses.map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
                  <Field label="Prioridade"><select value={taskDraft.priority} disabled={!assembly.canManage} onChange={(event) => setTaskDraft((current) => ({ ...current, priority: event.target.value }))}>{assemblyPriorities.map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select></Field>
                  <Field label="Empresa responsável"><select value={taskDraft.responsibleCompanyId} disabled={!assembly.canManage} onChange={(event) => setTaskDraft((current) => ({ ...current, responsibleCompanyId: event.target.value, responsibleUserId: "" }))}><option value="">Não definida</option>{(assembly.members || []).map((member) => <option value={member.companyId} key={member.companyId}>{member.companyName}</option>)}</select></Field>
                  <Field label="Profissional responsável"><select value={taskDraft.responsibleUserId} disabled={!assembly.canManage} onChange={(event) => setTaskDraft((current) => ({ ...current, responsibleUserId: event.target.value }))}><option value="">Não definido</option>{availableProfessionals.map((professional) => <option value={professional.id} key={professional.id}>{professional.fullName} · {professional.companyName}</option>)}</select></Field>
                  <Field label="Prazo"><input type="date" value={taskDraft.dueDate} min={todayISO} max={openingDate || undefined} disabled={!assembly.canManage || !canSchedule} onChange={(event) => setTaskDraft((current) => ({ ...current, dueDate: event.target.value }))} />{assembly.canManage && !canSchedule && <small>{openingDate ? "A abertura do edital já ocorreu." : "Data de abertura do edital não informada."}</small>}</Field>
                </div>
                <div className="actions">
                  {assembly.canWork && <Button onClick={saveTask} disabled={savingTask || deletingTask}>{savingTask ? "Salvando..." : assembly.canManage ? "Salvar tarefa" : "Atualizar andamento"}</Button>}
                  {assembly.canManage && <Button variant="danger" onClick={deleteTask} disabled={deletingTask || savingTask}>{deletingTask ? "Excluindo..." : "Excluir tarefa"}</Button>}
                </div>
              </section>

              <section className="assemblyTaskSection">
                <h4>Comentários</h4>
                <div className="assemblyComments">{(selectedTask.comments || []).length === 0 ? <p>Nenhum comentário nesta tarefa.</p> : selectedTask.comments.map((item) => <article key={item.id}><strong>{item.userName} · {item.companyName}</strong><p>{item.content}</p><time>{new Date(item.createdAt).toLocaleString("pt-BR")}</time></article>)}</div>
                {assembly.canWork && <div className="assemblyComposer"><textarea value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Registre uma orientação, pendência ou retorno da revisão" maxLength={2000} /><Button onClick={addComment} disabled={savingComment || !comment.trim()}>{savingComment ? "Enviando..." : "Comentar"}</Button></div>}
              </section>

              <section className="assemblyTaskSection">
                <h4>Documentos e evidências</h4>
                <div className="assemblyEvidenceList">{(selectedTask.evidences || []).length === 0 ? <p>Nenhum documento anexado.</p> : selectedTask.evidences.map((item) => <article key={item.id}><div><strong>{item.title}</strong><small>Versão {item.versionNumber} · {item.userName || item.companyName}</small></div>{item.url ? <a href={item.url} target="_blank" rel="noreferrer">Abrir</a> : <span>{item.note}</span>}</article>)}</div>
                {assembly.canWork && <div className="assemblyEvidenceForm">
                  <div className="assemblyFieldGrid">
                    <Field label="Tipo"><select value={evidence.evidenceType} onChange={(event) => setEvidence((current) => ({ ...current, evidenceType: event.target.value }))}><option value="file">Arquivo</option><option value="link">Link externo</option><option value="note">Anotação</option></select></Field>
                    <Field label="Título"><input value={evidence.title} onChange={(event) => setEvidence((current) => ({ ...current, title: event.target.value }))} placeholder="Ex.: CAT do coordenador" /></Field>
                  </div>
                  {evidence.evidenceType === "file" && <Field label="Arquivo"><input type="file" onChange={(event) => setEvidence((current) => ({ ...current, file: event.target.files?.[0] || null }))} /></Field>}
                  {evidence.evidenceType === "link" && <Field label="Endereço do documento"><input type="url" value={evidence.externalUrl} onChange={(event) => setEvidence((current) => ({ ...current, externalUrl: event.target.value }))} placeholder="https://" /></Field>}
                  {evidence.evidenceType === "note" && <Field label="Anotação"><textarea value={evidence.note} onChange={(event) => setEvidence((current) => ({ ...current, note: event.target.value }))} /></Field>}
                  <Button onClick={addEvidence} disabled={savingEvidence}>{savingEvidence ? "Incluindo..." : "Incluir no dossiê"}</Button>
                </div>}
              </section>
            </div>
          </aside>
        </div>
      )}

      {dossierOpen && (
        <div className="assemblyPanelBackdrop" onMouseDown={(event) => event.target === event.currentTarget && setDossierOpen(false)}>
          <aside className="assemblyDossierPanel">
            <header><div><span>Dossiê consolidado</span><h3>Documentos da montagem</h3></div><button type="button" onClick={() => setDossierOpen(false)} aria-label="Fechar dossiê">×</button></header>
            <div className="assemblyDossierSummary"><strong>{dossierItems.length}</strong><span>documentos e evidências reunidos por fase e tarefa</span></div>
            <div className="assemblyDossierBody">
              {stages.map((stage) => {
                const stageItems = (stage.tasks || []).flatMap((task) => (task.evidences || []).map((item) => ({ ...item, taskTitle: task.title })));
                if (!stageItems.length) return null;
                return <section key={stage.id}><h4>{stage.title}</h4>{stageItems.map((item) => <article key={item.id}><div><strong>{item.title}</strong><span>{item.taskTitle}</span><small>{item.companyName} · versão {item.versionNumber}</small></div>{item.url ? <a href={item.url} target="_blank" rel="noreferrer">Abrir documento</a> : <p>{item.note}</p>}</article>)}</section>;
              })}
              {!dossierItems.length && <div className="assemblyEmptyDossier"><strong>O dossiê ainda está vazio</strong><p>Os documentos incluídos nas tarefas aparecerão automaticamente aqui, organizados por fase.</p></div>}
            </div>
          </aside>
        </div>
      )}
    </Page>
  );
}

function TenderTable({ navigate = () => {}, openTenderInterestCompanies = () => {} }) {
  return <Table columns={["Órgão", "Número", "Objeto", "Local", "Abertura", "Valor", "Critério", "Status", "Ações"]} rows={tenders.map((tender) => [
    tender.agency,
    tender.number,
    tender.object,
    tender.location,
    tender.opening,
    tender.value,
    tender.criterion,
    <span className={`statusPill ${tender.status === "Publicado" ? "open" : "review"}`} key={`${tender.number}-status`}>{tender.status}</span>,
    <div className="rowActions compactRowActions" key={`${tender.number}-actions`}>
      <button className="iconButton secondaryIcon" title="Ver detalhe do edital" aria-label="Ver detalhe do edital" onClick={() => navigate("tender-detail")}>{"\u25C9"}</button>
      <button className="iconButton successIcon" title="Marcar interesse no edital" aria-label="Marcar interesse no edital" onClick={() => navigate("tender-interest")}>{"\u2713"}</button>
      <button className="iconButton partnerIcon" title="Ver empresas interessadas neste edital" aria-label="Ver empresas interessadas neste edital" onClick={() => openTenderInterestCompanies(tender.id)}>{"\u2637"}</button>
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
      <span className="badge">Anúncio de parceria</span>
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

