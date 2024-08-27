package model

type Comment struct {
	Uname         string    //姓名
	Sex           string    //性别
	Content       string    //评论内容
	Rpid          int64     //评论id
	Oid           int       //评论区id
	Bvid          string    //视频bv
	Mid           int       //发送者id
	Parent        int       //若为一级评论则为 0若为二级评论则为根评论 rpid大于二级评论为上一级评 论 rpid
	Fansgrade     int       //是否粉丝标签
	Ctime         int       //评论时间戳
	Like          int       //喜欢数
	Following     bool      //是否关注
	Current_level int       //当前等级
	Location      string    //位置
	Pictures      []Picture // 图片
}

type Picture struct {
	Img_src string `json:"img_src"`
}
type ReplyItem struct {
	Rpid      int64  `json:"rpid"`
	Oid       int    `json:"oid"`
	Type      int    `json:"type"`
	Mid       int    `json:"mid"`
	Root      int    `json:"root"`
	Parent    int    `json:"parent"`
	Dialog    int    `json:"dialog"`
	Count     int    `json:"count"`
	Rcount    int    `json:"rcount"`
	State     int    `json:"state"`
	Fansgrade int    `json:"fansgrade"`
	Attr      int    `json:"attr"`
	Ctime     int    `json:"ctime"`
	MidStr    string `json:"mid_str"`
	OidStr    string `json:"oid_str"`
	RpidStr   string `json:"rpid_str"`
	Like      int    `json:"like"`
	Action    int    `json:"action"`
	Member    struct {
		Mid            string `json:"mid"`
		Uname          string `json:"uname"`
		Sex            string `json:"sex"`
		Sign           string `json:"sign"`
		Avatar         string `json:"avatar"`
		Rank           string `json:"rank"`
		FaceNftNew     int    `json:"face_nft_new"`
		IsSeniorMember int    `json:"is_senior_member"`
		LevelInfo      struct {
			CurrentLevel int `json:"current_level"`
			CurrentMin   int `json:"current_min"`
			CurrentExp   int `json:"current_exp"`
			NextExp      int `json:"next_exp"`
		} `json:"level_info"`
		Vip struct {
			VipType       int    `json:"vipType"`
			VipDueDate    int64  `json:"vipDueDate"`
			DueRemark     string `json:"dueRemark"`
			AccessStatus  int    `json:"accessStatus"`
			VipStatus     int    `json:"vipStatus"`
			VipStatusWarn string `json:"vipStatusWarn"`
		} `json:"vip"`
		FansDetail any `json:"fans_detail"`
	} `json:"member"`
	Content struct {
		Message  string    `json:"message"`
		Pictures []Picture `json:"pictures"`
		Members  []any     `json:"members"`
		Emote    struct {
			NAMING_FAILED struct {
				ID        int    `json:"id"`
				PackageID int    `json:"package_id"`
				State     int    `json:"state"`
				Type      int    `json:"type"`
				Attr      int    `json:"attr"`
				Text      string `json:"text"`
				URL       string `json:"url"`
				Meta      struct {
					Size int `json:"size"`
				} `json:"meta"`
				Mtime     int    `json:"mtime"`
				JumpTitle string `json:"jump_title"`
			} `json:"[吃瓜]"`
		} `json:"emote"`
		JumpURL struct {
		} `json:"jump_url"`
		MaxLine int `json:"max_line"`
	} `json:"content"`
	Replies      []ReplyItem `json:"replies"`
	Invisible    bool        `json:"invisible"`
	ReplyControl struct {
		Following bool   `json:"following"`
		MaxLine   int    `json:"max_line"`
		TimeDesc  string `json:"time_desc"`
		Location  string `json:"location"`
	} `json:"reply_control"`
}

type CommentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Page struct {
			Num    int `json:"num"`
			Size   int `json:"size"`
			Count  int `json:"count"`
			Acount int `json:"acount"`
		} `json:"page"`
		Replies    []ReplyItem `json:"replies"`
		TopReplies []ReplyItem `json:"top_replies"`
	} `json:"data"`
}
