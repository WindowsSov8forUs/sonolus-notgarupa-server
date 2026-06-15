package level

import (
	"net/http"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	coreserver "github.com/WindowsSov8forUs/sonolus-core-go/core/server"
	"github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-server-go"
)

const (
	pageSize = 20
)

type Service struct {
	items func() []sonolus.LevelItemModel
}

func NewServiceFromItems(items func() []sonolus.LevelItemModel) *Service {
	return &Service{items: items}
}

func (s *Service) Install(app *sonolus.Sonolus) {
	app.Level.Searches = levelSearches()
	app.Level.InfoHandler = s.Info
	app.Level.ListHandler = s.List
	app.Level.DetailsHandler = s.Details
}

func (s *Service) Info(ctx sonolus.Context, typ string) (sonolus.ItemInfoModel[sonolus.LevelItemModel], *sonolus.ServerError) {
	list, err := s.List(ctx, sonolus.FormValue{}, 0, "")
	if err != nil {
		return sonolus.ItemInfoModel[sonolus.LevelItemModel]{}, err
	}
	count := min(len(list.Items), 5)
	return sonolus.ItemInfoModel[sonolus.LevelItemModel]{
		Sections: []sonolus.ItemSectionModel{
			{
				Title:    localized("#NEWEST", "#NEWEST"),
				ItemType: core.ItemTypeLevel,
				Items:    list.Items[:count],
			},
		},
	}, nil
}

func (s *Service) List(ctx sonolus.Context, search sonolus.FormValue, page int, cursor string) (sonolus.ItemListModel[sonolus.LevelItemModel], *sonolus.ServerError) {
	return s.listItems(search, page), nil
}

func (s *Service) listItems(search sonolus.FormValue, page int) sonolus.ItemListModel[sonolus.LevelItemModel] {
	name, ok := searchName(search)
	if !ok {
		return sonolus.ItemListModel[sonolus.LevelItemModel]{PageCount: 0, Items: []sonolus.LevelItemModel{}}
	}
	var filtered []sonolus.LevelItemModel
	for _, item := range s.items() {
		if name == "" || item.Name == name {
			filtered = append(filtered, item)
		}
	}
	start := page * pageSize
	if start >= len(filtered) {
		return sonolus.ItemListModel[sonolus.LevelItemModel]{PageCount: pageCount(len(filtered)), Items: []sonolus.LevelItemModel{}}
	}
	end := min(start+pageSize, len(filtered))
	return sonolus.ItemListModel[sonolus.LevelItemModel]{
		PageCount: pageCount(len(filtered)),
		Items:     filtered[start:end],
	}
}

func levelSearches() sonolus.FormsModel {
	return sonolus.FormsModel{
		"uid": {
			Title: localized("Search", "Search"),
			Options: sonolus.OptionsModel{
				"keywords": sonolus.TextOption{
					OptionBase: sonolus.OptionBase{
						Name: localized("UID", "UID"),
					},
					Placeholder: localized("Enter level UID", "Enter level UID"),
					Shortcuts:   []string{},
				},
			},
		},
	}
}

func (s *Service) Details(ctx sonolus.Context, name string) (model sonolus.ItemDetailsModel[sonolus.LevelItemModel], serverErr *sonolus.ServerError) {
	defer func() {
		if serverErr != nil {
			return
		}
		if model.Leaderboards == nil {
			model.Leaderboards = []coreserver.ServerItemLeaderboard{}
		}
		if model.Sections == nil {
			model.Sections = []sonolus.ItemSectionModel{}
		}
	}()

	for _, item := range s.items() {
		if item.Name == name {
			return sonolus.ItemDetailsModel[sonolus.LevelItemModel]{
				Item:        item,
				Description: item.Description,
			}, nil
		}
	}
	return sonolus.ItemDetailsModel[sonolus.LevelItemModel]{}, sonolus.StatusError(http.StatusNotFound)
}

func searchName(search sonolus.FormValue) (string, bool) {
	if search.Type != "quick" && search.Type != "uid" {
		return "", true
	}
	keywords, _ := search.Options["keywords"].(string)
	if keywords == "" {
		return "", true
	}
	return keywords, true
}

func pageCount(count int) int {
	return count/pageSize + 1
}

func localized(en string, zhs string) database.LocalizationText {
	return database.LocalizationText{"en": core.Text(en), "zhs": core.Text(zhs), "zht": core.Text(zhs)}
}
