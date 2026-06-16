package level

import (
	"testing"

	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	sonolus "github.com/WindowsSov8forUs/sonolus-server-go"
)

func TestListMatchesLevelByUUID(t *testing.T) {
	service := NewServiceFromItems(func() []sonolus.LevelItemModel {
		return []sonolus.LevelItemModel{
			{DatabaseLevelItem: database.DatabaseLevelItem{Name: "notgarupa_12345678-abcd-ef01"}},
			{DatabaseLevelItem: database.DatabaseLevelItem{Name: "notgarupa_87654321-abcd-ef01"}},
		}
	})

	list, err := service.List(sonolus.Context{}, sonolus.FormValue{
		Type: "uuid",
		Options: map[string]any{
			"keywords": "12345678-abcd-ef01",
		},
	}, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Name != "notgarupa_12345678-abcd-ef01" {
		t.Fatalf("items=%#v, want matching UUID level", list.Items)
	}
}

func TestListStillMatchesFullLevelName(t *testing.T) {
	service := NewServiceFromItems(func() []sonolus.LevelItemModel {
		return []sonolus.LevelItemModel{
			{DatabaseLevelItem: database.DatabaseLevelItem{Name: "notgarupa_12345678-abcd-ef01"}},
		}
	})

	list, err := service.List(sonolus.Context{}, sonolus.FormValue{
		Type: "uuid",
		Options: map[string]any{
			"keywords": "notgarupa_12345678-abcd-ef01",
		},
	}, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Name != "notgarupa_12345678-abcd-ef01" {
		t.Fatalf("items=%#v, want matching level name", list.Items)
	}
}
