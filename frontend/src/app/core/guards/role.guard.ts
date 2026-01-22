import { Injectable } from '@angular/core';
import { Router, UrlTree, ActivatedRouteSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService, UserRole } from '../services/auth.service';

/**
 * Guard para proteger rutas según el rol del usuario
 *
 * Uso en rutas:
 * ```typescript
 * {
 *   path: 'admin',
 *   canActivate: [RoleGuard],
 *   data: { roles: ['admin', 'director'] }
 * }
 * ```
 */
@Injectable({
  providedIn: 'root'
})
export class RoleGuard {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(route: ActivatedRouteSnapshot): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree {
    const requiredRoles = route.data['roles'] as UserRole[];

    if (!requiredRoles || requiredRoles.length === 0) {
      // Si no se especifican roles, solo verificar autenticación
      return this.authService.isAuthenticated() || this.router.createUrlTree(['/login']);
    }

    if (!this.authService.isAuthenticated()) {
      console.warn('🚫 Usuario no autenticado - Redirigiendo a login');
      return this.router.createUrlTree(['/login']);
    }

    if (this.authService.hasAnyRole(requiredRoles)) {
      return true;
    }

    // Usuario autenticado pero sin permisos
    console.warn('🚫 Acceso denegado - Rol insuficiente');
    return this.router.createUrlTree(['/dashboard']);
  }

  canActivateChild(route: ActivatedRouteSnapshot): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree {
    return this.canActivate(route);
  }
}
