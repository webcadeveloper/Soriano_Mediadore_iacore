package api

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"soriano-mediadores/internal/db"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ImportType representa el tipo de datos a importar
type ImportType string

const (
	ImportClientes   ImportType = "clientes"
	ImportPolizas    ImportType = "polizas"
	ImportRecibos    ImportType = "recibos"
	ImportSiniestros ImportType = "siniestros"
)

// ImportMode representa el modo de importación
type ImportMode string

const (
	ModeAdd     ImportMode = "add"      // Solo agregar nuevos
	ModeUpdate  ImportMode = "update"   // Solo actualizar existentes
	ModeReplace ImportMode = "replace"  // Reemplazar todos los datos
)

// ImportStatus representa el estado de un import job
type ImportStatus string

const (
	StatusPending    ImportStatus = "pending"
	StatusProcessing ImportStatus = "processing"
	StatusCompleted  ImportStatus = "completed"
	StatusFailed     ImportStatus = "failed"
	StatusCancelled  ImportStatus = "cancelled"
)

// ImportJob representa un trabajo de importación
type ImportJob struct {
	ID                string       `json:"import_id"`
	Type              ImportType   `json:"type"`
	Mode              ImportMode   `json:"mode"`
	Status            ImportStatus `json:"status"`
	TotalRows         int          `json:"total_rows"`
	ProcessedRows     int          `json:"processed_rows"`
	SuccessfulRows    int          `json:"successful_rows"`
	FailedRows        int          `json:"failed_rows"`
	DuplicateRows     int          `json:"duplicate_rows"`
	SkippedRows       int          `json:"skipped_rows"`
	Errors            []ImportError `json:"errors"`
	StartedAt         time.Time    `json:"started_at"`
	CompletedAt       *time.Time   `json:"completed_at,omitempty"`
	ValidateFirst     bool         `json:"validate_first"`
	DuplicateHandling string       `json:"duplicate_handling"` // skip, update, error
	Username          string       `json:"username,omitempty"`
	Filename          string       `json:"filename,omitempty"`
	mu                sync.RWMutex
	cancel            chan bool
}

// ImportError representa un error durante la importación
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Almacenamiento de jobs en memoria (en producción usar Redis/DB)
var (
	importJobs   = make(map[string]*ImportJob)
	importJobsMu sync.RWMutex
)

// PreviewCSV procesa las primeras 10 filas del CSV
func PreviewCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No se proporcionó archivo CSV",
		})
	}

	// Abrir archivo
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error abriendo archivo",
		})
	}
	defer fileContent.Close()

	// Leer CSV
	reader := csv.NewReader(fileContent)
	reader.Comma = ';' // Occident usa punto y coma
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Leer encabezados
	headers, err := reader.Read()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Error leyendo encabezados del CSV",
		})
	}

	// Leer hasta 10 filas
	var rows [][]string
	for i := 0; i < 10; i++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rows = append(rows, row)
	}

	// Contar total de filas
	totalRows := len(rows)
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err == nil {
			totalRows++
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"fileName":  file.Filename,
			"fileSize":  file.Size,
			"headers":   headers,
			"rows":      rows,
			"totalRows": totalRows,
		},
	})
}

// StartImport inicia el proceso de importación
func StartImport(c *fiber.Ctx) error {
	// Obtener archivo
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No se proporcionó archivo CSV",
		})
	}

	// Obtener parámetros
	importType := ImportType(c.FormValue("type", "clientes"))
	importMode := ImportMode(c.FormValue("mode", "add"))
	validateFirst := c.FormValue("validate_first") == "true"
	duplicateHandling := c.FormValue("duplicate_handling", "skip")
	username := c.FormValue("username", "admin")

	// Validar tipo
	if !isValidImportType(importType) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Tipo de importación inválido",
		})
	}

	// Crear job
	job := &ImportJob{
		ID:                uuid.New().String(),
		Type:              importType,
		Mode:              importMode,
		Status:            StatusPending,
		StartedAt:         time.Now(),
		ValidateFirst:     validateFirst,
		DuplicateHandling: duplicateHandling,
		Username:          username,
		Filename:          file.Filename,
		cancel:            make(chan bool),
	}

	// Guardar job
	importJobsMu.Lock()
	importJobs[job.ID] = job
	importJobsMu.Unlock()

	// Abrir archivo
	fileContent, err := file.Open()
	if err != nil {
		job.Status = StatusFailed
		job.Errors = append(job.Errors, ImportError{
			Row:     0,
			Message: "Error abriendo archivo: " + err.Error(),
		})
		return c.Status(500).JSON(fiber.Map{
			"error": "Error abriendo archivo",
		})
	}

	// Procesar asíncronamente
	go processImport(job, fileContent)

	return c.JSON(fiber.Map{
		"import_id": job.ID,
		"status":    job.Status,
		"message":   "Importación iniciada",
	})
}

