package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestParseFlags тестирует парсинг флагов командной строки
func TestParseFlags(t *testing.T) {
	// Сохраняем оригинальные аргументы и восстанавливаем после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "default flags",
			args: []string{"cmd"},
			expected: Config{
				keyColumn:       0,
				numeric:         false,
				reverse:         false,
				unique:          false,
				month:           false,
				checkSorted:     false,
				humanNumeric:    false,
				columnSeparator: "\t",
			},
		},
		{
			name: "numeric flag",
			args: []string{"cmd", "-n"},
			expected: Config{
				keyColumn:       0,
				numeric:         true,
				reverse:         false,
				unique:          false,
				month:           false,
				checkSorted:     false,
				humanNumeric:    false,
				columnSeparator: "\t",
			},
		},
		{
			name: "reverse flag",
			args: []string{"cmd", "-r"},
			expected: Config{
				keyColumn:       0,
				numeric:         false,
				reverse:         true,
				unique:          false,
				month:           false,
				checkSorted:     false,
				humanNumeric:    false,
				columnSeparator: "\t",
			},
		},
		{
			name: "key column flag",
			args: []string{"cmd", "-k", "2"},
			expected: Config{
				keyColumn:       2,
				numeric:         false,
				reverse:         false,
				unique:          false,
				month:           false,
				checkSorted:     false,
				humanNumeric:    false,
				columnSeparator: "\t",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			// Сбрасываем флаги для каждого теста
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			config := parseFlags()

			if config.keyColumn != tt.expected.keyColumn {
				t.Errorf("keyColumn: got %d, want %d", config.keyColumn, tt.expected.keyColumn)
			}
			if config.numeric != tt.expected.numeric {
				t.Errorf("numeric: got %v, want %v", config.numeric, tt.expected.numeric)
			}
			if config.reverse != tt.expected.reverse {
				t.Errorf("reverse: got %v, want %v", config.reverse, tt.expected.reverse)
			}
		})
	}
}

// TestGetColumn тестирует извлечение столбцов из строки
func TestGetColumn(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		column       int
		ignoreBlanks bool
		expected     string
	}{
		{
			name:         "basic column extraction",
			line:         "a\tb\tc",
			column:       2,
			ignoreBlanks: false,
			expected:     "b",
		},
		{
			name:         "column out of range",
			line:         "a\tb\tc",
			column:       5,
			ignoreBlanks: false,
			expected:     "",
		},
		{
			name:         "ignore trailing blanks",
			line:         "a\tb  \tc",
			column:       2,
			ignoreBlanks: true,
			expected:     "b",
		},
		{
			name:         "first column",
			line:         "first\tsecond\tthird",
			column:       1,
			ignoreBlanks: false,
			expected:     "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColumn(tt.line, tt.column, tt.ignoreBlanks)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseHumanSize тестирует парсинг человекочитаемых размеров
func TestParseHumanSize(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		success  bool
	}{
		{"1K", 1024, true},
		{"2.5M", 2.5 * 1024 * 1024, true},
		{"1G", 1024 * 1024 * 1024, true},
		{"123", 123, true},
		{"abc", 0, false},
		{"", 0, false},
		{"1.5K", 1.5 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, success := parseHumanSize(tt.input)
			if success != tt.success {
				t.Errorf("success: got %v, want %v", success, tt.success)
			}
			if success && result != tt.expected {
				t.Errorf("value: got %f, want %f", result, tt.expected)
			}
		})
	}
}

