package bots

import (
	"strings"
)

// PatternResponse representa un patrón de pregunta y su respuesta
type PatternResponse struct {
	Keywords []string // Palabras clave para detectar
	Response string   // Respuesta a dar
	Priority int      // Prioridad (mayor = más específico)
}

// FallbackResponses contiene respuestas pre-programadas por bot
var FallbackResponses = map[string][]PatternResponse{
	"bot_atencion": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes", "saludos"},
			Response: "¡Hola! Bienvenido a Soriano Mediadores. ¿En qué puedo ayudarte hoy? Puedo ayudarte con:\n- Información sobre tus pólizas\n- Consultar recibos\n- Estado de siniestros\n- Datos de contacto\n\n¿Qué necesitas?",
			Priority: 1,
		},
		{
			Keywords: []string{"poliza", "polizas", "seguro", "seguros"},
			Response: "📋 Para consultar tus pólizas, puedo ayudarte con:\n\n1️⃣ Ver todas tus pólizas activas\n2️⃣ Detalles de una póliza específica\n3️⃣ Fechas de vencimiento\n4️⃣ Primas y coberturas\n\nPor favor, indícame tu nombre o NIF para buscar tu información.",
			Priority: 2,
		},
		{
			Keywords: []string{"recibo", "recibos", "pago", "pagos", "cobro"},
			Response: "💰 Gestión de Recibos:\n\n✅ Consultar recibos pendientes\n✅ Ver historial de pagos\n✅ Información de vencimientos\n✅ Métodos de pago disponibles\n\n¿Qué información necesitas sobre tus recibos?",
			Priority: 2,
		},
		{
			Keywords: []string{"siniestro", "siniestros", "accidente", "daño", "incidente"},
			Response: "🚨 Información sobre Siniestros:\n\nPuedo ayudarte con:\n- Consultar estado de siniestros\n- Reportar un nuevo siniestro\n- Ver historial de siniestros\n- Documentación necesaria\n\n¿Qué necesitas saber?",
			Priority: 2,
		},
		{
			Keywords: []string{"contacto", "telefono", "email", "direccion", "oficina", "donde", "ubicacion"},
			Response: "📞 Datos de Contacto - Soriano Mediadores:\n\n🏢 Dirección: Calle Constitución 5, Villajoyosa, 03570 (Alicante)\n📱 Teléfono: +34 96 681 02 90\n📧 Email: info@sorianomediadores.es\n🌐 Web: www.sorianomediadores.es\n⏰ Horario: Lunes a Domingo, 09:00 - 17:00\n\n📍 También operamos en: Barcelona, Valladolid y Valencia\n\n¿Necesitas algo más?",
			Priority: 2,
		},
		{
			Keywords: []string{"quienes sois", "quien sois", "sobre nosotros", "que es soriano", "empresa", "quienes somos"},
			Response: "🏢 Sobre Soriano Mediadores:\n\nSomos mediadores de seguros confiables con más de 30 años de experiencia.\n\n✨ Nuestros valores:\n- Prometer es cumplir: Trabajo meticuloso y atención al detalle\n- Experiencia: Años de trayectoria en el sector asegurador\n- Transparencia: No se negocia, siempre claros y honestos\n\n💼 Nuestros servicios:\n- Seguros personalizados\n- Telecomunicaciones\n- Contratos energéticos\n- Inmuebles (compra, venta, alquiler)\n\n📍 Sede: Villajoyosa (Alicante)\n📍 Presencia: Barcelona, Valladolid, Valencia\n\n¡Queremos ser parte de tu familia!",
			Priority: 2,
		},
		{
			Keywords: []string{"servicios", "que ofreceis", "que haceis", "productos"},
			Response: "💼 Servicios de Soriano Mediadores:\n\n🛡️ SEGUROS - Coberturas personalizadas:\n   • Vida\n   • Hogar\n   • Accidentes\n   • Ahorro e inversión\n   • Protección jurídica\n\n📡 TELECOM - Asesoría e instalación de telecomunicaciones\n\n⚡ ENERGÍA - Gestión y negociación de contratos energéticos\n\n🏠 INMUEBLES - Servicios inmobiliarios:\n   • Compra y venta\n   • Alquiler\n   • Propiedades vacacionales\n\n¿Te interesa algún servicio en particular?",
			Priority: 2,
		},
		{
			Keywords: []string{"redes sociales", "facebook", "instagram", "linkedin", "social"},
			Response: "📱 Redes Sociales de Soriano Mediadores:\n\n📘 Facebook: Soriano Mediadores\n📸 Instagram: @soriano_mediadores\n💼 LinkedIn: Soriano Mediadores de Seguros\n\n¡Síguenos para estar al día de nuestras novedades y promociones!",
			Priority: 2,
		},
		{
			Keywords: []string{"ayuda", "help", "que puedes hacer", "funciones"},
			Response: "🤖 Soy tu asistente virtual de Soriano Mediadores.\n\nPuedo ayudarte con:\n\n📋 Pólizas - Consultar y gestionar tus seguros\n💰 Recibos - Ver pagos y vencimientos\n🚨 Siniestros - Estado y tramitación\n👤 Datos - Actualizar información personal\n📞 Contacto - Información de la mediadora\n\n¿En qué te puedo ayudar?",
			Priority: 1,
		},
		{
			Keywords: []string{"gracias", "vale", "ok", "perfecto", "bien"},
			Response: "¡De nada! Si necesitas algo más, estoy aquí para ayudarte. 😊",
			Priority: 1,
		},
		{
			Keywords: []string{"adios", "hasta luego", "chao", "bye"},
			Response: "¡Hasta pronto! Que tengas un buen día. Si necesitas ayuda, no dudes en volver. 👋",
			Priority: 1,
		},
	},
	"bot_cobranza": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes"},
			Response: "👋 Hola, soy el Bot de Gestión de Cobranza de Soriano Mediadores.\n\nPuedo ayudarte con:\n- Recibos pendientes de pago\n- Historial de cobros\n- Métodos de pago\n- Domiciliación bancaria\n\n¿Qué necesitas?",
			Priority: 1,
		},
		{
			Keywords: []string{"pendiente", "pendientes", "debo", "pagar"},
			Response: "💳 Recibos Pendientes:\n\nPara consultar tus recibos pendientes, necesito tu:\n- Nombre completo o\n- NIF\n\nCon esta información puedo mostrarte:\n✅ Recibos pendientes de pago\n✅ Fechas de vencimiento\n✅ Importes\n✅ Métodos de pago disponibles",
			Priority: 2,
		},
		{
			Keywords: []string{"domiciliar", "domiciliacion", "cuenta", "banco"},
			Response: "🏦 Domiciliación Bancaria:\n\nVentajas de domiciliar tus recibos:\n✅ Pago automático - No olvides ningún recibo\n✅ Sin comisiones adicionales\n✅ Puedes cancelar cuando quieras\n\nPara domiciliar, necesitamos:\n- Número de cuenta (IBAN)\n- Titular de la cuenta\n- Pólizas a domiciliar\n\n¿Quieres proceder?",
			Priority: 2,
		},
		{
			Keywords: []string{"pago", "como pagar", "metodo", "forma pago"},
			Response: "💰 Métodos de Pago Disponibles:\n\n1️⃣ Domiciliación bancaria (Recomendado)\n2️⃣ Transferencia bancaria\n3️⃣ Pago en oficina (efectivo/tarjeta)\n4️⃣ Bizum (para importes menores)\n\n¿Cuál prefieres usar?",
			Priority: 2,
		},
	},
	"bot_siniestros": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes"},
			Response: "👋 Hola, soy el Bot de Gestión de Siniestros.\n\nPuedo ayudarte con:\n🚨 Reportar un nuevo siniestro\n📊 Consultar estado de siniestros\n📄 Documentación necesaria\n⏱️ Tiempos de tramitación\n\n¿En qué puedo ayudarte?",
			Priority: 1,
		},
		{
			Keywords: []string{"reportar", "nuevo", "accidente", "daño"},
			Response: "🚨 Reportar Nuevo Siniestro:\n\nPara abrir un parte de siniestro necesitamos:\n\n1️⃣ Datos del asegurado (nombre, NIF, póliza)\n2️⃣ Fecha y hora del incidente\n3️⃣ Descripción detallada de lo ocurrido\n4️⃣ Fotografías del daño (si es posible)\n5️⃣ Parte de accidente (si hay terceros)\n\n¿Deseas proceder con el reporte?",
			Priority: 2,
		},
		{
			Keywords: []string{"estado", "consultar", "como va", "tramite"},
			Response: "📊 Consultar Estado de Siniestro:\n\nPara consultar el estado necesito:\n- Número de siniestro o\n- Número de póliza + Fecha aproximada\n\nPuedo informarte sobre:\n✅ Estado actual del expediente\n✅ Peritaje realizado\n✅ Documentación pendiente\n✅ Tiempo estimado de resolución",
			Priority: 2,
		},
		{
			Keywords: []string{"documentos", "documentacion", "papeles", "que necesito"},
			Response: "📄 Documentación para Siniestros:\n\nSegún tipo de siniestro:\n\n🚗 Auto: Parte amistoso, fotos, DNI, permiso circulación\n🏠 Hogar: Fotos daños, presupuesto reparación, facturas\n💼 Vida/Salud: Informes médicos, facturas, recetas\n⚖️ RC: Reclamación tercero, documentación incidente\n\n¿Qué tipo de siniestro es?",
			Priority: 2,
		},
	},
	"bot_agente": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes"},
			Response: "👋 ¡Hola! Soy tu Agente Comercial Virtual de Soriano Mediadores.\n\n¿Te interesa?\n🏠 Seguro de Hogar\n🚗 Seguro de Auto\n👨‍👩‍👧 Seguro de Vida\n💼 Seguro de Negocio\n🏥 Seguro de Salud\n\n¡Cuéntame qué necesitas!",
			Priority: 1,
		},
		{
			Keywords: []string{"presupuesto", "cotizar", "precio", "cuanto cuesta"},
			Response: "💰 Solicitar Presupuesto:\n\nPara prepararte un presupuesto personalizado necesito:\n\n1️⃣ Tipo de seguro que te interesa\n2️⃣ Datos básicos (edad, profesión, etc.)\n3️⃣ Coberturas deseadas\n4️⃣ Datos del bien a asegurar (coche, vivienda, etc.)\n\n¿Qué tipo de seguro te interesa?",
			Priority: 2,
		},
		{
			Keywords: []string{"contratar", "quiero", "nuevo seguro"},
			Response: "✅ Contratar Nuevo Seguro:\n\n¡Excelente decisión! Para ayudarte mejor:\n\n1️⃣ ¿Qué tipo de seguro necesitas?\n2️⃣ ¿Tienes alguna póliza actualmente?\n3️⃣ ¿Cuándo necesitas que comience la cobertura?\n\nUn agente comercial se pondrá en contacto contigo para finalizar la contratación. 📞",
			Priority: 2,
		},
		{
			Keywords: []string{"comparar", "comparativa", "diferencias", "cual mejor"},
			Response: "🔍 Comparativa de Seguros:\n\nPuedo ayudarte a comparar:\n- Coberturas incluidas\n- Precios y franquicias\n- Compañías aseguradoras\n- Servicios adicionales\n\n¿Qué productos quieres comparar?",
			Priority: 2,
		},
	},
	"bot_analista": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes"},
			Response: "👋 Hola, soy el Bot Analista de Datos.\n\nPuedo proporcionarte:\n📊 Estadísticas generales\n📈 Análisis de cartera\n🏆 Rankings y top clientes\n📉 Tendencias temporales\n💰 Análisis financiero\n\n¿Qué análisis necesitas?",
			Priority: 1,
		},
		{
			Keywords: []string{"estadisticas", "stats", "numeros", "datos"},
			Response: "📊 Estadísticas del Sistema:\n\nTenemos datos actualizados sobre:\n- Total de clientes y pólizas\n- Distribución por ramos\n- Primas totales\n- Siniestralidad\n- Ratios de cobranza\n\n¿Qué estadística específica te interesa?",
			Priority: 2,
		},
		{
			Keywords: []string{"top", "mejores", "ranking", "principales"},
			Response: "🏆 Análisis Top Clientes:\n\nPuedo mostrarte:\n- Top 20 clientes por primas\n- Clientes con más pólizas\n- Mejores ramos\n- Productos más vendidos\n\n¿Qué ranking quieres ver?",
			Priority: 2,
		},
	},
	"bot_auditor": {
		{
			Keywords: []string{"hola", "buenos dias", "buenas tardes"},
			Response: "👋 Hola, soy el Bot Auditor de Calidad.\n\nRealizo auditorías sobre:\n🔍 Calidad de datos\n✅ Integridad referencial\n📋 Cumplimiento normativo\n⚠️ Detección de anomalías\n\n¿Qué auditoría necesitas?",
			Priority: 1,
		},
		{
			Keywords: []string{"auditoria", "revision", "calidad", "datos"},
			Response: "🔍 Auditoría de Calidad:\n\nPuedo revisar:\n✅ Completitud de datos de clientes\n✅ Integridad referencial (pólizas-recibos-siniestros)\n✅ Datos duplicados\n✅ Inconsistencias en fechas\n✅ Validación de NIFs\n\n¿Qué aspecto quieres auditar?",
			Priority: 2,
		},
		{
			Keywords: []string{"errores", "problemas", "inconsistencias"},
			Response: "⚠️ Detección de Problemas:\n\nPuedo identificar:\n- Recibos sin póliza asociada\n- Siniestros huérfanos\n- Clientes sin email/teléfono\n- Pólizas vencidas sin renovación\n- Datos incompletos\n\n¿Qué tipo de problema buscas?",
			Priority: 2,
		},
	},
}

