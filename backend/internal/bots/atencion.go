package bots

import (
	"fmt"
	"soriano-mediadores/internal/ai"
	"soriano-mediadores/internal/db"
	"strings"
	"time"
)

// BotAtencion - Bot de Atención al Cliente
// Funciones: Consultar pólizas, recibos, siniestros, información general
type BotAtencion struct {
	ID   string
	Name string
}

func NewBotAtencion() *BotAtencion {
	return &BotAtencion{
		ID:   "bot_atencion",
		Name: "Asistente de Atención al Cliente",
	}
}

// ProcesarConsulta procesa una consulta del cliente
func (b *BotAtencion) ProcesarConsulta(sessionID string, mensaje string) (string, error) {
	// Guardar mensaje en MongoDB
	db.GuardarSesionBot(sessionID, b.ID, map[string]interface{}{
		"tipo":    "consulta",
		"mensaje": mensaje,
	})

	// PASO 1: Intentar respuesta desde fallback (más rápido)
	if respuesta, found := FindBestMatch(b.ID, mensaje); found {
		// Cachear en Redis
		db.CacheSet("bot_response:"+b.ID+":"+mensaje, respuesta, 24*time.Hour)
		return respuesta, nil
	}

	// PASO 2: Verificar cache de Redis
	var cachedResponse string
	cacheKey := "bot_response:" + b.ID + ":" + mensaje
	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &cachedResponse); err == nil {
			return cachedResponse, nil
		}
	}

	// PASO 3: Detectar tipo de consulta usando AI (solo si no hay match)
	systemPrompt := `Eres el asistente de atención al cliente de SORIANO MEDIADORES, correduría de seguros española
colaboradora exclusiva de GRUPO OCCIDENT (Catalana Occidente, Plus Ultra Seguros, Seguros Bilbao, NorteHispana).

Analiza la consulta del cliente y clasifícala en una de estas categorías:
- BUSCAR_CLIENTE: buscar información de un cliente por nombre, NIF (DNI/NIE/CIF) o IdAccount
- CONSULTAR_POLIZAS: ver pólizas de un cliente (auto, hogar, vida, salud, accidentes, decesos, comercio, RC)
- CONSULTAR_RECIBOS: ver recibos, primas, pagos o estado de cobro de un cliente
- CONSULTAR_SINIESTROS: ver siniestros, partes o tramitaciones de un cliente
- INFORMACION_GENERAL: preguntas sobre coberturas, productos Occident, horarios, contacto, documentación

Responde SOLO con la categoría exacta, sin explicaciones adicionales.`

	categoria, err := ai.ConsultarAI(mensaje, systemPrompt)
	if err != nil {
		// Si AI falla, devolver respuesta por defecto
		return GetDefaultResponse(b.ID), nil
	}

	categoria = strings.TrimSpace(strings.ToUpper(categoria))

	// Procesar según categoría
	switch categoria {
	case "BUSCAR_CLIENTE":
		return b.BuscarCliente(mensaje)
	case "CONSULTAR_POLIZAS":
		return b.ConsultarPolizas(mensaje)
	case "CONSULTAR_RECIBOS":
		return b.ConsultarRecibos(mensaje)
	case "CONSULTAR_SINIESTROS":
		return b.ConsultarSiniestros(mensaje)
	default:
		return b.InformacionGeneral(mensaje)
	}
}

// BuscarCliente busca un cliente
func (b *BotAtencion) BuscarCliente(consulta string) (string, error) {
	// Intentar usar cache de Redis primero
	cacheKey := fmt.Sprintf("busqueda_cliente:%s", consulta)
	var clientesCache []db.Cliente

	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &clientesCache); err == nil {
			return b.FormatearResultadosClientes(clientesCache, true), nil
		}
	}

	// Extraer término de búsqueda
	termino := extraerTerminoBusqueda(consulta)

	// Buscar en PostgreSQL
	clientes, err := db.BuscarClientes(termino, 10)
	if err != nil {
		return "", fmt.Errorf("error buscando clientes: %w", err)
	}

	if len(clientes) == 0 {
		return "No se encontraron clientes con ese criterio de búsqueda.", nil
	}

	// Guardar en cache (5 minutos)
	db.CacheSet(cacheKey, clientes, 5*time.Minute)

	// Guardar búsqueda en MongoDB para analytics
	db.GuardarBusquedaCache(termino, clientes)

	return b.FormatearResultadosClientes(clientes, false), nil
}

