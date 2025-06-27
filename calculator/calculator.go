/*
	Это просто примитивный калькулятор. Однако, показалось интересной идея представления калькулятора в виде map, а не банальным switch-case.
*/

package main

import (
	"fmt"
)

var Calculator map[string](func(float64, float64) float64) = map[string](func(float64, float64) float64){

	"+": func(x, y float64) float64 {
		return x + y
	},

	"-": func(x, y float64) float64 {
		return x - y
	},

	"*": func(x, y float64) float64 {
		return x * y
	},

	"/": func(x, y float64) float64 {
		if y == 0 {
			fmt.Println("Делить на ноль нельзя!")
			return 0
		}
		return x / y
	},

	"%": func(x, y float64) float64 {
		if y == 0 {
			fmt.Println("Делить на ноль нельзя!")
			return 0
		}
		return float64(int(x) % int(y))
	},
}

func main() {

	var num1, num2 float64
	var sign string

	fmt.Print("Введите первое число (у дроби разделитель '.'): ")
	fmt.Scan(&num1)

	fmt.Print("Введите знак действия ('+', '-', '*', '/', '%'): ")
	fmt.Scan(&sign)

	fmt.Print("Введите второе число (у дроби разделитель '.'): ")
	fmt.Scan(&num2)

	if _, ok := Calculator[sign]; ok {
		fmt.Printf("Ответ: %v", Calculator[sign](num1, num2))
	} else {
		fmt.Println("неизвестная операция")
	}
}
