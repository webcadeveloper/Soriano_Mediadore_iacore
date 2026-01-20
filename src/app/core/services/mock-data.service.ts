import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';

/**
 * Servicio de datos mock para desarrollo sin backend
 * Proporciona datos simulados para todas las funcionalidades
 */
@Injectable({
  providedIn: 'root'
})
export class MockDataService {

  /**
   * Genera estadísticas mock para el dashboard
   */
  getStatsMock(): Observable<any> {
    const stats = {
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

    return of(stats).pipe(delay(300)); // Simular latencia de red
  }

  /**
   * Genera KPIs de recibos mock
   */
  getRecibosKPIMock(params?: any): Observable<any> {
    const recibos = [
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

    const response = {
      success: true,
      data: recibos,
      total: recibos.length,
      pagina: params?.pagina || 1,
      limite: params?.limite || 10,
      total_paginas: 1,
      kpis: {
        total_recibos: recibos.length,
        total_importe: recibos.reduce((sum, r) => sum + r.importe, 0),
        pendientes: recibos.filter(r => r.situacion === 'Pendiente').length,
        cobrados: recibos.filter(r => r.situacion === 'Cobrado').length,
        vencidos: recibos.filter(r => r.situacion === 'Vencido').length,
        tasa_cobranza: 67.5
      }
    };

    return of(response).pipe(delay(400));
  }

  /**
   * Genera lista de clientes mock
   */
  getClientesMock(query?: string): Observable<any> {
    const clientes = [
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

    // Filtrar por query si existe
    let filtered = clientes;
    if (query && query.trim()) {
      const lowerQuery = query.toLowerCase();
      filtered = clientes.filter(c =>
        c.nombre_completo.toLowerCase().includes(lowerQuery) ||
        c.nif.toLowerCase().includes(lowerQuery) ||
        c.email.toLowerCase().includes(lowerQuery)
      );
    }

    return of({
      success: true,
      data: filtered,
      total: filtered.length
    }).pipe(delay(300));
  }

  /**
   * Genera detalle de un cliente mock
   */
  getClienteDetailMock(id: string): Observable<any> {
    const cliente = {
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

    return of(cliente).pipe(delay(200));
  }

  /**
   * Genera pólizas de un cliente mock
   */
  getPolizasClienteMock(clienteId: string): Observable<any> {
    const polizas = [
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

    return of({
      success: true,
      data: polizas,
      total: polizas.length
    }).pipe(delay(300));
  }

  /**
   * Genera lista de bots mock
   */
  getBotsMock(): Observable<any> {
    const bots = [
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

    return of({
      success: true,
      data: bots,
      total: bots.length
    }).pipe(delay(250));
  }

  /**
   * Simula respuesta de chat con un bot
   */
  chatMock(botType: string, message: string): Observable<any> {
    const responses: { [key: string]: string[] } = {
      'atencion': [
        'Encantado de ayudarte. ¿En qué puedo asistirte hoy?',
        'Claro, permíteme revisar esa información para ti.',
        'Entiendo tu consulta. Puedo ayudarte con eso.',
        '¿Hay algo más en lo que pueda ayudarte?'
      ],
      'cobranza': [
        'He revisado el estado de tus recibos. Todo está al día.',
        'Encontré los recibos pendientes. ¿Deseas ver un resumen?',
        'Puedo ayudarte a configurar un recordatorio de pago.',
        'El pago se procesará en las próximas 24 horas.'
      ],
      'siniestros': [
        'Lamento escuchar sobre el incidente. Vamos a tramitar el siniestro.',
        'He registrado tu siniestro con el número SIN-2024-001.',
        'La documentación está completa. Procederemos con la tramitación.',
        'El perito se pondrá en contacto contigo en 48 horas.'
      ],
      'default': [
        'Entiendo. ¿Puedes darme más detalles?',
        'Perfecto, procesando tu solicitud...',
        'He registrado tu petición correctamente.',
        'Gracias por la información.'
      ]
    };

    const botResponses = responses[botType] || responses['default'];
    const randomResponse = botResponses[Math.floor(Math.random() * botResponses.length)];

    return of({
      success: true,
      message: randomResponse,
      timestamp: new Date().toISOString(),
      bot: botType
    }).pipe(delay(800)); // Simular tiempo de pensamiento del bot
  }

  /**
   * Genera datos para reportes mock
   */
  getReportesMock(): Observable<any> {
    const reportes = {
      produccion_mensual: {
        labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun'],
        data: [45000, 52000, 48000, 61000, 58000, 67000]
      },
      distribucion_productos: {
        labels: ['Auto', 'Hogar', 'Vida', 'Salud', 'Empresas'],
        data: [35, 25, 20, 12, 8]
      },
      cobranza_mensual: {
        labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun'],
        data: [82, 85, 78, 88, 84, 89]
      },
      top_mediadores: [
        { nombre: 'Ana Martínez', primas: 125000, clientes: 89 },
        { nombre: 'Carlos Rodríguez', primas: 98000, clientes: 67 },
        { nombre: 'Pedro López', primas: 87000, clientes: 54 },
        { nombre: 'Laura García', primas: 76000, clientes: 48 },
        { nombre: 'Miguel Sánchez', primas: 65000, clientes: 42 }
      ]
    };

    return of({
      success: true,
      data: reportes
    }).pipe(delay(400));
  }
}
