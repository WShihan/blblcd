package model

type VideoItem struct {
	Comment          int    `json:"comment"`
	Typeid           int    `json:"typeid"`
	Play             int    `json:"play"`
	Pic              string `json:"pic"`
	Subtitle         string `json:"subtitle"`
	Description      string `json:"description"`
	Copyright        string `json:"copyright"`
	Title            string `json:"title"`
	Review           int    `json:"review"`
	Author           string `json:"author"`
	Mid              int    `json:"mid"`
	Created          int    `json:"created"`
	Length           string `json:"length"`
	VideoReview      int    `json:"video_review"`
	Aid              int    `json:"aid"`
	Bvid             string `json:"bvid"`
	HideClick        bool   `json:"hide_click"`
	IsPay            int    `json:"is_pay"`
	IsUnionVideo     int    `json:"is_union_video"`
	IsSteinsGate     int    `json:"is_steins_gate"`
	IsLivePlayback   int    `json:"is_live_playback"`
	IsLessonVideo    int    `json:"is_lesson_video"`
	IsLessonFinished int    `json:"is_lesson_finished"`
	LessonUpdateInfo string `json:"lesson_update_info"`
	JumpURL          string `json:"jump_url"`
	Meta             any    `json:"meta"`
	IsAvoided        int    `json:"is_avoided"`
	SeasonID         int    `json:"season_id"`
	Attribute        int    `json:"attribute"`
	IsChargingArc    bool   `json:"is_charging_arc"`
	Vt               int    `json:"vt"`
	EnableVt         int    `json:"enable_vt"`
	VtDisplay        string `json:"vt_display"`
	PlaybackPosition int    `json:"playback_position"`
}

type VideoListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		List struct {
			Vlist []VideoItem `json:"vlist"`
			Slist []any       `json:"slist"`
		} `json:"list"`
		Page struct {
			Pn    int `json:"pn"`
			Ps    int `json:"ps"`
			Count int `json:"count"`
		} `json:"page"`
		EpisodicButton struct {
			Text string `json:"text"`
			URI  string `json:"uri"`
		} `json:"episodic_button"`
		IsRisk      bool `json:"is_risk"`
		GaiaResType int  `json:"gaia_res_type"`
		GaiaData    any  `json:"gaia_data"`
	} `json:"data"`
}
