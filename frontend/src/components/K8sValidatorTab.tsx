"use client";

import React from "react";
import { ShieldAlert, Play, RefreshCw } from "lucide-react";

interface K8sValidatorTabProps {
  yamlInput: string;
  setYamlInput: (v: string) => void;
  k8sLoading: boolean;
  onValidate: () => void;
  k8sValid: boolean | null;
  k8sViolations: any[];
}

export default function K8sValidatorTab({
  yamlInput, setYamlInput, k8sLoading, onValidate, k8sValid, k8sViolations
}: K8sValidatorTabProps) {
  return (
    <div className="flex flex-col gap-6">
      <div className="glass-panel flex flex-col gap-4">
        <h2 className="text-lg font-bold text-white flex items-center gap-2">
          <ShieldAlert className="h-5 w-5 text-[#EF4444]" />
          Kubernetes Documentation Schema Validator
        </h2>
        <p className="text-xs text-slate-400">
          Scan embedded YAML code fences or full manifest files. Detects <span className="font-semibold text-white">apiVersion drifts</span>, legacy schemas (e.g. extensions/v1beta1 Deployments), and yields self-healing corrections.
        </p>

        <div className="mt-2">
          <textarea
            value={yamlInput}
            onChange={(e) => setYamlInput(e.target.value)}
            rows={12}
            className="form-input w-full font-mono text-xs leading-relaxed p-4 bg-[#0D111C] border border-white/10 text-slate-300 focus:border-[#00F2FE] rounded-lg"
          />
        </div>

        <div className="flex gap-4 justify-between items-center mt-2 border-t border-white/5 pt-4">
          <span className="text-xs text-slate-500">Evaluates schema compatibility against Kubernetes v1.25 rules.</span>
          <button onClick={onValidate} disabled={k8sLoading} className="btn-primary cursor-pointer">
            {k8sLoading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Scan Yaml Schema
          </button>
        </div>
      </div>

      {k8sValid !== null && (
        <div className="glass-panel flex flex-col gap-4">
          <div className="flex items-center justify-between border-b border-white/5 pb-3">
            <h3 className="text-sm font-bold text-white flex items-center gap-2">
              <ShieldAlert className="h-4 w-4 text-[#00F2FE]" />
              Schema Diagnostics Results
            </h3>
            <span className={`status-pill ${k8sValid ? "status-active" : "status-error"} font-mono`}>
              {k8sValid ? "MANIFEST_COMPLIANT" : "LEGACY_DRIFTS_DETECTED"}
            </span>
          </div>

          {k8sViolations.length > 0 ? (
            <div className="flex flex-col gap-3">
              {k8sViolations.map((v: any, i: number) => (
                <div key={i} className="p-4 rounded-lg bg-red-500/5 border border-red-500/10 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 text-xs font-mono">
                  <div className="flex flex-col gap-1.5 max-w-xl">
                    <span className="font-bold text-slate-200">DRIFT: {v.kind || v.Kind} uses outdated apiVersion &apos;{v.api_version || v.apiVersion}&apos;</span>
                    <p className="text-slate-400 text-[10px] leading-relaxed">{v.message || `Kubernetes deprecated '${v.api_version || v.apiVersion}' for '${v.kind || v.Kind}' schemas.`}</p>
                  </div>
                  <div className="flex flex-col gap-2 shrink-0 items-end">
                    <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${v.severity === "CRITICAL" ? "bg-red-500/10 text-red-400 border border-red-500/20" : "bg-amber-500/10 text-amber-400 border border-amber-500/20"}`}>{v.severity || "WARNING"}</span>
                    <div className="text-[10px] text-slate-500">Fix: <code className="text-emerald-400 font-bold font-mono">{v.suggested_fix || v.suggestedFix}</code></div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-4 rounded bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-mono">
              Manifest completely healthy. All apiVersions match contemporary Kubernetes standards.
            </div>
          )}
        </div>
      )}
    </div>
  );
}
