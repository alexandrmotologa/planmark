import React from 'react';
import { Session } from '../types';
import { 
  FileText, 
  GitCommit, 
  Layout, 
  GitFork, 
  Send, 
  CheckCircle2, 
  XCircle, 
  Code2, 
  Bot 
} from 'lucide-react';

interface NavbarProps {
  session: Session | null;
  onOpenPacket: () => void;
  onSubmit: (status: 'approved' | 'changes_requested' | 'rejected') => void;
  annotationCount: number;
}

export const Navbar: React.FC<NavbarProps> = ({
  session,
  onOpenPacket,
  onSubmit,
  annotationCount,
}) => {
  const getArtifactIcon = (type?: string) => {
    switch (type) {
      case 'markdown':
        return <FileText className="w-4 h-4 text-blue-400" />;
      case 'diff':
        return <GitCommit className="w-4 h-4 text-emerald-400" />;
      case 'html':
        return <Layout className="w-4 h-4 text-purple-400" />;
      case 'mermaid':
        return <GitFork className="w-4 h-4 text-amber-400" />;
      default:
        return <FileText className="w-4 h-4 text-gray-400" />;
    }
  };

  return (
    <header className="h-14 border-b border-gray-800 bg-[#0c1220] px-4 flex items-center justify-between sticky top-0 z-40">
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 font-bold text-slate-100 text-base tracking-tight">
          <span className="w-2.5 h-2.5 rounded-full bg-blue-500 shadow-sm shadow-blue-500/50 animate-pulse"></span>
          <span>PlanMark</span>
          <span className="text-xs font-mono font-medium px-1.5 py-0.5 rounded bg-blue-950/80 text-blue-400 border border-blue-800/50">
            Studio
          </span>
        </div>

        {session && (
          <div className="flex items-center gap-2 pl-3 border-l border-gray-800">
            <span className="flex items-center gap-1.5 text-xs text-gray-400 bg-gray-900 px-2 py-1 rounded border border-gray-800">
              {getArtifactIcon(session.artifact.type)}
              <span className="font-medium text-slate-200 capitalize">{session.artifact.type}</span>
            </span>

            <span className="text-xs text-gray-300 font-mono max-w-[280px] truncate" title={session.artifact.filePath || session.title}>
              {session.artifact.filePath || session.title}
            </span>

            <span className="flex items-center gap-1 text-[11px] text-gray-400 bg-slate-900/60 px-2 py-0.5 rounded border border-slate-800">
              <Bot className="w-3 h-3 text-cyan-400" />
              <span>{session.agent}</span>
            </span>
          </div>
        )}
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={onOpenPacket}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 bg-gray-900 hover:bg-gray-800 hover:text-white rounded border border-gray-700 transition"
          title="Preview the generated surgical patch blocks"
        >
          <Code2 className="w-3.5 h-3.5 text-blue-400" />
          <span>Surgical Patch</span>
          {annotationCount > 0 && (
            <span className="px-1.5 py-0.2 rounded-full bg-blue-600 text-[10px] font-mono text-white">
              {annotationCount}
            </span>
          )}
        </button>

        <button
          onClick={() => onSubmit('rejected')}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-rose-300 bg-rose-950/40 hover:bg-rose-900/50 rounded border border-rose-800/60 transition"
        >
          <XCircle className="w-3.5 h-3.5 text-rose-400" />
          <span>Reject</span>
        </button>

        <button
          onClick={() => onSubmit(annotationCount > 0 ? 'changes_requested' : 'approved')}
          className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-500 rounded shadow-sm shadow-blue-600/30 transition"
        >
          {annotationCount > 0 ? (
            <>
              <Send className="w-3.5 h-3.5" />
              <span>Submit Feedback ({annotationCount})</span>
            </>
          ) : (
            <>
              <CheckCircle2 className="w-3.5 h-3.5" />
              <span>Approve Artifact</span>
            </>
          )}
        </button>
      </div>
    </header>
  );
};
