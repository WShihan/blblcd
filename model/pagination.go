package model

type PaginationOffsetData struct {
	Pn int `json:"pn"`
}

type PaginationOffset struct {
	Type      int                  `json:"type"`
	Direction int                  `json:"direction"`
	Data      PaginationOffsetData `json:"data"`
	SessionId string               `json:"session_id"`
}

type Pagination struct {
	Offset string `json:"offset"`
}
