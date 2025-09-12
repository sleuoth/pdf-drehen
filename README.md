# PDF Reorder for Duplex Front-Side Printing

This small Go utility reorders the pages of a PDF so that you can print it **duplex (double-sided)** on a standard printer and end up with **front sides in sequential order** (1,2,3,…) while the back sides contain the continuation (N/2+1, N/2+2, …).  
The result lets you flip the entire stack halfway through and continue reading in order—perfect for checklists or lists you want to read from front sides only.

## How It Works
Given a PDF with an even number of pages `N`, the tool creates a new PDF where the page order is:
```
1, N/2 + 1, 2, N/2 + 2, 3, N/2 + 3, ...
```
This means:
* Sheet 1: Front = Page 1,  Back = Page N/2 + 1
* Sheet 2: Front = Page 2,  Back = Page N/2 + 2
* ...

Optionally, you can pair the back sides in descending order (e.g., N, N-2, ...) instead of ascending, and you can also rotate all back-side pages by 180 degrees in the output PDF.

When you print the resulting PDF **double-sided (short edge binding)** and read only the front sides, you see pages 1 → 2 → 3… sequentially. After half the sheets, flip the entire stack to continue with the second half.

## Features
- Splits input PDF into single pages
- Reorders them numerically (robust parsing of split filenames)
- Merges them back in the correct sequence
- Automatically adds a blank page if the input PDF has an odd page count

## Requirements
- Go 1.21 or newer
- [pdfcpu](https://github.com/pdfcpu/pdfcpu) library

## Build
```bash
go mod tidy
go build -o pdfreorder .
```

## Usage
```bash
./pdfreorder -in input.pdf -out reordered.pdf
```
Optional flags:
* `-work <dir>`  : Use a specific working directory for split pages
* `-keep`        : Keep the working directory (for debugging)
* `-backdesc`    : Pair backs in descending order (e.g., N, N-2, ...) instead of ascending (S+1, S+2, ...).
* `-rotateback`  : Rotate all back-side pages by 180 degrees in the output PDF.

Example:
```bash
./pdfreorder -in checklist.pdf -out checklist_duplex.pdf -backdesc -rotateback
```

## Printing Instructions
1. Open `checklist_duplex.pdf`.
2. In your printer dialog, enable **Two-Sided / Duplex** printing.
3. Set binding to **Short Edge** (sometimes called “flip on short edge”).
4. Print normally.

Your output will have sequential pages on the front sides, with the second half printed on the back sides of the same sheets.

## License
MIT License
