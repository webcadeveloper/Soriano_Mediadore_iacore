import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Cliente, ClientesResponse } from '../../shared/models/cliente.model';
import { Bot, BotsResponse, ChatMessage, ChatResponse } from '../../shared/models/bot.model';
import { StatsResponse, HealthResponse } from '../../shared/models/stats.model';
import {
  ImportType,
  PreviewResponse,
  ImportResponse,
  ImportStatusResponse,
  ImportHistoryResponse,
  RevertResponse
} from '../../shared/models/import.model';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly API_URL = '/api'; // Usa el proxy
  private readonly DIRECT_API_URL = 'http://localhost:8080/api'; // Directo al backend para archivos grandes

  constructor(private http: HttpClient) {}

  // Health Check
  getHealth(): Observable<HealthResponse> {
    return this.http.get<HealthResponse>('/health');
  }

  // Estadísticas
  getStats(): Observable<StatsResponse> {
    return this.http.get<StatsResponse>(`${this.API_URL}/stats`);
  }

  // Estadísticas generales del sistema
  getStatsGeneral(): Observable<any> {
    return this.http.get<any>(`${this.API_URL}/stats/general`);
  }

  // KPIs de recibos con filtros y paginación
  getRecibosKPI(params?: {
    situacion?: string,
    fecha_desde?: string,
    fecha_hasta?: string,
    mediador?: string,
    cliente?: string,
    importe_min?: number,
    importe_max?: number,
    ordenar_por?: string,
    orden?: string,
    pagina?: number,
    limite?: number
  }): Observable<any> {
    let httpParams = new HttpParams();
    if (params) {
      Object.keys(params).forEach(key => {
        const value = (params as any)[key];
        if (value !== undefined && value !== null && value !== '') {
          httpParams = httpParams.set(key, value.toString());
        }
      });
    }
    return this.http.get<any>(`${this.API_URL}/stats/recibos-kpi`, { params: httpParams });
  }

  // Bots
  getBots(): Observable<BotsResponse> {
    return this.http.get<BotsResponse>(`${this.API_URL}/bots`);
  }

  // Clientes
  buscarClientes(query: string): Observable<ClientesResponse> {
    const params = new HttpParams().set('q', query);
    return this.http.get<ClientesResponse>(`${this.API_URL}/clientes`, { params });
  }

  getCliente(id: string): Observable<Cliente> {
    return this.http.get<Cliente>(`${this.API_URL}/clientes/${id}`);
  }

  getPolizasCliente(id: string): Observable<any> {
    return this.http.get<any>(`${this.API_URL}/clientes/${id}/polizas`);
  }

  // Chat con bots
  chatAtencion(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/atencion`, message);
  }

  chatCobranza(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/cobranza`, message);
  }

  chatSiniestros(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/siniestros`, message);
  }

  chatAgente(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/agente`, message);
  }

  chatAnalista(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/analista`, message);
  }

  chatAuditor(message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/auditor`, message);
  }

  // Chat genérico
  chat(botType: string, message: ChatMessage): Observable<ChatResponse> {
    return this.http.post<ChatResponse>(`${this.API_URL}/chat/${botType}`, message);
  }

  // ===== IMPORT DATA ENDPOINTS =====

  // Preview CSV file - Ahora pasa por NGINX con límite de 100MB
  previewCSV(formData: FormData): Observable<PreviewResponse> {
    return this.http.post<PreviewResponse>(`${this.API_URL}/admin/import/preview`, formData);
  }

  // Start import based on type - Ahora pasa por NGINX con límite de 100MB
  startImport(type: ImportType, formData: FormData): Observable<ImportResponse> {
    return this.http.post<ImportResponse>(`${this.API_URL}/admin/import/start`, formData);
  }

  // Get import status (for polling)
  getImportStatus(importId: string): Observable<ImportStatusResponse> {
    return this.http.get<ImportStatusResponse>(`${this.API_URL}/admin/import/status/${importId}`);
  }

  // Cancel ongoing import
  cancelImport(importId: string): Observable<ImportResponse> {
    return this.http.post<ImportResponse>(`${this.API_URL}/admin/import/cancel/${importId}`, {});
  }

  // Get import history
  getImportHistory(limit?: number, offset?: number): Observable<ImportHistoryResponse> {
    let params = new HttpParams();
    if (limit) params = params.set('limit', limit.toString());
    if (offset) params = params.set('offset', offset.toString());
    return this.http.get<ImportHistoryResponse>(`${this.API_URL}/admin/import/history`, { params });
  }

  // Revert a specific import
  revertImport(importId: string): Observable<RevertResponse> {
    return this.http.post<RevertResponse>(`${this.API_URL}/admin/import/revert/${importId}`, {});
  }
}
