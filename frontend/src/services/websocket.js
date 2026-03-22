// WebSocket service for real-time notifications
// Uses native WebSocket API to connect to backend /ws endpoint

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:3000/ws';

// Event types matching backend
export const WS_EVENTS = {
    NEW_REFERRAL: 'new_referral',
    REFERRAL_UPDATE: 'referral_updated',
    REFERRAL_REDIRECT: 'referral_redirected',
};

// Callback storage
let ws = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000;

let messageHandlers = [];
let connectionHandlers = [];
let isConnecting = false;

// Get auth token
const getToken = () => localStorage.getItem('token');

// Connect to WebSocket
export const connect = () => {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
        return;
    }

    const token = getToken();
    if (!token) {
        console.warn('[WS] No auth token, skipping WebSocket connection');
        return;
    }

    isConnecting = true;

    try {
        // Append token as query param for authentication
        const wsUrl = `${WS_URL}?token=${encodeURIComponent(token)}`;
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('[WS] Connected');
            isConnecting = false;
            reconnectAttempts = 0;
            connectionHandlers.forEach(cb => cb(true));
        };

        ws.onclose = (event) => {
            console.log('[WS] Disconnected', event.code, event.reason);
            isConnecting = false;
            ws = null;
            connectionHandlers.forEach(cb => cb(false));

            // Auto reconnect if not a clean close
            if (event.code !== 1000 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                reconnectAttempts++;
                console.log(`[WS] Reconnecting... attempt ${reconnectAttempts}`);
                setTimeout(connect, RECONNECT_DELAY);
            }
        };

        ws.onerror = (error) => {
            console.error('[WS] Error:', error);
            isConnecting = false;
        };

        ws.onmessage = (event) => {
            try {
                // Handle multiple messages separated by newline
                const messages = event.data.split('\n').filter(m => m.trim());

                messages.forEach(msgStr => {
                    const message = JSON.parse(msgStr);

                    // Handle ping/pong
                    if (message.type === 'ping') {
                        // Pong is handled automatically by browser
                        return;
                    }

                    // Notify all handlers
                    messageHandlers.forEach(handler => {
                        try {
                            handler(message);
                        } catch (err) {
                            console.error('[WS] Handler error:', err);
                        }
                    });
                });
            } catch (err) {
                console.error('[WS] Failed to parse message:', err);
            }
        };
    } catch (err) {
        console.error('[WS] Connection error:', err);
        isConnecting = false;
    }
};

// Disconnect from WebSocket
export const disconnect = () => {
    if (ws) {
        ws.close(1000, 'Client disconnect');
        ws = null;
    }
    reconnectAttempts = MAX_RECONNECT_ATTEMPTS; // Prevent auto reconnect
};

// Subscribe to messages
export const subscribe = (handler) => {
    messageHandlers.push(handler);
    return () => {
        messageHandlers = messageHandlers.filter(h => h !== handler);
    };
};

// Subscribe to connection status changes
export const onConnectionChange = (handler) => {
    connectionHandlers.push(handler);
    return () => {
        connectionHandlers = connectionHandlers.filter(h => h !== handler);
    };
};

// Get connection status
export const isConnected = () => ws && ws.readyState === WebSocket.OPEN;

// Send message to server (for future use)
export const send = (data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(data));
    }
};

export default {
    connect,
    disconnect,
    subscribe,
    onConnectionChange,
    isConnected,
    send,
    WS_EVENTS,
};
