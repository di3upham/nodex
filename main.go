package main

import (
	"fmt"
)

func printContent(bm *BieuMau) {
	maxRi := 0
	maxCi := 0
	for idx := range bm.Content {
		if idx.Ri > maxRi {
			maxRi = idx.Ri
		}
		if idx.Ci > maxCi {
			maxCi = idx.Ci
		}
	}

	matrix := make([][]string, maxRi+1)
	for i := range matrix {
		matrix[i] = make([]string, maxCi+1)
	}

	for idx, cell := range bm.Content {
		matrix[idx.Ri][idx.Ci] = cell.Value
	}

	for i, row := range matrix {
		for _, val := range row {
			if val == "" {
				val = "-"
			}
			fmt.Printf("%15s\t", val)
		}
		fmt.Println()
		if i == getDepth(bm.ColTree)-1 {
			for range row {
				fmt.Print("----------------\t")
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

func createSampleBieuMau() *BieuMau {
	return &BieuMau{
		ChiTieus: []*ChiTieu{
			{
				Name: "Doanh thu",
				PhanTos: []*PhanTo{
					{Name: "Vùng miền", Values: []string{"Miền Bắc", "Miền Nam"}},
					{Name: "Hình thức", Values: []string{"Bán lẻ"}},
				},
			},
			{
				Name: "Chi phí",
				PhanTos: []*PhanTo{
					{Name: "Vùng miền", Values: []string{"Miền Bắc", "Miền Nam"}},
					{Name: "Loại CP", Values: []string{"Vận hành"}},
				},
			},
		},
		BangChiTieus: []*BangChiTieu{
			{
				ChiTieuName: "Doanh thu",
				DongDuLieus: []*DongDuLieu{
					{
						Dims: []*KV{
							{Key: "Vùng miền", Value: "Miền Bắc"},
							{Key: "Hình thức", Value: "Bán lẻ"},
						},
						Solieu: "100",
					},
					{
						Dims: []*KV{
							{Key: "Vùng miền", Value: "Miền Nam"},
							{Key: "Hình thức", Value: "Bán lẻ"},
						},
						Solieu: "150",
					},
				},
			},
			{
				ChiTieuName: "Chi phí",
				DongDuLieus: []*DongDuLieu{
					{
						Dims: []*KV{
							{Key: "Vùng miền", Value: "Miền Bắc"},
							{Key: "Loại CP", Value: "Vận hành"},
						},
						Solieu: "80",
					},
					{
						Dims: []*KV{
							{Key: "Vùng miền", Value: "Miền Nam"},
							{Key: "Loại CP", Value: "Vận hành"},
						},
						Solieu: "120",
					},
				},
			},
		},
	}
}

func main() {
	bm := createSampleBieuMau()

	fmt.Println("=== 1. KHỞI TẠO BẢNG FLAT ===")
	bm.HeaderType = "flat"
	bm.setupFull()
	printContent(bm)

	fmt.Println("=== 2. GIẢ LẬP EDIT TRÊN EXCEL FLAT (Thêm Miền Trung và CP Bán buôn) ===")
	editedMatrix := [][]string{
		{"Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền"},
		{"Miền Bắc", "Miền Trung", "Miền Nam", "Miền Bắc", "Miền Trung", "Miền Nam", "Miền Nam"},
		{"Doanh thu", "Doanh thu", "Doanh thu", "Chi phí", "Chi phí", "Chi phí", "Chi phí"},
		{"Hình thức", "Hình thức", "Hình thức", "Loại CP", "Loại CP", "Loại CP", "Loại CP"},
		{"Bán lẻ", "Bán lẻ", "Bán lẻ", "Vận hành", "Vận hành", "Vận hành", "Bán buôn"},
		{"100", "110", "150", "80", "90", "120", "200"},
	}

	bm.replaceContent(editedMatrix)

	fmt.Println("=== 3. KẾT QUẢ SAU KHI IMPORT NGƯỢC FLAT ===")
	printContent(bm)

	bm.reset()

	fmt.Println("=== 4. KHỞI TẠO BẢNG MATRIX (CHỈ TIÊU Ở DÒNG) ===")
	bmMatrix := createSampleBieuMau()
	bmMatrix.HeaderType = "matrix_chitieu_in_rows"
	bmMatrix.setupFull()
	printContent(bmMatrix)

	fmt.Println("=== 5. GIẢ LẬP EDIT TRÊN EXCEL MATRIX (Thêm Miền Trung và CP Bán buôn) ===")
	// Old matrix layout has:
	// Columns: 3 header columns (ChiTieu, PhanTo name, PhanTo value) + 2 data columns (Miền Bắc, Miền Nam)
	// Rows: 2 header rows (Vùng miền, value) + 2 data rows (Doanh thu, Chi phí)
	// We edit it to:
	// - Add "Miền Trung" as a 3rd data column
	// - Add "Bán buôn" under "Chi phí" as an extra row
	editedMatrixInRows := [][]string{
		{"-", "-", "-", "Vùng miền", "Vùng miền", "Vùng miền"},
		{"-", "-", "-", "Miền Bắc", "Miền Trung", "Miền Nam"},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "110", "150"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "90", "120"},
		{"Chi phí", "Loại CP", "Bán buôn", "-", "-", "200"},
	}

	bmMatrix.replaceContent(editedMatrixInRows)

	fmt.Println("=== 6. KẾT QUẢ SAU KHI IMPORT NGƯỢC MATRIX ===")
	printContent(bmMatrix)

	fmt.Println("=== 7. KHỞI TẠO BẢNG MATRIX (CHỈ TIÊU Ở CỘT) ===")
	bmMatrixCols := createSampleBieuMau()
	bmMatrixCols.HeaderType = "matrix_chitieu_in_cols"
	bmMatrixCols.setupFull()
	printContent(bmMatrixCols)

	fmt.Println("=== 8. GIẢ LẬP EDIT TRÊN EXCEL MATRIX (CHỈ TIÊU Ở CỘT) (Thêm Miền Trung và CP Bán buôn) ===")
	editedMatrixInCols := [][]string{
		{"-", "-", "Doanh thu", "Chi phí", "Chi phí"},
		{"-", "-", "Hình thức", "Loại CP", "Loại CP"},
		{"-", "-", "Bán lẻ", "Vận hành", "Bán buôn"},
		{"Vùng miền", "Miền Bắc", "100", "80", "-"},
		{"Vùng miền", "Miền Trung", "110", "90", "-"},
		{"Vùng miền", "Miền Nam", "150", "120", "200"},
	}

	bmMatrixCols.replaceContent(editedMatrixInCols)

	fmt.Println("=== 9. KẾT QUẢ SAU KHI IMPORT NGƯỢC MATRIX (CHỈ TIÊU Ở CỘT) ===")
	printContent(bmMatrixCols)
}
