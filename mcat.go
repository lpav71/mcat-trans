package main

import (
	"fmt"
	"github.com/fatih/color"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxGoroutines = 3 // Максимальное количество одновременно выполняемых горутин

var mu sync.Mutex       // Создаем mutex для защиты переменной totalProcessed
var totalProcessed uint // Переменная для хранения общего количества обработанных записей

func formatNumberWithSpaces(num int) string {
	// Конвертируем число в строку
	strNum := strconv.Itoa(num)
	n := len(strNum)
	var result strings.Builder

	for i, digit := range strNum {
		// Добавляем пробелы в нужные места
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(digit)
	}

	return result.String()
}

// Функция для очистки имени файла от специальных символов
func sanitizeName(record string) string {
	re := regexp.MustCompile(`[<>:"'/\\|?*]+`)
	cleaned := re.ReplaceAllString(record, "")
	return strings.TrimSpace(cleaned)
}

func processRecords(db *gorm.DB, record string) {
	red := color.New(color.FgRed).SprintFunc()

	recordSanitized := sanitizeName(record)
	var fgMcatParamsList = FgMcatParamsList{
		Name: recordSanitized,
	}
	if err := db.Create(&fgMcatParamsList).Error; err != nil {
		log.Print(red("Не удалось создать запись:"), err)
		return
	}

	const batchSize = 500
	var params []FgMcatParams

	// Извлекаем блоки записей
	var lastProcessedId uint32
	for {
		var batch []FgMcatParams
		if err := db.Model(&FgMcatParams{}).
			Where("ParamName = ? AND Id > ?", record, lastProcessedId).
			Order("Id").
			Limit(batchSize).
			Find(&batch).Error; err != nil {
			log.Print(red("не удалось извлечь записи по ParamName:"), err)
			return
		}

		if len(batch) == 0 {
			break // выход из цикла, если больше нет записей
		}

		params = append(params, batch...)
		lastProcessedId = batch[len(batch)-1].ID // обновляем на последний элемент
	}

	tx := db.Begin() // Начинаем транзакцию
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Откат транзакции при панике
		}
	}()

	for _, param := range params {
		cleanedParam := sanitizeName(param.ParamValue)
		paramValue := FgMcatParamsValues{
			ParamID: fgMcatParamsList.ID,
			Value:   cleanedParam,
		}

		if err := tx.Create(&paramValue).Error; err != nil {
			log.Print(red("не удалось выполнить вставку в базу данных:"), err)
			tx.Rollback() // Откат при ошибке вставки
			return
		}
		valueId := paramValue.ID
		var fgMcatParamsListValues = FgMcatParamsListValues{
			FgMcatParamsItemHash: param.ItemHash,
			FgMcatParamsValuesID: valueId,
		}
		if err := tx.Create(&fgMcatParamsListValues).Error; err != nil {
			log.Print(red("не удалось выполнить вставку в базу данных:"), err)
			tx.Rollback() // Откат при ошибке вставки
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Print(red("не удалось подтвердить транзакцию:"), err)
	}

	mu.Lock()
	totalProcessed += uint(len(params)) // Обновляем общее количество обработанных записей
	// Форматируем число с пробелами
	formatted := formatNumberWithSpaces(int(totalProcessed))
	fmt.Println("Обработано записей:", formatted)
	mu.Unlock()
}

func main() {
	// Создаем экземпляры цветов
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	dsn := "mysql:mysql@tcp(127.0.0.1:3306)/fg_main_catalogue?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(red("не удалось подключиться к базе данных:"), err)
	}

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(red("не удалось получить доступ к базовым соединениям:"), err)
	}
	sqlDB.SetMaxOpenConns(50)                 // Максимум 50 открытых соединений
	sqlDB.SetMaxIdleConns(20)                 // Максимум 20 простаивающих соединений
	sqlDB.SetConnMaxLifetime(time.Minute * 5) // Время жизни соединения

	db.Logger = logger.Default.LogMode(logger.Error)

	fmt.Println(blue("Начало обработки..."))

	startTime := time.Now() // Запоминаем начальное время

	// 1. Извлечение уникальных значений ParamName из fg_mcat_params
	var uniqueParams []string
	if err := db.Model(&FgMcatParams{}).Distinct("ParamName").Pluck("ParamName", &uniqueParams).Error; err != nil {
		log.Fatal("не удалось извлечь уникальные значения ParamName:", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxGoroutines) // Семафор для ограничения количества горутин

	for _, paramName := range uniqueParams {
		wg.Add(1) // Увеличиваем счетчик горутин
		go func(paramName string) {
			defer wg.Done()                // Уменьшаем счетчик при завершении
			semaphore <- struct{}{}        // Блокируем, если достигли лимита
			defer func() { <-semaphore }() // Освобождаем после завершения

			// Используем существующее соединение
			processRecords(db, paramName)
		}(paramName)
	}

	wg.Wait()                            // Ждем завершения всех горутин
	elapsedTime := time.Since(startTime) // Вычисляем время выполнения
	fmt.Printf(green("Время выполнения (форматированный вывод): %.2f секунд\n"), elapsedTime.Seconds())
	fmt.Println(green("Время выполнения (стандарный вывод):", elapsedTime))
	formatted := formatNumberWithSpaces(int(totalProcessed))
	fmt.Println("Обработано записей:", formatted)
	fmt.Println(yellow("Обработка завершена"))
	fmt.Scanln()
}
