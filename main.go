package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

type Game struct {
	Map     [][]Tile
	Rooms   []Room
	Player  Player
	Monsters []Monster
	Chests   []Chest
	Potions  []Potion
	StairsX, StairsY int
	Floor    int
	Log      []string
	GameOver bool
	MapGen   *MapGenerator
	Renderer *Renderer
}

func NewGame() *Game {
	g := &Game{
		Floor:    1,
		GameOver: false,
		Log:      make([]string, 0, 10),
	}
	g.Player = NewPlayer()
	g.MapGen = NewMapGenerator(g)
	g.Renderer = NewRenderer(g)
	g.MapGen.GenerateFloor()
	return g
}

func (g *Game) AddLog(msg string) {
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

	m := g.MonsterAt(nx, ny)
	if m != nil {
		dmg := g.Player.Atk + rand.Intn(5)
		m.HP -= dmg
		g.AddLog(fmt.Sprintf("你攻击%s造成 %d 伤害", m.Name(), dmg))
		if m.HP <= 0 {
			m.Alive = false
			g.AddLog(fmt.Sprintf("%s被消灭了！", m.Name()))
			if m.ShouldDrop() {
				g.Potions = append(g.Potions, Potion{X: m.X, Y: m.Y})
				g.AddLog("怪物掉落了一瓶药水！")
			}
		}
		g.monstersTurn()
		return
	}

	c := g.ChestAt(nx, ny)
	if c != nil {
		c.Opened = true
		if rand.Intn(2) == 0 {
			heal := 15 + rand.Intn(20)
			g.Player.HP += heal
			if g.Player.HP > g.Player.MaxHP {
				g.Player.MaxHP = g.Player.HP
			}
			g.AddLog(fmt.Sprintf("宝箱！恢复 %d HP", heal))
		} else {
			bonus := 2 + rand.Intn(4)
			g.Player.Atk += bonus
			g.AddLog(fmt.Sprintf("宝箱！攻击力 +%d", bonus))
		}
		g.monstersTurn()
		return
	}

	if !g.IsWalkable(nx, ny) {
		return
	}

	g.Player.X = nx
	g.Player.Y = ny

	pi := g.PotionAt(nx, ny)
	if pi >= 0 {
		g.Player.Potions++
		g.Potions = append(g.Potions[:pi], g.Potions[pi+1:]...)
		g.AddLog(fmt.Sprintf("捡到药水！当前 %d 瓶", g.Player.Potions))
	}

	if nx == g.StairsX && ny == g.StairsY {
		g.Floor++
		g.AddLog("进入下一层...")
		g.MapGen.GenerateFloor()
		return
	}

	g.monstersTurn()
}

func (g *Game) UsePotion() {
	if g.GameOver {
		return
	}
	_, msg := g.Player.DrinkPotion()
	g.AddLog(msg)
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
			g.AddLog(fmt.Sprintf("%s攻击你造成 %d 伤害", m.Name(), dmg))
			if g.Player.HP <= 0 {
				g.Player.HP = 0
				g.GameOver = true
				g.AddLog("你死了！按 r 重新开始")
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
		if g.IsWalkable(nx, ny) && g.MonsterAt(nx, ny) == nil &&
			!(nx == g.Player.X && ny == g.Player.Y) && g.PotionAt(nx, ny) < 0 {
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

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - ('a' - 'A')
	}
	return b
}

func main() {
	rand.Seed(time.Now().UnixNano())
	exec.Command("cmd", "/c", "mode con cols=80 lines=40").Run()

	g := NewGame()

	for {
		g.Renderer.Render()
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
		case 'P':
			g.UsePotion()
		case 'Q':
			return
		case 'R':
			g = NewGame()
		}
	}
}
