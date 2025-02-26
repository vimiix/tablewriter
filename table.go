// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

// Package tablewriter Create & Generate text based table
package tablewriter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
)

const (
	MAX_ROW_WIDTH = 30
)

const (
	CENTER  = "+"
	ROW     = "-"
	COLUMN  = "|"
	SPACE   = " "
	NEWLINE = "\n"
)

const (
	ALIGN_DEFAULT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
	ALIGN_LEFT
)

var (
	decimal = regexp.MustCompile(`^-?(?:\d{1,3}(?:,\d{3})*|\d+)(?:\.\d+)?$`)
	percent = regexp.MustCompile(`^-?\d+\.?\d*$%$`)
)

type Border struct {
	Left   bool
	Right  bool
	Top    bool
	Bottom bool
}

type symbolID int

// Symbol ID constants which indicates the compass points, in order NESW, where
// a given symbol has connections to. The order here matches the order of box
// drawing unicode symbols.
const (
	symEW symbolID = iota
	symNS
	symES
	symSW
	symNE
	symNW
	symNES
	symNSW
	symESW
	symNEW
	symNESW
)

type Table struct {
	out                     io.Writer
	rows                    [][]string
	lines                   [][][]string
	cs                      map[int]int
	rs                      map[int]int
	headers                 [][]string
	footers                 [][]string
	caption                 bool
	captionText             string
	autoFmt                 bool
	autoWrap                bool
	reflowText              bool
	mW                      int
	syms                    []string
	pCenter                 string
	pRow                    string
	pColumn                 string
	tColumn                 int
	tRow                    int
	hAlign                  int
	fAlign                  int
	align                   int
	newLine                 string
	rowLine                 bool
	autoMergeCells          bool
	columnsToAutoMergeCells map[int]bool
	noWhiteSpace            bool
	tablePadding            string
	hdrLine                 bool
	borders                 Border
	colSize                 int
	headerParams            []string
	columnsParams           []string
	footerParams            []string
	columnsAlign            []int
}

// NewWriter Start New Table
// Take io.Writer Directly
func NewWriter(writer io.Writer) *Table {
	t := &Table{
		out:           writer,
		rows:          [][]string{},
		lines:         [][][]string{},
		cs:            make(map[int]int),
		rs:            make(map[int]int),
		headers:       [][]string{},
		footers:       [][]string{},
		caption:       false,
		captionText:   "Table caption.",
		autoFmt:       true,
		autoWrap:      true,
		reflowText:    true,
		mW:            MAX_ROW_WIDTH,
		syms:          simpleSyms(CENTER, ROW, COLUMN),
		pCenter:       CENTER,
		pRow:          ROW,
		pColumn:       COLUMN,
		tColumn:       -1,
		tRow:          -1,
		hAlign:        ALIGN_DEFAULT,
		fAlign:        ALIGN_DEFAULT,
		align:         ALIGN_DEFAULT,
		newLine:       NEWLINE,
		rowLine:       false,
		hdrLine:       true,
		borders:       Border{Left: true, Right: true, Bottom: true, Top: true},
		colSize:       -1,
		headerParams:  []string{},
		columnsParams: []string{},
		footerParams:  []string{},
		columnsAlign:  []int{}}
	return t
}

// Render table output
func (t *Table) Render() {
	if t.borders.Top {
		t.printLine(true, false)
	}
	t.printHeading()
	if t.autoMergeCells {
		t.printRowsMergeCells()
	} else {
		t.printRows()
	}
	if !t.rowLine && t.borders.Bottom {
		t.printLine(false, len(t.footers) == 0)
	}
	t.printFooter()

	if t.caption {
		t.printCaption()
	}
}

const (
	headerRowIdx = -1
	footerRowIdx = -2
)

// SetHeader Set table header
func (t *Table) SetHeader(keys []string) {
	t.colSize = len(keys)
	for i, v := range keys {
		lines := t.parseDimension(v, i, headerRowIdx)
		t.headers = append(t.headers, lines)
	}
}