// ConsultarPolizas consulta las pólizas de un cliente
func (b *BotAtencion) ConsultarPolizas(consulta string) (string, error) {
	// Extraer ID del cliente de la consulta
	idAccount := extraerIDCliente(consulta)
	if idAccount == "" {
		return "Por favor proporciona el ID del cliente o NIF para consultar sus pólizas.", nil
	}

	// Verificar cache
	cacheKey := fmt.Sprintf("polizas_cliente:%s", idAccount)
	var polizasCache []db.Poliza

	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &polizasCache); err == nil {
			return b.FormatearPolizas(polizasCache, true), nil
		}
	}

	// Obtener pólizas de PostgreSQL
	polizas, err := db.ObtenerPolizasCliente(idAccount)
	if err != nil {
		return "", fmt.Errorf("error obteniendo pólizas: %w", err)
	}

	if len(polizas) == 0 {
		return "No se encontraron pólizas para este cliente.", nil
	}

	// Cache por 10 minutos
	db.CacheSet(cacheKey, polizas, 10*time.Minute)

	return b.FormatearPolizas(polizas, false), nil
}

// ConsultarRecibos consulta los recibos de un cliente
func (b *BotAtencion) ConsultarRecibos(consulta string) (string, error) {
	idAccount := extraerIDCliente(consulta)
	if idAccount == "" {
		return "Por favor proporciona el ID del cliente para consultar sus recibos.", nil
	}

	cacheKey := fmt.Sprintf("recibos_cliente:%s", idAccount)
	var recibosCache []db.Recibo

	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &recibosCache); err == nil {
			return b.FormatearRecibos(recibosCache, true), nil
		}
	}

	recibos, err := db.ObtenerRecibosCliente(idAccount, 20)
	if err != nil {
		return "", fmt.Errorf("error obteniendo recibos: %w", err)
	}

	if len(recibos) == 0 {
		return "No se encontraron recibos para este cliente.", nil
	}

	db.CacheSet(cacheKey, recibos, 10*time.Minute)

	return b.FormatearRecibos(recibos, false), nil
}

// ConsultarSiniestros consulta los siniestros de un cliente
func (b *BotAtencion) ConsultarSiniestros(consulta string) (string, error) {
	idAccount := extraerIDCliente(consulta)
	if idAccount == "" {
		return "Por favor proporciona el ID del cliente para consultar sus siniestros.", nil
	}

	siniestros, err := db.ObtenerSiniestrosCliente(idAccount)
	if err != nil {
		return "", fmt.Errorf("error obteniendo siniestros: %w", err)
	}

	if len(siniestros) == 0 {
		return "No se encontraron siniestros para este cliente.", nil
	}

	return b.FormatearSiniestros(siniestros), nil
}

