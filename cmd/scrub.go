package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/sqlstream/scrubbing"
)

// Used for flags.
var (
	parallelism int = runtime.NumCPU()

	scrubCmd = &cobra.Command{
		Use:   "scrub",
		Short: "Mask sensitive data in a MySQL dump",
		Long:  `Parses stdin as SQL; prints masked SQL to stdout.`,
		Run:   scrub,
	}
)

func init() {
	scrubCmd.PersistentFlags().IntVar(&parallelism, "parallelism", runtime.NumCPU(), "lines to scrub at once")
}

func scrub(cmd *cobra.Command, args []string) {
	N := parallelism

	in := make([]chan string, N)
	out := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		out[i] = make(chan string)
		go scrubbing.Scrub(in[i], out[i])
	}
	drain := func(to int) {
		for i := 0; i < to; i++ {
			fmt.Print(<-out[i])
		}
	}
	done := func() {
		for i := 0; i < N; i++ {
			close(in[i])
			close(out[i])
		}
	}

	reader := bufio.NewReader(os.Stdin)
	l := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		in[l] <- line
		l = (l + 1) % N
		if l == 0 {
			drain(N)
		}
	}
	drain(l)
	done()
}