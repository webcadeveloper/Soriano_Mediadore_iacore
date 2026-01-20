import { Component, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatGridListModule } from '@angular/material/grid-list';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { provideHttpClient } from '@angular/common/http';
import { Subject, takeUntil } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Stats } from '../../shared/models/stats.model';
import { Cliente } from '../../shared/models/cliente.model';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatCardModule,
    MatGridListModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatButtonModule,
    MatTableModule,
    MatChipsModule,
    MatTooltipModule
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class DashboardComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  stats: Stats | null = null;
  recibosKPI: any = null;  // KPIs de recibos
  statsGeneral: any = null; // Estadísticas generales
  recentClientes: Cliente[] = [];
  loading = true;
  loadingClientes = true;
  loadingRecibosKPI = true;
  error: string | null = null;

  displayedColumns: string[] = [
    'nombre_completo',
    'nif',
    'email',
    'telefono',
    'provincia',
    'total_primas',
    'acciones'
  ];

  constructor(private apiService: ApiService) {}

  ngOnInit(): void {
    this.loadStats();
    this.loadStatsGeneral();
    this.loadRecibosKPI();
    this.loadRecentClientes();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadStats(): void {
    this.loading = true;
    this.apiService.getStats()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.stats = response.estadisticas;
          this.loading = false;
        },
        error: (err) => {
          this.error = 'Error cargando estadísticas';
          this.loading = false;
          console.error('Error:', err);
        }
      });
  }

  loadStatsGeneral(): void {
    this.apiService.getStatsGeneral()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.statsGeneral = response.data;
        },
        error: (err) => {
          console.error('Error cargando estadísticas generales:', err);
        }
      });
  }

  loadRecibosKPI(): void {
    this.loadingRecibosKPI = true;
    this.apiService.getRecibosKPI({ situacion: 'Retornado', limite: 200 })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.recibosKPI = response.data;
          this.loadingRecibosKPI = false;
        },
        error: (err) => {
          console.error('Error cargando recibos KPI:', err);
          this.loadingRecibosKPI = false;
        }
      });
  }

  loadRecentClientes(): void {
    this.loadingClientes = true;
    // Get all clientes with empty search to get first 20
    this.apiService.buscarClientes('')
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.recentClientes = response.clientes || [];
          this.loadingClientes = false;
        },
        error: (err) => {
          console.error('Error cargando clientes:', err);
          this.loadingClientes = false;
        }
      });
  }

  formatCurrency(value: number): string {
    return new Intl.NumberFormat('es-ES', {
      style: 'currency',
      currency: 'EUR'
    }).format(value);
  }

  // TrackBy function for performance optimization
  trackByClienteId(index: number, cliente: Cliente): string {
    return cliente.nif || `cliente-${index}`;
  }
}
