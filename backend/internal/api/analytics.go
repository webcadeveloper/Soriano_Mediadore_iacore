package api

import (
	"database/sql"
	"fmt"
	"log"
	"soriano-mediadores/internal/db"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ========================================
// STRUCTS DE RESPUESTA
// ========================================

// FinancialKPIsResponse - KPIs financieros
type FinancialKPIsResponse struct {
	TotalPrimasMesActual      float64                `json:"total_primas_mes_actual"`
	TotalPrimasMesAnterior    float64                `json:"total_primas_mes_anterior"`
	TotalPrimasAnioActual     float64                `json:"total_primas_anio_actual"`
	TotalComisionesMes        float64                `json:"total_comisiones_mes"`
	TotalComisionesAnio       float64                `json:"total_comisiones_anio"`
	PromedioPrimaPoliza       float64                `json:"promedio_prima_poliza"`
	MargenComisionPromedio    float64                `json:"margen_comision_promedio"`
	PocketSharePorCompania    []PocketShareItem      `json:"pocket_share_por_compania"`
	EvolucionMensualPrimas    []EvolucionMensualItem `json:"evolucion_mensual_primas"`
}

type PocketShareItem struct {
	Gestora       string  `json:"gestora"`
	NumPolizas    int     `json:"num_polizas"`
	TotalPrimas   float64 `json:"total_primas"`
	Porcentaje    float64 `json:"porcentaje"`
}

type EvolucionMensualItem struct {
	Mes          string  `json:"mes"`
	TotalPrimas  float64 `json:"total_primas"`
	NumPolizas   int     `json:"num_polizas"`
}

// PortfolioAnalysisResponse - Análisis de cartera
type PortfolioAnalysisResponse struct {
	DistribucionPorRamo       []DistribucionRamoItem `json:"distribucion_por_ramo"`
	DistribucionPorProvincia  []DistribucionProvinciaItem `json:"distribucion_por_provincia"`
	Top10ClientesVsResto      Top10Analysis          `json:"top_10_clientes_vs_resto"`
	ConcentracionRiesgo       float64                `json:"concentracion_riesgo"`
	EvolucionCarteraMensual   []EvolucionCarteraItem `json:"evolucion_cartera_mensual"`
}

type DistribucionRamoItem struct {
	Ramo         string  `json:"ramo"`
	NumPolizas   int     `json:"num_polizas"`
	TotalPrimas  float64 `json:"total_primas"`
	Porcentaje   float64 `json:"porcentaje"`
}

type DistribucionProvinciaItem struct {
	Provincia    string  `json:"provincia"`
	NumClientes  int     `json:"num_clientes"`
	NumPolizas   int     `json:"num_polizas"`
	TotalPrimas  float64 `json:"total_primas"`
}

type Top10Analysis struct {
	Top10Primas    float64 `json:"top_10_primas"`
	RestoPrimas    float64 `json:"resto_primas"`
	Top10Count     int     `json:"top_10_count"`
	RestoCount     int     `json:"resto_count"`
	PorcentajeTop10 float64 `json:"porcentaje_top_10"`
}

type EvolucionCarteraItem struct {
	Mes            string `json:"mes"`
	NumClientes    int    `json:"num_clientes"`
	NumPolizas     int    `json:"num_polizas"`
}

// CollectionsPerformanceResponse - Rendimiento de cobros
type CollectionsPerformanceResponse struct {
	RatioCobroMes                float64                    `json:"ratio_cobro_mes"`
	RecibosDevueltosTendencia    []TendenciaDevueltosItem   `json:"recibos_devueltos_tendencia"`
	MorosidadPorRangoDias        []MorosidadRangoItem       `json:"morosidad_por_rango_dias"`
	DeudaTotalVsCartera          DeudaVsCarteraData         `json:"deuda_total_vs_cartera"`
	ClientesMorososRecurrentes   []ClienteMorosoRecurrente  `json:"clientes_morosos_recurrentes"`
}

type TendenciaDevueltosItem struct {
	Mes              string  `json:"mes"`
	NumDevueltos     int     `json:"num_devueltos"`
	ImporteDevuelto  float64 `json:"importe_devuelto"`
}

type MorosidadRangoItem struct {
	Rango           string  `json:"rango"`
	NumRecibos      int     `json:"num_recibos"`
	ImporteTotal    float64 `json:"importe_total"`
}

type DeudaVsCarteraData struct {
	DeudaTotal         float64 `json:"deuda_total"`
	PrimasCartera      float64 `json:"primas_cartera"`
	PorcentajeMorosidad float64 `json:"porcentaje_morosidad"`
}

type ClienteMorosoRecurrente struct {
	IDAccount          string  `json:"id_account"`
	NombreCompleto     string  `json:"nombre_completo"`
	NumDevoluciones    int     `json:"num_devoluciones"`
	DeudaTotal         float64 `json:"deuda_total"`
}

// ClaimsAnalysisResponse - Análisis de siniestros
type ClaimsAnalysisResponse struct {
	SiniestralididadPorRamo     []SiniestralididadRamoItem `json:"siniestralidad_por_ramo"`
	SiniestrosAbiertos          int                        `json:"siniestros_abiertos"`
	SiniestrosCerrados          int                        `json:"siniestros_cerrados"`
	TiempoMedioResolucion       float64                    `json:"tiempo_medio_resolucion_dias"`
	SiniestrosPorMes            []SiniestrosMesItem        `json:"siniestros_por_mes"`
}

type SiniestralididadRamoItem struct {
	Ramo                string  `json:"ramo"`
	NumPolizas          int     `json:"num_polizas"`
	NumSiniestros       int     `json:"num_siniestros"`
	Siniestralidad      float64 `json:"siniestralidad"`
}

type SiniestrosMesItem struct {
	Mes            string `json:"mes"`
	NumSiniestros  int    `json:"num_siniestros"`
}

// PerformanceTrendsResponse - Tendencias de rendimiento
type PerformanceTrendsResponse struct {
	EvolucionClientes            []TrendItem `json:"evolucion_clientes"`
	EvolucionPolizas             []TrendItem `json:"evolucion_polizas"`
	EvolucionPrimas              []TrendItem `json:"evolucion_primas"`
	ComparativaPeriodoAnterior   Comparative `json:"comparativa_periodo_anterior"`
}

type TrendItem struct {
	Fecha   string  `json:"fecha"`
	Valor   float64 `json:"valor"`
}

type Comparative struct {
	ClientesCambio  float64 `json:"clientes_cambio_porcentaje"`
	PolizasCambio   float64 `json:"polizas_cambio_porcentaje"`
	PrimasCambio    float64 `json:"primas_cambio_porcentaje"`
}

// ========================================
// HANDLERS DE ENDPOINTS
// ========================================

// GetFinancialKPIs - GET /api/analytics/financial-kpis
func GetFinancialKPIs(c *fiber.Ctx) error {
	var response FinancialKPIsResponse

	// Fechas para cálculos
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	lastMonth := now.AddDate(0, -1, 0).Month()

	// 1. Total primas mes actual
	err := db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND EXTRACT(MONTH FROM fecha_emision) = $2
		  AND activo = TRUE
	`, currentYear, currentMonth).Scan(&response.TotalPrimasMesActual)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando primas mes actual: %v", err)
	}

	// 2. Total primas mes anterior
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND EXTRACT(MONTH FROM fecha_emision) = $2
		  AND activo = TRUE
	`, now.AddDate(0, -1, 0).Year(), int(lastMonth)).Scan(&response.TotalPrimasMesAnterior)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando primas mes anterior: %v", err)
	}

	// 3. Total primas año actual
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND activo = TRUE
	`, currentYear).Scan(&response.TotalPrimasAnioActual)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando primas año actual: %v", err)
	}

	// 4. Total comisiones mes
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(comision_neta), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND EXTRACT(MONTH FROM fecha_emision) = $2
		  AND activo = TRUE
	`, currentYear, currentMonth).Scan(&response.TotalComisionesMes)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando comisiones mes: %v", err)
	}

	// 5. Total comisiones año
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(comision_neta), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND activo = TRUE
	`, currentYear).Scan(&response.TotalComisionesAnio)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando comisiones año: %v", err)
	}

	// 6. Promedio prima póliza
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(AVG(prima_anual), 0)
		FROM polizas
		WHERE situacion_poliza = 'Vigor'
		  AND activo = TRUE
	`).Scan(&response.PromedioPrimaPoliza)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error calculando promedio prima póliza: %v", err)
	}

	// 7. Margen comisión promedio
	if response.TotalPrimasMesActual > 0 {
		response.MargenComisionPromedio = (response.TotalComisionesMes / response.TotalPrimasMesActual) * 100
	}

	// 8. Pocket Share por compañía
	rows, err := db.PostgresDB.Query(`
		SELECT
			gestora,
			COUNT(*) as num_polizas,
			COALESCE(SUM(prima_anual), 0) as total_primas
		FROM polizas
		WHERE situacion_poliza = 'Vigor'
		  AND activo = TRUE
		  AND gestora IS NOT NULL
		  AND gestora != ''
		GROUP BY gestora
		ORDER BY total_primas DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		var totalPrimasGlobal float64 = 0
		var items []PocketShareItem

		for rows.Next() {
			var item PocketShareItem
			rows.Scan(&item.Gestora, &item.NumPolizas, &item.TotalPrimas)
			totalPrimasGlobal += item.TotalPrimas
			items = append(items, item)
		}

		// Calcular porcentajes
		for i := range items {
			if totalPrimasGlobal > 0 {
				items[i].Porcentaje = (items[i].TotalPrimas / totalPrimasGlobal) * 100
			}
		}
		response.PocketSharePorCompania = items
	}

	// 9. Evolución mensual primas (últimos 12 meses)
	rows, err = db.PostgresDB.Query(`
		SELECT
			TO_CHAR(fecha_emision, 'YYYY-MM') as mes,
			COALESCE(SUM(prima_total), 0) as total_primas,
			COUNT(*) as num_polizas
		FROM recibos
		WHERE fecha_emision >= NOW() - INTERVAL '12 months'
		  AND activo = TRUE
		GROUP BY mes
		ORDER BY mes ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item EvolucionMensualItem
			rows.Scan(&item.Mes, &item.TotalPrimas, &item.NumPolizas)
			response.EvolucionMensualPrimas = append(response.EvolucionMensualPrimas, item)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetPortfolioAnalysis - GET /api/analytics/portfolio-analysis
func GetPortfolioAnalysis(c *fiber.Ctx) error {
	var response PortfolioAnalysisResponse

	// 1. Distribución por ramo
	rows, err := db.PostgresDB.Query(`
		SELECT
			ramo,
			COUNT(*) as num_polizas,
			COALESCE(SUM(prima_anual), 0) as total_primas
		FROM polizas
		WHERE situacion_poliza = 'Vigor'
		  AND activo = TRUE
		  AND ramo IS NOT NULL
		  AND ramo != ''
		GROUP BY ramo
		ORDER BY total_primas DESC
	`)
	if err == nil {
		defer rows.Close()
		var totalPrimas float64 = 0
		var items []DistribucionRamoItem

		for rows.Next() {
			var item DistribucionRamoItem
			rows.Scan(&item.Ramo, &item.NumPolizas, &item.TotalPrimas)
			totalPrimas += item.TotalPrimas
			items = append(items, item)
		}

		// Calcular porcentajes
		for i := range items {
			if totalPrimas > 0 {
				items[i].Porcentaje = (items[i].TotalPrimas / totalPrimas) * 100
			}
		}
		response.DistribucionPorRamo = items
	}

	// 2. Distribución por provincia
	rows, err = db.PostgresDB.Query(`
		SELECT
			c.provincia,
			COUNT(DISTINCT c.id) as num_clientes,
			COUNT(DISTINCT p.id) as num_polizas,
			COALESCE(SUM(p.prima_anual), 0) as total_primas
		FROM clientes c
		LEFT JOIN polizas p ON c.id_account = p.id_account AND p.situacion_poliza = 'Vigor' AND p.activo = TRUE
		WHERE c.activo = TRUE
		  AND c.provincia IS NOT NULL
		  AND c.provincia != ''
		GROUP BY c.provincia
		ORDER BY total_primas DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item DistribucionProvinciaItem
			rows.Scan(&item.Provincia, &item.NumClientes, &item.NumPolizas, &item.TotalPrimas)
			response.DistribucionPorProvincia = append(response.DistribucionPorProvincia, item)
		}
	}

	// 3. Top 10 clientes vs resto
	var top10Analysis Top10Analysis

	// Top 10 primas
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(total_primas_cartera), 0), COUNT(*)
		FROM (
			SELECT total_primas_cartera
			FROM clientes
			WHERE activo = TRUE
			ORDER BY total_primas_cartera DESC
			LIMIT 10
		) as top10
	`).Scan(&top10Analysis.Top10Primas, &top10Analysis.Top10Count)

	// Resto primas
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(total_primas_cartera), 0), COUNT(*)
		FROM (
			SELECT total_primas_cartera
			FROM clientes
			WHERE activo = TRUE
			ORDER BY total_primas_cartera DESC
			OFFSET 10
		) as resto
	`).Scan(&top10Analysis.RestoPrimas, &top10Analysis.RestoCount)

	totalCartera := top10Analysis.Top10Primas + top10Analysis.RestoPrimas
	if totalCartera > 0 {
		top10Analysis.PorcentajeTop10 = (top10Analysis.Top10Primas / totalCartera) * 100
	}
	response.Top10ClientesVsResto = top10Analysis

	// 4. Concentración de riesgo (% top 20)
	response.ConcentracionRiesgo = top10Analysis.PorcentajeTop10 // Simplificado, puede extenderse a top 20

	// 5. Evolución cartera mensual (últimos 12 meses)
	rows, err = db.PostgresDB.Query(`
		SELECT
			TO_CHAR(mes, 'YYYY-MM') as mes,
			num_clientes,
			num_polizas
		FROM (
			SELECT
				DATE_TRUNC('month', creado_en) as mes,
				COUNT(DISTINCT CASE WHEN tipo = 'cliente' THEN id END) as num_clientes,
				COUNT(DISTINCT CASE WHEN tipo = 'poliza' THEN id END) as num_polizas
			FROM (
				SELECT id, creado_en, 'cliente' as tipo FROM clientes WHERE activo = TRUE
				UNION ALL
				SELECT id, creado_en, 'poliza' as tipo FROM polizas WHERE activo = TRUE
			) combined
			WHERE creado_en >= NOW() - INTERVAL '12 months'
			GROUP BY mes
		) monthly
		ORDER BY mes ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item EvolucionCarteraItem
			rows.Scan(&item.Mes, &item.NumClientes, &item.NumPolizas)
			response.EvolucionCarteraMensual = append(response.EvolucionCarteraMensual, item)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetCollectionsPerformance - GET /api/analytics/collections-performance
func GetCollectionsPerformance(c *fiber.Ctx) error {
	var response CollectionsPerformanceResponse

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// 1. Ratio de cobro mes
	var totalEmitidos, totalCobrados float64

	db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND EXTRACT(MONTH FROM fecha_emision) = $2
		  AND activo = TRUE
	`, currentYear, currentMonth).Scan(&totalEmitidos)

	db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE EXTRACT(YEAR FROM fecha_emision) = $1
		  AND EXTRACT(MONTH FROM fecha_emision) = $2
		  AND situacion_recibo = 'Cobrado'
		  AND activo = TRUE
	`, currentYear, currentMonth).Scan(&totalCobrados)

	if totalEmitidos > 0 {
		response.RatioCobroMes = (totalCobrados / totalEmitidos) * 100
	}

	// 2. Tendencia recibos devueltos (últimos 12 meses)
	rows, err := db.PostgresDB.Query(`
		SELECT
			TO_CHAR(fecha_situacion, 'YYYY-MM') as mes,
			COUNT(*) as num_devueltos,
			COALESCE(SUM(prima_total), 0) as importe_devuelto
		FROM recibos
		WHERE fecha_situacion >= NOW() - INTERVAL '12 months'
		  AND situacion_recibo = 'Retornado'
		  AND activo = TRUE
		GROUP BY mes
		ORDER BY mes ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item TendenciaDevueltosItem
			rows.Scan(&item.Mes, &item.NumDevueltos, &item.ImporteDevuelto)
			response.RecibosDevueltosTendencia = append(response.RecibosDevueltosTendencia, item)
		}
	}

	// 3. Morosidad por rango de días
	morosidadRangos := []MorosidadRangoItem{
		{Rango: "0-30 días", NumRecibos: 0, ImporteTotal: 0},
		{Rango: "30-60 días", NumRecibos: 0, ImporteTotal: 0},
		{Rango: "60-90 días", NumRecibos: 0, ImporteTotal: 0},
		{Rango: "90+ días", NumRecibos: 0, ImporteTotal: 0},
	}

	rows, err = db.PostgresDB.Query(`
		SELECT
			CASE
				WHEN EXTRACT(DAY FROM NOW() - fecha_situacion) <= 30 THEN '0-30 días'
				WHEN EXTRACT(DAY FROM NOW() - fecha_situacion) <= 60 THEN '30-60 días'
				WHEN EXTRACT(DAY FROM NOW() - fecha_situacion) <= 90 THEN '60-90 días'
				ELSE '90+ días'
			END as rango,
			COUNT(*) as num_recibos,
			COALESCE(SUM(prima_total), 0) as importe_total
		FROM recibos
		WHERE situacion_recibo = 'Retornado'
		  AND activo = TRUE
		GROUP BY rango
	`)
	if err == nil {
		defer rows.Close()
		rangosMap := make(map[string]*MorosidadRangoItem)
		for i := range morosidadRangos {
			rangosMap[morosidadRangos[i].Rango] = &morosidadRangos[i]
		}

		for rows.Next() {
			var rango string
			var numRecibos int
			var importeTotal float64
			rows.Scan(&rango, &numRecibos, &importeTotal)
			if item, ok := rangosMap[rango]; ok {
				item.NumRecibos = numRecibos
				item.ImporteTotal = importeTotal
			}
		}
	}
	response.MorosidadPorRangoDias = morosidadRangos

	// 4. Deuda total vs cartera
	var deudaData DeudaVsCarteraData

	db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0)
		FROM recibos
		WHERE situacion_recibo = 'Retornado'
		  AND activo = TRUE
	`).Scan(&deudaData.DeudaTotal)

	db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(total_primas_cartera), 0)
		FROM clientes
		WHERE activo = TRUE
	`).Scan(&deudaData.PrimasCartera)

	if deudaData.PrimasCartera > 0 {
		deudaData.PorcentajeMorosidad = (deudaData.DeudaTotal / deudaData.PrimasCartera) * 100
	}
	response.DeudaTotalVsCartera = deudaData

	// 5. Clientes morosos recurrentes
	rows, err = db.PostgresDB.Query(`
		SELECT
			r.id_account,
			c.nombre_completo,
			COUNT(*) as num_devoluciones,
			COALESCE(SUM(r.prima_total), 0) as deuda_total
		FROM recibos r
		LEFT JOIN clientes c ON r.id_account = c.id_account
		WHERE r.situacion_recibo = 'Retornado'
		  AND r.activo = TRUE
		  AND r.id_account IS NOT NULL
		GROUP BY r.id_account, c.nombre_completo
		HAVING COUNT(*) > 1
		ORDER BY deuda_total DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item ClienteMorosoRecurrente
			rows.Scan(&item.IDAccount, &item.NombreCompleto, &item.NumDevoluciones, &item.DeudaTotal)
			response.ClientesMorososRecurrentes = append(response.ClientesMorososRecurrentes, item)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetClaimsAnalysis - GET /api/analytics/claims-analysis
