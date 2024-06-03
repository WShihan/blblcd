package model

type Comment struct {
	Uname         string //姓名
	Sex           string //性别
	Content       string //评论内容
	Rpid          int64  //评论id
	Oid           int    //评论区id
	Bvid          string //视频bv
	Mid           int    //发送者id
	Parent        int    //若为一级评论则为 0若为二级评论则为根评论 rpid大于二级评论为上一级评 论 rpid
	Fansgrade     int    //是否粉丝标签
	Ctime         int    //评论时间戳
	Like          int    //喜欢数
	Following     int    //是否关注
	Current_level int    //当前等级
	Location      string //位置
	Time_desc     string //时间间隔
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
	RootStr   string `json:"root_str"`
	ParentStr string `json:"parent_str"`
	DialogStr string `json:"dialog_str"`
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
		Following      int    `json:"following"`
		Senior         struct {
		} `json:"senior"`
		LevelInfo struct {
			CurrentLevel int `json:"current_level"`
			CurrentMin   int `json:"current_min"`
			CurrentExp   int `json:"current_exp"`
			NextExp      int `json:"next_exp"`
		} `json:"level_info"`
	} `json:"member"`
	Content struct {
		Message string `json:"message"`
		Members []any  `json:"members"`
		JumpURL struct {
		} `json:"jump_url"`
		MaxLine int `json:"max_line"`
	} `json:"content"`
	Replies      []ReplyItem `json:"replies"`
	ReplyControl struct {
		MaxLine           int    `json:"max_line"`
		SubReplyEntryText string `json:"sub_reply_entry_text"`
		SubReplyTitleText string `json:"sub_reply_title_text"`
		TimeDesc          string `json:"time_desc"`
		Location          string `json:"location"`
		FoldPictures      bool   `json:"fold_pictures"`
	} `json:"reply_control"`
}

type CommentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Cursor struct {
			IsBegin   bool   `json:"is_begin"`
			Prev      int    `json:"prev"`
			Next      int    `json:"next"`
			IsEnd     bool   `json:"is_end"`
			Mode      int    `json:"mode"`
			ModeText  string `json:"mode_text"`
			AllCount  int    `json:"all_count"`
			SessionID string `json:"session_id"`
		} `json:"cursor"`
		Replies []ReplyItem `json:"replies"`
	} `json:"data"`
}
