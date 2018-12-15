package flutter

import (
	"sort"
	"unicode"
)

type argsEditingState struct {
	Text                   string `json:"text"`
	SelectionBase          int    `json:"selectionBase"`
	SelectionExtent        int    `json:"selectionExtent"`
	SelectionAffinity      string `json:"selectionAffinity"`
	SelectionIsDirectional bool   `json:"selectionIsDirectional"`
	ComposingBase          int    `json:"composingBase"`
	ComposingExtent        int    `json:"composingExtent"`
}

func (p *textinputPlugin) isSelected() bool {
	return p.selectionBase != p.selectionExtent
}

func (p *textinputPlugin) addChar(char []rune) {
	p.RemoveSelectedText()
	newWord := make([]rune, 0, len(char)+len(p.word))
	newWord = append(newWord, p.word[:p.selectionBase]...)
	newWord = append(newWord, char...)
	newWord = append(newWord, p.word[p.selectionBase:]...)

	p.word = newWord

	p.selectionBase += len(char)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorHomeSimple() {
	p.selectionBase = 0
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorHomeSelect() {
	p.selectionBase = 0
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorEndSimple() {
	p.selectionBase = len(p.word)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorEndSelect() {
	p.selectionBase = len(p.word)
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorLeftSimple() {
	p.selectionExtent--
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorLeftWord() {
	p.selectionBase = indexStartLeadingWord(p.word, p.selectionBase)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()

}

func (p *textinputPlugin) MoveCursorLeftLine() {
	if p.isSelected() {
		p.selectionExtent = indexStartLeadingWord(p.word, p.selectionExtent)
	} else {
		p.selectionExtent = indexStartLeadingWord(p.word, p.selectionBase)
	}
	p.updateEditingState()

}

func (p *textinputPlugin) MoveCursorLeftReset() {
	if !p.isSelected() {
		if p.selectionBase > 0 {
			p.selectionBase--
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorRightSimple() {
	p.selectionExtent++
	p.updateEditingState()
}

func (p *textinputPlugin) MoveCursorRightWord() {
	p.selectionBase = indexEndForwardWord(p.word, p.selectionBase)
	p.selectionExtent = p.selectionBase
	p.updateEditingState()

}

func (p *textinputPlugin) MoveCursorRightLine() {
	if p.isSelected() {
		p.selectionExtent = indexEndForwardWord(p.word, p.selectionExtent)
	} else {
		p.selectionExtent = indexEndForwardWord(p.word, p.selectionBase)
	}
	p.updateEditingState()

}

func (p *textinputPlugin) MoveCursorRightReset() {
	if !p.isSelected() {
		if p.selectionBase < len(p.word) {
			p.selectionBase++
			p.selectionExtent = p.selectionBase
		}
	} else {
		p.selectionBase = p.selectionExtent
	}
	p.updateEditingState()
}

func (p *textinputPlugin) SelectAll() {
	p.selectionBase = 0
	p.selectionExtent = len(p.word)
	p.updateEditingState()
}

func (p *textinputPlugin) DeleteSimple() {
	if p.selectionBase < len(p.word) {
		p.word = append(p.word[:p.selectionBase], p.word[p.selectionBase+1:]...)
		p.updateEditingState()
	}
}

func (p *textinputPlugin) DeleteWord() {
	UpTo := indexEndForwardWord(p.word, p.selectionBase)
	p.word = append(p.word[:p.selectionBase], p.word[UpTo:]...)
	p.updateEditingState()
}

func (p *textinputPlugin) DeleteLine() {
	p.word = p.word[:p.selectionBase]
	p.updateEditingState()
}


func (p *textinputPlugin) BackspaceSimple(){
	if len(p.word) > 0 && p.selectionBase > 0 {
		p.word = append(p.word[:p.selectionBase-1], p.word[p.selectionBase:]...)
		p.selectionBase--
		p.selectionExtent = p.selectionBase
		p.updateEditingState()
	}
}

func (p *textinputPlugin) BackspaceWord(){
	if len(p.word) > 0 && p.selectionBase > 0 {
		deleteUpTo := indexStartLeadingWord(p.word, p.selectionBase)
		p.word = append(p.word[:deleteUpTo], p.word[p.selectionBase:]...)
		p.selectionBase = deleteUpTo
		p.selectionExtent = deleteUpTo
		p.updateEditingState()
	}
}

func (p *textinputPlugin) BackspaceLine(){
	p.word = p.word[:0]
	p.selectionBase = 0
	p.selectionExtent = 0
	p.updateEditingState()
}

// RemoveSelectedText do nothing if no text is selected
// return true if the state has been updated
func (p *textinputPlugin) RemoveSelectedText() bool {
	if p.isSelected() {
		selectionIndexStart, selectionIndexEnd, _ := p.GetSelectedText()
		p.word = append(p.word[:selectionIndexStart], p.word[selectionIndexEnd:]...)
		p.selectionBase = selectionIndexStart
		p.selectionExtent = selectionIndexStart
		p.selectionExtent = p.selectionBase
		p.updateEditingState()
		return true
	}
	return false

}

// GetSelectedText return
// (left index of the selection, right index of the selection,
// the content of the selection)
func (p *textinputPlugin) GetSelectedText() (int, int, string) {
	selectionIndex := []int{p.selectionBase, p.selectionExtent}
	sort.Ints(selectionIndex)
	return selectionIndex[0],
		selectionIndex[1],
		string(p.word[selectionIndex[0]:selectionIndex[1]])
}

// Helpers
func indexStartLeadingWord(line []rune, start int) int {
	pos := start
	// Remove whitespace to the left
	for {
		if pos == 0 || !unicode.IsSpace(line[pos-1]) {
			break
		}
		pos--
	}
	// Remove non-whitespace to the left
	for {
		if pos == 0 || unicode.IsSpace(line[pos-1]) {
			break
		}
		pos--
	}
	return pos
}

func indexEndForwardWord(line []rune, start int) int {
	pos := start
	lineSize := len(line)
	// Remove whitespace to the right
	for {
		if pos == lineSize || !unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	// Remove non-whitespace to the right
	for {
		if pos == lineSize || unicode.IsSpace(line[pos]) {
			break
		}
		pos++
	}
	return pos
}

// UpupdateEditingState updates the TextInput with the current state by invoking
// TextInputClient.updateEditingState in the Flutter Framework.
func (p *textinputPlugin) updateEditingState() {
	editingState := argsEditingState{
		Text:                   string(p.word),
		SelectionAffinity:      "TextAffinity.downstream",
		SelectionBase:          p.selectionBase,
		SelectionExtent:        p.selectionExtent,
		SelectionIsDirectional: false,
	}
	arguments := []interface{}{
		p.clientID,
		editingState,
	}
	p.channel.InvokeMethod("TextInputClient.updateEditingState", arguments)
}

// performAction invokes the TextInputClient performAction method in the Flutter
// Framework.
func (p *textinputPlugin) performAction(action string) {
	p.channel.InvokeMethod("TextInputClient.performAction", []interface{}{
		p.clientID,
		"TextInputAction." + action,
	})
}
