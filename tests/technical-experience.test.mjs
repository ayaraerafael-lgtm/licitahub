import assert from "node:assert/strict";
import test from "node:test";
import { buildTechnicalExperience, formatExperienceDays } from "../src/technicalExperience.js";

test("desconta a sobreposicao entre atestados do mesmo profissional", () => {
  const [group] = buildTechnicalExperience([
    { id: "a", technicalProfessionalId: "p1", catProfessional: "Ana", executionStart: "2022-01-01", executionEnd: "2022-12-31" },
    { id: "b", technicalProfessionalId: "p1", catProfessional: "Ana", executionStart: "2022-07-01", executionEnd: "2023-06-30" }
  ]);
  assert.equal(group.rawDays, 730);
  assert.equal(group.uniqueDays, 546);
  assert.equal(group.overlapDays, 184);
});

test("nao mistura periodos de profissionais diferentes", () => {
  const groups = buildTechnicalExperience([
    { id: "a", technicalProfessionalId: "p1", catProfessional: "Ana", executionStart: "2020-01-01", executionEnd: "2020-12-31" },
    { id: "b", technicalProfessionalId: "p2", catProfessional: "Bruno", executionStart: "2020-01-01", executionEnd: "2020-12-31" }
  ]);
  assert.equal(groups.length, 2);
  assert.ok(groups.every((group) => group.overlapDays === 0));
});

test("separa atestado sem datas validas do calculo", () => {
  const [group] = buildTechnicalExperience([
    { id: "a", technicalProfessionalId: "p1", catProfessional: "Ana", executionStart: "", executionEnd: "" }
  ]);
  assert.equal(group.certificates.length, 0);
  assert.equal(group.invalidCertificates.length, 1);
  assert.equal(group.uniqueDays, 0);
});

test("formata o total como anos sem perder o total exato em dias", () => {
  const duration = formatExperienceDays(913);
  assert.equal(duration.days, 913);
  assert.ok(duration.decimalYears > 2.49 && duration.decimalYears < 2.51);
  assert.match(duration.label, /2 anos/);
});
