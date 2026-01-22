package main

import (
	"log"
	"os"
	"strings"
	"soriano-mediadores/internal/api"
	"soriano-mediadores/internal/auth"
	"soriano-mediadores/internal/db"
	"soriano-mediadores/internal/scraper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No se encontró archivo .env, usando variables de entorno del sistema")
	}

	log.Println("🚀 Iniciando Soriano Mediadores - Sistema de Gestión")
	log.Println(strings.Repeat("=", 60))

	// Inicializar bases de datos
	log.Println("\n📊 Conectando a bases de datos...")

	if err := db.InitPostgres(); err != nil {
		log.Fatalf("❌ Error iniciando PostgreSQL: %v", err)
	}

	if err := db.InitRedis(); err != nil {
		log.Printf("⚠️  Error iniciando Redis: %v (continuando sin cache)", err)
	}

	if err := db.InitMongo(); err != nil {
		log.Printf("⚠️  Error iniciando MongoDB: %v (continuando sin MongoDB)", err)
	}

	// Inicializar bots
	log.Println("\n🤖 Inicializando bots AI...")
	api.InitBots()
	log.Println("✅ 6 bots inicializados correctamente")

	// Inicializar scraper GCO
	log.Println("\n🔄 Inicializando scraper GCO...")
	scraper.InitScraper()
	if err := scraper.InitScheduler(); err != nil {
		log.Printf("⚠️  Error iniciando scheduler: %v", err)
	} else {
		log.Println("✅ Scraper GCO inicializado")
	}

	// Inicializar autenticación Microsoft
	log.Println("\n🔐 Inicializando autenticación Microsoft...")
	auth.InitAuth()
	log.Println("✅ Autenticación Microsoft inicializada")

	// Crear servidor Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Soriano Mediadores API",
		ServerHeader: "Soriano",
		Prefork:      false,
		BodyLimit:    100 * 1024 * 1024, // 100 MB límite para archivos CSV grandes
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Rutas de autenticación (antes del middleware de auth)
	setupAuthRoutes(app)

	// Middleware de autenticación (protege todas las rutas siguientes)
	app.Use(auth.AuthMiddleware)

	// Rutas protegidas
	setupRoutes(app)

	// Puerto
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("\n✅ Servidor listo en http://localhost:%s\n", port)
	log.Println(strings.Repeat("=", 60))
	log.Println("\n📡 Endpoints disponibles:")
	log.Println("   GET  /health              - Estado del sistema")
	log.Println("   GET  /api/stats           - Estadísticas generales")
	log.Println("   GET  /api/bots            - Lista de bots disponibles")
	log.Println("   GET  /api/clientes?q=     - Buscar clientes")
	log.Println("   GET  /api/clientes/:id    - Obtener cliente")
	log.Println("   GET  /api/clientes/:id/polizas - Pólizas del cliente")
	log.Println("   POST /api/chat/atencion   - Chat con Bot Atención")
	log.Println("   POST /api/chat/cobranza   - Chat con Bot Cobranza")
	log.Println("   POST /api/chat/siniestros - Chat con Bot Siniestros")
	log.Println("   POST /api/chat/agente     - Chat con Bot Agente")
	log.Println("   POST /api/chat/analista   - Chat con Bot Analista")
	log.Println("   POST /api/chat/auditor    - Chat con Bot Auditor")
	log.Println("\n📥 Importación CSV:")
	log.Println("   POST /api/admin/import/preview  - Previsualizar CSV")
	log.Println("   POST /api/admin/import/start    - Iniciar importación")
	log.Println("   GET  /api/admin/import/status/:id - Estado de importación")
	log.Println("   GET  /api/admin/import/history  - Historial de importaciones")
	log.Println("\n🔗 N8N Webhooks:")
	log.Println("   POST /api/n8n/cliente/creado    - Crear cliente desde N8N")
	log.Println("   POST /api/n8n/poliza/creada     - Crear póliza desde N8N")
	log.Println("   POST /api/n8n/cliente/consulta  - Consultar cliente desde N8N")
	log.Println("   POST /api/n8n/webhook           - Webhook genérico N8N")
	log.Println("\n📧 Recobros - Microsoft Graph:")
	log.Println("   POST /api/recobros/send-email   - Enviar email de recobro")
	log.Println("   POST /api/recobros/send-email-template - Enviar email con plantilla (1, 2 o 3)")
	log.Println("   POST /api/recobros/send-bulk    - Envío masivo de emails")
	log.Println("   POST /api/recobros/test-email   - Enviar email de prueba")
	log.Println("   GET  /api/recobros/test-graph   - Probar conexión con Microsoft Graph")
	log.Println("   GET  /api/recobros/devueltos    - Lista de recibos devueltos")
	log.Println("   GET  /api/recobros/clientes-deuda - Clientes con deudas")
	log.Println("\n📊 Estadísticas:")
	log.Println("   GET  /api/stats/general         - Estadísticas generales del sistema")
	log.Println("   GET  /api/stats/recibos-devueltos - Recibos devueltos (con paginación)")
	log.Println("   GET  /api/stats/clientes-deuda  - Clientes con deuda (con paginación)")
	log.Println("   GET  /api/stats/recibos-kpi     - KPIs completos + historial de recibos (con filtros)")
	log.Println("\n📈 Business Intelligence:")
	log.Println("   GET  /api/analytics/financial-kpis        - KPIs financieros (primas, comisiones, pocket share)")
	log.Println("   GET  /api/analytics/portfolio-analysis    - Análisis de cartera (por ramo, provincia, top clientes)")
	log.Println("   GET  /api/analytics/collections-performance - Rendimiento cobros (morosidad, ratios)")
	log.Println("   GET  /api/analytics/claims-analysis       - Análisis siniestros (siniestralidad por ramo)")
	log.Println("   GET  /api/analytics/performance-trends    - Tendencias de rendimiento (clientes, pólizas, primas)")
	log.Println("\n🔄 Scraper GCO:")
	log.Println("   POST /api/scraper/run           - Ejecutar scraper manualmente")
	log.Println("   GET  /api/scraper/status        - Estado del scraper")
	log.Println("   GET  /api/scraper/metrics       - Métricas de ejecución")
	log.Println("   POST /api/scraper/stop          - Detener scraper")
	log.Println("   GET  /api/scraper/schedule      - Estado del scheduler")
	log.Println("   POST /api/scraper/schedule      - Configurar scheduler")
	log.Println("   POST /api/scraper/test-login    - Probar login en GCO")
	log.Println()

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

