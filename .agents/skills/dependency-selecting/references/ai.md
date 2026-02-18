# AI & RAG

## `github.com/agentable/unifai` — Unified AI Providers

- Single interface for multiple AI service providers
- Provider-agnostic code

**When to use:** Calling LLM APIs (OpenAI, Anthropic, etc.) with a unified interface, swapping providers without code changes, multi-provider fallback.

## `github.com/agentable/knora` — RAG Systems

- Building Retrieval-Augmented Generation systems

**When to use:** Document Q&A, knowledge base search with LLM, context-augmented generation pipelines.

## Decision Tree

```
Need AI integration?
├── Call LLM APIs with unified interface → agentable/unifai
├── Build RAG / knowledge retrieval → agentable/knora
└── Just HTTP calls to AI API → kaptinlin/requests or net/http
```
