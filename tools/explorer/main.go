package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"gopkg.in/yaml.v3"
)

var (
	outfitTag    string
	outputFolder string
)

func init() {
	flag.StringVar(&outfitTag, "tag", "", "outfit tag")
	flag.StringVar(&outputFolder, "output", "", "output folder")
	flag.Parse()
}

func main() {
	httpClient := &http.Client{}
	censusClient := census2.NewClient("https://census.daybreakgames.com", "", httpClient)
	outfits, err := censusClient.Execute(
		context.Background(),
		census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Outfit).
			Where(
				census2.Cond("alias_lower").
					Equals(census2.Str(strings.ToLower(outfitTag))),
			).
			Resolve("member_character"),
	)
	if err != nil {
		log.Fatalln(err)
	}
	if len(outfits) == 0 {
		log.Fatalf("outfit %q not found", outfitTag)
	}
	out, err := yaml.Marshal(outfits[0])
	if err != nil {
		log.Fatalln(err)
	}
	outputFileName := fmt.Sprintf("%s.yaml", strings.ToLower(outfitTag))
	outputPath := path.Join(outputFolder, outputFileName)
	err = os.WriteFile(outputPath, out, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("wrote %s", outputPath)
}
