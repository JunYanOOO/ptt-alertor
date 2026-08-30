package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models"
	"github.com/Ptt-Alertor/ptt-alertor/models/subscription"
	"github.com/Ptt-Alertor/ptt-alertor/models/user"
	"github.com/Ptt-Alertor/ptt-alertor/myutil"
	"github.com/julienschmidt/httprouter"
)

func UserFind(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	u := visibleUser(models.User().Find(params.ByName("account")))
	uJSON, err := json.Marshal(u)
	if err != nil {
		myutil.LogJSONEncode(err, u)
	}
	fmt.Fprintf(w, "%s", uJSON)
}

func UserAll(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	us := models.User().All()

	data := struct {
		Total, Line, Messenger, Telegram, Discord, IdleUser, BlockUser int
		SubCount, BoardCount, KeywordCount, AuthorCount, PushSumCount  int
		User, Room, Group                                              int
		Users                                                          []*user.User
		Features                                                       config.Features
	}{}
	data.Features = config.Current().Features
	data.Users = make([]*user.User, 0, len(us))
	data.Total = len(us)
	for _, u := range us {
		visible := visibleUser(*u)
		u = &visible
		data.Users = append(data.Users, u)
		if !u.Enable {
			data.BlockUser++
		}
		if u.Profile.Line != "" {
			data.Line++
		}
		if u.Profile.Messenger != "" {
			data.Messenger++
		}
		if u.Profile.Telegram != "" {
			data.Telegram++
		}
		if u.Profile.DiscordChannel != "" {
			data.Discord++
		}
		switch u.Profile.Type {
		case "user", "":
			data.User++
		case "room":
			data.Room++
		case "group":
			data.Group++
		}
		subCount := len(u.Subscribes)
		data.SubCount += subCount
		if subCount == 0 {
			data.IdleUser++
		}
		data.BoardCount += subCount
		for _, s := range u.Subscribes {
			data.KeywordCount += len(s.Keywords)
			data.AuthorCount += len(s.Authors)
			if s.PushSum.Up != 0 || s.PushSum.Down != 0 {
				data.PushSumCount++
			}
		}
	}
	t, err := template.ParseFiles("public/user.tpl")
	if err != nil {
		panic(err)
	}
	t.Execute(w, data)
}

func UserCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	u := models.User()
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		myutil.LogJSONDecode(err, r.Body)
		http.Error(w, "not a json valid format", 400)
		return
	}
	if err := preserveAndValidateDisabledSubscriptions(u, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.Save(); err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func UserModify(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	account := params.ByName("account")
	u := models.User()
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		myutil.LogJSONDecode(err, r.Body)
		http.Error(w, "not a json valid format", 400)
		return
	}

	if u.Profile.Account != account {
		http.Error(w, "account does not match", 400)
		return
	}
	existing := models.User().Find(account)
	if err := preserveAndValidateDisabledSubscriptions(u, &existing); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := u.Update(); err != nil {
		http.Error(w, err.Error(), 400)
	}

}

func visibleUser(source user.User) user.User {
	features := config.Current().Features
	visible := source
	visible.Subscribes = make(subscription.Subscriptions, 0, len(source.Subscribes))
	for _, sourceSubscription := range source.Subscribes {
		sub := sourceSubscription
		if !features.KeywordTracking {
			sub.Keywords = nil
		}
		if !features.AuthorTracking {
			sub.Authors = nil
		}
		if !features.PushSumTracking {
			sub.PushSum = subscription.EmptyPushSum
		}
		if !features.ArticleCommentTracking {
			sub.Articles = nil
		}
		if len(sub.Keywords) != 0 || len(sub.Authors) != 0 || len(sub.Articles) != 0 || sub.PushSum != subscription.EmptyPushSum {
			visible.Subscribes = append(visible.Subscribes, sub)
		}
	}
	return visible
}

func preserveAndValidateDisabledSubscriptions(incoming *user.User, existing *user.User) error {
	features := config.Current().Features
	existingByBoard := make(map[string]subscription.Subscription)
	if existing != nil {
		for _, sub := range existing.Subscribes {
			existingByBoard[strings.ToLower(sub.Board)] = sub
		}
	}

	for _, sub := range incoming.Subscribes {
		previous, existed := existingByBoard[strings.ToLower(sub.Board)]
		if !features.KeywordTracking && len(sub.Keywords) != 0 && (!existed || !sameStrings(sub.Keywords, previous.Keywords)) {
			return errors.New("關鍵字追蹤功能目前停用，無法新增或變更關鍵字設定")
		}
		if !features.AuthorTracking && len(sub.Authors) != 0 && (!existed || !sameStrings(sub.Authors, previous.Authors)) {
			return errors.New("作者追蹤功能目前停用，無法新增或變更作者設定")
		}
		if !features.PushSumTracking && sub.PushSum != subscription.EmptyPushSum && (!existed || sub.PushSum != previous.PushSum) {
			return errors.New("推噓文數追蹤功能目前停用，無法新增或變更推噓文數設定")
		}
		if !features.ArticleCommentTracking && len(sub.Articles) != 0 && (!existed || !sameStrings(sub.Articles, previous.Articles)) {
			return errors.New("單篇文章推文追蹤功能目前停用，無法新增或變更文章設定")
		}
	}

	if existing == nil {
		return nil
	}

	incomingByBoard := make(map[string]int, len(incoming.Subscribes))
	for index, sub := range incoming.Subscribes {
		incomingByBoard[strings.ToLower(sub.Board)] = index
	}
	for _, previous := range existing.Subscribes {
		key := strings.ToLower(previous.Board)
		index, found := incomingByBoard[key]
		if !found {
			preserved := subscription.Subscription{Board: previous.Board}
			copyDisabledSubscriptionFields(&preserved, previous, features)
			if len(preserved.Keywords) != 0 || len(preserved.Authors) != 0 || len(preserved.Articles) != 0 || preserved.PushSum != subscription.EmptyPushSum {
				incoming.Subscribes = append(incoming.Subscribes, preserved)
			}
			continue
		}
		copyDisabledSubscriptionFields(&incoming.Subscribes[index], previous, features)
	}
	return nil
}

func copyDisabledSubscriptionFields(target *subscription.Subscription, source subscription.Subscription, features config.Features) {
	if !features.KeywordTracking {
		target.Keywords = append(target.Keywords[:0], source.Keywords...)
	}
	if !features.AuthorTracking {
		target.Authors = append(target.Authors[:0], source.Authors...)
	}
	if !features.PushSumTracking {
		target.PushSum = source.PushSum
	}
	if !features.ArticleCommentTracking {
		target.Articles = append(target.Articles[:0], source.Articles...)
	}
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
