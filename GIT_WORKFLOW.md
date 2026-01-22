# Guía de Git - Soriano Mediadores

Guía rápida para trabajar con el repositorio y mantenerlo sincronizado.

## Comandos Básicos Diarios

### 1. Ver Estado del Repositorio

```bash
cd /opt/soriano
git status
```

### 2. Ver Cambios Realizados

```bash
# Ver cambios en archivos modificados
git diff

# Ver cambios en un archivo específico
git diff frontend/src/app/pages/dashboard/dashboard.component.ts
git diff backend/internal/api/handlers.go
```

### 3. Agregar Cambios al Staging

```bash
# Agregar todos los cambios
git add .

# Agregar archivos específicos
git add frontend/src/app/pages/dashboard/
git add backend/internal/api/handlers.go

# Agregar por tipo
git add frontend/
git add backend/
```

### 4. Crear Commit

```bash
# Commit con mensaje descriptivo
git commit -m "Descripción clara de los cambios realizados"

# Ejemplos de buenos mensajes:
git commit -m "Add dashboard KPI widgets for client analytics"
git commit -m "Fix authentication redirect loop in auth.guard.ts"
git commit -m "Update bot response templates for cobranza"
git commit -m "Refactor API handlers to improve error handling"
```

### 5. Sincronizar con GitHub (Push)

```bash
# Push al repositorio remoto
git push origin master

# Si hay cambios en el remoto que no tienes localmente
git pull origin master
git push origin master
```

## Workflow Completo - Actualizar Repositorio

### Opción A: Workflow Simple (Recomendado)

```bash
cd /opt/soriano

# 1. Ver qué cambió
git status
git diff

# 2. Agregar todos los cambios
git add .

# 3. Crear commit
git commit -m "Descripción de los cambios"

# 4. Push a GitHub
git push origin master
```

### Opción B: Workflow Selectivo (Para cambios específicos)

```bash
cd /opt/soriano

# 1. Ver cambios
git status

# 2. Agregar solo archivos específicos
git add frontend/src/app/pages/dashboard/
git add backend/internal/api/handlers.go

# 3. Ver qué se va a commitear
git diff --staged

# 4. Crear commit
git commit -m "Update dashboard and API handlers"

# 5. Push
git push origin master
```

## Script de Sincronización Automática

Puedes usar este script para sincronizar rápidamente:

```bash
#!/bin/bash
# Archivo: /opt/soriano/git-sync.sh

cd /opt/soriano

echo "📊 Estado actual:"
git status

echo ""
echo "➕ Agregando cambios..."
git add .

echo ""
echo "📝 Creando commit..."
read -p "Mensaje del commit: " commit_msg
git commit -m "$commit_msg

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

echo ""
echo "🚀 Sincronizando con GitHub..."
git push origin master

echo ""
echo "✅ Sincronización completada"
git status
```

### Hacer el script ejecutable:

```bash
chmod +x /opt/soriano/git-sync.sh
```

### Usar el script:

```bash
cd /opt/soriano
./git-sync.sh
```

## Comandos Útiles

### Ver Historial de Commits

```bash
# Últimos 10 commits
git log --oneline -10

# Ver cambios de un commit específico
git show abc1234
```

### Ver Archivos Modificados

```bash
# Ver solo nombres de archivos modificados
git diff --name-only

# Ver archivos en el último commit
git diff --name-only HEAD~1
```

### Descartar Cambios

```bash
# Descartar cambios en un archivo específico
git restore frontend/src/app/pages/dashboard/dashboard.component.ts

# Descartar todos los cambios (¡CUIDADO!)
git restore .
```

### Sincronizar desde GitHub

```bash
# Descargar cambios del remoto
git fetch origin

# Ver si hay cambios nuevos
git status

# Traer y fusionar cambios
git pull origin master
```

## Buenas Prácticas

### ✅ Mensajes de Commit

**Buenos ejemplos:**
```
✅ Add export functionality to recobros page
✅ Fix authentication token expiration handling
✅ Update database schema for import jobs
✅ Refactor email templates for better readability
✅ Improve error messages in API responses
```

**Malos ejemplos:**
```
❌ fix
❌ cambios
❌ updates
❌ wip
❌ asdf
```

### ✅ Commits Frecuentes

- Haz commits pequeños y frecuentes
- Cada commit debe ser una unidad lógica de cambio
- No esperes a acumular muchos cambios

### ✅ Push Regular

- Haz push al menos una vez al día
- Haz push después de completar una funcionalidad
- Haz push antes de cambiar de máquina

### ❌ NO Commitear

Estos archivos están en `.gitignore` y NO deben commitearse:

- `backend/.env` (contiene secrets)
- `backend/soriano-*` (binarios compilados)
- `frontend/node_modules/`
- `frontend/dist/`
- `frontend/.angular/`
- `logs/`

## Seguridad

### Verificar antes de Push

```bash
# Ver qué archivos se van a pushear
git diff --staged --name-only

# Buscar posibles secrets
git diff --staged | grep -i "password\|secret\|key\|token"

# Si encuentras secrets, NO hagas push y elimínalos primero
```

### Si Accidentalmente Commiteas un Secret

```bash
# 1. Eliminar el secret del archivo
# 2. Crear un nuevo commit
git add .
git commit -m "Remove accidentally committed secrets"

# 3. Push (el nuevo commit reemplaza el anterior)
git push origin master
```

## Troubleshooting

### Error: "Your branch is behind"

```bash
# Primero traer cambios del remoto
git pull origin master

# Luego hacer push
git push origin master
```

### Error: "Push rejected - secrets detected"

```bash
# 1. Identificar qué archivo tiene secrets
# 2. Eliminar los secrets del archivo
# 3. Hacer un nuevo commit
git add .
git commit --amend --no-edit

# 4. Force push (solo si es necesario)
git push --force origin master
```

### Error: "Merge conflict"

```bash
# 1. Ver archivos en conflicto
git status

# 2. Editar archivos manualmente y resolver conflictos
# 3. Marcar como resueltos
git add .

# 4. Completar el merge
git commit -m "Resolve merge conflicts"
```

## Configuración de Git (Una sola vez)

```bash
# Configurar nombre y email
git config --global user.name "Tu Nombre"
git config --global user.email "tu@email.com"

# Verificar configuración
git config --list
```

## URLs Útiles

- **Repositorio**: https://github.com/webcadeveloper/Soriano_Mediadore_iacore
- **SSH Config**: ~/.ssh/config (configuración de clave SSH)
- **Git Docs**: https://git-scm.com/docs

## Resumen Rápido

```bash
# Workflow diario en 4 comandos:
cd /opt/soriano
git add .
git commit -m "Descripción de cambios"
git push origin master
```

---

**Última actualización:** 2026-01-22
**Repositorio:** git@github.com:webcadeveloper/Soriano_Mediadore_iacore.git
