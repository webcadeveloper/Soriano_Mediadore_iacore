package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TokenResponse representa la respuesta de Microsoft OAuth
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope"`
}

// UserInfo representa la información del usuario de Microsoft
type UserInfo struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
	JobTitle          string `json:"jobTitle,omitempty"`
}

// Session representa una sesión de usuario autenticado
type Session struct {
	UserInfo    UserInfo
	AccessToken string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

var (
	// Almacén de sesiones en memoria
	sessions     = make(map[string]*Session)
	sessionMutex sync.RWMutex

	// Almacén de estados OAuth (para prevenir CSRF)
	oauthStates     = make(map[string]time.Time)
	oauthStateMutex sync.RWMutex

	// Configuración
	clientID     string
	clientSecret string
	tenantID     string
	redirectURI  string
	baseURL      string
)

// InitAuth inicializa la configuración de autenticación
func InitAuth() {
	clientID = os.Getenv("MS_CLIENT_ID")
	if clientID == "" {
		clientID = os.Getenv("MICROSOFT_CLIENT_ID")
	}

	clientSecret = os.Getenv("MS_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = os.Getenv("MICROSOFT_CLIENT_SECRET")
	}

	tenantID = os.Getenv("MS_TENANT_ID")
	if tenantID == "" {
		tenantID = os.Getenv("MICROSOFT_TENANT_ID")
	}

	// URL base de la aplicación (para construir redirect URI)
	baseURL = os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
		baseURL = "http://localhost:" + port
	}
	redirectURI = baseURL + "/auth/callback"

	log.Printf("🔐 Auth inicializado - Redirect URI: %s", redirectURI)

	// Limpiar estados OAuth expirados periódicamente
	go cleanupOAuthStates()
	// Limpiar sesiones expiradas periódicamente
	go cleanupExpiredSessions()
}

// generateState genera un estado aleatorio para OAuth
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	oauthStateMutex.Lock()
	oauthStates[state] = time.Now().Add(10 * time.Minute)
	oauthStateMutex.Unlock()

	return state, nil
}

// validateState valida el estado OAuth
func validateState(state string) bool {
	oauthStateMutex.Lock()
	defer oauthStateMutex.Unlock()

	expiry, exists := oauthStates[state]
	if !exists {
		return false
	}

	delete(oauthStates, state)
	return time.Now().Before(expiry)
}

// cleanupOAuthStates limpia estados OAuth expirados
func cleanupOAuthStates() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		oauthStateMutex.Lock()
		now := time.Now()
		for state, expiry := range oauthStates {
			if now.After(expiry) {
				delete(oauthStates, state)
			}
		}
		oauthStateMutex.Unlock()
	}
}

// cleanupExpiredSessions limpia sesiones expiradas
func cleanupExpiredSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		sessionMutex.Lock()
		now := time.Now()
		for sessionID, session := range sessions {
			if now.After(session.ExpiresAt) {
				delete(sessions, sessionID)
				log.Printf("🔐 Sesión expirada eliminada: %s", sessionID[:8])
			}
		}
		sessionMutex.Unlock()
	}
}

// generateSessionID genera un ID de sesión aleatorio
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL devuelve la URL de autenticación de Microsoft
func GetAuthURL() (string, error) {
	state, err := generateState()
	if err != nil {
		return "", err
	}

	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"response_mode": {"query"},
		"scope":         {"openid profile email User.Read"},
		"state":         {state},
	}

	authURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?%s",
		tenantID,
		params.Encode(),
	)

	return authURL, nil
}

// ExchangeCodeForToken intercambia el código de autorización por tokens
func ExchangeCodeForToken(code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
		"scope":         {"openid profile email User.Read"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error obteniendo token: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// GetUserInfo obtiene la información del usuario desde Microsoft Graph
func GetUserInfo(accessToken string) (*UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error obteniendo usuario: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// CreateSession crea una nueva sesión para el usuario
func CreateSession(userInfo *UserInfo, accessToken string, expiresIn int) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	session := &Session{
		UserInfo:    *userInfo,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
		CreatedAt:   time.Now(),
	}

	sessionMutex.Lock()
	sessions[sessionID] = session
	sessionMutex.Unlock()

	log.Printf("🔐 Nueva sesión creada para: %s (%s)", userInfo.DisplayName, userInfo.Mail)

	return sessionID, nil
}

// GetSession obtiene una sesión por su ID
func GetSession(sessionID string) (*Session, error) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := sessions[sessionID]
	if !exists {
		return nil, errors.New("sesión no encontrada")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("sesión expirada")
	}

	return session, nil
}

// DeleteSession elimina una sesión
func DeleteSession(sessionID string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessions, sessionID)
}

// ValidateStateHandler valida el estado OAuth en el callback
func ValidateStateHandler(state string) bool {
	return validateState(state)
}

