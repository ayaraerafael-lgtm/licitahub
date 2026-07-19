package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type tenderGeminiFile struct {
	Name     string `json:"name"`
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	State    string `json:"state"`
}

func tenderAIInitialProvider() (string, string) {
	if technicalCertificateAIProviderConfigured("openai") {
		return "openai", tenderAIOpenAIModel()
	}
	return "gemini", tenderAIGeminiModel()
}

func tenderAIOpenAIModel() string {
	return strings.TrimSpace(getenv("OPENAI_ANALYSIS_MODEL", "gpt-5.6"))
}

func tenderAIGeminiModel() string {
	return strings.TrimSpace(getenv("GEMINI_TENDER_ANALYSIS_MODEL", getenv("GEMINI_TECHNICAL_ANALYSIS_MODEL", getenv("GEMINI_MODEL", "gemini-3.5-flash"))))
}

func (a *app) requestTenderAnalysis(ctx context.Context, content []map[string]any) (tenderAIAnalysisResult, error) {
	failures := make([]string, 0, 2)
	if technicalCertificateAIProviderConfigured("openai") {
		result, err := a.requestTenderAnalysisOpenAI(ctx, tenderAIOpenAIModel(), content)
		if err == nil {
			return result, nil
		}
		failures = append(failures, "OpenAI: "+err.Error())
	}
	if technicalCertificateAIProviderConfigured("gemini") {
		result, err := a.requestTenderAnalysisGemini(ctx, tenderAIGeminiModel(), content)
		if err == nil {
			return result, nil
		}
		failures = append(failures, "Google Gemini: "+err.Error())
	}
	if len(failures) == 0 {
		return tenderAIAnalysisResult{}, errors.New("nenhum provedor de IA esta configurado nesta instalacao")
	}
	return tenderAIAnalysisResult{}, errors.New(strings.Join(failures, " | "))
}

func (a *app) requestTenderAnalysisGemini(ctx context.Context, model string, content []map[string]any) (tenderAIAnalysisResult, error) {
	parts := make([]map[string]any, 0, len(content))
	uploaded := make([]tenderGeminiFile, 0, len(content))
	defer func() {
		for _, file := range uploaded {
			a.deleteTenderGeminiFile(file.Name)
		}
	}()

	for _, item := range content {
		switch itemType, _ := item["type"].(string); itemType {
		case "input_text":
			text, _ := item["text"].(string)
			if strings.TrimSpace(text) != "" {
				parts = append(parts, map[string]any{"text": text})
			}
		case "input_file":
			filename, _ := item["filename"].(string)
			dataURL, _ := item["file_data"].(string)
			mimeType, fileBytes, err := decodeTenderAIDataURL(dataURL)
			if err != nil {
				return tenderAIAnalysisResult{}, fmt.Errorf("nao foi possivel preparar o documento %q para o Gemini", filename)
			}
			file, err := a.uploadTenderGeminiFile(ctx, filename, mimeType, fileBytes)
			if err != nil {
				return tenderAIAnalysisResult{}, fmt.Errorf("nao foi possivel enviar o documento %q ao Gemini: %w", filename, err)
			}
			uploaded = append(uploaded, file)
			file, err = a.waitTenderGeminiFile(ctx, file)
			if err != nil {
				return tenderAIAnalysisResult{}, fmt.Errorf("o Gemini nao conseguiu preparar o documento %q: %w", filename, err)
			}
			parts = append(parts, map[string]any{
				"file_data": map[string]any{
					"mime_type": file.MimeType,
					"file_uri":  file.URI,
				},
			})
		}
	}
	if len(parts) < 2 {
		return tenderAIAnalysisResult{}, errors.New("nenhum documento valido foi preparado para a analise")
	}

	body, err := json.Marshal(map[string]any{
		"contents": []map[string]any{{
			"role":  "user",
			"parts": parts,
		}},
		"generationConfig": map[string]any{"maxOutputTokens": 32000},
	})
	if err != nil {
		return tenderAIAnalysisResult{}, errors.New("nao foi possivel preparar o pedido")
	}
	baseURL := strings.TrimRight(getenv("GEMINI_API_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/models/"+model+":generateContent", bytes.NewReader(body))
	if err != nil {
		return tenderAIAnalysisResult{}, errors.New("nao foi possivel criar o pedido")
	}
	req.Header.Set("x-goog-api-key", os.Getenv("GEMINI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	response, err := a.httpClient.Do(req)
	if err != nil {
		return tenderAIAnalysisResult{}, errors.New("nao foi possivel comunicar com a API; tente novamente")
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 16*1024*1024))
	if err != nil {
		return tenderAIAnalysisResult{}, errors.New("nao foi possivel ler a resposta")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return tenderAIAnalysisResult{}, technicalCertificateAIGeminiError(response.StatusCode, responseBody)
	}
	var parsed struct {
		ResponseID   string `json:"responseId"`
		ModelVersion string `json:"modelVersion"`
		Candidates   []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return tenderAIAnalysisResult{}, errors.New("a API retornou uma resposta invalida")
	}
	var result strings.Builder
	for _, candidate := range parsed.Candidates {
		for _, part := range candidate.Content.Parts {
			result.WriteString(part.Text)
		}
		if strings.TrimSpace(result.String()) != "" {
			break
		}
	}
	if strings.TrimSpace(result.String()) == "" {
		return tenderAIAnalysisResult{}, errors.New("a API nao retornou o HTML da analise")
	}
	resolvedModel := model
	if strings.TrimSpace(parsed.ModelVersion) != "" {
		resolvedModel = strings.TrimSpace(parsed.ModelVersion)
	}
	return tenderAIAnalysisResult{
		HTML:       strings.TrimSpace(result.String()),
		ResponseID: parsed.ResponseID,
		Provider:   "gemini",
		Model:      resolvedModel,
	}, nil
}

func decodeTenderAIDataURL(dataURL string) (string, []byte, error) {
	header, encoded, ok := strings.Cut(dataURL, ",")
	if !ok || !strings.HasPrefix(header, "data:") || !strings.Contains(header, ";base64") {
		return "", nil, errors.New("arquivo invalido")
	}
	mimeType := strings.TrimSuffix(strings.TrimPrefix(header, "data:"), ";base64")
	fileBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(fileBytes) == 0 {
		return "", nil, errors.New("arquivo invalido")
	}
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType, fileBytes, nil
}

func (a *app) uploadTenderGeminiFile(ctx context.Context, filename, mimeType string, contents []byte) (tenderGeminiFile, error) {
	metadata, _ := json.Marshal(map[string]any{"file": map[string]any{"display_name": filename}})
	uploadBaseURL := strings.TrimRight(getenv("GEMINI_UPLOAD_BASE_URL", "https://generativelanguage.googleapis.com/upload/v1beta"), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadBaseURL+"/files", bytes.NewReader(metadata))
	if err != nil {
		return tenderGeminiFile{}, errors.New("nao foi possivel iniciar o envio")
	}
	req.Header.Set("x-goog-api-key", os.Getenv("GEMINI_API_KEY"))
	req.Header.Set("X-Goog-Upload-Protocol", "resumable")
	req.Header.Set("X-Goog-Upload-Command", "start")
	req.Header.Set("X-Goog-Upload-Header-Content-Length", strconv.Itoa(len(contents)))
	req.Header.Set("X-Goog-Upload-Header-Content-Type", mimeType)
	req.Header.Set("Content-Type", "application/json")
	response, err := a.httpClient.Do(req)
	if err != nil {
		return tenderGeminiFile{}, errors.New("nao foi possivel comunicar com o servico de arquivos")
	}
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return tenderGeminiFile{}, technicalCertificateAIGeminiError(response.StatusCode, responseBody)
	}
	uploadURL := strings.TrimSpace(response.Header.Get("X-Goog-Upload-URL"))
	if uploadURL == "" {
		return tenderGeminiFile{}, errors.New("o servico nao forneceu o endereco de envio")
	}

	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(contents))
	if err != nil {
		return tenderGeminiFile{}, errors.New("nao foi possivel preparar o envio")
	}
	uploadReq.ContentLength = int64(len(contents))
	uploadReq.Header.Set("X-Goog-Upload-Offset", "0")
	uploadReq.Header.Set("X-Goog-Upload-Command", "upload, finalize")
	uploadResponse, err := a.httpClient.Do(uploadReq)
	if err != nil {
		return tenderGeminiFile{}, errors.New("nao foi possivel concluir o envio")
	}
	defer uploadResponse.Body.Close()
	uploadBody, err := io.ReadAll(io.LimitReader(uploadResponse.Body, 2*1024*1024))
	if err != nil {
		return tenderGeminiFile{}, errors.New("nao foi possivel ler o arquivo enviado")
	}
	if uploadResponse.StatusCode < 200 || uploadResponse.StatusCode >= 300 {
		return tenderGeminiFile{}, technicalCertificateAIGeminiError(uploadResponse.StatusCode, uploadBody)
	}
	var uploaded struct {
		File tenderGeminiFile `json:"file"`
	}
	if err := json.Unmarshal(uploadBody, &uploaded); err != nil || uploaded.File.Name == "" || uploaded.File.URI == "" {
		return tenderGeminiFile{}, errors.New("o servico retornou um arquivo invalido")
	}
	if uploaded.File.MimeType == "" {
		uploaded.File.MimeType = mimeType
	}
	return uploaded.File, nil
}

