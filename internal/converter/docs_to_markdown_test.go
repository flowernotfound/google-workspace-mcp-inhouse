package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	docs "google.golang.org/api/docs/v1"
)

// --- helpers ---

func paragraph(style string, text string) *docs.StructuralElement {
	return &docs.StructuralElement{
		Paragraph: &docs.Paragraph{
			ParagraphStyle: &docs.ParagraphStyle{
				NamedStyleType: style,
			},
			Elements: []*docs.ParagraphElement{
				{TextRun: &docs.TextRun{Content: text}},
			},
		},
	}
}

func styledRun(content string, style *docs.TextStyle) *docs.ParagraphElement {
	return &docs.ParagraphElement{
		TextRun: &docs.TextRun{
			Content:   content,
			TextStyle: style,
		},
	}
}

// --- TC1: HEADING_1 ---

func TestConvertDocsToMarkdown_Heading1(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("HEADING_1", "Hello"),
			},
		},
	}
	assert.Equal(t, "# Hello", ConvertDocsToMarkdown(doc))
}

// --- TC2: All heading levels ---

func TestConvertDocsToMarkdown_AllHeadingLevels(t *testing.T) {
	cases := []struct {
		style    string
		expected string
	}{
		{"HEADING_1", "# text"},
		{"HEADING_2", "## text"},
		{"HEADING_3", "### text"},
		{"HEADING_4", "#### text"},
		{"HEADING_5", "##### text"},
		{"HEADING_6", "###### text"},
	}
	for _, tc := range cases {
		doc := &docs.Document{
			Body: &docs.Body{
				Content: []*docs.StructuralElement{
					paragraph(tc.style, "text"),
				},
			},
		}
		assert.Equal(t, tc.expected, ConvertDocsToMarkdown(doc), tc.style)
	}
}

// --- TC3: Bold ---

func TestConvertDocsToMarkdown_Bold(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("bold", &docs.TextStyle{Bold: true}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "**bold**", ConvertDocsToMarkdown(doc))
}

// --- TC4: Italic ---

func TestConvertDocsToMarkdown_Italic(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("italic", &docs.TextStyle{Italic: true}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "*italic*", ConvertDocsToMarkdown(doc))
}

// --- TC5: Bold + Italic ---

func TestConvertDocsToMarkdown_BoldItalic(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("bolditalic", &docs.TextStyle{Bold: true, Italic: true}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "***bolditalic***", ConvertDocsToMarkdown(doc))
}

// --- TC6: Strikethrough ---

func TestConvertDocsToMarkdown_Strikethrough(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("strike", &docs.TextStyle{Strikethrough: true}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "~~strike~~", ConvertDocsToMarkdown(doc))
}

// --- TC7: Link ---

func TestConvertDocsToMarkdown_Link(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("click here", &docs.TextStyle{
							Link: &docs.Link{Url: "https://example.com"},
						}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "[click here](https://example.com)", ConvertDocsToMarkdown(doc))
}

// --- TC8: Bullet list ---

func TestConvertDocsToMarkdown_BulletList(t *testing.T) {
	doc := &docs.Document{
		Lists: map[string]docs.List{
			"list1": {
				ListProperties: &docs.ListProperties{
					NestingLevels: []*docs.NestingLevel{
						{GlyphType: "BULLET"},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "item one"}},
					},
				}},
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "item two"}},
					},
				}},
			},
		},
	}
	expected := "- item one\n- item two"
	assert.Equal(t, expected, ConvertDocsToMarkdown(doc))
}

// --- TC9: Numbered list ---

func TestConvertDocsToMarkdown_NumberedList(t *testing.T) {
	doc := &docs.Document{
		Lists: map[string]docs.List{
			"list1": {
				ListProperties: &docs.ListProperties{
					NestingLevels: []*docs.NestingLevel{
						{GlyphType: "DECIMAL"},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "first"}},
					},
				}},
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "second"}},
					},
				}},
			},
		},
	}
	expected := "1. first\n1. second"
	assert.Equal(t, expected, ConvertDocsToMarkdown(doc))
}

