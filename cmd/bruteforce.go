/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	proto "github.com/ta04/auth-service/model/proto"
	"github.com/ta04/brute-force-client/client"
)

// bruteforceCmd represents the bruteforce command
var bruteforceCmd = &cobra.Command{
	Use:   "bruteforce",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		bruteforce(args)
	},
}

func init() {
	rootCmd.AddCommand(bruteforceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bruteforceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bruteforceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type resultStruct struct {
	Result int64 `json:"result"`
}

func bruteforce(args []string) {
	client := client.NewAuthSC()

	username := args[0]
	generatorValue := args[1]
	y := args[2]
	primeNumber, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		log.Println(err)
	}

	rand.Seed(time.Now().UnixNano())
	min := 1
	max := int(primeNumber)
	v := rand.Intn(max-min+1) + min

	// generate possible Xs
	var possibleXs []string
	for i := 0; i < 4096; i++ {
		calculateYURL := fmt.Sprintf("http://localhost:5000/calculateY?g=%s&x=%d&n=%d",
			generatorValue, i, primeNumber)
		calculateYRes, err := http.Get(calculateYURL)
		if err != nil {
			log.Println(err)
		}

		defer calculateYRes.Body.Close()

		body, err := ioutil.ReadAll(calculateYRes.Body)
		if err != nil {
			log.Println(err)
		}

		var unmarshalledBody resultStruct
		err = json.Unmarshal(body, &unmarshalledBody)
		if err != nil {
			log.Println(err)
		}

		calculatedY := strconv.FormatInt(unmarshalledBody.Result, 10)
		if y == calculatedY {
			possibleXs = append(possibleXs, strconv.FormatInt(int64(i), 10))
		}
	}

	log.Println("here are some possible Xs: ", possibleXs)

	// calculate t
	calculateTURL := fmt.Sprintf("http://localhost:5000/calculateT?g=%s&v=%d&n=%d",
		generatorValue, v, primeNumber)
	calculateTRes, err := http.Get(calculateTURL)
	if err != nil {
		log.Println(err)
	}

	defer calculateTRes.Body.Close()

	body, err := ioutil.ReadAll(calculateTRes.Body)
	if err != nil {
		log.Println(err)
	}

	var unmarshalledBody resultStruct
	err = json.Unmarshal(body, &unmarshalledBody)
	if err != nil {
		log.Println(err)
	}

	t := strconv.FormatInt(unmarshalledBody.Result, 10)

	// Call AuthRPC1 from auth client
	auth1Res, err := client.AuthRPC1(context.Background(), &proto.Auth1{Username: username, T: t})
	if err != nil {
		log.Println("failed to authenticate in auth1. err: ", err)
	}

	c, err := strconv.ParseInt(auth1Res.C, 10, 64)
	if err != nil {
		log.Println(err)
	}

	for _, possibleX := range possibleXs {
		x, err := strconv.ParseInt(possibleX, 10, 64)
		if err != nil {
			log.Println(err)
		}

		r := strconv.FormatInt(int64(v-int(c)*int(x)), 10)

		// Call AuthRPC2 from auth client
		log.Println("trying to login with x = ", x)
		auth2Res, err := client.AuthRPC2(context.Background(), &proto.Auth2{Username: username, R: r})
		if err != nil {
			log.Println("failed to authenticate in auth2. err: ", err)
		}

		if auth2Res != nil && auth2Res.Token != "" {
			token := auth2Res.Token
			log.Println("logged in successfully. token: ", token)

			break
		}
	}
}
