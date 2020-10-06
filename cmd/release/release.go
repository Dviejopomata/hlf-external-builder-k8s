package main

import (
	"fmt"
	"github.com/kungfusoftware/externalbuilder/pkg/utils"
	"os"
	"path"
)

func main() {
	args := os.Args[1:]
	buildOutpurDir := args[0]
	releaseOutputDir := args[1]
	metadataPath := path.Join(buildOutpurDir, "metadata")
	if utils.Exists(metadataPath) {
		metadataDestPath := path.Join(releaseOutputDir, "metadata")
		err := utils.Copy(metadataPath, metadataDestPath)
		utils.HandleErr(err, fmt.Sprintf("failed to copy metadata directory from %s to %s", metadataPath, metadataDestPath))
	}
	// copy code
	chaincodeServerPath := path.Join(releaseOutputDir, "chaincode", "server")
	err := os.MkdirAll(chaincodeServerPath, 777)
	utils.HandleErr(err, fmt.Sprintf("failed to create chaincode server directory %s", chaincodeServerPath))

	// copy connection.json
	connectionJsonPath := path.Join(buildOutpurDir, "connection.json")
	connectionDestJsonPath := path.Join(chaincodeServerPath, "connection.json")
	err = utils.Copy(connectionJsonPath, connectionDestJsonPath)
	utils.HandleErr(
		err,
		fmt.Sprintf("failed to copy connection.json file from %s to %s", connectionJsonPath, connectionDestJsonPath),
	)
	// TODO: check if tls required to copy files

}
