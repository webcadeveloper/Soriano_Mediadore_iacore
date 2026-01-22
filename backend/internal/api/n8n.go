package api

import (
	"encoding/json"
	"fmt"
	"log"
	"soriano-mediadores/internal/db"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// N8NWebhookRequest representa una solicitud de webhook desde N8N
type N8NWebhookRequest struct {
	WorkflowID   string                 `json:"workflow_id"`
	ExecutionID  string                 `json:"execution_id"`
	Action       string                 `json:"action"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    string                 `json:"timestamp"`
}

// N8NWebhookResponse representa la respuesta al webhook
type N8NWebhookResponse struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ExecutionID string                 `json:"execution_id"`
	Timestamp   string                 `json:"timestamp"`
}

// N8NClienteCreado webhook para cuando se crea un cliente
func N8NClienteCreado(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	// Extraer datos del cliente
	nif, _ := req.Data["nif"].(string)
	nombre, _ := req.Data["nombre_completo"].(string)
	email, _ := req.Data["email"].(string)
	telefono, _ := req.Data["telefono"].(string)

	if nif == "" || nombre == "" {
		return sendN8NError(c, "NIF y nombre son requeridos", nil)
	}

	// Insertar cliente
	sql := `
		INSERT INTO clientes (
			nif, id_account, nombre_completo, email_contacto, telefono_contacto,
			activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, TRUE, NOW())
		RETURNING id
	`

	idAccount := uuid.New().String()
	var clienteID int
	err := db.PostgresDB.QueryRow(sql, nif, idAccount, nombre, email, telefono).Scan(&clienteID)
	if err != nil {
		return sendN8NError(c, "Error creando cliente", err)
	}

	log.Printf("📥 N8N: Cliente creado desde workflow %s - ID: %d", req.WorkflowID, clienteID)

	return sendN8NSuccess(c, "Cliente creado exitosamente", map[string]interface{}{
		"cliente_id": clienteID,
		"id_account": idAccount,
		"nif":        nif,
		"nombre":     nombre,
	})
}

// N8NPolizaCreada webhook para cuando se crea una póliza
func N8NPolizaCreada(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	numeroPoliza, _ := req.Data["numero_poliza"].(string)
	idAccount, _ := req.Data["id_account"].(string)
	ramo, _ := req.Data["ramo"].(string)
	primaAnual, _ := req.Data["prima_anual"].(string)

	if numeroPoliza == "" || idAccount == "" {
		return sendN8NError(c, "numero_poliza e id_account son requeridos", nil)
	}

	sql := `
		INSERT INTO polizas (
			numero_poliza, id_account, nombre_cliente, ramo, prima_anual,
			situacion, activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, 'Vigente', TRUE, NOW())
		RETURNING id
	`

	nombreCliente, _ := req.Data["nombre_cliente"].(string)
	var polizaID int
	err := db.PostgresDB.QueryRow(sql, numeroPoliza, idAccount, nombreCliente, ramo, primaAnual).Scan(&polizaID)
	if err != nil {
		return sendN8NError(c, "Error creando póliza", err)
	}

	log.Printf("📥 N8N: Póliza creada desde workflow %s - ID: %d", req.WorkflowID, polizaID)

	return sendN8NSuccess(c, "Póliza creada exitosamente", map[string]interface{}{
		"poliza_id":      polizaID,
		"numero_poliza":  numeroPoliza,
		"id_account":     idAccount,
	})
}

// N8NReciboCreado webhook para cuando se crea un recibo
func N8NReciboCreado(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	numeroRecibo, _ := req.Data["numero_recibo"].(string)
	numeroPoliza, _ := req.Data["numero_poliza"].(string)
	primaTotal, _ := req.Data["prima_total"].(float64)

	if numeroRecibo == "" || numeroPoliza == "" {
		return sendN8NError(c, "numero_recibo y numero_poliza son requeridos", nil)
	}

	sql := `
		INSERT INTO recibos (
			numero_recibo, numero_poliza, id_account, prima_total,
			situacion_recibo, activo, creado_en
		) VALUES ($1, $2, $3, $4, 'Pendiente', TRUE, NOW())
		RETURNING id
	`

	idAccount, _ := req.Data["id_account"].(string)
	var reciboID int
	err := db.PostgresDB.QueryRow(sql, numeroRecibo, numeroPoliza, idAccount, primaTotal).Scan(&reciboID)
	if err != nil {
		return sendN8NError(c, "Error creando recibo", err)
	}

	log.Printf("📥 N8N: Recibo creado desde workflow %s - ID: %d", req.WorkflowID, reciboID)

	return sendN8NSuccess(c, "Recibo creado exitosamente", map[string]interface{}{
		"recibo_id":      reciboID,
		"numero_recibo":  numeroRecibo,
		"numero_poliza":  numeroPoliza,
		"prima_total":    primaTotal,
	})
}

// N8NSiniestroCreado webhook para cuando se crea un siniestro
func N8NSiniestroCreado(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	numeroSiniestro, _ := req.Data["numero_siniestro"].(string)
	numeroPoliza, _ := req.Data["numero_poliza"].(string)

	if numeroSiniestro == "" || numeroPoliza == "" {
		return sendN8NError(c, "numero_siniestro y numero_poliza son requeridos", nil)
	}

	sql := `
		INSERT INTO siniestros (
			numero_siniestro, numero_poliza, id_account, situacion_siniestro,
			fecha_ocurrencia, activo, creado_en
		) VALUES ($1, $2, $3, 'Abierto', NOW(), TRUE, NOW())
		RETURNING id
	`

	idAccount, _ := req.Data["id_account"].(string)
	var siniestroID int
	err := db.PostgresDB.QueryRow(sql, numeroSiniestro, numeroPoliza, idAccount).Scan(&siniestroID)
	if err != nil {
		return sendN8NError(c, "Error creando siniestro", err)
	}

	log.Printf("📥 N8N: Siniestro creado desde workflow %s - ID: %d", req.WorkflowID, siniestroID)

	return sendN8NSuccess(c, "Siniestro creado exitosamente", map[string]interface{}{
		"siniestro_id":      siniestroID,
		"numero_siniestro":  numeroSiniestro,
		"numero_poliza":     numeroPoliza,
	})
}

// N8NConsultaCliente webhook para consultar datos de un cliente
func N8NConsultaCliente(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	query, _ := req.Data["query"].(string)
	if query == "" {
		return sendN8NError(c, "query requerido (NIF, nombre o ID)", nil)
	}

	clientes, err := db.BuscarClientes(query, 10)
	if err != nil {
		return sendN8NError(c, "Error buscando clientes", err)
	}

	log.Printf("📥 N8N: Consulta de cliente desde workflow %s - Resultados: %d", req.WorkflowID, len(clientes))

	return sendN8NSuccess(c, "Clientes encontrados", map[string]interface{}{
		"total":    len(clientes),
		"clientes": clientes,
	})
}

// N8NConsultaPolizas webhook para consultar pólizas de un cliente
func N8NConsultaPolizas(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	idAccount, _ := req.Data["id_account"].(string)
	if idAccount == "" {
		return sendN8NError(c, "id_account requerido", nil)
	}

	polizas, err := db.ObtenerPolizasCliente(idAccount)
	if err != nil {
		return sendN8NError(c, "Error obteniendo pólizas", err)
	}

	log.Printf("📥 N8N: Consulta de pólizas desde workflow %s - ID Account: %s", req.WorkflowID, idAccount)

	return sendN8NSuccess(c, "Pólizas encontradas", map[string]interface{}{
		"total":   len(polizas),
		"polizas": polizas,
	})
}

// N8NConsultaRecibos webhook para consultar recibos pendientes
func N8NConsultaRecibos(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	idAccount, _ := req.Data["id_account"].(string)
	situacion, _ := req.Data["situacion"].(string)

	if idAccount == "" {
		return sendN8NError(c, "id_account requerido", nil)
	}

	recibos, err := db.ObtenerRecibosCliente(idAccount, 50)
	if err != nil {
		return sendN8NError(c, "Error obteniendo recibos", err)
	}

	// Filtrar por situación si se especifica
	var filtrados []db.Recibo
	if situacion != "" {
		for _, r := range recibos {
			if r.Situacion == situacion {
				filtrados = append(filtrados, r)
			}
		}
	} else {
		filtrados = recibos
	}

	log.Printf("📥 N8N: Consulta de recibos desde workflow %s - ID Account: %s", req.WorkflowID, idAccount)

	return sendN8NSuccess(c, "Recibos encontrados", map[string]interface{}{
		"total":   len(filtrados),
		"recibos": filtrados,
	})
}

// N8NEstadisticas webhook para obtener estadísticas generales
func N8NEstadisticas(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	var stats struct {
		TotalClientes   int     `json:"total_clientes"`
		TotalPolizas    int     `json:"total_polizas"`
		TotalRecibos    int     `json:"total_recibos"`
		TotalSiniestros int     `json:"total_siniestros"`
		PrimasTotal     float64 `json:"primas_total"`
	}

	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE").Scan(&stats.TotalClientes)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM polizas WHERE activo = TRUE").Scan(&stats.TotalPolizas)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM recibos WHERE activo = TRUE").Scan(&stats.TotalRecibos)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM siniestros WHERE activo = TRUE").Scan(&stats.TotalSiniestros)
	db.PostgresDB.QueryRow("SELECT COALESCE(SUM(total_primas_cartera), 0) FROM clientes WHERE activo = TRUE").Scan(&stats.PrimasTotal)

	log.Printf("📥 N8N: Solicitud de estadísticas desde workflow %s", req.WorkflowID)

	return sendN8NSuccess(c, "Estadísticas obtenidas", map[string]interface{}{
		"estadisticas": stats,
	})
}

// N8NActualizarCliente webhook para actualizar datos de un cliente
func N8NActualizarCliente(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	idAccount, _ := req.Data["id_account"].(string)
	if idAccount == "" {
		return sendN8NError(c, "id_account requerido", nil)
	}

	// Construir UPDATE dinámico
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if email, ok := req.Data["email"].(string); ok && email != "" {
		updates = append(updates, fmt.Sprintf("email_contacto = $%d", argPos))
		args = append(args, email)
		argPos++
	}

	if telefono, ok := req.Data["telefono"].(string); ok && telefono != "" {
		updates = append(updates, fmt.Sprintf("telefono_contacto = $%d", argPos))
		args = append(args, telefono)
		argPos++
	}

	if domicilio, ok := req.Data["domicilio"].(string); ok && domicilio != "" {
		updates = append(updates, fmt.Sprintf("domicilio = $%d", argPos))
		args = append(args, domicilio)
		argPos++
	}

	if len(updates) == 0 {
		return sendN8NError(c, "No se proporcionaron campos para actualizar", nil)
	}

	updates = append(updates, "actualizado_en = NOW()")
	args = append(args, idAccount)

	sql := fmt.Sprintf(`
		UPDATE clientes SET %s
		WHERE id_account = $%d
	`, strings.Join(updates, ", "), argPos)

	result, err := db.PostgresDB.Exec(sql, args...)
	if err != nil {
		return sendN8NError(c, "Error actualizando cliente", err)
	}

	rowsAffected, _ := result.RowsAffected()

	log.Printf("📥 N8N: Cliente actualizado desde workflow %s - ID Account: %s", req.WorkflowID, idAccount)

	return sendN8NSuccess(c, "Cliente actualizado exitosamente", map[string]interface{}{
		"id_account":     idAccount,
		"rows_affected":  rowsAffected,
	})
}

// N8NNotificarCliente webhook para enviar notificación a un cliente
func N8NNotificarCliente(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	idAccount, _ := req.Data["id_account"].(string)
	mensaje, _ := req.Data["mensaje"].(string)
	tipoNotificacion, _ := req.Data["tipo"].(string) // email, sms, whatsapp

	if idAccount == "" || mensaje == "" {
		return sendN8NError(c, "id_account y mensaje son requeridos", nil)
	}

	// En producción, aquí se enviaría la notificación real
	// Por ahora, solo registramos en MongoDB

	notification := map[string]interface{}{
		"id_account":  idAccount,
		"tipo":        tipoNotificacion,
		"mensaje":     mensaje,
		"enviado_en":  time.Now(),
		"workflow_id": req.WorkflowID,
		"estado":      "enviado",
	}

	notificationJSON, _ := json.Marshal(notification)

	db.GuardarMetrica("notificacion_enviada", map[string]interface{}{
		"data": string(notificationJSON),
	})

	log.Printf("📥 N8N: Notificación enviada desde workflow %s - Cliente: %s, Tipo: %s",
		req.WorkflowID, idAccount, tipoNotificacion)

	return sendN8NSuccess(c, "Notificación enviada exitosamente", map[string]interface{}{
		"id_account": idAccount,
		"tipo":       tipoNotificacion,
		"mensaje":    mensaje,
	})
}

// N8NWebhookGenerico webhook genérico para acciones personalizadas
func N8NWebhookGenerico(c *fiber.Ctx) error {
	var req N8NWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return sendN8NError(c, "JSON inválido", err)
	}

	// Guardar evento en MongoDB para auditoría
	db.GuardarMetrica("webhook_n8n", map[string]interface{}{
		"workflow_id":  req.WorkflowID,
		"execution_id": req.ExecutionID,
		"action":       req.Action,
		"data":         req.Data,
	})

	log.Printf("📥 N8N: Webhook genérico recibido - Workflow: %s, Action: %s",
		req.WorkflowID, req.Action)

	return sendN8NSuccess(c, "Webhook procesado exitosamente", map[string]interface{}{
		"action":      req.Action,
		"workflow_id": req.WorkflowID,
	})
}

// Helper functions
func sendN8NSuccess(c *fiber.Ctx, message string, data map[string]interface{}) error {
	return c.JSON(N8NWebhookResponse{
		Success:     true,
		Message:     message,
		Data:        data,
		ExecutionID: uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
	})
}

func sendN8NError(c *fiber.Ctx, message string, err error) error {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	return c.Status(400).JSON(N8NWebhookResponse{
		Success:     false,
		Message:     message,
		Error:       errorMsg,
		ExecutionID: uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
	})
}
