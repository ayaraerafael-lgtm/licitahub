package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type academyCertificateData struct {
	VerificationCode string                     `json:"verificationCode"`
	IssuedAt         string                     `json:"issuedAt"`
	StudentName      string                     `json:"studentName"`
	JobTitle         string                     `json:"jobTitle"`
	CompanyName      string                     `json:"companyName"`
	CourseTitle      string                     `json:"courseTitle"`
	CourseCategory   string                     `json:"courseCategory"`
	CoverImageURL    string                     `json:"coverImageUrl"`
	WorkloadHours    int                        `json:"workloadHours"`
	Themes           []string                   `json:"themes"`
	Lessons          []academyCertificateLesson `json:"lessons"`
	ValidationURL    string                     `json:"-"`
}

type academyCertificateLesson struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationSeconds int    `json:"durationSeconds"`
}

func (a *app) downloadAcademyCertificate(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT certificate.verification_code AS "verificationCode", certificate.issued_at AS "issuedAt",
				users.full_name AS "studentName", COALESCE(users.job_title, '') AS "jobTitle",
				COALESCE(company.trade_name, '') AS "companyName", course.title AS "courseTitle",
				course.category AS "courseCategory", COALESCE(course.cover_image_url, '') AS "coverImageUrl",
				course.workload_hours AS "workloadHours",
				COALESCE((
					SELECT json_agg(lesson.title ORDER BY lesson.display_order)
					FROM academy_lessons lesson
					WHERE lesson.course_id = course.id AND lesson.is_published
				), '[]'::json) AS themes,
				COALESCE((
					SELECT json_agg(json_build_object(
						'title', lesson.title,
						'description', lesson.description,
						'durationSeconds', lesson.duration_seconds
					) ORDER BY lesson.display_order)
					FROM academy_lessons lesson
					WHERE lesson.course_id = course.id AND lesson.is_published
				), '[]'::json) AS lessons
			FROM academy_certificates certificate
			JOIN academy_courses course ON course.id = certificate.course_id
			JOIN users ON users.id = certificate.user_id
			LEFT JOIN companies company ON company.id = users.company_id
			WHERE certificate.course_id = %s::uuid AND certificate.user_id = %s::uuid
		) item;
	`, sqlQuote(courseID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar o certificado")
		return
	}
	var data academyCertificateData
	if strings.TrimSpace(string(payload)) == "null" || json.Unmarshal(payload, &data) != nil || data.VerificationCode == "" {
		writeError(w, http.StatusForbidden, "o certificado sera liberado somente apos a conclusao do curso")
		return
	}
	data.ValidationURL = strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/certificates/verify?code=" + data.VerificationCode
	pdf := buildAcademyCertificatePDF(data)
	fileName := "certificado-" + strings.ToLower(data.VerificationCode) + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (a *app) verifyAcademyCertificate(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("code")))
	if !regexp.MustCompile(`^[A-Z0-9]{12}$`).MatchString(code) {
		renderAcademyCertificateVerification(w, nil)
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT certificate.verification_code AS "verificationCode", certificate.issued_at AS "issuedAt",
				users.full_name AS "studentName", COALESCE(users.job_title, '') AS "jobTitle",
				COALESCE(company.trade_name, '') AS "companyName", course.title AS "courseTitle",
				course.category AS "courseCategory", course.workload_hours AS "workloadHours",
				COALESCE((
					SELECT json_agg(lesson.title ORDER BY lesson.display_order)
					FROM academy_lessons lesson
					WHERE lesson.course_id = course.id AND lesson.is_published
				), '[]'::json) AS themes
			FROM academy_certificates certificate
			JOIN academy_courses course ON course.id = certificate.course_id
			JOIN users ON users.id = certificate.user_id
			LEFT JOIN companies company ON company.id = users.company_id
			WHERE certificate.verification_code = %s
		) item;
	`, sqlQuote(code)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		renderAcademyCertificateVerification(w, nil)
		return
	}
	var data academyCertificateData
	if json.Unmarshal(payload, &data) != nil {
		renderAcademyCertificateVerification(w, nil)
		return
	}
	renderAcademyCertificateVerification(w, &data)
}

