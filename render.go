package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Renderer struct {
	Game *Game
}

func NewRenderer(g *Game) *Renderer {
	return &Renderer{Game: g}
}

func (r *Renderer) Render() {
	clearScreen()
	g := r.Game

	totalWidth := MapWidth + SideBarWidth + 3
	fmt.Println("+" + repeat("-", totalWidth-2) + "+")

	for y := 0; y < MapHeight; y++ {
		line := "| "
		for x := 0; x < MapWidth; x++ {
			ch := byte(g.Map[y][x])

			if g.Player.X == x && g.Player.Y == y {
				ch = PlayerChar
			} else if m := g.MonsterAt(x, y); m != nil {
				ch = m.Char()
			} else if g.ChestAt(x, y) != nil {
				ch = ChestChar
			} else if g.PotionAt(x, y) >= 0 {
				ch = PotionChar
			}
			line += string(ch)
		}

		line += " | "
		line += r.sidebarLine(y)
		for len(line) < totalWidth-1 {
			line += " "
		}
		line += "|"
		fmt.Println(line)
	}

	logAreaHeight := 10
	logCopy := make([]string, len(g.Log))
	copy(logCopy, g.Log)
	for i := 0; i < logAreaHeight; i++ {
		line := "| "
		for x := 0; x < MapWidth; x++ {
			line += " "
		}
		line += " | "
		logLen := len(logCopy)
		if logLen > 0 && i < logLen {
			idx := logLen - 1 - i
			if idx >= 0 && idx < logLen {
				msg := logCopy[idx]
				if len(msg) > SideBarWidth {
					msg = msg[:SideBarWidth]
				}
				line += msg
			}
		}
		for len(line) < totalWidth-1 {
			line += " "
		}
		line += "|"
		fmt.Println(line)
	}

	fmt.Println("+" + repeat("-", totalWidth-2) + "+")
	fmt.Println("  移动: WASD | 喝药: P | 退出: Q | 重开: R")
}

func (r *Renderer) sidebarLine(y int) string {
	g := r.Game
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
		return fmt.Sprintf("药水: %d 瓶 (P键使用)", g.Player.Potions)
	case 6:
		return ""
	case 7:
		return "== 图例 =="
	case 8:
		return fmt.Sprintf("%c 你", PlayerChar)
	case 9:
		return fmt.Sprintf("%c 哥布林  %c 骷髅", GoblinChar, SkeletonChar)
	case 10:
		return fmt.Sprintf("%c 小龙    %c 宝箱", DragonChar, ChestChar)
	case 11:
		return fmt.Sprintf("%c 药水    %c 楼梯", PotionChar, TileStairs)
	case 12:
		return fmt.Sprintf("%c 墙      %c 地板", TileWall, TileFloor)
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
