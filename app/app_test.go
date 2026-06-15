package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	coredb "github.com/WindowsSov8forUs/sonolus-core-go/database"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/config"
	"github.com/gin-gonic/gin"
)

const (
	notGarupaEngineName         = "notgarupa"
	notGarupaHabahiroEngineName = "notgarupa-habahiro"
)

type fakeRepository struct {
	t       *testing.T
	uploads []repositoryUpload
	levels  []coredb.DatabaseLevelItem
	server  *httptest.Server
}

type repositoryUpload struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Artists     string   `json:"artists"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Rating      int      `json:"rating"`
	Engine      string   `json:"engine"`
	Cover       []byte   `json:"cover"`
	BGM         []byte   `json:"bgm"`
	Data        []byte   `json:"data"`
	Chart       []byte   `json:"chart,omitempty"`
}

func newFakeRepository(t *testing.T) *fakeRepository {
	t.Helper()
	repo := &fakeRepository{t: t}
	mux := http.NewServeMux()
	mux.HandleFunc("/manifest.json", repo.manifest)
	mux.HandleFunc("/admin/levels", repo.levelsHandler)
	repo.server = httptest.NewServer(mux)
	t.Cleanup(repo.server.Close)
	return repo
}

func (r *fakeRepository) manifest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"version":     len(r.levels) + 1,
		"generatedAt": "2026-06-15T00:00:00Z",
		"db": map[string]any{
			"info": map[string]any{
				"title":       map[string]string{"en": "Repository Test"},
				"description": map[string]string{"en": "Repository-backed test"},
			},
			"posts":       []any{},
			"playlists":   []any{},
			"levels":      r.levels,
			"skins":       []coredb.DatabaseSkinItem{testDBSkin()},
			"backgrounds": []coredb.DatabaseBackgroundItem{testDBBackground()},
			"effects":     []coredb.DatabaseEffectItem{testDBEffect()},
			"particles":   []coredb.DatabaseParticleItem{testDBParticle()},
			"engines": []coredb.DatabaseEngineItem{
				testDBEngine(notGarupaEngineName),
				testDBEngine(notGarupaHabahiroEngineName),
			},
			"replays": []any{},
		},
	})
}

func (r *fakeRepository) levelsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var upload repositoryUpload
	if err := json.NewDecoder(req.Body).Decode(&upload); err != nil {
		r.t.Fatalf("decode upload: %v", err)
	}
	r.uploads = append(r.uploads, upload)
	r.levels = append(r.levels, coredb.DatabaseLevelItem{
		Name:          upload.Name,
		Version:       1,
		Rating:        float64(upload.Rating),
		Title:         text(upload.Title),
		Artists:       text(upload.Artists),
		Author:        text(upload.Author),
		Tags:          tags(upload.Tags),
		Description:   text(upload.Description),
		Engine:        upload.Engine,
		UseSkin:       coredb.DatabaseUseItem{UseDefault: true},
		UseBackground: coredb.DatabaseUseItem{UseDefault: true},
		UseEffect:     coredb.DatabaseUseItem{UseDefault: true},
		UseParticle:   coredb.DatabaseUseItem{UseDefault: true},
		Cover:         testSrl(r.server.URL + "/sonolus/repository/cover"),
		BGM:           testSrl(r.server.URL + "/sonolus/repository/bgm"),
		Data:          testSrl(r.server.URL + "/sonolus/repository/data"),
	})
	_ = json.NewEncoder(w).Encode(map[string]int{"version": len(r.levels) + 1})
}

func TestUploadPublishesToRepositoryAndRefreshesManifest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepository(t)
	router, err := BuildRouter(config.Config{
		Address:                "http://127.0.0.1",
		Listen:                 "127.0.0.1:0",
		RepositoryAdminURL:     repo.server.URL,
		RepositoryManifestURL:  repo.server.URL + "/manifest.json",
		RepositoryPollInterval: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	info := getJSON(t, router, "/sonolus/info")
	assertButtonPresent(t, info, "level")

	uid := uploadLevel(t, router, "flow")
	assertLevelName(t, uid, "notgarupa-")
	if len(repo.uploads) != 1 {
		t.Fatalf("uploads=%d, want 1", len(repo.uploads))
	}
	upload := repo.uploads[0]
	if upload.Title != "flow" || upload.Artists != "Uploaded Artists" || upload.Author != "Uploaded Author" {
		t.Fatalf("upload metadata=%#v", upload)
	}
	if upload.Engine != notGarupaEngineName {
		t.Fatalf("engine=%q", upload.Engine)
	}
	if len(upload.Cover) == 0 || len(upload.BGM) == 0 || len(upload.Data) == 0 || len(upload.Chart) == 0 {
		t.Fatalf("upload missing binary payloads: cover=%d bgm=%d data=%d chart=%d", len(upload.Cover), len(upload.BGM), len(upload.Data), len(upload.Chart))
	}

	levelList := getJSON(t, router, "/sonolus/levels/list?localization=zhs&page=0")
	assertItemPresent(t, levelList, uid)
	details := getJSON(t, router, "/sonolus/levels/"+uid)
	assertLevelMetadata(t, details, 21, "Uploaded Artists", "Uploaded Author")
	assertLevelDescription(t, details, "Uploaded Description")
	assertLevelTags(t, details, []string{"upload", "test"})
	assertLevelUsesEngine(t, details, notGarupaEngineName)
	assertRepositoryURL(t, details, "cover", repo.server.URL+"/sonolus/repository/cover")
}

func TestUploadSelectsHabahiroForSingleWidth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepository(t)
	router, err := BuildRouter(config.Config{
		Address:                "http://127.0.0.1",
		Listen:                 "127.0.0.1:0",
		RepositoryAdminURL:     repo.server.URL,
		RepositoryManifestURL:  repo.server.URL + "/manifest.json",
		RepositoryPollInterval: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	uid := uploadLevelWithChart(t, router, "single-width", `[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Single","beat":1,"lane":3,"width":2}
	]`)

	assertLevelName(t, uid, "habahiro-")
	if repo.uploads[0].Engine != notGarupaHabahiroEngineName {
		t.Fatalf("engine=%q", repo.uploads[0].Engine)
	}
	details := getJSON(t, router, "/sonolus/levels/"+uid)
	assertLevelUsesEngine(t, details, notGarupaHabahiroEngineName)
}

func TestUploadRequiresRepositoryAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := BuildRouter(config.Config{
		Address: "http://127.0.0.1",
		Listen:  "127.0.0.1:0",
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := uploadLevelWithChartResponse(t, router, uploadOptions{title: "no-repo"})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("upload status=%d, want 500 body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadRejectsInvalidTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeRepository(t)
	router, err := BuildRouter(config.Config{
		Address:               "http://127.0.0.1",
		Listen:                "127.0.0.1:0",
		RepositoryAdminURL:    repo.server.URL,
		RepositoryManifestURL: repo.server.URL + "/manifest.json",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tagValue := range []string{"not-json", `{"tag":"x"}`, `["ok", 1]`} {
		rec := uploadLevelWithChartResponse(t, router, uploadOptions{title: "bad-tags", tags: tagValue})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("tags=%q upload status=%d, want %d body=%s", tagValue, rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}
}

type uploadOptions struct {
	title       string
	chart       string
	description string
	tags        string
}

func uploadLevel(t *testing.T, handler http.Handler, title string) string {
	t.Helper()
	return uploadLevelWithChart(t, handler, title, `[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Single","beat":1,"lane":3,"width":1},
		{"type":"Skill","beat":2,"lane":4,"width":1}
	]`)
}

func uploadLevelWithChart(t *testing.T, handler http.Handler, title string, chart string) string {
	t.Helper()
	rec := uploadLevelWithChartResponse(t, handler, uploadOptions{title: title, chart: chart})
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", rec.Code, rec.Body.String())
	}
	var upload struct {
		UID string `json:"uid"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &upload); err != nil {
		t.Fatal(err)
	}
	return upload.UID
}

