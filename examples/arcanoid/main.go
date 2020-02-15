package main

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/life4/gweb/canvas"
	"github.com/life4/gweb/web"
)

const (
	BGColor       = "#ecf0f1"
	BallColor     = "#27ae60"
	PlatformColor = "#2c3e50"
	TextColor     = "#2c3e50"
)

const PlatformWidth = 120
const PlatformHeight = 20
const PlatformMaxSpeed = 40

const BallSize = 20

const BrickHeight = 20
const BrickRows = 8
const BrickCols = 14
const BrickMarginLeft = 120 // pixels
const BrickMarginTop = 10   // pixels
const BrickMarginX = 5      // pixels
const BrickMarginY = 5      // pixels

type Platform struct {
	context canvas.Context2D
	element web.Canvas
	// geometry
	width int
	x, y  int
	// movement
	mouseX int
	// borders
	windowWidth  int
	windowHeight int
}

func (platform Platform) contains(x, y int) bool {
	if y < platform.y { // ball upper
		return false
	}
	if y > platform.y+PlatformHeight { // ball downer
		return false
	}
	if x > platform.x+platform.width { // ball righter
		return false
	}
	if x < platform.x { // ball lefter
		return false
	}
	return true

}

func (ctx *Platform) changePosition() {
	path := ctx.mouseX - (ctx.x + ctx.width/2)
	if path == 0 {
		return
	}

	// don't move too fast
	if path > 0 && path > PlatformMaxSpeed {
		path = PlatformMaxSpeed
	} else if path < 0 && path < -PlatformMaxSpeed {
		path = -PlatformMaxSpeed
	}

	// don't move out of playground
	if ctx.x+path <= 0 {
		ctx.x = 0
		return
	}
	if ctx.x+path >= ctx.windowWidth-ctx.width {
		ctx.x = ctx.windowWidth - ctx.width
		return
	}

	ctx.x += path
}

func (ctx *Platform) handleMouse(event web.Event) {
	ctx.mouseX = event.Get("clientX").Int()
}

func (ctx *Platform) handleFrame() {
	// clear out previous render
	ctx.context.SetFillStyle(BGColor)
	ctx.context.Rectangle(ctx.x, ctx.y, ctx.width, PlatformHeight).Filled().Draw()

	// change platform coordinates
	ctx.changePosition()

	// draw the platform
	ctx.context.SetFillStyle(PlatformColor)
	ctx.context.Rectangle(ctx.x, ctx.y, ctx.width, PlatformHeight).Filled().Draw()
}

type Ball struct {
	context canvas.Context2D
	// position
	x, y int
	// movement
	vectorX int
	vectorY int
	// borders
	windowWidth  int
	windowHeight int
	platform     *Platform
}

func (ctx *Ball) changeDirection() {
	ballX := ctx.x + ctx.vectorX
	ballY := ctx.y + ctx.vectorY
	// bounce from text box (where we draw FPS and score)
	if ballX < 110+BallSize && ballY < 60 {
		ctx.vectorX = -ctx.vectorX
	}
	if ballX < 110 && ballY < 60+BallSize {
		ctx.vectorY = -ctx.vectorY
	}

	// right and left
	if ballX > ctx.windowWidth-BallSize {
		ctx.vectorX = -ctx.vectorX
	} else if ballX < BallSize {
		ctx.vectorX = -ctx.vectorX
	}

	// bottom and top
	if ballY > ctx.windowHeight-BallSize {
		ctx.vectorY = -ctx.vectorY
	} else if ballY < BallSize {
		ctx.vectorY = -ctx.vectorY
	}

	// bounce from platform top
	if ctx.vectorY > 0 && ctx.platform.contains(ballX, ballY+BallSize) {
		ctx.vectorY = -ctx.vectorY
	}
	// bounce from platform bottom
	if ctx.vectorY < 0 && ctx.platform.contains(ballX, ballY-BallSize) {
		ctx.vectorY = -ctx.vectorY
	}
	// bounce from platform left
	if ctx.vectorX > 0 && ctx.platform.contains(ballX+BallSize, ballY) {
		ctx.vectorX = -ctx.vectorX
	}
	// bounce from platform right
	if ctx.vectorX < 0 && ctx.platform.contains(ballX-BallSize, ballY) {
		ctx.vectorX = -ctx.vectorX
	}
}

func (ctx *Ball) handle() {
	// clear out previous render
	ctx.context.SetFillStyle(BGColor)
	ctx.context.BeginPath()
	ctx.context.Arc(ctx.x, ctx.y, BallSize+1, 0, math.Pi*2)
	ctx.context.Fill()
	ctx.context.ClosePath()

	ctx.changeDirection()

	// move the ball
	ctx.x += ctx.vectorX
	ctx.y += ctx.vectorY

	// draw the ball
	ctx.context.SetFillStyle(BallColor)
	ctx.context.BeginPath()
	ctx.context.Arc(ctx.x, ctx.y, BallSize, 0, math.Pi*2)
	ctx.context.Fill()
	ctx.context.ClosePath()
}

type Brick struct {
	context canvas.Context2D
	x, y    int
	width   int
	height  int
	removed bool
}

func (brick Brick) contains(x, y int) bool {
	if y < brick.y { // ball upper
		return false
	}
	if y > brick.y+brick.height { // ball downer
		return false
	}
	if x > brick.x+brick.width { // ball righter
		return false
	}
	if x < brick.x { // ball lefter
		return false
	}
	return true
}

