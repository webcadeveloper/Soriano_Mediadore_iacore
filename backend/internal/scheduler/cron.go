package scheduler

import (
	"log"

	"soriano-mediadores/internal/reports"
	"soriano-mediadores/internal/sharepoint"

	"github.com/robfig/cron/v3"
)

// Scheduler gestiona los trabajos programados
type Scheduler struct {
	cron             *cron.Cron
	reportGenerator  *reports.Generator
	sharepointClient *sharepoint.Client
}

// NewScheduler crea un nuevo scheduler
func NewScheduler(generator *reports.Generator, spClient *sharepoint.Client) *Scheduler {
	// Crear cron con soporte de segundos y logging
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:             c,
		reportGenerator:  generator,
		sharepointClient: spClient,
	}
}

// Start inicia el scheduler con todos los trabajos programados
func (s *Scheduler) Start() error {
	log.Println("📅 Iniciando scheduler de reportes automáticos...")

	// Reporte diario: todos los días a las 8:00 AM
	_, err := s.cron.AddFunc("0 0 8 * * *", func() {
		log.Println("⏰ Ejecutando generación de reporte diario...")
		if err := s.reportGenerator.GenerateDailyReport(); err != nil {
			log.Printf("❌ Error generando reporte diario: %v", err)
		} else {
			log.Println("✅ Reporte diario generado exitosamente")
		}
	})
	if err != nil {
		return err
	}

	// Reporte semanal: todos los lunes a las 9:00 AM
	_, err = s.cron.AddFunc("0 0 9 * * MON", func() {
		log.Println("⏰ Ejecutando generación de reporte semanal...")
		if err := s.reportGenerator.GenerateWeeklyReport(); err != nil {
			log.Printf("❌ Error generando reporte semanal: %v", err)
		} else {
			log.Println("✅ Reporte semanal generado exitosamente")
		}
	})
	if err != nil {
		return err
	}

	// Reporte mensual: primer día del mes a las 10:00 AM
	_, err = s.cron.AddFunc("0 0 10 1 * *", func() {
		log.Println("⏰ Ejecutando generación de reporte mensual...")
		if err := s.reportGenerator.GenerateMonthlyReport(); err != nil {
			log.Printf("❌ Error generando reporte mensual: %v", err)
		} else {
			log.Println("✅ Reporte mensual generado exitosamente")
		}
	})
	if err != nil {
		return err
	}

	// Iniciar el scheduler
	s.cron.Start()

	log.Println("✅ Scheduler iniciado con los siguientes trabajos:")
	log.Println("   • Reporte Diario:   Todos los días a las 08:00")
	log.Println("   • Reporte Semanal:  Todos los lunes a las 09:00")
	log.Println("   • Reporte Mensual:  Día 1 de cada mes a las 10:00")

	return nil
}

// Stop detiene el scheduler
func (s *Scheduler) Stop() {
	log.Println("🛑 Deteniendo scheduler de reportes...")
	s.cron.Stop()
}

// AddCustomJob agrega un trabajo personalizado
func (s *Scheduler) AddCustomJob(schedule string, job func()) error {
	_, err := s.cron.AddFunc(schedule, job)
	return err
}

// GetEntries devuelve todos los trabajos programados
func (s *Scheduler) GetEntries() []cron.Entry {
	return s.cron.Entries()
}
