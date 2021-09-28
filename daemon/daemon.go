package main

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/api/profile"
	"keentune/daemon/api/sensitize"
	"keentune/daemon/api/system"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"keentune/daemon/common/file"
	"net"
	"net/rpc"
	"os/signal"
	"syscall"
	"fmt"
	"os"
)

func main() {
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

	fmt.Println("KeenTune daemon running...")

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
	if !file.IsPathExist(m.GetProfileWorkPath("")) {
		os.MkdirAll(m.GetProfileWorkPath(""), os.ModePerm)
	}

	if !file.IsPathExist(m.GetTuningWorkPath("")) {
		os.MkdirAll(m.GetTuningWorkPath(""), os.ModePerm)
	}

	if !file.IsPathExist(m.GetSensitizePath()) {
		os.MkdirAll(m.GetSensitizePath(), os.ModePerm)
	}
}
