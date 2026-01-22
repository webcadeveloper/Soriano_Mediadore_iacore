import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { AuthService, User, UserRole } from './auth.service';
import { SecureStorageService } from './secure-storage.service';

describe('AuthService', () => {
  let service: AuthService;
  let router: jasmine.SpyObj<Router>;
  let secureStorage: jasmine.SpyObj<SecureStorageService>;

  beforeEach(() => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    const storageSpy = jasmine.createSpyObj('SecureStorageService', ['getItem', 'setItem', 'removeItem']);

    TestBed.configureTestingModule({
      providers: [
        AuthService,
        { provide: Router, useValue: routerSpy },
        { provide: SecureStorageService, useValue: storageSpy }
      ]
    });

    service = TestBed.inject(AuthService);
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    secureStorage = TestBed.inject(SecureStorageService) as jasmine.SpyObj<SecureStorageService>;
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('login', () => {
    it('should login successfully with valid credentials', (done) => {
      service.login('admin', 'admin123').subscribe({
        next: (response) => {
          expect(response.success).toBe(true);
          expect(response.user).toBeDefined();
          expect(response.user?.username).toBe('admin');
          expect(response.user?.role).toBe('admin' as UserRole);
          expect(response.token).toBeDefined();
          done();
        }
      });
    });

    it('should fail login with invalid credentials', (done) => {
      service.login('admin', 'wrongpassword').subscribe({
        next: (response) => {
          expect(response.success).toBe(false);
          expect(response.error).toContain('incorrectas');
          expect(response.user).toBeUndefined();
          done();
        }
      });
    });

    it('should store token and user after successful login', (done) => {
      service.login('admin', 'admin123').subscribe({
        next: () => {
          expect(secureStorage.setItem).toHaveBeenCalledWith('auth_token', jasmine.any(String));
          expect(secureStorage.setItem).toHaveBeenCalledWith('current_user', jasmine.any(Object));
          done();
        }
      });
    });

    it('should emit user through observable after login', (done) => {
      service.currentUser$.subscribe(user => {
        if (user) {
          expect(user.username).toBe('admin');
          done();
        }
      });

      service.login('admin', 'admin123').subscribe();
    });
  });

  describe('logout', () => {
    beforeEach((done) => {
      service.login('admin', 'admin123').subscribe(() => done());
    });

    it('should clear storage on logout', () => {
      service.logout();
      expect(secureStorage.removeItem).toHaveBeenCalledWith('auth_token');
      expect(secureStorage.removeItem).toHaveBeenCalledWith('current_user');
    });

    it('should navigate to login on logout', () => {
      service.logout();
      expect(router.navigate).toHaveBeenCalledWith(['/login']);
    });

    it('should emit null user after logout', (done) => {
      service.currentUser$.subscribe(user => {
        if (user === null) {
          done();
        }
      });

      service.logout();
    });
  });

  describe('isAuthenticated', () => {
    it('should return false when not authenticated', () => {
      secureStorage.getItem.and.returnValue(null);
      expect(service.isAuthenticated()).toBe(false);
    });

    it('should return true when authenticated', (done) => {
      service.login('admin', 'admin123').subscribe(() => {
        secureStorage.getItem.and.returnValue('mock-token');
        expect(service.isAuthenticated()).toBe(true);
        done();
      });
    });
  });

  describe('hasRole', () => {
    beforeEach((done) => {
      service.login('admin', 'admin123').subscribe(() => done());
    });

    it('should return true when user has the specified role', () => {
      expect(service.hasRole('admin')).toBe(true);
    });

    it('should return false when user does not have the specified role', () => {
      expect(service.hasRole('agente')).toBe(false);
    });
  });

  describe('hasAnyRole', () => {
    beforeEach((done) => {
      service.login('supervisor', 'supervisor123').subscribe(() => done());
    });

    it('should return true when user has any of the specified roles', () => {
      expect(service.hasAnyRole(['admin', 'supervisor'])).toBe(true);
    });

    it('should return false when user has none of the specified roles', () => {
      expect(service.hasAnyRole(['admin', 'agente'])).toBe(false);
    });

    it('should return false when roles array is empty', () => {
      expect(service.hasAnyRole([])).toBe(false);
    });
  });

  describe('getToken', () => {
    it('should return null when no token exists', () => {
      secureStorage.getItem.and.returnValue(null);
      expect(service.getToken()).toBeNull();
    });

    it('should return token when exists', (done) => {
      service.login('admin', 'admin123').subscribe(() => {
        const token = service.getToken();
        expect(token).toBeDefined();
        expect(token).toContain('Bearer ');
        done();
      });
    });
  });

  describe('getCurrentUser', () => {
    it('should return null when no user is logged in', () => {
      expect(service.getCurrentUser()).toBeNull();
    });

    it('should return current user when logged in', (done) => {
      service.login('admin', 'admin123').subscribe(() => {
        const user = service.getCurrentUser();
        expect(user).toBeDefined();
        expect(user?.username).toBe('admin');
        done();
      });
    });
  });

  describe('refreshToken', () => {
    beforeEach((done) => {
      service.login('admin', 'admin123').subscribe(() => done());
    });

    it('should refresh token successfully', (done) => {
      service.refreshToken().subscribe({
        next: (response) => {
          expect(response.success).toBe(true);
          expect(response.token).toBeDefined();
          done();
        }
      });
    });

    it('should update stored token after refresh', (done) => {
      const callCount = secureStorage.setItem.calls.count();

      service.refreshToken().subscribe({
        next: () => {
          expect(secureStorage.setItem.calls.count()).toBeGreaterThan(callCount);
          done();
        }
      });
    });
  });

  describe('Demo users', () => {
    const testUsers = [
      { username: 'admin', password: 'admin123', role: 'admin' },
      { username: 'agente', password: 'agente123', role: 'agente' },
      { username: 'supervisor', password: 'supervisor123', role: 'supervisor' },
      { username: 'director', password: 'director123', role: 'director' },
      { username: 'auditor', password: 'auditor123', role: 'auditor' }
    ];

    testUsers.forEach(({ username, password, role }) => {
      it(`should login ${role} user successfully`, (done) => {
        service.login(username, password).subscribe({
          next: (response) => {
            expect(response.success).toBe(true);
            expect(response.user?.role).toBe(role as UserRole);
            done();
          }
        });
      });
    });
  });
});