// ========== HANDLERS HTTP ==========

// LoginHandler redirige al usuario a Microsoft para autenticación
func LoginHandler(c *fiber.Ctx) error {
	authURL, err := GetAuthURL()
	if err != nil {
		log.Printf("❌ Error generando URL de auth: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Error iniciando autenticación",
		})
	}

	return c.Redirect(authURL, 302)
}

// CallbackHandler maneja el callback de Microsoft OAuth
func CallbackHandler(c *fiber.Ctx) error {
	// Verificar errores de Microsoft
	if errMsg := c.Query("error"); errMsg != "" {
		errDesc := c.Query("error_description")
		log.Printf("❌ Error de Microsoft OAuth: %s - %s", errMsg, errDesc)
		return c.Redirect("/login?error=" + url.QueryEscape(errDesc), 302)
	}

	// Validar estado (prevenir CSRF)
	state := c.Query("state")
	if !ValidateStateHandler(state) {
		log.Printf("❌ Estado OAuth inválido")
		return c.Redirect("/login?error=invalid_state", 302)
	}

	// Obtener código de autorización
	code := c.Query("code")
	if code == "" {
		return c.Redirect("/login?error=no_code", 302)
	}

	// Intercambiar código por token
	tokenResp, err := ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("❌ Error intercambiando código: %v", err)
		return c.Redirect("/login?error=token_exchange_failed", 302)
	}

	// Obtener información del usuario
	userInfo, err := GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("❌ Error obteniendo usuario: %v", err)
		return c.Redirect("/login?error=user_info_failed", 302)
	}

	// Crear sesión
	sessionID, err := CreateSession(userInfo, tokenResp.AccessToken, tokenResp.ExpiresIn)
	if err != nil {
		log.Printf("❌ Error creando sesión: %v", err)
		return c.Redirect("/login?error=session_failed", 302)
	}

	// Establecer cookie de sesión
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		HTTPOnly: true,
		Secure:   strings.HasPrefix(baseURL, "https"),
		SameSite: "Lax",
		Path:     "/",
	})

	log.Printf("✅ Login exitoso: %s (%s)", userInfo.DisplayName, userInfo.Mail)

	// Redirigir a la aplicación principal
	return c.Redirect("/", 302)
}

// LogoutHandler cierra la sesión del usuario
func LogoutHandler(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID != "" {
		DeleteSession(sessionID)
	}

	// Eliminar cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})

	// Redirigir a Microsoft logout y luego a login
	logoutURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/logout?post_logout_redirect_uri=%s",
		tenantID,
		url.QueryEscape(baseURL+"/login"),
	)

	return c.Redirect(logoutURL, 302)
}

// MeHandler devuelve información del usuario actual
func MeHandler(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(401).JSON(fiber.Map{
			"error":         "No autenticado",
			"authenticated": false,
		})
	}

	session, err := GetSession(sessionID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":         err.Error(),
			"authenticated": false,
		})
	}

	return c.JSON(fiber.Map{
		"authenticated": true,
		"user":          session.UserInfo,
		"expiresAt":     session.ExpiresAt,
	})
}

// AuthMiddleware middleware que protege las rutas
func AuthMiddleware(c *fiber.Ctx) error {
	// Rutas públicas que no requieren autenticación
	path := c.Path()
	publicPaths := []string{
		"/login",
		"/auth/login",
		"/auth/callback",
		"/auth/logout",
		"/health",
		"/favicon.ico",
	}

	for _, p := range publicPaths {
		if path == p || strings.HasPrefix(path, "/auth/") {
			return c.Next()
		}
	}

	// Permitir archivos estáticos del login
	if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".ico") {
		return c.Next()
	}

	// Verificar sesión
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		// Si es una petición AJAX/API, devolver 401
		if c.Get("Accept") == "application/json" || strings.HasPrefix(path, "/api/") {
			return c.Status(401).JSON(fiber.Map{
				"error":         "No autenticado",
				"authenticated": false,
			})
		}
		// Si es una petición de navegador, redirigir a login
		return c.Redirect("/login", 302)
	}

	session, err := GetSession(sessionID)
	if err != nil {
		// Sesión inválida o expirada
		c.Cookie(&fiber.Cookie{
			Name:     "session_id",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HTTPOnly: true,
			Path:     "/",
		})

		if c.Get("Accept") == "application/json" || strings.HasPrefix(path, "/api/") {
			return c.Status(401).JSON(fiber.Map{
				"error":         "Sesión expirada",
				"authenticated": false,
			})
		}
		return c.Redirect("/login", 302)
	}

	// Añadir información del usuario al contexto
	c.Locals("user", session.UserInfo)
	c.Locals("session", session)

	return c.Next()
}
