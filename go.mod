module github.com/TrueBlocks/trueblocks-art/differ

go 1.25.1

require (
	github.com/TrueBlocks/trueblocks-art/packages/docxzip v0.0.0-00010101000000-000000000000
	golang.org/x/term v0.40.0
	golang.org/x/text v0.35.0
)

require golang.org/x/sys v0.41.0 // indirect

replace github.com/TrueBlocks/trueblocks-art/packages/docxzip => ../packages/docxzip
