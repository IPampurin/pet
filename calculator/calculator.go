/*
	Это просто примитивный калькулятор.
*/

package main

import (
	"fmt"
)

func main() {

	var num1, num2, result float64
	var sign, message string

	fmt.Print("Введите первое число (у дроби разделитель '.'): ")
	fmt.Scan(&num1)

	fmt.Print("Введите знак действия ('+', '-', '*', '/'): ")
	fmt.Scan(&sign)

	fmt.Print("Введите второе число (у дроби разделитель '.'): ")
	fmt.Scan(&num2)

	switch sign {
	case "+":
		result = num1 + num2
	case "-":
		result = num1 - num2
	case "*":
		result = num1 * num2
	case "/":
		if num2 == 0 {
			message = "Делить на ноль нельзя!"
			fmt.Println(message)
			return
		} else {
			result = num1 / num2
		}
	case "%":
		result = float64(int(num1) % int(num2))
	default:
		message = "неизвестная операция"
		fmt.Println(message)
		return
	}

	fmt.Printf("Ответ: %v", result)
}
