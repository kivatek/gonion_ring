package model

type Report struct {
	OriginalWidth  int `json:"originalWidth"`
	OriginalHeight int `json:"originalHeight"`
	Left           int `json:"left"`
	Top            int `json:"top"`
	Right          int `json:"right"`
	Bottom         int `json:"bottom"`
}