func (a *app) waitTenderGeminiFile(ctx context.Context, file tenderGeminiFile) (tenderGeminiFile, error) {
	for attempts := 0; attempts < 60; attempts++ {
		state := strings.ToUpper(strings.TrimSpace(file.State))
		if state == "" || state == "ACTIVE" {
			return file, nil
		}
		if state == "FAILED" {
			return tenderGeminiFile{}, errors.New("o processamento do arquivo falhou")
		}
		select {
		case <-ctx.Done():
			return tenderGeminiFile{}, errors.New("o processamento do arquivo excedeu o tempo limite")
		case <-time.After(2 * time.Second):
		}
		baseURL := strings.TrimRight(getenv("GEMINI_API_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"), "/")
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/"+strings.TrimLeft(file.Name, "/"), nil)
		if err != nil {
			return tenderGeminiFile{}, errors.New("nao foi possivel consultar o arquivo")
		}
		req.Header.Set("x-goog-api-key", os.Getenv("GEMINI_API_KEY"))
		response, err := a.httpClient.Do(req)
		if err != nil {
			return tenderGeminiFile{}, errors.New("nao foi possivel consultar o arquivo")
		}
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
		response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return tenderGeminiFile{}, technicalCertificateAIGeminiError(response.StatusCode, body)
		}
		if err := json.Unmarshal(body, &file); err != nil {
			return tenderGeminiFile{}, errors.New("o servico retornou um arquivo invalido")
		}
	}
	return tenderGeminiFile{}, errors.New("o processamento do arquivo excedeu o tempo limite")
}

func (a *app) deleteTenderGeminiFile(name string) {
	if strings.TrimSpace(name) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	baseURL := strings.TrimRight(getenv("GEMINI_API_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, baseURL+"/"+strings.TrimLeft(name, "/"), nil)
	if err != nil {
		return
	}
	req.Header.Set("x-goog-api-key", os.Getenv("GEMINI_API_KEY"))
	if response, err := a.httpClient.Do(req); err == nil {
		response.Body.Close()
	}
}
