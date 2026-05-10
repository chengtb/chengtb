package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

const (
	pageWidthMM  = 297.0
	pageHeightMM = 210.0
)

type BinIngredient struct {
	Ingredient string `json:"ingredient"`
	Amount     string `json:"amount"`
	Remark     string `json:"remark"`
}

type RecipeCardInput struct {
	RecipeName       string                   `json:"recipeName"`
	RecipeNumber     string                   `json:"recipeNumber"`
	DevelopmentTime  string                   `json:"developmentTime"`
	MachineType      string                   `json:"machineType"`
	Description      string                   `json:"description"`
	PreparationSteps []string                 `json:"preparationSteps"`
	Bins             map[string]BinIngredient `json:"bins"`
	RawMaterialImage string                   `json:"rawMaterialImage"`
	FinishedImage    string                   `json:"finishedImage"`
}

func main() {
	inputPath := flag.String("input", "", "absolute path of input recipe json")
	outputPath := flag.String("output", "", "absolute path of output pdf")
	fontPath := flag.String("font", "", "optional absolute path to a UTF-8 TTF font (recommended for Chinese text)")
	timeout := flag.Duration("timeout", 20*time.Second, "image download timeout")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		exitWithError(errors.New("both -input and -output are required"))
	}

	in, err := loadInput(*inputPath)
	if err != nil {
		exitWithError(fmt.Errorf("load input: %w", err))
	}
	if err := in.Validate(); err != nil {
		exitWithError(fmt.Errorf("invalid input: %w", err))
	}

	if err := generateRecipeCardPDF(in, *outputPath, *fontPath, *timeout); err != nil {
		exitWithError(fmt.Errorf("generate pdf: %w", err))
	}

	fmt.Printf("recipe card generated successfully: %s\n", *outputPath)
}

func loadInput(path string) (RecipeCardInput, error) {
	var in RecipeCardInput
	f, err := os.Open(path)
	if err != nil {
		return in, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&in); err != nil {
		return in, err
	}
	return in, nil
}

func (in RecipeCardInput) Validate() error {
	requiredTextFields := map[string]string{
		"recipeName":      in.RecipeName,
		"recipeNumber":    in.RecipeNumber,
		"developmentTime": in.DevelopmentTime,
		"machineType":     in.MachineType,
		"description":     in.Description,
	}
	for k, v := range requiredTextFields {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("missing required field: %s", k)
		}
	}
	if strings.TrimSpace(in.RawMaterialImage) == "" {
		return errors.New("missing required field: rawMaterialImage")
	}
	if strings.TrimSpace(in.FinishedImage) == "" {
		return errors.New("missing required field: finishedImage")
	}
	if len(in.PreparationSteps) == 0 {
		return errors.New("preparationSteps cannot be empty")
	}
	for _, bin := range []string{"A", "B", "C", "D"} {
		item, ok := in.Bins[bin]
		if !ok {
			return fmt.Errorf("missing required bins.%s", bin)
		}
		if strings.TrimSpace(item.Ingredient) == "" {
			return fmt.Errorf("missing required bins.%s.ingredient", bin)
		}
	}
	return nil
}

func generateRecipeCardPDF(in RecipeCardInput, outputPath, fontPath string, timeout time.Duration) error {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: pageWidthMM, Ht: pageHeightMM},
		OrientationStr: "L",
	})
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	fontFamily := "Arial"
	if fontPath != "" {
		fontPath = filepath.Clean(fontPath)
		if _, err := os.Stat(fontPath); err != nil {
			return fmt.Errorf("invalid font file: %w", err)
		}
		pdf.AddUTF8Font("custom", "", fontPath)
		if err := pdf.Err(); err != nil {
			return fmt.Errorf("load UTF-8 font: %w", err)
		}
		fontFamily = "custom"
	}

	if err := drawTemplate(pdf, in, fontFamily, timeout); err != nil {
		return err
	}
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return err
	}
	return nil
}

