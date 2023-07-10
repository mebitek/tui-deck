package deck_markdown

import (
	"fmt"
	"github.com/rivo/tview"
	"regexp"
	"strings"
	"tui-deck/utils"
)

func GetMarkDownDescription(text string, configuration utils.Configuration) string {
	var lineContainer = ""
	var isCodeBlock = false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "##") || strings.HasPrefix(line, "###") {
			// # heading
			line = fmt.Sprintf("[:%s:b]%s[-:-:-]\n", configuration.Color, strings.ReplaceAll(line, "#", "")[1:])
		} else if strings.Contains(line, "- [ ]") {
			// task list unselected
			line = fmt.Sprintf("%s\n", strings.ReplaceAll(line, "- [ ]", "[ ]"))
		} else if strings.Contains(line, "- [x]") {
			// task list selected
			line = fmt.Sprintf("%s\n", strings.ReplaceAll(line, "- [x]", "[✓]"))
		} else if strings.HasPrefix(line, "- ") {
			// unordered list
			line = fmt.Sprintf("%s\n", strings.ReplaceAll(line, "- ", "• "))
		} else if strings.HasPrefix(line, "> ") {
			// Blockquotes
			line = fmt.Sprintf("%s\n", strings.Replace(line, "> ", "│ ", 1))
		} else if strings.HasPrefix(line, "```") {
			//codeBlock
			if !isCodeBlock {
				line = fmt.Sprintf("[#af0000:#4e4e4e:-]%s\n", strings.ReplaceAll(line, "```", ""))
				isCodeBlock = true
			} else {
				line = fmt.Sprintf("[-:-:-]%s\n", strings.ReplaceAll(line, "```", ""))
			}
		} else if len(checkLink(line)) > 0 {
			result := checkLink(line)
			for _, r := range result {
				rTmp := strings.ReplaceAll(r[1], "[", "")
				r[1] = strings.ReplaceAll(rTmp, "]", "")
				line = fmt.Sprintf("%s", strings.ReplaceAll(line, r[0], fmt.Sprintf("[#00ffaf]%s [#5f5fff:-:u]%s[-:-:-]", strings.ReplaceAll(r[1], "[", ""), r[2])))
			}
		} else {
			line = fmt.Sprintf("%s\n", tview.Escape(line))
		}

		// bold + italic
		result := getStringInBetween(line, "\\*\\*\\*", "\\*\\*\\*")
		if len(result) > 0 {
			for _, r := range result {
				line = fmt.Sprintf("%s", strings.ReplaceAll(line, r[1], fmt.Sprintf("[::bi]%s[::-]", r[2])))
			}
			line = fmt.Sprintf("%s", strings.ReplaceAll(line, "***", ""))
		}
		// bold
		result = getStringInBetween(line, "\\*\\*", "\\*\\*")
		if len(result) > 0 {
			for _, r := range result {
				line = fmt.Sprintf("%s", strings.ReplaceAll(line, r[1], fmt.Sprintf("[::b]%s[::-]", r[2])))
			}
			line = fmt.Sprintf("%s", strings.ReplaceAll(line, "**", ""))
		}
		// italic
		result = getStringInBetween(line, "\\*", "\\*")
		if len(result) > 0 {

			for _, r := range result {
				line = fmt.Sprintf("%s", strings.ReplaceAll(line, r[1], fmt.Sprintf("[::i]%s[::-]", r[2])))
			}
			line = fmt.Sprintf("%s", strings.ReplaceAll(line, "*", ""))
		}
		// code inline
		result = getStringInBetween(line, "\\`", "\\`")
		if len(result) > 0 {
			for _, r := range result {
				line = fmt.Sprintf("%s", strings.ReplaceAll(line, r[1], fmt.Sprintf("[#af0000:#4e4e4e:-] %s [::-]", r[2])))
			}
			line = fmt.Sprintf("%s", strings.ReplaceAll(line, "`", ""))
		}

		lineContainer = lineContainer + fmt.Sprintf("%s", line)
	}

	return lineContainer
}

func getStringInBetween(str string, start string, end string) (result [][]string) {
	re := regexp.MustCompile("(" + start + `(.*?)` + end + ")")
	match := re.FindAllStringSubmatch(str, -1)
	if len(match) > 0 {
		return match
	} else {
		return
	}
}

func checkLink(str string) (result [][]string) {
	re := regexp.MustCompile(`(\[[^][]+])\((https?://[^()]+)\)`)
	match := re.FindAllStringSubmatch(str, -1)
	if len(match) > 0 {
		return match
	} else {
		return
	}
}

func CountCheckList(description string) (int, int, error) {
	re := regexp.MustCompile("- \\[ \\]")
	matchUncheckd := re.FindAllStringSubmatch(description, -1)

	re = regexp.MustCompile("- \\[x\\]")
	matchChecked := re.FindAllStringSubmatch(description, -1)

	if len(matchChecked) > 0 || len(matchUncheckd) > 0 {
		return len(matchChecked), len(matchUncheckd) + len(matchChecked), nil
	}
	return 0, 0, fmt.Errorf("no check list found")

}
