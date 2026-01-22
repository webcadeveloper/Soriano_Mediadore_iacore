#!/bin/bash
# Script de sincronización rápida con GitHub
# Uso: ./git-sync.sh "Mensaje del commit"

cd /opt/soriano

echo "================================================"
echo "🔄 Git Sync - Soriano Mediadores"
echo "================================================"
echo ""

# Verificar si estamos en un repositorio git
if [ ! -d .git ]; then
    echo "❌ Error: No estás en un repositorio Git"
    exit 1
fi

echo "📊 Estado actual del repositorio:"
echo "================================================"
git status
echo ""

# Verificar si hay cambios
if git diff-index --quiet HEAD --; then
    echo "✅ No hay cambios para commitear"
    echo ""
    echo "📡 Verificando sincronización con remoto..."
    git fetch origin

    if [ $(git rev-list HEAD...origin/master --count) -eq 0 ]; then
        echo "✅ Repositorio ya está sincronizado"
    else
        echo "⚠️  Hay cambios en el remoto. Ejecuta: git pull origin master"
    fi
    exit 0
fi

echo "➕ Agregando todos los cambios al staging..."
git add .

echo ""
echo "📝 Archivos que serán commiteados:"
echo "================================================"
git diff --staged --name-only
echo ""

# Verificar si se proporcionó mensaje como argumento
if [ -n "$1" ]; then
    commit_msg="$1"
else
    # Pedir mensaje de commit
    read -p "💬 Mensaje del commit: " commit_msg

    # Verificar que no esté vacío
    if [ -z "$commit_msg" ]; then
        echo "❌ Error: El mensaje del commit no puede estar vacío"
        exit 1
    fi
fi

echo ""
echo "📝 Creando commit..."
git commit -m "$commit_msg

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

if [ $? -ne 0 ]; then
    echo "❌ Error al crear el commit"
    exit 1
fi

echo ""
echo "🚀 Sincronizando con GitHub..."
echo "================================================"

# Intentar push
git push origin master

if [ $? -eq 0 ]; then
    echo ""
    echo "================================================"
    echo "✅ Sincronización completada exitosamente"
    echo "================================================"
    echo ""
    echo "📊 Estado final:"
    git log --oneline -3
    echo ""
    git status
else
    echo ""
    echo "================================================"
    echo "❌ Error al hacer push"
    echo "================================================"
    echo ""
    echo "Posibles causas:"
    echo "1. Hay cambios en el remoto que no tienes localmente"
    echo "   Solución: git pull origin master && git push origin master"
    echo ""
    echo "2. GitHub detectó secrets en el código"
    echo "   Solución: Revisar y eliminar secrets, luego volver a commitear"
    echo ""
    echo "3. Problemas de autenticación SSH"
    echo "   Solución: Verificar configuración SSH"
    exit 1
fi
