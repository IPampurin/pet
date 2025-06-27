/*README
			Добро пожаловать в best of the best планировщик задач, потрясающий своим богатейшим функционалом и бессмысленностью!

Описание:
	При запуске программа проверяет наличие и в случае отсутствия создаёт файл базы данных (БД) в папке, в которой находится (можете указать
	местоположение и имя файла БД в константе "dbFile"). Чтобы первый запуск стал успешным, ознакомьтесь с разделом "Запуск псевдоприложения".
	Все задания хранятся в отдельном файле БД, поэтому завершение программы не приведёт к потере записанных задач. Добавить задание можно только для будущего
	времени - если вводится дата в неправильном формате или указывает на прошедшее время, программа с завидной упёртостью предложит указать корректную дату.

Управление:
	Работа с планировщиком строится на основании следующих незамысловатых команд (поддерживаются и короткие варианты по первой букве), вводимых в консоли.
	create (c)		- при помощи "create" можно добавлять задачу (описание, дата в формате гггг.мм.дд) в базу данных.
	read (r)		- выводит список всех имеющихся задач, отсортированный по дате, количество одновременно выведенных на экран задач можно изменить в константе "Limit".
	update (u)		- запрашивает id задачи, которую надо изменить, и предлагает ввести новые значения описания и даты (всё в том же формате гггг.мм.дд).
	delete (d)		- удаляет задачу по id.
	basedelete (b)	- удаляет файл БД, запускает ракету на Марс и выпивает цианиду, чтобы вражеской разведке ничего не досталось.
	search (s)		- выводит задачи, содержащие поисковый запрос. Запросы на кириллице чувствительны к регистру.
	exit (e)		- выход из программы.

Запуск псевдоприложения:
	Для установки драйвера подключения БД необходимо, находясь в папке с фалом main.go, в консоли выполнить команды:
	1. "go mod init consoleToDoList" (consoleToDoList для примера, введите имя папки, в которой лежит файл с программой);
	2. "go get modernc.org/sqlite" (для подключения драйвера работы с БД);
	3. "go mod tidy" (для актуализации всех связей).
	Они превратят папку с программой в модуль с указанием всех связей (появятся два файла с указанием связей go.mod и go.sum, не удаляйте их).
	Далее просто запустите программу, например, командой "go run main.go" и следуйте инструкциям в консоли.

Комментарии:
	Недостатком программы является остановка с помощью panic() при некоторых внутренних ошибках, но доводить до ума и так уже много букв.
*/

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const (
	welcomeMessage       = "Welcome to the TO DO List CLI app!"                                           // приветствие при запуске программы
	commandMessage       = "Enter your command (create, read, update, delete, basedelete, search, exit):" // приглашение ввести команду
	inputContentMessage  = "Enter task content:"                                                          // приглашение ввести описание задачи
	inputDateMessage     = "Enter task date in format yyyy.mm.dd:"                                        // приглашение ввести дату, на которую запланирована задача
	updateMassage        = "Enter id task for update:"                                                    // приглашение ввести id задачи для обновления
	deleteMessage        = "Enter id task for delete:"                                                    // приглашение ввести id задачи для её удаления
	deleteBaseMessage    = "Database has been deleted. Restart the program."                              // сообщение об удалении БД
	dateInvTimeMessage   = "Enter correct date:"                                                          // приглашение ввести корректную дату
	searchMessage        = "Enter search query:"                                                          // приглашение к вводу искомой подстроки
	byeMessage           = "The program is completed. All data is saved. Good luck!"                      // сообщение при завершении программы
	errorCommandMessage  = "Invalid command! Please, try again!"                                          // сообщение о неверном вводе команды
	errorIdUpdateMassage = "Bad id for updating task."                                                    // сообщение о вводе неверного id задачи при обновлении
	errorPrefix          = "oops, something went wrong, programm is stopped, error: "                     // сообщение об ошибке, приведшей к завершению программы
)

const (
	dbFile      = "tasksDB.db" // название файла базы данных
	Limit       = 100          // количество единовременно выводимых на экран строк с заданиями
	dateFormfat = "2006.01.02" // формат ввода даты
)

// table - схема таблицы БД
const table = `
CREATE TABLE dataTask (
id INTEGER PRIMARY KEY AUTOINCREMENT,
content TEXT NOT NULL DEFAULT "",
date CHAR(8) NOT NULL DEFAULT ""
);
CREATE INDEX dataTask_date ON dataTask (date);`

// Task описывает структуру задачи
type Task struct {
	id      string
	content string
	date    string
}

