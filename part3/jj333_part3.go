package main

import (
	"log"
	"math/rand"
	"time"
)

/*
3.b:

The only case of deadlock I can find that has been fixed in 3 but not 2,
is the case mentioned in 2.b where the doctor goes into a sleeping state
while the patients are being added to the queue.

This was fixed, by introducing a "ready" channel, to signal that the dentist
is ready to recieve patients.

The introduction of the assistant doesn't help a lot in this case, as the low
priority patients actually get pushed further down the queue, since the placement
on the queue is no longer blocked by how fast the dentist is.
If the "main" queue is given a length greater than or equal to the high priority queue,
then the low priority patients will just be placed directly after the high priority
patients, which is debatably unfair, since their actual appointment time will be much later.

*/

func printDentist(message string) {
	log.Printf("Dentist %s", message)
}

func printAssistant(message string) {
	log.Printf("Assistant %s", message)
}

func printPatient(id int, message string) {
	log.Printf("Patient %d %s", id, message)
}

func dentist(wait chan chan int, dent <-chan chan int, ready chan bool) {
	for {
		select {
		case patient := <-wait: // There is a patient in the queue
			dentistSeeingPatient(patient)
		default: // No patient in the high priority queue
			printDentist("is sleeping")
			ready <- true              // Signals the dentist is ready
			impatientPatient := <-dent // Sleeps by waiting for a "wake" on the dent channel
			printDentist("woken up by patient")
			dentistSeeingPatient(impatientPatient)
		}
	}
}

func dentistSeeingPatient(patient chan<- int) {
	patient <- 0 // Indicates the patient is being seen
	r := rand.New(rand.NewSource(99))
	randomTime := r.Intn(3)
	time.Sleep(time.Duration(randomTime) * time.Second) // Indicates how long the treatment takes
	patient <- 0                                        // The patient has been seen
}

func patient(wait chan<- chan int, dent chan<- chan int, id int) {
	printPatient(id, "wants a treatment")
	appointment := make(chan int)
	select {
	case dent <- appointment: // Wakes up the dentist if asleep
		patientBeingTreated(id, appointment)
	default:
		wait <- appointment
		printPatient(id, "is waiting")
		patientBeingTreated(id, appointment)
	}
}

func patientBeingTreated(id int, appointment <-chan int) {
	<-appointment
	printPatient(id, "is having a treatment")
	<-appointment
	printPatient(id, "has shiny teeth!")
}

func assistant(hwait chan chan int, lwait <-chan chan int, wait chan<- chan int) {
	go fairness(hwait, lwait, wait)

	for {
		select {
		case hPatient := <-hwait: // There is a patient in the high priority queue
			wait <- hPatient
		default: // No patient in the high priority queue
			select {
			case lPatient := <-lwait: // Checks for patients in the low priority queue
				wait <- lPatient
			default: // No patients in either queue
			}
		}
	}
}

func fairness(hwait chan chan int, lwait <-chan chan int, wait chan<- chan int) {
	m := 1000
	timer := time.NewTimer(time.Duration(m) * time.Millisecond)

	for {
		select {
		case <-timer.C:
			select {
			case lPatient := <-lwait:
				hwait <- lPatient
				printAssistant("moving patient from low priority to high priority")
			default:
				// There was nobody waiting in low priority
			}
			timer = time.NewTimer(time.Duration(m) * time.Millisecond) // reset timer
		default:
			// Timer is still going on
		}
	}
}

type config struct {
	highPriorityPatients  int
	lowPriorityPatients   int
	highPriorityQueueSize int
	lowPriorityQueueSize  int
	mainQueueSize         int
}

func main() {
	c := config{
		highPriorityPatients:  10,
		lowPriorityPatients:   3,
		highPriorityQueueSize: 100,
		lowPriorityQueueSize:  5,
		mainQueueSize:         5,
	}

	dent := make(chan chan int)
	hwait := make(chan chan int, c.highPriorityQueueSize)
	lwait := make(chan chan int, c.lowPriorityQueueSize)
	wait := make(chan chan int, c.mainQueueSize)
	ready := make(chan bool)

	go dentist(wait, dent, ready)
	go assistant(hwait, lwait, wait)

	<-ready // Signals the dentist is ready to recieve patients.

	for a := c.lowPriorityPatients; a < c.highPriorityPatients; a++ {
		go patient(hwait, dent, a)
	}

	for b := 0; b < c.lowPriorityPatients; b++ {
		go patient(lwait, dent, b)
	}

	time.Sleep(50 * time.Second)
}
