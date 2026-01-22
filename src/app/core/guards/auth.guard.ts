import { Injectable } from '@angular/core';
import { Router, UrlTree } from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService } from '../services/auth.service';

/**
 * Guard para proteger rutas que requieren autenticacion
 * La autenticacion se maneja via backend con Microsoft OAuth
 */
@Injectable({
  providedIn: 'root'
})
export class AuthGuard {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree {
    if (this.authService.isAuthenticated()) {
      return true;
    }

    // No autenticado - el backend redirigira a login de Microsoft
    console.warn('Acceso denegado - Redirigiendo a login');
    // Redirigir al login de Microsoft via backend
    window.location.href = '/auth/login';
    return false;
  }

  canActivateChild(): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree {
    return this.canActivate();
  }
}
