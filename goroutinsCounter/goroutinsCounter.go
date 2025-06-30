/*
				Многопоточный счётчик.

Данная программа является примером решения следующей задачи:

	Имеется горутина, которая генерирует числа и отправляет их в канал. Дальше несколько горутин читают и раскидывают их по нескольким каналам.
	В конце происходит обратный процесс: все числа из каналов пишутся в один результирующий канал.
	Нужно, чтобы количество и сумма входящих чисел совпала с количеством и суммой чисел, которые получены из канала вывода.

*/

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const NumOut = 5 // количество обрабатывающих горутин и каналов

// generator генерирует последовательность чисел 1,2,3 и т.д. и отправляет их в канал ch,
// после записи в канал для каждого числа вызывается функция fn (для подсчёта количества и суммы сгенерированных чисел)
func generator(ctx context.Context, ch chan<- int64, fn func(int64)) {

	defer close(ch)
	var i int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			i++
			ch <- i
			fn(i)
		}
	}
}

// worker читает число из канала in и пишет его в канал out
func worker(in <-chan int64, out chan<- int64) {

	defer close(out)
	for {
		v, ok := <-in
		if !ok {
			return
		}
		out <- v
		time.Sleep(1 * time.Millisecond)
	}
}

func main() {

	// определим контекст
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// chIn - канал, куда будет писать числа generator
	chIn := make(chan int64)

	var inputSum int64   // сумма сгенерированных чисел
	var inputCount int64 // количество сгенерированных чисел

	// генерируем числа, подсчитывая их количество и сумму (применим потокобезопасный вариант)
	go generator(ctx, chIn, func(i int64) {
		atomic.AddInt64(&inputSum, i)
		atomic.AddInt64(&inputCount, i)
	})

	// outs - слайс каналов, куда будут записываться числа из chIn
	outs := make([]chan int64, NumOut)

	// создаём каналы и для каждого из них вызываем горутину worker
	for i := 0; i < NumOut; i++ {
		outs[i] = make(chan int64)
		go worker(chIn, outs[i])
	}

	// amounts - слайс, в который собирается статистика по горутинам
	amounts := make([]int64, NumOut)

	// chOut - канал, в который будут отправляться числа из горутин worker-ов
	chOut := make(chan int64, NumOut)

	var wg sync.WaitGroup

	// собираем числа из каналов outs и добавляем счетчик в статистике
	for i := 0; i < NumOut; i++ {
		wg.Add(1)
		go func(in <-chan int64, i int64) {
			defer wg.Done()
			for v := range in {
				chOut <- v
				amounts[i]++
			}
		}(outs[i], int64(i))
	}

	// ждём завершения работы всех горутин для outs и закрываем результирующий канал
	go func() {
		wg.Wait()
		close(chOut)
	}()

	var count int64 // количество чисел результирующего канала
	var sum int64   // сумма чисел результирующего канала

	// читаем числа из результирующего канала
	for v := range chOut {
		sum += v
		count++
	}

	fmt.Println("Количество чисел", inputCount, count)
	fmt.Println("Сумма чисел", inputSum, sum)
	fmt.Println("Разбивка по каналам", amounts)

	// проверка результатов
	if inputSum != sum {
		log.Fatalf("Ошибка: суммы чисел не равны: %d != %d\n", inputSum, sum)
	}
	if inputCount != count {
		log.Fatalf("Ошибка: количество чисел не равно: %d != %d\n", inputCount, count)
	}
	for _, v := range amounts {
		inputCount -= v
	}
	if inputCount != 0 {
		log.Fatalf("Ошибка: разделение чисел по каналам неверное\n")
	}
}