func drawTemplate(pdf *gofpdf.Fpdf, in RecipeCardInput, fontFamily string, timeout time.Duration) error {
	leftMargin := 8.0
	topMargin := 8.0
	contentW := pageWidthMM - 16.0
	contentH := pageHeightMM - 16.0

	pdf.SetLineWidth(0.3)
	pdf.Rect(leftMargin, topMargin, contentW, contentH, "D")

	pdf.SetFont(fontFamily, "B", 18)
	pdf.SetXY(leftMargin+3, topMargin+3)
	pdf.CellFormat(contentW-6, 10, "Recipe Card", "", 0, "C", false, 0, "")

	metaTop := topMargin + 15
	metaH := 18.0
	pdf.Rect(leftMargin+2, metaTop, contentW-4, metaH, "D")
	colW := (contentW - 4) / 4
	for i := 1; i < 4; i++ {
		x := leftMargin + 2 + float64(i)*colW
		pdf.Line(x, metaTop, x, metaTop+metaH)
	}
	pdf.SetFont(fontFamily, "", 10)
	drawLabelValue(pdf, leftMargin+2, metaTop, colW, "Name", in.RecipeName)
	drawLabelValue(pdf, leftMargin+2+colW, metaTop, colW, "No.", in.RecipeNumber)
	drawLabelValue(pdf, leftMargin+2+2*colW, metaTop, colW, "R&D Time", in.DevelopmentTime)
	drawLabelValue(pdf, leftMargin+2+3*colW, metaTop, colW, "Machine", in.MachineType)

	upperTop := metaTop + metaH + 2
	leftW := contentW * 0.58
	rightW := contentW - leftW - 4

	descH := 28.0
	pdf.Rect(leftMargin+2, upperTop, leftW, descH, "D")
	pdf.SetFont(fontFamily, "B", 11)
	pdf.SetXY(leftMargin+4, upperTop+2)
	pdf.CellFormat(leftW-4, 6, "Description", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetXY(leftMargin+4, upperTop+8)
	pdf.MultiCell(leftW-6, 5, in.Description, "", "L", false)

	stepsTop := upperTop + descH + 2
	stepsH := contentH - (stepsTop-topMargin) - 2
	pdf.Rect(leftMargin+2, stepsTop, leftW, stepsH, "D")
	pdf.SetFont(fontFamily, "B", 11)
	pdf.SetXY(leftMargin+4, stepsTop+2)
	pdf.CellFormat(leftW-4, 6, "Preparation Steps", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetXY(leftMargin+4, stepsTop+8)
	pdf.MultiCell(leftW-6, 5, formatSteps(in.PreparationSteps), "", "L", false)

	rightLeft := leftMargin + 2 + leftW + 2
	binH := 58.0
	pdf.Rect(rightLeft, upperTop, rightW, binH, "D")
	pdf.SetFont(fontFamily, "B", 11)
	pdf.SetXY(rightLeft+2, upperTop+2)
	pdf.CellFormat(rightW-4, 6, "Ingredients (A/B/C/D)", "", 0, "L", false, 0, "")

	tableTop := upperTop + 10
	pdf.SetFont(fontFamily, "", 9)
	pdf.Line(rightLeft+9, tableTop, rightLeft+9, upperTop+binH)
	pdf.Line(rightLeft+rightW*0.48, tableTop, rightLeft+rightW*0.48, upperTop+binH)
	pdf.Line(rightLeft+rightW*0.70, tableTop, rightLeft+rightW*0.70, upperTop+binH)
	pdf.Line(rightLeft, tableTop, rightLeft+rightW, tableTop)
	rowH := (binH - 10) / 4
	for i := 1; i < 4; i++ {
		y := tableTop + rowH*float64(i)
		pdf.Line(rightLeft, y, rightLeft+rightW, y)
	}

	for i, key := range []string{"A", "B", "C", "D"} {
		item := in.Bins[key]
		y := tableTop + rowH*float64(i)
		pdf.SetXY(rightLeft+1, y+2)
		pdf.CellFormat(7, 5, key, "", 0, "C", false, 0, "")
		pdf.SetXY(rightLeft+10, y+1.5)
		pdf.MultiCell(rightW*0.48-12, 4, safeText(item.Ingredient), "", "L", false)
		pdf.SetXY(rightLeft+rightW*0.48+1, y+1.5)
		pdf.MultiCell(rightW*0.22-2, 4, safeText(item.Amount), "", "L", false)
		pdf.SetXY(rightLeft+rightW*0.70+1, y+1.5)
		pdf.MultiCell(rightW*0.30-2, 4, safeText(item.Remark), "", "L", false)
	}

	imageTop := upperTop + binH + 2
	imageH := contentH - (imageTop-topMargin) - 2
	imageGap := 2.0
	imageBoxW := (rightW - imageGap) / 2

	pdf.Rect(rightLeft, imageTop, imageBoxW, imageH, "D")
	pdf.Rect(rightLeft+imageBoxW+imageGap, imageTop, imageBoxW, imageH, "D")

	pdf.SetFont(fontFamily, "B", 10)
	pdf.SetXY(rightLeft+1.5, imageTop+1.5)
	pdf.CellFormat(imageBoxW-3, 5, "Raw Material", "", 0, "L", false, 0, "")
	pdf.SetXY(rightLeft+imageBoxW+imageGap+1.5, imageTop+1.5)
	pdf.CellFormat(imageBoxW-3, 5, "Finished Dish", "", 0, "L", false, 0, "")

	if err := drawImageFromSource(pdf, "raw-material-image", in.RawMaterialImage, rightLeft+1, imageTop+8, imageBoxW-2, imageH-9, timeout); err != nil {
		return fmt.Errorf("draw rawMaterialImage: %w", err)
	}
	if err := drawImageFromSource(pdf, "finished-image", in.FinishedImage, rightLeft+imageBoxW+imageGap+1, imageTop+8, imageBoxW-2, imageH-9, timeout); err != nil {
		return fmt.Errorf("draw finishedImage: %w", err)
	}

	if err := pdf.Err(); err != nil {
		return err
	}
	return nil
}

func drawLabelValue(pdf *gofpdf.Fpdf, x, y, w float64, label, value string) {
	pdf.SetFontStyle("B")
	pdf.SetXY(x+2, y+2)
	pdf.CellFormat(w-4, 5, label, "", 0, "L", false, 0, "")
	pdf.SetFontStyle("")
	pdf.SetXY(x+2, y+8)
	pdf.MultiCell(w-4, 4, safeText(value), "", "L", false)
}

func drawImageFromSource(pdf *gofpdf.Fpdf, name, source string, x, y, maxW, maxH float64, timeout time.Duration) error {
	content, imageType, err := readImageContent(source, timeout)
	if err != nil {
		return err
	}

	pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{
		ImageType: strings.ToUpper(imageType),
		ReadDpi:   true,
	}, bytes.NewReader(content))
	if err := pdf.Err(); err != nil {
		return err
	}

	info := pdf.GetImageInfo(name)
	if info == nil {
		return errors.New("failed to parse image info")
	}
	w := info.Extent().Wd
	h := info.Extent().Ht
	if w <= 0 || h <= 0 {
		return errors.New("invalid image dimensions")
	}

	scale := min(maxW/w, maxH/h)
	drawW := w * scale
	drawH := h * scale
	centerX := x + (maxW-drawW)/2
	centerY := y + (maxH-drawH)/2

	pdf.ImageOptions(name, centerX, centerY, drawW, drawH, false, gofpdf.ImageOptions{
		ImageType: strings.ToUpper(imageType),
		ReadDpi:   true,
	}, 0, "")
	return nil
}

func readImageContent(source string, timeout time.Duration) ([]byte, string, error) {
	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: timeout}
		resp, reqErr := client.Get(source)
		if reqErr != nil {
			return nil, "", reqErr
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return nil, "", err
		}
	}

	imgCfg, imageType, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image config failed: %w", err)
	}
	if imgCfg.Width <= 0 || imgCfg.Height <= 0 {
		return nil, "", errors.New("image has invalid size")
	}
	if imageType == "jpeg" {
		imageType = "jpg"
	}
	return data, imageType, nil
}

func formatSteps(steps []string) string {
	var b strings.Builder
	count := 0
	for _, step := range steps {
		trimmed := strings.TrimSpace(step)
		if trimmed == "" {
			continue
		}
		count++
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%d. %s", count, trimmed)
	}
	return b.String()
}

func safeText(s string) string {
	return strings.TrimSpace(s)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
