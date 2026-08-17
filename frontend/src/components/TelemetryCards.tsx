"use client";

import React from "react";
import { Cpu, Database, GitBranch, ShieldAlert } from "lucide-react";

export default function TelemetryCards() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <div className="glass-panel telemetry-card border-t-2 border-t-[#00F2FE] bg-[#0D111C]/40 flex items-center gap-4">
        <div className="h-12 w-12 rounded-xl bg-cyan-500/10 flex items-center justify-center border border-cyan-500/20 shrink-0 shadow-lg shadow-cyan-500/5">
          <Cpu className="h-6 w-6 text-[#00F2FE]" />
        </div>
        <div className="flex flex-col">
          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Active Agents</span>
          <span className="text-xl font-bold text-white leading-none mt-1">4</span>
          <span className="text-[9.5px] text-emerald-400 font-semibold flex items-center gap-1 mt-1">
            Online <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          </span>
        </div>
      </div>

      <div className="glass-panel telemetry-card border-t-2 border-t-[#00F2FE] bg-[#0D111C]/40 flex items-center gap-4">
        <div className="h-12 w-12 rounded-xl bg-cyan-500/10 flex items-center justify-center border border-cyan-500/20 shrink-0 shadow-lg shadow-cyan-500/5">
          <Database className="h-6 w-6 text-[#00F2FE]" />
        </div>
        <div className="flex flex-col">
          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Indexed Repos</span>
          <span className="text-xl font-bold text-white leading-none mt-1">12</span>
          <span className="text-[9.5px] text-slate-400 font-medium mt-1">Repos Indexed</span>
        </div>
      </div>

      <div className="glass-panel telemetry-card border-t-2 border-t-purple-500 bg-[#0D111C]/40 flex items-center gap-4">
        <div className="h-12 w-12 rounded-xl bg-purple-500/10 flex items-center justify-center border border-purple-500/20 shrink-0 shadow-lg shadow-purple-500/5">
          <GitBranch className="h-6 w-6 text-purple-400" />
        </div>
        <div className="flex flex-col">
          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Benchmark score</span>
          <span className="text-sm font-bold text-white leading-none mt-1">Precision@5: 0.91</span>
          <span className="text-[9.5px] text-emerald-400 font-medium flex items-center gap-1 mt-1">
            Passed <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          </span>
        </div>
      </div>

      <div className="glass-panel telemetry-card border-t-2 border-t-amber-500 bg-[#0D111C]/40 flex items-center gap-4">
        <div className="h-12 w-12 rounded-xl bg-amber-500/10 flex items-center justify-center border border-amber-500/20 shrink-0 shadow-lg shadow-amber-500/5">
          <ShieldAlert className="h-6 w-6 text-amber-400" />
        </div>
        <div className="flex flex-col">
          <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Hallucination Risk</span>
          <span className="text-xl font-bold text-white leading-none mt-1">LOW</span>
          <span className="px-1.5 py-0.2 text-[8px] bg-emerald-500/10 text-emerald-400 rounded-sm font-bold border border-emerald-500/20 max-w-max mt-1 leading-none">PASSED</span>
        </div>
      </div>
    </div>
  );
}
