package main

import (
    "bytes"
    "image"
    "fmt"
    _ "image/png"
    "log"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
    utils "github.com/FriendlyUser/spaceshooter/utils"
    resources "github.com/FriendlyUser/spaceshooter/resources"
    // "github.com/hajimehoshi/go-inovation/ino/internal/input"
    // "github.com/hajimehoshi/ebiten/inpututil"
)

// hardcoded animation and sprite numbers from resources/sheet.xml
const (
  playerSpriteStartNum = 207
  playerSpriteEndNum = 215
  ScreenWidth = 1920
  ScreenHeight = 1440
  ScaleFactor = 0.5
)

// game images
var (
  // global metadata for images from sheet.xml
  images, _ = utils.ReadImageData("resources/sheet.xml")
  gameImages *ebiten.Image
  bgImage *ebiten.Image
  count = 0
)

type Game struct {
    Val   string
    // tracks location of player and maybe health
    Player struct {
        x         float64
        y         float64 
        health    int
        laserType int 
    }
}

// load images
func init() {
    // sprites
	img, _, err := image.Decode(bytes.NewReader(resources.Sprites_png))
	if err != nil {
		log.Fatal(err)
	}
    gameImages, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

    // backgrounds
    img, _, err = image.Decode(bytes.NewReader(resources.Starfieldreal_jpg))
	if err != nil {
		log.Fatal(err)
	}
	bgImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}

// background image logic from 
// # https://github.com/hajimehoshi/ebiten/blob/master/examples/infinitescroll/main.go
var (
	theViewport = &viewport{}
)

type viewport struct {
	x16 int
	y16 int
}

func (p *viewport) Move() {
	w, h := bgImage.Size()
	maxX16 := w * 16
	maxY16 := h * 16

	p.x16 += w / 32
	p.y16 += h / 32
	p.x16 %= maxX16
	p.y16 %= maxY16
}

func (p *viewport) Position() (int, int) {
	return p.x16, p.y16
}
func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.Val = "Testing"
	g.Player.x = 200.00
	g.Player.y = 100.00
}

func (g *Game) Update(screen *ebiten.Image) error {
    theViewport.Move()
    if ebiten.IsDrawingSkipped() {
		return nil
    }
    
    x16, y16 := theViewport.Position()
	offsetX, offsetY := float64(-x16)/16, float64(-y16)/16

	// Draw bgImage on the screen repeatedly.
	const repeat = 5
	w, h := bgImage.Size()
	for j := 0; j < repeat; j++ {
		for i := 0; i < repeat; i++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(w*i), float64(h*j))
			op.GeoM.Translate(offsetX, offsetY)
			screen.DrawImage(bgImage, op)
		}
    }
    
    // TPS counter
    fps := fmt.Sprintf("TPS: %f", ebiten.CurrentTPS())
    ebitenutil.DebugPrint(screen, fps)
    g.moveShip()
    g.drawShip(screen)
    return nil
}

// give player laser type, add laser struct to Player struct
func (g *Game) shootLaser() {

}
// TODO Handle out of bounds cases
func (g *Game) moveShip() {
	// Controls
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		// Selects preloaded sprite
		g.Player.x += -3
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		// Moves character 3px left
		g.Player.x += 3
	} else if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.Player.y += 3
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
    g.Player.y -= 3
  }
}

// draws the player ship using game object player data 
func (g *Game) drawShip(screen *ebiten.Image) {
	count++
	op := &ebiten.DrawImageOptions{}
  // move to player location
  i := (count / 10) % 7
  op.GeoM.Translate(g.Player.x, g.Player.y)
  // player ships from number 207 to 215
	_, x, y, width, height := utils.ImageData(images[playerSpriteStartNum+i])
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
}

func main() {
    g := NewGame()
    // add const screenHeight and screenWidth later
    if err := ebiten.Run(g.Update, ScreenWidth, ScreenHeight, ScaleFactor, "Hello world!"); err != nil {
        panic(err)
    }
}