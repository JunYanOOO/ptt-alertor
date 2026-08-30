package jobs

import (
	"net/http"
	"time"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/Ptt-Alertor/ptt-alertor/config"
)

type pttMonitor struct {
	duration time.Duration
	retry    int
	features config.Features
}

func NewPttMonitor(features config.Features) *pttMonitor {
	return &pttMonitor{
		duration: 1 * time.Minute,
		retry:    3,
		features: features,
	}
}

func (pm pttMonitor) Run() {
	log.Info("Start Ptt Monitor")

	var errorCounter = 0
	var url = "https://www.ptt.cc/bbs/index.html"
	ticker := time.NewTicker(pm.duration)
	for range ticker.C {
		resp, err := http.Get(url)
		if err != nil {
			log.WithError(err).Error("HTTP Get Error")
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Info("Ptt is alive")
			if errorCounter >= pm.retry {
				log.Info("Ptt is back to life")
				pm.startTrackingJobs()
			}
			errorCounter = 0
		}
		if err == nil && resp.StatusCode != http.StatusOK {
			if errorCounter < pm.retry {
				log.Info("Ptt is dying")
			}
			if errorCounter == pm.retry {
				log.Info("Ptt is Dead")
				pm.stopTrackingJobs()
			}
			errorCounter++
		}
	}
}

func (pm pttMonitor) startTrackingJobs() {
	if pm.features.KeywordTracking || pm.features.AuthorTracking {
		go NewChecker(pm.features).Run()
	}
	if pm.features.PushSumTracking {
		go NewPushSumChecker().Run()
	}
	if pm.features.ArticleCommentTracking {
		go NewCommentChecker().Run()
	}
}

func (pm pttMonitor) stopTrackingJobs() {
	if pm.features.KeywordTracking || pm.features.AuthorTracking {
		go NewChecker(pm.features).Stop()
	}
	if pm.features.PushSumTracking {
		go NewPushSumChecker().Stop()
	}
	if pm.features.ArticleCommentTracking {
		go NewCommentChecker().Stop()
	}
}
