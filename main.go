package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/scheduler"
	sa "github.com/NodeFitter/NodeFitter/scheduler/abstraction"
)

func main() {

	//EXAMPLE STARTUP CODE
	a := context.InitialContext{}
	a.ReadConfig()

	var b sa.Ischeduler = &scheduler.Scheduler{} // import sa "github.com/NodeFitter/NodeFitter/scheduler/abstraction"

	//var c ca.Icontroller = &controller.Controller{} // import ca "github.com/NodeFitter/NodeFitter/controller/abstraction"

	err := b.Start(a.SchedulerContext)

	log.Println(err)

	b.GetVMs()

	// Allow the scheduler to run until forcefully stopped
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	fmt.Println("Scheduler closed")
}
