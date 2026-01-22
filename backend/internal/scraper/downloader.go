package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Downloader maneja la navegación y descarga de CSVs
type Downloader struct {
	DownloadDir string
	Configs     map[CSVType]DownloadConfig
}

// NewDownloader crea un nuevo Downloader
func NewDownloader(downloadDir string) *Downloader {
	return &Downloader{
		DownloadDir: downloadDir,
		Configs:     DefaultDownloadConfigs,
	}
}

// DownloadCSV descarga un CSV específico del portal
func (d *Downloader) DownloadCSV(ctx context.Context, csvType CSVType) (*DownloadedFile, error) {
	config, ok := d.Configs[csvType]
	if !ok {
		return nil, fmt.Errorf("configuración no encontrada para: %s", csvType)
	}

	log.Printf("[Downloader] Iniciando descarga de %s...", csvType)

	// Navegar por los menús hasta llegar a la sección deseada
	for i, menuItem := range config.MenuPath {
		log.Printf("[Downloader] Navegando a menú: %s (%d/%d)", menuItem, i+1, len(config.MenuPath))

		err := d.clickMenu(ctx, menuItem)
		if err != nil {
			// Tomar screenshot para debug
			d.takeScreenshot(ctx, fmt.Sprintf("menu_error_%s_%d.png", csvType, i))
			return nil, fmt.Errorf("error navegando a menú '%s': %w", menuItem, err)
		}

		// Esperar a que cargue la página
		time.Sleep(2 * time.Second)
	}

	// Esperar a que la tabla/grid de datos esté visible
	log.Println("[Downloader] Esperando a que carguen los datos...")
	err := d.waitForData(ctx, config.WaitSelector)
	if err != nil {
		log.Printf("[Downloader] Advertencia: no se detectó tabla de datos: %v", err)
		// Continuar de todos modos, puede que la estructura sea diferente
	}

	// Buscar y hacer click en el botón de exportar
	log.Println("[Downloader] Buscando botón de exportar...")
	downloadPath, err := d.clickExportAndDownload(ctx, config)
	if err != nil {
		d.takeScreenshot(ctx, fmt.Sprintf("export_error_%s.png", csvType))
		return nil, fmt.Errorf("error exportando %s: %w", csvType, err)
	}

	// Obtener información del archivo descargado
	fileInfo, err := os.Stat(downloadPath)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo info del archivo: %w", err)
	}

	result := &DownloadedFile{
		Type:         csvType,
		Path:         downloadPath,
		Size:         fileInfo.Size(),
		DownloadedAt: time.Now(),
	}

	log.Printf("[Downloader] %s descargado exitosamente: %s (%d bytes)", csvType, downloadPath, result.Size)

	return result, nil
}

