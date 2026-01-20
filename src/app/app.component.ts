import { Component, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule, MatSidenav } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatSidenavModule,
    MatListModule,
    MatTooltipModule,
    MatMenuModule,
    MatDividerModule
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  @ViewChild('sidenav') sidenav!: MatSidenav;

  title = 'Soriano Mediadores';
  isMobile = false;
  isCollapsed = false;
  currentYear = new Date().getFullYear();

  menuItems = [
    { path: '/dashboard', icon: 'dashboard', label: 'Dashboard', tooltip: 'Panel de control' },
    { path: '/clientes', icon: 'people', label: 'Clientes', tooltip: 'Gestión de clientes' },
    { path: '/bots', icon: 'smart_toy', label: 'Bots AI', tooltip: 'Asistentes virtuales' },
    { path: '/recobros', icon: 'account_balance_wallet', label: 'Recobros', tooltip: 'Gestión de recibos impagados' },
    { path: '/reportes', icon: 'assessment', label: 'Reportes', tooltip: 'Informes y estadísticas' },
    { path: '/admin/import', icon: 'upload_file', label: 'Importar Datos', tooltip: 'Importación de CSV' }
  ];

  constructor(private breakpointObserver: BreakpointObserver) {
    this.breakpointObserver.observe([Breakpoints.Handset, Breakpoints.Tablet])
      .subscribe(result => {
        this.isMobile = result.matches;
        if (this.isMobile) {
          this.isCollapsed = false;
        }
      });
  }

  toggleSidenav(): void {
    if (this.isMobile) {
      this.sidenav.toggle();
    } else {
      this.isCollapsed = !this.isCollapsed;
    }
  }

  closeSidenavOnMobile(): void {
    if (this.isMobile && this.sidenav) {
      this.sidenav.close();
    }
  }
}
