package main

import (
	"flag"
	"fmt"
	"rabbitmq/modified/server"
)

func main() {

	// configure/read flags
	var isAdaptivePtr = flag.Bool("is-adaptive", false, "is-adaptive is a boolean")
	var controllerTypePtr = flag.String("controller-type", "None", "controller-type is a string")
	var prefetchCountInitialPtr = flag.Int("prefetch-count-initial", 1, "prefetch-count-initial is an int")
	var monitorIntervalPtr = flag.Int("monitor-interval", 1, "monitor-interval is an int (ms)")
	var setPoint = flag.Int("set-point", 1601, "set-point is an int (goal rate)")
	var kp = flag.Int("kp", 1601, "kp is an int (constant K of PID)")
	flag.Parse()

	// create new server
	var server = server.NewServer(*isAdaptivePtr, *controllerTypePtr, *prefetchCountInitialPtr, *monitorIntervalPtr, *setPoint, *kp)

	// execute server
	fmt.Println("Server is running ...")
	server.Run()
}
