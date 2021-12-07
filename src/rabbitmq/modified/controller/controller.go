package controller

import (
	"fmt"
	"os"
	"rabbitmq/modified/monitor"
)

//const KP = 0.01
//const KI = 0.01
//const KD = 0.01

const MAX_PC_VALUE = 10000

type Controller struct {
	Type            string
	PC              int
	SP              int // set point
	KP              int
	KI              int
	KD              int
	F               func(int, int, ...float64) int
	Monitor         monitor.Monitor
	ProcRate        float64
	OldProcRate     float64
	DeliveryRate    float64
	OldDeliveryRate float64
	Counter         int
}

func NewController(ct string, m monitor.Monitor, pc int, sp int, kp int) Controller {
	r := Controller{}

	switch ct {
	case "Constant":
		r.F = r.Constant
	case "Delivery":
		r.F = r.DeliveryController
	case "Increasing":
		r.F = r.IncreasingController
	case "Alternating":
		r.F = r.AlternateController
	case "ProcRate":
		r.F = r.ProcRateController
	case "PID":
		r.F = r.PIDController
	case "P":
		r.F = r.PController
		r.KP = kp
	default:
		fmt.Println("Fatal Error:: Controller type does not exist")
		os.Exit(0)
	}

	r.Type = ct
	r.Monitor = m
	r.PC = pc
	r.SP = sp

	return r
}

func (c *Controller) ProcRateController(pc int, sp int, procRates ...float64) int {
	r := pc

	if procRates[0] > procRates[1] {
		r = pc * 2
		if r > MAX_PC_VALUE {
			r = MAX_PC_VALUE
		}
	} else {
		r = pc / 2
		if r == 0 {
			r = 1
		}
	}
	return r
}

func (Controller) ProcRateControllerOld(pc int, SP float64, KP float64, procRates []float64) int {
	r := pc
	last := len(procRates)

	if last <= 1 {
		return pc
	} else {
		e := procRates[last-1] - procRates[last-2]
		r = r + int(e*1.0)
	}

	if r <= 0 {
		r = 1
	}

	return r
}

func (c *Controller) PIDController(pc int, SP int, rates ...float64) int {
	// rates[0] is 'procRate'
	err := SP - int(rates[0])

	//return int(KP * err)
	return err
}

func (c *Controller) PController(pc int, SP int, rates ...float64) int {
	// rates[0] is 'procRate'
	err := SP - int(rates[0])

	return c.KP * err
}

func (c Controller) DeliveryController(pc int, sp int, rates ...float64) int {
	newPC := c.Monitor.GetDeliveryRate() * 2

	return int(newPC)
}

func (Controller) IncreasingController(pc int, sp int, rates ...float64) int {
	//r := int(math.Round(float64(pc) * 1.50))  // increasing 50%

	r := pc + 5 // linear increasing

	if r > MAX_PC_VALUE {
		r = MAX_PC_VALUE
	}
	return r
}

func (Controller) DoubleRateController(pc int, SP float64, KP float64, procRate float64) int {
	return int(2 * procRate)
}

func (Controller) Constant(pc int, sp int, rates ...float64) int {
	return pc
}

func (Controller) AlternateController(pc int, sp int, rates ...float64) int {

	// '0' is infinite - default
	switch pc {
	case 0:
		return 1
	case 1:
		return 0
	default:
		fmt.Println("Fatal error:: Controller 'Alternating' has something wrong")
		os.Exit(0)
	}
	return 0
}
