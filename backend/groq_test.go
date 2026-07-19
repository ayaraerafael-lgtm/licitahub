package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestGroqTextIdentifiesProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("rota Groq inesperada: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer groq-test-key" {
			t.Fatalf("chave Groq nao enviada no cabecalho")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "Analise este atestado") {
			t.Fatalf("instrucao nao enviada a Groq")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "groq-response-1",
			"model": "openai/gpt-oss-20b",
			"choices": [{"message": {"content": "Analise Groq concluida."}}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GROQ_API_KEY", "groq-test-key")
	t.Setenv("GROQ_API_BASE_URL", server.URL)
	application := &app{httpClient: server.Client()}

	result, err := application.requestGroqText(context.Background(), "openai/gpt-oss-20b", "Analise este atestado")
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "groq" || result.Model != "openai/gpt-oss-20b" {
		t.Fatalf("provedor ou modelo inesperado: %#v", result)
	}
	if result.Text != "Analise Groq concluida." {
		t.Fatalf("texto inesperado: %s", result.Text)
	}
}
