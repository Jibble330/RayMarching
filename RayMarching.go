package main

import (
    "fmt"
    "time"
    "math"
    "image/color"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/imdraw"
    "github.com/faiface/pixel/pixelgl"
)

const (
    MIN_RAY_DIST = 0.001
    MAX_RAY_DIST = 2203 //sqrt(1920^2 + 1080^2)
)

var (
    IN_SHAPE = color.RGBA{0, 0, 0, 255}
    OUT_SHAPE = color.RGBA{0, 0, 0, 255}

    IN_RAY = color.RGBA{20, 20, 20, 20}
    OUT_RAY = color.RGBA{30, 30, 30, 30}
    CENTER_RAY = color.RGBA{50, 50, 50, 100}
)

var (
    win *pixelgl.Window
    imd *imdraw.IMDraw

    shapes []Shape
)

type Shape interface {
    Dist(p pixel.Vec) float64
    Draw()
}

type Circle struct {
    Position pixel.Vec
    Radius float64
    Inner, Outline color.Color
}

func NewCircle(Position pixel.Vec, Radius float64, Inner, Outline color.Color) *Circle {
    return &Circle{Position, Radius, Inner, Outline}
}

func (c *Circle) Draw() {
    imd.Color = c.Inner
    imd.Push(c.Position)
    imd.Circle(c.Radius, 0)

    imd.Color = c.Outline
    imd.Push(c.Position)
    imd.Circle(c.Radius, 3)
}

func (c *Circle) Dist(p pixel.Vec) float64 {
    return c.Position.Sub(p).Len() - c.Radius
}

type Rect struct {
    Min, Max pixel.Vec
    Inner, Outline color.Color
}

func NewRect(Min, Max pixel.Vec, Inner, Outline color.Color) *Rect {
    return &Rect{Min, Max, Inner, Outline}
}

func (r *Rect) Draw() {
    imd.Color = r.Inner
    imd.Push(r.Min, r.Max)
    imd.Rectangle(0)

    imd.Color = r.Outline
    imd.Push(r.Min, r.Max)
    imd.Rectangle(3)
}

func (r *Rect) Dist(p pixel.Vec) float64 {
    dx := math.Max(math.Max(r.Min.X - p.X, p.X - r.Max.X), 0)
    dy := math.Max(math.Max(r.Min.Y - p.Y, p.Y - r.Max.Y), 0)
    return math.Sqrt(dx*dx + dy*dy)
}

type Ray struct {
    Position pixel.Vec
    Direction pixel.Vec
}

func NewRay(pos pixel.Vec, rad float64) *Ray {
    return &Ray{pos, pixel.V(math.Cos(rad), math.Sin(rad))}
}

func (r *Ray) MinDist() float64 {
    var min float64
    for i, s := range shapes {
        dist := s.Dist(r.Position)
        if i == 0 || dist < min {
            min = dist
        }
    }
    return min
}

func (r *Ray) March() (hit bool, dist float64) {
    for {
        r.Draw()

        min := r.MinDist()
        if min < MIN_RAY_DIST {
            return true, dist
        }
        if min > MAX_RAY_DIST {
            return false, -1
        }
        

        r.Position = r.Position.Add(r.Direction.Scaled(min))
        dist += min
    }
}

func (r *Ray) Draw() {
    imd.Color = CENTER_RAY
    imd.Push(r.Position)
    imd.Circle(5, 0)

    dist := r.MinDist()

    imd.Color = IN_RAY
    imd.Push(r.Position)
    imd.Circle(dist, 0)

    imd.Color = OUT_RAY
    imd.Push(r.Position)
    imd.Circle(dist, 3)
}

func run() {
    monitor := pixelgl.PrimaryMonitor()
    PositionX, PositionY  := monitor.Position()
    SizeX, SizeY := monitor.Size()
    screen := pixel.R(PositionX, PositionY, SizeX, SizeY)

    cfg := pixelgl.WindowConfig{
        Title:   "RayMarching",
        Bounds:  screen,
        Maximized: true,
    }
    var err error
    win, err = pixelgl.NewWindow(cfg)
    if err != nil {
        panic(err)
    }

    imd = imdraw.New(nil)
    fps := time.NewTicker(time.Second/60)
    defer fps.Stop()

    r1 := NewRect(pixel.V(20, 20), pixel.V(300, 300), IN_SHAPE, OUT_SHAPE)
    r2 := NewRect(pixel.V(1400, 300), pixel.V(1700, 400), IN_SHAPE, OUT_SHAPE)
    ci := NewCircle(pixel.V(1000, 1000), 40, IN_SHAPE, OUT_SHAPE)
    
    orig := pixel.V(1920/2, 1080/2)    

    shapes = append(shapes, Shape(r1), Shape(r2), Shape(ci))

    for !win.Closed() {
        
        imd.Clear()

        if win.JustPressed(pixelgl.KeyEscape) {
            win.SetClosed(true)
        }

        if win.Pressed(pixelgl.KeyW) {
            orig = orig.Add(pixel.V(0, 10))
        }
        if win.Pressed(pixelgl.KeyS) {
            orig = orig.Sub(pixel.V(0, 10))
        }
        if win.Pressed(pixelgl.KeyD) {
            orig = orig.Add(pixel.V(10, 0))
        }
        if win.Pressed(pixelgl.KeyA) {
            orig = orig.Sub(pixel.V(10, 0))
        }

        mouse := win.MousePosition()
        var min float64
        for i, s := range shapes {
            dist := s.Dist(mouse)
            if i == 0 || dist < min {
                min = dist
            }
        }

        dif := mouse.Sub(orig)
        r := NewRay(orig, math.Atan2(dif.Y, dif.X))
        fmt.Println(r.March())

        for _, s := range shapes {
            s.Draw()
        }

        win.Clear(color.RGBA{20, 20, 20, 255})
        imd.Draw(win)
        win.Update()
        <-fps.C
    }
}

func main() {
    pixelgl.Run(run)
}