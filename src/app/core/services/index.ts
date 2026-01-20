/**
 * Barrel export para servicios del core
 * Facilita las importaciones en otros módulos
 */

// Servicios de autenticación y seguridad
export * from './auth.service';
export * from './security.service';
export * from './secure-storage.service';

// Servicios de utilidad
export * from './logger.service';
export * from './accessibility.service';
export * from './pwa-update.service';
export * from './meta-tags.service';

// Servicios de API
export * from './api.service';
