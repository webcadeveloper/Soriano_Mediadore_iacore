package api

import (
	"database/sql"
	"log"
	"soriano-mediadores/internal/db"

	"github.com/gofiber/fiber/v2"
)

// StatsResponse contiene estadísticas generales del sistema
type StatsResponse struct {
	TotalClientes        int                   `json:"total_clientes"`
	TotalPolizas         int                   `json:"total_polizas"`
	TotalRecibos         int                   `json:"total_recibos"`
	RecibosDevueltos     int                   `json:"recibos_devueltos"`
	ClientesConDeuda     int                   `json:"clientes_con_deuda"`
	DeudaTotal           float64               `json:"deuda_total"`
	RecibosRetornados    []ReciboDevuelto      `json:"recibos_retornados,omitempty"`
	ClientesDeudores     []ClienteConDeuda     `json:"clientes_deudores,omitempty"`
}

// ReciboDevuelto representa un recibo devuelto/retornado
type ReciboDevuelto struct {
	NumeroRecibo     string  `json:"numero_recibo"`
	NumeroPoliza     string  `json:"numero_poliza"`
	Cliente          string  `json:"cliente"`
	NIF              string  `json:"nif,omitempty"`
	Email            string  `json:"email,omitempty"`
	Telefono         string  `json:"telefono,omitempty"`
	Importe          float64 `json:"importe"`
	FechaEmision     string  `json:"fecha_emision,omitempty"`
	FechaSituacion   string  `json:"fecha_situacion,omitempty"`
	Situacion        string  `json:"situacion"`
	DetalleRecibo    string  `json:"detalle_recibo,omitempty"`
	GestionCobro     string  `json:"gestion_cobro,omitempty"`
	DiasVencido      int     `json:"dias_vencido"`
	IDAccount        string  `json:"id_account"`
}

// ClienteConDeuda representa un cliente con recibos pendientes
type ClienteConDeuda struct {
	NIF              string  `json:"nif"`
	IDAccount        string  `json:"id_account"`
	NombreCompleto   string  `json:"nombre_completo"`
	Email            string  `json:"email,omitempty"`
	Telefono         string  `json:"telefono,omitempty"`
	Provincia        string  `json:"provincia,omitempty"`
	TotalRecibos     int     `json:"total_recibos_devueltos"`
	DeudaTotal       float64 `json:"deuda_total"`
	UltimaDevolucion string  `json:"ultima_devolucion,omitempty"`
}

