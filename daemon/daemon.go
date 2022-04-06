package main

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/api/profile"
	"keentune/daemon/api/sensitize"
	"keentune/daemon/api/system"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.Init()
	log.Init()
	
	m.StopSig = make(chan os.Signal, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go com.KeenTunedService(quit)
	rpc.RegisterName("system", new(system.Service))
	rpc.RegisterName("param", new(param.Service))
	rpc.RegisterName("profile", new(profile.Service))
	rpc.RegisterName("sensitize", new(sensitize.Service))

	listener, err := net.Listen("tcp", ":9870")
	if err != nil {
		log.Errorf(log.ParamTune, "ListenTCP error:%v", err)
		os.Exit(1)
	}

	go mkWorkDir()

	showStart()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf(log.ParamTune, "Accept error:%v", err)
			continue
		}
		rpc.ServeConn(conn)
	}
}

func mkWorkDir() {
	if !file.IsPathExist(config.GetProfileWorkPath("")) {
		os.MkdirAll(config.GetProfileWorkPath(""), os.ModePerm)
	}

	if !file.IsPathExist(config.GetTuningWorkPath("")) {
		os.MkdirAll(config.GetTuningWorkPath(""), os.ModePerm)
	}

	if !file.IsPathExist(config.GetSensitizePath()) {
		os.MkdirAll(config.GetSensitizePath(), os.ModePerm)
	}

	activeConf := config.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeConf) {
		fp, _ := os.Create(activeConf)
		if fp != nil {
			fp.Close()
		}
	}
}

func showStart() {
	fmt.Printf("Keentune Home: %v\nKeentune Workspace: %v\n", config.KeenTune.Home, config.KeenTune.DumpConf.DumpHome)

	fmt.Println("In order to ensure the security of sensitive information, IP is mapped to ID")
	for _, group := range config.KeenTune.Group {
		for index, ip := range group.IPs {
			fmt.Printf("\ttarget [%v]\t<--> id: group-%v.target-%v\n", ip, group.GroupNo, index+1)
		}
	}

	if err := com.StartCheck(); err != nil {
		fmt.Printf("[%v] %v\n", utils.ColorString("YELLOW", "Warning"), err)
	}

	fmt.Println("KeenTune daemon running...")
}

