package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

const (
	pageWidthMM   = 297.0
	pageHeightMM  = 210.0
	rightPanelMM  = 64.0
	maxImageBytes = 20 << 20
	httpTimeout   = 25 * time.Second
)

type RecipeCardInput struct {
	Title               string              `json:"title"`
	RecipeNo            string              `json:"recipe_no"`
	DevelopmentDate     string              `json:"development_date"`
	Machine             string              `json:"machine"`
	Description         string              `json:"description"`
	PreparationSteps    []string            `json:"preparation_steps"`
	Compartments        map[string][]string `json:"compartments"`
	IngredientsImageURL string              `json:"ingredients_image_url"`
	DishImageURL        string              `json:"dish_image_url"`
}

func main() {
	inputPath := flag.String("input", "", "Path to input JSON")
	outputPath := flag.String("output", "recipe-card.pdf", "Output PDF path")
	flag.Parse()

	if *inputPath == "" {
		exitWithError(errors.New("-input is required"))
	}

	in, err := loadInput(*inputPath)
	if err != nil {
		exitWithError(err)
	}

	if err := validateInput(&in); err != nil {
		exitWithError(err)
	}

	if err := generatePDF(context.Background(), in, *outputPath); err != nil {
		exitWithError(err)
	}

	fmt.Printf("PDF generated: %s\n", *outputPath)
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func loadInput(path string) (RecipeCardInput, error) {
	f, err := os.Open(path)
	if err != nil {
		return RecipeCardInput{}, fmt.Errorf("open input file: %w", err)
	}
	defer f.Close()

	var in RecipeCardInput
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return RecipeCardInput{}, fmt.Errorf("decode input json: %w", err)
	}
	return in, nil
}