// SetFooter Set table Footer
func (t *Table) SetFooter(keys []string) {
	//t.colSize = len(keys)
	for i, v := range keys {
		lines := t.parseDimension(v, i, footerRowIdx)
		t.footers = append(t.footers, lines)
	}
}

// SetCaption Set table Caption
func (t *Table) SetCaption(caption bool, captionText ...string) {
	t.caption = caption
	if len(captionText) == 1 {
		t.captionText = captionText[0]
	}
}

// SetAutoFormatHeaders Turn header autoformatting on/off. Default is on (true).
func (t *Table) SetAutoFormatHeaders(auto bool) {
	t.autoFmt = auto
}

// SetAutoWrapText Turn automatic multiline text adjustment on/off. Default is on (true).
func (t *Table) SetAutoWrapText(auto bool) {
	t.autoWrap = auto
}

// SetReflowDuringAutoWrap Turn automatic reflowing of multiline text when rewrapping. Default is on (true).
func (t *Table) SetReflowDuringAutoWrap(auto bool) {
	t.reflowText = auto
}

// SetColWidth Set the Default column width
func (t *Table) SetColWidth(width int) {
	t.mW = width
}

// SetColMinWidth Set the minimal width for a column
func (t *Table) SetColMinWidth(column int, width int) {
	t.cs[column] = width
}

// SetColumnSeparator Set the Column Separator
func (t *Table) SetColumnSeparator(sep string) {
	t.pColumn = sep
	t.syms = simpleSyms(t.pCenter, t.pRow, t.pColumn)
}

// SetRowSeparator Set the Row Separator
func (t *Table) SetRowSeparator(sep string) {
	t.pRow = sep
	t.syms = simpleSyms(t.pCenter, t.pRow, t.pColumn)
}

// SetCenterSeparator Set the center Separator
func (t *Table) SetCenterSeparator(sep string) {
	t.pCenter = sep
	t.syms = simpleSyms(t.pCenter, t.pRow, t.pColumn)
}

// SetHeaderAlignment Set Header Alignment
func (t *Table) SetHeaderAlignment(hAlign int) {
	t.hAlign = hAlign
}

// SetFooterAlignment Set Footer Alignment
func (t *Table) SetFooterAlignment(fAlign int) {
	t.fAlign = fAlign
}

// SetAlignment Set Table Alignment
func (t *Table) SetAlignment(align int) {
	t.align = align
}

// SetNoWhiteSpace Set No White Space
func (t *Table) SetNoWhiteSpace(allow bool) {
	t.noWhiteSpace = allow
}

// SetTablePadding Set Table Padding
func (t *Table) SetTablePadding(padding string) {
	t.tablePadding = padding
}

// SetColumnAlignment Set Column Alignment
func (t *Table) SetColumnAlignment(keys []int) {
	for _, v := range keys {
		switch v {
		case ALIGN_CENTER:
			break
		case ALIGN_LEFT:
			break
		case ALIGN_RIGHT:
			break
		default:
			v = ALIGN_DEFAULT
		}
		t.columnsAlign = append(t.columnsAlign, v)
	}
}

// SetNewLine Set New Line
func (t *Table) SetNewLine(nl string) {
	t.newLine = nl
}

// SetHeaderLine Set Header Line
// This would enable / disable a line after the header
func (t *Table) SetHeaderLine(line bool) {
	t.hdrLine = line
}

// SetRowLine Set Row Line
// This would enable / disable a line on each row of the table
func (t *Table) SetRowLine(line bool) {
	t.rowLine = line
}

// SetAutoMergeCells Set Auto Merge Cells
// This would enable / disable the merge of cells with identical values
func (t *Table) SetAutoMergeCells(auto bool) {
	t.autoMergeCells = auto
}

