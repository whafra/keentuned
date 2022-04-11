package common

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	"os"
	"os/signal"
	"strings"
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
	var offline = false
	offline = IsTargetOffline(clientName)

	benchStatus := IsBenchOffline(clientName)
	offline = offline || benchStatus

	// check brain
	if isBrainOffline() {
		*clientName += fmt.Sprintf("brain client")
		offline = true
	}

	*clientName = strings.TrimSuffix(*clientName, ", ")
	return offline
}

func IsTargetOffline(clientName *string) bool {
	var offline bool
	
	for _, group := range config.KeenTune.Group {
		for index, ip := range group.IPs {
			targetURI := fmt.Sprintf("%v:%v/status", ip, group.Port)
			if checkOffline(targetURI) {
				*clientName += fmt.Sprintf("group-%v.target-%v, ", group.GroupNo, index+1)
				offline = true
			}
		}

	}

	return offline
}

func IsBenchOffline(clientName *string) bool {
	var offline bool

	for groupID, benchgroup := range config.KeenTune.BenchGroup {
		for ipIndex, benchip := range benchgroup.SrcIPs {
			benchURI := fmt.Sprintf("%v:%v/status",benchip,benchgroup.SrcPort)
			if checkOffline(benchURI) {
				*clientName += fmt.Sprintf("group-%v.bench-%v, ", groupID+1, ipIndex+1)
				offline = true
			}
		}
	}

	return offline
}


func checkOffline(uri string) bool {
	bytes, err := http.RemoteCall("GET", uri, nil)
	if err != nil {
		return true
	}

	var clientState ClientState
	if err = json.Unmarshal(bytes, &clientState); err != nil {
		log.Errorf("", "unmashal client state [%+v] failed", string(bytes))
		return true
	}

	return clientState.Status != "alive"
}

func StartCheck() error {
	var clientName = new(string)
	if IsClientOffline(clientName) {
		return fmt.Errorf("Found %v offline, please get them (it) ready before use", *clientName)
	}

	return nil
}

func isBrainOffline() bool {
	url := fmt.Sprintf("%v:%v/sensitize_list", config.KeenTune.BrainIP, config.KeenTune.BrainPort)
	_, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		return true
	}

	return false
}

