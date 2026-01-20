# Soriano Mediadores - Sistema CRM

Sistema CRM moderno y seguro para la gestión integral de mediadores de seguros, desarrollado con Angular 18.2.21 y Material Design.

## 🚀 Características Principales

### Seguridad
- ✅ Autenticación JWT con refresh tokens
- ✅ Cifrado end-to-end de datos sensibles
- ✅ Validación XSS y sanitización HTML
- ✅ Guards de autenticación y roles
- ✅ Interceptores HTTP seguros
- ✅ Almacenamiento cifrado (SecureStorageService)

### Accesibilidad (WCAG 2.1 Level AA)
- ✅ Skip navigation links
- ✅ ARIA labels completos
- ✅ Navegación por teclado
- ✅ Lectores de pantalla (NVDA, JAWS, VoiceOver)
- ✅ Anuncios contextuales
- ✅ Focus management

### Testing
- ✅ Suite completa de tests unitarios (165+ tests)
- ✅ Cobertura de servicios, guards e interceptors
- ✅ Tests de accesibilidad
- ✅ Jasmine/Karma configurado

### Arquitectura
- ✅ Lazy loading en todas las rutas
- ✅ Preloading selectivo inteligente
- ✅ Barrel exports para importaciones limpias
- ✅ Standalone components (Angular 18)
- ✅ Estructura modular escalable

### UI/UX
- ✅ Material Design 3
- ✅ Paleta de colores personalizada (rojo semioscuro #8b4049)
- ✅ Fondo blanco con grises optimizados
- ✅ Tipografía mejorada y legible
- ✅ Contraste WCAG AA+
- ✅ Animaciones suaves

## 📦 Tecnologías

- **Framework**: Angular 18.2.21
- **UI Library**: Angular Material 18
- **Lenguaje**: TypeScript 5.5
- **Testing**: Jasmine + Karma
- **Build**: Angular CLI + esbuild
- **Estilos**: SCSS + CSS Variables

## 🛠️ Instalación

```bash
# Instalar dependencias
npm install

# Servidor de desarrollo
npm start
# Aplicación disponible en http://localhost:4200

# Build de producción
npm run build

# Ejecutar tests
npm test
```

## 📁 Estructura del Proyecto

```
src/
├── app/
│   ├── core/                    # Módulo core (servicios, guards, interceptors)
│   │   ├── guards/             # Guards de autenticación y roles
│   │   ├── interceptors/       # HTTP interceptors
│   │   ├── services/           # Servicios singleton
│   │   └── strategies/         # Estrategias de preloading
│   ├── pages/                   # Componentes de páginas
│   │   ├── login/
│   │   ├── dashboard/
│   │   ├── clientes/
│   │   ├── recobros/
│   │   └── ...
│   ├── shared/                  # Componentes y utilidades compartidas
│   │   └── components/
│   ├── app.component.*          # Componente raíz
│   └── app.routes.ts            # Configuración de rutas
├── styles.scss                  # Estilos globales
├── theme.scss                   # Tema Material personalizado
└── environments/                # Variables de entorno
```

## 🔐 Credenciales Demo

El sistema incluye usuarios demo para testing:

| Usuario | Contraseña | Rol |
|---------|------------|-----|
| admin | admin123 | Administrador |
| agente | agente123 | Agente |
| supervisor | supervisor123 | Supervisor |
| director | director123 | Director |
| auditor | auditor123 | Auditor |

## 🎨 Sistema de Diseño

### Paleta de Colores

```scss
// Primario (Rojo Semioscuro)
--primary-color: #8b4049
--primary-light: #a8545e
--primary-dark: #6d323a

// Fondos
--background-color: #ffffff
--surface-color: #ffffff

// Grises
--gray-900: #2c2c2c  // Texto primario
--gray-600: #757575  // Texto secundario
--gray-300: #e0e0e0  // Bordes
--gray-100: #f5f5f5  // Fondos alternativos
```

### Tipografía

- **Font Family**: Roboto, "Helvetica Neue", sans-serif
- **Headlines**: 700-600 weight, 2.5rem a 1rem
- **Body**: 400 weight, 1rem y 0.875rem
- **Line Height**: 1.5 para body, 1.2 para headlines

## 🔒 Seguridad

### Autenticación
- JWT con expiración configurable
- Refresh tokens automáticos
- Logout seguro con limpieza de sesión
- Guards para protección de rutas

### Cifrado
- AES-256 para datos sensibles
- Almacenamiento cifrado en localStorage
- Sanitización de inputs
- Validación de archivos

### Prevención de Vulnerabilidades
- XSS protection
- CSRF tokens
- Validación server-side
- Sanitización HTML
- Input validation

## 📊 Performance

### Bundle Size
- **Initial**: ~796 KB (179 KB gzipped)
- **Lazy chunks**: 10-127 KB cada uno
- **Styles**: 97.78 KB (9.50 KB gzipped)

### Optimizaciones
- Lazy loading en todas las rutas
- Preloading selectivo inteligente
- Tree shaking automático
- Minificación y compresión
- OnPush change detection

## 🧪 Testing

```bash
# Ejecutar todos los tests
npm test

# Tests con cobertura
npm run test:coverage

# Tests en modo watch
npm run test:watch
```

### Cobertura
- **Servicios**: 9 archivos, 165+ tests
- **Guards**: 2 archivos, 30+ tests
- **Interceptors**: 2 archivos, 40+ tests
- **Total**: ~165 tests unitarios

## 📱 Responsive

- ✅ Desktop (1920px+)
- ✅ Laptop (1024px-1919px)
- ✅ Tablet (768px-1023px)
- ✅ Mobile (320px-767px)

## ♿ Accesibilidad

### Cumplimiento WCAG 2.1
- **Level AA** cumplido
- Contraste mínimo 4.5:1 para texto normal
- Contraste mínimo 3:1 para texto grande
- Navegación completa por teclado
- Skip links funcionales

### Herramientas Compatibles
- NVDA (Windows)
- JAWS (Windows)
- VoiceOver (macOS/iOS)
- TalkBack (Android)

## 🚀 Deployment

### Build de Producción

```bash
npm run build
# Output en: dist/soriano-mediadores-web/
```

## 📝 Scripts NPM

```bash
npm start          # Servidor de desarrollo
npm run build      # Build de producción
npm test           # Ejecutar tests
npm run lint       # Linter
```

## 📄 Licencia

Copyright © 2026 Soriano Mediadores de Seguros. Todos los derechos reservados.

---

Desarrollado con ❤️ por el equipo de Soriano Mediadores
