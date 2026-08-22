package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/controller"
	ca "github.com/NodeFitter/NodeFitter/controller/abstraction"
	"github.com/NodeFitter/NodeFitter/scheduler"
	sa "github.com/NodeFitter/NodeFitter/scheduler/abstraction"
)

func main() {

	//EXAMPLE STARTUP CODE
	log.Println("[*] Reading configs...")
	a := context.InitialContext{}
	a.ReadConfig()

	log.Println("[*] Initializing scheduler")

	var b sa.Ischeduler = &scheduler.Scheduler{} // import sa "github.com/NodeFitter/NodeFitter/scheduler/abstraction"

	log.Println("[*] Starting scheduler...")
	err := b.Start(a.SchedulerContext)
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}

	var c ca.Icontroller = controller.NewController(a.ControllerContext, b)

	log.Println("[*] Starting RPC server...")
	socket := os.Getenv("NODEFITTER_SOCKET")
	if socket == "" {
		socket = "/run/nodefitter/NodeFitter.sock"
	}

	if err := controller.Serve(
		// a.ControllerContext.Socket,
		socket,
		c,
	); err != nil {
		log.Fatal(err)
	}

	log.Println("[*] Ready")

	// Allow the scheduler to run until forcefully stopped
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	fmt.Println("Scheduler closed")
}
