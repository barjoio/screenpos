package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"

	"github.com/barjoco/utils/colour"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	gridImg               *ebiten.Image
	errRegularTermination = errors.New("regular termination")
	cfg                   config
	mplusFont             font.Face
	keyCol                string   // selected key column
	w, h                  int      // width, height
	wCell, hCell          int      // width, height, of cells
	gridSize              int      // x.y grid size
	gx, gy                = 0, 0   // grid spawn coordinates
	cx, cy                = -1, -1 // screen coordinates
	exit                  bool
)

// keys used to select columns/rows
var keys = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// configurable options
type config struct {
	ArrowColour    string `json:"arrowColour,omitempty"`
	ArrowSize      int    `json:"arrowSize,omitempty"`
	ArrowOpacity   int    `json:"arrowOpacity,omitempty"`
	FontColour     string `json:"fontColour,omitempty"`
	FontDropColour string `json:"fontDropColour,omitempty"`
	FontOpacity    int    `json:"fontOpacity,omitempty"`
	LineColour     string `json:"lineColour,omitempty"`
	LineDropColour string `json:"lineDropColour,omitempty"`
	LineOpacity    int    `json:"lineOpacity,omitempty"`
	GridStep       int    `json:"gridStep,omitempty"`
}

type render struct {
	frames int
}

func init() {
	w, h = ebiten.ScreenSizeInFullscreen()

	// unmarshal default or custom config
	if len(os.Args) >= 3 && (os.Args[1] == "-c" || os.Args[1] == "--config") {
		customConfig, err := ioutil.ReadFile(os.Args[2])
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(customConfig), &cfg)
		if err != nil {
			panic(err)
		}
	} else {
		cfg = defaultConfig
	}

	gridSize = len(keys)
	wCell, hCell = (w+w%gridSize)/gridSize, (h+h%gridSize)/gridSize
}

// initialise font
func initFont() {
	f, _ := base64.StdEncoding.DecodeString(InterRegular)
	tt, err := opentype.Parse(f)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	mplusFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12 * ebiten.DeviceScaleFactor(),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// convert hex and opacity, to rgba colour
func getColour(hex string, opacity int) color.RGBA {
	r, g, b := colour.HexToRGB(hex)
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(opacity)}
}

// generate new grid image
func makeGrid() {
	grid := image.NewRGBA(image.Rectangle{image.Point{cfg.GridStep * gx, cfg.GridStep * gy}, image.Point{w, h}})

	arrowOpacity := cfg.ArrowOpacity
	if cx != -1 {
		arrowOpacity = 0
	}

	// traverse x axis, drawing lines along y at intervals of cell width
	for x := 0; x < gridSize; x++ {
		for y := 0; y < h; y++ {
			grid.Set(x*wCell, y, getColour(cfg.LineDropColour, cfg.LineOpacity))
			grid.Set(x*wCell-1, y, getColour(cfg.LineColour, cfg.LineOpacity))
			if y%hCell < cfg.ArrowSize {
				if keyCol == keys[x] {
					arrowOpacity = cfg.ArrowOpacity
				}
				grid.Set(x*wCell, y, getColour(cfg.ArrowColour, arrowOpacity))
				if cx != -1 {
					arrowOpacity = cfg.ArrowOpacity / 4
				}
			}
		}
	}

	// traverse y axis, drawing lines along x at intervals of cell height
	for y := 0; y < gridSize; y++ {
		for x := 0; x < w; x++ {
			grid.Set(x, y*hCell, getColour(cfg.LineDropColour, cfg.LineOpacity))
			grid.Set(x, y*hCell-1, getColour(cfg.LineColour, cfg.LineOpacity))
			if x%wCell < cfg.ArrowSize {
				if keyCol == keys[x/wCell] {
					arrowOpacity = cfg.ArrowOpacity
				}
				grid.Set(x, y*hCell, getColour(cfg.ArrowColour, arrowOpacity))
				if cx != -1 {
					arrowOpacity = cfg.ArrowOpacity / 4
				}
			}
		}
	}

	gridImg = ebiten.NewImageFromImage(grid)

	// draw key labels for cells
	fontOpacity := cfg.FontOpacity
	if cx != -1 {
		fontOpacity = cfg.FontOpacity / 4
	}
	for x := 0; x < gridSize; x++ {
		for y := 0; y < gridSize; y++ {
			if keyCol == keys[x] {
				fontOpacity = cfg.FontOpacity
			}
			text.Draw(gridImg, keys[x]+keys[y], mplusFont, x*wCell+11-cfg.GridStep*gx, y*hCell+19-cfg.GridStep*gy, getColour(cfg.FontDropColour, fontOpacity))
			text.Draw(gridImg, keys[x]+keys[y], mplusFont, x*wCell+10-cfg.GridStep*gx, y*hCell+18-cfg.GridStep*gy, getColour(cfg.FontColour, fontOpacity))
			if cx != -1 {
				fontOpacity = cfg.FontOpacity / 4
			}
		}
	}
}

