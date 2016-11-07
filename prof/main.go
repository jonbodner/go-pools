package main

import (
	"bytes"
	"flag"

	"github.com/jdamick/go-pools/buffers"
	"github.com/jdamick/go-pools/buffers2"
	"github.com/pkg/profile"
)

func main() {
	//mode := flag.String("profile.mode", "", "enable profiling mode, one of [cpu, mem, block]")
	function := flag.String("profile.func", "", "enable profiling mode, one of [norm, fixed]")

	flag.Parse()
	switch *function {
	case "norm":
		normalBuffs(profile.BlockProfile)
		normalBuffs(profile.MemProfile)
		normalBuffs(profile.CPUProfile)
	case "fixed":
		fixedPoolsBuffs(profile.BlockProfile)
		fixedPoolsBuffs(profile.MemProfile)
		fixedPoolsBuffs(profile.CPUProfile)
	case "fixed2":
		fixedPoolsBuffs2(profile.BlockProfile)
		fixedPoolsBuffs2(profile.MemProfile)
		fixedPoolsBuffs2(profile.CPUProfile)
	}
}

const (
	maxiters          = 10000000
	maxBytesPerBuffer = 4096
)

func normalBuffs(prof func(*profile.Profile)) {
	defer profile.Start(prof, profile.ProfilePath("./prof.norm")).Stop()
	for i := 0; i < maxiters; i++ {
		b := bytes.NewBuffer(make([]byte, maxBytesPerBuffer))
		if b == nil {
			b.WriteString("a")
		}
	}
}

func fixedPoolsBuffs(prof func(*profile.Profile)) {
	defer profile.Start(prof, profile.ProfilePath("./prof.fixed")).Stop()
	p := buffers.NewFixedBufferPool(10, maxBytesPerBuffer)
	for i := 0; i < maxiters; i++ {
		b := p.Get()
		if b != nil {
			b.WriteString("a")
			p.Put(b)
		}
	}
}

func fixedPoolsBuffs2(prof func(*profile.Profile)) {
	defer profile.Start(prof, profile.ProfilePath("./prof.fixed2")).Stop()
	p := buffers2.NewFixedBufferPool(10, maxBytesPerBuffer)
	for i := 0; i < maxiters; i++ {
		b := p.Get()
		if b != nil {
			b.WriteString("a")
			p.Put(b)
		}
	}
}
