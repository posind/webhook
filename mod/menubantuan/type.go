package menubantuan

type WhatsAppButton struct {
    Type  string `json:"type"`
    Title string `json:"title"`
    ID    string `json:"id"`
}

type WhatsAppInteractiveMessage struct {
    Type    string `json:"type"`
    Body    struct {
        Text string `json:"text"`
    } `json:"body"`
    Action struct {
        Buttons []WhatsAppButton `json:"buttons"`
    } `json:"action"`
}
