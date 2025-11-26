package model

type Event struct {
	EventName  string                 `json:"event_name"`
	Channel    string                 `json:"channel"`
	CampaignID string                 `json:"campaign_id"`
	UserID     string                 `json:"user_id"`
	Timestamp  int64                  `json:"timestamp"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type MetricsResponse struct {
	EventName   string          `json:"event_name"`
	TotalCount  int             `json:"total_count"`
	UniqueUsers int             `json:"unique_users"`
	ByChannel   []ChannelMetric `json:"by_channel,omitempty"`
	ByTime      []TimeMetric    `json:"by_time,omitempty"`
}

type ChannelMetric struct {
	Channel     string `json:"channel"`
	TotalCount  int    `json:"total_count"`
	UniqueUsers int    `json:"unique_users"`
}

type TimeMetric struct {
	Period      string `json:"period"`
	TotalCount  int    `json:"total_count"`
	UniqueUsers int    `json:"unique_users"`
}
