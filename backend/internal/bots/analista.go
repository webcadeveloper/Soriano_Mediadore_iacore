package bots

import (
	"fmt"
	"soriano-mediadores/internal/db"
	"strings"
)

// BotAnalista - Bot Analista
// Funciones: Generar reportes, métricas, detectar oportunidades
type BotAnalista struct {
	ID   string
	Name string
}

func NewBotAnalista() *BotAnalista {
	return &BotAnalista{
		ID:   "bot_analista",
		Name: "Analista",
	}
}

// ProcesarConsulta procesa consultas de análisis
func (b *BotAnalista) ProcesarConsulta(sessionID string, mensaje string) (string, error) {
	return ProcesarConFallback(b.ID, sessionID, mensaje, func(msg string) (string, error) {
		mensajeLower := strings.ToLower(msg)

		if strings.Contains(mensajeLower, "top") || strings.Contains(mensajeLower, "mejores") {
			return b.TopClientes()
		} else if strings.Contains(mensajeLower, "ramo") || strings.Contains(mensajeLower, "producto") {
			return b.AnalisisPorRamo()
		} else if strings.Contains(mensajeLower, "comision") {
			return b.AnalisisComisiones()
		}

		return b.ReporteGeneral()
	})
}

// TopClientes lista los mejores clientes
func (b *BotAnalista) TopClientes() (string, error) {
	query := `
		SELECT
			c.nombre_completo,
			c.id_account,
			c.total_primas_cartera,
			c.total_primas_relacion,
			COUNT(DISTINCT p.id) as num_polizas
		FROM clientes c
		LEFT JOIN polizas p ON c.id_account = p.id_account AND p.activo = TRUE
		WHERE c.activo = TRUE
		GROUP BY c.id, c.nombre_completo, c.id_account, c.total_primas_cartera, c.total_primas_relacion
		ORDER BY c.total_primas_cartera DESC
		LIMIT 20
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("🏆 TOP 20 CLIENTES POR VOLUMEN DE PRIMAS\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0
	var totalPrimas, totalComisiones float64

	for rows.Next() {
		var nombre, idAccount string
		var primas *float64
		var comisiones *float64
		var numPolizas int

		err := rows.Scan(&nombre, &idAccount, &primas, &comisiones, &numPolizas)
		if err != nil {
			continue
		}

		count++

		primasVal := 0.0
		if primas != nil {
			primasVal = *primas
			totalPrimas += primasVal
		}

		comisionesVal := 0.0
		if comisiones != nil {
			comisionesVal = *comisiones
			totalComisiones += comisionesVal
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, nombre))
		sb.WriteString(fmt.Sprintf("   ID: %s\n", idAccount))
		sb.WriteString(fmt.Sprintf("   Pólizas: %d\n", numPolizas))
		sb.WriteString(fmt.Sprintf("   💰 Primas: €%.2f\n", primasVal))
		sb.WriteString(fmt.Sprintf("   💵 Comisiones: €%.2f\n\n", comisionesVal))
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total Primas Top 20: €%.2f\n", totalPrimas))
	sb.WriteString(fmt.Sprintf("Total Comisiones Top 20: €%.2f\n", totalComisiones))

	return sb.String(), nil
}

// AnalisisPorRamo analiza distribución por tipo de seguro
func (b *BotAnalista) AnalisisPorRamo() (string, error) {
	query := `
		SELECT
			ramo,
			COUNT(*) as num_polizas
		FROM polizas
		WHERE activo = TRUE AND ramo IS NOT NULL
		GROUP BY ramo
		ORDER BY COUNT(*) DESC
		LIMIT 15
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📊 ANÁLISIS POR RAMO DE SEGURO\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0
	var totalPolizas int

	for rows.Next() {
		var ramo string
		var numPolizas int

		err := rows.Scan(&ramo, &numPolizas)
		if err != nil {
			continue
		}

		count++
		totalPolizas += numPolizas

		// Truncar ramo si es muy largo
		ramoDisplay := ramo
		if len(ramo) > 40 {
			ramoDisplay = ramo[:37] + "..."
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, ramoDisplay))
		sb.WriteString(fmt.Sprintf("   Pólizas: %d\n\n", numPolizas))
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total Pólizas: %d\n", totalPolizas))

	return sb.String(), nil
}