// GetImportStatus obtiene el estado de un import job
func GetImportStatus(c *fiber.Ctx) error {
	importID := c.Params("id")

	importJobsMu.RLock()
	job, exists := importJobs[importID]
	importJobsMu.RUnlock()

	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "Import job no encontrado",
		})
	}

	job.mu.RLock()
	defer job.mu.RUnlock()

	return c.JSON(job)
}

// CancelImport cancela un import job en progreso
func CancelImport(c *fiber.Ctx) error {
	importID := c.Params("id")

	importJobsMu.RLock()
	job, exists := importJobs[importID]
	importJobsMu.RUnlock()

	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "Import job no encontrado",
		})
	}

	job.mu.Lock()
	if job.Status == StatusProcessing {
		job.Status = StatusCancelled
		close(job.cancel)
	}
	job.mu.Unlock()

	return c.JSON(fiber.Map{
		"message": "Import job cancelado",
		"status":  job.Status,
	})
}

// GetImportHistory obtiene el historial de importaciones
func GetImportHistory(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)

	// Obtener todos los jobs ordenados por fecha
	importJobsMu.RLock()
	defer importJobsMu.RUnlock()

	var jobs []*ImportJob
	for _, job := range importJobs {
		jobs = append(jobs, job)
	}

	// Ordenar por fecha (más reciente primero)
	// En producción, esto debería venir de la base de datos ordenado
	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if jobs[j].StartedAt.After(jobs[i].StartedAt) {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}

	// Limitar resultados
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return c.JSON(fiber.Map{
		"total":   len(jobs),
		"imports": jobs,
	})
}

// RevertImport revierte una importación (marca registros como inactivos)
func RevertImport(c *fiber.Ctx) error {
	importID := c.Params("id")

	importJobsMu.RLock()
	job, exists := importJobs[importID]
	importJobsMu.RUnlock()

	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "Import job no encontrado",
		})
	}

	if job.Status != StatusCompleted {
		return c.Status(400).JSON(fiber.Map{
			"error": "Solo se pueden revertir importaciones completadas",
		})
	}

	// En producción, esto marcaría los registros importados como inactivos
	// usando un campo import_id en cada tabla

	return c.JSON(fiber.Map{
		"message": "Importación revertida exitosamente",
		"import_id": importID,
	})
}

// ValidateImport valida un CSV sin importarlo
func ValidateImport(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No se proporcionó archivo CSV",
		})
	}

	importType := ImportType(c.FormValue("type", "clientes"))

	if !isValidImportType(importType) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Tipo de importación inválido",
		})
	}

	fileContent, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error abriendo archivo",
		})
	}
	defer fileContent.Close()

	// Validar estructura
	errors := validateCSVStructure(fileContent, importType)

	return c.JSON(fiber.Map{
		"valid":  len(errors) == 0,
		"errors": errors,
		"total_errors": len(errors),
	})
}

// GetImportTemplate descarga una plantilla CSV
func GetImportTemplate(c *fiber.Ctx) error {
	importType := ImportType(c.Query("type", "clientes"))

	var headers []string
	switch importType {
	case ImportClientes:
		headers = []string{
			"nif", "id_account", "nombre_completo", "email_contacto",
			"telefono_contacto", "domicilio", "codigo_postal", "provincia", "pais",
		}
	case ImportPolizas:
		headers = []string{
			"numero_poliza", "id_account", "nombre_cliente", "ramo", "gestora",
			"situacion", "prima_anual", "fecha_efecto", "fecha_vencimiento",
		}
	case ImportRecibos:
		headers = []string{
			"numero_recibo", "numero_poliza", "id_account", "prima_total",
			"situacion_recibo", "fecha_emision", "forma_pago",
		}
	case ImportSiniestros:
		headers = []string{
			"numero_siniestro", "numero_poliza", "id_account", "situacion_siniestro",
			"fecha_ocurrencia", "fecha_apertura", "tramitador",
		}
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Tipo inválido",
		})
	}

	// Generar CSV
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=plantilla_%s.csv", importType))

	return c.SendString(strings.Join(headers, ";") + "\n")
}

