package controllers

import (
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models/subscription"
	"github.com/Ptt-Alertor/ptt-alertor/models/user"
	"github.com/Ptt-Alertor/ptt-alertor/myutil"
)

func TestVisibleUserHidesDisabledSubscriptions(t *testing.T) {
	setKeywordOnlyConfig(t)
	source := user.User{Subscribes: subscription.Subscriptions{{
		Board:    "gossiping",
		Keywords: myutil.StringSlice{"新聞"},
		Authors:  myutil.StringSlice{"author"},
		Articles: myutil.StringSlice{"M.123.A.123"},
		PushSum:  subscription.PushSum{Up: 10},
	}}}

	visible := visibleUser(source)
	if len(visible.Subscribes) != 1 {
		t.Fatalf("visible subscriptions = %d, want 1", len(visible.Subscribes))
	}
	got := visible.Subscribes[0]
	if len(got.Keywords) != 1 || len(got.Authors) != 0 || len(got.Articles) != 0 || got.PushSum != subscription.EmptyPushSum {
		t.Fatalf("visible subscription = %#v, want keyword data only", got)
	}
}

func TestPreserveDisabledSubscriptionsDuringKeywordUpdate(t *testing.T) {
	setKeywordOnlyConfig(t)
	existing := user.User{Subscribes: subscription.Subscriptions{{
		Board:    "gossiping",
		Keywords: myutil.StringSlice{"舊關鍵字"},
		Authors:  myutil.StringSlice{"author"},
		Articles: myutil.StringSlice{"M.123.A.123"},
		PushSum:  subscription.PushSum{Up: 10},
	}}}
	incoming := user.User{Subscribes: subscription.Subscriptions{{
		Board:    "gossiping",
		Keywords: myutil.StringSlice{"新關鍵字"},
	}}}

	if err := preserveAndValidateDisabledSubscriptions(&incoming, &existing); err != nil {
		t.Fatalf("preserveAndValidateDisabledSubscriptions() error = %v", err)
	}
	got := incoming.Subscribes[0]
	if len(got.Keywords) != 1 || got.Keywords[0] != "新關鍵字" {
		t.Errorf("enabled keyword update was not preserved: %#v", got.Keywords)
	}
	if len(got.Authors) != 1 || len(got.Articles) != 1 || got.PushSum.Up != 10 {
		t.Errorf("disabled subscription data was not preserved: %#v", got)
	}
}

func TestRejectsChangingDisabledSubscription(t *testing.T) {
	setKeywordOnlyConfig(t)
	existing := user.User{Subscribes: subscription.Subscriptions{{
		Board:   "gossiping",
		Authors: myutil.StringSlice{"old-author"},
	}}}
	incoming := user.User{Subscribes: subscription.Subscriptions{{
		Board:   "gossiping",
		Authors: myutil.StringSlice{"new-author"},
	}}}

	if err := preserveAndValidateDisabledSubscriptions(&incoming, &existing); err == nil {
		t.Error("changing a disabled subscription should be rejected")
	}
}

func setKeywordOnlyConfig(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.Set(config.Config{Features: config.Features{KeywordTracking: true}})
	t.Cleanup(func() { config.Set(previous) })
}
