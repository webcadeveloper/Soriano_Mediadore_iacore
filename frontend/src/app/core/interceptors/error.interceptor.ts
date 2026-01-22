import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';

/**
 * Interceptor para manejo centralizado de errores HTTP
 */
@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
  constructor(
    private router: Router,
    private snackBar: MatSnackBar
  ) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(req).pipe(
      // Reintentar automáticamente en caso de error de red (excepto POST/PUT/DELETE)
      retry({
        count: req.method === 'GET' ? 2 : 0,
        delay: 1000
      }),
      catchError((error: HttpErrorResponse) => {
        let errorMessage = 'Ha ocurrido un error';
        let showSnackbar = true;

        if (error.error instanceof ErrorEvent) {
          // Error del lado del cliente
          errorMessage = `Error: ${error.error.message}`;
          console.error('Error del cliente:', error.error.message);
        } else {
          // Error del lado del servidor
          switch (error.status) {
            case 0:
              errorMessage = 'No se puede conectar con el servidor. Verifique su conexión.';
              break;
            case 400:
              errorMessage = 'Solicitud incorrecta. Verifique los datos.';
              break;
            case 401:
              errorMessage = 'Su sesión ha expirado. Por favor, inicie sesión nuevamente.';
              // Redirigir a login después de 2 segundos
              setTimeout(() => {
                this.router.navigate(['/login']);
              }, 2000);
              break;
            case 403:
              errorMessage = 'No tiene permisos para realizar esta acción.';
              break;
            case 404:
              errorMessage = 'Recurso no encontrado.';
              break;
            case 408:
              errorMessage = 'Tiempo de espera agotado. Intente nuevamente.';
              break;
            case 429:
              errorMessage = 'Demasiadas solicitudes. Por favor, espere un momento.';
              break;
            case 500:
              errorMessage = 'Error interno del servidor. Intente más tarde.';
              break;
            case 502:
            case 503:
            case 504:
              errorMessage = 'El servidor no está disponible. Intente más tarde.';
              break;
            default:
              errorMessage = error.error?.message || `Error del servidor (${error.status})`;
          }

          console.error(
            `Error HTTP ${error.status}:`,
            error.message,
            '\nURL:', error.url,
            '\nDetalles:', error.error
          );
        }

        // Mostrar snackbar con el error (excepto en 401 que redirige)
        if (showSnackbar && error.status !== 401) {
          this.snackBar.open(errorMessage, 'Cerrar', {
            duration: 5000,
            panelClass: ['error-snackbar']
          });
        }

        // Propagar el error para que los componentes puedan manejarlo si lo necesitan
        return throwError(() => ({
          status: error.status,
          message: errorMessage,
          originalError: error
        }));
      })
    );
  }
}
