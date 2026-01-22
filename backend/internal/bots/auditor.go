package bots

import (
	"database/sql"
	"fmt"
	"soriano-mediadores/internal/db"
	"strings"
)

// BotAuditor - Bot Auditor de Calidad
// Funciones: Detectar duplicados, validar datos, controlar calidad
type BotAuditor struct {
	ID   string
	Name string
}

func NewBotAuditor() *BotAuditor {
	return &BotAuditor{
		ID:   "bot_auditor",
		Name: "Auditor de Calidad",
	}
}

// ProcesarConsulta procesa consultas de auditoría
func (b *BotAuditor) ProcesarConsulta(sessionID string, mensaje string) (string, error) {
	// Usar sistema de fallback primero
	return ProcesarConFallback(b.ID, sessionID, mensaje, func(msg string) (string, error) {
		// Si fallback no encuentra match, procesar con lógica original
		mensajeLower := strings.ToLower(msg)

		if strings.Contains(mensajeLower, "duplicado") {
			return b.DetectarDuplicados()
		} else if strings.Contains(mensajeLower, "calidad") || strings.Contains(mensajeLower, "integridad") {
			return b.AnalisisCalidadDatos()
		} else if strings.Contains(mensajeLower, "huerfano") || strings.Contains(mensajeLower, "fk") {
			return b.DetectarDatosHuerfanos()
		}

		return b.AuditoriaGeneral()
	})
}