// FindBestMatch busca la mejor respuesta basada en el mensaje
func FindBestMatch(botID string, mensaje string) (string, bool) {
	responses, exists := FallbackResponses[botID]
	if !exists {
		return "", false
	}

	mensajeLower := strings.ToLower(mensaje)

	var bestMatch *PatternResponse
	maxScore := 0

	for i := range responses {
		score := 0
		for _, keyword := range responses[i].Keywords {
			if strings.Contains(mensajeLower, keyword) {
				score += responses[i].Priority
			}
		}

		if score > maxScore {
			maxScore = score
			bestMatch = &responses[i]
		}
	}

	if bestMatch != nil && maxScore > 0 {
		return bestMatch.Response, true
	}

	return "", false
}

// GetDefaultResponse devuelve una respuesta genérica por bot
func GetDefaultResponse(botID string) string {
	defaults := map[string]string{
		"bot_atencion": "Lo siento, no entendí bien tu pregunta. ¿Puedes reformularla? Puedo ayudarte con:\n- Información de pólizas\n- Consulta de recibos\n- Estado de siniestros\n- Contacto y datos",
		"bot_cobranza": "No entendí tu consulta sobre cobranza. Puedo ayudarte con:\n- Recibos pendientes\n- Métodos de pago\n- Domiciliación bancaria\n- Historial de cobros",
		"bot_siniestros": "No comprendí tu consulta sobre siniestros. Puedo ayudarte con:\n- Reportar nuevo siniestro\n- Consultar estado\n- Documentación necesaria",
		"bot_agente": "No entendí bien. Puedo ayudarte con:\n- Nuevas pólizas y presupuestos\n- Comparativas de seguros\n- Productos disponibles",
		"bot_analista": "No entendí tu consulta. Puedo proporcionarte:\n- Estadísticas generales\n- Análisis de cartera\n- Rankings y comparativas",
		"bot_auditor": "No comprendí. Puedo realizar:\n- Auditorías de calidad\n- Revisión de integridad\n- Detección de inconsistencias",
	}

	if response, exists := defaults[botID]; exists {
		return response
	}

	return "Lo siento, no entendí tu pregunta. ¿Puedes ser más específico?"
}
