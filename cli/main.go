package cli

import (
	"datfpk/fox2"
	"datfpk/fpk"
	"datfpk/hashing"
	"datfpk/qar"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const dictionaryName = "dictionary.txt"
const dictUrl = "https://github.com/kapuragu/mgsv-lookup-strings/raw/refs/heads/master/GzsTool/qar_dictionary.txt"

type jsonInput struct {
	Type string `json:"type"`
}

func DecompileFox2(in string, out string) error {
	var err error
	input, err := os.Open(in)
	if err != nil {
		return err
	}
	defer input.Close()

	f := &fox2.Fox2{}
	if err = f.Read(input); err != nil {
		return err
	}

	if out == "" {
		out = in + ".xml"
	}

	outFile, err := os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err = f.ToXML(outFile); err != nil {
		return err
	}

	return nil
}

func CompileFox2(in string, out string) error {
	var err error
	input, err := os.Open(in)
	if err != nil {
		return err
	}
	defer input.Close()

	f := &fox2.Fox2{}
	if err = f.FromXML(input); err != nil {
		return err
	}

	if out == "" {
		out = strings.TrimSuffix(in, ".xml")
	}

	outFile, err := os.OpenFile(out, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err = f.Write(outFile); err != nil {
		return err
	}

	return nil
}

func ExtractQar(qarPath string, dictionaryPath string, outDir string) error {
	var err error

	if qarPath == "" {
		return fmt.Errorf("no dat/qar path provided")
	}

	if dictionaryPath == "" {
		return fmt.Errorf("no dictionary.txt path provided")
	}

	dict := hashing.Dictionary{}

	dictFile, err := os.Open(dictionaryPath)
	if err != nil {
		slog.Error("cannot open QAR dictionary", "path", dictionaryName, "error", err.Error())
		os.Exit(1)
	}
	defer dictFile.Close()

	if err = dict.Read(dictFile); err != nil {
		slog.Error("cannot read QAR dictionary", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("QAR dictionary entries", "count", len(dict.Hashes))

	if outDir != "" {
		o, err := os.Stat(outDir)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(outDir, 0755); err != nil {
				slog.Error("cannot create outdir", "error", err.Error())
				os.Exit(1)
			}
		}
		if err == nil && o != nil {
			if !o.IsDir() {
				slog.Error("not a directory", "path", outDir)
				os.Exit(1)
			}
		}
	}

	q := qar.Qar{}
	if err = q.ReadFrom(qarPath); err != nil {
		slog.Error("QAR read error", "error", err.Error())
		os.Exit(1)
	}

	defer q.Close()

	slog.Info("QAR",
		"version", q.Version,
		"flags", q.Flags,
		"file count", q.FileCount,
		"first file offset", q.OffsetFirstFile,
		"block file end", q.BlockFileEnd,
		"entries", len(q.Entries))

	for n, e := range q.Entries {
		entryName, resolved := dict.GetByHash(e.Header.PathHash)
		q.Entries[n].Header.FilePath = entryName
		if !resolved {
			q.Entries[n].Header.NameHashForPacking = e.Header.PathHash
		}
		slog.Info("qar", "entry", entryName, "pathHash", fmt.Sprintf("%x", e.Header.PathHash), "offset", e.Header.DataOffset, "encrypted", e.DataHeader.EncryptionMagic > 0, "key", fmt.Sprintf("%x", e.DataHeader.Key), "compressed", e.Header.Compressed)

		if _, err = q.Extract(entryName, e.Header.PathHash, outDir); err != nil {
			return err
		}
	}

	descName := qarPath + ".json"
	desc, err := os.OpenFile(descName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("cannot open description file %s for writing: %w", descName, err)
	}

	if err = q.SaveDefinition(desc); err != nil {
		return fmt.Errorf("cannot save description to %s: %w", descName, err)
	}

	return nil
}

func PackQar(jsonDefinitionPath string, outPath string, inputDir string) error {
	var f []byte
	var err error
	if f, err = os.ReadFile(jsonDefinitionPath); err != nil {
		return fmt.Errorf("read qar definition from %s: %w", jsonDefinitionPath, err)
	}

	q := &qar.Qar{}
	if err = json.Unmarshal(f, q); err != nil {
		return fmt.Errorf("unmarshal qar definition: %w", err)
	}

	slog.Info("QAR", "version", q.Version, "filecount", len(q.Entries), "flags", q.Flags)

	if outPath == "" {
		ext := filepath.Ext(jsonDefinitionPath)
		outPath = strings.TrimSuffix(jsonDefinitionPath, ext)
	}

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		slog.Error("open QAR file for writing", "error", err.Error())
		os.Exit(1)
	}

	if inputDir == "" {
		nojs := strings.TrimSuffix(jsonDefinitionPath, ".json")
		ext := filepath.Ext(nojs)
		inputDir = strings.TrimSuffix(nojs, ext) + strings.Replace(ext, ".", "_", -1)
	}
	slog.Info("input", "directory", inputDir, "output", outPath)

	if err = q.Write(out, inputDir); err != nil {
		slog.Error("write QAR", "error", err.Error())
		os.Exit(1)
	}

	return nil
}

func ExtractFpk(path string, outDir string) error {
	var err error
	if outDir != "" {
		o, err := os.Stat(outDir)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(outDir, 0755); err != nil {
				slog.Error("cannot create outdir", "error", err.Error())
				os.Exit(1)
			}
		}
		if err == nil && o != nil {
			if !o.IsDir() {
				slog.Error("not a directory", "path", outDir)
				os.Exit(1)
			}
		}
	}

	f := fpk.Fpk{}
	if err = f.ReadFrom(path); err != nil {
		return fmt.Errorf("fpk(d) read: %w", err)
	}

	slog.Info("extracting fpk(d)", "in", path, "out", outDir)

	for _, v := range f.Entries {
		if err = f.Extract(v.FilePath.Data, outDir); err != nil {
			slog.Error("fpk extract", "path", v.FilePath.Data, "error", err.Error())
			os.Exit(1)
		}
	}

	descName := path + ".json"
	desc, err := os.OpenFile(descName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("cannot open definition file %s for writing: %w", descName, err)
	}

	if err = f.SaveDefinition(desc); err != nil {
		return fmt.Errorf("cannot save definition to %s: %w", descName, err)
	}

	return nil
}

func PackFpk(jsonDefinitionPath string, outPath string, inputDir string) error {
	var f []byte
	var err error
	if f, err = os.ReadFile(jsonDefinitionPath); err != nil {
		return fmt.Errorf("read fpk definition from %s: %w", jsonDefinitionPath, err)
	}

	q := &fpk.Fpk{}
	if err = json.Unmarshal(f, q); err != nil {
		return fmt.Errorf("unmarshal fpk definition: %w", err)
	}

	slog.Info("Fpk", "fileCount", len(q.Entries), "references", len(q.References))

	if outPath == "" {
		ext := filepath.Ext(jsonDefinitionPath)
		outPath = strings.TrimSuffix(jsonDefinitionPath, ext)
	}

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		slog.Error("open fpk file for writing", "error", err.Error())
		os.Exit(1)
	}

	if inputDir == "" {
		nojs := strings.TrimSuffix(jsonDefinitionPath, ".json")
		ext := filepath.Ext(nojs)
		inputDir = strings.TrimSuffix(nojs, ext) + strings.Replace(ext, ".", "_", -1)
	}
	slog.Info("input", "directory", inputDir, "output", outPath)

	if err = q.Write(out, inputDir); err != nil {
		slog.Error("write fpk", "error", err.Error())
		os.Exit(1)
	}

	return nil
}

