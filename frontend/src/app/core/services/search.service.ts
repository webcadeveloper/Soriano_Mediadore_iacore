import { Injectable } from '@angular/core';
import { Observable, of, BehaviorSubject } from 'rxjs';
import { map, debounceTime, distinctUntilChanged, switchMap } from 'rxjs/operators';

/**
 * Interfaz para resultados de búsqueda
 */
export interface SearchResult {
  id: string;
  type: 'cliente' | 'recobro' | 'reporte' | 'bot';
  title: string;
  subtitle?: string;
  description?: string;
  icon: string;
  route: string;
  relevance: number;
}

/**
 * Servicio de búsqueda global
 * Permite buscar en todas las entidades del sistema
 */
@Injectable({
  providedIn: 'root'
})
export class SearchService {
  private searchHistorySubject = new BehaviorSubject<string[]>([]);
  public searchHistory$: Observable<string[]> = this.searchHistorySubject.asObservable();

  private recentSearchesKey = 'recent_searches';
  private maxHistoryItems = 10;

  constructor() {
    this.loadSearchHistory();
  }

  /**
   * Busca en todas las entidades
   */
  search(query: string): Observable<SearchResult[]> {
    if (!query || query.trim().length < 2) {
      return of([]);
    }

    const normalizedQuery = query.toLowerCase().trim();

    // Simular búsqueda (en producción esto haría llamadas a la API)
    return of(this.mockSearch(normalizedQuery));
  }

  /**
   * Búsqueda con debounce para input en tiempo real
   */
  searchWithDebounce(query$: Observable<string>): Observable<SearchResult[]> {
    return query$.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      switchMap(query => this.search(query))
    );
  }

  /**
   * Añade una búsqueda al historial
   */
  addToHistory(query: string): void {
    if (!query || query.trim().length < 2) {
      return;
    }

    let history = this.searchHistorySubject.value;

    // Eliminar si ya existe
    history = history.filter(item => item !== query);

    // Añadir al principio
    history.unshift(query);

    // Limitar tamaño
    if (history.length > this.maxHistoryItems) {
      history = history.slice(0, this.maxHistoryItems);
    }

    this.searchHistorySubject.next(history);
    this.saveSearchHistory();
  }

  /**
   * Limpia el historial de búsqueda
   */
  clearHistory(): void {
    this.searchHistorySubject.next([]);
    localStorage.removeItem(this.recentSearchesKey);
  }

  /**
   * Elimina un item del historial
   */
  removeFromHistory(query: string): void {
    const history = this.searchHistorySubject.value.filter(item => item !== query);
    this.searchHistorySubject.next(history);
    this.saveSearchHistory();
  }

  /**
   * Busca clientes
   */
  searchClientes(query: string): Observable<SearchResult[]> {
    return this.search(query).pipe(
      map(results => results.filter(r => r.type === 'cliente'))
    );
  }

  /**
   * Busca recobros
   */
  searchRecobros(query: string): Observable<SearchResult[]> {
    return this.search(query).pipe(
      map(results => results.filter(r => r.type === 'recobro'))
    );
  }

  /**
   * Mock de búsqueda (reemplazar con llamadas reales a la API)
   */
  private mockSearch(query: string): SearchResult[] {
    const results: SearchResult[] = [];

    // Mock de clientes
    const mockClientes = [
      { id: '1', nombre: 'Juan Pérez García', nif: '12345678A', telefono: '600123456' },
      { id: '2', nombre: 'María González López', nif: '87654321B', telefono: '600987654' },
      { id: '3', nombre: 'Carlos Rodríguez Martín', nif: '11223344C', telefono: '600555444' },
      { id: '4', nombre: 'Ana Fernández Sánchez', nif: '55667788D', telefono: '600111222' },
      { id: '5', nombre: 'Pedro Martínez Díaz', nif: '99887766E', telefono: '600333444' }
    ];

    mockClientes.forEach(cliente => {
      if (
        cliente.nombre.toLowerCase().includes(query) ||
        cliente.nif.toLowerCase().includes(query) ||
        cliente.telefono.includes(query)
      ) {
        results.push({
          id: cliente.id,
          type: 'cliente',
          title: cliente.nombre,
          subtitle: cliente.nif,
          description: cliente.telefono,
          icon: 'person',
          route: `/clientes/${cliente.id}`,
          relevance: this.calculateRelevance(query, cliente.nombre)
        });
      }
    });

    // Mock de recobros
    const mockRecobros = [
      { id: '1', recibo: 'REC-2024-001', cliente: 'Juan Pérez García', importe: 1250.50 },
      { id: '2', recibo: 'REC-2024-002', cliente: 'María González López', importe: 850.00 },
      { id: '3', recibo: 'REC-2024-003', cliente: 'Carlos Rodríguez Martín', importe: 2100.75 }
    ];

    mockRecobros.forEach(recobro => {
      if (
        recobro.recibo.toLowerCase().includes(query) ||
        recobro.cliente.toLowerCase().includes(query)
      ) {
        results.push({
          id: recobro.id,
          type: 'recobro',
          title: recobro.recibo,
          subtitle: recobro.cliente,
          description: `${recobro.importe.toFixed(2)} €`,
          icon: 'account_balance_wallet',
          route: `/recobros`,
          relevance: this.calculateRelevance(query, recobro.recibo)
        });
      }
    });

    // Mock de reportes
    if ('reportes'.includes(query) || 'informes'.includes(query)) {
      results.push({
        id: 'reportes',
        type: 'reporte',
        title: 'Reportes',
        subtitle: 'Ver todos los reportes',
        icon: 'assessment',
        route: '/reportes',
        relevance: 0.8
      });
    }

    // Mock de bots
    if ('bots'.includes(query) || 'asistente'.includes(query)) {
      results.push({
        id: 'bots',
        type: 'bot',
        title: 'Bots AI',
        subtitle: 'Asistentes virtuales',
        icon: 'smart_toy',
        route: '/bots',
        relevance: 0.8
      });
    }

    // Ordenar por relevancia
    return results.sort((a, b) => b.relevance - a.relevance);
  }

  /**
   * Calcula la relevancia de un resultado
   */
  private calculateRelevance(query: string, text: string): number {
    const normalizedText = text.toLowerCase();
    const normalizedQuery = query.toLowerCase();

    // Coincidencia exacta
    if (normalizedText === normalizedQuery) {
      return 1.0;
    }

    // Empieza con la query
    if (normalizedText.startsWith(normalizedQuery)) {
      return 0.9;
    }

    // Contiene la query
    if (normalizedText.includes(normalizedQuery)) {
      return 0.7;
    }

    // Palabras individuales
    const queryWords = normalizedQuery.split(' ');
    const matchedWords = queryWords.filter(word => normalizedText.includes(word));
    const wordMatchRatio = matchedWords.length / queryWords.length;

    return wordMatchRatio * 0.6;
  }

  /**
   * Guarda el historial en localStorage
   */
  private saveSearchHistory(): void {
    try {
      const history = this.searchHistorySubject.value;
      localStorage.setItem(this.recentSearchesKey, JSON.stringify(history));
    } catch (error) {
      console.error('Error saving search history:', error);
    }
  }

  /**
   * Carga el historial desde localStorage
   */
  private loadSearchHistory(): void {
    try {
      const stored = localStorage.getItem(this.recentSearchesKey);
      if (stored) {
        const history: string[] = JSON.parse(stored);
        this.searchHistorySubject.next(history);
      }
    } catch (error) {
      console.error('Error loading search history:', error);
    }
  }
}