// InformacionGeneral responde preguntas generales
func (b *BotAtencion) InformacionGeneral(consulta string) (string, error) {
	systemPrompt := `Eres el asistente virtual de SORIANO MEDIADORES, correduría de seguros española con más de 30 años
de experiencia, colaboradora exclusiva de GRUPO OCCIDENT.

INFORMACIÓN DE LA EMPRESA:
- Nombre: Soriano Mediadores
- Lema: "Somos mediadores de seguros confiables"
- Filosofía: "Queremos ser parte de tu familia"
- Web: www.sorianomediadores.es

UBICACIÓN Y CONTACTO:
- Sede Principal: Calle Constitución 5, Villajoyosa, 03570 (Alicante)
- Teléfono: +34 96 681 02 90
- Email: info@sorianomediadores.es
- Horario: Lunes a Domingo de 09:00 a 17:00
- Oficinas adicionales en: Barcelona, Valladolid, Valencia
- Cobertura: Toda España

REDES SOCIALES:
- Facebook: Soriano Mediadores
- Instagram: @soriano_mediadores
- LinkedIn: Soriano Mediadores de Seguros

VALORES DE LA EMPRESA:
1. "Prometer es cumplir" - Trabajo meticuloso y atención al detalle
2. "Experiencia" - Más de 30 años de trayectoria en el sector asegurador
3. "La transparencia no se negocia" - Prácticas claras y honestas

SERVICIOS QUE OFRECEMOS:
1. SEGUROS - Coberturas personalizadas:
   - Vida
   - Hogar
   - Accidentes
   - Ahorro e inversión
   - Protección jurídica

2. TELECOM - Asesoría e instalación de servicios de telecomunicaciones

3. CONTRATOS ENERGÉTICOS - Gestión y negociación de contratos de energía

4. INMUEBLES - Servicios inmobiliarios:
   - Compra y venta de propiedades
   - Alquiler
   - Propiedades vacacionales

COMPAÑÍAS DEL GRUPO OCCIDENT QUE COMERCIALIZAMOS:
- Catalana Occidente: Líder en seguros multirramo
- Plus Ultra Seguros: Especialistas en auto y hogar
- Seguros Bilbao: Seguros de vida y ahorro
- NorteHispana: Seguros de salud y dental

PRODUCTOS DE SEGUROS:
- AUTOMÓVILES: Todo riesgo, terceros ampliado, terceros básico
- HOGAR: Continente, contenido, RC familiar, asistencia 24h
- VIDA Y AHORRO: Vida riesgo, PIAS, Unit Linked, planes de pensiones
- SALUD: Cuadro médico, reembolso, dental, copago
- ACCIDENTES: Individual, colectivo, convenio
- DECESOS: Familiar, individual, repatriación
- COMERCIO Y PYMES: Multirriesgo, RC profesional, D&O
- COMUNIDADES: Multirriesgo edificios, RC comunitaria

NORMATIVA ESPAÑOLA APLICABLE:
- Ley 50/1980 de Contrato de Seguro
- Ley de Distribución de Seguros (mediación)
- Período de reflexión: 14 días en seguros de vida

INSTRUCCIONES:
- Responde de forma profesional, cercana y en español de España
- Usa terminología española: "póliza" (no policy), "prima" (no premium), "siniestro" (no claim)
- Si preguntan por precios o presupuestos, indica que un agente les contactará
- Para urgencias fuera de horario: teléfono de asistencia 24h de la compañía
- Cuando pregunten datos de contacto, proporciona la información real de arriba

Si no conoces la respuesta exacta, sugiere contactar con la oficina al +34 96 681 02 90 o por email a info@sorianomediadores.es.`

	respuesta, err := ai.ConsultarAI(consulta, systemPrompt)
	if err != nil {
		return "Lo siento, no puedo procesar tu consulta en este momento. Por favor, contacta con nuestra oficina en horario de atención (L-V 9:00-14:00 y 16:00-19:00).", err
	}

	return respuesta, nil
}

// Funciones auxiliares de formateo
func (b *BotAtencion) FormatearResultadosClientes(clientes []db.Cliente, fromCache bool) string {
	var sb strings.Builder

	if fromCache {
		sb.WriteString("📋 [Desde caché]\n\n")
	}

	sb.WriteString(fmt.Sprintf("Encontrados %d cliente(s):\n\n", len(clientes)))

	for i, c := range clientes {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, c.NombreCompleto))
		sb.WriteString(fmt.Sprintf("   NIF: %s | ID: %s\n", c.NIF, c.IDAccount))
		if c.Email != "" {
			sb.WriteString(fmt.Sprintf("   Email: %s\n", c.Email))
		}
		if c.Telefono != "" {
			sb.WriteString(fmt.Sprintf("   Tel: %s\n", c.Telefono))
		}
		sb.WriteString(fmt.Sprintf("   Total Primas: €%.2f | Comisiones: €%.2f\n\n", c.TotalPrimas, c.TotalComisiones))
	}

	return sb.String()
}

