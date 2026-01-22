import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from '../services/auth.service';

/**
 * Interceptor para añadir token de autenticación a todas las peticiones HTTP
 */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private authService: AuthService) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    // Obtener token del servicio de autenticación
    const token = this.authService.getToken();

    // Si existe token, clonar request y añadir header Authorization
    if (token) {
      const authReq = req.clone({
        setHeaders: {
          Authorization: token
        }
      });
      return next.handle(authReq);
    }

    // Si no hay token, continuar sin modificar
    return next.handle(req);
  }
}
