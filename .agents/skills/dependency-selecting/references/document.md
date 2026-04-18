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

## `github.com/agentable/go-mathconv` — Math Formula Conversion

- Bidirectional: LaTeX ↔ MathML ↔ OMML ↔ Typst
- High performance

**When to use:** Academic publishing, Word ↔ LaTeX workflows, documents with formulas.

## `github.com/agentable/go-image` — Image Processing

- Gamma-correct image processing with fluent API and pure-Go default backend
- Resize, crop, rotate, blur, sharpen, masks, compositing, and SVG rasterization
- Optional libvips acceleration later without changing the core abstraction

**When to use:** General image processing pipelines, thumbnails, image transforms, and pure-Go image workflows.

## `github.com/agentable/go-audio` — Audio Processing

- Probe-first audio library with explicit decoder/encoder activation
- WAV encode/decode is the stable pure-Go path today
- Supports probing, inspection, validation, and transcode workflows

**When to use:** Audio format detection, validation, WAV workflows, and controlled codec-based audio pipelines.

## `github.com/agentable/go-video` — Video Processing

- Probe metadata, trim clips, generate thumbnails, and run FFmpeg-backed transcodes
- Reader-based APIs, audio replacement, frame extraction, and hardware acceleration support
- Structured errors for missing FFmpeg/FFprobe and unsupported operations

**When to use:** Video metadata inspection, thumbnails, clip extraction, and FFmpeg-backed media pipelines.


## Decision Tree

```
Need document/image/media processing?
├── Create/edit Word (.docx) → agentable/godocx
├── Create/edit PDF → agentable/pdfkit
├── Markdown → DOCX/Typst → agentable/markconv
├── Parse documents (multi-format) → agentable/polyparse
├── Translate documents → agentable/polytrans
├── Math formulas across formats → agentable/go-mathconv
├── General image processing → agentable/go-image
├── Audio probing / transcode pipeline → agentable/go-audio
└── Video probing / thumbnails / transcode → agentable/go-video
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `unidoc/unioffice` | Commercial license |
| `jung-kurt/gofpdf` | Archived |
| `h2non/bimg` | Prefer `agentable/go-image` for our default image pipeline |
| `disintegration/imaging` | Prefer `agentable/go-image` for our default image pipeline |
