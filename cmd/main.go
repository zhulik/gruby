package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/chzyer/readline"
	"github.com/k0kubun/pp"
	"github.com/zhulik/gruby"
)

func main() {
	grb := gruby.Must(gruby.New())

	defer grb.Close()

	workDir := gruby.Must(os.Getwd())

	ctx := gruby.NewCompileContext(grb)
	ctx.SetFilename(path.Join(workDir, "main.rb"))
	defer ctx.Close()

	rln, err := readline.New(">> ")
	if err != nil {
		panic(fmt.Errorf("readline init error: %w", err))
	}
	defer rln.Close()

	for {
		line, rErr := rln.Readline()
		if rErr != nil {
			if errors.Is(rErr, io.EOF) || errors.Is(rErr, readline.ErrInterrupt) {
				return
			}

			panic(fmt.Errorf("readline error: %w", err))
		}

		result, err := grb.LoadStringWithContext(line, ctx)
		if err != nil {
			pp.Printf("ERROR: %s\n", err.Error())
			continue
		}

		pp.Println(result.String())
	}
}