// clickMenu hace click en un elemento del menú
func (d *Downloader) clickMenu(ctx context.Context, menuItem string) error {
	// Selectores para encontrar elementos del menú
	menuSelectors := []string{
		fmt.Sprintf(`a:contains("%s")`, menuItem),
		fmt.Sprintf(`span:contains("%s")`, menuItem),
		fmt.Sprintf(`li:contains("%s") > a`, menuItem),
		fmt.Sprintf(`[data-menu="%s"]`, menuItem),
		fmt.Sprintf(`[title="%s"]`, menuItem),
		fmt.Sprintf(`nav a[href*="%s" i]`, strings.ToLower(menuItem)),
		fmt.Sprintf(`.menu-item:contains("%s")`, menuItem),
		fmt.Sprintf(`.nav-link:contains("%s")`, menuItem),
	}

	// Primero intentar con JavaScript para manejar :contains()
	for _, sel := range menuSelectors {
		var clicked bool

		// Usar evaluación JavaScript para selectores con :contains()
		if strings.Contains(sel, ":contains") {
			// Convertir :contains a XPath o búsqueda JS
			searchText := strings.ToLower(menuItem)
			jsCode := fmt.Sprintf(`
				(function() {
					const elements = document.querySelectorAll('a, span, li, button, div');
					for (let el of elements) {
						if (el.textContent.toLowerCase().includes('%s')) {
							el.click();
							return true;
						}
					}
					return false;
				})()
			`, searchText)

			err := chromedp.Run(ctx, chromedp.Evaluate(jsCode, &clicked))
			if err == nil && clicked {
				log.Printf("[Downloader] Menú '%s' clickeado via JS", menuItem)
				return nil
			}
		} else {
			// Selector CSS normal
			var exists bool
			err := chromedp.Run(ctx,
				chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
			)
			if err == nil && exists {
				err = chromedp.Run(ctx,
					chromedp.Click(sel, chromedp.ByQuery),
				)
				if err == nil {
					log.Printf("[Downloader] Menú '%s' clickeado via selector: %s", menuItem, sel)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no se encontró elemento de menú: %s", menuItem)
}

// waitForData espera a que los datos estén visibles
func (d *Downloader) waitForData(ctx context.Context, waitSelector string) error {
	selectors := strings.Split(waitSelector, ", ")

	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			// Esperar a que sea visible
			ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err = chromedp.Run(ctx2, chromedp.WaitVisible(sel, chromedp.ByQuery))
			if err == nil {
				log.Printf("[Downloader] Datos visibles (selector: %s)", sel)
				return nil
			}
		}
	}

	return fmt.Errorf("no se encontró indicador de datos cargados")
}

// clickExportAndDownload hace click en exportar y espera la descarga
func (d *Downloader) clickExportAndDownload(ctx context.Context, config DownloadConfig) (string, error) {
	exportSelectors := strings.Split(config.ExportButton, ", ")

	// También buscar por texto
	exportTexts := []string{"Exportar", "Export", "Excel", "CSV", "Descargar", "Download"}

	// Intentar selectores CSS primero
	for _, sel := range exportSelectors {
		sel = strings.TrimSpace(sel)
		if strings.Contains(sel, ":contains") {
			continue // Lo manejaremos con JS
		}

		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			log.Printf("[Downloader] Botón export encontrado: %s", sel)

			// Hacer click
			err = chromedp.Run(ctx, chromedp.Click(sel, chromedp.ByQuery))
			if err == nil {
				return d.waitForDownload(config.FileName)
			}
		}
	}

	// Buscar por texto con JavaScript
	for _, text := range exportTexts {
		var clicked bool
		jsCode := fmt.Sprintf(`
			(function() {
				const elements = document.querySelectorAll('button, a, input[type="button"], input[type="submit"]');
				for (let el of elements) {
					if (el.textContent.toLowerCase().includes('%s') ||
						(el.title && el.title.toLowerCase().includes('%s')) ||
						(el.value && el.value.toLowerCase().includes('%s'))) {
						el.click();
						return true;
					}
				}
				return false;
			})()
		`, strings.ToLower(text), strings.ToLower(text), strings.ToLower(text))

		err := chromedp.Run(ctx, chromedp.Evaluate(jsCode, &clicked))
		if err == nil && clicked {
			log.Printf("[Downloader] Botón '%s' clickeado", text)
			return d.waitForDownload(config.FileName)
		}
	}

	// Último intento: buscar cualquier botón/link con icono de descarga
	var clicked bool
	jsCode := `
		(function() {
			// Buscar iconos comunes de descarga/export
			const icons = document.querySelectorAll('[class*="download"], [class*="export"], [class*="excel"], [class*="csv"]');
			for (let icon of icons) {
				let clickable = icon.closest('button, a');
				if (clickable) {
					clickable.click();
					return true;
				}
			}
			return false;
		})()
	`
	err := chromedp.Run(ctx, chromedp.Evaluate(jsCode, &clicked))
	if err == nil && clicked {
		log.Println("[Downloader] Botón con icono de descarga clickeado")
		return d.waitForDownload(config.FileName)
	}

	return "", fmt.Errorf("no se encontró botón de exportar")
}

// waitForDownload espera a que se complete la descarga
func (d *Downloader) waitForDownload(expectedFileName string) (string, error) {
	log.Printf("[Downloader] Esperando descarga de %s...", expectedFileName)

	// Ruta esperada del archivo
	expectedPath := filepath.Join(d.DownloadDir, expectedFileName)

	// También buscar archivos recientes en el directorio de descarga
	timeout := 120 * time.Second
	deadline := time.Now().Add(timeout)
	checkInterval := 1 * time.Second

	var lastSize int64 = -1
	stableCount := 0

	for time.Now().Before(deadline) {
		// Buscar el archivo esperado
		if info, err := os.Stat(expectedPath); err == nil {
			currentSize := info.Size()

			// Verificar que el archivo no está siendo escrito (tamaño estable)
			if currentSize == lastSize && currentSize > 0 {
				stableCount++
				if stableCount >= 3 {
					log.Printf("[Downloader] Archivo descargado: %s (%d bytes)", expectedPath, currentSize)
					return expectedPath, nil
				}
			} else {
				stableCount = 0
				lastSize = currentSize
			}
		}

		// Buscar cualquier archivo CSV reciente en el directorio
		files, _ := filepath.Glob(filepath.Join(d.DownloadDir, "*.csv"))
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}

			// Si el archivo fue modificado en los últimos 2 minutos
			if time.Since(info.ModTime()) < 2*time.Minute {
				currentSize := info.Size()

				// Verificar estabilidad
				if currentSize > 0 {
					// Renombrar al nombre esperado si es diferente
					if file != expectedPath {
						os.Rename(file, expectedPath)
						file = expectedPath
					}

					time.Sleep(2 * time.Second)

					// Verificar tamaño final
					if finalInfo, err := os.Stat(file); err == nil && finalInfo.Size() == currentSize {
						log.Printf("[Downloader] Archivo descargado (detectado): %s (%d bytes)", file, currentSize)
						return file, nil
					}
				}
			}
		}

		// También buscar archivos .xlsx (Excel) que podrían necesitar conversión
		xlsxFiles, _ := filepath.Glob(filepath.Join(d.DownloadDir, "*.xlsx"))
		for _, file := range xlsxFiles {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}

			if time.Since(info.ModTime()) < 2*time.Minute && info.Size() > 0 {
				log.Printf("[Downloader] Archivo Excel descargado: %s", file)
				// Por ahora retornamos el xlsx, después se puede añadir conversión
				return file, nil
			}
		}

		time.Sleep(checkInterval)
	}

	return "", fmt.Errorf("timeout esperando descarga de %s", expectedFileName)
}

// takeScreenshot toma una captura de pantalla para debug
func (d *Downloader) takeScreenshot(ctx context.Context, filename string) {
	var screenshot []byte
	err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&screenshot))
	if err != nil {
		log.Printf("[Downloader] Error tomando screenshot: %v", err)
		return
	}

	screenshotPath := filepath.Join(os.Getenv("GCO_LOG_DIR"), filename)
	err = os.WriteFile(screenshotPath, screenshot, 0644)
	if err != nil {
		log.Printf("[Downloader] Error guardando screenshot: %v", err)
		return
	}

	log.Printf("[Downloader] Screenshot guardado: %s", screenshotPath)
}

// DownloadAll descarga todos los tipos de CSV
func (d *Downloader) DownloadAll(ctx context.Context) ([]*DownloadedFile, []error) {
	var files []*DownloadedFile
	var errors []error

	csvTypes := []CSVType{CSVClientes, CSVPolizas, CSVRecibos, CSVSiniestros}

	for _, csvType := range csvTypes {
		file, err := d.DownloadCSV(ctx, csvType)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", csvType, err))
			continue
		}
		files = append(files, file)
	}

	return files, errors
}
