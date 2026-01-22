import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  nombre: string;
  displayName?: string;
  jobTitle?: string;
}

export type UserRole = 'agente' | 'supervisor' | 'director' | 'auditor' | 'admin';

export interface MicrosoftUser {
  id: string;
  displayName: string;
  givenName: string;
  surname: string;
  mail: string;
  userPrincipalName: string;
  jobTitle?: string;
}

export interface AuthResponse {
  authenticated: boolean;
  user?: MicrosoftUser;
  expiresAt?: string;
  error?: string;
}

/**
 * Servicio de autenticacion con Microsoft OAuth
 * La autenticacion se maneja via backend (cookies de sesion)
 */
@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  private isLoadingSubject = new BehaviorSubject<boolean>(true);

  public currentUser$ = this.currentUserSubject.asObservable();
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();
  public isLoading$ = this.isLoadingSubject.asObservable();

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    this.checkAuth();
  }

  /**
   * Verifica si el usuario esta autenticado via backend
   */
  checkAuth(): void {
    this.isLoadingSubject.next(true);

    this.http.get<AuthResponse>('/auth/me', { withCredentials: true })
      .pipe(
        catchError(err => {
          console.log('No autenticado');
          return of({ authenticated: false });
        })
      )
      .subscribe(response => {
        if (response.authenticated && response.user) {
          const user = this.mapMicrosoftUser(response.user);
          this.currentUserSubject.next(user);
          this.isAuthenticatedSubject.next(true);
          console.log('Usuario autenticado:', user.nombre);
        } else {
          this.currentUserSubject.next(null);
          this.isAuthenticatedSubject.next(false);
        }
        this.isLoadingSubject.next(false);
      });
  }

  /**
   * Mapea usuario de Microsoft a formato interno
   */
  private mapMicrosoftUser(msUser: MicrosoftUser): User {
    return {
      id: msUser.id,
      username: msUser.userPrincipalName,
      email: msUser.mail || msUser.userPrincipalName,
      role: this.determineRole(msUser),
      nombre: msUser.displayName,
      displayName: msUser.displayName,
      jobTitle: msUser.jobTitle
    };
  }

  /**
   * Determina el rol basado en el cargo o email
   */
  private determineRole(msUser: MicrosoftUser): UserRole {
    const jobTitle = (msUser.jobTitle || '').toLowerCase();
    const email = (msUser.mail || msUser.userPrincipalName || '').toLowerCase();

    if (jobTitle.includes('director') || jobTitle.includes('gerente')) {
      return 'director';
    }
    if (jobTitle.includes('supervisor') || jobTitle.includes('jefe')) {
      return 'supervisor';
    }
    if (jobTitle.includes('auditor')) {
      return 'auditor';
    }
    if (email.includes('admin')) {
      return 'admin';
    }
    return 'agente';
  }

  /**
   * Inicia el login con Microsoft
   * Redirige al endpoint de OAuth del backend
   */
  login(): void {
    window.location.href = '/auth/login';
  }

  /**
   * Logout de usuario
   * Redirige al endpoint de logout del backend que limpia la sesion
   * y redirige a Microsoft para cerrar sesion alli tambien
   */
  logout(): void {
    console.log('Cerrando sesion...');

    // Limpiar estado local
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);

    // Llamar al endpoint de logout del backend
    this.http.post('/auth/logout', {}, { withCredentials: true })
      .subscribe({
        next: () => {
          console.log('Sesión cerrada correctamente');
          // Redirigir a la página de login
          this.router.navigate(['/login']);
        },
        error: (err) => {
          console.error('Error al cerrar sesión:', err);
          // Redirigir a login de todas formas
          this.router.navigate(['/login']);
        }
      });
  }

  /**
   * Obtiene el usuario actual
   */
  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  /**
   * Verifica si el usuario esta autenticado
   */
  isAuthenticated(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  /**
   * Verifica si el usuario tiene un rol especifico
   */
  hasRole(role: UserRole): boolean {
    const user = this.getCurrentUser();
    return user?.role === role;
  }

  /**
   * Verifica si el usuario tiene al menos uno de los roles
   */
  hasAnyRole(roles: UserRole[]): boolean {
    const user = this.getCurrentUser();
    return user ? roles.includes(user.role) : false;
  }

  /**
   * Obtiene informacion actualizada del usuario
   */
  refreshUser(): Observable<User | null> {
    return this.http.get<AuthResponse>('/auth/me', { withCredentials: true })
      .pipe(
        map(response => {
          if (response.authenticated && response.user) {
            const user = this.mapMicrosoftUser(response.user);
            this.currentUserSubject.next(user);
            this.isAuthenticatedSubject.next(true);
            return user;
          }
          return null;
        }),
        catchError(() => {
          this.currentUserSubject.next(null);
          this.isAuthenticatedSubject.next(false);
          return of(null);
        })
      );
  }
}
