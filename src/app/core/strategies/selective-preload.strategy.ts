import { Injectable } from '@angular/core';
import { PreloadingStrategy, Route } from '@angular/router';
import { Observable, of, timer } from 'rxjs';
import { mergeMap } from 'rxjs/operators';

/**
 * Estrategia de precarga selectiva
 * Precarga rutas marcadas con data.preload = true después de un delay
 * Mejora el rendimiento inicial mientras prepara rutas frecuentes
 */
@Injectable({
  providedIn: 'root'
})
export class SelectivePreloadStrategy implements PreloadingStrategy {
  preload(route: Route, load: () => Observable<any>): Observable<any> {
    if (route.data && route.data['preload']) {
      // Delay de 2 segundos antes de precargar
      const delay = route.data['preloadDelay'] || 2000;

      return timer(delay).pipe(
        mergeMap(() => {
          console.log(`Precargando ruta: ${route.path}`);
          return load();
        })
      );
    }

    return of(null);
  }
}
