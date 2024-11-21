// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package translations

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en": &dictionary{index: enIndex, data: enData},
		"ru": &dictionary{index: ruIndex, data: ruData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"\n\n**Characters:**\n":        59,
	"\n**Left the outfit:**":       9,
	"\n**Other characters:**":      44,
	"\n**Welcome to the outfit:**": 8,
	"\nStatus: _":                  64,
	"# PlanetSide 2 Spy\n\nSimple discord bot for PlanetSide 2 outfits\n\n## Links\n\n- [GitHub](https://github.com/x0k/ps2-spy)\n\t\t\n": 14,
	"%s (%s) is now offline (%s)":                          3,
	"%s (%s) is now online (%s)":                           1,
	"%s - %d":                                              49,
	"%s - %s (Ends %s)":                                    52,
	"%s - No alerts":                                       54,
	"%s [%s] captured %s (%s) on %s (%s)":                  12,
	"%s [%s] lost %s (%s) on %s (%s)":                      13,
	"%s alerts":                                            55,
	") by `":                                               63,
	"Characters online:":                                   42,
	"Enter the character names separated by comma":         37,
	"Enter the outfit tags separated by comma":             35,
	"Failed to load %s population with %s":                 17,
	"Failed to load %s territory control":                  18,
	"Failed to load character names for %v (%s)":           30,
	"Failed to load character: %s":                         10,
	"Failed to load characters %v (%s)":                    24,
	"Failed to load characters: %v (%s)":                   5,
	"Failed to load facility: %s":                          11,
	"Failed to load global alerts from %s":                 20,
	"Failed to load global population with %s":             16,
	"Failed to load online members for %s channel (%s)":    21,
	"Failed to load outfit tags for %v (%s)":               29,
	"Failed to load outfit: %s (%s)":                       4,
	"Failed to load outfits %v (%s)":                       22,
	"Failed to load outfits by tags %v (%s)":               23,
	"Failed to load tracking settings for %s channel (%s)": 28,
	"Failed to load world alerts for %s from %s":           19,
	"Failed to parse language %q":                          38,
	"Failed to save language %q":                           39,
	"Failed to save tracking settings for %s channel (%s)": 25,
	"Invalid population type: %s":                          15,
	"Language for this channel has been set to %q":         40,
	"Locked ":              62,
	"NC":                   46,
	"No":                   69,
	"No alerts":            56,
	"No characters":        60,
	"No characters online": 41,
	"No outfits":           58,
	"Period":               51,
	"Settings are saved, but failed to load character names %v (%s)": 27,
	"Settings are saved, but failed to load outfit tags %v (%s)":     26,
	"Settings are updated.\n\n**Outfits:**\n":                        57,
	"Source: %s":                             48,
	"Stable":                                 65,
	"TR":                                     45,
	"Territory Control":                      53,
	"Total population - %d":                  50,
	"Tracking Settings (PC)":                 31,
	"Tracking Settings (PS4 EU)":             32,
	"Tracking Settings (PS4 US)":             33,
	"Unlocked ":                              61,
	"Unstable":                               66,
	"Update of ":                             6,
	"VS":                                     47,
	"Which characters do you want to track?": 36,
	"Which outfits do you want to track?":    34,
	"Yes":                                    68,
	"[%s] %s (%s) is now offline (%s)":       2,
	"[%s] %s (%s) is now online (%s)":        0,
	"] outfit (":                             43,
	"] outfit members:":                      7,
	"_\nAlerts: _":                           67,
}

var enIndex = []uint32{ // 71 elements
	// Entry 0 - 1F
	0x00000000, 0x0000002c, 0x00000050, 0x0000007d,
	0x000000a2, 0x000000c7, 0x000000f0, 0x000000ff,
	0x00000111, 0x00000131, 0x0000014b, 0x0000016b,
	0x0000018a, 0x000001c0, 0x000001f2, 0x00000270,
	0x0000028f, 0x000002bb, 0x000002e6, 0x0000030d,
	0x0000033e, 0x00000366, 0x0000039e, 0x000003c3,
	0x000003f0, 0x00000418, 0x00000453, 0x00000494,
	0x000004d9, 0x00000514, 0x00000541, 0x00000572,
	// Entry 20 - 3F
	0x00000589, 0x000005a4, 0x000005bf, 0x000005e3,
	0x0000060c, 0x00000633, 0x00000660, 0x0000067f,
	0x0000069d, 0x000006cd, 0x000006e2, 0x000006f5,
	0x00000700, 0x0000071b, 0x0000071e, 0x00000721,
	0x00000724, 0x00000732, 0x00000740, 0x00000759,
	0x00000760, 0x0000077b, 0x0000078d, 0x0000079f,
	0x000007ac, 0x000007b6, 0x000007df, 0x000007ea,
	0x00000801, 0x0000080f, 0x0000081d, 0x00000829,
	// Entry 40 - 5F
	0x00000830, 0x0000083f, 0x00000846, 0x0000084f,
	0x0000085b, 0x0000085f, 0x00000862,
} // Size: 308 bytes

