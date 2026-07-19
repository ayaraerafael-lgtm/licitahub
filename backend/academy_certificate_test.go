package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
)

func TestBuildAcademyCertificatePDF(t *testing.T) {
	data := academyCertificateData{
		VerificationCode: "A1B2C3D4E5F6",
		IssuedAt:         "2026-07-18T15:30:00-03:00",
		StudentName:      "Mariana de Albuquerque Souza",
		JobTitle:         "Engenheira civil",
		CompanyName:      "Empresa Consultiva de Engenharia",
		CourseTitle:      "Planejamento e Gestão de Licitações de Engenharia Consultiva",
		CourseCategory:   "Licitações e propostas",
		CoverImageURL:    testAcademyCoverDataURL(t),
		WorkloadHours:    20,
		Themes:           []string{"Leitura estratégica do edital", "Formação de consórcios", "Planejamento da proposta", "Gestão de documentos técnicos"},
		Lessons: []academyCertificateLesson{
			{Title: "Leitura estratégica do edital", Description: "Interpretação do objeto, dos critérios de julgamento, dos requisitos técnicos e dos riscos que devem orientar a decisão de participação.", DurationSeconds: 1800},
			{Title: "Formação de consórcios", Description: "Avaliação de complementaridade entre empresas, definição de liderança e organização das responsabilidades da futura composição.", DurationSeconds: 2400},
			{Title: "Planejamento da proposta", Description: "Estruturação das etapas de trabalho, equipe, documentos, orçamento e controles necessários para uma entrega consistente.", DurationSeconds: 2700},
			{Title: "Gestão de documentos técnicos", Description: "Organização dos comprovantes de capacidade técnica, revisão e consolidação final dos arquivos da proposta.", DurationSeconds: 1500},
		},
		ValidationURL: "http://127.0.0.1:8080/certificates/verify?code=A1B2C3D4E5F6",
	}
	pdf := buildAcademyCertificatePDF(data)
	if !strings.HasPrefix(string(pdf), "%PDF-1.4") {
		t.Fatal("o arquivo gerado nao possui cabecalho PDF")
	}
	if len(pdf) < 2000 {
		t.Fatal("o certificado PDF foi gerado sem conteudo suficiente")
	}
	if path := os.Getenv("ACADEMY_CERTIFICATE_PREVIEW"); path != "" {
		if err := os.WriteFile(path, pdf, 0o644); err != nil {
			t.Fatalf("nao foi possivel salvar a previa: %v", err)
		}
	}
}

func testAcademyCoverDataURL(t *testing.T) string {
	t.Helper()
	cover := image.NewRGBA(image.Rect(0, 0, 320, 180))
	for y := 0; y < 180; y++ {
		for x := 0; x < 320; x++ {
			if x < 105 {
				cover.Set(x, y, color.RGBA{R: 5, G: 83, B: 78, A: 255})
			} else {
				cover.Set(x, y, color.RGBA{R: 220, G: 177, B: 48, A: 255})
			}
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, cover); err != nil {
		t.Fatalf("nao foi possivel preparar a capa de teste: %v", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
}
