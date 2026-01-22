package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"soriano-mediadores/internal/sharepoint"
)

// ReportType tipo de reporte a generar
type ReportType string

const (
	ReportTypeDaily   ReportType = "daily"
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

// Generator generador de reportes
type Generator struct {
	sharepointClient *sharepoint.Client
	apiBaseURL       string // URL base de la API de analytics (ej: http://localhost:8080/api)
}

// NewGenerator crea un nuevo generador de reportes
func NewGenerator(spClient *sharepoint.Client, apiBaseURL string) *Generator {
	return &Generator{
		sharepointClient: spClient,
		apiBaseURL:       apiBaseURL,
	}
}

// GenerateReport genera un reporte y lo sube a SharePoint
func (g *Generator) GenerateReport(reportType ReportType, period string, publishToSharePoint bool) (*GenerateReportResult, error) {
	log.Printf("📊 Generando reporte %s para periodo %s", reportType, period)

	// Obtener datos de analytics desde la API
	analyticsData, err := g.fetchAnalyticsData()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo datos de analytics: %w", err)
	}

	result := &GenerateReportResult{
		ReportType:  string(reportType),
		Period:      period,
		GeneratedAt: time.Now(),
	}

	// Generar PDF
	pdfContent, err := g.generatePDF(reportType, period, analyticsData)
	if err != nil {
		return nil, fmt.Errorf("error generando PDF: %w", err)
	}
	result.PDFSizeBytes = len(pdfContent)

	// Generar Excel (opcional para reportes semanales y mensuales)
	if reportType != ReportTypeDaily {
		excelContent, err := g.generateExcel(reportType, period, analyticsData)
		if err != nil {
			log.Printf("Advertencia: error generando Excel: %v", err)
		} else {
			result.ExcelSizeBytes = len(excelContent)

			// Subir Excel a SharePoint si está habilitado
			if publishToSharePoint && g.sharepointClient != nil {
				excelFileName := fmt.Sprintf("%s_%s_Reporte_BI.xlsx", period, reportType)
				folderPath := g.getFolderPath(reportType)

				uploadResult, err := g.sharepointClient.UploadFile(&sharepoint.UploadFileOptions{
					FolderPath:  folderPath,
					FileName:    excelFileName,
					FileContent: excelContent,
					Overwrite:   true,
				})

				if err != nil {
					log.Printf("Error subiendo Excel a SharePoint: %v", err)
				} else {
					result.ExcelURL = uploadResult.WebURL
					log.Printf("✅ Excel subido: %s", uploadResult.WebURL)
				}
			}
		}
	}

	// Subir PDF a SharePoint si está habilitado
	if publishToSharePoint && g.sharepointClient != nil {
		pdfFileName := fmt.Sprintf("%s_%s_Reporte_BI.pdf", period, reportType)
		folderPath := g.getFolderPath(reportType)

		uploadResult, err := g.sharepointClient.UploadFile(&sharepoint.UploadFileOptions{
			FolderPath:  folderPath,
			FileName:    pdfFileName,
			FileContent: pdfContent,
			Overwrite:   true,
		})

		if err != nil {
			return nil, fmt.Errorf("error subiendo PDF a SharePoint: %w", err)
		}

		result.PDFURL = uploadResult.WebURL
		log.Printf("✅ PDF subido: %s", uploadResult.WebURL)
	}

	log.Printf("✅ Reporte generado exitosamente")

	return result, nil
}

// getFolderPath determina la carpeta de SharePoint según el tipo de reporte
func (g *Generator) getFolderPath(reportType ReportType) string {
	basePath := "Reportes BI"

	switch reportType {
	case ReportTypeDaily:
		return basePath + "/Diarios"
	case ReportTypeWeekly:
		return basePath + "/Semanales"
	case ReportTypeMonthly:
		return basePath + "/Mensuales"
	default:
		return basePath + "/Personalizados"
	}
}

// fetchAnalyticsData obtiene los datos de analytics desde la API
func (g *Generator) fetchAnalyticsData() (*AnalyticsData, error) {
	data := &AnalyticsData{}

	// Obtener Financial KPIs
	financialKPIs, err := g.fetchEndpoint("/analytics/financial-kpis")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo financial-kpis: %w", err)
	}
	data.FinancialKPIs = financialKPIs

	// Obtener Portfolio Analysis
	portfolioAnalysis, err := g.fetchEndpoint("/analytics/portfolio-analysis")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo portfolio-analysis: %w", err)
	}
	data.PortfolioAnalysis = portfolioAnalysis

	// Obtener Collections Performance
	collectionsPerformance, err := g.fetchEndpoint("/analytics/collections-performance")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo collections-performance: %w", err)
	}
	data.CollectionsPerformance = collectionsPerformance

	// Obtener Claims Analysis
	claimsAnalysis, err := g.fetchEndpoint("/analytics/claims-analysis")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo claims-analysis: %w", err)
	}
	data.ClaimsAnalysis = claimsAnalysis

	// Obtener Performance Trends
	performanceTrends, err := g.fetchEndpoint("/analytics/performance-trends?period=30days")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo performance-trends: %w", err)
	}
	data.PerformanceTrends = performanceTrends

	return data, nil
}