// --- TC10: Nested list ---

func TestConvertDocsToMarkdown_NestedList(t *testing.T) {
	doc := &docs.Document{
		Lists: map[string]docs.List{
			"list1": {
				ListProperties: &docs.ListProperties{
					NestingLevels: []*docs.NestingLevel{
						{GlyphType: "BULLET"},
						{GlyphType: "BULLET"},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "parent"}},
					},
				}},
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 1},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "child"}},
					},
				}},
			},
		},
	}
	expected := "- parent\n  - child"
	assert.Equal(t, expected, ConvertDocsToMarkdown(doc))
}

// --- TC11: Table ---

func TestConvertDocsToMarkdown_Table(t *testing.T) {
	makeCell := func(text string) *docs.TableCell {
		return &docs.TableCell{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements:       []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: text}}},
				}},
			},
		}
	}
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Table: &docs.Table{
					TableRows: []*docs.TableRow{
						{TableCells: []*docs.TableCell{makeCell("Name"), makeCell("Age")}},
						{TableCells: []*docs.TableCell{makeCell("Alice"), makeCell("30")}},
					},
				}},
			},
		},
	}
	expected := "| Name | Age |\n| --- | --- |\n| Alice | 30 |"
	assert.Equal(t, expected, ConvertDocsToMarkdown(doc))
}

// --- TC12: Image ---

func TestConvertDocsToMarkdown_Image(t *testing.T) {
	doc := &docs.Document{
		InlineObjects: map[string]docs.InlineObject{
			"obj1": {
				InlineObjectProperties: &docs.InlineObjectProperties{
					EmbeddedObject: &docs.EmbeddedObject{
						Description: "a cat",
						Title:       "cat image",
						ImageProperties: &docs.ImageProperties{
							ContentUri: "https://example.com/cat.png",
						},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "obj1"}},
					},
				}},
			},
		},
	}
	assert.Equal(t, "![a cat](https://example.com/cat.png)", ConvertDocsToMarkdown(doc))
}

// TC12b: Image with no description falls back to title.
func TestConvertDocsToMarkdown_ImageFallbackToTitle(t *testing.T) {
	doc := &docs.Document{
		InlineObjects: map[string]docs.InlineObject{
			"obj1": {
				InlineObjectProperties: &docs.InlineObjectProperties{
					EmbeddedObject: &docs.EmbeddedObject{
						Title: "cat image",
						ImageProperties: &docs.ImageProperties{
							ContentUri: "https://example.com/cat.png",
						},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "obj1"}},
					},
				}},
			},
		},
	}
	assert.Equal(t, "![cat image](https://example.com/cat.png)", ConvertDocsToMarkdown(doc))
}

// --- TC13: Inline code (monospace font) ---

func TestConvertDocsToMarkdown_InlineCode(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("fmt.Println()", &docs.TextStyle{
							WeightedFontFamily: &docs.WeightedFontFamily{FontFamily: "Courier New"},
						}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "`fmt.Println()`", ConvertDocsToMarkdown(doc))
}

// --- TC14: Empty document ---

func TestConvertDocsToMarkdown_EmptyDocument(t *testing.T) {
	assert.Equal(t, "", ConvertDocsToMarkdown(nil))
	assert.Equal(t, "", ConvertDocsToMarkdown(&docs.Document{}))
	assert.Equal(t, "", ConvertDocsToMarkdown(&docs.Document{
		Body: &docs.Body{Content: []*docs.StructuralElement{}},
	}))
}

// --- TC15: Complex document ---

