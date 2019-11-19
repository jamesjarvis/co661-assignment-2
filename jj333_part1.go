package main

import (
	"log"
	"math/rand"
	"time"
)

func printDentist(message string) {
	log.Printf("Dentist %s", message)
}

func printPatient(id int, message string) {
	log.Printf("Patient %d %s", id, message)
}

func dentist(wait <-chan chan int, dent <-chan chan int) {
	for {
		select {
		case nextPatient := <-wait: // There is a patient in the queue
			dentistSeeingPatient(nextPatient)
		default: // No patient in the queue
			printDentist("is sleeping")
			impatientPatient := <-dent // Sleeps by waiting for a "wake" on the dent channel
			printDentist("woken up by patient")
			dentistSeeingPatient(impatientPatient)
		}
	}
}

func dentistSeeingPatient(patient chan<- int) {
	r := rand.New(rand.NewSource(99))
	randomTime := r.Intn(6)
	patient <- 0                                        // Indicates the patient is being seen
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
	m := 10
	// n is the size of the queue to the dentist
	n := 5

	dent := make(chan chan int)    // creates a synchronous channel
	wait := make(chan chan int, n) // creates an asynchronous channel of size n

	go dentist(wait, dent)
	time.Sleep(3 * time.Second)

	for i := 0; i < m; i++ {
		go patient(wait, dent, i)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(100000 * time.Millisecond)
}
