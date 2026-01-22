export interface TopCliente {
  nombre_completo: string;
  nif: string;
  total_primas: number;
  num_polizas: number;
}

export interface AnalisisPorRamo {
  ramo: string;
  num_polizas: number;
}

export interface Stats {
  total_clientes: number;
  total_polizas: number;
  total_recibos: number;
  total_siniestros: number;
  top_20_clientes?: TopCliente[];
  analisis_por_ramo?: AnalisisPorRamo[];
}

export interface StatsResponse {
  estadisticas: Stats;
  timestamp: string;
}

export interface HealthResponse {
  status: string;
  database: string;
  cache: string;
  bots: number;
  timestamp: string;
}
