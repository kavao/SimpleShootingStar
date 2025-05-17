package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

// GameState はゲームの状態を表す定数
const (
	GameStateTitle = iota
	GameStatePlaying
	GameStateStageClear
	GameStatePlayerExplosion
	GameStateGameOver
)

// Bullet は弾の状態を保持する構造体です
type Bullet struct {
	x, y   float64
	vx, vy float64
}

// Star は背景の流れる星を表す構造体
type Star struct {
	x, y   float64
	speed  float64
	length float64
	color  color.RGBA
}

// EnemyType は敵の種類を表す定数
const (
	EnemyTypeStraight = iota // まっすぐ進む敵
	EnemyTypeSine            // サインカーブで動く敵
	EnemyTypeSpecial         // 特殊な動きをする敵
)

// EnemyBullet構造体を追加
type EnemyBullet struct {
	x, y   float64
	vx, vy float64
}

// Enemy は敵の状態を保持する構造体
type Enemy struct {
	x, y           float64
	speed          float64
	enemyType      int
	time           float64 // 時間経過（サインカーブ用）
	phase          int     // 特殊な動きのフェーズ
	hp             int     // 耐久度を追加
	shootsBullet   bool    // 弾を撃つ敵かどうか
	bulletType     int     // 0:主人公狙い, 1:真下, 2:斜め
	bulletCooldown int     // 弾発射クールダウン
	turnDirection  int     // 追加
}

// Wave は敵の出現パターンを表す構造体
type Wave struct {
	enemyType     int
	x             int
	delay         int
	shootsBullet  bool
	bulletType    int
	speed         float64 // 追加
	turnDirection int     // 追加（1:右, -1:左, 0:デフォルト右）
}

// Particle はパーティクルの状態を保持する構造体
type Particle struct {
	x, y     float64
	vx, vy   float64 // 速度
	size     float64 // サイズ
	alpha    float64 // 透明度
	lifetime int     // 生存時間
	ptype    int     // 0:通常, 1:発射ライン
}

// Stage はステージの情報を保持する構造体
type Stage struct {
	Name  string
	Waves []Wave
}

var stages = []Stage{
	{
		Name: "Stage 1: 基本編",
		Waves: []Wave{
			{enemyType: EnemyTypeStraight, x: 100, delay: 0, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeStraight, x: 320, delay: 30, shootsBullet: true, bulletType: 0}, // 主人公狙い
			{enemyType: EnemyTypeStraight, x: 540, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 200, delay: 60, shootsBullet: true, bulletType: 1}, // 真下
			{enemyType: EnemyTypeSine, x: 440, delay: 30, shootsBullet: false, bulletType: 0},
		},
	},
	{
		Name: "Stage 2: 波状攻撃",
		Waves: []Wave{
			{enemyType: EnemyTypeSine, x: 100, delay: 0, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 540, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 200, delay: 60, shootsBullet: true, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 440, delay: 30, shootsBullet: false, bulletType: 0, turnDirection: -1},
		},
	},
	{
		Name: "Stage 3: 特殊攻撃",
		Waves: []Wave{
			{enemyType: EnemyTypeSpecial, x: 100, delay: 0, shootsBullet: false, bulletType: 0, speed: 2.0, turnDirection: 1},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 40, shootsBullet: false, bulletType: 0, speed: 4.0, turnDirection: 1},  // 高速右
			{enemyType: EnemyTypeSpecial, x: 540, delay: 40, shootsBullet: false, bulletType: 0, speed: 2.0, turnDirection: -1}, // 左曲がり
			{enemyType: EnemyTypeStraight, x: 200, delay: 60, shootsBullet: true, bulletType: 0, speed: 2.0, turnDirection: 0},
			{enemyType: EnemyTypeSine, x: 440, delay: 30, shootsBullet: false, bulletType: 0, speed: 2.0, turnDirection: 0},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 30, shootsBullet: false, bulletType: 0, speed: 4.0, turnDirection: -1}, // 高速左
		},
	},
	{
		Name: "Stage 4: 複合攻撃",
		Waves: []Wave{
			{enemyType: EnemyTypeStraight, x: 100, delay: 0, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 540, delay: 30, shootsBullet: false, bulletType: 0, turnDirection: -1},
			{enemyType: EnemyTypeStraight, x: 200, delay: 60, shootsBullet: true, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 440, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeStraight, x: 540, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 100, delay: 30, shootsBullet: false, bulletType: 0},
		},
	},
	{
		Name: "Stage 5: 最終決戦",
		Waves: []Wave{
			{enemyType: EnemyTypeSpecial, x: 100, delay: 0, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 540, delay: 30, shootsBullet: false, bulletType: 0, turnDirection: -1},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 30, shootsBullet: false, bulletType: 0, turnDirection: -1},
			{enemyType: EnemyTypeSine, x: 100, delay: 60, shootsBullet: true, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSine, x: 540, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeStraight, x: 200, delay: 60, shootsBullet: true, bulletType: 0},
			{enemyType: EnemyTypeStraight, x: 320, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeStraight, x: 440, delay: 30, shootsBullet: false, bulletType: 0},
			{enemyType: EnemyTypeSpecial, x: 320, delay: 60, shootsBullet: false, bulletType: 0},
		},
	},
}