func uploadLevelWithChartResponse(t *testing.T, handler http.Handler, options uploadOptions) *httptest.ResponseRecorder {
	t.Helper()
	if options.title == "" {
		options.title = "flow"
	}
	if options.chart == "" {
		options.chart = `[
			{"type":"BPM","beat":0,"value":120},
			{"type":"Single","beat":1,"lane":3,"width":1}
		]`
	}
	if options.description == "" && options.tags == "" {
		options.description = "Uploaded Description"
		options.tags = `["upload","test"]`
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", options.title)
	_ = writer.WriteField("rating", "21")
	_ = writer.WriteField("artists", "Uploaded Artists")
	_ = writer.WriteField("author", "Uploaded Author")
	if options.description != "" {
		_ = writer.WriteField("description", options.description)
	}
	if options.tags != "" {
		_ = writer.WriteField("tags", options.tags)
	}
	writeFormFile(t, writer, "chart", "chart.json", []byte(options.chart))
	writeFormFile(t, writer, "bgm", "test.wav", testWAV())
	writeFormFile(t, writer, "cover", "cover.png", testPNG())
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sonolus/levels", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(rec, req)
	return rec
}

func writeFormFile(t *testing.T, writer *multipart.Writer, field string, name string, data []byte) {
	t.Helper()
	part, err := writer.CreateFormFile(field, name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
}

func getJSON(t *testing.T, handler http.Handler, path string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("%s invalid json: %v body=%s", path, err, rec.Body.String())
	}
	return response
}

func assertLevelName(t *testing.T, name string, prefix string) {
	t.Helper()
	if !strings.HasPrefix(name, prefix) {
		t.Fatalf("level name %q does not have prefix %q", name, prefix)
	}
	pattern := regexp.MustCompile(`^(notgarupa|habahiro)-[23456789abcdefghijkmnpqrstuvwxyz]{9}_[23456789abcdefghijkmnpqrstuvwxyz]{6}$`)
	if !pattern.MatchString(name) {
		t.Fatalf("level name %q does not match expected format", name)
	}
}

func assertButtonPresent(t *testing.T, response map[string]any, typ string) {
	t.Helper()
	buttons, ok := response["buttons"].([]any)
	if !ok {
		t.Fatalf("missing buttons in response: %#v", response)
	}
	for _, button := range buttons {
		object, ok := button.(map[string]any)
		if ok && object["type"] == typ {
			return
		}
	}
	t.Fatalf("expected button %q in response: %#v", typ, buttons)
}

func assertItemPresent(t *testing.T, response map[string]any, name string) {
	t.Helper()
	items, ok := response["items"].([]any)
	if !ok {
		t.Fatalf("missing items in response: %#v", response)
	}
	for _, item := range items {
		object, ok := item.(map[string]any)
		if ok && object["name"] == name {
			return
		}
	}
	t.Fatalf("expected item %q in response: %#v", name, items)
}

func assertLevelUsesEngine(t *testing.T, response map[string]any, name string) {
	t.Helper()
	item := response["item"].(map[string]any)
	engine := item["engine"].(map[string]any)
	if engine["name"] != name {
		t.Fatalf("engine name=%#v, want %q", engine["name"], name)
	}
}

func assertLevelMetadata(t *testing.T, response map[string]any, rating float64, artists string, author string) {
	t.Helper()
	item := response["item"].(map[string]any)
	if item["rating"] != rating || item["artists"] != artists || item["author"] != author {
		t.Fatalf("metadata item=%#v", item)
	}
}

func assertLevelDescription(t *testing.T, response map[string]any, description string) {
	t.Helper()
	if response["description"] != description {
		t.Fatalf("details description=%#v, want %q", response["description"], description)
	}
}

func assertLevelTags(t *testing.T, response map[string]any, want []string) {
	t.Helper()
	item := response["item"].(map[string]any)
	rawTags := item["tags"].([]any)
	if len(rawTags) != len(want) {
		t.Fatalf("tag count=%d, want %d: %#v", len(rawTags), len(want), rawTags)
	}
	for i, wantTag := range want {
		tag := rawTags[i].(map[string]any)
		if tag["title"] != wantTag {
			t.Fatalf("tag[%d] title=%#v, want %q", i, tag["title"], wantTag)
		}
	}
}

func assertRepositoryURL(t *testing.T, response map[string]any, field string, url string) {
	t.Helper()
	item := response["item"].(map[string]any)
	srl := item[field].(map[string]any)
	if srl["url"] != url {
		t.Fatalf("%s url=%#v, want %q", field, srl["url"], url)
	}
}

func tags(values []string) []coredb.DatabaseTag {
	result := make([]coredb.DatabaseTag, 0, len(values))
	for _, value := range values {
		result = append(result, coredb.DatabaseTag{Title: text(value)})
	}
	return result
}

func testDBSkin() coredb.DatabaseSkinItem {
	return coredb.DatabaseSkinItem{Name: "skin", Version: 1, Title: text("Skin"), Subtitle: text("Skin"), Author: text("Author"), Thumbnail: testSrl("/skin-thumb"), Data: testSrl("/skin-data"), Texture: testSrl("/skin-texture")}
}

func testDBBackground() coredb.DatabaseBackgroundItem {
	return coredb.DatabaseBackgroundItem{Name: "background", Version: 1, Title: text("Background"), Subtitle: text("Background"), Author: text("Author"), Thumbnail: testSrl("/background-thumb"), Data: testSrl("/background-data"), Image: testSrl("/background-image"), Configuration: testSrl("/background-config")}
}

func testDBEffect() coredb.DatabaseEffectItem {
	return coredb.DatabaseEffectItem{Name: "effect", Version: 1, Title: text("Effect"), Subtitle: text("Effect"), Author: text("Author"), Thumbnail: testSrl("/effect-thumb"), Data: testSrl("/effect-data"), Audio: testSrl("/effect-audio")}
}

func testDBParticle() coredb.DatabaseParticleItem {
	return coredb.DatabaseParticleItem{Name: "particle", Version: 1, Title: text("Particle"), Subtitle: text("Particle"), Author: text("Author"), Thumbnail: testSrl("/particle-thumb"), Data: testSrl("/particle-data"), Texture: testSrl("/particle-texture")}
}

func testDBEngine(name string) coredb.DatabaseEngineItem {
	return coredb.DatabaseEngineItem{Name: name, Version: 13, Title: text(name), Subtitle: text("Engine"), Author: text("Author"), Skin: "skin", Background: "background", Effect: "effect", Particle: "particle", Thumbnail: testSrl("/engine-thumb"), PlayData: testSrl("/engine-play"), WatchData: testSrl("/engine-watch"), PreviewData: testSrl("/engine-preview"), TutorialData: testSrl("/engine-tutorial"), Configuration: testSrl("/engine-config")}
}

func text(value string) coredb.LocalizationText {
	return coredb.LocalizationText{"en": core.Text(value), "zhs": core.Text(value), "zht": core.Text(value)}
}

func testSrl(url string) core.Srl {
	hash := core.Value("hash" + url)
	value := core.Value(url)
	return core.Srl{Hash: &hash, URL: &value}
}

func testWAV() []byte {
	const sampleRate = 8000
	dataSize := sampleRate * 2
	buf := &bytes.Buffer{}
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, uint32(dataSize))
	for range sampleRate {
		_ = binary.Write(buf, binary.LittleEndian, uint16(0))
	}
	return buf.Bytes()
}

func testPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

func readAll(t *testing.T, reader io.Reader) []byte {
	t.Helper()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
