const baseUrl = (process.env.LICITAHUB_TEST_URL || "http://127.0.0.1:8081").replace(/\/$/, "");
const password = process.env.QA_TEST_PASSWORD;
const adminEmail = "qa.automacao@licitahub.local";

if (!password) throw new Error("Defina QA_TEST_PASSWORD antes de executar a homologacao.");

const runId = Date.now().toString();
const results = [];

function record(id, passed, detail = "") {
  results.push({ id, passed, detail });
  console.log(`${passed ? "APROVADO" : "FALHOU"} | ${id}${detail ? ` | ${detail}` : ""}`);
  if (!passed) throw new Error(`${id}: ${detail}`);
}

function cookieFrom(response) {
  const raw = response.headers.get("set-cookie") || "";
  return raw.split(";")[0];
}

async function request(session, path, options = {}) {
  const headers = { ...(options.headers || {}) };
  if (session?.cookie) headers.Cookie = session.cookie;
  if (options.body && !headers["Content-Type"]) headers["Content-Type"] = "application/json";
  const response = await fetch(`${baseUrl}${path}`, { ...options, headers });
  const text = await response.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  return { response, data };
}

async function expectStatus(id, session, path, status, options = {}) {
  const result = await request(session, path, options);
  record(id, result.response.status === status, `HTTP ${result.response.status}${result.data?.error ? `: ${result.data.error}` : ""}`);
  return result.data;
}

async function login(email, userPassword) {
  const { response, data } = await request(null, "/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password: userPassword })
  });
  if (response.status !== 200 || !data?.authenticated || !cookieFrom(response)) {
    throw new Error(`Login nao realizado para ${email}: ${data?.error || "credenciais recusadas"}`);
  }
  return { cookie: cookieFrom(response), user: data.user };
}

function futureDate(days = 180) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toISOString().slice(0, 10);
}

function requirements(prefix) {
  return [
    ["operational_qualification", "fully_meets"],
    ["professional_qualification", "partially_meets"],
    ["technical_proposal", "seeks_partner"],
    ["certifications", "under_review"]
  ].map(([requirementKey, statusKey]) => ({
    requirementKey,
    statusKey,
    whatWeHave: `${prefix}: informacao de teste`,
    whatWeSeek: `${prefix}: complemento de teste`
  }));
}

async function createApprovedCompany(admin, label, sequence) {
  const email = `qa.${label.toLowerCase()}.${runId}@licitahub.local`;
  const cnpj = `99${runId.slice(-8)}${String(sequence).padStart(4, "0")}`.slice(0, 14);
  const invitation = await expectStatus(`EM-${sequence} convite ${label}`, admin, "/api/company-invitations", 201, {
    method: "POST",
    body: JSON.stringify({
      tradeName: `QA ${label} ${runId}`,
      cnpj,
      contactName: `Responsavel ${label}`,
      email,
      phone: "61999990001",
      state: "DF",
      internalNote: "Criado pela homologacao automatizada"
    })
  });
  const accepted = await expectStatus(`EM-${sequence} aceite ${label}`, null, `/api/company-invitations/${invitation.id}/accept`, 201, {
    method: "POST",
    body: JSON.stringify({
      invitationId: invitation.id,
      token: invitation.token,
      website: "https://example.com",
      institutionalDescription: `Empresa ${label} de teste automatizado.`,
      city: "Brasilia",
      state: "DF",
      adminFullName: `Administrador ${label}`,
      adminEmail: email,
      adminPhone: "61999990001",
      adminJobTitle: "Administrador",
      password
    })
  });
  await expectStatus(`EM-${sequence} aprovacao ${label}`, admin, `/api/company-invitations/${invitation.id}/review`, 200, {
    method: "PATCH",
    body: JSON.stringify({ decision: "approved", adjustmentRequest: "", reviewNote: "Aprovado por teste automatizado." })
  });
  const session = await login(email, password);
  record(`EM-${sequence} login ${label}`, session.user.roleKey === "company_admin", session.user.roleKey);
  return { ...accepted, session, email, label };
}

