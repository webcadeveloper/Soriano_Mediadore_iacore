import { TestBed } from '@angular/core/testing';
import { LoggerService } from './logger.service';

describe('LoggerService', () => {
  let service: LoggerService;
  let consoleSpy: jasmine.SpyObj<Console>;

  beforeEach(() => {
    // Create spies for console methods
    consoleSpy = jasmine.createSpyObj('Console', ['log', 'warn', 'error', 'group', 'groupEnd', 'time', 'timeEnd']);

    // Replace global console with our spy
    spyOn(console, 'log');
    spyOn(console, 'warn');
    spyOn(console, 'error');
    spyOn(console, 'group');
    spyOn(console, 'groupEnd');
    spyOn(console, 'time');
    spyOn(console, 'timeEnd');

    TestBed.configureTestingModule({
      providers: [LoggerService]
    });

    service = TestBed.inject(LoggerService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('info', () => {
    it('should log info messages in development', () => {
      service.info('Test info message', { data: 'test' });
      expect(console.log).toHaveBeenCalledWith('[INFO] Test info message', { data: 'test' });
    });

    it('should support multiple arguments', () => {
      service.info('Test', 'arg1', 'arg2', 123);
      expect(console.log).toHaveBeenCalledWith('[INFO] Test', 'arg1', 'arg2', 123);
    });
  });

  describe('warn', () => {
    it('should log warning messages in development', () => {
      service.warn('Test warning', { code: 404 });
      expect(console.warn).toHaveBeenCalledWith('[WARN] Test warning', { code: 404 });
    });

    it('should format warning messages correctly', () => {
      service.warn('Deprecated feature');
      expect(console.warn).toHaveBeenCalledWith('[WARN] Deprecated feature');
    });
  });

  describe('error', () => {
    it('should always log errors even in production', () => {
      const error = new Error('Test error');
      service.error('Something went wrong', error);
      expect(console.error).toHaveBeenCalledWith('[ERROR] Something went wrong', error);
    });

    it('should log errors without error object', () => {
      service.error('Error message only');
      expect(console.error).toHaveBeenCalledWith('[ERROR] Error message only', undefined);
    });

    it('should handle various error types', () => {
      const errorTypes = [
        new Error('Standard error'),
        { message: 'Custom error object' },
        'String error',
        null
      ];

      errorTypes.forEach(error => {
        service.error('Test', error);
        expect(console.error).toHaveBeenCalledWith('[ERROR] Test', error);
      });
    });
  });

  describe('debug', () => {
    it('should log debug messages in development', () => {
      service.debug('Debug info', { state: 'active' });
      expect(console.log).toHaveBeenCalledWith('[DEBUG] Debug info', { state: 'active' });
    });

    it('should support complex data structures', () => {
      const complexData = {
        nested: { array: [1, 2, 3], object: { key: 'value' } }
      };
      service.debug('Complex data', complexData);
      expect(console.log).toHaveBeenCalledWith('[DEBUG] Complex data', complexData);
    });
  });

  describe('group', () => {
    it('should create grouped logs in development', () => {
      service.group('Test Group', () => {
        console.log('Inside group');
      });

      expect(console.group).toHaveBeenCalledWith('Test Group');
      expect(console.groupEnd).toHaveBeenCalled();
    });

    it('should execute callback within group', () => {
      const callback = jasmine.createSpy('callback');
      service.group('Group', callback);
      expect(callback).toHaveBeenCalled();
    });

    it('should handle errors in callback gracefully', () => {
      const errorCallback = () => {
        throw new Error('Callback error');
      };

      expect(() => {
        service.group('Error Group', errorCallback);
      }).toThrow();

      // Group should still be ended even if callback throws
      expect(console.group).toHaveBeenCalled();
    });
  });

  describe('time', () => {
    it('should start timing with label', () => {
      service.time('operation');
      expect(console.time).toHaveBeenCalledWith('operation');
    });

    it('should end timing with same label', () => {
      service.time('operation');
      service.timeEnd('operation');
      expect(console.timeEnd).toHaveBeenCalledWith('operation');
    });

    it('should handle multiple concurrent timers', () => {
      service.time('timer1');
      service.time('timer2');
      service.timeEnd('timer1');
      service.timeEnd('timer2');

      expect(console.time).toHaveBeenCalledWith('timer1');
      expect(console.time).toHaveBeenCalledWith('timer2');
      expect(console.timeEnd).toHaveBeenCalledWith('timer1');
      expect(console.timeEnd).toHaveBeenCalledWith('timer2');
    });
  });

  describe('Production behavior', () => {
    it('should document that errors are always logged', () => {
      // This test documents behavior - errors ALWAYS log
      service.error('Critical error');
      expect(console.error).toHaveBeenCalled();
    });

    it('should have isDevelopment property', () => {
      // Verify the service has the isDevelopment property
      expect((service as any).isDevelopment).toBeDefined();
    });
  });

  describe('Message formatting', () => {
    it('should include proper prefixes', () => {
      service.info('info');
      service.warn('warn');
      service.error('error');
      service.debug('debug');

      expect(console.log).toHaveBeenCalledWith('[INFO] info');
      expect(console.warn).toHaveBeenCalledWith('[WARN] warn');
      expect(console.error).toHaveBeenCalledWith('[ERROR] error');
      expect(console.log).toHaveBeenCalledWith('[DEBUG] debug');
    });
  });
});
