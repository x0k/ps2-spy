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
	resource     string
	query        string
	outputFolder string
)

func init() {
	flag.StringVar(&resource, "resource", "", "resource")
	flag.StringVar(&query, "query", "", "query")
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
					Equals(census2.Str(strings.ToLower(tag))),
			).
			WithJoin(
				census2.Join(collections.CharactersWorld).
					On("leader_character_id").
					To("character_id").
					InjectAt("characters_world"),
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

func loadOutfits(c *census2.Client, tagsStr string) (any, error) {
	const op = "loadOutfits"
	tags := stringsx.SplitAndTrim(strings.ToLower(tagsStr), ",")
	q := census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Outfit).
		Where(
			census2.Cond("alias_lower").Equals(census2.StrList(tags...)),
		).
		SetLimit(len(tags))
	log.Printf("%s run query: %s", op, q.String())
	outfits, err := c.Execute(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return outfits, nil
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

func loadOutfitMembers(c *census2.Client, outfitTag string) (any, error) {
	const op = "loadOutfitMembers"
	q := census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Outfit).
		Where(
			census2.Cond("alias_lower").Equals(census2.Str(strings.ToLower(outfitTag))),
		).
		Show("outfit_id").
		WithJoin(
			census2.Join(collections.OutfitMember).
				Show("character_id").
				InjectAt("members").
				IsList(true),
		)
	log.Printf("run query: %s", q.String())
	outfits, err := c.Execute(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(outfits) == 0 {
		return nil, fmt.Errorf("%s: outfit %q not found", op, outfitTag)
	}
	return outfits[0], nil
}

func loadWorldState(c *census2.Client, query string) (any, error) {
	const op = "loadWorldState"
	q := census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Map).
		Where(
			census2.Cond("world_id").Equals(census2.Str(query)),
			census2.Cond("zone_ids").Equals(census2.Str("2,4,6,8,14,344")),
		).
		WithJoin(
			census2.Join(collections.MapRegion).
				Show(
					"zone_id",
					"facility_id",
					"facility_name",
					"facility_type",
				).
				InjectAt("map_region").
				On("Regions.Row.RowData.RegionId").
				To("map_region_id"),
		)
	log.Printf("run query: %s", q.String())
	events, err := c.Execute(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return events, nil
}

var handlers = map[string]func(c *census2.Client, query string) (any, error){
	"outfit":     loadOutfitInfo,
	"outfits":    loadOutfits,
	"character":  loadCharacterInfo,
	"characters": loadCharacters,
	"members":    loadOutfitMembers,
	"world":      loadWorldState,
}

func main() {
	httpClient := &http.Client{}
	censusClient := census2.NewClient("https://census.daybreakgames.com", "", httpClient)
	handler, ok := handlers[resource]
	if !ok {
		log.Fatalf("unknown resource: %s", resource)
	}
	data, err := handler(censusClient, query)
	if err != nil {
		log.Fatalf("failed to handle flags: %s", err)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	outputFileName := fmt.Sprintf("%s.yaml", resource)
	outputPath := path.Join(outputFolder, outputFileName)
	err = os.WriteFile(outputPath, out, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("wrote %s", outputPath)
}
