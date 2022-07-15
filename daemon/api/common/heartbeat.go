package common

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
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

	go monitorClientStatus(IsClientOffline, clientName, nil)
	return nil
}

func monitorClientStatus(monitor interface{}, clientName *string, group []bool) {
	signalChan := make(chan os.Signal, 1)

	// notify system signal, such as ctrl + C
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var faultCount int
	ticker := time.NewTicker(time.Duration(config.KeenTune.HeartbeatTime) * time.Second)

	for {
		select {
		case <-ticker.C:
			if monitorImply(monitor, clientName, group) {
				faultCount++
				log.Infof("", "keentuned detected that '%v' was offline for the %vth time", *clientName, faultCount)
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

			if m.GetRunningTask() != "" {
				ResetJob()
				http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			config.ServeFinish <- true
			return
		}
	}
}

func monitorImply(monitor interface{}, clientName *string, group []bool) bool {
	switch f := monitor.(type) {
	case func(*string) bool:
		return f(clientName)
	case func([]bool, *string) bool:
		return f(group, clientName)
	}

	return false
}

func IsClientOffline(clientName *string) bool {
	var offline = false
	offline = IsTargetOffline(clientName)

	benchStatus := IsBenchOffline(clientName)
	offline = offline || benchStatus

	// check brain
	brainOffline := IsBrainOffline(clientName)
	offline = offline || brainOffline

	*clientName = strings.TrimSuffix(*clientName, ", ")
	return offline
}

func IsTargetOffline(clientName *string) bool {
	var offline bool

	for _, group := range config.KeenTune.Group {
		for index, ip := range group.IPs {
			targetURI := fmt.Sprintf("%v:%v/status", ip, group.Port)
			if checkOffline(targetURI) {
				*clientName += fmt.Sprintf("target%v-%v, ", group.GroupNo, index+1)
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
			benchURI := fmt.Sprintf("%v:%v/status", benchip, benchgroup.SrcPort)
			if checkOffline(benchURI) {
				*clientName += fmt.Sprintf("bench_src%v-%v, ", groupID+1, ipIndex+1)
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

func IsBrainOffline(clientName *string) bool {
	url := fmt.Sprintf("%v:%v/avaliable", config.KeenTune.BrainIP, config.KeenTune.BrainPort)
	_, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		*clientName += fmt.Sprintf("brain client")
		return true
	}

	return false
}

func checkSettableTarget(group []bool) error {
	inputLen := len(group)
	hasLen := len(config.KeenTune.Group)
	if inputLen != hasLen {
		return fmt.Errorf("target group out of range, has count %v, input count %v", hasLen, inputLen)
	}

	return nil
}

func IsSetTargetOffline(group []bool, clientName *string) bool {
	err := checkSettableTarget(group)
	if err != nil {
		*clientName += err.Error()
		return true
	}

	var offline bool
	for i, settable := range group {
		if !settable {
			continue
		}

		for index, ip := range config.KeenTune.Group[i].IPs {
			targetURI := fmt.Sprintf("%v:%v/status", ip, config.KeenTune.Group[i].Port)
			if checkOffline(targetURI) {
				*clientName += fmt.Sprintf("target%v-%v, ", config.KeenTune.Group[i].GroupNo, index+1)
				offline = true
			}
		}

	}

	return offline
}

func CheckBrainClient() error {
	clientName := new(string)
	if IsBrainOffline(clientName) {
		return fmt.Errorf("brain client is offline, please get it ready")
	}

	go monitorClientStatus(IsBrainOffline, clientName, nil)
	return nil
}

