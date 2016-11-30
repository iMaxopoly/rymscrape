package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

// generateExcelReport generates complete excel file report using the saved debug file
func generateExcelReport(timestamp, folder string) {
	var (
		file        *xlsx.File
		sheet       *xlsx.Sheet
		row         *xlsx.Row
		cell        *xlsx.Cell
		headerStyle *xlsx.Style
		dataStyle   *xlsx.Style
		err         error
	)

	file = xlsx.NewFile()

	// create sheets
	sheet, err = file.AddSheet("links")
	if err != nil {
		fmt.Printf(err.Error())
	}
	sheet.SetColWidth(0, 0, 15)
	sheet.SetColWidth(1, 1, 60)
	sheet.SetColWidth(2, 2, 60)
	sheet.SetColWidth(3, 3, 10)
	sheet.SetColWidth(4, 4, 50)

	// setup style
	headerStyle = xlsx.NewStyle()
	headerStyle.Font.Bold = true
	headerStyle.Font.Name = "Calibri"
	headerStyle.Font.Size = 11
	headerStyle.Alignment.Horizontal = "center"
	headerStyle.Border.Top = "thin"
	headerStyle.Border.Bottom = "thin"
	headerStyle.Border.Right = "thin"
	headerStyle.Border.Left = "thin"

	dataStyle = xlsx.NewStyle()
	dataStyle.Font.Name = "Calibri"
	dataStyle.Font.Size = 11

	// header internal sheet
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "index"
	cell.SetStyle(headerStyle)

	cell = row.AddCell()
	cell.Value = "cyberlocker_link"
	cell.SetStyle(headerStyle)

	cell = row.AddCell()
	cell.Value = "illegal_website_pagetitle"
	cell.SetStyle(headerStyle)

	cell = row.AddCell()
	cell.Value = "licensor"
	cell.SetStyle(headerStyle)

	cell = row.AddCell()
	cell.Value = "site_url"
	cell.SetStyle(headerStyle)

	data := readFileIntoList(folder + "/debug")

	for i, v := range data {
		// [0] = cyberlockerLink; [1] = pageTitle; [2] = licensor; [3] = siteUrl
		values := strings.Split(v, "\t")
		if len(values) < 4 {
			continue
		}

		pagetitle := values[1]
		licensor := values[2]
		siteUrl := values[0]
		cyberlockerLink := values[3]

		thisRow := sheet.AddRow()

		// index
		thisCell := thisRow.AddCell()
		thisCell.Value = strconv.Itoa(i)
		thisCell.SetStyle(headerStyle)

		// cyberlockerLink
		thisCell = thisRow.AddCell()
		thisCell.Value = cyberlockerLink
		thisCell.SetStyle(dataStyle)

		// pageTitle
		thisCell = thisRow.AddCell()
		thisCell.Value = pagetitle
		thisCell.SetStyle(dataStyle)

		// licensor
		thisCell = thisRow.AddCell()
		thisCell.Value = licensor
		thisCell.SetStyle(dataStyle)

		// siteUrl
		thisCell = thisRow.AddCell()
		thisCell.Value = siteUrl
		thisCell.SetStyle(dataStyle)

	}

	err = file.Save(fmt.Sprintf(folder+"/report_%s.xlsx", timestamp))
	if err != nil {
		fmt.Printf(err.Error())
	}
}
