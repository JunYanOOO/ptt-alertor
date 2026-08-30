package jobs

import (
	"sync"

	log "github.com/Ptt-Alertor/logrus"

	"time"

	"github.com/Ptt-Alertor/ptt-alertor/config"
	"github.com/Ptt-Alertor/ptt-alertor/models"
	"github.com/Ptt-Alertor/ptt-alertor/models/board"
)

type Fetcher struct {
	features config.Features
}

func NewFetcher(features config.Features) *Fetcher {
	return &Fetcher{features: features}
}

func (f Fetcher) Run() {
	boards := models.Board().All()

	var wg sync.WaitGroup
	for _, bd := range boards {
		if !boardHasEnabledSubscribers(bd.Name, f.features) {
			continue
		}
		wg.Add(1)
		go func(bd board.Board) {
			defer wg.Done()
			bd.Articles = bd.FetchArticles()
			bd.Save()
			log.WithField("board", bd.Name).Info("Fetched")
		}(*bd)
		time.Sleep(50 * time.Millisecond)
	}
	wg.Wait()
	log.Info("All fetcher done")
}
