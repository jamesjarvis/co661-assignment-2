package main

import (
	"log"
	"math/rand"
	"time"
)

/*
2.b:

The only case of starvation I can find, is in the case of all the patients
being initialised in the time period between when the dentist checks
for patients in the queue, and going into it's sleeping state.
If this happens, then the patients will first check if the dentist is sleeping,
find that the dentist is not, and then move into the queue.
But, the dentist would have checked the queue, found nobody in it, and then
gone to sleep.

In this case, the situation will remain in starvation, until another patient
enters, and triggers the dentist to wake up.

*/

func printDentist(message string) {
	log.Printf("Dentist %s", message)
}

func printPatient(id int, message string) {
	log.Printf("Patient %d %s", id, message)
}

func dentist(hwait chan chan int, lwait <-chan chan int, dent <-chan chan int) {
	m := 10000
	timer := time.NewTimer(time.Duration(m) * time.Millisecond)
	for {
		select {
		case <-timer.C:
			select {
			case lPatient := <-lwait:
				hwait <- lPatient
				printDentist("moving patient from low priority to high priority")
			default:
				// There was nobody waiting in low priority
			}
			timer = time.NewTimer(time.Duration(m) * time.Millisecond) // reset timer
		default:
			// Timer is still going on
		}

		select {
		case hPatient := <-hwait: // There is a patient in the high priority queue
			dentistSeeingPatient(hPatient)
		default: // No patient in the high priority queue
			select {
			case lPatient := <-lwait: // Checks for patients in the low priority queue
				printDentist("taking care of low priority...")
				dentistSeeingPatient(lPatient)
			default: // No patients in either queue
				printDentist("is sleeping")
				impatientPatient := <-dent // Sleeps by waiting for a "wake" on the dent channel
				printDentist("woken up by patient")
				dentistSeeingPatient(impatientPatient)
			}
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

func main() {
	// m is the number of patients
	high := 10
	// n is the size of the queue to the dentist
	low := 3

	dent := make(chan chan int)
	hwait := make(chan chan int, 100)
	lwait := make(chan chan int, 5)

	go dentist(hwait, lwait, dent)

	for a := low; a < high; a++ {
		go patient(hwait, dent, a)
	}

	for b := 0; b < low; b++ {
		go patient(lwait, dent, b)
	}

	time.Sleep(50 * time.Second)
}
