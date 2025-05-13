datfpk
=====

Based on [FoxTool](https://github.com/Atvaark/FoxTool) and [GzsTool](https://github.com/Atvaark/GzsTool) by Atvaark.

```
Usage of ./datfpk:
Pack/unpack MGSV:TPP file formats.

Unpack (short syntax):
	./datfpk file.dat [dictionary.txt]
	./datfpk file.dat [output dir] [dictionary.txt]
	./datfpk file.fpk [output dir]
	./datfpk file.fox2 [output file]

Pack (short syntax):
	./datfpk definition.json [output file] [input dir]
	./datfpk file.fox2.xml [output file]

Options:

Tips:
  - Get dictionary.txt from https://github.com/kapuragu/mgsv-lookup-strings/raw/refs/heads/master/GzsTool/qar_dictionary.txt
  - Create empty dictionary.txt to skip filename resolution.
```