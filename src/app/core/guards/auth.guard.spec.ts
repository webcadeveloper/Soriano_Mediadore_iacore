import { TestBed } from '@angular/core/testing';
import { Router, UrlTree } from '@angular/router';
import { AuthGuard } from './auth.guard';
import { AuthService } from '../services/auth.service';

describe('AuthGuard', () => {
  let guard: AuthGuard;
  let authService: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;
  let mockUrlTree: UrlTree;

  beforeEach(() => {
    const authServiceSpy = jasmine.createSpyObj('AuthService', ['isAuthenticated', 'autoRenewToken']);
    const routerSpy = jasmine.createSpyObj('Router', ['createUrlTree']);

    TestBed.configureTestingModule({
      providers: [
        AuthGuard,
        { provide: AuthService, useValue: authServiceSpy },
        { provide: Router, useValue: routerSpy }
      ]
    });

    guard = TestBed.inject(AuthGuard);
    authService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;

    // Create a mock UrlTree
    mockUrlTree = new UrlTree();
    router.createUrlTree.and.returnValue(mockUrlTree);
  });

  it('should be created', () => {
    expect(guard).toBeTruthy();
  });

  describe('canActivate', () => {
    it('should allow access when user is authenticated', () => {
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate();

      expect(result).toBe(true);
      expect(authService.isAuthenticated).toHaveBeenCalled();
      expect(authService.autoRenewToken).toHaveBeenCalled();
    });

    it('should auto-renew token when user is authenticated', () => {
      authService.isAuthenticated.and.returnValue(true);

      guard.canActivate();

      expect(authService.autoRenewToken).toHaveBeenCalled();
    });

    it('should redirect to login when user is not authenticated', () => {
      authService.isAuthenticated.and.returnValue(false);
      spyOn(console, 'warn');

      const result = guard.canActivate();

      expect(result).toBe(mockUrlTree);
      expect(router.createUrlTree).toHaveBeenCalledWith(['/login']);
      expect(console.warn).toHaveBeenCalledWith('🚫 Acceso denegado - Redirigiendo a login');
    });

    it('should not call autoRenewToken when user is not authenticated', () => {
      authService.isAuthenticated.and.returnValue(false);

      guard.canActivate();

      expect(authService.autoRenewToken).not.toHaveBeenCalled();
    });

    it('should check authentication status before allowing access', () => {
      authService.isAuthenticated.and.returnValue(true);

      guard.canActivate();

      expect(authService.isAuthenticated).toHaveBeenCalledBefore(authService.autoRenewToken);
    });
  });

  describe('canActivateChild', () => {
    it('should delegate to canActivate when user is authenticated', () => {
      authService.isAuthenticated.and.returnValue(true);
      spyOn(guard, 'canActivate').and.returnValue(true);

      const result = guard.canActivateChild();

      expect(guard.canActivate).toHaveBeenCalled();
      expect(result).toBe(true);
    });

    it('should delegate to canActivate when user is not authenticated', () => {
      authService.isAuthenticated.and.returnValue(false);
      spyOn(guard, 'canActivate').and.returnValue(mockUrlTree);

      const result = guard.canActivateChild();

      expect(guard.canActivate).toHaveBeenCalled();
      expect(result).toBe(mockUrlTree);
    });

    it('should return same result as canActivate', () => {
      authService.isAuthenticated.and.returnValue(true);

      const canActivateResult = guard.canActivate();
      const canActivateChildResult = guard.canActivateChild();

      expect(canActivateChildResult).toEqual(canActivateResult);
    });
  });

  describe('Multiple calls', () => {
    it('should check authentication on each call', () => {
      authService.isAuthenticated.and.returnValue(true);

      guard.canActivate();
      guard.canActivate();
      guard.canActivate();

      expect(authService.isAuthenticated).toHaveBeenCalledTimes(3);
      expect(authService.autoRenewToken).toHaveBeenCalledTimes(3);
    });

    it('should handle authentication state changes', () => {
      // First call: authenticated
      authService.isAuthenticated.and.returnValue(true);
      let result1 = guard.canActivate();
      expect(result1).toBe(true);

      // Second call: not authenticated
      authService.isAuthenticated.and.returnValue(false);
      let result2 = guard.canActivate();
      expect(result2).toBe(mockUrlTree);

      // Third call: authenticated again
      authService.isAuthenticated.and.returnValue(true);
      let result3 = guard.canActivate();
      expect(result3).toBe(true);
    });
  });

  describe('Return types', () => {
    it('should return boolean true for authenticated users', () => {
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate();

      expect(typeof result).toBe('boolean');
      expect(result).toBe(true);
    });

    it('should return UrlTree for unauthenticated users', () => {
      authService.isAuthenticated.and.returnValue(false);

      const result = guard.canActivate();

      expect(result).toBeInstanceOf(UrlTree);
      expect(result).toBe(mockUrlTree);
    });
  });
});
