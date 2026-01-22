package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// AuthManager gestiona la autenticación con el portal GCO
type AuthManager struct {
	Username   string
	Password   string
	BaseURL    string
	IsLoggedIn bool
	LastLogin  time.Time
}

// NewAuthManager crea un nuevo AuthManager con las credenciales de .env
func NewAuthManager() *AuthManager {
	return &AuthManager{
		Username: os.Getenv("GCO_USERNAME"),
		Password: os.Getenv("GCO_PASSWORD"),
		BaseURL:  os.Getenv("GCO_BASE_URL"),
	}
}

// Login realiza el login en el portal GCO
func (am *AuthManager) Login(ctx context.Context) error {
	log.Printf("[Auth] Iniciando login en %s con usuario %s", am.BaseURL, am.Username)

	// Selectores comunes para formularios de login
	// NOTA: Estos se ajustarán según el HTML real del portal GCO
	usernameSelectors := []string{
		`input[name="username"]`,
		`input[name="user"]`,
		`input[name="login"]`,
		`input[name="Usuario"]`,
		`input[id="username"]`,
		`input[id="user"]`,
		`input[id="txtUsuario"]`,
		`input[type="text"]:first-of-type`,
		`#username`,
		`#user`,
		`#login`,
	}

	passwordSelectors := []string{
		`input[name="password"]`,
		`input[name="pass"]`,
		`input[name="Password"]`,
		`input[name="Clave"]`,
		`input[id="password"]`,
		`input[id="pass"]`,
		`input[id="txtPassword"]`,
		`input[type="password"]`,
		`#password`,
		`#pass`,
	}

	submitSelectors := []string{
		`button[type="submit"]`,
		`input[type="submit"]`,
		`button[name="login"]`,
		`button[id="btnLogin"]`,
		`button[id="btnEntrar"]`,
		`.btn-login`,
		`.login-button`,
		`button:contains("Entrar")`,
		`button:contains("Acceder")`,
		`input[value="Entrar"]`,
		`input[value="Acceder"]`,
	}

	// Navegar al portal
	err := chromedp.Run(ctx,
		chromedp.Navigate(am.BaseURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return fmt.Errorf("error navegando a %s: %w", am.BaseURL, err)
	}

	log.Println("[Auth] Página cargada, buscando formulario de login...")

	// Intentar encontrar y rellenar el campo de usuario
	var usernameSelector string
	for _, sel := range usernameSelectors {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			usernameSelector = sel
			log.Printf("[Auth] Campo usuario encontrado: %s", sel)
			break
		}
	}

	if usernameSelector == "" {
		// Tomar screenshot para debug
		var screenshot []byte
		chromedp.Run(ctx, chromedp.CaptureScreenshot(&screenshot))
		if len(screenshot) > 0 {
			os.WriteFile("/opt/soriano/logs/scraper/login_page.png", screenshot, 0644)
			log.Println("[Auth] Screenshot guardado en /opt/soriano/logs/scraper/login_page.png")
		}
		return fmt.Errorf("no se encontró campo de usuario en la página")
	}

	// Encontrar campo de password
	var passwordSelector string
	for _, sel := range passwordSelectors {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			passwordSelector = sel
			log.Printf("[Auth] Campo password encontrado: %s", sel)
			break
		}
	}

	if passwordSelector == "" {
		return fmt.Errorf("no se encontró campo de password en la página")
	}

	// Encontrar botón de submit
	var submitSelector string
	for _, sel := range submitSelectors {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			submitSelector = sel
			log.Printf("[Auth] Botón submit encontrado: %s", sel)
			break
		}
	}

	if submitSelector == "" {
		return fmt.Errorf("no se encontró botón de submit en la página")
	}

	// Rellenar formulario y hacer login
	log.Println("[Auth] Rellenando formulario de login...")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.Clear(usernameSelector),
		chromedp.SendKeys(usernameSelector, am.Username),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Clear(passwordSelector),
		chromedp.SendKeys(passwordSelector, am.Password),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(submitSelector, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("error rellenando formulario: %w", err)
	}

	log.Println("[Auth] Formulario enviado, esperando respuesta...")

	// Esperar a que cargue el dashboard o detectar error de login
	err = chromedp.Run(ctx,
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("error esperando respuesta de login: %w", err)
	}

	// Verificar si el login fue exitoso
	// Buscar indicadores de login exitoso (menú, dashboard, nombre de usuario, etc.)
	dashboardSelectors := []string{
		`.dashboard`,
		`.main-content`,
		`.menu-principal`,
		`.sidebar`,
		`.nav-menu`,
		`#menu`,
		`.header-user`,
		`.user-info`,
		`nav`,
		`.layout-main`,
	}

	var loginSuccess bool
	for _, sel := range dashboardSelectors {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			loginSuccess = true
			log.Printf("[Auth] Dashboard detectado: %s", sel)
			break
		}
	}

	// Verificar si hay mensajes de error
	errorSelectors := []string{
		`.error`,
		`.alert-danger`,
		`.error-message`,
		`.login-error`,
		`#error`,
		`.validation-error`,
	}

	for _, sel := range errorSelectors {
		var errorText string
		err := chromedp.Run(ctx,
			chromedp.Text(sel, &errorText, chromedp.ByQuery, chromedp.AtLeast(0)),
		)
		if err == nil && errorText != "" {
			return fmt.Errorf("error de login: %s", errorText)
		}
	}

	// Si no detectamos dashboard pero tampoco error, verificar URL
	var currentURL string
	chromedp.Run(ctx, chromedp.Location(&currentURL))
	log.Printf("[Auth] URL actual: %s", currentURL)

	// Si la URL cambió del login, probablemente el login fue exitoso
	if currentURL != am.BaseURL && currentURL != am.BaseURL+"/" {
		loginSuccess = true
	}

	if !loginSuccess {
		// Tomar screenshot para debug
		var screenshot []byte
		chromedp.Run(ctx, chromedp.CaptureScreenshot(&screenshot))
		if len(screenshot) > 0 {
			os.WriteFile("/opt/soriano/logs/scraper/after_login.png", screenshot, 0644)
			log.Println("[Auth] Screenshot post-login guardado")
		}

		// Obtener HTML para debug
		var html string
		chromedp.Run(ctx, chromedp.OuterHTML("html", &html))
		if len(html) > 500 {
			html = html[:500] + "..."
		}
		log.Printf("[Auth] HTML parcial: %s", html)

		return fmt.Errorf("no se pudo verificar el login exitoso")
	}

	am.IsLoggedIn = true
	am.LastLogin = time.Now()
	log.Println("[Auth] Login exitoso!")

	return nil
}

// IsSessionValid verifica si la sesión sigue siendo válida
func (am *AuthManager) IsSessionValid() bool {
	if !am.IsLoggedIn {
		return false
	}

	// Sesión válida por 30 minutos
	return time.Since(am.LastLogin) < 30*time.Minute
}

// Logout cierra la sesión
func (am *AuthManager) Logout(ctx context.Context) error {
	log.Println("[Auth] Cerrando sesión...")

	// Intentar encontrar y hacer click en logout
	logoutSelectors := []string{
		`a[href*="logout"]`,
		`a[href*="salir"]`,
		`button:contains("Salir")`,
		`button:contains("Cerrar sesión")`,
		`.logout`,
		`#logout`,
	}

	for _, sel := range logoutSelectors {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists),
		)
		if err == nil && exists {
			chromedp.Run(ctx, chromedp.Click(sel, chromedp.ByQuery))
			break
		}
	}

	am.IsLoggedIn = false
	return nil
}
