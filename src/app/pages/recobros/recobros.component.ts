import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatTableModule } from '@angular/material/table';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTabsModule } from '@angular/material/tabs';
import { MatDividerModule } from '@angular/material/divider';
import { Subject, takeUntil } from 'rxjs';

import { RecobrosService } from '../../core/services/recobros.service';
import { EmailTemplateService } from '../../core/services/email-template.service';
import { RecobrosEmailService } from '../../core/services/recobros-email.service';
import { ApiService } from '../../core/services/api.service';
import { Recibo, Template, ConfigRecobros, ReciboEstado, CanalComunicacion } from '../../shared/models/recibo.model';
import { EmailTemplate } from '../../shared/models/email-template.model';

type SeccionRecobros = 'inicio' | 'bandeja' | 'carga' | 'plantillas' | 'analitica' | 'ajustes';

@Component({
  selector: 'app-recobros',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatButtonModule,
    MatCardModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    MatTableModule,
    MatCheckboxModule,
    MatChipsModule,
    MatDialogModule,
    MatSnackBarModule,
    MatTabsModule,
    MatDividerModule
  ],
  templateUrl: './recobros.component.html',
  styleUrls: ['./recobros.component.scss']
})
export class RecobrosComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  // State
  seccionActual: SeccionRecobros = 'inicio';
  recibos: Recibo[] = [];
  templates: Template[] = [];
  emailTemplates: EmailTemplate[] = [];
  config: ConfigRecobros;

  // KPIs from Database
  recibosKPI: any = null;
  statsGeneral: any = null;
  loadingKPIs = false;

  // Filters
  query = '';
  estadoFilter = 'TODOS';
  motivoFilter = 'ALL';
  selected: string[] = [];

  // CSV paste
  pastedCsv = '';

  // Template editor
  templateSeleccionado: string = '';
  emailTemplateSeleccionado: string = '';
  previewHtml: string = '';

  // Enums para template
  ReciboEstado = ReciboEstado;
  CanalComunicacion = CanalComunicacion;

  constructor(
    private recobrosService: RecobrosService,
    private emailTemplateService: EmailTemplateService,
    private recobrosEmailService: RecobrosEmailService,
    private apiService: ApiService,
    private snackBar: MatSnackBar,
    private dialog: MatDialog
  ) {
    this.config = this.recobrosService.getConfig();
  }

  ngOnInit(): void {
    this.recobrosService.recibos$
      .pipe(takeUntil(this.destroy$))
      .subscribe(recibos => {
        this.recibos = recibos
          .filter(r => !r.deleted)
          .map(r => ({ ...r, _score: this.recobrosService.calcularScore(r) }))
          .sort((a, b) => (b._score || 0) - (a._score || 0));
      });

    this.recobrosService.templates$
      .pipe(takeUntil(this.destroy$))
      .subscribe(templates => {
        this.templates = templates;
        if (!this.templateSeleccionado && templates.length > 0) {
          this.templateSeleccionado = templates[0].id;
        }
      });

    this.recobrosService.config$
      .pipe(takeUntil(this.destroy$))
      .subscribe(config => this.config = config);

    this.emailTemplateService.templates$
      .pipe(takeUntil(this.destroy$))
      .subscribe(templates => {
        this.emailTemplates = templates;
        if (!this.emailTemplateSeleccionado && templates.length > 0) {
          this.emailTemplateSeleccionado = templates[0].id;
          this.actualizarPreview();
        }
      });

    // Cargar KPIs reales de la base de datos
    this.loadKPIs();
  }

  // ===== CARGAR KPIs DE LA BASE DE DATOS =====
  loadKPIs(): void {
    this.loadingKPIs = true;

    // Cargar estadísticas generales
    this.apiService.getStatsGeneral().subscribe({
      next: (response) => {
        this.statsGeneral = response.data;
        console.log('Stats General:', this.statsGeneral);
      },
      error: (err) => {
        console.error('Error cargando stats general:', err);
      }
    });

    // Cargar KPIs de recibos CON FILTRO DE RETORNADOS y límite de 200
    this.apiService.getRecibosKPI({ situacion: 'Retornado', limite: 200 }).subscribe({
      next: (response) => {
        this.recibosKPI = response.data;
        this.loadingKPIs = false;
        console.log(`Cargados ${this.recibosKPI.recibos?.length || 0} recibos retornados con email/teléfono`);
      },
      error: (err) => {
        console.error('Error cargando recibos KPI:', err);
        this.loadingKPIs = false;
        this.snackBar.open('Error cargando KPIs de la base de datos', 'Cerrar', { duration: 3000 });
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  // ===== NAVEGACIÓN =====
  cambiarSeccion(seccion: SeccionRecobros): void {
    this.seccionActual = seccion;
  }

  // ===== FILTROS =====
  get recibosFiltrados(): Recibo[] {
    let filtrados = [...this.recibos];

    if (this.estadoFilter !== 'TODOS') {
      filtrados = filtrados.filter(r => r.estado === this.estadoFilter);
    }

    if (this.motivoFilter !== 'ALL') {
      const motivoLower = this.motivoFilter.toLowerCase();
      filtrados = filtrados.filter(r => (r.motivo || '').toLowerCase().includes(motivoLower));
    }

    if (this.query.trim()) {
      const q = this.query.toLowerCase();
      filtrados = filtrados.filter(r =>
        [r.cliente, r.nif, r.poliza, r.num_recibo, r.tel, r.email, r.motivo].join(' ').toLowerCase().includes(q)
      );
    }

    return filtrados;
  }

  // ===== SELECCIÓN =====
  toggleSelection(id: string): void {
    if (this.selected.includes(id)) {
      this.selected = this.selected.filter(x => x !== id);
    } else {
      this.selected = [...this.selected, id];
    }
  }

  selectAll(): void {
    if (this.selected.length === this.recibosFiltrados.length) {
      this.selected = [];
    } else {
      this.selected = this.recibosFiltrados.map(r => r.id);
    }
  }

  clearSelection(): void {
    this.selected = [];
  }

  // ===== CARGA CSV =====
  processPastedCsv(): void {
    if (!this.pastedCsv.trim()) return;

    try {
      const parsed = this.parseCsv(this.pastedCsv);
      if (parsed.length === 0) {
        this.snackBar.open('No se encontraron filas válidas en el CSV', 'Cerrar', { duration: 3000 });
        return;
      }

      const mapped: Recibo[] = parsed.map((o: any, i: number) => ({
        id: `imp_${Date.now()}_${i}`,
        cliente: o.cliente || o.Cliente || '',
        nif: o.nif || o.NIF || '',
        poliza: o.poliza || o.POLIZA || '',
        num_recibo: o.num_recibo || o.recibo || '',
        venc: String(o.venc || o.vencimiento || '').slice(0, 10).replace(/\//g, '-'),
        importe: this.parseImporte(o.importe || '0'),
        motivo: String(o.motivo || ''),
        tel: o.tel || o.telefono || '',
        email: o.email || o.Email || '',
        estado: ReciboEstado.DEVUELTO,
        canal: CanalComunicacion.WA,
        notas: [],
        historial: [{
          id: `h${i}`,
          ts: new Date().toISOString(),
          type: 'IMPORT' as any,
          by: 'sistema'
        }]
      }));

      this.recobrosService.addRecibos(mapped);
      this.snackBar.open(`✅ Importadas ${mapped.length} filas correctamente`, 'Cerrar', { duration: 3000 });
      this.pastedCsv = '';
    } catch (err: any) {
      this.snackBar.open(`Error al importar: ${err?.message}`, 'Cerrar', { duration: 5000 });
    }
  }

  private parseCsv(text: string): any[] {
    const lines = text.split(/\r?\n/).filter(Boolean);
    if (lines.length < 2) return [];

    const headers = lines[0].split(';');
    return lines.slice(1).map((ln) => {
      const cells = ln.split(';');
      const obj: any = {};
      headers.forEach((h, idx) => obj[h] = cells[idx] || '');
      return obj;
    });
  }

  private parseImporte(raw: string): number {
    const cleaned = raw.toString().replace(/\./g, '').replace(',', '.');
    return Math.round(Number(cleaned) * 100) || 0;
  }

  // ===== UTILIDADES =====
  formatImporte(centimos: number): string {
    return this.recobrosService.formatImporte(centimos);
  }

  formatFecha(iso?: string): string {
    return this.recobrosService.formatFecha(iso);
  }

  diasDesde(iso: string): number {
    return this.recobrosService.diasDesdeVencimiento(iso);
  }

  // ===== KPIs =====
  get totalCasos(): number {
    return this.recibos.length;
  }

  get totalDevueltos(): number {
    return this.recibos.filter(r => r.estado === ReciboEstado.DEVUELTO).length;
  }

  get totalPromesas(): number {
    return this.recibos.filter(r => !!r.promesa_pago).length;
  }

  get totalRecuperados(): number {
    return this.recibos.filter(r => r.estado === ReciboEstado.RECUPERADO).length;
  }

  get importeTotal(): number {
    return this.recibos.reduce((a, b) => a + b.importe, 0);
  }

  get ticketMedio(): number {
    return this.recibos.length ? Math.round(this.importeTotal / this.recibos.length) : 0;
  }

  get diasMedios(): number {
    return this.recibos.length
      ? Math.round(this.recibos.reduce((a, b) => a + this.diasDesde(b.venc), 0) / this.recibos.length)
      : 0;
  }

  // ===== TEMPLATES =====
  get templateActual(): Template | undefined {
    return this.templates.find(t => t.id === this.templateSeleccionado);
  }

  actualizarTemplate(field: keyof Template, value: any): void {
    if (!this.templateSeleccionado) return;
    this.recobrosService.updateTemplate(this.templateSeleccionado, { [field]: value });
  }

  get templatesByCategoria(): { [key: string]: Template[] } {
    const categorias = ['Previo al cargo', 'Devuelto', 'Seguimiento', 'Incidencias', 'Cierre', 'Multicanal'];
    const result: { [key: string]: Template[] } = {};
    categorias.forEach(cat => {
      result[cat] = this.templates.filter(t => t.categoria === cat);
    });
    return result;
  }

  // ===== CONFIG =====
  actualizarConfig(field: keyof ConfigRecobros, value: any): void {
    this.recobrosService.updateConfig({ [field]: value });
  }

  validarIBAN(iban?: string): boolean {
    return this.recobrosService.validarIBAN(iban);
  }

  esHttps(url?: string): boolean {
    return this.recobrosService.esHttps(url);
  }

  get todosEnlacesHttps(): boolean {
    return this.esHttps(this.config.urlTPV) &&
           this.esHttps(this.config.urlHub) &&
           this.esHttps(this.config.urlPayPal);
  }

  // ===== ANALÍTICA =====
  get recibosFondos(): number {
    return this.recibos.filter(r => r.motivo.toLowerCase().includes('fondos')).length;
  }

  get recibosCuenta(): number {
    return this.recibos.filter(r => r.motivo.toLowerCase().includes('cuenta')).length;
  }

  get recibosRevocacion(): number {
    return this.recibos.filter(r => r.motivo.toLowerCase().includes('revocación')).length;
  }

  // ===== EMAIL TEMPLATES =====
  get emailTemplateActual(): EmailTemplate | undefined {
    return this.emailTemplates.find(t => t.id === this.emailTemplateSeleccionado);
  }

  get emailTemplatesByCategoria(): { [key: string]: EmailTemplate[] } {
    const categorias = ['Previo al cargo', 'Devuelto', 'Seguimiento', 'Incidencias', 'Cierre'];
    const result: { [key: string]: EmailTemplate[] } = {};
    categorias.forEach(cat => {
      result[cat] = this.emailTemplates.filter(t => t.categoria === cat);
    });
    return result;
  }

  seleccionarEmailTemplate(id: string): void {
    this.emailTemplateSeleccionado = id;
    this.actualizarPreview();
  }

  actualizarEmailTemplate(field: keyof EmailTemplate, value: any): void {
    if (!this.emailTemplateSeleccionado) return;
    this.emailTemplateService.updateTemplate(this.emailTemplateSeleccionado, { [field]: value });
    this.actualizarPreview();
  }

  actualizarPreview(): void {
    if (!this.emailTemplateSeleccionado) return;
    const template = this.emailTemplateActual;
    if (template) {
      this.previewHtml = this.emailTemplateService.previewTemplate(template, this.config);
    }
  }

  getVariablesDisponibles(): string[] {
    return this.emailTemplateService.getVariablesDisponibles();
  }

  // ===== ENVÍO DE EMAILS =====
  enviarEmailRecibo(recibo: Recibo): void {
    if (!recibo.email) {
      this.snackBar.open('El recibo no tiene email registrado', 'Cerrar', { duration: 3000 });
      return;
    }

    // Mostrar diálogo para seleccionar plantilla
    const templateOptions = [
      { value: 1, label: 'Plantilla 1 - Normal', description: 'Primer aviso profesional' },
      { value: 2, label: 'Plantilla 2 - Urgente', description: 'Recordatorio con tono más directo' },
      { value: 3, label: 'Plantilla 3 - Crítico', description: 'Último aviso antes de acciones legales' }
    ];

    // Por ahora, usar plantilla 1 por defecto (TODO: crear diálogo de selección)
    const templateNumber = 1;

    // Preparar request para el backend
    const request = {
      numero_recibo: recibo.num_recibo,
      email_destino: recibo.email,
      template_number: templateNumber
    };

    // Enviar email con plantilla HTML tipo factura
    this.snackBar.open('Enviando email...', '', { duration: 1000 });

    this.recobrosEmailService.sendEmailWithTemplate(request).subscribe({
      next: (response) => {
        this.snackBar.open('✅ Email enviado correctamente', 'Cerrar', { duration: 3000 });
        console.log('Email enviado:', response);
      },
      error: (error) => {
        this.snackBar.open('❌ Error al enviar email: ' + error.message, 'Cerrar', { duration: 5000 });
        console.error('Error enviando email:', error);
      }
    });
  }

  enviarEmailsSeleccionados(): void {
    if (this.selected.length === 0) {
      this.snackBar.open('No hay recibos seleccionados', 'Cerrar', { duration: 3000 });
      return;
    }

    const recibosSeleccionados = this.recibos.filter(r => this.selected.includes(r.id));
    const recibosConEmail = recibosSeleccionados.filter(r => r.email);

    if (recibosConEmail.length === 0) {
      this.snackBar.open('Ningún recibo seleccionado tiene email', 'Cerrar', { duration: 3000 });
      return;
    }

    // Usar plantilla 1 por defecto para envío masivo
    const templateNumber = 1;
    let enviados = 0;
    let errores = 0;

    this.snackBar.open(`Enviando ${recibosConEmail.length} emails...`, '', { duration: 2000 });

    // Enviar cada email individualmente con la plantilla HTML tipo factura
    recibosConEmail.forEach((recibo, index) => {
      const request = {
        numero_recibo: recibo.num_recibo,
        email_destino: recibo.email!,
        template_number: templateNumber
      };

      this.recobrosEmailService.sendEmailWithTemplate(request).subscribe({
        next: (response) => {
          enviados++;
          console.log(`Email ${index + 1}/${recibosConEmail.length} enviado:`, response);

          // Mostrar mensaje final cuando todos terminen
          if (enviados + errores === recibosConEmail.length) {
            if (errores === 0) {
              this.snackBar.open(`✅ ${enviados} emails enviados correctamente`, 'Cerrar', { duration: 4000 });
            } else {
              this.snackBar.open(`✅ ${enviados} enviados, ❌ ${errores} con error`, 'Cerrar', { duration: 5000 });
            }
            this.clearSelection();
          }
        },
        error: (error) => {
          errores++;
          console.error(`Error enviando email ${index + 1}:`, error);

          // Mostrar mensaje final cuando todos terminen
          if (enviados + errores === recibosConEmail.length) {
            this.snackBar.open(`✅ ${enviados} enviados, ❌ ${errores} con error`, 'Cerrar', { duration: 5000 });
            this.clearSelection();
          }
        }
      });
    });
  }

  probarConexionGraph(): void {
    this.snackBar.open('Probando conexión con Microsoft Graph...', '', { duration: 2000 });

    this.recobrosEmailService.testGraphConnection().subscribe({
      next: (response) => {
        this.snackBar.open('✅ Conexión exitosa con Microsoft Graph', 'Cerrar', { duration: 4000 });
        console.log('Test Graph:', response);
      },
      error: (error) => {
        this.snackBar.open('❌ Error de conexión: ' + error.message, 'Cerrar', { duration: 5000 });
        console.error('Error probando Graph:', error);
      }
    });
  }

  private reemplazarVariablesAsunto(asunto: string, recibo: Recibo): string {
    return asunto
      .replace(/{num_recibo}/g, recibo.num_recibo)
      .replace(/{poliza}/g, recibo.poliza)
      .replace(/{cliente}/g, recibo.cliente)
      .replace(/{importe}/g, this.formatImporte(recibo.importe));
  }
}
