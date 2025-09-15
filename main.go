package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	in := flag.String("in", "", "Input PDF")
	out := flag.String("out", "reordered.pdf", "Output PDF")
	work := flag.String("work", "", "Working directory (optional, default temp)")
	keepWork := flag.Bool("keep", false, "Keep working directory (debug)")
	backDesc := flag.Bool("backdesc", false, "Pair backs in descending order (e.g., N, N-2, ...) instead of ascending (S+1, S+2, ...)")
	rotateBack := flag.Bool("rotateback", false, "Rotate all back-side pages by 180 degrees in the output PDF")
	flag.Parse()

	if *in == "" {
		log.Fatal("Please provide -in <input.pdf>")
	}
	absIn, _ := filepath.Abs(*in)

	// Working directory
	workDir := *work
	var err error
	if workDir == "" {
		workDir, err = os.MkdirTemp("", "pdfreorder_*")
		if err != nil {
			log.Fatalf("Creating temp directory failed: %v", err)
		}
		defer func() {
			if !*keepWork {
				os.RemoveAll(workDir)
			} else {
				log.Printf("Keeping working directory: %s", workDir)
			}
		}()
	} else {
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			log.Fatalf("Failed to create working directory: %v", err)
		}
	}

	conf := model.NewDefaultConfiguration()

	// 1) Split input into single pages
	if err := api.SplitFile(absIn, workDir, 1, conf); err != nil {
		log.Fatalf("Split failed: %v", err)
	}

	// 2) List single-page PDFs (usually named like <name>_0001.pdf etc.)
	entries, err := os.ReadDir(workDir)
	if err != nil {
		log.Fatalf("ReadDir failed: %v", err)
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
			log.Printf("Warning: Could not extract page number from %q, skipping file.", name)
			continue
		}
		pfiles = append(pfiles, pageFile{path: full, idx: idx})
	}
	if len(pfiles) == 0 {
		log.Fatal("No pages found after split.")
	}

	// sort numerically
	sort.Slice(pfiles, func(i, j int) bool { return pfiles[i].idx < pfiles[j].idx })

	pages := make([]string, len(pfiles))
	for i, pf := range pfiles {
		pages[i] = pf.path
	}
	N := len(pages)

	// Ensure even number of pages for perfect pattern
	if N%2 != 0 {
		fmt.Printf("Warning: PDF has an odd number of pages (%d). ", N)
		fmt.Println("Adding a blank page at the end to maintain correct ordering.")

		// Create blank page: here we duplicate the last page as a simple workaround.
		// Workaround: pdfcpu CLI can create blank page easily, API is more complex.
		// So we duplicate the last page; usually fine for plain lists.
		// (Alternatively replace with a true blank-page creation if desired.)
		last := pages[len(pages)-1]
		copyPath := filepath.Join(workDir, "ZZZ_blank_clone.pdf")
		if err := copyFile(last, copyPath); err != nil {
			log.Fatalf("Could not create dummy page: %v", err)
		}
		pages = append(pages, copyPath)
		N++
	}

	S := N / 2
	var order []string
	order = make([]string, 0, N)
	for i := 0; i < S; i++ {
		// front side i+1
		order = append(order, pages[i])
		// back side
		if *backDesc {
			b := N - 1 - i
			if b >= 0 && b < N {
				order = append(order, pages[b])
			}
		} else {
			order = append(order, pages[S+i])
		}
	}

	// Debug: print sheet mapping
	for i := 0; i < S; i++ {
		if *backDesc {
			fmt.Printf("Sheet %2d: Front=Page %d, Back=Page %d (descending)\n", i+1, i+1, N-i)
		} else {
			fmt.Printf("Sheet %2d: Front=Page %d, Back=Page %d (ascending)\n", i+1, i+1, S+i+1)
		}
	}

	// 3) Merge pages in desired order
	if err := api.MergeCreateFile(order, *out, false, conf); err != nil {
		log.Fatalf("Merge failed: %v", err)
	}
	if *rotateBack {
		selectedPages := []string{}
		for i := 2; i <= len(order); i += 2 {
			selectedPages = append(selectedPages, strconv.Itoa(i))
		}
		if err := api.RotateFile(*out, *out, 180, selectedPages, conf); err != nil {
			log.Fatalf("Rotate failed: %v", err)
		}
	}
	fmt.Printf("Done: %s\n", *out)
}

// simple copy helper
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
