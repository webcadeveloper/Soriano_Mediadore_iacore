import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTabsModule } from '@angular/material/tabs';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDividerModule } from '@angular/material/divider';
import { ApiService } from '../../core/services/api.service';
import { Cliente } from '../../shared/models/cliente.model';

interface Poliza {
  id: number;
  numero_poliza: string;
  ramo: string;
  compania: string;
  producto: string;
  situacion: string;
  prima_anual: string;
  fecha_efecto: string;
  fecha_vencimiento: string;
}

@Component({
  selector: 'app-cliente-detalle',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTabsModule,
    MatTableModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatDividerModule
  ],
  templateUrl: './cliente-detalle.component.html',
  styleUrl: './cliente-detalle.component.scss'
})
export class ClienteDetalleComponent implements OnInit {
  clienteId: string = '';
  cliente: Cliente | null = null;
  polizas: Poliza[] = [];
  loading = true;
  error = false;

  displayedColumns: string[] = ['numero_poliza', 'ramo', 'compania', 'producto', 'situacion', 'prima_anual', 'vencimiento'];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService
  ) {}

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.clienteId = params['id'];
      if (this.clienteId) {
        this.cargarCliente();
        this.cargarPolizas();
      }
    });
  }

  cargarCliente(): void {
    this.loading = true;
    this.apiService.getCliente(this.clienteId).subscribe({
      next: (data: any) => {
        // El backend devuelve cliente con polizas incluidas
        this.cliente = data;
        // Las pólizas vienen en el array 'polizas' de la respuesta
        if (data.polizas && Array.isArray(data.polizas)) {
          this.polizas = data.polizas;
        }
        this.loading = false;
        console.log('Cliente cargado:', this.cliente);
        console.log('Pólizas cargadas:', this.polizas);
      },
      error: (err) => {
        console.error('Error cargando cliente:', err);
        this.error = true;
        this.loading = false;
      }
    });
  }

  cargarPolizas(): void {
    // Las pólizas ya vienen con el cliente, pero si no están
    // hacer llamada separada al endpoint de pólizas
    if (!this.polizas || this.polizas.length === 0) {
      this.apiService.getPolizasCliente(this.clienteId).subscribe({
        next: (response: any) => {
          this.polizas = response.polizas || [];
          console.log('Pólizas cargadas desde endpoint separado:', this.polizas);
        },
        error: (err) => {
          console.error('Error cargando pólizas:', err);
        }
      });
    }
  }

  volver(): void {
    this.router.navigate(['/clientes']);
  }

  getSituacionColor(situacion: string): string {
    const situacionLower = situacion?.toLowerCase() || '';
    if (situacionLower.includes('vigente') || situacionLower.includes('activ')) {
      return 'primary';
    } else if (situacionLower.includes('anulad') || situacionLower.includes('cancelad')) {
      return 'warn';
    }
    return '';
  }

  formatCurrency(value: string): string {
    const num = parseFloat(value);
    if (isNaN(num)) return value;
    return new Intl.NumberFormat('es-ES', {
      style: 'currency',
      currency: 'EUR'
    }).format(num);
  }

  formatDate(date: string): string {
    if (!date) return '-';
    return new Date(date).toLocaleDateString('es-ES');
  }
}