func renderAcademyCertificateVerification(w http.ResponseWriter, data *academyCertificateData) {
	const page = `<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Validação de certificado | LicitaHub</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#edf4f2;color:#163330;font-family:Arial,sans-serif}.top{height:8px;background:#075e58}
main{width:min(760px,calc(100% - 32px));margin:52px auto}.brand{display:flex;align-items:center;gap:14px;margin-bottom:24px}.mark{display:grid;place-items:center;width:48px;height:48px;background:#075e58;color:white;font-weight:800;border-radius:6px}.brand strong{display:block;font-size:20px}.brand span{color:#607370;font-size:13px}
.panel{background:white;border:1px solid #cddbd8;border-radius:8px;padding:32px;box-shadow:0 12px 36px rgba(22,51,48,.08)}.valid{display:inline-block;color:#086d4c;background:#e4f5ed;padding:7px 11px;border-radius:4px;font-weight:700}.invalid{color:#a52c2c;background:#fff0f0}.panel h1{font-size:25px;margin:20px 0 8px}.panel p{color:#607370;line-height:1.55}.grid{display:grid;grid-template-columns:1fr 1fr;gap:18px;margin-top:28px;padding-top:24px;border-top:1px solid #dce6e3}.grid span{display:block;color:#6c7e7a;font-size:12px;margin-bottom:5px}.grid strong{font-size:15px}.themes{grid-column:1/-1}.code{font-family:monospace;letter-spacing:1px}@media(max-width:580px){main{margin:24px auto}.panel{padding:22px}.grid{grid-template-columns:1fr}}
</style></head><body><div class="top"></div><main><div class="brand"><div class="mark">LH</div><div><strong>LicitaHub</strong><span>Validação pública de certificados</span></div></div>
<section class="panel">{{if .}}<span class="valid">Certificado autêntico</span><h1>{{.StudentName}}</h1><p>Este certificado foi emitido pela Academia LicitaHub após a conclusão integral do curso e das avaliações obrigatórias.</p>
<div class="grid"><div><span>Curso</span><strong>{{.CourseTitle}}</strong></div><div><span>Carga horária</span><strong>{{.WorkloadHours}} hora(s)</strong></div><div><span>Categoria</span><strong>{{.CourseCategory}}</strong></div><div><span>Emissão</span><strong>{{dateBR .IssuedAt}}</strong></div><div><span>Empresa</span><strong>{{if .CompanyName}}{{.CompanyName}}{{else}}Não informada{{end}}</strong></div><div><span>Código</span><strong class="code">{{.VerificationCode}}</strong></div><div class="themes"><span>Temáticas</span><strong>{{join .Themes}}</strong></div></div>
{{else}}<span class="valid invalid">Certificado não localizado</span><h1>Não foi possível validar o código</h1><p>Confira o código informado no certificado. Um certificado somente é emitido após a conclusão integral do curso.</p>{{end}}</section></main></body></html>`
	functions := template.FuncMap{
		"dateBR": certificateDateBR,
		"join":   func(values []string) string { return strings.Join(values, " | ") },
	}
	view, err := template.New("certificate-verification").Funcs(functions).Parse(page)
	if err != nil {
		http.Error(w, "Não foi possível abrir a validação.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = view.Execute(w, data)
}

func buildAcademyCertificatePDF(data academyCertificateData) []byte {
	cover := decodeCertificateCover(data.CoverImageURL)
	var firstPage bytes.Buffer
	writeCertificatePageFrame(&firstPage)
	firstPage.WriteString("0.035 0.310 0.295 rg 58 500 54 54 re f\n")
	firstPage.WriteString("1 1 1 rg\n")
	pdfText(&firstPage, "F2", 20, 85, 518, "LH", true)
	firstPage.WriteString("0.035 0.310 0.295 rg\n")
	pdfText(&firstPage, "F2", 13, 128, 535, "LICITAHUB", false)
	pdfText(&firstPage, "F1", 9, 128, 519, "ACADEMIA DE ENGENHARIA CONSULTIVA", false)
	drawCertificateCover(&firstPage, cover)
	firstPage.WriteString("0.035 0.310 0.295 rg\n")
	pdfText(&firstPage, "F2", 31, 421, 458, "CERTIFICADO", true)
	pdfText(&firstPage, "F1", 12, 421, 430, "Certificamos que", true)

	nameSize := 25.0
	if utf8.RuneCountInString(data.StudentName) > 42 {
		nameSize = 20
	}
	pdfText(&firstPage, "F2", nameSize, 421, 391, data.StudentName, true)
	firstPage.WriteString("0.820 0.650 0.180 RG 1.2 w 170 375 m 672 375 l S\n")

	pdfText(&firstPage, "F1", 11, 421, 352, "concluiu integralmente o curso", true)
	courseLines := wrapCertificateText(data.CourseTitle, 66)
	courseY := 321.0
	for _, line := range courseLines[:minInt(len(courseLines), 2)] {
		pdfText(&firstPage, "F2", 17, 421, courseY, line, true)
		courseY -= 21
	}

	details := fmt.Sprintf("%s | Carga horaria: %d hora(s)", data.CourseCategory, data.WorkloadHours)
	pdfText(&firstPage, "F1", 10, 421, courseY-3, details, true)
	pdfText(&firstPage, "F1", 8.5, 421, courseY-35, "Conteudo programatico detalhado nas paginas seguintes.", true)

	issued := certificateDateBR(data.IssuedAt)
	studentDetails := strings.Join(nonEmptyStrings(data.CompanyName, data.JobTitle), " | ")
	if studentDetails != "" {
		pdfText(&firstPage, "F1", 8.5, 421, 122, studentDetails, true)
	}
	pdfText(&firstPage, "F1", 9, 75, 82, "Conclusao e emissao: "+issued, false)
	pdfText(&firstPage, "F2", 9, 75, 64, "Codigo de validacao: "+data.VerificationCode, false)
	pdfText(&firstPage, "F1", 7.2, 535, 80, "Validacao publica:", false)
	pdfText(&firstPage, "F1", 7.2, 535, 64, data.ValidationURL, false)

	pages := [][]byte{firstPage.Bytes()}
	pages = append(pages, buildAcademyCurriculumPages(data)...)
	for index := range pages {
		var numbered bytes.Buffer
		numbered.Write(pages[index])
		pdfText(&numbered, "F1", 7.5, 767, 42, fmt.Sprintf("Pagina %d de %d", index+1, len(pages)), true)
		pages[index] = numbered.Bytes()
	}
	return assembleCertificatePDF(pages, cover)
}

func writeCertificatePageFrame(content *bytes.Buffer) {
	content.WriteString("0.965 0.980 0.975 rg 0 0 842 595 re f\n")
	content.WriteString("0.035 0.310 0.295 RG 3 w 22 22 798 551 re S\n")
	content.WriteString("0.820 0.650 0.180 RG 1 w 31 31 780 533 re S\n")
	content.WriteString("0.035 0.310 0.295 rg\n")
}

func drawCertificateCover(content *bytes.Buffer, cover *pdfEmbeddedImage) {
	content.WriteString("0.925 0.950 0.945 rg 650 486 140 76 re f\n")
	content.WriteString("0.760 0.835 0.820 RG 0.8 w 650 486 140 76 re S\n")
	if cover == nil {
		pdfText(content, "F1", 7.5, 720, 520, "CAPA DO CURSO", true)
		return
	}
	maxWidth, maxHeight := 130.0, 68.0
	width, height := maxWidth, maxWidth*float64(cover.Height)/float64(cover.Width)
	if height > maxHeight {
		height = maxHeight
		width = maxHeight * float64(cover.Width) / float64(cover.Height)
	}
	x := 655 + (maxWidth-width)/2
	y := 490 + (maxHeight-height)/2
	fmt.Fprintf(content, "q %.2f 0 0 %.2f %.2f %.2f cm /Im1 Do Q\n", width, height, x, y)
}

func buildAcademyCurriculumPages(data academyCertificateData) [][]byte {
	lessons := data.Lessons
	if len(lessons) == 0 {
		for _, title := range data.Themes {
			lessons = append(lessons, academyCertificateLesson{Title: title})
		}
	}
	pages := make([][]byte, 0)
	var page bytes.Buffer
	y := 0.0
	startPage := func() {
		page.Reset()
		writeCertificatePageFrame(&page)
		pdfText(&page, "F2", 21, 65, 520, "CONTEUDO PROGRAMATICO", false)
		pdfText(&page, "F1", 9, 65, 500, data.CourseTitle, false)
		page.WriteString("0.820 0.650 0.180 RG 1 w 65 484 m 777 484 l S\n")
		y = 454
	}
	finishPage := func() {
		copyOfPage := append([]byte(nil), page.Bytes()...)
		pages = append(pages, copyOfPage)
	}
	startPage()
	if len(lessons) == 0 {
		pdfText(&page, "F1", 10, 65, y, "Nenhuma aula publicada foi encontrada para este curso.", false)
		finishPage()
		return pages
	}
	for index, lesson := range lessons {
		descriptionLines := wrapCertificateText(strings.TrimSpace(lesson.Description), 112)
		if len(descriptionLines) == 1 && descriptionLines[0] == "" {
			descriptionLines = []string{"Descricao nao informada."}
		}
		if y < 115 {
			finishPage()
			startPage()
		}
		title := fmt.Sprintf("Aula %02d - %s", index+1, lesson.Title)
		pdfText(&page, "F2", 11, 65, y, title, false)
		pdfText(&page, "F1", 8.5, 700, y, certificateDuration(lesson.DurationSeconds), false)
		y -= 18
		for lineIndex, line := range descriptionLines {
			if y < 75 {
				finishPage()
				startPage()
				pdfText(&page, "F2", 10, 65, y, fmt.Sprintf("Aula %02d - continuacao", index+1), false)
				y -= 18
			}
			pdfText(&page, "F1", 8.5, 78, y, line, false)
			y -= 12
			if lineIndex == len(descriptionLines)-1 {
				y -= 4
			}
		}
		page.WriteString(fmt.Sprintf("0.850 0.890 0.880 RG 0.5 w 65 %.1f m 777 %.1f l S\n", y, y))
		y -= 15
	}
	finishPage()
	return pages
}

func certificateDuration(seconds int) string {
	if seconds <= 0 {
		return "Duracao nao informada"
	}
	minutes := (seconds + 59) / 60
	return fmt.Sprintf("%d min", minutes)
}

func pdfText(content *bytes.Buffer, font string, size, x, y float64, value string, centered bool) {
	if centered {
		x -= approximatePDFTextWidth(value, size) / 2
	}
	fmt.Fprintf(content, "BT /%s %.1f Tf %.1f %.1f Td (%s) Tj ET\n", font, size, x, y, escapePDFText(value))
}

func approximatePDFTextWidth(value string, size float64) float64 {
	return float64(utf8.RuneCountInString(value)) * size * 0.49
}

func escapePDFText(value string) string {
	var result bytes.Buffer
	for _, char := range value {
		switch char {
		case '\\', '(', ')':
			result.WriteByte('\\')
			result.WriteByte(byte(char))
		case '–', '—':
			result.WriteByte('-')
		case '“', '”':
			result.WriteByte('"')
		case '‘', '’':
			result.WriteByte('\'')
		case '•':
			result.WriteString(" - ")
		default:
			if char >= 32 && char <= 255 {
				result.WriteByte(byte(char))
			} else if char == '\n' || char == '\r' || char == '\t' {
				result.WriteByte(' ')
			} else {
				result.WriteByte('?')
			}
		}
	}
	return result.String()
}

func wrapCertificateText(value string, maxRunes int) []string {
	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{""}
	}
	lines := make([]string, 0)
	current := ""
	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if utf8.RuneCountInString(candidate) <= maxRunes {
			current = candidate
			continue
		}
		if current != "" {
			lines = append(lines, current)
		}
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

type pdfEmbeddedImage struct {
	Data   []byte
	Width  int
	Height int
}

func decodeCertificateCover(dataURL string) *pdfEmbeddedImage {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(dataURL)), "data:image/") {
		return nil
	}
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 || !strings.Contains(strings.ToLower(parts[0]), ";base64") {
		return nil
	}
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	source, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	resized := resizeCertificateImage(source, 600, 360)
	var encoded bytes.Buffer
	if jpeg.Encode(&encoded, resized, &jpeg.Options{Quality: 84}) != nil {
		return nil
	}
	bounds := resized.Bounds()
	return &pdfEmbeddedImage{Data: encoded.Bytes(), Width: bounds.Dx(), Height: bounds.Dy()}
}

