"use client";

import React from "react";
import { Database, Moon, Bell } from "lucide-react";

interface HeaderBarProps {
  cacheRatio: number;
}

export default function HeaderBar({ cacheRatio }: HeaderBarProps) {
  return (
    <header className="sticky top-0 z-30 border-b border-white/5 py-4 px-8 flex items-center justify-between bg-[#07090E]/80 backdrop-blur-md">
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
          <span className="text-xs font-bold font-mono text-slate-300">Backend Connection: Live</span>
        </div>
        <div className="h-4 w-[1px] bg-white/10 hidden sm:block"></div>
        <div className="flex items-center gap-2 text-xs text-slate-400">
          <Database className="h-4 w-4 text-slate-500" />
          <span className="font-bold font-mono">Cache Hit Ratio: {(cacheRatio * 100).toFixed(1)}%</span>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <button className="text-slate-400 hover:text-white transition-colors" title="Toggle theme">
          <Moon className="h-5 w-5" />
        </button>
        <button className="text-slate-400 hover:text-white transition-colors relative" title="Notifications">
          <Bell className="h-5 w-5" />
          <span className="absolute -top-1 -right-1 h-1.5 w-1.5 bg-cyan-500 rounded-full animate-ping"></span>
        </button>
      </div>
    </header>
  );
}
