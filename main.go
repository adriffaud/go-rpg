package main

import (
	"image"
	"image/color"
	"log"

	"github.com/adriffaud/go-rpg/entities"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		rect := image.Rect(int(sprite.X), int(sprite.Y), int(sprite.X)+16, int(sprite.Y)+16)

		if collider.Overlaps(rect) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - 16
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		rect := image.Rect(int(sprite.X), int(sprite.Y), int(sprite.X)+16, int(sprite.Y)+16)

		if collider.Overlaps(rect) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - 16
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}

type Game struct {
	player      *entities.Player
	enemies     []*entities.Enemy
	potions     []*entities.Potion
	tilemapJSON *TilemapJSON
	tilesets    []Tileset
	tilemapImg  *ebiten.Image
	camera      *Camera
	colliders   []image.Rectangle
}

func (g *Game) Update() error {
	g.player.Dx = 0
	g.player.Dy = 0

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = 2
	}

	g.player.X += g.player.Dx
	CheckCollisionHorizontal(g.player.Sprite, g.colliders)
	g.player.Y += g.player.Dy
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	for _, sprite := range g.enemies {
		sprite.Dx = 0
		sprite.Dy = 0

		if sprite.FollowsPlayer {
			if sprite.X < g.player.X {
				sprite.Dx = 1
			} else if sprite.X > g.player.X {
				sprite.Dx = -1
			}
			if sprite.Y < g.player.Y {
				sprite.Dy = 1
			} else if sprite.Y > g.player.Y {
				sprite.Dy = -1
			}
		}

		sprite.X += sprite.Dx
		CheckCollisionHorizontal(sprite.Sprite, g.colliders)
		sprite.Y += sprite.Dy
		CheckCollisionVertical(sprite.Sprite, g.colliders)
	}

	for _, potion := range g.potions {
		if g.player.X > potion.X {
			g.player.Health += potion.HealAmt
			// fmt.Printf("Picked up potion! Health: %d\n", g.player.Health)
		}
	}

	g.camera.FollowTarget(g.player.X+8, g.player.Y+8, 320, 240)
	g.camera.Constrain(float64(g.tilemapJSON.Layers[0].Width*16), float64(g.tilemapJSON.Layers[0].Height*16), 320, 240)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	for layerIndex, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {
			if id == 0 {
				continue
			}

			x := index % layer.Width
			y := index / layer.Width

			x *= 16
			y *= 16

			img := g.tilesets[layerIndex].Img(id)
			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(0, -(float64(img.Bounds().Dy()) + 16))
			opts.GeoM.Translate(g.camera.X, g.camera.Y)
			screen.DrawImage(img, &opts)
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

	opts.GeoM.Reset()

	for _, potion := range g.potions {
		opts.GeoM.Translate(potion.X, potion.Y)
		opts.GeoM.Translate(g.camera.X, g.camera.Y)
		screen.DrawImage(potion.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
		opts.GeoM.Reset()
	}

	opts.GeoM.Reset()

	for _, collider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(collider.Min.X)+float32(g.camera.X),
			float32(collider.Min.Y)+float32(g.camera.Y),
			float32(collider.Dx()),
			float32(collider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			true)
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

	tilesets, err := tilemapJSON.GenTilesets()
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
		tilesets:    tilesets,
		camera:      NewCamera(0, 0),
		colliders: []image.Rectangle{
			image.Rect(100, 100, 116, 116)},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
