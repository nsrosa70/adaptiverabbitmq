package main

import (
	"flag"
	"rabbitmq/modified/client"
)

func main() {

	// configure/read flags
	var clientIdPtr = flag.String("client-id", "1", "client-id is an int")
	var fibonacciNumberPtr = flag.Int("fibonacci-number", 0, "fibonacci-number is an int")
	var sampleSizePtr = flag.Int("sample-size", 1, "sample-size is an int")
	var meanRequestTimePtr = flag.Int("mean-request-time", 1, "mean-request-time is an int (ms)")
	var stdDevMeanRequestTimePtr = flag.Int("std-dev-mean-request-time", 0, "std-dev-mean-request-time is an int")
	flag.Parse()

	// create client
	c := client.NewClient(*clientIdPtr, *fibonacciNumberPtr, *sampleSizePtr, *meanRequestTimePtr, *stdDevMeanRequestTimePtr)

	// make requests to client
	totalTime := c.Run()

	// print time
	//meanTime := float64(totalTime) / 1000000.0 / float64(c.SampleSize)
	_ = float64(totalTime) / 1000000.0 / float64(c.SampleSize)
	//fmt.Printf("Mean 'response time': %.3f (ms) \n", meanTime)
	//fmt.Printf("%.3f\n", meanTime)
	//fmt.Printf("%.3f \n", meanTime)
}