// GetStats obtiene estadísticas generales del sistema
func GetStats(c *fiber.Ctx) error {
	var stats StatsResponse

	// Total de clientes activos
	err := db.PostgresDB.QueryRow(`
		SELECT COUNT(*) FROM clientes WHERE activo = TRUE
	`).Scan(&stats.TotalClientes)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo total de clientes: %v", err)
		stats.TotalClientes = 0
	}

	// Total de pólizas activas
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(*) FROM polizas WHERE situacion_poliza = 'Vigor'
	`).Scan(&stats.TotalPolizas)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo total de pólizas: %v", err)
		stats.TotalPolizas = 0
	}

	// Total de recibos
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(*) FROM recibos WHERE activo = TRUE
	`).Scan(&stats.TotalRecibos)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo total de recibos: %v", err)
		stats.TotalRecibos = 0
	}

	// Recibos devueltos/retornados
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(*) FROM recibos
		WHERE activo = TRUE
		  AND (situacion_recibo = 'Retornado' OR detalle_recibo LIKE '%Devuelto%')
	`).Scan(&stats.RecibosDevueltos)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo recibos devueltos: %v", err)
		stats.RecibosDevueltos = 0
	}

	// Clientes con deuda (distintos)
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(DISTINCT id_account) FROM recibos
		WHERE activo = TRUE
		  AND id_account IS NOT NULL
		  AND (situacion_recibo = 'Retornado' OR detalle_recibo LIKE '%Devuelto%')
	`).Scan(&stats.ClientesConDeuda)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo clientes con deuda: %v", err)
		stats.ClientesConDeuda = 0
	}

	// Deuda total (suma de importes de recibos devueltos)
	err = db.PostgresDB.QueryRow(`
		SELECT COALESCE(SUM(prima_total), 0) FROM recibos
		WHERE activo = TRUE
		  AND (situacion_recibo = 'Retornado' OR detalle_recibo LIKE '%Devuelto%')
	`).Scan(&stats.DeudaTotal)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error obteniendo deuda total: %v", err)
		stats.DeudaTotal = 0
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetRecibosDevueltos obtiene lista de recibos devueltos con información del cliente
func GetRecibosDevueltos(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	query := `
		SELECT
			r.numero_recibo,
			r.numero_poliza,
			r.nombre_cliente,
			r.id_account,
			r.prima_total,
			r.fecha_emision,
			r.fecha_situacion,
			r.situacion_recibo,
			r.detalle_recibo,
			r.gestion_cobro,
			COALESCE(c.nif, '') as nif,
			COALESCE(c.email_contacto, '') as email,
			COALESCE(c.telefono_contacto, '') as telefono,
			EXTRACT(DAY FROM (NOW() - r.fecha_situacion)) as dias_vencido
		FROM recibos r
		LEFT JOIN clientes c ON r.id_account = c.id_account
		WHERE r.activo = TRUE
		  AND (r.situacion_recibo = 'Retornado' OR r.detalle_recibo LIKE '%Devuelto%')
		ORDER BY r.fecha_situacion DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.PostgresDB.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error obteniendo recibos devueltos: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error obteniendo recibos devueltos",
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var recibos []ReciboDevuelto
	for rows.Next() {
		var r ReciboDevuelto
		var fechaEmision, fechaSituacion sql.NullTime
		var diasVencido sql.NullFloat64

		err := rows.Scan(
			&r.NumeroRecibo,
			&r.NumeroPoliza,
			&r.Cliente,
			&r.IDAccount,
			&r.Importe,
			&fechaEmision,
			&fechaSituacion,
			&r.Situacion,
			&r.DetalleRecibo,
			&r.GestionCobro,
			&r.NIF,
			&r.Email,
			&r.Telefono,
			&diasVencido,
		)
		if err != nil {
			log.Printf("Error escaneando recibo: %v", err)
			continue
		}

		if fechaEmision.Valid {
			r.FechaEmision = fechaEmision.Time.Format("2006-01-02")
		}
		if fechaSituacion.Valid {
			r.FechaSituacion = fechaSituacion.Time.Format("2006-01-02")
		}
		if diasVencido.Valid {
			r.DiasVencido = int(diasVencido.Float64)
		}

		recibos = append(recibos, r)
	}

	// Contar total de recibos devueltos
	var total int
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(*) FROM recibos
		WHERE activo = TRUE
		  AND (situacion_recibo = 'Retornado' OR detalle_recibo LIKE '%Devuelto%')
	`).Scan(&total)
	if err != nil {
		total = len(recibos)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    recibos,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetClientesConDeuda obtiene lista de clientes con recibos pendientes
func GetClientesConDeuda(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	query := `
		SELECT
			c.nif,
			c.id_account,
			c.nombre_completo,
			COALESCE(c.email_contacto, '') as email,
			COALESCE(c.telefono_contacto, '') as telefono,
			COALESCE(c.provincia, '') as provincia,
			COUNT(r.id) as total_recibos,
			SUM(r.prima_total) as deuda_total,
			MAX(r.fecha_situacion) as ultima_devolucion
		FROM clientes c
		INNER JOIN recibos r ON c.id_account = r.id_account
		WHERE c.activo = TRUE
		  AND r.activo = TRUE
		  AND (r.situacion_recibo = 'Retornado' OR r.detalle_recibo LIKE '%Devuelto%')
		GROUP BY c.nif, c.id_account, c.nombre_completo, c.email_contacto, c.telefono_contacto, c.provincia
		ORDER BY deuda_total DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.PostgresDB.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error obteniendo clientes con deuda: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Error obteniendo clientes con deuda",
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var clientes []ClienteConDeuda
	for rows.Next() {
		var cliente ClienteConDeuda
		var ultimaDevolucion sql.NullTime

		err := rows.Scan(
			&cliente.NIF,
			&cliente.IDAccount,
			&cliente.NombreCompleto,
			&cliente.Email,
			&cliente.Telefono,
			&cliente.Provincia,
			&cliente.TotalRecibos,
			&cliente.DeudaTotal,
			&ultimaDevolucion,
		)
		if err != nil {
			log.Printf("Error escaneando cliente: %v", err)
			continue
		}

		if ultimaDevolucion.Valid {
			cliente.UltimaDevolucion = ultimaDevolucion.Time.Format("2006-01-02")
		}

		clientes = append(clientes, cliente)
	}

	// Contar total de clientes con deuda
	var total int
	err = db.PostgresDB.QueryRow(`
		SELECT COUNT(DISTINCT c.id_account) FROM clientes c
		INNER JOIN recibos r ON c.id_account = r.id_account
		WHERE c.activo = TRUE
		  AND r.activo = TRUE
		  AND (r.situacion_recibo = 'Retornado' OR r.detalle_recibo LIKE '%Devuelto%')
	`).Scan(&total)
	if err != nil {
		total = len(clientes)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    clientes,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
