// Package service provides external integration services for MedConnect.
// This file contains WhatsApp notification message templates.
package service

import (
	"fmt"
	"time"
)

// NotificationLanguage defines the language for notification messages.
type NotificationLanguage string

const (
	LanguageFrench NotificationLanguage = "fr"
	LanguageArabic NotificationLanguage = "ar"
)

// ──────────────────────────────────────────────────────────────────────
// Notification Templates
// ──────────────────────────────────────────────────────────────────────

// AppointmentNotificationData contains data for appointment notifications.
type AppointmentNotificationData struct {
	PatientName     string
	DepartmentName  string
	DoctorName      string
	AppointmentDate time.Time
	CHUAddress      string
	CHUContact      string
	Instructions    string
	Language        NotificationLanguage
}

// ReferralDeniedNotificationData contains data for denied referral notifications.
type ReferralDeniedNotificationData struct {
	PatientName    string
	DepartmentName string
	DoctorName     string
	Reason         string
	CHUContact     string
	Language       NotificationLanguage
}

// ReferralRedirectedNotificationData contains data for redirected referral notifications.
type ReferralRedirectedNotificationData struct {
	PatientName    string
	FromDepartment string
	ToDepartment   string
	DoctorName     string
	Reason         string
	CHUContact     string
	Language       NotificationLanguage
}

// AppointmentScheduledTemplate generates a WhatsApp message for appointment scheduling.
func AppointmentScheduledTemplate(data AppointmentNotificationData) string {
	dateStr := data.AppointmentDate.Format("02/01/2006 à 15:04")
	dayOfWeek := getDayOfWeekFrench(data.AppointmentDate.Weekday())

	if data.Language == LanguageArabic {
		return fmt.Sprintf(`مرحباً %s 📋

✅ تم تحديد موعدك في مستشفى CHU Mohammed VI

📅 الموعد: %s %s
🏥 القسم: %s
👨‍⚕️ الطبيب: %s

📍 العنوان: %s

📝 تعليمات:
%s

⚠️ هام: يرجى إحضار بطاقة الهوية ( CIN ) والحضور قبل 30 دقيقة من الموعد.

للأسئلة أو التأجيل، يرجى الاتصال: %s

شكراً لثقتكم بنا 🏥`, data.PatientName, dayOfWeek, dateStr, data.DepartmentName, data.DoctorName, data.CHUAddress, data.Instructions, data.CHUContact)
	}

	return fmt.Sprintf(`Bonjour %s 📋

✅ Votre rendez-vous a été programmé au CHU Mohammed VI

📅 Date: %s %s
🏥 Service: %s
👨‍⚕️ Médecin: %s

📍 Adresse: %s

📝 Instructions:
%s

⚠️ Important: Veuillez apporter votre CIN et arriver 30 minutes avant le rendez-vous.

Pour toute question ou report, contactez: %s

Merci de votre confiance 🏥`, data.PatientName, dayOfWeek, dateStr, data.DepartmentName, data.DoctorName, data.CHUAddress, data.Instructions, data.CHUContact)
}

// ReferralDeniedTemplate generates a WhatsApp message for denied referrals.
func ReferralDeniedTemplate(data ReferralDeniedNotificationData) string {
	if data.Language == LanguageArabic {
		return fmt.Sprintf(`مرحباً %s 📋

❌ نأسف لإخبارك أن طلب الإحالة الخاص بك إلى %s قد تم رفضه.

👨‍⚕️ الطبيب: %s
📝 السبب: %s

لمزيد من المعلومات أو الاستئناف، يرجى الاتصال: %s

نعتذر عن أي إزعاج 🏥`, data.PatientName, data.DepartmentName, data.DoctorName, data.Reason, data.CHUContact)
	}

	return fmt.Sprintf(`Bonjour %s 📋

❌ Nous regrettons de vous informer que votre demande de référence vers %s a été refusée.

👨‍⚕️ Médecin: %s
📝 Raison: %s

Pour plus d'informations ou pour faire appel, veuillez contacter: %s

Nous vous prions de bien vouloir nous excuser pour la gêne occasionnée 🏥`, data.PatientName, data.DepartmentName, data.DoctorName, data.Reason, data.CHUContact)
}

