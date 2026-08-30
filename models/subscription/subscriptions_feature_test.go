package subscription

import (
	"strings"
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/myutil"
)

func TestStringWithOptionsHidesDisabledSubscriptions(t *testing.T) {
	subscriptions := Subscriptions{{
		Board:    "gossiping",
		Keywords: myutil.StringSlice{"新聞"},
		Authors:  myutil.StringSlice{"author"},
		Articles: myutil.StringSlice{"M.123.A.123"},
		PushSum:  PushSum{Up: 10},
	}}

	got := subscriptions.StringWithOptions(true, false, false, false)
	if !strings.Contains(got, "新聞") {
		t.Error("enabled keyword subscription should be rendered")
	}
	for _, hidden := range []string{"作者", "推文數", "推文追蹤", "author", "M.123.A.123"} {
		if strings.Contains(got, hidden) {
			t.Errorf("disabled subscription content %q should be hidden in %q", hidden, got)
		}
	}
}
