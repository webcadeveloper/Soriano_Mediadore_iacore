# Soriano Mediadores - Sistema de Gestión

Monorepo unificado del sistema de gestión de Soriano Mediadores, incluyendo frontend Angular y backend Go.

## 🚀 Inicio Rápido

**Para sincronizar cambios con GitHub:**

```bash
./git-sync.sh "Descripción de tus cambios"
```

Ver [GIT_WORKFLOW.md](GIT_WORKFLOW.md) para documentación completa de Git.

## Estructura del Proyecto

```
soriano/
├── frontend/          # Aplicación Angular 18
│   ├── src/          # Código fuente
│   ├── public/       # Recursos estáticos
│   └── dist/         # Build de producción
├── backend/           # API Server en Go
│   ├── cmd/          # Puntos de entrada
│   ├── internal/     # Lógica del negocio
│   └── migrations/   # Migraciones de BD
└── logs/             # Logs de aplicación
```

## Frontend (Angular 18)

### Características
- ✅ Angular 18 con standalone components
- ✅ Autenticación Microsoft OAuth via backend
- ✅ Material Design
- ✅ Progressive Web App (PWA)
- ✅ MockInterceptor para desarrollo sin backend
- ✅ Lazy loading de módulos
- ✅ Accesibilidad (WCAG 2.1 AA)

### Setup Frontend

```bash
cd frontend
npm install
npm start
```

La aplicación estará disponible en `http://localhost:4200`

### Build Frontend

```bash
cd frontend
npm run build
```

Los archivos compilados estarán en `frontend/dist/`

Para más detalles, consulta [frontend/README.md](frontend/README.md)

## Backend (Go + PostgreSQL)

### Características
- ✅ API RESTful en Go
- ✅ Autenticación Microsoft OAuth 2.0
- ✅ PostgreSQL para datos principales
- ✅ MongoDB para logs y analytics
- ✅ Redis para cache y sesiones
- ✅ Integración con Groq AI
- ✅ Scraper automatizado
- ✅ Sistema de bots (cobranza, auditoría, siniestros)

### Tecnologías Backend
- **Framework**: Gin (HTTP)
- **Base de datos**: PostgreSQL 14+
- **Cache**: Redis
- **Analytics**: MongoDB
- **AI**: Groq API
- **Auth**: Microsoft Graph API

### Setup Backend

1. **Configurar variables de entorno**

Crea un archivo `backend/.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu_password
DB_NAME=soriano_mediadores

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DB=soriano_logs

# Microsoft OAuth
MICROSOFT_CLIENT_ID=tu_client_id
MICROSOFT_CLIENT_SECRET=tu_client_secret
MICROSOFT_TENANT_ID=tu_tenant_id
MICROSOFT_REDIRECT_URI=http://localhost:8080/auth/callback

# Groq AI
GROQ_API_KEY=tu_groq_api_key

# Server
PORT=8080
ENV=development
```

2. **Instalar dependencias**

```bash
cd backend
go mod download
```

3. **Ejecutar migraciones**

```bash
psql -U postgres -d soriano_mediadores -f migrations/001_initial_schema.sql
psql -U postgres -d soriano_mediadores -f migrations/002_add_indexes.sql
psql -U postgres -d soriano_mediadores -f migrations/003_add_bots.sql
psql -U postgres -d soriano_mediadores -f migrations/004_create_import_jobs.sql
```

4. **Compilar y ejecutar**

```bash
go build -o soriano-api ./cmd/server
./soriano-api
```

El servidor estará disponible en `http://localhost:8080`

### Endpoints Principales

- `GET /health` - Health check
- `GET /auth/login` - Iniciar sesión con Microsoft
- `GET /auth/callback` - Callback de Microsoft OAuth
- `GET /auth/me` - Obtener usuario autenticado
- `POST /auth/logout` - Cerrar sesión
- `GET /api/stats` - Estadísticas generales
- `GET /api/clientes` - Listar clientes
- `GET /api/recobros` - Listar recobros
- `GET /api/bots` - Listar bots activos
- `POST /api/import` - Importar datos CSV

## Desarrollo

### Requisitos
- Node.js 18+
- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- MongoDB 6+

### Desarrollo Local

1. **Terminal 1: Backend**
```bash
cd backend
go run ./cmd/server
```

2. **Terminal 2: Frontend**
```bash
cd frontend
npm start
```

3. **Acceder a la aplicación**
   - Frontend: http://localhost:4200
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/health

### MockInterceptor

El frontend incluye un `MockInterceptor` que detecta automáticamente si el backend está disponible:
- ✅ Si el backend responde → usa datos reales
- ✅ Si el backend no responde → usa datos mock

Esto permite desarrollar el frontend sin necesidad de tener el backend corriendo.

## Despliegue

### Docker

Cada componente tiene su propio Dockerfile:

**Frontend:**
```bash
cd frontend
docker build -t soriano-frontend .
docker run -p 80:80 soriano-frontend
```

**Backend:**
```bash
cd backend
docker build -t soriano-backend .
docker run -p 8080:8080 soriano-backend
```

### PM2 (Producción)

El proyecto incluye configuración PM2:

```bash
pm2 start ecosystem.config.js
pm2 save
pm2 startup
```

## Seguridad

- **NO commitear** archivos sensibles:
  - `backend/.env` (contiene secrets)
  - Binarios compilados
  - Archivos de configuración con credenciales

- **Archivos ignorados en Git:**
  - `backend/.env`
  - `backend/soriano-*` (binarios)
  - `frontend/node_modules/`
  - `frontend/dist/`
  - `logs/`

## Contribuir

1. Crear una rama para tu feature: `git checkout -b feature/nombre-feature`
2. Hacer commits descriptivos
3. Push a tu rama: `git push origin feature/nombre-feature`
4. Crear Pull Request

## Licencia

Propiedad de Soriano Mediadores. Todos los derechos reservados.

## Soporte

Para soporte técnico, contactar al equipo de desarrollo.

---

**Última actualización:** 2026-01-22
**Versión:** 1.0.0
**Repositorio:** https://github.com/webcadeveloper/Soriano_Mediadore_iacore
