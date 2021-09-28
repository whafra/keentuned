package common

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ClientState struct {
	Status string `json:"status"`
}

const (
	// MaxReconnectionTime ...
	MaxReconnectionTime = 3
)

func HeartbeatCheck() {
	signalChan := make(chan os.Signal, 1)

	// notify sysem signal, such as ctrl + C
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var faultCount int
	uri := config.KeenTune.TargetIP + ":" + config.KeenTune.TargetPort + "/status"
	ticker := time.NewTicker(time.Duration(config.KeenTune.HeartbeatTime) * time.Second)

	if IsClientOffline(uri) {
		log.Error("", "target machine client is offline, please retry after check it is ready")
		return
	} else {
		log.Info("", "\tKeenTuned heartbeat check : keentuned and target connection success")
	}

	for {
		select {
		case <-ticker.C:
			if IsClientOffline(uri) {
				faultCount++
				log.Infof("", "keentuned detected that the client was offline for the %vth time", faultCount)
			}

			if faultCount == MaxReconnectionTime {
				log.Info("", "Heartbeat Check the client is offline")
				config.ServeFinish <- true
				return
			}

		case <-config.ProgramNeedExit:
			log.Debug("", "Heartbeat Check program is finish")
			config.ServeFinish <- true
			return

		case <-signalChan:
			log.Debug("", "Heartbeat Check program is interrupt")
			http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			config.ServeFinish <- true
			return
		}
	}
}

func IsClientOffline(uri string) bool {
	bytes, err := http.RemoteCall("GET", uri, nil)
	if err != nil {
		log.Errorf("", "remotcall return err: %v", err)
		return true
	}

	var clientState ClientState
	if err = json.Unmarshal(bytes, &clientState); err != nil {
		log.Errorf("", "unmashal client state [%+v] failed", string(bytes))
		return true
	}

	return clientState.Status != "alive"
}

