package client

import (
	"fmt"
	"github.com/streadway/amqp"
	"math/rand"
	"rabbitmq/shared"
	"strconv"
	"time"
)

type Client struct {
	Id         string
	N          int
	SampleSize int
	Mean       float64
	StdDev     float64
	Conn       *amqp.Connection
	Ch         *amqp.Channel
	Queue      amqp.Queue
	Msgs       <-chan amqp.Delivery
}

func NewClient(clientIdPtr string, fibonacciNumberPtr int, sampleSizePtr int, meanRequestTimePtr int, stdDevMeanRequestTimePtr int) Client {
	c := Client{}

	// random setup
	rand.Seed(time.Now().UTC().UnixNano())

	// configure client

	c.Id = clientIdPtr
	c.N = fibonacciNumberPtr
	c.SampleSize = sampleSizePtr
	c.Mean = float64(meanRequestTimePtr)
	c.StdDev = float64(stdDevMeanRequestTimePtr)

	// Configure rabbitmq elements
	c.configureRabbitMQ()

	return c
}

func (c Client) Run() time.Duration {

	err := error(nil)
	totalTime := time.Duration(0)

	// Close channels and connections (when finish)
	defer c.Conn.Close()
	defer c.Ch.Close()

	for i := 0; i < c.SampleSize; i++ {
		corrId := shared.RandomString(32)

		// make resquest randomly distributed
		interTime := c.Mean + rand.NormFloat64()*c.StdDev
		time.Sleep(time.Duration(interTime) * time.Millisecond)

		// remove if N is fixed
		//c.N = rand.Intn(25)  // TODO

		startTime := time.Now()

		err = c.Ch.Publish(
			"",          // exchange
			"rpc_queue", // routing key
			false,       // mandatory
			false,       // immediate

			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrId,
				ReplyTo:       c.Queue.Name,
				Body:          []byte(strconv.Itoa(c.N)),
				AppId:         c.Id, // TODO - include
			})
		shared.FailOnError(err, "Failed to publish a message")

		// Receive response
		for d := range c.Msgs {
			if corrId == d.CorrelationId {
				_, err := strconv.Atoi(string(d.Body)) // discard result
				shared.FailOnError(err, "Failed to convert body to integer")
				endTime := time.Now()
				if c.Id == "50" { // only client '50' TODO
					//fmt.Println(time.Now(),endTime.Sub(startTime))
					fmt.Println(endTime.Sub(startTime).Milliseconds())
				}
				totalTime += endTime.Sub(startTime)
				break
			}
		}
	}
	return totalTime
}

func (c *Client) configureRabbitMQ() {

	err := error(nil)

	//conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/") // local
	c.Conn, err = amqp.Dial("amqp://nsr:nsr@localhost:5672/") // Docker
	shared.FailOnError(err, "Failed to connect to RabbitMQ")
	//defer conn.Close()

	c.Ch, err = c.Conn.Channel()
	shared.FailOnError(err, "Failed to open a channel")
	//defer ch.Close()

	c.Queue, err = c.Ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	shared.FailOnError(err, "Failed to declare a queue")

	c.Msgs, err = c.Ch.Consume(
		c.Queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	shared.FailOnError(err, "Failed to register a consumer")
}
