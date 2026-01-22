import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, shareReplay } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  AnalyticsApiResponse,
  FinancialKPIsResponse,
  PortfolioAnalysisResponse,
  CollectionsPerformanceResponse,
  ClaimsAnalysisResponse,
  PerformanceTrendsResponse
} from '../../shared/models/analytics.model';

/**
 * Servicio para consumir endpoints de Business Intelligence
 * Proporciona acceso a KPIs financieros, análisis de cartera, cobros, siniestros y tendencias
 */
@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  private readonly API_BASE = '/api/analytics';

  // Cache de resultados (5 minutos)
  private cache = new Map<string, { data: any, timestamp: number }>();
  private CACHE_DURATION = 5 * 60 * 1000; // 5 minutos en milisegundos

  constructor(private http: HttpClient) {}

  /**
   * Obtiene KPIs financieros
   * - Primas mes actual/anterior/año
   * - Comisiones mes/año
   * - Promedio prima póliza
   * - Margen comisión
   * - Pocket Share por compañía
   * - Evolución mensual primas (12 meses)
   */
  getFinancialKPIs(): Observable<FinancialKPIsResponse> {
    const cacheKey = 'financial-kpis';
    const cached = this.getFromCache(cacheKey);
    if (cached) {
      return cached;
    }

    const request = this.http
      .get<AnalyticsApiResponse<FinancialKPIsResponse>>(`${this.API_BASE}/financial-kpis`)
      .pipe(
        map(response => response.data),
        shareReplay(1)
      );

    this.setCache(cacheKey, request);
    return request;
  }

  /**
   * Obtiene análisis de cartera
   * - Distribución por ramo
   * - Distribución por provincia
   * - Top 10 clientes vs resto
   * - Concentración de riesgo
   * - Evolución cartera mensual
   */
  getPortfolioAnalysis(): Observable<PortfolioAnalysisResponse> {
    const cacheKey = 'portfolio-analysis';
    const cached = this.getFromCache(cacheKey);
    if (cached) {
      return cached;
    }

    const request = this.http
      .get<AnalyticsApiResponse<PortfolioAnalysisResponse>>(`${this.API_BASE}/portfolio-analysis`)
      .pipe(
        map(response => response.data),
        shareReplay(1)
      );

    this.setCache(cacheKey, request);
    return request;
  }

  /**
   * Obtiene rendimiento de cobros
   * - Ratio de cobro mensual
   * - Recibos devueltos tendencia (12 meses)
   * - Morosidad por rango de días (0-30, 30-60, 60-90, 90+)
   * - Deuda total vs cartera
   * - Clientes morosos recurrentes
   */
  getCollectionsPerformance(): Observable<CollectionsPerformanceResponse> {
    const cacheKey = 'collections-performance';
    const cached = this.getFromCache(cacheKey);
    if (cached) {
      return cached;
    }

    const request = this.http
      .get<AnalyticsApiResponse<CollectionsPerformanceResponse>>(`${this.API_BASE}/collections-performance`)
      .pipe(
        map(response => response.data),
        shareReplay(1)
      );

    this.setCache(cacheKey, request);
    return request;
  }

  /**
   * Obtiene análisis de siniestros
   * - Siniestralidad por ramo
   * - Siniestros abiertos vs cerrados
   * - Tiempo medio de resolución
   * - Siniestros por mes (12 meses)
   */
  getClaimsAnalysis(): Observable<ClaimsAnalysisResponse> {
    const cacheKey = 'claims-analysis';
    const cached = this.getFromCache(cacheKey);
    if (cached) {
      return cached;
    }

    const request = this.http
      .get<AnalyticsApiResponse<ClaimsAnalysisResponse>>(`${this.API_BASE}/claims-analysis`)
      .pipe(
        map(response => response.data),
        shareReplay(1)
      );

    this.setCache(cacheKey, request);
    return request;
  }

  /**
   * Obtiene tendencias de rendimiento
   * - Evolución clientes
   * - Evolución pólizas
   * - Evolución primas
   * - Comparativa con periodo anterior
   *
   * @param period - Periodo: 7days, 30days, 90days, 12months, ytd, custom
   */
  getPerformanceTrends(period: string = '30days'): Observable<PerformanceTrendsResponse> {
    const cacheKey = `performance-trends-${period}`;
    const cached = this.getFromCache(cacheKey);
    if (cached) {
      return cached;
    }

    const request = this.http
      .get<AnalyticsApiResponse<PerformanceTrendsResponse>>(`${this.API_BASE}/performance-trends?period=${period}`)
      .pipe(
        map(response => response.data),
        shareReplay(1)
      );

    this.setCache(cacheKey, request);
    return request;
  }

  /**
   * Limpia la caché de analytics
   * Útil cuando se necesita refrescar los datos
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * Limpia una entrada específica de la caché
   */
  clearCacheEntry(key: string): void {
    this.cache.delete(key);
  }

  // ========== Métodos privados de caché ==========

  private getFromCache(key: string): Observable<any> | null {
    const cached = this.cache.get(key);
    if (!cached) {
      return null;
    }

    const now = Date.now();
    if (now - cached.timestamp > this.CACHE_DURATION) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  private setCache(key: string, data: Observable<any>): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now()
    });
  }
}
