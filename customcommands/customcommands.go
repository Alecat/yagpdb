package customcommands

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"github.com/jonas747/yagpdb/bot"
	"github.com/jonas747/yagpdb/web"
	"log"
	"sort"
)

type Plugin struct{}

func RegisterPlugin() {
	plugin := &Plugin{}
	web.RegisterPlugin(plugin)
	bot.RegisterPlugin(plugin)
}

func (p *Plugin) InitBot() {
	bot.Session.AddHandler(bot.CustomMessageCreate(HandleMessageCreate))
}

func (p *Plugin) Name() string {
	return "Custom commands"
}

type CommandTriggerType int

const (
	CommandTriggerCommand CommandTriggerType = iota
	CommandTriggerStartsWith
	CommandTriggerContains
	CommandTriggerRegex
	CommandTriggerExact
)

type CustomCommand struct {
	TriggerType   CommandTriggerType `json:"trigger_type"`
	Trigger       string             `json:"trigger"`
	Response      string             `json:"response"`
	CaseSensitive bool               `json:"case_sensitive"`
	ID            int                `json:"id"`
}

func GetCommands(client *redis.Client, guild string) ([]*CustomCommand, int, error) {
	hash, err := client.Cmd("HGETALL", "custom_commands:"+guild).Hash()
	if err != nil {
		// Check if the error was that it didnt exist, if so return an empty slice
		// If not, there was an actual error
		if _, ok := err.(*redis.CmdError); ok {
			return []*CustomCommand{}, 0, nil
		} else {
			return nil, 0, err
		}
	}

	highest := 0
	result := make([]*CustomCommand, len(hash))

	// Decode the commands, and also calculate the highest id
	i := 0
	for k, raw := range hash {
		var decoded *CustomCommand
		err = json.Unmarshal([]byte(raw), &decoded)
		if err != nil {
			log.Println("Failed decoding custom command", k, guild, err)
			result[i] = &CustomCommand{}
		} else {
			result[i] = decoded
			if decoded.ID > highest {
				highest = decoded.ID
			}
		}
		i++
	}

	// Sort by id
	sort.Sort(CustomCommandSlice(result))

	return result, highest, nil
}

type CustomCommandSlice []*CustomCommand

// Len is the number of elements in the collection.
func (c CustomCommandSlice) Len() int {
	return len(c)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (c CustomCommandSlice) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}

// Swap swaps the elements with indexes i and j.
func (c CustomCommandSlice) Swap(i, j int) {
	temp := c[i]
	c[i] = c[j]
	c[j] = temp
}