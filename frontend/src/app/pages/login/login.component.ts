import { Component, AfterViewInit, ElementRef, ViewChild, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { AuthService } from '../../core/services/auth.service';
import { AccessibilityService } from '../../core/services/accessibility.service';
import { MetaTagsService } from '../../core/services/meta-tags.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatSnackBarModule
  ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit, AfterViewInit {
  @ViewChild('usernameInput') usernameInput!: ElementRef<HTMLInputElement>;

  username = '';
  password = '';
  loading = false;
  hidePassword = true;
  currentYear = new Date().getFullYear();

  constructor(
    private authService: AuthService,
    private router: Router,
    private snackBar: MatSnackBar,
    private a11y: AccessibilityService,
    private metaTagsService: MetaTagsService
  ) {
    // Si ya está autenticado, redirigir al dashboard
    if (this.authService.isAuthenticated()) {
      this.router.navigate(['/dashboard']);
    }
  }

  ngOnInit(): void {
    // Configurar meta tags para la página de login
    this.metaTagsService.updateMetaTags({
      title: 'Iniciar Sesión - Soriano Mediadores CRM',
      description: 'Accede al sistema CRM de Soriano Mediadores de Seguros. Gestiona clientes, recobros y reportes de forma segura.',
      canonical: 'https://sorianomediadores.com/login',
      ogTitle: 'Iniciar Sesión - Soriano Mediadores',
      ogDescription: 'Acceso al sistema de gestión de mediadores de seguros',
      robots: 'noindex, nofollow' // No indexar página de login por seguridad
    });
  }

  ngAfterViewInit(): void {
    // Enfocar el campo de usuario al cargar
    this.a11y.focusElement(this.usernameInput?.nativeElement, 100);
  }

  onSubmit(): void {
    if (!this.username || !this.password) {
      const message = 'Por favor complete todos los campos';
      this.snackBar.open(message, 'Cerrar', {
        duration: 3000
      });
      this.a11y.announce(message, 'assertive');
      return;
    }

    this.loading = true;
    this.a11y.announce('Redirigiendo a Microsoft para autenticación', 'polite');

    // Redirigir a Microsoft OAuth
    this.authService.login();
  }

  togglePasswordVisibility(): void {
    this.hidePassword = !this.hidePassword;
  }
}
