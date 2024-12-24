package goidagif

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

const (
	GIPHY_API_KEY      = "qNkyLYgSNm4fwmkGMI5ucBAE6iUS1lUW"
	GIPHY_TRENDING_URL = "https://api.giphy.com/v1/gifs/trending"
)

// GiphyResponse represents the structure of the Giphy API response
type GiphyResponse struct {
	Data []struct {
		Images struct {
			Original struct {
				URL string `json:"url"`
			} `json:"original"`
		} `json:"images"`
	} `json:"data"`
}

// FetchTrendingGIF fetches a trending GIF from Giphy API
func FetchTrendingGIF() (*gif.GIF, error) {
	url := fmt.Sprintf("%s?api_key=%s&limit=1&offset=%v", GIPHY_TRENDING_URL, GIPHY_API_KEY, rand.Intn(499))
	fmt.Println("Requesting URL:", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Go-http-client/1.1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	fmt.Printf("Response status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, string(body))
	}

	var giphyResp GiphyResponse
	if err := json.NewDecoder(resp.Body).Decode(&giphyResp); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	if len(giphyResp.Data) == 0 {
		return nil, fmt.Errorf("no GIFs found in the response")
	}

	gifURL := giphyResp.Data[0].Images.Original.URL
	fmt.Println("GIF URL:", gifURL)

	gifResp, err := http.Get(gifURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching GIF data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error closing body: %v", err)
		}
	}(gifResp.Body)

	gifData, err := gif.DecodeAll(gifResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding GIF: %w", err)
	}

	fmt.Println("Successfully fetched and decoded GIF")
	return gifData, nil
}

func AddTextToFrame(frame image.Image, text string) image.Image {
	fmt.Println("Adding text to frame with shadow and word wrapping:", text)
	bounds := frame.Bounds()
	dc := gg.NewContext(bounds.Dx(), bounds.Dy())
	dc.DrawImage(frame, 0, 0)

	fontPath := "assets/fonts/DejaVuSans-Bold.ttf"
	fontSize := 70
	if err := dc.LoadFontFace(fontPath, float64(fontSize)); err != nil {
		fmt.Printf("Error loading font: %v\n", err)
		return frame
	}

	// Define text box dimensions and position
	margin := 20
	maxWidth := float64(bounds.Dx() - 2*margin)
	x := float64(bounds.Dx() / 2)
	y := float64(bounds.Dy() - 80)

	// Wrap text if it exceeds maxWidth
	wrappedText := wrapText(dc, text, maxWidth)

	// Draw shadow
	dc.SetRGBA(0, 0, 0, 0.5) // Black with 50% opacity
	dc.DrawStringWrapped(wrappedText, x+2, y+2, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	// Draw main text
	dc.SetRGB(1, 1, 1) // White color
	dc.DrawStringWrapped(wrappedText, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	fmt.Println("Text with shadow and wrapping added successfully")
	return dc.Image()
}

// wrapText breaks text into lines that fit within the given width
func wrapText(dc *gg.Context, text string, maxWidth float64) string {
	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine + " " + word
		if currentLine == "" {
			testLine = word
		}
		measureStringWeight, _ := dc.MeasureString(testLine)
		if measureStringWeight <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// GenerateTextVariant generates a random variation of the text
func GenerateTextVariant(base string) string {
	phrases := []string{
		"Братья! %s",
		"Друзья! %s",
		"%s навсегда!",
		"%s — это сила!",
		"Вперёд, %s!",
		"%s! %s! %s!",
	}
	keywords := []string{"гойда", "гол", "слон", "наш", "медведь"}
	rand.Seed(time.Now().UnixNano())
	base = keywords[rand.Intn(len(keywords))]
	phrase := phrases[rand.Intn(len(phrases))]

	result := fmt.Sprintf(phrase, strings.ToUpper(base))
	fmt.Println("Generated text variant:", result)
	return result
}

// ProcessGIF adds text overlays to all frames and saves a new GIF
func ProcessGIF(gifData *gif.GIF, text, outputPath string) error {
	fmt.Println("Processing GIF with text:", text)
	for i, frame := range gifData.Image {
		fmt.Printf("Processing frame %d\n", i)
		img := AddTextToFrame(frame, text)
		var buf bytes.Buffer
		png.Encode(&buf, img)
		decodedFrame, _ := png.Decode(&buf)
		draw.Draw(frame, frame.Bounds(), decodedFrame, image.Point{}, draw.Over)
	}

	fmt.Println("Saving processed GIF to:", outputPath)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, gifData); err != nil {
		return fmt.Errorf("error encoding GIF: %w", err)
	}

	fmt.Println("GIF saved successfully")
	return nil
}

// GenerateGIF generates a GIF with text overlay
func GenerateGIF(outputPath, text string) error {
	fmt.Println("Starting GIF generation")
	gifData, err := FetchTrendingGIF()
	if err != nil {
		return fmt.Errorf("error fetching GIF: %w", err)
	}

	textVariant := GenerateTextVariant(text)
	if err := ProcessGIF(gifData, textVariant, outputPath); err != nil {
		return fmt.Errorf("error processing GIF: %w", err)
	}

	fmt.Println("GIF generation completed successfully")
	return nil
}