// SetAutoMergeCellsByColumnIndex Set Auto Merge Cells By Column Index
// This would enable / disable the merge of cells with identical values for specific columns
// If cols is empty, it is the same as `SetAutoMergeCells(true)`.
func (t *Table) SetAutoMergeCellsByColumnIndex(cols []int) {
	t.autoMergeCells = true
	if len(cols) > 0 {
		m := make(map[int]bool)
		for _, col := range cols {
			m[col] = true
		}
		t.columnsToAutoMergeCells = m
	}
}

// SetBorder Set Table Border
// This would enable / disable line around the table
// Deprecated: use EnableBorder
func (t *Table) SetBorder(border bool) {
	t.EnableBorder(border)
}

// EnableBorder Set Table Border
// This would enable / disable line around the table
func (t *Table) EnableBorder(border bool) {
	t.SetBorders(Border{border, border, border, border})
}

// SetBorders SetBorder Set Custom Table Border
func (t *Table) SetBorders(border Border) {
	t.borders = border
}

// SetStructs sets header and rows from slice of struct.
// If something that is not a slice is passed, error will be returned.
// The tag specified by "tablewriter" for the struct becomes the header.
// If not specified or empty, the field name will be used.
// The field of the first element of the slice is used as the header.
// If the element implements fmt.Stringer, the result will be used.
// And the slice contains nil, it will be skipped without rendering.
func (t *Table) SetStructs(v interface{}) error {
	if v == nil {
		return errors.New("nil value")
	}
	vt := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)
	switch vt.Kind() {
	case reflect.Slice, reflect.Array:
		if vv.Len() < 1 {
			return errors.New("empty value")
		}

		// check first element to set header
		first := vv.Index(0)
		e := first.Type()
		switch e.Kind() {
		case reflect.Struct:
			// OK
		case reflect.Ptr:
			if first.IsNil() {
				return errors.New("the first element is nil")
			}
			e = first.Elem().Type()
			if e.Kind() != reflect.Struct {
				return fmt.Errorf("invalid kind %s", e.Kind())
			}
		default:
			return fmt.Errorf("invalid kind %s", e.Kind())
		}
		n := e.NumField()
		headers := make([]string, n)
		for i := 0; i < n; i++ {
			f := e.Field(i)
			header := f.Tag.Get("tablewriter")
			if header == "" {
				header = f.Name
			}
			headers[i] = header
		}
		t.SetHeader(headers)

		for i := 0; i < vv.Len(); i++ {
			item := reflect.Indirect(vv.Index(i))
			itemType := reflect.TypeOf(item)
			switch itemType.Kind() {
			case reflect.Struct:
				// OK
			default:
				return fmt.Errorf("invalid item type %v", itemType.Kind())
			}
			if !item.IsValid() {
				// skip rendering
				continue
			}
			nf := item.NumField()
			if n != nf {
				return errors.New("invalid num of field")
			}
			rows := make([]string, nf)
			for j := 0; j < nf; j++ {
				f := reflect.Indirect(item.Field(j))
				if f.Kind() == reflect.Ptr {
					f = f.Elem()
				}
				if f.IsValid() {
					if s, ok := f.Interface().(fmt.Stringer); ok {
						rows[j] = s.String()
						continue
					}
					rows[j] = fmt.Sprint(f)
				} else {
					rows[j] = "nil"
				}
			}
			t.Append(rows)
		}
	default:
		return fmt.Errorf("invalid type %T", v)
	}
	return nil
}

// Append row to table
func (t *Table) Append(row []string) {
	rowSize := len(t.headers)
	if rowSize > t.colSize {
		t.colSize = rowSize
	}

	n := len(t.lines)
	line := [][]string{}
	for i, v := range row {

		// Detect string  width
		// Detect String height
		// Break strings into words
		out := t.parseDimension(v, i, n)

		// Append broken words
		line = append(line, out)
	}
	t.lines = append(t.lines, line)
}

