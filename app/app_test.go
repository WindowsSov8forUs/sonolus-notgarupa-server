package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/config"
	"github.com/gin-gonic/gin"
)

const (
	notGarupaEngineName         = "notgarupa"
	notGarupaHabahiroEngineName = "notgarupa-habahiro"
)

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

func TestUploadPublishesToRepositoryAndRefreshesSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestConfig(t)
	router, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	info := getJSON(t, router, "/sonolus/info")
	assertButtonPresent(t, info, "level")

	id := uploadLevel(t, router, "flow")
	assertUUID(t, id)
	levelName := levelNameForUUID(notGarupaEngineName, id)
	assertLevelName(t, levelName, notGarupaEngineName, id)
	upload := readUploadedLevel(t, cfg.Repository.SourceDir, levelName)
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
	assertItemPresent(t, levelList, levelName)
	details := getJSON(t, router, "/sonolus/levels/"+levelName)
	assertLevelMetadata(t, details, 21, "Uploaded Artists", "Uploaded Author")
	assertLevelDescription(t, details, "Uploaded Description")
	assertLevelTags(t, details, []string{"upload", "test"})
	assertLevelUsesEngine(t, details, notGarupaEngineName)
	assertRepositoryURLPrefix(t, details, "cover", "/sonolus/repository/")
}

func TestUploadSelectsHabahiroForSingleWidth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestConfig(t)
	router, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	id := uploadLevelWithChart(t, router, "single-width", `[
		{"type":"BPM","beat":0,"value":120},
		{"type":"Single","beat":1,"lane":3,"width":2}
	]`)

	assertUUID(t, id)
	levelName := levelNameForUUID(notGarupaHabahiroEngineName, id)
	assertLevelName(t, levelName, notGarupaHabahiroEngineName, id)
	upload := readUploadedLevel(t, cfg.Repository.SourceDir, levelName)
	if upload.Engine != notGarupaHabahiroEngineName {
		t.Fatalf("engine=%q", upload.Engine)
	}
	details := getJSON(t, router, "/sonolus/levels/"+levelName)
	assertLevelUsesEngine(t, details, notGarupaHabahiroEngineName)
}

func TestUploadWritesLevelSourceFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestConfig(t)
	router, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	id := uploadLevel(t, router, "local-store")
	levelName := levelNameForUUID(notGarupaEngineName, id)
	for _, name := range []string{"item.json", "cover.png", "bgm.mp3", "data.json", "chart.json"} {
		if _, err := os.Stat(filepath.Join(cfg.Repository.SourceDir, "levels", levelName, name)); err != nil {
			t.Fatalf("missing uploaded source file %s: %v", name, err)
		}
	}
}

func TestBuildRouterDoesNotSeedBuiltinCatalogWhenRepositoryUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{
		Server: config.ServerConfig{
			Listen: "127.0.0.1:0",
		},
		Repository: config.RepositoryConfig{
			SourceDir: filepath.Join(t.TempDir(), "missing-source"),
			PackDir:   filepath.Join(t.TempDir(), "missing-pack"),
			TmpDir:    filepath.Join(t.TempDir(), "tmp"),
		},
	}
	router, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	info := getJSON(t, router, "/sonolus/info")
	assertButtonPresent(t, info, "level")
	assertButtonAbsent(t, info, "engine")
	assertButtonAbsent(t, info, "skin")
	assertButtonAbsent(t, info, "background")
	assertButtonAbsent(t, info, "effect")
	assertButtonAbsent(t, info, "particle")
}

func TestUploadRejectsInvalidTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestConfig(t)
	router, err := BuildRouter(cfg)
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

func newTestConfig(t *testing.T) config.Config {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeTestSource(t, source)
	return config.Config{
		Server: config.ServerConfig{
			Listen: "127.0.0.1:0",
		},
		Repository: config.RepositoryConfig{
			SourceDir: source,
			PackDir:   filepath.Join(root, "pack"),
			TmpDir:    filepath.Join(root, "tmp"),
		},
	}
}

