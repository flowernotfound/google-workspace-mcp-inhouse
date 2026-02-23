package converter

import (
	"strings"

	docs "google.golang.org/api/docs/v1"
)

var headingPrefix = map[string]string{
	"TITLE":     "#",
	"HEADING_1": "#",
	"HEADING_2": "##",
	"HEADING_3": "###",
	"HEADING_4": "####",
	"HEADING_5": "#####",
	"HEADING_6": "######",
}

// ConvertDocsToMarkdown converts a Google Docs document to Markdown text.
func ConvertDocsToMarkdown(doc *docs.Document) string {
	if doc == nil || doc.Body == nil {
		return ""
	}

	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		convertStructuralElement(&sb, elem, doc)
	}
	return strings.TrimSpace(sb.String())
}

func convertStructuralElement(sb *strings.Builder, elem *docs.StructuralElement, doc *docs.Document) {
	if elem == nil {
		return
	}
	switch {
	case elem.Paragraph != nil:
		convertParagraph(sb, elem.Paragraph, doc)
	case elem.Table != nil:
		convertTable(sb, elem.Table, doc)
	case elem.SectionBreak != nil:
		sb.WriteString("---\n\n")
	}
}

func convertParagraph(sb *strings.Builder, para *docs.Paragraph, doc *docs.Document) {
	// List items are handled separately.
	if para.Bullet != nil {
		convertListItem(sb, para, doc)
		return
	}

	text := convertParagraphElements(para.Elements, doc)
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return
	}

	// Heading
	if para.ParagraphStyle != nil {
		if prefix, ok := headingPrefix[para.ParagraphStyle.NamedStyleType]; ok {
			sb.WriteString(prefix + " " + text + "\n\n")
			return
		}
	}

	// Normal paragraph
	sb.WriteString(text + "\n\n")
}

// convertParagraphElements converts the inline elements of a paragraph to a string.
// Inline style conversion (bold, italic, etc.) is applied per element.
func convertParagraphElements(elements []*docs.ParagraphElement, doc *docs.Document) string {
	var sb strings.Builder
	for _, elem := range elements {
		if elem == nil {
			continue
		}
		switch {
		case elem.TextRun != nil:
			sb.WriteString(convertTextRun(elem.TextRun))
		case elem.InlineObjectElement != nil:
			sb.WriteString(convertInlineObject(elem.InlineObjectElement, doc))
		}
	}
	return sb.String()
}

// convertTextRun converts a single text run, applying inline styles.
func convertTextRun(run *docs.TextRun) string {
	content := run.Content
	// Trailing newline is a paragraph separator in the Docs API; strip it.
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return ""
	}

	if run.TextStyle == nil {
		return content
	}
	style := run.TextStyle

	// Inline code: monospace fonts (conservative list).
	// Inline code takes precedence over bold/italic; additional styles are intentionally
	// dropped because Markdown does not support styled inline code spans.
	if isMonospaceFont(style) {
		return "`" + content + "`"
	}

	// Link takes precedence over bold/italic when both are present.
	if style.Link != nil && style.Link.Url != "" {
		inner := applyTextDecorations(content, style)
		return "[" + inner + "](" + style.Link.Url + ")"
	}

	return applyTextDecorations(content, style)
}

// applyTextDecorations applies bold, italic, and strikethrough markers.
func applyTextDecorations(text string, style *docs.TextStyle) string {
	if style == nil {
		return text
	}

	result := text
	if style.Strikethrough {
		result = "~~" + result + "~~"
	}
	if style.Bold && style.Italic {
		result = "***" + result + "***"
	} else if style.Bold {
		result = "**" + result + "**"
	} else if style.Italic {
		result = "*" + result + "*"
	}
	return result
}

// isMonospaceFont reports whether the text run uses a monospace font.
func isMonospaceFont(style *docs.TextStyle) bool {
	if style.WeightedFontFamily == nil {
		return false
	}
	switch style.WeightedFontFamily.FontFamily {
	case "Courier New", "Courier":
		return true
	}
	return false
}

// convertListItem converts a list item paragraph.
func convertListItem(sb *strings.Builder, para *docs.Paragraph, doc *docs.Document) {
	text := convertParagraphElements(para.Elements, doc)
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return
	}

	bullet := para.Bullet
	level := int(bullet.NestingLevel)

	isBullet := true
	if doc != nil {
		if list, ok := doc.Lists[bullet.ListId]; ok && list.ListProperties != nil {
			levels := list.ListProperties.NestingLevels
			if level < len(levels) {
				isBullet = isGlyphBullet(levels[level].GlyphType)
			}
		}
	}

	if isBullet {
		indent := strings.Repeat("  ", level)
		sb.WriteString(indent + "- " + text + "\n")
	} else {
		indent := strings.Repeat("  ", level)
		sb.WriteString(indent + "1. " + text + "\n")
	}
}

