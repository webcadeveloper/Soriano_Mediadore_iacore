package api

import (
	"log"
	"net/url"
	"soriano-mediadores/internal/bots"
	"soriano-mediadores/internal/db"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	botAtencion   *bots.BotAtencion
	botCobranza   *bots.BotCobranza
	botSiniestros *bots.BotSiniestros
	botAgente     *bots.BotAgente
	botAnalista   *bots.BotAnalista
	botAuditor    *bots.BotAuditor
)

// InitBots inicializa todos los bots
func InitBots() {
	botAtencion = bots.NewBotAtencion()
	botCobranza = bots.NewBotCobranza()
	botSiniestros = bots.NewBotSiniestros()
	botAgente = bots.NewBotAgente()
	botAnalista = bots.NewBotAnalista()
	botAuditor = bots.NewBotAuditor()
}

// HealthCheck verifica el estado del sistema
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  "connected",
		"cache":     "connected",
		"bots":      6,
	})
}

// BuscarClientes endpoint para buscar clientes
func BuscarClientes(c *fiber.Ctx) error {
	query := c.Query("q")

	// Permitir query vacío para obtener todos los clientes
	// Cuando query está vacío, se pasa 0 como límite que significa "todos"
	limit := 0
	if query != "" {
		limit = 100 // Límite razonable para búsquedas específicas
	}

	clientes, err := db.BuscarClientes(query, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ERROR_NUEVO_COMPILADO: " + err.Error(),
		})
	}

	// Guardar métrica
	db.GuardarMetrica("busqueda_clientes", map[string]interface{}{
		"query":      query,
		"resultados": len(clientes),
	})

	return c.JSON(fiber.Map{
		"total":    len(clientes),
		"clientes": clientes,
	})
}

// ObtenerCliente obtiene detalles de un cliente específico
func ObtenerCliente(c *fiber.Ctx) error {
	idAccountRaw := c.Params("id")
	idAccount, _ := url.QueryUnescape(idAccountRaw)

	// Verificar cache
	cacheKey := "cliente:" + idAccount
	type ClienteConPolizas struct {
		db.Cliente
		Polizas []db.Poliza `json:"polizas"`
	}
	var clienteCache ClienteConPolizas

	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &clienteCache); err == nil {
			return c.JSON(clienteCache)
		}
	}

	cliente, err := db.ObtenerClientePorID(idAccount)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Cliente no encontrado",
		})
	}

	// Obtener pólizas del cliente
	polizas, err := db.ObtenerPolizasCliente(idAccount)
	if err != nil {
		polizas = []db.Poliza{} // Si hay error, devolver array vacío
	}

	// Combinar cliente con pólizas
	response := ClienteConPolizas{
		Cliente: *cliente,
		Polizas: polizas,
	}

	// Cache por 15 minutos
	db.CacheSet(cacheKey, response, 15*time.Minute)

	return c.JSON(response)
}

