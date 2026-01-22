package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// GCOScraper es el scraper principal para el portal GCO
type GCOScraper struct {
	Auth        *AuthManager
	Downloader  *Downloader
	Metrics     *ScraperMetrics
	BaseURL     string
	DownloadDir string
	mu          sync.Mutex
	IsRunning   bool
	LastRun     time.Time
	LastError   error
	Status      ScraperStatus
	cancelFunc  context.CancelFunc
}

// GlobalScraper instancia global del scraper
var GlobalScraper *GCOScraper

// InitScraper inicializa el scraper global
func InitScraper() {
	GlobalScraper = NewGCOScraper()
	log.Println("[Scraper] Scraper GCO inicializado")
}

// NewGCOScraper crea un nuevo scraper
func NewGCOScraper() *GCOScraper {
	downloadDir := os.Getenv("GCO_DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = "/opt/soriano/backend/CSV/scraped"
	}

	// Crear directorio si no existe
	os.MkdirAll(downloadDir, 0755)

	logDir := os.Getenv("GCO_LOG_DIR")
	if logDir == "" {
		logDir = "/opt/soriano/logs/scraper"
	}
	os.MkdirAll(logDir, 0755)

	return &GCOScraper{
		Auth:        NewAuthManager(),
		Downloader:  NewDownloader(downloadDir),
		Metrics:     &ScraperMetrics{},
		BaseURL:     os.Getenv("GCO_BASE_URL"),
		DownloadDir: downloadDir,
		Status:      StatusIdle,
	}
}

// createChromeContext crea un contexto de Chrome con las opciones necesarias
func createChromeContext(downloadDir string) (context.Context, context.CancelFunc) {
	headless := os.Getenv("GCO_HEADLESS") != "false"
	ignoreSSL := os.Getenv("GCO_IGNORE_SSL_ERRORS") == "true"

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),

		// Directorio de descargas
		chromedp.Flag("download.default_directory", downloadDir),
		chromedp.Flag("download.prompt_for_download", false),
		chromedp.Flag("download.directory_upgrade", true),
		chromedp.Flag("safebrowsing.enabled", false),
	)

	if ignoreSSL {
		opts = append(opts,
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("ignore-ssl-errors", true),
		)
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(log.Printf),
	)

	// Combinar cancels
	cancel := func() {
		ctxCancel()
		allocCancel()
	}

	// Configurar el comportamiento de descarga
	chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Esto se ejecutará cuando se abra el navegador
			return nil
		}),
	)

	return ctx, cancel
}

// Run ejecuta el scraper completo
func (s *GCOScraper) Run(parentCtx context.Context) (*ScraperResult, error) {
	s.mu.Lock()
	if s.IsRunning {
		s.mu.Unlock()
		return nil, fmt.Errorf("el scraper ya está en ejecución")
	}
	s.IsRunning = true
	s.Status = StatusRunning
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.IsRunning = false
		s.LastRun = time.Now()
		s.mu.Unlock()
	}()

	result := &ScraperResult{
		StartTime: time.Now(),
		Files:     []DownloadedFile{},
		Errors:    []string{},
	}

	log.Println("[Scraper] ========================================")
	log.Println("[Scraper] Iniciando ejecución del scraper GCO")
	log.Println("[Scraper] ========================================")

	// Crear contexto Chrome
	chromeCtx, cancel := createChromeContext(s.DownloadDir)
	s.cancelFunc = cancel
	defer cancel()

	// Timeout global de 15 minutos
	chromeCtx, cancelTimeout := context.WithTimeout(chromeCtx, 15*time.Minute)
	defer cancelTimeout()

	// 1. Login
	log.Println("[Scraper] Paso 1/3: Autenticación...")
	err := s.Auth.Login(chromeCtx)
	if err != nil {
		errMsg := fmt.Sprintf("Error en login: %v", err)
		log.Println("[Scraper] " + errMsg)
		result.Errors = append(result.Errors, errMsg)
		s.Status = StatusFailed
		s.LastError = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		return result, err
	}
	log.Println("[Scraper] Login exitoso")

	// 2. Descargar CSVs
	log.Println("[Scraper] Paso 2/3: Descargando CSVs...")
	csvTypes := []CSVType{CSVClientes, CSVPolizas, CSVRecibos, CSVSiniestros}

	for i, csvType := range csvTypes {
		log.Printf("[Scraper] Descargando %s (%d/%d)...", csvType, i+1, len(csvTypes))

		file, err := s.Downloader.DownloadCSV(chromeCtx, csvType)
		if err != nil {
			errMsg := fmt.Sprintf("Error descargando %s: %v", csvType, err)
			log.Println("[Scraper] " + errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		result.Files = append(result.Files, *file)
		log.Printf("[Scraper] %s descargado exitosamente", csvType)
	}

	// 3. Logout
	log.Println("[Scraper] Paso 3/3: Cerrando sesión...")
	s.Auth.Logout(chromeCtx)

	// Finalizar
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()
	result.Success = len(result.Errors) == 0 && len(result.Files) > 0

	if result.Success {
		s.Status = StatusSuccess
		log.Printf("[Scraper] Ejecución completada exitosamente. Archivos: %d", len(result.Files))
	} else if len(result.Files) > 0 {
		s.Status = StatusSuccess // Parcialmente exitoso
		log.Printf("[Scraper] Ejecución parcial. Archivos: %d, Errores: %d", len(result.Files), len(result.Errors))
	} else {
		s.Status = StatusFailed
		log.Printf("[Scraper] Ejecución fallida. Errores: %d", len(result.Errors))
	}

	// Registrar métricas
	s.Metrics.RecordRun(result)

	log.Println("[Scraper] ========================================")
	log.Printf("[Scraper] Duración total: %s", result.Duration)
	log.Println("[Scraper] ========================================")

	return result, nil
}

// Stop detiene la ejecución del scraper
func (s *GCOScraper) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsRunning {
		return fmt.Errorf("el scraper no está en ejecución")
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	s.Status = StatusCancelled
	log.Println("[Scraper] Ejecución cancelada")

	return nil
}

// GetStatus devuelve el estado actual del scraper
func (s *GCOScraper) GetStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErrorStr string
	if s.LastError != nil {
		lastErrorStr = s.LastError.Error()
	}

	return map[string]interface{}{
		"status":     s.Status,
		"is_running": s.IsRunning,
		"last_run":   s.LastRun,
		"last_error": lastErrorStr,
		"base_url":   s.BaseURL,
	}
}

// RunWithRetry ejecuta el scraper con reintentos
func (s *GCOScraper) RunWithRetry(ctx context.Context, config RetryConfig) (*ScraperResult, error) {
	var lastResult *ScraperResult
	var lastErr error

	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[Scraper] Reintento %d/%d después de %v", attempt, config.MaxRetries, delay)
			time.Sleep(delay)

			// Exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		result, err := s.Run(ctx)
		lastResult = result
		lastErr = err

		if err == nil && result.Success {
			return result, nil
		}

		// Si hay algunos archivos descargados, consideramos parcialmente exitoso
		if result != nil && len(result.Files) > 0 {
			log.Printf("[Scraper] Ejecución parcialmente exitosa con %d archivos", len(result.Files))
			return result, nil
		}
	}

	return lastResult, fmt.Errorf("falló después de %d reintentos: %w", config.MaxRetries, lastErr)
}
