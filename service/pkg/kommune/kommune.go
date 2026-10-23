package kommune

import (
	"fmt"
	"strings"
)

// kommuner maps official 4-digit Danish municipal codes (Kommunekoder)
// to their official names.
var kommuner = map[string]string{
	// Region Hovedstaden
	"0101": "København",
	"0147": "Frederiksberg",
	"0151": "Ballerup",
	"0153": "Brøndby",
	"0155": "Dragør",
	"0157": "Gentofte",
	"0159": "Gladsaxe",
	"0161": "Glostrup",
	"0163": "Herlev",
	"0165": "Albertslund",
	"0167": "Hvidovre",
	"0169": "Høje-Taastrup",
	"0173": "Lyngby-Taarbæk",
	"0175": "Rødovre",
	"0183": "Ishøj",
	"0185": "Tårnby",
	"0187": "Vallensbæk",
	"0190": "Furesø",
	"0201": "Allerød",
	"0210": "Fredensborg",
	"0217": "Helsingør",
	"0219": "Hillerød",
	"0223": "Hørsholm",
	"0230": "Rudersdal",
	"0240": "Egedal",
	"0250": "Frederikssund",
	"0260": "Halsnæs",
	"0270": "Gribskov",
	"0400": "Bornholm",
	"0411": "Christiansø",

	// Region Sjælland
	"0253": "Greve",
	"0259": "Køge",
	"0265": "Roskilde",
	"0269": "Solrød",
	"0306": "Odsherred",
	"0316": "Holbæk",
	"0320": "Faxe",
	"0326": "Kalundborg",
	"0329": "Ringsted",
	"0330": "Slagelse",
	"0336": "Stevns",
	"0340": "Sorø",
	"0350": "Lejre",
	"0360": "Lolland",
	"0370": "Næstved",
	"0376": "Guldborgsund",
	"0390": "Vordingborg",

	// Region Syddanmark
	"0420": "Middelfart",
	"0430": "Faaborg-Midtfyn",
	"0440": "Kerteminde",
	"0450": "Nyborg",
	"0461": "Odense",
	"0479": "Svendborg",
	"0480": "Nordfyns",
	"0482": "Langeland",
	"0492": "Ærø",
	"0510": "Haderslev",
	"0530": "Billund",
	"0540": "Sønderborg",
	"0550": "Tønder",
	"0561": "Esbjerg",
	"0563": "Fanø",
	"0573": "Varde",
	"0575": "Vejen",
	"0580": "Vejle",
	"0607": "Fredericia",
	"0621": "Kolding",

	// Region Midtjylland
	"0615": "Horsens",
	"0657": "Herning",
	"0661": "Holstebro",
	"0665": "Lemvig",
	"0671": "Struer",
	"0706": "Syddjurs",
	"0707": "Norddjurs",
	"0710": "Favrskov",
	"0727": "Odder",
	"0730": "Randers",
	"0740": "Silkeborg",
	"0741": "Samsø",
	"0746": "Skanderborg",
	"0751": "Aarhus",
	"0756": "Ikast-Brande",
	"0760": "Ringkøbing-Skjern",
	"0766": "Hedensted",
	"0779": "Skive",
	"0791": "Viborg",

	// Region Nordjylland
	"0773": "Morsø",
	"0787": "Thisted",
	"0810": "Brønderslev",
	"0813": "Frederikshavn",
	"0820": "Vesthimmerland",
	"0825": "Læsø",
	"0840": "Rebild",
	"0846": "Mariagerfjord",
	"0849": "Jammerbugt",
	"0851": "Aalborg",
	"0860": "Hjørring",
}

// yderkommuner holds the set of official Danish municipalities eligible for
// the elevated / full befordringsfradrag under Ligningslovens § 9 C, stk. 3.
var yderkommuner = map[string]bool{
	"0306": true, // Odsherred
	"0330": true, // Slagelse
	"0360": true, // Lolland
	"0376": true, // Guldborgsund
	"0390": true, // Vordingborg
	"0400": true, // Bornholm
	"0430": true, // Faaborg-Midtfyn
	"0479": true, // Svendborg
	"0482": true, // Langeland
	"0492": true, // Ærø
	"0540": true, // Sønderborg
	"0550": true, // Tønder
	"0563": true, // Fanø
	"0665": true, // Lemvig
	"0671": true, // Struer
	"0707": true, // Norddjurs
	"0741": true, // Samsø
	"0760": true, // Ringkøbing-Skjern
	"0773": true, // Morsø
	"0779": true, // Skive
	"0787": true, // Thisted
	"0810": true, // Brønderslev
	"0813": true, // Frederikshavn
	"0820": true, // Vesthimmerland
	"0825": true, // Læsø
	"0849": true, // Jammerbugt
	"0860": true, // Hjørring
}

// LookupKommune resolves a 4-digit municipal code to its municipality name and
// whether it qualifies as an yderkommune under Ligningslovens § 9 C, stk. 3.
func LookupKommune(code string) (name string, isYderkommune bool, found bool) {
	code = strings.TrimSpace(code)
	if len(code) == 3 {
		code = fmt.Sprintf("0%s", code)
	}
	name, exists := kommuner[code]
	if !exists {
		return "", false, false
	}
	return name, yderkommuner[code], true
}
