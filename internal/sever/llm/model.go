package llm

type TTSMessage struct {
	Index      int32  `json:"index"`
	MessageID  int64  `json:"message_id"`
	Text       string `json:"text"`
	SenderName string `json:"sender_name"`
	Audio      []byte `json:"audio"`
}

type TTSRequest struct {
	Text      string
	MessageID int64
	Index     int32
	Sender    string
}

type HistoryMessage struct {
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	Time       string `json:"time"`
}
