import { Injectable } from '@angular/core';
import { HttpEvent, HttpHandler, HttpInterceptor, HttpRequest, HttpResponse } from '@angular/common/http';
import { Observable, of, delay, catchError, throwError } from 'rxjs';

/**
 * Interceptor que simula respuestas del backend
 * Se activa solo cuando no hay backend disponible
 */
@Injectable()
export class MockInterceptor implements HttpInterceptor {
  private useRealBackend = false;
  private backendChecked = false;

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    // Solo interceptar llamadas a /api
    if (!req.url.includes('/api')) {
      return next.handle(req);
    }

    // Si ya verificamos el backend y funciona, usar backend real
    if (this.useRealBackend) {
      return next.handle(req);
    }

    // Si aún no verificamos el backend, intentar conectar
    if (!this.backendChecked) {
      this.backendChecked = true;
      // Intentar llamada real primero
      return next.handle(req).pipe(
        catchError((error) => {
          // Si falla, usar mock para esta y futuras llamadas
          console.warn('Backend no disponible, usando datos mock:', error.message);
          const mockResponse = this.getMockResponse(req);
          if (mockResponse) {
            return of(new HttpResponse({
              status: 200,
              body: mockResponse
            })).pipe(delay(300));
          }
          return throwError(() => error);
        })
      );
    }

    // Simular respuestas según el endpoint
    const mockResponse = this.getMockResponse(req);

    if (mockResponse) {
      return of(new HttpResponse({
        status: 200,
        body: mockResponse
      })).pipe(delay(300)); // Simular latencia de red
    }

