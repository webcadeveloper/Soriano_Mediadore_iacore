import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, of, throwError } from 'rxjs';
import { delay, tap, catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { SecureStorageService } from './secure-storage.service';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  nombre: string;
}

export type UserRole = 'agente' | 'supervisor' | 'director' | 'auditor' | 'admin';

export interface AuthResponse {
  success: boolean;
  token?: string;
  refreshToken?: string;
  user?: User;
  error?: string;
}

/**
 * Servicio de autenticación con JWT
 */
@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly TOKEN_KEY = 'auth_token';
  private readonly REFRESH_TOKEN_KEY = 'refresh_token';
  private readonly USER_KEY = 'current_user';
  private readonly TOKEN_EXPIRY_KEY = 'token_expiry';

  private currentUserSubject = new BehaviorSubject<User | null>(null);
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);

  public currentUser$ = this.currentUserSubject.asObservable();
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  constructor(
    private storage: SecureStorageService,
    private router: Router
  ) {
    this.checkStoredAuth();
  }

  /**
   * Verifica si hay una sesión guardada al iniciar
   */
  private checkStoredAuth(): void {
    const token = this.storage.getItem<string>(this.TOKEN_KEY);
    const user = this.storage.getItem<User>(this.USER_KEY);
    const expiry = this.storage.getItem<number>(this.TOKEN_EXPIRY_KEY);

    if (token && user && expiry) {
      // Verificar si el token no ha expirado
      if (Date.now() < expiry) {
        this.currentUserSubject.next(user);
        this.isAuthenticatedSubject.next(true);
      } else {
        // Token expirado, intentar refresh
        this.refreshToken().subscribe();
      }
    }
  }

  /**
   * Login de usuario
   * @param username - Nombre de usuario
   * @param password - Contraseña
   * @returns Observable con resultado del login
   */
  login(username: string, password: string): Observable<AuthResponse> {
    // Simulación de login (en producción llamaría al backend)
    console.log('🔐 Intentando login:', username);

    // Validar credenciales (demo)
    const validUsers: { [key: string]: { password: string; user: User } } = {
      'admin': {
        password: 'admin123',
        user: {
          id: '1',
          username: 'admin',
          email: 'admin@sorianomediadores.es',
          role: 'admin',
          nombre: 'Administrador'
        }
      },
      'agente': {
        password: 'agente123',
        user: {
          id: '2',
          username: 'agente',
          email: 'agente@sorianomediadores.es',
          role: 'agente',
          nombre: 'Laura García'
        }
      },
      'supervisor': {
        password: 'supervisor123',
        user: {
          id: '3',
          username: 'supervisor',
          email: 'supervisor@sorianomediadores.es',
          role: 'supervisor',
          nombre: 'Carlos Ruiz'
        }
      }
    };

    const userData = validUsers[username.toLowerCase()];

    if (userData && userData.password === password) {
      // Login exitoso
      const token = this.generateMockToken();
      const refreshToken = this.generateMockToken();
      const expiry = Date.now() + (60 * 60 * 1000); // 1 hora

      // Guardar en storage seguro
      this.storage.setItem(this.TOKEN_KEY, token);
      this.storage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
      this.storage.setItem(this.USER_KEY, userData.user);
      this.storage.setItem(this.TOKEN_EXPIRY_KEY, expiry);

      // Actualizar subjects
      this.currentUserSubject.next(userData.user);
      this.isAuthenticatedSubject.next(true);

      return of({
        success: true,
        token,
        refreshToken,
        user: userData.user
      }).pipe(delay(500)); // Simular delay de red
    } else {
      return of({
        success: false,
        error: 'Credenciales inválidas'
      }).pipe(delay(500));
    }
  }

  /**
   * Logout de usuario
   */
  logout(): void {
    console.log('🚪 Cerrando sesión...');

    // Limpiar storage
    this.storage.removeItem(this.TOKEN_KEY);
    this.storage.removeItem(this.REFRESH_TOKEN_KEY);
    this.storage.removeItem(this.USER_KEY);
    this.storage.removeItem(this.TOKEN_EXPIRY_KEY);

    // Actualizar subjects
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);

    // Redirigir a login
    this.router.navigate(['/login']);
  }

  /**
   * Refresca el token de autenticación
   */
  refreshToken(): Observable<AuthResponse> {
    const refreshToken = this.storage.getItem<string>(this.REFRESH_TOKEN_KEY);

    if (!refreshToken) {
      return throwError(() => new Error('No refresh token available'));
    }

    // Simulación de refresh (en producción llamaría al backend)
    const user = this.storage.getItem<User>(this.USER_KEY);
    const newToken = this.generateMockToken();
    const newExpiry = Date.now() + (60 * 60 * 1000);

    this.storage.setItem(this.TOKEN_KEY, newToken);
    this.storage.setItem(this.TOKEN_EXPIRY_KEY, newExpiry);

    return of({
      success: true,
      token: newToken,
      user: user || undefined
    }).pipe(delay(300));
  }

  /**
   * Obtiene el token actual
   */
  getToken(): string | null {
    return this.storage.getItem<string>(this.TOKEN_KEY);
  }

  /**
   * Obtiene el usuario actual
   */
  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  /**
   * Verifica si el usuario está autenticado
   */
  isAuthenticated(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  /**
   * Verifica si el usuario tiene un rol específico
   * @param role - Rol a verificar
   */
  hasRole(role: UserRole): boolean {
    const user = this.getCurrentUser();
    return user?.role === role;
  }

  /**
   * Verifica si el usuario tiene al menos uno de los roles
   * @param roles - Array de roles permitidos
   */
  hasAnyRole(roles: UserRole[]): boolean {
    const user = this.getCurrentUser();
    return user ? roles.includes(user.role) : false;
  }

  /**
   * Genera un token mock para desarrollo
   * En producción esto vendría del backend
   */
  private generateMockToken(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let token = '';
    for (let i = 0; i < 64; i++) {
      token += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return `Bearer.${token}.${Date.now()}`;
  }

  /**
   * Verifica si el token está próximo a expirar (menos de 5 minutos)
   */
  isTokenExpiringSoon(): boolean {
    const expiry = this.storage.getItem<number>(this.TOKEN_EXPIRY_KEY);
    if (!expiry) return true;

    const fiveMinutes = 5 * 60 * 1000;
    return Date.now() > (expiry - fiveMinutes);
  }

  /**
   * Renueva automáticamente el token si está próximo a expirar
   */
  autoRenewToken(): void {
    if (this.isAuthenticated() && this.isTokenExpiringSoon()) {
      console.log('🔄 Renovando token automáticamente...');
      this.refreshToken().subscribe({
        next: () => console.log('✅ Token renovado'),
        error: () => {
          console.error('❌ Error renovando token, cerrando sesión');
          this.logout();
        }
      });
    }
  }
}
