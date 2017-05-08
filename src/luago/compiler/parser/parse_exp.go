package parser

import . "luago/number"
import . "luago/compiler/ast"
import . "luago/compiler/lexer"

// explist ::= exp {‘,’ exp}
func parseExpList(lexer *Lexer) []Exp {
	exps := make([]Exp, 0, 4)
	exps = append(exps, parseExp(lexer))
	for lexer.LookAhead(1) == TOKEN_SEP_COMMA {
		lexer.NextToken()
		exps = append(exps, parseExp(lexer))
	}
	return exps
}

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp
*/
func parseExp(lexer *Lexer) Exp {
	return parseExp12(lexer)
}

// x or y
func parseExp12(lexer *Lexer) Exp {
	exp := parseExp11(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_OR {
			line, op, _ := lexer.NextToken()
			lor := &BinopExp{line, op, 12, exp, parseExp11(lexer)}
			exp = optimizeLogicalOr(lor)
		} else {
			break
		}
	}
	// for the convenience of codegen
	return changeAssociative(exp, 12)
}

// x and y
func parseExp11(lexer *Lexer) Exp {
	exp := parseExp10(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_AND {
			line, op, _ := lexer.NextToken()
			land := &BinopExp{line, op, 11, exp, parseExp10(lexer)}
			last := lexer.LookAhead(1) != TOKEN_OP_AND // todo
			exp = optimizeLogicalAnd(land, last)
		} else {
			break
		}
	}
	// for the convenience of codegen
	return changeAssociative(exp, 11)
}

// compare
func parseExp10(lexer *Lexer) Exp {
	exp := parseExp9(lexer)
	for {
		switch lexer.LookAhead(1) {
		case TOKEN_OP_LT, TOKEN_OP_GT, TOKEN_OP_NE,
			TOKEN_OP_LE, TOKEN_OP_GE, TOKEN_OP_EQ:
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{line, op, 10, exp, parseExp9(lexer)}
		default:
			return exp
		}
	}
	return exp
}

// x | y
func parseExp9(lexer *Lexer) Exp {
	exp := parseExp8(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_BOR {
			line, op, _ := lexer.NextToken()
			bor := &BinopExp{line, op, 9, exp, parseExp8(lexer)}
			exp = optimizeBitwiseBinaryOp(bor)
		} else {
			break
		}
	}
	return changeAssociative(exp, 9)
}

// x ~ y
func parseExp8(lexer *Lexer) Exp {
	exp := parseExp7(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_BXOR {
			line, op, _ := lexer.NextToken()
			bxor := &BinopExp{line, op, 8, exp, parseExp7(lexer)}
			exp = optimizeBitwiseBinaryOp(bxor)
		} else {
			break
		}
	}
	return changeAssociative(exp, 8)
}

// x & y
func parseExp7(lexer *Lexer) Exp {
	exp := parseExp6(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_BAND {
			line, op, _ := lexer.NextToken()
			band := &BinopExp{line, op, 7, exp, parseExp6(lexer)}
			exp = optimizeBitwiseBinaryOp(band)
		} else {
			break
		}
	}
	return changeAssociative(exp, 7)
}

// shift
func parseExp6(lexer *Lexer) Exp {
	exp := parseExp5(lexer)
	loop: for {
		switch lexer.LookAhead(1) {
		case TOKEN_OP_SHL, TOKEN_OP_SHR:
			line, op, _ := lexer.NextToken()
			shx := &BinopExp{line, op, 6, exp, parseExp5(lexer)}
			exp = optimizeBitwiseBinaryOp(shx)
		default:
			break loop
		}
	}
	return changeAssociative(exp, 6)
}

// a .. b
func parseExp5(lexer *Lexer) Exp {
	exp := parseExp4(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_CONCAT {
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{line, op, 5, exp, parseExp4(lexer)}
		} else {
			break
		}
	}
	// concat is right associative
	return changeAssociative(exp, 5)
}

