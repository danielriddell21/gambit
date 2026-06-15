package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func parseFlags() config {
	c := config{}
	flag.StringVar(&c.white, "white", "minimax", "white agent strategy")
	flag.StringVar(&c.black, "black", "random", "black agent strategy")
	flag.IntVar(&c.depth, "depth", 4, "search depth for depth-limited agents")
	flag.IntVar(&c.iterations, "iterations", 20000, "playout budget for the mcts agent")
	flag.Int64Var(&c.seed, "seed", time.Now().UnixNano(), "RNG seed for stochastic agents")
	flag.StringVar(&c.fen, "fen", "", "starting position FEN (default: standard start)")
	flag.DurationVar(&c.delay, "delay", 400*time.Millisecond, "pause between moves in the GUI")
	flag.StringVar(&c.record, "record", "", "record the game to this GIF path, then exit (GUI only)")
	flag.IntVar(&c.square, "square", 80, "board square size in pixels (GUI only)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("gambit", version)
		os.Exit(0)
	}
	return c
}
