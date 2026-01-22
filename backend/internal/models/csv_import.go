package models

import "time"

// CSVReciboOccident representa un recibo del CSV de Occident
type CSVReciboOccident struct {
	NumRecibo           string    `json:"num_recibo"`
	Mediador            string    `json:"mediador"`
	Ramo                string    `json:"ramo"`
	Origen              string    `json:"origen"`
	PrimaTotal          string    `json:"prima_total"`
	FechaInicioCobertura time.Time `json:"fecha_inicio_cobertura"`
	SituacionRecibo     string    `json:"situacion_recibo"`
	FechaEmision        time.Time `json:"fecha_emision"`
	FechaSituacion      time.Time `json:"fecha_situacion"`
	FechaFinCobertura   time.Time `json:"fecha_fin_cobertura"`
	GestionCobro        string    `json:"gestion_cobro"`
	DetalleRecibo       string    `json:"detalle_recibo"`
	Cliente             string    `json:"cliente"`
	NumPoliza           string    `json:"num_poliza"`
	FormaPago           string    `json:"forma_pago"`
	DescripcionRiesgo   string    `json:"descripcion_riesgo"`
	IdAccount           string    `json:"id_account"`
}

// ImportCSVRequest es la petición de importación
type ImportCSVRequest struct {
	ArchivoNombre string `json:"archivo_nombre"`
	ArchivoBase64 string `json:"archivo_base64"` // CSV en base64
}

// ImportCSVResponse es la respuesta de la importación
type ImportCSVResponse struct {
	Success         bool                  `json:"success"`
	Message         string                `json:"message"`
	TotalLeidos     int                   `json:"total_leidos"`
	RecibosDevueltos int                  `json:"recibos_devueltos"`
	CasosCreados    int                   `json:"casos_creados"`
	Duplicados      int                   `json:"duplicados"`
	Errores         int                   `json:"errores"`
	Detalles        []ImportDetailItem    `json:"detalles,omitempty"`
	RecibosImportados []ReciboImportado    `json:"recibos_importados,omitempty"`
}

// ImportDetailItem detalla cada registro procesado
type ImportDetailItem struct {
	Linea       int    `json:"linea"`
	NumRecibo   string `json:"num_recibo"`
	Cliente     string `json:"cliente"`
	Estado      string `json:"estado"` // "creado", "duplicado", "error"
	Mensaje     string `json:"mensaje,omitempty"`
}

// ReciboImportado representa un recibo que fue importado y debe crear un caso de recobro
type ReciboImportado struct {
	NumRecibo       string  `json:"num_recibo"`
	Cliente         string  `json:"cliente"`
	NumPoliza       string  `json:"num_poliza"`
	Importe         float64 `json:"importe"`
	FechaVencimiento time.Time `json:"fecha_vencimiento"`
	Motivo          string  `json:"motivo"` // R01, R02, etc.
	Estado          string  `json:"estado"` // Retornado, Devuelto, etc.
	DiasVencido     int     `json:"dias_vencido"`
	Email           string  `json:"email,omitempty"`
	Telefono        string  `json:"telefono,omitempty"`
	NIF             string  `json:"nif,omitempty"`
}
