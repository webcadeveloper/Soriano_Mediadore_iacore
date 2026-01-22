import { TestBed } from '@angular/core/testing';
import { DomSanitizer } from '@angular/platform-browser';
import { SecurityService } from './security.service';

describe('SecurityService', () => {
  let service: SecurityService;
  let sanitizer: DomSanitizer;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SecurityService);
    sanitizer = TestBed.inject(DomSanitizer);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('escapeHtml', () => {
    it('should escape HTML special characters', () => {
      const input = '<script>alert("XSS")</script>';
      const output = service.escapeHtml(input);
      expect(output).not.toContain('<script>');
      expect(output).toContain('&lt;script&gt;');
    });

    it('should escape quotes', () => {
      const input = 'Test "quoted" text';
      const output = service.escapeHtml(input);
      expect(output).toContain('&quot;');
    });

    it('should handle empty string', () => {
      expect(service.escapeHtml('')).toBe('');
    });

    it('should handle null and undefined', () => {
      expect(service.escapeHtml(null as any)).toBe('');
      expect(service.escapeHtml(undefined as any)).toBe('');
    });
  });

  describe('validateHtml', () => {
    it('should reject HTML with script tags', () => {
      const result = service.validateHtml('<script>alert("XSS")</script>');
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should reject HTML with iframe tags', () => {
      const result = service.validateHtml('<iframe src="evil.com"></iframe>');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Contiene etiquetas <iframe> no permitidas');
    });

    it('should reject HTML with event handlers', () => {
      const result = service.validateHtml('<div onclick="alert()">Click</div>');
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should reject javascript: URLs', () => {
      const result = service.validateHtml('<a href="javascript:alert()">Link</a>');
      expect(result.valid).toBe(false);
    });

    it('should accept safe HTML', () => {
      const result = service.validateHtml('<div><p>Safe content</p></div>');
      expect(result.valid).toBe(true);
      expect(result.errors.length).toBe(0);
    });

    it('should accept empty HTML', () => {
      const result = service.validateHtml('');
      expect(result.valid).toBe(true);
    });
  });

  describe('validateIBAN', () => {
    it('should validate correct Spanish IBAN', () => {
      expect(service.validateIBAN('ES9121000418450200051332')).toBe(true);
    });

    it('should reject invalid IBAN format', () => {
      expect(service.validateIBAN('ES00123')).toBe(false);
    });

    it('should reject IBAN with invalid checksum', () => {
      expect(service.validateIBAN('ES9121000418450200051333')).toBe(false);
    });

    it('should handle empty or null IBAN', () => {
      expect(service.validateIBAN('')).toBe(false);
      expect(service.validateIBAN(null as any)).toBe(false);
      expect(service.validateIBAN(undefined as any)).toBe(false);
    });

    it('should normalize IBAN (remove spaces)', () => {
      expect(service.validateIBAN('ES91 2100 0418 4502 0005 1332')).toBe(true);
    });

    it('should reject IBAN with invalid characters', () => {
      expect(service.validateIBAN('ES91-2100-0418')).toBe(false);
    });
  });

  describe('isHttps', () => {
    it('should return true for HTTPS URLs', () => {
      expect(service.isHttps('https://example.com')).toBe(true);
    });

    it('should return false for HTTP URLs', () => {
      expect(service.isHttps('http://example.com')).toBe(false);
    });

    it('should return false for empty or invalid URLs', () => {
      expect(service.isHttps('')).toBe(false);
      expect(service.isHttps(null as any)).toBe(false);
      expect(service.isHttps('not-a-url')).toBe(false);
    });
  });

  describe('sanitizeEmail', () => {
    it('should sanitize valid email', () => {
      const result = service.sanitizeEmail('test@example.com');
      expect(result).toBe('test@example.com');
    });

    it('should remove invalid characters', () => {
      const result = service.sanitizeEmail('test<script>@example.com');
      expect(result).not.toContain('<');
      expect(result).not.toContain('>');
    });

    it('should handle empty email', () => {
      expect(service.sanitizeEmail('')).toBe('');
    });
  });

  describe('sanitizePhone', () => {
    it('should keep only valid phone characters', () => {
      const result = service.sanitizePhone('+34 123-456-789');
      expect(result).toMatch(/^[\d+\-() ]+$/);
    });

    it('should remove invalid characters', () => {
      const result = service.sanitizePhone('+34abc123<script>');
      expect(result).not.toContain('abc');
      expect(result).not.toContain('<');
    });

    it('should handle empty phone', () => {
      expect(service.sanitizePhone('')).toBe('');
    });
  });

  describe('validateFile', () => {
    it('should validate file with correct type and size', () => {
      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const result = service.validateFile(file, ['application/pdf'], 5);
      expect(result.valid).toBe(true);
    });

    it('should reject file with wrong type', () => {
      const file = new File(['content'], 'test.exe', { type: 'application/x-msdownload' });
      const result = service.validateFile(file, ['application/pdf'], 5);
      expect(result.valid).toBe(false);
      expect(result.error).toContain('Tipo de archivo no permitido');
    });

    it('should reject file exceeding size limit', () => {
      const largeContent = new Array(6 * 1024 * 1024).join('a'); // 6MB
      const file = new File([largeContent], 'large.pdf', { type: 'application/pdf' });
      const result = service.validateFile(file, ['application/pdf'], 5);
      expect(result.valid).toBe(false);
      expect(result.error).toContain('demasiado grande');
    });

    it('should accept file within size limit', () => {
      const content = new Array(1024 * 1024).join('a'); // 1MB
      const file = new File([content], 'small.pdf', { type: 'application/pdf' });
      const result = service.validateFile(file, ['application/pdf'], 5);
      expect(result.valid).toBe(true);
    });
  });

  describe('encrypt and decrypt', () => {
    it('should encrypt and decrypt data correctly', () => {
      const original = 'sensitive data';
      const encrypted = service.encrypt(original);
      const decrypted = service.decrypt(encrypted);

      expect(encrypted).not.toBe(original);
      expect(decrypted).toBe(original);
    });

    it('should handle empty string', () => {
      const encrypted = service.encrypt('');
      const decrypted = service.decrypt(encrypted);
      expect(decrypted).toBe('');
    });

    it('should handle special characters', () => {
      const original = 'Test with émojis 🚀 and symbols ñ@#$%';
      const encrypted = service.encrypt(original);
      const decrypted = service.decrypt(encrypted);
      expect(decrypted).toBe(original);
    });

    it('should return empty string for invalid encrypted data', () => {
      const decrypted = service.decrypt('invalid-base64-data!!!');
      expect(decrypted).toBe('');
    });
  });
});
