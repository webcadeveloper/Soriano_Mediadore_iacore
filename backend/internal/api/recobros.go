package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"soriano-mediadores/internal/db"
	"soriano-mediadores/internal/email"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ReciboEmailRequest - Request para enviar email de recobro
type ReciboEmailRequest struct {
	ReciboID     string `json:"recibo_id"`
	ClienteEmail string `json:"cliente_email"`
	TemplateID   string `json:"template_id"`
	From         string `json:"from"`
	Subject      string `json:"subject"`
	HTMLBody     string `json:"html_body"`
}

// BulkEmailRequest - Request para envío masivo de emails
type BulkEmailRequest struct {
	From   string              `json:"from"`
	Emails []SingleEmailData   `json:"emails"`
}

// SingleEmailData - Datos de un solo email
type SingleEmailData struct {
	ReciboID string `json:"recibo_id"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
}

// EmailResponse - Respuesta del endpoint de email
type EmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// SendReciboEmail - Envía un email de recobro a un cliente
func SendReciboEmail(c *fiber.Ctx) error {
	var req ReciboEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Error al parsear request: " + err.Error(),
		})
	}

	// Validaciones
	if req.ClienteEmail == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El email del cliente es requerido",
		})
	}

	if req.Subject == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El asunto del email es requerido",
		})
	}

	if req.HTMLBody == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El cuerpo HTML del email es requerido",
		})
	}

	if req.From == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El remitente (from) es requerido",
		})
	}

	// Crear cliente de Graph
	graphClient, err := email.NewGraphClient()
	if err != nil {
		log.Printf("❌ Error creando cliente de Graph: %v", err)
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error de configuración de Microsoft Graph: " + err.Error(),
		})
	}

	// Enviar email
	if err := graphClient.SendEmail(req.From, req.ClienteEmail, req.Subject, req.HTMLBody); err != nil {
		log.Printf("❌ Error enviando email a %s: %v", req.ClienteEmail, err)
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error al enviar email: " + err.Error(),
		})
	}

	log.Printf("✅ Email enviado exitosamente a %s (Recibo: %s)", req.ClienteEmail, req.ReciboID)

	return c.JSON(EmailResponse{
		Success: true,
		Message: fmt.Sprintf("Email enviado exitosamente a %s", req.ClienteEmail),
		Data: map[string]any{
			"recibo_id":     req.ReciboID,
			"cliente_email": req.ClienteEmail,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	})
}

// SendBulkReciboEmails - Envía emails masivos de recobro
func SendBulkReciboEmails(c *fiber.Ctx) error {
	var req BulkEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Error al parsear request: " + err.Error(),
		})
	}

	if req.From == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El remitente (from) es requerido",
		})
	}

	if len(req.Emails) == 0 {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "No se proporcionaron emails para enviar",
		})
	}

	// Crear cliente de Graph
	graphClient, err := email.NewGraphClient()
	if err != nil {
		log.Printf("❌ Error creando cliente de Graph: %v", err)
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error de configuración de Microsoft Graph: " + err.Error(),
		})
	}

	// Preparar datos para envío masivo
	bulkEmails := make([]struct {
		To      string
		Subject string
		Body    string
	}, len(req.Emails))

	for i, email := range req.Emails {
		bulkEmails[i] = struct {
			To      string
			Subject string
			Body    string
		}{
			To:      email.To,
			Subject: email.Subject,
			Body:    email.Body,
		}
	}

	// Enviar emails
	log.Printf("📧 Enviando %d emails masivos...", len(bulkEmails))
	errors, err := graphClient.SendBulkEmails(req.From, bulkEmails)

	if err != nil {
		// Hubo errores, pero algunos pueden haber sido exitosos
		errorMessages := make([]string, 0)
		if errors != nil {
			for _, e := range errors {
				errorMessages = append(errorMessages, e.Error())
			}
		}

		return c.Status(207).JSON(EmailResponse{ // 207 Multi-Status
			Success: false,
			Message: fmt.Sprintf("Se completó con %d errores", len(errors)),
			Data: map[string]any{
				"total":          len(req.Emails),
				"errors_count":   len(errors),
				"success_count":  len(req.Emails) - len(errors),
				"error_messages": errorMessages,
			},
		})
	}

	log.Printf("✅ Todos los emails enviados exitosamente (%d)", len(req.Emails))

	return c.JSON(EmailResponse{
		Success: true,
		Message: fmt.Sprintf("Todos los %d emails fueron enviados exitosamente", len(req.Emails)),
		Data: map[string]any{
			"total":     len(req.Emails),
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

// TestGraphConnection - Endpoint para probar la conexión con Microsoft Graph
func TestGraphConnection(c *fiber.Ctx) error {
	graphClient, err := email.NewGraphClient()
	if err != nil {
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error de configuración: " + err.Error(),
		})
	}

	// Intentar obtener un token
	if err := graphClient.GetAccessToken(); err != nil {
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error al autenticar con Microsoft Graph: " + err.Error(),
		})
	}

	return c.JSON(EmailResponse{
		Success: true,
		Message: "Conexión exitosa con Microsoft Graph",
		Data: map[string]any{
			"tenant_id":    graphClient.TenantID,
			"client_id":    graphClient.ClientID,
			"token_expiry": graphClient.TokenExpiry.Format(time.RFC3339),
		},
	})
}

// GetEmailTemplates - Obtiene las plantillas de email (desde localStorage del frontend)
// Este endpoint es opcional, ya que las plantillas están en el frontend
func GetEmailTemplates(c *fiber.Ctx) error {
	// Por ahora, las plantillas están en el frontend (localStorage)
	// Este endpoint podría usarse en el futuro para sincronizar plantillas
	return c.JSON(EmailResponse{
		Success: true,
		Message: "Las plantillas se gestionan desde el frontend",
		Data: map[string]any{
			"info": "Las plantillas HTML están almacenadas en localStorage del navegador",
		},
	})
}

// ReplaceVariables - Función auxiliar para reemplazar variables en plantillas
func ReplaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// SendReciboEmailWithTemplate - Envía un email de recobro usando plantilla
func SendReciboEmailWithTemplate(c *fiber.Ctx) error {
	type TemplateEmailRequest struct {
		From               string `json:"from"`
		To                 string `json:"to"`
		TemplateNumber     int    `json:"template_number"` // 1, 2, o 3
		NumeroRecibo       string `json:"numero_recibo"`   // REQUERIDO - Se usa para buscar en BD

		// Campos opcionales - si no se proveen, se obtienen de la BD
		NombreCliente      string `json:"nombre_cliente,omitempty"`
		NumeroPoliza       string `json:"numero_poliza,omitempty"`
		Ramo               string `json:"ramo,omitempty"`
		Tomador            string `json:"tomador,omitempty"`
		DescripcionRiesgo  string `json:"descripcion_riesgo,omitempty"`
		MotivoDevolucion   string `json:"motivo_devolucion,omitempty"`
		Importe            string `json:"importe,omitempty"`
	}

	var req TemplateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Error al parsear request: " + err.Error(),
		})
	}

	// Validaciones
	if req.From == "" || req.To == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Los campos 'from' y 'to' son requeridos",
		})
	}

	if req.NumeroRecibo == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "El campo 'numero_recibo' es requerido",
		})
	}

	if req.TemplateNumber < 1 || req.TemplateNumber > 3 {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "template_number debe ser 1, 2 o 3",
		})
	}

	// Si los campos no fueron proporcionados, obtenerlos de la base de datos
	if req.NombreCliente == "" || req.Ramo == "" || req.Tomador == "" || req.DescripcionRiesgo == "" {
		// Consultar recibo desde la base de datos
		query := `
			SELECT
				r.nombre_cliente,
				r.numero_poliza,
				r.ramo,
				r.descripcion_riesgo,
				r.prima_total,
				r.detalle_recibo,
				r.situacion_recibo
			FROM recibos r
			WHERE r.numero_recibo = $1 AND r.activo = true
			LIMIT 1
		`

		var nombreCliente, numeroPoliza, ramo, descripcionRiesgo, detalleRecibo, situacion string
		var primaTotal float64

		err := db.PostgresDB.QueryRow(query, req.NumeroRecibo).Scan(
			&nombreCliente,
			&numeroPoliza,
			&ramo,
			&descripcionRiesgo,
			&primaTotal,
			&detalleRecibo,
			&situacion,
		)

		if err != nil {
			log.Printf("❌ Error consultando recibo %s: %v", req.NumeroRecibo, err)
			return c.Status(404).JSON(EmailResponse{
				Success: false,
				Message: fmt.Sprintf("No se encontró el recibo %s en la base de datos", req.NumeroRecibo),
			})
		}

		// Asignar valores obtenidos de la BD si no fueron proporcionados
		if req.NombreCliente == "" {
			req.NombreCliente = nombreCliente
		}
		if req.NumeroPoliza == "" {
			req.NumeroPoliza = numeroPoliza
		}
		if req.Ramo == "" {
			req.Ramo = ramo
		}
		if req.Tomador == "" {
			req.Tomador = nombreCliente // El tomador es el nombre del cliente
		}
		if req.DescripcionRiesgo == "" {
			req.DescripcionRiesgo = descripcionRiesgo
		}
		if req.MotivoDevolucion == "" {
			req.MotivoDevolucion = detalleRecibo
		}
		if req.Importe == "" {
			req.Importe = fmt.Sprintf("%.2f", primaTotal)
		}

		log.Printf("📋 Datos obtenidos de BD para recibo %s: Cliente=%s, Póliza=%s, Ramo=%s",
			req.NumeroRecibo, nombreCliente, numeroPoliza, ramo)
	}

	// Leer plantilla
	templatePath := filepath.Join("templates", "email", fmt.Sprintf("plantilla_email_recobro_%d.html", req.TemplateNumber))
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("❌ Error leyendo plantilla %s: %v", templatePath, err)
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error al cargar plantilla: " + err.Error(),
		})
	}

	// Reemplazar variables
	htmlBody := string(templateContent)
	htmlBody = strings.ReplaceAll(htmlBody, "{{NOMBRE_CLIENTE}}", req.NombreCliente)
	htmlBody = strings.ReplaceAll(htmlBody, "{{NUMERO_RECIBO}}", req.NumeroRecibo)
	htmlBody = strings.ReplaceAll(htmlBody, "{{NUMERO_POLIZA}}", req.NumeroPoliza)
	htmlBody = strings.ReplaceAll(htmlBody, "{{RAMO}}", req.Ramo)
	htmlBody = strings.ReplaceAll(htmlBody, "{{TOMADOR}}", req.Tomador)
	htmlBody = strings.ReplaceAll(htmlBody, "{{DESCRIPCION_RIESGO}}", req.DescripcionRiesgo)
	htmlBody = strings.ReplaceAll(htmlBody, "{{MOTIVO_DEVOLUCION}}", req.MotivoDevolucion)
	htmlBody = strings.ReplaceAll(htmlBody, "{{IMPORTE}}", req.Importe)

	// Definir asuntos según plantilla
	subjects := map[int]string{
		1: "Aviso Importante - Recibo Devuelto",
		2: "Recordatorio Urgente - Pago Pendiente",
		3: "ÚLTIMO AVISO - Anulación de Póliza",
	}

	subject := subjects[req.TemplateNumber]

	// Crear cliente de Graph
	graphClient, err := email.NewGraphClient()
	if err != nil {
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error de configuración: " + err.Error(),
		})
	}

	// Enviar email
	if err := graphClient.SendEmail(req.From, req.To, subject, htmlBody); err != nil {
		log.Printf("❌ Error enviando email a %s: %v", req.To, err)
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error al enviar email: " + err.Error(),
		})
	}

	log.Printf("✅ Email enviado exitosamente a %s (Recibo: %s, Plantilla: %d)", req.To, req.NumeroRecibo, req.TemplateNumber)

	return c.JSON(EmailResponse{
		Success: true,
		Message: fmt.Sprintf("Email enviado exitosamente usando plantilla %d", req.TemplateNumber),
		Data: map[string]any{
			"from":            req.From,
			"to":              req.To,
			"template":        req.TemplateNumber,
			"numero_recibo":   req.NumeroRecibo,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
	})
}

// SendTestEmail - Envía un email de prueba
func SendTestEmail(c *fiber.Ctx) error {
	type TestEmailRequest struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req TestEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Error al parsear request: " + err.Error(),
		})
	}

	// Validaciones
	if req.From == "" || req.To == "" {
		return c.Status(400).JSON(EmailResponse{
			Success: false,
			Message: "Los campos 'from' y 'to' son requeridos",
		})
	}

	// Crear cliente de Graph
	graphClient, err := email.NewGraphClient()
	if err != nil {
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error de configuración: " + err.Error(),
		})
	}

	// Si no hay body, usar uno por defecto
	body := req.Body
	if body == "" {
		body = `
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"></head>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h1 style="color: #c2185b;">Email de Prueba</h1>
			<p>Este es un email de prueba enviado desde Soriano Mediadores CRM.</p>
			<p>Si recibes este mensaje, la configuración de Microsoft Graph está funcionando correctamente.</p>
			<hr>
			<p style="color: #666; font-size: 12px;">Enviado el: ` + time.Now().Format("02/01/2006 15:04:05") + `</p>
		</body>
		</html>
		`
	}

	subject := req.Subject
	if subject == "" {
		subject = "Test Email - Soriano Mediadores"
	}

	// Enviar email
	if err := graphClient.SendEmail(req.From, req.To, subject, body); err != nil {
		return c.Status(500).JSON(EmailResponse{
			Success: false,
			Message: "Error al enviar email: " + err.Error(),
		})
	}

	return c.JSON(EmailResponse{
		Success: true,
		Message: "Email de prueba enviado exitosamente",
		Data: map[string]any{
			"from":      req.From,
			"to":        req.To,
			"subject":   subject,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}
