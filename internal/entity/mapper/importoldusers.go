package mapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"main/internal/entity"
	"os"
	"strconv"
	"strings"
	"time"
)

func ParseUsersFromCSV(filename string) ([]*entity.OldUser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать заголовок: %w", err)
	}
	expected := 15
	if len(header) != expected {
		log.Printf("Warning: заголовок имеет %d полей вместо %d. Продолжаем.", len(header), expected)
	}

	users := make([]*entity.OldUser, 0, 3000)
	lineNum := 1
	warnings := 0
	processed := 0

	layout := "2006-01-02 15:04:05"

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Warning: ошибка чтения строки %d: %v. Создаём user с defaults.", lineNum+1, err)
			warnings++
			users = append(users, &entity.OldUser{})
			lineNum++
			continue
		}
		lineNum++

		processed++
		user := &entity.OldUser{}

		if len(row) > 0 && row[0] != "" {
			user.UserID = strings.TrimSpace(row[0])
		}
		if len(row) > 1 && row[1] != "" {
			if reg, err := time.Parse(layout, strings.TrimSpace(row[1])); err == nil {
				user.Registration = reg
			} else {
				log.Printf("Warning: строка %d: неверная дата регистрации '%s'.", lineNum, row[1])
				warnings++
			}
		}
		if len(row) > 2 {
			user.Username = strings.TrimSpace(row[2])
		}
		if len(row) > 3 {
			user.Name = strings.TrimSpace(row[3])
		}
		if len(row) > 4 {
			user.RefLink = strings.TrimSpace(row[4])
		}
		if len(row) > 5 {
			user.Phone = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			user.Email = strings.TrimSpace(row[6])
		}
		if len(row) > 7 {
			user.Comment = strings.TrimSpace(row[7])
		}
		if len(row) > 8 && row[8] != "" {
			user.Active = strings.TrimSpace(row[8]) == "1"
		}
		if len(row) > 9 {
			user.Plan = strings.TrimSpace(row[9])
		}
		if len(row) > 10 && row[10] != "" {
			if end, err := time.Parse(layout, strings.TrimSpace(row[10])); err == nil {
				user.EndDate = end
			} else {
				log.Printf("Warning: строка %d: неверная end_date '%s'.", lineNum, row[10])
				warnings++
			}
		}
		if len(row) > 11 && row[11] != "" {
			trial := strings.TrimSpace(row[11])
			user.UseTrial = trial == "+" || trial == "1"
		}
		if len(row) > 12 && row[12] != "" {
			if count, err := strconv.Atoi(strings.TrimSpace(row[12])); err == nil {
				user.PayCount = count
			} else {
				log.Printf("Warning: строка %d: неверный pay_count '%s'. Устанавливаем 0.", lineNum, row[12])
				warnings++
				user.PayCount = 0
			}
		}
		if len(row) > 13 && row[13] != "" {
			if last, err := time.Parse(layout, strings.TrimSpace(row[13])); err == nil {
				user.PayLast = last
			} else {
				log.Printf("Warning: строка %d: неверная pay_last '%s'.", lineNum, row[13])
				warnings++
			}
		}
		if len(row) > 14 && row[14] != "" {
			if first, err := time.Parse(layout, strings.TrimSpace(row[14])); err == nil {
				user.Pay1st = first
			} else {
				log.Printf("Warning: строка %d: неверная pay_1st '%s'.", lineNum, row[14])
				warnings++
			}
		}

		users = append(users, user)
	}

	log.Printf("Обработано %d строк, warnings: %d. Получено %d пользователей.", processed, warnings, len(users))
	return users, nil
}
