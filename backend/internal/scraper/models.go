package scraper

import (
	"sync"
	"time"
)

// CSVType representa los tipos de CSV que se pueden descargar
type CSVType string

const (
	CSVClientes   CSVType = "clientes"
	CSVPolizas    CSVType = "polizas"
	CSVRecibos    CSVType = "recibos"
	CSVSiniestros CSVType = "siniestros"
)

// ScraperStatus representa el estado del scraper
type ScraperStatus string

const (
	StatusIdle      ScraperStatus = "idle"
	StatusRunning   ScraperStatus = "running"
	StatusSuccess   ScraperStatus = "success"
	StatusFailed    ScraperStatus = "failed"
	StatusCancelled ScraperStatus = "cancelled"
)

// DownloadConfig configuración para descarga de cada tipo de CSV
type DownloadConfig struct {
	Type         CSVType
	MenuPath     []string // Ruta de navegación por menús
	ExportButton string   // Selector del botón de export
	FileName     string   // Nombre del archivo descargado
	WaitSelector string   // Selector para esperar antes de exportar
}

// DownloadedFile representa un archivo descargado
type DownloadedFile struct {
	Type         CSVType   `json:"type"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	DownloadedAt time.Time `json:"downloaded_at"`
	ImportJobID  string    `json:"import_job_id,omitempty"`
}

// ScraperResult resultado de una ejecución del scraper
type ScraperResult struct {
	Success      bool             `json:"success"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Duration     string           `json:"duration"`
	Files        []DownloadedFile `json:"files"`
	Errors       []string         `json:"errors"`
	ImportJobIDs []string         `json:"import_job_ids,omitempty"`
}

// ScraperMetrics métricas de ejecución del scraper
type ScraperMetrics struct {
	TotalRuns            int64         `json:"total_runs"`
	SuccessfulRuns       int64         `json:"successful_runs"`
	FailedRuns           int64         `json:"failed_runs"`
	TotalFilesDownloaded int64         `json:"total_files_downloaded"`
	TotalBytesDownloaded int64         `json:"total_bytes_downloaded"`
	LastRunTime          time.Time     `json:"last_run_time"`
	LastRunDuration      time.Duration `json:"last_run_duration"`
	LastRunStatus        ScraperStatus `json:"last_run_status"`
	AverageRunDuration   time.Duration `json:"average_run_duration"`
	mu                   sync.RWMutex
}

// RecordRun registra el resultado de una ejecución
func (m *ScraperMetrics) RecordRun(result *ScraperResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRuns++
	m.LastRunTime = result.StartTime
	m.LastRunDuration = result.EndTime.Sub(result.StartTime)

	if result.Success {
		m.SuccessfulRuns++
		m.LastRunStatus = StatusSuccess
	} else {
		m.FailedRuns++
		m.LastRunStatus = StatusFailed
	}

	for _, file := range result.Files {
		m.TotalFilesDownloaded++
		m.TotalBytesDownloaded += file.Size
	}

	// Recalcular promedio
	if m.TotalRuns > 0 {
		totalDuration := time.Duration(int64(m.AverageRunDuration)*int64(m.TotalRuns-1)) + m.LastRunDuration
		m.AverageRunDuration = totalDuration / time.Duration(m.TotalRuns)
	}
}

// ToJSON convierte métricas a mapa para JSON
func (m *ScraperMetrics) ToJSON() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	successRate := float64(0)
	if m.TotalRuns > 0 {
		successRate = float64(m.SuccessfulRuns) / float64(m.TotalRuns) * 100
	}

	return map[string]interface{}{
		"total_runs":             m.TotalRuns,
		"successful_runs":        m.SuccessfulRuns,
		"failed_runs":            m.FailedRuns,
		"success_rate":           successRate,
		"total_files_downloaded": m.TotalFilesDownloaded,
		"total_bytes_downloaded": m.TotalBytesDownloaded,
		"last_run_time":          m.LastRunTime,
		"last_run_duration":      m.LastRunDuration.String(),
		"last_run_status":        m.LastRunStatus,
		"average_run_duration":   m.AverageRunDuration.String(),
	}
}

// RetryConfig configuración de reintentos
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryConfig configuración por defecto de reintentos
var DefaultRetryConfig = RetryConfig{
	MaxRetries:    3,
	InitialDelay:  5 * time.Second,
	MaxDelay:      60 * time.Second,
	BackoffFactor: 2.0,
}

// DefaultDownloadConfigs configuraciones por defecto para cada tipo de CSV
// NOTA: Los selectores se ajustarán según el HTML real del portal
var DefaultDownloadConfigs = map[CSVType]DownloadConfig{
	CSVClientes: {
		Type:         CSVClientes,
		MenuPath:     []string{"Cartera", "Clientes"},
		ExportButton: `button[title*="Exportar"], .btn-export, a[href*="export"], button:contains("Excel"), button:contains("CSV")`,
		FileName:     "CLIENTES.csv",
		WaitSelector: `table, .grid, .data-table, .ag-body`,
	},
	CSVPolizas: {
		Type:         CSVPolizas,
		MenuPath:     []string{"Cartera", "Pólizas"},
		ExportButton: `button[title*="Exportar"], .btn-export, a[href*="export"], button:contains("Excel"), button:contains("CSV")`,
		FileName:     "POLIZAS.csv",
		WaitSelector: `table, .grid, .data-table, .ag-body`,
	},
	CSVRecibos: {
		Type:         CSVRecibos,
		MenuPath:     []string{"Recibos", "Listado"},
		ExportButton: `button[title*="Exportar"], .btn-export, a[href*="export"], button:contains("Excel"), button:contains("CSV")`,
		FileName:     "RECIBOS.csv",
		WaitSelector: `table, .grid, .data-table, .ag-body`,
	},
	CSVSiniestros: {
		Type:         CSVSiniestros,
		MenuPath:     []string{"Siniestros", "Listado"},
		ExportButton: `button[title*="Exportar"], .btn-export, a[href*="export"], button:contains("Excel"), button:contains("CSV")`,
		FileName:     "SINIESTROS.csv",
		WaitSelector: `table, .grid, .data-table, .ag-body`,
	},
}