func GetClaimsAnalysis(c *fiber.Ctx) error {
	var response ClaimsAnalysisResponse

	// 1. Siniestralidad por ramo
	rows, err := db.PostgresDB.Query(`
		SELECT
			p.ramo,
			COUNT(DISTINCT p.id) as num_polizas,
			COUNT(DISTINCT s.id) as num_siniestros
		FROM polizas p
		LEFT JOIN siniestros s ON p.numero_poliza = s.numero_poliza AND s.activo = TRUE
		WHERE p.situacion_poliza = 'Vigor'
		  AND p.activo = TRUE
		  AND p.ramo IS NOT NULL
		  AND p.ramo != ''
		GROUP BY p.ramo
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY num_siniestros DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item SiniestralididadRamoItem
			rows.Scan(&item.Ramo, &item.NumPolizas, &item.NumSiniestros)
			if item.NumPolizas > 0 {
				item.Siniestralidad = (float64(item.NumSiniestros) / float64(item.NumPolizas)) * 100
			}
			response.SiniestralididadPorRamo = append(response.SiniestralididadPorRamo, item)
		}
	}

	// 2. Siniestros abiertos vs cerrados
	db.PostgresDB.QueryRow(`
		SELECT COUNT(*)
		FROM siniestros
		WHERE situacion_siniestro NOT IN ('Cerrado', 'Terminado', 'Finalizado')
		  AND activo = TRUE
	`).Scan(&response.SiniestrosAbiertos)

	db.PostgresDB.QueryRow(`
		SELECT COUNT(*)
		FROM siniestros
		WHERE situacion_siniestro IN ('Cerrado', 'Terminado', 'Finalizado')
		  AND activo = TRUE
	`).Scan(&response.SiniestrosCerrados)

	// 3. Tiempo medio de resolución (estimado)
	db.PostgresDB.QueryRow(`
		SELECT COALESCE(AVG(EXTRACT(DAY FROM NOW() - fecha_apertura)), 0)
		FROM siniestros
		WHERE situacion_siniestro IN ('Cerrado', 'Terminado', 'Finalizado')
		  AND activo = TRUE
		  AND fecha_apertura IS NOT NULL
	`).Scan(&response.TiempoMedioResolucion)

	// 4. Siniestros por mes (últimos 12 meses)
	rows, err = db.PostgresDB.Query(`
		SELECT
			TO_CHAR(fecha_ocurrencia, 'YYYY-MM') as mes,
			COUNT(*) as num_siniestros
		FROM siniestros
		WHERE fecha_ocurrencia >= NOW() - INTERVAL '12 months'
		  AND activo = TRUE
		GROUP BY mes
		ORDER BY mes ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item SiniestrosMesItem
			rows.Scan(&item.Mes, &item.NumSiniestros)
			response.SiniestrosPorMes = append(response.SiniestrosPorMes, item)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetPerformanceTrends - GET /api/analytics/performance-trends
func GetPerformanceTrends(c *fiber.Ctx) error {
	// Parámetro de periodo (7days, 30days, 90days, 12months, ytd, custom)
	period := c.Query("period", "30days")

	var interval string
	switch period {
	case "7days":
		interval = "7 days"
	case "30days":
		interval = "30 days"
	case "90days":
		interval = "90 days"
	case "12months":
		interval = "12 months"
	case "ytd":
		interval = fmt.Sprintf("%d days", time.Now().YearDay())
	default:
		interval = "30 days"
	}

	var response PerformanceTrendsResponse

	// 1. Evolución clientes
	rows, err := db.PostgresDB.Query(`
		SELECT
			TO_CHAR(DATE_TRUNC('day', creado_en), 'YYYY-MM-DD') as fecha,
			COUNT(*) as valor
		FROM clientes
		WHERE creado_en >= NOW() - INTERVAL '` + interval + `'
		  AND activo = TRUE
		GROUP BY fecha
		ORDER BY fecha ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item TrendItem
			rows.Scan(&item.Fecha, &item.Valor)
			response.EvolucionClientes = append(response.EvolucionClientes, item)
		}
	}

	// 2. Evolución pólizas
	rows, err = db.PostgresDB.Query(`
		SELECT
			TO_CHAR(DATE_TRUNC('day', creado_en), 'YYYY-MM-DD') as fecha,
			COUNT(*) as valor
		FROM polizas
		WHERE creado_en >= NOW() - INTERVAL '` + interval + `'
		  AND activo = TRUE
		GROUP BY fecha
		ORDER BY fecha ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item TrendItem
			rows.Scan(&item.Fecha, &item.Valor)
			response.EvolucionPolizas = append(response.EvolucionPolizas, item)
		}
	}

	// 3. Evolución primas
	rows, err = db.PostgresDB.Query(`
		SELECT
			TO_CHAR(DATE_TRUNC('day', fecha_emision), 'YYYY-MM-DD') as fecha,
			COALESCE(SUM(prima_total), 0) as valor
		FROM recibos
		WHERE fecha_emision >= NOW() - INTERVAL '` + interval + `'
		  AND activo = TRUE
		GROUP BY fecha
		ORDER BY fecha ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item TrendItem
			rows.Scan(&item.Fecha, &item.Valor)
			response.EvolucionPrimas = append(response.EvolucionPrimas, item)
		}
	}

	// 4. Comparativa con periodo anterior
	// TODO: Implementar cálculo de cambio porcentual vs periodo anterior
	response.ComparativaPeriodoAnterior = Comparative{
		ClientesCambio: 0,
		PolizasCambio:  0,
		PrimasCambio:   0,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
