package models

type Event struct {
	ID       int64  `json:"id" example:"1" json:"id"`
	Alias    string `json:"alias"`
	Title    string `json:"title"`
	Info     string `json:"info"`
	FullInfo string `json:"full_info"`
	Visible  bool   `json:"visible"`

	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`

	Site string `json:"site"`
}
