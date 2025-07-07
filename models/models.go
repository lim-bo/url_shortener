package models

type ClicksStat struct {
	Code   string `json:"code"`
	OGLink string `json:"link"`
	Clicks uint64 `json:"clicks"`
}
