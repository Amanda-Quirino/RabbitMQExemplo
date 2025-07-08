package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue",
		false,
		true,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	word, err := getRandomWord("words.txt")

	if err != nil {
		fmt.Println("Error reading word file")
		return
	}
	attemps := 6
	currentWordState := initializeCurrentWordState(word)
	var userInput string

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for d := range msgs {
			n := string(d.Body)

			log.Printf(" [.] opt %s", n)
			response := validarOpcao(n, string(d.CorrelationId), string(d.CorrelationId), currentWordState)
			userInput = n

			err = ch.PublishWithContext(ctx,
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			failOnError(err, "Failed to publish a message")

			d.Ack(false)
		}
	}()

	correctGuess := updateGuessd(word, currentWordState, userInput)

	if !correctGuess {
		attemps--
	}

	<-forever

	// if isWordGuessed(currentWordState, word) {
	// 	fmt.Println("Congratulations! You won!")
	// 	return
	// } else if attemps == 0 {
	// 	fmt.Println("Game Over! The word is", word)
	// 	return
	// }
}

func validarOpcao(opcao string, corrId string, turn string, guessed []string) string {
	switch opcao {
	case "1":
		fmt.Println("Sala criada! O código da sala é:", 1234)
	case "2":
		fmt.Print("Digite o código da sala:")
	case "3":
		if turn == corrId {
			fmt.Println("Qual o seu palpite ?")
		} else {
			fmt.Println("Ainda não é a sua vez")
		}
	case "4":
		fmt.Println("Qual palavra ?")
	case "5":
		fmt.Println("A palavra atual é", strings.Join(guessed, ""))
	case "6":
		fmt.Println("Jogador", corrId, "saiu do jogo")
	default:
		fmt.Println("Opção inválida")
	}
	return "oi"
}

func updateGuessd(word string, guessed []string, letter string) bool {
	correctGuess := false

	for i, char := range word {
		if string(char) == letter {
			guessed[i] = letter
			correctGuess = true
		}
	}

	return correctGuess
}

func getRandomWord(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	words := strings.Split(string(data), "\n")
	return words[rand.Intn(len(words))], nil
}

func initializeCurrentWordState(word string) []string {
	currentWordState := make([]string, len(word))

	for i := range currentWordState {
		currentWordState[i] = "_"
	}

	return currentWordState
}
