package mapper

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func TelegramFormat[Anything any](array []Anything) (string, error) {
	if len(array) == 0 {
		return "", errors.New("Нет данных.")
	}
	t := reflect.TypeOf(array[0])
	if t.Kind() != reflect.Struct {
		return "", errors.New("Ошибка: тип должен быть структурой.")
	}
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			headers = append(headers, field.Name)
		}
	}
	if len(headers) == 0 {
		return "", errors.New("Ошибка: нет экспортированных полей.")
	}

	var sections []string
	for _, item := range array {
		rv := reflect.ValueOf(item)
		var lines []string
		for _, header := range headers {
			field := rv.FieldByName(header)
			if !field.IsValid() {
				continue
			}
			value := ""
			switch field.Kind() {
			case reflect.String:
				value = field.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value = strconv.Itoa(int(field.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value = strconv.FormatUint(field.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				value = fmt.Sprintf("%.2f", field.Float())
			case reflect.Bool:
				value = strings.ToLower(fmt.Sprintf("%t", field.Bool()))
			case reflect.Struct:
				if tm, ok := field.Interface().(time.Time); ok {
					value = tm.Format("2006-01-02 15:04")
				} else {
					value = fmt.Sprintf("%v", field.Interface())
				}
			default:
				value = fmt.Sprintf("%v", field.Interface())
			}
			// Экранирование для MarkdownV2
			value = strings.ReplaceAll(value, "\\", "\\\\")
			value = strings.ReplaceAll(value, "_", "\\_")
			value = strings.ReplaceAll(value, "*", "\\*")
			value = strings.ReplaceAll(value, "[", "\\[")
			value = strings.ReplaceAll(value, "]", "\\]")
			value = strings.ReplaceAll(value, "(", "\\(")
			value = strings.ReplaceAll(value, ")", "\\)")
			value = strings.ReplaceAll(value, "~", "\\~")
			value = strings.ReplaceAll(value, "`", "\\`")
			value = strings.ReplaceAll(value, ">", "\\>")
			line := fmt.Sprintf("*%s* - %s", header, value)
			lines = append(lines, line)
		}
		if len(lines) > 0 {
			sections = append(sections, strings.Join(lines, "\n"))
		}
	}

	if len(sections) == 0 {
		return "", errors.New("Нет данных.")
	}
	return strings.Join(sections, "\n\n"), nil
}