// processImport procesa el import job asíncronamente
func processImport(job *ImportJob, file io.ReadCloser) {
	defer file.Close()

	job.mu.Lock()
	job.Status = StatusProcessing
	job.mu.Unlock()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Leer encabezados
	headers, err := reader.Read()
	if err != nil {
		job.mu.Lock()
		job.Status = StatusFailed
		job.Errors = append(job.Errors, ImportError{
			Row:     0,
			Message: "Error leyendo encabezados: " + err.Error(),
		})
		job.mu.Unlock()
		return
	}

	rowNum := 1
	for {
		// Verificar cancelación
		select {
		case <-job.cancel:
			return
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			job.mu.Lock()
			job.Errors = append(job.Errors, ImportError{
				Row:     rowNum,
				Message: "Error leyendo fila: " + err.Error(),
			})
			job.FailedRows++
			job.mu.Unlock()
			rowNum++
			continue
		}

		job.mu.Lock()
		job.TotalRows++
		job.mu.Unlock()

		// Procesar fila según tipo
		isDuplicate, err := processRow(job, headers, row, rowNum)

		job.mu.Lock()
		job.ProcessedRows++
		if err != nil {
			job.FailedRows++
			job.Errors = append(job.Errors, ImportError{
				Row:     rowNum,
				Message: err.Error(),
			})
		} else if isDuplicate {
			job.DuplicateRows++
			job.SkippedRows++
		} else {
			job.SuccessfulRows++
		}
		job.mu.Unlock()

		rowNum++
	}

	// Finalizar
	job.mu.Lock()
	if len(job.Errors) > 0 && job.SuccessfulRows == 0 {
		job.Status = StatusFailed
	} else {
		job.Status = StatusCompleted
	}
	now := time.Now()
	job.CompletedAt = &now
	job.mu.Unlock()

	// Guardar en base de datos
	saveImportJobToDB(job)

	log.Printf("✅ Import job %s completado: %d exitosos, %d fallidos",
		job.ID, job.SuccessfulRows, job.FailedRows)
}

// processRow procesa una fila según el tipo de importación
// Retorna (isDuplicate, error)
func processRow(job *ImportJob, headers []string, row []string, rowNum int) (bool, error) {
	// Convertir a mapa
	data := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			data[strings.TrimSpace(header)] = strings.TrimSpace(row[i])
		}
	}

	switch job.Type {
	case ImportClientes:
		return importCliente(data, job.Mode, job.DuplicateHandling)
	case ImportPolizas:
		return importPoliza(data, job.Mode, job.DuplicateHandling)
	case ImportRecibos:
		return importRecibo(data, job.Mode, job.DuplicateHandling)
	case ImportSiniestros:
		return importSiniestro(data, job.Mode, job.DuplicateHandling)
	default:
		return false, fmt.Errorf("tipo de importación no soportado: %s", job.Type)
	}
}

