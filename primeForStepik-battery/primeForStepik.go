/*
	Данная программа написана под влиянием задачи на образовательной платформе Stepik.
	Задача называлась "Батарейка" и сводилась к показу уровня заряда батареи, например телефона,
	в виде [    XXXXXX] в зависимости от сигналов от секций аккумулятора.
	Цель написания: наглядно показать обучающимся как устроены интерфейсы.
*/

package main

import (
	"fmt"
	"slices"
	"strings"
)

// объявляем константы для сообщений о проверках (ну захотелось проверить входные данные)
const (
	сheckForTen = "считано не 10 цифр"
	notOnly     = "во введённой строке есть не только 0 и 1"
)

// Capacity - базовая структура, для которой переопределим метод String
type Capacity struct {
	Charge  string
	Message string
}

// String - переопределённый метод вывода для объектов типа "структура Capacity"
func (c Capacity) String() string {
	if c.Message != "" {
		return fmt.Sprintf("%s", c.Message) // печать поля структуры с возможными казусами
	}
	return fmt.Sprintf("[%s]", c.Charge) // печать поля структуры по ТЗ
}

// Str - интерфейс с двумя методами для структур SliceUint, StringString, SliceString
type Str interface {
	// методы интерфейса будут работать с каждой из описаных ниже структур по-своему
	// и нам не надо будет выбирать конкретный метод
	Transform() (string, string)
	MessageOutput() Capacity
}

// SliceUint - структура интерфейсного типа Str
type SliceUint struct {
	Capacity          // базовая структура входит в SliceUint как одно из полей
	InputSlice []uint // поле структуры для записи входного []uint
}

// Transform преобразует данные из входного формата []uint в string
func (sl SliceUint) Transform() (result string, mes string) {
	// проверяем входные данные по количеству
	if len(sl.InputSlice) != 10 {
		mes = fmt.Sprint(сheckForTen)
		return "", mes
	}
	slices.Sort(sl.InputSlice)     // сортируем входной []uint
	slice := make([]string, 0, 10) // создаём новый []string, а потом записываем в него пробелы и Х в нужном количестве
	for i, v := range sl.InputSlice {
		if v == 0 {
			slice = append(slice, " ")
		} else if v == 1 {
			slice = append(slice, "X")
		} else {
			mes = fmt.Sprintf("по индексу %d введено неверное значение. %v", i, notOnly) // если среди цифр не только 1 и 0
			return "", mes
		}
	}
	result = fmt.Sprint(strings.Join(slice, "")) // объединяем заполненный []string в одну строку
	return result, ""
}

// MessageOutput определяет структуру Capacity, чтобы применить метод String
func (sl SliceUint) MessageOutput() Capacity {
	res, _ := sl.Transform() // присваиваем переменной первое выводимое методом значение
	_, mes := sl.Transform() // присваиваем переменной второе выводимое методом значение
	// заполняем поля Capacity и выдаём структуру
	return Capacity{
		Charge:  res,
		Message: mes,
	}
}

// StringString - структура интерфейсного типа Str
type StringString struct {
	Capacity           // базовая структура входит в StringString как одно из полей
	StringInput string // поле структуры для записи входной строки
}

// Transform преобразует данные из входного формата string (с проверкой) в string
func (str StringString) Transform() (result string, mes string) {
	// проверяем входные данные по количеству
	if len(str.StringInput) != 10 {
		mes = fmt.Sprint(сheckForTen)
		return "", mes
	}
	cntZero := strings.Count(str.StringInput, "0") // определяем количество 0 в строке
	cntOne := strings.Count(str.StringInput, "1")  // определяем количество 1 в строке
	// проверяем строку на посторонние символы
	if cntZero+cntOne != 10 {
		mes = fmt.Sprint(notOnly)
		return "", mes
	}
	result = fmt.Sprintf("%s%s", strings.Repeat(" ", cntZero), strings.Repeat("X", cntOne))
	return result, ""
}

