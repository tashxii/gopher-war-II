package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

// TargetInfo model
type TargetInfo struct {
	ID, NAME, CHARGE, CHARACTER string
	X, Y, LIFE, SIZE            int
	PrevX, PrevY                int
	HasMoved                    bool
	Paralyzed                   int // Frames remaining for paralysis
}

// BulletInfo model
type BulletInfo struct {
	ID                             string
	X, Y, LIFE, MAXLIFE, DIRECTION int
	DAMAGE, SPEED, FIRERANGE, SIZE int
	FIRE, SPECIAL                  bool
	ParalyzeAmount                 int // Amount of paralysis to apply when hitting
}

// Config model
type Config struct {
	MaxLife           int     `json:"maxLife"`
	TargetSize        int     `json:"maxSize"`
	MaxSpeed          int     `json:"maxSpeed"`
	Speed             float64 `json:"speed"`
	BombLife          int     `json:"bombLife"`
	BombSpeed         int     `json:"bombSpeed"`
	BombFire          int     `json:"bombFire"`
	BombSize          int     `json:"bombSize"`
	BombDmg           int     `json:"bombDmg"`
	MissileLife       int     `json:"missileLife"`
	MissileSpeed      int     `json:"missileSpeed"`
	MissileFire       int     `json:"missileFire"`
	MissileSize       int     `json:"missileSize"`
	MissileDmg        int     `json:"missileDmg"`
	DmgSize           int     `json:"dmgSize"`
	MissileCharging   int     `json:"missileCharging"`
	MissileCharged    int     `json:"missileCharged"`
	DmgMessage        string  `json:"dmgMessage"`
	MissileMessage    string  `json:"missileMessage"`
	ExplosionDuration int     `json:"explosionDuration"`
	ParalyzeDuration  int     `json:"paralyzeDuration"`
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	router := gin.Default()
	mrouter := melody.New()
	targets := make(map[*melody.Session]*TargetInfo)
	bombs := make(map[*melody.Session]*BulletInfo)
	missiles := make(map[*melody.Session]*BulletInfo)
	configs := make(map[*melody.Session]*Config)
	lock := new(sync.Mutex)
	counter := 0
	previousMillisecond := time.Now().UnixNano() / int64(time.Millisecond)
	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/index.html")
	})
	router.GET("/main", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/main.html")
	})
	router.Static("/static", "./static")

	router.GET("/ws", func(c *gin.Context) {
		mrouter.HandleRequest(c.Writer, c.Request)
	})

	mrouter.HandleConnect(func(s *melody.Session) {
		lock.Lock()
		for _, target := range targets {
			message := fmt.Sprintf("show %s %d %d %d %s %s %s %d", target.ID, target.X, target.Y, target.LIFE, target.NAME, target.CHARGE, target.CHARACTER, target.Paralyzed)
			s.Write([]byte(message))
		}
		//appear(id)
		// Generate random spawn position (avoid screen edges)
		spawnX := rand.Intn(600) + 100 // Random X between 100-700
		spawnY := rand.Intn(400) + 100 // Random Y between 100-500
		targets[s] = &TargetInfo{ID: strconv.Itoa(counter), NAME: "", CHARGE: "none", CHARACTER: "", X: spawnX, Y: spawnY, PrevX: spawnX, PrevY: spawnY, HasMoved: false, Paralyzed: 0}
		bombs[s] = &BulletInfo{ID: targets[s].ID, SPECIAL: false}
		missiles[s] = &BulletInfo{ID: targets[s].ID, SPECIAL: true}
		message := fmt.Sprintf("appear %s", targets[s].ID)
		s.Write([]byte(message))
		counter++
		lock.Unlock()
	})

	mrouter.HandleDisconnect(func(s *melody.Session) {
		lock.Lock()
		if _, exists := targets[s]; exists {
			mrouter.BroadcastOthers([]byte(fmt.Sprintf("dead %s", targets[s].ID)), s)
		}
		delete(targets, s)
		delete(bombs, s)
		delete(missiles, s)
		delete(configs, s)
		lock.Unlock()
	})

	mrouter.HandleMessage(func(s *melody.Session, msg []byte) {
		params := strings.Split(string(msg), " ")
		lock.Lock()
		if params[0] == "init" && len(params) >= 4 {
			// Parse player-specific config
			r := []rune(string(msg))
			r1 := []rune(string(params[1]))
			r2 := []rune(string(params[2]))
			playerConfig := &Config{}
			err := json.Unmarshal([]byte(string(r[len("init ")+len(r1)+1+len(r2)+1:])), playerConfig)
			if err != nil {
				message := fmt.Sprintf("Failed to configure by json [%s]", string(msg))
				fmt.Println(message)
				lock.Unlock()
				return
			}
			// Store player-specific config
			configs[s] = playerConfig

			target := targets[s]
			target.NAME = params[1]
			target.CHARACTER = params[2]
			target.LIFE = playerConfig.MaxLife
			target.SIZE = playerConfig.TargetSize
			bomb := bombs[s]
			bomb.MAXLIFE = playerConfig.BombLife
			bomb.LIFE = playerConfig.BombLife
			bomb.FIRERANGE = playerConfig.BombFire
			bomb.SPEED = playerConfig.BombSpeed
			bomb.SIZE = playerConfig.BombSize
			bomb.DAMAGE = playerConfig.BombDmg
			bomb.ParalyzeAmount = 0 // Bombs don't paralyze
			missile := missiles[s]
			missile.MAXLIFE = playerConfig.MissileLife
			missile.LIFE = playerConfig.MissileLife
			missile.FIRERANGE = playerConfig.MissileFire
			missile.SPEED = playerConfig.MissileSpeed
			missile.SIZE = playerConfig.MissileSize
			missile.DAMAGE = playerConfig.MissileDmg
			missile.ParalyzeAmount = playerConfig.ParalyzeDuration
		}
		//["show", e.pageX, e.pageY, charge]
		if params[0] == "show" && len(params) == 4 {
			// Check if target and config exist
			if targets[s] != nil && configs[s] != nil {
				moveTarget(targets[s], params, configs[s], mrouter)
			}
		}
		//["fire-xxx", e.pageX, e.pageY, direction]
		if params[0] == "fire-bomb" && len(params) == 4 {
			if bombs[s] != nil && targets[s] != nil {
				fireBullet(bombs[s], params, mrouter, s, targets)
			}
		}
		if params[0] == "fire-missile" && len(params) == 4 {
			if missiles[s] != nil && targets[s] != nil {
				fireBullet(missiles[s], params, mrouter, s, targets)
			}
		}
		if params[0] == "refresh" {
			currentMillisecond := time.Now().UnixNano() / int64(time.Millisecond)
			//Max FPS = 50
			if currentMillisecond-previousMillisecond >= 20 {
				previousMillisecond = currentMillisecond
				for _, missile := range missiles {
					moveBullet(missile, mrouter, targets)
				}
				for _, bomb := range bombs {
					moveBullet(bomb, mrouter, targets)
				}
				deadTargets := make([]*melody.Session, 0)
				for targetSession, target := range targets {
					targetConfig := configs[targetSession]
					targetDied := false
					for _, missile := range missiles {
						if judgeHitBullet(target, missile, targetConfig, mrouter) {
							targetDied = true
							break
						}
					}
					if !targetDied {
						for _, bomb := range bombs {
							if judgeHitBullet(target, bomb, targetConfig, mrouter) {
								targetDied = true
								break
							}
						}
					}
					if targetDied {
						deadTargets = append(deadTargets, targetSession)
					}
				}
				// Remove dead targets from maps
				for _, deadSession := range deadTargets {
					delete(targets, deadSession)
					delete(bombs, deadSession)
					delete(missiles, deadSession)
					delete(configs, deadSession)
				}
			}
		}
		lock.Unlock()
	})

	// Environment variables for server configuration
	ip := os.Getenv("IP")
	if ip == "" {
		fmt.Println("Environment variable [IP] is not set, using default")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
		fmt.Println("Environment variable [PORT] is not set, using default port 5000")
	}

	router.Run(ip + ":" + port)
}

