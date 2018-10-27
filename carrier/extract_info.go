package carrier

type InfoDict struct {
	Creator            string               `json:"creator"`
	Description        string               `json:"description"`
	Title              string               `json:"title"`
	RequestedSubtitles map[string]Subtitles `json:"requested_subtitles"`
}

type Subtitles struct {
	Extension string `json:"ext"`
	URL       string `json:"url"`
}
