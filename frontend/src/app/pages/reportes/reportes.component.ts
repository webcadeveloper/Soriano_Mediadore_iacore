import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subject, forkJoin, takeUntil } from 'rxjs';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatTabsModule } from '@angular/material/tabs';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { BaseChartDirective } from 'ng2-charts';
import { ChartConfiguration, ChartData, ChartType } from 'chart.js';
import { ApiService } from '../../core/services/api.service';
import { AnalyticsService } from '../../core/services/analytics.service';
import {
  FinancialKPIsResponse,
  PortfolioAnalysisResponse,
  CollectionsPerformanceResponse,
  ClaimsAnalysisResponse,
  PerformanceTrendsResponse
} from '../../shared/models/analytics.model';

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
    MatTabsModule,
    MatButtonModule,
    MatTooltipModule,
    BaseChartDirective
  ],
  templateUrl: './reportes.component.html',
  styleUrl: './reportes.component.scss'
})
export class ReportesComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  loading = false;
  topClientes: TopCliente[] = [];
  analisisPorRamo: AnalisisPorRamo[] = [];

  displayedColumnsClientes: string[] = ['posicion', 'nombre_completo', 'nif', 'num_polizas', 'total_primas'];
  displayedColumnsRamos: string[] = ['ramo', 'num_polizas'];

  // Stats básicos (del endpoint antiguo)
  totalClientes = 0;
  totalPolizas = 0;
  totalRecibos = 0;
  totalSiniestros = 0;

  // Datos de Analytics (nuevos endpoints)
  financialKPIs: FinancialKPIsResponse | null = null;
  portfolioAnalysis: PortfolioAnalysisResponse | null = null;
  collectionsPerformance: CollectionsPerformanceResponse | null = null;
  claimsAnalysis: ClaimsAnalysisResponse | null = null;
  performanceTrends: PerformanceTrendsResponse | null = null;

  // Gráficos Chart.js
  // 1. Evolución Mensual de Primas (Tab 1 - Resumen Ejecutivo)
  primasEvolucionChartData: ChartData<'line'> = { labels: [], datasets: [] };
  primasEvolucionChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true, position: 'top' },
      tooltip: { mode: 'index', intersect: false }
    },
    scales: {
      x: { grid: { display: false } },
      y: { beginAtZero: true, ticks: { callback: (value) => this.formatCurrency(Number(value)) } }
    }
  };
  primasEvolucionChartType: ChartType = 'line';

  // 2. Top 10 Clientes (Gráfico de Barras)
  topClientesChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  topClientesChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    indexAxis: 'y',
    plugins: {
      legend: { display: false },
      tooltip: { callbacks: { label: (context) => this.formatCurrency(context.parsed.x ?? 0) } }
    },
    scales: {
      x: { ticks: { callback: (value) => this.formatCurrency(Number(value)) } }
    }
  };
  topClientesChartType: ChartType = 'bar';

  // 3. Distribución por Ramo (Pie Chart)
  distribucionRamoChartData: ChartData<'pie'> = { labels: [], datasets: [] };
  distribucionRamoChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true, position: 'right' },
      tooltip: {
        callbacks: {
          label: (context) => {
            const label = context.label || '';
            const value = this.formatCurrency(context.parsed);
            const percentage = ((context.parsed / (context.dataset.data as number[]).reduce((a, b) => a + b, 0)) * 100).toFixed(1);
            return `${label}: ${value} (${percentage}%)`;
          }
        }
      }
    }
  };
  distribucionRamoChartType: ChartType = 'pie';

  // 4. Pocket Share por Compañía (Doughnut Chart)
  pocketShareChartData: ChartData<'doughnut'> = { labels: [], datasets: [] };
  pocketShareChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true, position: 'right' },
      tooltip: {
        callbacks: {
          label: (context) => {
            const label = context.label || '';
            const value = this.formatCurrency(context.parsed);
            const data = this.financialKPIs?.pocket_share_por_compania.find(item => item.gestora === label);
            const percentage = data?.porcentaje.toFixed(1) || '0';
            return `${label}: ${value} (${percentage}%)`;
          }
        }
      }
    }
  };
  pocketShareChartType: ChartType = 'doughnut';

  // 5. Morosidad por Rango de Días (Bar Chart)
  morosidadRangoChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  morosidadRangoChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => {
            const value = this.formatCurrency(context.parsed.y ?? 0);
            const recibos = this.collectionsPerformance?.morosidad_por_rango_dias[context.dataIndex].num_recibos || 0;
            return [`Importe: ${value}`, `Recibos: ${recibos}`];
          }
        }
      }
    },
    scales: {
      y: { beginAtZero: true, ticks: { callback: (value) => this.formatCurrency(Number(value)) } }
    }
  };
  morosidadRangoChartType: ChartType = 'bar';

  // 6. Siniestralidad por Ramo (Bar Chart)
  siniestralididadRamoChartData: ChartData<'bar'> = { labels: [], datasets: [] };
  siniestralididadRamoChartOptions: ChartConfiguration['options'] = {
    responsive: true,
    maintainAspectRatio: false,
    indexAxis: 'y',
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => {
            const data = this.claimsAnalysis?.siniestralidad_por_ramo[context.dataIndex];
            return [
              `Siniestralidad: ${(context.parsed.x ?? 0).toFixed(2)}%`,
              `Siniestros: ${data?.num_siniestros || 0}`,
              `Pólizas: ${data?.num_polizas || 0}`
            ];
          }
        }
      }
    },
    scales: {
      x: { beginAtZero: true, ticks: { callback: (value) => `${value}%` } }
    }
  };
  siniestralididadRamoChartType: ChartType = 'bar';

  constructor(
    private apiService: ApiService,
    private analyticsService: AnalyticsService
  ) {}

  ngOnInit(): void {
    this.cargarTodosLosDatos();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  cargarTodosLosDatos(): void {
    this.loading = true;

    // Cargar stats básicos + todos los endpoints de analytics en paralelo
    forkJoin({
      stats: this.apiService.getStats(),
      financialKPIs: this.analyticsService.getFinancialKPIs(),
      portfolioAnalysis: this.analyticsService.getPortfolioAnalysis(),
      collectionsPerformance: this.analyticsService.getCollectionsPerformance(),
      claimsAnalysis: this.analyticsService.getClaimsAnalysis(),
      performanceTrends: this.analyticsService.getPerformanceTrends('30days')
    })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          // Stats básicos
          const stats = response.stats.estadisticas;
          this.totalClientes = stats.total_clientes;
          this.totalPolizas = stats.total_polizas;
          this.totalRecibos = stats.total_recibos;
          this.totalSiniestros = stats.total_siniestros;
          this.topClientes = stats.top_20_clientes || [];
          this.analisisPorRamo = stats.analisis_por_ramo || [];

          // Analytics data
          this.financialKPIs = response.financialKPIs;
          this.portfolioAnalysis = response.portfolioAnalysis;
          this.collectionsPerformance = response.collectionsPerformance;
          this.claimsAnalysis = response.claimsAnalysis;
          this.performanceTrends = response.performanceTrends;

          // Preparar gráficos
          this.prepararGraficos();

          this.loading = false;
        },
        error: (err) => {
          console.error('Error cargando datos:', err);
          this.loading = false;
        }
      });
  }

  prepararGraficos(): void {
    // 1. Evolución Mensual de Primas
    if (this.financialKPIs?.evolucion_mensual_primas) {
      this.primasEvolucionChartData = {
        labels: this.financialKPIs.evolucion_mensual_primas.map(item => item.mes),
        datasets: [
          {
            label: 'Primas Totales',
            data: this.financialKPIs.evolucion_mensual_primas.map(item => item.total_primas),
            borderColor: '#c2185b',
            backgroundColor: 'rgba(194, 24, 91, 0.1)',
            tension: 0.4,
            fill: true
          },
          {
            label: 'Número de Pólizas',
            data: this.financialKPIs.evolucion_mensual_primas.map(item => item.num_polizas * 100), // Escalar para visualización
            borderColor: '#1976d2',
            backgroundColor: 'rgba(25, 118, 210, 0.1)',
            tension: 0.4,
            fill: false,
            yAxisID: 'y1'
          }
        ]
      };

      // Agregar segundo eje Y para número de pólizas
      if (this.primasEvolucionChartOptions && this.primasEvolucionChartOptions.scales) {
        this.primasEvolucionChartOptions.scales['y1'] = {
          type: 'linear',
          position: 'right',
          grid: { drawOnChartArea: false },
          ticks: { callback: (value) => Math.round(Number(value) / 100).toString() }
        };
      }
    }

    // 2. Top 10 Clientes (Barras horizontales)
    const top10 = this.topClientes.slice(0, 10);
    this.topClientesChartData = {
      labels: top10.map(c => c.nombre_completo.length > 25 ? c.nombre_completo.substring(0, 25) + '...' : c.nombre_completo),
      datasets: [
        {
          label: 'Total Primas',
          data: top10.map(c => c.total_primas),
          backgroundColor: [
            '#c2185b', '#c2185b', '#c2185b', // Top 3 en color principal
            '#e91e63', '#e91e63', '#e91e63', '#e91e63', '#e91e63', // Top 4-8 en variante
            '#f06292', '#f06292' // Top 9-10 en variante clara
          ]
        }
      ]
    };

    // 3. Distribución por Ramo (Pie Chart)
    if (this.portfolioAnalysis?.distribucion_por_ramo) {
      const coloresPie = [
        '#c2185b', '#1976d2', '#388e3c', '#f57c00', '#7b1fa2',
        '#0097a7', '#c62828', '#5d4037', '#455a64', '#616161'
      ];

      this.distribucionRamoChartData = {
        labels: this.portfolioAnalysis.distribucion_por_ramo.map(item => item.ramo),
        datasets: [
          {
            data: this.portfolioAnalysis.distribucion_por_ramo.map(item => item.total_primas),
            backgroundColor: coloresPie,
            hoverOffset: 10
          }
        ]
      };
    }

    // 4. Pocket Share por Compañía (Doughnut Chart)
    if (this.financialKPIs?.pocket_share_por_compania) {
      const coloresDoughnut = [
        '#c2185b', '#e91e63', '#f06292', '#f48fb1', '#f8bbd0',
        '#1976d2', '#42a5f5', '#90caf9', '#bbdefb', '#e3f2fd'
      ];

      this.pocketShareChartData = {
        labels: this.financialKPIs.pocket_share_por_compania.map(item => item.gestora),
        datasets: [
          {
            data: this.financialKPIs.pocket_share_por_compania.map(item => item.total_primas),
            backgroundColor: coloresDoughnut,
            hoverOffset: 15
          }
        ]
      };
    }

    // 5. Morosidad por Rango de Días
    if (this.collectionsPerformance?.morosidad_por_rango_dias) {
      this.morosidadRangoChartData = {
        labels: this.collectionsPerformance.morosidad_por_rango_dias.map(item => item.rango),
        datasets: [
          {
            label: 'Importe Total',
            data: this.collectionsPerformance.morosidad_por_rango_dias.map(item => item.importe_total),
            backgroundColor: ['#4caf50', '#ff9800', '#ff5722', '#d32f2f'], // Verde -> Amarillo -> Naranja -> Rojo
            borderWidth: 0
          }
        ]
      };
    }

    // 6. Siniestralidad por Ramo
    if (this.claimsAnalysis?.siniestralidad_por_ramo) {
      this.siniestralididadRamoChartData = {
        labels: this.claimsAnalysis.siniestralidad_por_ramo.map(item => item.ramo),
        datasets: [
          {
            label: 'Siniestralidad (%)',
            data: this.claimsAnalysis.siniestralidad_por_ramo.map(item => item.siniestralidad),
            backgroundColor: this.claimsAnalysis.siniestralidad_por_ramo.map(item =>
              item.siniestralidad > 10 ? '#d32f2f' : // Rojo si > 10%
              item.siniestralidad > 5 ? '#ff9800' : // Naranja si > 5%
              '#4caf50' // Verde si <= 5%
            )
          }
        ]
      };
    }
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

  formatPercentage(value: number): string {
    return `${value.toFixed(2)}%`;
  }

  // Calcular comparativas (% cambio)
  getPrimasChangePercentage(): string {
    if (!this.financialKPIs) return '-';
    const actual = this.financialKPIs.total_primas_mes_actual;
    const anterior = this.financialKPIs.total_primas_mes_anterior;
    if (anterior === 0) return '-';
    const cambio = ((actual - anterior) / anterior) * 100;
    return this.formatPercentage(cambio);
  }

  isPrimasIncreasing(): boolean {
    if (!this.financialKPIs) return false;
    return this.financialKPIs.total_primas_mes_actual > this.financialKPIs.total_primas_mes_anterior;
  }

  getRatioCobro(): string {
    if (!this.collectionsPerformance) return '-';
    return this.formatPercentage(this.collectionsPerformance.ratio_cobro_mes);
  }

  getMorosidadPercentage(): string {
    if (!this.collectionsPerformance) return '-';
    return this.formatPercentage(this.collectionsPerformance.deuda_total_vs_cartera.porcentaje_morosidad);
  }

  getSiniestrosAbiertos(): number {
    return this.claimsAnalysis?.siniestros_abiertos || 0;
  }

  getConcentracionRiesgo(): string {
    if (!this.portfolioAnalysis) return '-';
    return this.formatPercentage(this.portfolioAnalysis.concentracion_riesgo);
  }
}
