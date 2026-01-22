package scraper

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RunScraperHandler ejecuta el scraper manualmente
// POST /api/scraper/run
func RunScraperHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		InitScraper()
	}

	if GlobalScraper.IsRunning {
		return c.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "El scraper ya está en ejecución",
			"status":  GlobalScraper.Status,
		})
	}

	// Ejecutar en background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()

		log.Println("[Handler] Iniciando scraper manualmente...")
		result, err := GlobalScraper.RunWithRetry(ctx, DefaultRetryConfig)

		if err != nil {
			log.Printf("[Handler] Error en scraper: %v", err)
		} else {
			log.Printf("[Handler] Scraper completado. Archivos: %d", len(result.Files))
		}
	}()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Scraper iniciado en background",
		"status":  StatusRunning,
	})
}

// GetScraperStatusHandler devuelve el estado actual del scraper
// GET /api/scraper/status
func GetScraperStatusHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		InitScraper()
	}

	status := GlobalScraper.GetStatus()

	// Añadir info del scheduler si está disponible
	if GlobalScheduler != nil {
		status["scheduler"] = GlobalScheduler.GetStatus()
	}

	return c.JSON(status)
}

// GetScraperMetricsHandler devuelve las métricas del scraper
// GET /api/scraper/metrics
func GetScraperMetricsHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		InitScraper()
	}

	return c.JSON(GlobalScraper.Metrics.ToJSON())
}

// StopScraperHandler detiene el scraper si está en ejecución
// POST /api/scraper/stop
func StopScraperHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Scraper no inicializado",
		})
	}

	err := GlobalScraper.Stop()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Solicitud de detención enviada",
	})
}

// ConfigureScheduleHandler configura el schedule del scraper
// POST /api/scraper/schedule
func ConfigureScheduleHandler(c *fiber.Ctx) error {
	var req struct {
		Cron    string `json:"cron"`
		Enabled *bool  `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "JSON inválido",
		})
	}

	if GlobalScheduler == nil {
		if GlobalScraper == nil {
			InitScraper()
		}
		GlobalScheduler = NewScraperScheduler(GlobalScraper)
	}

	// Actualizar cron si se proporciona
	if req.Cron != "" {
		err := GlobalScheduler.UpdateCron(req.Cron)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   "Expresión cron inválida: " + err.Error(),
			})
		}
	}

	// Habilitar/deshabilitar si se proporciona
	if req.Enabled != nil {
		if *req.Enabled {
			GlobalScheduler.Enable()
		} else {
			GlobalScheduler.Disable()
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Schedule actualizado",
		"schedule": GlobalScheduler.GetStatus(),
	})
}

// GetScheduleStatusHandler devuelve el estado del scheduler
// GET /api/scraper/schedule
func GetScheduleStatusHandler(c *fiber.Ctx) error {
	if GlobalScheduler == nil {
		return c.JSON(fiber.Map{
			"enabled": false,
			"message": "Scheduler no inicializado",
		})
	}

	return c.JSON(GlobalScheduler.GetStatus())
}

// TestLoginHandler prueba solo el login sin descargar
// POST /api/scraper/test-login
func TestLoginHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		InitScraper()
	}

	if GlobalScraper.IsRunning {
		return c.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "El scraper está en ejecución",
		})
	}

	log.Println("[Handler] Probando login...")

	// Crear contexto Chrome
	ctx, cancel := createChromeContext(GlobalScraper.DownloadDir)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 2*time.Minute)
	defer cancelTimeout()

	// Intentar login
	err := GlobalScraper.Auth.Login(ctx)
	if err != nil {
		log.Printf("[Handler] Error en test de login: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error en login",
			"error":   err.Error(),
		})
	}

	// Logout
	GlobalScraper.Auth.Logout(ctx)

	log.Println("[Handler] Test de login exitoso")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login exitoso",
	})
}

// GetConfigHandler devuelve la configuración actual del scraper
// GET /api/scraper/config
func GetConfigHandler(c *fiber.Ctx) error {
	if GlobalScraper == nil {
		InitScraper()
	}

	// No devolver contraseña por seguridad
	username := GlobalScraper.Auth.Username
	if len(username) > 4 {
		username = username[:4] + "****"
	}

	return c.JSON(fiber.Map{
		"base_url":     GlobalScraper.BaseURL,
		"username":     username,
		"download_dir": GlobalScraper.DownloadDir,
		"csv_types": []string{
			string(CSVClientes),
			string(CSVPolizas),
			string(CSVRecibos),
			string(CSVSiniestros),
		},
	})
}

// RegisterScraperRoutes registra todas las rutas del scraper
func RegisterScraperRoutes(app *fiber.App) {
	scraper := app.Group("/api/scraper")

	scraper.Post("/run", RunScraperHandler)
	scraper.Get("/status", GetScraperStatusHandler)
	scraper.Get("/metrics", GetScraperMetricsHandler)
	scraper.Post("/stop", StopScraperHandler)
	scraper.Get("/schedule", GetScheduleStatusHandler)
	scraper.Post("/schedule", ConfigureScheduleHandler)
	scraper.Post("/test-login", TestLoginHandler)
	scraper.Get("/config", GetConfigHandler)

	log.Println("[Scraper] Rutas del scraper registradas en /api/scraper/*")
}
