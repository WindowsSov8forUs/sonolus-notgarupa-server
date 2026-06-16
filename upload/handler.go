package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/chart"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/repository"
	"github.com/WindowsSov8forUs/sonolus-notgarupa-server/uid"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	maxBGMSize          = 20 * 1024 * 1024
	maxCoverSize        = 5 * 1024 * 1024
	maxNameAttempts     = 5
	maxDescriptionRunes = 2048
	maxTagCount         = 16
	maxTagRunes         = 64
)

type Handler struct {
	Engines         map[string]bool
	LevelNames      LevelNameIndex
	Publisher       LevelPublisher
	RefreshSnapshot func(repository.Snapshot)
	GenerateName    func(engine string, now time.Time) (string, error)
	mu              sync.Mutex
}

type LevelNameIndex interface {
	HasLevel(name string) bool
}

type LevelPublisher interface {
	UploadLevel(ctx context.Context, upload repository.LevelUpload) (repository.Snapshot, error)
}

func (h *Handler) Install(router gin.IRouter) {
	router.POST("/sonolus/levels", h.post)
}

func (h *Handler) post(ctx *gin.Context) {
	form, err := bind(ctx)
	if err != nil {
		writeError(ctx, formError(err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()

	bgmData, err := readBGM(form.BGM)
	if err != nil {
		writeError(ctx, err)
		return
	}
	coverData, err := readCover(form.Cover)
	if err != nil {
		writeError(ctx, err)
		return
	}

	var garupaChart chart.GarupaChart
	if err := json.Unmarshal([]byte(form.Chart), &garupaChart); err != nil {
		writeError(ctx, chartError(fmt.Errorf("invalid chart json: %w", err)))
		return
	}
	engine := h.preferredEngine(garupaChart)
	chartData, err := json.Marshal(garupaChart)
	if err != nil {
		writeError(ctx, chartError(err))
		return
	}
	sonolusChart, err := garupaChart.ConvertToSonolus()
	if err != nil {
		writeError(ctx, chartConvertError(err))
		return
	}
	sonolusData, err := json.Marshal(sonolusChart)
	if err != nil {
		writeError(ctx, chartConvertError(err))
		return
	}
	if h.Publisher == nil {
		writeError(ctx, databaseError(fmt.Errorf("level publisher is not configured")))
		return
	}
	name, snapshot, err := h.uploadWithGeneratedName(ctx, engine, now, form, coverData, bgmData, sonolusData, chartData)
	if err != nil {
		writeError(ctx, databaseError(err))
		return
	}
	if h.RefreshSnapshot != nil {
		h.RefreshSnapshot(snapshot)
	}

	ctx.JSON(http.StatusOK, gin.H{"uid": name})
}

func (h *Handler) uploadWithGeneratedName(ctx *gin.Context, engine string, now time.Time, form form, coverData []byte, bgmData []byte, sonolusData []byte, chartData []byte) (string, repository.Snapshot, error) {
	for i := 0; i < maxNameAttempts; i++ {
		name, err := h.generateUniqueName(engine, now)
		if err != nil {
			return "", repository.Snapshot{}, err
		}
		snapshot, err := h.uploadToRepository(ctx, name, form, engine, coverData, bgmData, sonolusData, chartData)
		if errors.Is(err, repository.ErrLevelExists) {
			continue
		}
		if err != nil {
			return "", repository.Snapshot{}, err
		}
		return name, snapshot, nil
	}
	return "", repository.Snapshot{}, fmt.Errorf("failed to publish unique level name after %d attempts", maxNameAttempts)
}

func (h *Handler) uploadToRepository(ctx *gin.Context, name string, form form, engine string, coverData []byte, bgmData []byte, sonolusData []byte, chartData []byte) (repository.Snapshot, error) {
	snapshot, err := h.Publisher.UploadLevel(ctx.Request.Context(), repository.LevelUpload{
		Name:        name,
		Title:       form.Title,
		Artists:     form.Artists,
		Author:      form.Author,
		Description: form.Description,
		Tags:        form.Tags,
		Rating:      form.Rating,
		Engine:      engine,
		Cover:       coverData,
		BGM:         bgmData,
		Data:        sonolusData,
		Chart:       chartData,
	})
	if err != nil {
		return repository.Snapshot{}, err
	}
	return snapshot, nil
}

func (h *Handler) preferredEngine(garupaChart chart.GarupaChart) string {
	preferred := garupaChart.PreferredEngine()
	if h.hasEngine(preferred) {
		return preferred
	}
	if h.hasEngine(chart.EngineNotGarupa) {
		return chart.EngineNotGarupa
	}
	if h.hasEngine(chart.EngineNotGarupaHabahiro) {
		return chart.EngineNotGarupaHabahiro
	}
	return preferred
}

func (h *Handler) hasEngine(name string) bool {
	if len(h.Engines) == 0 {
		return true
	}
	return h.Engines[name]
}

type form struct {
	Title       string
	Artists     string
	Author      string
	BGM         *multipart.FileHeader
	Cover       *multipart.FileHeader
	Chart       string
	Rating      int
	Description string
	Tags        []string
}

func bind(ctx *gin.Context) (form, error) {
	title := strings.TrimSpace(ctx.PostForm("title"))
	if title == "" {
		return form{}, fmt.Errorf("title is required")
	}
	title = truncateRunes(title, 256)
	artists := strings.TrimSpace(ctx.PostForm("artists"))
	if artists == "" {
		return form{}, fmt.Errorf("artists is required")
	}
	artists = truncateRunes(artists, 256)
	author := strings.TrimSpace(ctx.PostForm("author"))
	if author == "" {
		return form{}, fmt.Errorf("author is required")
	}
	author = truncateRunes(author, 256)
	bgm, err := ctx.FormFile("bgm")
	if err != nil {
		return form{}, fmt.Errorf("bgm is required: %w", err)
	}
	cover, err := ctx.FormFile("cover")
	if err != nil {
		return form{}, fmt.Errorf("cover is required: %w", err)
	}
	chartValue, err := readChart(ctx)
	if err != nil {
		return form{}, err
	}
	chartValue = strings.TrimPrefix(strings.TrimSpace(chartValue), "\ufeff")
	if chartValue == "" {
		return form{}, fmt.Errorf("chart is required")
	}
	ratingValue := ctx.PostForm("rating")
	rating, err := parsePositiveInt(ratingValue)
	if err != nil {
		return form{}, fmt.Errorf("rating is required: %w", err)
	}
	description := truncateRunes(strings.TrimSpace(ctx.PostForm("description")), maxDescriptionRunes)
	tags, err := parseTags(ctx.PostForm("tags"))
	if err != nil {
		return form{}, err
	}
	return form{
		Title:       title,
		Artists:     artists,
		Author:      author,
		BGM:         bgm,
		Cover:       cover,
		Chart:       chartValue,
		Rating:      rating,
		Description: description,
		Tags:        tags,
	}, nil
}

func readChart(ctx *gin.Context) (string, error) {
	if value := ctx.PostForm("chart"); value != "" {
		return value, nil
	}
	header, err := ctx.FormFile("chart")
	if err != nil {
		return "", nil
	}
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (h *Handler) generateUniqueName(engine string, now time.Time) (string, error) {
	generate := h.GenerateName
	if generate == nil {
		generate = uid.Generate
	}
	for i := 0; i < maxNameAttempts; i++ {
		name, err := generate(engine, now)
		if err != nil {
			return "", err
		}
		exists := false
		if h.LevelNames != nil && h.LevelNames.HasLevel(name) {
			exists = true
		}
		if !exists {
			return name, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique level name after %d attempts", maxNameAttempts)
}

func readBGM(header *multipart.FileHeader) ([]byte, error) {
	if header.Size >= maxBGMSize {
		return nil, fileTooBigError(fmt.Errorf("bgm too big: %.1f MB", float64(header.Size)/1024.0/1024.0))
	}
	file, err := header.Open()
	if err != nil {
		return nil, bgmProcessError(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, bgmProcessError(err)
	}
	if len(data) < 261 {
		return nil, badBGMError(fmt.Errorf("bgm file is too small"))
	}
	head := data[:261]
	if !filetype.IsAudio(head) {
		kind, _ := filetype.Match(head)
		return nil, badBGMError(fmt.Errorf("bgm is %s (%s), not audio", kind.Extension, kind.MIME.Value))
	}
	return data, nil
}

func readCover(header *multipart.FileHeader) ([]byte, error) {
	file, err := header.Open()
	if err != nil {
		return nil, coverProcessError(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, coverProcessError(err)
	}
	if len(data) < 16 {
		return nil, badCoverError(fmt.Errorf("cover file is too small"))
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, badCoverError(fmt.Errorf("cover is not a supported image: %w", err))
	}

	cover, err := encodeCoverPNG(img)
	if err != nil {
		return nil, coverProcessError(err)
	}
	cover, err = fitCoverPNG(cover, img)
	if err != nil {
		return nil, coverProcessError(err)
	}
	return cover, nil
}

func encodeCoverPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fitCoverPNG(data []byte, img image.Image) ([]byte, error) {
	if len(data) <= maxCoverSize {
		return data, nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	for len(data) > maxCoverSize {
		if width <= 1 && height <= 1 {
			return nil, fmt.Errorf("cover png remains too large after resizing: %d bytes", len(data))
		}

		scale := 0.9
		if len(data) > 0 {
			targetScale := 0.95 * math.Sqrt(float64(maxCoverSize)/float64(len(data)))
			if targetScale < scale {
				scale = targetScale
			}
		}
		if scale <= 0 || scale >= 1 {
			scale = 0.9
		}

		nextWidth := maxInt(1, int(float64(width)*scale))
		nextHeight := maxInt(1, int(float64(height)*scale))
		if nextWidth == width && width > 1 {
			nextWidth--
		}
		if nextHeight == height && height > 1 {
			nextHeight--
		}

		resized := image.NewRGBA(image.Rect(0, 0, nextWidth, nextHeight))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

		nextData, err := encodeCoverPNG(resized)
		if err != nil {
			return nil, err
		}
		data = nextData
		img = resized
		bounds = img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}
	return data, nil
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	return parsed, nil
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func parseTags(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}, nil
	}
	var raw []string
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return nil, fmt.Errorf("tags must be a JSON string array: %w", err)
	}
	tags := make([]string, 0, min(len(raw), maxTagCount))
	for _, tag := range raw {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tags = append(tags, truncateRunes(tag, maxTagRunes))
		if len(tags) == maxTagCount {
			break
		}
	}
	return tags, nil
}

func writeError(ctx *gin.Context, err error) {
	var uploadErr uploadError
	if errors.As(err, &uploadErr) {
		ctx.JSON(uploadErr.status, gin.H{
			"code":        uploadErr.code,
			"description": uploadErr.description,
			"detail":      uploadErr.Error(),
		})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"code":        500,
		"description": "internal server error",
		"detail":      err.Error(),
	})
}
