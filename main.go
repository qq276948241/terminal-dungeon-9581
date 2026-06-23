package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

const (
	MapWidth  = 50
	MapHeight = 25
	SideBarWidth = 25

	TileWall   = '#'
	TileFloor  = '.'
	TileStairs = '>'

	PlayerChar  = '@'
	MonsterChar = 'M'
	ChestChar   = 'C'
)

type Tile byte

type Room struct {
	X, Y, W, H int
}

type Player struct {
	X, Y int
	HP   int
	MaxHP int
	Atk  int
}

type Monster struct {
	X, Y int
	HP   int
	Atk  int
	Alive bool
}

type Chest struct {
	X, Y int
	Opened bool
}

type Game struct {
	Map     [][]Tile
	Rooms   []Room
	Player  Player
	Monsters []Monster
	Chests   []Chest
	StairsX, StairsY int
	Floor    int
	Log      []string
	GameOver bool
}

func NewGame() *Game {
	g := &Game{
		Floor:    1,
		GameOver: false,
		Log:      make([]string, 0, 10),
	}
	g.Player = Player{HP: 100, MaxHP: 100, Atk: 10}
	g.GenerateFloor()
	return g
}

func (g *Game) GenerateFloor() {
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
	g.Log = g.Log[:0]
	g.addLog(fmt.Sprintf("=== 第 %d 层 ===", g.Floor))

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
			g.carveRoom(newRoom)
			if len(g.Rooms) > 0 {
				prev := g.Rooms[len(g.Rooms)-1]
				prevCX := prev.X + prev.W/2
				prevCY := prev.Y + prev.H/2
				newCX := newRoom.X + newRoom.W/2
				newCY := newRoom.Y + newRoom.H/2
				if rand.Intn(2) == 0 {
					g.carveHTunnel(prevCX, newCX, prevCY)
					g.carveVTunnel(prevCY, newCY, newCX)
				} else {
					g.carveVTunnel(prevCY, newCY, prevCX)
					g.carveHTunnel(prevCX, newCX, newCY)
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
			baseHP := 20 + g.Floor*10
			baseAtk := 5 + g.Floor*2
			g.Monsters = append(g.Monsters, Monster{
				X: mx, Y: my,
				HP: baseHP + rand.Intn(10),
				Atk: baseAtk + rand.Intn(3),
				Alive: true,
			})
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

func (g *Game) carveRoom(r Room) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			g.Map[y][x] = TileFloor
		}
	}
}

func (g *Game) carveHTunnel(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		g.Map[y][x] = TileFloor
	}
}

func (g *Game) carveVTunnel(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		g.Map[y][x] = TileFloor
	}
}

func (g *Game) isWalkable(x, y int) bool {
	if x < 0 || x >= MapWidth || y < 0 || y >= MapHeight {
		return false
	}
	t := g.Map[y][x]
	return t == TileFloor || t == TileStairs
}

func (g *Game) monsterAt(x, y int) *Monster {
	for i := range g.Monsters {
		if g.Monsters[i].Alive && g.Monsters[i].X == x && g.Monsters[i].Y == y {
			return &g.Monsters[i]
		}
	}
	return nil
}

func (g *Game) chestAt(x, y int) *Chest {
	for i := range g.Chests {
		if !g.Chests[i].Opened && g.Chests[i].X == x && g.Chests[i].Y == y {
			return &g.Chests[i]
		}
	}
	return nil
}

func (g *Game) addLog(msg string) {
	g.Log = append(g.Log, msg)
	if len(g.Log) > 8 {
		g.Log = g.Log[len(g.Log)-8:]
	}
}

func (g *Game) MovePlayer(dx, dy int) {
	if g.GameOver {
		return
	}
	nx := g.Player.X + dx
	ny := g.Player.Y + dy

	m := g.monsterAt(nx, ny)
	if m != nil {
		dmg := g.Player.Atk + rand.Intn(5)
		m.HP -= dmg
		g.addLog(fmt.Sprintf("你攻击怪物造成 %d 伤害", dmg))
		if m.HP <= 0 {
			m.Alive = false
			g.addLog("怪物被消灭了！")
		}
		g.monstersTurn()
		return
	}

	c := g.chestAt(nx, ny)
	if c != nil {
		c.Opened = true
		if rand.Intn(2) == 0 {
			heal := 15 + rand.Intn(20)
			g.Player.HP += heal
			if g.Player.HP > g.Player.MaxHP {
				g.Player.MaxHP = g.Player.HP
			}
			g.addLog(fmt.Sprintf("宝箱！恢复 %d HP", heal))
		} else {
			bonus := 2 + rand.Intn(4)
			g.Player.Atk += bonus
			g.addLog(fmt.Sprintf("宝箱！攻击力 +%d", bonus))
		}
		g.monstersTurn()
		return
	}

	if !g.isWalkable(nx, ny) {
		return
	}

	g.Player.X = nx
	g.Player.Y = ny

	if nx == g.StairsX && ny == g.StairsY {
		g.Floor++
		g.addLog("进入下一层...")
		g.GenerateFloor()
		return
	}

	g.monstersTurn()
}