func main() {

	err := initDB(dbFile)
	if err != nil {
		fmt.Println("call error initDB")
		panic(fmt.Sprint(errorPrefix, err))
	}

	fmt.Println(welcomeMessage)

	var input string

	for {
		fmt.Println(commandMessage)
		input = scanInput()

		switch {
		case input == "create" || input == "c" || input == "с": // на всякий случай и в кириллице
			create()
		case input == "read" || input == "r":
			read()
		case input == "update" || input == "u":
			update()
		case input == "delete" || input == "d":
			delTask()
		case input == "basedelete" || input == "b":
			basedelete()
			return
		case input == "search" || input == "s":
			search()
		case input == "exit" || input == "e":
			fmt.Println(byeMessage)
			return
		default:
			fmt.Println(errorCommandMessage)
		}
	}

}

// scanInput сканирует введённые данные
func scanInput() string {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	return scanner.Text()
}

// initDB проверяет наличие и создаёт БД, если её нет
func initDB(dbFile string) error {

	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Printf("opening error %s: ", dbFile)
		return err
	}
	defer db.Close()

	if install {
		_, err = db.Exec(table)
		if err != nil {
			fmt.Printf("error creating a table in %v or adding an index: ", dbFile)
			return err
		}
	}

	return nil
}

// checkDate проверяет корректность введённой даты
func checkDate(in string) bool {

	now := time.Now()

	date, err := time.Parse(dateFormfat, in)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return false
	}

	if !date.After(now) && (now.Format(dateFormfat) != in) {
		fmt.Println(dateInvTimeMessage)
		return false
	}

	return true
}

// create добавляет задачу в БД
func create() {

	var task Task

	fmt.Println(inputContentMessage)
	task.content = scanInput()

	var quest bool
	for !quest {
		fmt.Println(inputDateMessage)
		task.date = scanInput()
		quest = checkDate(task.date)
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer db.Close()

	var id int64

	query := "INSERT INTO dataTask (content, date) VALUES (:content, :date)"
	res, err := db.Exec(query,
		sql.Named("content", task.content),
		sql.Named("date", task.date))
	if err == nil {
		id, err = res.LastInsertId()
		if err != nil {
			panic(fmt.Sprint(errorPrefix, err))
		}
	}

	fmt.Printf("Task with id = %d added.\n", id)
}

// read выводит список всех задач, отсортированных по дате в максимальном количестве limit на странице
func read() {

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer db.Close()

	var allTasks []*Task
	var rows *sql.Rows

	rows, err = db.Query("SELECT * FROM dataTask ORDER BY date LIMIT :limit",
		sql.Named("limit", Limit))
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err = rows.Scan(&task.id, &task.content, &task.date)
		if err != nil {
			panic(fmt.Sprint(errorPrefix, err))
		}
		allTasks = append(allTasks, &task)
	}

	if err := rows.Err(); err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	fmt.Printf("%5s. %10s %v\n", "id", "date", "content")
	for _, val := range allTasks {
		fmt.Printf("%5s. %10s %v\n", val.id, val.date, val.content)
	}
}

// update позволяет обновить задание по введённому id задачи
func update() {

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer db.Close()

	var task Task

	fmt.Println(updateMassage)
	task.id = scanInput()

	fmt.Println(inputContentMessage)
	task.content = scanInput()

	var quest bool
	for !quest {
		fmt.Println(inputDateMessage)
		task.date = scanInput()
		quest = checkDate(task.date)
	}

	res, err := db.Exec("UPDATE dataTask SET content = :content, date = :date WHERE id = :id",
		sql.Named("content", &task.content),
		sql.Named("date", &task.date),
		sql.Named("id", &task.id))
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	count, err := res.RowsAffected()
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	if count == 0 {
		fmt.Println(errorIdUpdateMassage)
		return
	}
}

// delTask удаляет задачу по введённоу id
func delTask() {

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer db.Close()

	var task Task

	fmt.Println(deleteMessage)
	task.id = scanInput()

	_, err = db.Exec("DELETE FROM dataTask WHERE id = :id",
		sql.Named("id", &task.id))
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
}

// basedelete удаляет файл базы данных и запускает ракету к Марсу
func basedelete() {

	_, err := os.Stat(dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	err = os.Remove(dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	fmt.Println(deleteBaseMessage)
}

// search позволяет найти задачи, в описании или дате которых, есть введённая подстрока
func search() {

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer db.Close()

	fmt.Println(searchMessage)
	searching := scanInput()
	searching = "%" + searching + "%"

	var allTasks []*Task
	var rows *sql.Rows

	rows, err = db.Query("SELECT * FROM dataTask WHERE content LIKE :searching OR date LIKE :searching ORDER BY date LIMIT :limit",
		sql.Named("searching", searching),
		sql.Named("limit", Limit))
	if err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err = rows.Scan(&task.id, &task.content, &task.date)
		if err != nil {
			panic(fmt.Sprint(errorPrefix, err))
		}
		allTasks = append(allTasks, &task)
	}

	if err := rows.Err(); err != nil {
		panic(fmt.Sprint(errorPrefix, err))
	}

	fmt.Printf("%5s. %10s %v\n", "id", "date", "content")
	for _, val := range allTasks {
		fmt.Printf("%5s. %10s %v\n", val.id, val.date, val.content)
	}
}
