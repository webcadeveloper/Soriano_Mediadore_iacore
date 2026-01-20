import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { EmailTemplate, PLANTILLAS_EMAIL_PREDEFINIDAS, EMAIL_BASE_TEMPLATE, VARIABLES_DISPONIBLES } from '../../shared/models/email-template.model';
import { Recibo } from '../../shared/models/recibo.model';
import { ConfigRecobros } from '../../shared/models/recibo.model';

@Injectable({
  providedIn: 'root'
})
export class EmailTemplateService {
  private readonly STORAGE_KEY = 'soriano_email_templates';
  private templatesSubject = new BehaviorSubject<EmailTemplate[]>([]);

  public templates$: Observable<EmailTemplate[]> = this.templatesSubject.asObservable();

  constructor() {
    this.loadTemplates();
  }

  private loadTemplates(): void {
    const stored = localStorage.getItem(this.STORAGE_KEY);
    if (stored) {
      try {
        const templates = JSON.parse(stored) as EmailTemplate[];
        this.templatesSubject.next(templates);
      } catch (err) {
        console.error('Error al cargar plantillas de email:', err);
        this.initializeDefaultTemplates();
      }
    } else {
      this.initializeDefaultTemplates();
    }
  }

  private initializeDefaultTemplates(): void {
    const templates: EmailTemplate[] = PLANTILLAS_EMAIL_PREDEFINIDAS.map(t => ({
      id: t.id!,
      nombre: t.nombre!,
      asunto: t.asunto!,
      categoria: t.categoria!,
      motivo: t.motivo,
      htmlContent: t.htmlContent!,
      variables: t.variables!,
      incluirBloquePago: t.incluirBloquePago!,
      activo: t.activo!
    }));

    this.templatesSubject.next(templates);
    this.saveTemplates(templates);
  }

  private saveTemplates(templates: EmailTemplate[]): void {
    localStorage.setItem(this.STORAGE_KEY, JSON.stringify(templates));
    this.templatesSubject.next(templates);
  }

  getTemplates(): EmailTemplate[] {
    return this.templatesSubject.value;
  }

  getTemplateById(id: string): EmailTemplate | undefined {
    return this.templatesSubject.value.find(t => t.id === id);
  }

  updateTemplate(id: string, updates: Partial<EmailTemplate>): void {
    const templates = this.templatesSubject.value.map(t =>
      t.id === id ? { ...t, ...updates } : t
    );
    this.saveTemplates(templates);
  }

  addTemplate(template: EmailTemplate): void {
    const templates = [...this.templatesSubject.value, template];
    this.saveTemplates(templates);
  }

  deleteTemplate(id: string): void {
    const templates = this.templatesSubject.value.filter(t => t.id !== id);
    this.saveTemplates(templates);
  }

  /**
   * Genera el HTML del bloque de pago basado en la configuración
   */
  generarBloquePago(config: ConfigRecobros): string {
    const opciones: string[] = [];

    if (config.urlTPV && this.esHttps(config.urlTPV)) {
      opciones.push(`
        <div class="payment-option">
          <strong>💳 Pago con Tarjeta (TPV)</strong>
          <p>Acceda de forma segura a nuestro TPV:</p>
          <a href="${config.urlTPV}" class="button">Pagar con Tarjeta</a>
        </div>
      `);
    }

    if (config.urlHub && this.esHttps(config.urlHub)) {
      opciones.push(`
        <div class="payment-option">
          <strong>📱 Bizum</strong>
          <p>Pague de forma rápida con Bizum:</p>
          <a href="${config.urlHub}" class="button">Pagar con Bizum</a>
        </div>
      `);
    }

    if (config.urlPayPal && this.esHttps(config.urlPayPal)) {
      opciones.push(`
        <div class="payment-option">
          <strong>🅿️ PayPal</strong>
          <p>Utilice su cuenta de PayPal:</p>
          <a href="${config.urlPayPal}" class="button">Pagar con PayPal</a>
        </div>
      `);
    }

    if (config.iban && this.validarIBAN(config.iban)) {
      opciones.push(`
        <div class="payment-option">
          <strong>🏦 Transferencia Bancaria</strong>
          <p>Puede realizar una transferencia a nuestra cuenta:</p>
          <p style="font-family: monospace; background-color: #f5f5f5; padding: 8px; border-radius: 4px;">
            <strong>IBAN:</strong> ${this.formatearIBAN(config.iban)}
          </p>
          <p style="font-size: 12px; color: #666;">Por favor, indique el número de recibo en el concepto.</p>
        </div>
      `);
    }

    if (opciones.length === 0) {
      return `
        <div class="payment-section">
          <h3>📞 Métodos de Pago</h3>
          <p>Por favor, contacte con nosotros para conocer los métodos de pago disponibles.</p>
          <p><strong>Teléfono:</strong> {telefono_empresa}</p>
          <p><strong>Email:</strong> {email_empresa}</p>
        </div>
      `;
    }

    return `
      <div class="payment-section">
        <h3>💰 Métodos de Pago Disponibles</h3>
        <p style="margin-bottom: 16px;">Puede regularizar su situación mediante cualquiera de estos métodos seguros:</p>
        ${opciones.join('\n')}
      </div>
    `;
  }

