package main

import (
	"fmt"
	"math/rand"
)

type MapGenerator struct {
	Game *Game
}

func NewMapGenerator(g *Game) *MapGenerator {
	return &MapGenerator{Game: g}
}

func (mg *MapGenerator) GenerateFloor() {
	g := mg.Game
	g.Map = make([][]Tile, MapHeight)
	for y := range g.Map {
		g.Map[y] = make([]Tile, MapWidth)
		for x := range g.Map[y] {
			g.Map[y][x] = TileWall
		}
	}

	g.Rooms = nil
	g.Monsters = nil
	g.Chests = nil
	g.Potions = nil
	g.Log = g.Log[:0]
	g.AddLog(fmt.Sprintf("=== 第 %d 层 ===", g.Floor))

	numRooms := 6 + rand.Intn(5)
	for i := 0; i < numRooms; i++ {
		w := 5 + rand.Intn(6)
		h := 4 + rand.Intn(4)
		x := 1 + rand.Intn(MapWidth - w - 2)
		y := 1 + rand.Intn(MapHeight - h - 2)
		newRoom := Room{X: x, Y: y, W: w, H: h}

		overlap := false
		for _, r := range g.Rooms {
			if newRoom.X < r.X+r.W+1 && newRoom.X+newRoom.W+1 > r.X &&
				newRoom.Y < r.Y+r.H+1 && newRoom.Y+newRoom.H+1 > r.Y {
				overlap = true
				break
			}
		}
		if !overlap {
			mg.carveRoom(newRoom)
			if len(g.Rooms) > 0 {
				prev := g.Rooms[len(g.Rooms)-1]
				prevCX := prev.X + prev.W/2
				prevCY := prev.Y + prev.H/2
				newCX := newRoom.X + newRoom.W/2
				newCY := newRoom.Y + newRoom.H/2
				if rand.Intn(2) == 0 {
					mg.carveHTunnel(prevCX, newCX, prevCY)
					mg.carveVTunnel(prevCY, newCY, newCX)
				} else {
					mg.carveVTunnel(prevCY, newCY, prevCX)
					mg.carveHTunnel(prevCX, newCX, newCY)
				}
			}
			g.Rooms = append(g.Rooms, newRoom)
		}
	}

	first := g.Rooms[0]
	g.Player.X = first.X + first.W/2
	g.Player.Y = first.Y + first.H/2

	last := g.Rooms[len(g.Rooms)-1]
	g.StairsX = last.X + last.W/2
	g.StairsY = last.Y + last.H/2
	g.Map[g.StairsY][g.StairsX] = TileStairs

	for i := 1; i < len(g.Rooms); i++ {
		r := g.Rooms[i]
		numMonsters := 1 + rand.Intn(2)
		for j := 0; j < numMonsters; j++ {
			mx := r.X + rand.Intn(r.W)
			my := r.Y + rand.Intn(r.H)
			if mx == g.StairsX && my == g.StairsY {
				continue
			}
			g.Monsters = append(g.Monsters, NewMonster(mx, my, g.Floor))
		}
		if rand.Intn(3) == 0 {
			cx := r.X + rand.Intn(r.W)
			cy := r.Y + rand.Intn(r.H)
			if cx == g.StairsX && cy == g.StairsY {
				continue
			}
			g.Chests = append(g.Chests, Chest{X: cx, Y: cy, Opened: false})
		}
	}
}

func (mg *MapGenerator) carveRoom(r Room) {
	g := mg.Game
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			g.Map[y][x] = TileFloor
		}
	}
}

func (mg *MapGenerator) carveHTunnel(x1, x2, y int) {
	g := mg.Game
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		g.Map[y][x] = TileFloor
	}
}

func (mg *MapGenerator) carveVTunnel(y1, y2, x int) {
	g := mg.Game
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		g.Map[y][x] = TileFloor
	}
}

func (g *Game) IsWalkable(x, y int) bool {
	if x < 0 || x >= MapWidth || y < 0 || y >= MapHeight {
		return false
	}
	t := g.Map[y][x]
	return t == TileFloor || t == TileStairs
}

func (g *Game) MonsterAt(x, y int) *Monster {
	for i := range g.Monsters {
		if g.Monsters[i].Alive && g.Monsters[i].X == x && g.Monsters[i].Y == y {
			return &g.Monsters[i]
		}
	}
	return nil
}

func (g *Game) ChestAt(x, y int) *Chest {
	for i := range g.Chests {
		if !g.Chests[i].Opened && g.Chests[i].X == x && g.Chests[i].Y == y {
			return &g.Chests[i]
		}
	}
	return nil
}

func (g *Game) PotionAt(x, y int) int {
	for i := range g.Potions {
		if g.Potions[i].X == x && g.Potions[i].Y == y {
			return i
		}
	}
	return -1
}
