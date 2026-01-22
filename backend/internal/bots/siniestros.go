package bots

import (
	"fmt"
	"soriano-mediadores/internal/ai"
	"soriano-mediadores/internal/db"
	"strings"
)

// BotSiniestros - Bot de Gestión de Siniestros para SORIANO MEDIADORES (Occident)
// Funciones: Registrar siniestros, consultar estado, asesorar documentación, notificar clientes
type BotSiniestros struct {
	ID   string
	Name string
}

func NewBotSiniestros() *BotSiniestros {
	return &BotSiniestros{
		ID:   "bot_siniestros",
		Name: "Gestor de Siniestros - Soriano Mediadores",
	}
}

// ProcesarConsulta procesa consultas de siniestros
func (b *BotSiniestros) ProcesarConsulta(sessionID string, mensaje string) (string, error) {
	return ProcesarConFallback(b.ID, sessionID, mensaje, func(msg string) (string, error) {
		mensajeLower := strings.ToLower(msg)

		if strings.Contains(mensajeLower, "abierto") || strings.Contains(mensajeLower, "pendiente") {
			return b.ListarSiniestrosAbiertos()
		} else if strings.Contains(mensajeLower, "estadistica") || strings.Contains(mensajeLower, "resumen") {
			return b.ResumenSiniestros()
		} else if strings.Contains(mensajeLower, "tramitador") {
			return b.EstadisticasPorTramitador()
		} else if strings.Contains(mensajeLower, "documento") || strings.Contains(mensajeLower, "necesito") || strings.Contains(mensajeLower, "parte") {
			return b.AsesorarDocumentacion(msg)
		}

		return b.ResumenSiniestros()
	})
}

// AsesorarDocumentacion asesora sobre documentación necesaria para siniestros
func (b *BotSiniestros) AsesorarDocumentacion(consulta string) (string, error) {
	systemPrompt := `Eres el DEPARTAMENTO DE SINIESTROS de SORIANO MEDIADORES, correduría de seguros española
colaboradora exclusiva de GRUPO OCCIDENT (Catalana Occidente, Plus Ultra, Seguros Bilbao, NorteHispana).

INFORMACIÓN DE LA EMPRESA:
- Nombre: Soriano Mediadores
- Lema: "Somos mediadores de seguros confiables"
- Filosofía: "Queremos ser parte de tu familia"
- Sede: Calle Constitución 5, Villajoyosa, 03570 (Alicante)
- Teléfono: +34 96 681 02 90
- Email: info@sorianomediadores.es
- Web: www.sorianomediadores.es
- Horario: Lunes a Domingo de 09:00 a 17:00
- Oficinas en: Alicante (sede), Barcelona, Valladolid, Valencia

PROCESO DE TRAMITACIÓN DE SINIESTROS EN ESPAÑA:

1. COMUNICACIÓN DEL SINIESTRO:
   - Plazo legal: 7 días desde conocimiento (Art. 16 Ley Contrato de Seguro)
   - Excepciones: Robo (24-48h), Fallecimiento (inmediato)
   - Canales: Teléfono 24h compañía, app, email, presencial en correduría

2. DOCUMENTACIÓN POR TIPO DE SINIESTRO:

AUTOMÓVIL:
- Parte amistoso de accidente (DAA) firmado por ambas partes
- Fotografías del siniestro y daños
- DNI del conductor y tomador
- Permiso de circulación y ficha técnica
- Carné de conducir vigente
- Atestado policial (si intervino policía)
- Informe médico (si hay lesiones)
- Facturas de reparación (si ya reparado)

HOGAR:
- Fotografías de los daños
- Facturas o presupuestos de reparación
- Denuncia policial (robo, vandalismo)
- Informe de bomberos (incendio)
- Parte de la comunidad (daños por agua de vecino)
- Facturas de objetos dañados/robados
- Informe de cerrajero (robo con fuerza)

SALUD:
- Informe médico detallado
- Pruebas diagnósticas
- Facturas de tratamiento (reembolso)
- Autorización previa (hospitalización programada)
- Tarjeta sanitaria de la compañía

VIDA/FALLECIMIENTO:
- Certificado de defunción
- Certificado médico de causa de muerte
- DNI del fallecido
- Póliza original
- Libro de familia
- Certificado de últimas voluntades
- Testamento (si existe)
- DNI de beneficiarios

ACCIDENTES:
- Parte de accidente (trabajo, tráfico, doméstico)
- Informe médico de urgencias
- Partes de baja/alta laboral
- Informe de secuelas (si procede)

RESPONSABILIDAD CIVIL:
- Reclamación del tercero
- Fotografías de los daños
- Presupuestos de reparación
- Testigos (si los hay)

3. TELÉFONOS DE ASISTENCIA 24H GRUPO OCCIDENT:
- Catalana Occidente: 900 300 400
- Plus Ultra: 900 103 283
- Seguros Bilbao: 900 101 600
- NorteHispana: 900 100 247

4. PLAZOS IMPORTANTES:
- Comunicación: 7 días
- Aportación documentación: 10 días tras solicitud
- Respuesta compañía: 40 días (Art. 18 LCS)
- Intereses de demora: Si supera 3 meses

5. CONSEJOS PRÁCTICOS:
- NUNCA firmar documentos sin leer
- SIEMPRE hacer fotos ANTES de reparar
- CONSERVAR facturas y tickets originales
- NO admitir culpabilidad ante terceros
- ANOTAR datos de testigos

Responde de forma clara, práctica y en español de España. Indica siempre los documentos específicos necesarios.`

	respuesta, err := ai.ConsultarAI(consulta, systemPrompt)
	if err != nil {
		return "Error procesando consulta de siniestros. Contacte con el departamento de siniestros en horario de oficina.", err
	}

	return respuesta, nil
}

