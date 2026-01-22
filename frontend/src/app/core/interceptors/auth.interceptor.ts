import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { Observable } from 'rxjs';

/**
 * Interceptor para asegurar que las credenciales (cookies) se envíen con cada petición
 * La autenticación ahora se maneja via cookies de sesión del backend
 */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    // Clonar la petición para incluir credenciales (cookies)
    const authReq = req.clone({
      withCredentials: true
    });

    return next.handle(authReq);
  }
}
