import { TestBed } from '@angular/core/testing';
import { Router, UrlTree, ActivatedRouteSnapshot } from '@angular/router';
import { RoleGuard } from './role.guard';
import { AuthService, UserRole } from '../services/auth.service';

describe('RoleGuard', () => {
  let guard: RoleGuard;
  let authService: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;
  let mockUrlTree: UrlTree;
  let mockRoute: ActivatedRouteSnapshot;

  beforeEach(() => {
    const authServiceSpy = jasmine.createSpyObj('AuthService', ['isAuthenticated', 'hasAnyRole']);
    const routerSpy = jasmine.createSpyObj('Router', ['createUrlTree']);

    TestBed.configureTestingModule({
      providers: [
        RoleGuard,
        { provide: AuthService, useValue: authServiceSpy },
        { provide: Router, useValue: routerSpy }
      ]
    });

    guard = TestBed.inject(RoleGuard);
    authService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;

    mockUrlTree = new UrlTree();
    router.createUrlTree.and.returnValue(mockUrlTree);

    // Create mock route
    mockRoute = new ActivatedRouteSnapshot();
  });

  it('should be created', () => {
    expect(guard).toBeTruthy();
  });

  describe('canActivate - No roles specified', () => {
    it('should allow access if authenticated and no roles required', () => {
      mockRoute.data = {};
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
      expect(authService.isAuthenticated).toHaveBeenCalled();
      expect(authService.hasAnyRole).not.toHaveBeenCalled();
    });

    it('should redirect to login if not authenticated and no roles required', () => {
      mockRoute.data = {};
      authService.isAuthenticated.and.returnValue(false);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(mockUrlTree);
      expect(router.createUrlTree).toHaveBeenCalledWith(['/login']);
    });

    it('should allow access when roles array is empty', () => {
      mockRoute.data = { roles: [] };
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });
  });

  describe('canActivate - With required roles', () => {
    it('should allow access when user has required role', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
      expect(authService.hasAnyRole).toHaveBeenCalledWith(['admin']);
    });

    it('should allow access when user has any of the required roles', () => {
      mockRoute.data = { roles: ['admin', 'supervisor'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
      expect(authService.hasAnyRole).toHaveBeenCalledWith(['admin', 'supervisor']);
    });

    it('should redirect to login when not authenticated', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(false);
      spyOn(console, 'warn');

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(mockUrlTree);
      expect(router.createUrlTree).toHaveBeenCalledWith(['/login']);
      expect(console.warn).toHaveBeenCalledWith('🚫 Usuario no autenticado - Redirigiendo a login');
      expect(authService.hasAnyRole).not.toHaveBeenCalled();
    });

    it('should redirect to dashboard when authenticated but lacks required role', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(false);
      spyOn(console, 'warn');

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(mockUrlTree);
      expect(router.createUrlTree).toHaveBeenCalledWith(['/dashboard']);
      expect(console.warn).toHaveBeenCalledWith('🚫 Acceso denegado - Rol insuficiente');
    });
  });

  describe('canActivate - Different role scenarios', () => {
    it('should handle admin role correctly', () => {
      mockRoute.data = { roles: ['admin' as UserRole] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle agente role correctly', () => {
      mockRoute.data = { roles: ['agente' as UserRole] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle supervisor role correctly', () => {
      mockRoute.data = { roles: ['supervisor' as UserRole] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle director role correctly', () => {
      mockRoute.data = { roles: ['director' as UserRole] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle auditor role correctly', () => {
      mockRoute.data = { roles: ['auditor' as UserRole] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle multiple roles correctly', () => {
      mockRoute.data = { roles: ['admin', 'director', 'supervisor'] as UserRole[] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
      expect(authService.hasAnyRole).toHaveBeenCalledWith(['admin', 'director', 'supervisor']);
    });
  });

  describe('canActivateChild', () => {
    it('should delegate to canActivate with roles', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);
      spyOn(guard, 'canActivate').and.returnValue(true);

      const result = guard.canActivateChild(mockRoute);

      expect(guard.canActivate).toHaveBeenCalledWith(mockRoute);
      expect(result).toBe(true);
    });

    it('should delegate to canActivate without roles', () => {
      mockRoute.data = {};
      authService.isAuthenticated.and.returnValue(true);
      spyOn(guard, 'canActivate').and.returnValue(true);

      const result = guard.canActivateChild(mockRoute);

      expect(guard.canActivate).toHaveBeenCalledWith(mockRoute);
      expect(result).toBe(true);
    });

    it('should return same result as canActivate', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const canActivateResult = guard.canActivate(mockRoute);
      const canActivateChildResult = guard.canActivateChild(mockRoute);

      expect(canActivateChildResult).toEqual(canActivateResult);
    });
  });

  describe('Edge cases', () => {
    it('should check authentication before checking roles', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(false);

      guard.canActivate(mockRoute);

      expect(authService.isAuthenticated).toHaveBeenCalled();
      expect(authService.hasAnyRole).not.toHaveBeenCalled();
    });

    it('should handle null roles data', () => {
      mockRoute.data = { roles: null };
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });

    it('should handle undefined roles data', () => {
      mockRoute.data = { roles: undefined };
      authService.isAuthenticated.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(result).toBe(true);
    });
  });

  describe('Return types', () => {
    it('should return boolean true when access is granted', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);

      const result = guard.canActivate(mockRoute);

      expect(typeof result).toBe('boolean');
      expect(result).toBe(true);
    });

    it('should return UrlTree when redirecting to login', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(false);

      const result = guard.canActivate(mockRoute);

      expect(result).toBeInstanceOf(UrlTree);
    });

    it('should return UrlTree when redirecting to dashboard', () => {
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(false);

      const result = guard.canActivate(mockRoute);

      expect(result).toBeInstanceOf(UrlTree);
    });
  });

  describe('Multiple guard calls', () => {
    it('should evaluate each route independently', () => {
      // First route: admin only
      mockRoute.data = { roles: ['admin'] };
      authService.isAuthenticated.and.returnValue(true);
      authService.hasAnyRole.and.returnValue(true);
      expect(guard.canActivate(mockRoute)).toBe(true);

      // Second route: supervisor only
      mockRoute.data = { roles: ['supervisor'] };
      authService.hasAnyRole.and.returnValue(false);
      expect(guard.canActivate(mockRoute)).toBe(mockUrlTree);

      // Third route: no roles
      mockRoute.data = {};
      expect(guard.canActivate(mockRoute)).toBe(true);
    });
  });
});
