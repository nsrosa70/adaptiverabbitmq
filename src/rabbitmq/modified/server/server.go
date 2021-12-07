package server

import (
	"fmt"
	"github.com/streadway/amqp"
	"rabbitmq/modified/controller"
	"rabbitmq/modified/impl"
	"rabbitmq/modified/monitor"
	"rabbitmq/shared"
	"strconv"
	"time"
)

type Server struct {
	IsAdaptive bool
	//Cnt             controller.Controller
	MonitorInterval time.Duration
	ConnSub         *amqp.Connection
	ConnPub         *amqp.Connection
	ChSub           *amqp.Channel
	ChPub           *amqp.Channel
	Queue           amqp.Queue
	Msgs            <-chan amqp.Delivery
	ChStart         chan bool
	ChStop          chan bool
	Mnt             monitor.Monitor
	Ctler           controller.Controller
}

func NewServer(isAdaptive bool, controllerType string, prefetchCountInitial int, monitorInterval int, setPoint int, kp int) Server {
	s := Server{}

	// Configure server
	s.IsAdaptive = isAdaptive
	s.MonitorInterval = time.Duration(monitorInterval) * time.Millisecond

	// Initialise channel to communicate with Monitor
	s.ChStart = make(chan bool)
	s.ChStop = make(chan bool)

	// create Monitor
	s.Mnt = monitor.NewMonitor(s.MonitorInterval)

	// create controller
	s.Ctler = controller.NewController(controllerType, s.Mnt, prefetchCountInitial, setPoint, kp)
	s.MonitorInterval = time.Duration(monitorInterval) * time.Millisecond

	// Configure RabbitMQ
	s.configureRabbitMQ()

	return s
}

// Run server
func (s Server) Run() {

	// close all rabbitmq elements before exit
	defer s.ConnSub.Close()
	defer s.ConnPub.Close()
	defer s.ChSub.Close()
	defer s.ChPub.Close()

	// start monitor
	go s.Mnt.Monitoring(s.ChStart, s.ChStop)

	// handle requests
	s.handlRequests()
}

// Handle requests
func (s Server) handlRequests() {
	forever := make(chan bool)

	go func(chStart, chStop chan bool) {
		count := 0

		for {
		myLoop:
			for d := range s.Msgs {

				// send ack to broker as soon the message has been received
				d.Ack(false)

				// unmarshall message
				N, err := strconv.Atoi(string(d.Body))
				shared.FailOnError(err, "Failed to convert body to time")

				// invoke fibonacci
				response := impl.Fib(N)

				// publish response
				err = s.ChPub.Publish(
					"",        // exchange
					d.ReplyTo, // routing key
					false,     // mandatory
					false,     // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: d.CorrelationId,
						Body:          []byte(strconv.Itoa(response)),
					})
				shared.FailOnError(err, "Failed to publish a message")

				// interact with Monitor/Controller

				select {
				case <-chStart: // start monitor
					count = 0
					s.Ctler.OldDeliveryRate = s.Ctler.DeliveryRate
					s.Ctler.OldProcRate = s.Ctler.ProcRate
				case <-chStop: // stop monitor
					s.Ctler.ProcRate = float64(count) / float64(s.Mnt.MonitorInterval.Seconds())
					fmt.Println(s.Ctler.ProcRate)

					// Reconfigure QoS (Ineffective if autoAck is true)
					if s.IsAdaptive {
						s.Ctler.PC = s.Ctler.F(s.Ctler.PC, s.Ctler.SP, s.Ctler.ProcRate, s.Ctler.OldProcRate, s.Ctler.DeliveryRate, s.Ctler.OldDeliveryRate)
						err := s.ChSub.Qos(
							s.Ctler.PC, // prefetch count
							0,          // prefetch size
							true,       // global TODO default is false
						)
						shared.FailOnError(err, "Failed to set QoS")
					}
					break myLoop
				default: // normal processing
					count++
				}
			}
		}
	}(s.ChStart, s.ChStop)
	<-forever
}

// Configure client RabbitMQ (server-side)
func (s *Server) configureRabbitMQ() {
	err := error(nil)

	///connPub, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	s.ConnPub, err = amqp.Dial("amqp://nsr:nsr@localhost:5672/") // Docker 'some-rabbit'
	shared.FailOnError(err, "Failed to connect to RabbitMQ - Publisher")

	//connSub, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	s.ConnSub, err = amqp.Dial("amqp://nsr:nsr@localhost:5672/") // Docker 'some-rabbit'
	shared.FailOnError(err, "Failed to connect to RabbitMQ - Subscriber")
	//defer conn.Close()

	s.ChPub, err = s.ConnPub.Channel()
	shared.FailOnError(err, "Failed to open a channel")
	s.ChSub, err = s.ConnSub.Channel()
	shared.FailOnError(err, "Failed to open a channel")
	//defer ch.Close()

	q, err := s.ChPub.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	shared.FailOnError(err, "Failed to declare a queue")

	if s.Ctler.PC != 0 { // start having a infinite prefetch buffer
		err = s.ChSub.Qos( // TODO - check if it is ok
			s.Ctler.PC, // prefetch count
			0,          // prefetch size
			true,       // global - default false
		)
		shared.FailOnError(err, "Failed to set QoS")
	}

	s.Msgs, err = s.ChSub.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	shared.FailOnError(err, "Failed to register a consumer")

	// configure initial QoS
	err = s.ChSub.Qos(
		s.Ctler.PC, // prefetch count
		0,          // prefetch size
		true,       // global TODO default is false
	)
	shared.FailOnError(err, "Failed to set QoS")
	return
}
