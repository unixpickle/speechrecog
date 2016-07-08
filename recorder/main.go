package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/unixpickle/speechrecog/speechdata"
)

func main() {
	var portNum int
	flag.IntVar(&portNum, "port", 80, "HTTP port number")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: recorder [flags] data_dir\n\nAvailable flags:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	indexPath := flag.Args()[0]

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		if err := os.Mkdir(indexPath, speechdata.IndexPerms); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to make data dir:", err)
			os.Exit(1)
		}
	}

	index, err := speechdata.LoadIndex(indexPath)
	if err != nil {
		index := &speechdata.Index{
			DirPath: indexPath,
		}
		err = index.Save()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create DB:", err)
		os.Exit(1)
	}

	http.ListenAndServe(":"+strconv.Itoa(portNum), &Server{
		Index: index,
	})
}
