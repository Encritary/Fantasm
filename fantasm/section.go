package fantasm

import (
	vkapi "github.com/himidori/golang-vk-api"
	"strconv"
)

type Section struct {
	ID      string
	Text    string
	Images  []string
	Actions []*Action
}

type Action struct {
	Template string
	Label    string
	Color    string
	Section  string
}

func (sec *Section) BuildKeyboard(inline bool) vkapi.Keyboard {
	mergeRows := 0
	mergeLen := 0

	rows := len(sec.Actions)
	if len(sec.Actions) > 10 {
		mergeRows = len(sec.Actions) - 10
		mergeLen = 2

		rows = 10
	}
	buttons := make([][]vkapi.Button, rows)

	y := 0
	x := 0
	for i, action := range sec.Actions {
		buttons[y] = append(buttons[y], vkapi.Button{
			Action: map[string]string{
				"type":    "text",
				"label":   action.Label,
				"payload": "{\"action\":" + strconv.Itoa(i) + "}", // TODO
			},
			Color: action.Color,
		})

		if mergeRows > 0 {
			x++

			if x >= mergeLen {
				mergeRows--
				x = 0
				y++
			}
		} else {
			y++
		}
	}
	return vkapi.Keyboard{
		Buttons: buttons,
		Inline:  inline,
		OneTime: false,
	}
}
