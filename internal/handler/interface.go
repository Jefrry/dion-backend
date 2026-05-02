package handler

type createRecordingRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	ConcertDate *string `json:"concertDate"`
	ExternalURL string  `json:"externalURL"`
	ArtistName  string  `json:"artistName"`
}
