package jobs

import (
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models/subscription"
	"github.com/Ptt-Alertor/ptt-alertor/myutil"
)

func TestGeneratorOnlyRecognizesEnabledSubscriptions(t *testing.T) {
	generator := NewGenerator(config.Features{KeywordTracking: true})
	if !generator.hasEnabledSubscription(subscription.Subscription{Keywords: myutil.StringSlice{"新聞"}}) {
		t.Error("enabled keyword subscription should be recognized")
	}
	for _, disabled := range []subscription.Subscription{
		{Authors: myutil.StringSlice{"author"}},
		{Articles: myutil.StringSlice{"M.123.A.123"}},
		{PushSum: subscription.PushSum{Up: 10}},
	} {
		if generator.hasEnabledSubscription(disabled) {
			t.Errorf("disabled subscription should not be recognized: %#v", disabled)
		}
	}
}
