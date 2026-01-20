import { Component, AfterViewInit, ElementRef, ViewChild } from '@angular/core';
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
export class LoginComponent implements AfterViewInit {
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
    private a11y: AccessibilityService
  ) {
    // Si ya está autenticado, redirigir al dashboard
    if (this.authService.isAuthenticated()) {
      this.router.navigate(['/dashboard']);
    }
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
    this.a11y.announce('Iniciando sesión, por favor espere', 'polite');

    this.authService.login(this.username, this.password).subscribe({
      next: (response) => {
        this.loading = false;
        if (response.success && response.user) {
          const message = `¡Bienvenido ${response.user.nombre}! Redirigiendo al panel principal`;
          this.snackBar.open(message, 'Cerrar', {
            duration: 3000
          });
          this.a11y.announce(message, 'polite');
          this.router.navigate(['/dashboard']);
        } else {
          const message = response.error || 'Error al iniciar sesión';
          this.snackBar.open(message, 'Cerrar', {
            duration: 4000
          });
          this.a11y.announce(message, 'assertive');
        }
      },
      error: (error) => {
        this.loading = false;
        const message = 'Error de conexión. Intente nuevamente.';
        this.snackBar.open(message, 'Cerrar', {
          duration: 4000
        });
        this.a11y.announce(message, 'assertive');
        console.error('Login error:', error);
      }
    });
  }

  togglePasswordVisibility(): void {
    this.hidePassword = !this.hidePassword;
  }
}
