package common

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
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

func HeartbeatCheck() error {
	var clientName = new(string)
	if IsClientOffline(clientName) {
		return fmt.Errorf("%v is offline, please retry after check it is ready", *clientName)
	}
	log.Info("", "\tKeenTuned heartbeat check : keentuned and client connection success")

	go monitorClientStatus(clientName)
	return nil
}

func monitorClientStatus(clientName *string) {
	signalChan := make(chan os.Signal, 1)

	// notify sysem signal, such as ctrl + C
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var faultCount int
	ticker := time.NewTicker(time.Duration(config.KeenTune.HeartbeatTime) * time.Second)

	for {
		select {
		case <-ticker.C:
			if IsClientOffline(clientName) {
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

			if GetRunningTask() != "" {
				http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			config.ServeFinish <- true
			return
		}
	}
}

func IsClientOffline(clientName *string) bool {
	for _, group := range config.KeenTune.Group {
		for _, ip := range group.IPs {
			targetURI := fmt.Sprintf("%v:%v/status", ip, group.Port)
			id := config.KeenTune.Target.IPMap[ip]
			if checkOffline(targetURI) {
				*clientName = fmt.Sprintf("target %v", id)
				return true
			}
		}

	}

	for _, benchgroup := range config.KeenTune.BenchGroup {
		for _, benchip := range benchgroup.SrcIPs {
			benchURI := fmt.Sprintf("%v:%v/status",benchip,benchgroup.SrcPort)
			id := config.KeenTune.Bench.BenchIPMap[benchip]
			if checkOffline(benchURI) {
				*clientName = fmt.Sprintf("bench %v", id)
				return true
			}
		}
	}
	return false
}

func checkOffline(uri string) bool {
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
