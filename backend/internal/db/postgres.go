package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var PostgresDB *sql.DB

// InitPostgres conecta a PostgreSQL
func InitPostgres() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	var err error
	PostgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error abriendo conexión PostgreSQL: %w", err)
	}

	// Verificar conexión
	if err = PostgresDB.Ping(); err != nil {
		return fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	// Configurar pool de conexiones
	PostgresDB.SetMaxOpenConns(25)
	PostgresDB.SetMaxIdleConns(5)

	log.Println("✅ Conectado a PostgreSQL")
	return nil
}

// Cliente representa un cliente de la base de datos
type Cliente struct {
	ID             int     `json:"id"`
	NIF            string  `json:"nif"`
	IDAccount      string  `json:"id_account"`
	NombreCompleto string  `json:"nombre_completo"`
	Email          string  `json:"email,omitempty"`
	Telefono       string  `json:"telefono,omitempty"`
	Direccion      string  `json:"direccion,omitempty"`
	CodigoPostal   string  `json:"codigo_postal,omitempty"`
	Provincia      string  `json:"provincia,omitempty"`
	Pais           string  `json:"pais,omitempty"`
	TotalPrimas    float64 `json:"total_primas"`
	TotalComisiones float64 `json:"total_comisiones"`
}

// Poliza representa una póliza
type Poliza struct {
	ID               int    `json:"id"`
	NumeroPoliza     string `json:"numero_poliza"`
	IDAccount        string `json:"id_account"`
	NombreCliente    string `json:"nombre_cliente"`
	Ramo             string `json:"ramo"`
	Compania         string `json:"compania"`
	Producto         string `json:"producto"`
	Situacion        string `json:"situacion"`
	PrimaAnual       string `json:"prima_anual"`
	FechaEfecto      string `json:"fecha_efecto,omitempty"`
	FechaVencimiento string `json:"fecha_vencimiento,omitempty"`
}

// Recibo representa un recibo
type Recibo struct {
	ID             int     `json:"id"`
	NumeroRecibo   string  `json:"numero_recibo"`
	NumeroPoliza   string  `json:"numero_poliza"`
	IDAccount      string  `json:"id_account"`
	PrimaTotal     float64 `json:"prima_total"`
	Situacion      string  `json:"situacion_recibo"`
	FechaEmision   string  `json:"fecha_emision,omitempty"`
	FormaPago      string  `json:"forma_pago,omitempty"`
}

// Siniestro representa un siniestro
type Siniestro struct {
	ID              int    `json:"id"`
	NumeroSiniestro string `json:"numero_siniestro"`
	NumeroPoliza    string `json:"numero_poliza"`
	IDAccount       string `json:"id_account"`
	Situacion       string `json:"situacion_siniestro"`
	FechaOcurrencia string `json:"fecha_ocurrencia,omitempty"`
	FechaApertura   string `json:"fecha_apertura,omitempty"`
	Tramitador      string `json:"tramitador,omitempty"`
}

