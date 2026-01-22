export interface EmailTemplate {
  id: string;
  nombre: string;
  asunto: string;
  categoria: 'Previo al cargo' | 'Devuelto' | 'Seguimiento' | 'Incidencias' | 'Cierre';
  motivo?: string;
  htmlContent: string; // HTML con variables {nombre}, {importe}, etc.
  variables: string[]; // Lista de variables disponibles
  incluirBloquePago: boolean;
  activo: boolean;
}

export const VARIABLES_DISPONIBLES = [
  '{nombre}',
  '{nif}',
  '{poliza}',
  '{num_recibo}',
  '{importe}',
  '{venc}',
  '{dias_vencido}',
  '{motivo}',
  '{agente}',
  '{empresa}',
  '{telefono_empresa}',
  '{email_empresa}',
  '{bloquePago}' // Se reemplaza con HTML de medios de pago
];

// Plantilla base HTML para emails
export const EMAIL_BASE_TEMPLATE = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{asunto}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .email-container {
            max-width: 600px;
            margin: 20px auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #c2185b 0%, #880e4f 100%);
            color: #ffffff;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 30px 20px;
        }
        .greeting {
            font-size: 18px;
            color: #333;
            margin-bottom: 20px;
        }
        .info-box {
            background-color: #f8f9fa;
            border-left: 4px solid #c2185b;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-box strong {
            color: #c2185b;
        }
        .payment-section {
            background-color: #fff9f0;
            border: 2px solid #ffb74d;
            border-radius: 8px;
            padding: 20px;
            margin: 25px 0;
        }
        .payment-section h3 {
            color: #f57c00;
            margin-top: 0;
            font-size: 18px;
        }
        .payment-option {
            margin: 12px 0;
            padding: 12px;
            background-color: #ffffff;
            border-radius: 6px;
            border: 1px solid #e0e0e0;
        }
        .payment-option strong {
            color: #c2185b;
            display: block;
            margin-bottom: 5px;
        }
        .button {
            display: inline-block;
            padding: 14px 32px;
            background-color: #c2185b;
            color: #ffffff !important;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 600;
            margin: 10px 5px;
            text-align: center;
        }
        .button:hover {
            background-color: #880e4f;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 20px;
            text-align: center;
            font-size: 13px;
            color: #666;
            border-top: 1px solid #e0e0e0;
        }
        .footer p {
            margin: 5px 0;
        }
        .security-notice {
            background-color: #e8f5e9;
            border-left: 4px solid #4caf50;
            padding: 12px;
            margin: 20px 0;
            font-size: 13px;
            color: #2e7d32;
        }
        @media only screen and (max-width: 600px) {
            .email-container {
                margin: 0;
                border-radius: 0;
            }
            .content {
                padding: 20px 15px;
            }
            .button {
                display: block;
                margin: 10px 0;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <!-- Header -->
        <div class="header">
            <h1>🏛️ Soriano Mediadores</h1>
        </div>

        <!-- Content -->
        <div class="content">
            {content}
        </div>

        <!-- Footer -->
        <div class="footer">
            <p><strong>Soriano Mediadores de Seguros</strong></p>
            <p>📧 {email_empresa} | ☎️ {telefono_empresa}</p>
            <p style="color: #999; font-size: 11px; margin-top: 15px;">
                Este es un correo automático. Por favor, no responder directamente.
            </p>
        </div>
    </div>
</body>
</html>
`;

// Plantillas predefinidas
export const PLANTILLAS_EMAIL_PREDEFINIDAS: Partial<EmailTemplate>[] = [
  {
    id: 'email_r01_fondos',
    nombre: 'Devolución — Fondos insuficientes (R01)',
    asunto: 'Recibo {num_recibo} devuelto - Acción requerida',
    categoria: 'Devuelto',
    motivo: 'R01',
    incluirBloquePago: true,
    activo: true,
    variables: ['{nombre}', '{num_recibo}', '{poliza}', '{importe}', '{venc}', '{motivo}', '{bloquePago}'],
    htmlContent: `
<p class="greeting">Estimado/a <strong>{nombre}</strong>,</p>

<p>Le informamos que el recibo <strong>{num_recibo}</strong> correspondiente a su póliza <strong>{poliza}</strong> ha sido devuelto por su entidad bancaria con el motivo:</p>

<div class="info-box">
    <strong>Motivo:</strong> {motivo}<br>
    <strong>Importe:</strong> {importe}<br>
    <strong>Fecha de vencimiento:</strong> {venc}<br>
    <strong>Días vencidos:</strong> {dias_vencido}
</div>

<p>Para regularizar su situación y mantener la cobertura de su seguro activa, le solicitamos que realice el pago lo antes posible.</p>

{bloquePago}

<div class="security-notice">
    🔒 <strong>Seguridad:</strong> Este email proviene de un dominio verificado de Soriano Mediadores. Nunca solicitamos datos bancarios por email.
</div>

<p>Si ya ha realizado el pago o tiene alguna consulta, no dude en contactarnos.</p>

<p style="margin-top: 30px;">
    Atentamente,<br>
    <strong>{agente}</strong><br>
    Soriano Mediadores
</p>
`
  },
  {
    id: 'email_recordatorio_d5',
    nombre: 'Recordatorio D+5 - Amigable',
    asunto: 'Recordatorio: Recibo {num_recibo} pendiente de pago',
    categoria: 'Seguimiento',
    motivo: 'Recordatorio',
    incluirBloquePago: true,
    activo: true,
    variables: ['{nombre}', '{num_recibo}', '{poliza}', '{importe}', '{venc}', '{dias_vencido}', '{bloquePago}'],
    htmlContent: `
<p class="greeting">Hola <strong>{nombre}</strong>,</p>

<p>Le recordamos amablemente que el recibo <strong>{num_recibo}</strong> de su póliza <strong>{poliza}</strong> se encuentra pendiente de pago.</p>

<div class="info-box">
    <strong>Póliza:</strong> {poliza}<br>
    <strong>Recibo:</strong> {num_recibo}<br>
    <strong>Importe:</strong> {importe}<br>
    <strong>Vencimiento:</strong> {venc}<br>
    <strong>Días transcurridos:</strong> {dias_vencido}
</div>

<p>Puede regularizar su situación de forma rápida y segura mediante cualquiera de estos métodos:</p>

{bloquePago}

<p style="background-color: #fff3cd; padding: 15px; border-radius: 6px; border-left: 4px solid #ffc107;">
    ⚠️ <strong>Importante:</strong> Mantener su póliza al día evita la suspensión de coberturas y posibles recargos.
</p>

<p>Quedamos a su disposición para cualquier consulta.</p>

<p style="margin-top: 30px;">
    Un cordial saludo,<br>
    <strong>{agente}</strong><br>
    Soriano Mediadores
</p>
`
  },
  {
    id: 'email_confirmacion_pago',
    nombre: 'Confirmación de pago recibido',
    asunto: '✅ Pago recibido - Recibo {num_recibo}',
    categoria: 'Cierre',
    motivo: 'Confirmación',
    incluirBloquePago: false,
    activo: true,
    variables: ['{nombre}', '{num_recibo}', '{poliza}', '{importe}'],
    htmlContent: `
<p class="greeting">Estimado/a <strong>{nombre}</strong>,</p>

<p>Nos complace confirmarle que hemos recibido correctamente el pago del recibo <strong>{num_recibo}</strong>.</p>

<div class="info-box" style="border-left-color: #4caf50; background-color: #e8f5e9;">
    <strong style="color: #2e7d32;">✅ Pago confirmado</strong><br><br>
    <strong>Póliza:</strong> {poliza}<br>
    <strong>Recibo:</strong> {num_recibo}<br>
    <strong>Importe:</strong> {importe}
</div>

<p>Su cobertura de seguro continúa activa sin interrupciones.</p>

<p>Si necesita un justificante de pago o tiene alguna consulta, no dude en contactarnos respondiendo a este correo.</p>

<p style="margin-top: 30px;">
    Gracias por su confianza,<br>
    <strong>{agente}</strong><br>
    Soriano Mediadores
</p>
`
  },
  {
    id: 'email_preaviso_cargo',
    nombre: 'Preaviso de cargo (D-5 a D-2)',
    asunto: 'Próximo cargo: Recibo {num_recibo}',
    categoria: 'Previo al cargo',
    motivo: 'Preaviso',
    incluirBloquePago: false,
    activo: true,
    variables: ['{nombre}', '{num_recibo}', '{poliza}', '{importe}', '{venc}'],
    htmlContent: `
<p class="greeting">Estimado/a <strong>{nombre}</strong>,</p>

<p>Le informamos que en los próximos días se realizará el cargo del siguiente recibo:</p>

<div class="info-box">
    <strong>Póliza:</strong> {poliza}<br>
    <strong>Recibo:</strong> {num_recibo}<br>
    <strong>Importe:</strong> {importe}<br>
    <strong>Fecha prevista de cargo:</strong> {venc}
</div>

<p>Por favor, asegúrese de tener fondos suficientes en su cuenta para evitar devoluciones.</p>

<p style="background-color: #e3f2fd; padding: 15px; border-radius: 6px; border-left: 4px solid #2196f3;">
    💡 <strong>¿Desea cambiar su forma de pago?</strong><br>
    Puede realizar el pago por otros medios como tarjeta o transferencia. Contáctenos si lo necesita.
</p>

<p>Gracias por su atención.</p>

<p style="margin-top: 30px;">
    Atentamente,<br>
    <strong>{agente}</strong><br>
    Soriano Mediadores
</p>
`
  }
];
