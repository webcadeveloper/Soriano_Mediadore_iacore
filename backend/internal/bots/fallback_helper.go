package bots

import (
	"soriano-mediadores/internal/db"
	"time"
)

// ProcesarConFallback es una función helper que intenta responder usando:
// 1. Respuestas pre-programadas (fallback)
// 2. Cache de Redis
// 3. AI (solo si los anteriores fallan)
func ProcesarConFallback(botID string, sessionID string, mensaje string, aiFunc func(string) (string, error)) (string, error) {
	// Guardar en MongoDB
	db.GuardarSesionBot(sessionID, botID, map[string]interface{}{
		"tipo":    "consulta",
		"mensaje": mensaje,
	})

	// PASO 1: Intentar respuesta desde fallback (instantáneo)
	if respuesta, found := FindBestMatch(botID, mensaje); found {
		// Cachear la respuesta para siguiente vez
		cacheKey := "bot_response:" + botID + ":" + mensaje
		db.CacheSet(cacheKey, respuesta, 24*time.Hour)
		return respuesta, nil
	}

	// PASO 2: Verificar cache de Redis (muy rápido)
	var cachedResponse string
	cacheKey := "bot_response:" + botID + ":" + mensaje
	if db.CacheExists(cacheKey) {
		if err := db.CacheGet(cacheKey, &cachedResponse); err == nil {
			return cachedResponse, nil
		}
	}

	// PASO 3: Intentar con AI (más lento, solo si no hay alternativa)
	if aiFunc != nil {
		respuesta, err := aiFunc(mensaje)
		if err == nil {
			// Cachear la respuesta de AI para futuras consultas similares
			db.CacheSet(cacheKey, respuesta, 24*time.Hour)
			return respuesta, nil
		}
	}

	// PASO 4: Si todo falla, respuesta genérica por defecto
	return GetDefaultResponse(botID), nil
}
