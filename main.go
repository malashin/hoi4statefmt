package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

var stateRule = `
	entry                = '' scopeBody$;

	declr                 = (@varNumber '=' @number
						   |@varString '=' @string
						   |@varIdent '=' @ident
						   |@varBool '=' @bool) [@comment]
						   |@ERR_wrong_value_type;
	declrScope            = lval '=' scope [@comment];

	lval                 = @date|@int|@ident;
	rval                 = @date|@number|@ident|@string;

	varNumber            = keywordsNumber|stateBuildings|provinceBuildings|resources|'buildings_max_level_factor'|'add_extra_state_shared_building_slots';
	varString            = 'state_category'|'name'|'has_dlc';
	varIdent             = 'state_category'|'owner'|'add_core_of'|'controller'|'add_claim_by'|'remove_core_of'|'transfer_state'|'type'|'remove_claim_by';
	varBool              = 'impassable'|'set_demilitarized_zone';

	keywordsNumber       = 'id'|'manpower'|'set_province_controller'|'level';
	resources            = 'fuel'|'advanced_technology'|'water'|'electricity'|'metal'|'electronics'|'aluminium'|'chromium'|'steel'|'tungsten'|'oil'|'rubber';
	stateBuildings       = 'infrastructure'|'arms_factory'|'industrial_complex'|'anti_air_building'|'radar_station'|'air_base'|'dockyard'|'metal_generator'|'water_generator'|'electricity_generator'|'synthetic_refinery';
	provinceBuildings    = 'coastal_bunker'|'bunker2'|'bunker'|'naval_base';

	ERR_wrong_value_type = lval '=' rval [@comment];

	numberList           = @number {@number};
	scope                = '{' (scopeBody|@empty) '}';
	scopeBody            = (@declr|@declrScope|@numberList){@declr|@declrScope|@numberList};
	comment              = '#'#anyRune#{#!\x0a#!\x0d#!$#anyRune};

	int                  = digit#{#digit};
	float                = [int]#'.'#int;
	number               = float|int;
	string               = '"'#anyRune#{#!'"'#anyRune}#'"';
	ident                = symbol#{#symbol};
	date                 = int#'.'#int#'.'#int#['.'#int];
	bool                 = 'yes'|'no';

	                     = {spaces|@comment};
	spaces               = \x00..\x20;
	anyRune              = \x00..\xff;
	digit                = '0'..'9';
	letter               = 'a'..'z'|'A'..'Z';
	symbol               = digit|letter|'_'|':'|'@'|'.';
	empty                = '';
`

var test = `
state= {
	id=99
	name="STATE_99"
	manpower = 1672235
	resources={
		aluminium=6 # was: 10
	}

	state_category = town

	history=
	{
		owner = DEN
		victory_points = {
			6364 3
		}
		buildings = {
			infrastructure = 8
			industrial_complex = 3
			air_base = 2
			394 = {
				naval_base = 3
			}
			6364 = {
				naval_base = 1
			}
		}
		add_core_of = DEN
	}
	provinces=
	{
316 332 394 399 3206 3277 3341 6235 6364 11251 	}
}

`

var p *ptool.TParser

func main() {
	var err error
	p, err = ptool.NewBuilder().FromString(stateRule).Entries("entry").Build()
	if err != nil {
		fmt.Println("\nparser error: ", err)
		return
	}

	// fmt.Println("============")
	// builder := ptool.NewBuilder().FromString(stateRule)
	// _, _ = builder.Build()
	// s := builder.TreeToString()
	// fmt.Println(s)
	// fmt.Println("============")

	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("\nioutil.ReadDir error: ", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") && file.Name() != "output.txt" {
			fmt.Println("\n>>>>", file.Name())
			f, err := readFile(file.Name())
			if err != nil {
				fmt.Println("\nreadFile error: ", err)
				return
			}

			fmt.Println(f)
			node, err := p.Parse(f)
			if err != nil {
				fmt.Println("\n*TParser.Parse error: ", err)
				return
			}
			fmt.Println(ptool.TreeToString(node, p.ByID))
			err = traverse(node, 0)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	// node, err := p.Parse(test)
	// if err != nil {
	// 	fmt.Println("\n*TParser.Parse error: ", err)
	// 	// return
	// }
	// fmt.Println(ptool.TreeToString(node, p.ByID))
	// err = traverse(node, 0)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

func traverse(root *ptool.TNode, tabSize int) error {
	tabSize++
	for _, node := range root.Links {
		nodeType := p.ByID(node.Type)
		if strings.HasPrefix(nodeType, "ERR") {
			fmt.Println(nodeType+":", strings.Join(nodesToString(node), " = "))
		}
		switch nodeType {
		case "declrScope", "declr":
			err := traverse(node, tabSize)
			if err != nil {
				return err
			}
		case "number", "ident", "string", "bool":
			fmt.Println(strings.Repeat("  ", tabSize), nodeType, node.Value)
		case "intList":
			list := []string{}
			for _, s := range node.Links {
				list = append(list, s.Value)
			}
			fmt.Println(strings.Repeat("  ", tabSize), nodeType, strings.Join(list, ", "))
		}
	}
	return nil
}

func nodesToString(node *ptool.TNode) []string {
	s := []string{}
	for _, n := range node.Links {
		s = append(s, nodesToString(n)...)
	}
	if node.Value != "" {
		s = append(s, node.Value)
	}
	return s
}

func readFile(path string) (string, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(f), nil
}
