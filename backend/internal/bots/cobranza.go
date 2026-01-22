package bots

import (
	"fmt"
	"soriano-mediadores/internal/ai"
	"soriano-mediadores/internal/db"
	"strings"
	"time"
)

// BotCobranza - Bot de Gestión de Cobranza para SORIANO MEDIADORES (Occident)
// Funciones: Identificar recibos pendientes, generar listas de contacto, seguimiento, mensajes de recobro
type BotCobranza struct {
	ID   string
	Name string
}

func NewBotCobranza() *BotCobranza {
	return &BotCobranza{
		ID:   "bot_cobranza",
		Name: "Gestor de Cobranza - Soriano Mediadores",
	}
}

// ProcesarConsulta procesa consultas relacionadas con cobranza
func (b *BotCobranza) ProcesarConsulta(sessionID string, mensaje string) (string, error) {
	return ProcesarConFallback(b.ID, sessionID, mensaje, func(msg string) (string, error) {
		mensajeLower := strings.ToLower(msg)

		// Detectar tipo de consulta
		if strings.Contains(mensajeLower, "pendiente") || strings.Contains(mensajeLower, "impagado") {
			return b.ListarRecibosPendientes(100)
		} else if strings.Contains(mensajeLower, "vencido") || strings.Contains(mensajeLower, "atrasado") {
			return b.ListarRecibosVencidos()
		} else if strings.Contains(mensajeLower, "contacto") || strings.Contains(mensajeLower, "llamar") {
			return b.GenerarListaContacto()
		} else if strings.Contains(mensajeLower, "estadistica") || strings.Contains(mensajeLower, "resumen") {
			return b.ResumenCobranza()
		} else if strings.Contains(mensajeLower, "mensaje") || strings.Contains(mensajeLower, "email") || strings.Contains(mensajeLower, "carta") {
			return b.GenerarMensajeRecobro(msg)
		}

		return b.ResumenCobranza()
	})
}

// GenerarMensajeRecobro genera un mensaje de recobro personalizado usando AI
func (b *BotCobranza) GenerarMensajeRecobro(contexto string) (string, error) {
	systemPrompt := `Eres el departamento de GESTIÓN DE COBROS de SORIANO MEDIADORES, correduría de seguros española
colaboradora exclusiva de GRUPO OCCIDENT.

INFORMACIÓN DE LA EMPRESA:
- Nombre: Soriano Mediadores
- Lema: "Somos mediadores de seguros confiables"
- Filosofía: "Queremos ser parte de tu familia"
- Sede: Calle Constitución 5, Villajoyosa, 03570 (Alicante)
- Teléfono: +34 96 681 02 90
- Email: info@sorianomediadores.es / cobros@sorianomediadores.es
- Web: www.sorianomediadores.es
- Horario: Lunes a Domingo de 09:00 a 17:00
- Oficinas en: Alicante (sede), Barcelona, Valladolid, Valencia
- Cobertura: Toda España

VALORES DE LA EMPRESA:
1. "Prometer es cumplir" - Trabajo meticuloso
2. "Experiencia" - Más de 30 años en el sector
3. "La transparencia no se negocia"

CONTEXTO DE LA EMPRESA:
- Correduría con más de 30 años de experiencia
- Trabajamos con Catalana Occidente, Plus Ultra, Seguros Bilbao y NorteHispana
- Sede en Villajoyosa (Alicante), operamos en toda España

GENERA MENSAJES DE RECOBRO según el nivel indicado:

NIVEL 1 - RECORDATORIO AMABLE (1-15 días):
- Tono: Cordial, servicial
- Asunto: "Recordatorio de pago - Póliza [RAMO]"
- Contenido: Recordar vencimiento, ofrecer ayuda, facilitar formas de pago
- Incluir: Datos bancarios para transferencia, teléfono de contacto

NIVEL 2 - AVISO FORMAL (16-30 días):
- Tono: Profesional, firme pero respetuoso
- Asunto: "Aviso importante - Recibo pendiente"
- Contenido: Informar de la situación, advertir posible suspensión de cobertura
- Incluir: Consecuencias del impago, plazo para regularizar

NIVEL 3 - REQUERIMIENTO (31-60 días):
- Tono: Formal, serio
- Asunto: "Requerimiento de pago - Acción necesaria"
- Contenido: Advertir suspensión inminente, mencionar posibles recargos
- Incluir: Fecha límite, consecuencias legales posibles

NIVEL 4 - ÚLTIMA NOTIFICACIÓN (>60 días):
- Tono: Muy formal
- Asunto: "Última notificación antes de anulación"
- Contenido: Informar anulación inminente, pérdida de bonificaciones
- Incluir: Opción de fraccionamiento de deuda si procede

NORMATIVA ESPAÑOLA A CONSIDERAR:
- Art. 15 Ley Contrato de Seguro: Impago de prima
- Período de gracia: 1 mes desde vencimiento
- Suspensión de cobertura: A partir del mes de impago
- Resolución del contrato: A los 6 meses de impago

FORMATO DEL MENSAJE:
- Siempre incluir: Nombre cliente, número de recibo, importe, fecha vencimiento
- Firmar como: "Departamento de Gestión de Cobros - SORIANO MEDIADORES"
- Incluir: Teléfono 96X XXX XXX, email cobros@sorianomediadores.es
- Datos bancarios: ES00 0000 0000 0000 0000 0000 (Concepto: Nº Recibo)

Genera el mensaje apropiado según el contexto proporcionado.`

	respuesta, err := ai.ConsultarAI(contexto, systemPrompt)
	if err != nil {
		return "Error generando mensaje de recobro. Por favor, contacte con el departamento de cobros.", err
	}

	return respuesta, nil
}

