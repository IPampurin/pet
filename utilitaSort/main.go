/* ВАРИАНТ №1 - решение задачи l2.10 - Утилита sort */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Config - конфигурация сортировки
type Config struct {
	keyColumn            int    // флаг сортировки по столбцам
	numeric              bool   // флаг числовой сортировки
	reverse              bool   // флаг сортировки в обратном порядке
	unique               bool   // флаг выдачи без повторов
	month                bool   // сортировка по месяцам
	ignoreTrailingBlanks bool   // флаг игнора хвостовых пробелов
	checkSorted          bool   // флаг проверки на отсортированность
	humanNumeric         bool   // флаг сортировки по размерам
	columnSeparator      string // разделитель по умолчанию (табуляция)
}

// MonthMap служит для преобразования названия месяца в число
var MonthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

// HumanSuffixes - множители для человекочитаемых размеров
var HumanSuffixes = map[byte]float64{
	'K': 1024,
	'M': 1024 * 1024,
	'G': 1024 * 1024 * 1024,
	'T': 1024 * 1024 * 1024 * 1024,
	'P': 1024 * 1024 * 1024 * 1024 * 1024,
	'E': 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
}

// parseFlags парсит флаги строки запуска программы
func parseFlags() Config {

	config := Config{}
	flag.IntVar(&config.keyColumn, "k", 0, "sort by column N")
	flag.BoolVar(&config.numeric, "n", false, "sort numerically")
	flag.BoolVar(&config.reverse, "r", false, "sort in reverse order")
	flag.BoolVar(&config.unique, "u", false, "output only unique lines")
	flag.BoolVar(&config.month, "M", false, "sort by month")
	flag.BoolVar(&config.ignoreTrailingBlanks, "b", false, "ignore trailing blanks")
	flag.BoolVar(&config.checkSorted, "c", false, "check if sorted")
	flag.BoolVar(&config.humanNumeric, "h", false, "sort by human-readable sizes")

	flag.Parse() // парсим флаги из командной строки (Must be called after all flags are defined and before flags are accessed by the program)

	config.columnSeparator = "\t" // разделитель по умолчанию (табуляция)

	// возвращаем структуру конфигурации
	return config
}

// readLines считывает ввод по строкам
func readLines(r io.Reader) []string {

	scanner := bufio.NewScanner(r)
	// увеличиваем буфер для чтения больших файлов (по умолчанию )
	const maxCapacity = 1024 * 1024 * 1024 // 1GB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// считываем строки и добавляем в слайс
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// обрабатываем ошибку сканера
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка считывания: %v\n", err)
		os.Exit(1)
	}

	// возвращаем слайс считанных строк
	return lines
}

// getColumn извлекает указанный столбец из строки
func getColumn(line string, column int, ignoreBlanks bool) string {

	columns := strings.Split(line, "\t") // разделитель столбцов - табуляция
	if column <= 0 || column > len(columns) {
		return "" // если столбец не существует, возвращаем пустую строку
	}

	result := columns[column-1]
	// если ignoreTrailingBlanks установлен в конфигурации запуска (флаг -b),
	// удаляем хвостовые пробелы из значения столбца
	if ignoreBlanks {
		result = strings.TrimRightFunc(result, unicode.IsSpace)
	}

	return result
}

// parseHumanSize парсит строки, представляющие размеры данных в человекочитаемом формате,
// поддерживает суффиксы K, M, G, T, P, E для килобайт, мегабайт и т.д.
// возвращает числовое значение в байтах и флаг успешного парсинга,
// например: "1K" -> 1024, "2.5M" -> 2621440
func parseHumanSize(s string) (float64, bool) {

	// удаляем пробелы
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false // если были только пробелы, то это что-то не то
	}

	// пытаемся распарсить как обычное число
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num, true
	}

	// пытаемся распарсить число с суффиксом
	lastChar := s[len(s)-1]
	if multiplier, exists := HumanSuffixes[lastChar]; exists {
		numStr := s[:len(s)-1]
		if num, err := strconv.ParseFloat(numStr, 64); err == nil {
			return num * multiplier, true
		}
	}

	return 0, false
}

