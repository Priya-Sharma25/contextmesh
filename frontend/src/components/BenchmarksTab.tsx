"use client";

import React from "react";
import { Award, Play, RefreshCw } from "lucide-react";

interface BenchmarksTabProps {
  benchLoading: boolean;
  onRunBenchmark: () => void;
  benchResponse: any;
}

export default function BenchmarksTab({ benchLoading, onRunBenchmark, benchResponse }: BenchmarksTabProps) {
  return (
    <div className="flex flex-col gap-6">
      <div className="glass-panel panel-cyan flex flex-col gap-4">
        <div className="flex justify-between items-start">
          <div>
            <h2 className="text-lg font-bold text-white flex items-center gap-2">
              <Award className="h-5 w-5 text-[#00F2FE]" />
              Evaluation Benchmarks Engine
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              Run evaluation suites comparing vector retrieval strategies against ground truth queries. Evaluates semantic precision, recall, and hallucination safety ratios.
            </p>
          </div>
          <button onClick={onRunBenchmark} disabled={benchLoading} className="btn-primary cursor-pointer">
            {benchLoading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Execute Benchmark Suite
          </button>
        </div>
      </div>

      {benchResponse && (
        <div className="flex flex-col gap-6">
          {/* Metrics Grid */}
          <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
            <MetricCard label="Avg Precision" value={`${(benchResponse.averageMetrics.precisionAtK * 100).toFixed(0)}%`} sub="Precision@5 metric" color="text-white" subColor="text-[#00F2FE]" />
            <MetricCard label="Avg Recall" value={`${(benchResponse.averageMetrics.recallAtK * 100).toFixed(0)}%`} sub="Recall@5 metric" color="text-slate-300" subColor="text-purple-400" />
            <MetricCard label="Hallucination Risk" value={`${(benchResponse.averageMetrics.hallucinationRisk * 100).toFixed(0)}%`} sub="Lower is optimal" color="text-[#10B981]" subColor="text-emerald-400" />
            <MetricCard label="Latency" value={`${benchResponse.averageMetrics.latencyMs.toFixed(1)} ms`} sub="Mean response cycle" color="text-slate-200" subColor="text-amber-400" />
            <MetricCard label="Citation Conf" value={`${(benchResponse.averageMetrics.citationConfidence * 100).toFixed(0)}%`} sub="Exact overlap metric" color="text-slate-200" subColor="text-cyan-400" />
          </div>

          {/* Side-by-Side */}
          <h3 className="text-sm font-bold text-slate-300 border-b border-white/5 pb-2 font-mono">Side-by-Side Strategy Performance</h3>
          {benchResponse.comparisons.map((c: any, i: number) => (
            <div key={i} className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <StrategyCard name={c.strategyAName} label="Strategy A" metrics={c.metricsA} chunks={c.chunksA} borderColor="border-[#00F2FE]" dotColor="bg-[#00F2FE]" badgeBg="bg-cyan-500/10" badgeText="text-[#00F2FE]" badgeBorder="border-cyan-500/20" precisionColor="text-[#00F2FE]" />
              <StrategyCard name={c.strategyBName} label="Strategy B" metrics={c.metricsB} chunks={c.chunksB} borderColor="border-purple-500" dotColor="bg-purple-500" badgeBg="bg-purple-500/10" badgeText="text-purple-400" badgeBorder="border-purple-500/20" precisionColor="text-purple-400" />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function MetricCard({ label, value, sub, color, subColor }: { label: string; value: string; sub: string; color: string; subColor: string }) {
  return (
    <div className="bg-[#0D111C]/60 p-4 rounded-lg border border-white/5 flex flex-col gap-1">
      <span className="text-[10px] text-slate-500 uppercase tracking-widest font-bold">{label}</span>
      <span className={`text-xl font-bold ${color} font-mono`}>{value}</span>
      <span className={`text-[9px] ${subColor}`}>{sub}</span>
    </div>
  );
}

function StrategyCard({ name, label, metrics, chunks, borderColor, dotColor, badgeBg, badgeText, badgeBorder, precisionColor }: any) {
  return (
    <div className={`glass-panel border-l-2 ${borderColor}`}>
      <div className="flex justify-between items-center border-b border-white/5 pb-3 mb-4">
        <h4 className="font-bold text-white flex items-center gap-2">
          <span className={`h-2 w-2 rounded-full ${dotColor}`}></span>
          {name}
        </h4>
        <span className={`text-[10px] px-2 py-0.5 rounded ${badgeBg} ${badgeText} border ${badgeBorder} font-bold uppercase font-mono`}>{label}</span>
      </div>
      <div className="grid grid-cols-3 gap-4 mb-6 font-mono">
        <div className="bg-[#121824]/40 p-3 rounded flex flex-col">
          <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">Precision</span>
          <span className={`text-sm font-bold ${precisionColor}`}>{(metrics.precisionAtK * 100).toFixed(0)}%</span>
        </div>
        <div className="bg-[#121824]/40 p-3 rounded flex flex-col">
          <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">Hallucination Risk</span>
          <span className="text-sm font-bold text-[#10B981]">{(metrics.hallucinationRisk * 100).toFixed(0)}%</span>
        </div>
        <div className="bg-[#121824]/40 p-3 rounded flex flex-col">
          <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">Latency</span>
          <span className="text-sm font-bold text-slate-200">{metrics.latencyMs.toFixed(1)} ms</span>
        </div>
      </div>
      <p className="text-xs text-slate-400 font-bold mb-2">Simulated Retrieved Context Block:</p>
      <pre className="p-3 rounded bg-[#07090E] border border-white/5 text-[11px] font-mono text-slate-300 leading-relaxed overflow-x-auto whitespace-pre-wrap">
        {chunks[0] ? chunks[0].text : "No chunks retrieved."}
      </pre>
    </div>
  );
}