// Game はゲームの状態を保持する構造体です
type Game struct {
	playerX               float64
	playerY               float64
	bullets               []Bullet
	shootCooldown         int    // 連射防止用
	stars                 []Star // 星のスライスを追加
	enemies               []Enemy
	waves                 []Wave
	waveTimer             int
	currentSpawn          int
	score                 int
	gameState             int        // ゲームの状態
	highScore             int        // ハイスコア
	particles             []Particle // パーティクルを追加
	currentStage          int        // 現在のステージ番号
	stageClearTimer       int        // ステージクリア演出用
	stageClearKeyReleased bool       // ステージクリア画面でキーリリースを検知
	playerExplosionTimer  int        // 爆発演出用
	enemyBullets          []EnemyBullet
}

var (
	gameFont font.Face
)

// NewGame は新しいゲームインスタンスを作成します
func NewGame() *Game {
	// 星の色バリエーション
	starColors := []color.RGBA{
		{180, 180, 255, 100}, // 白
		{140, 180, 255, 100}, // 青白
		{100, 140, 255, 100}, // 青
		{200, 200, 255, 80},  // 明るい白
		{80, 120, 255, 80},   // 暗い青
	}
	stars := make([]Star, 60)
	for i := range stars {
		c := starColors[rand.Intn(len(starColors))]
		stars[i] = Star{
			x:      rand.Float64() * screenWidth,
			y:      rand.Float64() * screenHeight,
			speed:  2 + rand.Float64()*3,
			length: 8 + rand.Float64()*8,
			color:  c,
		}
	}

	return &Game{
		playerX:               screenWidth / 2,
		playerY:               screenHeight / 2 * 1.7,
		bullets:               []Bullet{},
		stars:                 stars,
		enemies:               []Enemy{},
		waves:                 stages[0].Waves,
		waveTimer:             0,
		currentSpawn:          0,
		score:                 0,
		gameState:             GameStateTitle,
		highScore:             0,
		particles:             []Particle{},
		currentStage:          0,
		stageClearTimer:       0,
		stageClearKeyReleased: false,
		playerExplosionTimer:  0,
		enemyBullets:          []EnemyBullet{},
	}
}

// createExplosion は爆発エフェクトのパーティクルを生成します
func (g *Game) createExplosion(x, y float64, color color.RGBA) {
	particleCount := 20
	for i := 0; i < particleCount; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 2 + rand.Float64()*3
		particle := Particle{
			x:        x,
			y:        y,
			vx:       math.Cos(angle) * speed,
			vy:       math.Sin(angle) * speed,
			size:     4 + rand.Float64()*4,
			alpha:    1.0,
			lifetime: 30 + rand.Intn(20),
			ptype:    0,
		}
		g.particles = append(g.particles, particle)
	}
}

