// build +generate
package main

import (
	"datfpk/hashing"
	"fmt"
	"os"
)

var extensions = []string{
	"1.ftexs",
	"1.nav2",
	"2.ftexs",
	"3.ftexs",
	"4.ftexs",
	"5.ftexs",
	"6.ftexs",
	"ag.evf",
	"aia",
	"aib",
	"aibc",
	"aig",
	"aigc",
	"aim",
	"aip",
	"ait",
	"atsh",
	"bnd",
	"bnk",
	"cc.evf",
	"clo",
	"csnav",
	"dat",
	"des",
	"dnav",
	"dnav2",
	"eng.lng",
	"ese",
	"evb",
	"evf",
	"fag",
	"fage",
	"fago",
	"fagp",
	"fagx",
	"fclo",
	"fcnp",
	"fcnpx",
	"fdes",
	"fdmg",
	"ffnt",
	"fmdl",
	"fmdlb",
	"fmtt",
	"fnt",
	"fova",
	"fox",
	"fox2",
	"fpk",
	"fpkd",
	"fpkl",
	"frdv",
	"fre.lng",
	"frig",
	"frt",
	"fsd",
	"fsm",
	"fsml",
	"fsop",
	"fstb",
	"ftex",
	"fv2",
	"fx.evf",
	"fxp",
	"gani",
	"geom",
	"ger.lng",
	"gpfp",
	"grxla",
	"grxoc",
	"gskl",
	"htre",
	"info",
	"ita.lng",
	"jpn.lng",
	"json",
	"lad",
	"ladb",
	"lani",
	"las",
	"lba",
	"lng",
	"lpsh",
	"lua",
	"mas",
	"mbl",
	"mog",
	"mtar",
	"mtl",
	"nav2",
	"nta",
	"obr",
	"obrb",
	"param",
	"parts",
	"path",
	"pftxs",
	"ph",
	"phep",
	"phsd",
	"por.lng",
	"qar",
	"rbs",
	"rdb",
	"rdf",
	"rnav",
	"rus.lng",
	"sad",
	"sand",
	"sani",
	"sbp",
	"sd.evf",
	"sdf",
	"sim",
	"simep",
	"snav",
	"spa.lng",
	"spch",
	"sub",
	"subp",
	"tgt",
	"tre2",
	"txt",
	"uia",
	"uif",
	"uig",
	"uigb",
	"uil",
	"uilb",
	"utxl",
	"veh",
	"vfx",
	"vfxbin",
	"vfxdb",
	"vnav",
	"vo.evf",
	"vpc",
	"wem",
	"wmv",
	"xml",
}

func main() {
	res := make(map[string]uint64)
	for _, v := range extensions {
		res[v] = hashing.HashFileName(v, false) & 0x1fff
	}

	var out *os.File
	var err error
	if out, err = os.OpenFile("extension.go", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644); err != nil {
		panic(err)
	}
	defer out.Close()

	if _, err = out.WriteString(`package hashing
var Extensions = map[string]uint64{
`); err != nil {
		panic(err)
	}

	for k, v := range res {
		if _, err = out.WriteString(fmt.Sprintf("\"%s\":0x%x,\n", k, v)); err != nil {
			panic(err)
		}
	}

	if _, err = out.WriteString("}\n"); err != nil {
		panic(err)
	}

	out.WriteString("var ExtensionsByHash = map[uint64]string{\n")
	for k, v := range res {
		if _, err = out.WriteString(fmt.Sprintf("0x%x: \"%s\",\n", v, k)); err != nil {
			panic(err)
		}
	}
	if _, err = out.WriteString("}\n"); err != nil {
		panic(err)
	}
}
