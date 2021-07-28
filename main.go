package main

import (
	"fmt"
	"fofa/fetch"
	"fofa/logger"
	"fofa/option"
	"fofa/report"
	"fofa/utils"
	"os"
)

func main() {
	logger.InitPlatform()
	logger.AsciiBanner()

	email, apiKey, query, ruleFile, size, output, err := option.ParseCli(os.Args[1:])
	if err != nil {
		if err == option.PrintUsage {
			logger.Usage()
		} else {
			logger.Warn(err.Error())
		}
		return
	}

	logger.Success(fmt.Sprintf("Current Email: %v", email))
	logger.Success(fmt.Sprintf("Current Email: %v", apiKey))

	var querys []string

	if query != "" {
		logger.Success(fmt.Sprintf("Current Query: %v", query))
		querys = append(querys, query)
	} else {
		logger.Success(fmt.Sprintf("Current Rules File: %v", ruleFile))
		querys = utils.ScanFile(ruleFile)
	}

	logger.Success(fmt.Sprintf("Current Fetch Size: %v", size))

	clt := fetch.NewFofaClient(email, apiKey, size)

	vaild, err := clt.Auth()
	if vaild != true {
		logger.Warn(err.Error())
		return
	} else {
		logger.Success("Account(email & key) verified successfully.")
	}

	clt.QueryAllT(querys)

	report.WriteXlsx(fetch.FetchResultT.M, output)

	return
}