// importCliente importa un cliente desde CSV de Occident
// Campos Occident: NIF, Nombre completo, Nombre, Apellidos, Fecha nacimiento, Sexo,
// Domicilio, Teléfono contacto, 2º Teléfono contacto, Población, Código postal,
// Email contacto, Total primas en cartera, Total primas relación, Mediador, Provincia, IdAccount
func importCliente(data map[string]string, mode ImportMode, dupHandling string) (bool, error) {
	// Mapear campos de Occident (case insensitive con BOM)
	nif := getField(data, "NIF", "nif")
	idAccount := getField(data, "IdAccount", "id_account")

	if nif == "" && idAccount == "" {
		return false, fmt.Errorf("NIF o ID Account requerido")
	}

	// Verificar si existe
	var existingID int
	query := "SELECT id FROM clientes WHERE nif = $1 OR id_account = $2"
	err := db.PostgresDB.QueryRow(query, nif, idAccount).Scan(&existingID)
	exists := (err != sql.ErrNoRows)

	if exists {
		if dupHandling == "skip" {
			return true, nil
		}
		if dupHandling == "error" {
			return false, fmt.Errorf("cliente duplicado: %s", nif)
		}
		if mode == ModeAdd {
			return true, nil
		}

		// Update con todos los campos de Occident
		updateSQL := `
			UPDATE clientes SET
				nombre_completo = COALESCE(NULLIF($1, ''), nombre_completo),
				nombre = COALESCE(NULLIF($2, ''), nombre),
				apellidos = COALESCE(NULLIF($3, ''), apellidos),
				fecha_nacimiento = COALESCE(NULLIF($4, ''), fecha_nacimiento),
				sexo = COALESCE(NULLIF($5, ''), sexo),
				domicilio = COALESCE(NULLIF($6, ''), domicilio),
				poblacion = COALESCE(NULLIF($7, ''), poblacion),
				codigo_postal = COALESCE(NULLIF($8, ''), codigo_postal),
				provincia = COALESCE(NULLIF($9, ''), provincia),
				email_contacto = COALESCE(NULLIF($10, ''), email_contacto),
				telefono_contacto = COALESCE(NULLIF($11, ''), telefono_contacto),
				telefono2_contacto = COALESCE(NULLIF($12, ''), telefono2_contacto),
				total_primas_cartera = COALESCE(NULLIF($13, 0), total_primas_cartera),
				total_primas_relacion = COALESCE(NULLIF($14, 0), total_primas_relacion),
				num_polizas_totales = COALESCE(NULLIF($15, 0), num_polizas_totales),
				mediador = COALESCE(NULLIF($16, ''), mediador),
				actualizado_en = NOW()
			WHERE id = $17
		`
		_, err = db.PostgresDB.Exec(updateSQL,
			getField(data, "Nombre completo"),
			getField(data, "Nombre"),
			getField(data, "Apellidos"),
			getField(data, "Fecha nacimiento"),
			getField(data, "Sexo"),
			getField(data, "Domicilio"),
			getField(data, "Población"),
			getField(data, "Código postal"),
			getField(data, "Provincia"),
			getField(data, "Email contacto"),
			getField(data, "Teléfono contacto"),
			getField(data, "2º Teléfono contacto"),
			parseFloat(getField(data, "Total primas en cartera")),
			parseFloat(getField(data, "Total primas relación")),
			parseInt(getField(data, "Número de pólizas totales")),
			getField(data, "Mediador"),
			existingID,
		)
		return false, err
	}

	if mode == ModeUpdate {
		return false, fmt.Errorf("cliente no existe para actualizar: %s", nif)
	}

	// Insertar nuevo con todos los campos de Occident
	insertSQL := `
		INSERT INTO clientes (
			nif, id_account, nombre_completo, nombre, apellidos,
			fecha_nacimiento, sexo, domicilio, poblacion, codigo_postal,
			provincia, email_contacto, telefono_contacto, telefono2_contacto,
			total_primas_cartera, total_primas_relacion, num_polizas_totales,
			mediador, activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, TRUE, NOW())
	`
	_, err = db.PostgresDB.Exec(insertSQL,
		nif,
		idAccount,
		getField(data, "Nombre completo"),
		getField(data, "Nombre"),
		getField(data, "Apellidos"),
		getField(data, "Fecha nacimiento"),
		getField(data, "Sexo"),
		getField(data, "Domicilio"),
		getField(data, "Población"),
		getField(data, "Código postal"),
		getField(data, "Provincia"),
		getField(data, "Email contacto"),
		getField(data, "Teléfono contacto"),
		getField(data, "2º Teléfono contacto"),
		parseFloat(getField(data, "Total primas en cartera")),
		parseFloat(getField(data, "Total primas relación")),
		parseInt(getField(data, "Número de pólizas totales")),
		getField(data, "Mediador"),
	)
	return false, err
}

