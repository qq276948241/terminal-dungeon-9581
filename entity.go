package main

import (
	"fmt"
	"math/rand"
)

const (
	GoblinChar  = 'G'
	SkeletonChar = 'S'
	DragonChar  = 'D'
	PotionChar  = '!'
)

type MonsterType int

const (
	TypeGoblin MonsterType = iota
	TypeSkeleton
	TypeDragon
)

type MonsterTypeInfo struct {
	Char    byte
	Name    string
	BaseHP  int
	BaseAtk int
	DropRate float64
}

var MonsterTypes = map[MonsterType]MonsterTypeInfo{
	TypeGoblin:   {Char: GoblinChar, Name: "哥布林", BaseHP: 15, BaseAtk: 4, DropRate: 0.3},
	TypeSkeleton: {Char: SkeletonChar, Name: "骷髅", BaseHP: 30, BaseAtk: 8, DropRate: 0.4},
	TypeDragon:   {Char: DragonChar, Name: "小龙", BaseHP: 60, BaseAtk: 15, DropRate: 0.6},
}

type Monster struct {
	X, Y int
	HP   int
	Atk  int
	Alive bool
	Type  MonsterType
}

type Chest struct {
	X, Y int
	Opened bool
}

type Potion struct {
	X, Y int
}

func (m *Monster) Char() byte {
	return MonsterTypes[m.Type].Char
}

func (m *Monster) Name() string {
	return MonsterTypes[m.Type].Name
}

func pickMonsterType(floor int) MonsterType {
	goblinW := 10 - floor*2
	if goblinW < 1 {
		goblinW = 1
	}
	skeletonW := floor * 2
	if skeletonW < 1 {
		skeletonW = 1
	}
	dragonW := 0
	if floor >= 3 {
		dragonW = floor - 2
	}
	total := goblinW + skeletonW + dragonW
	r := rand.Intn(total)
	if r < goblinW {
		return TypeGoblin
	}
	r -= goblinW
	if r < skeletonW {
		return TypeSkeleton
	}
	return TypeDragon
}

func NewMonster(x, y, floor int) Monster {
	t := pickMonsterType(floor)
	info := MonsterTypes[t]
	return Monster{
		X: x, Y: y,
		HP:    info.BaseHP + floor*8 + rand.Intn(8),
		Atk:   info.BaseAtk + floor*2 + rand.Intn(3),
		Alive: true,
		Type:  t,
	}
}

func (m *Monster) ShouldDrop() bool {
	return rand.Float64() < MonsterTypes[m.Type].DropRate
}

func NewPlayer() Player {
	return Player{HP: 100, MaxHP: 100, Atk: 10}
}

type Player struct {
	X, Y int
	HP    int
	MaxHP int
	Atk   int
	Potions int
}

func (p *Player) DrinkPotion() (int, string) {
	if p.Potions <= 0 {
		return 0, "没有药水了！"
	}
	heal := 30 + rand.Intn(20)
	p.Potions--
	oldHP := p.HP
	p.HP += heal
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	actualHeal := p.HP - oldHP
	return actualHeal, fmt.Sprintf("喝下药水，恢复 %d HP", actualHeal)
}