// ListarRecibosPendientes lista recibos pendientes de pago
func (b *BotCobranza) ListarRecibosPendientes(limit int) (string, error) {
	query := `
		SELECT
			r.numero_recibo,
			r.id_account,
			r.prima_total,
			r.fecha_emision,
			r.nombre_cliente,
			r.numero_poliza
		FROM recibos r
		WHERE r.situacion_recibo = 'Pendiente'
		  AND r.activo = TRUE
		ORDER BY r.fecha_emision ASC
		LIMIT $1
	`

	rows, err := db.PostgresDB.Query(query, limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📋 RECIBOS PENDIENTES DE COBRO\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0
	totalImporte := 0.0

	for rows.Next() {
		var numeroRecibo, idAccount, nombreCliente string
		var numeroPoliza, fechaEmision *string
		var primaTotal float64

		err := rows.Scan(&numeroRecibo, &idAccount, &primaTotal, &fechaEmision, &nombreCliente, &numeroPoliza)
		if err != nil {
			continue
		}

		count++
		totalImporte += primaTotal

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, nombreCliente))
		sb.WriteString(fmt.Sprintf("   Recibo: %s\n", numeroRecibo))
		sb.WriteString(fmt.Sprintf("   Importe: €%.2f\n", primaTotal))
		if fechaEmision != nil {
			sb.WriteString(fmt.Sprintf("   Fecha: %s\n", *fechaEmision))
		}
		sb.WriteString(fmt.Sprintf("   ID Cliente: %s\n\n", idAccount))
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total recibos pendientes: %d\n", count))
	sb.WriteString(fmt.Sprintf("💰 Importe total pendiente: €%.2f\n", totalImporte))

	// Guardar métrica
	db.GuardarMetrica("recibos_pendientes", map[string]interface{}{
		"cantidad": count,
		"total":    totalImporte,
		"fecha":    time.Now(),
	})

	return sb.String(), nil
}

