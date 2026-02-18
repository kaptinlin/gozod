# Documents & Images

## `github.com/agentable/godocx` — Word Documents

- Pure Go, no external dependencies
- Modern V3 architecture
- Create, manipulate, template Word documents

**When to use:** Report generation, document templating, mail merge, automated DOCX creation.

## `github.com/agentable/pdfkit` — PDF Documents

- API-first design based on go-pdfium
- Create/edit documents, draw content, read structure
- Inspect/edit page objects

**When to use:** Invoice generation, certificate creation, PDF form filling, document annotation.

## `github.com/agentable/markconv` — Markdown Conversion

- Dual output: DOCX and Typst
- Math formula rendering (via mathconv)
- Mermaid diagram conversion
- Intelligent typography for 15+ languages

**When to use:** Documentation publishing, Markdown → Word/PDF pipelines, technical documents.

## `github.com/agentable/polyparse` — Document Parsing

- Modular architecture (Go 1.24+)
- OCR integration capability
- Multi-format support

**When to use:** Document ingestion pipelines, content extraction, format conversion input.

## `github.com/agentable/polytrans` — Document Translation

- Multi-service translation engine

**When to use:** Translating document content across languages, batch translation pipelines.

## `github.com/agentable/mathconv` — Math Formula Conversion

- Bidirectional: LaTeX ↔ MathML ↔ OMML ↔ Typst
- High performance

**When to use:** Academic publishing, Word ↔ LaTeX workflows, documents with formulas.

## `github.com/cshum/vipsgen` — Image Processing

- Auto-generated type-safe Go bindings for libvips (~300 operations)
- Streaming via `io.Reader`/`io.Writer`
- 4-8x faster than ImageMagick, lower memory
- Pre-generated packages for common libvips versions

**Requires:** libvips C library installed on the system.

**When to use:** Image resizing/cropping, thumbnail generation, format conversion, watermarking, high-throughput image pipelines.

## Decision Tree

```
Need document/image processing?
├── Create/edit Word (.docx) → agentable/godocx
├── Create/edit PDF → agentable/pdfkit
├── Markdown → DOCX/Typst → agentable/markconv
├── Parse documents (multi-format) → agentable/polyparse
├── Translate documents → agentable/polytrans
├── Math formulas across formats → agentable/mathconv
└── Image processing → cshum/vipsgen (requires libvips)
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `unidoc/unioffice` | Commercial license |
| `jung-kurt/gofpdf` | Archived |
| `h2non/bimg` | Use vipsgen for comprehensive auto-generated bindings |
| `disintegration/imaging` | Pure Go but much slower than vips |