// parseMonthValue парсит строку по названию месяца,
// поддерживает английские сокращения месяцев: jan, feb, mar, apr, may, jun, jul, aug, sep, oct, nov, dec
func parseMonthValue(s string) (int, bool) {

	// убираем лишние пробелы и приводим к нижнему регистру
	s = strings.ToLower(strings.TrimSpace(s))

	if month, exists := MonthMap[s]; exists {
		return month, true // возвращаем номер месяца (1-12) и флаг успешного распознавания
	}

	return 0, false
}

// функция сравнения для сортировки
type comparer func(i, j int) bool

// createComparer создаёт функцию сравнения для сортировки строк с поддержкой различных типов данных,
// учитывает все установленные флаги конфигурации и возвращает компаратор для sort.Slice.
// приоритет проверки типов: человекочитаемые размеры -> месяцы -> числовая сортировка -> строковое сравнение.
// если указана колонка, сравнение происходит только по значениям в этой колонке
func createComparer(lines []string, config Config) comparer {

	return func(i, j int) bool {
		a, b := lines[i], lines[j]

		// если указана колонка, берем только её
		if config.keyColumn > 0 {
			a = getColumn(a, config.keyColumn, config.ignoreTrailingBlanks)
			b = getColumn(b, config.keyColumn, config.ignoreTrailingBlanks)
		}

		// игнорируем хвостовые пробелы если нужно
		if config.ignoreTrailingBlanks {
			a = strings.TrimRightFunc(a, unicode.IsSpace)
			b = strings.TrimRightFunc(b, unicode.IsSpace)
		}

		// человекочитаемые размеры
		if config.humanNumeric {
			numA, okA := parseHumanSize(a)
			numB, okB := parseHumanSize(b)
			if okA && okB {
				return numA < numB
			}
			// если не удалось распарсить, переходим к следующему типу сравнения
		}

		// сортировка по месяцам
		if config.month {
			monthA, okA := parseMonthValue(a)
			monthB, okB := parseMonthValue(b)
			if okA && okB {
				return monthA < monthB
			}
			// если не удалось распарсить, переходим к следующему типу сравнения
		}

		// числовая сортировка
		if config.numeric {
			numA, errA := strconv.ParseFloat(a, 64)
			numB, errB := strconv.ParseFloat(b, 64)
			if errA == nil && errB == nil {
				return numA < numB
			}
			// если не удалось распарсить, переходим к следующему типу сравнения
		}

		// строковое сравнение
		return a < b
	}
}

// processSimple обрабатывает данные без дополнительных флагов (базовая сортировка).
// выполняет сортировку строк с учетом конфигурации (какие флаги установлены)
func processSimple(lines []string, config Config) {

	// если строка одна, сортировать нечего
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, comparer)               // сортируем данные
	}

	outputLines(lines)
}

// processReverse выполняет обратную сортировку строк (флаг -r)
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processReverse(lines []string, config Config) {

	// если строка одна, сортировать нечего
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(lines)
}

// processNumeric выполняет числовую сортировку строк (флаг -n)
// если строки не могут быть преобразованы в числа, используется обычное строковое сравнение
func processNumeric(lines []string, config Config) {

	// временно устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true

	// если строка одна, сортировать нечего
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, comparer)               // сортируем данные
	}

	outputLines(lines)
}

// processReverseNumeric выполняет числовую сортировку строк в обратном порядке (флаг -rn)
// инвертирует результат сравнения для получения обратного порядка сортировки.
// если строки не могут быть преобразованы в числа, используется обычное строковое сравнение
func processReverseNumeric(lines []string, config Config) {

	// временно устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true

	// если строка одна, сортировать нечего
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(lines)
}

