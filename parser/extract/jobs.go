package extract

import (
	"fmt"
	"sort"
	"strings"
	"github.com/InstIDEA/ddjj/parser/declaration"	
)

// Experimental version

func Jobs(e *Extractor, parser *ParserData) []*declaration.Job {

	e.BindFlag(EXTRACTOR_FLAG_1)
	e.BindFlag(EXTRACTOR_FLAG_2)

	var instituciones []*declaration.Job
	var resultsPositions []int // for valid and invalid results
	var counter = countJobs(e)
	var successful int

	e.Rewind()

	job := &declaration.Job{ }

	if counter > 0 &&
	e.MoveUntilStartWith(CurrToken, "DATOS LABORALES") {
		e.SaveLine()

		for e.Scan() {
			if counter == successful {
				break
			}

			if job.Cargo == "" {
				value := getJobTitle(e, &resultsPositions)

				if !isJobFormField(value) {
					job.Cargo = value
				}
			}

			if job.Cargo != "" &&
			job.Institucion == "" {
				value := getJobInst(e, &resultsPositions)

				if !isJobFormField(value) {
					job.Institucion = value
				}
			}

			if job.Cargo != "" && job.Institucion != "" {
				successful++
				instituciones = append(instituciones, job)
				job = &declaration.Job{ }
				e.MoveUntilSavedLine()
			}
		}
	}

	if successful != counter {
		parser.addMessage(fmt.Sprintf("ignored jobs: %d/%d", counter - successful, counter))
	}

	if instituciones == nil {
		parser.addError(fmt.Errorf("failed when extracting jobs"))
		return nil
	}

	return instituciones
}

func getJobTitle(e *Extractor, pos *[]int) string {

	if strings.Contains(e.CurrToken, "CARGO") &&
	!ContainsIntItem(*pos, e.CurrLineNum()) {
		val, check := isKeyValuePair(e.CurrToken, "CARGO")
		if check {
			*pos = append(*pos, e.CurrLineNum())
			e.MoveUntilSavedLine()
			return val
		}
	}

	if isCurrLine(e.PrevToken, "CARGO") &&
	!ContainsIntItem(*pos, e.CurrLineNum()) {
		*pos = append(*pos, e.CurrLineNum())
		subStr := e.CurrToken[: getLCSpacePos(e.CurrToken)]
		e.MoveUntilSavedLine()
		return strings.TrimSpace(subStr)		
	}

	return ""
}

func getJobInst(e *Extractor, pos *[]int) string {

	if strings.Contains(e.CurrToken, "INSTITUCIÓN") &&
	!ContainsIntItem(*pos, e.CurrLineNum()) {
		val, check := isKeyValuePair(e.CurrToken, "INSTITUCIÓN")
		if check {
			*pos = append(*pos, e.CurrLineNum())
			e.MoveUntilSavedLine()
			return removeSubstring(val, "ACTO ADM. COM.")
		}
	}

	fields := strings.Fields(e.CurrToken)

	if isCurrLine(e.PrevToken, "INSTITUCIÓN") &&
	isNumber(fields[0]) &&
	!ContainsIntItem(*pos, e.CurrLineNum()) {
		*pos = append(*pos, e.CurrLineNum())
		numPos := strings.Index(e.CurrToken, fields[0]) +2
		subStr := strings.TrimSpace(e.CurrToken[numPos:])
		e.MoveUntilSavedLine()
		return strings.TrimSpace(subStr[:getLCSpacePos(subStr)])
	}

	return ""
}

func countJobs(e *Extractor) int {
	var counter int

	for e.Scan() {
		// first position
		if isCurrLine(e.CurrToken, "CARGO") {
			counter++
			continue
		}

		// middle position
		if hasLeadingSpaces(e.CurrToken, "CARGO") &&
		!endsWith(e.CurrToken, "CARGO") {
			counter++
		}
	}
	return counter
}

func isJobFormField(s string) bool {
	formField := []string {
		"TIPO",
		"INSTITUCION",
		"DIRECCION",
		"DEPENDENCIA",
		"CATEGORIA",
		"NOMBRADO/CONTRATADO",
		"CARGO",
		"FECHA ASUNC./CESE/OTROS",
		"ACTO ADMINIST",
		"FECHA ACT. ADM",
		"TELEFONO",
		"COMISIONADO",
		"FECHA INGRESO",
		"FECHA EGRESO",
	}

	s = removeAccents(s)
	for _, value := range formField {
		if strings.Contains(s, value) {
			return true
		}
	}

	return false
}

// return the longest continuos space position
func getLCSpacePos(s string) int {
	var spaces int
	var results [][]int

	for pos, letter := range s {
		if letter == ' ' {
			spaces += 1
			continue
		}

		if spaces > 0 {
			results = append(results, []int{ pos - 1, spaces })
			spaces = 0
		}
	}

	sort.SliceStable(results, func(i, j int) bool {return results[i][1] > results[j][1]})
	return results[0][0]
}

func removeSubstring(line string, sub string) string {
	index := strings.Index(line, sub)

	if index == -1 {
		return line
	}

	return strings.TrimSpace(line[:index])
}
