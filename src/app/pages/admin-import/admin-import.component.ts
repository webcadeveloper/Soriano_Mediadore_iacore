import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatStepperModule } from '@angular/material/stepper';
import { MatRadioModule } from '@angular/material/radio';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { MatTabsModule } from '@angular/material/tabs';
import { ApiService } from '../../core/services/api.service';
import { LoggerService } from '../../core/services/logger.service';
import {
  ImportType,
  ImportMode,
  ImportStatus,
  ImportConfig,
  CSVPreview,
  ImportProgress,
  ImportHistory,
  ImportTypeDescriptions,
  ImportError
} from '../../shared/models/import.model';
import { Subscription, interval, Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-admin-import',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatStepperModule,
    MatRadioModule,
    MatCheckboxModule,
    MatTableModule,
    MatChipsModule,
    MatSnackBarModule,
    MatDialogModule,
    MatTooltipModule,
    MatDividerModule,
    MatTabsModule
  ],
  templateUrl: './admin-import.component.html',
  styleUrl: './admin-import.component.scss'
})
export class AdminImportComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  // Enums for template
  ImportType = ImportType;
  ImportMode = ImportMode;
  ImportStatus = ImportStatus;
  ImportTypeDescriptions = ImportTypeDescriptions;

  // File handling
  selectedFile: File | null = null;
  isDragging = false;
  pastedCsv = '';

  // Configuration
  importConfig: ImportConfig = {
    type: ImportType.CLIENTES,
    mode: ImportMode.ADD,
    validateBeforeImport: true,
    handleDuplicates: 'skip'
  };

  // Preview
  csvPreview: CSVPreview | null = null;
  isLoadingPreview = false;

  // Import process
  currentImport: ImportProgress | null = null;
  isImporting = false;
  importSubscription: Subscription | null = null;

  // History
  importHistory: ImportHistory[] = [];
  isLoadingHistory = false;
  historyDisplayedColumns: string[] = ['fileName', 'type', 'userName', 'stats', 'status', 'date', 'actions'];

  // Current step (0: upload, 1: config, 2: preview, 3: process, 4: results)
  currentStep = 0;

  constructor(
    private apiService: ApiService,
    private snackBar: MatSnackBar,
    private dialog: MatDialog,
    private logger: LoggerService
  ) {}

  ngOnInit(): void {
    this.loadHistory();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    if (this.importSubscription) {
      this.importSubscription.unsubscribe();
    }
  }

  // ===== FILE HANDLING =====
  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.handleFile(input.files[0]);
    }
  }

  onDragOver(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging = true;
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging = false;
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging = false;

    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      this.handleFile(files[0]);
    }
  }

  // Seleccionar tipo de importación
  selectImportType(type: ImportType): void {
    this.importConfig.type = type;
  }

  // Obtener nombre de archivo esperado
  getExpectedFileName(): string {
    switch (this.importConfig.type) {
      case ImportType.CLIENTES:
        return 'DatosExportados_*.csv';
      case ImportType.POLIZAS:
        return 'POLIZAS.csv';
      case ImportType.RECIBOS:
        return 'RECIBOS.csv';
      case ImportType.SINIESTROS:
        return 'SINIESTROS.csv';
      default:
        return '';
    }
  }

  // Obtener columnas requeridas según tipo
  getRequiredColumns(): string[] {
    switch (this.importConfig.type) {
      case ImportType.CLIENTES:
        return ['NIF', 'Nombre completo', 'IdAccount', 'Email contacto', 'Provincia'];
      case ImportType.POLIZAS:
        return ['Número de la póliza', 'Ramo', 'Nombre del cliente', 'IdAccount', 'Situación de la póliza'];
      case ImportType.RECIBOS:
        return ['Nº recibo', 'Nº póliza', 'Cliente', 'Prima total', 'Situación del recibo'];
      case ImportType.SINIESTROS:
        return ['Número de siniestro', 'Número de póliza', 'Cliente', 'Situación del siniestro', 'IdAccount'];
      default:
        return [];
    }
  }

  // Validar columnas del CSV contra las esperadas
  validateCSVColumns(columns: string[]): boolean {
    const required = this.getRequiredColumns();
    const normalizedColumns = columns.map(col => col.trim().toLowerCase());

    // Verificar que todas las columnas requeridas estén presentes
    for (const reqCol of required) {
      const normalized = reqCol.toLowerCase();
      if (!normalizedColumns.some(col => col.includes(normalized.substring(0, 10)))) {
        this.snackBar.open(`El archivo no contiene la columna requerida: "${reqCol}"`, 'Cerrar', {
          duration: 5000,
          panelClass: ['error-snackbar']
        });
        return false;
      }
    }
    return true;
  }

  handleFile(file: File): void {
    this.logger.debug('Archivo seleccionado', { fileName: file.name, size: file.size });

    // Validate file type
    if (!file.name.endsWith('.csv')) {
      this.snackBar.open('Por favor, selecciona un archivo CSV válido', 'Cerrar', {
        duration: 3000,
        panelClass: ['error-snackbar']
      });
      return;
    }

    // Validate file size (max 100MB)
    const maxSize = 100 * 1024 * 1024;
    if (file.size > maxSize) {
      this.snackBar.open('El archivo es demasiado grande. Máximo 100MB', 'Cerrar', {
        duration: 3000,
        panelClass: ['error-snackbar']
      });
      return;
    }

    this.selectedFile = file;
    this.loadPreview();
    this.currentStep = 3; // Paso 3 = Vista Previa (0=Tipo, 1=Cargar, 2=Config, 3=Preview)
  }

  removeFile(): void {
    this.selectedFile = null;
    this.csvPreview = null;
    this.currentStep = 0;
  }

  // ===== PREVIEW =====
  loadPreview(): void {
    if (!this.selectedFile) return;

    this.isLoadingPreview = true;
    const formData = new FormData();
    formData.append('file', this.selectedFile);

    this.apiService.previewCSV(formData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          if (response.success) {
            this.csvPreview = response.data;

            // Validar columnas del CSV contra las esperadas
            if (this.csvPreview && this.csvPreview.headers) {
              const isValid = this.validateCSVColumns(this.csvPreview.headers);
              if (!isValid) {
                // Si no es válido, limpiar archivo
                this.selectedFile = null;
                this.csvPreview = null;
                this.currentStep = 1; // Volver al paso de carga
                this.isLoadingPreview = false;
                return;
              }
            }

            this.snackBar.open(`✅ Archivo validado correctamente para ${ImportTypeDescriptions[this.importConfig.type].title}`, 'Cerrar', {
              duration: 3000,
              panelClass: ['success-snackbar']
            });
          } else {
            this.snackBar.open(response.message || 'Error al cargar la vista previa', 'Cerrar', {
              duration: 3000,
              panelClass: ['error-snackbar']
            });
          }
          this.isLoadingPreview = false;
        },
        error: (error) => {
          this.logger.error('Error al cargar vista previa del archivo', error);
          this.snackBar.open(`Error al cargar la vista previa del archivo`, 'Cerrar', {
            duration: 5000,
            panelClass: ['error-snackbar']
          });
          this.isLoadingPreview = false;
        }
      });
  }

  // ===== IMPORT =====
  startImport(): void {
    // Permitir importar incluso sin preview para archivos muy grandes
    if (!this.selectedFile) {
      this.snackBar.open('Por favor, selecciona un archivo primero', 'Cerrar', {
        duration: 3000,
        panelClass: ['error-snackbar']
      });
      return;
    }

    if (!this.importConfig.type) {
      this.snackBar.open('Por favor, selecciona el tipo de importación', 'Cerrar', {
        duration: 3000,
        panelClass: ['error-snackbar']
      });
      return;
    }

    this.isImporting = true;
    this.currentStep = 3;

    const formData = new FormData();
    formData.append('file', this.selectedFile);
    formData.append('config', JSON.stringify(this.importConfig));

    this.apiService.startImport(this.importConfig.type, formData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          if (response.success) {
            this.currentImport = response.data;
            this.pollImportStatus(response.data.id);
            this.snackBar.open('Importación iniciada', 'Cerrar', {
              duration: 2000
            });
          } else {
            this.snackBar.open(response.message || 'Error al iniciar la importación', 'Cerrar', {
              duration: 3000,
              panelClass: ['error-snackbar']
            });
            this.isImporting = false;
          }
        },
        error: (error) => {
          this.logger.error('Error al iniciar importación', error);
          this.snackBar.open('Error al iniciar la importación', 'Cerrar', {
            duration: 3000,
            panelClass: ['error-snackbar']
          });
          this.isImporting = false;
        }
      });
  }

  pollImportStatus(importId: string): void {
    // Exponential backoff: start at 1s, increase by 1.5x each time, max 30s
    let pollInterval = 1000;
    const maxInterval = 30000;
    let pollCount = 0;

    const doPoll = () => {
      this.apiService.getImportStatus(importId)
        .pipe(takeUntil(this.destroy$))
        .subscribe({
          next: (response) => {
            if (response.success) {
              this.currentImport = response.data;

              if (
                response.data.status === ImportStatus.COMPLETED ||
                response.data.status === ImportStatus.ERROR ||
                response.data.status === ImportStatus.CANCELLED
              ) {
                this.importSubscription?.unsubscribe();
                this.isImporting = false;
                this.currentStep = 4;
                this.loadHistory();

                if (response.data.status === ImportStatus.COMPLETED) {
                  this.snackBar.open('Importación completada exitosamente', 'Cerrar', {
                    duration: 3000,
                    panelClass: ['success-snackbar']
                  });
                }
              } else {
                // Continue polling with exponential backoff
                pollCount++;
                if (pollCount > 5) {
                  pollInterval = Math.min(pollInterval * 1.5, maxInterval);
                }
                setTimeout(() => doPoll(), pollInterval);
              }
            }
          },
          error: (error) => {
            this.logger.error('Error al consultar estado de importación', error);
            this.importSubscription?.unsubscribe();
            this.isImporting = false;
          }
        });
    };

    doPoll();
  }

  cancelImport(): void {
    if (this.currentImport) {
      this.apiService.cancelImport(this.currentImport.id)
        .pipe(takeUntil(this.destroy$))
        .subscribe({
          next: (response) => {
            if (response.success) {
              this.snackBar.open('Importación cancelada', 'Cerrar', {
                duration: 2000
              });
              this.importSubscription?.unsubscribe();
              this.isImporting = false;
            }
          },
          error: (error) => {
            this.logger.error('Error al cancelar importación', error);
            this.snackBar.open('Error al cancelar la importación', 'Cerrar', {
              duration: 3000,
              panelClass: ['error-snackbar']
            });
          }
        });
    }
  }

  // ===== HISTORY =====
  loadHistory(): void {
    this.isLoadingHistory = true;
    this.apiService.getImportHistory()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          if (response.success) {
            this.importHistory = response.data;
          }
          this.isLoadingHistory = false;
        },
        error: (error) => {
          this.logger.error('Error al cargar historial de importaciones', error);
          this.isLoadingHistory = false;
        }
      });
  }

  downloadErrorReport(importId: string): void {
    const historyItem = this.importHistory.find(h => h.id === importId);
    if (historyItem && historyItem.errors) {
      this.generateErrorReport(historyItem.errors, historyItem.fileName);
    }
  }

  generateErrorReport(errors: ImportError[], fileName: string): void {
    const csvContent = [
      ['Fila', 'Campo', 'Mensaje', 'Valor'],
      ...errors.map(e => [e.row, e.field, e.message, e.value || ''])
    ].map(row => row.join(',')).join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `errores_${fileName}_${new Date().getTime()}.csv`;
    link.click();
    window.URL.revokeObjectURL(url);

    this.snackBar.open('Reporte de errores descargado', 'Cerrar', {
      duration: 2000
    });
  }

  revertImport(importId: string): void {
    // Confirmation dialog would go here
    this.apiService.revertImport(importId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          if (response.success) {
            this.snackBar.open(response.message, 'Cerrar', {
              duration: 3000,
              panelClass: ['success-snackbar']
            });
            this.loadHistory();
          } else {
            this.snackBar.open(response.message, 'Cerrar', {
              duration: 3000,
              panelClass: ['error-snackbar']
            });
          }
        },
        error: (error) => {
          this.logger.error('Error al revertir importación', error);
          this.snackBar.open('Error al revertir la importación', 'Cerrar', {
            duration: 3000,
            panelClass: ['error-snackbar']
          });
        }
      });
  }

  // ===== UTILITIES =====
  getTypeDescription(type: string) {
    return ImportTypeDescriptions[type as ImportType];
  }

  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  }

  getStatusIcon(status: ImportStatus): string {
    const icons = {
      [ImportStatus.PENDING]: 'schedule',
      [ImportStatus.VALIDATING]: 'sync',
      [ImportStatus.PROCESSING]: 'sync',
      [ImportStatus.COMPLETED]: 'check_circle',
      [ImportStatus.ERROR]: 'error',
      [ImportStatus.CANCELLED]: 'cancel'
    };
    return icons[status];
  }

  getStatusColor(status: ImportStatus): string {
    const colors = {
      [ImportStatus.PENDING]: 'accent',
      [ImportStatus.VALIDATING]: 'primary',
      [ImportStatus.PROCESSING]: 'primary',
      [ImportStatus.COMPLETED]: 'success',
      [ImportStatus.ERROR]: 'warn',
      [ImportStatus.CANCELLED]: 'warn'
    };
    return colors[status];
  }

  resetImport(): void {
    this.selectedFile = null;
    this.csvPreview = null;
    this.currentImport = null;
    this.isImporting = false;
    this.currentStep = 0;
    this.importSubscription?.unsubscribe();
  }

  goToStep(step: number): void {
    if (step <= this.currentStep) {
      this.currentStep = step;
    }
  }

  processPastedCsv(): void {
    // TODO: Implementar procesamiento de CSV pegado
    this.snackBar.open('Función de CSV pegado en desarrollo', 'Cerrar', { duration: 3000 });
  }
}
