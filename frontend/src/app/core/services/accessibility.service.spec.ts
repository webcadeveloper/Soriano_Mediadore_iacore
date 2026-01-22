import { TestBed } from '@angular/core/testing';
import { AccessibilityService } from './accessibility.service';

describe('AccessibilityService', () => {
  let service: AccessibilityService;
  let mockLiveRegion: HTMLElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [AccessibilityService]
    });

    service = TestBed.inject(AccessibilityService);

    // Get reference to the live region created by the service
    mockLiveRegion = document.querySelector('.sr-only') as HTMLElement;
  });

  afterEach(() => {
    // Clean up the live region after each test
    if (mockLiveRegion && mockLiveRegion.parentNode) {
      mockLiveRegion.parentNode.removeChild(mockLiveRegion);
    }
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('Live Region Creation', () => {
    it('should create a live region on initialization', () => {
      expect(mockLiveRegion).toBeTruthy();
      expect(mockLiveRegion.getAttribute('role')).toBe('status');
      expect(mockLiveRegion.getAttribute('aria-live')).toBe('polite');
      expect(mockLiveRegion.getAttribute('aria-atomic')).toBe('true');
    });

    it('should position live region off-screen', () => {
      expect(mockLiveRegion.style.position).toBe('absolute');
      expect(mockLiveRegion.style.left).toBe('-10000px');
      expect(mockLiveRegion.style.width).toBe('1px');
      expect(mockLiveRegion.style.height).toBe('1px');
      expect(mockLiveRegion.style.overflow).toBe('hidden');
    });

    it('should have sr-only class', () => {
      expect(mockLiveRegion.className).toBe('sr-only');
    });
  });

  describe('announce', () => {
    beforeEach(() => {
      jasmine.clock().install();
    });

    afterEach(() => {
      jasmine.clock().uninstall();
    });

    it('should announce message with polite priority by default', () => {
      service.announce('Test announcement');

      expect(mockLiveRegion.textContent).toBe('Test announcement');
      expect(mockLiveRegion.getAttribute('aria-live')).toBe('polite');
    });

    it('should announce message with assertive priority', () => {
      service.announce('Urgent message', 'assertive');

      expect(mockLiveRegion.textContent).toBe('Urgent message');
      expect(mockLiveRegion.getAttribute('aria-live')).toBe('assertive');
    });

    it('should clear message after 5 seconds', () => {
      service.announce('Temporary message');
      expect(mockLiveRegion.textContent).toBe('Temporary message');

      jasmine.clock().tick(5000);
      expect(mockLiveRegion.textContent).toBe('');
    });

    it('should handle multiple announcements', () => {
      service.announce('First message');
      expect(mockLiveRegion.textContent).toBe('First message');

      jasmine.clock().tick(2000);
      service.announce('Second message');
      expect(mockLiveRegion.textContent).toBe('Second message');

      jasmine.clock().tick(5000);
      expect(mockLiveRegion.textContent).toBe('');
    });
  });

  describe('focusElement', () => {
    it('should focus element immediately when delay is 0', (done) => {
      const button = document.createElement('button');
      document.body.appendChild(button);
      spyOn(button, 'focus');

      service.focusElement(button, 0);

      setTimeout(() => {
        expect(button.focus).toHaveBeenCalled();
        document.body.removeChild(button);
        done();
      }, 10);
    });

    it('should focus element after delay', (done) => {
      const button = document.createElement('button');
      document.body.appendChild(button);
      spyOn(button, 'focus');

      service.focusElement(button, 100);

      setTimeout(() => {
        expect(button.focus).not.toHaveBeenCalled();
      }, 50);

      setTimeout(() => {
        expect(button.focus).toHaveBeenCalled();
        document.body.removeChild(button);
        done();
      }, 150);
    });

    it('should handle null element gracefully', () => {
      expect(() => {
        service.focusElement(null);
      }).not.toThrow();
    });
  });

  describe('focusFirstInteractive', () => {
    it('should focus first button', () => {
      const container = document.createElement('div');
      const button1 = document.createElement('button');
      const button2 = document.createElement('button');
      container.appendChild(button1);
      container.appendChild(button2);
      document.body.appendChild(container);

      spyOn(button1, 'focus');
      service.focusFirstInteractive(container);

      expect(button1.focus).toHaveBeenCalled();
      document.body.removeChild(container);
    });

    it('should focus first link', () => {
      const container = document.createElement('div');
      const link = document.createElement('a');
      link.href = '#';
      container.appendChild(link);
      document.body.appendChild(container);

      spyOn(link, 'focus');
      service.focusFirstInteractive(container);

      expect(link.focus).toHaveBeenCalled();
      document.body.removeChild(container);
    });

    it('should focus first input', () => {
      const container = document.createElement('div');
      const input = document.createElement('input');
      container.appendChild(input);
      document.body.appendChild(container);

      spyOn(input, 'focus');
      service.focusFirstInteractive(container);

      expect(input.focus).toHaveBeenCalled();
      document.body.removeChild(container);
    });

    it('should handle container with no focusable elements', () => {
      const container = document.createElement('div');
      const div = document.createElement('div');
      container.appendChild(div);
      document.body.appendChild(container);

      expect(() => {
        service.focusFirstInteractive(container);
      }).not.toThrow();

      document.body.removeChild(container);
    });
  });

  describe('getFocusableElements', () => {
    it('should find all focusable elements', () => {
      const container = document.createElement('div');
      const button = document.createElement('button');
      const link = document.createElement('a');
      link.href = '#';
      const input = document.createElement('input');
      const select = document.createElement('select');

      container.appendChild(button);
      container.appendChild(link);
      container.appendChild(input);
      container.appendChild(select);
      document.body.appendChild(container);

      const focusable = service.getFocusableElements(container);

      expect(focusable.length).toBe(4);
      expect(focusable).toContain(button);
      expect(focusable).toContain(link);
      expect(focusable).toContain(input);
      expect(focusable).toContain(select);

      document.body.removeChild(container);
    });

    it('should exclude disabled elements', () => {
      const container = document.createElement('div');
      const button = document.createElement('button');
      button.disabled = true;
      const input = document.createElement('input');
      input.disabled = true;

      container.appendChild(button);
      container.appendChild(input);
      document.body.appendChild(container);

      const focusable = service.getFocusableElements(container);

      expect(focusable.length).toBe(0);

      document.body.removeChild(container);
    });

    it('should exclude elements with tabindex="-1"', () => {
      const container = document.createElement('div');
      const div = document.createElement('div');
      div.setAttribute('tabindex', '-1');

      container.appendChild(div);
      document.body.appendChild(container);

      const focusable = service.getFocusableElements(container);

      expect(focusable.length).toBe(0);

      document.body.removeChild(container);
    });

    it('should include elements with positive tabindex', () => {
      const container = document.createElement('div');
      const div = document.createElement('div');
      div.setAttribute('tabindex', '0');

      container.appendChild(div);
      document.body.appendChild(container);

      const focusable = service.getFocusableElements(container);

      expect(focusable.length).toBe(1);
      expect(focusable[0]).toBe(div);

      document.body.removeChild(container);
    });
  });

  describe('trapFocus', () => {
    let container: HTMLElement;
    let firstButton: HTMLButtonElement;
    let lastButton: HTMLButtonElement;

    beforeEach(() => {
      container = document.createElement('div');
      firstButton = document.createElement('button');
      const middleButton = document.createElement('button');
      lastButton = document.createElement('button');

      container.appendChild(firstButton);
      container.appendChild(middleButton);
      container.appendChild(lastButton);
      document.body.appendChild(container);
    });

    afterEach(() => {
      document.body.removeChild(container);
    });

    it('should trap Tab at last element to first', () => {
      lastButton.focus();
      const event = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: false });
      spyOn(event, 'preventDefault');
      spyOn(firstButton, 'focus');

      service.trapFocus(container, event);

      expect(event.preventDefault).toHaveBeenCalled();
      expect(firstButton.focus).toHaveBeenCalled();
    });

    it('should trap Shift+Tab at first element to last', () => {
      firstButton.focus();
      const event = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true });
      spyOn(event, 'preventDefault');
      spyOn(lastButton, 'focus');

      service.trapFocus(container, event);

      expect(event.preventDefault).toHaveBeenCalled();
      expect(lastButton.focus).toHaveBeenCalled();
    });

    it('should not trap focus on non-Tab keys', () => {
      const event = new KeyboardEvent('keydown', { key: 'Enter' });
      spyOn(event, 'preventDefault');

      service.trapFocus(container, event);

      expect(event.preventDefault).not.toHaveBeenCalled();
    });

    it('should handle container with no focusable elements', () => {
      const emptyContainer = document.createElement('div');
      document.body.appendChild(emptyContainer);

      const event = new KeyboardEvent('keydown', { key: 'Tab' });

      expect(() => {
        service.trapFocus(emptyContainer, event);
      }).not.toThrow();

      document.body.removeChild(emptyContainer);
    });
  });

  describe('isInViewport', () => {
    it('should return true for element in viewport', () => {
      const element = document.createElement('div');
      element.style.position = 'absolute';
      element.style.top = '100px';
      element.style.left = '100px';
      element.style.width = '100px';
      element.style.height = '100px';
      document.body.appendChild(element);

      const result = service.isInViewport(element);

      expect(result).toBe(true);
      document.body.removeChild(element);
    });

    it('should return false for element above viewport', () => {
      const element = document.createElement('div');
      element.style.position = 'absolute';
      element.style.top = '-200px';
      element.style.left = '100px';
      document.body.appendChild(element);

      const result = service.isInViewport(element);

      expect(result).toBe(false);
      document.body.removeChild(element);
    });
  });

  describe('scrollToElement', () => {
    beforeEach(() => {
      jasmine.clock().install();
    });

    afterEach(() => {
      jasmine.clock().uninstall();
    });

    it('should scroll to element smoothly', () => {
      const element = document.createElement('div');
      document.body.appendChild(element);
      spyOn(element, 'scrollIntoView');

      service.scrollToElement(element, false);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: 'smooth',
        block: 'start'
      });

      document.body.removeChild(element);
    });

    it('should announce with aria-label if present', () => {
      const element = document.createElement('div');
      element.setAttribute('aria-label', 'Main Content');
      document.body.appendChild(element);
      spyOn(element, 'scrollIntoView');
      spyOn(service, 'announce');

      service.scrollToElement(element, true);

      expect(service.announce).toHaveBeenCalledWith('Main Content');
      document.body.removeChild(element);
    });

    it('should announce with title if aria-label not present', () => {
      const element = document.createElement('div');
      element.setAttribute('title', 'Section Title');
      document.body.appendChild(element);
      spyOn(element, 'scrollIntoView');
      spyOn(service, 'announce');

      service.scrollToElement(element, true);

      expect(service.announce).toHaveBeenCalledWith('Section Title');
      document.body.removeChild(element);
    });

    it('should use default message if no label or title', () => {
      const element = document.createElement('div');
      document.body.appendChild(element);
      spyOn(element, 'scrollIntoView');
      spyOn(service, 'announce');

      service.scrollToElement(element, true);

      expect(service.announce).toHaveBeenCalledWith('Sección cargada');
      document.body.removeChild(element);
    });

    it('should not announce if announce parameter is false', () => {
      const element = document.createElement('div');
      document.body.appendChild(element);
      spyOn(element, 'scrollIntoView');
      spyOn(service, 'announce');

      service.scrollToElement(element, false);

      expect(service.announce).not.toHaveBeenCalled();
      document.body.removeChild(element);
    });
  });
});
