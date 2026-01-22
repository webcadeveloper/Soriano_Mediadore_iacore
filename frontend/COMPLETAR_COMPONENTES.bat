@echo off
chcp 65001 >nul
echo Completando todos los componentes del frontend...
cd src\app\pages

echo Componente Bots - TypeScript
(
echo import { Component, OnInit } from '@angular/core';
echo import { CommonModule } from '@angular/common';
echo import { FormsModule } from '@angular/forms';
echo import { MatCardModule } from '@angular/material/card';
echo import { MatButtonModule } from '@angular/material/button';
echo import { MatIconModule } from '@angular/material/icon';
echo import { MatFormFieldModule } from '@angular/material/form-field';
echo import { MatInputModule } from '@angular/material/input';
echo import { MatChipsModule } from '@angular/material/chips';
echo import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
echo import { ApiService } from '../../core/services/api.service';
echo import { Bot } from '../../shared/models/bot.model';
echo.
echo interface Message { texto: string; esUsuario: boolean; timestamp: Date; }
echo.
echo @Component({
echo   selector: 'app-bots',
echo   standalone: true,
echo   imports: [CommonModule, FormsModule, MatCardModule, MatButtonModule, MatIconModule, MatFormFieldModule, MatInputModule, MatChipsModule, MatProgressSpinnerModule],
echo   templateUrl: './bots.component.html',
echo   styleUrl: './bots.component.scss'
echo ^}^)
echo export class BotsComponent implements OnInit {
echo   bots: Bot[] = [];
echo   botSeleccionado: Bot ^| null = null;
echo   mensajes: Message[] = [];
echo   nuevoMensaje = '';
echo   loading = false;
echo   sessionId = 'session-' + Date.now(^);
echo   constructor(private apiService: ApiService^) {}
echo   ngOnInit(^): void { this.apiService.getBots(^).subscribe({ next: (r^) =^> { this.bots = r.bots; if(this.bots.length ^> 0^) this.seleccionarBot(this.bots[0]^); }, error: (e^) =^> console.error(e^) }^); }
echo   seleccionarBot(bot: Bot^): void { this.botSeleccionado = bot; this.mensajes = []; this.sessionId = 'session-' + Date.now(^); }
echo   enviarMensaje(^): void { if(!this.nuevoMensaje.trim(^) ^|^| !this.botSeleccionado^) return; const texto = this.nuevoMensaje; this.mensajes.push({ texto, esUsuario: true, timestamp: new Date(^) }^); this.nuevoMensaje = ''; this.loading = true; this.apiService.chat(this.botSeleccionado.id, { session_id: this.sessionId, mensaje: texto }^).subscribe({ next: (r^) =^> { this.mensajes.push({ texto: r.respuesta, esUsuario: false, timestamp: new Date(^) }^); this.loading = false; }, error: (e^) =^> { this.mensajes.push({ texto: 'Error: ' + e.message, esUsuario: false, timestamp: new Date(^) }^); this.loading = false; } }^); }
echo }
) > bots\bots.component.ts

echo Componente Bots - HTML
(
echo ^<div class="bots-container"^>
echo   ^<h1^>Chat con Bots AI^</h1^>
echo   ^<div class="bots-selector"^>
echo     ^<mat-chip-listbox^>
echo       ^<mat-chip-option *ngFor="let bot of bots" (click^)="seleccionarBot(bot^)" [selected]="botSeleccionado?.id === bot.id"^>{{ bot.nombre }}^</mat-chip-option^>
echo     ^</mat-chip-listbox^>
echo   ^</div^>
echo   ^<mat-card *ngIf="botSeleccionado" class="chat-card"^>
echo     ^<div class="chat-messages"^>
echo       ^<div *ngFor="let msg of mensajes" [class]="msg.esUsuario ? 'message user' : 'message bot'"^>
echo         ^<p^>{{ msg.texto }}^</p^>
echo       ^</div^>
echo     ^</div^>
echo     ^<mat-divider^>^</mat-divider^>
echo     ^<div class="chat-input"^>
echo       ^<mat-form-field appearance="outline"^>
echo         ^<input matInput [(ngModel^)]="nuevoMensaje" (keyup.enter^)="enviarMensaje(^)" placeholder="Escribe tu mensaje..."^>
echo       ^</mat-form-field^>
echo       ^<button mat-raised-button color="primary" (click^)="enviarMensaje(^)" [disabled]="loading"^>Enviar^</button^>
echo     ^</div^>
echo   ^</mat-card^>
echo ^</div^>
) > bots\bots.component.html

echo ✓ Componentes completados
pause
