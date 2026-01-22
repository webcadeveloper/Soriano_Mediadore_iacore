package reports

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ExcelBuilder constructor de archivos Excel
type ExcelBuilder struct {
	file *excelize.File
}

// NewExcelBuilder crea un nuevo constructor de Excel
func NewExcelBuilder() *ExcelBuilder {
	f := excelize.NewFile()

	return &ExcelBuilder{
		file: f,
	}
}

// BuildWeeklyReport genera un reporte semanal en Excel
func (b *ExcelBuilder) BuildWeeklyReport(period string, data *AnalyticsData) ([]byte, error) {
	// Crear hojas
	b.file.SetSheetName("Sheet1", "Resumen")
	b.file.NewSheet("KPIs Financieros")
	b.file.NewSheet("Cartera")
	b.file.NewSheet("Cobros")
	b.file.NewSheet("Siniestros")

	// Hoja 1: Resumen
	b.addSummarySheet(period, data)

	// Hoja 2: KPIs Financieros
	b.addFinancialKPIsSheet(data)

	// Hoja 3: Cartera
	b.addPortfolioSheet(data)

	// Hoja 4: Cobros
	b.addCollectionsSheet(data)

	// Hoja 5: Siniestros
	b.addClaimsSheet(data)

	return b.getOutput()
}

// BuildMonthlyReport genera un reporte mensual completo en Excel
func (b *ExcelBuilder) BuildMonthlyReport(period string, data *AnalyticsData) ([]byte, error) {
	// Crear hojas
	b.file.SetSheetName("Sheet1", "Resumen Ejecutivo")
	b.file.NewSheet("KPIs Financieros")
	b.file.NewSheet("Pocket Share")
	b.file.NewSheet("Cartera")
	b.file.NewSheet("Distribucion Ramo")
	b.file.NewSheet("Cobros y Morosidad")
	b.file.NewSheet("Siniestros")
	b.file.NewSheet("Tendencias")

	// Añadir contenido a cada hoja
	b.addExecutiveSummarySheet(period, data)
	b.addFinancialKPIsSheet(data)
	b.addPocketShareSheet(data)
	b.addPortfolioSheet(data)
	b.addDistributionByRamoSheet(data)
	b.addCollectionsSheet(data)
	b.addClaimsSheet(data)
	b.addTrendsSheet(data)

	return b.getOutput()
}

// addSummarySheet agrega hoja de resumen
func (b *ExcelBuilder) addSummarySheet(period string, data *AnalyticsData) {
	sheet := "Resumen"

	// Título
	b.file.SetCellValue(sheet, "A1", "REPORTE SEMANAL - SORIANO MEDIADORES")
	b.file.SetCellValue(sheet, "A2", fmt.Sprintf("Periodo: %s", period))

	// Estilo del título
	style, _ := b.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "C2185B"},
	})
	b.file.SetCellStyle(sheet, "A1", "A1", style)

	// KPIs principales
	if data.FinancialKPIs != nil {
		b.file.SetCellValue(sheet, "A4", "KPI")
		b.file.SetCellValue(sheet, "B4", "Valor")

		row := 5
		kpis := []struct {
			label string
			key   string
		}{
			{"Primas Mes Actual", "total_primas_mes_actual"},
			{"Comisiones del Mes", "total_comisiones_mes"},
			{"Primas Año Actual", "total_primas_anio_actual"},
			{"Comisiones Año", "total_comisiones_anio"},
		}

		for _, kpi := range kpis {
			b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), kpi.label)
			b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(data.FinancialKPIs, kpi.key))
			row++
		}

		// Formato de moneda
		currencyStyle, _ := b.file.NewStyle(&excelize.Style{
			NumFmt: 4, // Formato de moneda
		})
		b.file.SetCellStyle(sheet, "B5", fmt.Sprintf("B%d", row-1), currencyStyle)
	}

	// Ajustar anchos de columna
	b.file.SetColWidth(sheet, "A", "A", 30)
	b.file.SetColWidth(sheet, "B", "B", 20)
}

