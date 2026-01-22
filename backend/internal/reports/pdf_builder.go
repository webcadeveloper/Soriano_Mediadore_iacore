package reports

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// PDFBuilder constructor de PDFs
type PDFBuilder struct {
	pdf *gofpdf.Fpdf
}

// NewPDFBuilder crea un nuevo constructor de PDFs
func NewPDFBuilder() *PDFBuilder {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	return &PDFBuilder{
		pdf: pdf,
	}
}

// BuildDailyReport genera un reporte diario en PDF
func (b *PDFBuilder) BuildDailyReport(period string, data *AnalyticsData) ([]byte, error) {
	b.addCoverPage("Reporte Diario", period)
	b.addFinancialKPIsPage(data)
	b.addSummaryPage(data)

	return b.getOutput()
}

// BuildWeeklyReport genera un reporte semanal en PDF
func (b *PDFBuilder) BuildWeeklyReport(period string, data *AnalyticsData) ([]byte, error) {
	b.addCoverPage("Reporte Semanal", period)
	b.addFinancialKPIsPage(data)
	b.addPortfolioAnalysisPage(data)
	b.addCollectionsPage(data)
	b.addClaimsPage(data)

	return b.getOutput()
}

// BuildMonthlyReport genera un reporte mensual completo en PDF
func (b *PDFBuilder) BuildMonthlyReport(period string, data *AnalyticsData) ([]byte, error) {
	b.addCoverPage("Reporte Mensual Completo", period)
	b.addExecutiveSummaryPage(data)
	b.addFinancialKPIsPage(data)
	b.addPortfolioAnalysisPage(data)
	b.addCollectionsPage(data)
	b.addClaimsPage(data)
	b.addPerformanceTrendsPage(data)

	return b.getOutput()
}

// addCoverPage agrega portada
func (b *PDFBuilder) addCoverPage(title, period string) {
	b.pdf.AddPage()

	// Logo/Header (placeholder)
	b.pdf.SetFont("Arial", "B", 32)
	b.pdf.SetTextColor(194, 24, 91) // Color principal: #c2185b
	b.pdf.CellFormat(0, 30, "", "", 0, "C", false, 0, "")
	b.pdf.Ln(20)

	// Título
	b.pdf.CellFormat(0, 20, title, "", 1, "C", false, 0, "")
	b.pdf.Ln(10)

	// Subtítulo
	b.pdf.SetFont("Arial", "", 18)
	b.pdf.SetTextColor(102, 102, 102)
	b.pdf.CellFormat(0, 10, "Soriano Mediadores", "", 1, "C", false, 0, "")
	b.pdf.Ln(5)

	// Periodo
	b.pdf.SetFont("Arial", "B", 14)
	b.pdf.SetTextColor(0, 0, 0)
	b.pdf.CellFormat(0, 10, fmt.Sprintf("Periodo: %s", period), "", 1, "C", false, 0, "")
	b.pdf.Ln(5)

	// Fecha de generación
	b.pdf.SetFont("Arial", "", 10)
	b.pdf.SetTextColor(153, 153, 153)
	b.pdf.CellFormat(0, 10, fmt.Sprintf("Generado: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "C", false, 0, "")
	b.pdf.Ln(50)

	// Footer de portada
	b.pdf.SetY(-40)
	b.pdf.SetFont("Arial", "I", 8)
	b.pdf.SetTextColor(153, 153, 153)
	b.pdf.CellFormat(0, 5, "Business Intelligence - Sistema Automatico de Reportes", "", 1, "C", false, 0, "")
	b.pdf.CellFormat(0, 5, "Generado automaticamente por el sistema ERP", "", 1, "C", false, 0, "")
}

// addExecutiveSummaryPage agrega resumen ejecutivo
func (b *PDFBuilder) addExecutiveSummaryPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Resumen Ejecutivo")

	b.pdf.SetFont("Arial", "", 11)
	b.pdf.SetTextColor(0, 0, 0)

	// Análisis general
	b.pdf.MultiCell(0, 6, "Este reporte mensual presenta un análisis completo de la actividad comercial y operativa de Soriano Mediadores, incluyendo:\n", "", "L", false)
	b.pdf.Ln(3)

	bulletPoints := []string{
		"KPIs financieros: primas, comisiones y pocket share",
		"Analisis de cartera: distribucion por ramo y concentracion de riesgo",
		"Rendimiento de cobros: ratio de cobro y morosidad",
		"Analisis de siniestros: siniestralidad por ramo",
		"Tendencias de rendimiento: evolucion de clientes y polizas",
	}

	for _, point := range bulletPoints {
		b.pdf.SetFont("Arial", "", 10)
		b.pdf.Cell(10, 6, "•")
		b.pdf.MultiCell(0, 6, point, "", "L", false)
	}

	b.pdf.Ln(5)
}

// addFinancialKPIsPage agrega página de KPIs financieros
func (b *PDFBuilder) addFinancialKPIsPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("KPIs Financieros")

	if data.FinancialKPIs != nil {
		b.addKPIBox("Primas Mes Actual", formatCurrency(getFloat(data.FinancialKPIs, "total_primas_mes_actual")))
		b.addKPIBox("Comisiones del Mes", formatCurrency(getFloat(data.FinancialKPIs, "total_comisiones_mes")))
		b.addKPIBox("Primas Año Actual", formatCurrency(getFloat(data.FinancialKPIs, "total_primas_anio_actual")))
		b.addKPIBox("Comisiones Año", formatCurrency(getFloat(data.FinancialKPIs, "total_comisiones_anio")))
		b.addKPIBox("Promedio Prima/Poliza", formatCurrency(getFloat(data.FinancialKPIs, "promedio_prima_poliza")))
		b.addKPIBox("Margen Comision Promedio", formatPercentage(getFloat(data.FinancialKPIs, "margen_comision_promedio")))
	} else {
		b.pdf.Cell(0, 10, "No hay datos disponibles")
	}

	b.pdf.Ln(10)
}

