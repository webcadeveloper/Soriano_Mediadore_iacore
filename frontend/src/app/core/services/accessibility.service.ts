import { Injectable } from '@angular/core';

/**
 * Servicio para gestionar funcionalidades de accesibilidad
 * - Anuncios para lectores de pantalla
 * - Gestión de focus
 * - Skip links
 */
@Injectable({
  providedIn: 'root'
})
export class AccessibilityService {
  private liveRegion: HTMLElement | null = null;

  constructor() {
    this.createLiveRegion();
  }

  /**
   * Crea una región ARIA live para anuncios a lectores de pantalla
   */
  private createLiveRegion(): void {
    if (typeof document !== 'undefined') {
      this.liveRegion = document.createElement('div');
      this.liveRegion.setAttribute('role', 'status');
      this.liveRegion.setAttribute('aria-live', 'polite');
      this.liveRegion.setAttribute('aria-atomic', 'true');
      this.liveRegion.className = 'sr-only'; // Oculto visualmente pero accesible
      this.liveRegion.style.position = 'absolute';
      this.liveRegion.style.left = '-10000px';
      this.liveRegion.style.width = '1px';
      this.liveRegion.style.height = '1px';
      this.liveRegion.style.overflow = 'hidden';
      document.body.appendChild(this.liveRegion);
    }
  }

  /**
   * Anuncia un mensaje a lectores de pantalla
   */
  announce(message: string, priority: 'polite' | 'assertive' = 'polite'): void {
    if (this.liveRegion) {
      this.liveRegion.setAttribute('aria-live', priority);
      this.liveRegion.textContent = message;

      // Limpiar después de 5 segundos
      setTimeout(() => {
        if (this.liveRegion) {
          this.liveRegion.textContent = '';
        }
      }, 5000);
    }
  }

  /**
   * Mueve el focus a un elemento específico
   */
  focusElement(element: HTMLElement | null, delay: number = 0): void {
    if (element) {
      setTimeout(() => {
        element.focus();
      }, delay);
    }
  }

  /**
   * Mueve el focus al primer elemento interactivo del contenedor
   */
  focusFirstInteractive(container: HTMLElement): void {
    const focusable = container.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );

    if (focusable.length > 0) {
      focusable[0].focus();
    }
  }

  /**
   * Obtiene todos los elementos enfocables en un contenedor
   */
  getFocusableElements(container: HTMLElement): HTMLElement[] {
    return Array.from(
      container.querySelectorAll<HTMLElement>(
        'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
    );
  }

  /**
   * Crea un trap de focus para modales/diálogos
   */
  trapFocus(container: HTMLElement, event: KeyboardEvent): void {
    const focusableElements = this.getFocusableElements(container);

    if (focusableElements.length === 0) return;

    const firstFocusable = focusableElements[0];
    const lastFocusable = focusableElements[focusableElements.length - 1];

    if (event.key === 'Tab') {
      if (event.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstFocusable) {
          event.preventDefault();
          lastFocusable.focus();
        }
      } else {
        // Tab
        if (document.activeElement === lastFocusable) {
          event.preventDefault();
          firstFocusable.focus();
        }
      }
    }
  }

  /**
   * Verifica si un elemento es visible en el viewport
   */
  isInViewport(element: HTMLElement): boolean {
    const rect = element.getBoundingClientRect();
    return (
      rect.top >= 0 &&
      rect.left >= 0 &&
      rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
      rect.right <= (window.innerWidth || document.documentElement.clientWidth)
    );
  }

  /**
   * Scroll suave a un elemento con anuncio
   */
  scrollToElement(element: HTMLElement, announce: boolean = true): void {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' });

    if (announce) {
      const label = element.getAttribute('aria-label') ||
                    element.getAttribute('title') ||
                    'Sección cargada';
      this.announce(label);
    }
  }
}
