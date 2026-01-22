export interface Cliente {
  id: number;
  nif: string;
  id_account: string;
  nombre_completo: string;
  email?: string;
  email_contacto?: string;
  telefono?: string;
  telefono_contacto?: string;
  direccion?: string;
  domicilio?: string;
  codigo_postal?: string;
  provincia?: string;
  total_primas: number;
  total_primas_cartera?: number;
  total_primas_relacion?: number;
  total_comisiones?: number;
  polizas?: any[];
}

export interface ClientesResponse {
  clientes: Cliente[];
  total: number;
}
