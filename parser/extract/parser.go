package extract

import (
	"bufio"
	"code.sajari.com/docconv"
	"encoding/json"
	"fmt"
	"github.com/InstIDEA/ddjj/parser/declaration"
	"io"
	"strings"
	"time"
)

type ParserData struct {
	Message []string                 `json:"message"`
	Status  int                      `json:"status"`
	Data    *declaration.Declaration `json:"data"`
	Raw     []string                 `json:"raw"`
}

func (parser *ParserData) addMessage(msg string) {
	parser.Message = append(parser.Message, msg)
}

func (parser *ParserData) addError(msg error) {
	parser.Message = append(parser.Message, msg.Error())
}

func (parser *ParserData) rawData(s string) {
	parser.Raw = append(parser.Raw, s)
}

func (parser *ParserData) Print() {
	b, err := json.MarshalIndent(parser, "", "\t")
	if err != nil {
		fmt.Println("{ message: null, status: 0, data: null, raw: null }")
		return
	}
	fmt.Println(string(b))
}

func (parser *ParserData) checkStr(val string, err error) string {
	if err != nil {
		parser.addError(err)
	}
	return val
}

func (parser *ParserData) checkInt(val int, err error) int {
	if err != nil {
		parser.addError(err)
	}
	return val
}

func (parser *ParserData) check(val time.Time, err error) time.Time {
	if err != nil {
		parser.addError(err)
	}
	return val
}

func CreateError(msg string) ParserData {
	return ParserData{
		Message: []string{msg},
		Status:  -1,
		Data:    nil,
		Raw:     nil,
	}
}

func ParsePDF(file io.Reader) ParserData {

	var parser = ParserData{
		Message: make([]string, 0),
		Status:  0,
		Data:    nil,
		Raw:     make([]string, 0),
	}
	res, err := docconv.Convert(file, "application/pdf", false)

	if err != nil {
		parser.addError(err)
		parser.Status = 1
		return parser
	}

	res2, err2 := docconv.Convert(file, "application/pdf", true)

	if err2 != nil {
		parser.addError(err2)
		parser.Status = 1
		return parser
	}

	parser.rawData(res.Body)

	d := &declaration.Declaration{}	

	// Basic Info.
	d.Fecha = parser.check(Date(NewExtractor(res.Body)))
	d.Cedula = parser.checkInt(Cedula(NewExtractor(res.Body)))
	d.Nombre = parser.checkStr(Name(NewExtractor(res.Body)))
	d.Apellido = parser.checkStr(Lastname(NewExtractor(res.Body)))

	// Spouse
	d.Conyuge = parser.checkStr(Spouse(NewExtractor(res.Body)))

	// Jobs
	d.Instituciones = Jobs(NewExtractor(res2.Body), &parser)

	// Deposits
	scanner := bufio.NewScanner(strings.NewReader(res.Body))
	d.Deposits, err = Deposits(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Debtors.
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.Debtors, err = Debtors(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Real state.
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.RealStates, err = RealStates(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Vehicles
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.Vehicles = Vehicles(scanner, &parser)

	// Agricultural activity
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.Agricultural, err = Agricultural(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Furniture
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.Furniture, err = Furniture(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Other assets
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.OtherAssets, err = Assets(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Debts
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.Debts, err = Debts(scanner)
	if err != nil {
		parser.addError(err)
	}

	// Income and Expenses
	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.IncomeMonthly = MonthlyIncome(scanner)

	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.IncomeAnnual = AnnualIncome(scanner)

	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.ExpensesMonthly = MonthlyExpenses(scanner)

	scanner = bufio.NewScanner(strings.NewReader(res.Body))
	d.ExpensesAnnual = AnnualExpenses(scanner)

	// Summary
	d.Resumen = Summary(NewExtractor(res.Body), &parser)

	d.CalculatePatrimony()

	if d.Resumen != nil {
		if d.Assets != d.Resumen.TotalActivo {
			parser.addMessage("calculated assets and summary assets does not match")
		}
	
		if d.Liabilities != d.Resumen.TotalPasivo {
			parser.addMessage("calculated liabilities and summary liabilities does not match")
		}
	
		if d.NetPatrimony != d.Resumen.PatrimonioNeto {
			parser.addMessage("calculated net patrimony and summary net patrimony does not match")
		}
	}

	parser.Data = d
	return parser
}
