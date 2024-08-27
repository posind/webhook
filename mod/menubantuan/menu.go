package menubantuan

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// SendWhatsAppMenu sends an interactive WhatsApp message.
func SendWhatsAppMenu(Pesan string) string {
    // Use CreateWhatsAppMessage to create the message
    message := CreateWhatsAppMessage()

    // Convert the message to JSON
    payloadBytes, err := json.Marshal(message)
    if err != nil {
        return fmt.Sprintf("Failed to marshal message: %v", err)
    }

    // Create the POST request to the WhatsApp API
    req, err := http.NewRequest("POST", "https://api.whatsapp.com/v1/messages", strings.NewReader(string(payloadBytes)))
    if err != nil {
        return fmt.Sprintf("Failed to create request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer YOUR_WHATSAPP_API_TOKEN")

    // Send the request
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return fmt.Sprintf("Failed to send message: %v", err)
    }
    defer res.Body.Close()

    // Check the response status code
    if res.StatusCode != http.StatusOK {
        return fmt.Sprintf("Message failed to send. Status code: %d", res.StatusCode)
    }

    return "Message sent successfully"
}

// CreateWhatsAppMessage constructs the interactive message.
func CreateWhatsAppMessage() WhatsAppInteractiveMessage {
    // Create buttons based on your business logic
    buttons := []WhatsAppButton{
        {Type: "reply", Title: "0. Kembali ke menu utama", ID: "menu_utama"},
        {Type: "reply", Title: "1. Terkait Proses Pendaftaran dan Verifikasi", ID: "pendaftaran_verifikasi"},
        {Type: "reply", Title: "2. Terkait Masalah Teknis", ID: "masalah_teknis"},
        // Add more buttons as needed...
    }

    // Create the interactive message
    message := WhatsAppInteractiveMessage{
        Type: "button", // Define the type as "button"
        Body: struct {
            Text string `json:"text"`
        }{
            Text: "Selamat datang Bapak/Ibu,\n\nTerima kasih telah menghubungi kami Helpdesk LMS Pamong Desa.\n\nUntuk mendapatkan layanan yang lebih baik, mohon bantuan Bapak/Ibu untuk memilih kendala Anda terlebih dahulu:",
        },
        Action: struct {
            Buttons []WhatsAppButton `json:"buttons"`
        }{
            Buttons: buttons,
        },
    }

    return message
}