// ObtenerPolizasCliente obtiene pólizas de un cliente
func ObtenerPolizasCliente(c *fiber.Ctx) error {
	idAccountRaw := c.Params("id")
	idAccount, _ := url.QueryUnescape(idAccountRaw)

	// Log para debug
	log.Printf("🔍 Buscando pólizas para id_account: %s (raw: %s)", idAccount, idAccountRaw)

	polizas, err := db.ObtenerPolizasCliente(idAccount)
	if err != nil {
		log.Printf("❌ Error obteniendo pólizas: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("✅ Encontradas %d pólizas para %s", len(polizas), idAccount)

	return c.JSON(fiber.Map{
		"total":   len(polizas),
		"polizas": polizas,
	})
}

// ChatBotAtencion endpoint para chatear con Bot de Atención
func ChatBotAtencion(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botAtencion.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "atencion",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ChatBotCobranza endpoint para chatear con Bot de Cobranza
func ChatBotCobranza(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botCobranza.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "cobranza",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ChatBotSiniestros endpoint para chatear con Bot de Siniestros
func ChatBotSiniestros(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botSiniestros.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "siniestros",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ChatBotAgente endpoint para chatear con Bot Agente/Comercial
func ChatBotAgente(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botAgente.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "agente",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ChatBotAnalista endpoint para chatear con Bot Analista
func ChatBotAnalista(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botAnalista.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "analista",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ChatBotAuditor endpoint para chatear con Bot Auditor
func ChatBotAuditor(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		Mensaje   string `json:"mensaje"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "JSON inválido",
		})
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	respuesta, err := botAuditor.ProcesarConsulta(req.SessionID, req.Mensaje)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": req.SessionID,
		"bot":        "auditor",
		"respuesta":  respuesta,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ListarBots lista todos los bots disponibles
func ListarBots(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"bots": []fiber.Map{
			{
				"id":          "atencion",
				"nombre":      "Atención al Cliente",
				"descripcion": "Busca clientes por nombre/NIF, consulta pólizas, recibos y siniestros de un cliente específico. Pregunta: 'Busca a Juan García' o 'Pólizas del cliente 12345678A'",
				"endpoint":    "/api/chat/atencion",
				"ejemplos":    []string{"Busca a Héctor Pérez", "Pólizas del cliente 20777103/000", "¿Qué productos ofrece Occident?"},
			},
			{
				"id":          "cobranza",
				"nombre":      "Gestor de Cobranza",
				"descripcion": "Lista clientes morosos, recibos pendientes/vencidos, genera mensajes de recobro. Pregunta: 'Clientes que deben' o 'Recibos vencidos'",
				"endpoint":    "/api/chat/cobranza",
				"ejemplos":    []string{"Recibos pendientes", "Clientes morosos", "Genera carta de recobro nivel 2", "Lista de contacto para llamar"},
			},
			{
				"id":          "siniestros",
				"nombre":      "Gestor de Siniestros",
				"descripcion": "Lista siniestros abiertos, estadísticas por tramitador, documentación necesaria para partes. Pregunta: 'Siniestros abiertos' o '¿Qué documentos necesito para un parte de auto?'",
				"endpoint":    "/api/chat/siniestros",
				"ejemplos":    []string{"Siniestros abiertos", "Estadísticas por tramitador", "¿Qué documentos necesito para un siniestro de hogar?"},
			},
			{
				"id":          "agente",
				"nombre":      "Agente Comercial",
				"descripcion": "Información sobre productos Occident, precios orientativos, comparativas. Pregunta sobre seguros de auto, hogar, vida, salud, PIAS, etc.",
				"endpoint":    "/api/chat/agente",
				"ejemplos":    []string{"¿Qué seguros de vida tenéis?", "Información sobre PIAS de Seguros Bilbao", "Diferencia entre todo riesgo con y sin franquicia"},
			},
			{
				"id":          "analista",
				"nombre":      "Analista de Datos",
				"descripcion": "Estadísticas de cartera, top clientes por primas, distribución por ramos, análisis de tendencias.",
				"endpoint":    "/api/chat/analista",
				"ejemplos":    []string{"Top 20 clientes", "Distribución por ramos", "Estadísticas generales"},
			},
			{
				"id":          "auditor",
				"nombre":      "Auditor de Calidad",
				"descripcion": "Detecta clientes duplicados, datos incompletos, problemas de integridad referencial.",
				"endpoint":    "/api/chat/auditor",
				"ejemplos":    []string{"Busca duplicados", "Análisis de calidad de datos", "Datos huérfanos"},
			},
		},
	})
}

// ActualizarClienteRequest estructura para actualizar cliente
type ActualizarClienteRequest struct {
	EmailContacto     *string `json:"email_contacto"`
	TelefonoContacto  *string `json:"telefono_contacto"`
	Telefono2Contacto *string `json:"telefono2_contacto"`
	Domicilio         *string `json:"domicilio"`
	Poblacion         *string `json:"poblacion"`
	CodigoPostal      *string `json:"codigo_postal"`
	Provincia         *string `json:"provincia"`
}

// ActualizarCliente endpoint para actualizar datos de un cliente (CRM)
func ActualizarCliente(c *fiber.Ctx) error {
	clienteIDRaw := c.Params("id")
	clienteID, _ := url.QueryUnescape(clienteIDRaw)

	var req ActualizarClienteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "JSON inválido",
			"error":   err.Error(),
		})
	}

	// Construir query dinámicamente solo con campos que se envían
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.EmailContacto != nil {
		updates = append(updates, "email_contacto = $"+string(rune(argCount+48)))
		args = append(args, *req.EmailContacto)
		argCount++
	}
	if req.TelefonoContacto != nil {
		updates = append(updates, "telefono_contacto = $"+string(rune(argCount+48)))
		args = append(args, *req.TelefonoContacto)
		argCount++
	}
	if req.Telefono2Contacto != nil {
		updates = append(updates, "telefono2_contacto = $"+string(rune(argCount+48)))
		args = append(args, *req.Telefono2Contacto)
		argCount++
	}
	if req.Domicilio != nil {
		updates = append(updates, "domicilio = $"+string(rune(argCount+48)))
		args = append(args, *req.Domicilio)
		argCount++
	}
	if req.Poblacion != nil {
		updates = append(updates, "poblacion = $"+string(rune(argCount+48)))
		args = append(args, *req.Poblacion)
		argCount++
	}
	if req.CodigoPostal != nil {
		updates = append(updates, "codigo_postal = $"+string(rune(argCount+48)))
		args = append(args, *req.CodigoPostal)
		argCount++
	}
	if req.Provincia != nil {
		updates = append(updates, "provincia = $"+string(rune(argCount+48)))
		args = append(args, *req.Provincia)
		argCount++
	}

	if len(updates) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "No hay campos para actualizar",
		})
	}

	// Agregar actualizado_en
	updates = append(updates, "actualizado_en = NOW()")

	// Construir query
	query := "UPDATE clientes SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE (id::text = $" + string(rune(argCount+48)) + " OR id_account = $" + string(rune(argCount+48)) + ")"
	args = append(args, clienteID)

	// Ejecutar actualización
	result, err := db.PostgresDB.Exec(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error al actualizar cliente",
			"error":   err.Error(),
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Cliente no encontrado",
		})
	}

	// Invalidar cache del cliente
	db.CacheDelete("cliente:" + clienteID)

	// Obtener cliente actualizado
	cliente, err := db.ObtenerClientePorID(clienteID)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Cliente actualizado correctamente",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Cliente actualizado correctamente",
		"cliente": cliente,
	})
}

// GetRamos obtiene todos los tipos de póliza disponibles
func GetRamos(c *fiber.Ctx) error {
	rows, err := db.PostgresDB.Query(`
		SELECT DISTINCT ramo
		FROM polizas
		WHERE ramo IS NOT NULL AND TRIM(ramo) != ''
		ORDER BY ramo
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var ramos []string
	for rows.Next() {
		var ramo string
		if err := rows.Scan(&ramo); err == nil {
			ramos = append(ramos, ramo)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"total":   len(ramos),
		"ramos":   ramos,
	})
}

// CrearClienteRequest estructura para crear un nuevo cliente
type CrearClienteRequest struct {
	NIF               string  `json:"nif"`
	NombreCompleto    string  `json:"nombre_completo"`
	EmailContacto     *string `json:"email_contacto"`
	TelefonoContacto  *string `json:"telefono_contacto"`
	Telefono2Contacto *string `json:"telefono2_contacto"`
	Domicilio         *string `json:"domicilio"`
	Poblacion         *string `json:"poblacion"`
	CodigoPostal      *string `json:"codigo_postal"`
	Provincia         *string `json:"provincia"`
}

// CrearCliente crea un nuevo cliente en la base de datos
func CrearCliente(c *fiber.Ctx) error {
	var req CrearClienteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "JSON inválido",
			"error":   err.Error(),
		})
	}

	// Validar campos obligatorios
	if req.NIF == "" || req.NombreCompleto == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "NIF y Nombre Completo son obligatorios",
		})
	}

	// Verificar si ya existe un cliente con ese NIF
	var existingID int
	err := db.PostgresDB.QueryRow("SELECT id FROM clientes WHERE nif = $1", req.NIF).Scan(&existingID)
	if err == nil {
		return c.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "Ya existe un cliente con ese NIF",
			"cliente_id": existingID,
		})
	}

	// Generar id_account único
	idAccount := "CLI-" + uuid.New().String()[:8]

	// Insertar el nuevo cliente
	var newID int
	err = db.PostgresDB.QueryRow(`
		INSERT INTO clientes (
			id_account, nif, nombre_completo, email_contacto,
			telefono_contacto, telefono2_contacto, domicilio,
			poblacion, codigo_postal, provincia, activo, creado_en, actualizado_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, NOW(), NOW())
		RETURNING id
	`, idAccount, req.NIF, req.NombreCompleto, req.EmailContacto,
	   req.TelefonoContacto, req.Telefono2Contacto, req.Domicilio,
	   req.Poblacion, req.CodigoPostal, req.Provincia).Scan(&newID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error al crear cliente",
			"error":   err.Error(),
		})
	}

	// Obtener el cliente creado
	cliente, err := db.ObtenerClientePorID(idAccount)
	if err != nil {
		return c.Status(201).JSON(fiber.Map{
			"success":    true,
			"message":    "Cliente creado correctamente",
			"cliente_id": newID,
			"id_account": idAccount,
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success":    true,
		"message":    "Cliente creado correctamente",
		"cliente":    cliente,
		"id_account": idAccount,
	})
}

// Estadisticas generales del sistema
func Estadisticas(c *fiber.Ctx) error {
	var stats struct {
		TotalClientes   int
		TotalPolizas    int
		TotalRecibos    int
		TotalSiniestros int
	}

	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE activo = TRUE").Scan(&stats.TotalClientes)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM polizas WHERE activo = TRUE").Scan(&stats.TotalPolizas)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM recibos WHERE activo = TRUE").Scan(&stats.TotalRecibos)
	db.PostgresDB.QueryRow("SELECT COUNT(*) FROM siniestros WHERE activo = TRUE").Scan(&stats.TotalSiniestros)

	return c.JSON(fiber.Map{
		"estadisticas": stats,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}
