// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/hoisie/web"
)

var (
	cfg      = ReadConfig().Explorer
	server   = web.NewServer()
	tpl      = new(template.Template)
	ExtIDMap map[string]bool
)

func main() {
	var (
		err error
		dir string
	)

	server.Config.StaticDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.StaticDir != "" {
		server.Config.StaticDir = cfg.StaticDir
		dir = cfg.StaticDir
	}

	tpl = template.Must(template.New("main").Funcs(template.FuncMap{
		"hashfilter": hashfilter,
		"hextotext": hextotext,
	}).ParseFiles(
		dir+"/views/404.html",
		dir+"/views/chain.html",
		dir+"/views/chains.html",
		dir+"/views/cheader.html",
		dir+"/views/dblock.html",
		dir+"/views/eblock.html",
		dir+"/views/entries.html",
		dir+"/views/header.html",
		dir+"/views/index.html",
		dir+"/views/sentry.html",
	))

	server.Get(`/(?:home)?`, handleHome)
	server.Get(`/`, handleDBlocks)
	server.Get(`/index.html`, handleDBlocks)
	server.Get(`/chains/?`, handleChains)
	server.Get(`/chain/([^/]+)?`, handleChain)
	server.Get(`/dblocks/?`, handleDBlocks)
	server.Get(`/dblock/([^/]+)?`, handleDBlock)
	server.Get(`/eblock/([^/]+)?`, handleEBlock)
	server.Get(`/entry/([^/]+)?`, handleEntry)
	server.Get(`/sentry/([^/]+)?`, handleEntry)
	server.Post(`/search/?`, handleSearch)
	server.Get(`/.*`, handle404)

	server.Run(fmt.Sprintf(":%d", cfg.PortNumber))
}

func handle404(ctx *web.Context) {
	var c interface{}
	tpl.ExecuteTemplate(ctx, "404.html", c)
}

func handleSearch(ctx *web.Context) {
	fmt.Println("r.Form:", ctx.Params["searchType"])
	fmt.Println("r.Form:", ctx.Params["searchText"])

	//	pagesize := 1000
	//	hashArray := make([]*notaryapi.Hash, 0, 5)
	searchText := strings.ToLower(strings.TrimSpace(ctx.Params["searchText"]))

	switch searchType := ctx.Params["searchType"]; searchType {
	case "entry":
		handleEntry(ctx, searchText)

	case "eblock":
		handleEBlock(ctx, searchText)

	case "dblock":
		handleDBlock(ctx, searchText)
	case "extID":
		handleEntryEid(ctx, searchText)
	default:
	}
}

func handleChain(ctx *web.Context, hash string) {
	chain, err := factom.GetChain(hash)
	if err != nil {
		log.Println(err)
		handle404(ctx)
		return
	}

	tpl.ExecuteTemplate(ctx, "chain.html", chain)
}

func handleChains(ctx *web.Context) {
	chains, err := factom.GetChains()
	if err != nil {
		log.Println(err)
	}
	
	tpl.ExecuteTemplate(ctx, "chains.html", chains)
}

func handleDBlock(ctx *web.Context, hash string) {
	type fullblock struct {
		dblock *factom.DBlock
		dbinfo *factom.DBInfo
	}
	
	var b fullblock
	
	b.dblock, err := factom.GetDBlock(hash)
	if err != nil {
		log.Println(err)
		handle404(ctx)
		return
	}
	b.dbinfo, err := factom.GetDBInfo(hash)
	if err != nil {
		log.Println(err)
	}

	tpl.ExecuteTemplate(ctx, "dblock.html", b)
}

func handleDBlocks(ctx *web.Context) {
	height, err := factom.GetBlockHeight()
	if err != nil {
		log.Println(err)
	}
	dBlocks, err := factom.GetDBlocks(0, height)
	if err != nil {
		log.Println(err)
	}
	if dBlocks == nil {
		log.Println("dBlocks is nil")
	}

	tpl.ExecuteTemplate(ctx, "index.html", dBlocks)
}

func handleEBlock(ctx *web.Context, mr string) {
	type eblockPlus struct {
		EBlock *factom.EBlock
		Hash   string
		Count  int
	}
	
	eblock, err := factom.GetEBlock(mr)
	if err != nil {
		log.Println(err)
		handle404(ctx)
		return
	}
	
	e := eblockPlus{
		EBlock: eblock,
		Hash:   mr,
		Count:  len(eblock.EBEntries),
	}

	tpl.ExecuteTemplate(ctx, "eblock.html", e)
}

func handleEntry(ctx *web.Context, hash string) {
	entry, err := factom.GetEntry(hash)
	if err != nil {
		log.Println(err)
		handle404(ctx)
		return
	}

	tpl.ExecuteTemplate(ctx, "sentry.html", entry)
}

func handleEntryEid(ctx *web.Context, eid string) {
	entries, err := factom.GetEntriesByExtID(eid)
	if err != nil {
		log.Println(err)
		handle404(ctx)
		return
	}

	tpl.ExecuteTemplate(ctx, "entries.html", entries)
}

func handleHome(ctx *web.Context) {
	handleDBlocks(ctx)
}

func hextotext(h string) string {
	p, err := hex.DecodeString(h)
	if err != nil {
		log.Println(err)
	}
	return string(p)
}

func hashfilter(s string) string {
	var filter = []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}
	
	for _, v := range filter {
		if s == v {
			return "None"
		}
	}
	
	return s
}