// nextWave は次のウェーブに進みます
func (g *Game) nextWave() {
	g.currentSpawn = 0
	g.waveTimer = 0

	// 現在のステージの全ウェーブをクリアした場合
	if g.currentSpawn >= len(g.waves) {
		g.currentStage++
		// 全ステージクリア
		if g.currentStage >= len(stages) {
			g.gameState = GameStateGameOver
			if g.score > g.highScore {
				g.highScore = g.score
			}
			return
		}
		// 次のステージのウェーブを設定
		g.waves = stages[g.currentStage].Waves
		g.currentSpawn = 0
	}
}

// Update はゲームの状態を更新します
func (g *Game) Update() error {
	// 星の移動（どの状態でも動く）
	for i := range g.stars {
		g.stars[i].y += g.stars[i].speed
		if g.stars[i].y > screenHeight {
			g.stars[i].x = rand.Float64() * screenWidth
			g.stars[i].y = -g.stars[i].length
			g.stars[i].speed = 2 + rand.Float64()*3
			g.stars[i].length = 8 + rand.Float64()*8
		}
	}

	// パーティクルの更新（どの状態でも動く）
	newParticles := g.particles[:0]
	for _, p := range g.particles {
		if p.ptype != 1 {
			p.x += p.vx
			p.y += p.vy
			p.vy += 0.1 // 重力効果
		}
		p.alpha -= 1.0 / float64(p.lifetime)
		p.lifetime--
		if p.lifetime > 0 && p.alpha > 0 {
			newParticles = append(newParticles, p)
		}
	}
	g.particles = newParticles

	switch g.gameState {
	case GameStateTitle:
		// スペースキーでゲーム開始
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.gameState = GameStatePlaying
		}
	case GameStatePlaying:
		// 既存のゲームプレイ処理
		moveSpeed := 8.0
		// プレイヤーの移動処理
		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.playerX -= moveSpeed
			if g.playerX < 20 {
				g.playerX = 20
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.playerX += moveSpeed
			if g.playerX > screenWidth-40 {
				g.playerX = screenWidth - 40
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.playerY -= moveSpeed
			if g.playerY < 40 {
				g.playerY = 40
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.playerY += moveSpeed
			if g.playerY > screenHeight-20 {
				g.playerY = screenHeight - 20
			}
		}

		// 敵の出現処理
		if g.currentSpawn < len(g.waves) {
			// 累積delay方式
			totalDelay := 0
			for i := 0; i <= g.currentSpawn; i++ {
				totalDelay += g.waves[i].delay
			}
			if g.waveTimer >= totalDelay {
				wave := g.waves[g.currentSpawn]
				hp := 1
				switch wave.enemyType {
				case EnemyTypeStraight:
					hp = 2
				case EnemyTypeSine:
					hp = 3
				case EnemyTypeSpecial:
					hp = 4
				}
				speed := wave.speed
				if speed == 0 {
					speed = 2.0 // デフォルト
				}
				turnDir := wave.turnDirection
				if turnDir == 0 {
					turnDir = 1 // デフォルト右
				}
				enemy := Enemy{
					x:              float64(wave.x),
					y:              -20,
					speed:          speed,
					enemyType:      wave.enemyType,
					time:           0,
					phase:          0,
					hp:             hp,
					shootsBullet:   wave.shootsBullet,
					bulletType:     wave.bulletType,
					bulletCooldown: 60 + rand.Intn(60), // 1〜2秒ごとに発射
					turnDirection:  turnDir,
				}
				g.enemies = append(g.enemies, enemy)
				g.currentSpawn++
			}
		}
		g.waveTimer++

		// 敵の移動処理
		for i := range g.enemies {
			e := &g.enemies[i]
			e.time += 0.05

			switch e.enemyType {
			case EnemyTypeStraight:
				e.y += e.speed
			case EnemyTypeSine:
				e.y += e.speed
				e.x += math.Sin(e.time) * 3
			case EnemyTypeSpecial:
				switch e.phase {
				case 0: // 上昇
					e.y += e.speed
					if e.y > screenHeight/2 {
						e.phase = 1
					}
				case 1: // 横移動
					e.x += e.speed * float64(e.turnDirection)
					if (e.turnDirection == 1 && e.x > screenWidth-40) || (e.turnDirection == -1 && e.x < 20) {
						e.phase = 2
					}
				case 2: // 下降
					e.y += e.speed
				}
			}

			// 弾発射
			if e.shootsBullet {
				e.bulletCooldown--
				if e.bulletCooldown <= 0 {
					switch e.bulletType {
					case 0: // 主人公狙い
						dx := g.playerX - e.x
						dy := g.playerY - e.y
						dist := math.Hypot(dx, dy)
						speed := 4.0
						vx := dx / dist * speed
						vy := dy / dist * speed
						g.enemyBullets = append(g.enemyBullets, EnemyBullet{x: e.x + 10, y: e.y + 20, vx: vx, vy: vy})
						g.particles = append(g.particles, Particle{x: e.x + 10, y: e.y + 20, vx: vx, vy: vy, size: 80, alpha: 1.0, lifetime: 5, ptype: 1})
					case 1: // 真下
						g.enemyBullets = append(g.enemyBullets, EnemyBullet{x: e.x + 10, y: e.y + 20, vx: 0, vy: 4.0})
						g.particles = append(g.particles, Particle{x: e.x + 10, y: e.y + 20, vx: 0, vy: 4.0, size: 80, alpha: 1.0, lifetime: 5, ptype: 1})
					case 2: // 斜め右下
						g.enemyBullets = append(g.enemyBullets, EnemyBullet{x: e.x + 10, y: e.y + 20, vx: 2.0, vy: 4.0})
						g.particles = append(g.particles, Particle{x: e.x + 10, y: e.y + 20, vx: 2.0, vy: 4.0, size: 80, alpha: 1.0, lifetime: 5, ptype: 1})
					case 3: // 斜め左下
						g.enemyBullets = append(g.enemyBullets, EnemyBullet{x: e.x + 10, y: e.y + 20, vx: -2.0, vy: 4.0})
						g.particles = append(g.particles, Particle{x: e.x + 10, y: e.y + 20, vx: -2.0, vy: 4.0, size: 80, alpha: 1.0, lifetime: 5, ptype: 1})
					}
					e.bulletCooldown = 60 + rand.Intn(60)
				}
			}
		}

		// 画面外に出た敵を削除
		newEnemies := g.enemies[:0]
		for _, e := range g.enemies {
			if e.y < screenHeight+20 {
				newEnemies = append(newEnemies, e)
			}
		}
		g.enemies = newEnemies

		// 全ての敵が出現し、かつ全滅したら次のステージへ
		if g.currentSpawn >= len(g.waves) && len(g.enemies) == 0 {
			g.gameState = GameStateStageClear
			g.stageClearTimer = 0
			g.stageClearKeyReleased = false
		}

		// 弾の発射（スペースキー）
		if ebiten.IsKeyPressed(ebiten.KeySpace) && g.shootCooldown == 0 {
			angles := []float64{-3, 0, 3}  // 度
			offsets := []float64{0, 8, 16} // 左・中央・右
			for i, deg := range angles {
				rad := (math.Pi / 180) * deg
				speed := 12.0
				bullet := Bullet{
					x:  g.playerX + offsets[i],
					y:  g.playerY,
					vx: math.Sin(rad) * speed,
					vy: -math.Cos(rad) * speed,
				}
				g.bullets = append(g.bullets, bullet)
			}
			g.shootCooldown = 5
		}
		if g.shootCooldown > 0 {
			g.shootCooldown--
		}

		// 弾の移動と当たり判定
		newBullets := g.bullets[:0]
		for _, b := range g.bullets {
			hit := false
			for i := range g.enemies {
				if b.x < g.enemies[i].x+20 && b.x+4 > g.enemies[i].x &&
					b.y < g.enemies[i].y+20 && b.y+8 > g.enemies[i].y {
					hit = true
					g.enemies[i].hp--
					if g.enemies[i].hp <= 0 {
						g.score += 100
						// 敵の種類に応じた色で爆発エフェクト
						var explosionColor color.RGBA
						switch g.enemies[i].enemyType {
						case EnemyTypeStraight:
							explosionColor = color.RGBA{255, 0, 0, 255}
						case EnemyTypeSine:
							explosionColor = color.RGBA{255, 165, 0, 255}
						case EnemyTypeSpecial:
							explosionColor = color.RGBA{255, 0, 255, 255}
						}
						g.createExplosion(g.enemies[i].x+10, g.enemies[i].y+10, explosionColor)
						g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
					}
					break
				}
			}
			if !hit {
				b.x += b.vx
				b.y += b.vy
				if b.y > -8 && b.x > -8 && b.x < screenWidth+8 {
					newBullets = append(newBullets, b)
				}
			}
		}
		g.bullets = newBullets

		// 敵弾の移動・当たり判定
		newEnemyBullets := g.enemyBullets[:0]
		for _, eb := range g.enemyBullets {
			eb.x += eb.vx
			eb.y += eb.vy
			// プレイヤーとの当たり判定
			if eb.x < g.playerX+20 && eb.x+4 > g.playerX && eb.y < g.playerY+24 && eb.y+8 > g.playerY {
				g.createExplosion(g.playerX+10, g.playerY+12, color.RGBA{0, 255, 0, 255})
				g.gameState = GameStatePlayerExplosion
				g.playerExplosionTimer = 0
				break
			}
			// 画面内に残す
			if eb.y < screenHeight+8 && eb.x > -8 && eb.x < screenWidth+8 {
				newEnemyBullets = append(newEnemyBullets, eb)
			}
		}
		g.enemyBullets = newEnemyBullets

		// プレイヤーと敵の当たり判定
		for _, e := range g.enemies {
			if g.playerX < e.x+20 && g.playerX+20 > e.x &&
				g.playerY < e.y+20 && g.playerY+24 > e.y {
				if g.score > g.highScore {
					g.highScore = g.score
				}
				// プレイヤーの爆発エフェクト
				g.createExplosion(g.playerX+10, g.playerY+12, color.RGBA{0, 255, 0, 255})
				g.gameState = GameStatePlayerExplosion
				g.playerExplosionTimer = 0
				break
			}
		}

	case GameStatePlayerExplosion:
		g.playerExplosionTimer++
		if g.playerExplosionTimer > 60 {
			g.gameState = GameStateGameOver
		}

	case GameStateStageClear:
		g.stageClearTimer++
		// 1秒経過後、スペースキーが一度離されてから押された場合のみ進行
		if g.stageClearTimer > 60 {
			if !ebiten.IsKeyPressed(ebiten.KeySpace) {
				g.stageClearKeyReleased = true
			}
			if g.stageClearKeyReleased && ebiten.IsKeyPressed(ebiten.KeySpace) {
				g.currentStage++
				if g.currentStage >= len(stages) {
					g.gameState = GameStateGameOver
					if g.score > g.highScore {
						g.highScore = g.score
					}
				} else {
					g.waves = stages[g.currentStage].Waves
					g.currentSpawn = 0
					g.waveTimer = 0
					g.enemies = []Enemy{}
					g.bullets = []Bullet{}
					g.enemyBullets = []EnemyBullet{}
					g.gameState = GameStatePlaying
				}
				return nil
			}
		}
		// 2秒経過で自動進行
		if g.stageClearTimer > 120 {
			g.currentStage++
			if g.currentStage >= len(stages) {
				g.gameState = GameStateGameOver
				if g.score > g.highScore {
					g.highScore = g.score
				}
			} else {
				g.waves = stages[g.currentStage].Waves
				g.currentSpawn = 0
				g.waveTimer = 0
				g.enemies = []Enemy{}
				g.bullets = []Bullet{}
				g.enemyBullets = []EnemyBullet{}
				g.gameState = GameStatePlaying
			}
		}

	case GameStateGameOver:
		// 敵の移動処理（ゲームオーバー時も継続）
		for i := range g.enemies {
			e := &g.enemies[i]
			e.time += 0.05

			switch e.enemyType {
			case EnemyTypeStraight:
				e.y += e.speed
			case EnemyTypeSine:
				e.y += e.speed
				e.x += math.Sin(e.time) * 3
			case EnemyTypeSpecial:
				switch e.phase {
				case 0: // 上昇
					e.y += e.speed
					if e.y > screenHeight/2 {
						e.phase = 1
					}
				case 1: // 横移動
					e.x += e.speed
					if e.x > screenWidth-40 {
						e.phase = 2
					}
				case 2: // 下降
					e.y += e.speed
				}
			}
		}

		// 画面外に出た敵を削除
		newEnemies := g.enemies[:0]
		for _, e := range g.enemies {
			if e.y < screenHeight+20 {
				newEnemies = append(newEnemies, e)
			}
		}
		g.enemies = newEnemies

		// Rキーでリスタート
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			*g = *NewGame()
			g.gameState = GameStatePlaying
		}
	}

	return nil
}

