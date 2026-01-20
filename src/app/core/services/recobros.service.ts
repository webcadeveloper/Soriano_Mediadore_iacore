import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { Recibo, Template, ConfigRecobros, ReciboEstado, CanalComunicacion, EventoTipo } from '../../shared/models/recibo.model';

@Injectable({
  providedIn: 'root'
})
export class RecobrosService {
  private recibosSubject = new BehaviorSubject<Recibo[]>(this.getSeedData());
  private templatesSubject = new BehaviorSubject<Template[]>(this.getSeedTemplates());
  private configSubject = new BehaviorSubject<ConfigRecobros>(this.getDefaultConfig());

  recibos$ = this.recibosSubject.asObservable();
  templates$ = this.templatesSubject.asObservable();
  config$ = this.configSubject.asObservable();

  constructor() {
    // Cargar datos del localStorage si existen
    const savedRecibos = localStorage.getItem('recobros_recibos');
    const savedTemplates = localStorage.getItem('recobros_templates');
    const savedConfig = localStorage.getItem('recobros_config');

    if (savedRecibos) this.recibosSubject.next(JSON.parse(savedRecibos));
    if (savedTemplates) this.templatesSubject.next(JSON.parse(savedTemplates));
    if (savedConfig) this.configSubject.next(JSON.parse(savedConfig));
  }

  // ===== RECIBOS =====
  getRecibos(): Recibo[] {
    return this.recibosSubject.value;
  }

  addRecibos(recibos: Recibo[]): void {
    const current = this.recibosSubject.value;
    const updated = [...recibos, ...current];
    this.recibosSubject.next(updated);
    localStorage.setItem('recobros_recibos', JSON.stringify(updated));
  }

  updateRecibo(id: string, updates: Partial<Recibo>): void {
    const current = this.recibosSubject.value;
    const updated = current.map(r => r.id === id ? { ...r, ...updates } : r);
    this.recibosSubject.next(updated);
    localStorage.setItem('recobros_recibos', JSON.stringify(updated));
  }

  deleteRecibo(id: string): void {
    const current = this.recibosSubject.value;
    const updated = current.map(r => r.id === id ? { ...r, deleted: true } : r);
    this.recibosSubject.next(updated);
    localStorage.setItem('recobros_recibos', JSON.stringify(updated));
  }

  // ===== TEMPLATES =====
  getTemplates(): Template[] {
    return this.templatesSubject.value;
  }

  updateTemplate(id: string, updates: Partial<Template>): void {
    const current = this.templatesSubject.value;
    const updated = current.map(t => t.id === id ? { ...t, ...updates } : t);
    this.templatesSubject.next(updated);
    localStorage.setItem('recobros_templates', JSON.stringify(updated));
  }

  // ===== CONFIG =====
  getConfig(): ConfigRecobros {
    return this.configSubject.value;
  }

  updateConfig(updates: Partial<ConfigRecobros>): void {
    const current = this.configSubject.value;
    const updated = { ...current, ...updates };
    this.configSubject.next(updated);
    localStorage.setItem('recobros_config', JSON.stringify(updated));
  }

  // ===== UTILIDADES =====
  calcularScore(r: Recibo): number {
    const A = r.importe >= 30000 ? 1 : r.importe >= 15000 ? 0.6 : r.importe >= 5000 ? 0.3 : 0.1;
    const days = Math.max(0, (Date.now() - new Date(r.venc).getTime()) / 86400000);
    const V = days <= 7 ? 1 : days <= 14 ? 0.6 : 0.3;
    const C = r.tel && r.email ? 1 : r.tel ? 0.6 : r.email ? 0.3 : 0;
    const M = (r.motivo || '').includes('R01') ? 1 : ((r.motivo || '').includes('R03') || (r.motivo || '').includes('R04')) ? 0.7 : 0.4;
    return Math.round(40 * A + 30 * V + 20 * C + 10 * M);
  }

  formatImporte(centimos: number): string {
    return (centimos / 100).toLocaleString('es-ES', { style: 'currency', currency: 'EUR' });
  }

  formatFecha(iso?: string): string {
    return iso ? iso.split('-').reverse().join('/') : '';
  }

  diasDesdeVencimiento(iso: string): number {
    return Math.floor((Date.now() - new Date(iso).getTime()) / 86400000);
  }

  validarIBAN(iban?: string): boolean {
    if (!iban) return false;
    const s = iban.replace(/\s+/g, '').toUpperCase();
    if (!/^([A-Z]{2}\d{2}[A-Z0-9]{1,30})$/.test(s)) return false;
    const re = s.slice(4) + s.slice(0, 4);
    const num = re.replace(/[A-Z]/g, (c) => (c.charCodeAt(0) - 55).toString());
    let mod = 0;
    for (let i = 0; i < num.length; i += 7) {
      const part = (mod.toString() + num.substring(i, i + 7));
      mod = parseInt(part, 10) % 97;
    }
    return mod === 1;
  }

