package core

import (
	"blblcd/model"
	"testing"
)

func TestCollectSubCommentsKeepsPreviews(t *testing.T) {
	roots := []model.ReplyItem{
		{Rpid: 1, Rcount: 2, Replies: []model.ReplyItem{{Rpid: 11}, {Rpid: 12}}},
		{Rpid: 2, Rcount: 0},
		{Rpid: 3, Rcount: 1, Replies: []model.ReplyItem{{Rpid: 31}}},
	}
	got := collectSubComments(roots, &model.Option{Concurrency: 1})
	if len(got) != 3 {
		t.Fatalf("got %d sub comments, want 3", len(got))
	}
	want := []int64{11, 12, 31}
	for i, id := range want {
		if got[i].Rpid != id {
			t.Errorf("sub[%d] rpid = %d, want %d", i, got[i].Rpid, id)
		}
	}
}
