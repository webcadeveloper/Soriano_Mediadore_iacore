import { Injectable, ApplicationRef } from '@angular/core';
import { SwUpdate, VersionReadyEvent } from '@angular/service-worker';
import { concat, interval } from 'rxjs';
import { first, filter } from 'rxjs/operators';

/**
 * Servicio para gestionar actualizaciones de PWA
 * Maneja la detección y aplicación de nuevas versiones de la aplicación
 */
@Injectable({
  providedIn: 'root'
})
export class PwaUpdateService {

  constructor(
    private appRef: ApplicationRef,
    private swUpdate: SwUpdate
  ) {}

  /**
   * Inicializa la comprobación de actualizaciones
   * Comprueba cada 6 horas si hay una nueva versión disponible
   */
  checkForUpdates(): void {
    if (!this.swUpdate.isEnabled) {
      console.log('Service Worker no está habilitado');
      return;
    }

    // Espera a que la app esté estable antes de comprobar actualizaciones
    const appIsStable$ = this.appRef.isStable.pipe(
      first(isStable => isStable === true)
    );

    // Comprueba actualizaciones cada 6 horas
    const everySixHours$ = interval(6 * 60 * 60 * 1000);

    const everySixHoursOnceAppIsStable$ = concat(appIsStable$, everySixHours$);

    everySixHoursOnceAppIsStable$.subscribe(() => {
      this.swUpdate.checkForUpdate().then(updateFound => {
        if (updateFound) {
          console.log('Nueva versión disponible');
        }
      }).catch(err => {
        console.error('Error al comprobar actualizaciones:', err);
      });
    });
  }

  /**
   * Escucha eventos de nueva versión disponible
   * Notifica al usuario y permite actualizar
   */
  listenForUpdates(): void {
    if (!this.swUpdate.isEnabled) {
      return;
    }

    this.swUpdate.versionUpdates
      .pipe(
        filter((evt): evt is VersionReadyEvent => evt.type === 'VERSION_READY')
      )
      .subscribe(evt => {
        console.log('Nueva versión disponible:', evt.latestVersion);

        const updateConfirmed = confirm(
          '¡Hay una nueva versión disponible! ¿Deseas actualizar ahora?'
        );

        if (updateConfirmed) {
          this.activateUpdate();
        }
      });
  }

  /**
   * Activa la actualización y recarga la página
   */
  private activateUpdate(): void {
    this.swUpdate.activateUpdate().then(() => {
      console.log('Actualización activada. Recargando...');
      document.location.reload();
    }).catch(err => {
      console.error('Error al activar actualización:', err);
    });
  }

  /**
   * Maneja errores no recuperables del Service Worker
   */
  handleUnrecoverableState(): void {
    if (!this.swUpdate.isEnabled) {
      return;
    }

    this.swUpdate.unrecoverable.subscribe(event => {
      console.error('Estado no recuperable del Service Worker:', event.reason);

      const reloadConfirmed = confirm(
        'La aplicación ha encontrado un error. ¿Deseas recargar?'
      );

      if (reloadConfirmed) {
        document.location.reload();
      }
    });
  }
}
