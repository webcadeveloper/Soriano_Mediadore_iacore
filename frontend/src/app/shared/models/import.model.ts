// ===== ENUMS =====
export enum ImportType {
  CLIENTES = 'clientes',
  POLIZAS = 'polizas',
  RECIBOS = 'recibos',
  SINIESTROS = 'siniestros'
}

export enum ImportMode {
  ADD = 'add',
  REPLACE = 'replace'
}

export enum ImportStatus {
  PENDING = 'pending',
  VALIDATING = 'validating',
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  ERROR = 'error',
  CANCELLED = 'cancelled'
}

// ===== INTERFACES =====
export interface ImportConfig {
  type: ImportType;
  mode: ImportMode;
  validateBeforeImport: boolean;
  handleDuplicates: 'skip' | 'update' | 'error';
}

export interface CSVPreview {
  headers: string[];
  rows: any[][];
  totalRows: number;
  fileSize: number;
  fileName: string;
}

export interface ImportError {
  row: number;
  field: string;
  message: string;
  value?: any;
}

export interface ImportStats {
  totalRows: number;
  processedRows: number;
  successfulRows: number;
  errorRows: number;
  duplicateRows: number;
  skippedRows: number;
}

export interface ImportProgress {
  id: string;
  status: ImportStatus;
  stats: ImportStats;
  errors: ImportError[];
  startTime: Date;
  endTime?: Date;
  progress: number; // 0-100
  message?: string;
}

export interface ImportHistory {
  id: string;
  type: ImportType;
  fileName: string;
  fileSize: number;
  status: ImportStatus;
  stats: ImportStats;
  userName: string;
  startTime: Date;
  endTime?: Date;
  errors?: ImportError[];
  canRevert: boolean;
}

// ===== API RESPONSES =====
export interface PreviewResponse {
  success: boolean;
  data: CSVPreview;
  message?: string;
}

export interface ImportResponse {
  success: boolean;
  data: ImportProgress;
  message?: string;
}

export interface ImportStatusResponse {
  success: boolean;
  data: ImportProgress;
  message?: string;
}

export interface ImportHistoryResponse {
  success: boolean;
  data: ImportHistory[];
  total: number;
  message?: string;
}

export interface RevertResponse {
  success: boolean;
  message: string;
  rowsReverted?: number;
}

// ===== TYPE DESCRIPTIONS =====
export const ImportTypeDescriptions: Record<ImportType, { title: string; description: string; icon: string }> = {
  [ImportType.CLIENTES]: {
    title: 'Clientes',
    description: 'Importar información de clientes: NIF, nombre, contacto, dirección',
    icon: 'people'
  },
  [ImportType.POLIZAS]: {
    title: 'Pólizas',
    description: 'Importar pólizas de seguros: número de póliza, asegurado, coberturas',
    icon: 'description'
  },
  [ImportType.RECIBOS]: {
    title: 'Recibos',
    description: 'Importar recibos de pago: fecha, importe, estado de pago',
    icon: 'receipt'
  },
  [ImportType.SINIESTROS]: {
    title: 'Siniestros',
    description: 'Importar siniestros: fecha, descripción, estado, indemnización',
    icon: 'warning'
  }
};
