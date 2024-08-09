package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/adriffaud/go-rpg/entities"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	player      *entities.Player
	enemies     []*entities.Enemy
	potions     []*entities.Potion
	tilemapJSON *TilemapJSON
	tilemapImg  *ebiten.Image
	camera      *Camera
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Y -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Y += 2
	}

	for _, sprite := range g.enemies {
		if sprite.FollowsPlayer {
			if sprite.X < g.player.X {
				sprite.X += 1
			} else if sprite.X > g.player.X {
				sprite.X -= 1
			}
			if sprite.Y < g.player.Y {
				sprite.Y += 1
			} else if sprite.Y > g.player.Y {
				sprite.Y -= 1
			}
		}
	}

	for _, potion := range g.potions {
		if g.player.X > potion.X {
			g.player.Health += potion.HealAmt
			fmt.Printf("Picked up potion! Health: %d\n", g.player.Health)
		}
	}

	g.camera.FollowTarget(g.player.X+8, g.player.Y+8, 320, 240)
	g.camera.Constrain(float64(g.tilemapJSON.Layers[0].Width*16), float64(g.tilemapJSON.Layers[0].Height*16), 320, 240)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	for _, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {
			x := index % layer.Width
			y := index / layer.Width

			x *= 16
			y *= 16

			srcX := (id - 1) % 22
			srcY := (id - 1) / 22

			srcX *= 16
			srcY *= 16

			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(g.camera.X, g.camera.Y)
			screen.DrawImage(g.tilemapImg.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image), &opts)
			opts.GeoM.Reset()
		}
	}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.camera.X, g.camera.Y)
	screen.DrawImage(g.player.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
	opts.GeoM.Reset()

	for _, enemy := range g.enemies {
		opts.GeoM.Translate(enemy.X, enemy.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)
		screen.DrawImage(enemy.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
		opts.GeoM.Reset()
	}

	for _, potion := range g.potions {
		opts.GeoM.Translate(potion.X, potion.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)
		screen.DrawImage(potion.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
		opts.GeoM.Reset()
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/ninja.png")
	if err != nil {
		log.Fatal(err)
	}
	skeletonImg, _, err := ebitenutil.NewImageFromFile("assets/images/skeleton.png")
	if err != nil {
		log.Fatal(err)
	}
	potionImg, _, err := ebitenutil.NewImageFromFile("assets/images/potion.png")
	if err != nil {
		log.Fatal(err)
	}
	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/TilesetFloor.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("assets/maps/spawn.json")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		player: &entities.Player{Sprite: &entities.Sprite{Img: playerImg, X: 50, Y: 50}, Health: 3},
		enemies: []*entities.Enemy{
			{Sprite: &entities.Sprite{Img: skeletonImg, X: 100, Y: 100}, FollowsPlayer: true},
			{Sprite: &entities.Sprite{Img: skeletonImg, X: 150, Y: 150}, FollowsPlayer: false},
			{Sprite: &entities.Sprite{Img: skeletonImg, X: 75, Y: 75}, FollowsPlayer: false},
		},
		potions: []*entities.Potion{
			{Sprite: &entities.Sprite{Img: potionImg, X: 210, Y: 100}, HealAmt: 1},
		},
		tilemapJSON: tilemapJSON,
		tilemapImg:  tilemapImg,
		camera:      NewCamera(0, 0),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
