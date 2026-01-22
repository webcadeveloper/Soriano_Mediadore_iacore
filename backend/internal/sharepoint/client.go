package sharepoint

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/sites"
)

// Client es el cliente de SharePoint que encapsula el SDK de Microsoft Graph
type Client struct {
	graphClient *msgraphsdk.GraphServiceClient
	siteID      string
	driveID     string
}

// SharePointConfig contiene la configuración para conectarse a SharePoint
type SharePointConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	SiteURL      string // Ej: "te966987723r1.sharepoint.com/sites/PROYECTOS"
}

// NewClient crea una nueva instancia del cliente de SharePoint
func NewClient(config *SharePointConfig) (*Client, error) {
	// Validar configuración
	if config.TenantID == "" || config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("faltan credenciales de Azure AD (TENANT_ID, CLIENT_ID, CLIENT_SECRET)")
	}

	// Crear credenciales usando Client Credentials Flow
	cred, err := azidentity.NewClientSecretCredential(
		config.TenantID,
		config.ClientID,
		config.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creando credenciales: %w", err)
	}

	// Crear cliente de Microsoft Graph
	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("error creando cliente Graph: %w", err)
	}

	client := &Client{
		graphClient: graphClient,
	}

	// Obtener Site ID desde la URL del sitio
	if config.SiteURL != "" {
		siteID, err := client.getSiteIDByURL(config.SiteURL)
		if err != nil {
			log.Printf("Advertencia: no se pudo obtener Site ID automáticamente: %v", err)
			log.Printf("Deberás configurar SHAREPOINT_SITE_ID manualmente")
		} else {
			client.siteID = siteID
			log.Printf("✅ Site ID obtenido: %s", siteID)

			// Obtener Drive ID (Documents library del sitio)
			driveID, err := client.getDefaultDriveID()
			if err != nil {
				log.Printf("Advertencia: no se pudo obtener Drive ID: %v", err)
			} else {
				client.driveID = driveID
				log.Printf("✅ Drive ID obtenido: %s", driveID)
			}
		}
	}

	return client, nil
}

// NewClientFromEnv crea un cliente de SharePoint usando variables de entorno
func NewClientFromEnv() (*Client, error) {
	config := &SharePointConfig{
		TenantID:     os.Getenv("AZURE_TENANT_ID"),
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		SiteURL:      os.Getenv("SHAREPOINT_SITE_URL"),
	}

	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	// Si hay Site ID y Drive ID configurados manualmente, usarlos
	if siteID := os.Getenv("SHAREPOINT_SITE_ID"); siteID != "" {
		client.siteID = siteID
	}
	if driveID := os.Getenv("SHAREPOINT_DRIVE_ID"); driveID != "" {
		client.driveID = driveID
	}

	return client, nil
}

// getSiteIDByURL obtiene el Site ID de SharePoint usando la URL del sitio
func (c *Client) getSiteIDByURL(siteURL string) (string, error) {
	ctx := context.Background()

	// Parsear URL para obtener hostname y path
	// Ej: "te966987723r1.sharepoint.com/sites/PROYECTOS" -> hostname: te966987723r1.sharepoint.com, path: /sites/PROYECTOS
	hostname, sitePath, err := parseSiteURL(siteURL)
	if err != nil {
		return "", fmt.Errorf("URL de sitio inválida: %w", err)
	}

	// Llamar a Graph API para obtener el sitio
	site, err := c.graphClient.Sites().ByHostnamePath(hostname, sitePath).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("error obteniendo sitio: %w", err)
	}

	if site.GetId() == nil {
		return "", fmt.Errorf("sitio no encontrado")
	}

	return *site.GetId(), nil
}

// getDefaultDriveID obtiene el Drive ID de la biblioteca de documentos predeterminada
func (c *Client) getDefaultDriveID() (string, error) {
	if c.siteID == "" {
		return "", fmt.Errorf("site ID no configurado")
	}

	ctx := context.Background()

	// Obtener el drive predeterminado del sitio
	drive, err := c.graphClient.Sites().BySiteId(c.siteID).Drive().Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("error obteniendo drive: %w", err)
	}

	if drive.GetId() == nil {
		return "", fmt.Errorf("drive no encontrado")
	}

	return *drive.GetId(), nil
}

// GetSiteID devuelve el Site ID configurado
func (c *Client) GetSiteID() string {
	return c.siteID
}

// GetDriveID devuelve el Drive ID configurado
func (c *Client) GetDriveID() string {
	return c.driveID
}

// SetSiteID configura manualmente el Site ID
func (c *Client) SetSiteID(siteID string) {
	c.siteID = siteID
}

// SetDriveID configura manualmente el Drive ID
func (c *Client) SetDriveID(driveID string) {
	c.driveID = driveID
}

// ListDrives lista todos los drives disponibles en el sitio
func (c *Client) ListDrives() ([]models.Driveable, error) {
	if c.siteID == "" {
		return nil, fmt.Errorf("site ID no configurado")
	}

	ctx := context.Background()

	result, err := c.graphClient.Sites().BySiteId(c.siteID).Drives().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error listando drives: %w", err)
	}

	return result.GetValue(), nil
}

// parseSiteURL parsea una URL de SharePoint y extrae hostname y path
func parseSiteURL(siteURL string) (hostname, sitePath string, err error) {
	// Soportar formatos:
	// 1. "hostname.sharepoint.com/sites/SITENAME"
	// 2. "https://hostname.sharepoint.com/sites/SITENAME"

	// Remover protocolo si existe
	if len(siteURL) > 8 && siteURL[:8] == "https://" {
		siteURL = siteURL[8:]
	} else if len(siteURL) > 7 && siteURL[:7] == "http://" {
		siteURL = siteURL[7:]
	}

	// Dividir en hostname y path
	var found bool
	for i, c := range siteURL {
		if c == '/' {
			hostname = siteURL[:i]
			sitePath = siteURL[i:]
			found = true
			break
		}
	}

	if !found {
		return "", "", fmt.Errorf("formato de URL inválido, se esperaba: hostname.sharepoint.com/sites/SITENAME")
	}

	return hostname, sitePath, nil
}

// TestConnection prueba la conexión con SharePoint
func (c *Client) TestConnection() error {
	if c.siteID == "" {
		return fmt.Errorf("site ID no configurado")
	}

	ctx := context.Background()

	// Intentar obtener información del sitio
	site, err := c.graphClient.Sites().BySiteId(c.siteID).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("error probando conexión: %w", err)
	}

	siteName := "Desconocido"
	if site.GetDisplayName() != nil {
		siteName = *site.GetDisplayName()
	}

	log.Printf("✅ Conexión exitosa con SharePoint")
	log.Printf("   Sitio: %s", siteName)
	log.Printf("   Site ID: %s", c.siteID)
	if c.driveID != "" {
		log.Printf("   Drive ID: %s", c.driveID)
	}

	return nil
}
