import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  {
    path: 'dashboard',
    loadComponent: () => import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'clientes',
    loadComponent: () => import('./pages/clientes/clientes.component').then(m => m.ClientesComponent)
  },
  {
    path: 'clientes/:id',
    loadComponent: () => import('./pages/cliente-detalle/cliente-detalle.component').then(m => m.ClienteDetalleComponent)
  },
  {
    path: 'bots',
    loadComponent: () => import('./pages/bots/bots.component').then(m => m.BotsComponent)
  },
  {
    path: 'reportes',
    loadComponent: () => import('./pages/reportes/reportes.component').then(m => m.ReportesComponent)
  },
  {
    path: 'admin/import',
    loadComponent: () => import('./pages/admin-import/admin-import.component').then(m => m.AdminImportComponent)
  },
  {
    path: 'recobros',
    loadComponent: () => import('./pages/recobros/recobros.component').then(m => m.RecobrosComponent)
  }
];
