package filepatcher

import (
	"fmt"
	"strings"
)

func (bop *blockOp) patch(lines []string, remove bool) ([]string, error) {
	if remove {
		return bop.unPatch(lines)
	}
	// A first pass. Just to detect if block is present
	for idx, line := range lines {
		switch bop.state {
		case blockInit:
			if strings.Contains(line, bop.beginMarker) {
				bop.state = blockSkip
			}
		case blockSkip:
			if strings.Contains(line, bop.endMarker) {
				bop.state = blockFound
			} // else continue
		case blockFound:
			// Sanity check
			if strings.Contains(line, bop.beginMarker) || strings.Contains(line, bop.endMarker) {
				return nil, fmt.Errorf("duplicated block marker")
			}
		default:
			return nil, fmt.Errorf("unhandlded state '%d' at line %d on first pass", bop.state, idx)
		}
	}
	if bop.state != blockInit && bop.state != blockFound {
		return nil, fmt.Errorf("invalid state '%d' at EOF on first pass", bop.state)
	}
	// The second pass will do the real job.
	newLines := make([]string, 0, len(lines)+20)
	for idx, line := range lines {
		switch bop.state {
		case blockInit:
			if bop.insertAfter.Match([]byte(line)) {
				// Must insert the block right after
				newLines = append(newLines, line)
				newLines = bop.pushBlock(newLines)
				bop.state = blockDone
			} else {
				// Nothing to do, just copy
				newLines = append(newLines, line)
			}
		case blockFound:
			if strings.Contains(line, bop.beginMarker) {
				// We must re-insert the block (Maybe the content has been changed)
				newLines = bop.pushBlock(newLines)
				bop.state = blockSkip
			} else {
				// Nothing to do, just copy
				newLines = append(newLines, line)
			}
		case blockSkip:
			if strings.Contains(line, bop.endMarker) {
				bop.state = blockDone
			} // else continue
		case blockDone:
			// Nothing to do, just copy
			newLines = append(newLines, line)
		default:
			return nil, fmt.Errorf("unhandlded state '%d' at line %d on second pass", bop.state, idx)
		}
	}
	if bop.state == blockInit {
		// We are at the end of the file. Must add the block if not done
		newLines = bop.pushBlock(newLines)
		bop.state = blockDone
	}
	if bop.state != blockDone {
		return nil, fmt.Errorf("invalid state '%d' at EOF on second pass", bop.state)
	}
	return newLines, nil
}

func (bop *blockOp) pushBlock(lines []string) []string {
	lines = append(lines, bop.padding+bop.beginMarker)
	blockLines := strings.Split(bop.op.Block, "\n")
	for _, l := range blockLines {
		if len(l) != 0 { // Skip empty line (may occurs
			lines = append(lines, bop.padding+l)
		}
	}
	lines = append(lines, bop.padding+bop.endMarker)
	return lines
}

func (bop *blockOp) unPatch(lines []string) ([]string, error) {

	newLines := make([]string, 0, len(lines)+20)
	for idx, line := range lines {
		switch bop.state {
		case blockInit:
			if strings.Contains(line, bop.beginMarker) {
				bop.state = blockSkip
			} else {
				// Nothing to do, just copy
				newLines = append(newLines, line)
			}
		case blockSkip:
			if strings.Contains(line, bop.endMarker) {
				bop.state = blockInit
			} // else just forget
		default:
			return nil, fmt.Errorf("unhandlded state '%d' at line %d on unPatch", bop.state, idx)
		}
	}
	if bop.state != blockInit {
		return nil, fmt.Errorf("invalid state '%d' at EOF on second pass", bop.state)
	}
	return newLines, nil
}
