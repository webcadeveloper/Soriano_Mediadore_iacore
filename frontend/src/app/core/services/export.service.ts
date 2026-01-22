import { Injectable } from '@angular/core';

/**
 * Opciones para exportación
 */
export interface ExportOptions {
  filename?: string;
  sheetName?: string;
  includeHeaders?: boolean;
  dateFormat?: string;
}

/**
 * Servicio de exportación de datos
 * Permite exportar datos a CSV, Excel y PDF
 */
@Injectable({
  providedIn: 'root'
})
export class ExportService {

  /**
   * Exporta datos a CSV
   */
  exportToCSV(data: any[], options: ExportOptions = {}): void {
    if (!data || data.length === 0) {
      console.warn('No hay datos para exportar');
      return;
    }

    const filename = options.filename || `export_${this.getTimestamp()}.csv`;
    const includeHeaders = options.includeHeaders !== false;

    // Obtener headers de las claves del primer objeto
    const headers = Object.keys(data[0]);

    // Crear contenido CSV
    let csvContent = '';

    // Añadir headers si está habilitado
    if (includeHeaders) {
      csvContent += headers.map(h => this.escapeCSV(h)).join(',') + '\n';
    }

    // Añadir filas
    data.forEach(row => {
      const values = headers.map(header => {
        const value = row[header];
        return this.escapeCSV(this.formatValue(value));
      });
      csvContent += values.join(',') + '\n';
    });

    // Crear Blob y descargar
    this.downloadFile(csvContent, filename, 'text/csv;charset=utf-8;');
  }

  /**
   * Exporta datos a Excel (formato XLSX simulado con CSV)
   */
  exportToExcel(data: any[], options: ExportOptions = {}): void {
    // Para una implementación completa, usar librería como xlsx
    // Por ahora, exportamos como CSV con extensión .xlsx
    const filename = options.filename || `export_${this.getTimestamp()}.xlsx`;
    this.exportToCSV(data, { ...options, filename });
  }

  /**
   * Exporta datos a JSON
   */
  exportToJSON(data: any[], options: ExportOptions = {}): void {
    if (!data || data.length === 0) {
      console.warn('No hay datos para exportar');
      return;
    }

    const filename = options.filename || `export_${this.getTimestamp()}.json`;
    const jsonContent = JSON.stringify(data, null, 2);

    this.downloadFile(jsonContent, filename, 'application/json;charset=utf-8;');
  }

  /**
   * Exporta datos a PDF (requiere implementación adicional)
   */
  exportToPDF(data: any[], options: ExportOptions = {}): void {
    // Para implementación completa, usar librería como jsPDF
    console.warn('Exportación a PDF requiere librería adicional (jsPDF)');

    // Por ahora, exportamos como texto plano
    const filename = options.filename || `export_${this.getTimestamp()}.txt`;
    const textContent = this.convertToText(data);

    this.downloadFile(textContent, filename, 'text/plain;charset=utf-8;');
  }

  /**
   * Exporta tabla HTML a CSV
   */
  exportTableToCSV(tableId: string, options: ExportOptions = {}): void {
    const table = document.getElementById(tableId) as HTMLTableElement;
    if (!table) {
      console.error(`Tabla con ID "${tableId}" no encontrada`);
      return;
    }

    const filename = options.filename || `table_export_${this.getTimestamp()}.csv`;
    let csvContent = '';

    // Procesar filas
    const rows = table.querySelectorAll('tr');
    rows.forEach((row, index) => {
      const cells = row.querySelectorAll('th, td');
      const values: string[] = [];

      cells.forEach(cell => {
        const text = cell.textContent?.trim() || '';
        values.push(this.escapeCSV(text));
      });

      if (values.length > 0) {
        csvContent += values.join(',') + '\n';
      }
    });

    this.downloadFile(csvContent, filename, 'text/csv;charset=utf-8;');
  }

  /**
   * Imprime datos (abre diálogo de impresión)
   */
  print(data: any[], title?: string): void {
    const printWindow = window.open('', '_blank');
    if (!printWindow) {
      console.error('No se pudo abrir ventana de impresión');
      return;
    }

    const html = this.generatePrintHTML(data, title);
    printWindow.document.write(html);
    printWindow.document.close();

    // Esperar a que se cargue el contenido antes de imprimir
    printWindow.onload = () => {
      printWindow.print();
    };
  }

  /**
   * Escapa caracteres especiales para CSV
   */
  private escapeCSV(value: string): string {
    if (typeof value !== 'string') {
      value = String(value);
    }

    // Si contiene coma, comillas o saltos de línea, envolver en comillas
    if (value.includes(',') || value.includes('"') || value.includes('\n')) {
      // Escapar comillas dobles
      value = value.replace(/"/g, '""');
      // Envolver en comillas
      return `"${value}"`;
    }

    return value;
  }

  /**
   * Formatea un valor para exportación
   */
  private formatValue(value: any): string {
    if (value === null || value === undefined) {
      return '';
    }

    if (value instanceof Date) {
      return value.toISOString().split('T')[0]; // YYYY-MM-DD
    }

    if (typeof value === 'object') {
      return JSON.stringify(value);
    }

    return String(value);
  }

  /**
   * Convierte datos a texto plano
   */
  private convertToText(data: any[]): string {
    let text = '';

    data.forEach((item, index) => {
      text += `--- Registro ${index + 1} ---\n`;

      Object.keys(item).forEach(key => {
        text += `${key}: ${this.formatValue(item[key])}\n`;
      });

      text += '\n';
    });

    return text;
  }

  /**
   * Genera HTML para impresión
   */
  private generatePrintHTML(data: any[], title?: string): string {
    const headers = data.length > 0 ? Object.keys(data[0]) : [];

    let html = `
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="utf-8">
        <title>${title || 'Impresión'}</title>
        <style>
          body {
            font-family: Arial, sans-serif;
            margin: 20px;
          }
          h1 {
            color: #8b4049;
            border-bottom: 2px solid #8b4049;
            padding-bottom: 10px;
          }
          table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
          }
          th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
          }
          th {
            background-color: #8b4049;
            color: white;
          }
          tr:nth-child(even) {
            background-color: #f5f5f5;
          }
          @media print {
            button {
              display: none;
            }
          }
        </style>
      </head>
      <body>
        <h1>${title || 'Datos Exportados'}</h1>
        <p>Fecha: ${new Date().toLocaleString('es-ES')}</p>
        <table>
          <thead>
            <tr>
              ${headers.map(h => `<th>${h}</th>`).join('')}
            </tr>
          </thead>
          <tbody>
            ${data.map(row => `
              <tr>
                ${headers.map(h => `<td>${this.formatValue(row[h])}</td>`).join('')}
              </tr>
            `).join('')}
          </tbody>
        </table>
      </body>
      </html>
    `;

    return html;
  }

  /**
   * Descarga un archivo
   */
  private downloadFile(content: string, filename: string, mimeType: string): void {
    const blob = new Blob([content], { type: mimeType });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');

    link.href = url;
    link.download = filename;
    link.style.display = 'none';

    document.body.appendChild(link);
    link.click();

    // Limpiar
    setTimeout(() => {
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    }, 100);
  }

  /**
   * Genera timestamp para nombres de archivo
   */
  private getTimestamp(): string {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');

    return `${year}${month}${day}_${hours}${minutes}${seconds}`;
  }
}