// addExecutiveSummarySheet agrega resumen ejecutivo
func (b *ExcelBuilder) addExecutiveSummarySheet(period string, data *AnalyticsData) {
	sheet := "Resumen Ejecutivo"

	b.file.SetCellValue(sheet, "A1", "REPORTE MENSUAL COMPLETO")
	b.file.SetCellValue(sheet, "A2", fmt.Sprintf("Periodo: %s", period))

	style, _ := b.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "C2185B"},
	})
	b.file.SetCellStyle(sheet, "A1", "A1", style)

	b.file.SetCellValue(sheet, "A4", "Este reporte mensual incluye:")
	b.file.SetCellValue(sheet, "A5", "• KPIs financieros completos")
	b.file.SetCellValue(sheet, "A6", "• Análisis de pocket share por compañía")
	b.file.SetCellValue(sheet, "A7", "• Análisis de cartera y concentración de riesgo")
	b.file.SetCellValue(sheet, "A8", "• Distribución por ramo")
	b.file.SetCellValue(sheet, "A9", "• Rendimiento de cobros y morosidad")
	b.file.SetCellValue(sheet, "A10", "• Análisis de siniestros")
	b.file.SetCellValue(sheet, "A11", "• Tendencias de rendimiento")

	b.file.SetColWidth(sheet, "A", "A", 50)
}

// addFinancialKPIsSheet agrega hoja de KPIs financieros
func (b *ExcelBuilder) addFinancialKPIsSheet(data *AnalyticsData) {
	sheet := "KPIs Financieros"

	// Cabeceras
	b.file.SetCellValue(sheet, "A1", "KPI Financiero")
	b.file.SetCellValue(sheet, "B1", "Valor")

	headerStyle, _ := b.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F5F5F5"}, Pattern: 1},
	})
	b.file.SetCellStyle(sheet, "A1", "B1", headerStyle)

	if data.FinancialKPIs != nil {
		kpis := []struct {
			label string
			key   string
		}{
			{"Primas Mes Actual", "total_primas_mes_actual"},
			{"Primas Mes Anterior", "total_primas_mes_anterior"},
			{"Primas Año Actual", "total_primas_anio_actual"},
			{"Comisiones del Mes", "total_comisiones_mes"},
			{"Comisiones del Año", "total_comisiones_anio"},
			{"Promedio Prima por Póliza", "promedio_prima_poliza"},
			{"Margen Comisión Promedio (%)", "margen_comision_promedio"},
		}

		for i, kpi := range kpis {
			row := i + 2
			b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), kpi.label)
			b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(data.FinancialKPIs, kpi.key))
		}

		// Formato de moneda para columna B
		currencyStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 4})
		b.file.SetCellStyle(sheet, "B2", "B8", currencyStyle)
	}

	b.file.SetColWidth(sheet, "A", "A", 35)
	b.file.SetColWidth(sheet, "B", "B", 20)
}

// addPocketShareSheet agrega hoja de pocket share
func (b *ExcelBuilder) addPocketShareSheet(data *AnalyticsData) {
	sheet := "Pocket Share"

	// Cabeceras
	headers := []string{"Compañía Aseguradora", "Nº Pólizas", "Total Primas", "Porcentaje (%)"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		b.file.SetCellValue(sheet, cell, header)
	}

	headerStyle, _ := b.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F5F5F5"}, Pattern: 1},
	})
	b.file.SetCellStyle(sheet, "A1", "D1", headerStyle)

	// Datos
	if data.FinancialKPIs != nil {
		if pocketShare, ok := data.FinancialKPIs["pocket_share_por_compania"].([]interface{}); ok {
			for i, item := range pocketShare {
				if itemMap, ok := item.(map[string]interface{}); ok {
					row := i + 2
					b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), getString(itemMap, "gestora"))
					b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(itemMap, "num_polizas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("C%d", row), getFloat(itemMap, "total_primas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("D%d", row), getFloat(itemMap, "porcentaje"))
				}
			}

			// Formatos
			currencyStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 4})
			b.file.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", len(pocketShare)+1), currencyStyle)

			percentStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 10})
			b.file.SetCellStyle(sheet, "D2", fmt.Sprintf("D%d", len(pocketShare)+1), percentStyle)
		}
	}

	b.file.SetColWidth(sheet, "A", "A", 30)
	b.file.SetColWidth(sheet, "B", "D", 15)
}

