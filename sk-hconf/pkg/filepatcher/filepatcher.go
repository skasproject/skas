package filepatcher

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

// Public structures

type BlockOperation struct {
	Block       string `yaml:"block"`
	Marker      string `yaml:"marker"`      // Text surrounding the block. Must contain '{mark}', which will be substituted with 'BEGIN' and END
	InsertAfter string `yaml:"insertAfter"` // A regex. Block will be inserted after if not present. If not found will be inserted at the end of file
	Indent      int    `yaml:"indent"`
}

type LineOperation struct {
	Line        string `yaml:"line"`        // The line to insert
	Regex       string `yaml:"regex"`       // If matched, The 'Line' will be inserted at this place. If no match, the line will be inserted under 'InsertAfter'
	InsertAfter string `yaml:"insertAfter"` // Regex. The Line will be inserted after this if regex had no matches. Or at the end if InsertAfter has no match (or is "")
	Indent      int    `yaml:"indent"`
}

type PatchOperation struct {
	File            string           `yaml:"file"`
	Backup          bool             `yaml:"backup"`
	BackupFolder    string           `yaml:"backupFolder"`
	TmpFolder       string           `yaml:"tmpFolder"`
	Remove          bool             `yaml:"remove"`
	BlockOperations []BlockOperation `yaml:"blockOperations"`
	LineOperations  []LineOperation  `yaml:"lineOperations"`
}

// Internal structures

type blockState int

const (
	blockInit blockState = iota
	blockSkip
	blockFound
	blockDone
)

type lineState int

const (
	lineInit lineState = iota
	lineFound
)

type blockOp struct {
	op          *BlockOperation
	beginMarker string
	endMarker   string
	insertAfter *regexp.Regexp
	state       blockState
	padding     string
}

type lineOp struct {
	op          *LineOperation
	regex       *regexp.Regexp
	insertAfter *regexp.Regexp
	state       lineState
	padding     string
}

type patchOp struct {
	patch    *PatchOperation
	blockOps []*blockOp
	lineOps  []*lineOp
}

func (patch *PatchOperation) Run() error {
	if patch.Backup {
		err := backup(patch.File, patch.BackupFolder)
		if err != nil {
			return err
		}
	}
	patchOp, err := buildPatch(patch)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(patch.File)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	lines, err = patchOp.doPatch(lines)
	if err != nil {
		return err
	}
	// And write the file
	tmpFile, err := os.CreateTemp(patch.TmpFolder, "filepatcher")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }() // clean up
	if _, err := tmpFile.WriteString(strings.Join(lines, "\n")); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpFile.Name(), patch.File); err != nil {
		return err
	}

	return nil
}

func backup(srcFile string, backupFileFolder string) error {
	var backupRoot string
	if backupFileFolder != "" {
		backupRoot = path.Join(backupFileFolder, path.Base(srcFile))
	} else {
		backupRoot = srcFile
	}
	dstFile := fmt.Sprintf("%s.%d.%s.bck", backupRoot, os.Getpid(), time.Now().Format(time.RFC3339))

	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	// preserve permissions from srcFile to dstFile
	srcStat, _ := src.Stat()
	err = os.Chmod(dstFile, srcStat.Mode())
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

const spaces = "                                                                                                       "

func buildPatch(patchOperation *PatchOperation) (*patchOp, error) {
	patchOp := &patchOp{
		patch:    patchOperation,
		blockOps: make([]*blockOp, 0, len(patchOperation.BlockOperations)),
		lineOps:  make([]*lineOp, 0, len(patchOperation.LineOperations)),
	}
	var err error
	for idx, blockOperation := range patchOperation.BlockOperations {
		bo := &blockOp{
			op:          &patchOperation.BlockOperations[idx],
			state:       blockInit,
			beginMarker: strings.Replace(blockOperation.Marker, "{mark}", "BEGIN", 1),
			endMarker:   strings.Replace(blockOperation.Marker, "{mark}", "END", 1),
			padding:     spaces[0:blockOperation.Indent],
		}
		bo.insertAfter, err = regexp.Compile(blockOperation.InsertAfter)
		if err != nil {
			return nil, fmt.Errorf("error while compiling '%s': %v", blockOperation.InsertAfter, err)
		}
		patchOp.blockOps = append(patchOp.blockOps, bo)
	}
	for idx, lineOperation := range patchOperation.LineOperations {
		lo := &lineOp{
			op:      &patchOperation.LineOperations[idx],
			state:   lineInit,
			padding: spaces[0:lineOperation.Indent],
		}
		lo.regex, err = regexp.Compile(lineOperation.Regex)
		if err != nil {
			return nil, fmt.Errorf("error while compiling '%s': %v", lineOperation.Regex, err)
		}
		lo.insertAfter, err = regexp.Compile(lineOperation.InsertAfter)
		if err != nil {
			return nil, fmt.Errorf("error while compiling '%s': %v", lineOperation.InsertAfter, err)
		}
		patchOp.lineOps = append(patchOp.lineOps, lo)
	}
	return patchOp, nil
}

func (op *patchOp) doPatch(lines []string) ([]string, error) {
	for idx, bop := range op.blockOps {
		var err error
		lines, err = bop.patch(lines, op.patch.Remove)
		if err != nil {
			return nil, fmt.Errorf("blockOp#%d: %v", idx, err)
		}
	}
	for idx, lop := range op.lineOps {
		var err error
		lines, err = lop.patch(lines, op.patch.Remove)
		if err != nil {
			return nil, fmt.Errorf("lineOp#%d: %v", idx, err)
		}
	}
	return lines, nil
}
