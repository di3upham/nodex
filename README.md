# IndexCell

`IndexCell` is a Go library designed to manage, generate, and synchronize complex hierarchical spreadsheet-like structures (similar to Microsoft Excel). It helps map multi-dimensional data (multiple indicators/metrics and analysis dimensions/breakdowns) onto a 2D row-column grid and supports reverse-syncing data from edited Excel-like matrix layouts back to Go structures.

---

## Key Features

- **Hierarchical Tree Headers (`ColTree` & `RowTree`)**: Automatically aggregates shared analysis dimensions (dimensions appearing in all metrics) and manages custom/specific breakdowns per metric.
- **Multiple Layout Types (`HeaderType`)**:
  - `flat`: Flat layout where all metrics and dimensions are displayed on columns, and data is represented in flat rows.
  - `matrix_chitieu_in_rows`: Matrix layout where metrics and breakdowns are placed in vertical rows, and shared dimensions are placed on column headers.
  - `matrix_chitieu_in_cols`: Matrix layout where metrics and breakdowns are placed on column headers, and shared dimensions are placed in vertical rows.
- **Header Collapsing (`ColCollapse` & `RowRollapse`)**:
  - Allows rendering the entire hierarchical structure on a **single row** or a **single column**, saving grid space by avoiding empty cell pads.
- **Excel-to-Data Reverse Sync (`importFromMatrix`)**:
  - Parses an edited 2D string matrix (`[][]string`) from an Excel spreadsheet (e.g., when users append new rows, new columns, or update data values).
  - Automatically synchronizes and updates the list of metrics (`ChiTieu`), breakdowns (`PhanTo`), and numerical values (`DongDuLieu`) in the `BieuMau` model.

---

## Core Data Structures

All key structs are defined in [indexcell.go](file:///Users/di3upham/workspace/indexcell/indexcell.go):

- **`ChiTieu`**: Represents a metric or indicator (e.g., *Revenue*, *Expenses*).
- **`PhanTo`**: Represents an analysis dimension or category (e.g., *Region* with values *North*, *South*).
- **`BangChiTieu` & `DongDuLieu`**: Stores the actual data point values indexed by dimensions (`Dims` as Key-Value pairs).
- **`BieuMau`**: The main struct holding configurations, header trees, and cell maps (`Content`).
- **`Node`**: A node in the header trees storing calculated coordinates (`Ri`, `Ci`) for rendering.
- **`Cell`**: Represents a grid cell with a value at a specific `(Ri, Ci)` location.

---

## Getting Started

### 1. Generating a Layout
Here is an example of creating a model and rendering it using a `flat` layout:

```go
package main

import (
	"fmt"
)

func main() {
	// Initialize sample model
	bm := &BieuMau{
		ChiTieus: []*ChiTieu{
			{
				Name: "Revenue",
				PhanTos: []*PhanTo{
					{Name: "Region", Values: []string{"North", "South"}},
					{Name: "Type", Values: []string{"Retail"}},
				},
			},
		},
		HeaderType: "flat",
	}

	// Calculate and generate headers & grid content
	bm.setupFull()
	
	// Access generated cell coordinates via bm.Content
	for idx, cell := range bm.Content {
		fmt.Printf("Cell [%d, %d]: %s\n", idx.Ri, idx.Ci, cell.Value)
	}
}
```

### 2. Importing from an Edited Excel Matrix
If a user edits the spreadsheet layout (such as inserting columns/rows or updating values), pass the modified 2D string matrix `[][]string` to `importFromMatrix`:

```go
// Simulated matrix returned after user edit (added Central Region)
editedMatrix := [][]string{
    {"Region",  "Region",  "Region"},
    {"North",   "Central", "South"}, // "Central" added
    {"Revenue", "Revenue", "Revenue"},
    {"Type",    "Type",    "Type"},
    {"Retail",  "Retail",  "Retail"},
    {"100",     "110",     "150"},   // Values
}

bm.importFromMatrix(editedMatrix)
// The 'bm' instance is now automatically updated with "Central" in the "Region" dimension,
// and the new value of 110 is imported.
```

---

## Running Demo & Tests

To execute the demo:
```bash
go run .
```

To run all unit tests for verifying tree layout calculation, header collapse, and importing logic:
```bash
go test -v
```
