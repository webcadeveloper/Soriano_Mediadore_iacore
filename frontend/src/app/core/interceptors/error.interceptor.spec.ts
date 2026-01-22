import { TestBed, fakeAsync, tick, flush } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { HTTP_INTERCEPTORS, HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ErrorInterceptor } from './error.interceptor';

describe('ErrorInterceptor', () => {
  let interceptor: ErrorInterceptor;
  let httpMock: HttpTestingController;
  let httpClient: HttpClient;
  let router: jasmine.SpyObj<Router>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;

  beforeEach(() => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        ErrorInterceptor,
        { provide: Router, useValue: routerSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
        {
          provide: HTTP_INTERCEPTORS,
          useClass: ErrorInterceptor,
          multi: true
        }
      ]
    });

    interceptor = TestBed.inject(ErrorInterceptor);
    httpMock = TestBed.inject(HttpTestingController);
    httpClient = TestBed.inject(HttpClient);
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(interceptor).toBeTruthy();
  });

  describe('Successful requests', () => {
    it('should not interfere with successful requests', () => {
      httpClient.get('/api/test').subscribe(response => {
        expect(response).toEqual({ success: true });
      });

      const req = httpMock.expectOne('/api/test');
      req.flush({ success: true });

      expect(snackBar.open).not.toHaveBeenCalled();
    });
  });

  describe('Retry logic', () => {
    it('should retry GET requests 2 times on failure', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });

      // First attempt
      const req1 = httpMock.expectOne('/api/test');
      req1.flush('Error', { status: 500, statusText: 'Server Error' });

      // Second attempt (retry 1)
      const req2 = httpMock.expectOne('/api/test');
      req2.flush('Error', { status: 500, statusText: 'Server Error' });

      // Third attempt (retry 2)
      const req3 = httpMock.expectOne('/api/test');
      req3.flush('Error', { status: 500, statusText: 'Server Error' });

      httpMock.verify();
    });

    it('should not retry POST requests', () => {
      httpClient.post('/api/test', {}).subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Server Error' });

      httpMock.verify();
    });

    it('should not retry PUT requests', () => {
      httpClient.put('/api/test', {}).subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Server Error' });

      httpMock.verify();
    });

    it('should not retry DELETE requests', () => {
      httpClient.delete('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Server Error' });

      httpMock.verify();
    });
  });

  describe('Client-side errors', () => {
    it('should handle client-side errors', () => {
      spyOn(console, 'error');

      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toContain('Error:');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.error(new ErrorEvent('Network error', {
        message: 'Connection failed'
      }));

      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('HTTP status codes', () => {
    it('should handle 0 (network error)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('No se puede conectar con el servidor. Verifique su conexión.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 0, statusText: 'Unknown Error' });

      expect(snackBar.open).toHaveBeenCalledWith(
        'No se puede conectar con el servidor. Verifique su conexión.',
        'Cerrar',
        { duration: 5000, panelClass: ['error-snackbar'] }
      );
    });

    it('should handle 400 (Bad Request)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Solicitud incorrecta. Verifique los datos.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 400, statusText: 'Bad Request' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 401 (Unauthorized) and redirect to login', fakeAsync(() => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Su sesión ha expirado. Por favor, inicie sesión nuevamente.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 401, statusText: 'Unauthorized' });

      // Should not show snackbar for 401
      expect(snackBar.open).not.toHaveBeenCalled();

      // Should redirect after 2 seconds
      tick(2000);
      expect(router.navigate).toHaveBeenCalledWith(['/login']);

      flush();
    }));

    it('should handle 403 (Forbidden)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('No tiene permisos para realizar esta acción.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 403, statusText: 'Forbidden' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 404 (Not Found)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Recurso no encontrado.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 404, statusText: 'Not Found' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 408 (Request Timeout)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Tiempo de espera agotado. Intente nuevamente.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 408, statusText: 'Request Timeout' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 429 (Too Many Requests)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Demasiadas solicitudes. Por favor, espere un momento.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 429, statusText: 'Too Many Requests' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 500 (Internal Server Error)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Error interno del servidor. Intente más tarde.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Internal Server Error' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 502 (Bad Gateway)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('El servidor no está disponible. Intente más tarde.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 502, statusText: 'Bad Gateway' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 503 (Service Unavailable)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('El servidor no está disponible. Intente más tarde.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 503, statusText: 'Service Unavailable' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle 504 (Gateway Timeout)', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('El servidor no está disponible. Intente más tarde.');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 504, statusText: 'Gateway Timeout' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle unknown status codes with custom message', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Custom error message');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush({ message: 'Custom error message' }, { status: 418, statusText: 'I am a teapot' });

      expect(snackBar.open).toHaveBeenCalled();
    });

    it('should handle unknown status codes without custom message', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.message).toBe('Error del servidor (418)');
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 418, statusText: 'I am a teapot' });

      expect(snackBar.open).toHaveBeenCalled();
    });
  });

  describe('Error propagation', () => {
    it('should return error object with status and message', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.status).toBe(404);
          expect(error.message).toBeDefined();
          expect(error.originalError).toBeInstanceOf(HttpErrorResponse);
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 404, statusText: 'Not Found' });
    });

    it('should include original error in propagated error', () => {
      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: (error) => {
          expect(error.originalError).toBeDefined();
          expect(error.originalError.status).toBe(500);
        }
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Internal Server Error' });
    });
  });

  describe('Console logging', () => {
    it('should log server errors to console', () => {
      spyOn(console, 'error');

      httpClient.get('/api/test').subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });

      const req = httpMock.expectOne('/api/test');
      req.flush('Error', { status: 500, statusText: 'Internal Server Error' });

      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('Multiple errors', () => {
    it('should handle multiple consecutive errors', () => {
      // First error
      httpClient.get('/api/test1').subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });
      const req1 = httpMock.expectOne('/api/test1');
      req1.flush('Error', { status: 404, statusText: 'Not Found' });

      // Second error
      httpClient.get('/api/test2').subscribe({
        next: () => fail('should have failed'),
        error: () => {}
      });
      const req2 = httpMock.expectOne('/api/test2');
      req2.flush('Error', { status: 500, statusText: 'Internal Server Error' });

      // Both should show snackbar
      expect(snackBar.open).toHaveBeenCalledTimes(2);
    });
  });
});
