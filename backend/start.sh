#!/bin/bash
# Load environment variables from .env file
# Copy .env.example to .env and configure your credentials

# Database Configuration
export POSTGRES_HOST=${POSTGRES_HOST:-localhost}
export POSTGRES_PORT=${POSTGRES_PORT:-5432}
export POSTGRES_USER=${POSTGRES_USER:-soriano}
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-your_password}
export POSTGRES_DB=${POSTGRES_DB:-soriano_occident_db}

# Server Configuration
export SERVER_PORT=${SERVER_PORT:-8080}
export SERVER_HOST=${SERVER_HOST:-0.0.0.0}

# Microsoft OAuth Configuration
export MICROSOFT_CLIENT_ID=${MICROSOFT_CLIENT_ID:-your_client_id}
export MICROSOFT_CLIENT_SECRET=${MICROSOFT_CLIENT_SECRET:-your_client_secret}
export MICROSOFT_TENANT_ID=${MICROSOFT_TENANT_ID:-your_tenant_id}

# Legacy Microsoft variables (for compatibility)
export MS_CLIENT_ID=${MS_CLIENT_ID:-$MICROSOFT_CLIENT_ID}
export MS_CLIENT_SECRET=${MS_CLIENT_SECRET:-$MICROSOFT_CLIENT_SECRET}
export MS_TENANT_ID=${MS_TENANT_ID:-$MICROSOFT_TENANT_ID}

# Groq AI Configuration
export GROQ_API_KEY=${GROQ_API_KEY:-your_groq_api_key}
export GROQ_MODEL=${GROQ_MODEL:-llama-3.3-70b-versatile}

exec ./soriano-server
