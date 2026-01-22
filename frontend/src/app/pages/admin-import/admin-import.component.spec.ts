import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AdminImportComponent } from './admin-import.component';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

describe('AdminImportComponent', () => {
  let component: AdminImportComponent;
  let fixture: ComponentFixture<AdminImportComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        AdminImportComponent,
        HttpClientTestingModule,
        BrowserAnimationsModule
      ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AdminImportComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with default values', () => {
    expect(component.selectedFile).toBeNull();
    expect(component.isDragging).toBeFalse();
    expect(component.currentStep).toBe(0);
    expect(component.isImporting).toBeFalse();
  });

  it('should format file size correctly', () => {
    expect(component.formatFileSize(0)).toBe('0 Bytes');
    expect(component.formatFileSize(1024)).toBe('1 KB');
    expect(component.formatFileSize(1048576)).toBe('1 MB');
  });

  it('should handle file selection', () => {
    const file = new File(['test'], 'test.csv', { type: 'text/csv' });
    component.handleFile(file);
    expect(component.selectedFile).toBe(file);
  });

  it('should reject non-CSV files', () => {
    const file = new File(['test'], 'test.txt', { type: 'text/plain' });
    component.handleFile(file);
    expect(component.selectedFile).toBeNull();
  });

  it('should reset import state', () => {
    component.selectedFile = new File(['test'], 'test.csv', { type: 'text/csv' });
    component.currentStep = 2;
    component.resetImport();
    expect(component.selectedFile).toBeNull();
    expect(component.currentStep).toBe(0);
  });
});
