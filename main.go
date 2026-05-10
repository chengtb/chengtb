package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"
)

type kvFlag []string

func (k *kvFlag) String() string {
	return strings.Join(*k, ",")
}

func (k *kvFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("param cannot be empty")
	}
	*k = append(*k, value)
	return nil
}

type inputConfig struct {
	imagePath string
	outputPDF string
	rawParams []string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := inputConfig{}
	var params kvFlag
	flag.StringVar(&cfg.imagePath, "image", "", "template image path (required)")
	flag.StringVar(&cfg.outputPDF, "output", "", "output pdf path (required)")
	flag.Var(&params, "param", "required parameter in key=value format, can be repeated")
	flag.Parse()

	cfg.rawParams = params
	if err := validateInput(cfg); err != nil {
		return err
	}

	parsed, err := parseParams(cfg.rawParams)
	if err != nil {
		return err
	}

	return generatePDF(cfg.imagePath, cfg.outputPDF, parsed)
}

func validateInput(cfg inputConfig) error {
	if strings.TrimSpace(cfg.imagePath) == "" {
		return errors.New("missing required -image")
	}
	if strings.TrimSpace(cfg.outputPDF) == "" {
		return errors.New("missing required -output")
	}
	if len(cfg.rawParams) == 0 {
		return errors.New("missing required -param (key=value), provide at least one")
	}
	if _, err := os.Stat(cfg.imagePath); err != nil {
		return fmt.Errorf("cannot access image %q: %w", cfg.imagePath, err)
	}
	return nil
}

func parseParams(raw []string) (map[string]string, error) {
	result := make(map[string]string, len(raw))
	for _, item := range raw {
		idx := strings.Index(item, "=")
		if idx <= 0 || idx == len(item)-1 {
			return nil, fmt.Errorf("invalid param %q, expected key=value", item)
		}

		key := strings.TrimSpace(item[:idx])
		val := strings.TrimSpace(item[idx+1:])
		if key == "" || val == "" {
			return nil, fmt.Errorf("invalid param %q, key/value cannot be empty", item)
		}
		result[key] = val
	}
	return result, nil
}

func generatePDF(imagePath, outputPath string, params map[string]string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	imgOpts := fpdf.ImageOptions{ReadDpi: true}
	info := pdf.RegisterImageOptions(imagePath, imgOpts)
	if info == nil {
		return fmt.Errorf("failed to read image file %q", imagePath)
	}

	pageW, pageH := pdf.GetPageSize()
	imgW, imgH := info.Extent()
	if imgW <= 0 || imgH <= 0 {
		return fmt.Errorf("invalid image size for %q", imagePath)
	}

	scale := min(pageW/imgW, pageH/imgH)
	drawW := imgW * scale
	drawH := imgH * scale
	offsetX := (pageW - drawW) / 2
	offsetY := (pageH - drawH) / 2
	pdf.ImageOptions(imagePath, offsetX, offsetY, drawW, drawH, false, imgOpts, 0, "")

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	const (
		boxX      = 10.0
		boxY      = 10.0
		lineH     = 6.0
		boxWidth  = 120.0
		boxPadX   = 4.0
		boxPadY   = 3.0
		titleSize = 12.0
		textSize  = 10.0
	)
	boxHeight := boxPadY*2 + lineH + float64(len(keys))*lineH
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(boxX, boxY, boxWidth, boxHeight, "F")

	pdf.SetXY(boxX+boxPadX, boxY+boxPadY)
	pdf.SetFont("Helvetica", "B", titleSize)
	pdf.CellFormat(boxWidth-boxPadX*2, lineH, "Parameters", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", textSize)
	for _, k := range keys {
		line := fmt.Sprintf("%s: %s", k, params[k])
		pdf.CellFormat(boxWidth-boxPadX*2, lineH, line, "", 1, "", false, 0, "")
	}

	outDir := filepath.Dir(outputPath)
	if outDir != "." && outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outDir, err)
		}
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("failed to write output pdf %q: %w", outputPath, err)
	}
	return nil
}
