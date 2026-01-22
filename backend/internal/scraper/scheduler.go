package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

// ScraperScheduler maneja la ejecución programada del scraper
type ScraperScheduler struct {
	Scheduler *gocron.Scheduler
	Scraper   *GCOScraper
	CronExpr  string
	Enabled   bool
	Job       *gocron.Job
	Location  *time.Location
}

// GlobalScheduler instancia global del scheduler
var GlobalScheduler *ScraperScheduler

// InitScheduler inicializa el scheduler global
func InitScheduler() error {
	if GlobalScraper == nil {
		InitScraper()
	}

	GlobalScheduler = NewScraperScheduler(GlobalScraper)
	return GlobalScheduler.Start()
}

// NewScraperScheduler crea un nuevo scheduler
func NewScraperScheduler(scraper *GCOScraper) *ScraperScheduler {
	// Usar zona horaria de España
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("[Scheduler] Error cargando timezone, usando UTC: %v", err)
		loc = time.UTC
	}

	s := gocron.NewScheduler(loc)

	cronExpr := os.Getenv("GCO_SCHEDULE_CRON")
	if cronExpr == "" {
		cronExpr = "0 6 * * *" // Default: 6:00 AM diario
	}

	enabled := os.Getenv("GCO_SCHEDULE_ENABLED") == "true"

	return &ScraperScheduler{
		Scheduler: s,
		Scraper:   scraper,
		CronExpr:  cronExpr,
		Enabled:   enabled,
		Location:  loc,
	}
}

// Start inicia el scheduler
func (ss *ScraperScheduler) Start() error {
	if !ss.Enabled {
		log.Println("[Scheduler] Scheduler deshabilitado (GCO_SCHEDULE_ENABLED=false)")
		return nil
	}

	log.Printf("[Scheduler] Configurando scheduler con cron: %s", ss.CronExpr)
	log.Printf("[Scheduler] Zona horaria: %s", ss.Location.String())

	job, err := ss.Scheduler.Cron(ss.CronExpr).Do(ss.runTask)
	if err != nil {
		return err
	}

	ss.Job = job
	ss.Scheduler.StartAsync()

	nextRun := job.NextRun()
	log.Printf("[Scheduler] Scheduler iniciado")
	log.Printf("[Scheduler] Próxima ejecución programada: %s", nextRun.Format("2006-01-02 15:04:05 MST"))

	return nil
}

// runTask ejecuta el scraper como tarea programada
func (ss *ScraperScheduler) runTask() {
	log.Println("[Scheduler] ========================================")
	log.Println("[Scheduler] Ejecutando tarea programada")
	log.Println("[Scheduler] ========================================")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// Ejecutar con reintentos
	result, err := ss.Scraper.RunWithRetry(ctx, DefaultRetryConfig)

	if err != nil {
		log.Printf("[Scheduler] Error en ejecución programada: %v", err)
		ss.sendErrorNotification(err)
		return
	}

	log.Printf("[Scheduler] Ejecución completada. Éxito: %v, Archivos: %d, Errores: %d",
		result.Success, len(result.Files), len(result.Errors))

	if result.Success {
		ss.sendSuccessNotification(result)
	} else if len(result.Files) > 0 {
		// Parcialmente exitoso
		ss.sendPartialSuccessNotification(result)
	} else {
		ss.sendErrorNotification(fmt.Errorf("no se descargaron archivos"))
	}

	// Mostrar próxima ejecución
	if ss.Job != nil {
		log.Printf("[Scheduler] Próxima ejecución: %s", ss.Job.NextRun().Format("2006-01-02 15:04:05 MST"))
	}
}

// Stop detiene el scheduler
func (ss *ScraperScheduler) Stop() {
	ss.Scheduler.Stop()
	log.Println("[Scheduler] Scheduler detenido")
}

// UpdateCron actualiza la expresión cron
func (ss *ScraperScheduler) UpdateCron(cronExpr string) error {
	log.Printf("[Scheduler] Actualizando cron de '%s' a '%s'", ss.CronExpr, cronExpr)

	ss.Scheduler.Stop()
	ss.CronExpr = cronExpr

	if ss.Enabled {
		return ss.Start()
	}

	return nil
}

// Enable habilita el scheduler
func (ss *ScraperScheduler) Enable() error {
	ss.Enabled = true
	return ss.Start()
}

// Disable deshabilita el scheduler
func (ss *ScraperScheduler) Disable() {
	ss.Enabled = false
	ss.Stop()
}

// GetNextRun devuelve la próxima ejecución programada
func (ss *ScraperScheduler) GetNextRun() *time.Time {
	if ss.Job == nil || !ss.Enabled {
		return nil
	}
	nextRun := ss.Job.NextRun()
	return &nextRun
}

// GetStatus devuelve el estado del scheduler
func (ss *ScraperScheduler) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":  ss.Enabled,
		"cron":     ss.CronExpr,
		"timezone": ss.Location.String(),
	}

	if nextRun := ss.GetNextRun(); nextRun != nil {
		status["next_run"] = nextRun.Format("2006-01-02 15:04:05 MST")
		status["next_run_in"] = time.Until(*nextRun).String()
	}

	return status
}

// RunNow ejecuta el scraper inmediatamente (fuera de schedule)
func (ss *ScraperScheduler) RunNow() {
	go ss.runTask()
}

// Funciones de notificación (pueden integrarse con email, Slack, etc.)

func (ss *ScraperScheduler) sendSuccessNotification(result *ScraperResult) {
	log.Printf("[Notification] ✅ Scraper GCO completado exitosamente")
	log.Printf("[Notification] Archivos descargados: %d", len(result.Files))
	for _, file := range result.Files {
		log.Printf("[Notification]   - %s: %s (%d bytes)", file.Type, file.Path, file.Size)
	}
	log.Printf("[Notification] Duración: %s", result.Duration)

	// TODO: Integrar con sistema de email (Microsoft Graph) para notificaciones
	// Ejemplo:
	// email.SendEmail("Scraper GCO - Éxito", formatSuccessEmail(result), "admin@sorianomediadores.es")
}

func (ss *ScraperScheduler) sendPartialSuccessNotification(result *ScraperResult) {
	log.Printf("[Notification] ⚠️ Scraper GCO parcialmente exitoso")
	log.Printf("[Notification] Archivos descargados: %d", len(result.Files))
	log.Printf("[Notification] Errores: %d", len(result.Errors))
	for _, errMsg := range result.Errors {
		log.Printf("[Notification]   - %s", errMsg)
	}

	// TODO: Enviar notificación por email
}

func (ss *ScraperScheduler) sendErrorNotification(err error) {
	log.Printf("[Notification] ❌ Error en Scraper GCO: %v", err)

	// TODO: Enviar notificación urgente por email
	// email.SendEmail("⚠️ Scraper GCO - Error", formatErrorEmail(err), "admin@sorianomediadores.es")
}

// formatError formatea un error para logs
func formatError(err error) string {
	return fmt.Sprintf("Error: %v", err)
}
