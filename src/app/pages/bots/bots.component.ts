import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ApiService } from '../../core/services/api.service';
import { Bot } from '../../shared/models/bot.model';

interface Message {
  texto: string;
  esUsuario: boolean;
  timestamp: Date;
}

// Declaraciones para Web Speech API
declare var webkitSpeechRecognition: any;

@Component({
  selector: 'app-bots',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatDividerModule,
    MatTooltipModule
  ],
  templateUrl: './bots.component.html',
  styleUrl: './bots.component.scss'
})
export class BotsComponent implements OnInit {
  bots: Bot[] = [];
  botSeleccionado: Bot | null = null;
  mensajes: Message[] = [];
  nuevoMensaje = '';
  loading = false;
  sessionId = 'session-' + Date.now();

  // Voice properties
  recognition: any = null;
  isListening = false;
  isSpeaking = false;
  speechSynthesis: SpeechSynthesis;
  voiceEnabled = false;
  selectedVoice: SpeechSynthesisVoice | null = null;

  constructor(private apiService: ApiService) {
    this.speechSynthesis = window.speechSynthesis;
  }

  ngOnInit(): void {
    this.cargarBots();
    this.initVoiceRecognition();
    this.loadSpanishVoice();
  }

  cargarBots(): void {
    this.apiService.getBots().subscribe({
      next: (response) => {
        this.bots = response.bots;
        if (this.bots.length > 0) {
          this.seleccionarBot(this.bots[0]);
        }
      },
      error: (err) => {
        console.error('Error cargando bots:', err);
      }
    });
  }

  seleccionarBot(bot: Bot): void {
    this.botSeleccionado = bot;
    this.mensajes = [];
    this.sessionId = 'session-' + Date.now();

    // Mensaje de bienvenida del bot
    this.mensajes.push({
      texto: `Hola, soy ${bot.nombre}. ${bot.descripcion}`,
      esUsuario: false,
      timestamp: new Date()
    });
  }

  enviarMensaje(): void {
    if (!this.nuevoMensaje.trim() || !this.botSeleccionado) {
      return;
    }

    const texto = this.nuevoMensaje;
    this.mensajes.push({
      texto,
      esUsuario: true,
      timestamp: new Date()
    });

    this.nuevoMensaje = '';
    this.loading = true;

    this.apiService.chat(this.botSeleccionado.id, {
      session_id: this.sessionId,
      mensaje: texto
    }).subscribe({
      next: (response) => {
        this.mensajes.push({
          texto: response.respuesta,
          esUsuario: false,
          timestamp: new Date()
        });
        this.loading = false;

        // Auto-speak bot response if voice is enabled
        if (this.voiceEnabled) {
          this.speak(response.respuesta);
        }
      },
      error: (err) => {
        const errorMsg = 'Error al procesar tu mensaje. Por favor intenta de nuevo.';
        this.mensajes.push({
          texto: errorMsg,
          esUsuario: false,
          timestamp: new Date()
        });
        this.loading = false;
        console.error('Error en chat:', err);

        // Speak error message
        if (this.voiceEnabled) {
          this.speak(errorMsg);
        }
      }
    });
  }

