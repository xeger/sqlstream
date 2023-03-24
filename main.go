package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	_ "github.com/pingcap/tidb/parser/test_driver"

	"gonum.org/v1/gonum/mathext/prng"
)

// Preserves non-parseable lines (assuming they are comments).
const doComments = true

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = true

// Turns an AST back into a string.
func restore(stmt ast.StmtNode) string {
	buf := new(bytes.Buffer)
	ctx := format.NewRestoreCtx(format.RestoreKeyWordUppercase|format.RestoreNameBackQuotes|format.RestoreStringSingleQuotes|format.RestoreStringWithoutDefaultCharset, buf)
	err := stmt.Restore(ctx)
	if err != nil {
		panic(err)
	}
	s := buf.String()
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s + ";\n"
}

// Attempts to remove sensitive data from an AST. Returns nil if the entire statement should be dropped.
func sanitize(stmt ast.StmtNode) (ast.StmtNode, bool) {
	switch st := stmt.(type) {
	// for table name: st.Table.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name
	// for raw values: st.Lists[0][0], etc...
	case *ast.InsertStmt:
		if doInserts {
			v := &scrubber{source: prng.NewMT19937()}
			st.Accept(v)
			return st, true
		} else {
			return nil, true
		}
	default:
		if doMisc {
			return stmt, false
		}
		return nil, true
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	p := parser.New()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		stmts, _, err := p.Parse(line, "", "")
		if (err != nil || len(stmts) == 0) && doComments {
			fmt.Print(line)
		}

		for _, in := range stmts {
			out, processed := sanitize(in)
			if !processed {
				fmt.Println(out.OriginalText())
			} else if out != nil {
				fmt.Print(restore(out))
			}
		}
	}
}
