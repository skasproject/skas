package filepatcher

import (
	"fmt"
)

func (lop *lineOp) patch(lines []string, remove bool) ([]string, error) {
	if remove {
		return lop.unPatch(lines)
	}
	// If the line is present, just update it, and job is done, so exit
	for idx, line := range lines {
		if lop.regex.Match([]byte(line)) {
			lines[idx] = lop.padding + lop.op.Line
			return lines, nil
		}
	}
	// Line was not found. Will insert it
	newLines := make([]string, 0, len(lines)+20)
	for idx, line := range lines {
		switch lop.state {
		case lineInit:
			//fmt.Printf("will try '%s'\n", line)
			if lop.insertAfter.Match([]byte(line)) {
				newLines = append(newLines, line)
				newLines = append(newLines, lop.padding+lop.op.Line)
				lop.state = lineFound
			} else {
				newLines = append(newLines, line)
			}
		case lineFound:
			newLines = append(newLines, line)
		default:
			return nil, fmt.Errorf("unhandlded state '%d' at line %d on first pass", lop.state, idx)
		}
	}
	if lop.state == lineInit {
		// Was not added. Add at the end
		newLines = append(newLines, lop.padding+lop.op.Line)
		lop.state = lineFound
	}
	return newLines, nil
}

func (lop *lineOp) unPatch(lines []string) ([]string, error) {
	newLines := make([]string, 0, len(lines)+20)
	for _, line := range lines {
		if !lop.regex.Match([]byte(line)) {
			newLines = append(newLines, line)
		}
	}
	return newLines, nil
}
