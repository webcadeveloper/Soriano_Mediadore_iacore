import { Routes } from '@angular/router';
import { AuthGuard } from './core/guards/auth.guard';
import { RoleGuard } from './core/guards/role.guard';

/**
 * Configuración de rutas con lazy loading y metadata
 * - Todas las rutas usan lazy loading para optimizar bundle inicial
 * - data.preload: true marca rutas para precarga selectiva
 * - data.title: título de la página para SEO
 * - data.roles: roles requeridos para acceso
 */
export const routes: Routes = [
  // Login (pública)
  {
    path: 'login',
    loadComponent: () => import('./pages/login/login.component').then(m => m.LoginComponent),
    data: {
      title: 'Iniciar Sesión - Soriano Mediadores'
    }
  },

  // Redirigir raíz a dashboard
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full'
  },

  // Rutas protegidas con autenticación
  {
    path: 'dashboard',
    loadComponent: () => import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Dashboard - Soriano Mediadores',
      preload: true, // Precargar esta ruta (muy frecuente)
      preloadDelay: 1000
    }
  },
  {
    path: 'clientes',
    loadComponent: () => import('./pages/clientes/clientes.component').then(m => m.ClientesComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Clientes - Soriano Mediadores',
      preload: true, // Precargar (frecuente)
      preloadDelay: 2000
    }
  },
  {
    path: 'clientes/:id',
    loadComponent: () => import('./pages/cliente-detalle/cliente-detalle.component').then(m => m.ClienteDetalleComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Detalle de Cliente - Soriano Mediadores'
    }
  },
  {
    path: 'bots',
    loadComponent: () => import('./pages/bots/bots.component').then(m => m.BotsComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Bots AI - Soriano Mediadores'
    }
  },
  {
    path: 'reportes',
    loadComponent: () => import('./pages/reportes/reportes.component').then(m => m.ReportesComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Reportes - Soriano Mediadores'
    }
  },
  {
    path: 'recobros',
    loadComponent: () => import('./pages/recobros/recobros.component').then(m => m.RecobrosComponent),
    canActivate: [AuthGuard],
    data: {
      title: 'Recobros - Soriano Mediadores',
      preload: true, // Precargar (frecuente)
      preloadDelay: 2500
    }
  },

  // Rutas de admin protegidas con rol
  {
    path: 'admin/import',
    loadComponent: () => import('./pages/admin-import/admin-import.component').then(m => m.AdminImportComponent),
    canActivate: [AuthGuard, RoleGuard],
    data: {
      title: 'Importar Datos - Soriano Mediadores',
      roles: ['admin', 'director', 'supervisor']
    }
  },

  // Ruta 404 (opcional)
  {
    path: '**',
    redirectTo: '/dashboard'
  }
];