func moveTarget(target *TargetInfo, params []string, config *Config, mrouter *melody.Melody) {
	// Decrease paralysis counter
	if target.Paralyzed > 0 {
		target.Paralyzed--
	}

	// If still paralyzed, don't allow movement but update charge
	if target.Paralyzed > 0 {
		target.CHARGE = params[3]
		message := fmt.Sprintf("show %s %d %d %d %s %s %s %d",
			target.ID,
			target.X,
			target.Y,
			target.LIFE,
			target.NAME,
			target.CHARGE,
			target.CHARACTER,
			target.Paralyzed)
		mrouter.Broadcast([]byte(message))
		return
	}

	newX, _ := strconv.Atoi(params[1])
	newY, _ := strconv.Atoi(params[2])

	// Calculate movement distance
	dx := float64(newX - target.PrevX)
	dy := float64(newY - target.PrevY)
	distance := math.Sqrt(dx*dx + dy*dy)

	// Apply speed limit
	maxMoveDistance := float64(config.MaxSpeed) * config.Speed
	if distance > maxMoveDistance && maxMoveDistance > 0 {
		// Normalize movement vector and apply speed limit
		ratio := maxMoveDistance / distance
		dx *= ratio
		dy *= ratio
		newX = target.PrevX + int(dx)
		newY = target.PrevY + int(dy)
	}

	// Update position
	target.PrevX = target.X
	target.PrevY = target.Y
	target.X = newX
	target.Y = newY
	target.CHARGE = params[3]

	// Mark that player has moved from initial position
	if !target.HasMoved && (target.X != 0 || target.Y != 0) {
		target.HasMoved = true
	}

	message := fmt.Sprintf("show %s %d %d %d %s %s %s %d",
		target.ID,
		target.X,
		target.Y,
		target.LIFE,
		target.NAME,
		target.CHARGE,
		target.CHARACTER,
		target.Paralyzed)
	mrouter.Broadcast([]byte(message))
}