// Rich Append row to table with color attributes
func (t *Table) Rich(row []string, colors []Colors) {
	rowSize := len(t.headers)
	if rowSize > t.colSize {
		t.colSize = rowSize
	}

	n := len(t.lines)
	line := [][]string{}
	for i, v := range row {

		// Detect string  width
		// Detect String height
		// Break strings into words
		out := t.parseDimension(v, i, n)

		if len(colors) > i {
			color := colors[i]
			for idx := range out {
				out[idx] = format(out[idx], color)
			}
		}

		// Append broken words
		line = append(line, out)
	}
	t.lines = append(t.lines, line)
}

// AppendBulk Allow Support for Bulk Append
// Eliminates repeated for loops
func (t *Table) AppendBulk(rows [][]string) {
	for _, row := range rows {
		t.Append(row)
	}
}

// NumLines to get the number of lines
func (t *Table) NumLines() int {
	return len(t.lines)
}

// ClearRows Clear rows
func (t *Table) ClearRows() {
	t.lines = [][][]string{}
}

// ClearFooter Clear footer
func (t *Table) ClearFooter() {
	t.footers = [][]string{}
}

// Center based on position and border.
func (t *Table) center(i int, isFirstRow, isLastRow bool) string {
	if i == -1 {
		if !t.borders.Left {
			return t.syms[symEW]
		}
		if isFirstRow {
			return t.syms[symES]
		}
		if isLastRow {
			return t.syms[symNE]
		}
		return t.syms[symNES]
	}

	if i == len(t.cs)-1 {
		if !t.borders.Right {
			return t.syms[symEW]
		}
		if isFirstRow {
			return t.syms[symSW]
		}
		if isLastRow {
			return t.syms[symNW]
		}
		return t.syms[symNSW]
	}

	if isFirstRow {
		return t.syms[symESW]
	}
	if isLastRow {
		return t.syms[symNEW]
	}
	return t.syms[symNESW]
}

// Print line based on row width
func (t *Table) printLine(isFirst, isLast bool) {
	fmt.Fprint(t.out, t.center(-1, isFirst, isLast))
	for i := 0; i < len(t.cs); i++ {
		v := t.cs[i]
		fmt.Fprintf(t.out, "%s%s%s%s",
			t.syms[symEW],
			strings.Repeat(t.syms[symEW], v),
			t.syms[symEW],
			t.center(i, isFirst, isLast))
	}
	fmt.Fprint(t.out, t.newLine)
}

// Print line based on row width with our without cell separator
func (t *Table) printLineOptionalCellSeparators(nl bool, displayCellSeparator []bool) {
	fmt.Fprint(t.out, t.syms[symNES])
	centerSym := symNESW
	for i := 0; i < len(t.cs); i++ {
		v := t.cs[i]
		if i == len(t.cs)-1 {
			centerSym = symNSW
		}
		if i > len(displayCellSeparator) || displayCellSeparator[i] {
			// Display the cell separator
			fmt.Fprintf(t.out, "%s%s%s%s",
				t.syms[symEW],
				strings.Repeat(string(t.syms[symEW]), v),
				t.syms[symEW],
				t.syms[centerSym])
		} else {
			// Don't display the cell separator for this cell
			fmt.Fprintf(t.out, "%s%s",
				strings.Repeat(" ", v+2),
				t.syms[centerSym])
		}
	}
	if nl {
		fmt.Fprint(t.out, t.newLine)
	}
}

// Return the PadRight function if align is left, PadLeft if align is right,
// and Pad by default
func pad(align int) func(string, string, int) string {
	padFunc := Pad
	switch align {
	case ALIGN_LEFT:
		padFunc = PadRight
	case ALIGN_RIGHT:
		padFunc = PadLeft
	}
	return padFunc
}