// ListarRecibosVencidos lista recibos vencidos (más de 30 días)
func (b *BotCobranza) ListarRecibosVencidos() (string, error) {
	query := `
		SELECT
			r.numero_recibo,
			r.nombre_cliente,
			r.prima_total,
			r.fecha_emision,
			r.id_account,
			c.telefono,
			c.email
		FROM recibos r
		LEFT JOIN clientes c ON r.id_account = c.id_account
		WHERE r.situacion_recibo = 'Pendiente'
		  AND r.fecha_emision < CURRENT_DATE - INTERVAL '30 days'
		  AND r.activo = TRUE
		ORDER BY r.fecha_emision ASC
		LIMIT 50
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("⚠️  RECIBOS VENCIDOS (más de 30 días)\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0
	totalImporte := 0.0

	for rows.Next() {
		var numeroRecibo, nombreCliente, idAccount string
		var primaTotal float64
		var fechaEmision *string
		var telefono *string
		var email *string

		err := rows.Scan(&numeroRecibo, &nombreCliente, &primaTotal, &fechaEmision, &idAccount, &telefono, &email)
		if err != nil {
			continue
		}

		count++
		totalImporte += primaTotal

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, nombreCliente))
		sb.WriteString(fmt.Sprintf("   Recibo: %s\n", numeroRecibo))
		sb.WriteString(fmt.Sprintf("   Importe: €%.2f\n", primaTotal))
		if fechaEmision != nil {
			sb.WriteString(fmt.Sprintf("   Vencido desde: %s\n", *fechaEmision))
		}
		if telefono != nil && *telefono != "" {
			sb.WriteString(fmt.Sprintf("   📞 Tel: %s\n", *telefono))
		}
		if email != nil && *email != "" {
			sb.WriteString(fmt.Sprintf("   📧 Email: %s\n", *email))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("⚠️  Total recibos vencidos: %d\n", count))
	sb.WriteString(fmt.Sprintf("💰 Importe total: €%.2f\n", totalImporte))

	return sb.String(), nil
}

// GenerarListaContacto genera lista de clientes para contactar
func (b *BotCobranza) GenerarListaContacto() (string, error) {
	query := `
		SELECT DISTINCT
			c.id_account,
			c.nombre_completo,
			c.telefono,
			c.email,
			COUNT(r.id) as recibos_pendientes,
			SUM(r.prima_total) as total_deuda
		FROM clientes c
		INNER JOIN recibos r ON c.id_account = r.id_account
		WHERE r.situacion_recibo = 'Pendiente'
		  AND r.activo = TRUE
		  AND c.activo = TRUE
		GROUP BY c.id_account, c.nombre_completo, c.telefono, c.email
		HAVING COUNT(r.id) > 0
		ORDER BY total_deuda DESC
		LIMIT 30
	`

	rows, err := db.PostgresDB.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📞 LISTA DE CONTACTO - RECIBOS PENDIENTES\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	count := 0

	for rows.Next() {
		var idAccount, nombreCompleto string
		var telefono *string
		var email *string
		var recibosPendientes int
		var totalDeuda float64

		err := rows.Scan(&idAccount, &nombreCompleto, &telefono, &email, &recibosPendientes, &totalDeuda)
		if err != nil {
			continue
		}

		count++

		sb.WriteString(fmt.Sprintf("%d. %s\n", count, nombreCompleto))
		sb.WriteString(fmt.Sprintf("   ID: %s\n", idAccount))
		sb.WriteString(fmt.Sprintf("   Recibos pendientes: %d\n", recibosPendientes))
		sb.WriteString(fmt.Sprintf("   💰 Deuda total: €%.2f\n", totalDeuda))

		if telefono != nil && *telefono != "" {
			sb.WriteString(fmt.Sprintf("   📞 LLAMAR: %s\n", *telefono))
		} else {
			sb.WriteString("   📞 Sin teléfono\n")
		}

		if email != nil && *email != "" {
			sb.WriteString(fmt.Sprintf("   📧 Email: %s\n", *email))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total clientes a contactar: %d\n", count))

	return sb.String(), nil
}

// ResumenCobranza genera resumen estadístico
func (b *BotCobranza) ResumenCobranza() (string, error) {
	var stats struct {
		TotalRecibos          int
		RecibosCobrados       int
		RecibosPendientes     int
		RecibosAnulados       int
		ImporteTotalPendiente float64
		ImporteTotalCobrado   float64
	}

	// Estadísticas generales
	db.PostgresDB.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN situacion_recibo = 'Cobrado' THEN 1 END) as cobrados,
			COUNT(CASE WHEN situacion_recibo = 'Pendiente' THEN 1 END) as pendientes,
			COUNT(CASE WHEN situacion_recibo = 'Anulado' THEN 1 END) as anulados,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Pendiente' THEN prima_total ELSE 0 END), 0) as total_pendiente,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Cobrado' THEN prima_total ELSE 0 END), 0) as total_cobrado
		FROM recibos
		WHERE activo = TRUE
	`).Scan(
		&stats.TotalRecibos,
		&stats.RecibosCobrados,
		&stats.RecibosPendientes,
		&stats.RecibosAnulados,
		&stats.ImporteTotalPendiente,
		&stats.ImporteTotalCobrado,
	)

	var sb strings.Builder
	sb.WriteString("📊 RESUMEN DE COBRANZA\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString(fmt.Sprintf("Total recibos: %d\n\n", stats.TotalRecibos))

	if stats.TotalRecibos > 0 {
		pctCobrado := float64(stats.RecibosCobrados) / float64(stats.TotalRecibos) * 100
		pctPendiente := float64(stats.RecibosPendientes) / float64(stats.TotalRecibos) * 100

		sb.WriteString(fmt.Sprintf("✅ Cobrados: %d (%.1f%%)\n", stats.RecibosCobrados, pctCobrado))
		sb.WriteString(fmt.Sprintf("⏳ Pendientes: %d (%.1f%%)\n", stats.RecibosPendientes, pctPendiente))
		sb.WriteString(fmt.Sprintf("❌ Anulados: %d\n\n", stats.RecibosAnulados))

		sb.WriteString(strings.Repeat("-", 50) + "\n\n")
		sb.WriteString(fmt.Sprintf("💰 Importe cobrado: €%.2f\n", stats.ImporteTotalCobrado))
		sb.WriteString(fmt.Sprintf("⚠️  Importe pendiente: €%.2f\n", stats.ImporteTotalPendiente))

		if stats.ImporteTotalPendiente > 0 {
			sb.WriteString(fmt.Sprintf("\n⚡ ACCIÓN REQUERIDA: %.2f€ por cobrar\n", stats.ImporteTotalPendiente))
		}
	}

	return sb.String(), nil
}
