package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type Flashcard struct {
	term       string
	definition string
	mistakes   int
}

type SessionMistakes struct {
	max   int
	terms []string
}

func definitionInFlashcards(definition string, flashcards map[string]Flashcard) bool {
	for _, card := range flashcards {
		if card.definition == definition {
			return true
		}
	}
	return false
}

func logAndPrint(log *[]string, rawMessage string, additional ...interface{}) {
	var message string
	if len(additional) != 0 {
		message = fmt.Sprintf(rawMessage, additional...)
	} else {
		message = fmt.Sprint(rawMessage)
	}
	fmt.Print(message)
	*log = append(*log, message)
}

func scanAndLog(log *[]string, reader *bufio.Reader) string {
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	*log = append(*log, choice)
	return choice
}

func mistakes(flashcards map[string]Flashcard) *SessionMistakes {
	max := 0
	sessionMistakes := new(SessionMistakes)
	for _, card := range flashcards {
		if card.mistakes > max {
			max = card.mistakes
		}
	}

	sessionMistakes.max = max

	if max > 0 {
		for _, card := range flashcards {
			if card.mistakes == max {
				sessionMistakes.terms = append(sessionMistakes.terms, `"`+card.term+`"`)
			}
		}
	}

	return sessionMistakes
}

func main() {
	importFrom := flag.String("import_from", "IMPORT", "Select file to import")
	exportTo := flag.String("export_to", "EXPORT", "Select file to save to once you exit. Think of this as \"autosave\"")
	flag.Parse()

	var log []string
	flashcards := make(map[string]Flashcard)
	reader := bufio.NewReader(os.Stdin)

	if importFrom != nil {
		content, err := os.ReadFile(*importFrom)
		if err != nil {
			logAndPrint(&log, "File not found.\n")
		} else {
			lines := strings.Split(string(content), "\n")
			for i := 0; i < len(lines)-1; i += 3 {
				mistakes, _ := strconv.Atoi(lines[i+2])
				flashcards[lines[i]] = Flashcard{term: lines[i], definition: lines[i+1], mistakes: mistakes}
			}
			logAndPrint(&log, "%d cards have been loaded.\n", len(lines)/3)
		}
	}

	for {
		logAndPrint(&log, "Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats)\n")
		choice := scanAndLog(&log, reader)

		switch choice {
		case "add":
			var term string
			wasErr := false

			for {
				if !wasErr {
					logAndPrint(&log, "The card:\n")
				}
				term = scanAndLog(&log, reader)

				if _, ok := flashcards[term]; ok {
					logAndPrint(&log, "The card \"%s\" already exists. Try again:\n", term)
					wasErr = true
					continue
				} else {
					wasErr = false
					break
				}
			}

			var definition string
			wasErr = false
			for {
				if !wasErr {
					logAndPrint(&log, "The definition of the card:\n")
				}
				definition = scanAndLog(&log, reader)

				if definitionInFlashcards(definition, flashcards) {
					logAndPrint(&log, "The definition \"%s\" already exists. Try again:\n", definition)
					wasErr = true
					continue
				} else {
					wasErr = false
					break
				}
			}

			flashcards[term] = Flashcard{term, definition, 0}
			logAndPrint(&log, "The pair (\"%s\":\"%s\") has been added.\n", term, definition)

		case "remove":
			logAndPrint(&log, "Which card?\n")
			term := scanAndLog(&log, reader)

			if _, ok := flashcards[term]; ok {
				delete(flashcards, term)
				logAndPrint(&log, "The card has been removed.\n")
			} else {
				logAndPrint(&log, "Can't remove \"%s\": there is no such card.\n", term)
			}

		case "import":
			logAndPrint(&log, "File name:\n")
			fileName := scanAndLog(&log, reader)

			content, err := os.ReadFile(fileName)
			if err != nil {
				logAndPrint(&log, "File not found.\n")
			} else {
				lines := strings.Split(string(content), "\n")
				for i := 0; i < len(lines)-1; i += 3 {
					mistakes, _ := strconv.Atoi(lines[i+2])
					flashcards[lines[i]] = Flashcard{term: lines[i], definition: lines[i+1], mistakes: mistakes}
				}
				logAndPrint(&log, "%d cards have been loaded.\n", len(lines)/3)
			}

		case "export":
			logAndPrint(&log, "File name:\n")
			fileName := scanAndLog(&log, reader)
			file, err := os.Create(fileName)

			if err != nil {
				logAndPrint(&log, "Error creating file.\n")
			} else {
				for _, card := range flashcards {
					_, err := file.WriteString(card.term + "\n" + card.definition + "\n" + strconv.Itoa(card.mistakes) + "\n")
					if err != nil {
						break
					}
				}
				logAndPrint(&log, "%d cards have been saved.\n", len(flashcards))
			}
			err = file.Close()
			if err != nil {
				break
			}

		case "ask":
			if len(flashcards) == 0 {
				logAndPrint(&log, "No cards available.\n")
			} else {
				logAndPrint(&log, "How many times to ask?\n")
				times, err := strconv.Atoi(scanAndLog(&log, reader))

				if err != nil {
					break
				}

				terms := make([]string, 0, len(flashcards))
				for term := range flashcards {
					terms = append(terms, term)
				}

				for i := 0; i < times; i++ {
					randomIndex := rand.Intn(len(terms))
					randomTerm := terms[randomIndex]
					card := flashcards[randomTerm]

					logAndPrint(&log, "Print the definition of \"%s\":\n", randomTerm)

					receivedDefinition := scanAndLog(&log, reader)

					if receivedDefinition == card.definition {
						logAndPrint(&log, "Correct!\n")
					} else {
						flashcards[randomTerm] = Flashcard{
							flashcards[randomTerm].term,
							flashcards[randomTerm].definition,
							flashcards[randomTerm].mistakes + 1,
						}
						correct := false
						for _, otherCard := range flashcards {
							if receivedDefinition == otherCard.definition {
								logAndPrint(&log, "Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", card.definition, otherCard.term)
								correct = true
								break
							}
						}
						if !correct {
							//incrementMistakesForCard(randomTerm, flashcards)
							logAndPrint(&log, "Wrong. The right answer if \"%s\".\n", card.definition)
						}
					}
				}
			}

		case "log":
			dumpLog(log, reader)

		case "hardest card":
			mistakes := mistakes(flashcards)
			if len(flashcards) == 0 || mistakes.max == 0 || len(mistakes.terms) == 0 {
				logAndPrint(&log, "There are no cards with errors.\n")
			} else if len(mistakes.terms) == 1 {
				logAndPrint(&log, "The hardest card is %s. You have %s errors answering it\n", mistakes.terms[0], strconv.Itoa(mistakes.max))
			} else if len(mistakes.terms) > 1 {
				terms := strings.Join(mistakes.terms, ", ")
				logAndPrint(&log, "The hardest cards are %s\n", terms)
			}

		case "reset stats":
			resetFlashcardsMistakeStats(flashcards, log)

		case "exit":
			fmt.Println("Bye bye!")
			if exportTo != nil {
				file, err := os.Create(*exportTo)

				if err != nil {
					logAndPrint(&log, "Error creating file.\n")
				} else {
					for _, card := range flashcards {
						_, err := file.WriteString(card.term + "\n" + card.definition + "\n" + strconv.Itoa(card.mistakes) + "\n")
						if err != nil {
							break
						}
					}
					logAndPrint(&log, "%d cards have been saved.\n", len(flashcards))
				}
				err = file.Close()
				if err != nil {
					break
				}
			}
			break
		}
	}
}

func dumpLog(log []string, reader *bufio.Reader) {
	logAndPrint(&log, "File name:\n")
	fileName := scanAndLog(&log, reader)
	file, err := os.Create(fileName)

	if err != nil {
		logAndPrint(&log, "Error creating file.\n")
	} else {
		for _, line := range log {
			_, err := file.WriteString(line + "\n")
			if err != nil {
				break
			}
		}
		logAndPrint(&log, "The log has been saved.\n")
	}
	err = file.Close()
	if err != nil {
		return
	}
}

func resetFlashcardsMistakeStats(flashcards map[string]Flashcard, log []string) {
	for term := range flashcards {
		flashcards[term] = Flashcard{
			flashcards[term].term,
			flashcards[term].definition,
			0,
		}
	}
	logAndPrint(&log, "Card statistics have been reset.\n")
}