// Print heading information
func (t *Table) printHeading() {
	// Check if headers is available
	if len(t.headers) < 1 {
		return
	}

	// Identify last column
	end := len(t.cs) - 1

	// Get pad function
	padFunc := pad(t.hAlign)

	// Checking for ANSI escape sequences for header
	is_esc_seq := false
	if len(t.headerParams) > 0 {
		is_esc_seq = true
	}

	// Maximum height.
	max := t.rs[headerRowIdx]

	// Print Heading
	for x := 0; x < max; x++ {
		// Check if border is set
		// Replace with space if not set
		if !t.noWhiteSpace {
			fmt.Fprint(t.out, ConditionString(t.borders.Left, t.syms[symNS], SPACE))
		}

		for y := 0; y <= end; y++ {
			v := t.cs[y]
			h := ""

			if y < len(t.headers) && x < len(t.headers[y]) {
				h = t.headers[y][x]
			}
			if t.autoFmt {
				h = Title(h)
			}
			pad := ConditionString((y == end && !t.borders.Left), SPACE, t.syms[symNS])
			if t.noWhiteSpace {
				pad = ConditionString((y == end && !t.borders.Left), SPACE, t.tablePadding)
			}
			if is_esc_seq {
				if !t.noWhiteSpace {
					fmt.Fprintf(t.out, " %s %s",
						format(padFunc(h, SPACE, v),
							t.headerParams[y]), pad)
				} else {
					fmt.Fprintf(t.out, "%s %s",
						format(padFunc(h, SPACE, v),
							t.headerParams[y]), pad)
				}
			} else {
				if !t.noWhiteSpace {
					fmt.Fprintf(t.out, " %s %s",
						padFunc(h, SPACE, v),
						pad)
				} else {
					// the spaces between breaks the kube formatting
					fmt.Fprintf(t.out, "%s%s",
						padFunc(h, SPACE, v),
						pad)
				}
			}
		}
		// Next line
		fmt.Fprint(t.out, t.newLine)
	}
	if t.hdrLine {
		t.printLine(false, false)
	}
}

// Print heading information
func (t *Table) printFooter() {
	// Check if headers is available
	if len(t.footers) < 1 {
		return
	}

	// Only print line if border is not set
	if !t.borders.Bottom {
		t.printLine(false, false)
	}

	// Identify last column
	end := len(t.cs) - 1

	// Get pad function
	padFunc := pad(t.fAlign)

	// Checking for ANSI escape sequences for header
	is_esc_seq := false
	if len(t.footerParams) > 0 {
		is_esc_seq = true
	}

	// Maximum height.
	max := t.rs[footerRowIdx]

	// Print Footer
	for i := 0; i < (len(t.cs) - len(t.footers)); i++ {
		lines := t.parseDimension(" ", len(t.footers), footerRowIdx)
		t.footers = append(t.footers, lines)
	}
	erasePad := make([]bool, len(t.footers))
	for x := 0; x < max; x++ {
		// Check if border is set
		// Replace with space if not set
		fmt.Fprint(t.out, ConditionString(t.borders.Bottom, t.syms[symNS], SPACE))

		for y := 0; y <= end; y++ {
			v := t.cs[y]
			f := ""
			if y < len(t.footers) && x < len(t.footers[y]) {
				f = t.footers[y][x]
			}
			if t.autoFmt {
				f = Title(f)
			}
			pad := ConditionString((y == end && !t.borders.Top), SPACE, t.syms[symNS])

			if erasePad[y] || (x == 0 && len(f) == 0) {
				pad = SPACE
				erasePad[y] = true
			}

			if is_esc_seq {
				fmt.Fprintf(t.out, " %s %s",
					format(padFunc(f, SPACE, v),
						t.footerParams[y]), pad)
			} else {
				fmt.Fprintf(t.out, " %s %s",
					padFunc(f, SPACE, v),
					pad)
			}

			//fmt.Fprintf(t.out, " %s %s",
			//	padFunc(f, SPACE, v),
			//	pad)
		}
		// Next line
		fmt.Fprint(t.out, t.newLine)
	}

	hasPrinted := false

	for i := 0; i <= end; i++ {
		v := t.cs[i]
		pad := t.syms[symEW]
		center := t.syms[symNEW]
		length := len(t.footers[i][0])

		if length > 0 {
			hasPrinted = true
		}

		// Set center to be space if length is 0
		if length == 0 && !t.borders.Right {
			center = SPACE
		}

		// Print first junction
		if i == 0 {
			if length > 0 && !t.borders.Left {
				center = t.syms[symEW]
			} else if center != SPACE {
				center = t.syms[symNE]
			}
			fmt.Fprint(t.out, center)
		}

		// Pad With space of length is 0
		if length == 0 {
			pad = SPACE
		}
		// Ignore left space as it has printed before
		if hasPrinted || t.borders.Left {
			pad = t.syms[symEW]
			center = t.syms[symNEW]
		}

		// Change Center end position
		if center != SPACE {
			if i == end {
				if t.borders.Right {
					center = t.syms[symNW]
				} else {
					center = t.syms[symEW]
				}
			}
		}

		// Change Center start position
		if center == SPACE {
			if i < end && len(t.footers[i+1][0]) != 0 {
				if !t.borders.Left {
					center = t.syms[symEW]
				} else {
					center = t.syms[symNEW]
				}
			}
		}

		// Print the footer
		fmt.Fprintf(t.out, "%s%s%s%s",
			pad,
			strings.Repeat(string(pad), v),
			pad,
			center)

	}

	fmt.Fprint(t.out, t.newLine)
}

