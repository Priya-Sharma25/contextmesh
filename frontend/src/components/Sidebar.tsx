"use client";

import React from "react";
import {
  Search, Settings, ShieldAlert, Database, GitBranch, Cpu, Award,
  Activity, FileText, RefreshCw, ArrowRight
} from "lucide-react";

type TabId = "orchestrator" | "benchmarks" | "agents-md" | "k8s-validator" | "sync" | "storage";

interface SidebarProps {
  activeTab: TabId;
  setActiveTab: (tab: TabId) => void;
  syncLoading: boolean;
  onTriggerSync: () => void;
}

const NAV_ITEMS: { id: TabId; label: string; icon: React.ReactNode }[] = [
  { id: "orchestrator", label: "Parallel Orchestrator", icon: <Search className="h-3.5 w-3.5" /> },
  { id: "benchmarks", label: "Evaluation Benchmarks", icon: <Award className="h-3.5 w-3.5" /> },
  { id: "agents-md", label: "AGENTS.md Configuration", icon: <Settings className="h-3.5 w-3.5" /> },
  { id: "k8s-validator", label: "Kubernetes Validator", icon: <ShieldAlert className="h-3.5 w-3.5" /> },
  { id: "sync", label: "Multi-Repo Delta Sync", icon: <GitBranch className="h-3.5 w-3.5" /> },
  { id: "storage", label: "pgvector Storage & Cache", icon: <Database className="h-3.5 w-3.5" /> },
];

