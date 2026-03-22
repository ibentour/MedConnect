// Package service provides external integration services for MedConnect.
// This file implements the Evolution API WhatsApp gateway client.
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WhatsAppService handles sending messages via the local Evolution API instance.
type WhatsAppService struct {
	baseURL    string
	apiToken   string
	instanceID string
	httpClient *http.Client
}

// NewWhatsAppService creates a WhatsApp gateway client.
// baseURL: Evolution API URL (e.g., "http://localhost:8080")
// apiToken: Bearer token for authentication
// instanceID: Evolution API instance name (default: "medconnect")
func NewWhatsAppService(baseURL, apiToken, instanceID string) *WhatsAppService {
	if instanceID == "" {
		instanceID = "medconnect"
	}
	return &WhatsAppService{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiToken:   apiToken,
		instanceID: instanceID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ──────────────────────────────────────────────────────────────────────
// Evolution API request/response types
// ──────────────────────────────────────────────────────────────────────

type sendTextRequest struct {
	Number  string          `json:"number"`
	Text    string          `json:"text"`
	Options *messageOptions `json:"options,omitempty"`
}

type messageOptions struct {
	Delay   int  `json:"delay,omitempty"`
	Linkify bool `json:"linkPreview"`
}

type sendTextResponse struct {
	Key struct {
		RemoteJID string `json:"remoteJid"`
		ID        string `json:"id"`
	} `json:"key"`
	Status string `json:"status"`
}

// ──────────────────────────────────────────────────────────────────────
// Send Text Message
// ──────────────────────────────────────────────────────────────────────

// SendTextMessage sends a WhatsApp text message to the specified phone number.
// Phone number should be in Moroccan format (e.g., "0661234567" or "+212661234567").
// Returns the message ID on success.
func (ws *WhatsAppService) SendTextMessage(phone, message string) (string, error) {
	// Normalize Moroccan phone number to international format
	normalizedPhone := normalizeMoroccanPhone(phone)

	reqBody := sendTextRequest{
		Number: normalizedPhone,
		Text:   message,
		Options: &messageOptions{
			Delay:   1200, // 1.2s delay for natural feel
			Linkify: false,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("whatsapp: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/message/sendText/%s", ws.baseURL, ws.instanceID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("whatsapp: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", ws.apiToken)

	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("whatsapp: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("whatsapp: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("whatsapp: Evolution API returned status %d: %s", resp.StatusCode, string(body))
	}

	var sendResp sendTextResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		// Message was sent but response parsing failed — not critical
		return "sent-parse-error", nil
	}

	return sendResp.Key.ID, nil
}

// ──────────────────────────────────────────────────────────────────────
// Phone Number Normalization
// ──────────────────────────────────────────────────────────────────────

// normalizeMoroccanPhone converts Moroccan phone numbers to the
// format expected by WhatsApp: country code + number (no + prefix).
//
// Examples:
//   - "0661234567"    → "212661234567"
//   - "+212661234567" → "212661234567"
//   - "212661234567"  → "212661234567"
func normalizeMoroccanPhone(phone string) string {
	// Remove spaces, dashes, and parentheses
	phone = strings.NewReplacer(
		" ", "", "-", "", "(", "", ")", "",
	).Replace(phone)

	// Remove leading "+"
	phone = strings.TrimPrefix(phone, "+")

	// Convert local format (06/07) to international
	if strings.HasPrefix(phone, "06") || strings.HasPrefix(phone, "07") || strings.HasPrefix(phone, "05") {
		phone = "212" + phone[1:] // Replace leading "0" with "212"
	}

	return phone
}

// ──────────────────────────────────────────────────────────────────────
// Patient Notification Methods
// ──────────────────────────────────────────────────────────────────────

// SendAppointmentNotification sends a WhatsApp notification to a patient
// when their appointment is scheduled.
func (ws *WhatsAppService) SendAppointmentNotification(phone string, data AppointmentNotificationData) (string, error) {
	message := AppointmentScheduledTemplate(data)
	return ws.SendTextMessage(phone, message)
}

// SendReferralDeniedNotification sends a WhatsApp notification to a patient
// when their referral is denied.
func (ws *WhatsAppService) SendReferralDeniedNotification(phone string, data ReferralDeniedNotificationData) (string, error) {
	message := ReferralDeniedTemplate(data)
	return ws.SendTextMessage(phone, message)
}

// SendReferralRedirectedNotification sends a WhatsApp notification to a patient
// when their referral is redirected to another department.
func (ws *WhatsAppService) SendReferralRedirectedNotification(phone string, data ReferralRedirectedNotificationData) (string, error) {
	message := ReferralRedirectedTemplate(data)
	return ws.SendTextMessage(phone, message)
}