// addPortfolioAnalysisPage agrega análisis de cartera
func (b *PDFBuilder) addPortfolioAnalysisPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Analisis de Cartera")

	if data.PortfolioAnalysis != nil {
		concentracion := getFloat(data.PortfolioAnalysis, "concentracion_riesgo")
		b.addKPIBox("Concentracion de Riesgo (Top 20)", formatPercentage(concentracion))
	}

	b.pdf.Ln(5)
}

// addCollectionsPage agrega rendimiento de cobros
func (b *PDFBuilder) addCollectionsPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Rendimiento de Cobros y Morosidad")

	if data.CollectionsPerformance != nil {
		ratioCobro := getFloat(data.CollectionsPerformance, "ratio_cobro_mes")
		b.addKPIBox("Ratio de Cobro Mensual", formatPercentage(ratioCobro))

		// Deuda total vs cartera
		if deudaInfo, ok := data.CollectionsPerformance["deuda_total_vs_cartera"].(map[string]interface{}); ok {
			deudaTotal := getFloat(deudaInfo, "deuda_total")
			porcentajeMorosidad := getFloat(deudaInfo, "porcentaje_morosidad")

			b.addKPIBox("Deuda Total", formatCurrency(deudaTotal))
			b.addKPIBox("Porcentaje Morosidad", formatPercentage(porcentajeMorosidad))
		}
	}

	b.pdf.Ln(5)
}

// addClaimsPage agrega análisis de siniestros
func (b *PDFBuilder) addClaimsPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Analisis de Siniestros")

	if data.ClaimsAnalysis != nil {
		siniestrosAbiertos := getFloat(data.ClaimsAnalysis, "siniestros_abiertos")
		siniestrosCerrados := getFloat(data.ClaimsAnalysis, "siniestros_cerrados")
		tiempoMedio := getFloat(data.ClaimsAnalysis, "tiempo_medio_resolucion_dias")

		b.addKPIBox("Siniestros Abiertos", fmt.Sprintf("%.0f", siniestrosAbiertos))
		b.addKPIBox("Siniestros Cerrados", fmt.Sprintf("%.0f", siniestrosCerrados))
		b.addKPIBox("Tiempo Medio Resolucion", fmt.Sprintf("%.0f dias", tiempoMedio))
	}

	b.pdf.Ln(5)
}

// addPerformanceTrendsPage agrega tendencias de rendimiento
func (b *PDFBuilder) addPerformanceTrendsPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Tendencias de Rendimiento")

	b.pdf.SetFont("Arial", "", 10)
	b.pdf.MultiCell(0, 6, "Analisis de tendencias basado en los ultimos 30 dias de actividad.", "", "L", false)

	b.pdf.Ln(5)
}

// addSummaryPage agrega resumen breve
func (b *PDFBuilder) addSummaryPage(data *AnalyticsData) {
	b.pdf.AddPage()

	b.addSectionHeader("Resumen del Dia")

	b.pdf.SetFont("Arial", "", 10)
	b.pdf.MultiCell(0, 6, "Principales indicadores del dia:", "", "L", false)
	b.pdf.Ln(3)

	if data.FinancialKPIs != nil {
		primasMes := formatCurrency(getFloat(data.FinancialKPIs, "total_primas_mes_actual"))
		comisionesMes := formatCurrency(getFloat(data.FinancialKPIs, "total_comisiones_mes"))

		b.pdf.SetFont("Arial", "", 10)
		b.pdf.Cell(10, 6, "•")
		b.pdf.Cell(0, 6, fmt.Sprintf("Primas acumuladas del mes: %s", primasMes))
		b.pdf.Ln(6)

		b.pdf.Cell(10, 6, "•")
		b.pdf.Cell(0, 6, fmt.Sprintf("Comisiones del mes: %s", comisionesMes))
		b.pdf.Ln(6)
	}

	b.pdf.Ln(5)
}

// Helper functions

func (b *PDFBuilder) addSectionHeader(title string) {
	b.pdf.SetFont("Arial", "B", 16)
	b.pdf.SetTextColor(194, 24, 91)
	b.pdf.CellFormat(0, 12, title, "B", 1, "L", false, 0, "")
	b.pdf.Ln(5)
	b.pdf.SetTextColor(0, 0, 0)
}

func (b *PDFBuilder) addKPIBox(label, value string) {
	b.pdf.SetFillColor(250, 250, 250)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.SetTextColor(102, 102, 102)
	b.pdf.CellFormat(90, 6, label, "1", 0, "L", true, 0, "")

	b.pdf.SetFont("Arial", "B", 11)
	b.pdf.SetTextColor(0, 0, 0)
	b.pdf.CellFormat(0, 6, value, "1", 1, "R", true, 0, "")

	b.pdf.Ln(2)
}

func (b *PDFBuilder) getOutput() ([]byte, error) {
	var buf []byte
	buf = b.pdf.Output(buf)

	if b.pdf.Error() != nil {
		return nil, b.pdf.Error()
	}

	return buf, nil
}

// Helper para extraer float de map
func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

// Formatear moneda
func formatCurrency(amount float64) string {
	return fmt.Sprintf("€ %.2f", amount)
}

// Formatear porcentaje
func formatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}
