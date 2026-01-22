// Modelos para respuestas de Analytics API

export interface FinancialKPIsResponse {
  total_primas_mes_actual: number;
  total_primas_mes_anterior: number;
  total_primas_anio_actual: number;
  total_comisiones_mes: number;
  total_comisiones_anio: number;
  promedio_prima_poliza: number;
  margen_comision_promedio: number;
  pocket_share_por_compania: PocketShareItem[];
  evolucion_mensual_primas: EvolucionMensualItem[];
}

export interface PocketShareItem {
  gestora: string;
  num_polizas: number;
  total_primas: number;
  porcentaje: number;
}

export interface EvolucionMensualItem {
  mes: string;
  total_primas: number;
  num_polizas: number;
}

export interface PortfolioAnalysisResponse {
  distribucion_por_ramo: DistribucionRamoItem[];
  distribucion_por_provincia: DistribucionProvinciaItem[];
  top_10_clientes_vs_resto: Top10Analysis;
  concentracion_riesgo: number;
  evolucion_cartera_mensual: EvolucionCarteraItem[];
}

export interface DistribucionRamoItem {
  ramo: string;
  num_polizas: number;
  total_primas: number;
  porcentaje: number;
}

export interface DistribucionProvinciaItem {
  provincia: string;
  num_clientes: number;
  num_polizas: number;
  total_primas: number;
}

export interface Top10Analysis {
  top_10_primas: number;
  resto_primas: number;
  top_10_count: number;
  resto_count: number;
  porcentaje_top_10: number;
}

export interface EvolucionCarteraItem {
  mes: string;
  num_clientes: number;
  num_polizas: number;
}

export interface CollectionsPerformanceResponse {
  ratio_cobro_mes: number;
  recibos_devueltos_tendencia: TendenciaDevueltosItem[];
  morosidad_por_rango_dias: MorosidadRangoItem[];
  deuda_total_vs_cartera: DeudaVsCarteraData;
  clientes_morosos_recurrentes: ClienteMorosoRecurrente[];
}

export interface TendenciaDevueltosItem {
  mes: string;
  num_devueltos: number;
  importe_devuelto: number;
}

export interface MorosidadRangoItem {
  rango: string;
  num_recibos: number;
  importe_total: number;
}

export interface DeudaVsCarteraData {
  deuda_total: number;
  primas_cartera: number;
  porcentaje_morosidad: number;
}

export interface ClienteMorosoRecurrente {
  id_account: string;
  nombre_completo: string;
  num_devoluciones: number;
  deuda_total: number;
}

export interface ClaimsAnalysisResponse {
  siniestralidad_por_ramo: SiniestralididadRamoItem[];
  siniestros_abiertos: number;
  siniestros_cerrados: number;
  tiempo_medio_resolucion_dias: number;
  siniestros_por_mes: SiniestrosMesItem[];
}

export interface SiniestralididadRamoItem {
  ramo: string;
  num_polizas: number;
  num_siniestros: number;
  siniestralidad: number;
}

export interface SiniestrosMesItem {
  mes: string;
  num_siniestros: number;
}

export interface PerformanceTrendsResponse {
  evolucion_clientes: TrendItem[];
  evolucion_polizas: TrendItem[];
  evolucion_primas: TrendItem[];
  comparativa_periodo_anterior: Comparative;
}

export interface TrendItem {
  fecha: string;
  valor: number;
}

export interface Comparative {
  clientes_cambio_porcentaje: number;
  polizas_cambio_porcentaje: number;
  primas_cambio_porcentaje: number;
}

// Tipos de respuesta genérica de API
export interface AnalyticsApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}