func resizeCertificateImage(source image.Image, maxWidth, maxHeight int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maxWidth && height <= maxHeight {
		return source
	}
	scale := minFloat(float64(maxWidth)/float64(width), float64(maxHeight)/float64(height))
	targetWidth := maxInt(1, int(float64(width)*scale))
	targetHeight := maxInt(1, int(float64(height)*scale))
	target := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		sourceY := bounds.Min.Y + y*height/targetHeight
		for x := 0; x < targetWidth; x++ {
			sourceX := bounds.Min.X + x*width/targetWidth
			original := source.At(sourceX, sourceY)
			_, _, _, alpha := original.RGBA()
			if alpha < 0xffff {
				background := color.RGBA{R: 245, G: 249, B: 248, A: 255}
				target.Set(x, y, blendCertificatePixel(background, original))
			} else {
				target.Set(x, y, original)
			}
		}
	}
	return target
}

func blendCertificatePixel(background color.RGBA, foreground color.Color) color.RGBA {
	red, green, blue, alpha := foreground.RGBA()
	a := float64(alpha) / 65535
	return color.RGBA{
		R: uint8(float64(red>>8)*a + float64(background.R)*(1-a)),
		G: uint8(float64(green>>8)*a + float64(background.G)*(1-a)),
		B: uint8(float64(blue>>8)*a + float64(background.B)*(1-a)),
		A: 255,
	}
}

