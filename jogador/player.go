package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	win := false
	scanner := bufio.NewScanner(os.Stdin)

	var res string

	// guessedLetters := make(map[string]bool)

	fmt.Println("Welcome to Hangman!")
	corrId := randomString(rand.Intn(10))
	fmt.Println("Your id is", corrId)

	fmt.Println("Let's begin")

	// Conectando com servidor
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for !win {
		op := menu(*scanner)

		err = ch.PublishWithContext(ctx,
			"",          // exchange
			"rpc_queue", // routing key
			false,       // mandatory
			false,       // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrId,
				ReplyTo:       q.Name,
				Body:          []byte(op),
			})
		failOnError(err, "Failed to publish a message")

		for d := range msgs {
			if corrId == d.CorrelationId {
				res = string(d.Body)
				break
			}
		}

		fmt.Println("Resposta:", res)

	}
}

func menu(scanner bufio.Scanner) string {
	fmt.Println("\n------ MENU ------")
	fmt.Println("1. Criar jogo com amigos")
	fmt.Println("2. Entrar em jogo")
	fmt.Println("3. Palpitar letra")
	fmt.Println("4. Palpitar palavra")
	fmt.Println("5. Obter estado do jogo") //Quando o rabbitMQ for implementado isso vai ser desnecessario?
	fmt.Println("0. Sair")
	fmt.Print("Escolha uma opção: ")

	op := getUserInput(&scanner)
	return op
}

func isWordGuessed(guessed []string, word string) bool {
	return strings.Join(guessed, "") == word
}

func isValidInput(userInput string) bool {
	return utf8.RuneCountInString(userInput) == 1
}

func getUserInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}

func displayCurrentState(currentWordState []string, attempts int) {
	fmt.Println("Current word state:", strings.Join(currentWordState, " "))
	fmt.Println("Attemps left:", attempts)
}