func (g *Game) monstersTurn() {
	for i := range g.Monsters {
		m := &g.Monsters[i]
		if !m.Alive {
			continue
		}
		absX := abs(g.Player.X - m.X)
		absY := abs(g.Player.Y - m.Y)
		if absX+absY == 1 {
			dmg := m.Atk + rand.Intn(3)
			g.Player.HP -= dmg
			g.addLog(fmt.Sprintf("怪物攻击你造成 %d 伤害", dmg))
			if g.Player.HP <= 0 {
				g.Player.HP = 0
				g.GameOver = true
				g.addLog("你死了！按 r 重新开始")
				return
			}
			continue
		}
		if absX+absY > 8 {
			continue
		}
		dx, dy := 0, 0
		if absX > absY {
			if g.Player.X > m.X {
				dx = 1
			} else {
				dx = -1
			}
		} else {
			if g.Player.Y > m.Y {
				dy = 1
			} else {
				dy = -1
			}
		}
		nx := m.X + dx
		ny := m.Y + dy
		if g.isWalkable(nx, ny) && g.monsterAt(nx, ny) == nil &&
			!(nx == g.Player.X && ny == g.Player.Y) {
			m.X = nx
			m.Y = ny
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) Render() {
	clearScreen()

	totalWidth := MapWidth + SideBarWidth + 3
	fmt.Println("+" + repeat("-", totalWidth-2) + "+")

	for y := 0; y < MapHeight; y++ {
		line := "| "
		for x := 0; x < MapWidth; x++ {
			ch := byte(g.Map[y][x])

			if g.Player.X == x && g.Player.Y == y {
				ch = PlayerChar
			} else if g.monsterAt(x, y) != nil {
				ch = MonsterChar
			} else if g.chestAt(x, y) != nil {
				ch = ChestChar
			}
			line += string(ch)
		}

		line += " | "
		line += g.sidebarLine(y)
		for len(line) < totalWidth-1 {
			line += " "
		}
		line += "|"
		fmt.Println(line)
	}

	logAreaStart := MapHeight
	logAreaHeight := 10
	for i := 0; i < logAreaHeight; i++ {
		line := "| "
		for x := 0; x < MapWidth; x++ {
			line += " "
		}
		line += " | "
		y := logAreaStart + i
		if i < len(g.Log) {
			msg := g.Log[len(g.Log)-1-i]
			if len(msg) > SideBarWidth {
				msg = msg[:SideBarWidth]
			}
			line += msg
		}
		for len(line) < totalWidth-1 {
			line += " "
		}
		line += "|"
		if y < logAreaStart+logAreaHeight {
			fmt.Println(line)
		}
	}

	fmt.Println("+" + repeat("-", totalWidth-2) + "+")
	fmt.Println("  移动: WASD | 退出: Q | 重开: R")
}

func (g *Game) sidebarLine(y int) string {
	switch y {
	case 0:
		return "== 角色状态 =="
	case 1:
		return fmt.Sprintf("楼层: %d", g.Floor)
	case 2:
		return fmt.Sprintf("HP: %d / %d", g.Player.HP, g.Player.MaxHP)
	case 3:
		barLen := 20
		hp := g.Player.HP
		max := g.Player.MaxHP
		if max == 0 {
			max = 1
		}
		filled := hp * barLen / max
		if filled < 0 {
			filled = 0
		}
		bar := "[" + repeat("=", filled) + repeat(" ", barLen-filled) + "]"
		return bar
	case 4:
		return fmt.Sprintf("攻击力: %d", g.Player.Atk)
	case 5:
		return ""
	case 6:
		return "== 图例 =="
	case 7:
		return fmt.Sprintf("%c 你", PlayerChar)
	case 8:
		return fmt.Sprintf("%c 怪物", MonsterChar)
	case 9:
		return fmt.Sprintf("%c 宝箱", ChestChar)
	case 10:
		return fmt.Sprintf("%c 楼梯", TileStairs)
	case 11:
		return fmt.Sprintf("%c 墙", TileWall)
	case 12:
		return fmt.Sprintf("%c 地板", TileFloor)
	case 24:
		if g.GameOver {
			return "*** 游戏结束 ***"
		}
		return ""
	default:
		return ""
	}
}

func repeat(s string, n int) string {
	res := ""
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func readKey() byte {
	var b [3]byte
	os.Stdin.Read(b[:])
	return b[0]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	exec.Command("cmd", "/c", "mode con cols=80 lines=40").Run()

	g := NewGame()

	for {
		g.Render()
		fmt.Print("\n> ")
		k := readKey()
		k = toUpper(k)

		if g.GameOver {
			if k == 'R' {
				g = NewGame()
			} else if k == 'Q' {
				break
			}
			continue
		}

		switch k {
		case 'W':
			g.MovePlayer(0, -1)
		case 'S':
			g.MovePlayer(0, 1)
		case 'A':
			g.MovePlayer(-1, 0)
		case 'D':
			g.MovePlayer(1, 0)
		case 'Q':
			return
		case 'R':
			g = NewGame()
		}
	}
}

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - ('a' - 'A')
	}
	return b
}
