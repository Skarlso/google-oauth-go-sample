package structs

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Enemy represents an enemy combatant.
type Enemy struct {
	Name string
	ID   string
	Race int
	Cast int
	// Calculated based on player's HP
	Hp int
	// Items which the player can loot. Will be crossreferenced with Items, from items.json
	Items []Item
	// Gold which the player can loot
	Gold int
	// Xp is calculated based on level and rareness
	Xp int
	// Level is calculated based on the Players level. +-5%
	Level int
	// RarenessLevel is 1-10 where 10 is highly rare
	RarenessLevel int
	// Armor is given for now. Should be increased with level
	Armor  int
	Damage int
}

// SpawnEnemy spawns an enemy combatand who's stats are based on the player's character.
// Also, based on RarenessLevel.
func SpawnEnemy(c Character) Enemy {
	// Monster Level will be +- 20% of Character Level
	m := Enemy{}
	// TODO: Generated Monster names and stats. Instead of saved ones.
	m.initializeStatsFromJSON()
	m.Hp = calculateEnemyHp(c.MaxHp)
	m.Level = calculateEnemyLevel(c.Level)
	m.Xp = calculateEnemyXpaward(c.Level, m.Xp)
	return m
}

// MonsterItem is an item in the monsters.json file.
type MonsterItem struct {
	ID     int `json:"id"`
	Chance int `json:"chance"`
}

// Monster is a monster from the monsters.json file.
type Monster struct {
	Name   string        `json:"name"`
	ID     int           `json:"id"`
	Race   int           `json:"race"`
	Cast   int           `json:"cast"`
	Items  []MonsterItem `json:"items"`
	Gold   int           `json:"gold"`
	Rare   int           `json:"rare"`
	Xp     int           `json:"xp"`
	Armor  int           `json:"armor"`
	Damage int           `json:"damage"`
}

// Monsters is a collection of monsters.
type Monsters struct {
	Monster []Monster `json:"monsters"`
}

// initializeStatsFromJSON initialize the stats of an enemy with stats from the monsters JSON file.
func (e *Enemy) initializeStatsFromJSON() {

	m := Monsters{}

	file, err := os.Open("monsters.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)

	err = json.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(m.Monster))
	e.Cast = m.Monster[index].Cast
	e.Race = m.Monster[index].Race
	e.RarenessLevel = m.Monster[index].Rare
	e.Gold = m.Monster[index].Gold
	e.ID = strconv.Itoa(m.Monster[index].ID)
	e.Name = m.Monster[index].Name
	e.Xp = m.Monster[index].Xp
	e.Armor = m.Monster[index].Armor
	e.Damage = m.Monster[index].Damage

	var monsterItems []Item
	// for _, i := range m.Monster[index].Items {
	// 	monsterItems = append(monsterItems, ItemsMap[i.ID])
	// }
	e.Items = monsterItems
}

// Items is a collection of all the items from the items file.
type Items struct {
	Items []Item `json:"items"`
}

// calculateEnemyHp calculates eveny hp.
// TODO: For now this is the same as level. Come up with something better!
func calculateEnemyHp(playerHp int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	limiter := int(float64(playerHp) * 0.2)
	if limiter <= 0 {
		limiter = 1
	}

	hp := (playerHp - limiter) + r.Intn(limiter*2)

	if hp < 100 {
		hp = 100
	}

	return hp
}

// calculateEnemyLevel calculates enemy's level based on the player.
func calculateEnemyLevel(playerLevel int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	limiter := int(float64(playerLevel) * 0.2)
	if limiter <= 0 {
		limiter = 1
	}

	level := (playerLevel - limiter) + r.Intn(limiter*2)

	if level < 0 {
		level = 0
	}

	return level
}

// calculateEnemyXpaward currently it just multiplies the base xp of a monsters
// by 30% of players level.
// Later this will be a graded system with Challenge Rating for monsters based on
// RarenessLevel, Gold, Xp, HP.
func calculateEnemyXpaward(playerLevel, initialMonsterXp int) int {
	multiplier := float64(playerLevel) * 0.3
	newXp := int(float64(initialMonsterXp)*multiplier) + initialMonsterXp

	return newXp
}
