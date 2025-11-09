package mapper

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func ToExcel[Anything any](array []Anything) (filename string, err error) {
	if len(array) == 0 {
		return "", errors.New("Not info")
	}
	t := reflect.TypeOf(array[0])
	if t.Kind() != reflect.Struct {
		return "", errors.New("ToExcel: Anything должен быть типом структуры")
	}
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			headers = append(headers, field.Name)
		}
	}
	if len(headers) == 0 {
		return "", errors.New("ToExcel: Не найдено экспортированных полей в структуре")
	}
	f := excelize.NewFile()
	defer f.Close()
	sheetName := f.GetSheetName(0)
	f.SetActiveSheet(0)
	for j, header := range headers {
		cellName, err := excelize.CoordinatesToCellName(j+1, 1)
		if err != nil {
			return "", errors.New(fmt.Sprintf("ToExcel: Ошибка генерации имени ячейки: %v", err))
		}
		if err := f.SetCellValue(sheetName, cellName, header); err != nil {
			return "", errors.New(fmt.Sprintf("ToExcel: Ошибка установки заголовка '%s': %v", header, err))
		}
	}
	for i, item := range array {
		rv := reflect.ValueOf(item)
		for j, header := range headers {
			field := rv.FieldByName(header)
			if !field.IsValid() {
				continue
			}
			cellName, err := excelize.CoordinatesToCellName(j+1, i+2)
			if err != nil {
				return "", errors.New(fmt.Sprintf("ToExcel: Ошибка генерации имени ячейки: %v", err))
			}
			var fallback string
			switch field.Kind() {
			case reflect.Bool:
				fallback = strings.ToLower(fmt.Sprintf("%t", field.Bool()))
			case reflect.String:
				fallback = field.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fallback = strconv.Itoa(int(field.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fallback = strconv.FormatUint(field.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				fallback = fmt.Sprintf("%.2f", field.Float())
			case reflect.Struct:
				if tm, ok := field.Interface().(time.Time); ok {
					fallback = tm.Format(time.RFC3339)
				} else {
					fallback = fmt.Sprintf("%v", field.Interface())
				}
			default:
				fallback = fmt.Sprintf("%v", field.Interface())
			}
			if err := f.SetCellValue(sheetName, cellName, fallback); err != nil {
				return "", errors.New(fmt.Sprintf("ToExcel: Ошибка установки значения '%s': %v", fallback, err))
			}
		}
	}
	filename = "output.xlsx"
	if err := f.SaveAs(filename); err != nil {
		return "", errors.New(fmt.Sprintf("ToExcel: Ошибка сохранения файла '%s': %v", filename, err))
	}
	return filename, nil
}

func FromExcel[Anything any](filename string) ([]Anything, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("FromExcel: Файл '%s' не найден", filename))
	}
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("FromExcel: Ошибка открытия файла '%s': %v", filename, err))
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("FromExcel: Ошибка чтения строк из листа '%s': %v", sheetName, err))
	}
	if len(rows) == 0 {
		return nil, errors.New("Empty Rows")
	}
	headers := rows[0]
	t := reflect.TypeOf((*Anything)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, errors.New("FromExcel: Anything должен быть типом структуры")
	}
	var result []Anything
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}
		rv := reflect.New(t).Elem()
		for j, header := range headers {
			if j >= len(row) {
				continue
			}
			cellValue := row[j]
			field, ok := t.FieldByName(header)
			if !ok {
				continue
			}
			fVal := rv.FieldByName(header)
			if !fVal.CanSet() {
				continue
			}
			switch field.Type.Kind() {
			case reflect.String:
				fVal.SetString(cellValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if intVal, err := strconv.ParseInt(cellValue, 10, 64); err == nil {
					fVal.SetInt(intVal)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if uintVal, err := strconv.ParseUint(cellValue, 10, 64); err == nil {
					fVal.SetUint(uintVal)
				}
			case reflect.Float32, reflect.Float64:
				if floatVal, err := strconv.ParseFloat(cellValue, 64); err == nil {
					fVal.SetFloat(floatVal)
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(cellValue); err == nil {
					fVal.SetBool(boolVal)
				}
			case reflect.Struct:
				if field.Type == reflect.TypeOf(time.Time{}) {
					if tm, err := time.Parse(time.RFC3339, cellValue); err == nil {
						fVal.Set(reflect.ValueOf(tm))
					} else {
						fmt.Println("ParseTime failed for", cellValue, ":", err)
					}
				} else {
					fmt.Println("Unsupported struct type for", header)
				}
			default:
				fmt.Println("Unsupported type for", header, ":", field.Type.Kind())
			}
		}
		newVal := rv.Interface().(Anything)
		result = append(result, newVal)
	}
	return result, nil
}