  onKeyPress(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.enviarMensaje();
    }
  }

  // ============= VOICE FUNCTIONALITY =============

  initVoiceRecognition(): void {
    // Check if Speech Recognition is supported
    if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
      const SpeechRecognition = (window as any).SpeechRecognition || webkitSpeechRecognition;
      this.recognition = new SpeechRecognition();
      this.recognition.continuous = false;
      this.recognition.interimResults = false;
      this.recognition.lang = 'es-ES'; // Spanish language
      this.recognition.maxAlternatives = 1;

      this.recognition.onresult = (event: any) => {
        const transcript = event.results[0][0].transcript;
        this.nuevoMensaje = transcript;
        this.isListening = false;

        // Auto-send after voice input
        setTimeout(() => {
          this.enviarMensaje();
        }, 500);
      };

      this.recognition.onerror = (event: any) => {
        console.error('Speech recognition error:', event.error);
        this.isListening = false;
        this.mensajes.push({
          texto: 'Error al reconocer voz. Por favor intenta de nuevo.',
          esUsuario: false,
          timestamp: new Date()
        });
      };

      this.recognition.onend = () => {
        this.isListening = false;
      };

      this.voiceEnabled = true;
    } else {
      console.warn('Speech Recognition no está soportado en este navegador');
      this.voiceEnabled = false;
    }
  }

  loadSpanishVoice(): void {
    // Wait for voices to be loaded
    const setVoice = () => {
      const voices = this.speechSynthesis.getVoices();

      // Prefer Spanish voices (Spain or Latin America)
      const spanishVoices = voices.filter(voice =>
        voice.lang.startsWith('es-') || voice.lang === 'es'
      );

      if (spanishVoices.length > 0) {
        // Prefer female voices for better experience
        const femaleVoice = spanishVoices.find(v =>
          v.name.toLowerCase().includes('female') ||
          v.name.toLowerCase().includes('mujer') ||
          v.name.toLowerCase().includes('monica') ||
          v.name.toLowerCase().includes('lucia')
        );

        this.selectedVoice = femaleVoice || spanishVoices[0];
        console.log('Selected Spanish voice:', this.selectedVoice.name);
      } else {
        // Fallback to any voice
        this.selectedVoice = voices[0] || null;
        console.warn('No Spanish voice found, using default');
      }
    };

    // Load voices
    if (this.speechSynthesis.getVoices().length > 0) {
      setVoice();
    } else {
      this.speechSynthesis.onvoiceschanged = setVoice;
    }
  }

  toggleVoiceInput(): void {
    if (!this.voiceEnabled) {
      alert('La funcionalidad de voz no está disponible en este navegador. Prueba con Chrome, Edge o Safari.');
      return;
    }

    if (this.isListening) {
      this.stopListening();
    } else {
      this.startListening();
    }
  }

  startListening(): void {
    if (!this.recognition || this.loading) {
      return;
    }

    // Stop any ongoing speech
    if (this.isSpeaking) {
      this.speechSynthesis.cancel();
      this.isSpeaking = false;
    }

    this.isListening = true;
    this.nuevoMensaje = '';

    try {
      this.recognition.start();
      console.log('Listening for voice input...');
    } catch (error) {
      console.error('Error starting recognition:', error);
      this.isListening = false;
    }
  }

  stopListening(): void {
    if (this.recognition && this.isListening) {
      this.recognition.stop();
      this.isListening = false;
    }
  }

  speak(text: string): void {
    if (!this.voiceEnabled || !text) {
      return;
    }

    // Stop any ongoing speech
    this.speechSynthesis.cancel();

    const utterance = new SpeechSynthesisUtterance(text);
    utterance.lang = 'es-ES';
    utterance.rate = 0.95; // Slightly slower for clarity
    utterance.pitch = 1.0;
    utterance.volume = 1.0;

    if (this.selectedVoice) {
      utterance.voice = this.selectedVoice;
    }

    utterance.onstart = () => {
      this.isSpeaking = true;
    };

    utterance.onend = () => {
      this.isSpeaking = false;
    };

    utterance.onerror = (event) => {
      console.error('Speech synthesis error:', event);
      this.isSpeaking = false;
    };

    this.isSpeaking = true;
    this.speechSynthesis.speak(utterance);
  }

  stopSpeaking(): void {
    if (this.isSpeaking) {
      this.speechSynthesis.cancel();
      this.isSpeaking = false;
    }
  }

  toggleAutoSpeak(): void {
    // This could be enhanced to toggle auto-speak mode
    if (this.isSpeaking) {
      this.stopSpeaking();
    }
  }
}