func TestConvertDocsToMarkdown_ComplexDocument(t *testing.T) {
	doc := &docs.Document{
		Lists: map[string]docs.List{
			"list1": {
				ListProperties: &docs.ListProperties{
					NestingLevels: []*docs.NestingLevel{{GlyphType: "BULLET"}},
				},
			},
		},
		InlineObjects: map[string]docs.InlineObject{
			"img1": {
				InlineObjectProperties: &docs.InlineObjectProperties{
					EmbeddedObject: &docs.EmbeddedObject{
						Description: "logo",
						ImageProperties: &docs.ImageProperties{
							ContentUri: "https://example.com/logo.png",
						},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("HEADING_1", "Title"),
				paragraph("NORMAL_TEXT", "Introduction paragraph."),
				{Paragraph: &docs.Paragraph{
					Bullet: &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "list item"}},
					},
				}},
				{Table: &docs.Table{
					TableRows: []*docs.TableRow{
						{TableCells: []*docs.TableCell{
							{Content: []*docs.StructuralElement{
								{Paragraph: &docs.Paragraph{
									ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
									Elements:       []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "Col"}}},
								}},
							}},
						}},
						{TableCells: []*docs.TableCell{
							{Content: []*docs.StructuralElement{
								{Paragraph: &docs.Paragraph{
									ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
									Elements:       []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "Val"}}},
								}},
							}},
						}},
					},
				}},
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "img1"}},
					},
				}},
			},
		},
	}

	result := ConvertDocsToMarkdown(doc)
	assert.Contains(t, result, "# Title")
	assert.Contains(t, result, "Introduction paragraph.")
	assert.Contains(t, result, "- list item")
	assert.Contains(t, result, "| Col |")
	assert.Contains(t, result, "![logo](https://example.com/logo.png)")
}

// --- Section break ---

func TestConvertDocsToMarkdown_SectionBreak(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("NORMAL_TEXT", "before"),
				{SectionBreak: &docs.SectionBreak{}},
				paragraph("NORMAL_TEXT", "after"),
			},
		},
	}
	result := ConvertDocsToMarkdown(doc)
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "before")
	assert.Contains(t, result, "after")
}

// --- Nested numbered list ---

func TestConvertDocsToMarkdown_NestedNumberedList(t *testing.T) {
	doc := &docs.Document{
		Lists: map[string]docs.List{
			"list1": {
				ListProperties: &docs.ListProperties{
					NestingLevels: []*docs.NestingLevel{
						{GlyphType: "DECIMAL"},
						{GlyphType: "DECIMAL"},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					Bullet:   &docs.Bullet{ListId: "list1", NestingLevel: 0},
					Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "parent"}}},
				}},
				{Paragraph: &docs.Paragraph{
					Bullet:   &docs.Bullet{ListId: "list1", NestingLevel: 1},
					Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "child"}}},
				}},
			},
		},
	}
	expected := "1. parent\n  1. child"
	assert.Equal(t, expected, ConvertDocsToMarkdown(doc))
}

// --- Style combinations ---

func TestConvertDocsToMarkdown_StrikethroughBold(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("text", &docs.TextStyle{Strikethrough: true, Bold: true}),
					},
				}},
			},
		},
	}
	// applyTextDecorations applies strikethrough first (inner), then bold (outer).
	assert.Equal(t, "**~~text~~**", ConvertDocsToMarkdown(doc))
}

func TestConvertDocsToMarkdown_StrikethroughItalic(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("text", &docs.TextStyle{Strikethrough: true, Italic: true}),
					},
				}},
			},
		},
	}
	// applyTextDecorations applies strikethrough first (inner), then italic (outer).
	assert.Equal(t, "*~~text~~*", ConvertDocsToMarkdown(doc))
}

func TestConvertDocsToMarkdown_BoldLink(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("text", &docs.TextStyle{
							Bold: true,
							Link: &docs.Link{Url: "https://example.com"},
						}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "[**text**](https://example.com)", ConvertDocsToMarkdown(doc))
}

// --- Monospace overrides link ---

func TestConvertDocsToMarkdown_MonospaceOverridesLink(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("code", &docs.TextStyle{
							WeightedFontFamily: &docs.WeightedFontFamily{FontFamily: "Courier New"},
							Link:               &docs.Link{Url: "https://example.com"},
						}),
					},
				}},
			},
		},
	}
	// Monospace detection takes precedence; link is dropped.
	assert.Equal(t, "`code`", ConvertDocsToMarkdown(doc))
}