    return next.handle(req);
  }

  private getMockResponse(req: HttpRequest<any>): any {
    const url = req.url;

    // Stats general
    if (url.includes('/api/stats/general') || url.includes('/api/stats') && !url.includes('recibos-kpi')) {
      return {
        clientes_totales: 1247,
        clientes_activos: 1089,
        polizas_activas: 3456,
        primas_mes_actual: 456789.50,
        recibos_pendientes: 234,
        recibos_cobrados: 1123,
        tasa_cobranza: 82.7,
        siniestros_abiertos: 45,
        timestamp: new Date().toISOString()
      };
    }

    // Recibos KPI
    if (url.includes('/api/stats/recibos-kpi')) {
      return {
        success: true,
        data: this.getMockRecibos(),
        total: 5,
        pagina: 1,
        limite: 10,
        total_paginas: 1,
        kpis: {
          total_recibos: 5,
          total_importe: 6326.50,
          pendientes: 2,
          cobrados: 2,
          vencidos: 1,
          tasa_cobranza: 67.5
        }
      };
    }

    // Clientes
    if (url.includes('/api/clientes/') && !url.includes('/polizas')) {
      const id = url.split('/').pop();
      return this.getMockClienteDetail(id || '1');
    }

    if (url.includes('/api/clientes')) {
      return {
        success: true,
        data: this.getMockClientes(),
        total: 5
      };
    }

    // Pólizas de cliente
    if (url.includes('/polizas')) {
      return {
        success: true,
        data: this.getMockPolizas(),
        total: 3
      };
    }

    // Bots
    if (url.includes('/api/bots')) {
      return {
        success: true,
        data: this.getMockBots(),
        total: 6
      };
    }

    // Chat
    if (url.includes('/api/chat')) {
      return {
        success: true,
        message: this.getRandomChatResponse(),
        timestamp: new Date().toISOString()
      };
    }

    return null;
  }

  private getMockRecibos(): any[] {
    return [
      {
        id: 'REC-2024-001',
        cliente: 'Juan Pérez García',
        poliza: 'POL-2024-1234',
        mediador: 'Ana Martínez',
        importe: 1250.50,
        situacion: 'Pendiente',
        fecha_emision: '2024-01-15',
        fecha_vencimiento: '2024-02-15',
        dias_vencido: 5
      },
      {
        id: 'REC-2024-002',
        cliente: 'María González López',
        poliza: 'POL-2024-5678',
        mediador: 'Carlos Rodríguez',
        importe: 850.00,
        situacion: 'Cobrado',
        fecha_emision: '2024-01-10',
        fecha_vencimiento: '2024-02-10',
        fecha_cobro: '2024-02-08',
        dias_vencido: 0
      },
      {
        id: 'REC-2024-003',
        cliente: 'Carlos Rodríguez Martín',
        poliza: 'POL-2024-9012',
        mediador: 'Ana Martínez',
        importe: 2100.75,
        situacion: 'Vencido',
        fecha_emision: '2023-12-20',
        fecha_vencimiento: '2024-01-20',
        dias_vencido: 30
      },
      {
        id: 'REC-2024-004',
        cliente: 'Ana Fernández Sánchez',
        poliza: 'POL-2024-3456',
        mediador: 'Pedro López',
        importe: 675.25,
        situacion: 'Pendiente',
        fecha_emision: '2024-01-18',
        fecha_vencimiento: '2024-02-18',
        dias_vencido: 0
      },
      {
        id: 'REC-2024-005',
        cliente: 'Pedro Martínez Díaz',
        poliza: 'POL-2024-7890',
        mediador: 'Carlos Rodríguez',
        importe: 1450.00,
        situacion: 'Cobrado',
        fecha_emision: '2024-01-05',
        fecha_vencimiento: '2024-02-05',
        fecha_cobro: '2024-02-03',
        dias_vencido: 0
      }
    ];
  }

  private getMockClientes(): any[] {
    return [
      {
        id: '1',
        nif: '12345678A',
        nombre_completo: 'Juan Pérez García',
        email: 'juan.perez@email.com',
        telefono: '600123456',
        direccion: 'Calle Mayor 123, Madrid',
        fecha_alta: '2020-03-15',
        polizas_activas: 3,
        total_primas: 4500.00,
        estado: 'Activo'
      },
      {
        id: '2',
        nif: '87654321B',
        nombre_completo: 'María González López',
        email: 'maria.gonzalez@email.com',
        telefono: '600987654',
        direccion: 'Avenida Libertad 45, Barcelona',
        fecha_alta: '2019-07-22',
        polizas_activas: 2,
        total_primas: 3200.00,
        estado: 'Activo'
      },
      {
        id: '3',
        nif: '11223344C',
        nombre_completo: 'Carlos Rodríguez Martín',
        email: 'carlos.rodriguez@email.com',
        telefono: '600555444',
        direccion: 'Plaza España 7, Valencia',
        fecha_alta: '2021-01-10',
        polizas_activas: 5,
        total_primas: 6800.00,
        estado: 'Activo'
      },
      {
        id: '4',
        nif: '55667788D',
        nombre_completo: 'Ana Fernández Sánchez',
        email: 'ana.fernandez@email.com',
        telefono: '600111222',
        direccion: 'Calle Sol 89, Sevilla',
        fecha_alta: '2018-11-05',
        polizas_activas: 1,
        total_primas: 1200.00,
        estado: 'Activo'
      },
      {
        id: '5',
        nif: '99887766E',
        nombre_completo: 'Pedro Martínez Díaz',
        email: 'pedro.martinez@email.com',
        telefono: '600333444',
        direccion: 'Avenida Constitución 234, Bilbao',
        fecha_alta: '2022-05-18',
        polizas_activas: 4,
        total_primas: 5400.00,
        estado: 'Activo'
      }
    ];
  }

  private getMockClienteDetail(id: string): any {
    return {
      id,
      nif: '12345678A',
      nombre_completo: 'Juan Pérez García',
      email: 'juan.perez@email.com',
      telefono: '600123456',
      telefono_alternativo: '912345678',
      direccion: 'Calle Mayor 123, 3º A',
      codigo_postal: '28013',
      ciudad: 'Madrid',
      provincia: 'Madrid',
      fecha_alta: '2020-03-15',
      fecha_nacimiento: '1985-06-20',
      estado_civil: 'Casado',
      profesion: 'Ingeniero',
      polizas_activas: 3,
      polizas_canceladas: 1,
      total_primas_cartera: 4500.00,
      total_primas_relacion: 5200.00,
      estado: 'Activo',
      notas: 'Cliente premium, renovación automática'
    };
  }

  private getMockPolizas(): any[] {
    return [
      {
        id: 'POL-2024-1234',
        numero_poliza: 'AUTO-2024-1234',
        tipo: 'Automóvil',
        compania: 'Mapfre',
        fecha_efecto: '2024-01-01',
        fecha_vencimiento: '2024-12-31',
        prima_anual: 1500.00,
        estado: 'Activa',
        coberturas: ['Todo Riesgo', 'Asistencia en Viaje', 'Conductor Joven']
      },
      {
        id: 'POL-2024-5678',
        numero_poliza: 'HOGAR-2024-5678',
        tipo: 'Hogar',
        compania: 'Allianz',
        fecha_efecto: '2024-03-01',
        fecha_vencimiento: '2025-02-28',
        prima_anual: 850.00,
        estado: 'Activa',
        coberturas: ['Continente', 'Contenido', 'Responsabilidad Civil']
      },
      {
        id: 'POL-2024-9012',
        numero_poliza: 'VIDA-2024-9012',
        tipo: 'Vida',
        compania: 'Mutua Madrileña',
        fecha_efecto: '2023-06-01',
        fecha_vencimiento: '2033-05-31',
        prima_anual: 2150.00,
        estado: 'Activa',
        coberturas: ['Fallecimiento', 'Invalidez', 'Ahorro']
      }
    ];
  }

  private getMockBots(): any[] {
    return [
      {
        id: 'atencion',
        nombre: 'Asistente de Atención al Cliente',
        descripcion: 'Ayuda con consultas generales, información de pólizas y trámites',
        icon: 'support_agent',
        activo: true,
        conversaciones_totales: 1250
      },
      {
        id: 'cobranza',
        nombre: 'Bot de Gestión de Cobranza',
        descripcion: 'Asiste en seguimiento de pagos, recordatorios y planes de pago',
        icon: 'account_balance_wallet',
        activo: true,
        conversaciones_totales: 890
      },
      {
        id: 'siniestros',
        nombre: 'Gestor de Siniestros',
        descripcion: 'Tramitación de siniestros, seguimiento y documentación',
        icon: 'report_problem',
        activo: true,
        conversaciones_totales: 456
      },
      {
        id: 'agente',
        nombre: 'Asistente de Agente',
        descripcion: 'Apoyo para agentes con información de clientes y productos',
        icon: 'badge',
        activo: true,
        conversaciones_totales: 2100
      },
      {
        id: 'analista',
        nombre: 'Analista de Datos',
        descripcion: 'Genera reportes y análisis de cartera',
        icon: 'analytics',
        activo: true,
        conversaciones_totales: 567
      },
      {
        id: 'auditor',
        nombre: 'Auditor de Cumplimiento',
        descripcion: 'Verifica cumplimiento normativo y detecta anomalías',
        icon: 'verified_user',
        activo: true,
        conversaciones_totales: 234
      }
    ];
  }

  private getRandomChatResponse(): string {
    const responses = [
      'Encantado de ayudarte. ¿En qué puedo asistirte hoy?',
      'Claro, permíteme revisar esa información para ti.',
      'Entiendo tu consulta. Puedo ayudarte con eso.',
      'He revisado el estado. Todo está correcto.',
      'Perfecto, procesando tu solicitud...',
      '¿Hay algo más en lo que pueda ayudarte?'
    ];
    return responses[Math.floor(Math.random() * responses.length)];
  }
}
