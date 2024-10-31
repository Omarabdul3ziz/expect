package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Job struct {
	cmd          string
	args         []string
	expectations map[string]string // prompt:answer
}

func runInterpreter() error {
	if len(os.Args) != 2 {
		return errors.New("require a single expect file as an argument")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", os.Args[1], err)
	}

	jobs, err := parse(bufio.NewScanner(file))
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := handle(job); err != nil {
			return err
		}
	}

	return nil
}

func parse(scanner *bufio.Scanner) ([]Job, error) {
	var jobs []Job

	emptyJob := Job{
		expectations: make(map[string]string),
	}

	job := emptyJob
	crtPrompt := ""

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		// TODO: check for syntax errors? invalid line
		if len(line) < 2 {
			continue
		}

		switch line[0] {
		case "spawn":
			job.cmd = line[1]
			if len(line) > 2 {
				job.args = line[2:]
			}

		case "expect":
			if line[1] == "eof" {
				// finish the current job
				jobs = append(jobs, job)

				// clear for the next job
				job = emptyJob
				crtPrompt = ""
				continue
			}

			key := buildString(line[1:])
			job.expectations[key] = ""
			crtPrompt = key

		case "send":
			// assign answer to a prompt
			job.expectations[crtPrompt] = buildString(line[1:])
		}
	}

	return jobs, nil
}

func handle(job Job) error {
	keyChan := make(chan string)

	cmd := exec.Command(job.cmd, job.args...)
	cmd.Stderr = os.Stderr // TODO: should be handled?

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	defer cmd.Wait()

	go startReader(out, keyChan)                                       // from the output
	if err := startWriter(in, keyChan, job.expectations); err != nil { // to the input
		return err
	}

	return nil
}

func startReader(out io.ReadCloser, keyChan chan string) {
	key := make([]byte, 1024) // is that a good limit?

	for {
		n, err := out.Read(key)
		if err != nil {
			close(keyChan)
			if err == io.EOF { // ?
				// log.Println("Done")
				break
			}
			log.Println("Error: failed to read")
			continue
		}

		// removing the empty bytes and the \n
		// TODO: find a better way
		keyChan <- string(key[:n-1])
	}

}

func startWriter(in io.WriteCloser, keyChan chan string, conv map[string]string) error {
	for key := range keyChan {
		answer := getAnswer(conv, key)
		// log.Printf("prompt: %q, answer: %q", key, answer)

		if _, err := in.Write([]byte(answer)); err != nil {
			return fmt.Errorf("failed to write %q: %w", answer, err)
		}
	}

	return nil
}

func buildString(args []string) string {
	return strings.Join(args, " ")
}

func getAnswer(conv map[string]string, got string) string {
	// TODO: find a better way to match (maybe: regex)
	for prompt, answer := range conv {
		// this just to remove the wrong \"
		prompt = strings.ReplaceAll(prompt, "\"", "")
		if prompt == got {
			answer = strings.ReplaceAll(answer, "\"", "")

			return answer + "\n" // new line to send
		}
	}

	return ""
}

func main() {
	if err := runInterpreter(); err != nil {
		log.Fatal(err)
	}
}
