import { Injectable } from '@angular/core';
import { MatSnackBar, MatSnackBarConfig } from '@angular/material/snack-bar';
import { BehaviorSubject, Observable } from 'rxjs';

/**
 * Interfaz para notificaciones
 */
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  read: boolean;
  action?: {
    label: string;
    callback: () => void;
  };
}

/**
 * Servicio de notificaciones
 * Gestiona notificaciones toast (snackbar) y notificaciones persistentes
 */
@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private notificationsSubject = new BehaviorSubject<Notification[]>([]);
  public notifications$: Observable<Notification[]> = this.notificationsSubject.asObservable();

  private unreadCountSubject = new BehaviorSubject<number>(0);
  public unreadCount$: Observable<number> = this.unreadCountSubject.asObservable();

  constructor(private snackBar: MatSnackBar) {
    this.loadNotificationsFromStorage();
  }

  /**
   * Muestra una notificación toast (snackbar)
   */
  showToast(
    message: string,
    type: 'info' | 'success' | 'warning' | 'error' = 'info',
    duration: number = 3000
  ): void {
    const config: MatSnackBarConfig = {
      duration,
      horizontalPosition: 'end',
      verticalPosition: 'top',
      panelClass: [`snackbar-${type}`]
    };

    this.snackBar.open(message, 'Cerrar', config);
  }

  /**
   * Muestra notificación de éxito
   */
  success(message: string, duration: number = 3000): void {
    this.showToast(message, 'success', duration);
  }

  /**
   * Muestra notificación de error
   */
  error(message: string, duration: number = 5000): void {
    this.showToast(message, 'error', duration);
  }

  /**
   * Muestra notificación de advertencia
   */
  warning(message: string, duration: number = 4000): void {
    this.showToast(message, 'warning', duration);
  }

  /**
   * Muestra notificación informativa
   */
  info(message: string, duration: number = 3000): void {
    this.showToast(message, 'info', duration);
  }

  /**
   * Añade una notificación persistente
   */
  addNotification(
    title: string,
    message: string,
    type: 'info' | 'success' | 'warning' | 'error' = 'info',
    action?: { label: string; callback: () => void }
  ): void {
    const notification: Notification = {
      id: this.generateId(),
      type,
      title,
      message,
      timestamp: new Date(),
      read: false,
      action
    };

    const notifications = this.notificationsSubject.value;
    notifications.unshift(notification);

    // Limitar a 50 notificaciones
    if (notifications.length > 50) {
      notifications.pop();
    }

    this.notificationsSubject.next(notifications);
    this.updateUnreadCount();
    this.saveNotificationsToStorage();
  }

  /**
   * Marca una notificación como leída
   */
  markAsRead(notificationId: string): void {
    const notifications = this.notificationsSubject.value;
    const notification = notifications.find(n => n.id === notificationId);

    if (notification && !notification.read) {
      notification.read = true;
      this.notificationsSubject.next(notifications);
      this.updateUnreadCount();
      this.saveNotificationsToStorage();
    }
  }

  /**
   * Marca todas las notificaciones como leídas
   */
  markAllAsRead(): void {
    const notifications = this.notificationsSubject.value;
    notifications.forEach(n => n.read = true);
    this.notificationsSubject.next(notifications);
    this.unreadCountSubject.next(0);
    this.saveNotificationsToStorage();
  }

  /**
   * Elimina una notificación
   */
  deleteNotification(notificationId: string): void {
    let notifications = this.notificationsSubject.value;
    notifications = notifications.filter(n => n.id !== notificationId);
    this.notificationsSubject.next(notifications);
    this.updateUnreadCount();
    this.saveNotificationsToStorage();
  }

  /**
   * Elimina todas las notificaciones
   */
  clearAll(): void {
    this.notificationsSubject.next([]);
    this.unreadCountSubject.next(0);
    this.saveNotificationsToStorage();
  }

  /**
   * Obtiene todas las notificaciones
   */
  getNotifications(): Notification[] {
    return this.notificationsSubject.value;
  }

  /**
   * Obtiene notificaciones no leídas
   */
  getUnreadNotifications(): Notification[] {
    return this.notificationsSubject.value.filter(n => !n.read);
  }

  /**
   * Actualiza el contador de no leídas
   */
  private updateUnreadCount(): void {
    const unreadCount = this.getUnreadNotifications().length;
    this.unreadCountSubject.next(unreadCount);
  }

  /**
   * Genera un ID único para la notificación
   */
  private generateId(): string {
    return `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Guarda notificaciones en localStorage
   */
  private saveNotificationsToStorage(): void {
    try {
      const notifications = this.notificationsSubject.value;
      // No guardar callbacks en storage
      const serializableNotifications = notifications.map(n => ({
        ...n,
        action: n.action ? { label: n.action.label } : undefined
      }));
      localStorage.setItem('app_notifications', JSON.stringify(serializableNotifications));
    } catch (error) {
      console.error('Error saving notifications to storage:', error);
    }
  }

  /**
   * Carga notificaciones desde localStorage
   */
  private loadNotificationsFromStorage(): void {
    try {
      const stored = localStorage.getItem('app_notifications');
      if (stored) {
        const notifications: Notification[] = JSON.parse(stored);
        // Convertir timestamps de string a Date
        notifications.forEach(n => {
          n.timestamp = new Date(n.timestamp);
        });
        this.notificationsSubject.next(notifications);
        this.updateUnreadCount();
      }
    } catch (error) {
      console.error('Error loading notifications from storage:', error);
    }
  }
}
