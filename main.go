package main

// imported packages, most of them are for the golang game engine
import (
    "bytes"
    "image"
    "fmt"
    _ "image/png"
    "math"
    "log"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
    "time"
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
  // delay beginning on ticker until actually game mode
  // collision checker 
  collisionTicker = time.NewTicker(250 * time.Millisecond)
  shootTicker = time.NewTicker(2000 * time.Millisecond)
  gameImages *ebiten.Image
  bgImage *ebiten.Image
  count = 0
)

// basic information to draw sprites, track position and update position
type Body struct {
    // positions
    x float64 
    y float64
    // velocities
    vx float64 
    vy float64
    // get height and width from sheet.xml using sp
    width int 
    height int 
}

type Enemy struct {
    Body 
    sp int
}

type Laser struct {
    Body
    sp int
}

// in the future have a laser type struct, spriteImgNum, and number of animations
type Game struct {
    Val   string
    // tracks location of player and maybe health
    Player struct {
        Body
        health    int
        laserType int 
        canShoot  bool
        sp        int
        // consider adding in height and width of player object
        // all of the sprites seem to be the same
        // TODO set global width
    }
    PLasers []*Laser
    Enemies []*Enemy
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

// initial Player
func (g *Game) init() {
        
    _, _, _, width, height := utils.ImageData(images[playerSpriteStartNum])
	g.Val = "Testing"
    g.Player.canShoot = true 
    g.Player.sp = playerSpriteStartNum
    g.Player.Body.x = ScreenWidth / 2
    g.Player.Body.y = ScreenHeight - 100
    g.Player.Body.vx = 5
    g.Player.Body.vy = 5
    g.Player.Body.width = width 
    g.Player.Body.height = height
    g.addEnemy(50)
}

// main game loop
func (g *Game) Update(screen *ebiten.Image) error {
    if ebiten.IsDrawingSkipped() {
		return nil
    }
    // draw background
    ScrollBG(screen)
    // TPS counter
    fps := fmt.Sprintf("TPS: %f", ebiten.CurrentTPS())
    ebitenutil.DebugPrint(screen, fps)
    // show if ship should move
    g.moveShip()
    // check if laser is shot
    g.shootLaser()
    // draw and update lasers
    // maybe goroutine some of this
    g.moveAndDrawLasers(screen)
    g.moveAndDrawEnemies(screen)
    g.checkCollisions()
    // g.drawLasers(screen)
    g.drawShip(screen)
    return nil
}

func (g *Game) checkCollisions() {
    for i := 0; i < len(g.Enemies); i++ {
        s := g.Enemies[i]
        go func() {
            for _ = range collisionTicker.C {
                // fmt.Println("Can shoot laser")
                collision := g.checkPlayerEnemyCollision(s)
                fmt.Println(collision)
            }
        }()

        // for j :=0; j < len(g.PLasers); j++ {
        //    pl := g.PLasers[j]
        // }
	}
}

// Collision Functions

func (g* Game) checkPlayerEnemyCollision(e* Enemy) (bool) {
    // fmt.Println(g.Player.Body)
    return BodyCollided(&e.Body,&g.Player.Body)
}
// check if bodies have collided
func BodyCollided(r1 *Body,  r2 *Body) (bool) {
    // compute rectangle 1
    r1L, r1R, r1T, r1B := ComputeRect(r1)
    // compute rectangle 2
    r2L, r2R, r2T, r2B := ComputeRect(r2)
    // fmt.Println(r1L, r1R, r1T, r1B)
    // fmt.Println(r2L, r2R, r2T, r2B)
    return (r1L < r2R && r1R > r2L &&
        r1B > r2T && r1T < r2B)
}

func ComputeRect(rect *Body) (float64, float64, float64, float64) {
    rectL := rect.x - float64(rect.width / 2)
    rectR := rect.x + float64(rect.width / 2)
    rectT := rect.y - float64(rect.height / 2)
    rectB := rect.y + float64(rect.height / 2)

    return rectL, rectR, rectT, rectB
}

/* func (g* Game) enemyPlayerCollided(s *Enemy) (bool) {
    p := g.Player

    eleft := s.x - float64(s.width / 2)
    eright := s.x + float64(s.width / 2)
    etop := s.y - float64(s.height / 2)
    ebottom := s.y + float64(s.height / 2)

    pleft := p.x - float64(p.width / 2)
    pright := p.x + float64(p.width / 2)
    ptop := p.y - float64(p.height / 2)
    pbottom := p.y + float64(p.height / 2)

    return (eleft < pright && eright > pleft &&
        ebottom > ptop && etop < pbottom)
    // return (RectA.X1 < RectB.X2 && RectA.X2 > RectB.X1 &&
    //    RectA.Y1 > RectB.Y2 && RectA.Y2 < RectB.Y1) 
} */

func (g *Game) removeLaser(i int) {
    s := g.PLasers
    s[i] = s[len(s)-1]
    g.PLasers = s[:len(s)-1]
    // https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-array-in-golang/37335777
    // s[i] = s[len(s)-1]
    // # We do not need to put s[i] at the end, as it will be discarded anyway
    //return s[:len(s)-1]
}

// give player laser type, add laser struct to Player struct
func (g *Game) shootLaser() {
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
        // Selects preloaded sprite
        if (g.Player.canShoot) {
            // make new laser
            g.addLaser()
            g.Player.canShoot = false
        }
    }
    go func() {
        for _ = range shootTicker.C {
            // fmt.Println("Can shoot laser")
            g.Player.canShoot = true
        }
    }()
}

