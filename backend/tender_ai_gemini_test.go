package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRequestTenderAnalysisFallsBackToGeminiWithDocuments(t *testing.T) {
	var server *httptest.Server
	var mutex sync.Mutex
	uploadedDocument := false
	deletedDocument := false

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/responses":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded","code":"insufficient_quota"}}`))
		case r.URL.Path == "/upload/v1beta/files":
			if r.Header.Get("X-Goog-Upload-Command") != "start" {
				t.Fatalf("comando inicial de upload inesperado")
			}
			w.Header().Set("X-Goog-Upload-URL", server.URL+"/upload-session/1")
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/upload-session/1":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			mutex.Lock()
			uploadedDocument = string(body) == "conteudo do edital"
			mutex.Unlock()
			_, _ = w.Write([]byte(`{"file":{"name":"files/tender-test","uri":"gemini://files/tender-test","mimeType":"application/pdf","state":"ACTIVE"}}`))
		case strings.HasSuffix(r.URL.Path, "/models/gemini-test:generateContent"):
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			requestBody := string(body)
			if !strings.Contains(requestBody, "gemini://files/tender-test") || !strings.Contains(requestBody, "Entregue somente HTML") {
				t.Fatalf("prompt ou arquivo nao foi enviado ao Gemini: %s", requestBody)
			}
			_, _ = w.Write([]byte(`{
				"responseId":"gemini-tender-response",
				"modelVersion":"gemini-test-001",
				"candidates":[{"content":{"parts":[{"text":"<!doctype html><html><body>Analise</body></html>"}]}}]
			}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/files/tender-test":
			mutex.Lock()
			deletedDocument = true
			mutex.Unlock()
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("OPENAI_API_KEY", "openai-test")
	t.Setenv("OPENAI_API_BASE_URL", server.URL)
	t.Setenv("OPENAI_ANALYSIS_MODEL", "openai-test")
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("GEMINI_API_BASE_URL", server.URL)
	t.Setenv("GEMINI_UPLOAD_BASE_URL", server.URL+"/upload/v1beta")
	t.Setenv("GEMINI_TENDER_ANALYSIS_MODEL", "gemini-test")
	application := &app{httpClient: server.Client()}
	content := []map[string]any{
		{"type": "input_text", "text": "Entregue somente HTML"},
		{"type": "input_file", "filename": "edital.pdf", "file_data": "data:application/pdf;base64,Y29udGV1ZG8gZG8gZWRpdGFs"},
	}

	result, err := application.requestTenderAnalysis(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "gemini" || result.Model != "gemini-test-001" {
		t.Fatalf("fallback inesperado: provedor=%s modelo=%s", result.Provider, result.Model)
	}
	if result.ResponseID != "gemini-tender-response" || !strings.Contains(result.HTML, "<html>") {
		t.Fatalf("resposta Gemini inesperada: %#v", result)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !uploadedDocument {
		t.Fatal("o documento original nao foi enviado ao Gemini")
	}
	if !deletedDocument {
		t.Fatal("o arquivo temporario nao foi removido do Gemini")
	}
}