  esHttps(url?: string): boolean {
    try {
      return !!url && new URL(url).protocol === 'https:';
    } catch {
      return false;
    }
  }

  // ===== SEED DATA =====
  private getSeedData(): Recibo[] {
    return [
      {
        id: 'r1',
        cliente: 'Marta López',
        nif: '12345678Z',
        poliza: 'AUTO-123',
        num_recibo: '784512',
        venc: '2025-09-14',
        importe: 15235,
        motivo: 'R01 – Fondos insuficientes',
        tel: '600123456',
        email: 'marta@ejemplo.es',
        estado: ReciboEstado.DEVUELTO,
        canal: CanalComunicacion.WA,
        notas: [],
        historial: []
      },
      {
        id: 'r2',
        cliente: 'Javier Torres',
        nif: '50223311H',
        poliza: 'HOG-778',
        num_recibo: '784513',
        venc: '2025-09-12',
        importe: 8920,
        motivo: 'R03 – Cuenta/orden incorrecta',
        tel: '677888999',
        email: 'jtorres@dom.es',
        estado: ReciboEstado.EN_GESTION,
        canal: CanalComunicacion.EMAIL,
        notas: [],
        historial: []
      },
      {
        id: 'r3',
        cliente: 'Nuria Pérez',
        nif: '22446688K',
        poliza: 'VID-901',
        num_recibo: '784514',
        venc: '2025-09-09',
        importe: 4020,
        motivo: 'R08 – Revocación',
        tel: '961111222',
        email: 'nuria@perez.es',
        estado: ReciboEstado.PENDIENTE,
        canal: CanalComunicacion.WA,
        notas: [],
        historial: []
      }
    ];
  }

  private getSeedTemplates(): Template[] {
    return [
      {
        id: 'preaviso_cargo',
        canal: CanalComunicacion.EMAIL,
        nombre: 'Preaviso de cargo (D-5 a D-2)',
        variant: 'A',
        categoria: 'Previo al cargo',
        motivo: 'Preaviso',
        textoA: 'Hola {nombre}, en los próximos días se cargará el recibo {num_recibo} ({poliza}) por {importe} (vto {venc}). Si deseas cambiar método de pago usa {hub}. Gracias, {agente}.',
        textoB: 'Hola {nombre}. En breve se cargará tu recibo {num_recibo} ({poliza}) por {importe} (vto {venc}). Gestiona aquí: {hub}.'
      },
      {
        id: 'r01_fondos',
        canal: CanalComunicacion.WA,
        nombre: 'Devolución — Fondos insuficientes (R01)',
        variant: 'A',
        categoria: 'Devuelto',
        motivo: 'R01',
        textoA: 'Hola {nombre}, soy {agente} de Soriano Mediadores. Tu recibo {num_recibo} ({poliza}) por {importe} (vto {venc}) figura devuelto por fondos insuficientes. {tono}\nOpciones de pago:\n{bloquePago}\nSeguridad: sólo {dominio}.',
        textoB: 'Hola {nombre}. Recibo {num_recibo} ({poliza}) por {importe} (vto {venc}) devuelto (R01). {tono}\nPaga ahora: {bloquePago}'
      },
      {
        id: 'recordatorio_d2',
        canal: CanalComunicacion.WA,
        nombre: 'Recordatorio D+2',
        variant: 'A',
        categoria: 'Seguimiento',
        motivo: 'Recordatorio',
        textoA: 'Hola {nombre}. Te recordamos el recibo {num_recibo} ({poliza}) de {importe} vencido el {venc}. {tono}\nPaga aquí: {bloquePago}',
        textoB: '{nombre}, recibo {num_recibo} por {importe} (vto {venc}). {tono}\nOpciones: {bloquePago}'
      }
    ];
  }

  private getDefaultConfig(): ConfigRecobros {
    return {
      dominioSeguro: 'pagos.sorianomediadores.es',
      urlTPV: 'https://pagos.sorianomediadores.es/tpv',
      urlHub: 'https://pagos.sorianomediadores.es/hub',
      iban: 'ES12 3456 7890 1234 5678 9012',
      telBizum: '+34600123456',
      urlPayPal: 'https://paypal.me/sorianomediadores',
      agente: 'Laura García',
      role: 'Agente',
      telefonoEmpresa: '961 23 45 67',
      emailEmpresa: 'info@sorianomediadores.es'
    };
  }
}