// Draw はゲームの描画を行います
func (g *Game) Draw(screen *ebiten.Image) {
	// 背景の星を描画（どの状態でも表示）
	for _, s := range g.stars {
		ebitenutil.DrawLine(screen, s.x, s.y, s.x, s.y+s.length, s.color)
	}

	switch g.gameState {
	case GameStateTitle:
		// タイトル画面
		titleText := "SPACE SHOOTER"
		startText := "Press SPACE to Start"
		highScoreText := fmt.Sprintf("High Score: %d", g.highScore)

		text.Draw(screen, titleText, gameFont, (screenWidth-len(titleText)*6)/2, screenHeight/3, color.White)
		text.Draw(screen, startText, gameFont, (screenWidth-len(startText)*6)/2, screenHeight/2, color.White)
		text.Draw(screen, highScoreText, gameFont, (screenWidth-len(highScoreText)*6)/2, screenHeight*2/3, color.White)

	case GameStatePlaying:
		// スコアとステージ表示
		scoreText := fmt.Sprintf("Score: %d", g.score)
		stageText := fmt.Sprintf("Stage: %s", stages[g.currentStage].Name)
		text.Draw(screen, scoreText, gameFont, 0, int(20*1.2), color.White)
		text.Draw(screen, stageText, gameFont, 0, int(20*2.0), color.White)

		// 敵を描画
		for _, e := range g.enemies {
			var enemyColor color.RGBA
			switch e.enemyType {
			case EnemyTypeStraight:
				enemyColor = color.RGBA{255, 0, 0, 255}
			case EnemyTypeSine:
				enemyColor = color.RGBA{255, 165, 0, 255}
			case EnemyTypeSpecial:
				enemyColor = color.RGBA{255, 0, 255, 255}
			}
			ebitenutil.DrawRect(screen, e.x, e.y, 20, 20, enemyColor)
			// HPバーを表示
			hpBarWidth := float64(e.hp) * 5
			ebitenutil.DrawRect(screen, e.x, e.y-5, hpBarWidth, 2, color.RGBA{0, 255, 0, 255})
		}

		// 自機を描画
		ebitenutil.DrawRect(screen, g.playerX, g.playerY, 4, 16, color.RGBA{0, 255, 0, 255})
		ebitenutil.DrawRect(screen, g.playerX+8, g.playerY-8, 4, 24, color.RGBA{0, 255, 0, 255})
		ebitenutil.DrawRect(screen, g.playerX+16, g.playerY, 4, 16, color.RGBA{0, 255, 0, 255})

		// 自機弾の描画
		for _, b := range g.bullets {
			ebitenutil.DrawRect(screen, b.x, b.y, 4, 8, color.RGBA{255, 255, 0, 255})
		}

		// 敵弾の描画（追加）
		for _, eb := range g.enemyBullets {
			ebitenutil.DrawRect(screen, eb.x, eb.y, 6, 12, color.RGBA{255, 0, 0, 255})
		}

		// パーティクルを描画
		for _, p := range g.particles {
			if p.ptype == 1 {
				norm := math.Hypot(p.vx, p.vy)
				if norm == 0 {
					norm = 1
				}
				length := 1000.0 // 画面端まで
				dx := p.vx / norm * length
				dy := p.vy / norm * length
				ebitenutil.DrawLine(screen, p.x, p.y, p.x+dx, p.y+dy, color.RGBA{255, 255, 0, uint8(p.alpha * 255)})
			} else {
				alpha := uint8(p.alpha * 255)
				ebitenutil.DrawRect(screen, p.x, p.y, p.size, p.size, color.RGBA{255, 255, 255, alpha})
			}
		}

	case GameStatePlayerExplosion:
		// 敵を描画
		for _, e := range g.enemies {
			var enemyColor color.RGBA
			switch e.enemyType {
			case EnemyTypeStraight:
				enemyColor = color.RGBA{255, 0, 0, 255}
			case EnemyTypeSine:
				enemyColor = color.RGBA{255, 165, 0, 255}
			case EnemyTypeSpecial:
				enemyColor = color.RGBA{255, 0, 255, 255}
			}
			ebitenutil.DrawRect(screen, e.x, e.y, 20, 20, enemyColor)
			// HPバーを表示
			hpBarWidth := float64(e.hp) * 5
			ebitenutil.DrawRect(screen, e.x, e.y-5, hpBarWidth, 2, color.RGBA{0, 255, 0, 255})
		}

		// 弾を描画
		for _, eb := range g.enemyBullets {
			ebitenutil.DrawRect(screen, eb.x, eb.y, 6, 12, color.RGBA{255, 128, 128, 255})
		}

		// パーティクルを描画
		for _, p := range g.particles {
			if p.ptype == 1 {
				norm := math.Hypot(p.vx, p.vy)
				if norm == 0 {
					norm = 1
				}
				length := 1000.0 // 画面端まで
				dx := p.vx / norm * length
				dy := p.vy / norm * length
				ebitenutil.DrawLine(screen, p.x, p.y, p.x+dx, p.y+dy, color.RGBA{255, 255, 0, uint8(p.alpha * 255)})
			} else {
				alpha := uint8(p.alpha * 255)
				ebitenutil.DrawRect(screen, p.x, p.y, p.size, p.size, color.RGBA{255, 255, 255, alpha})
			}
		}

	case GameStateStageClear:
		clearText := "STAGE CLEAR!"
		nextText := "Press SPACE or wait for next stage"
		text.Draw(screen, clearText, gameFont, (screenWidth-len(clearText)*6)/2, screenHeight/2-20, color.White)
		text.Draw(screen, nextText, gameFont, (screenWidth-len(nextText)*6)/2, screenHeight/2+20, color.White)

	case GameStateGameOver:
		// ゲームオーバー画面
		gameOverText := "GAME OVER"
		scoreText := fmt.Sprintf("Score: %d", g.score)
		highScoreText := fmt.Sprintf("High Score: %d", g.highScore)
		restartText := "Press R to Restart"

		text.Draw(screen, gameOverText, gameFont, (screenWidth-len(gameOverText)*6)/2, screenHeight/3, color.White)
		text.Draw(screen, scoreText, gameFont, 0, int(20*1.2), color.White)
		text.Draw(screen, highScoreText, gameFont, (screenWidth-len(highScoreText)*6)/2, screenHeight*2/3-20, color.White)
		text.Draw(screen, restartText, gameFont, (screenWidth-len(restartText)*6)/2, screenHeight*2/3+20, color.White)
	}
}

// Layout はゲームのレイアウトを設定します
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func loadFont() font.Face {
	fontBytes, err := os.ReadFile("assets/NotoSansJP-Regular.ttf")
	if err != nil {
		panic(err)
	}
	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	const fontSize = 20.0 // 1.5倍相当のサイズ
	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
	return face
}

func main() {
	gameFont = loadFont()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Simple Game")

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