func (brick *Brick) Collide(ball *Ball) bool {
	if brick.removed {
		return false
	}

	// quick checks of ball position
	if ball.x-BallSize > brick.x+brick.width { // ball righter
		return false
	}
	if ball.x+BallSize < brick.x { // ball lefter
		return false
	}
	if ball.y+BallSize < brick.y { // ball upper
		return false
	}
	if ball.y-BallSize > brick.y+brick.height { // ball downer
		return false
	}

	// bottom of brick collision
	if ball.vectorY < 0 && brick.contains(ball.x, ball.y-BallSize) {
		ball.vectorY = -ball.vectorY
		return true
	}
	// top of brick collision
	if ball.vectorY > 0 && brick.contains(ball.x, ball.y+BallSize) {
		ball.vectorY = -ball.vectorY
		return true
	}
	// left of brick collision
	if ball.vectorX > 0 && brick.contains(ball.x+BallSize, ball.y) {
		ball.vectorX = -ball.vectorX
		return true
	}
	// right of brick collision
	if ball.vectorX < 0 && brick.contains(ball.x-BallSize, ball.y) {
		ball.vectorX = -ball.vectorX
		return true
	}

	return false
}

func (brick *Brick) Draw(color string) {
	brick.context.SetFillStyle(color)
	brick.context.Rectangle(brick.x, brick.y, brick.width, brick.height).Filled().Draw()
	brick.removed = false
}

func (brick *Brick) Remove() {
	brick.context.SetFillStyle(BGColor)
	brick.context.Rectangle(brick.x, brick.y, brick.width, brick.height).Filled().Draw()
	brick.removed = true
}

type Bricks struct {
	context      canvas.Context2D
	registry     []*Brick
	ready        bool
	windowWidth  int
	windowHeight int
}

func (bricks *Bricks) Draw() {
	bricks.registry = make([]*Brick, BrickCols*BrickRows)
	width := (bricks.windowWidth-BrickMarginLeft)/BrickCols - BrickMarginX
	colors := []string{"#c0392b", "#d35400", "#f39c12", "#f1c40f"}
	for i := 0; i < BrickCols; i++ {
		for j := 0; j < BrickRows; j++ {
			x := BrickMarginLeft + (width+BrickMarginX)*i
			y := BrickMarginTop + (BrickHeight+BrickMarginY)*j
			brick := Brick{
				context: bricks.context,
				x:       x,
				y:       y,
				width:   width,
				height:  BrickHeight,
			}
			color := colors[(j/2)%len(colors)]
			brick.Draw(color)
			bricks.registry[BrickRows*i+j] = &brick
		}
	}
	bricks.ready = true
}

func (bricks *Bricks) Handle(ball *Ball) {
	for _, brick := range bricks.registry {
		if !brick.Collide(ball) {
			continue
		}
		brick.Remove()
	}
}

type FPS struct {
	context canvas.Context2D
	updated time.Time
}

func (h *FPS) drawFPS(now time.Time) {
	// calculate FPS
	fps := time.Second / now.Sub(h.updated)
	text := fmt.Sprintf("%d FPS", int64(fps))

	// clear
	h.context.SetFillStyle(BGColor)
	h.context.Rectangle(10, 10, 100, 20).Filled().Draw()

	// write
	h.context.Text().SetFont("bold 20px Roboto")
	h.context.SetFillStyle(TextColor)
	h.context.Text().Fill(text, 10, 30, 100)
}

func (h *FPS) handle() {
	now := time.Now()
	// update FPS counter every second
	if h.updated.Second() != now.Second() {
		h.drawFPS(now)
	}
	h.updated = now
}

func main() {
	window := web.GetWindow()
	doc := window.Document()
	doc.SetTitle("Arkanoid")
	body := doc.Body()

	// create canvas
	h := window.InnerHeight() - 40
	w := window.InnerWidth() - 40
	canvas := doc.CreateCanvas()
	canvas.SetHeight(h)
	canvas.SetWidth(w)
	body.Node().AppendChild(canvas.Node())

	context := canvas.Context2D()

	// draw background
	context.SetFillStyle(BGColor)
	context.BeginPath()
	context.Rectangle(0, 0, w, h).Filled().Draw()
	context.Fill()
	context.ClosePath()

	// make handlers
	platform := Platform{
		context:      context,
		element:      canvas,
		x:            w / 2,
		y:            h - 60,
		mouseX:       w / 2,
		width:        PlatformWidth,
		windowWidth:  w,
		windowHeight: h,
	}
	fps := FPS{context: context, updated: time.Now()}
	ball := Ball{
		context: context,
		vectorX: 5, vectorY: -5,
		x:           platform.x,
		y:           platform.y - BallSize,
		windowWidth: w, windowHeight: h,
		platform: &platform,
	}
	bricks := Bricks{
		context:      context,
		windowWidth:  w,
		windowHeight: h,
		ready:        false,
	}
	go bricks.Draw()

	// register mouse movement handler
	body.EventTarget().Listen(web.EventTypeMouseMove, platform.handleMouse)

	// register frame updaters
	handler := func() {
		wg := sync.WaitGroup{}
		wg.Add(4)
		go func() {
			fps.handle()
			wg.Done()
		}()
		go func() {
			platform.handleFrame()
			wg.Done()
		}()
		go func() {
			bricks.Handle(&ball)
			wg.Done()
		}()
		go func() {
			ball.handle()
			wg.Done()
		}()
		wg.Wait()
	}
	window.RequestAnimationFrame(handler, true)

	// prevent ending of the program
	select {}
}