// DetectarDuplicados busca posibles duplicados
func (b *BotAuditor) DetectarDuplicados() (string, error) {
	query := `
		SELECT
			nombre_completo,
			COUNT(*) as duplicados,
			STRING_AGG(id_account, ', ') as ids
		FROM clientes
		WHERE activo = TRUE
		GROUP BY LOWER(TRIM(nombre_completo))
		HAVING COUNT(*) > 1
		ORDER BY duplicados DESC
		LIMIT 20
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("🔍 CLIENTES DUPLICADOS (POSIBLES)\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0
	totalDuplicados := 0

	for rows.Next() {
		var nombre, ids string
		var numDuplicados int

		err := rows.Scan(&nombre, &numDuplicados, &ids)
		if err != nil {
			continue
		}

		count++
		totalDuplicados += numDuplicados

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, nombre))
		sb.WriteString(fmt.Sprintf("   Ocurrencias: %d\n", numDuplicados))
		sb.WriteString(fmt.Sprintf("   IDs: %s\n\n", ids))
	}

	if count == 0 {
		sb.WriteString("✅ No se detectaron duplicados evidentes\n")
	} else {
		sb.WriteString(strings.Repeat("=", 50) + "\n")
		sb.WriteString(fmt.Sprintf("⚠️  Total grupos duplicados: %d\n", count))
		sb.WriteString(fmt.Sprintf("⚠️  Total registros afectados: %d\n", totalDuplicados))
	}

	return sb.String(), nil
}

// AnalisisCalidadDatos analiza la calidad de los datos
func (b *BotAuditor) AnalisisCalidadDatos() (string, error) {
	var stats struct {
		TotalClientes      int
		ClientesSinEmail   int
		ClientesSinTel     int
		ClientesSinDir     int
		PolizasSinCliente  int
		RecibosSinPoliza   int
		SiniestrosSinPoliza int
	}

	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE").Scan(&stats.TotalClientes)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE AND (email_contacto IS NULL OR email_contacto = '')").Scan(&stats.ClientesSinEmail)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE AND (telefono_contacto IS NULL OR telefono_contacto = '')").Scan(&stats.ClientesSinTel)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE AND (domicilio IS NULL OR domicilio = '')").Scan(&stats.ClientesSinDir)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM polizas WHERE activo = TRUE AND id_account IS NULL").Scan(&stats.PolizasSinCliente)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM recibos WHERE activo = TRUE AND numero_poliza IS NULL").Scan(&stats.RecibosSinPoliza)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM siniestros WHERE activo = TRUE AND numero_poliza IS NULL").Scan(&stats.SiniestrosSinPoliza)

	var sb strings.Builder
	sb.WriteString("📊 ANÁLISIS DE CALIDAD DE DATOS\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString("👥 Clientes:\n")
	sb.WriteString(fmt.Sprintf("   Total: %d\n", stats.TotalClientes))

	if stats.TotalClientes > 0 {
		pctEmail := float64(stats.TotalClientes-stats.ClientesSinEmail) / float64(stats.TotalClientes) * 100
		pctTel := float64(stats.TotalClientes-stats.ClientesSinTel) / float64(stats.TotalClientes) * 100
		pctDir := float64(stats.TotalClientes-stats.ClientesSinDir) / float64(stats.TotalClientes) * 100

		sb.WriteString(fmt.Sprintf("   Con email: %.1f%%\n", pctEmail))
		sb.WriteString(fmt.Sprintf("   Con teléfono: %.1f%%\n", pctTel))
		sb.WriteString(fmt.Sprintf("   Con dirección: %.1f%%\n\n", pctDir))
	}

	sb.WriteString("🔗 Integridad Referencial:\n")
	sb.WriteString(fmt.Sprintf("   ⚠️  Pólizas sin cliente: %d\n", stats.PolizasSinCliente))
	sb.WriteString(fmt.Sprintf("   ⚠️  Recibos sin póliza: %d\n", stats.RecibosSinPoliza))
	sb.WriteString(fmt.Sprintf("   ⚠️  Siniestros sin póliza: %d\n", stats.SiniestrosSinPoliza))

	totalProblemas := stats.PolizasSinCliente + stats.RecibosSinPoliza + stats.SiniestrosSinPoliza
	if totalProblemas > 0 {
		sb.WriteString(fmt.Sprintf("\n⚠️  Total problemas de integridad: %d\n", totalProblemas))
	} else {
		sb.WriteString("\n✅ Integridad referencial perfecta\n")
	}

	return sb.String(), nil
}

// DetectarDatosHuerfanos detecta registros sin referencias
func (b *BotAuditor) DetectarDatosHuerfanos() (string, error) {
	var sb strings.Builder
	sb.WriteString("🔍 DATOS HUÉRFANOS (Sin Referencias)\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Pólizas sin cliente
	var polizasSinCliente int
	db.PostgresDB.QueryRow(`
		SELECT COUNT(*)
		FROM polizas
		WHERE activo = TRUE AND id_account IS NULL
	`).Scan(&polizasSinCliente)

	sb.WriteString(fmt.Sprintf("📋 Pólizas sin cliente vinculado: %d\n", polizasSinCliente))

	if polizasSinCliente > 0 && polizasSinCliente <= 10 {
		rows, _ := db.PostgresDB.Query(`
			SELECT numero_poliza, nombre_cliente, ramo
			FROM polizas
			WHERE activo = TRUE AND id_account IS NULL
			LIMIT 10
		`)
		defer rows.Close()

		for rows.Next() {
			var numeroPoliza, nombreCliente string
			var ramo sql.NullString
			rows.Scan(&numeroPoliza, &nombreCliente, &ramo)

			sb.WriteString(fmt.Sprintf("   - %s: %s\n", numeroPoliza, nombreCliente))
		}
	}

	// Recibos sin póliza
	var recibosSinPoliza int
	db.PostgresDB.QueryRow(`
		SELECT COUNT(*)
		FROM recibos
		WHERE activo = TRUE AND numero_poliza IS NULL
	`).Scan(&recibosSinPoliza)

	sb.WriteString(fmt.Sprintf("\n💳 Recibos sin póliza vinculada: %d\n", recibosSinPoliza))

	// Siniestros sin póliza
	var siniestrosSinPoliza int
	db.PostgresDB.QueryRow(`
		SELECT COUNT(*)
		FROM siniestros
		WHERE activo = TRUE AND numero_poliza IS NULL
	`).Scan(&siniestrosSinPoliza)

	sb.WriteString(fmt.Sprintf("🚨 Siniestros sin póliza vinculada: %d\n", siniestrosSinPoliza))

	return sb.String(), nil
}

// AuditoriaGeneral genera auditoría completa
func (b *BotAuditor) AuditoriaGeneral() (string, error) {
	var sb strings.Builder

	// Combinar todos los análisis
	duplicados, _ := b.DetectarDuplicados()
	calidad, _ := b.AnalisisCalidadDatos()

	sb.WriteString("🔍 AUDITORÍA GENERAL DEL SISTEMA\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString(calidad)
	sb.WriteString("\n\n")
	sb.WriteString(duplicados)

	return sb.String(), nil
}
