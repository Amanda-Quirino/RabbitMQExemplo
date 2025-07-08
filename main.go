package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {
	word, err := getRandomWord("words.txt")

	if err != nil {
		fmt.Println("Error reading word file")
		return
	}
	attemps := 6
	currentWordState := initializeCurrentWordState(word)

	scanner := bufio.NewScanner(os.Stdin)

	guessedLetters := make(map[string]bool)

	fmt.Println("Welcome to Hangman!")

	for attemps > 0 {
		displayCurrentState(currentWordState, attemps)
		userInput := getUserInput(scanner)

		if !isValidInput(userInput) {
			fmt.Println("Invalid input. Please enter a single letter.")
			continue
		}

		if guessedLetters[userInput] {
			fmt.Println("You already guessed that letter.")
			continue
		}

		guessedLetters[userInput] = true

		correctGuess := updateGuessd(word, currentWordState, userInput)

		if !correctGuess {
			attemps--
		}

		displayHangman(6 - attemps)

		if isWordGuessed(currentWordState, word) {
			fmt.Println("Congratulations! You won!")
			return
		} else if attemps == 0 {
			fmt.Println("Game Over! The word is", word)
			return
		}
	}
}

func getRandomWord(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	words := strings.Split(string(data), "\n")
	return words[rand.Intn(len(words))], nil
}

func isWordGuessed(guessed []string, word string) bool {
	return strings.Join(guessed, "") == word
}

func displayHangman(incorrectGuesses int) {
	if incorrectGuesses >= 0 && incorrectGuesses < len(hangmanStates) {
		fmt.Println(hangmanStates[incorrectGuesses])
	}
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

func isValidInput(userInput string) bool {
	return utf8.RuneCountInString(userInput) == 1
}

func getUserInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}

func initializeCurrentWordState(word string) []string {
	currentWordState := make([]string, len(word))

	for i := range currentWordState {
		currentWordState[i] = "_"
	}

	return currentWordState
}

func displayCurrentState(currentWordState []string, attempts int) {
	fmt.Println("Current word state:", strings.Join(currentWordState, " "))
	fmt.Println("Attemps left:", attempts)
}
