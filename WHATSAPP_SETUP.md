# WhatsApp Setup Guide for MedConnect

This guide explains how to connect Evolution API v2 to enable AI-powered WhatsApp messaging in MedConnect.

## Prerequisites

1. Docker and Docker Compose installed
2. A WhatsApp account (phone number) to use for the bot

## Quick Start

### 1. Start the Services

```bash
docker-compose up -d
```

This will start:

- PostgreSQL database
- Ollama AI
- Evolution API v2
- MedConnect Backend

### 2. Connect WhatsApp to Evolution API

After starting the services, you need to connect your WhatsApp number:

1. **Access Evolution API Dashboard**
   - Open your browser and go to: `http://localhost:8080`
   - Login with the API key: `evolution-secret-key` (or your custom WA_TOKEN)

2. **Create an Instance**
   - Click on "New Instance" or use the API to create one
   - Instance name: `medconnect`
   - Save the instance API key returned

3. **Scan QR Code**
   - A QR code will be displayed
   - Open WhatsApp on your phone
   - Go to Settings → Linked Devices → Link a Device
   - Scan the QR code

4. **Update Environment Variables**
   - Edit `.env` file and add your instance API key:

   ```env
   WA_TOKEN=your-instance-api-key
   WA_INSTANCE=medconnect
   ```

5. **Restart Backend**

   ```bash
   docker-compose restart backend
   ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `WA_URL` | Evolution API URL | <http://localhost:8080> |
| `WA_TOKEN` | API Key for authentication | evolution-secret-key |
| `WA_INSTANCE` | Instance name | medconnect |

## Testing the Integration

### Check Evolution API Health

```bash
curl -X GET http://localhost:8080/health
```

### Check Instance Status

```bash
curl -X GET http://localhost:8080/instance/connectionState/medconnect \
  -H "apikey: your-api-key"
```

### Send a Test Message via API

```bash
curl -X POST http://localhost:8080/message/sendText/medconnect \
  -H "Content-Type: application/json" \
  -H "apikey: your-api-key" \
  -d '{
    "number": "212661234567",
    "text": "Hello from MedConnect!"
  }'
```

## Troubleshooting

### WhatsApp Disconnected

If WhatsApp disconnects:

1. Access Evolution API dashboard
2. Find the instance
3. Click "Reconnect" or scan QR code again

### Messages Not Sending

1. Check Evolution API is running: `docker-compose ps`
2. Check logs: `docker-compose logs evolution-api`
3. Verify WA_TOKEN matches the instance API key
4. Ensure WhatsApp is connected (check dashboard)

### Backend Cannot Connect to Evolution API

1. Check that both services are on the same network
2. Verify WA_URL is correct (use service name in Docker: `http://evolution-api:8080`)
3. Check backend logs: `docker-compose logs backend`

## API Endpoints Used by MedConnect

The backend uses these Evolution API endpoints:

- `POST /message/sendText/{instanceName}` - Send text message

## Security Notes

- Change the default `WA_TOKEN` in production
- Keep your API keys secure
- Don't commit `.env` file to version control
