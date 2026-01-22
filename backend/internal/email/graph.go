package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GraphClient - Cliente para Microsoft Graph API
type GraphClient struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	AccessToken  string
	TokenExpiry  time.Time
	HTTPClient   *http.Client
}

// TokenResponse - Respuesta del endpoint de autenticación
type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// EmailMessage - Estructura del mensaje de email
type EmailMessage struct {
	Message         MessageContent `json:"message"`
	SaveToSentItems bool           `json:"saveToSentItems"`
}

// MessageContent - Contenido del mensaje
type MessageContent struct {
	Subject      string             `json:"subject"`
	Body         MessageBody        `json:"body"`
	ToRecipients []EmailRecipient   `json:"toRecipients"`
	CcRecipients []EmailRecipient   `json:"ccRecipients,omitempty"`
	From         *EmailRecipient    `json:"from,omitempty"`
}

// MessageBody - Cuerpo del mensaje
type MessageBody struct {
	ContentType string `json:"contentType"` // "Text" o "HTML"
	Content     string `json:"content"`
}

// EmailRecipient - Destinatario del email
type EmailRecipient struct {
	EmailAddress EmailAddress `json:"emailAddress"`
}

// EmailAddress - Dirección de email
type EmailAddress struct {
	Address string `json:"address"`
}

// NewGraphClient - Crea un nuevo cliente de Microsoft Graph
func NewGraphClient() (*GraphClient, error) {
	tenantID := os.Getenv("MS_TENANT_ID")
	clientID := os.Getenv("MS_CLIENT_ID")
	clientSecret := os.Getenv("MS_CLIENT_SECRET")

	if tenantID == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("faltan credenciales de Microsoft Graph en variables de entorno")
	}

	return &GraphClient{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetAccessToken - Obtiene un access token de Microsoft Graph
func (gc *GraphClient) GetAccessToken() error {
	// Si ya tenemos un token válido, no pedir uno nuevo
	if gc.AccessToken != "" && time.Now().Before(gc.TokenExpiry) {
		return nil
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", gc.TenantID)

	formData := fmt.Sprintf(
		"client_id=%s&scope=https://graph.microsoft.com/.default&client_secret=%s&grant_type=client_credentials",
		gc.ClientID,
		gc.ClientSecret,
	)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(formData))
	if err != nil {
		return fmt.Errorf("error creando request de token: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error obteniendo token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error en autenticación (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("error decodificando respuesta de token: %w", err)
	}

	gc.AccessToken = tokenResp.AccessToken
	gc.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 5 min antes de expirar

	return nil
}

// SendEmail - Envía un email usando Microsoft Graph
func (gc *GraphClient) SendEmail(from, to, subject, htmlBody string) error {
	// Asegurar que tenemos un token válido
	if err := gc.GetAccessToken(); err != nil {
		return fmt.Errorf("error obteniendo access token: %w", err)
	}

	// Construir el mensaje
	message := EmailMessage{
		Message: MessageContent{
			Subject: subject,
			Body: MessageBody{
				ContentType: "HTML",
				Content:     htmlBody,
			},
			ToRecipients: []EmailRecipient{
				{
					EmailAddress: EmailAddress{
						Address: to,
					},
				},
			},
		},
		SaveToSentItems: true,
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %w", err)
	}

	// URL del endpoint de envío
	sendURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", from)

	// Crear request
	req, err := http.NewRequest("POST", sendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request de envío: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gc.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Enviar
	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando email: %w", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error en envío de email (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendEmailWithCC - Envía un email con copia (CC)
func (gc *GraphClient) SendEmailWithCC(from, to string, cc []string, subject, htmlBody string) error {
	// Asegurar que tenemos un token válido
	if err := gc.GetAccessToken(); err != nil {
		return fmt.Errorf("error obteniendo access token: %w", err)
	}

	// Construir destinatarios CC
	ccRecipients := make([]EmailRecipient, 0, len(cc))
	for _, ccAddr := range cc {
		ccRecipients = append(ccRecipients, EmailRecipient{
			EmailAddress: EmailAddress{
				Address: ccAddr,
			},
		})
	}

	// Construir el mensaje
	message := EmailMessage{
		Message: MessageContent{
			Subject: subject,
			Body: MessageBody{
				ContentType: "HTML",
				Content:     htmlBody,
			},
			ToRecipients: []EmailRecipient{
				{
					EmailAddress: EmailAddress{
						Address: to,
					},
				},
			},
			CcRecipients: ccRecipients,
		},
		SaveToSentItems: true,
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %w", err)
	}

	// URL del endpoint de envío
	sendURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", from)

	// Crear request
	req, err := http.NewRequest("POST", sendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request de envío: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gc.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Enviar
	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando email: %w", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error en envío de email (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendBulkEmails - Envía múltiples emails en lote
func (gc *GraphClient) SendBulkEmails(from string, emails []struct {
	To      string
	Subject string
	Body    string
}) ([]error, error) {
	// Asegurar que tenemos un token válido
	if err := gc.GetAccessToken(); err != nil {
		return nil, fmt.Errorf("error obteniendo access token: %w", err)
	}

	errors := make([]error, 0)

	for _, email := range emails {
		if err := gc.SendEmail(from, email.To, email.Subject, email.Body); err != nil {
			errors = append(errors, fmt.Errorf("error enviando a %s: %w", email.To, err))
		}
		// Pequeña pausa entre envíos para no saturar la API
		time.Sleep(100 * time.Millisecond)
	}

	if len(errors) > 0 {
		return errors, fmt.Errorf("se encontraron %d errores al enviar emails", len(errors))
	}

	return nil, nil
}
