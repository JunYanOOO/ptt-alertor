package subscription

import (
	"sort"
	"strings"

	"github.com/Ptt-Alertor/ptt-alertor/models/board"
)

type Subscriptions []Subscription

func (ss Subscriptions) String() string {
	return ss.StringWithOptions(true, true, true, true)
}

// StringWithOptions renders only enabled subscription types without changing
// the stored subscriptions.
func (ss Subscriptions) StringWithOptions(keywords, authors, pushSum, articles bool) string {

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Board < ss[j].Board
	})

	sections := make([]string, 0, 4)
	if keywords {
		str := "關鍵字"
		for _, sub := range ss {
			if sub.String() != "" {
				str += "\n" + sub.String()
			}
		}
		sections = append(sections, str)
	}
	if authors {
		str := "作者"
		for _, sub := range ss {
			if sub.StringAuthor() != "" {
				str += "\n" + sub.StringAuthor()
			}
		}
		sections = append(sections, str)
	}
	if pushSum {
		str := "推文數"
		for _, sub := range ss {
			if sub.StringPushSum() != "" {
				str += "\n" + sub.StringPushSum()
			}
		}
		sections = append(sections, str)
	}
	if articles {
		sections = append(sections, "推文\n請輸入「推文清單」查看推文追蹤列表。")
	}
	return strings.Join(sections, "\n----\n")
}

func (ss Subscriptions) StringCommentList() string {
	var str string
	for _, sub := range ss {
		if sub.StringArticle() != "" {
			str += sub.StringArticle() + "\n"
		}
	}
	return str
}

func (ss *Subscriptions) Add(sub Subscription) error {
	if ok, suggestion := board.CheckBoardExist(sub.Board); !ok {
		return board.BoardNotExistError{Suggestion: suggestion}
	}
	sub.CleanUp()
	for i, s := range *ss {
		if strings.EqualFold(s.Board, sub.Board) {
			s.Keywords.AppendNonRepeat(sub.Keywords, false)
			s.Authors.AppendNonRepeat(sub.Authors, false)
			s.Articles.AppendNonRepeat(sub.Articles, false)
			(*ss)[i] = s
			return nil
		}
	}
	*ss = append(*ss, sub)

	return nil
}

func (ss *Subscriptions) Remove(sub Subscription) error {
	sub.CleanUp()
	for i := 0; i < len(*ss); i++ {
		s := (*ss)[i]
		if strings.EqualFold(s.Board, sub.Board) {
			s.DeleteKeywords(sub.Keywords)
			s.DeleteAuthors(sub.Authors)
			s.DeleteArticles(sub.Articles)
			(*ss)[i] = s
			if isSubEmpty((*ss)[i]) {
				*ss = append((*ss)[:i], (*ss)[i+1:]...)
				i--
				return nil
			}
		}
	}
	return nil
}

func (ss *Subscriptions) Update(sub Subscription) error {
	if ok, suggestion := board.CheckBoardExist(sub.Board); !ok {
		return board.BoardNotExistError{Suggestion: suggestion}
	}
	for i := 0; i < len(*ss); i++ {
		s := (*ss)[i]
		if strings.EqualFold(s.Board, sub.Board) {
			s.PushSum = sub.PushSum
			(*ss)[i] = s
			if isSubEmpty((*ss)[i]) {
				*ss = append((*ss)[:i], (*ss)[i+1:]...)
				i--
			}
			return nil
		}
	}
	*ss = append(*ss, sub)
	return nil
}

func (ss *Subscriptions) Delete(sub Subscription) error {
	for i := 0; i < len(*ss); i++ {
		s := (*ss)[i]
		if strings.EqualFold(s.Board, sub.Board) {
			*ss = append((*ss)[:i], (*ss)[i+1:]...)
			i--
			return nil
		}
	}
	return nil
}

func isSubEmpty(sub Subscription) bool {
	return len(sub.Keywords) == 0 && len(sub.Authors) == 0 && len(sub.Articles) == 0 && sub.PushSum == PushSum{}
}