  /**
   * Renderiza una plantilla completa reemplazando todas las variables
   */
  renderTemplate(
    template: EmailTemplate,
    recibo: Recibo,
    config: ConfigRecobros,
    agente: string = 'Equipo Soriano Mediadores'
  ): string {
    let content = template.htmlContent;

    // Generar bloque de pago si es necesario
    if (template.incluirBloquePago) {
      const bloquePago = this.generarBloquePago(config);
      content = content.replace(/{bloquePago}/g, bloquePago);
    } else {
      content = content.replace(/{bloquePago}/g, '');
    }

    // Reemplazar variables del recibo
    const replacements: { [key: string]: string } = {
      '{nombre}': recibo.cliente || 'Cliente',
      '{nif}': recibo.nif || '',
      '{poliza}': recibo.poliza || '',
      '{num_recibo}': recibo.num_recibo || '',
      '{importe}': this.formatImporte(recibo.importe),
      '{venc}': this.formatFecha(recibo.venc),
      '{dias_vencido}': this.diasDesdeVencimiento(recibo.venc).toString(),
      '{motivo}': recibo.motivo || 'Sin especificar',
      '{agente}': agente,
      '{empresa}': 'Soriano Mediadores de Seguros',
      '{telefono_empresa}': config.telefonoEmpresa || '',
      '{email_empresa}': config.emailEmpresa || ''
    };

    // Aplicar reemplazos
    Object.entries(replacements).forEach(([variable, valor]) => {
      const regex = new RegExp(variable.replace(/[{}]/g, '\\$&'), 'g');
      content = content.replace(regex, valor);
    });

    // Insertar contenido en la plantilla base
    let fullHtml = EMAIL_BASE_TEMPLATE.replace('{content}', content);
    fullHtml = fullHtml.replace('{asunto}', template.asunto);
    fullHtml = fullHtml.replace(/{email_empresa}/g, config.emailEmpresa || '');
    fullHtml = fullHtml.replace(/{telefono_empresa}/g, config.telefonoEmpresa || '');

    return fullHtml;
  }

  /**
   * Genera una vista previa con datos de ejemplo
   */
  previewTemplate(template: EmailTemplate, config: ConfigRecobros): string {
    const reciboEjemplo: Recibo = {
      id: 'preview',
      cliente: 'Juan Pérez García',
      nif: '12345678A',
      poliza: 'POL-2024-001234',
      num_recibo: 'REC-2024-005678',
      venc: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      importe: 25000, // 250.00€
      motivo: 'R01 - Fondos insuficientes',
      tel: '600123456',
      email: 'juan.perez@example.com',
      estado: 'DEVUELTO' as any,
      canal: 'EMAIL' as any,
      notas: [],
      historial: []
    };

    return this.renderTemplate(template, reciboEjemplo, config, 'María González');
  }

  // Utilidades
  private formatImporte(centimos: number): string {
    return new Intl.NumberFormat('es-ES', {
      style: 'currency',
      currency: 'EUR'
    }).format(centimos / 100);
  }

  private formatFecha(iso?: string): string {
    if (!iso) return '—';
    const d = new Date(iso);
    return d.toLocaleDateString('es-ES', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  private diasDesdeVencimiento(venc: string): number {
    const vencDate = new Date(venc);
    const hoy = new Date();
    const diff = hoy.getTime() - vencDate.getTime();
    return Math.floor(diff / (1000 * 60 * 60 * 24));
  }

  private validarIBAN(iban?: string): boolean {
    if (!iban) return false;
    const clean = iban.replace(/\s/g, '').toUpperCase();
    if (!/^[A-Z]{2}\d{22}$/.test(clean)) return false;

    const reordered = clean.slice(4) + clean.slice(0, 4);
    const numericIban = reordered.replace(/[A-Z]/g, char => (char.charCodeAt(0) - 55).toString());

    let remainder = '';
    for (let i = 0; i < numericIban.length; i++) {
      remainder = (parseInt(remainder + numericIban[i]) % 97).toString();
    }

    return parseInt(remainder) === 1;
  }

  private formatearIBAN(iban: string): string {
    const clean = iban.replace(/\s/g, '');
    return clean.match(/.{1,4}/g)?.join(' ') || iban;
  }

  private esHttps(url?: string): boolean {
    if (!url) return false;
    return url.toLowerCase().startsWith('https://');
  }

  /**
   * Obtiene las variables disponibles para una plantilla
   */
  getVariablesDisponibles(): string[] {
    return [...VARIABLES_DISPONIBLES];
  }

  /**
   * Valida que una plantilla HTML contenga las etiquetas básicas necesarias
   */
  validarPlantillaHTML(html: string): { valido: boolean; errores: string[] } {
    const errores: string[] = [];

    if (!html || html.trim().length === 0) {
      errores.push('El contenido HTML no puede estar vacío');
    }

    // Verificar que no contenga scripts maliciosos
    if (/<script/i.test(html)) {
      errores.push('No se permiten etiquetas <script> en las plantillas');
    }

    if (/<iframe/i.test(html)) {
      errores.push('No se permiten etiquetas <iframe> en las plantillas');
    }

    return {
      valido: errores.length === 0,
      errores
    };
  }
}
