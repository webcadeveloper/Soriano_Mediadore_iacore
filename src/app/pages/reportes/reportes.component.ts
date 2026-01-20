import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatTabsModule } from '@angular/material/tabs';
import { ApiService } from '../../core/services/api.service';

interface TopCliente {
  nombre_completo: string;
  nif: string;
  total_primas: number;
  num_polizas: number;
}

interface AnalisisPorRamo {
  ramo: string;
  num_polizas: number;
}

@Component({
  selector: 'app-reportes',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatTableModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatTabsModule
  ],
  templateUrl: './reportes.component.html',
  styleUrl: './reportes.component.scss'
})
export class ReportesComponent implements OnInit {
  loading = false;
  topClientes: TopCliente[] = [];
  analisisPorRamo: AnalisisPorRamo[] = [];

  displayedColumnsClientes: string[] = ['posicion', 'nombre_completo', 'nif', 'num_polizas', 'total_primas'];
  displayedColumnsRamos: string[] = ['ramo', 'num_polizas'];

  totalClientes = 0;
  totalPolizas = 0;
  totalRecibos = 0;
  totalSiniestros = 0;

  constructor(private apiService: ApiService) {}

  ngOnInit(): void {
    this.cargarReportes();
  }

  cargarReportes(): void {
    this.loading = true;

    this.apiService.getStats().subscribe({
      next: (response) => {
        const stats = response.estadisticas;

        // Estadísticas generales
        this.totalClientes = stats.total_clientes;
        this.totalPolizas = stats.total_polizas;
        this.totalRecibos = stats.total_recibos;
        this.totalSiniestros = stats.total_siniestros;

        // Top 20 clientes
        if (stats.top_20_clientes) {
          this.topClientes = stats.top_20_clientes;
        }

        // Análisis por ramo
        if (stats.analisis_por_ramo) {
          this.analisisPorRamo = stats.analisis_por_ramo;
        }

        this.loading = false;
      },
      error: (err) => {
        console.error('Error cargando reportes:', err);
        this.loading = false;
      }
    });
  }

  formatCurrency(value: number): string {
    return new Intl.NumberFormat('es-ES', {
      style: 'currency',
      currency: 'EUR'
    }).format(value);
  }

  formatNumber(value: number): string {
    return new Intl.NumberFormat('es-ES').format(value);
  }
}
