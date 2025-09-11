package relations

import (
	"encoding/json"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func PrintPipeline(pipeline []bson.M) {
	println("**************************************************")
	for _, stage := range pipeline {
		jsonData, err := json.MarshalIndent(stage, "", "  ")
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "relations: Failed to marshal json"),
			}
			fmt.Println(err.Error())
			continue
		}

		fmt.Printf("%s\n", string(jsonData))
	}
	println("**************************************************")
}

func PrintResults(results []bson.M) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "relations: Failed to marshal json"),
		}
		fmt.Println(err.Error())
		return
	}

	println("**************************************************")
	fmt.Printf("%s\n", string(jsonData))
	println("**************************************************")
}
