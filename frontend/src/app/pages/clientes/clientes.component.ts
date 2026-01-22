import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { Subject, takeUntil } from 'rxjs';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';
import { MatBadgeModule } from '@angular/material/badge';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ApiService } from '../../core/services/api.service';
import { Cliente } from '../../shared/models/cliente.model';

@Component({
  selector: 'app-clientes',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatTableModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatCardModule,
    MatDividerModule,
    MatBadgeModule,
    MatChipsModule,
    MatTooltipModule
  ],
  templateUrl: './clientes.component.html',
  styleUrl: './clientes.component.scss'
})
export class ClientesComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  searchQuery = '';
  clientes: Cliente[] = [];
  loading = false;
  searched = false;
  displayedColumns: string[] = ['nif', 'nombre_completo', 'email', 'telefono', 'provincia', 'total_primas', 'acciones'];

  constructor(private apiService: ApiService) {}

  ngOnInit(): void {
    // Cargar TODOS los clientes al inicio
    this.cargarTodosLosClientes();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  cargarTodosLosClientes(): void {
    this.loading = true;
    this.searched = true;

    // Query vacío = obtener todos los clientes
    this.apiService.buscarClientes('')
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.clientes = response.clientes;
          this.loading = false;
        },
        error: (err) => {
          console.error('Error cargando clientes:', err);
          this.loading = false;
          this.clientes = [];
        }
      });
  }

  buscar(query?: string): void {
    const searchTerm = query !== undefined ? query : this.searchQuery;

    // Si el término está vacío, cargar todos
    if (!searchTerm || searchTerm.trim() === '') {
      this.cargarTodosLosClientes();
      return;
    }

    // Si es muy corto, no buscar (pero permitir vacío para "todos")
    if (searchTerm.length < 2) {
      return;
    }

    this.loading = true;
    this.searched = true;

    this.apiService.buscarClientes(searchTerm)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.clientes = response.clientes;
          this.loading = false;
        },
        error: (err) => {
          console.error('Error buscando clientes:', err);
          this.loading = false;
          this.clientes = [];
        }
      });
  }

  onSearch(): void {
    this.buscar();
  }

  clearSearch(): void {
    this.searchQuery = '';
    // Al limpiar, recargar todos los clientes
    this.cargarTodosLosClientes();
  }

  // TrackBy function for performance optimization
  trackByClienteId(index: number, cliente: Cliente): string {
    return cliente.nif || `cliente-${index}`;
  }
}
