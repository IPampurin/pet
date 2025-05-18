/*
							Конвертер валют.

	Программа предназначена для того, чтобы сумму в одной валюте представить суммой в другой валюте. Имеющийся в распоряжении перечень валют и банковский
	курс обновляются посредством связи с ресурсом в сети интернет.

Функционал.

	Настоящая программа при запуске:
	- осуществляет актуализацию курсов валют по данным ЦБ РФ через ресурс www.cbr-xml-daily.ru;
	- выводит коды и названия доступных к конвертации валют;
	- запрашивает валюту, которую необходимо конвертировать, и имеющуюся сумму;
	- запрашивает валюту, в которую требуется перевести запрошенную сумму;
	- рассчитывает и выводит эквивалентное количество валюты на основе заданного курса;
	- завершение программы осуществляется вводом команды "exit" при любом запросе.

Примечания:

	Вероятным недостатком программы является отсутствие сохранённого крайнего курса валют - при каждом запуске программы требуется новый запрос.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	dateFormat      = "02.01.2006"                          // формат даты для информирования пользователя
	outOfProgramm   = "EXIT"                                // команда для завершения программы
	programComplete = "Программа завершена."                // сообщение о завершении программы
	invalidInput    = "Некорректный ввод. Уточните данные." // сообщение о валидности ввода

)

type ValuteInfo struct {
	ID       string  `json:"ID"`
	NumCode  string  `json:"NumCode"`
	CharCode string  `json:"CharCode"`
	Nominal  int     `json:"Nominal"`
	Name     string  `json:"Name"`
	Value    float64 `json:"Value"`
	Previous float64 `json:"Previous"`
}

type AllValute map[string]ValuteInfo

type Info struct {
	Date         string    `json:"Date"`
	PreviousDate string    `json:"PreviousDate"`
	PreviousURL  string    `json:"PreviousURL"`
	Timestamp    string    `json:"Timestamp"`
	Valute       AllValute `json:"Valute"`
}

func main() {

	var info Info
	err := updatingCourses(&info)
	if err != nil {
		fmt.Println(err)
		return
	}

	appendRubel(&info)

	// Узнаем время для информирования пользователя
	date, err := time.Parse(time.RFC3339, info.Date)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Отсортируем перечень доступных валют по имени для красивого предъявления пользователю
	keys := slices.Sorted(maps.Keys(info.Valute))

	// kodName - мапа для связи имени валюты в зависимости от её кода
	kodName := make(map[string]string)

	// Проинформируем пользователя о том, какие валюты есть в списке и параллельно заполним kodName
	fmt.Printf("\nПо курсу на %s доступны для конвертации следующие валюты:\n\n", date.Format(dateFormat))
	fmt.Printf("%3s %6s  %s\n", "Код", "Имя", "Полное наименование")
	fmt.Printf("%3s %6s  %s\n", "---", "---", "-------------------")
	for _, v := range keys {
		fmt.Printf("%3s %6s  ( %s )\n", info.Valute[v].NumCode, v, info.Valute[v].Name)
		kodName[info.Valute[v].NumCode] = v
	}

	var firstValute, secondValute string
	var ok bool
	var summFirstValute float64
	for {
		// firstValute - валюта, которую необходимо перевести в другую валюту
		message := "Введите какую валюту Вы хотите обменять (имя или код)"
		if firstValute, ok = inputValute(&info, kodName, message); ok {
			return
		}

		// summFirstValute - сумма, которую необходимо перевести в другую валюту
		message = "Введите сумму валюты, которую Вы хотите обменять (разделитель дробной части '.')"
		if summFirstValute, ok = inputSumm(&info, kodName, message); ok {
			return
		}

		// secondValute - валюта, которую необходимо перевести в другую валюту
		message = "Введите валюту, на которую Вы хотите обменять первую валюту (имя или код)"
		if secondValute, ok = inputValute(&info, kodName, message); ok {
			return
		}

		// количество рублей за единицу firstValute
		rubInFirstValute := info.Valute[firstValute].Value

		// количество рублей за единицу secondValute
		rubInSecondValute := info.Valute[secondValute].Value

		out := (summFirstValute * rubInFirstValute) / rubInSecondValute

		fmt.Printf("%.2f %s (%s) = %.2f %s (%s)\n\n", summFirstValute, firstValute, info.Valute[firstValute].Name, out, secondValute, info.Valute[secondValute].Name)
	}
}

// updatingCourses обновляет информацию о курсах валют по отношению к рублю с помощью ресурса www.cbr-xml-daily.ru (курс валюты рубля к валюте, якобы от ЦБ)
func updatingCourses(nowInfo *Info) error {

	baseURL := "https://www.cbr-xml-daily.ru/daily_json.js"

	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Println("ошибка связи с сервисом обновления курсов валют:", err)
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ошибка ответа от сервиса обновления курсов валют:", err)
		return err
	}

	err = json.Unmarshal(data, nowInfo)
	if err != nil {
		fmt.Println("ошибка десериализации ответа от сервера обновления:", err)
		return err
	}

	return nil
}

// appendRubel добавляет рубль в перечень валют, чтобы через него производить конвертацию
func appendRubel(nowInfo *Info) {

	nowInfo.Valute["RUB"] = ValuteInfo{
		ID:       "",
		NumCode:  "643",
		CharCode: "RUB",
		Nominal:  1,
		Name:     "Российский рубль",
		Value:    1,
		Previous: 1,
	}
}

func inputValute(nowInfo *Info, kodName map[string]string, mes string) (string, bool) {

	var input, valute string

	for {
		fmt.Println(mes)

		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println(err)
			fmt.Println(programComplete)
			return "", true
		}

		input = strings.ToUpper(input)

		if _, ok := nowInfo.Valute[input]; ok {
			valute = input
			break
		} else if v, ok := kodName[input]; ok {
			valute = v
			break
		} else if input == outOfProgramm {
			fmt.Println(programComplete)
			return "", true
		} else {
			fmt.Println(invalidInput)
		}
	}

	return valute, false
}

func inputSumm(nowInfo *Info, kodName map[string]string, mes string) (float64, bool) {

	var input string
	var summ float64

	for {
		fmt.Println(mes)

		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println(err)
			fmt.Println(programComplete)
			return 0, true
		}

		if input == outOfProgramm {
			fmt.Println(programComplete)
			return 0, true
		} else if v, err := strconv.ParseFloat(input, 64); err == nil {
			summ = v
			if summ <= 0 {
				fmt.Println(invalidInput)
				continue
			}
			break
		} else {
			fmt.Println(invalidInput)
		}
	}

	return summ, false
}
