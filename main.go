package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	glog "github.com/go-kit/kit/log"
	"github.com/grafana/loki/pkg/promtail/client"
	lokiflag "github.com/grafana/loki/pkg/util/flagext"
	"github.com/lrsmith/go-icinga2-api/iapi"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type Events struct {
	types []string `json:"types"`
	queue string   `json:"queue"`
}

func main() {
	icinga_api, _ := iapi.New(
		"icinga2-director",
		"icinga2",
		"https://localhost:5665/v1",
		true,
	)
	if err := icinga_api.Connect(); err != nil {
		log.Fatal(err)
	}

	var ClientConfigs []client.Config
	u, _ := url.Parse("http://localhost:3100/loki/api/v1/push")
	cfg := client.Config{
		URL:            flagext.URLValue{URL: u},
		BatchWait:      100 * time.Millisecond,
		BatchSize:      10,
		Client:         config.HTTPClientConfig{},
		BackoffConfig:  util.BackoffConfig{MinBackoff: 1 * time.Millisecond, MaxBackoff: 2 * time.Millisecond, MaxRetries: 3},
		ExternalLabels: lokiflag.LabelSet{},
		Timeout:        1 * time.Second,
		TenantID:       "",
	}
	ClientConfigs = append(ClientConfigs, cfg)

	c, err := client.NewMulti(glog.NewNopLogger(), ClientConfigs...)
	if err != nil {
		log.Fatal(err)
	}

	for {
		response, responseErr := icinga_api.NewAPIRequest("POST", "/events", []byte(`{"types": ["CheckResult","StateChange","Notification","AcknowledgementSet","AcknowledgementCleared","CommentAdded","CommentRemoved","DowntimeAdded","DowntimeRemoved","DowntimeStarted","DowntimeTriggered"],"queue":"icingatail"}`))

		if responseErr == nil {
			if response != nil {
				err = c.Handle(model.LabelSet{"foo": "bar"}, time.Now(), fmt.Sprintf("%v", response.Results))
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			log.Fatal(responseErr)
		}
	}
}