const enData string = "" + // Size: 2146 bytes
	"\x02[%[1]s] %[2]s (%[3]s) is now online (%[4]s)\x02%[1]s (%[2]s) is now " +
	"online (%[3]s)\x02[%[1]s] %[2]s (%[3]s) is now offline (%[4]s)\x02%[1]s " +
	"(%[2]s) is now offline (%[3]s)\x02Failed to load outfit: %[1]s (%[2]s)" +
	"\x02Failed to load characters: %[1]v (%[2]s)\x04\x00\x01 \x0a\x02Update " +
	"of\x02] outfit members:\x04\x01\x0a\x00\x1b\x02**Welcome to the outfit:*" +
	"*\x04\x01\x0a\x00\x15\x02**Left the outfit:**\x02Failed to load characte" +
	"r: %[1]s\x02Failed to load facility: %[1]s\x02%[1]s [%[2]s] captured %[3" +
	"]s (%[4]s) on %[5]s (%[6]s)\x02%[1]s [%[2]s] lost %[3]s (%[4]s) on %[5]s" +
	" (%[6]s)\x04\x00\x04\x0a\x09\x09\x0av\x02# PlanetSide 2 Spy\x0a\x0aSimpl" +
	"e discord bot for PlanetSide 2 outfits\x0a\x0a## Links\x0a\x0a- [GitHub]" +
	"(https://github.com/x0k/ps2-spy)\x02Invalid population type: %[1]s\x02Fa" +
	"iled to load global population with %[1]s\x02Failed to load %[1]s popula" +
	"tion with %[2]s\x02Failed to load %[1]s territory control\x02Failed to l" +
	"oad world alerts for %[1]s from %[2]s\x02Failed to load global alerts fr" +
	"om %[1]s\x02Failed to load online members for %[1]s channel (%[2]s)\x02F" +
	"ailed to load outfits %[1]v (%[2]s)\x02Failed to load outfits by tags %[" +
	"1]v (%[2]s)\x02Failed to load characters %[1]v (%[2]s)\x02Failed to save" +
	" tracking settings for %[1]s channel (%[2]s)\x02Settings are saved, but " +
	"failed to load outfit tags %[1]v (%[2]s)\x02Settings are saved, but fail" +
	"ed to load character names %[1]v (%[2]s)\x02Failed to load tracking sett" +
	"ings for %[1]s channel (%[2]s)\x02Failed to load outfit tags for %[1]v (" +
	"%[2]s)\x02Failed to load character names for %[1]v (%[2]s)\x02Tracking S" +
	"ettings (PC)\x02Tracking Settings (PS4 EU)\x02Tracking Settings (PS4 US)" +
	"\x02Which outfits do you want to track?\x02Enter the outfit tags separat" +
	"ed by comma\x02Which characters do you want to track?\x02Enter the chara" +
	"cter names separated by comma\x02Failed to parse language %[1]q\x02Faile" +
	"d to save language %[1]q\x02Language for this channel has been set to %[" +
	"1]q\x02No characters online\x02Characters online:\x02] outfit (\x04\x01" +
	"\x0a\x00\x16\x02**Other characters:**\x02TR\x02NC\x02VS\x02Source: %[1]s" +
	"\x02%[1]s - %[2]d\x02Total population - %[1]d\x02Period\x02%[1]s - %[2]s" +
	" (Ends %[3]s)\x02Territory Control\x02%[1]s - No alerts\x02%[1]s alerts" +
	"\x02No alerts\x04\x00\x01\x0a$\x02Settings are updated.\x0a\x0a**Outfits" +
	":**\x02No outfits\x04\x02\x0a\x0a\x01\x0a\x10\x02**Characters:**\x02No c" +
	"haracters\x04\x00\x01 \x09\x02Unlocked\x04\x00\x01 \x07\x02Locked\x02) b" +
	"y `\x04\x01\x0a\x00\x0a\x02Status: _\x02Stable\x02Unstable\x02_\x0aAlert" +
	"s: _\x02Yes\x02No"

var ruIndex = []uint32{ // 71 elements
	// Entry 0 - 1F
	0x00000000, 0x0000002d, 0x00000052, 0x00000081,
	0x000000a8, 0x000000e4, 0x00000128, 0x00000142,
	0x0000015f, 0x00000199, 0x000001c1, 0x000001f9,
	0x0000022d, 0x0000026b, 0x000002a7, 0x0000033c,
	0x00000370, 0x000003c0, 0x000003ff, 0x00000450,
	0x00000490, 0x000004d8, 0x00000535, 0x00000572,
	0x000005bf, 0x00000602, 0x00000668, 0x000006e5,
	0x0000076a, 0x000007cc, 0x0000081b, 0x00000872,
	// Entry 20 - 3F
	0x000008a3, 0x000008d8, 0x0000090d, 0x0000094c,
	0x0000098f, 0x000009ce, 0x00000a19, 0x00000a60,
	0x00000aa6, 0x00000af0, 0x00000b1d, 0x00000b40,
	0x00000b51, 0x00000b7b, 0x00000b80, 0x00000b85,
	0x00000b8a, 0x00000ba2, 0x00000bb0, 0x00000be0,
	0x00000bed, 0x00000c14, 0x00000c3a, 0x00000c56,
	0x00000c6b, 0x00000c7f, 0x00000cc2, 0x00000cda,
	0x00000cf9, 0x00000d17, 0x00000d37, 0x00000d55,
	// Entry 40 - 5F
	0x00000d6a, 0x00000d7f, 0x00000d94, 0x00000dae,
	0x00000dc2, 0x00000dc7, 0x00000dce,
} // Size: 308 bytes