// isGlyphBullet reports whether a GlyphType represents a bullet (unordered) list.
func isGlyphBullet(glyphType string) bool {
	switch glyphType {
	case "DECIMAL", "ZERO_DECIMAL", "UPPER_ALPHA", "ALPHA", "UPPER_ROMAN", "ROMAN":
		return false
	default:
		// "BULLET", "GLYPH_TYPE_UNSPECIFIED", or empty → treat as bullet.
		return true
	}
}

// convertTable converts a table element to a Markdown table.
func convertTable(sb *strings.Builder, table *docs.Table, doc *docs.Document) {
	if len(table.TableRows) == 0 {
		return
	}

	// Collect rows as string slices.
	rows := make([][]string, 0, len(table.TableRows))
	for _, row := range table.TableRows {
		if row == nil {
			continue
		}
		cells := make([]string, 0, len(row.TableCells))
		for _, cell := range row.TableCells {
			cells = append(cells, extractCellText(cell, doc))
		}
		rows = append(rows, cells)
	}

	// Determine column count from the widest row.
	cols := 0
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return
	}

	// Header row.
	sb.WriteString("| " + strings.Join(padRow(rows[0], cols), " | ") + " |\n")

	// Separator row.
	sep := make([]string, cols)
	for i := range sep {
		sep[i] = "---"
	}
	sb.WriteString("| " + strings.Join(sep, " | ") + " |\n")

	// Data rows.
	for _, row := range rows[1:] {
		sb.WriteString("| " + strings.Join(padRow(row, cols), " | ") + " |\n")
	}
	sb.WriteString("\n")
}

// extractCellText returns the text content of a table cell as a single line.
func extractCellText(cell *docs.TableCell, doc *docs.Document) string {
	if cell == nil {
		return ""
	}
	var parts []string
	for _, elem := range cell.Content {
		if elem.Paragraph == nil {
			continue
		}
		text := convertParagraphElements(elem.Paragraph.Elements, doc)
		text = strings.TrimRight(text, "\n")
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

// padRow pads a row to the given column count with empty strings.
func padRow(row []string, cols int) []string {
	padded := make([]string, cols)
	copy(padded, row)
	return padded
}

// ConvertDocsToPlainText converts a Google Docs document to plain text (no Markdown markers).
func ConvertDocsToPlainText(doc *docs.Document) string {
	if doc == nil || doc.Body == nil {
		return ""
	}

	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		if elem == nil {
			continue
		}
		if elem.Paragraph != nil {
			extractParagraphText(&sb, elem.Paragraph)
		}
	}
	return strings.TrimSpace(sb.String())
}

// extractParagraphText appends the raw text of a paragraph (no inline markers).
func extractParagraphText(sb *strings.Builder, para *docs.Paragraph) {
	text := extractPlainTextElements(para.Elements)
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return
	}
	sb.WriteString(text + "\n\n")
}

// extractPlainTextElements extracts raw text content from paragraph elements, ignoring styles.
func extractPlainTextElements(elements []*docs.ParagraphElement) string {
	var sb strings.Builder
	for _, elem := range elements {
		if elem == nil {
			continue
		}
		if elem.TextRun != nil {
			content := strings.TrimRight(elem.TextRun.Content, "\n")
			sb.WriteString(content)
		}
	}
	return sb.String()
}

// convertInlineObject converts an inline object element (e.g. image) to Markdown.
func convertInlineObject(elem *docs.InlineObjectElement, doc *docs.Document) string {
	if doc == nil || doc.InlineObjects == nil {
		return ""
	}
	obj, ok := doc.InlineObjects[elem.InlineObjectId]
	if !ok || obj.InlineObjectProperties == nil {
		return ""
	}
	embedded := obj.InlineObjectProperties.EmbeddedObject
	if embedded == nil || embedded.ImageProperties == nil {
		return ""
	}

	url := embedded.ImageProperties.ContentUri
	alt := embedded.Description
	if alt == "" {
		alt = embedded.Title
	}
	return "![" + alt + "](" + url + ")"
}
