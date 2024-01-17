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
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"gopkg.in/yaml.v3"
)

var (
	outfitTag      string
	characterName  string
	characterNames string
	outputFolder   string
)

func init() {
	flag.StringVar(&outfitTag, "tag", "", "outfit tag")
	flag.StringVar(&characterName, "character", "", "character name")
	flag.StringVar(&characterNames, "characters", "", "character names")
	flag.StringVar(&outputFolder, "output", "", "output folder")
	flag.Parse()
}

func loadOutfitInfo(c *census2.Client, tag string) (any, error) {
	const op = "loadOutfitInfo"
	outfits, err := c.Execute(
		context.Background(),
		census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Outfit).
			Where(
				census2.Cond("alias_lower").
					Equals(census2.Str(strings.ToLower(outfitTag))),
			).
			Resolve("member_character"),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(outfits) == 0 {
		return nil, fmt.Errorf("%s: outfit %q not found", op, tag)
	}
	return outfits[0], nil
}

func loadCharacterInfo(c *census2.Client, name string) (any, error) {
	const op = "loadCharacterInfo"
	q := census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Character).
		Where(
			census2.Cond("name.first_lower").
				Equals(census2.Str(strings.ToLower(name))),
		).
		Resolve("outfit", "world")
	log.Printf("%s run query: %s", op, q.String())
	characters, err := c.Execute(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(characters) == 0 {
		return nil, fmt.Errorf("%s: character %q not found", op, name)
	}
	return characters[0], nil
}

func loadCharacters(c *census2.Client, namesStr string) (any, error) {
	const op = "loadCharacters"
	names := stringsx.SplitAndTrim(strings.ToLower(namesStr), ",")
	q := census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Character).
		Where(
			census2.Cond("name.first_lower").
				Equals(census2.StrList(names...)),
		).
		SetLimit(len(names))
	log.Printf("%s run query: %s", op, q.String())
	characters, err := c.Execute(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return characters, nil
}

func handleFlags(c *census2.Client) (string, any, error) {
	if outfitTag != "" {
		info, err := loadOutfitInfo(c, outfitTag)
		return strings.ToLower(outfitTag), info, err
	}
	if characterName != "" {
		info, err := loadCharacterInfo(c, characterName)
		return strings.ToLower(characterName), info, err
	}
	if characterNames != "" {
		info, err := loadCharacters(c, characterNames)
		return "characters", info, err
	}
	return "", nil, fmt.Errorf("Invalid flags combination")
}

func main() {
	httpClient := &http.Client{}
	censusClient := census2.NewClient("https://census.daybreakgames.com", "", httpClient)
	filename, data, err := handleFlags(censusClient)
	if err != nil {
		log.Fatalf("failed to handle flags: %s", err)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	outputFileName := fmt.Sprintf("%s.yaml", filename)
	outputPath := path.Join(outputFolder, outputFileName)
	err = os.WriteFile(outputPath, out, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("wrote %s", outputPath)
}
