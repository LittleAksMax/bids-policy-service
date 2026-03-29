package convert

import (
	"github.com/LittleAksMax/bidscript"
	"github.com/LittleAksMax/bidscript/lexer"
	"github.com/LittleAksMax/bidscript/parser"
)

func GetScriptErrors(script string) bidscript.ParseErrors {
	l := lexer.NewLexer(script)
	p := parser.NewParser(l)
	_ = p.ParseProgram()

	errs := p.Errors()
	if len(errs) == 0 {
		return nil
	}

	return bidscript.ParseErrors(errs)
}
