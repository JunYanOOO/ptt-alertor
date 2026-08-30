package command

import (
	"strconv"
	"strings"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models"
	"github.com/Ptt-Alertor/ptt-alertor/models/author"
	"github.com/Ptt-Alertor/ptt-alertor/models/keyword"
	"github.com/Ptt-Alertor/ptt-alertor/models/pushsum"
	"github.com/Ptt-Alertor/ptt-alertor/models/subscription"
	"github.com/Ptt-Alertor/ptt-alertor/models/user"
	"github.com/Ptt-Alertor/ptt-alertor/myutil"
)

type updateAction func(u *user.User, sub subscription.Subscription, inputs ...string) error

// EnabledCommands returns only commands for features enabled in config.yaml.
func EnabledCommands() map[string]map[string]string {
	features := config.Current().Features
	enabled := map[string]map[string]string{
		"一般": {
			"指令": "可使用的指令清單",
			"清單": "設定的追蹤清單",
			"排行": "熱門追蹤排行",
		},
	}
	if features.KeywordTracking {
		enabled["關鍵字相關"] = Commands["關鍵字相關"]
		enabled["進階應用"] = Commands["進階應用"]
	}
	if features.AuthorTracking {
		enabled["作者相關"] = Commands["作者相關"]
	}
	if features.PushSumTracking {
		enabled["推噓文數相關"] = Commands["推噓文數相關"]
	}
	if features.ArticleCommentTracking {
		enabled["推文相關"] = Commands["推文相關"]
	}
	return enabled
}

func addKeywords(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Keywords = inputs
	err := u.Subscribes.Add(sub)
	if err == nil {
		err = keyword.AddSubscriber(sub.Board, u.Profile.Account)
	}
	return err
}

func removeKeywords(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Keywords = inputs
	if inputs[0] == "*" {
		for _, uSub := range u.Subscribes {
			if strings.EqualFold(uSub.Board, sub.Board) {
				sub.Keywords = make(myutil.StringSlice, len(uSub.Keywords))
				copy(sub.Keywords, uSub.Keywords)
			}
		}
	}
	err := u.Subscribes.Remove(sub)
	if err == nil {
		for _, uSub := range u.Subscribes {
			if strings.EqualFold(sub.Board, uSub.Board) && len(uSub.Keywords) > 0 {
				return nil
			}
		}
		err = keyword.RemoveSubscriber(sub.Board, u.Profile.Account)
	}
	return err
}

func addAuthors(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Authors = inputs
	err := u.Subscribes.Add(sub)
	if err == nil {
		err = author.AddSubscriber(sub.Board, u.Profile.Account)
	}
	return err
}

func removeAuthors(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Authors = inputs
	if inputs[0] == "*" {
		for _, uSub := range u.Subscribes {
			if strings.EqualFold(uSub.Board, sub.Board) {
				sub.Authors = make(myutil.StringSlice, len(uSub.Authors))
				copy(sub.Authors, uSub.Authors)
			}
		}
	}
	err := u.Subscribes.Remove(sub)
	if err == nil {
		for _, uSub := range u.Subscribes {
			if strings.EqualFold(sub.Board, uSub.Board) && len(uSub.Authors) > 0 {
				return nil
			}
		}
		err = author.RemoveSubscriber(sub.Board, u.Profile.Account)
	}
	return err
}

func updatePushUp(u *user.User, sub subscription.Subscription, inputs ...string) error {
	up, err := strconv.Atoi(inputs[0])
	if err != nil {
		return err
	}
	for _, s := range u.Subscribes {
		if strings.EqualFold(s.Board, sub.Board) {
			sub.PushSum.Down = s.PushSum.Down
		}
	}
	sub.PushSum.Up = up
	err = u.Subscribes.Update(sub)
	if err == nil {
		err = dealPushSum(u.Profile.Account, sub)
	}
	return err
}

func updatePushDown(u *user.User, sub subscription.Subscription, inputs ...string) error {
	down, err := strconv.Atoi(inputs[0])
	if err != nil {
		return err
	}
	for _, s := range u.Subscribes {
		if strings.EqualFold(s.Board, sub.Board) {
			sub.PushSum.Up = s.PushSum.Up
		}
	}
	sub.PushSum.Down = down
	err = u.Subscribes.Update(sub)
	if err == nil {
		err = dealPushSum(u.Profile.Account, sub)
	}
	return err
}

func dealPushSum(account string, sub subscription.Subscription) (err error) {
	if !pushsum.Exist(sub.Board) {
		err = pushsum.Add(sub.Board)
	}
	if sub.Up == 0 {
		err = pushsum.DelDiffList(account, sub.Board, "up")
	}
	if sub.Down == 0 {
		err = pushsum.DelDiffList(account, sub.Board, "down")
	}
	if sub.Up == 0 && sub.Down == 0 {
		err = pushsum.RemoveSubscriber(sub.Board, account)
	} else {
		err = pushsum.AddSubscriber(sub.Board, account)
	}
	return err
}

func addArticles(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Articles = inputs
	a := models.Article()
	a.Code = inputs[0]
	a.AddSubscriber(u.Profile.Account)
	return u.Subscribes.Add(sub)
}

func removeArticles(u *user.User, sub subscription.Subscription, inputs ...string) error {
	sub.Articles = inputs
	a := models.Article()
	a.Code = inputs[0]
	a.RemoveSubscriber(u.Profile.Account)
	return u.Subscribes.Remove(sub)
}