// processUnique удаляет дубликаты и сортирует строку (флаг -u)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации
func processUnique(lines []string, config Config) {

	seen := make(map[string]struct{}) // мапа уникальности
	var uniqueLines []string          // слайс уникальных строк

	for _, line := range lines {
		key := line
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если значение не встречалось, добавляем его в мапу и в слайс
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если уникальная строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, comparer)               // сортируем данные
	}

	outputLines(uniqueLines)
}

// processUniqueReverse удаляет дубликаты и сортирует строку в обратном порядке (флаг -ur)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processUniqueReverse(lines []string, config Config) {

	seen := make(map[string]struct{}) // мапа уникальности
	var uniqueLines []string          // слайс уникальных строк

	for _, line := range lines {
		key := line
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если значение не встречалось, добавляем его в мапу и в слайс
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если уникальная строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(uniqueLines)
}

// processUniqueNumeric удаляет дубликаты и сортирует строки как числа (флаг -un)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации
// если строки не могут быть преобразованы в числа, используется обычное строковое сравнение
func processUniqueNumeric(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true

	seen := make(map[string]struct{}) // мапа уникальности
	var uniqueLines []string          // слайс уникальных строк

	for _, line := range lines {
		key := line
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если значение не встречалось, добавляем его в мапу и в слайс
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если уникальная строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, comparer)               // сортируем данные
	}

	outputLines(uniqueLines)
}

// processUniqueReverseNumeric удаляет дубликаты и сортирует строки как числа в обратном порядке (флаг -unr)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации
// если строки не могут быть преобразованы в числа, используется обычное строковое сравнение
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processUniqueReverseNumeric(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true
	seen := make(map[string]struct{}) // мапа уникальности
	var uniqueLines []string          // слайс уникальных строк

	for _, line := range lines {
		key := line
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если значение не встречалось, добавляем его в мапу и в слайс
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если уникальная строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}
	outputLines(uniqueLines)
}

// функции для работы с колонками

// processKey выполняет сортировку строк по указанной колонке (флаг -k)
// использует компаратор, который учитывает все установленные флаги конфигурации
// если указана колонка, сравнение происходит только по значениям в этой колонке
// поддерживает числовую сортировку, сортировку по месяцам и другие типы сравнения
func processKey(lines []string, config Config) {

	// если строка одна, пропускаем сортировку
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, comparer)               // сортируем данные
	}

	outputLines(lines)
}

// processReverseKey выполняет сортировку строк по указанной колонке в обратном порядке (флаг -rk)
// использует компаратор, который учитывает все установленные флаги конфигурации
// если указана колонка, сравнение происходит только по значениям в этой колонке
// поддерживает числовую сортировку, сортировку по месяцам и другие типы сравнения
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processReverseKey(lines []string, config Config) {

	// если строка одна, пропускаем сортировку
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(lines)
}

// processNumericKey выполняет числовую сортировку строк по указанной колонке (флаг -kn)
// если указана колонка, сравнение происходит только по значениям в этой колонке
// если строки в колонке не могут быть преобразованы в числа, используется обычное строковое сравнение
func processNumericKey(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true

	// если строка одна, пропускаем сортировку
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, comparer)               // сортируем данные
	}

	outputLines(lines)
}

