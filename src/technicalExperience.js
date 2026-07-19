const DAY_MS = 24 * 60 * 60 * 1000;
const AVERAGE_YEAR_DAYS = 365.2425;
const AVERAGE_MONTH_DAYS = AVERAGE_YEAR_DAYS / 12;

function parseDateOnly(value) {
  const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (!match) return null;
  const date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
  return Number.isNaN(date.getTime()) ? null : date;
}

function normalizeName(value) {
  return String(value || "").trim().toLocaleLowerCase("pt-BR");
}

function monthKey(date) {
  return date.getUTCFullYear() * 12 + date.getUTCMonth();
}

function buildMonths(start, end) {
  const months = [];
  const first = monthKey(start);
  const last = monthKey(end);
  for (let key = first; key <= last; key += 1) {
    months.push({
      key,
      year: Math.floor(key / 12),
      month: key % 12
    });
  }
  return months;
}

function professionalIdentity(certificate) {
  const id = String(certificate.technicalProfessionalId || "").trim();
  const name = String(certificate.catProfessional || "").trim();
  if (id) return { key: `id:${id}`, name: name || "Profissional sem nome informado" };
  if (name) return { key: `name:${normalizeName(name)}`, name };
  return {
    key: `unlinked:${certificate.id || certificate.certificateNumber || "unknown"}`,
    name: "Profissional não informado"
  };
}

function mergeIntervals(intervals) {
  const merged = [];
  intervals
    .map((item) => ({ start: item.start, end: item.end }))
    .sort((a, b) => a.start - b.start || a.end - b.end)
    .forEach((interval) => {
      const current = merged[merged.length - 1];
      if (!current || interval.start.getTime() > current.end.getTime() + DAY_MS) {
        merged.push({ ...interval });
        return;
      }
      if (interval.end > current.end) current.end = interval.end;
    });
  return merged;
}

function intervalDays(start, end) {
  return Math.floor((end.getTime() - start.getTime()) / DAY_MS) + 1;
}

export function formatExperienceDays(days) {
  const safeDays = Math.max(0, Number(days) || 0);
  const years = Math.floor(safeDays / AVERAGE_YEAR_DAYS);
  const remainingAfterYears = safeDays - years * AVERAGE_YEAR_DAYS;
  const months = Math.floor(remainingAfterYears / AVERAGE_MONTH_DAYS);
  const decimalYears = safeDays / AVERAGE_YEAR_DAYS;
  const parts = [];
  if (years) parts.push(`${years} ${years === 1 ? "ano" : "anos"}`);
  if (months) parts.push(`${months} ${months === 1 ? "mês" : "meses"}`);
  if (!parts.length) parts.push(`${Math.round(safeDays)} ${Math.round(safeDays) === 1 ? "dia" : "dias"}`);
  return {
    label: parts.join(" e "),
    decimalYears,
    days: Math.round(safeDays)
  };
}

export function buildTechnicalExperience(certificates = []) {
  const groups = new Map();
  certificates.forEach((certificate, index) => {
    const identity = professionalIdentity({ ...certificate, id: certificate.id || `certificate-${index}` });
    if (!groups.has(identity.key)) {
      groups.set(identity.key, {
        key: identity.key,
        professionalName: identity.name,
        certificates: [],
        invalidCertificates: []
      });
    }
    const group = groups.get(identity.key);
    const start = parseDateOnly(certificate.executionStart);
    const end = parseDateOnly(certificate.executionEnd);
    const row = {
      ...certificate,
      timelineLabel: certificate.certificateNumber || certificate.fileName || `Atestado ${index + 1}`,
      start,
      end
    };
    if (!start || !end || end < start) {
      group.invalidCertificates.push(row);
      return;
    }
    group.certificates.push({
      ...row,
      rawDays: intervalDays(start, end)
    });
  });

  return [...groups.values()].map((group) => {
    const sorted = group.certificates.sort((a, b) => a.start - b.start || a.end - b.end);
    if (!sorted.length) {
      return {
        ...group,
        months: [],
        mergedIntervals: [],
        rawDays: 0,
        uniqueDays: 0,
        overlapDays: 0,
        occupancy: []
      };
    }
    const mergedIntervals = mergeIntervals(sorted);
    const rawDays = sorted.reduce((total, item) => total + item.rawDays, 0);
    const uniqueDays = mergedIntervals.reduce((total, item) => total + intervalDays(item.start, item.end), 0);
    const months = buildMonths(sorted[0].start, sorted.reduce((latest, item) => item.end > latest ? item.end : latest, sorted[0].end));
    const firstMonthKey = months[0].key;
    const occupancy = months.map(() => 0);
    const timelineCertificates = sorted.map((item) => {
      const startIndex = monthKey(item.start) - firstMonthKey;
      const endIndex = monthKey(item.end) - firstMonthKey;
      for (let index = startIndex; index <= endIndex; index += 1) occupancy[index] += 1;
      return { ...item, startIndex, endIndex };
    });
    return {
      ...group,
      certificates: timelineCertificates,
      months,
      mergedIntervals,
      rawDays,
      uniqueDays,
      overlapDays: Math.max(0, rawDays - uniqueDays),
      occupancy
    };
  }).sort((a, b) => a.professionalName.localeCompare(b.professionalName, "pt-BR"));
}