// --- Multiple text runs concatenation ---

func TestConvertDocsToMarkdown_MultipleTextRunsConcatenation(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						styledRun("Hello ", nil),
						styledRun("world", &docs.TextStyle{Bold: true}),
					},
				}},
			},
		},
	}
	assert.Equal(t, "Hello **world**", ConvertDocsToMarkdown(doc))
}

// --- Image with empty alt ---

func TestConvertDocsToMarkdown_ImageEmptyAlt(t *testing.T) {
	doc := &docs.Document{
		InlineObjects: map[string]docs.InlineObject{
			"obj1": {
				InlineObjectProperties: &docs.InlineObjectProperties{
					EmbeddedObject: &docs.EmbeddedObject{
						Description: "",
						Title:       "",
						ImageProperties: &docs.ImageProperties{
							ContentUri: "https://example.com/img.png",
						},
					},
				},
			},
		},
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "obj1"}},
					},
				}},
			},
		},
	}
	assert.Equal(t, "![](https://example.com/img.png)", ConvertDocsToMarkdown(doc))
}

// --- TITLE style ---

func TestConvertDocsToMarkdown_TitleStyle(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("TITLE", "My Document"),
			},
		},
	}
	assert.Equal(t, "# My Document", ConvertDocsToMarkdown(doc))
}

// --- nil element guards ---

func TestConvertDocsToMarkdown_NilStructuralElement(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("NORMAL_TEXT", "before"),
				nil,
				paragraph("NORMAL_TEXT", "after"),
			},
		},
	}
	// Should not panic; nil element is skipped.
	result := ConvertDocsToMarkdown(doc)
	assert.Contains(t, result, "before")
	assert.Contains(t, result, "after")
}

func TestConvertDocsToMarkdown_NilParagraphElement(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{Paragraph: &docs.Paragraph{
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: "hello"}},
						nil,
						{TextRun: &docs.TextRun{Content: " world"}},
					},
				}},
			},
		},
	}
	// Should not panic; nil element is skipped.
	assert.Equal(t, "hello world", ConvertDocsToMarkdown(doc))
}

func TestConvertDocsToPlainText_Paragraph(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				paragraph("NORMAL_TEXT", "Hello world"),
			},
		},
	}
	assert.Equal(t, "Hello world", ConvertDocsToPlainText(doc))
}

func TestConvertDocsToPlainText_Table(t *testing.T) {
	doc := &docs.Document{
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					Table: &docs.Table{
						TableRows: []*docs.TableRow{
							{
								TableCells: []*docs.TableCell{
									{Content: []*docs.StructuralElement{
										{Paragraph: &docs.Paragraph{
											Elements: []*docs.ParagraphElement{
												{TextRun: &docs.TextRun{Content: "A"}},
											},
										}},
									}},
									{Content: []*docs.StructuralElement{
										{Paragraph: &docs.Paragraph{
											Elements: []*docs.ParagraphElement{
												{TextRun: &docs.TextRun{Content: "B"}},
											},
										}},
									}},
								},
							},
							{
								TableCells: []*docs.TableCell{
									{Content: []*docs.StructuralElement{
										{Paragraph: &docs.Paragraph{
											Elements: []*docs.ParagraphElement{
												{TextRun: &docs.TextRun{Content: "C"}},
											},
										}},
									}},
									{Content: []*docs.StructuralElement{
										{Paragraph: &docs.Paragraph{
											Elements: []*docs.ParagraphElement{
												{TextRun: &docs.TextRun{Content: "D"}},
											},
										}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}
	result := ConvertDocsToPlainText(doc)
	assert.Contains(t, result, "A\tB")
	assert.Contains(t, result, "C\tD")
}

func TestConvertDocsToPlainText_NilAndEmpty(t *testing.T) {
	assert.Equal(t, "", ConvertDocsToPlainText(nil))
	assert.Equal(t, "", ConvertDocsToPlainText(&docs.Document{}))
}
