"use client";

import React from "react";
import { GitBranch, Play, RefreshCw, Terminal } from "lucide-react";

interface SyncTabProps {
  syncRepo: string;
  setSyncRepo: (v: string) => void;
  syncBranch: string;
  setSyncBranch: (v: string) => void;
  syncLoading: boolean;
  onTriggerSync: () => void;
  syncStats: any;
  syncLogs: string[];
}

export default function SyncTab({
  syncRepo, setSyncRepo, syncBranch, setSyncBranch,
  syncLoading, onTriggerSync, syncStats, syncLogs
}: SyncTabProps) {
  return (
    <div className="flex flex-col gap-6">
      <div className="glass-panel panel-cyan flex flex-col gap-4">
        <h2 className="text-lg font-bold text-white flex items-center gap-2">
          <GitBranch className="h-5 w-5 text-[#00F2FE]" />
          Multi-Repository Webhook delta Ingestion Engine
        </h2>
        <p className="text-xs text-slate-400">
          Simulate repository pushes. The Sync agent tracks incremental changes, generates file hashes, and logs delta indexings.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-semibold text-slate-400">Repository URL</label>
            <input type="text" value={syncRepo} onChange={(e) => setSyncRepo(e.target.value)} className="form-input bg-[#0D111C]" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs font-semibold text-slate-400">Branch Scope</label>
            <select value={syncBranch} onChange={(e) => setSyncBranch(e.target.value)} className="form-input bg-[#0D111C]">
              <option value="main">main (Release focus)</option>
              <option value="develop">develop (Experimental overrides)</option>
            </select>
          </div>
        </div>

        <div className="flex gap-4 justify-between items-center mt-2 border-t border-white/5 pt-4">
          <span className="text-xs text-slate-500">Calculates checksum differentials using MD5 hashing.</span>
          <button onClick={onTriggerSync} disabled={syncLoading} className="btn-primary cursor-pointer">
            {syncLoading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Simulate Webhook Push
          </button>
        </div>
      </div>

      {syncStats && (
        <div className="glass-panel flex flex-col gap-3">
          <h3 className="text-sm font-bold text-white border-b border-white/5 pb-2 font-mono">Active Index Ingestion Catalog</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs font-mono">
            {syncStats.map((st: any, i: number) => (
              <div key={i} className="p-4 rounded-lg bg-[#121824]/40 border border-white/5 flex flex-col gap-2">
                <div className="flex justify-between font-semibold border-b border-white/5 pb-1">
                  <span className="text-slate-200">Branch: &apos;{st.branch}&apos;</span>
                  <span className="text-[#10B981] font-mono text-[10px]">{st.syncStatus}</span>
                </div>
                <div className="flex flex-col gap-1 text-[10px] text-slate-400 leading-relaxed font-mono">
                  <div>Commit Hash: {st.lastCommitHash}</div>
                  <div>Total Chunks Indexed: {st.totalChunksIndexed}</div>
                  <div>Files Modified: {st.filesChanged} | Unchanged: {st.filesUnchanged}</div>
                  <div className="text-[9px] text-slate-500 mt-1">Last Synced: {new Date(st.updatedAt).toLocaleString()}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {syncLogs.length > 0 && (
        <div className="glass-panel flex flex-col gap-3">
          <h3 className="text-sm font-bold text-white flex items-center gap-2 font-mono">
            <Terminal className="h-4 w-4 text-[#00F2FE]" />
            Sync Execution logs (Delta calculations)
          </h3>
          <div className="p-4 rounded bg-[#07090E] border border-white/5 text-[10px] font-mono leading-relaxed text-[#10B981] max-h-56 overflow-y-auto">
            {syncLogs.map((log, idx) => (
              <div key={idx} className="truncate">{log}</div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