export default function Sidebar({ activeTab, setActiveTab, syncLoading, onTriggerSync }: SidebarProps) {
  return (
    <aside className="w-[270px] bg-[#090D16] border border-white/[0.04] rounded-2xl flex flex-col justify-between px-4 py-6 shrink-0 h-full overflow-y-auto">
      <div className="flex flex-col gap-5">
        {/* Logo */}
        <div className="flex items-center gap-4 border-b border-white/5 pb-5">
          <div className="h-14 w-14 rounded-lg bg-gradient-to-tr from-[#00F2FE] to-[#9d4edd] p-[1.5px] flex items-center justify-center shadow-lg shadow-cyan-500/10">
            <div className="h-full w-full rounded-lg bg-[#07090E] flex items-center justify-center">
              <Cpu className="h-7 w-7 text-[#00F2FE]" />
            </div>
          </div>
          <div>
            <h1 className="text-lg font-bold tracking-tight text-white flex flex-col leading-none">
              ContextMesh
              <span className="text-[9px] font-semibold text-[#00F2FE] mt-1">Agentic Context Platform</span>
            </h1>
          </div>
        </div>

        {/* Navigation */}
        <div className="flex flex-col justify-between flex-1 mt-4">
          <div className="flex flex-col gap-6">
            <p className="text-[11px] font-bold text-slate-500 uppercase tracking-widest px-3 mt-8 mb-3">ORCHESTRATION & AGENTS</p>
            {NAV_ITEMS.map((item) => (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`flex items-center gap-3 px-2 py-3 mx-2 mb-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${
                  activeTab === item.id
                    ? "sidebar-glow-active"
                    : "bg-[#121824]/20 border border-white/5 text-slate-400 hover:bg-[#121824]/50 hover:text-white"
                }`}
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </div>

          {/* Analytics shortcuts */}
          <div className="flex flex-col gap-6 mt-6">
            <p className="text-[11px] font-bold text-slate-500 uppercase tracking-widest px-3 mb-2">ANALYTICS & INSIGHTS</p>
            <button onClick={() => setActiveTab("benchmarks")} className="flex items-center gap-3 px-2 py-3 mx-2 mb-2 rounded-lg text-xs font-bold bg-[#121824]/20 border border-white/5 text-slate-400 hover:bg-[#121824]/50 hover:text-white transition-all cursor-pointer">
              <Activity className="h-3.5 w-3.5 text-slate-500" /> Retrieval Benchmarks
            </button>
            <button onClick={() => setActiveTab("orchestrator")} className="flex items-center gap-3 px-2 py-3 mx-2 mb-2 rounded-lg text-xs font-bold bg-[#121824]/20 border border-white/5 text-slate-400 hover:bg-[#121824]/50 hover:text-white transition-all cursor-pointer">
              <GitBranch className="h-3.5 w-3.5 text-slate-500" /> Resolved Citations
            </button>
          </div>

          {/* System shortcuts */}
          <div className="flex flex-col gap-6 mt-6">
            <p className="text-[11px] font-bold text-slate-500 uppercase tracking-widest px-3 mb-2">SYSTEM & SETTINGS</p>
            <button onClick={() => setActiveTab("agents-md")} className="flex items-center gap-3 px-2 py-3 mx-2 mb-2 rounded-lg text-xs font-bold bg-[#121824]/20 border border-white/5 text-slate-400 hover:bg-[#121824]/50 hover:text-white transition-all cursor-pointer">
              <Settings className="h-3.5 w-3.5 text-slate-500" /> Settings
            </button>
            <button onClick={() => setActiveTab("sync")} className="flex items-center gap-3 px-2 py-3 mx-2 mb-2 rounded-lg text-xs font-bold bg-[#121824]/20 border border-white/5 text-slate-400 hover:bg-[#121824]/50 hover:text-white transition-all cursor-pointer">
              <FileText className="h-3.5 w-3.5 text-slate-500" /> Audit Logs
            </button>
          </div>
        </div>
      </div>

      {/* Sidebar bottom */}
      <div className="flex flex-col gap-4 border-t border-white/5 pt-5">
        <div className="p-5 rounded-lg bg-[#0D111C] border border-white/5 flex flex-col gap-3 shadow-sm">
          <span className="text-[9px] font-bold text-slate-500 uppercase tracking-wider">Active Ingestion Target</span>
          <div className="flex items-start justify-between">
            <div className="flex flex-col leading-tight font-mono">
              <span className="text-xs font-bold text-[#00F2FE]">contextmesh</span>
              <span className="text-[9px] text-slate-400">Branch: main</span>
            </div>
            <span className="text-xs font-bold text-[#00F2FE] font-mono">78%</span>
          </div>
          <div className="h-1.5 w-full rounded bg-slate-800 overflow-hidden">
            <div className="h-full rounded bg-[#00F2FE] w-[78%]"></div>
          </div>
          <button
            onClick={onTriggerSync}
            disabled={syncLoading}
            className="mt-1 w-full py-1.5 rounded bg-cyan-500/10 hover:bg-cyan-500/20 text-[#00F2FE] text-[10px] font-bold border border-cyan-500/20 flex items-center justify-center gap-1.5 cursor-pointer transition-colors"
          >
            <RefreshCw className={`h-3 w-3 ${syncLoading ? "animate-spin" : ""}`} />
            Webhook sync delta
          </button>
        </div>

        <div className="p-3 rounded-lg bg-[#270E14] border border-red-500/20 flex items-center justify-between shadow-sm cursor-pointer hover:bg-[#33121A] transition-all">
          <div className="flex items-center gap-3">
            <div className="h-7 w-7 rounded bg-red-500/20 flex items-center justify-center text-xs font-bold text-red-400">N</div>
            <div className="flex flex-col">
              <span className="text-xs font-bold text-white leading-none">1 Issue</span>
              <span className="text-[9.5px] text-slate-400 mt-1">View details</span>
            </div>
          </div>
          <ArrowRight className="h-4 w-4 text-red-400" />
        </div>

        <div className="flex items-center justify-between text-[10px] text-slate-500 px-1 font-mono">
          <div className="flex items-center gap-1.5">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>ContextMesh v1.0.0</span>
          </div>
        </div>
      </div>
    </aside>
  );
}

export type { TabId };
