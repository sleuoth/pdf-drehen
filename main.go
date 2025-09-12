package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	in := flag.String("in", "", "Eingabe-PDF")
	out := flag.String("out", "reordered.pdf", "Ausgabe-PDF")
	work := flag.String("work", "", "Arbeitsordner (optional, sonst temp)")
	keepWork := flag.Bool("keep", false, "Arbeitsordner behalten (Debug)")
	flag.Parse()

	if *in == "" {
		log.Fatal("Bitte -in <eingabe.pdf> angeben")
	}
	absIn, _ := filepath.Abs(*in)

	// Arbeitsordner
	workDir := *work
	var err error
	if workDir == "" {
		workDir, err = os.MkdirTemp("", "pdfreorder_*")
		if err != nil {
			log.Fatalf("Temp-Ordner fehlgeschlagen: %v", err)
		}
		defer func() {
			if !*keepWork {
				os.RemoveAll(workDir)
			} else {
				log.Printf("Arbeitsordner behalten: %s", workDir)
			}
		}()
	} else {
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			log.Fatalf("Arbeitsordner anlegen fehlgeschlagen: %v", err)
		}
	}

	conf := model.NewDefaultConfiguration()

	// 1) In Einzelseiten aufspalten
	if err := api.SplitFile(absIn, workDir, 1, conf); err != nil {
		log.Fatalf("Split fehlgeschlagen: %v", err)
	}

	// 2) Einzel-PDFs auflisten (werden i. d. R. als <name>_0001.pdf etc. benannt)
	entries, err := os.ReadDir(workDir)
	if err != nil {
		log.Fatalf("ReadDir: %v", err)
	}

	type pageFile struct {
		path string
		idx  int
	}
	var pfiles []pageFile

	re := regexp.MustCompile(`_(\d+)\.pdf$`) // match ..._0001.pdf or ..._1.pdf
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
			continue
		}
		full := filepath.Join(workDir, name)

		// Try to extract numeric suffix as page index
		idx := -1
		if m := re.FindStringSubmatch(name); len(m) == 2 {
			fmt.Sscanf(m[1], "%d", &idx)
		}
		if idx < 0 {
			// Fallback: strip extension and parse trailing digits
			base := strings.TrimSuffix(name, filepath.Ext(name))
			// find last run of digits
			pos := len(base) - 1
			for pos >= 0 && base[pos] >= '0' && base[pos] <= '9' {
				pos--
			}
			if pos < len(base)-1 {
				fmt.Sscanf(base[pos+1:], "%d", &idx)
			}
		}
		if idx < 0 {
			log.Printf("Warnung: Konnte Seitenzahl aus %q nicht extrahieren, ignoriere Datei.", name)
			continue
		}
		pfiles = append(pfiles, pageFile{path: full, idx: idx})
	}
	if len(pfiles) == 0 {
		log.Fatal("Keine Seiten nach Split gefunden.")
	}

	// numerisch sortieren
	sort.Slice(pfiles, func(i, j int) bool { return pfiles[i].idx < pfiles[j].idx })

	pages := make([]string, len(pfiles))
	for i, pf := range pfiles {
		pages[i] = pf.path
	}
	N := len(pages)

	// Gerade Seitenzahl sicherstellen (für perfektes Muster)
	if N%2 != 0 {
		fmt.Printf("Warnung: PDF hat eine ungerade Seitenzahl (%d). ", N)
		fmt.Println("Ich hänge eine leere Seite ans Ende an, damit die Anordnung aufgeht.")

		// Leerseite erzeugen: wir nehmen Seite 1 als Template und löschen deren Inhalt – einfacher:
		// Workaround: pdfcpu kann leere Seite per CLI, im API ist’s umständlicher.
		// Deshalb duplizieren wir die letzte Seite; für reine Listen ist das meist okay.
		// (Alternativ: diese Stelle gegen eine echte Blank-Page-Erzeugung tauschen.)
		last := pages[len(pages)-1]
		copyPath := filepath.Join(workDir, "ZZZ_blank_clone.pdf")
		if err := copyFile(last, copyPath); err != nil {
			log.Fatalf("Konnte Dummy-Seite nicht erstellen: %v", err)
		}
		pages = append(pages, copyPath)
		N++
	}

	S := N / 2
	var order []string
	order = make([]string, 0, N)
	for i := 0; i < S; i++ {
		// front i+1
		order = append(order, pages[i])
		// back S+i+1
		order = append(order, pages[S+i])
	}

	// Debug: Ausgabe der Blatt-Zuordnung
	for i := 0; i < S; i++ {
		fmt.Printf("Blatt %2d: Vorderseite=Seite %d, Rückseite=Seite %d\n", i+1, i+1, S+i+1)
	}

	// 3) Neu zusammenführen in gewünschter Reihenfolge
	if err := api.MergeCreateFile(order, *out, false, conf); err != nil {
		log.Fatalf("Merge fehlgeschlagen: %v", err)
	}
	fmt.Printf("Fertig: %s\n", *out)
}

// einfacher Copy helper
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	_, err = out.ReadFrom(in)
	if err != nil {
		return err
	}
	return out.Sync()
}
