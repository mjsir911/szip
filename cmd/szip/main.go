package main

import (
	"flag"
	"os"
	"code.sirabella.org/szip"
	"code.sirabella.org/szip/tar"
)


func main() {
	r := szip.NewReader(os.Stdin);

	var (
		outputDir string
		permissionless bool
		extractTar bool
	)
	flag.StringVar(&outputDir, "d", ".", "output directory to extract")
	flag.BoolVar(&permissionless, "p", false, "Ignore permissions in central directory header? (faster)")
	flag.BoolVar(&extractTar, "t", false, "Translate to tar file on stdout")

	flag.Parse()

	if extractTar {
		if err := szip2tar.Write(os.Stdout, r); err != nil {
			panic(err)
		}
	} else {
		if err := extractZip(outputDir, &r); err != nil {
			panic(err)
		}
		if ! permissionless {
			files, err := r.CentralDirectory()
			if err != nil {
				panic(err)
			}
			if err := extractPermissions(outputDir, files); err != nil {
				panic(err)
			}
		}
	}
}