// ListarSiniestrosAbiertos lista siniestros en estado abierto
func (b *BotSiniestros) ListarSiniestrosAbiertos() (string, error) {
	query := `
		SELECT
			s.numero_siniestro,
			s.nombre_cliente,
			s.id_account,
			s.numero_poliza,
			s.fecha_ocurrencia,
			s.fecha_apertura,
			s.tramitador,
			p.ramo
		FROM siniestros s
		LEFT JOIN polizas p ON s.numero_poliza = p.numero_poliza
		WHERE s.situacion_siniestro = 'Abierto'
		  AND s.activo = TRUE
		ORDER BY s.fecha_ocurrencia DESC
		LIMIT 50
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📋 SINIESTROS ABIERTOS (En proceso)\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0

	for rows.Next() {
		var numeroSiniestro, nombreCliente, idAccount string
		var numeroPoliza *string
		var fechaOcurrencia *string
		var fechaApertura *string
		var tramitador *string
		var ramo *string

		err := rows.Scan(&numeroSiniestro, &nombreCliente, &idAccount, &numeroPoliza,
			&fechaOcurrencia, &fechaApertura, &tramitador, &ramo)
		if err != nil {
			continue
		}

		count++

		sb.WriteString(fmt.Sprintf("%d. Siniestro: %s\n", count, numeroSiniestro))
		sb.WriteString(fmt.Sprintf("   Cliente: %s (ID: %s)\n", nombreCliente, idAccount))

		if numeroPoliza != nil {
			sb.WriteString(fmt.Sprintf("   Póliza: %s\n", *numeroPoliza))
		}
		if ramo != nil {
			sb.WriteString(fmt.Sprintf("   Ramo: %s\n", *ramo))
		}
		if fechaOcurrencia != nil {
			sb.WriteString(fmt.Sprintf("   Fecha ocurrencia: %s\n", *fechaOcurrencia))
		}
		if tramitador != nil {
			sb.WriteString(fmt.Sprintf("   Tramitador: %s\n", *tramitador))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total siniestros abiertos: %d\n", count))

	return sb.String(), nil
}

// ResumenSiniestros genera resumen estadístico
func (b *BotSiniestros) ResumenSiniestros() (string, error) {
	var stats struct {
		Total    int
		Abiertos int
		Cerrados int
		Anulados int
	}

	db.PostgresDB.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN situacion_siniestro = 'Abierto' THEN 1 END) as abiertos,
			COUNT(CASE WHEN situacion_siniestro = 'Cerrado' THEN 1 END) as cerrados,
			COUNT(CASE WHEN situacion_siniestro = 'Anulado' THEN 1 END) as anulados
		FROM siniestros
		WHERE activo = TRUE
	`).Scan(&stats.Total, &stats.Abiertos, &stats.Cerrados, &stats.Anulados)

	var sb strings.Builder
	sb.WriteString("📊 RESUMEN DE SINIESTROS\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString(fmt.Sprintf("Total siniestros: %d\n\n", stats.Total))

	if stats.Total > 0 {
		pctCerrados := float64(stats.Cerrados) / float64(stats.Total) * 100
		pctAbiertos := float64(stats.Abiertos) / float64(stats.Total) * 100

		sb.WriteString(fmt.Sprintf("✅ Cerrados: %d (%.1f%%)\n", stats.Cerrados, pctCerrados))
		sb.WriteString(fmt.Sprintf("⏳ Abiertos: %d (%.1f%%)\n", stats.Abiertos, pctAbiertos))
		sb.WriteString(fmt.Sprintf("❌ Anulados: %d\n", stats.Anulados))

		if stats.Abiertos > 0 {
			sb.WriteString(fmt.Sprintf("\n⚡ ATENCIÓN: %d siniestros pendientes de resolver\n", stats.Abiertos))
		}
	}

	return sb.String(), nil
}

// EstadisticasPorTramitador muestra estadísticas por tramitador
func (b *BotSiniestros) EstadisticasPorTramitador() (string, error) {
	query := `
		SELECT
			tramitador,
			COUNT(*) as total,
			COUNT(CASE WHEN situacion_siniestro = 'Abierto' THEN 1 END) as abiertos,
			COUNT(CASE WHEN situacion_siniestro = 'Cerrado' THEN 1 END) as cerrados
		FROM siniestros
		WHERE activo = TRUE AND tramitador IS NOT NULL
		GROUP BY tramitador
		ORDER BY total DESC
		LIMIT 20
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📊 SINIESTROS POR TRAMITADOR\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0

	for rows.Next() {
		var tramitador string
		var total, abiertos, cerrados int

		err := rows.Scan(&tramitador, &total, &abiertos, &cerrados)
		if err != nil {
			continue
		}

		count++

		pctResolucion := 0.0
		if total > 0 {
			pctResolucion = float64(cerrados) / float64(total) * 100
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, tramitador))
		sb.WriteString(fmt.Sprintf("   Total casos: %d\n", total))
		sb.WriteString(fmt.Sprintf("   ⏳ Abiertos: %d\n", abiertos))
		sb.WriteString(fmt.Sprintf("   ✅ Cerrados: %d (%.1f%%)\n\n", cerrados, pctResolucion))
	}

	return sb.String(), nil
}
