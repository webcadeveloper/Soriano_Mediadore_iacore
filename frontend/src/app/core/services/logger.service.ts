import { Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';

/**
 * Servicio centralizado de logging
 * - En desarrollo: muestra todos los logs
 * - En producción: solo muestra errores críticos
 */
@Injectable({
  providedIn: 'root'
})
export class LoggerService {
  private readonly isDevelopment = !environment.production;

  /**
   * Log de información general (solo en desarrollo)
   */
  info(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.log(`[INFO] ${message}`, ...args);
    }
  }

  /**
   * Log de advertencias (solo en desarrollo)
   */
  warn(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.warn(`[WARN] ${message}`, ...args);
    }
  }

  /**
   * Log de errores (siempre se muestra)
   */
  error(message: string, error?: any): void {
    console.error(`[ERROR] ${message}`, error);

    // TODO: Aquí se podría enviar a un servicio de monitoreo externo
    // como Sentry, LogRocket, etc.
    // this.sendToMonitoring(message, error);
  }

  /**
   * Log de debug (solo en desarrollo)
   */
  debug(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.log(`[DEBUG] ${message}`, ...args);
    }
  }

  /**
   * Agrupa logs relacionados (solo en desarrollo)
   */
  group(title: string, callback: () => void): void {
    if (this.isDevelopment) {
      console.group(title);
      callback();
      console.groupEnd();
    }
  }

  /**
   * Mide el tiempo de ejecución (solo en desarrollo)
   */
  time(label: string): void {
    if (this.isDevelopment) {
      console.time(label);
    }
  }

  timeEnd(label: string): void {
    if (this.isDevelopment) {
      console.timeEnd(label);
    }
  }
}
