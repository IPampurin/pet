/*
		Сортировка набора строк

	Данная программа представляет собой сортирвщик слайса строк:
	- в лексикографическом порядке
	- по длине строк
	- по первому символу
	- по последнему символу
	- по количеству цифр.
	Программа служит иллюстрацией передачи функции компаратора в функцию сортировщик.
*/

package main

import (
	"fmt"
	"unicode"
)

const (
	Lexical       = "lexical"
	Length        = "length"
	FirstChar     = "firstChar"
	LastChar      = "lastChar"
	NumbersAmount = "numbersAmount"
)

// Comparators ключевая мапа выбора вариантов сортировки
var Comparators = map[string]func(string, string) bool{
	// Сравнение строк в лексикографическом порядке
	Lexical: func(firstWord, secondWord string) bool {
		if firstWord > secondWord {
			return true
		}
		return false
	},
	// Сравнение строк по длине
	Length: func(firstWord, secondWord string) bool {
		if len(firstWord) > len(secondWord) {
			return true
		}
		return false
	},
	// Сравнение строк по первому символу
	FirstChar: func(firstWord, secondWord string) bool {
		if firstWord[0] > secondWord[0] {
			return true
		}
		return false
	},
	// Сравнение строк по последнему символу
	LastChar: func(firstWord, secondWord string) bool {
		if firstWord[len(firstWord)-1] > secondWord[len(secondWord)-1] {
			return true
		}
		return false
	},
	// Сравнение строк по количеству цифр
	NumbersAmount: func(firstWord, secondWord string) bool {
		if digitalCounter(firstWord) > digitalCounter(secondWord) {
			return true
		}
		return false
	},
}

// bubbleSort выполняет сортировку "пузырьком"
func bubbleSort(slice []string, comparator func(string, string) bool) []string {

	for i := 0; i < len(slice); i++ {
		for j := 0; j < len(slice)-1-i; j++ {
			if comparator(slice[j], slice[j+1]) {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
	return slice
}

// digitalCounter осуществляет подсчёт цифр в строке
func digitalCounter(word string) int {

	count := 0
	for _, v := range word {
		if unicode.IsDigit(v) {
			count++
		}
	}
	return count
}

func main() {

	// Считывание данных
	var n int
	fmt.Scan(&n)

	strings := make([]string, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&strings[i])
	}

	var operation string
	fmt.Scan(&operation)

	if comparator, exists := Comparators[operation]; exists {
		sortedStrings := bubbleSort(strings, comparator)
		for _, str := range sortedStrings {
			fmt.Print(str, " ")
		}
	}
}