// TestParseMonthValue тестирует парсинг месяцев
func TestParseMonthValue(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		success  bool
	}{
		{"jan", 1, true},
		{"FEB", 2, true},
		{"Mar", 3, true},
		{"april", 0, false},
		{"", 0, false},
		{"dec", 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, success := parseMonthValue(tt.input)
			if success != tt.success {
				t.Errorf("success: got %v, want %v", success, tt.success)
			}
			if success && result != tt.expected {
				t.Errorf("value: got %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestProcessSimple тестирует базовую сортировку
func TestProcessSimple(t *testing.T) {
	input := []string{"banana", "apple", "cherry"}
	expected := []string{"apple", "banana", "cherry"}

	config := Config{}

	// Перехватываем вывод
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processSimple(input, config)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	got := strings.Split(output, "\n")

	if len(got) != len(expected) {
		t.Errorf("length mismatch: got %d, want %d", len(got), len(expected))
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

// TestProcessNumeric тестирует числовую сортировку
func TestProcessNumeric(t *testing.T) {
	input := []string{"10", "2", "1", "20"}
	expected := []string{"1", "2", "10", "20"}

	config := Config{}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processNumeric(input, config)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	got := strings.Split(output, "\n")

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

// TestProcessReverse тестирует обратную сортировку
func TestProcessReverse(t *testing.T) {
	input := []string{"apple", "banana", "cherry"}
	expected := []string{"cherry", "banana", "apple"}

	config := Config{}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processReverse(input, config)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	got := strings.Split(output, "\n")

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

// TestProcessUnique тестирует удаление дубликатов
func TestProcessUnique(t *testing.T) {
	input := []string{"apple", "banana", "apple", "cherry", "banana"}
	expected := []string{"apple", "banana", "cherry"}

	config := Config{}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processUnique(input, config)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	got := strings.Split(output, "\n")

	if len(got) != len(expected) {
		t.Errorf("length mismatch: got %d, want %d", len(got), len(expected))
		return
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

// TestIsSorted тестирует проверку отсортированности
func TestIsSorted(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		config   Config
		expected bool
	}{
		{
			name:     "sorted strings",
			lines:    []string{"a", "b", "c"},
			config:   Config{},
			expected: true,
		},
		{
			name:     "unsorted strings",
			lines:    []string{"c", "a", "b"},
			config:   Config{},
			expected: false,
		},
		{
			name:     "sorted numbers",
			lines:    []string{"1", "2", "10"},
			config:   Config{numeric: true},
			expected: true,
		},
		{
			name:     "single element",
			lines:    []string{"single"},
			config:   Config{},
			expected: true,
		},
		{
			name:     "empty slice",
			lines:    []string{},
			config:   Config{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSorted(tt.lines, tt.config)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCreateComparer тестирует создание компаратора
func TestCreateComparer(t *testing.T) {
	lines := []string{"banana", "apple", "cherry"}
	config := Config{}

	comparer := createComparer(lines, config)

	// Проверяем, что компаратор правильно сравнивает элементы
	if !comparer(1, 0) { // apple < banana
		t.Error("comparer should return true for apple < banana")
	}
	if comparer(0, 1) { // banana > apple
		t.Error("comparer should return false for banana > apple")
	}
}

// TestProcessKey тестирует сортировку по колонке
func TestProcessKey(t *testing.T) {
	input := []string{
		"3\tbanana",
		"1\tapple",
		"2\tcherry",
	}
	expected := []string{
		"1\tapple",
		"2\tcherry",
		"3\tbanana", // Сортировка по первой колонке (числам)
	}

	config := Config{keyColumn: 1}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	processKey(input, config)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	got := strings.Split(output, "\n")

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], expected[i])
		}
	}
}

// TestReadLines тестирует чтение строк
func TestReadLines(t *testing.T) {
	input := "line1\nline2\nline3"
	expected := []string{"line1", "line2", "line3"}

	reader := strings.NewReader(input)
	result := readLines(reader)

	if len(result) != len(expected) {
		t.Errorf("length mismatch: got %d, want %d", len(result), len(expected))
		return
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, result[i], expected[i])
		}
	}
}

// TestIntegration тестирует интеграционные сценарии
func TestIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		config   Config
		expected []string
	}{
		{
			name:  "numeric sort with duplicates",
			input: []string{"5", "3", "5", "1", "2"},
			config: Config{
				numeric: true,
				unique:  true,
			},
			expected: []string{"1", "2", "3", "5"},
		},
		{
			name:  "reverse numeric sort",
			input: []string{"1", "3", "2"},
			config: Config{
				numeric: true,
				reverse: true,
			},
			expected: []string{"3", "2", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Выбираем обработчик в зависимости от конфигурации
			switch {
			case tt.config.unique && tt.config.reverse && tt.config.numeric:
				processUniqueReverseNumeric(tt.input, tt.config)
			case tt.config.unique && tt.config.numeric:
				processUniqueNumeric(tt.input, tt.config)
			case tt.config.unique && tt.config.reverse:
				processUniqueReverse(tt.input, tt.config)
			case tt.config.unique:
				processUnique(tt.input, tt.config)
			case tt.config.reverse && tt.config.numeric:
				processReverseNumeric(tt.input, tt.config)
			case tt.config.numeric:
				processNumeric(tt.input, tt.config)
			case tt.config.reverse:
				processReverse(tt.input, tt.config)
			default:
				processSimple(tt.input, tt.config)
			}

			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())
			got := strings.Split(output, "\n")

			if len(got) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(got), len(tt.expected))
				return
			}

			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// BenchmarkSorting бенчмарк для измерения производительности сортировки
func BenchmarkSorting(b *testing.B) {
	// Создаем большой набор данных для тестирования
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", 1000-i)
	}

	config := Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Копируем данные для каждого запуска
		testLines := make([]string, len(lines))
		copy(testLines, lines)

		processSimple(testLines, config)
	}
}
