package util

import (
	"fmt"
	"strconv"
	"strings"
)

const banner = `   ____               __           
  / __/__  ___  ___  / /_ __ _ ___ 
 _\ \/ _ \/ _ '/ _ \/  '_/ // / _ \
/___/_//_/\_,_/ .__/_/\_\\_,_/ .__/
             /_/############/_/
`

const toReplace = "############"

func PrintBanner(version string) {
	rsLen := len(toReplace)
	verLen := len(version)
	padLeft := (rsLen - verLen) / 2
	padRight := rsLen - verLen - padLeft
	paddedLeft := fmt.Sprintf("%"+strconv.Itoa(padLeft+verLen)+"v", version)
	padded := fmt.Sprintf("%-"+strconv.Itoa(padRight+len(paddedLeft))+"v", paddedLeft)
	println(strings.Replace(banner, toReplace, padded, 1))
}