func assembleCertificatePDF(streams [][]byte, cover *pdfEmbeddedImage) []byte {
	objects := make([][]byte, 4)
	objects[0] = []byte("<< /Type /Catalog /Pages 2 0 R >>")
	objects[2] = []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	objects[3] = []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	imageObjectID := 0
	if cover != nil {
		imageObjectID = len(objects) + 1
		imageHeader := fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n", cover.Width, cover.Height, len(cover.Data))
		imageObject := append([]byte(imageHeader), cover.Data...)
		imageObject = append(imageObject, []byte("\nendstream")...)
		objects = append(objects, imageObject)
	}
	pageObjectIDs := make([]int, 0, len(streams))
	for _, stream := range streams {
		pageObjectID := len(objects) + 1
		contentObjectID := pageObjectID + 1
		pageObjectIDs = append(pageObjectIDs, pageObjectID)
		resources := "/Resources << /Font << /F1 3 0 R /F2 4 0 R >>"
		if imageObjectID > 0 {
			resources += fmt.Sprintf(" /XObject << /Im1 %d 0 R >>", imageObjectID)
		}
		resources += " >>"
		pageObject := []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 842 595] %s /Contents %d 0 R >>", resources, contentObjectID))
		contentObject := append([]byte(fmt.Sprintf("<< /Length %d >>\nstream\n", len(stream))), stream...)
		contentObject = append(contentObject, []byte("\nendstream")...)
		objects = append(objects, pageObject, contentObject)
	}
	kids := make([]string, 0, len(pageObjectIDs))
	for _, objectID := range pageObjectIDs {
		kids = append(kids, fmt.Sprintf("%d 0 R", objectID))
	}
	objects[1] = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pageObjectIDs)))

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n", index+1)
		pdf.Write(object)
		pdf.WriteString("\nendobj\n")
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n", len(objects)+1)
	pdf.WriteString("0000000000 65535 f \n")
	for index := 1; index <= len(objects); index++ {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offsets[index])
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return pdf.Bytes()
}

func certificateDateBR(value string) string {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return strings.Split(value, "T")[0]
	}
	return parsed.Format("02/01/2006")
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}
