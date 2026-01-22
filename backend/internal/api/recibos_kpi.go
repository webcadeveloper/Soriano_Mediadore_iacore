package api

import (
	"database/sql"
	"log"
	"soriano-mediadores/internal/db"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RecibosKPIResponse contiene todos los KPIs de recibos organizados
type RecibosKPIResponse struct {
	// KPIs Generales
	TotalRecibos       int     `json:"total_recibos"`
	TotalImporte       float64 `json:"total_importe"`

	// Por Situación
	TotalCobrados      int     `json:"total_cobrados"`
	ImporteCobrado     float64 `json:"importe_cobrado"`
	TotalPendientes    int     `json:"total_pendientes"`
	ImportePendiente   float64 `json:"importe_pendiente"`
	TotalAnulados      int     `json:"total_anulados"`
	ImporteAnulado     float64 `json:"importe_anulado"`
	TotalRetornados    int     `json:"total_retornados"`
	ImporteRetornado   float64 `json:"importe_retornado"`

	// Historial de Recibos
	Recibos            []ReciboCompleto `json:"recibos"`

	// Metadata de filtrado
	Filtros            FiltrosAplicados `json:"filtros_aplicados"`
	TotalPaginas       int              `json:"total_paginas"`
	PaginaActual       int              `json:"pagina_actual"`
	RegistrosPorPagina int              `json:"registros_por_pagina"`
}

// ReciboCompleto representa un recibo con toda su información
type ReciboCompleto struct {
	ID                   int     `json:"recibo_id"`
	NumeroRecibo         string  `json:"recibo"`
	NumeroPoliza         string  `json:"poliza"`
	IDAccount            string  `json:"id_account"`
	Cliente              string  `json:"nombre_completo"`
	NIF                  string  `json:"nif_cliente,omitempty"`
	Email                string  `json:"email,omitempty"`
	Telefono             string  `json:"telefono,omitempty"`
	Ramo                 string  `json:"ramo,omitempty"`
	Mediador             string  `json:"mediador,omitempty"`

	// Datos financieros
	PrimaTotal           float64 `json:"importe"`
	ComisionBruta        float64 `json:"comision_bruta,omitempty"`
	ComisionNeta         float64 `json:"comision_neta,omitempty"`

	// Estado
	SituacionRecibo      string  `json:"situacion"`
	DetalleRecibo        string  `json:"detalle_recibo,omitempty"`
	GestionCobro         string  `json:"gestion_cobro,omitempty"`
	FormaPago            string  `json:"forma_pago,omitempty"`

	// Fechas
	FechaEmision         string  `json:"vencimiento,omitempty"`
	FechaSituacion       string  `json:"fecha_situacion,omitempty"`
	FechaInicioCobertura string  `json:"fecha_inicio_cobertura,omitempty"`
	FechaFinCobertura    string  `json:"fecha_fin_cobertura,omitempty"`

	// Cálculos
	DiasDesdeEmision     int     `json:"dias_desde_emision"`
	DiasDesdeUltimoCambio int    `json:"dias_vencido"`
}

// FiltrosAplicados muestra qué filtros se aplicaron
type FiltrosAplicados struct {
	Situacion       []string `json:"situacion,omitempty"`
	FechaDesde      string   `json:"fecha_desde,omitempty"`
	FechaHasta      string   `json:"fecha_hasta,omitempty"`
	Mediador        string   `json:"mediador,omitempty"`
	Cliente         string   `json:"cliente,omitempty"`
	ImporteMinimo   float64  `json:"importe_minimo,omitempty"`
	ImporteMaximo   float64  `json:"importe_maximo,omitempty"`
	OrdenarPor      string   `json:"ordenar_por"`
	Orden           string   `json:"orden"`
}

// GetRecibosKPI obtiene todos los recibos organizados por KPI con filtros
func GetRecibosKPI(c *fiber.Ctx) error {
	// Parsear parámetros de query
	pagina := c.QueryInt("pagina", 1)
	limite := c.QueryInt("limite", 50) // Por defecto 50 recibos por página
	if limite > 500 {
		limite = 500 // Máximo 500 por página
	}
	if pagina < 1 {
		pagina = 1
	}

	// Filtros
	situaciones := c.Query("situacion", "") // "Cobrado,Pendiente,Anulado,Retornado"
	fechaDesde := c.Query("fecha_desde", "")
	fechaHasta := c.Query("fecha_hasta", "")
	mediador := c.Query("mediador", "")
	cliente := c.Query("cliente", "")
	importeMin := c.QueryFloat("importe_min", 0)
	importeMax := c.QueryFloat("importe_max", 0)
	ordenarPor := c.Query("ordenar_por", "fecha_emision") // fecha_emision, prima_total, situacion_recibo
	orden := c.Query("orden", "DESC")                      // ASC o DESC

	// Construir WHERE clause
	var whereClauses []string
	var args []interface{}
	argCounter := 1

	whereClauses = append(whereClauses, "r.activo = TRUE")

	// Filtro por situación
	situacionList := []string{}
	if situaciones != "" {
		situacionList = strings.Split(situaciones, ",")
		placeholders := []string{}
		for _, sit := range situacionList {
			placeholders = append(placeholders, "$"+string(rune(argCounter+48)))
			args = append(args, strings.TrimSpace(sit))
			argCounter++
		}
		whereClauses = append(whereClauses, "r.situacion_recibo IN ("+strings.Join(placeholders, ",")+")")
	}

	// Filtro por fecha
	if fechaDesde != "" {
		whereClauses = append(whereClauses, "r.fecha_emision >= $"+string(rune(argCounter+48)))
		args = append(args, fechaDesde)
		argCounter++
	}
	if fechaHasta != "" {
		whereClauses = append(whereClauses, "r.fecha_emision <= $"+string(rune(argCounter+48)))
		args = append(args, fechaHasta)
		argCounter++
	}

	// Filtro por mediador
	if mediador != "" {
		whereClauses = append(whereClauses, "r.mediador = $"+string(rune(argCounter+48)))
		args = append(args, mediador)
		argCounter++
	}

	// Filtro por cliente
	if cliente != "" {
		whereClauses = append(whereClauses, "LOWER(r.nombre_cliente) LIKE LOWER($"+string(rune(argCounter+48))+")")
		args = append(args, "%"+cliente+"%")
		argCounter++
	}

	// Filtro por importe
	if importeMin > 0 {
		whereClauses = append(whereClauses, "r.prima_total >= $"+string(rune(argCounter+48)))
		args = append(args, importeMin)
		argCounter++
	}
	if importeMax > 0 {
		whereClauses = append(whereClauses, "r.prima_total <= $"+string(rune(argCounter+48)))
		args = append(args, importeMax)
		argCounter++
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// Validar ordenamiento
	validOrderBy := map[string]bool{
		"fecha_emision":    true,
		"fecha_situacion":  true,
		"prima_total":      true,
		"situacion_recibo": true,
		"numero_recibo":    true,
	}
	if !validOrderBy[ordenarPor] {
		ordenarPor = "fecha_emision"
	}
	if orden != "ASC" && orden != "DESC" {
		orden = "DESC"
	}

	// === 1. Obtener KPIs Generales ===
	var response RecibosKPIResponse

	kpiQuery := `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(prima_total), 0) as total_importe,
			COUNT(CASE WHEN situacion_recibo = 'Cobrado' THEN 1 END) as total_cobrados,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Cobrado' THEN prima_total ELSE 0 END), 0) as importe_cobrado,
			COUNT(CASE WHEN situacion_recibo = 'Pendiente' THEN 1 END) as total_pendientes,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Pendiente' THEN prima_total ELSE 0 END), 0) as importe_pendiente,
			COUNT(CASE WHEN situacion_recibo = 'Anulado' THEN 1 END) as total_anulados,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Anulado' THEN prima_total ELSE 0 END), 0) as importe_anulado,
			COUNT(CASE WHEN situacion_recibo = 'Retornado' THEN 1 END) as total_retornados,
			COALESCE(SUM(CASE WHEN situacion_recibo = 'Retornado' THEN prima_total ELSE 0 END), 0) as importe_retornado
		FROM recibos r
		` + whereClause

	err := db.PostgresDB.QueryRow(kpiQuery, args...).Scan(
		&response.TotalRecibos,
		&response.TotalImporte,
		&response.TotalCobrados,
		&response.ImporteCobrado,
		&response.TotalPendientes,
		&response.ImportePendiente,
		&response.TotalAnulados,
		&response.ImporteAnulado,
		&response.TotalRetornados,
		&response.ImporteRetornado,
	)
	if err != nil {
		log.Printf("Error obteniendo KPIs: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error obteniendo KPIs de recibos",
			"error":   err.Error(),
		})
	}

	// === 2. Obtener Recibos con Paginación ===
	offset := (pagina - 1) * limite

	recibosQuery := `
		SELECT
			r.id,
			r.numero_recibo,
			r.numero_poliza,
			r.id_account,
			r.nombre_cliente,
			r.ramo,
			r.mediador,
			r.prima_total,
			r.comision_bruta,
			r.comision_neta,
			r.situacion_recibo,
			r.detalle_recibo,
			r.gestion_cobro,
			r.forma_pago,
			r.fecha_emision,
			r.fecha_situacion,
			r.fecha_inicio_cobertura,
			r.fecha_fin_cobertura,
			COALESCE(c.nif, '') as nif,
			COALESCE(c.email_contacto, '') as email,
			COALESCE(c.telefono_contacto, '') as telefono,
			EXTRACT(DAY FROM (NOW() - r.fecha_emision)) as dias_desde_emision,
			EXTRACT(DAY FROM (NOW() - r.fecha_situacion)) as dias_desde_cambio
		FROM recibos r
		LEFT JOIN clientes c ON r.id_account = c.id_account
		` + whereClause + `
		ORDER BY r.` + ordenarPor + ` ` + orden + `
		LIMIT $` + string(rune(argCounter+48)) + ` OFFSET $` + string(rune(argCounter+49))

	args = append(args, limite, offset)

	rows, err := db.PostgresDB.Query(recibosQuery, args...)
	if err != nil {
		log.Printf("Error obteniendo recibos: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error obteniendo recibos",
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	response.Recibos = []ReciboCompleto{}
	for rows.Next() {
		var r ReciboCompleto
		var fechaEmision, fechaSituacion, fechaInicio, fechaFin sql.NullTime
		var comisionBruta, comisionNeta sql.NullFloat64
		var diasEmision, diasCambio sql.NullFloat64
		var detalleRecibo, gestionCobro, formaPago, ramo, mediador sql.NullString
		var numeroRecibo, numeroPoliza, idAccount, nombreCliente sql.NullString

		err := rows.Scan(
			&r.ID,
			&numeroRecibo,
			&numeroPoliza,
			&idAccount,
			&nombreCliente,
			&ramo,
			&mediador,
			&r.PrimaTotal,
			&comisionBruta,
			&comisionNeta,
			&r.SituacionRecibo,
			&detalleRecibo,
			&gestionCobro,
			&formaPago,
			&fechaEmision,
			&fechaSituacion,
			&fechaInicio,
			&fechaFin,
			&r.NIF,
			&r.Email,
			&r.Telefono,
			&diasEmision,
			&diasCambio,
		)
		if err != nil {
			log.Printf("Error escaneando recibo: %v", err)
			continue
		}

		// Asignar valores obligatorios con manejo de NULL
		if numeroRecibo.Valid {
			r.NumeroRecibo = numeroRecibo.String
		}
		if numeroPoliza.Valid {
			r.NumeroPoliza = numeroPoliza.String
		}
		if idAccount.Valid {
			r.IDAccount = idAccount.String
		}
		if nombreCliente.Valid {
			r.Cliente = nombreCliente.String
		}

		// Asignar valores opcionales
		if ramo.Valid {
			r.Ramo = ramo.String
		}
		if mediador.Valid {
			r.Mediador = mediador.String
		}
		if comisionBruta.Valid {
			r.ComisionBruta = comisionBruta.Float64
		}
		if comisionNeta.Valid {
			r.ComisionNeta = comisionNeta.Float64
		}
		if detalleRecibo.Valid {
			r.DetalleRecibo = detalleRecibo.String
		}
		if gestionCobro.Valid {
			r.GestionCobro = gestionCobro.String
		}
		if formaPago.Valid {
			r.FormaPago = formaPago.String
		}
		if fechaEmision.Valid {
			r.FechaEmision = fechaEmision.Time.Format("2006-01-02")
		}
		if fechaSituacion.Valid {
			r.FechaSituacion = fechaSituacion.Time.Format("2006-01-02")
		}
		if fechaInicio.Valid {
			r.FechaInicioCobertura = fechaInicio.Time.Format("2006-01-02")
		}
		if fechaFin.Valid {
			r.FechaFinCobertura = fechaFin.Time.Format("2006-01-02")
		}
		if diasEmision.Valid {
			r.DiasDesdeEmision = int(diasEmision.Float64)
		}
		if diasCambio.Valid {
			r.DiasDesdeUltimoCambio = int(diasCambio.Float64)
		}

		response.Recibos = append(response.Recibos, r)
	}

	// === 3. Calcular paginación ===
	response.TotalPaginas = (response.TotalRecibos + limite - 1) / limite
	response.PaginaActual = pagina
	response.RegistrosPorPagina = limite

	// === 4. Documentar filtros aplicados ===
	response.Filtros = FiltrosAplicados{
		Situacion:     situacionList,
		FechaDesde:    fechaDesde,
		FechaHasta:    fechaHasta,
		Mediador:      mediador,
		Cliente:       cliente,
		ImporteMinimo: importeMin,
		ImporteMaximo: importeMax,
		OrdenarPor:    ordenarPor,
		Orden:         orden,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
