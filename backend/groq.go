package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type groqChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (a *app) requestGroqText(ctx context.Context, model, input string) (technicalCertificateAIResult, error) {
	return a.requestGroq(ctx, model, input, nil)
}

func (a *app) requestGroq(ctx context.Context, model, input string, responseFormat map[string]any) (technicalCertificateAIResult, error) {
	requestBody := map[string]any{
		"model": model,
		"messages": []map[string]string{{
			"role": "user", "content": input,
		}},
		"temperature":           0.1,
		"max_completion_tokens": 16000,
	}
	if responseFormat != nil {
		requestBody["response_format"] = responseFormat
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return technicalCertificateAIResult{}, errors.New("nao foi possivel preparar o pedido")
	}
	baseURL := strings.TrimRight(getenv("GROQ_API_BASE_URL", "https://api.groq.com/openai/v1"), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return technicalCertificateAIResult{}, errors.New("nao foi possivel criar o pedido")
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	response, err := a.httpClient.Do(req)
	if err != nil {
		return technicalCertificateAIResult{}, errors.New("nao foi possivel comunicar com a API; tente novamente")
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 12*1024*1024))
	if err != nil {
		return technicalCertificateAIResult{}, errors.New("nao foi possivel ler a resposta")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return technicalCertificateAIResult{}, groqAPIError(response.StatusCode, responseBody)
	}
	var parsed groqChatResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return technicalCertificateAIResult{}, errors.New("a API retornou uma resposta invalida")
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return technicalCertificateAIResult{}, errors.New("a API nao retornou uma analise")
	}
	resolvedModel := strings.TrimSpace(parsed.Model)
	if resolvedModel == "" {
		resolvedModel = model
	}
	return technicalCertificateAIResult{
		Text:       strings.TrimSpace(parsed.Choices[0].Message.Content),
		ResponseID: parsed.ID,
		Provider:   "groq",
		Model:      resolvedModel,
	}, nil
}

func groqAPIError(status int, payload []byte) error {
	var providerError struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(payload, &providerError)
	details := strings.ToLower(strings.Join([]string{providerError.Error.Message, providerError.Error.Type, providerError.Error.Code}, " "))
	if status == http.StatusTooManyRequests {
		if strings.Contains(details, "quota") || strings.Contains(details, "billing") || strings.Contains(details, "credit") || strings.Contains(details, "limit") {
			return errors.New("a cota gratuita ou o limite da conta foi atingido")
		}
		return errors.New("a conta atingiu um limite temporario")
	}
	if status == http.StatusForbidden || status == http.StatusUnauthorized {
		return errors.New("a chave nao foi aceita ou nao possui permissao")
	}
	return fmt.Errorf("a API recusou a analise (codigo %d)", status)
}

func groqAIModel() string {
	return strings.TrimSpace(getenv("GROQ_MODEL", "openai/gpt-oss-20b"))
}

func groqTechnicalCertificateModel() string {
	return strings.TrimSpace(getenv("GROQ_TECHNICAL_ANALYSIS_MODEL", groqAIModel()))
}

func groqCaptureClassificationModel() string {
	return strings.TrimSpace(getenv("GROQ_CAPTURE_CLASSIFICATION_MODEL", groqAIModel()))
}
