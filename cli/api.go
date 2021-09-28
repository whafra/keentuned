package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"strings"
)

// TuneFlag tune options
type TuneFlag struct {
	Name      string
	Round     int
	BenchConf string
	ParamConf string
	Verbose   bool
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output string
	Force  bool
}

type SetFlag struct {
	Name string
}

type TrainFlag struct {
	Output string
	Data   string
	Trials int
	Force  bool
}

type DeleteFlag struct {
	Name  string
	Cmd   string
	Force bool
}

type RollbackFlag struct {
	Cmd string
}

var (
	outputTips = "\n\toutput: When a duplicate name appears, the original file is overwritten, y(yes) or n(no)?\n"
	deleteTips = "\n\tdelete: Be careful! It cannot be restored after deletion, y(yes) or n(no)?\n"
)

func remoteImpl(callName string, flag interface{}) error {
	client, err := rpc.Dial("tcp", "localhost:9870")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Call(callName, flag, &reply)
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(reply, "displayed in the terminal.") {
		start := strings.Index(reply, "displayed in the terminal.")
		end := strings.Index(reply, "show table end.")

		fmt.Printf("\n%v\n", reply[:start+len("displayed in the terminal.")])
		showInTable(reply[start+len("displayed in the terminal.") : end])
		fmt.Printf("\n%v", strings.TrimLeft(reply[end+len("show table end."):], "\n"))
		return nil
	}

	fmt.Printf("\n%v", reply)
	return nil
}

func RunTuneRemote(ctx context.Context, flag TuneFlag) error {
	return remoteImpl("param.Tune", flag)
}

func RunDumpRemote(ctx context.Context, flag DumpFlag) error {
	fmt.Printf("%s", outputTips)
	flag.Force = confirm()
	return remoteImpl("param.Dump", flag)
}

func RunListRemote(ctx context.Context, flag string) error {
	return remoteImpl(fmt.Sprintf("%s.List", flag), flag)
}

func RunRollbackRemote(ctx context.Context, flag RollbackFlag) error {
	return remoteImpl(fmt.Sprintf("%s.Rollback", flag.Cmd), flag)
}

func RunDeleteRemote(ctx context.Context, flag DeleteFlag) error {
	fmt.Printf("%s", deleteTips)
	if !confirm() {
		fmt.Println("process exit")
		return nil
	}

	return remoteImpl(fmt.Sprintf("%s.Delete", flag.Cmd), flag)
}

func RunInfoRemote(ctx context.Context, flag string) error {
	return remoteImpl("profile.Info", flag)
}

func RunSetRemote(ctx context.Context, flag SetFlag) error {
	return remoteImpl("profile.Set", flag)
}

func RunGenerateRemote(ctx context.Context, flag DumpFlag) error {
	fmt.Printf("%s", outputTips)
	flag.Force = confirm()
	return remoteImpl("profile.Generate", flag)
}

func RunCollectRemote(ctx context.Context, flag TuneFlag) error {
	return remoteImpl("sensitize.Collect", flag)
}

func RunTrainRemote(ctx context.Context, flag TrainFlag) error {
	fmt.Printf("%s", outputTips)
	flag.Force = confirm()
	return remoteImpl("sensitize.Train", flag)
}

func MsgRemote(tx context.Context, flag string) error {
	return remoteImpl("system.Message", flag)
}

func StopRemote(ctx context.Context, flag string) error {
	fmt.Printf("stop remote %v task of running. \n", flag)
	return remoteImpl(fmt.Sprintf("%s.Stop", flag), flag)
}
