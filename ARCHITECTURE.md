# Roguelike 地牢探险 - 架构说明

写给接手这个项目的你，尽量口语化，能看懂就好。

---

## 文件总览

| 文件 | 管什么 | 核心内容 |
|------|--------|----------|
| [entity.go](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/entity.go) | 所有实体定义 + 常量 | Player/Monster/Chest/Potion 结构体、怪物类型配置、属性计算 |
| [mapgen.go](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/mapgen.go) | 地图生成 + 位置查询 | 房间+走廊算法、怪物/宝箱/楼梯放置、IsWalkable/MonsterAt 等查询方法 |
| [render.go](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/render.go) | 终端显示 | 清屏、地图绘制、侧边栏（HP/攻击/药水/图例）、日志栏、读键盘 |
| [main.go](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/main.go) | 胶水层 + 游戏逻辑 | Game 结构体、战斗/移动/捡物/喝药逻辑、主循环 |
| [go.mod](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/go.mod) | 依赖 | 零外部依赖，纯标准库 |

整个项目没有用任何第三方库，`go build .` 直接出 exe。

---

## 它们怎么串起来的：Game 结构体

核心就是 [Game](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/main.go#L10-L23)，它是全局状态的大总管，也是各个模块之间的桥梁。

```
Game
├── Map / Rooms / StairsX / StairsY      ← 地图数据（来自 mapgen）
├── Player / Monsters / Chests / Potions  ← 场上所有实体（定义在 entity）
├── Floor / Log / GameOver                ← 游戏进度
├── MapGen   *MapGenerator                ← 组合：管地图生成
└── Renderer *Renderer                    ← 组合：管屏幕渲染
```

**组合模式**：Game 里不是继承 MapGenerator 和 Renderer，而是把它们当字段嵌进去。NewGame 时一起构造：

```go
func NewGame() *Game {
    g := &Game{...}
    g.Player = NewPlayer()          // entity.go 里的工厂
    g.MapGen   = NewMapGenerator(g) // 把自己传进去，让 mapgen 能改 g 的字段
    g.Renderer = NewRenderer(g)     // 同上，让 renderer 能读 g 的数据
    g.MapGen.GenerateFloor()        // 先生成第 1 层
    return g
}
```

**数据流**：
- `mapgen.go` 写 Game 的字段（改 Map、放怪物、设玩家坐标…）
- `main.go` 改 Game 的字段（战斗扣血、移动、捡物…）
- `render.go` **只读** Game 的字段来画图

因为是单线程轮询式主循环，没有并发，所以不用锁。

---

## 地图生成算法（房间 + 走廊）

实现位置：[mapgen.go GenerateFloor()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/mapgen.go#L16-L118)

### 流程

```
1. 初始化：整张地图填成墙 (#)
           清空 Rooms/Monsters/Chests/Potions/Log
           加一条 "=== 第 N 层 ===" 日志
                │
                ▼
2. 随机放房间（6~10 间，不重叠）
   ┌─ 随机宽高 (w=5~10, h=4~7)
   ├─ 随机坐标（不撞边界）
   ├─ 和已有房间做 AABB 碰撞检测（留 1 格间隙）
   ├─ 重叠就放弃重试
   └─ 不重叠：
        ├─ 把这块区域挖成地板 (.)
        ├─ 和上一个房间之间连 L 型走廊
        └─ 加入 Rooms 切片
                │
                ▼
3. 确定玩家出生点 = 第 1 间房的中心
   确定楼梯位置   = 最后一间房的中心 (>)
                │
                ▼
4. 在第 2~N 间房里放东西：
   ├─ 每间房随机 1~2 只怪（距玩家曼哈顿距离 ≥5 才放）
   ├─ 30% 概率放一个宝箱
   └─ 楼梯格不放任何东西
```

### L 型走廊怎么连

取上一个房间中心 (px, py) 和新房间中心 (nx, ny)，50% 概率先横后竖，50% 先竖后横：

```
先横后竖：      先竖后横：
  px───►nx       py
  py  .          │
      . ny       ▼
                 nx
```

挖走廊就是把路径上的墙改成地板。

---

## 三种怪物配置

定义在 [entity.go MonsterTypes](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/entity.go#L46-L50)：

| 符号 | 名字 | 基础HP | 基础ATK | 掉药率 | 出现条件 |
|------|------|--------|---------|--------|----------|
| G | 哥布林 | 15 | 4 | 30% | 任何层 |
| S | 骷髅   | 30 | 8 | 40% | 第2层起权重快速上升 |
| D | 小龙   | 60 | 15 | 60% | 第3层起开始出现 |

实际 HP/ATK 还会叠层数加成，见 [NewMonster()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/entity.go#L110-L119)：
```
最终HP  = 基础HP + 层数×8 + rand(0~7)
最终ATK = 基础ATK + 层数×2 + rand(0~2)
```

**刷怪权重**：[pickMonsterType()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/entity.go#L85-L108)
- 哥布林权重 = `max(1, 10 - 层数×2)`（越高层越少见）
- 骷髅权重   = `max(1, 层数×2)`
- 小龙权重   = `max(0, 层数-2)`（第3层起才有）
- 然后按权重随机 roll 一只

---

## 一帧的流程（主循环）

实现位置：[main.go main()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/main.go#L185-L220)

```
┌─────────────────────────────────────────┐
│  Renderer.Render()                      │  ← 清屏 + 画地图 + 侧边栏 + 日志
│  打印 "移动: WASD | 喝药: P..." 提示    │
│  阻塞等待键盘输入 (readKey)             │  ← 这行卡住，直到用户按键
└────────────────────┬────────────────────┘
                     │ 读到一个按键
                     ▼
           ┌─────────────────┐
           │ GameOver 了吗？ │────是──► 只处理 R(重开) Q(退出)
           └────────┬────────┘
                    │否
                    ▼
         ┌──WASD──► MovePlayer(dx, dy) ──┐
         │                                │
         ├──P─────► UsePotion()           │
         │                                ├──► 无论哪个动作，
         ├──R─────► NewGame() 重开        │     都会触发怪物行动
         │                                │     monstersTurn()
         └──Q─────► return 退出           │
                                          │
                                          ▼
                              回到循环开头（下一帧）
```

### MovePlayer 内部做了什么（按一下方向键触发）

实现位置：[MovePlayer()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/main.go#L51-L110)

```
计算目标格 (nx, ny)
    │
    ├─ 目标格有怪物？ ──是──► 攻击它（你先手）
    │                        ├─ 怪物死了？掉落药水
    │                        └─ 然后轮到怪物行动
    │
    ├─ 目标格有宝箱？ ──是──► 开宝箱（回血 or 加攻击）
    │                        └─ 然后轮到怪物行动
    │
    ├─ 是墙？─────是──► 什么都不做，return
    │
    └─ 能走 ──► 移动过去
         │
         ├─ 踩到药水？自动捡（背包+1，地上消失）
         │
         ├─ 踩到楼梯 > ？ 楼层+1，重新生成地图
         │
         └─ 轮到怪物行动
```

### 怪物回合 monstersTurn()

实现位置：[monstersTurn()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/main.go#L121-L166)

对每只活着的怪：
```
与玩家曼哈顿距离 = 1？
    │
    ├─ 是 ──► 攻击玩家（扣血检查死亡）
    │
    └─ 否 ──► 距离 > 8？不动（没发现你）
              │
              └─ 距离 ≤ 8？朝你走一步（贪心：选 X 或 Y 差距大的方向追）
                    └─ 走之前检查：不撞墙、不撞怪、不踩玩家、不踩药水
```

---

## 战斗伤害公式

玩家攻击怪物（MovePlayer 内）：
```
伤害 = 玩家ATK + rand(0~4)
怪物HP -= 伤害
```

怪物攻击玩家（monstersTurn 内）：
```
伤害 = 怪物ATK + rand(0~2)
玩家HP -= 伤害
玩家HP ≤ 0 → GameOver
```

都是**先手规则**：玩家先动，打得到怪就先打；怪反击是在怪物回合，前提是它没死且还在你旁边。

---

## 药水系统

**获得方式**：
1. 怪物死亡按掉药率 roll（哥布林30% / 骷髅40% / 小龙60%）→ 在怪死亡位置生成 `!`
2. 宝箱有 50% 概率给 HP 恢复（虽然不是药水，但效果类似）

**拾取**：玩家踩到药水格自动 `Potions++`，药水从地图消失。

**使用**：按 P 键调用 [Player.DrinkPotion()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/entity.go#L130-L143)
```
没药水 → 提示 "没有药水了！"
有药水：
    Potions--
    回血量 = 30 + rand(0~19)
    HP += 回血量（不超过 MaxHP）
    然后轮到怪物行动（喝药要消耗回合！）
```

注意：喝药也会触发怪物回合，残血时别浪。

---

## 渲染怎么画的

实现位置：[render.go Render()](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo27/project27/render.go#L17-L79)

整体布局（大概 80 列 × 40 行）：
```
+---------------------------------------------------------------+
|  地图区 (50列×25行)         |  侧边栏 (25列)                 |
|  ##########....#######...   |  == 角色状态 ==               |
|  #@........G....#........   |  楼层: 3                      |
|  #.............D#.....C.   |  HP: 73 / 120 [=========     ]|
|  #####+######...#.....M.   |  攻击力: 16                   |
|  ...               >.....  |  药水: 2 瓶 (P键使用)          |
|  ########################  |                                |
|                             |  == 图例 ==                   |
|                             |  @ 你   G 哥布林  S 骷髅      |
|                             |  D 小龙  C 宝箱   ! 药水      |
+-----------------------------+--------------------------------+
|  （地图下方空白，留给日志）  |  小龙攻击你造成 12 伤害       |
|                             |  你攻击小龙造成 19 伤害       |
|                             |  捡到药水！当前 2 瓶          |
|                             |  === 第 3 层 ===              |
+---------------------------------------------------------------+
  移动: WASD | 喝药: P | 退出: Q | 重开: R
```

**绘制优先级**（同一格有多个东西时谁先显示）：
```
玩家 @ > 怪物 G/S/D > 宝箱 C > 药水 ! > 楼梯 > > 地板/墙
```
这个判断在 Render 的内层循环里，从上到下按优先级覆盖。

**日志栏防 bug**：渲染前先 `copy` 一份 Log 快照，索引时做 `idx >= 0 && idx < logLen` 双校验，避免 Log 切片在战斗中被 append/截断导致越界 panic（之前踩过坑）。

---

## 改代码的话从哪下手

| 想改什么 | 去哪个文件 |
|----------|------------|
| 地图大小、房间数量、走廊样式 | `mapgen.go` 的 GenerateFloor |
| 再加一种怪物（比如 Boss） | `entity.go`：加 MonsterType 常量、配 MonsterTypes、改 pickMonsterType 权重 |
| 怪物平衡（血量/攻击/掉率） | `entity.go` 的 MonsterTypes 表 + NewMonster 加成公式 |
| 加新道具（武器/护甲） | `entity.go` 加结构体 + `mapgen.go` 生成 + `main.go` MovePlayer 踩到的逻辑 + `render.go` 显示字符 |
| 改战斗公式/先手规则 | `main.go` MovePlayer（玩家攻击）和 monstersTurn（怪物攻击） |
| 换一种 UI 布局/配色 | `render.go` 的 Render + sidebarLine |
| 药水回复量/效果 | `entity.go` 的 DrinkPotion |
| 存档功能（目前没有） | 新加一个 `save.go`，把 Game 结构体 json 序列化到文件就行 |

---

最后提醒：这个游戏是**同步阻塞式**的，readKey 会一直等按键，没有敌人会自己动。想做带时间条的实时 roguelike 的话就得把主循环换成非阻塞输入+定时器，改动会比较大。
