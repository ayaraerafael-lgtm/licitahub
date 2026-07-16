const baseUrl = (process.env.LICITAHUB_TEST_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
const results = [];

async function check(id, path, expectedStatus, options = {}) {
  let status = 0;
  let detail = "";
  try {
    const response = await fetch(`${baseUrl}${path}`, { redirect: "manual", ...options });
    status = response.status;
    detail = await response.text();
  } catch (error) {
    detail = error.message;
  }

  const passed = expectedStatus.includes(status);
  results.push({ id, status, passed, detail });
  console.log(`${passed ? "APROVADO" : "FALHOU"} | ${id} | HTTP ${status}`);
}

await check("SM-01 Saude da API", "/health", [200]);
await check("SM-02 Frontend compilado", "/", [200]);
await check("SM-03 Sessao sem login protegida", "/api/auth/session", [401]);
await check("SM-04 Login invalido recusado", "/api/auth/login", [401], {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ email: "teste-invalido@licitahub.local", password: "senha-invalida" })
});
await check("SM-05 Editais protegidos sem sessao", "/api/tenders", [401]);
await check("SM-06 Consorcios protegidos sem sessao", "/api/matches", [401]);
await check("SM-07 Montagens protegidas sem sessao", "/api/assemblies/list", [401]);

const failed = results.filter((result) => !result.passed);
if (failed.length > 0) {
  console.error(`\n${failed.length} verificacao(oes) falharam.`);
  process.exitCode = 1;
} else {
  console.log(`\n${results.length} verificacoes aprovadas.`);
}
