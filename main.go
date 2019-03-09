package main

// imported packages, most of them are for the golang game engine
import (
    "image/color"
    "bytes"
    "image"
    "fmt"
    _ "image/png"
    "log"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
    "time"
    "math/rand"
    utils "github.com/FriendlyUser/spaceshooter/utils"
    resources "github.com/FriendlyUser/spaceshooter/resources"
    // "github.com/hajimehoshi/go-inovation/ino/internal/input"
    // "github.com/hajimehoshi/ebiten/inpututil"
    "github.com/hajimehoshi/ebiten/audio"
    "github.com/hajimehoshi/ebiten/audio/wav"
	"github.com/hajimehoshi/ebiten/audio/vorbis"
    "github.com/golang/freetype/truetype"
    "golang.org/x/image/font"
    "github.com/hajimehoshi/ebiten/examples/resources/fonts"
    "github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
)

// hardcoded animation and sprite numbers from resources/sheet.xml
const (
  playerSpriteStartNum = 207
  playerSpriteEndNum = 215
  ScreenWidth = 1920
  ScreenHeight = 1440
  ScaleFactor = 0.5
  fontSize         = 64
  smallFontSize    = fontSize / 2
)

// game images
var (
  // global metadata for images from sheet.xml
  images, _ = utils.ReadImageData()
  // delay beginning on ticker until actually game mode
  // collision checker 
  gameImages *ebiten.Image
  bgImage *ebiten.Image
  count = 0
)

const (
    loopLengthInSecond  = 17
)
// sounds 
var (
	audioContext *audio.Context
	laserPlayer  *audio.Player
	musicPlayer  *audio.Player
)

func init() {
    // sampling := 
	audioContext, _ = audio.NewContext(22050)

	laser, err := wav.Decode(audioContext, audio.BytesReadSeekCloser(resources.Laser_wav))
	if err != nil {
		log.Fatal(err)
	}
    laserPlayer, err = audio.NewPlayer(audioContext, laser)
	if err != nil {
		log.Fatal(err)
    }
    laserPlayer.SetVolume(0.25)


    oggS, err := vorbis.Decode(audioContext, audio.BytesReadSeekCloser(resources.OST_ogg))
    if err != nil {
        log.Fatal(err)
    }

    // Create an infinite loop stream from the decoded bytes.
    // s is still an io.ReadCloser and io.Seeker.
    s := audio.NewInfiniteLoop(oggS, loopLengthInSecond*600*22050)

    musicPlayer, err = audio.NewPlayer(audioContext, s)
    if err != nil {
        log.Fatal(err)
    }
    musicPlayer.SetVolume(0.5)
    // Play the infinite-length stream. This never ends.
    musicPlayer.Play()
}
// consider having a laser type to deal with orientation, etc
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
    health int
}

type Laser struct {
    Body
    sp int
}

