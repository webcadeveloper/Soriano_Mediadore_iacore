import { TestBed } from '@angular/core/testing';
import { SecureStorageService } from './secure-storage.service';
import { SecurityService } from './security.service';

describe('SecureStorageService', () => {
  let service: SecureStorageService;
  let securityService: jasmine.SpyObj<SecurityService>;
  let localStorageSpy: jasmine.SpyObj<Storage>;
  let sessionStorageSpy: jasmine.SpyObj<Storage>;

  beforeEach(() => {
    // Create spy for SecurityService
    const securitySpy = jasmine.createSpyObj('SecurityService', ['encrypt', 'decrypt']);

    // Create spies for localStorage and sessionStorage
    localStorageSpy = jasmine.createSpyObj('localStorage', ['getItem', 'setItem', 'removeItem', 'clear']);
    sessionStorageSpy = jasmine.createSpyObj('sessionStorage', ['getItem', 'setItem', 'removeItem', 'clear']);

    // Replace global storage objects
    spyOnProperty(window, 'localStorage', 'get').and.returnValue(localStorageSpy);
    spyOnProperty(window, 'sessionStorage', 'get').and.returnValue(sessionStorageSpy);

    TestBed.configureTestingModule({
      providers: [
        SecureStorageService,
        { provide: SecurityService, useValue: securitySpy }
      ]
    });

    service = TestBed.inject(SecureStorageService);
    securityService = TestBed.inject(SecurityService) as jasmine.SpyObj<SecurityService>;
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('setItem', () => {
    it('should encrypt and store data in localStorage', () => {
      const testData = { username: 'testuser', role: 'admin' };
      const encrypted = 'encrypted_data';

      securityService.encrypt.and.returnValue(encrypted);

      service.setItem('user', testData);

      expect(securityService.encrypt).toHaveBeenCalledWith(JSON.stringify(testData));
      expect(localStorageSpy.setItem).toHaveBeenCalledWith('user', encrypted);
    });

    it('should handle primitive types', () => {
      securityService.encrypt.and.returnValue('encrypted');

      service.setItem('count', 42);
      expect(securityService.encrypt).toHaveBeenCalledWith('42');

      service.setItem('flag', true);
      expect(securityService.encrypt).toHaveBeenCalledWith('true');

      service.setItem('name', 'John');
      expect(securityService.encrypt).toHaveBeenCalledWith('"John"');
    });

    it('should handle arrays', () => {
      const testArray = [1, 2, 3, 4, 5];
      securityService.encrypt.and.returnValue('encrypted_array');

      service.setItem('numbers', testArray);

      expect(securityService.encrypt).toHaveBeenCalledWith(JSON.stringify(testArray));
      expect(localStorageSpy.setItem).toHaveBeenCalled();
    });

    it('should handle errors gracefully', () => {
      securityService.encrypt.and.throwError('Encryption failed');
      spyOn(console, 'error');

      service.setItem('key', 'value');

      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('getItem', () => {
    it('should retrieve and decrypt data from localStorage', () => {
      const encrypted = 'encrypted_data';
      const decrypted = '{"username":"testuser","role":"admin"}';
      const expected = { username: 'testuser', role: 'admin' };

      localStorageSpy.getItem.and.returnValue(encrypted);
      securityService.decrypt.and.returnValue(decrypted);

      const result = service.getItem('user');

      expect(localStorageSpy.getItem).toHaveBeenCalledWith('user');
      expect(securityService.decrypt).toHaveBeenCalledWith(encrypted);
      expect(result).toEqual(expected);
    });

    it('should return null if key does not exist', () => {
      localStorageSpy.getItem.and.returnValue(null);

      const result = service.getItem('nonexistent');

      expect(result).toBeNull();
      expect(securityService.decrypt).not.toHaveBeenCalled();
    });

    it('should return null if decryption fails', () => {
      localStorageSpy.getItem.and.returnValue('encrypted');
      securityService.decrypt.and.returnValue('');

      const result = service.getItem('corrupted');

      expect(result).toBeNull();
    });

    it('should handle JSON parse errors gracefully', () => {
      localStorageSpy.getItem.and.returnValue('encrypted');
      securityService.decrypt.and.returnValue('invalid json {');
      spyOn(console, 'error');

      const result = service.getItem('invalid');

      expect(result).toBeNull();
      expect(console.error).toHaveBeenCalled();
    });

    it('should handle various data types', () => {
      localStorageSpy.getItem.and.returnValue('encrypted');

      // Number
      securityService.decrypt.and.returnValue('42');
      expect(service.getItem('number')).toBe(42);

      // Boolean
      securityService.decrypt.and.returnValue('true');
      expect(service.getItem('boolean')).toBe(true);

      // Array
      securityService.decrypt.and.returnValue('[1,2,3]');
      expect(service.getItem('array')).toEqual([1, 2, 3]);
    });
  });

  describe('removeItem', () => {
    it('should remove item from localStorage', () => {
      service.removeItem('test_key');

      expect(localStorageSpy.removeItem).toHaveBeenCalledWith('test_key');
    });
  });

  describe('clear', () => {
    it('should clear all items from localStorage', () => {
      service.clear();

      expect(localStorageSpy.clear).toHaveBeenCalled();
    });
  });

  describe('hasItem', () => {
    it('should return true if item exists', () => {
      localStorageSpy.getItem.and.returnValue('some_data');

      expect(service.hasItem('existing_key')).toBe(true);
      expect(localStorageSpy.getItem).toHaveBeenCalledWith('existing_key');
    });

    it('should return false if item does not exist', () => {
      localStorageSpy.getItem.and.returnValue(null);

      expect(service.hasItem('nonexistent_key')).toBe(false);
    });
  });

  describe('setSessionItem', () => {
    it('should encrypt and store data in sessionStorage', () => {
      const testData = { temp: 'data' };
      const encrypted = 'encrypted_session_data';

      securityService.encrypt.and.returnValue(encrypted);

      service.setSessionItem('session_key', testData);

      expect(securityService.encrypt).toHaveBeenCalledWith(JSON.stringify(testData));
      expect(sessionStorageSpy.setItem).toHaveBeenCalledWith('session_key', encrypted);
    });

    it('should handle errors gracefully', () => {
      securityService.encrypt.and.throwError('Encryption failed');
      spyOn(console, 'error');

      service.setSessionItem('key', 'value');

      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('getSessionItem', () => {
    it('should retrieve and decrypt data from sessionStorage', () => {
      const encrypted = 'encrypted_session_data';
      const decrypted = '{"temp":"data"}';
      const expected = { temp: 'data' };

      sessionStorageSpy.getItem.and.returnValue(encrypted);
      securityService.decrypt.and.returnValue(decrypted);

      const result = service.getSessionItem('session_key');

      expect(sessionStorageSpy.getItem).toHaveBeenCalledWith('session_key');
      expect(securityService.decrypt).toHaveBeenCalledWith(encrypted);
      expect(result).toEqual(expected);
    });

    it('should return null if key does not exist', () => {
      sessionStorageSpy.getItem.and.returnValue(null);

      const result = service.getSessionItem('nonexistent');

      expect(result).toBeNull();
      expect(securityService.decrypt).not.toHaveBeenCalled();
    });

    it('should handle errors gracefully', () => {
      sessionStorageSpy.getItem.and.returnValue('encrypted');
      securityService.decrypt.and.throwError('Decryption failed');
      spyOn(console, 'error');

      const result = service.getSessionItem('key');

      expect(result).toBeNull();
      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('migrateUnencryptedData', () => {
    it('should migrate plain JSON data to encrypted format', () => {
      const plainData = '{"key":"value"}';
      const encrypted = 'encrypted_migrated_data';

      localStorageSpy.getItem.and.returnValue(plainData);
      securityService.encrypt.and.returnValue(encrypted);
      spyOn(console, 'log');

      service.migrateUnencryptedData('legacy_key');

      expect(localStorageSpy.getItem).toHaveBeenCalledWith('legacy_key');
      expect(securityService.encrypt).toHaveBeenCalled();
      expect(localStorageSpy.setItem).toHaveBeenCalledWith('legacy_key', encrypted);
      expect(console.log).toHaveBeenCalledWith('✅ Datos migrados y cifrados: legacy_key');
    });

    it('should skip if key does not exist', () => {
      localStorageSpy.getItem.and.returnValue(null);

      service.migrateUnencryptedData('nonexistent');

      expect(securityService.encrypt).not.toHaveBeenCalled();
    });

    it('should warn on invalid JSON format', () => {
      localStorageSpy.getItem.and.returnValue('invalid json {');
      spyOn(console, 'warn');

      service.migrateUnencryptedData('invalid_key');

      expect(console.warn).toHaveBeenCalledWith('⚠️ No se pudo migrar invalid_key: formato inválido');
      expect(securityService.encrypt).not.toHaveBeenCalled();
    });

    it('should handle migration errors', () => {
      localStorageSpy.getItem.and.throwError('Storage error');
      spyOn(console, 'error');

      service.migrateUnencryptedData('error_key');

      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('migrateAllKnownKeys', () => {
    it('should migrate all known keys', () => {
      spyOn(service, 'migrateUnencryptedData');
      spyOn(console, 'log');

      service.migrateAllKnownKeys();

      expect(service.migrateUnencryptedData).toHaveBeenCalledWith('recobros_recibos');
      expect(service.migrateUnencryptedData).toHaveBeenCalledWith('recobros_templates');
      expect(service.migrateUnencryptedData).toHaveBeenCalledWith('recobros_config');
      expect(service.migrateUnencryptedData).toHaveBeenCalledWith('soriano_email_templates');
      expect(console.log).toHaveBeenCalledWith('🔄 Iniciando migración de datos a formato cifrado...');
      expect(console.log).toHaveBeenCalledWith('✅ Migración completada');
    });
  });

  describe('Complex data structures', () => {
    it('should handle nested objects', () => {
      const complexData = {
        user: {
          profile: {
            name: 'John',
            settings: {
              theme: 'dark',
              notifications: true
            }
          }
        }
      };

      const encrypted = 'encrypted_complex';
      const decrypted = JSON.stringify(complexData);

      securityService.encrypt.and.returnValue(encrypted);
      localStorageSpy.getItem.and.returnValue(encrypted);
      securityService.decrypt.and.returnValue(decrypted);

      service.setItem('complex', complexData);
      const result = service.getItem('complex');

      expect(result).toEqual(complexData);
    });

    it('should handle arrays of objects', () => {
      const arrayData = [
        { id: 1, name: 'Item 1' },
        { id: 2, name: 'Item 2' },
        { id: 3, name: 'Item 3' }
      ];

      const encrypted = 'encrypted_array';
      const decrypted = JSON.stringify(arrayData);

      securityService.encrypt.and.returnValue(encrypted);
      localStorageSpy.getItem.and.returnValue(encrypted);
      securityService.decrypt.and.returnValue(decrypted);

      service.setItem('items', arrayData);
      const result = service.getItem('items');

      expect(result).toEqual(arrayData);
    });
  });
});