func readUploadedLevel(t *testing.T, sourceDir string, name string) repositoryUpload {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(sourceDir, "levels", name, "item.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Name        string            `json:"name"`
		Title       map[string]string `json:"title"`
		Artists     map[string]string `json:"artists"`
		Author      map[string]string `json:"author"`
		Description map[string]string `json:"description"`
		Rating      int               `json:"rating"`
		Engine      string            `json:"engine"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	return repositoryUpload{
		Name:        raw.Name,
		Title:       raw.Title["en"],
		Artists:     raw.Artists["en"],
		Author:      raw.Author["en"],
		Description: raw.Description["en"],
		Rating:      raw.Rating,
		Engine:      raw.Engine,
		Cover:       readFile(t, filepath.Join(sourceDir, "levels", name, "cover.png")),
		BGM:         readFile(t, filepath.Join(sourceDir, "levels", name, "bgm.mp3")),
		Data:        readFile(t, filepath.Join(sourceDir, "levels", name, "data.json")),
		Chart:       readFile(t, filepath.Join(sourceDir, "levels", name, "chart.json")),
	}
}

func writeTestSource(t *testing.T, source string) {
	t.Helper()
	writeTextFile(t, filepath.Join(source, "info.json"), `{"title":{"en":"Repository Test","zhs":"Repository Test","zht":"Repository Test"},"description":{"en":"Repository-backed test","zhs":"Repository-backed test","zht":"Repository-backed test"}}`)
	writeFile(t, filepath.Join(source, "banner.png"), testPNG())

	writeItem(t, filepath.Join(source, "skins", "skin"), itemJSON(4, "Skin", "Skin"), map[string][]byte{
		"thumbnail.png": testPNG(),
		"data":          []byte(`{}`),
		"texture.png":   testPNG(),
	})
	writeItem(t, filepath.Join(source, "backgrounds", "background"), itemJSON(2, "Background", "Background"), map[string][]byte{
		"thumbnail.png": testPNG(),
		"data":          []byte(`{}`),
		"image.png":     testPNG(),
		"configuration": []byte(`{}`),
	})
	writeItem(t, filepath.Join(source, "effects", "effect"), itemJSON(5, "Effect", "Effect"), map[string][]byte{
		"thumbnail": testPNG(),
		"data":      []byte(`{}`),
		"audio":     []byte("audio"),
	})
	writeItem(t, filepath.Join(source, "particles", "particle"), itemJSON(3, "Particle", "Particle"), map[string][]byte{
		"thumbnail": testPNG(),
		"data":      []byte(`{}`),
		"texture":   testPNG(),
	})
	for _, engine := range []string{notGarupaEngineName, notGarupaHabahiroEngineName} {
		writeItem(t, filepath.Join(source, "engines", engine), engineItemJSON(engine), map[string][]byte{
			"thumbnail":     testPNG(),
			"playData":      []byte(`{}`),
			"watchData":     []byte(`{}`),
			"previewData":   []byte(`{}`),
			"tutorialData":  []byte(`{}`),
			"configuration": []byte(`{}`),
		})
	}
}

func itemJSON(version int, title string, subtitle string) string {
	return `{"version":` + strconv.Itoa(version) + `,"title":{"en":"` + title + `","zhs":"` + title + `","zht":"` + title + `"},"subtitle":{"en":"` + subtitle + `","zhs":"` + subtitle + `","zht":"` + subtitle + `"},"author":{"en":"Author","zhs":"Author","zht":"Author"},"tags":[],"description":{"en":"Description","zhs":"Description","zht":"Description"}}`
}

func engineItemJSON(name string) string {
	return `{"version":13,"title":{"en":"` + name + `","zhs":"` + name + `","zht":"` + name + `"},"subtitle":{"en":"Engine","zhs":"Engine","zht":"Engine"},"author":{"en":"Author","zhs":"Author","zht":"Author"},"tags":[],"description":{"en":"Description","zhs":"Description","zht":"Description"},"skin":"skin","background":"background","effect":"effect","particle":"particle"}`
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
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &upload); err != nil {
		t.Fatal(err)
	}
	return upload.UUID
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

func assertUUID(t *testing.T, id string) {
	t.Helper()
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}$`)
	if !pattern.MatchString(id) {
		t.Fatalf("uuid %q does not match expected format", id)
	}
}

func levelNameForUUID(engine string, id string) string {
	return engine + "_" + id
}

func assertLevelName(t *testing.T, name string, engine string, id string) {
	t.Helper()
	want := levelNameForUUID(engine, id)
	if name != want {
		t.Fatalf("level name=%q, want %q", name, want)
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

func assertButtonAbsent(t *testing.T, response map[string]any, typ string) {
	t.Helper()
	buttons, ok := response["buttons"].([]any)
	if !ok {
		t.Fatalf("missing buttons in response: %#v", response)
	}
	for _, button := range buttons {
		object, ok := button.(map[string]any)
		if ok && object["type"] == typ {
			t.Fatalf("unexpected button %q in response: %#v", typ, buttons)
		}
	}
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

func assertRepositoryURLPrefix(t *testing.T, response map[string]any, field string, prefix string) {
	t.Helper()
	item := response["item"].(map[string]any)
	srl := item[field].(map[string]any)
	url, ok := srl["url"].(string)
	if !ok || !strings.HasPrefix(url, prefix) {
		t.Fatalf("%s url=%#v, want prefix %q", field, srl["url"], prefix)
	}
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

func writeItem(t *testing.T, dir string, item string, files map[string][]byte) {
	t.Helper()
	writeTextFile(t, filepath.Join(dir, "item.json"), item)
	for name, data := range files {
		writeFile(t, filepath.Join(dir, name), data)
	}
}

func writeTextFile(t *testing.T, path string, content string) {
	t.Helper()
	writeFile(t, path, []byte(content))
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readAll(t *testing.T, reader io.Reader) []byte {
	t.Helper()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
