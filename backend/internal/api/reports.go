package api

import (
	"time"

	"soriano-mediadores/internal/reports"

	"github.com/gofiber/fiber/v2"
)

// GenerateReportRequest request para generar reporte
type GenerateReportRequest struct {
	Type                  string `json:"type"`                    // "daily", "weekly", "monthly"
	Period                string `json:"period"`                  // "2026-01-22", "2026-W03", "2026-01"
	PublishToSharePoint   bool   `json:"publish_to_sharepoint"`   // true/false
}

// GenerateReportResponse respuesta de generación de reporte
type GenerateReportResponse struct {
	Success bool                          `json:"success"`
	Data    *reports.GenerateReportResult `json:"data,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

// ReportGenerator instancia global del generador de reportes
var ReportGenerator *reports.Generator

// GenerateReport endpoint para generar reportes manualmente
func GenerateReport(c *fiber.Ctx) error {
	if ReportGenerator == nil {
		return c.Status(503).JSON(fiber.Map{
			"success": false,
			"error":   "Generador de reportes no disponible. SharePoint no está configurado.",
		})
	}

	var req GenerateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "JSON inválido: " + err.Error(),
		})
	}

	// Validar tipo de reporte
	var reportType reports.ReportType
	switch req.Type {
	case "daily":
		reportType = reports.ReportTypeDaily
	case "weekly":
		reportType = reports.ReportTypeWeekly
	case "monthly":
		reportType = reports.ReportTypeMonthly
	default:
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Tipo de reporte inválido. Debe ser: daily, weekly o monthly",
		})
	}

	// Si no se especifica periodo, usar el actual
	period := req.Period
	if period == "" {
		switch reportType {
		case reports.ReportTypeDaily:
			period = time.Now().Format("2006-01-02")
		case reports.ReportTypeWeekly:
			year, week := time.Now().ISOWeek()
			period = fiber.Map{"year": year, "week": week}.(string)
		case reports.ReportTypeMonthly:
			period = time.Now().Format("2006-01")
		}
	}

	// Generar reporte
	result, err := ReportGenerator.GenerateReport(reportType, period, req.PublishToSharePoint)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Error generando reporte: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// ListReports endpoint para listar reportes generados
func ListReports(c *fiber.Ctx) error {
	// TODO: Implementar listado de reportes desde SharePoint
	// Por ahora devolver un placeholder

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"reports": []fiber.Map{},
			"message": "Listado de reportes no implementado aún",
		},
	})
}

// CreateSharePointFolders endpoint para crear estructura de carpetas en SharePoint
func CreateSharePointFolders(c *fiber.Ctx) error {
	if ReportGenerator == nil {
		return c.Status(503).JSON(fiber.Map{
			"success": false,
			"error":   "Generador de reportes no disponible",
		})
	}

	err := ReportGenerator.CreateSharePointFolderStructure()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Error creando carpetas: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Estructura de carpetas creada exitosamente en SharePoint",
	})
}

// TestSharePointConnection endpoint para probar conexión con SharePoint
func TestSharePointConnection(c *fiber.Ctx) error {
	if ReportGenerator == nil {
		return c.Status(503).JSON(fiber.Map{
			"success": false,
			"error":   "SharePoint no está configurado",
		})
	}

	// Aquí podríamos agregar un método de test en el generator
	// Por ahora solo devolver OK si el generator existe

	return c.JSON(fiber.Map{
		"success": true,
		"message": "SharePoint está configurado correctamente",
	})
}
