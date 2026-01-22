import { Injectable } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

/**
 * Servicio de seguridad para sanitización, cifrado y validación
 */
@Injectable({
  providedIn: 'root'
})
export class SecurityService {
  // Clave de cifrado (en producción debería venir del backend)
  private readonly ENCRYPTION_KEY = 'SorianoMediadores2024SecureKey!';

  constructor(private sanitizer: DomSanitizer) {}

  /**
   * Sanitiza HTML para prevenir XSS
   * @param html - HTML sin sanitizar
   * @returns HTML sanitizado
   */
  sanitizeHtml(html: string): SafeHtml {
    return this.sanitizer.sanitize(1, html) || '';
  }

  /**
   * Escapa caracteres HTML para prevenir inyección
   * @param text - Texto a escapar
   * @returns Texto escapado
   */
  escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  /**
   * Valida que un HTML no contenga scripts maliciosos
   * @param html - HTML a validar
   * @returns Objeto con resultado de validación
   */
  validateHtml(html: string): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!html || html.trim().length === 0) {
      errors.push('El contenido HTML no puede estar vacío');
      return { valid: false, errors };
    }

    // Validar scripts (case-insensitive)
    if (/<script[\s>]/gi.test(html)) {
      errors.push('No se permiten etiquetas <script>');
    }

    // Validar iframes
    if (/<iframe[\s>]/gi.test(html)) {
      errors.push('No se permiten etiquetas <iframe>');
    }

    // Validar event handlers peligrosos
    const dangerousHandlers = [
      'onload', 'onerror', 'onclick', 'onmouseover', 'onmouseout',
      'onkeydown', 'onkeyup', 'onfocus', 'onblur', 'onchange',
      'onsubmit', 'ondblclick', 'oncontextmenu'
    ];

    dangerousHandlers.forEach(handler => {
      const regex = new RegExp(`${handler}\\s*=`, 'gi');
      if (regex.test(html)) {
        errors.push(`No se permite el event handler: ${handler}`);
      }
    });

    // Validar imports CSS maliciosos
    if (/@import/gi.test(html)) {
      errors.push('No se permiten @import en estilos');
    }

    // Validar expresiones javascript: en atributos
    if (/javascript:/gi.test(html)) {
      errors.push('No se permiten URLs javascript:');
    }

    // Validar data: URIs que puedan ejecutar código
    if (/data:text\/html/gi.test(html)) {
      errors.push('No se permiten data URIs de HTML');
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Cifra datos sensibles usando AES-256 (simulado con Base64 + XOR)
   * NOTA: Para producción, usar una librería como CryptoJS
   * @param data - Datos a cifrar
   * @returns Datos cifrados en Base64
   */
  encrypt(data: string): string {
    try {
      // Cifrado simple XOR + Base64 (para demo)
      // En producción usar crypto-js con AES-256
      const encrypted = this.xorEncrypt(data, this.ENCRYPTION_KEY);
      return btoa(encrypted);
    } catch (error) {
      console.error('Error cifrando datos:', error);
      return '';
    }
  }

  /**
   * Descifra datos cifrados
   * @param encryptedData - Datos cifrados en Base64
   * @returns Datos descifrados
   */
  decrypt(encryptedData: string): string {
    try {
      if (!encryptedData) return '';
      const decoded = atob(encryptedData);
      return this.xorEncrypt(decoded, this.ENCRYPTION_KEY);
    } catch (error) {
      console.error('Error descifrando datos:', error);
      return '';
    }
  }

  /**
   * Cifrado XOR simple (para demo, usar AES en producción)
   */
  private xorEncrypt(text: string, key: string): string {
    let result = '';
    for (let i = 0; i < text.length; i++) {
      result += String.fromCharCode(
        text.charCodeAt(i) ^ key.charCodeAt(i % key.length)
      );
    }
    return result;
  }

  /**
   * Valida que una URL sea HTTPS
   * @param url - URL a validar
   * @returns true si es HTTPS o vacía
   */
  isHttps(url?: string): boolean {
    if (!url) return false;
    try {
      return new URL(url).protocol === 'https:';
    } catch {
      return false;
    }
  }

  /**
   * Valida un IBAN usando el algoritmo mod-97
   * @param iban - IBAN a validar
   * @returns true si el IBAN es válido
   */
  validateIBAN(iban?: string): boolean {
    if (!iban) return false;

    // Eliminar espacios y convertir a mayúsculas
    const cleanIban = iban.replace(/\s+/g, '').toUpperCase();

    // Validar formato básico (2 letras + 2 dígitos + hasta 30 caracteres alfanuméricos)
    if (!/^([A-Z]{2}\d{2}[A-Z0-9]{1,30})$/.test(cleanIban)) {
      return false;
    }

    // Mover los primeros 4 caracteres al final
    const reordered = cleanIban.slice(4) + cleanIban.slice(0, 4);

    // Convertir letras a números (A=10, B=11, ..., Z=35)
    const numericIban = reordered.replace(/[A-Z]/g, (char) =>
      (char.charCodeAt(0) - 55).toString()
    );

    // Calcular módulo 97
    let remainder = 0;
    for (let i = 0; i < numericIban.length; i += 7) {
      const part = remainder.toString() + numericIban.substring(i, i + 7);
      remainder = parseInt(part, 10) % 97;
    }

    return remainder === 1;
  }

  /**
   * Genera un hash simple de un string (para comparaciones, no para seguridad)
   * @param text - Texto a hashear
   * @returns Hash numérico
   */
  simpleHash(text: string): number {
    let hash = 0;
    for (let i = 0; i < text.length; i++) {
      const char = text.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash);
  }

  /**
   * Sanitiza un email para prevenir inyección
   * @param email - Email a sanitizar
   * @returns Email sanitizado o vacío si inválido
   */
  sanitizeEmail(email: string): string {
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    const trimmed = email.trim().toLowerCase();
    return emailRegex.test(trimmed) ? trimmed : '';
  }

  /**
   * Sanitiza un teléfono eliminando caracteres peligrosos
   * @param phone - Teléfono a sanitizar
   * @returns Teléfono sanitizado
   */
  sanitizePhone(phone: string): string {
    // Permitir solo dígitos, +, espacios, guiones y paréntesis
    return phone.replace(/[^0-9+\s\-()]/g, '');
  }

  /**
   * Valida que un archivo sea del tipo esperado
   * @param file - Archivo a validar
   * @param allowedTypes - Tipos MIME permitidos
   * @param maxSizeMB - Tamaño máximo en MB
   * @returns Objeto con resultado de validación
   */
  validateFile(
    file: File,
    allowedTypes: string[],
    maxSizeMB: number = 10
  ): { valid: boolean; error?: string } {
    // Validar tamaño
    const maxSizeBytes = maxSizeMB * 1024 * 1024;
    if (file.size > maxSizeBytes) {
      return {
        valid: false,
        error: `El archivo supera el tamaño máximo de ${maxSizeMB}MB`
      };
    }

    // Validar tipo MIME
    if (!allowedTypes.includes(file.type)) {
      return {
        valid: false,
        error: `Tipo de archivo no permitido. Permitidos: ${allowedTypes.join(', ')}`
      };
    }

    // Validar extensión (doble verificación)
    const extension = file.name.split('.').pop()?.toLowerCase();
    const validExtensions = allowedTypes.map(type => {
      const parts = type.split('/');
      return parts[parts.length - 1];
    });

    if (extension && !validExtensions.includes(extension)) {
      return {
        valid: false,
        error: 'La extensión del archivo no coincide con su tipo'
      };
    }

    return { valid: true };
  }
}