func fireBullet(bullet *BulletInfo, params []string, mrouter *melody.Melody, s *melody.Session, targets map[*melody.Session]*TargetInfo) {
	bullet.FIRE = true
	bullet.LIFE = bullet.MAXLIFE
	// Use player's actual position instead of mouse position
	target := targets[s]
	bullet.X = target.X
	bullet.Y = target.Y
	bullet.DIRECTION, _ = strconv.Atoi(params[3])
	message := fmt.Sprintf("bullet %s %d %d %d %t %s", bullet.ID, bullet.X, bullet.Y, bullet.DIRECTION, bullet.SPECIAL, target.CHARACTER)
	mrouter.Broadcast([]byte(message))
}

func moveBullet(bullet *BulletInfo, mrouter *melody.Melody, targets map[*melody.Session]*TargetInfo) {
	if bullet.FIRE == false {
		return
	}
	bullet.LIFE = bullet.LIFE - 1
	if bullet.LIFE <= 0 {
		bullet.FIRE = false
		message := fmt.Sprintf("miss %s %t", bullet.ID, bullet.SPECIAL)
		mrouter.Broadcast([]byte(message))
		return
	}
	var dx, dy int
	switch bullet.DIRECTION {
	case 0:
		dy = bullet.SPEED
	case 1:
		dx = bullet.SPEED
	case 2:
		dy = -bullet.SPEED
	case 3:
		dx = -bullet.SPEED
	}
	bullet.X += dx
	bullet.Y += dy
	// Find the character for this bullet by matching bullet ID with target ID
	var character string = "gopher" // default
	for _, target := range targets {
		if target.ID == bullet.ID {
			character = target.CHARACTER
			break
		}
	}
	message := fmt.Sprintf("bullet %s %d %d %d %t %s", bullet.ID, bullet.X, bullet.Y, bullet.DIRECTION, bullet.SPECIAL, character)
	mrouter.Broadcast([]byte(message))
}

func judgeHitBullet(target *TargetInfo, bullet *BulletInfo, targetConfig *Config, mrouter *melody.Melody) bool {
	if bullet.FIRE == false || target.LIFE <= 0 || bullet.LIFE >= bullet.FIRERANGE {
		return false
	}

	// Skip collision detection if target hasn't moved from initial position
	if !target.HasMoved {
		return false
	}

	// Calculate current target size based on character config and current life
	currentTargetSize := targetConfig.TargetSize - targetConfig.DmgSize*(targetConfig.MaxLife-target.LIFE)

	// Circular collision detection with margin
	dx := float64(target.X - bullet.X)
	dy := float64(target.Y - bullet.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	// Calculate collision radius with 5px margin
	targetRadius := float64(currentTargetSize) / 2
	bulletRadius := float64(bullet.SIZE) / 2
	hitRadius := targetRadius + bulletRadius + 5 // 5px margin

	if distance <= hitRadius {
		target.LIFE = target.LIFE - bullet.DAMAGE
		bullet.LIFE = 0
		bullet.FIRE = false
		if target.LIFE > targetConfig.MaxLife {
			target.LIFE = targetConfig.MaxLife // Cap the life to the max life
		}

		// Update target size after damage
		target.SIZE = targetConfig.TargetSize - targetConfig.DmgSize*(targetConfig.MaxLife-target.LIFE)

		// Apply paralysis based on bullet's paralyze amount
		if bullet.ParalyzeAmount > 0 {
			target.Paralyzed = bullet.ParalyzeAmount
		}

		if target.LIFE <= 0 {
			message := fmt.Sprintf("dead %s %s %t", target.ID, bullet.ID, bullet.SPECIAL)
			mrouter.Broadcast([]byte(message))
			return true // Target died, should be removed
		} else {
			message := fmt.Sprintf("hit %s %s %d %t %s %d", target.ID, bullet.ID, target.LIFE, bullet.SPECIAL, target.CHARACTER, target.Paralyzed)
			mrouter.Broadcast([]byte(message))
		}
	}
	return false
}