// BuscarClientes busca clientes por nombre, NIF o ID
func BuscarClientes(query string, limit int) ([]Cliente, error) {
	var sqlQuery string
	var rows *sql.Rows
	var err error

	// Si limit es 0, NO aplicar límite (retornar TODOS los clientes)
	if limit <= 0 {
		sqlQuery = `
			SELECT id, nif, id_account, nombre_completo, email_contacto, telefono_contacto,
			       domicilio, codigo_postal, provincia,
			       total_primas_cartera, total_primas_relacion
			FROM clientes
			WHERE activo = TRUE
			  AND (
			    LOWER(nombre_completo) LIKE LOWER($1)
			    OR LOWER(nif) LIKE LOWER($1)
			    OR LOWER(id_account) LIKE LOWER($1)
			  )
			ORDER BY nombre_completo
		`
		rows, err = PostgresDB.Query(sqlQuery, "%"+query+"%")
	} else {
		sqlQuery = `
			SELECT id, nif, id_account, nombre_completo, email_contacto, telefono_contacto,
			       domicilio, codigo_postal, provincia,
			       total_primas_cartera, total_primas_relacion
			FROM clientes
			WHERE activo = TRUE
			  AND (
			    LOWER(nombre_completo) LIKE LOWER($1)
			    OR LOWER(nif) LIKE LOWER($1)
			    OR LOWER(id_account) LIKE LOWER($1)
			  )
			ORDER BY nombre_completo
			LIMIT $2
		`
		rows, err = PostgresDB.Query(sqlQuery, "%"+query+"%", limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientes []Cliente
	for rows.Next() {
		var c Cliente
		var email, telefono, direccion, cp, provincia *string
		var totalPrimas, totalRelacion *float64

		err := rows.Scan(
			&c.ID, &c.NIF, &c.IDAccount, &c.NombreCompleto,
			&email, &telefono, &direccion, &cp, &provincia,
			&totalPrimas, &totalRelacion,
		)
		if err != nil {
			continue
		}

		if email != nil {
			c.Email = *email
		}
		if telefono != nil {
			c.Telefono = *telefono
		}
		if direccion != nil {
			c.Direccion = *direccion
		}
		if cp != nil {
			c.CodigoPostal = *cp
		}
		if provincia != nil {
			c.Provincia = *provincia
		}
		if totalPrimas != nil {
			c.TotalPrimas = *totalPrimas
		}
		if totalRelacion != nil {
			c.TotalComisiones = *totalRelacion
		}

		clientes = append(clientes, c)
	}

	return clientes, nil
}

// ObtenerClientePorID obtiene un cliente específico
func ObtenerClientePorID(idAccount string) (*Cliente, error) {
	sql := `
		SELECT id, nif, id_account, nombre_completo, email_contacto, telefono_contacto,
		       domicilio, codigo_postal, provincia,
		       total_primas_cartera, total_primas_relacion
		FROM clientes
		WHERE (id_account = $1 OR id::text = $1) AND activo = TRUE
	`

	var c Cliente
	var email, telefono, direccion, cp, provincia *string
	var totalPrimas, totalRelacion *float64

	err := PostgresDB.QueryRow(sql, idAccount).Scan(
		&c.ID, &c.NIF, &c.IDAccount, &c.NombreCompleto,
		&email, &telefono, &direccion, &cp, &provincia,
		&totalPrimas, &totalRelacion,
	)

	if err != nil {
		return nil, err
	}

	if email != nil {
		c.Email = *email
	}
	if telefono != nil {
		c.Telefono = *telefono
	}
	if direccion != nil {
		c.Direccion = *direccion
	}
	if cp != nil {
		c.CodigoPostal = *cp
	}
	if provincia != nil {
		c.Provincia = *provincia
	}
	if totalPrimas != nil {
		c.TotalPrimas = *totalPrimas
	}
	if totalRelacion != nil {
		c.TotalComisiones = *totalRelacion
	}

	return &c, nil
}

// ObtenerPolizasCliente obtiene las pólizas de un cliente
func ObtenerPolizasCliente(idAccount string) ([]Poliza, error) {
	sql := `
		SELECT p.id, p.numero_poliza, p.id_account, p.nombre_cliente, p.ramo,
		       p.gestora, p.situacion_poliza, p.prima_anual, p.fecha_efecto, p.fecha_vencimiento
		FROM polizas p
		WHERE p.id_account = $1 AND p.activo = TRUE
		ORDER BY p.fecha_efecto DESC
	`

	rows, err := PostgresDB.Query(sql, idAccount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polizas []Poliza
	for rows.Next() {
		var p Poliza
		var fechaEfecto *string
		var fechaVencimiento *string
		var primaAnual *string
		var gestora *string
		var ramo *string
		var situacion *string

		err := rows.Scan(
			&p.ID, &p.NumeroPoliza, &p.IDAccount, &p.NombreCliente,
			&ramo, &gestora, &situacion, &primaAnual, &fechaEfecto, &fechaVencimiento,
		)
		if err != nil {
			log.Printf("Error escaneando póliza: %v", err)
			continue
		}

		// Set values with defaults
		if ramo != nil {
			p.Ramo = *ramo
			p.Producto = *ramo
		}
		if gestora != nil {
			p.Compania = *gestora
		}
		if situacion != nil {
			p.Situacion = *situacion
		}
		if primaAnual != nil {
			p.PrimaAnual = *primaAnual
		}
		if fechaEfecto != nil {
			p.FechaEfecto = *fechaEfecto
		}
		if fechaVencimiento != nil {
			p.FechaVencimiento = *fechaVencimiento
		}

		polizas = append(polizas, p)
	}

	return polizas, nil
}

// ObtenerRecibosCliente obtiene los recibos de un cliente
func ObtenerRecibosCliente(idAccount string, limit int) ([]Recibo, error) {
	if limit <= 0 {
		limit = 50
	}

	sql := `
		SELECT id, numero_recibo, numero_poliza, id_account,
		       prima_total, situacion_recibo, fecha_emision, forma_pago
		FROM recibos
		WHERE id_account = $1 AND activo = TRUE
		ORDER BY fecha_emision DESC
		LIMIT $2
	`

	rows, err := PostgresDB.Query(sql, idAccount, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recibos []Recibo
	for rows.Next() {
		var r Recibo
		var numeroPoliza *string
		var fechaEmision *string
		var formaPago *string
		var primaTotal *float64

		err := rows.Scan(
			&r.ID, &r.NumeroRecibo, &numeroPoliza, &r.IDAccount,
			&primaTotal, &r.Situacion, &fechaEmision, &formaPago,
		)
		if err != nil {
			continue
		}

		if numeroPoliza != nil {
			r.NumeroPoliza = *numeroPoliza
		}
		if primaTotal != nil {
			r.PrimaTotal = *primaTotal
		}
		if fechaEmision != nil {
			r.FechaEmision = *fechaEmision
		}
		if formaPago != nil {
			r.FormaPago = *formaPago
		}

		recibos = append(recibos, r)
	}

	return recibos, nil
}

// ObtenerSiniestrosCliente obtiene los siniestros de un cliente
func ObtenerSiniestrosCliente(idAccount string) ([]Siniestro, error) {
	sql := `
		SELECT id, numero_siniestro, numero_poliza, id_account,
		       situacion_siniestro, fecha_ocurrencia, fecha_apertura, tramitador
		FROM siniestros
		WHERE id_account = $1 AND activo = TRUE
		ORDER BY fecha_ocurrencia DESC
	`

	rows, err := PostgresDB.Query(sql, idAccount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var siniestros []Siniestro
	for rows.Next() {
		var s Siniestro
		var numeroPoliza *string
		var fechaOcurrencia *string
		var fechaApertura *string
		var tramitador *string

		err := rows.Scan(
			&s.ID, &s.NumeroSiniestro, &numeroPoliza, &s.IDAccount,
			&s.Situacion, &fechaOcurrencia, &fechaApertura, &tramitador,
		)
		if err != nil {
			continue
		}

		if numeroPoliza != nil {
			s.NumeroPoliza = *numeroPoliza
		}
		if fechaOcurrencia != nil {
			s.FechaOcurrencia = *fechaOcurrencia
		}
		if fechaApertura != nil {
			s.FechaApertura = *fechaApertura
		}
		if tramitador != nil {
			s.Tramitador = *tramitador
		}

		siniestros = append(siniestros, s)
	}

	return siniestros, nil
}
