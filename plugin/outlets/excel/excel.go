package excel

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/xuri/excelize/v2"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "excel",
		Factory: ExcelOutlet,
	})
}

func ExcelOutlet(ctx *engine.Context) engine.Outlet {
	recordsPerFile := ctx.Config().GetInt("records_per_file", 20_000)
	path := ctx.Config().GetString("path", "")
	ret := &excelOutlet{
		ctx:            ctx,
		recordsPerFile: recordsPerFile,
		path:           path,
		sheet:          "Sheet1",
	}
	return ret
}

type excelOutlet struct {
	ctx            *engine.Context
	recordsPerFile int
	path           string
	buffer         []engine.Record
	sheet          string
}

func (eo *excelOutlet) Open() error {
	return nil
}

func (eo *excelOutlet) Close() error {
	if err := eo.writeFile(eo.buffer); err != nil {
		return err
	}
	return nil
}

func (eo *excelOutlet) writeFile(records []engine.Record) error {
	if eo.path == "" {
		return nil
	}
	destPath := eo.path

	if eo.recordsPerFile > 0 {
		pathDir := filepath.Dir(eo.path)
		pathBase := filepath.Base(eo.path)
		pathExt := filepath.Ext(eo.path)
		if pathExt == "" {
			pathExt = ".xlsx"
		} else {
			pathBase = strings.TrimSuffix(pathBase, pathExt)
		}
		destPath = filepath.Join(pathDir, fmt.Sprintf("%s-%s%s", pathBase, time.Now().Format("20060102_150405"), pathExt))
	}

	excelFile := excelize.NewFile()
	headStyle, err := excelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Italic: false,
			Family: "Calibri",
			Size:   13,
			Color:  "777777",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		return err
	}

	cols := []string{}
	colsMap := map[string]int{}
	rows := 0
	for _, row := range records {
		rows++
		for _, name := range row.Names() {
			if _, ok := colsMap[name]; !ok {
				cols = append(cols, name)
				idx := len(cols) - 1
				// A: 65, Z: 90
				colRunes := []rune{}
				if cycle := idx / 26; cycle > 0 {
					colRunes = append(colRunes, rune('A'+cycle-1))
				}
				colRunes = append(colRunes, rune('A'+(idx%26)))
				col := string(colRunes)
				cell := fmt.Sprintf("%s%d", col, 1)

				colsMap[name] = idx
				excelFile.SetCellValue(eo.sheet, cell, name)
				excelFile.SetCellStyle(eo.sheet, cell, cell, headStyle)
				excelFile.SetColWidth(eo.sheet, col, col, 16)
			}
		}

		for i, field := range row.Fields(cols...) {
			if field == nil {
				continue
			}
			// A: 65, Z: 90
			col := []rune{}
			if cycle := i / 26; cycle > 0 {
				col = append(col, rune('A'+cycle-1))
			}
			col = append(col, rune('A'+(i%26)))

			cell := fmt.Sprintf("%s%d", string(col), rows+1)
			excelFile.SetCellValue(eo.sheet, cell, field.Value)
		}
	}

	if err := excelFile.SaveAs(destPath); err != nil {
		return err
	}
	if err := excelFile.Close(); err != nil {
		return err
	}

	return nil
}

func (eo *excelOutlet) Handle(recs []engine.Record) error {
	eo.buffer = append(eo.buffer, recs...)
	if eo.recordsPerFile <= 0 {
		return nil
	}
	rows := len(eo.buffer)
	if rows >= int(eo.recordsPerFile) {
		old := eo.buffer[0:eo.recordsPerFile]
		eo.buffer = eo.buffer[eo.recordsPerFile:]
		go func(recs []engine.Record) {
			if err := eo.writeFile(old); err != nil {
				eo.ctx.LogError("failed to write excel file: %v", err)
			}
		}(old)
	}
	return nil
}