type Mode int
const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)
var (
    arcadeFont      font.Face
	smallArcadeFont font.Face
)
// fonts and sizes
func init() {
	tt, err := truetype.Parse(fonts.ArcadeN_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	arcadeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	smallArcadeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    smallFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}
// in the future have a laser type struct, spriteImgNum, and number of animations
type Game struct {
    mode Mode
    level int
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
    ELasers []*Laser
    gameoverCount int
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

// player health bar
var (
    playerHB *ebiten.Image
    playerHBoutline *ebiten.Image
)

func (p *viewport) Position() (int, int) {
	return p.x16, p.y16
}

func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

// random number generation
func init() {
    rand.Seed(time.Now().UnixNano())
}

// initial Player
func (g *Game) init() {
        
    _, _, _, width, height := utils.ImageData(images[playerSpriteStartNum])
    g.Player.sp = playerSpriteStartNum
    g.Player.Body.x = ScreenWidth / 2
    g.Player.Body.y = ScreenHeight - 100
    g.Player.Body.width = width 
    g.Player.Body.height = height
    g.Player.health = 100
    g.Player.canShoot = true 
    g.level = 0
    g.Player.Body.vx = 5
    g.Player.Body.vy = 5

    g.CreateLevel()
}

func (g *Game) CreateLevel() {
    g.level = g.level + 1 
    gameLevel := g.level + 10
    enemySprite := 50
    // spawn enemies based on level
    for i := 1; i <= gameLevel; i++ {
        enemySprite = (rand.Intn(10)) + 50
        g.addEnemyRand(enemySprite)
    }
}
// copied from flappyplane, but basic functionality any key pressed
func jump() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	if len(inpututil.JustPressedTouchIDs()) > 0 {
		return true
	}
	return false
}

// main game loop
func (g *Game) Update(screen *ebiten.Image) error {
    switch g.mode {
    // allow user to jump from title screen
    // and title screen
    case ModeTitle:
        msg := []string{
			"Game Created Using Ebiten",
			"Created By David Li.",
		}
		for i, l := range msg {
            x := (ScreenWidth - len(l)) / 2
            y := (ScreenHeight) / 2
			text.Draw(screen, l, smallArcadeFont, x, y + i*40, color.White)
		}
        if jump() {
			g.mode = ModeGame
        }
    // main game mode with enemies, move objects and draw objects
    // main game loop
    // TODO refactor to separate movement and drawing, perhaps for efficiency reasons
    case ModeGame:
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
        g.drawShip(screen)   
        g.drawHealthBar(screen)
        g.moveAndDrawEnemyLasers(screen)
        g.checkCollisions()
        g.checkPlayerLaserCollision()
        // g.drawLasers(screen)
        // check if game over
        g.checkGameOver()
    // game over instance
    case ModeGameOver:
        // TODO DELETE ALL LASER, RESET PLAYER HEALTH, ETC, or add it to init
        msg := []string{
			"Game Over :(",
			"Created By David Li.",
		}
		for i, l := range msg {
            x := (ScreenWidth - len(l)) / 2
            y := (ScreenHeight) / 2
			text.Draw(screen, l, smallArcadeFont, x, y + i*40, color.White)
		}
        if g.gameoverCount > 0 {
            g.gameoverCount--
        }
        if g.gameoverCount == 0 && jump() {
            g.init()
            g.mode = ModeTitle
        }
    }
    return nil
}

// check if player got hit by enemy lasers
func (g *Game) checkPlayerLaserCollision() {
    // issues occur if you delete the same object during iteration
    // instead have a list of lasers to delete and then delete them outside the loop
    // check player lasers to player collisions
    if (count % 5 == 0) {
        for j := 0; j < len(g.ELasers); j++ {
            pelHit := g.checkPlayerELaserCollision(j)
            if (pelHit) {
                g.Player.health -= 1
                pelHit = false
                g.removeEnemyLaser(j)
            }
        }
    }
}

// not sure if I should "goroutine it all", it appears game is using goroutines already
// goroutine still causes issues, I think I want collisions to be synced.
func (g *Game) checkCollisions() {
    if (count % 5 == 0) {
        for i := 0; i < len(g.Enemies); i++ {
            s := g.Enemies[i]
            // since we are only checking collision between all enemies and a single player, this approach is fine
            pHit := g.checkPlayerEnemyCollision(s)
            if (pHit) {
                g.Player.health -= 1
            }
            for j :=0; j < len(g.PLasers); j++ {

                eHit := g.checkEnemyLaserCollision(s,j)
                if (eHit) {
                    eHit = false
                    g.removeLaser(j)
                    // make sure enemy exists before removing in, thats what you got to do with goroutines
                    if len(g.Enemies) > i {
                        g.Enemies[i].health -= 3
                    }
                    // remove enemy, convert to explosion object that will be erased later

                }
            }
        }
    }
}

// Collision Functions

// check laserEnemyCollision
func  (g* Game) checkPlayerELaserCollision(j int) (bool) {
    el := g.ELasers[j]
    // adjust body to correct looking 
    p := g.Player
    return BodyCollided(&p.Body, &el.Body)
}

func (g* Game) checkEnemyLaserCollision(e* Enemy, j int) (bool) {
    p := g.PLasers[j]
    return BodyCollided(&e.Body, &p.Body)
}

// check laserEnemyCollision
func  (g* Game) checkPlayerEnemyCollision(e* Enemy) (bool) {
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

// since I am using goroutines I need to check if the specified enemy exists every time this function 
// is called
func (g *Game) removeEnemy(i int) {
    s := g.Enemies
    if i >=0 && i < len(s) {
        s[i] = s[len(s)-1]
        g.Enemies =  s[:len(s)-1]
    }
}

func (g *Game) removeEnemyLaser(i int) {
    s := g.ELasers
    if i >= 0 && i < len(s) {
        s[i] = s[len(s)-1]
        g.ELasers =  s[:len(s)-1]
    }
}

func (g *Game) removeLaser(i int) {
    s := g.PLasers
    if i >= 0 && i < len(s) {
        s[i] = s[len(s)-1]
        g.PLasers =  s[:len(s)-1]
    }
}

// give player laser type, add laser struct to Player struct
func (g *Game) shootLaser() {
    // if issue persists, use count instead of goroutine to determine when player can shoot
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
        // Selects preloaded sprite
        if (g.Player.canShoot) {
            // make new laser
            g.addLaser()
            g.Player.canShoot = false
        }
    }
    // wait 200 millseconds before player can shoot
    if (count % 20 == 0) {
        g.Player.canShoot = true
    }

}

/*
 *
 * Adding Lasers, Enemies functions, and powerups in the future
 */
// adding new 
func (g *Game) addLaser() {
    laserPlayer.Rewind()
    laserPlayer.Play()
    _, _, _, ipw, iph := utils.ImageData(images[playerSpriteStartNum])
    pw := float64(ipw)
    ph := float64(iph)
    px := g.Player.Body.x + pw / 2
    py := g.Player.Body.y - ph
    // vx not used outside of initialization
    vx := 1.00
    vy := 3.00
    snum := 105
    _, _, _, width, height := utils.ImageData(images[snum])
    // fmt.Println("shooting a laser")
    g.PLasers = append(g.PLasers, &Laser{Body{px, py, vx, vy, width, height}, snum})

}


// TODO FInish this function
// ADD SPECIFIC FUNCTIONALITY for enemy
func (g *Game) addEnemyLoc(snum int) {
    emax := 5 
    emin := -5
    px := float64(rand.Intn(ScreenWidth))
    py := float64(rand.Intn(ScreenHeight / 2) + ScreenHeight / 4) 
    vx := float64(emin + rand.Intn(emax-emin+1))
    vy := float64(emin + rand.Intn(emax-emin+1))

    _, _, _, width, height := utils.ImageData(images[snum])

    health := 2
    // fmt.Println("shooting a laser")
    g.Enemies = append(g.Enemies, &Enemy{Body{px, py, vx, vy, width, height}, snum, health})

}

// TODO Make the spawn location within a randomized region
// sum is the sprite num corresponding to sheet.xml
func (g *Game) addEnemyRand(snum int) {
    emax := 5 
    emin := -5
    spawny := 0.75
    px := float64(rand.Intn(ScreenWidth))
    py := float64(spawny*float64(rand.Intn(ScreenHeight)) + ScreenHeight / 8) 
    vx := float64(emin + rand.Intn(emax-emin+1))
    vy := float64(emin + rand.Intn(emax-emin+1))

    _, _, _, width, height := utils.ImageData(images[snum])

    health := 2
    g.Enemies = append(g.Enemies, &Enemy{Body{px, py, vx, vy, width, height}, snum, health})

}

// destroy all generated game objects, return player to starting point, etc ...
func (g* Game) resetLevel() {

}

// create new enemy laser
func (g *Game) enemyShootLaser(x float64, y float64) {
    // vx not used outside of initialization
    vx := 1.00 
    vy := rand.Float64() * 4 + 0.01
    snum := rand.Intn(10) + 110
    _, _, _, width, height := utils.ImageData(images[snum])
    g.ELasers = append(g.ELasers, &Laser{Body{x, y, vx, vy, width, height}, snum})
}

// move and draw lasers
func (g *Game) moveAndDrawEnemyLasers(screen *ebiten.Image) {

    for i := 0; i < len(g.ELasers); i++ {
        s := g.ELasers[i]
        _, x, y, width, height := utils.ImageData(images[s.sp])
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(float64(s.x), float64(s.y))
        op.Filter = ebiten.FilterLinear
        screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
        if (s.y > ScreenHeight) {
            g.removeEnemyLaser(i)
        } else {
            g.ELasers[i].y += g.ELasers[i].vy
        }
	}
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


// draw health bar
func (g *Game) drawHealthBar(screen *ebiten.Image) {
    w := 10
    h := 10
    health := g.Player.health
    _, _, _, width, height := utils.ImageData(images[13])
    if playerHBoutline == nil {
        // Create an 16x16 image
        playerHBoutline, _ = ebiten.NewImage(width-5, height-5, ebiten.FilterNearest)
    } 

    // Fill the square with the white color
    playerHBoutline.Fill(color.NRGBA{0xff, 0x00, 0x00, 0xff})
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(w+1), float64(h+1))

    // Draw the square image to the screen with an empty option
    screen.DrawImage(playerHBoutline, op)
    if health > 0 {
        playerHB, _ = ebiten.NewImage(health*width/100, height, ebiten.FilterNearest)
    }
    playerHB.Fill(color.NRGBA{0x00, 0xff, 0x00, 0xff})
    // if playerHB == nil {
    // Create an 16x16 image    
   
    // } 
    op = &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(w), float64(h))
    screen.DrawImage(playerHB, op)

}