// Print caption text
func (t *Table) printCaption() {
	width := t.getTableWidth()
	paragraph, _ := WrapString(t.captionText, width)
	for linecount := 0; linecount < len(paragraph); linecount++ {
		fmt.Fprintln(t.out, paragraph[linecount])
	}
}

// Calculate the total number of characters in a row
func (t *Table) getTableWidth() int {
	var chars int
	for _, v := range t.cs {
		chars += v
	}

	// Add chars, spaces, seperators to calculate the total width of the table.
	// ncols := t.colSize
	// spaces := ncols * 2
	// seps := ncols + 1

	return (chars + (3 * t.colSize) + 2)
}

// printRows - print all the rows
func (t *Table) printRows() {
	for i, lines := range t.lines {
		t.printRow(lines, i)
	}
}

// fillAlignment - fill the alignment
func (t *Table) fillAlignment(num int) {
	if len(t.columnsAlign) < num {
		t.columnsAlign = make([]int, num)
		for i := range t.columnsAlign {
			t.columnsAlign[i] = t.align
		}
	}
}

// Print Row Information
// Adjust column alignment based on type
func (t *Table) printRow(columns [][]string, rowIdx int) {
	// Get Maximum Height
	max := t.rs[rowIdx]
	total := len(columns)

	// TODO Fix uneven col size
	// if total < t.colSize {
	//	for n := t.colSize - total; n < t.colSize ; n++ {
	//		columns = append(columns, []string{SPACE})
	//		t.cs[n] = t.mW
	//	}
	//}

	// Pad Each Height
	pads := []int{}

	// Checking for ANSI escape sequences for columns
	is_esc_seq := false
	if len(t.columnsParams) > 0 {
		is_esc_seq = true
	}
	t.fillAlignment(total)

	for i, line := range columns {
		length := len(line)
		pad := max - length
		pads = append(pads, pad)
		for n := 0; n < pad; n++ {
			columns[i] = append(columns[i], "  ")
		}
	}
	//fmt.Println(max, "\n")
	for x := 0; x < max; x++ {
		for y := 0; y < total; y++ {

			// Check if border is set
			if !t.noWhiteSpace {
				fmt.Fprint(t.out, ConditionString((!t.borders.Left && y == 0), SPACE, t.syms[symNS]))
				fmt.Fprintf(t.out, SPACE)
			}

			str := columns[y][x]

			// Embedding escape sequence with column value
			if is_esc_seq {
				str = format(str, t.columnsParams[y])
			}

			// This would print alignment
			// Default alignment  would use multiple configuration
			switch t.columnsAlign[y] {
			case ALIGN_CENTER: //
				fmt.Fprintf(t.out, "%s", Pad(str, SPACE, t.cs[y]))
			case ALIGN_RIGHT:
				fmt.Fprintf(t.out, "%s", PadLeft(str, SPACE, t.cs[y]))
			case ALIGN_LEFT:
				fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))
			default:
				if decimal.MatchString(strings.TrimSpace(str)) || percent.MatchString(strings.TrimSpace(str)) {
					fmt.Fprintf(t.out, "%s", PadLeft(str, SPACE, t.cs[y]))
				} else {
					fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))

					// TODO Custom alignment per column
					//if max == 1 || pads[y] > 0 {
					//	fmt.Fprintf(t.out, "%s", Pad(str, SPACE, t.cs[y]))
					//} else {
					//	fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))
					//}

				}
			}
			if !t.noWhiteSpace {
				fmt.Fprintf(t.out, SPACE)
			} else {
				fmt.Fprintf(t.out, t.tablePadding)
			}
		}
		// Check if border is set
		// Replace with space if not set
		if !t.noWhiteSpace {
			fmt.Fprint(t.out, ConditionString(t.borders.Left, t.syms[symNS], SPACE))
		}
		fmt.Fprint(t.out, t.newLine)
	}

	if t.rowLine {
		t.printLine(false, rowIdx == len(t.lines)-1 && len(t.footers) == 0)
	}
}