// processReverseNumericKey выполняет числовую сортировку строк по указанной колонке в обратном порядке (флаг -rkn)
// если указана колонка, сравнение происходит только по значениям в этой колонке
// если строки в колонке не могут быть преобразованы в числа, используется обычное строковое сравнение
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processReverseNumericKey(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true

	// если строка одна, пропускаем сортировку
	if len(lines) > 1 {
		comparer := createComparer(lines, config) // создаём компаратор
		sort.Slice(lines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(lines)
}

// processUniqueKey удаляет дубликаты по указанной колонке и сортирует строки (флаг -ku)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации (включая тип сортировки и колонку)
func processUniqueKey(lines []string, config Config) {

	// мапа для отслеживания уникальных значений в указанной колонке (или всей строки, если колонка не указана)
	seen := make(map[string]struct{})
	var uniqueLines []string

	for _, line := range lines {
		key := line
		// проверяем, что номер столбца положителен
		if config.keyColumn > 0 {
			key = getColumn(line, config.keyColumn, config.ignoreTrailingBlanks)
		}
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если ключа не встречали, кладём его в мапу и в добавляем в результат
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, comparer)               // сортируем данные
	}

	outputLines(uniqueLines)
}

// processUniqueReverseKey удаляет дубликаты по указанной колонке и сортирует строки в обратном порядке (флаг -rku)
// учитывает флаг ignoreTrailingBlanks при сравнении строк на уникальность
// после фильтрации дубликатов выполняет сортировку с учетом конфигурации (включая тип сортировки и колонку)
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processUniqueReverseKey(lines []string, config Config) {

	// мапа для отслеживания уникальных значений в указанной колонке (или всей строки, если колонка не указана)
	seen := make(map[string]struct{})
	var uniqueLines []string // слайс с результатом

	for _, line := range lines {
		key := line
		// проверяем, что номер столбца положителен
		if config.keyColumn > 0 {
			key = getColumn(line, config.keyColumn, config.ignoreTrailingBlanks)
		}
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если ключа не встречали, кладём его в мапу и в добавляем в результат
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(uniqueLines)
}

// processUniqueNumericKey удаляет дубликаты по указанной колонке и выполняет числовую сортировку (флаги -kun)
// после фильтрации дубликатов выполняет числовую сортировку с учетом конфигурации
// если строки в колонке не могут быть преобразованы в числа, используется обычное строковое сравнение
func processUniqueNumericKey(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true
	// мапа для отслеживания уникальных значений в указанной колонке (или всей строки, если колонка не указана)
	seen := make(map[string]struct{})
	var uniqueLines []string // слайс с результатом

	for _, line := range lines {
		key := line
		// проверяем, что номер столбца положителен
		if config.keyColumn > 0 {
			key = getColumn(line, config.keyColumn, config.ignoreTrailingBlanks)
		}
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если ключа не встречали, кладём его в мапу и в добавляем в результат
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, comparer)               // сортируем данные
	}

	outputLines(uniqueLines)
}

// processUniqueReverseNumericKey удаляет дубликаты по указанной колонке и выполняет числовую сортировку в обратном порядке (флаги -kurn)
// после фильтрации дубликатов выполняет числовую сортировку с учетом конфигурации
// если строки в колонке не могут быть преобразованы в числа, используется обычное строковое сравнение
// использует тот же компаратор что и обычная сортировка, но инвертирует результат сравнения
func processUniqueReverseNumericKey(lines []string, config Config) {

	// устанавливаем флаг numeric в true для создания компаратора с числовым сравнением
	config.numeric = true
	// мапа для отслеживания уникальных значений в указанной колонке (или всей строки, если колонка не указана)
	seen := make(map[string]struct{})
	var uniqueLines []string // слайс с результатом

	for _, line := range lines {
		key := line
		// проверяем, что номер столбца положителен
		if config.keyColumn > 0 {
			key = getColumn(line, config.keyColumn, config.ignoreTrailingBlanks)
		}
		// учитываем флаг ignoreTrailingBlanks
		if config.ignoreTrailingBlanks {
			key = strings.TrimRightFunc(key, unicode.IsSpace)
		}
		// если ключа не встречали, кладём его в мапу и в добавляем в результат
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			uniqueLines = append(uniqueLines, line)
		}
	}

	// если строка одна, пропускаем сортировку
	if len(uniqueLines) > 1 {
		comparer := createComparer(uniqueLines, config) // создаём компаратор
		sort.Slice(uniqueLines, func(i, j int) bool {   // сортируем данные
			return !comparer(i, j)
		})
	}

	outputLines(uniqueLines)
}