const ruData string = "" + // Size: 3534 bytes
	"\x02[%[1]s] %[2]s (%[3]s) онлайн (%[4]s)\x02%[1]s (%[2]s) онлайн (%[3]" +
	"s)\x02[%[1]s] %[2]s (%[3]s) оффлайн (%[4]s)\x02%[1]s (%[2]s) оффлайн (" +
	"%[3]s)\x02Ошибка загрузки аутфита: %[1]s (%[2]s)\x02Ошибка загрузки перс" +
	"онажей: %[1]v (%[2]s)\x04\x00\x01 \x15\x02Обновление\x02] члены аутфит" +
	"а:\x04\x01\x0a\x005\x02**Добро пожаловать в аутфит:**\x04\x01\x0a\x00#" +
	"\x02**Покинули аутфит:**\x02Ошибка загрузки персонажа: %[1]s\x02Ошибка з" +
	"агрузки объекта: %[1]s\x02%[1]s [%[2]s] захватил %[3]s (%[4]s) в %[5]s " +
	"(%[6]s)\x02%[1]s [%[2]s] потерял %[3]s (%[4]s) в %[5]s (%[6]s)\x04\x00" +
	"\x04\x0a\x09\x09\x0a\x8c\x01\x02PlanetSide 2 Spy\x0a\x0aПростой бот для" +
	" PlanetSide 2 аутфитов\x0a\x0a## Ссылки\x0a\x0a- [GitHub](https://github" +
	".com/x0k/ps2-spy)\x02Неверный тип популяции: %[1]s\x02Ошибка загрузки г" +
	"лобальной популяции (%[1]s)\x02Ошибка загрузки %[1]s популяции (%[2]s)" +
	"\x02Ошибка загрузки контроля территорий для %[1]s\x02Ошибка загрузки тре" +
	"вог для %[1]s (%[2]s)\x02Ошибка загрузки глобальных тревог (%[1]s)\x02О" +
	"шибка загрузки онлайн участников канала %[1]s (%[2]s)\x02Ошибка загруз" +
	"ки аутфитов %[1]v (%[2]s)\x02Ошибка загрузки аутфитов по тегам %[1]v (%" +
	"[2]s)\x02Ошибка загрузки персонажей %[1]v (%[2]s)\x02Ошибка сохранения " +
	"настроек подписки для канала %[1]s (%[2]s)\x02Настройки сохранены, но " +
	"не удалось загрузить теги аутфитов %[1]v (%[2]s)\x02Настройки сохранен" +
	"ы, но не удалось загрузить имена персонажей %[1]v (%[2]s)\x02Ошибка за" +
	"грузки настроек подписки для канала %[1]s (%[2]s)\x02Не удалось загрузи" +
	"ть теги аутфитов %[1]v (%[2]s)\x02Не удалось загрузить имена персонажеи" +
	"̆ %[1]v (%[2]s)\x02Настройки отслеживания (PC)\x02Настройки отслеживани" +
	"я (PS4 EU)\x02Настройки отслеживания (PS4 US)\x02Какие аутфиты хотите о" +
	"тслеживать?\x02Введите теги аутфитов через запятую\x02Каких игроков хот" +
	"ите отслеживать?\x02Введите имена персонажей через запятую\x02Невозмож" +
	"но распознать локализацию %[1]q\x02Ошибка при сохранении локализации %[" +
	"1]q\x02Для этого канала был установлен язык %[1]q\x02Нет персонажей онл" +
	"айн\x02Персонажи онлайн:\x02] аутфит (\x04\x01\x0a\x00%\x02**Другие п" +
	"ерсонажи:**\x02ТР\x02НК\x02СВ\x02Источник: %[1]s\x02%[1]s - %[2]d\x02Гл" +
	"обальная популяция - %[1]d\x02Период\x02%[1]s - %[2]s (Кончится %[3]s)" +
	"\x02Контроль территорий\x02%[1]s - Нет тревог\x02%[1]s тревоги\x02Нет тр" +
	"евог\x04\x00\x01\x0a>\x02Настройки обновлены.\x0a\x0a**Аутфиты:**\x02Н" +
	"ет аутфитов\x04\x02\x0a\x0a\x01\x0a\x18\x02**Персонажи:**\x02Нет персон" +
	"ажей\x04\x00\x01 \x1b\x02Разблокирован\x04\x00\x01 \x19\x02Заблокирова" +
	"н\x02) фракцией `\x04\x01\x0a\x00\x10\x02Статус: _\x02Стабильный\x02Не " +
	"стабильный\x02_\x0aТревоги: _\x02Да\x02Нет"

	// Total table size 6296 bytes (6KiB); checksum: F8ACABF9