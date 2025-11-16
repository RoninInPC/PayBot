package mapper

import (
	"fmt"
	"main/internal/entity"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestToExcel(t *testing.T) {
	users := []entity.User{
		{ID: 1, ContainsSub: true, TotalSub: 5, PromocodeID: 101, UserName: "alice_user", FirstTime: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)},
		{ID: 2, ContainsSub: false, TotalSub: 0, PromocodeID: 102, UserName: "bob_user", FirstTime: time.Date(2023, 2, 20, 14, 45, 0, 0, time.UTC)},
	}

	filename, _ := ToExcel[entity.User](users)
	defer os.Remove(filename)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Ожидали файл %s, но его нет", filename)
	}

	f, err := excelize.OpenFile(filename)
	if err != nil {
		t.Fatalf("Не удалось открыть файл: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Не удалось прочитать строки: %v", err)
	}

	expectedHeaders := []string{"Id", "ContainsSub", "TotalSub", "PromocodeID", "UserName", "FirstTime"}
	if len(rows) < 1 || !reflect.DeepEqual(rows[0], expectedHeaders) {
		t.Errorf("Ожидали заголовки %v, получили %v", expectedHeaders, rows[0])
	}

	expectedFirstRow := []string{"1", "true", "5", "101", "alice_user", "2023-01-15T10:30:00Z"}
	if len(rows) < 2 || !reflect.DeepEqual(rows[1], expectedFirstRow) {
		t.Errorf("Ожидали первую строку %v, получили %v", expectedFirstRow, rows[1])
	}
}

func TestFromExcel(t *testing.T) {
	originalUsers := []entity.User{
		{ID: 1, ContainsSub: true, TotalSub: 5, PromocodeID: 101, UserName: "alice_user", FirstTime: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)},
		{ID: 2, ContainsSub: false, TotalSub: 0, PromocodeID: 102, UserName: "bob_user", FirstTime: time.Date(2023, 2, 20, 14, 45, 0, 0, time.UTC)},
	}

	filename, _ := ToExcel[entity.User](originalUsers)
	defer os.Remove(filename)

	imported, _ := FromExcel[entity.User](filename)

	if len(imported) != len(originalUsers) {
		t.Errorf("Ожидали %d записей, получили %d", len(originalUsers), len(imported))
	}

	for i, orig := range originalUsers {
		imp := imported[i]
		if orig.ID != imp.ID {
			t.Errorf("Id: ожидали %d, получили %d", orig.ID, imp.ID)
		}
		if orig.ContainsSub != imp.ContainsSub {
			t.Errorf("ContainsSub: ожидали %t, получили %t", orig.ContainsSub, imp.ContainsSub)
		}
		if orig.TotalSub != imp.TotalSub {
			t.Errorf("TotalSub: ожидали %d, получили %d", orig.TotalSub, imp.TotalSub)
		}
		if orig.PromocodeID != imp.PromocodeID {
			t.Errorf("PromocodeID: ожидали %d, получили %d", orig.PromocodeID, imp.PromocodeID)
		}
		if orig.UserName != imp.UserName {
			t.Errorf("UserName: ожидали %s, получили %s", orig.UserName, imp.UserName)
		}
		if !orig.FirstTime.Equal(imp.FirstTime) && orig.FirstTime.Sub(imp.FirstTime).Abs().Seconds() > 1 {
			t.Errorf("FirstTime: ожидали %v, получили %v", orig.FirstTime, imp.FirstTime)
		}
	}
}

func TestToExcelEmpty(t *testing.T) {
	var users []entity.User
	filename, _ := ToExcel[entity.User](users)
	if filename != "" {
		t.Errorf("Ожидали пустое имя файла для пустого слайса, получили %s", filename)
	}
}

func TestFromExcelEmptyFile(t *testing.T) {
	emptyFile := "empty.xlsx"
	f := excelize.NewFile()
	f.SaveAs(emptyFile)
	defer os.Remove(emptyFile)

	imported, _ := FromExcel[entity.User](emptyFile)
	if len(imported) != 0 {
		t.Errorf("Ожидали пустой слайс для пустого файла, получили %d", len(imported))
	}
}

func TestToExcelNonStruct(t *testing.T) {
	testString := []string{"Not struct"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ожидали панику для не-структуры, но не получили")
		} else {
			panicMsg := fmt.Sprintf("%v", r)
			if strings.Contains(panicMsg, "ToExcel: Anything должен быть типом структуры") {
				t.Log("Паника сработала правильно:", panicMsg)
			} else {
				t.Errorf("Неправильный текст паники: ожидали 'ToExcel: Anything должен быть типом структуры', получили %v", r)
			}
		}
	}()
	ToExcel[string](testString)
}