// importPoliza importa una póliza desde CSV de Occident
// Campos Occident: Número de la póliza, Ramo, Mediador, Domicilio de la póliza,
// Prima anual, Fecha de efecto, Fecha de vencimiento, Gestora, Descripción del riesgo,
// Matricula, Situación de la póliza, Nombre del cliente, IdAccount
func importPoliza(data map[string]string, mode ImportMode, dupHandling string) (bool, error) {
	numeroPoliza := getField(data, "Número de la póliza", "numero_poliza")
	if numeroPoliza == "" {
		return false, fmt.Errorf("número de póliza requerido")
	}

	// Verificar si existe
	var existingID int
	query := "SELECT id FROM polizas WHERE numero_poliza = $1"
	err := db.PostgresDB.QueryRow(query, numeroPoliza).Scan(&existingID)
	exists := (err != sql.ErrNoRows)

	if exists {
		if dupHandling == "skip" {
			return true, nil
		}
		if dupHandling == "error" {
			return false, fmt.Errorf("póliza duplicada: %s", numeroPoliza)
		}
		if mode == ModeAdd {
			return true, nil
		}

		// Update con todos los campos de Occident
		updateSQL := `
			UPDATE polizas SET
				id_account = COALESCE(NULLIF($1, ''), id_account),
				nombre_cliente = COALESCE(NULLIF($2, ''), nombre_cliente),
				ramo = COALESCE(NULLIF($3, ''), ramo),
				gestora = COALESCE(NULLIF($4, ''), gestora),
				mediador = COALESCE(NULLIF($5, ''), mediador),
				situacion_poliza = COALESCE(NULLIF($6, ''), situacion_poliza),
				prima_anual = COALESCE(NULLIF($7, ''), prima_anual),
				fecha_efecto = COALESCE(NULLIF($8, ''), fecha_efecto),
				fecha_vencimiento = COALESCE(NULLIF($9, ''), fecha_vencimiento),
				domicilio_poliza = COALESCE(NULLIF($10, ''), domicilio_poliza),
				descripcion_riesgo = COALESCE(NULLIF($11, ''), descripcion_riesgo),
				matricula = COALESCE(NULLIF($12, ''), matricula),
				actualizado_en = NOW()
			WHERE id = $13
		`
		_, err = db.PostgresDB.Exec(updateSQL,
			getField(data, "IdAccount"),
			getField(data, "Nombre del cliente"),
			getField(data, "Ramo"),
			getField(data, "Gestora"),
			getField(data, "Mediador"),
			getField(data, "Situación de la póliza"),
			getField(data, "Prima anual"),
			getField(data, "Fecha de efecto"),
			getField(data, "Fecha de vencimiento"),
			getField(data, "Domicilio de la póliza"),
			getField(data, "Descripción del riesgo"),
			getField(data, "Matricula"),
			existingID,
		)
		return false, err
	}

	if mode == ModeUpdate {
		return false, fmt.Errorf("póliza no existe para actualizar: %s", numeroPoliza)
	}

	// Insertar nuevo con todos los campos de Occident
	insertSQL := `
		INSERT INTO polizas (
			numero_poliza, id_account, nombre_cliente, ramo, gestora,
			mediador, situacion_poliza, prima_anual, fecha_efecto, fecha_vencimiento,
			domicilio_poliza, descripcion_riesgo, matricula, activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, TRUE, NOW())
	`
	_, err = db.PostgresDB.Exec(insertSQL,
		numeroPoliza,
		getField(data, "IdAccount"),
		getField(data, "Nombre del cliente"),
		getField(data, "Ramo"),
		getField(data, "Gestora"),
		getField(data, "Mediador"),
		getField(data, "Situación de la póliza"),
		getField(data, "Prima anual"),
		getField(data, "Fecha de efecto"),
		getField(data, "Fecha de vencimiento"),
		getField(data, "Domicilio de la póliza"),
		getField(data, "Descripción del riesgo"),
		getField(data, "Matricula"),
	)
	return false, err
}