func (b *BotAtencion) FormatearPolizas(polizas []db.Poliza, fromCache bool) string {
	var sb strings.Builder

	if fromCache {
		sb.WriteString("📋 [Desde caché]\n\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d póliza(s)\n\n", len(polizas)))

	for i, p := range polizas {
		sb.WriteString(fmt.Sprintf("%d. Póliza: %s\n", i+1, p.NumeroPoliza))
		sb.WriteString(fmt.Sprintf("   Ramo: %s\n", p.Ramo))
		sb.WriteString(fmt.Sprintf("   Situación: %s\n", p.Situacion))
		sb.WriteString(fmt.Sprintf("   Prima Anual: %s\n", p.PrimaAnual))
		if p.FechaEfecto != "" {
			sb.WriteString(fmt.Sprintf("   Vigencia: %s\n", p.FechaEfecto))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (b *BotAtencion) FormatearRecibos(recibos []db.Recibo, fromCache bool) string {
	var sb strings.Builder

	if fromCache {
		sb.WriteString("📋 [Desde caché]\n\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d recibo(s)\n\n", len(recibos)))

	var totalPrima float64
	for i, r := range recibos {
		sb.WriteString(fmt.Sprintf("%d. Recibo: %s\n", i+1, r.NumeroRecibo))
		sb.WriteString(fmt.Sprintf("   Situación: %s\n", r.Situacion))
		sb.WriteString(fmt.Sprintf("   Prima: €%.2f\n", r.PrimaTotal))
		if r.FechaEmision != "" {
			sb.WriteString(fmt.Sprintf("   Fecha: %s\n", r.FechaEmision))
		}
		sb.WriteString("\n")
		totalPrima += r.PrimaTotal
	}

	sb.WriteString(fmt.Sprintf("💰 Total: €%.2f\n", totalPrima))

	return sb.String()
}

func (b *BotAtencion) FormatearSiniestros(siniestros []db.Siniestro) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total: %d siniestro(s)\n\n", len(siniestros)))

	for i, s := range siniestros {
		sb.WriteString(fmt.Sprintf("%d. Siniestro: %s\n", i+1, s.NumeroSiniestro))
		sb.WriteString(fmt.Sprintf("   Situación: %s\n", s.Situacion))
		if s.FechaOcurrencia != "" {
			sb.WriteString(fmt.Sprintf("   Fecha: %s\n", s.FechaOcurrencia))
		}
		if s.Tramitador != "" {
			sb.WriteString(fmt.Sprintf("   Tramitador: %s\n", s.Tramitador))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Utilidades
func extraerTerminoBusqueda(consulta string) string {
	// Usar AI para extraer el término de búsqueda
	systemPrompt := `Eres un extractor de datos para SORIANO MEDIADORES (correduría de seguros española con Occident).
Extrae ÚNICAMENTE el término de búsqueda (nombre de persona/empresa, NIF/DNI/NIE/CIF, o IdAccount formato XXXXXXXX/XXX).
Responde SOLO con el término extraído, sin explicaciones ni texto adicional.`
	termino, err := ai.ConsultarAI(consulta, systemPrompt)
	if err != nil {
		// Fallback: usar la consulta completa
		return consulta
	}
	return strings.TrimSpace(termino)
}

func extraerIDCliente(consulta string) string {
	// Usar AI para extraer ID del cliente
	systemPrompt := `Eres un extractor de identificadores para SORIANO MEDIADORES (correduría de seguros española con Occident).
Extrae ÚNICAMENTE el identificador del cliente de la consulta:
- IdAccount de Occident: formato XXXXXXXX/XXX (ej: 20777103/000)
- NIF español: 8 dígitos + letra (ej: 12345678A)
- NIE: X/Y/Z + 7 dígitos + letra (ej: X1234567L)
- CIF empresa: letra + 8 dígitos (ej: B12345678)

Responde SOLO con el identificador encontrado, sin explicaciones. Si no encuentras ninguno, responde vacío.`
	id, err := ai.ConsultarAI(consulta, systemPrompt)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(id)
}
