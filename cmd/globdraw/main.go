package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/kenshaw/glob"
	"github.com/kenshaw/glob/match"
)

func main() {
	pattern := flag.String("p", "", "pattern to draw")
	sep := flag.String("s", "", "comma separated list of separators characters")
	filepath := flag.String("file", "", "path for patterns file")
	auto := flag.Bool("auto", false, "autoopen result")
	offset := flag.Int("offset", 0, "patterns to skip")
	flag.Parse()
	if err := run(*pattern, *sep, *filepath, *auto, *offset); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(pattern, sep, filepath string, auto bool, offset int) error {
	var patterns []string
	if pattern != "" {
		patterns = append(patterns, pattern)
	}
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			return fmt.Errorf("could not open %s: %w", filepath, err)
		}
		s := bufio.NewScanner(file)
		for s.Scan() {
			fmt.Println(offset)
			if offset > 0 {
				offset--
				fmt.Println("skipped")
				continue
			}
			patterns = append(patterns, s.Text())
		}
		file.Close()
	}
	if len(patterns) == 0 {
		return nil
	}
	var separators []rune
	if len(sep) > 0 {
		for c := range strings.SplitSeq(sep, ",") {
			r, w := utf8.DecodeRuneInString(c)
			if len(c) > w {
				return fmt.Errorf("only single charactered separators are allowed: %+q", c)
			}
			separators = append(separators, r)
		}
	}
	br := bufio.NewReader(os.Stdin)
	for _, p := range patterns {
		g, err := glob.Compile(p, separators...)
		if err != nil {
			return fmt.Errorf("could not compile pattern %+q: %w", p, err)
		}
		s := match.Graphviz(p, g.(match.Matcher))
		if auto {
			fmt.Fprintf(os.Stdout, "pattern: %+q: ", p)
			if err := open(s); err != nil {
				fmt.Printf("could not open graphviz: %v", err)
				os.Exit(1)
			}
			if !next(br) {
				return nil
			}
		} else {
			fmt.Fprintln(os.Stdout, s)
		}
	}
	return nil
}

func open(s string) error {
	file, err := os.Create("glob.graphviz.png")
	if err != nil {
		return err
	}
	defer file.Close()
	cmd := exec.Command("dot", "-Tpng")
	cmd.Stdin = strings.NewReader(s)
	cmd.Stdout = file
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	cmd = exec.Command("open", file.Name())
	return cmd.Run()
}

func next(in *bufio.Reader) bool {
	fmt.Fprint(os.Stdout, "cancel? [Y/n]: ")
	p, err := in.ReadBytes('\n')
	if err != nil {
		return false
	}
	if p[0] == 'Y' {
		return false
	}
	return true
}
