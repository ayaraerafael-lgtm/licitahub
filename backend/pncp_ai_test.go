package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParsePNCPCaptureAIResponseValidatesEveryID(t *testing.T) {
	opportunities := []map[string]any{
		{"id": "11111111-1111-1111-1111-111111111111", "objeto": "Supervisao de obras"},
	}
	response := `{"results":[{"id":"11111111-1111-1111-1111-111111111111","classification":"consultiva","confidence":94,"justification":"Supervisao tecnica de obras.","areas":["Supervisao de obras"]}]}`
	result, err := parsePNCPCaptureAIResponse(response, opportunities)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Classification != "consultiva" || result[0].Confidence != 94 {
		t.Fatalf("classificacao inesperada: %#v", result)
	}

	invalid := strings.Replace(response, "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", 1)
	if _, err := parsePNCPCaptureAIResponse(invalid, opportunities); err == nil {
		t.Fatal("uma resposta com id externo ao lote deveria ser recusada")
	}
}

func TestPNCPCaptureAIUsesGeminiWhenOpenAIFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/responses" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded","code":"insufficient_quota"}}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, ":generateContent") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), `"responseMimeType":"application/json"`) || !strings.Contains(string(body), `"responseSchema"`) {
				t.Fatalf("o Gemini nao recebeu a configuracao de JSON estruturado: %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"responseId":   "gemini-pncp-response",
				"modelVersion": "gemini-pncp-001",
				"candidates": []map[string]any{{
					"content": map[string]any{
						"parts": []map[string]string{{
							"text": `{"results":[{"id":"11111111-1111-1111-1111-111111111111","classification":"consultiva","confidence":91,"justification":"O objeto trata de projeto executivo.","areas":["Projetos de engenharia"]}]}`,
						}},
					},
				}},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "openai-test")
	t.Setenv("OPENAI_API_BASE_URL", server.URL)
	t.Setenv("OPENAI_CAPTURE_CLASSIFICATION_MODEL", "openai-test")
	t.Setenv("GEMINI_API_KEY", "gemini-test")
	t.Setenv("GEMINI_API_BASE_URL", server.URL)
	t.Setenv("GEMINI_CAPTURE_CLASSIFICATION_MODEL", "gemini-test")
	application := &app{httpClient: server.Client()}
	opportunities := []map[string]any{{
		"id": "11111111-1111-1111-1111-111111111111", "orgao": "DNIT",
		"numero": "001/2026", "objeto": "Elaboracao de projeto executivo",
	}}

	providerResult, classifications, err := application.requestPNCPCaptureClassification(context.Background(), opportunities)
	if err != nil {
		t.Fatal(err)
	}
	if providerResult.Provider != "gemini" || providerResult.Model != "gemini-pncp-001" {
		t.Fatalf("fallback inesperado: %#v", providerResult)
	}
	if len(classifications) != 1 || classifications[0].Classification != "consultiva" {
		t.Fatalf("classificacao inesperada: %#v", classifications)
	}
}

func TestPNCPCaptureAIUsesGroqWhenOtherProvidersFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/responses":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded","code":"insufficient_quota"}}`))
		case strings.HasSuffix(r.URL.Path, ":generateContent"):
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded","status":"RESOURCE_EXHAUSTED"}}`))
		case r.URL.Path == "/chat/completions":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			requestText := string(body)
			if !strings.Contains(requestText, `"strict":true`) ||
				!strings.Contains(requestText, `"additionalProperties":false`) {
				t.Fatalf("a Groq nao recebeu o esquema estrito esperado: %s", requestText)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "groq-pncp-response",
				"model": "openai/gpt-oss-20b",
				"choices": []map[string]any{{
					"message": map[string]string{
						"content": `{"results":[{"id":"11111111-1111-1111-1111-111111111111","classification":"consultiva","confidence":96,"justification":"O objeto trata de supervisao de obras.","areas":["Supervisao de obras"]}]}`,
					},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "openai-test")
	t.Setenv("OPENAI_API_BASE_URL", server.URL)
	t.Setenv("GEMINI_API_KEY", "gemini-test")
	t.Setenv("GEMINI_API_BASE_URL", server.URL)
	t.Setenv("GROQ_API_KEY", "groq-test")
	t.Setenv("GROQ_API_BASE_URL", server.URL)
	t.Setenv("GROQ_CAPTURE_CLASSIFICATION_MODEL", "openai/gpt-oss-20b")
	application := &app{httpClient: server.Client()}
	opportunities := []map[string]any{{
		"id": "11111111-1111-1111-1111-111111111111", "orgao": "DNIT",
		"numero": "002/2026", "objeto": "Supervisao de obras",
	}}

	providerResult, classifications, err := application.requestPNCPCaptureClassification(context.Background(), opportunities)
	if err != nil {
		t.Fatal(err)
	}
	if providerResult.Provider != "groq" {
		t.Fatalf("fallback usou provedor %q; esperado groq", providerResult.Provider)
	}
	if len(classifications) != 1 || classifications[0].Classification != "consultiva" {
		t.Fatalf("classificacao inesperada: %#v", classifications)
	}
}