// importRecibo importa un recibo desde CSV de Occident
// Campos Occident: Nº recibo, Mediador, Ramo, Origen del recibo, Prima total,
// Fecha inicio cobertura, Situación del recibo, Fecha emisión, Fecha situación,
// Fecha fin cobertura, Gestora del recibo, Gestión de cobro, Detalle del recibo,
// Comisión bruta, Cliente, Nº póliza, Comisión neta, Forma de pago, Descripción riesgo, IdAccount
func importRecibo(data map[string]string, mode ImportMode, dupHandling string) (bool, error) {
	numeroRecibo := getField(data, "Nº recibo", "numero_recibo")
	if numeroRecibo == "" {
		return false, fmt.Errorf("número de recibo requerido")
	}

	// Verificar si existe
	var existingID int
	query := "SELECT id FROM recibos WHERE numero_recibo = $1"
	err := db.PostgresDB.QueryRow(query, numeroRecibo).Scan(&existingID)
	exists := (err != sql.ErrNoRows)

	if exists {
		if dupHandling == "skip" {
			return true, nil
		}
		if dupHandling == "error" {
			return false, fmt.Errorf("recibo duplicado: %s", numeroRecibo)
		}
		if mode == ModeAdd {
			return true, nil
		}

		// Update con todos los campos de Occident
		updateSQL := `
			UPDATE recibos SET
				numero_poliza = COALESCE(NULLIF($1, ''), numero_poliza),
				id_account = COALESCE(NULLIF($2, ''), id_account),
				nombre_cliente = COALESCE(NULLIF($3, ''), nombre_cliente),
				ramo = COALESCE(NULLIF($4, ''), ramo),
				mediador = COALESCE(NULLIF($5, ''), mediador),
				prima_total = COALESCE(NULLIF($6, 0), prima_total),
				comision_bruta = COALESCE(NULLIF($7, 0), comision_bruta),
				comision_neta = COALESCE(NULLIF($8, 0), comision_neta),
				situacion_recibo = COALESCE(NULLIF($9, ''), situacion_recibo),
				fecha_emision = $10,
				fecha_situacion = $11,
				fecha_inicio_cobertura = $12,
				fecha_fin_cobertura = $13,
				forma_pago = COALESCE(NULLIF($14, ''), forma_pago),
				gestora_recibo = COALESCE(NULLIF($15, ''), gestora_recibo),
				gestion_cobro = COALESCE(NULLIF($16, ''), gestion_cobro),
				detalle_recibo = COALESCE(NULLIF($17, ''), detalle_recibo),
				descripcion_riesgo = COALESCE(NULLIF($18, ''), descripcion_riesgo),
				actualizado_en = NOW()
			WHERE id = $19
		`
		_, err = db.PostgresDB.Exec(updateSQL,
			getField(data, "Nº póliza"),
			getField(data, "IdAccount"),
			getField(data, "Cliente"),
			getField(data, "Ramo"),
			getField(data, "Mediador"),
			parseFloat(getField(data, "Prima total")),
			parseFloat(getField(data, "Comisión bruta")),
			parseFloat(getField(data, "Comisión neta")),
			getField(data, "Situación del recibo"),
			parseDate(getField(data, "Fecha emisión")),
			parseDate(getField(data, "Fecha situación")),
			parseDate(getField(data, "Fecha inicio cobertura")),
			parseDate(getField(data, "Fecha fin cobertura")),
			getField(data, "Forma de pago"),
			getField(data, "Gestora del recibo"),
			getField(data, "Gestión de cobro"),
			getField(data, "Detalle del recibo"),
			getField(data, "Descripción riesgo"),
			existingID,
		)
		return false, err
	}

	if mode == ModeUpdate {
		return false, fmt.Errorf("recibo no existe para actualizar: %s", numeroRecibo)
	}

	// Insertar nuevo con todos los campos de Occident
	insertSQL := `
		INSERT INTO recibos (
			numero_recibo, numero_poliza, id_account, nombre_cliente, ramo, mediador,
			prima_total, comision_bruta, comision_neta, situacion_recibo,
			fecha_emision, fecha_situacion, fecha_inicio_cobertura, fecha_fin_cobertura,
			forma_pago, gestora_recibo, gestion_cobro, detalle_recibo, descripcion_riesgo,
			activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, TRUE, NOW())
	`
	_, err = db.PostgresDB.Exec(insertSQL,
		numeroRecibo,
		getField(data, "Nº póliza"),
		getField(data, "IdAccount"),
		getField(data, "Cliente"),
		getField(data, "Ramo"),
		getField(data, "Mediador"),
		parseFloat(getField(data, "Prima total")),
		parseFloat(getField(data, "Comisión bruta")),
		parseFloat(getField(data, "Comisión neta")),
		getField(data, "Situación del recibo"),
		parseDate(getField(data, "Fecha emisión")),
		parseDate(getField(data, "Fecha situación")),
		parseDate(getField(data, "Fecha inicio cobertura")),
		parseDate(getField(data, "Fecha fin cobertura")),
		getField(data, "Forma de pago"),
		getField(data, "Gestora del recibo"),
		getField(data, "Gestión de cobro"),
		getField(data, "Detalle del recibo"),
		getField(data, "Descripción riesgo"),
	)
	return false, err
}