// addPortfolioSheet agrega análisis de cartera
func (b *ExcelBuilder) addPortfolioSheet(data *AnalyticsData) {
	sheet := "Cartera"

	b.file.SetCellValue(sheet, "A1", "ANÁLISIS DE CARTERA")

	if data.PortfolioAnalysis != nil {
		// Concentración de riesgo
		b.file.SetCellValue(sheet, "A3", "Concentración de Riesgo (Top 20)")
		concentracion := getFloat(data.PortfolioAnalysis, "concentracion_riesgo")
		b.file.SetCellValue(sheet, "B3", concentracion)

		percentStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 10})
		b.file.SetCellStyle(sheet, "B3", "B3", percentStyle)

		// Distribución por provincia
		b.file.SetCellValue(sheet, "A5", "Distribución por Provincia")
		headers := []string{"Provincia", "Nº Clientes", "Nº Pólizas", "Total Primas"}
		for i, header := range headers {
			cell := fmt.Sprintf("%c6", 'A'+i)
			b.file.SetCellValue(sheet, cell, header)
		}

		if distribProv, ok := data.PortfolioAnalysis["distribucion_por_provincia"].([]interface{}); ok {
			for i, item := range distribProv {
				if itemMap, ok := item.(map[string]interface{}); ok {
					row := i + 7
					b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), getString(itemMap, "provincia"))
					b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(itemMap, "num_clientes"))
					b.file.SetCellValue(sheet, fmt.Sprintf("C%d", row), getFloat(itemMap, "num_polizas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("D%d", row), getFloat(itemMap, "total_primas"))
				}
			}

			currencyStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 4})
			b.file.SetCellStyle(sheet, "D7", fmt.Sprintf("D%d", len(distribProv)+6), currencyStyle)
		}
	}

	b.file.SetColWidth(sheet, "A", "A", 25)
	b.file.SetColWidth(sheet, "B", "D", 15)
}

// addDistributionByRamoSheet agrega distribución por ramo
func (b *ExcelBuilder) addDistributionByRamoSheet(data *AnalyticsData) {
	sheet := "Distribucion Ramo"

	headers := []string{"Ramo", "Nº Pólizas", "Total Primas", "Porcentaje"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		b.file.SetCellValue(sheet, cell, header)
	}

	if data.PortfolioAnalysis != nil {
		if distribRamo, ok := data.PortfolioAnalysis["distribucion_por_ramo"].([]interface{}); ok {
			for i, item := range distribRamo {
				if itemMap, ok := item.(map[string]interface{}); ok {
					row := i + 2
					b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), getString(itemMap, "ramo"))
					b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(itemMap, "num_polizas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("C%d", row), getFloat(itemMap, "total_primas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("D%d", row), getFloat(itemMap, "porcentaje"))
				}
			}

			currencyStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 4})
			b.file.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", len(distribRamo)+1), currencyStyle)

			percentStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 10})
			b.file.SetCellStyle(sheet, "D2", fmt.Sprintf("D%d", len(distribRamo)+1), percentStyle)
		}
	}

	b.file.SetColWidth(sheet, "A", "A", 30)
	b.file.SetColWidth(sheet, "B", "D", 15)
}

// addCollectionsSheet agrega cobros y morosidad
func (b *ExcelBuilder) addCollectionsSheet(data *AnalyticsData) {
	sheet := "Cobros y Morosidad"

	b.file.SetCellValue(sheet, "A1", "RENDIMIENTO DE COBROS")

	if data.CollectionsPerformance != nil {
		// KPIs principales
		b.file.SetCellValue(sheet, "A3", "Ratio de Cobro Mensual")
		b.file.SetCellValue(sheet, "B3", getFloat(data.CollectionsPerformance, "ratio_cobro_mes"))

		if deudaInfo, ok := data.CollectionsPerformance["deuda_total_vs_cartera"].(map[string]interface{}); ok {
			b.file.SetCellValue(sheet, "A4", "Deuda Total")
			b.file.SetCellValue(sheet, "B4", getFloat(deudaInfo, "deuda_total"))

			b.file.SetCellValue(sheet, "A5", "Porcentaje Morosidad")
			b.file.SetCellValue(sheet, "B5", getFloat(deudaInfo, "porcentaje_morosidad"))
		}

		// Morosidad por rango
		b.file.SetCellValue(sheet, "A7", "Morosidad por Rango de Días")
		headers := []string{"Rango", "Nº Recibos", "Importe Total"}
		for i, header := range headers {
			cell := fmt.Sprintf("%c8", 'A'+i)
			b.file.SetCellValue(sheet, cell, header)
		}

		if morosidadRango, ok := data.CollectionsPerformance["morosidad_por_rango_dias"].([]interface{}); ok {
			for i, item := range morosidadRango {
				if itemMap, ok := item.(map[string]interface{}); ok {
					row := i + 9
					b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), getString(itemMap, "rango"))
					b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(itemMap, "num_recibos"))
					b.file.SetCellValue(sheet, fmt.Sprintf("C%d", row), getFloat(itemMap, "importe_total"))
				}
			}

			currencyStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 4})
			b.file.SetCellStyle(sheet, "C9", fmt.Sprintf("C%d", len(morosidadRango)+8), currencyStyle)
		}
	}

	b.file.SetColWidth(sheet, "A", "A", 30)
	b.file.SetColWidth(sheet, "B", "C", 15)
}

