package main

import (
	"encoding/json"
	"github.com/kungfusoftware/externalbuilder/pkg/utils"
	"io/ioutil"
	"os"
	"path"
)

func main() {
	args := os.Args[1:]
	chaincodeMetadataDir := args[1]

	metadataPath := path.Join(chaincodeMetadataDir, "metadata.json")
	metadataBytes, err := ioutil.ReadFile(metadataPath)
	utils.HandleErr(err, "reading metadata file")

	// unmarshall json
	metadata := map[string]interface{}{}
	err = json.Unmarshal(metadataBytes, &metadata)
	utils.HandleErr(err, "parsing metadata json")

	// check for chaincode type
	chaincodeType, ok := metadata["type"].(string)
	if ok && chaincodeType == "external" {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
