import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { HTTP_INTERCEPTORS, HttpClient, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { AuthInterceptor } from './auth.interceptor';
import { AuthService } from '../services/auth.service';
import { Observable, of } from 'rxjs';

describe('AuthInterceptor', () => {
  let interceptor: AuthInterceptor;
  let authService: jasmine.SpyObj<AuthService>;
  let httpMock: HttpTestingController;
  let httpClient: HttpClient;

  beforeEach(() => {
    const authServiceSpy = jasmine.createSpyObj('AuthService', ['getToken']);

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        AuthInterceptor,
        { provide: AuthService, useValue: authServiceSpy },
        {
          provide: HTTP_INTERCEPTORS,
          useClass: AuthInterceptor,
          multi: true
        }
      ]
    });

    interceptor = TestBed.inject(AuthInterceptor);
    authService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>;
    httpMock = TestBed.inject(HttpTestingController);
    httpClient = TestBed.inject(HttpClient);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(interceptor).toBeTruthy();
  });

  describe('intercept - With token', () => {
    it('should add Authorization header when token exists', () => {
      const token = 'Bearer test-token-123';
      authService.getToken.and.returnValue(token);

      httpClient.get('/api/test').subscribe();

      const req = httpMock.expectOne('/api/test');
      expect(req.request.headers.has('Authorization')).toBe(true);
      expect(req.request.headers.get('Authorization')).toBe(token);

      req.flush({ success: true });
    });

    it('should add Bearer token to request', () => {
      const token = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9';
      authService.getToken.and.returnValue(token);

      httpClient.post('/api/users', { name: 'Test' }).subscribe();

      const req = httpMock.expectOne('/api/users');
      expect(req.request.headers.get('Authorization')).toBe(token);

      req.flush({ id: 1, name: 'Test' });
    });

    it('should preserve existing headers when adding Authorization', () => {
      const token = 'Bearer test-token';
      authService.getToken.and.returnValue(token);

      httpClient.get('/api/test', {
        headers: {
          'Content-Type': 'application/json',
          'X-Custom-Header': 'custom-value'
        }
      }).subscribe();

      const req = httpMock.expectOne('/api/test');
      expect(req.request.headers.get('Authorization')).toBe(token);
      expect(req.request.headers.get('Content-Type')).toBe('application/json');
      expect(req.request.headers.get('X-Custom-Header')).toBe('custom-value');

      req.flush({ success: true });
    });

    it('should handle multiple concurrent requests', () => {
      const token = 'Bearer concurrent-token';
      authService.getToken.and.returnValue(token);

      httpClient.get('/api/request1').subscribe();
      httpClient.get('/api/request2').subscribe();
      httpClient.post('/api/request3', {}).subscribe();

      const req1 = httpMock.expectOne('/api/request1');
      const req2 = httpMock.expectOne('/api/request2');
      const req3 = httpMock.expectOne('/api/request3');

      expect(req1.request.headers.get('Authorization')).toBe(token);
      expect(req2.request.headers.get('Authorization')).toBe(token);
      expect(req3.request.headers.get('Authorization')).toBe(token);

      req1.flush({});
      req2.flush({});
      req3.flush({});
    });

    it('should work with different HTTP methods', () => {
      const token = 'Bearer method-token';
      authService.getToken.and.returnValue(token);

      // GET
      httpClient.get('/api/get').subscribe();
      const getReq = httpMock.expectOne('/api/get');
      expect(getReq.request.headers.get('Authorization')).toBe(token);
      getReq.flush({});

      // POST
      httpClient.post('/api/post', {}).subscribe();
      const postReq = httpMock.expectOne('/api/post');
      expect(postReq.request.headers.get('Authorization')).toBe(token);
      postReq.flush({});

      // PUT
      httpClient.put('/api/put', {}).subscribe();
      const putReq = httpMock.expectOne('/api/put');
      expect(putReq.request.headers.get('Authorization')).toBe(token);
      putReq.flush({});

      // DELETE
      httpClient.delete('/api/delete').subscribe();
      const deleteReq = httpMock.expectOne('/api/delete');
      expect(deleteReq.request.headers.get('Authorization')).toBe(token);
      deleteReq.flush({});

      // PATCH
      httpClient.patch('/api/patch', {}).subscribe();
      const patchReq = httpMock.expectOne('/api/patch');
      expect(patchReq.request.headers.get('Authorization')).toBe(token);
      patchReq.flush({});
    });
  });

  describe('intercept - Without token', () => {
    it('should not add Authorization header when token is null', () => {
      authService.getToken.and.returnValue(null);

      httpClient.get('/api/public').subscribe();

      const req = httpMock.expectOne('/api/public');
      expect(req.request.headers.has('Authorization')).toBe(false);

      req.flush({ success: true });
    });

    it('should not add Authorization header when token is undefined', () => {
      authService.getToken.and.returnValue(undefined as any);

      httpClient.get('/api/public').subscribe();

      const req = httpMock.expectOne('/api/public');
      expect(req.request.headers.has('Authorization')).toBe(false);

      req.flush({ success: true });
    });

    it('should not add Authorization header when token is empty string', () => {
      authService.getToken.and.returnValue('');

      httpClient.get('/api/public').subscribe();

      const req = httpMock.expectOne('/api/public');
      expect(req.request.headers.has('Authorization')).toBe(false);

      req.flush({ success: true });
    });

    it('should preserve other headers when no token', () => {
      authService.getToken.and.returnValue(null);

      httpClient.get('/api/test', {
        headers: {
          'Content-Type': 'application/json'
        }
      }).subscribe();

      const req = httpMock.expectOne('/api/test');
      expect(req.request.headers.has('Authorization')).toBe(false);
      expect(req.request.headers.get('Content-Type')).toBe('application/json');

      req.flush({});
    });
  });

  describe('intercept - Token changes', () => {
    it('should use current token for each request', () => {
      // First request with token 1
      authService.getToken.and.returnValue('Bearer token-1');
      httpClient.get('/api/test1').subscribe();
      const req1 = httpMock.expectOne('/api/test1');
      expect(req1.request.headers.get('Authorization')).toBe('Bearer token-1');
      req1.flush({});

      // Second request with token 2
      authService.getToken.and.returnValue('Bearer token-2');
      httpClient.get('/api/test2').subscribe();
      const req2 = httpMock.expectOne('/api/test2');
      expect(req2.request.headers.get('Authorization')).toBe('Bearer token-2');
      req2.flush({});

      // Third request with no token
      authService.getToken.and.returnValue(null);
      httpClient.get('/api/test3').subscribe();
      const req3 = httpMock.expectOne('/api/test3');
      expect(req3.request.headers.has('Authorization')).toBe(false);
      req3.flush({});
    });

    it('should handle token refresh scenarios', () => {
      // Initial request with old token
      authService.getToken.and.returnValue('Bearer old-token');
      httpClient.get('/api/data').subscribe();
      const req1 = httpMock.expectOne('/api/data');
      expect(req1.request.headers.get('Authorization')).toBe('Bearer old-token');
      req1.flush({});

      // After refresh - new token
      authService.getToken.and.returnValue('Bearer refreshed-token');
      httpClient.get('/api/data').subscribe();
      const req2 = httpMock.expectOne('/api/data');
      expect(req2.request.headers.get('Authorization')).toBe('Bearer refreshed-token');
      req2.flush({});
    });
  });

  describe('intercept - Direct method call', () => {
    it('should clone and modify request when token exists', () => {
      const token = 'Bearer direct-token';
      authService.getToken.and.returnValue(token);

      const mockRequest = new HttpRequest('GET', '/api/test');
      const mockNext: HttpHandler = {
        handle: jasmine.createSpy('handle').and.returnValue(of({} as HttpEvent<any>))
      };

      interceptor.intercept(mockRequest, mockNext);

      expect(mockNext.handle).toHaveBeenCalled();
      const modifiedRequest = (mockNext.handle as jasmine.Spy).calls.mostRecent().args[0];
      expect(modifiedRequest.headers.get('Authorization')).toBe(token);
    });

    it('should not modify request when no token', () => {
      authService.getToken.and.returnValue(null);

      const mockRequest = new HttpRequest('GET', '/api/test');
      const mockNext: HttpHandler = {
        handle: jasmine.createSpy('handle').and.returnValue(of({} as HttpEvent<any>))
      };

      interceptor.intercept(mockRequest, mockNext);

      expect(mockNext.handle).toHaveBeenCalledWith(mockRequest);
    });

    it('should call getToken for each intercept', () => {
      authService.getToken.and.returnValue('Bearer token');

      const mockRequest = new HttpRequest('GET', '/api/test');
      const mockNext: HttpHandler = {
        handle: jasmine.createSpy('handle').and.returnValue(of({} as HttpEvent<any>))
      };

      interceptor.intercept(mockRequest, mockNext);
      interceptor.intercept(mockRequest, mockNext);
      interceptor.intercept(mockRequest, mockNext);

      expect(authService.getToken).toHaveBeenCalledTimes(3);
    });
  });

  describe('Request cloning', () => {
    it('should not mutate original request', () => {
      const token = 'Bearer immutable-token';
      authService.getToken.and.returnValue(token);

      const originalRequest = new HttpRequest('GET', '/api/test');
      const mockNext: HttpHandler = {
        handle: jasmine.createSpy('handle').and.returnValue(of({} as HttpEvent<any>))
      };

      interceptor.intercept(originalRequest, mockNext);

      // Original request should remain unchanged
      expect(originalRequest.headers.has('Authorization')).toBe(false);

      // Modified request should have the header
      const modifiedRequest = (mockNext.handle as jasmine.Spy).calls.mostRecent().args[0];
      expect(modifiedRequest.headers.get('Authorization')).toBe(token);
    });
  });
});
