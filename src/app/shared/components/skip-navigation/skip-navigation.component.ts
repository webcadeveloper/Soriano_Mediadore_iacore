import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

/**
 * Componente de navegación de salto (skip links)
 * Permite a usuarios de teclado/lectores de pantalla saltar directamente al contenido principal
 */
@Component({
  selector: 'app-skip-navigation',
  standalone: true,
  imports: [CommonModule],
  template: `
    <nav class="skip-navigation" aria-label="Enlaces de navegación rápida">
      <a href="#main-content" class="skip-link">
        Saltar al contenido principal
      </a>
      <a href="#main-navigation" class="skip-link">
        Saltar a la navegación
      </a>
      <a href="#search" class="skip-link" *ngIf="showSearch">
        Saltar a la búsqueda
      </a>
    </nav>
  `,
  styles: [`
    .skip-navigation {
      position: relative;
    }

    .skip-link {
      position: absolute;
      top: -40px;
      left: 0;
      z-index: 10000;
      padding: 8px 16px;
      background: #1976d2;
      color: white;
      text-decoration: none;
      font-weight: 500;
      border-radius: 0 0 4px 0;
      transition: top 0.2s;
      white-space: nowrap;
    }

    .skip-link:focus {
      top: 0;
      outline: 2px solid #fff;
      outline-offset: 2px;
    }

    .skip-link:hover:focus {
      background: #1565c0;
    }
  `]
})
export class SkipNavigationComponent {
  showSearch = true;
}
