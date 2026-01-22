import { Injectable } from '@angular/core';
import { SecurityService } from './security.service';

/**
 * Servicio de almacenamiento seguro con cifrado
 * Reemplaza el uso directo de localStorage con datos cifrados
 */
@Injectable({
  providedIn: 'root'
})
export class SecureStorageService {
  constructor(private securityService: SecurityService) {}

  /**
   * Guarda un valor en localStorage de forma cifrada
   * @param key - Clave de almacenamiento
   * @param value - Valor a almacenar (será serializado a JSON)
   */
  setItem<T>(key: string, value: T): void {
    try {
      const jsonString = JSON.stringify(value);
      const encrypted = this.securityService.encrypt(jsonString);
      localStorage.setItem(key, encrypted);
    } catch (error) {
      console.error(`Error guardando ${key} en storage:`, error);
    }
  }

  /**
   * Recupera un valor cifrado de localStorage
   * @param key - Clave de almacenamiento
   * @returns Valor descifrado o null si no existe
   */
  getItem<T>(key: string): T | null {
    try {
      const encrypted = localStorage.getItem(key);
      if (!encrypted) return null;

      const decrypted = this.securityService.decrypt(encrypted);
      if (!decrypted) return null;

      return JSON.parse(decrypted) as T;
    } catch (error) {
      console.error(`Error recuperando ${key} de storage:`, error);
      return null;
    }
  }

  /**
   * Elimina un item del localStorage
   * @param key - Clave a eliminar
   */
  removeItem(key: string): void {
    localStorage.removeItem(key);
  }

  /**
   * Limpia todos los items del localStorage
   */
  clear(): void {
    localStorage.clear();
  }

  /**
   * Verifica si existe una clave en localStorage
   * @param key - Clave a verificar
   * @returns true si existe
   */
  hasItem(key: string): boolean {
    return localStorage.getItem(key) !== null;
  }

  /**
   * Guarda en sessionStorage (solo durante la sesión)
   * @param key - Clave
   * @param value - Valor
   */
  setSessionItem<T>(key: string, value: T): void {
    try {
      const jsonString = JSON.stringify(value);
      const encrypted = this.securityService.encrypt(jsonString);
      sessionStorage.setItem(key, encrypted);
    } catch (error) {
      console.error(`Error guardando ${key} en session storage:`, error);
    }
  }

  /**
   * Recupera de sessionStorage
   * @param key - Clave
   * @returns Valor o null
   */
  getSessionItem<T>(key: string): T | null {
    try {
      const encrypted = sessionStorage.getItem(key);
      if (!encrypted) return null;

      const decrypted = this.securityService.decrypt(encrypted);
      if (!decrypted) return null;

      return JSON.parse(decrypted) as T;
    } catch (error) {
      console.error(`Error recuperando ${key} de session storage:`, error);
      return null;
    }
  }

  /**
   * Migra datos no cifrados a cifrados (para actualización)
   * @param key - Clave a migrar
   */
  migrateUnencryptedData<T>(key: string): void {
    try {
      const plainData = localStorage.getItem(key);
      if (!plainData) return;

      // Intentar parsear como JSON
      try {
        const parsed = JSON.parse(plainData);
        // Si se puede parsear, re-guardar cifrado
        this.setItem(key, parsed);
        console.log(`✅ Datos migrados y cifrados: ${key}`);
      } catch {
        // Si no se puede parsear, no hacer nada
        console.warn(`⚠️ No se pudo migrar ${key}: formato inválido`);
      }
    } catch (error) {
      console.error(`Error migrando ${key}:`, error);
    }
  }

  /**
   * Migra todas las claves conocidas
   */
  migrateAllKnownKeys(): void {
    const knownKeys = [
      'recobros_recibos',
      'recobros_templates',
      'recobros_config',
      'soriano_email_templates'
    ];

    console.log('🔄 Iniciando migración de datos a formato cifrado...');
    knownKeys.forEach(key => this.migrateUnencryptedData(key));
    console.log('✅ Migración completada');
  }
}