func validateInput(in *RecipeCardInput) error {
	var missing []string
	if strings.TrimSpace(in.Title) == "" {
		missing = append(missing, "title")
	}
	if strings.TrimSpace(in.RecipeNo) == "" {
		missing = append(missing, "recipe_no")
	}
	if strings.TrimSpace(in.DevelopmentDate) == "" {
		missing = append(missing, "development_date")
	}
	if strings.TrimSpace(in.Machine) == "" {
		missing = append(missing, "machine")
	}
	if strings.TrimSpace(in.Description) == "" {
		missing = append(missing, "description")
	}
	if strings.TrimSpace(in.IngredientsImageURL) == "" {
		missing = append(missing, "ingredients_image_url")
	}
	if strings.TrimSpace(in.DishImageURL) == "" {
		missing = append(missing, "dish_image_url")
	}
	if len(in.PreparationSteps) == 0 {
		missing = append(missing, "preparation_steps")
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	normalizedCompartments := make(map[string][]string, 4)
	for k, v := range in.Compartments {
		label := strings.ToUpper(strings.TrimSpace(k))
		if label == "" {
			continue
		}
		var cleaned []string
		for _, line := range v {
			line = strings.TrimSpace(line)
			if line != "" {
				cleaned = append(cleaned, line)
			}
		}
		if len(cleaned) > 0 {
			normalizedCompartments[label] = cleaned
		}
	}

	for _, label := range []string{"A", "B", "C", "D"} {
		if len(normalizedCompartments[label]) == 0 {
			missing = append(missing, "compartments."+label)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	in.Compartments = normalizedCompartments

	if err := validateURL(in.IngredientsImageURL); err != nil {
		return fmt.Errorf("ingredients_image_url: %w", err)
	}
	if err := validateURL(in.DishImageURL); err != nil {
		return fmt.Errorf("dish_image_url: %w", err)
	}

	return nil
}

func validateURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http/https URLs are supported")
	}
	return nil
}

func generatePDF(ctx context.Context, in RecipeCardInput, outputPath string) error {
	client := &http.Client{Timeout: httpTimeout}
	ingredientsImg, ingredientsType, err := fetchImage(ctx, client, in.IngredientsImageURL)
	if err != nil {
		return fmt.Errorf("download ingredients image: %w", err)
	}
	dishImg, dishType, err := fetchImage(ctx, client, in.DishImageURL)
	if err != nil {
		return fmt.Errorf("download dish image: %w", err)
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		Size: gofpdf.SizeType{
			Wd: pageWidthMM,
			Ht: pageHeightMM,
		},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	drawLayout(pdf, in)

	if err := placeImage(pdf, "ingredients", ingredientsImg, ingredientsType, 5, 20, 34, 37); err != nil {
		return err
	}
	if err := placeImage(pdf, "dish", dishImg, dishType, pageWidthMM-rightPanelMM+2, 82, rightPanelMM-4, 118); err != nil {
		return err
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write pdf file: %w", err)
	}
	return nil
}

func fetchImage(ctx context.Context, client *http.Client, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	limited := io.LimitReader(resp.Body, maxImageBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > maxImageBytes {
		return nil, "", fmt.Errorf("image exceeds %d bytes", maxImageBytes)
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("unsupported image content: %w", err)
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return data, "JPG", nil
	case "png":
		return data, "PNG", nil
	case "gif":
		return data, "GIF", nil
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}
}

func placeImage(pdf *gofpdf.Fpdf, name string, data []byte, imgType string, x, y, boxW, boxH float64) error {
	opts := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	info := pdf.RegisterImageOptionsReader(name, opts, bytes.NewReader(data))
	if info == nil {
		return fmt.Errorf("register image %s failed", name)
	}

	imgW, imgH := info.Extent()
	if imgW <= 0 || imgH <= 0 {
		return fmt.Errorf("invalid dimensions for image %s", name)
	}
	scale := math.Min(boxW/imgW, boxH/imgH)
	if scale <= 0 {
		return fmt.Errorf("invalid target box for image %s", name)
	}
	renderW := imgW * scale
	renderH := imgH * scale

	px := x + (boxW-renderW)/2
	py := y + (boxH-renderH)/2
	pdf.ImageOptions(name, px, py, renderW, renderH, false, opts, 0, "")
	return nil
}

func drawLayout(pdf *gofpdf.Fpdf, in RecipeCardInput) {
	leftW := pageWidthMM - rightPanelMM

	pdf.SetFillColor(239, 239, 244)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	pdf.SetFillColor(190, 138, 23)
	pdf.Rect(leftW, 0, rightPanelMM, pageHeightMM, "F")

	pdf.SetFillColor(245, 245, 249)
	pdf.Rect(leftW+0.2, 46, rightPanelMM-0.2, 157, "F")

	pdf.SetTextColor(84, 46, 17)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(5, 4)
	pdf.CellFormat(leftW-75, 6, in.Title, "", 0, "L", false, 0, "")

	pdf.SetTextColor(67, 67, 67)
	pdf.SetFont("Arial", "", 7.5)
	pdf.SetXY(5, 10)
	pdf.CellFormat(65, 4.5, "Recipe No: "+in.RecipeNo, "", 0, "L", false, 0, "")

	pdf.SetXY(leftW-74, 4)
	pdf.CellFormat(69, 4.5, "Development Date: "+in.DevelopmentDate, "", 0, "R", false, 0, "")
	pdf.SetXY(leftW-74, 9)
	pdf.CellFormat(69, 4.5, "Machine: "+in.Machine, "", 0, "R", false, 0, "")

	pdf.SetTextColor(52, 52, 52)
	pdf.SetXY(38, 14)
	pdf.CellFormat(leftW-42, 4.5, "Load ingredients following arrow direction into compartment", "", 0, "C", false, 0, "")

	drawCompartments(pdf, in, leftW)
	drawPreparationSteps(pdf, in, leftW)
	drawRightPanelText(pdf, in, leftW)
}

func drawCompartments(pdf *gofpdf.Fpdf, in RecipeCardInput, leftW float64) {
	xGrid := 42.5
	yGrid := 18.2
	gap := 1.8
	colW := (leftW - xGrid - 5 - gap) / 2
	rowH := 18.5

	pdf.SetFillColor(219, 181, 106)
	pdf.Rect(5, 54.5, 34, 2.6, "F")

	types := []struct {
		Label string
		X     float64
		Y     float64
	}{
		{Label: "C", X: xGrid, Y: yGrid},
		{Label: "D", X: xGrid + colW + gap, Y: yGrid},
		{Label: "A", X: xGrid, Y: yGrid + rowH + gap},
		{Label: "B", X: xGrid + colW + gap, Y: yGrid + rowH + gap},
	}

	for _, c := range types {
		pdf.SetFillColor(247, 247, 247)
		pdf.SetDrawColor(233, 233, 233)
		pdf.RoundedRect(c.X, c.Y, colW, rowH, 1.4, "1234", "FD")

		pdf.SetTextColor(84, 46, 17)
		pdf.SetFont("Arial", "B", 11)
		pdf.SetXY(c.X+1.8, c.Y+1.3)
		pdf.CellFormat(10, 5, c.Label, "", 0, "L", false, 0, "")

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(34, 34, 34)
		pdf.SetXY(c.X+1.8, c.Y+6)
		lines := strings.Join(in.Compartments[c.Label], "\n")
		pdf.MultiCell(colW-3.6, 4.6, lines, "", "L", false)
	}
}

func drawPreparationSteps(pdf *gofpdf.Fpdf, in RecipeCardInput, leftW float64) {
	x := 5.0
	y := 58.5
	w := leftW - 10
	h := pageHeightMM - y - 7

	pdf.SetFillColor(247, 247, 247)
	pdf.SetDrawColor(233, 233, 233)
	pdf.RoundedRect(x, y, w, h, 1.4, "1234", "FD")

	pdf.SetFillColor(201, 148, 26)
	pdf.RoundedRect(x, y, w, 5.2, 1.2, "12", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9.2)
	pdf.SetXY(x+2.2, y+0.2)
	pdf.CellFormat(w-4.4, 4.6, "Preparation Steps:", "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	baseY := y + 7
	for i, step := range in.PreparationSteps {
		lineY := baseY + float64(i)*6
		if lineY+6 > y+h {
			break
		}
		pdf.SetTextColor(201, 148, 26)
		pdf.SetXY(x+2.2, lineY)
		pdf.CellFormat(5, 4.2, fmt.Sprintf("%d.", i+1), "", 0, "L", false, 0, "")

		pdf.SetTextColor(33, 33, 33)
		pdf.SetXY(x+7.4, lineY)
		pdf.MultiCell(w-9.6, 4.2, step, "", "L", false)
	}
}

func drawRightPanelText(pdf *gofpdf.Fpdf, in RecipeCardInput, leftW float64) {
	x := leftW + 3.5
	w := rightPanelMM - 7

	pdf.SetTextColor(20, 10, 0)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(x, 5)
	pdf.MultiCell(w, 6, in.Title, "", "L", false)

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetXY(x, 18)
	pdf.MultiCell(w, 5.2, in.Description, "", "L", false)
}
