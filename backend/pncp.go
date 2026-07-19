package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PNCP is an external source. Captures stay in their own queue until an administrator publishes them.
type pncpCaptureSearchRequest struct {
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
	State      string   `json:"state"`
	Source     string   `json:"source"`
	Page       int      `json:"page"`
	Modalities []string `json:"modalities"`
}

type pncpCaptureActionRequest struct {
	Action string `json:"action"`
}

type pncpCandidate struct {
	Source            string
	SourceKey         string
	PNCPControlNumber string
	Agency            string
	Number            string
	Object            string
	Modality          string
	JudgmentCriterion string
	State             string
	City              string
	OpeningDate       string
	EstimatedValue    string
	ExternalURL       string
	RelevanceScore    int
	RelevanceReasons  []string
	Raw               map[string]any
}

func (a *app) handlePNCPCaptures(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode acessar a captacao PNCP")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.listPNCPCaptures(w, r)
	case http.MethodPost:
		a.capturePNCPOpportunities(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handlePNCPCaptureByPath(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode decidir sobre captacoes PNCP")
		return
	}
	if r.URL.Path == "/api/pncp/captures/discard-pending" && r.Method == http.MethodPost {
		a.discardPendingPNCPCaptures(w, r, session)
		return
	}
	id, action := splitResourcePath(r.URL.Path, "/api/pncp/captures/")
	if id == "" || action != "decision" || r.Method != http.MethodPatch {
		writeError(w, http.StatusNotFound, "captacao PNCP nao encontrada")
		return
	}
	var req pncpCaptureActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if req.Action != "prepare" && req.Action != "discard" {
		writeError(w, http.StatusBadRequest, "decisao invalida")
		return
	}
	if req.Action == "discard" {
		a.discardPNCPCapture(w, r, id, session)
		return
	}
	a.preparePNCPCapture(w, r, id, session)
}

func (a *app) listPNCPCaptures(w http.ResponseWriter, r *http.Request) {
	payload, err := a.queryJSON(r.Context(), `
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				pc.id::text AS id, pc.source AS source, pc.source_key AS "sourceKey", COALESCE(pc.pncp_control_number, '') AS "pncpControlNumber", pc.agency, pc.number, pc.object,
				COALESCE(pc.modality, '') AS modality, COALESCE(pc.judgment_criterion, '') AS "judgmentCriterion", COALESCE(pc.state, '') AS state,
				COALESCE(pc.city, '') AS city, pc.opening_date AS "openingDate",
				pc.estimated_value AS "estimatedValue", COALESCE(pc.external_url, '') AS "externalUrl",
				pc.relevance_score AS "relevanceScore", pc.relevance_reasons AS "relevanceReasons",
				pc.status, pc.captured_at AS "capturedAt", pc.approved_at AS "approvedAt",
				pc.published_tender_id::text AS "publishedTenderId",
				ai.classification AS "aiClassification", ai.confidence AS "aiConfidence",
				COALESCE(ai.justification, '') AS "aiJustification", COALESCE(ai.areas, '[]'::jsonb) AS "aiAreas",
				COALESCE(ai.provider, '') AS "aiProvider", COALESCE(ai.model, '') AS "aiModel",
				ai.created_at AS "aiAnalyzedAt"
			FROM pncp_captures pc
			LEFT JOIN LATERAL (
				SELECT classification, confidence, justification, areas, provider, model, created_at
				FROM pncp_capture_ai_analyses
				WHERE capture_id = pc.id
				ORDER BY created_at DESC
				LIMIT 1
			) ai ON true
			ORDER BY CASE pc.status WHEN 'captured' THEN 0 WHEN 'prepared' THEN 1 WHEN 'approved' THEN 2 ELSE 3 END,
				pc.relevance_score DESC, pc.opening_date ASC NULLS LAST, pc.captured_at DESC
		) item;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as captacoes do PNCP")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) capturePNCPOpportunities(w http.ResponseWriter, r *http.Request, session sessionUser) {
	var req pncpCaptureSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if strings.TrimSpace(req.EndDate) == "" {
		req.EndDate = time.Now().Format("2006-01-02")
	}
	if strings.TrimSpace(req.StartDate) == "" {
		req.StartDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "data inicial invalida")
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil || end.Before(start) {
		writeError(w, http.StatusBadRequest, "periodo de captacao invalido")
		return
	}
	if end.Sub(start) > 31*24*time.Hour {
		writeError(w, http.StatusBadRequest, "consulte no maximo 31 dias por vez")
		return
	}

	if req.Source == "" {
		req.Source = "pncp"
	}
	if req.Page < 1 {
		req.Page = 1
	}
	var candidates []pncpCandidate
	switch req.Source {
	case "pncp":
		candidates, err = a.fetchPNCPCandidates(r.Context(), req)
	case "comprasgov":
		candidates, err = a.fetchComprasGovCandidates(r.Context(), req)
	default:
		writeError(w, http.StatusBadRequest, "fonte de captacao invalida")
		return
	}
	if err != nil {
		message := "nao foi possivel consultar o PNCP agora. Tente novamente em alguns minutos."
		if req.Source == "comprasgov" {
			message = "nao foi possivel consultar o Compras.gov.br agora. Tente novamente em alguns minutos."
		}
		if strings.Contains(err.Error(), "PNCP recusou") || strings.Contains(err.Error(), "PNCP limitou") || strings.Contains(err.Error(), "Compras.gov.br recusou") || strings.Contains(err.Error(), "Compras.gov.br limitou") {
			message = err.Error()
		}
		writeError(w, http.StatusBadGateway, message)
		return
	}
	for _, candidate := range candidates {
		if err := a.savePNCPCandidate(r.Context(), candidate); err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel registrar a captacao da fonte consultada")
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"captured": len(candidates), "message": "Consulta concluida. Revise a fila antes de publicar."})
}

func (a *app) fetchPNCPCandidates(ctx context.Context, req pncpCaptureSearchRequest) ([]pncpCandidate, error) {
	// The PNCP consultation endpoint requires a modality code. Keep the initial
	// capture focused on formats that commonly contain technical services and do
	// not make a burst of requests that causes the public source to rate-limit us.
	// Modalidades do PNCP priorizadas pelo LicitaHub para engenharia consultiva:
	// 6 - Mao de obra, 7 - Obras, 8 - Servicos e 9 - Servicos de engenharia.
	modalities := req.Modalities
	if len(modalities) == 0 {
		modalities = []string{"6", "7", "8", "9"}
	}
	resultByKey := map[string]pncpCandidate{}
	// The PNCP consultation service validates dates in the yyyyMMdd format.
	startDate := strings.ReplaceAll(req.StartDate, "-", "")
	endDate := strings.ReplaceAll(req.EndDate, "-", "")
	var firstQueryError error
	for _, modality := range modalities {
		for page := req.Page; page <= req.Page; page++ {
			query := url.Values{}
			query.Set("dataInicial", startDate)
			query.Set("dataFinal", endDate)
			query.Set("codigoModalidadeContratacao", modality)
			query.Set("pagina", strconv.Itoa(page))
			query.Set("tamanhoPagina", "50")
			if state := normalizeState(req.State); state != "" {
				query.Set("uf", state)
			}
			endpoint := "https://pncp.gov.br/api/consulta/v1/contratacoes/publicacao?" + query.Encode()
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return nil, err
			}
			httpReq.Header.Set("Accept", "application/json")
			httpReq.Header.Set("User-Agent", "LicitaHub/1.0 (consulta publica de oportunidades)")
			response, err := doPNCPRequestWithRetry(ctx, a.httpClient, httpReq)
			if err != nil {
				return nil, err
			}
			if response.StatusCode == http.StatusNoContent {
				response.Body.Close()
				break
			}
			if response.StatusCode == http.StatusTooManyRequests {
				response.Body.Close()
				return nil, fmt.Errorf("o PNCP limitou temporariamente as consultas (HTTP 429). Aguarde cerca de um minuto antes de tentar novamente")
			}
			if response.StatusCode < 200 || response.StatusCode > 299 {
				detail, _ := io.ReadAll(io.LimitReader(response.Body, 1200))
				response.Body.Close()
				message := strings.TrimSpace(strings.Join(strings.Fields(string(detail)), " "))
				if len(message) > 500 {
					message = message[:500]
				}
				if firstQueryError == nil {
					if message != "" {
						firstQueryError = fmt.Errorf("o PNCP recusou a consulta da modalidade %s (HTTP %d): %s", modality, response.StatusCode, message)
					} else {
						firstQueryError = fmt.Errorf("o PNCP recusou a consulta da modalidade %s (HTTP %d)", modality, response.StatusCode)
					}
				}
				break
			}
			var payload any
			decodeErr := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&payload)
			response.Body.Close()
			if decodeErr != nil {
				if firstQueryError == nil {
					firstQueryError = decodeErr
				}
				break
			}
			items := pncpItems(payload)
			for _, item := range items {
				candidate := candidateFromPNCP(item)
				if candidate.SourceKey != "" && candidate.Agency != "" && candidate.Number != "" && candidate.Object != "" {
					resultByKey[candidate.SourceKey] = candidate
				}
			}
			if len(items) == 0 || (payloadTotalPages(payload) > 0 && page >= payloadTotalPages(payload)) {
				break
			}
			// A small interval respects the public API and avoids a request burst.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
	}
	candidates := make([]pncpCandidate, 0, len(resultByKey))
	for _, candidate := range resultByKey {
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 && firstQueryError != nil {
		return nil, firstQueryError
	}
	return candidates, nil
}

func pncpItems(payload any) []map[string]any {
	if list, ok := payload.([]any); ok {
		return mapItems(list)
	}
	object, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	for _, key := range []string{"data", "items", "content", "resultado"} {
		if list, ok := object[key].([]any); ok {
			return mapItems(list)
		}
	}
	return nil
}

func doPNCPRequestWithRetry(ctx context.Context, client *http.Client, request *http.Request) (*http.Response, error) {
	retryDelays := []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second}
	for attempt := 0; ; attempt++ {
		response, err := client.Do(request)
		if err != nil || response.StatusCode != http.StatusTooManyRequests || attempt >= len(retryDelays) {
			return response, err
		}

		delay := retryDelays[attempt]
		if seconds, parseErr := strconv.Atoi(strings.TrimSpace(response.Header.Get("Retry-After"))); parseErr == nil && seconds > 0 {
			delay = time.Duration(seconds) * time.Second
		}
		response.Body.Close()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
}

func payloadTotalPages(payload any) int {
	object, ok := payload.(map[string]any)
	if !ok {
		return 0
	}
	for _, key := range []string{"totalPaginas", "totalPages", "numeroPaginas", "pages"} {
		value, ok := object[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case json.Number:
			if parsed, err := strconv.Atoi(string(typed)); err == nil {
				return parsed
			}
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func mapItems(items []any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			result = append(result, object)
		}
	}
	return result
}

func candidateFromPNCP(item map[string]any) pncpCandidate {
	agencyObject := mapValue(item, "orgaoEntidade")
	unitObject := mapValue(item, "unidadeOrgao")
	agency := firstNonEmptyPNCP(stringValue(agencyObject, "razaoSocial"), stringValue(agencyObject, "nome"), stringValue(unitObject, "nomeUnidade"), stringValue(item, "nomeOrgao"))
	number := firstNonEmpty(stringValue(item, "numeroCompra"), stringValue(item, "numero"))
	if year := stringValue(item, "anoCompra"); year != "" && !strings.Contains(number, year) {
		number += "/" + year
	}
	object := firstNonEmptyPNCP(stringValue(item, "objetoCompra"), stringValue(item, "objeto"), stringValue(item, "informacaoComplementar"))
	state := normalizeState(firstNonEmptyPNCP(stringValue(unitObject, "ufSigla"), stringValue(item, "ufSigla")))
	city := firstNonEmptyPNCP(stringValue(unitObject, "municipioNome"), stringValue(item, "municipioNome"))
	sourceKey := firstNonEmptyPNCP(stringValue(item, "numeroControlePNCP"), stringValue(item, "id"))
	if sourceKey == "" {
		sourceKey = strings.Join([]string{agency, number, object}, "|")
	}
	reasons, score := consultantRelevance(agency + " " + object)
	return pncpCandidate{
		Source:    "pncp",
		SourceKey: sourceKey, Agency: agency, Number: number, Object: object,
		PNCPControlNumber: stringValue(item, "numeroControlePNCP"),
		Modality:          firstNonEmptyPNCP(stringValue(item, "modalidadeNome"), stringValue(item, "modalidadeContratacaoNome")),
		JudgmentCriterion: firstNonEmptyPNCP(stringValue(item, "criterioJulgamentoNome"), stringValue(item, "criterioJulgamento"), stringValue(item, "criterioJulgamentoDescricao")),
		State:             state, City: city,
		OpeningDate:    firstNonEmptyPNCP(stringValue(item, "dataAberturaProposta"), stringValue(item, "dataHoraAberturaProposta"), stringValue(item, "dataAbertura"), stringValue(item, "dataHoraAbertura"), stringValue(item, "dataSessao")),
		EstimatedValue: firstNonEmptyPNCP(stringValue(item, "valorTotalEstimado"), stringValue(item, "valorEstimado")),
		ExternalURL:    pncpExternalURL(item, agencyObject),
		RelevanceScore: score, RelevanceReasons: reasons, Raw: item,
	}
}

// fetchComprasGovCandidates consumes the public Dados Abertos API. It remains
// separate from the PNCP client because the services have different limits and
// response contracts, even when some records overlap.
func (a *app) fetchComprasGovCandidates(ctx context.Context, req pncpCaptureSearchRequest) ([]pncpCandidate, error) {
	// A fonte Compras.gov possui sua propria tabela de modalidades; os codigos
	// desta consulta nao devem ser confundidos com os codigos do PNCP acima.
	modalities := []string{"3", "4", "5", "6", "7", "10"}
	resultByKey := map[string]pncpCandidate{}
	var firstQueryError error
	for _, modality := range modalities {
		for page := req.Page; page <= req.Page; page++ {
			query := url.Values{}
			query.Set("pagina", strconv.Itoa(page))
			query.Set("tamanhoPagina", "100")
			query.Set("dataPublicacaoPncpInicial", req.StartDate)
			query.Set("dataPublicacaoPncpFinal", req.EndDate)
			query.Set("codigoModalidade", modality)
			endpoint := "https://dadosabertos.compras.gov.br/modulo-contratacoes/1_consultarContratacoes_PNCP_14133?" + query.Encode()
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return nil, err
			}
			httpReq.Header.Set("Accept", "application/json")
			httpReq.Header.Set("User-Agent", "LicitaHub/1.0 (consulta publica de oportunidades)")
			response, err := doPNCPRequestWithRetry(ctx, a.httpClient, httpReq)
			if err != nil {
				return nil, err
			}
			if response.StatusCode == http.StatusNoContent {
				response.Body.Close()
				break
			}
			if response.StatusCode == http.StatusTooManyRequests {
				response.Body.Close()
				return nil, fmt.Errorf("o Compras.gov.br limitou temporariamente as consultas (HTTP 429). Aguarde antes de tentar novamente")
			}
			if response.StatusCode < 200 || response.StatusCode > 299 {
				detail, _ := io.ReadAll(io.LimitReader(response.Body, 1200))
				response.Body.Close()
				message := strings.TrimSpace(strings.Join(strings.Fields(string(detail)), " "))
				if len(message) > 500 {
					message = message[:500]
				}
				if firstQueryError == nil {
					firstQueryError = fmt.Errorf("o Compras.gov.br recusou a consulta da modalidade %s (HTTP %d): %s", modality, response.StatusCode, message)
				}
				break
			}
			var payload any
			decodeErr := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&payload)
			response.Body.Close()
			if decodeErr != nil {
				if firstQueryError == nil {
					firstQueryError = decodeErr
				}
				break
			}
			items := pncpItems(payload)
			for _, item := range items {
				candidate := candidateFromComprasGov(item)
				if state := normalizeState(req.State); state != "" && candidate.State != "" && candidate.State != state {
					continue
				}
				if candidate.SourceKey != "" && candidate.Agency != "" && candidate.Number != "" && candidate.Object != "" {
					resultByKey[candidate.SourceKey] = candidate
				}
			}
			if len(items) == 0 || (payloadTotalPages(payload) > 0 && page >= payloadTotalPages(payload)) {
				break
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
	}
	candidates := make([]pncpCandidate, 0, len(resultByKey))
	for _, candidate := range resultByKey {
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 && firstQueryError != nil {
		return nil, firstQueryError
	}
	return candidates, nil
}

func candidateFromComprasGov(item map[string]any) pncpCandidate {
	agencyObject := mapValue(item, "orgaoEntidade")
	unitObject := mapValue(item, "unidadeOrgao")
	agency := firstNonEmptyPNCP(stringValue(item, "orgaoEntidadeRazaoSocial"), stringValue(item, "orgaoEntidadeNome"), stringValue(agencyObject, "razaoSocial"), stringValue(agencyObject, "nome"), stringValue(item, "unidadeOrgaoNomeUnidade"), stringValue(unitObject, "nomeUnidade"))
	number := firstNonEmptyPNCP(stringValue(item, "numeroCompra"), stringValue(item, "numero"))
	if year := firstNonEmptyPNCP(stringValue(item, "anoCompra"), stringValue(item, "ano")); year != "" && !strings.Contains(number, year) {
		number += "/" + year
	}
	object := firstNonEmptyPNCP(stringValue(item, "objetoCompra"), stringValue(item, "objeto"), stringValue(item, "informacaoComplementar"))
	state := normalizeState(firstNonEmptyPNCP(stringValue(item, "unidadeOrgaoUfSigla"), stringValue(item, "ufSigla"), stringValue(unitObject, "ufSigla")))
	city := firstNonEmptyPNCP(stringValue(item, "unidadeOrgaoMunicipioNome"), stringValue(item, "municipioNome"), stringValue(unitObject, "municipioNome"))
	controlNumber := stringValue(item, "numeroControlePNCP")
	externalID := firstNonEmptyPNCP(stringValue(item, "idCompra"), controlNumber, stringValue(item, "id"))
	sourceKey := "comprasgov:" + externalID
	if externalID == "" {
		sourceKey = "comprasgov:" + strings.Join([]string{agency, number, object}, "|")
	}
	reasons, score := consultantRelevance(agency + " " + object)
	return pncpCandidate{Source: "comprasgov", SourceKey: sourceKey, PNCPControlNumber: controlNumber, Agency: agency, Number: number, Object: object, Modality: firstNonEmptyPNCP(stringValue(item, "modalidadeNome"), stringValue(item, "modalidadeContratacaoNome")), JudgmentCriterion: firstNonEmptyPNCP(stringValue(item, "criterioJulgamentoNome"), stringValue(item, "criterioJulgamento"), stringValue(item, "criterioJulgamentoDescricao")), State: state, City: city, OpeningDate: firstNonEmptyPNCP(stringValue(item, "dataAberturaProposta"), stringValue(item, "dataHoraAberturaProposta"), stringValue(item, "dataAbertura"), stringValue(item, "dataHoraAbertura"), stringValue(item, "dataSessao")), EstimatedValue: firstNonEmptyPNCP(stringValue(item, "valorTotalEstimado"), stringValue(item, "valorEstimado")), ExternalURL: comprasGovExternalURL(externalID), RelevanceScore: score, RelevanceReasons: reasons, Raw: item}
}

func comprasGovExternalURL(id string) string {
	if strings.TrimSpace(id) != "" {
		return "https://cnetmobile.estaleiro.serpro.gov.br/comprasnet-web/public/compras/acompanhamento-compra?compra=" + url.QueryEscape(strings.TrimSpace(id))
	}
	return "https://www.gov.br/compras/pt-br"
}

func pncpExternalURL(item map[string]any, agency map[string]any) string {
	cnpj := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, stringValue(agency, "cnpj"))
	year := strings.TrimSpace(stringValue(item, "anoCompra"))
	sequence := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, stringValue(item, "sequencialCompra"))
	if sequenceNumber, err := strconv.Atoi(sequence); len(cnpj) == 14 && len(year) == 4 && err == nil {
		return fmt.Sprintf("https://pncp.gov.br/app/editais/%s/%s/%06d", cnpj, year, sequenceNumber)
	}
	return "https://pncp.gov.br/app/editais"
}

func firstNonEmptyPNCP(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mapValue(value map[string]any, key string) map[string]any {
	if nested, ok := value[key].(map[string]any); ok {
		return nested
	}
	return map[string]any{}
}

func stringValue(value map[string]any, key string) string {
	item, ok := value[key]
	if !ok || item == nil {
		return ""
	}
	switch typed := item.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func consultantRelevance(value string) ([]string, int) {
	text := normalizeConsultantRelevanceText(value)
	keywords := []struct{ word, label string }{
		{"engenharia consultiva", "engenharia consultiva"},
		{"consultoria de engenharia", "consultoria de engenharia"},
		{"servicos tecnicos especializados de engenharia", "servicos tecnicos especializados de engenharia"},
		{"apoio tecnico especializado", "apoio tecnico especializado"},
		{"estudos e projetos", "estudos e projetos"},
		{"estudos de viabilidade", "estudos de viabilidade"},
		{"evtea", "EVTEA"}, {"oae", "OAE"}, {"bim", "BIM"}, {"eia", "EIA"},
		{"rima", "RIMA"}, {"pba", "PBA"}, {"prad", "PRAD"}, {"pae", "PAE"},
		{"pmo", "PMO"}, {"pnp", "PNP"}, {"pncv", "PNCV"}, {"pmv", "PMV"},
		{"ppd", "PPD"}, {"aet", "AET"},
		{"projeto basico", "projeto basico"},
		{"projeto executivo", "projeto executivo"},
		{"anteprojeto", "anteprojeto"},
		{"supervisao de obras", "supervisao de obras"},
		{"fiscalizacao de obras", "fiscalizacao de obras"},
		{"gerenciamento de obras", "gerenciamento de obras"},
		{"gerenciamento de contratos", "gerenciamento de contratos"},
		{"controle tecnologico", "controle tecnologico"},
		{"auditoria de engenharia", "auditoria de engenharia"},
		{"inspecao de oae", "inspecao de OAE"},
		{"monitoramento estrutural", "monitoramento estrutural"},
		{"gestao ambiental", "gestao ambiental"},
		{"supervisao ambiental", "supervisao ambiental"},
		{"levantamento topografico", "levantamento topografico"},
		{"geoprocessamento", "geoprocessamento"},
		{"engenharia rodoviaria", "engenharia rodoviaria"},
		{"supervisao rodoviaria", "supervisao rodoviaria"},
		{"operacao rodoviaria", "operacao rodoviaria"},
		{"seguranca viaria", "seguranca viaria"},
		{"engenharia de trafego", "engenharia de trafego"},
		{"gestao de pavimentos", "gestao de pavimentos"},
		{"mobilidade urbana", "mobilidade urbana"},
		{"saneamento", "saneamento"},
		{"drenagem", "drenagem"},
		{"recursos hidricos", "recursos hidricos"},
		{"apoio a fiscalizacao", "apoio a fiscalizacao"},
		{"apoio ao gerenciamento", "apoio ao gerenciamento"},
		{"elaboracao de projeto", "elaboracao de projeto"},
		{"contratacao integrada", "contratacao integrada"},
		{"supervis", "supervisao"}, {"fiscaliza", "fiscalizacao"},
		{"projeto", "projetos"}, {"estudo", "estudos"},
		{"licenciamento ambiental", "meio ambiente"}, {"ambiental", "meio ambiente"},
		{"arqueolog", "arqueologia"}, {"trabalho social", "projetos sociais"},
		{"rodovia", "infraestrutura viaria"}, {"ferrovia", "infraestrutura ferroviaria"},
		{"mobilidade", "mobilidade"}, {"geotec", "geotecnia"},
	}
	exclusions := []string{"aquisicao de material", "generos alimenticios", "merenda", "combustivel", "mobiliario", "limpeza", "vigilancia armada"}
	reasons := []string{}
	for _, keyword := range keywords {
		if containsConsultantTerm(text, keyword.word, len(keyword.word) <= 4) {
			reasons = append(reasons, keyword.label)
		}
	}
	for _, excluded := range exclusions {
		if strings.Contains(text, excluded) {
			return []string{"possivel item fora do foco: " + excluded}, 0
		}
	}
	if len(reasons) == 0 {
		return []string{"objeto sem termos tecnicos reconhecidos"}, 10
	}
	score := 45 + len(reasons)*12
	if score > 95 {
		score = 95
	}
	return reasons, score
}

func containsConsultantTerm(text, term string, wholeWord bool) bool {
	if !wholeWord {
		return strings.Contains(text, term)
	}
	start := 0
	for start < len(text) {
		relative := strings.Index(text[start:], term)
		if relative < 0 {
			return false
		}
		index := start + relative
		beforeIsWord := index > 0 && isConsultantWordByte(text[index-1])
		afterIndex := index + len(term)
		afterIsWord := afterIndex < len(text) && isConsultantWordByte(text[afterIndex])
		if !beforeIsWord && !afterIsWord {
			return true
		}
		start = afterIndex
	}
	return false
}

func isConsultantWordByte(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= '0' && value <= '9') || value == '_'
}

func normalizeConsultantRelevanceText(value string) string {
	text := strings.ToLower(value)
	return strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ç", "c",
	).Replace(text)
}

func (a *app) savePNCPCandidate(ctx context.Context, candidate pncpCandidate) error {
	raw, _ := json.Marshal(candidate.Raw)
	reasons, _ := json.Marshal(candidate.RelevanceReasons)
	openingSQL := "NULL"
	if parsed, ok := parsePNCPDate(candidate.OpeningDate); ok {
		openingSQL = sqlQuote(parsed.Format(time.RFC3339)) + "::timestamptz"
	}
	valueSQL := "NULL"
	if normalized := normalizePNCPNumber(candidate.EstimatedValue); normalized != "" {
		valueSQL = sqlQuote(normalized) + "::numeric"
	}
	// When the PNCP control number exists, merge the two public sources into
	// the same queue item. PNCP values are authoritative; Compras.gov.br only
	// fills fields that are still empty.
	if candidate.PNCPControlNumber != "" {
		preferIncoming := candidate.Source == "pncp"
		field := func(column, value string) string {
			if preferIncoming {
				return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s)", sqlQuote(value), column)
			}
			return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s)", column, sqlQuote(value))
		}
		openingExpr := "opening_date"
		if openingSQL != "NULL" {
			if preferIncoming {
				openingExpr = fmt.Sprintf("COALESCE(%s, opening_date)", openingSQL)
			} else {
				openingExpr = fmt.Sprintf("COALESCE(opening_date, %s)", openingSQL)
			}
		}
		valueExpr := "estimated_value"
		if valueSQL != "NULL" {
			if preferIncoming {
				valueExpr = fmt.Sprintf("COALESCE(%s, estimated_value)", valueSQL)
			} else {
				valueExpr = fmt.Sprintf("COALESCE(estimated_value, %s)", valueSQL)
			}
		}
		identityWhere := fmt.Sprintf("pncp_control_number = %s", sqlQuote(candidate.PNCPControlNumber))
		if candidate.PNCPControlNumber == "" && valueSQL != "NULL" {
			identityWhere = fmt.Sprintf("lower(trim(agency)) = lower(trim(%s)) AND lower(trim(number)) = lower(trim(%s)) AND estimated_value = %s AND source <> %s", sqlQuote(candidate.Agency), sqlQuote(candidate.Number), valueSQL, sqlQuote(candidate.Source))
		}
		mergedSource := "CASE WHEN source = 'pncp' AND " + sqlQuote(candidate.Source) + " = 'comprasgov' THEN 'pncp+comprasgov' WHEN source = 'comprasgov' AND " + sqlQuote(candidate.Source) + " = 'pncp' THEN 'pncp+comprasgov' ELSE source END"
		updated, updateErr := a.queryJSON(ctx, fmt.Sprintf(`
			UPDATE pncp_captures
			SET source = %s, agency = %s, number = %s, object = %s,
				modality = %s, judgment_criterion = %s, state = %s, city = %s,
				opening_date = %s, estimated_value = %s, external_url = COALESCE(NULLIF(external_url, ''), %s),
				relevance_score = GREATEST(relevance_score, %d), relevance_reasons = %s::jsonb,
				raw_payload = raw_payload || jsonb_build_object(%s, %s::jsonb), pncp_control_number = %s, updated_at = now()
			WHERE %s
			RETURNING id;
		`, mergedSource, field("agency", candidate.Agency), field("number", candidate.Number), field("object", candidate.Object),
			field("modality", candidate.Modality), field("judgment_criterion", candidate.JudgmentCriterion), field("state", candidate.State), field("city", candidate.City),
			openingExpr, valueExpr, sqlQuote(candidate.ExternalURL), candidate.RelevanceScore, sqlQuote(string(reasons)), sqlQuote(candidate.Source), sqlQuote(string(raw)), sqlQuote(candidate.PNCPControlNumber), identityWhere))
		if updateErr != nil {
			return updateErr
		}
		if strings.TrimSpace(string(updated)) != "null" && strings.TrimSpace(string(updated)) != "[]" {
			return nil
		}
	}

	_, err := a.runPSQL(ctx, fmt.Sprintf(`
		INSERT INTO pncp_captures (
			source, source_key, pncp_control_number, agency, number, object, modality, judgment_criterion, state, city, opening_date,
			estimated_value, external_url, relevance_score, relevance_reasons, raw_payload
		) VALUES (
			%s, %s, NULLIF(%s, ''), %s, %s, %s, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), %s,
			%s, NULLIF(%s, ''), %d, %s::jsonb, %s::jsonb
		)
		ON CONFLICT (source_key) DO UPDATE SET
			agency = COALESCE(NULLIF(pncp_captures.agency, ''), EXCLUDED.agency), number = COALESCE(NULLIF(pncp_captures.number, ''), EXCLUDED.number), object = COALESCE(NULLIF(pncp_captures.object, ''), EXCLUDED.object),
			pncp_control_number = COALESCE(pncp_captures.pncp_control_number, EXCLUDED.pncp_control_number), modality = COALESCE(NULLIF(pncp_captures.modality, ''), EXCLUDED.modality), state = COALESCE(NULLIF(pncp_captures.state, ''), EXCLUDED.state), city = COALESCE(NULLIF(pncp_captures.city, ''), EXCLUDED.city),
			judgment_criterion = COALESCE(NULLIF(pncp_captures.judgment_criterion, ''), EXCLUDED.judgment_criterion), opening_date = COALESCE(pncp_captures.opening_date, EXCLUDED.opening_date), estimated_value = COALESCE(pncp_captures.estimated_value, EXCLUDED.estimated_value),
			external_url = COALESCE(NULLIF(pncp_captures.external_url, ''), EXCLUDED.external_url), relevance_score = GREATEST(pncp_captures.relevance_score, EXCLUDED.relevance_score),
			relevance_reasons = EXCLUDED.relevance_reasons, raw_payload = pncp_captures.raw_payload || jsonb_build_object(EXCLUDED.source, EXCLUDED.raw_payload),
			updated_at = now();
	`, sqlQuote(candidate.Source), sqlQuote(candidate.SourceKey), sqlQuote(candidate.PNCPControlNumber), sqlQuote(candidate.Agency), sqlQuote(candidate.Number), sqlQuote(candidate.Object),
		sqlQuote(candidate.Modality), sqlQuote(candidate.JudgmentCriterion), sqlQuote(candidate.State), sqlQuote(candidate.City), openingSQL, valueSQL,
		sqlQuote(candidate.ExternalURL), candidate.RelevanceScore, sqlQuote(string(reasons)), sqlQuote(string(raw))))
	return err
}

func parsePNCPDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000", "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func normalizePNCPNumber(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, ".", ""))
	value = strings.ReplaceAll(value, ",", ".")
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return ""
	}
	return value
}

func (a *app) discardPNCPCapture(w http.ResponseWriter, r *http.Request, captureID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		DELETE FROM pncp_captures
		WHERE id = %s::uuid AND status = 'captured'
		RETURNING row_to_json(pncp_captures);
	`, sqlQuote(captureID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusBadRequest, "esta captacao nao pode mais ser descartada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) discardPendingPNCPCaptures(w http.ResponseWriter, r *http.Request, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), `
		WITH deleted AS (
			DELETE FROM pncp_captures
			WHERE status IN ('captured', 'discarded')
			RETURNING id
		)
		SELECT json_build_object('discarded', (SELECT count(*) FROM deleted));
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel descartar as captacoes pendentes")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) preparePNCPCapture(w http.ResponseWriter, r *http.Request, captureID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_capture AS (
			SELECT * FROM pncp_captures WHERE id = %s::uuid AND status = 'captured' FOR UPDATE
		), existing_tender AS (
			SELECT id FROM tenders
			WHERE source = (SELECT source FROM selected_capture) AND source_reference = (SELECT source_key FROM selected_capture)
			  AND deleted_at IS NULL
			LIMIT 1
		), inserted_tender AS (
			INSERT INTO tenders (
				agency, number, object, modality, judgment_criterion, estimated_value, state, city, opening_date,
				status, cloud_folder_url, source, source_reference, created_by_user_id
			)
			SELECT agency, number, object, modality, judgment_criterion, estimated_value, state, city, opening_date,
				'draft', external_url, source, source_key, %s::uuid
			FROM selected_capture
			WHERE NOT EXISTS (SELECT 1 FROM existing_tender)
			RETURNING id
		), chosen_tender AS (
			SELECT id FROM inserted_tender
			UNION ALL
			SELECT id FROM existing_tender
			LIMIT 1
		), updated_capture AS (
			UPDATE pncp_captures
			SET status = 'prepared', approved_by_user_id = %s::uuid,
				published_tender_id = (SELECT id FROM chosen_tender), updated_at = now()
			WHERE id = %s::uuid
			  AND EXISTS (SELECT 1 FROM chosen_tender)
			RETURNING published_tender_id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT uc.published_tender_id::text AS "tenderId", sc.number, sc.agency
			FROM updated_capture uc JOIN selected_capture sc ON true
		) item;
	`, sqlQuote(captureID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(captureID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusBadRequest, "esta captacao nao pode ser preparada agora")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}
