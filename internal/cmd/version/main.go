// Command version prints the embedded arrest-go version, so the release
// workflows can confirm version.txt made it into the build.
package main

import (
	"fmt"

	"github.com/zostay/arrest-go"
)

func main() {
	fmt.Println(arrest.Version)
}