// move and draw enemies
func (g *Game) moveAndDrawEnemies(screen *ebiten.Image) {
    for i := 0; i < len(g.Enemies); i++ {
        s := g.Enemies[i]
        // destroy enemy if health is low, maybe seperate loop to remove glittery behaviour
        if (s.health < 0) {
            g.removeEnemy(i)
            continue
        }
        // update enemies
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

        // make bullet be shot every 20 seconds
        if (count % 200 == 0) {
            elaserx := s.x
            elasery := s.y
            g.enemyShootLaser(elaserx, elasery)
        }
    }
    // go to next level if all enemies are dead
    if len(g.Enemies) == 0 {
        g.CreateLevel()
    }
}

// move and draw lasers
func (g *Game) moveAndDrawLasers(screen *ebiten.Image) {
    // get player data to determine where bullet should spawn
    // consider getting global height width for player object later
    for i := 0; i < len(g.PLasers); i++ {
        s := g.PLasers[i]
        _, x, y, width, height := utils.ImageData(images[s.sp])
        op := &ebiten.DrawImageOptions{}
        // op.GeoM.Rotate(90 * math.Pi / 180)
        op.GeoM.Translate(float64(s.x), float64(s.y))
        screen.DrawImage(gameImages.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image), op)
        if (s.y < -float64(height)) {
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

func (g *Game) moveShip() {
	// Controls
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
        // Selects preloaded sprite
        if (int(g.Player.Body.x) > - g.Player.Body.width / 2) {
            g.Player.Body.x -= g.Player.Body.vx
        }
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
        // Moves character 3px left
        if (int(g.Player.Body.x) < ScreenWidth - g.Player.Body.width / 2) {
            g.Player.Body.x += g.Player.Body.vx
        }
	} else if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
        if (int(g.Player.Body.y) > -g.Player.Body.height / 2) {
            g.Player.Body.y -= g.Player.Body.vy
        }
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {

        if (int(g.Player.Body.y) < ScreenHeight - g.Player.Body.width / 2) {
            g.Player.Body.y += g.Player.Body.vy
        }
  }
}

// functions to check if game is over, level should proceed, perhaps even boss level
func (g *Game) checkGameOver() {
    if g.Player.health < 0 {
        g.mode = ModeGameOver
    }
}

func main() {
    g := NewGame()
    // add const screenHeight and screenWidth later
    if err := ebiten.Run(g.Update, ScreenWidth, ScreenHeight, ScaleFactor, "Space Shooter!"); err != nil {
        panic(err)
    }
}