func Run() {
	datPath := flag.String("dat", "", "path to dat/qar file")

	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		slog.Error("cannot get executable path", "error", err.Error())
		os.Exit(1)
	}
	fp := filepath.Join(filepath.Dir(exePath), dictionaryName)

	dictPath := flag.String("dict", fp, "path to qar dict file")
	out := flag.String("out", "", "output file/directory (default <filename>_<extension>/)")
	jsonPath := flag.String("json", "", "path to qar definition file (.json)")
	inputDir := flag.String("in", "", "input directory path (default <jsonFilename>_<extension>/)")

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println("Pack/unpack MGSV:TPP file formats.")
		fmt.Println("")
		fmt.Println("Unpack (short syntax):")
		fmt.Printf("\t%s file.dat [dictionary.txt]\n", os.Args[0])
		fmt.Printf("\t%s file.dat [output dir] [dictionary.txt]\n", os.Args[0])
		fmt.Printf("\t%s file.fpk [output dir]\n", os.Args[0])
		fmt.Printf("\t%s file.fox2 [output file]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("Pack (short syntax):")
		fmt.Printf("\t%s definition.json [output file] [input dir]\n", os.Args[0])
		fmt.Printf("\t%s file.fox2.xml [output file]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("")
		fmt.Println("Tips:")
		fmt.Printf("  - Get dictionary.txt from %s\n", dictUrl)
		fmt.Printf("  - Create empty dictionary.txt to skip filename resolution.\n")
	}

	flag.Parse()

	if len(os.Args) > 1 {
		if strings.HasSuffix(os.Args[1], ".dat") {
			dp := fp
			if len(os.Args) > 2 {
				if strings.HasSuffix(os.Args[2], ".txt") {
					dp = os.Args[2]
				} else {
					*out = os.Args[2]
				}
			}

			if len(os.Args) > 3 {
				if strings.HasSuffix(os.Args[3], ".txt") {
					dp = os.Args[3]
				}
			}

			if err = ExtractQar(os.Args[1], dp, *out); err != nil {
				slog.Error("extract failed", "error", err.Error())
				os.Exit(1)
			}

			return
		}

		if strings.HasSuffix(os.Args[1], ".json") {
			jj := jsonInput{}
			data, err := os.ReadFile(os.Args[1])
			if err != nil {
				slog.Error("parse json", "error", err.Error())
				os.Exit(1)
			}

			if err = json.Unmarshal(data, &jj); err != nil {
				slog.Error("unmarshal json", "error", err.Error())
				os.Exit(1)
			}

			if len(os.Args) > 2 {
				if !strings.HasPrefix(os.Args[2], "-") {
					*out = os.Args[2]
				}
			}

			if len(os.Args) > 3 {
				if !strings.HasPrefix(os.Args[3], "-") {
					*inputDir = os.Args[3]
				}
			}

			switch jj.Type {
			case fpk.FpkdID, fpk.FpkID:
				slog.Info("fpk")
				if err = PackFpk(os.Args[1], *out, *inputDir); err != nil {
					slog.Error("pack failed", "error", err.Error())
					os.Exit(1)
				}
			case qar.QarID:
				if err = PackQar(os.Args[1], *out, *inputDir); err != nil {
					slog.Error("pack failed", "error", err.Error())
					os.Exit(1)
				}
			default:
				slog.Error("unknown type", "type", jj.Type)
				os.Exit(1)
			}

			return
		}

		if strings.HasSuffix(os.Args[1], ".fpk") || strings.HasSuffix(os.Args[1], ".fpkd") {
			slog.Info("extracting fpk(d)")
			if err = ExtractFpk(os.Args[1], *out); err != nil {
				slog.Error("extract failed", "error", err.Error())
				os.Exit(1)
			}
		}

		if strings.HasSuffix(os.Args[1], ".fox2") {
			slog.Info("decompiling fox2")
			if len(os.Args) > 2 {
				if !strings.HasPrefix(os.Args[2], "-") {
					*out = os.Args[2]
				}
			}
			if err = DecompileFox2(os.Args[1], *out); err != nil {
				slog.Error("fox2 decompilation failed", "error", err.Error())
				os.Exit(1)
			}
		}

		if strings.HasSuffix(os.Args[1], ".fox2.xml") {
			slog.Info("compiling fox2")
			if len(os.Args) > 2 {
				if !strings.HasPrefix(os.Args[2], "-") {
					*out = os.Args[2]
				}
			}
			if err = CompileFox2(os.Args[1], *out); err != nil {
				slog.Error("fox2 compilation failed", "error", err.Error())
				os.Exit(1)
			}
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}

	if *datPath != "" {
		if err = ExtractQar(*datPath, *dictPath, ""); err != nil {
			slog.Error("extract failed", "error", err.Error())
			os.Exit(1)
		}
		return
	}

	if *jsonPath != "" {
		if err = PackQar(*jsonPath, *out, *inputDir); err != nil {
			slog.Error("pack failed", "error", err.Error())
			os.Exit(1)
		}
		return
	}
}
