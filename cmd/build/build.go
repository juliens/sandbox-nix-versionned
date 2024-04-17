package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/juliens/sandbox-nix-versionned/pkg/foo"
	"github.com/juliens/sandbox-nix-versionned/pkg/nix"
)

var count int

type out struct {
	commit   string
	err      error
	packages map[string]nix.Pkg
	duration time.Duration
}

var previous *foo.Foo

func main() {
	var err error
	previous, err = foo.New("./pkg/foo/all.json")
	if err != nil {
		log.Fatal(err)
	}

	inCh := make(chan string, 100000)
	outCh := make(chan out)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)

	go feed(ctx, inCh)

	endCh := make(chan struct{})
	go func() {
		var handled int
		var average time.Duration
		var buf int
		for val := range outCh {
			buf++
			handled++
			log.Printf("%d handled / %d sent\n", handled, count)
			if errors.Is(val.err, context.Canceled) {
				continue
			}

			if val.err == nil && val.packages == nil {
				log.Println("Ignore because empty")
				continue
			}

			average = time.Duration((int(average)*(handled-1) + int(val.duration)) / handled)

			log.Printf("commit done in %s", val.duration)
			log.Printf("Approx. remaining time: %s\n", time.Duration(count)*average/5)

			previous.AddCommit(val.commit, val.err)
			previous.Merge(val.packages)

			if buf > 10 {
				err = previous.Write()
				buf = 0
			}
			if err != nil {
				log.Printf("Error while writing file: %v", err)
			}
		}

		close(endCh)
	}()

	startWorkers(ctx, 5, inCh, outCh)
	close(outCh)
	<-endCh
}

func startWorkers(ctx context.Context, nb int, inCh chan string, outCh chan out) {
	waitGroup := sync.WaitGroup{}
	for range nb {
		waitGroup.Add(1)
		go func() {
			parser(ctx, inCh, outCh)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}

func parser(ctx context.Context, inCh chan string, outCh chan out) {
	n := nix.NewNix("/Users/juliensalleyron/dev/nixpkgs")

	for commit := range inCh {
		if previous.ContainsCommit(commit) {
			log.Println("Ignore, already exists")
			continue
		}
		start := time.Now()
		packages, err := n.Packages(ctx, commit)
		outCh <- out{
			packages: packages,
			err:      err,
			commit:   commit,
			duration: time.Now().Sub(start),
		}
	}
}

func feed(ctx context.Context, inCh chan string) {
	defer close(inCh)

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			inCh <- scanner.Text()
			count++
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

	} else {
		for _, s := range os.Args[1:len(os.Args)] {
			select {
			case <-ctx.Done():
				return
			default:
			}
			inCh <- strings.TrimSpace(s)
			count++
		}
	}
}
