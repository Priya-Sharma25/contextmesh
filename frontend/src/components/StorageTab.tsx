"use client";

import React, { useState } from "react";
import { Database, Copy, Check } from "lucide-react";
import { copyToClipboard } from "@/lib/api";

const PG_VECTOR_SCHEMA = `
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Table to store repository information
CREATE TABLE repositories (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) UNIQUE NOT NULL
);

-- Table to store document catalog
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER REFERENCES repositories(id),
    filepath VARCHAR(512) NOT NULL,
    commit_hash CHAR(40) NOT NULL,
    content TEXT
);

-- Table to store document chunks and embeddings
CREATE TABLE document_chunks (
    id VARCHAR(64) PRIMARY KEY,
    document_id INTEGER REFERENCES documents(id),
    filepath VARCHAR(512) NOT NULL,
    text_content TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    hierarchy_breadcrumbs TEXT,
    embedding vector(1536) -- 1536-dim vector for LLM embeddings
);

-- Create HNSW index for cosine operations
CREATE INDEX ON document_chunks USING hnsw (embedding vector_cosine_ops);
`;

interface StorageTabProps {
  cacheStats: { hits: number; misses: number; ratio: number };
}

export default function StorageTab({ cacheStats }: StorageTabProps) {
  const [schemaCopied, setSchemaCopied] = useState(false);

  return (
    <div className="flex flex-col gap-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass-panel flex flex-col gap-1">
          <span className="text-[10px] text-slate-500 uppercase tracking-widest font-bold">Redis Cache Hits</span>
          <span className="text-2xl font-bold text-[#00F2FE] font-mono">{cacheStats.hits}</span>
          <p className="text-[10px] text-slate-400 mt-2">Successful matches returned from temporary key-value memory wrappers.</p>
        </div>
        <div className="glass-panel flex flex-col gap-1">
          <span className="text-[10px] text-slate-500 uppercase tracking-widest font-bold">Redis Cache Misses</span>
          <span className="text-2xl font-bold text-slate-400 font-mono">{cacheStats.misses}</span>
          <p className="text-[10px] text-slate-400 mt-2">Total queries routed to database vector indexing for semantic similarity checks.</p>
        </div>
        <div className="glass-panel flex flex-col gap-1">
          <span className="text-[10px] text-slate-500 uppercase tracking-widest font-bold">Vector DB Query Latency</span>
          <span className="text-2xl font-bold text-purple-400 font-mono">{(cacheStats.ratio * 100).toFixed(1)}%</span>
          <p className="text-[10px] text-slate-400 mt-2">Overall cache utility score representing overall performance efficiency.</p>
        </div>
      </div>

      {/* Schema */}
      <div className="glass-panel flex flex-col gap-3">
        <div className="flex justify-between items-center border-b border-white/5 pb-2">
          <h3 className="text-sm font-bold text-white flex items-center gap-2">
            <Database className="h-4 w-4 text-[#00F2FE]" />
            PostgreSQL Vector Database (pgvector) Schema Design
          </h3>
          <button onClick={() => copyToClipboard(PG_VECTOR_SCHEMA, setSchemaCopied)} className="btn-secondary px-3 py-1.5 text-xs flex items-center gap-1.5 cursor-pointer">
            {schemaCopied ? <Check className="h-3 w-3 text-[#00F2FE]" /> : <Copy className="h-3 w-3" />}
            Copy Schema SQL
          </button>
        </div>
        <p className="text-xs text-[#94A3B8] leading-relaxed">
          Design specifications utilizing <span className="font-semibold text-slate-200">pgvector HNSW</span> indexes to optimize OpenAI 1536-dimension embedding distance lookups:
        </p>
        <pre className="p-4 rounded bg-[#07090E] border border-white/5 text-[10px] font-mono leading-relaxed text-slate-300 overflow-x-auto whitespace-pre">
          {PG_VECTOR_SCHEMA}
        </pre>
      </div>
    </div>
  );
}