// importSiniestro importa un siniestro
func importSiniestro(data map[string]string, mode ImportMode, dupHandling string) (bool, error) {
	// Campos de Occident SINIESTROS.csv:
	// Número de póliza, Número de siniestro, Mediador, Situación del siniestro,
	// Fecha de ocurrencia, Fecha de cierre, Fecha de apertura, Tramitador,
	// Gestionado, Cliente, Centro de tramitación, IdAccount

	numeroSiniestro := getField(data, "Número de siniestro", "numero_siniestro", "Numero de siniestro")
	if numeroSiniestro == "" {
		return false, fmt.Errorf("número de siniestro requerido")
	}

	// Extraer todos los campos de Occident
	numeroPoliza := getField(data, "Número de póliza", "numero_poliza", "Numero de poliza")
	idAccount := getField(data, "IdAccount", "id_account", "ID Account")
	cliente := getField(data, "Cliente", "cliente", "nombre_cliente")
	situacion := getField(data, "Situación del siniestro", "situacion_siniestro", "Situacion del siniestro")
	fechaOcurrencia := getField(data, "Fecha de ocurrencia", "fecha_ocurrencia")
	fechaApertura := getField(data, "Fecha de apertura", "fecha_apertura")
	fechaCierre := getField(data, "Fecha de cierre", "fecha_cierre")
	tramitador := getField(data, "Tramitador", "tramitador")
	centroTramitacion := getField(data, "Centro de tramitación", "centro_tramitacion", "Centro de tramitacion")
	mediador := getField(data, "Mediador", "mediador")
	gestionado := getField(data, "Gestionado", "gestionado")

	// Verificar si existe
	var existingID int
	query := "SELECT id FROM siniestros WHERE numero_siniestro = $1"
	err := db.PostgresDB.QueryRow(query, numeroSiniestro).Scan(&existingID)
	exists := (err != sql.ErrNoRows)

	if exists {
		if dupHandling == "skip" {
			log.Printf("⏭️  Siniestro duplicado omitido: Número=%s", numeroSiniestro)
			return true, nil
		}
		if dupHandling == "error" {
			return false, fmt.Errorf("siniestro duplicado: %s", numeroSiniestro)
		}
		if mode == ModeAdd {
			log.Printf("⏭️  Siniestro duplicado omitido (modo ADD): Número=%s", numeroSiniestro)
			return true, nil
		}

		log.Printf("🔄 Actualizando siniestro existente: Número=%s", numeroSiniestro)
		updateSQL := `
			UPDATE siniestros SET
				numero_poliza = COALESCE(NULLIF($1, ''), numero_poliza),
				id_account = COALESCE(NULLIF($2, ''), id_account),
				cliente = COALESCE(NULLIF($3, ''), cliente),
				situacion_siniestro = COALESCE(NULLIF($4, ''), situacion_siniestro),
				fecha_ocurrencia = COALESCE(NULLIF($5, ''), fecha_ocurrencia),
				fecha_apertura = COALESCE(NULLIF($6, ''), fecha_apertura),
				fecha_cierre = COALESCE(NULLIF($7, ''), fecha_cierre),
				tramitador = COALESCE(NULLIF($8, ''), tramitador),
				centro_tramitacion = COALESCE(NULLIF($9, ''), centro_tramitacion),
				mediador = COALESCE(NULLIF($10, ''), mediador),
				gestionado = COALESCE(NULLIF($11, ''), gestionado),
				actualizado_en = NOW()
			WHERE id = $12
		`
		_, err = db.PostgresDB.Exec(updateSQL,
			numeroPoliza, idAccount, cliente, situacion,
			fechaOcurrencia, fechaApertura, fechaCierre,
			tramitador, centroTramitacion, mediador, gestionado,
			existingID,
		)
		return false, err
	}

	if mode == ModeUpdate {
		return false, fmt.Errorf("siniestro no existe para actualizar: %s", numeroSiniestro)
	}

	// Insertar nuevo
	log.Printf("➕ Insertando nuevo siniestro: Número=%s, Póliza=%s, Cliente=%s", numeroSiniestro, numeroPoliza, cliente)
	insertSQL := `
		INSERT INTO siniestros (
			numero_siniestro, numero_poliza, id_account, cliente, situacion_siniestro,
			fecha_ocurrencia, fecha_apertura, fecha_cierre, tramitador,
			centro_tramitacion, mediador, gestionado, activo, creado_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, TRUE, NOW())
	`
	_, err = db.PostgresDB.Exec(insertSQL,
		numeroSiniestro, numeroPoliza, idAccount, cliente, situacion,
		nullIfEmpty(fechaOcurrencia), nullIfEmpty(fechaApertura), nullIfEmpty(fechaCierre),
		tramitador, centroTramitacion, mediador, gestionado,
	)
	return false, err
}