// ReferralRedirectedTemplate generates a WhatsApp message for redirected referrals.
func ReferralRedirectedTemplate(data ReferralRedirectedNotificationData) string {
	if data.Language == LanguageArabic {
		return fmt.Sprintf(`مرحباً %s 📋

🔄 تم تحويل طلب الإحالة الخاص بك.

من: %s
إلى: %s
👨‍⚕️ الطبيب: %s
📝 السبب: %s

سيتم إخطارك بالموعد الجديد قريباً.

للأسئلة، يرجى الاتصال: %s

شكراً لتفهمكم 🏥`, data.PatientName, data.FromDepartment, data.ToDepartment, data.DoctorName, data.Reason, data.CHUContact)
	}

	return fmt.Sprintf(`Bonjour %s 📋

🔄 Votre demande de référence a été redirigée.

De: %s
Vers: %s
👨‍⚕️ Médecin: %s
📝 Raison: %s

Vous serez bientôt notifié du nouveau rendez-vous.

Pour toute question, contactez: %s

Merci de votre compréhension 🏥`, data.PatientName, data.FromDepartment, data.ToDepartment, data.DoctorName, data.Reason, data.CHUContact)
}

// DefaultCHUAddress returns the default CHU address for notifications.
func DefaultCHUAddress() string {
	return "CHU Mohammed VI, Boulevard Mohammed VI, Oujda, Morocco"
}

// DefaultCHUContact returns the default CHU contact number.
func DefaultCHUContact() string {
	return "+212 536 68 88 88"
}

// getDayOfWeekFrench returns the French name of a weekday.
func getDayOfWeekFrench(weekday time.Weekday) string {
	days := map[time.Weekday]string{
		time.Monday:    "Lundi",
		time.Tuesday:   "Mardi",
		time.Wednesday: "Mercredi",
		time.Thursday:  "Jeudi",
		time.Friday:    "Vendredi",
		time.Saturday:  "Samedi",
		time.Sunday:    "Dimanche",
	}
	if day, ok := days[weekday]; ok {
		return day
	}
	return ""
}

// GetDefaultInstructions returns default instructions based on department.
func GetDefaultInstructions(departmentName string) string {
	instructions := map[string]string{
		"Cardiologie":   "Apportez vos résultats ECG récents, analyses sanguines (NFS, profil lipidique), et tout rapport cardiologique précédent.",
		"Neurologie":    "Apportez vos IRM/CT cérébraux récents, rapports neurologiques précédents, et liste de vos médicaments.",
		"Chirurgie":     "Apportez vos analyses pré-opératoires (NFS, coagulation), radiographies de la zone concernée, et consentement signé.",
		"Pédiatrie":     "Apportez le carnet de santé de l'enfant, vaccins à jour, et tous rapports médicaux précédents.",
		"Gynécologie":   "Apportez vos échographies récentes, résultats de frottis, et historique médical gynécologique.",
		"Dermatologie":  "Apportez photos de l'évolution des lésions, traitements en cours, et rapports dermatologiques précédents.",
		"Ophtalmologie": "Apportez vos examens de vue récents, rapports d'ophtalmologie, et liste des médicaments oculaires.",
		"ORL":           "Apportez vos audiogrammes, scanners des sinus, et rapports ORL précédents.",
		"Urologie":      "Apportez vos analyses d'urine récentes, échographies, et rapports urologiques.",
		"Psychiatrie":   "Apportez vos rapports psychiatriques précédents, traitements en cours, et tout document pertinent.",
	}

	for key, value := range instructions {
		if containsIgnoreCase(departmentName, key) {
			return value
		}
	}

	return "Apportez votre CIN, tous documents médicaux récents, et liste de vos médicaments actuels."
}

// containsIgnoreCase checks if a string contains another string (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
