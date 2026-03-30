package processor

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"WBTech_L3.4/internal/storage"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/wb-go/wbf/zlog"
)

type result struct {
	watermarkPath string
	thumbPath     string
	err           error
}

type Job struct {
	OriginalPath string
	ResultChan   chan result
}

type Processor struct {
	storage       storage.Storage
	watermarkText string
	watermarkPath string
	jobs          chan Job
	logger        zlog.Zerolog
}

func NewProcessor(storage storage.Storage, watermarkPath, watermarkText string, workerCount int, logger zlog.Zerolog) *Processor {
	p := &Processor{
		storage:       storage,
		watermarkPath: watermarkPath,
		watermarkText: watermarkText,
		jobs:          make(chan Job, 100),
		logger:        logger,
	}

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return p
}

func (p *Processor) Process(ctx context.Context, originalPath string) (string, string, error) {
	resChan := make(chan result, 1)
	job := Job{OriginalPath: originalPath, ResultChan: resChan}

	select {
	case p.jobs <- job:
		res := <-resChan
		return res.watermarkPath, res.thumbPath, res.err
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
}

func (p *Processor) worker() {
	for job := range p.jobs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					p.logger.Error().Msgf("error: worker failed with panic: %v", r)
				}
			}()

			watermark, thumb, err := p.processLogic(job.OriginalPath)

			if err != nil {
				p.logger.Error().Err(err).Str("path", job.OriginalPath).Msg("error processing file")
			}

			job.ResultChan <- result{watermarkPath: watermark, thumbPath: thumb, err: err}
		}()
	}
}

func (p *Processor) processLogic(originalPath string) (string, string, error) {
	file, err := os.Open(originalPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot open file: %w", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			p.logger.Error().Err(err).Msg("cannot close file")
		}
	}(file)

	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("decoding error (unsupported format): %w", err)
	}

	var mainImg image.Image
	mainImg = imaging.Resize(img, 1024, 0, imaging.Lanczos)

	if p.watermarkText != "" {
		mainImg = p.addTextWatermark(mainImg, p.watermarkPath, p.watermarkText)
	}

	thumbImg := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)

	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	mainFilename := nameWithoutExt + "_processed" + ext
	watermarkPath, err := p.saveToStorage(mainFilename, mainImg, format)
	if err != nil {
		return "", "", fmt.Errorf("error while saving file path: %w", err)
	}

	thumbFilename := nameWithoutExt + "_thumb" + ext
	_, err = p.saveToStorage(thumbFilename, thumbImg, format)
	if err != nil {
		p.logger.Warn().Err(err).Str("path", thumbFilename).Msg("warning: thumb file was not saved")
	}

	return watermarkPath, thumbFilename, nil
}

func (p *Processor) addTextWatermark(base image.Image, path, text string) image.Image {
	bounds := base.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	dc := gg.NewContextForImage(base)

	dc.SetRGBA(1, 1, 1, 0.4)

	fontSize := w / 4
	if err := dc.LoadFontFace(path, fontSize); err != nil {
		p.logger.Warn().Msg("font not found, using default")
	}

	dc.Push()
	dc.Translate(w/2, h/2)

	dc.Rotate(gg.Radians(-45))

	dc.DrawStringAnchored(text, 0, 0, 0.5, 0.5)
	dc.Pop()

	return dc.Image()
}

func (p *Processor) saveToStorage(filename string, img image.Image, format string) (string, error) {
	buf := new(bytes.Buffer)
	var err error

	switch format {
	case "jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 100})
	case "png":
		err = png.Encode(buf, img)
	default:
		return "", fmt.Errorf("unsupported format: %s. Only JPG/PNG are allowed", format)
	}

	if err != nil {
		return "", err
	}
	return p.storage.Save(filename, buf)
}