// Print the rows of the table and merge the cells that are identical
func (t *Table) printRowsMergeCells() {
	var previousLine []string
	var displayCellBorder []bool
	var tmpWriter bytes.Buffer
	for i, lines := range t.lines {
		// We store the display of the current line in a tmp writer, as we need to know which border needs to be print above
		previousLine, displayCellBorder = t.printRowMergeCells(&tmpWriter, lines, i, previousLine)
		if i > 0 { //We don't need to print borders above first line
			if t.rowLine {
				t.printLineOptionalCellSeparators(true, displayCellBorder)
			}
		}
		tmpWriter.WriteTo(t.out)
	}
	//Print the end of the table
	if t.rowLine {
		t.printLine(false, true)
	}
}

// Print Row Information to a writer and merge identical cells.
// Adjust column alignment based on type
func (t *Table) printRowMergeCells(writer io.Writer, columns [][]string, rowIdx int, previousLine []string) ([]string, []bool) {
	// Get Maximum Height
	max := t.rs[rowIdx]
	total := len(columns)

	// Pad Each Height
	pads := []int{}

	// Checking for ANSI escape sequences for columns
	isEscSeq := false
	if len(t.columnsParams) > 0 {
		isEscSeq = true
	}
	for i, line := range columns {
		length := len(line)
		pad := max - length
		pads = append(pads, pad)
		for n := 0; n < pad; n++ {
			columns[i] = append(columns[i], "  ")
		}
	}

	var displayCellBorder []bool
	t.fillAlignment(total)
	for x := 0; x < max; x++ {
		for y := 0; y < total; y++ {

			// Check if border is set
			fmt.Fprint(writer, ConditionString((!t.borders.Left && y == 0), SPACE, t.syms[symNS]))

			fmt.Fprintf(writer, SPACE)

			str := columns[y][x]

			// Embedding escape sequence with column value
			if isEscSeq {
				str = format(str, t.columnsParams[y])
			}

			if t.autoMergeCells {
				var mergeCell bool
				if t.columnsToAutoMergeCells != nil {
					// Check to see if the column index is in columnsToAutoMergeCells.
					if t.columnsToAutoMergeCells[y] {
						mergeCell = true
					}
				} else {
					// columnsToAutoMergeCells was not set.
					mergeCell = true
				}
				//Store the full line to merge mutli-lines cells
				fullLine := strings.TrimRight(strings.Join(columns[y], " "), " ")
				if len(previousLine) > y && fullLine == previousLine[y] && fullLine != "" && mergeCell {
					// If this cell is identical to the one above but not empty, we don't display the border and keep the cell empty.
					displayCellBorder = append(displayCellBorder, false)
					str = ""
				} else {
					// First line or different content, keep the content and print the cell border
					displayCellBorder = append(displayCellBorder, true)
				}
			}

			// This would print alignment
			// Default alignment  would use multiple configuration
			switch t.columnsAlign[y] {
			case ALIGN_CENTER: //
				fmt.Fprintf(writer, "%s", Pad(str, SPACE, t.cs[y]))
			case ALIGN_RIGHT:
				fmt.Fprintf(writer, "%s", PadLeft(str, SPACE, t.cs[y]))
			case ALIGN_LEFT:
				fmt.Fprintf(writer, "%s", PadRight(str, SPACE, t.cs[y]))
			default:
				if decimal.MatchString(strings.TrimSpace(str)) || percent.MatchString(strings.TrimSpace(str)) {
					fmt.Fprintf(writer, "%s", PadLeft(str, SPACE, t.cs[y]))
				} else {
					fmt.Fprintf(writer, "%s", PadRight(str, SPACE, t.cs[y]))
				}
			}
			fmt.Fprintf(writer, SPACE)
		}
		// Check if border is set
		// Replace with space if not set
		fmt.Fprint(writer, ConditionString(t.borders.Left, t.syms[symNS], SPACE))
		fmt.Fprint(writer, t.newLine)
	}

	//The new previous line is the current one
	previousLine = make([]string, total)
	for y := 0; y < total; y++ {
		previousLine[y] = strings.TrimRight(strings.Join(columns[y], " "), " ") //Store the full line for multi-lines cells
	}
	//Returns the newly added line and wether or not a border should be displayed above.
	return previousLine, displayCellBorder
}