/*
 *
 * Adding Lasers, Enemies functions, and powerups in the future
 */
// adding new 
func (g *Game) addLaser() {
    px := g.Player.Body.x 
    py := g.Player.Body.y 
    // vx not used outside of initialization
    vx := 1.00
    vy := 3.00
    snum := 1
    _, _, _, width, height := utils.ImageData(images[snum])
    // fmt.Println("shooting a laser")
    g.PLasers = append(g.PLasers, &Laser{Body{px, py, vx, vy, width, height}, snum})

}

// TODO Make the spawn location within a randomized region
// sum is the sprite num corresponding to sheet.xml
func (g *Game) addEnemy(snum int) {

    px := g.Player.Body.x
    py := float64(ScreenHeight / 2) 
    vx := 0.00
    vy := 3.00

    _, _, _, width, height := utils.ImageData(images[snum])
    // fmt.Println("shooting a laser")
    g.Enemies = append(g.Enemies, &Enemy{Body{px, py, vx, vy, width, height}, snum})

}

/*
 *
 * Movement and Drawing Functions --- Ships, background and enemies, lasers
 */
// draws the player ship using game object player data 
func (g *Game) drawShip(screen *ebiten.Image) {
	count++
	op := &ebiten.DrawImageOptions{}
    // move to player location
    i := (count / 10) % 7
    op.GeoM.Translate(g.Player.Body.x, g.Player.Body.y)
    // player ships from number 207 to 215
	_, x, y, width, height := utils.ImageData(images[playerSpriteStartNum+i])
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
}

// move and draw lasers
func (g *Game) moveAndDrawEnemies(screen *ebiten.Image) {
    for i := 0; i < len(g.Enemies); i++ {
        // update enemies
        s := g.Enemies[i]
        if (s.x < 0 ) {
            g.Enemies[i].Body.vx = -g.Enemies[i].Body.vx
        } else if (s.x > ScreenWidth) {
            g.Enemies[i].Body.vx = -g.Enemies[i].Body.vx
        }
        g.Enemies[i].Body.x += g.Enemies[i].Body.vx
        // draw image
        _, x, y, width, height := utils.ImageData(images[s.sp])
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(float64(s.Body.x), float64(s.Body.y))
        screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
	}
}

// move and draw lasers
func (g *Game) moveAndDrawLasers(screen *ebiten.Image) {
    // get player data to determine where bullet should spawn
    // consider getting global height width for player object later
    _, _, _, ipw, iph := utils.ImageData(images[playerSpriteStartNum])
    pw := float64(ipw)
    ph := float64(iph)
    for i := 0; i < len(g.PLasers); i++ {
        s := g.PLasers[i]
        _, x, y, width, height := utils.ImageData(images[s.sp])
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Rotate(90 * math.Pi / 180)
        op.GeoM.Translate(float64(s.x) + 5 + pw / 2, float64(s.y) - ph / 2)
        screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
        if (s.y < -float64(height)) {
            // fmt.Println("Deleting Laser")
            g.removeLaser(i)
        } else {
            g.PLasers[i].y -= g.PLasers[i].vy
        }
	}
}


// make the background scroll
func ScrollBG(screen *ebiten.Image) {
    theViewport.Move()
    x16, y16 := theViewport.Position()
	offsetX, offsetY := float64(-x16) /16, float64(-y16) /16

	// Draw bgImage on the screen repeatedly.
	const repeat = 3
	w, h := bgImage.Size()
	for j := 0; j < repeat; j++ {
		for i := 0; i < repeat; i++ {
            op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(w*i), float64(h*j))
			op.GeoM.Translate(offsetX, offsetY)
			screen.DrawImage(bgImage, op)
		}
    }
}

// TODO Handle out of bounds cases
func (g *Game) moveShip() {
	// Controls
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		// Selects preloaded sprite
		g.Player.Body.x -= g.Player.Body.vx
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		// Moves character 3px left
		g.Player.Body.x += g.Player.Body.vx
	} else if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.Player.Body.y -= g.Player.Body.vy
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
        g.Player.Body.y += g.Player.Body.vy
  }
}

func main() {
    g := NewGame()
    // add const screenHeight and screenWidth later
    if err := ebiten.Run(g.Update, ScreenWidth, ScreenHeight, ScaleFactor, "Space Shooter!"); err != nil {
        panic(err)
    }
}