// AnalisisComisiones analiza comisiones
func (b *BotAnalista) AnalisisComisiones() (string, error) {
	var stats struct {
		TotalComisionesBrutas float64
		TotalComisionesNetas  float64
		NumRecibosCobrados    int
	}

	db.PostgresDB.QueryRow(`
		SELECT
			COALESCE(SUM(comision_bruta), 0) as bruta,
			COALESCE(SUM(comision_neta), 0) as neta,
			COUNT(*) as cobrados
		FROM recibos
		WHERE situacion_recibo = 'Cobrado' AND activo = TRUE
	`).Scan(&stats.TotalComisionesBrutas, &stats.TotalComisionesNetas, &stats.NumRecibosCobrados)

	var sb strings.Builder
	sb.WriteString("💰 ANÁLISIS DE COMISIONES\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString(fmt.Sprintf("Recibos cobrados: %d\n\n", stats.NumRecibosCobrados))
	sb.WriteString(fmt.Sprintf("Comisiones brutas: €%.2f\n", stats.TotalComisionesBrutas))
	sb.WriteString(fmt.Sprintf("Comisiones netas: €%.2f\n", stats.TotalComisionesNetas))

	if stats.TotalComisionesBrutas > 0 {
		pctNeta := (stats.TotalComisionesNetas / stats.TotalComisionesBrutas) * 100
		sb.WriteString(fmt.Sprintf("\nRetención neta: %.1f%%\n", pctNeta))
	}

	return sb.String(), nil
}

// ReporteGeneral genera un reporte general del sistema
func (b *BotAnalista) ReporteGeneral() (string, error) {
	var stats struct {
		TotalClientes   int
		TotalPolizas    int
		TotalRecibos    int
		TotalSiniestros int
		TotalPrimas     float64
		TotalComisiones float64
	}

	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE").Scan(&stats.TotalClientes)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM polizas WHERE activo = TRUE").Scan(&stats.TotalPolizas)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM recibos WHERE activo = TRUE").Scan(&stats.TotalRecibos)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM siniestros WHERE activo = TRUE").Scan(&stats.TotalSiniestros)
	db.PostgresDB.QueryRow("SELECT COALESCE(SUM(prima_total), 0) FROM recibos WHERE activo = TRUE").Scan(&stats.TotalPrimas)
	db.PostgresDB.QueryRow("SELECT COALESCE(SUM(comision_neta), 0) FROM recibos WHERE situacion_recibo = 'Cobrado' AND activo = TRUE").Scan(&stats.TotalComisiones)

	var sb strings.Builder
	sb.WriteString("📊 REPORTE GENERAL DEL SISTEMA\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString("📈 Datos Generales:\n")
	sb.WriteString(fmt.Sprintf("   Clientes: %d\n", stats.TotalClientes))
	sb.WriteString(fmt.Sprintf("   Pólizas: %d\n", stats.TotalPolizas))
	sb.WriteString(fmt.Sprintf("   Recibos: %d\n", stats.TotalRecibos))
	sb.WriteString(fmt.Sprintf("   Siniestros: %d\n\n", stats.TotalSiniestros))

	sb.WriteString("💰 Financiero:\n")
	sb.WriteString(fmt.Sprintf("   Total Primas: €%.2f\n", stats.TotalPrimas))
	sb.WriteString(fmt.Sprintf("   Total Comisiones: €%.2f\n", stats.TotalComisiones))

	if stats.TotalClientes > 0 {
		avgPolizas := float64(stats.TotalPolizas) / float64(stats.TotalClientes)
		sb.WriteString(fmt.Sprintf("\n📊 Promedios:\n"))
		sb.WriteString(fmt.Sprintf("   Pólizas por cliente: %.2f\n", avgPolizas))
	}

	return sb.String(), nil
}
