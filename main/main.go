package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/golang/freetype"

	githubColors "github.com/Jomy10/github-colors"
	githubApi "jonaseveraert.be/gh-api-go"
	jimage "jonaseveraert.be/gh-readme/image"
)

type LangPair struct {
	Lang string
	Perc float64
}

func main() {
	// Init image
	img := getImageRGBA("res/base_image.png")

	fontCtx := setUpFont()

	drawLanguages(img, fontCtx)

	drawFrameworks(img, fontCtx)

	drawContributions(img, fontCtx)

	// Export image
	jimage.ExportImage(img, "output.png")
}

func setUpFont() *freetype.Context {
	fontBytes, err := ioutil.ReadFile("/Users/jonaseveraert/Library/Fonts/FiraCode-Bold.ttf") // If you use this program, don't forget to replace this
	if err != nil {
		panic(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}

	ctx := freetype.NewContext()
	ctx.SetDPI(72) // DPI of the base image
	ctx.SetFont(font)

	return ctx
}

func drawLanguages(img *image.RGBA, fontCtx *freetype.Context) {
	// Draw title
	jimage.AddLabel(img, fontCtx, 1150, 25, "The languages I use", 64, color.White)
	jimage.AddLabel(img, fontCtx, 1206, 120, "(ordered according to usage in my repositories)", 24, color.White)

	// Get languages
	// This line is for debugging
	// languagesCount := map[string]uint{"Rust": 1000, "Swift": 500, "Java": 200, "JavaScript": 100, "HTML": 50, "CSS": 10, "Go": 1, "WebAssembly": 400, "Python": 65, "Shell": 457, "SCSS": 67}
	languagesCount := githubApi.FetchUserLanguages("jomy10", readToken(), false, true, []string{}, []string{}, []string{}, []string{"Jomy10/website"})
	// Add additional repos
	// Filter some languages
	delete(languagesCount, "Procfile")
	delete(languagesCount, "Svelte")

	// Count total to calc percentage
	langs := make([]LangPair, len(languagesCount))
	var total uint = 0
	for _, count := range languagesCount {
		total += count
	}

	// Sort languages by percentage into langs
	for lang, count := range languagesCount {
		perc := float64(count) / float64(total)
		sorted := false
		for i := 0; i < len(languagesCount); i++ {
			if langs[i].Perc > perc {
				continue
			} else {
				var pre, post []LangPair

				if i == 0 {
					pre = []LangPair{}
				} else {
					pre = langs[0:i]
				}

				post = langs[i:len(languagesCount)]

				// Needs to be in this order, otherwise the value post refers to changes
				langs = append([]LangPair{{lang, perc}}, post...)
				langs = append(pre, langs...)

				sorted = true
				break
			}
		}
		if !sorted {
			langs = append(langs, LangPair{lang, perc})
		}
	}

	// Get language colors
	colors := githubColors.GetGithubColors()

	// Draw languages
	const maxWidth int = 1000 // max width of a bar
	const height int = 55     // height of a bar
	const marginTop int = 200
	const marginBetween int = 50
	const logoSize int = 90
	const marginBarLogo = 25 // margin between the bar and the logo
	const maxLangs = 7       // 8 languages will be displayed, the rest will be put together in one single bar

	i := 0
	for _, pair := range langs {
		lang := pair.Lang
		perc := pair.Perc
		if lang == "" {
			continue
		}
		r, g, b, err := hexToRGB(colors[lang].Color)
		if err != nil {
			fmt.Println(lang, "does not have a color:", colors[lang])
			r, g, b = 0, 0, 0
		}

		var logoImg image.Image
		if i >= maxLangs {
			// other languages will be put together
			perc = 0
			r, g, b = 255, 255, 255

			var _langs []string
			for j := i; j < len(langs); j++ {
				_lang := langs[j].Lang
				if _lang != "" {
					perc += langs[j].Perc
					_langs = append(_langs, _lang)
				}
			}

			// Combine logos into 1 image
			// TODO: #2 add logo width and height. height will be = logoSize while width will be changed to accomodate
			//	     more languages so that there are only 2 rows of languages
			logoImg = getMultiLangImage(_langs, logoSize)

			jimage.ExportImage(logoImg, "test.png")
		} else {
			logoImg = getLangLogo(lang)
		}
		// Calculate bar position
		barX := img.Bounds().Max.X - int(float64(maxWidth)*perc)
		barY := marginTop + i*(height+marginBetween)

		// Draw logo
		jimage.DrawImage(img, logoImg, barX-marginBarLogo-logoSize, barY+(height-logoSize)/2)

		// Draw bar
		jimage.DrawRect(
			img,
			barX,
			barY,
			img.Bounds().Max.X,
			height,
			color.RGBA{r, g, b, 255},
		)

		if i >= maxLangs {
			break
		}

		i++
	}
}

// An icon (90 x 90) containing multiple logos
func getMultiLangImage(langs []string, imgSize int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	logoSize := imgSize / (len(langs) / 2)
	logosPerRow := imgSize / logoSize
	row := 0
	logosOnRow := 0 // logos currently already drawn on the row
	for i, lang := range langs {
		og := getLangLogo(lang)
		newImg := jimage.ResizeImage(og, logoSize, logoSize)

		var x, y int
		x = (i % logosPerRow) * logoSize
		y = row * logoSize

		jimage.DrawImage(dst, newImg, x, y)

		logosOnRow++
		if logosOnRow == logosPerRow {
			row++
			logosOnRow = 0
		}
	}

	jimage.ExportImage(dst, "test.png")

	return dst
}

// Draw frameworks and other tools
func drawFrameworks(img *image.RGBA, fontCtx *freetype.Context) {
	jimage.AddLabel(img, fontCtx, 50, 25, "The tools and", 64, color.White)
	jimage.AddLabel(img, fontCtx, 50, 100, "frameworks I use", 64, color.White)

	// Draw images from `frameworks` folder
	files, err := ioutil.ReadDir("res/frameworks")
	if err != nil {
		panic(err)
	}

	amtOfFramworks := len(files)
	const maxFrameworkRows int = 4

	framworksPerRow := int(math.Ceil(math.Max(float64(amtOfFramworks)/4, 2)))

	row := 0
	i := 0
	const logoSize int = 90
	const marginTop int = 200
	const marginLeft int = 75
	const marginBetween int = 15

	for _, file := range files {
		if !strings.Contains(file.Name(), "png") {
			// ignore non-image files (like .DS_Store)
			continue
		}
		filePath := "res/frameworks/" + file.Name()
		image := getImageRGBA(filePath)

		jimage.DrawImage(img, image, marginLeft+i*(logoSize+marginBetween), marginTop+row*(logoSize+marginBetween))

		i++

		if i >= framworksPerRow {
			i = 0
			row++
		}
	}
}

func drawContributions(img *image.RGBA, fontCtx *freetype.Context) {
	// fetch repos
	// TODO: #1 order repositories by date contributed to
	repos := githubApi.FetchReposContributedTo(readToken())

	// Draw title
	const titleMarginTop int = 550
	jimage.AddLabel(img, fontCtx, 50, titleMarginTop, "Projects I", 64, color.White)
	jimage.AddLabel(img, fontCtx, 50, titleMarginTop+75, "contributed to", 64, color.White)

	// Draw repos
	const marginTop int = titleMarginTop + 75 + 100
	const marginTopToBotom int = 35 // Size take by text and their bottom padding
	i := 0
	for _, repo := range repos {
		jimage.AddLabel(img, fontCtx, 75, marginTop+i*marginTopToBotom, repo, 24, color.White)
		i++
	}
}

// Get image from path
func getImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return img
}