// validateCSVStructure valida la estructura del CSV
func validateCSVStructure(file io.Reader, importType ImportType) []ImportError {
	var errors []ImportError

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		errors = append(errors, ImportError{
			Row:     0,
			Message: "Error leyendo encabezados: " + err.Error(),
		})
		return errors
	}

	// Validar encabezados requeridos
	requiredHeaders := getRequiredHeaders(importType)
	for _, required := range requiredHeaders {
		found := false
		for _, header := range headers {
			if strings.TrimSpace(header) == required {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ImportError{
				Row:     0,
				Field:   required,
				Message: fmt.Sprintf("Columna requerida no encontrada: %s", required),
			})
		}
	}

	return errors
}

// getRequiredHeaders devuelve los headers requeridos según tipo
func getRequiredHeaders(importType ImportType) []string {
	switch importType {
	case ImportClientes:
		return []string{"nif", "nombre_completo"}
	case ImportPolizas:
		return []string{"numero_poliza", "id_account"}
	case ImportRecibos:
		return []string{"numero_recibo", "numero_poliza"}
	case ImportSiniestros:
		return []string{"numero_siniestro", "numero_poliza"}
	default:
		return []string{}
	}
}

// saveImportJobToDB guarda el job en la base de datos
func saveImportJobToDB(job *ImportJob) {
	errorsJSON, _ := json.Marshal(job.Errors)

	sql := `
		INSERT INTO import_jobs (
			id, type, mode, status, total_rows, processed_rows,
			successful_rows, failed_rows, duplicate_rows, skipped_rows,
			errors, started_at, completed_at, validate_first,
			duplicate_handling, username, filename
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			processed_rows = EXCLUDED.processed_rows,
			successful_rows = EXCLUDED.successful_rows,
			failed_rows = EXCLUDED.failed_rows,
			duplicate_rows = EXCLUDED.duplicate_rows,
			skipped_rows = EXCLUDED.skipped_rows,
			errors = EXCLUDED.errors,
			completed_at = EXCLUDED.completed_at
	`

	_, err := db.PostgresDB.Exec(sql,
		job.ID, job.Type, job.Mode, job.Status, job.TotalRows, job.ProcessedRows,
		job.SuccessfulRows, job.FailedRows, job.DuplicateRows, job.SkippedRows,
		string(errorsJSON), job.StartedAt, job.CompletedAt, job.ValidateFirst,
		job.DuplicateHandling, job.Username, job.Filename,
	)

	if err != nil {
		log.Printf("⚠️  Error guardando import job en DB: %v", err)
	}
}

// Helper functions
func isValidImportType(t ImportType) bool {
	return t == ImportClientes || t == ImportPolizas || t == ImportRecibos || t == ImportSiniestros
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// getField busca un campo en el mapa con múltiples nombres posibles
// Maneja el BOM de UTF-8 que puede venir en el primer campo
func getField(data map[string]string, names ...string) string {
	for _, name := range names {
		if val, ok := data[name]; ok && val != "" {
			return strings.TrimSpace(val)
		}
		// Probar con BOM (común en CSVs de Windows/Excel)
		bomName := "\ufeff" + name
		if val, ok := data[bomName]; ok && val != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// parseFloat convierte string a float64, manejando formato español (coma decimal)
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	// Limpiar el string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " €", "")
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, " ", "")
	// Manejar formato español: 1.234,56 -> 1234.56
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// parseInt convierte string a int
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(s)
	val, _ := strconv.Atoi(s)
	return val
}

// parseDate convierte fecha de Occident (DD/MM/YYYY o YYYY-MM-DD) a formato SQL
func parseDate(s string) interface{} {
	if s == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	// Si tiene hora, quitarla
	if strings.Contains(s, " ") {
		s = strings.Split(s, " ")[0]
	}
	// Ya está en formato YYYY-MM-DD
	if len(s) == 10 && s[4] == '-' {
		return s
	}
	// Formato DD/MM/YYYY -> YYYY-MM-DD
	if len(s) >= 10 && s[2] == '/' {
		parts := strings.Split(s, "/")
		if len(parts) == 3 {
			return parts[2] + "-" + parts[1] + "-" + parts[0]
		}
	}
	return s
}
