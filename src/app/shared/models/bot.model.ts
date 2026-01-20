export interface Bot {
  id: string;
  nombre: string;
  descripcion: string;
  endpoint: string;
}

export interface BotsResponse {
  bots: Bot[];
}

export interface ChatMessage {
  session_id: string;
  mensaje: string;
}

export interface ChatResponse {
  bot: string;
  respuesta: string;
  session_id: string;
  timestamp: string;
}