// addClaimsSheet agrega análisis de siniestros
func (b *ExcelBuilder) addClaimsSheet(data *AnalyticsData) {
	sheet := "Siniestros"

	b.file.SetCellValue(sheet, "A1", "ANÁLISIS DE SINIESTROS")

	if data.ClaimsAnalysis != nil {
		// KPIs
		b.file.SetCellValue(sheet, "A3", "Siniestros Abiertos")
		b.file.SetCellValue(sheet, "B3", getFloat(data.ClaimsAnalysis, "siniestros_abiertos"))

		b.file.SetCellValue(sheet, "A4", "Siniestros Cerrados")
		b.file.SetCellValue(sheet, "B4", getFloat(data.ClaimsAnalysis, "siniestros_cerrados"))

		b.file.SetCellValue(sheet, "A5", "Tiempo Medio Resolución (días)")
		b.file.SetCellValue(sheet, "B5", getFloat(data.ClaimsAnalysis, "tiempo_medio_resolucion_dias"))

		// Siniestralidad por ramo
		b.file.SetCellValue(sheet, "A7", "Siniestralidad por Ramo")
		headers := []string{"Ramo", "Nº Pólizas", "Nº Siniestros", "Siniestralidad (%)"}
		for i, header := range headers {
			cell := fmt.Sprintf("%c8", 'A'+i)
			b.file.SetCellValue(sheet, cell, header)
		}

		if siniestralidad, ok := data.ClaimsAnalysis["siniestralidad_por_ramo"].([]interface{}); ok {
			for i, item := range siniestralidad {
				if itemMap, ok := item.(map[string]interface{}); ok {
					row := i + 9
					b.file.SetCellValue(sheet, fmt.Sprintf("A%d", row), getString(itemMap, "ramo"))
					b.file.SetCellValue(sheet, fmt.Sprintf("B%d", row), getFloat(itemMap, "num_polizas"))
					b.file.SetCellValue(sheet, fmt.Sprintf("C%d", row), getFloat(itemMap, "num_siniestros"))
					b.file.SetCellValue(sheet, fmt.Sprintf("D%d", row), getFloat(itemMap, "siniestralidad"))
				}
			}

			percentStyle, _ := b.file.NewStyle(&excelize.Style{NumFmt: 10})
			b.file.SetCellStyle(sheet, "D9", fmt.Sprintf("D%d", len(siniestralidad)+8), percentStyle)
		}
	}

	b.file.SetColWidth(sheet, "A", "A", 30)
	b.file.SetColWidth(sheet, "B", "D", 15)
}

// addTrendsSheet agrega tendencias
func (b *ExcelBuilder) addTrendsSheet(data *AnalyticsData) {
	sheet := "Tendencias"

	b.file.SetCellValue(sheet, "A1", "TENDENCIAS DE RENDIMIENTO (30 DÍAS)")

	b.file.SetCellValue(sheet, "A3", "Análisis de evolución de clientes, pólizas y primas.")

	b.file.SetColWidth(sheet, "A", "A", 50)
}

// getOutput devuelve el contenido del archivo Excel
func (b *ExcelBuilder) getOutput() ([]byte, error) {
	buf, err := b.file.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Helper para extraer string de map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