// isSorted проверяет, отсортированы ли строки в соответствии с заданной конфигурацией
// используется при установленном флаге -c для проверки отсортированности данных
// использует тот же компаратор, что и для сортировки
func isSorted(lines []string, config Config) bool {

	// если строка одна, то сортировать нечего
	if len(lines) <= 1 {
		return true
	}

	comparer := createComparer(lines, config) // создаём компаратор

	for i := 0; i < len(lines)-1; i++ {
		if !comparer(i, i+1) {
			return false // возвращаем false если все строки не отсортированы
		}
	}

	// возвращаем true если все строки отсортированы
	return true
}

// outputLines выводит данные в консоль
func outputLines(lines []string) {

	for _, line := range lines {
		fmt.Println(line)
	}
}

func main() {

	config := parseFlags() // парсим флаги запуска

	var input io.Reader // объявляем ридер

	if fileName := flag.Arg(0); fileName != "" {
		// если при запуске программы указано имя файла, то читаем из него
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка открытия файла: %v\n", err)
			os.Exit(1)
		}
		defer file.Close() // обеспечиваем закрытие файла
		input = file
	} else {
		// если при запуске программы имя файла не указано, то читаем из консоли
		input = os.Stdin
	}

	// считываем данные
	lines := readLines(input)

	// проверяем отсортированы ли данные, если установлен флаг -c
	if config.checkSorted {
		if isSorted(lines, config) {
			// если данные отсортированы - просто выходим
			os.Exit(0)
		} else {
			// если данные не отсортированы - выводим сообщение и выходим
			// со статусом 1 для указание на наличие не успешного состояния
			fmt.Fprintf(os.Stderr, "данные не отсортированы\n")
			os.Exit(1)
		}
	}

	// выбираем вариант обработки в зависимости от комбинации флагов
	// порядок от самых специфичных комбинаций к самым простым
	switch {
	case config.unique && config.reverse && config.numeric && config.keyColumn > 0:
		processUniqueReverseNumericKey(lines, config) // уникальность + обратная сортировка + числовая + сортировка по колонке
	case config.unique && config.numeric && config.keyColumn > 0:
		processUniqueNumericKey(lines, config) // уникальность + числовая + сортировка по колонке
	case config.unique && config.reverse && config.keyColumn > 0:
		processUniqueReverseKey(lines, config) // уникальность + обратная сортировка + сортировка по колонке
	case config.unique && config.keyColumn > 0:
		processUniqueKey(lines, config) // уникальность + сортировка по колонке
	case config.reverse && config.numeric && config.keyColumn > 0:
		processReverseNumericKey(lines, config) // обратная сортировка + числовая + сортировка по колонке
	case config.numeric && config.keyColumn > 0:
		processNumericKey(lines, config) // числовая + сортировка по колонке
	case config.reverse && config.keyColumn > 0:
		processReverseKey(lines, config) // обратная сортировка + сортировка по колонке
	case config.keyColumn > 0:
		processKey(lines, config) // только сортировка по колонке
	case config.unique && config.reverse && config.numeric:
		processUniqueReverseNumeric(lines, config) // уникальность + обратная сортировка + числовая
	case config.unique && config.numeric:
		processUniqueNumeric(lines, config) // уникальность + числовая
	case config.unique && config.reverse:
		processUniqueReverse(lines, config) // уникальность + обратная сортировка
	case config.unique:
		processUnique(lines, config) // только уникальность
	case config.reverse && config.numeric:
		processReverseNumeric(lines, config) // обратная сортировка + числовая
	case config.numeric:
		processNumeric(lines, config) // только числовая сортировка
	case config.reverse:
		processReverse(lines, config) // только обратная сортировка
	default:
		processSimple(lines, config) // простая сортировка (без дополнительных флагов) - случай по умолчанию
	}
}
