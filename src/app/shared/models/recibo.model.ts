// ===== ENUMS =====
export enum ReciboEstado {
  DEVUELTO = 'DEVUELTO',
  EN_GESTION = 'EN_GESTION',
  PROMESADO = 'PROMESADO',
  PARCIAL = 'PARCIAL',
  RECUPERADO = 'RECUPERADO',
  PENDIENTE = 'PENDIENTE',
  INCOBRABLE = 'INCOBRABLE'
}

export enum CanalComunicacion {
  WA = 'WA',
  EMAIL = 'EMAIL',
  SMS = 'SMS'
}

export enum EventoTipo {
  CONTACT = 'CONTACT',
  NOTE = 'NOTE',
  PROMISE = 'PROMISE',
  PAYMENT = 'PAYMENT',
  IMPORT = 'IMPORT',
  STATE = 'STATE',
  CLIENT_INPUT = 'CLIENT_INPUT'
}

// ===== INTERFACES =====
export interface HistorialEvento {
  id: string;
  ts: string;
  type: EventoTipo;
  by: string;
  data?: any;
}

export interface Recibo {
  id: string;
  cliente: string;
  nif: string;
  poliza: string;
  num_recibo: string;
  venc: string;
  importe: number; // en céntimos
  motivo: string;
  tel?: string;
  email?: string;
  iban?: string;
  estado: ReciboEstado;
  canal?: CanalComunicacion;
  notas?: string[];
  historial?: HistorialEvento[];
  promesa_pago?: string;
  pagado?: number;
  medio_pago?: string;
  ref_pago?: string;
  owner?: string;
  deleted?: boolean;
  _score?: number; // calculado en frontend
}

export interface Template {
  id: string;
  canal: CanalComunicacion;
  nombre: string;
  variant: 'A' | 'B';
  textoA: string;
  textoB: string;
  motivo?: string;
  categoria?: string;
}

export interface ConfigRecobros {
  dominioSeguro: string;
  urlTPV: string;
  urlHub: string;
  iban: string;
  telBizum: string;
  urlPayPal: string;
  agente: string;
  role: 'Agente' | 'Supervisor' | 'Direccion' | 'Auditor';
  telefonoEmpresa?: string;
  emailEmpresa?: string;
}

// ===== API RESPONSES =====
export interface RecibosResponse {
  success: boolean;
  data: Recibo[];
  total: number;
  message?: string;
}

export interface ReciboResponse {
  success: boolean;
  data: Recibo;
  message?: string;
}