func setupRoutes(app *fiber.App) {
	// Health check
	app.Get("/health", api.HealthCheck)

	// Ruta para el frontend de recobros
	app.Get("/recobros", func(c *fiber.Ctx) error {
		return c.SendFile("./templates/recobros.html")
	})

	// API v1
	v1 := app.Group("/api")

	// Estadísticas
	v1.Get("/stats", api.Estadisticas)

	// Bots
	v1.Get("/bots", api.ListarBots)

	// Clientes - CRM
	v1.Get("/clientes", api.BuscarClientes)
	v1.Post("/clientes", api.CrearCliente)           // CRM - Crear nuevo cliente
	v1.Get("/clientes/:id", api.ObtenerCliente)
	v1.Put("/clientes/:id", api.ActualizarCliente)   // CRM - Actualizar cliente
	v1.Get("/clientes/:id/polizas", api.ObtenerPolizasCliente)

	// Catálogos
	v1.Get("/ramos", api.GetRamos)                   // Obtener tipos de póliza

	// Chat con bots
	chat := v1.Group("/chat")
	chat.Post("/atencion", api.ChatBotAtencion)
	chat.Post("/cobranza", api.ChatBotCobranza)
	chat.Post("/siniestros", api.ChatBotSiniestros)
	chat.Post("/agente", api.ChatBotAgente)
	chat.Post("/analista", api.ChatBotAnalista)
	chat.Post("/auditor", api.ChatBotAuditor)

	// Admin - CSV Import
	admin := v1.Group("/admin")
	importRoutes := admin.Group("/import")
	importRoutes.Post("/preview", api.PreviewCSV)
	importRoutes.Post("/start", api.StartImport)
	importRoutes.Get("/status/:id", api.GetImportStatus)
	importRoutes.Post("/cancel/:id", api.CancelImport)
	importRoutes.Get("/history", api.GetImportHistory)
	importRoutes.Post("/revert/:id", api.RevertImport)
	importRoutes.Post("/validate", api.ValidateImport)
	importRoutes.Get("/template", api.GetImportTemplate)

	// N8N Webhooks
	n8n := v1.Group("/n8n")
	n8n.Post("/cliente/creado", api.N8NClienteCreado)
	n8n.Post("/poliza/creada", api.N8NPolizaCreada)
	n8n.Post("/recibo/creado", api.N8NReciboCreado)
	n8n.Post("/siniestro/creado", api.N8NSiniestroCreado)
	n8n.Post("/cliente/consulta", api.N8NConsultaCliente)
	n8n.Post("/polizas/consulta", api.N8NConsultaPolizas)
	n8n.Post("/recibos/consulta", api.N8NConsultaRecibos)
	n8n.Post("/estadisticas", api.N8NEstadisticas)
	n8n.Post("/cliente/actualizar", api.N8NActualizarCliente)
	n8n.Post("/cliente/notificar", api.N8NNotificarCliente)
	n8n.Post("/webhook", api.N8NWebhookGenerico)

	// Recobros - Microsoft Graph Email
	recobros := v1.Group("/recobros")
	recobros.Post("/send-email", api.SendReciboEmail)
	recobros.Post("/send-email-template", api.SendReciboEmailWithTemplate) // NUEVO - Enviar con plantilla
	recobros.Post("/send-bulk", api.SendBulkReciboEmails)
	recobros.Post("/test-email", api.SendTestEmail)
	recobros.Get("/test-graph", api.TestGraphConnection)
	recobros.Get("/templates", api.GetEmailTemplates)
	recobros.Get("/devueltos", api.GetRecibosDevueltos)
	recobros.Get("/clientes-deuda", api.GetClientesConDeuda)

	// Estadísticas y Analytics
	v1.Get("/stats/general", api.GetStats)
	v1.Get("/stats/recibos-devueltos", api.GetRecibosDevueltos)
	v1.Get("/stats/clientes-deuda", api.GetClientesConDeuda)
	v1.Get("/stats/recibos-kpi", api.GetRecibosKPI) // KPIs completos + historial

	// Analytics - Business Intelligence
	analytics := v1.Group("/analytics")
	analytics.Get("/financial-kpis", api.GetFinancialKPIs)
	analytics.Get("/portfolio-analysis", api.GetPortfolioAnalysis)
	analytics.Get("/collections-performance", api.GetCollectionsPerformance)
	analytics.Get("/claims-analysis", api.GetClaimsAnalysis)
	analytics.Get("/performance-trends", api.GetPerformanceTrends)

	// Scraper GCO
	scraper.RegisterScraperRoutes(app)

	// Servir el frontend Angular (archivos estáticos)
	app.Static("/", "../frontend")

	// Fallback para SPA - cualquier ruta no encontrada sirve index.html
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/index.html")
	})
}

// setupAuthRoutes configura las rutas de autenticación
func setupAuthRoutes(app *fiber.App) {
	// Página de login
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./templates/login.html")
	})

	// Rutas de autenticación OAuth
	authGroup := app.Group("/auth")
	authGroup.Get("/login", auth.LoginHandler)
	authGroup.Get("/callback", auth.CallbackHandler)
	authGroup.Get("/logout", auth.LogoutHandler)
	authGroup.Get("/me", auth.MeHandler)
}
