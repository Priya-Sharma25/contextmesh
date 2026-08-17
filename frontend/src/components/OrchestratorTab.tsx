"use client";

import React from "react";
import {
  Search, Play, RefreshCw, FileText, Cpu, Activity, GitBranch,
  Heart, ExternalLink, ArrowRight, Copy
} from "lucide-react";

interface OrchestratorTabProps {
  queryInput: string;
  setQueryInput: (v: string) => void;
  queryLoading: boolean;
  onRunQuery: () => void;
  bypassCache: boolean;
  setBypassCache: (v: boolean) => void;
  syncBranch: string;
  queryResponse: any;
  onCopy: (text: string, setCopied: (v: boolean) => void) => void;
}

export default function OrchestratorTab({
  queryInput, setQueryInput, queryLoading, onRunQuery,
  bypassCache, setBypassCache, syncBranch, queryResponse, onCopy
}: OrchestratorTabProps) {
  const getField = (snakeCase: string, camelCase: string) =>
    queryResponse?.[snakeCase] !== undefined ? queryResponse[snakeCase] : queryResponse?.[camelCase];

  const totalTokens = getField("total_tokens", "totalTokens") ?? 1842;
  const relevanceScore = getField("overall_relevance_score", "overallRelevanceScore") ?? 0.94;
  const compressionRatio = getField("compression_ratio", "compressionRatio") ?? 0.46;
  const chunks = getField("retrieved_chunks", "retrievedChunks") ?? [];

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-stretch">
      {/* LEFT COLUMN */}
      <div className="lg:col-span-8 flex flex-col gap-8">
        {/* Search Bar */}
        <div className="glass-panel dashboard-panel panel-cyan flex flex-col gap-5">
          <h2 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <Search className="h-5 w-5 text-[#00F2FE]" />
            Multi-Agent Cooperative Orchestrator
          </h2>
          <p className="text-sm text-slate-400 font-normal leading-relaxed">
            Coordinatively runs <span className="font-semibold text-white">Retriever</span>, <span className="font-semibold text-white">Validator</span>, and <span className="font-semibold text-white">Summarizer</span> agents in parallel. Applies branch-scoped priority path biases resolved from your repository policies.
          </p>

          <div className="flex gap-4 mt-1">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-4.5 h-5 w-5 text-[#64748B] transition-colors" />
              <input
                type="text"
                value={queryInput}
                onChange={(e) => setQueryInput(e.target.value)}
                placeholder="Enter agentic context query or command..."
                className="form-input-command w-full pl-12 pr-4 h-14 rounded-xl text-white font-medium focus:outline-none"
              />
            </div>
            <button onClick={onRunQuery} disabled={queryLoading} className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white font-bold px-12 h-16 rounded-xl flex items-center gap-2.5 text-sm shadow-md shadow-blue-500/5 hover:shadow-blue-500/10 transition-all hover:scale-[1.01] cursor-pointer shrink-0">
              {queryLoading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4 fill-current animate-pulse" />}
              Run Orchestrator
            </button>
          </div>

          <div className="flex items-center gap-6 text-xs text-[#94A3B8] border-t border-white/5 pt-4 mt-1 font-mono">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" checked={bypassCache} onChange={(e) => setBypassCache(e.target.checked)} className="accent-[#00F2FE] rounded" />
              Bypass Redis Cache
            </label>
            <span>Active Branch: <strong className="text-slate-200">{syncBranch}</strong></span>
            <span>Active Repository: <strong className="text-slate-200">github.com/Priyasharma620064/contextmesh</strong></span>
          </div>
        </div>

        {/* Pipeline Flow */}
        <div className="glass-panel dashboard-panel flex flex-col gap-5 border border-white/5 bg-[#0D111C]/40 backdrop-blur-lg">
          <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest flex items-center gap-2">
            <Cpu className="h-4 w-4 text-[#00F2FE]" />
            Active Retrieval Pipeline Execution Flow
          </h3>
          <div className="relative flex flex-col md:flex-row items-center justify-between gap-6 md:gap-0 py-6 px-4 max-w-[700px] mx-auto w-full">
            <div className="hidden md:block absolute top-1/2 left-0 w-full h-6 -translate-y-1/2 z-0 px-[52px]">
              <svg className="w-full h-full" fill="none" viewBox="0 0 100 24" preserveAspectRatio="none">
                <defs><linearGradient id="flowGrad" x1="0%" y1="0%" x2="100%" y2="0%"><stop offset="0%" stopColor="#00F2FE" stopOpacity="0.8" /><stop offset="50%" stopColor="#9d4edd" stopOpacity="0.8" /><stop offset="100%" stopColor="#10B981" stopOpacity="0.8" /></linearGradient></defs>
                <path d="M 0 12 L 100 12" stroke="url(#flowGrad)" strokeWidth="4" strokeDasharray="6 4" className="pulse-path" />
              </svg>
            </div>
            {[
              { num: 1, label: "Query Ingest", sub: "Input Received", time: "18ms", color: "cyan" },
              { num: 2, label: "Retriever", sub: "+25% Path Bias", time: "42ms", color: "purple" },
              { num: 3, label: "Intel Validator", sub: "API Version Check", time: "96ms", color: "amber" },
              { num: 4, label: "Compressor", sub: "Token budget bound", time: "72ms", color: "emerald" },
            ].map((node) => (
              <div key={node.num} className="flex flex-col items-center gap-2.5 shrink-0 z-10">
                <div className={`h-[72px] w-[72px] rounded-full border border-white/10 flex items-center justify-center text-xl font-extrabold text-white bg-slate-900 transition-all duration-300 ${queryLoading ? `animate-pulse border-${node.color}-400/50 shadow-[0_0_15px_rgba(0,242,254,0.15)]` : ""}`}>
                  {node.num}
                </div>
                <span className="text-[13px] font-semibold text-slate-300">{node.label}</span>
                <span className={`text-[10px] text-${node.color}-400/70 font-mono`}>{node.sub}</span>
                <span className={`px-2 py-0.5 rounded-full text-[8px] bg-${node.color}-500/10 text-${node.color}-400 border border-${node.color}-500/20 font-bold font-mono mt-1`}>{node.time}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Compiled Context */}
        <div className="glass-panel dashboard-panel flex flex-col gap-3">
          <div className="flex items-center justify-between border-b border-white/5 pb-3">
            <h3 className="text-sm font-bold text-white flex items-center gap-2">
              <FileText className="h-4 w-4 text-[#00F2FE]" />
              Compiled Prompt Context (Token compressed)
            </h3>
            <span className="text-[10px] text-slate-500 font-mono">XML format ready for LLMs</span>
          </div>
          <div className="relative">
            <div className="flex rounded-xl bg-[#05070B] border border-white/5 overflow-x-auto text-[12px] font-mono leading-[22px] max-h-[380px] overflow-y-auto">
              <div className="flex flex-col text-slate-600 select-none text-right px-3 py-4 border-r border-white/5 bg-[#07090F] min-h-full">
                {Array.from({ length: queryResponse ? queryResponse.compiledContext?.split('\n').length || 8 : 8 }).map((_, idx) => (
                  <span key={idx} className="block h-[22px] leading-[22px]">{idx + 1}</span>
                ))}
              </div>
              <pre className="flex-1 p-4 overflow-x-auto text-[12px] font-mono leading-[22px] whitespace-pre text-slate-300">
                <code>{queryResponse?.compiledContext || "Loading..."}</code>
              </pre>
            </div>
            <button onClick={() => onCopy(queryResponse?.compiledContext || "", () => {})} className="absolute right-4 top-4 p-2 rounded bg-slate-800 hover:bg-slate-700 text-white transition-colors cursor-pointer" title="Copy Compiled Prompt">
              <Copy className="h-4 w-4" />
            </button>
          </div>
          <div className="flex items-center justify-between text-xs text-slate-400 pt-3 border-t border-white/5 font-mono">
            <span>Tokens: {totalTokens} / 4,000</span>
            <span className="px-2 py-0.5 bg-emerald-500/10 text-[#10B981] rounded border border-emerald-500/20 text-[10px] font-bold">
              {((totalTokens / 4000) * 100).toFixed(0)}% of budget
            </span>
          </div>
        </div>
      </div>

      {/* RIGHT COLUMN */}
      <div className="lg:col-span-4 flex flex-col gap-8">
        {/* Telemetry Metrics */}
        <div className="glass-panel dashboard-panel flex flex-col gap-6">
          <h3 className="text-base font-semibold text-white border-b border-white/5 pb-3 flex items-center gap-2">
            <Activity className="h-4 w-4 text-[#00F2FE]" /> Ingested Telemetry Metrics
          </h3>
          <GaugeBar label="Context Relevance Score" value={relevanceScore} color="cyan" />
          <GaugeBar label="Context Compression Ratio" value={compressionRatio} color="purple" />
          <div className="flex flex-col gap-3 border-t border-white/5 pt-4 text-xs font-mono">
            <div className="flex justify-between"><span className="text-slate-400">Total Tokens</span><span className="text-white font-bold">{totalTokens} Tokens</span></div>
            <div className="flex justify-between"><span className="text-slate-400">Latency (p95)</span><span className="text-emerald-400 font-bold">2.1 ms</span></div>
            <div className="flex justify-between font-mono"><span className="text-slate-400">Cache Strategy</span><span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold text-[10px]">REDIS_CACHE_HIT</span></div>
          </div>
        </div>

        {/* Citations Graph */}
        <div className="glass-panel dashboard-panel flex flex-col gap-5">
          <div className="flex items-center justify-between border-b border-white/5 pb-2">
            <h3 className="text-base font-semibold text-white flex items-center gap-2"><GitBranch className="h-4 w-4 text-[#00F2FE]" /> Resolved Citations Graph</h3>
            <a href="#citations" className="text-[10px] text-cyan-400 hover:text-cyan-300 font-semibold flex items-center gap-1">View full graph <ExternalLink className="h-3 w-3" /></a>
          </div>
          <div className="flex flex-col gap-2">
            {(chunks.length > 0 ? chunks : [
              { filepath: "docs/architecture/...", score: 0.94, startLine: 120, endLine: 138, id: "SRC-1" },
              { filepath: "pkg/agents/orchestrator.go", score: 0.88, startLine: 80, endLine: 95, id: "SRC-2" }
            ]).map((chunk: any, i: number) => (
              <div key={i} className="p-3 rounded bg-[#121824]/50 border border-white/5 flex flex-col gap-1 text-xs">
                <div className="flex justify-between font-semibold">
                  <span className="text-slate-200 truncate max-w-[140px] font-mono">{chunk.filepath}</span>
                  <span className="text-[#00F2FE]">Score: {chunk.score ? chunk.score.toFixed(2) : "0.90"}</span>
                </div>
                <span className="text-[10px] text-slate-500 font-mono">Lines {chunk.startLine}-{chunk.endLine} | ID: CIT-{chunk.id ? chunk.id.substring(0, 6) : "d62a37"}</span>
              </div>
            ))}
          </div>
          <button className="w-full mt-2 py-2 rounded bg-slate-900/60 hover:bg-slate-900 border border-white/5 text-slate-300 text-xs font-bold flex items-center justify-center gap-1.5 cursor-pointer transition-colors">
            View all citations <ArrowRight className="h-3.5 w-3.5" />
          </button>
        </div>

        {/* System Health */}
        <div className="glass-panel dashboard-panel flex flex-col gap-6">
          <h3 className="text-base font-semibold text-white border-b border-white/5 pb-3 flex items-center gap-2">
            <Heart className="h-4 w-4 text-rose-500 fill-rose-500/20 animate-pulse" /> System Health
          </h3>
          <div className="flex flex-col gap-3 text-xs font-mono">
            {["Backend", "Redis Cache", "PgVector", "Kubernetes", "Webhooks"].map((name) => (
              <div key={name} className="flex justify-between items-center">
                <span className="text-slate-400">{name}</span>
                <span className="flex items-center gap-1.5 text-slate-300">
                  <span className="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  {name === "Redis Cache" ? "Hit Ratio: 58%" : "Healthy"}
                </span>
              </div>
            ))}
          </div>
          <button className="w-full mt-1 py-2 rounded bg-slate-900/60 hover:bg-slate-900 border border-white/5 text-slate-300 text-xs font-bold flex items-center justify-center gap-1.5 cursor-pointer transition-colors">
            View system logs <ArrowRight className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
}

function GaugeBar({ label, value, color }: { label: string; value: number; color: string }) {
  const colorClass = color === "cyan" ? "from-cyan-400 to-cyan-500" : "from-purple-400 to-purple-500";
  const textColor = color === "cyan" ? "text-[#00F2FE]" : "text-purple-400";
  return (
    <div className="flex flex-col gap-2">
      <div className="flex justify-between text-xs font-semibold">
        <span className="text-slate-400">{label}</span>
        <span className={`${textColor} font-mono`}>{(value * 100).toFixed(0)}%</span>
      </div>
      <div className="h-2 w-full rounded bg-slate-800 overflow-hidden">
        <div className={`h-full rounded bg-gradient-to-r ${colorClass} transition-all duration-500`} style={{ width: `${value * 100}%` }}></div>
      </div>
    </div>
  );
}