// fetchEndpoint hace una petición HTTP a un endpoint y devuelve el JSON
func (g *Generator) fetchEndpoint(endpoint string) (map[string]interface{}, error) {
	url := g.apiBaseURL + endpoint

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request a %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decodificando JSON: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("respuesta no exitosa de la API")
	}

	return result.Data, nil
}

// generatePDF genera el contenido del PDF según el tipo de reporte
func (g *Generator) generatePDF(reportType ReportType, period string, data *AnalyticsData) ([]byte, error) {
	builder := NewPDFBuilder()

	switch reportType {
	case ReportTypeDaily:
		return builder.BuildDailyReport(period, data)
	case ReportTypeWeekly:
		return builder.BuildWeeklyReport(period, data)
	case ReportTypeMonthly:
		return builder.BuildMonthlyReport(period, data)
	default:
		return nil, fmt.Errorf("tipo de reporte no soportado: %s", reportType)
	}
}

// generateExcel genera el contenido del Excel según el tipo de reporte
func (g *Generator) generateExcel(reportType ReportType, period string, data *AnalyticsData) ([]byte, error) {
	builder := NewExcelBuilder()

	switch reportType {
	case ReportTypeWeekly:
		return builder.BuildWeeklyReport(period, data)
	case ReportTypeMonthly:
		return builder.BuildMonthlyReport(period, data)
	default:
		return nil, fmt.Errorf("tipo de reporte no soportado para Excel: %s", reportType)
	}
}

// AnalyticsData contiene todos los datos de analytics
type AnalyticsData struct {
	FinancialKPIs          map[string]interface{}
	PortfolioAnalysis      map[string]interface{}
	CollectionsPerformance map[string]interface{}
	ClaimsAnalysis         map[string]interface{}
	PerformanceTrends      map[string]interface{}
}

// GenerateReportResult resultado de la generación de un reporte
type GenerateReportResult struct {
	ReportType     string    `json:"report_type"`
	Period         string    `json:"period"`
	PDFURL         string    `json:"pdf_url,omitempty"`
	ExcelURL       string    `json:"excel_url,omitempty"`
	GeneratedAt    time.Time `json:"generated_at"`
	PDFSizeBytes   int       `json:"pdf_size_bytes"`
	ExcelSizeBytes int       `json:"excel_size_bytes,omitempty"`
}

// GenerateDailyReport genera el reporte diario
func (g *Generator) GenerateDailyReport() error {
	period := time.Now().Format("2006-01-02")
	_, err := g.GenerateReport(ReportTypeDaily, period, true)
	return err
}

// GenerateWeeklyReport genera el reporte semanal
func (g *Generator) GenerateWeeklyReport() error {
	// Formato: 2026-W03
	year, week := time.Now().ISOWeek()
	period := fmt.Sprintf("%d-W%02d", year, week)
	_, err := g.GenerateReport(ReportTypeWeekly, period, true)
	return err
}

// GenerateMonthlyReport genera el reporte mensual
func (g *Generator) GenerateMonthlyReport() error {
	period := time.Now().Format("2006-01")
	_, err := g.GenerateReport(ReportTypeMonthly, period, true)
	return err
}

// CreateSharePointFolderStructure crea la estructura de carpetas en SharePoint
func (g *Generator) CreateSharePointFolderStructure() error {
	if g.sharepointClient == nil {
		return fmt.Errorf("cliente de SharePoint no configurado")
	}

	folders := []string{
		"Reportes BI",
		"Reportes BI/Diarios",
		"Reportes BI/Semanales",
		"Reportes BI/Mensuales",
		"Reportes BI/Personalizados",
	}

	for _, folder := range folders {
		if err := g.sharepointClient.CreateFolder(folder); err != nil {
			return fmt.Errorf("error creando carpeta %s: %w", folder, err)
		}
	}

	log.Printf("✅ Estructura de carpetas creada en SharePoint")

	return nil
}

// Helper para convertir bytes a buffer
func bytesToBuffer(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}