// set x.y screen coordinates for selected cell
func setCoords(key ebiten.Key) {
	if cx == -1 {
		cx = int(key)
		if cx > 25 {
			cx -= 18
		}
		if key == ebiten.Key0 {
			cx = 35
		}
		makeGrid()
		return
	}
	if cy == -1 {
		cy = int(key)
		if cy > 25 {
			cy -= 18
		}
		if key == ebiten.Key0 {
			cy = 35
		}
		fmt.Println(wCell*cx-cfg.GridStep*gx, hCell*cy-cfg.GridStep*gy)
		exit = true
	}
}

// translates the grid in the direction of the arrow key pressed
func moveGrid(key ebiten.Key) {
	switch key {
	case ebiten.KeyArrowUp:
		gy++
	case ebiten.KeyArrowRight:
		gx--
	case ebiten.KeyArrowDown:
		gy--
	case ebiten.KeyArrowLeft:
		gx++
	}
	makeGrid()
}

// main loop
func (g *render) Update() error {
	g.frames++

	// exit application with Esc
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errRegularTermination
	}

	// key listeners
	switch {
	case inpututil.IsKeyJustReleased(ebiten.KeyA):
		keyCol = "a"
		setCoords(ebiten.KeyA)
	case inpututil.IsKeyJustReleased(ebiten.KeyB):
		keyCol = "b"
		setCoords(ebiten.KeyB)
	case inpututil.IsKeyJustReleased(ebiten.KeyC):
		keyCol = "c"
		setCoords(ebiten.KeyC)
	case inpututil.IsKeyJustReleased(ebiten.KeyD):
		keyCol = "d"
		setCoords(ebiten.KeyD)
	case inpututil.IsKeyJustReleased(ebiten.KeyE):
		keyCol = "e"
		setCoords(ebiten.KeyE)
	case inpututil.IsKeyJustReleased(ebiten.KeyF):
		keyCol = "f"
		setCoords(ebiten.KeyF)
	case inpututil.IsKeyJustReleased(ebiten.KeyG):
		keyCol = "g"
		setCoords(ebiten.KeyG)
	case inpututil.IsKeyJustReleased(ebiten.KeyH):
		keyCol = "h"
		setCoords(ebiten.KeyH)
	case inpututil.IsKeyJustReleased(ebiten.KeyI):
		keyCol = "i"
		setCoords(ebiten.KeyI)
	case inpututil.IsKeyJustReleased(ebiten.KeyJ):
		keyCol = "j"
		setCoords(ebiten.KeyJ)
	case inpututil.IsKeyJustReleased(ebiten.KeyK):
		keyCol = "k"
		setCoords(ebiten.KeyK)
	case inpututil.IsKeyJustReleased(ebiten.KeyL):
		keyCol = "l"
		setCoords(ebiten.KeyL)
	case inpututil.IsKeyJustReleased(ebiten.KeyM):
		keyCol = "m"
		setCoords(ebiten.KeyM)
	case inpututil.IsKeyJustReleased(ebiten.KeyN):
		keyCol = "n"
		setCoords(ebiten.KeyN)
	case inpututil.IsKeyJustReleased(ebiten.KeyO):
		keyCol = "o"
		setCoords(ebiten.KeyO)
	case inpututil.IsKeyJustReleased(ebiten.KeyP):
		keyCol = "p"
		setCoords(ebiten.KeyP)
	case inpututil.IsKeyJustReleased(ebiten.KeyQ):
		keyCol = "q"
		setCoords(ebiten.KeyQ)
	case inpututil.IsKeyJustReleased(ebiten.KeyR):
		keyCol = "r"
		setCoords(ebiten.KeyR)
	case inpututil.IsKeyJustReleased(ebiten.KeyS):
		keyCol = "s"
		setCoords(ebiten.KeyS)
	case inpututil.IsKeyJustReleased(ebiten.KeyT):
		keyCol = "t"
		setCoords(ebiten.KeyT)
	case inpututil.IsKeyJustReleased(ebiten.KeyU):
		keyCol = "u"
		setCoords(ebiten.KeyU)
	case inpututil.IsKeyJustReleased(ebiten.KeyV):
		keyCol = "v"
		setCoords(ebiten.KeyV)
	case inpututil.IsKeyJustReleased(ebiten.KeyW):
		keyCol = "w"
		setCoords(ebiten.KeyW)
	case inpututil.IsKeyJustReleased(ebiten.KeyX):
		keyCol = "x"
		setCoords(ebiten.KeyX)
	case inpututil.IsKeyJustReleased(ebiten.KeyY):
		keyCol = "y"
		setCoords(ebiten.KeyY)
	case inpututil.IsKeyJustReleased(ebiten.KeyZ):
		keyCol = "z"
		setCoords(ebiten.KeyZ)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit1):
		keyCol = "1"
		setCoords(ebiten.KeyDigit1)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit2):
		keyCol = "2"
		setCoords(ebiten.KeyDigit2)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit3):
		keyCol = "3"
		setCoords(ebiten.KeyDigit3)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit4):
		keyCol = "4"
		setCoords(ebiten.KeyDigit4)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit5):
		keyCol = "5"
		setCoords(ebiten.KeyDigit5)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit6):
		keyCol = "6"
		setCoords(ebiten.KeyDigit6)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit7):
		keyCol = "7"
		setCoords(ebiten.KeyDigit7)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit8):
		keyCol = "8"
		setCoords(ebiten.KeyDigit8)
	case inpututil.IsKeyJustReleased(ebiten.KeyDigit9):
		keyCol = "9"
		setCoords(ebiten.KeyDigit9)
	case inpututil.IsKeyJustReleased(ebiten.KeyArrowUp):
		moveGrid(ebiten.KeyArrowUp)
	case inpututil.IsKeyJustReleased(ebiten.KeyArrowRight):
		moveGrid(ebiten.KeyArrowRight)
	case inpututil.IsKeyJustReleased(ebiten.KeyArrowDown):
		moveGrid(ebiten.KeyArrowDown)
	case inpututil.IsKeyJustReleased(ebiten.KeyArrowLeft):
		moveGrid(ebiten.KeyArrowLeft)
	}
	if exit {
		return errRegularTermination
	}
	return nil
}

// draw grid
func (g *render) Draw(screen *ebiten.Image) {
	screen.DrawImage(gridImg, &ebiten.DrawImageOptions{})
}

func (g *render) Layout(outsideWidth, outsideHeight int) (int, int) {
	s := ebiten.DeviceScaleFactor()
	return int(float64(outsideWidth) * s), int(float64(outsideHeight) * s)
}

func main() {
	initFont()
	makeGrid()

	ebiten.SetScreenTransparent(true)
	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Screenpos")
	if err := ebiten.RunGame(&render{}); err != nil && err != errRegularTermination {
		log.Fatal(err)
	}
}
