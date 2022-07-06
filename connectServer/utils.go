package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/xx_network/primitives/utils"
	"io/ioutil"
	"log"
	"os"
)

func initLog(threshold uint, logPath string) {
	if logPath != "-" && logPath != "" {
		// Disable stdout output
		jww.SetStdoutOutput(ioutil.Discard)
		// Use log file
		logOutput, err := os.OpenFile(logPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err.Error())
		}
		jww.SetLogOutput(logOutput)
	}

	if threshold > 1 {
		jww.INFO.Printf("log level set to: TRACE")
		jww.SetStdoutThreshold(jww.LevelTrace)
		jww.SetLogThreshold(jww.LevelTrace)
		jww.SetFlags(log.LstdFlags | log.Lmicroseconds)
	} else if threshold == 1 {
		jww.INFO.Printf("log level set to: DEBUG")
		jww.SetStdoutThreshold(jww.LevelDebug)
		jww.SetLogThreshold(jww.LevelDebug)
		jww.SetFlags(log.LstdFlags | log.Lmicroseconds)
	} else {
		jww.INFO.Printf("log level set to: INFO")
		jww.SetStdoutThreshold(jww.LevelInfo)
		jww.SetLogThreshold(jww.LevelInfo)
	}
}

func writeContact(outfilePath string, c contact.Contact) {
	err := utils.WriteFileDef(outfilePath, c.Marshal())

	if err != nil {
		jww.ERROR.Printf("could not write contact file: %+v", err)
	} else {
		jww.INFO.Printf("contact written to %s successfully", outfilePath)
	}
}