try {
  const health = await request(null, "/health");
  record("SM-01 homologacao disponivel", health.response.status === 200, `HTTP ${health.response.status}`);

  const admin = await login(adminEmail, password);
  record("AC-01 login administrador", admin.user.roleKey === "platform_admin", admin.user.roleKey);
  await expectStatus("AC-02 URL administrativa sem sessao", null, "/api/company-invitations", 401);

  const alpha = await createApprovedCompany(admin, "Alfa", 1);
  const beta = await createApprovedCompany(admin, "Beta", 2);
  const gamma = await createApprovedCompany(admin, "Gama", 3);

  const categories = await expectStatus("RA-01 categorias de noticias", admin, "/api/news/categories", 200);
  const category = categories[0];
  const news = await expectStatus("RA-02 cadastro de noticia", admin, "/api/news", 201, {
    method: "POST",
    body: JSON.stringify({
      title: `Noticia QA ${runId}`,
      categorySlug: category.slug,
      summary: "Resumo criado pelo teste de homologacao.",
      content: "Conteudo completo da noticia de homologacao.",
      status: "published",
      expiresAt: futureDate(30)
    })
  });
  const newsList = await expectStatus("RA-03 listagem de noticias", alpha.session, "/api/news", 200);
  record("RA-04 noticia publicada visivel", newsList.some((item) => item.id === news.id), "noticia criada nao encontrada na lista");

  const postTitle = `Atividade QA ${runId}`;
  await expectStatus("CO-01 publicacao da comunidade", alpha.session, "/api/community/posts", 201, {
    method: "POST",
    body: JSON.stringify({ title: postTitle, categorySlug: "atividades", visibility: "community", content: "Publicacao automatizada da empresa Alfa." })
  });
  const communityPosts = await expectStatus("CO-01A localizacao da publicacao", alpha.session, "/api/community/posts", 200);
  const post = communityPosts.find((item) => item.title === postTitle);
  record("CO-01B identificador da publicacao", Boolean(post?.id), "publicacao recem-criada nao localizada");
  await expectStatus("CO-02 curtida entre empresas", beta.session, `/api/community/posts/${post.id}/like`, 200, { method: "POST" });
  await expectStatus("CO-03 comentario entre empresas", beta.session, `/api/community/posts/${post.id}/comments`, 201, {
    method: "POST", body: JSON.stringify({ content: "Comentario automatizado da empresa Beta." })
  });

  const tenderPayload = {
    agency: "Orgao QA de Homologacao",
    number: `QA-${runId}`,
    object: "Servico de engenharia consultiva para teste automatizado",
    modality: "Concorrencia",
    judgmentCriterion: "Tecnica e preco",
    estimatedValue: String(100000 + Number(runId.slice(-4))),
    state: "DF",
    city: "Brasilia",
    openingDate: futureDate(180),
    status: "published",
    cloudFolderUrl: "https://example.com/documentos",
    confirmValueConflict: true
  };
  const tender = await expectStatus("ED-01 cadastro de edital", admin, "/api/tenders", 201, { method: "POST", body: JSON.stringify(tenderPayload) });
  const alphaTenders = await expectStatus("ED-02 edital visivel para empresa", alpha.session, "/api/tenders", 200);
  record("ED-03 edital publicado encontrado", alphaTenders.some((item) => item.id === tender.id), tender.number);

  for (const company of [alpha, beta]) {
    await expectStatus(`IN-${company.label} interesse com anuncio`, company.session, `/api/tenders/${tender.id}/interests`, 201, {
      method: "POST",
      body: JSON.stringify({
        participationMode: "seeking_partners",
        generalPosition: "interested",
        desiredRole: "seeks_partner",
        publicSummary: `${company.label} busca parceria para compor o edital QA.`,
        internalNote: "Registro criado por homologacao.",
        requirements: requirements(company.label)
      })
    });
  }

  const alphaAds = await expectStatus("IN-03 vitrine de parceiros", alpha.session, "/api/partnership-ads", 200);
  const alphaAd = alphaAds.find((ad) => ad.tenderId === tender.id && ad.companyId === alpha.session.user.companyId);
  const betaAd = alphaAds.find((ad) => ad.tenderId === tender.id && ad.companyId === beta.session.user.companyId);
  record("IN-04 dois anuncios criados", Boolean(alphaAd && betaAd), "anuncios das empresas nao encontrados");

  const chat = await expectStatus("CH-01 abrir conversa de anuncio", alpha.session, "/api/chats", 201, { method: "POST", body: JSON.stringify({ adId: betaAd.id }) });
  await expectStatus("CH-02 enviar mensagem", alpha.session, `/api/chats/${chat.id}/messages`, 201, { method: "POST", body: JSON.stringify({ content: "Mensagem automatizada para avaliacao da parceria." }) });
  const betaMessages = await expectStatus("CH-03 receber mensagem", beta.session, `/api/chats/${chat.id}/messages`, 200);
  record("CH-04 historico da conversa", betaMessages.some((message) => String(message.content || "").includes("Mensagem automatizada")), "mensagem nao encontrada");

  await expectStatus("MA-01 like da Alfa", alpha.session, `/api/partnership-ads/${betaAd.id}/evaluate`, 200, { method: "POST", body: JSON.stringify({ decision: "liked" }) });
  await expectStatus("MA-02 like reciproco da Beta", beta.session, `/api/partnership-ads/${alphaAd.id}/evaluate`, 200, { method: "POST", body: JSON.stringify({ decision: "liked" }) });
  const matches = await expectStatus("MA-03 match criado", alpha.session, "/api/matches", 200);
  const match = matches.find((item) => item.tenderId === tender.id);
  record("MA-04 consorcio visivel", Boolean(match), "match nao encontrado");

  await expectStatus("MA-05 definir lider", alpha.session, `/api/matches/${match.id}/leader`, 200, {
    method: "PUT", body: JSON.stringify({ leadCompanyId: alpha.session.user.companyId, notes: "Lideranca definida pela homologacao." })
  });

  await expectStatus("ED-04 suspender edital", admin, `/api/tenders/${tender.id}`, 200, { method: "PUT", body: JSON.stringify({ ...tenderPayload, status: "suspended" }) });
  const suspendedAds = await expectStatus("ED-05 anuncios ocultos na suspensao", gamma.session, "/api/partnership-ads", 200);
  record("ED-06 anuncio suspenso removido da vitrine", !suspendedAds.some((ad) => ad.tenderId === tender.id), "anuncio ainda visivel");
  await expectStatus("ED-07 republicar edital", admin, `/api/tenders/${tender.id}`, 200, { method: "PUT", body: JSON.stringify({ ...tenderPayload, status: "published" }) });
  const resumedAds = await expectStatus("ED-08 anuncios retomados", gamma.session, "/api/partnership-ads", 200);
  record("ED-09 anuncio restaurado apos retomada", resumedAds.some((ad) => ad.tenderId === tender.id), "anuncio nao voltou para a vitrine");

  await expectStatus("ED-10 pedido de impugnacao", gamma.session, `/api/tenders/${tender.id}/challenge`, 201, {
    method: "POST", body: JSON.stringify({ subject: "Pedido QA", rationale: "Fundamentacao registrada por teste automatizado.", documents: [] })
  });
  const notifications = await expectStatus("CH-05 notificacoes da empresa", beta.session, "/api/notifications", 200);
  record("CH-06 notificacoes registradas", Array.isArray(notifications) && notifications.length > 0, "nenhuma notificacao retornada");
} catch (error) {
  console.error(`\nINTERROMPIDO: ${error.message}`);
  process.exitCode = 1;
}

const failed = results.filter((result) => !result.passed);
console.log(`\nResumo: ${results.length - failed.length} aprovados, ${failed.length} falhos.`);
