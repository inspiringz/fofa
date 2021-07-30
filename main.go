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
		} else if err == option.PrintGrammar {
			logger.FofaGrammar()
		} else if err == option.IconHash {
			fetch.GetIconHash(err.Error())
		} else if err == option.FofaTip {
			fetch.GetFofaTip(err.Error())
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
	logger.Success(fmt.Sprintf("Current Output Path: %v", output))

	clt := fetch.NewFofaClient(email, apiKey, size)

	vaild, err := clt.Auth()
	if vaild != true {
		if err != nil {
			logger.Warn(err.Error())
		} else {
			logger.Warn("not vaild!")
		}
		return
	} else {
		logger.Success("Account(email & key) verified successfully.")
	}

	clt.QueryAllT(querys)

	report.WriteXlsx(fetch.FetchResultT.M, output)

	return
}