// x +/- y
func parseExp4(lexer *Lexer) Exp {
	exp := parseExp3(lexer)
	loop: for {
		switch lexer.LookAhead(1) {
		case TOKEN_OP_ADD, TOKEN_OP_SUB:
			line, op, _ := lexer.NextToken()
			arith := &BinopExp{line, op, 4, exp, parseExp3(lexer)}
			exp = optimizeArithBinaryOp(arith)
		default:
			break loop
		}
	}

	return changeAssociative(exp, 4)
}

// *, %, /, //
func parseExp3(lexer *Lexer) Exp {
	exp := parseExp2(lexer)
	loop: for {
		switch lexer.LookAhead(1) {
		case TOKEN_OP_MUL, TOKEN_OP_MOD, TOKEN_OP_DIV, TOKEN_OP_IDIV:
			line, op, _ := lexer.NextToken()
			arith := &BinopExp{line, op, 3, exp, parseExp2(lexer)}
			exp = optimizeArithBinaryOp(arith)
		default:
			break loop
		}
	}
	return changeAssociative(exp, 3)
}

// unary
func parseExp2(lexer *Lexer) Exp {
	switch lexer.LookAhead(1) {
	case TOKEN_OP_UNM, TOKEN_OP_BNOT, TOKEN_OP_LEN, TOKEN_OP_NOT:
		line, op, _ := lexer.NextToken()
		exp := &UnopExp{line, op, parseExp2(lexer)}
		return optimizeUnaryOp(exp)
	default:
		return parseExp1(lexer)
	}
}

// x ^ y
func parseExp1(lexer *Lexer) Exp {
	exp := parseExp0(lexer)
	for {
		if lexer.LookAhead(1) == TOKEN_OP_POW {
			line, op, _ := lexer.NextToken()
			exp = &BinopExp{line, op, 1, exp, parseExp0(lexer)}
		} else {
			break
		}
	}
	// pow is right associative
	exp = changeAssociative(exp, 1)
	return optimizePow(exp)
}

func parseExp0(lexer *Lexer) Exp {
	switch lexer.LookAhead(1) {
	case TOKEN_VARARG: // ...
		line, _, _ := lexer.NextToken()
		return &VarargExp{line}
	case TOKEN_KW_NIL: // nil
		line, _, _ := lexer.NextToken()
		return &NilExp{line}
	case TOKEN_KW_TRUE: // true
		line, _, _ := lexer.NextToken()
		return &TrueExp{line}
	case TOKEN_KW_FALSE: // false
		line, _, _ := lexer.NextToken()
		return &FalseExp{line}
	case TOKEN_STRING: // LiteralString
		line, _, token := lexer.NextToken()
		return &StringExp{line, token}
	case TOKEN_NUMBER: // Numeral
		return parseNumberExp(lexer, 1)
	case TOKEN_SEP_LCURLY: // tableconstructor
		return parseTableConstructorExp(lexer)
	case TOKEN_KW_FUNCTION: // functiondef
		lexer.NextToken()
		return parseFuncDefExp(lexer)
	default: // prefixexp
		return parsePrefixExp(lexer)
	}
}

func parseNumberExp(lexer *Lexer, sign int) Exp {
	line, _, token := lexer.NextToken()
	if i, ok := ParseInteger(token); ok {
		if sign >= 0 {
			return &IntegerExp{line, i}
		} else {
			return &IntegerExp{line, -i}
		}
	} else if f, ok := ParseFloat(token); ok {
		if sign >= 0 {
			return &FloatExp{line, f}
		} else {
			return &FloatExp{line, -f}
		}
	} else { // todo
		panic("not a number: " + token)
	}
}

/*
        op1          op2
       /   \        /   \
     op2    c  =>  a    op1
    /   \              /   \
   a     b            b     c
*/
func changeAssociative(exp Exp, prec int) Exp {
	if bexp, ok := exp.(*BinopExp); ok && bexp.Prec == prec {
		for {
			if exp1, ok := bexp.Exp1.(*BinopExp); ok && exp1.Prec == prec {
				bexp.Exp1 = exp1.Exp1
				exp1.Exp1 = exp1.Exp2
				exp1.Exp2 = bexp.Exp2
				bexp.Exp2 = exp1
				bexp.Op, exp1.Op = exp1.Op, bexp.Op
			} else {
				break
			}
		}
	}
	return exp
}
