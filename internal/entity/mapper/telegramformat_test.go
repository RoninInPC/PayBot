package mapper

import (
	"main/internal/entity"
	"strings"
	"testing"
	"time"
)

func TestTelegramFormat(t *testing.T) {
	users := []entity.User{
		{ID: 1, ContainsSub: true, TotalSub: 5, PromocodeID: 101, UserName: "alice_user", FirstTime: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)},
	}
	markdown, _ := TelegramFormat[entity.User](users)
	expectedParts := []string{
		"*ID* - 1",
		"*ContainsSub* - true",
		"*TotalSub* - 5",
		"*PromocodeID* - 101",
		"*UserName* - alice\\_user",
		"*FirstTime* - 2023-01-15 10:30",
	}
	for _, part := range expectedParts {
		if !strings.Contains(markdown, part) {
			t.Errorf("Ожидали '%s' в Markdown, но не нашли", part)
		}
	}
	users2 := append(users, entity.User{ID: 2, ContainsSub: false, TotalSub: 0, PromocodeID: 102, UserName: "bob_user", FirstTime: time.Date(2023, 2, 20, 14, 45, 0, 0, time.UTC)})
	markdown2, _ := TelegramFormat[entity.User](users2)
	if !strings.Contains(markdown2, "\n\n") {
		t.Error("Ожидали пустую строку между записями")
	}
}

func TestTelegramFormatEmpty(t *testing.T) {
	var users []entity.User
	markdown, _ := TelegramFormat[entity.User](users)
	if markdown != "Нет данных." {
		t.Errorf("Ожидали 'Нет данных.' для пустого слайса, получили %s", markdown)
	}
}
