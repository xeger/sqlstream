package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeger/sqlstream/nlp"
)

// Used for flags.
var (
	confidence float64

	recognizeCmd = &cobra.Command{
		Use:   "recognize",
		Short: "Test a model against input lines",
		Long: `Parses words/phrases from stdin, one per line.
Prints input lines that match the model.`,
		Run: recognize,
	}
)

func init() {
	recognizeCmd.PersistentFlags().Float64Var(&confidence, "confidence", 0.5, "minimum probability to consider a match")
}

func recognize(cmd *cobra.Command, args []string) {
	var modelFile string
	if len(args) == 1 {
		modelFile = args[0]
	} else {
		fmt.Fprintln(os.Stderr, "Usage: sqlstream train <sentences|words>")
		os.Exit(1)
	}

	var model *nlp.Model = nil
	data, err := os.ReadFile(modelFile)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(data, &model)
	if err != nil {
		panic(err.Error())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n\t")
		if model.Recognize(line, confidence) {
			fmt.Println(line)
		}
	}
}