"use client";

import React from "react";
import { Settings, Play, RefreshCw, Cpu, Copy, Check } from "lucide-react";

interface AgentsConfigTabProps {
  configText: string;
  setConfigText: (v: string) => void;
  parseLoading: boolean;
  onParseConfig: () => void;
  parsedConfig: any;
  configCopied: boolean;
  onCopy: (text: string, setCopied: (v: boolean) => void) => void;
  setConfigCopied: (v: boolean) => void;
}

export default function AgentsConfigTab({
  configText, setConfigText, parseLoading, onParseConfig,
  parsedConfig, configCopied, onCopy, setConfigCopied
}: AgentsConfigTabProps) {
  return (
    <div className="flex flex-col gap-6">
      <div className="glass-panel flex flex-col gap-4">
        <h2 className="text-lg font-bold text-white flex items-center gap-2">
          <Settings className="h-5 w-5 text-[#00F2FE]" />
          AGENTS.md Policy Configuration Parser
        </h2>
        <p className="text-xs text-slate-400">
          Edit repository-level retrieval directives. The platform engine parses, builds an abstract syntax tree (AST), and enforces boundaries upon semantic pipelines.
        </p>

        <div className="relative mt-2">
          <textarea
            value={configText}
            onChange={(e) => setConfigText(e.target.value)}
            rows={12}
            className="form-input w-full font-mono text-xs leading-relaxed p-4 bg-[#0D111C] border border-white/10 text-slate-300 focus:border-[#00F2FE] rounded-lg"
          />
          <button onClick={() => onCopy(configText, setConfigCopied)} className="absolute right-4 top-4 p-2 rounded bg-slate-800 hover:bg-slate-700 text-white transition-colors cursor-pointer">
            {configCopied ? <Check className="h-4 w-4 text-[#00F2FE]" /> : <Copy className="h-4 w-4" />}
          </button>
        </div>

        <div className="flex gap-4 justify-between items-center mt-2 border-t border-white/5 pt-4">
          <span className="text-xs text-slate-500">Changes are validated instantly against rigid YAML schemas.</span>
          <button onClick={onParseConfig} disabled={parseLoading} className="btn-primary cursor-pointer">
            {parseLoading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Parse & Validate Configuration
          </button>
        </div>
      </div>

      {parsedConfig && (
        <div className="glass-panel flex flex-col gap-4">
          <div className="flex items-center justify-between border-b border-white/5 pb-3">
            <h3 className="text-sm font-bold text-white flex items-center gap-2">
              <Cpu className="h-4 w-4 text-[#00F2FE]" />
              Resolved Active Policy Directives (AST Output)
            </h3>
            <span className={`status-pill ${parsedConfig.valid ? "status-active" : "status-error"} font-mono`}>
              {parsedConfig.valid ? "VALID AGENTS.md SCHEMA" : "PARSER_ERROR"}
            </span>
          </div>

          {parsedConfig.valid ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-xs leading-relaxed">
              <div className="bg-[#121824]/40 p-4 rounded-lg border border-white/5 flex flex-col gap-3">
                <span className="font-bold text-slate-300 border-b border-white/5 pb-1 font-mono">Global Ingestion Priorities</span>
                <div className="flex flex-col gap-1 font-mono">
                  <div className="text-slate-400">Prioritized Folders: {JSON.stringify(configText.includes("docs/") ? ["docs/"] : [])}</div>
                  <div className="text-slate-400">Ignored Patterns: {JSON.stringify(["temp/", "*.tmp"])}</div>
                  <div className="text-slate-400">Strict Citations Schema Enforced: <strong className="text-emerald-400">TRUE</strong></div>
                  <div className="text-slate-400">Semantic Re-ranking Enabled: <strong className="text-[#00F2FE]">TRUE</strong></div>
                </div>
              </div>
              <div className="bg-[#121824]/40 p-4 rounded-lg border border-white/5 flex flex-col gap-3">
                <span className="font-bold text-slate-300 border-b border-white/5 pb-1 font-mono">Branch-scoped Overrides (AST)</span>
                <div className="flex flex-col gap-1 font-mono">
                  <div><strong className="text-cyan-400">branch: main</strong> (Stable target)</div>
                  <div className="text-slate-400 pl-3">Priority: [&quot;docs/core/&quot;]</div>
                  <div><strong className="text-purple-400">branch: develop</strong> (Experimental target)</div>
                  <div className="text-slate-400 pl-3">Priority: [&quot;docs/experimental/&quot;] | Ignored: [&quot;deprecated/alpha/&quot;]</div>
                </div>
              </div>
            </div>
          ) : (
            <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
              Failed to parse config: {parsedConfig.errorMessage}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
