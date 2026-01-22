import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface SendEmailRequest {
  recibo_id: string;
  cliente_email: string;
  template_id: string;
  from: string;
  subject: string;
  html_body: string;
}

export interface BulkEmailRequest {
  from: string;
  emails: {
    recibo_id: string;
    to: string;
    subject: string;
    body: string;
  }[];
}

export interface SendEmailTemplateRequest {
  numero_recibo: string;
  email_destino: string;
  template_number: number; // 1, 2, or 3
  nombre_cliente?: string;
  numero_poliza?: string;
  ramo?: string;
  tomador?: string;
  descripcion_riesgo?: string;
  prima_total?: string;
  detalle_recibo?: string;
  situacion_recibo?: string;
}

export interface EmailResponse {
  success: boolean;
  message: string;
  data?: any;
}

@Injectable({
  providedIn: 'root'
})
export class RecobrosEmailService {
  private readonly API_URL = environment.apiUrl || 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  /**
   * Envía un email de recobro individual
   */
  sendReciboEmail(request: SendEmailRequest): Observable<EmailResponse> {
    const url = `${this.API_URL}/recobros/send-email`;
    return this.http.post<EmailResponse>(url, request).pipe(
      catchError(this.handleError)
    );
  }

  /**
   * Envía múltiples emails de recobro en lote
   */
  sendBulkEmails(request: BulkEmailRequest): Observable<EmailResponse> {
    const url = `${this.API_URL}/recobros/send-bulk`;
    return this.http.post<EmailResponse>(url, request).pipe(
      catchError(this.handleError)
    );
  }

  /**
   * Envía un email de prueba
   */
  sendTestEmail(from: string, to: string, subject?: string, body?: string): Observable<EmailResponse> {
    const url = `${this.API_URL}/recobros/test-email`;
    const payload = {
      from,
      to,
      subject: subject || 'Email de Prueba - Soriano Mediadores',
      body: body || ''
    };
    return this.http.post<EmailResponse>(url, payload).pipe(
      catchError(this.handleError)
    );
  }

  /**
   * Envía un email usando las plantillas HTML tipo factura (1, 2, o 3)
   */
  sendEmailWithTemplate(request: SendEmailTemplateRequest): Observable<EmailResponse> {
    const url = `${this.API_URL}/recobros/send-email-template`;
    return this.http.post<EmailResponse>(url, request).pipe(
      catchError(this.handleError)
    );
  }

  /**
   * Prueba la conexión con Microsoft Graph
   */
  testGraphConnection(): Observable<EmailResponse> {
    const url = `${this.API_URL}/recobros/test-graph`;
    return this.http.get<EmailResponse>(url).pipe(
      catchError(this.handleError)
    );
  }

  /**
   * Manejo de errores
   */
  private handleError(error: any): Observable<never> {
    let errorMessage = 'Error desconocido';

    if (error.error instanceof ErrorEvent) {
      // Error del lado del cliente
      errorMessage = `Error: ${error.error.message}`;
    } else {
      // Error del lado del servidor
      errorMessage = error.error?.message || `Error ${error.status}: ${error.statusText}`;
    }

    console.error('Error en RecobrosEmailService:', errorMessage);
    return throwError(() => new Error(errorMessage));
  }
}