// MessageOutput определяет структуру Capacity, чтобы применить метод String
func (str StringString) MessageOutput() Capacity {
	res, _ := str.Transform() // присваиваем переменной первое выводимое методом значение
	_, mes := str.Transform() // присваиваем переменной второе выводимое методом значение
	// заполняем поля Capacity и выдаём структуру
	return Capacity{
		Charge:  res,
		Message: mes,
	}
}

// SliceString - структура интерфейсного типа Str
type SliceString struct {
	Capacity                  // базовая структура входит в SliceString как одно из полей
	InputSliceString []string // поле структуры для записи входного []string
}

// Transform преобразует данные из входного формата []string в string
func (sls SliceString) Transform() (result string, mes string) {
	// проверяем входные данные по количеству
	if len(sls.InputSliceString) != 10 {
		mes = fmt.Sprint(сheckForTen)
		return "", mes
	}
	noSlice := strings.Join(sls.InputSliceString, "") // объединяем входной []string в строку
	cntZero := strings.Count(noSlice, "0")
	cntOne := strings.Count(noSlice, "1")
	if cntZero+cntOne != 10 {
		mes = fmt.Sprint(notOnly)
		return "", mes
	}
	result = fmt.Sprintf("%s%s", strings.Repeat(" ", cntZero), strings.Repeat("X", cntOne))
	return result, ""
}

// MessageOutput определяет структуру Capacity, чтобы применить метод String
func (sls SliceString) MessageOutput() Capacity {
	res, _ := sls.Transform() // присваиваем переменной первое выводимое методом значение
	_, mes := sls.Transform() // присваиваем переменной второе выводимое методом значение
	// заполняем поля Capacity и выдаём структуру
	return Capacity{
		Charge:  res,
		Message: mes,
	}
}

// самое интересное (!) - зачем нужны интерфейсы )))
// ReadAndPrint принимает переменную интерфейсного типа (в функции даём ей имя out) для интерфейса Str
func ReadAndPrint(out Str) string {
	// далее применяем метод интерфейса к переменной интерфейсного типа.
	// ничему не присваиваем и никуда не передаём значение - просто переопределяем структуры, объявленные выше.
	out.Transform()
	// вторым методом интерфейса Str определяем структуру, для которой прописали метод String и присваиваем её переменной mes
	mes := out.MessageOutput()
	// передаём переменную mes на выход (по сути это структура Capacity с заполненными
	// полями, которую метод String распечатает так, как мы захотели)
	return fmt.Sprint(mes)
}

// parseInput ищет соответствие типу для приведения и определяет переменную (здесь: структуру) для передачи в интерфейс Str
func parseInput(inp interface{}) string {
	switch v := inp.(type) {
	case []uint:
		one := SliceUint{
			Capacity: Capacity{
				Charge:  "",
				Message: "",
			},
			InputSlice: v,
		}
		return fmt.Sprint(ReadAndPrint(one))
	case string:
		two := StringString{
			Capacity: Capacity{
				Charge:  "",
				Message: "",
			},
			StringInput: v,
		}
		return fmt.Sprint(ReadAndPrint(two))
	case []string:
		three := SliceString{
			Capacity: Capacity{
				Charge:  "",
				Message: "",
			},
			InputSliceString: v,
		}
		return fmt.Sprint(ReadAndPrint(three))
	default:
		fmt.Printf("Считана переменная необрабатываемого типа - %T!\n", v)
	}
	return ""
}

func main() {

	input := []uint{0, 1, 1, 0, 0, 1, 1, 1, 0, 1}

	batteryForTest := parseInput(input)

	fmt.Println(batteryForTest)

	/*
		Варианты обрабатываемых входных данных
		inputOne := []uint{0, 1, 1, 0, 0, 1, 1, 1, 0, 1}
		inputTwo := "0110011101"
		inputThree := []string{"0", "1", "1", "0", "0", "1", "1", "1", "0", "1"}
	*/
}