// parseDimension - parse table dimensions
func (t *Table) parseDimension(str string, colKey, rowKey int) []string {
	var (
		raw      []string
		maxWidth int
	)

	raw = getLines(str)
	maxWidth = 0
	for _, line := range raw {
		if w := DisplayWidth(line); w > maxWidth {
			maxWidth = w
		}
	}

	// If wrapping, ensure that all paragraphs in the cell fit in the
	// specified width.
	if t.autoWrap {
		// If there's a maximum allowed width for wrapping, use that.
		if maxWidth > t.mW {
			maxWidth = t.mW
		}

		// In the process of doing so, we need to recompute maxWidth. This
		// is because perhaps a word in the cell is longer than the
		// allowed maximum width in t.mW.
		newMaxWidth := maxWidth
		newRaw := make([]string, 0, len(raw))

		if t.reflowText {
			// Make a single paragraph of everything.
			raw = []string{strings.Join(raw, " ")}
		}
		for i, para := range raw {
			paraLines, _ := WrapString(para, maxWidth)
			for _, line := range paraLines {
				if w := DisplayWidth(line); w > newMaxWidth {
					newMaxWidth = w
				}
			}
			if i > 0 {
				newRaw = append(newRaw, " ")
			}
			newRaw = append(newRaw, paraLines...)
		}
		raw = newRaw
		maxWidth = newMaxWidth
	}

	// Store the new known maximum width.
	v, ok := t.cs[colKey]
	if !ok || v < maxWidth || v == 0 {
		t.cs[colKey] = maxWidth
	}

	// Remember the number of lines for the row printer.
	h := len(raw)
	v, ok = t.rs[rowKey]

	if !ok || v < h || v == 0 {
		t.rs[rowKey] = h
	}
	//fmt.Printf("Raw %+v %d\n", raw, len(raw))
	return raw
}
