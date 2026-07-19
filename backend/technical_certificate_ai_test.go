package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestTechnicalCertificateGeminiIdentifiesProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("chave Gemini nao enviada no cabecalho")
		}
		if !strings.HasSuffix(r.URL.Path, "/models/gemini-test:generateContent") {
			t.Fatalf("rota Gemini inesperada: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "JSON dos atestados") {
			t.Fatalf("instrucao nao enviada ao Gemini")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"responseId": "gemini-response-1",
			"modelVersion": "gemini-test-001",
			"candidates": [{"content": {"parts": [{"text": "Analise concluida."}]}}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_API_BASE_URL", server.URL)
	application := &app{httpClient: server.Client()}

	result, err := application.requestTechnicalCertificateGemini(context.Background(), "gemini-test", "JSON dos atestados")
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "gemini" {
		t.Fatalf("provedor inesperado: %s", result.Provider)
	}
	if result.Model != "gemini-test-001" {
		t.Fatalf("modelo inesperado: %s", result.Model)
	}
	if result.ResponseID != "gemini-response-1" {
		t.Fatalf("identificador inesperado: %s", result.ResponseID)
	}
	if result.Text != "Analise concluida." {
		t.Fatalf("texto inesperado: %s", result.Text)
	}
}

func TestNormalizeTechnicalCertificateAIProvider(t *testing.T) {
	tests := map[string]string{
		"":             "automatic",
		"auto":         "automatic",
		"OpenAI":       "openai",
		"gemini":       "gemini",
		"Groq":         "groq",
		"desconhecido": "",
	}
	for input, expected := range tests {
		if actual := normalizeTechnicalCertificateAIProvider(input); actual != expected {
			t.Fatalf("normalize(%q) = %q; esperado %q", input, actual, expected)
		}
	}
}

func TestRequestTechnicalCertificateAIFallsBackToGemini(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/responses" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded","code":"insufficient_quota"}}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, ":generateContent") {
			_, _ = w.Write([]byte(`{
				"responseId": "fallback-response",
				"modelVersion": "gemini-fallback-001",
				"candidates": [{"content": {"parts": [{"text": "Resposta alternativa."}]}}]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "openai-test")
	t.Setenv("OPENAI_API_BASE_URL", server.URL)
	t.Setenv("OPENAI_TECHNICAL_ANALYSIS_MODEL", "openai-test")
	t.Setenv("GEMINI_API_KEY", "gemini-test")
	t.Setenv("GEMINI_API_BASE_URL", server.URL)
	t.Setenv("GEMINI_TECHNICAL_ANALYSIS_MODEL", "gemini-test")
	application := &app{httpClient: server.Client()}

	result, err := application.requestTechnicalCertificateAI(context.Background(), "automatic", "Analise este JSON")
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "gemini" {
		t.Fatalf("fallback usou provedor %q; esperado gemini", result.Provider)
	}
	if result.Model != "gemini-fallback-001" {
		t.Fatalf("modelo do fallback inesperado: %s", result.Model)
	}
	if result.Text != "Resposta alternativa." {
		t.Fatalf("resposta do fallback inesperada: %s", result.Text)
	}
}
