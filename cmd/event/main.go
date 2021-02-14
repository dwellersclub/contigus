package main

import (
	logging "github.com/dwellersclub/contigus/log"
	_ "github.com/lib/pq"
)

var log = logging.GetLogger()
var (
	version    string
	buildstamp string
	githash    string
)

func main() {}