// Get image from path as `*image.RGBA`
func getImageRGBA(path string) *image.RGBA {
	img := getImage(path)

	image := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	jimage.DrawImage(image, img, 0, 0)

	return image
}

// Get a logo from a language stored in `res/langs/*.png`
func getLangLogo(lang string) image.Image {
	return getImage(fmt.Sprintf("res/langs/%s.png", lang))
}

// Convert hexadecimal color string to RGBA values
func hexToRGB(hexa string) (uint8, uint8, uint8, error) {
	if hexa == "" {
		return 0, 0, 0, errors.New("emtpy color")
	}
	regex := regexp.MustCompile(`(?i)#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})`)
	parsed := regex.FindAllStringSubmatch(hexa, -1)[0]
	// return (parsed[1],parsed[2],parsed[3])
	r, err := hex.DecodeString(parsed[1])
	if err != nil {
		fmt.Println("Error decoding string")
		panic(err)
	}
	g, err := hex.DecodeString(parsed[2])
	if err != nil {
		fmt.Println("Error decoding string")
		panic(err)
	}
	b, err := hex.DecodeString(parsed[3])
	if err != nil {
		fmt.Println("Error decoding string")
		panic(err)
	}
	// r := uint8(rHex)
	return r[0], g[0], b[0], nil
}

// Read the GitHub API token
func readToken() string {
	tokenBytes, err := os.ReadFile("../gh-api/GH_TOKEN")
	if err != nil {
		panic(err)
	}

	return strings.Trim(string(tokenBytes), "\n")
}
