package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/andig/sensonet/sensonet"
	"github.com/evcc-io/evcc/util"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
)

const TOKEN_FILE = ".sensonet-token.json"

func readToken(filename string) (*oauth2.Token, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	err = json.Unmarshal(b, &token)

	return &token, err
}

func writeToken(filename string, token *oauth2.Token) error {
	b, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, b, 0o644)
}

func main() {
	// util.LogLevel("trace", nil)
	logger := util.NewLogger("sensonet")

	identity, err := sensonet.NewIdentity(logger, sensonet.REALM_GERMANY)
	if err != nil {
		log.Fatal(err)
	}

	var ts oauth2.TokenSource

	token, err := readToken(TOKEN_FILE)
	if err == nil {
		ts, err = identity.TokenSource(token)

		if err == nil {
			// save token in case of refresh
			if tok, err := ts.Token(); err == nil && tok.Valid() && (tok.AccessToken != token.AccessToken) {
				_ = writeToken(TOKEN_FILE, tok)
			}
		}
	}

	if err != nil {
		user := os.Getenv("SENSONET_USER")
		password := os.Getenv("SENSONET_PASSWORD")

		token, err = identity.Login(user, password)
		if err != nil {
			log.Fatal(err)
		}

		ts, err = identity.TokenSource(token)
		if err != nil {
			log.Fatal(err)
		}

		if err := writeToken(TOKEN_FILE, token); err != nil {
			log.Fatal(err)
		}
	}

	conn, err := sensonet.NewConnection(logger, ts)
	if err != nil {
		log.Fatal(err)
	}

	state, err := conn.GetSystem()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("OutdoorTemperature: %.1f°C\n", state.State.System.OutdoorTemperature)
	for i, c := range state.State.Circuits {
		fmt.Printf("Zone %s: %.1f°C (%.1f°C)\n", state.Configuration.Zones[i].General.Name, c.CurrentCircuitFlowTemperature, c.HeatingCircuitFlowSetpoint)
	}
	if len(state.State.DomesticHotWater)*len(state.Configuration.DomesticHotWater) > 0 {
		fmt.Printf("HotWaterTemperature: %.1f°C (%.1f°C)\n", state.State.DomesticHotWater[0].CurrentDomesticHotWaterTemperature, state.Configuration.DomesticHotWater[0].TappingSetpoint)
	}
}
