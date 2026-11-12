import React from 'react';
import { Session } from '../types';
import { FileText, GitCommit, Layout, GitFork, Plus } from 'lucide-react';

interface SessionTabsProps {
  sessions: Session[];
  activeSessionId: string | null;
  onSelectSession: (id: string) => void;
  onNewSession?: () => void;
}

export const SessionTabs: React.FC<SessionTabsProps> = ({
  sessions,
  activeSessionId,
  onSelectSession,
  onNewSession,
}) => {
  if (sessions.length <= 1) return null;

  const getIcon = (type: string) => {
    switch (type) {
      case 'markdown':
        return <FileText className="w-3.5 h-3.5 text-blue-400" />;
      case 'diff':
        return <GitCommit className="w-3.5 h-3.5 text-emerald-400" />;
      case 'html':
        return <Layout className="w-3.5 h-3.5 text-purple-400" />;
      case 'mermaid':
        return <GitFork className="w-3.5 h-3.5 text-amber-400" />;
      default:
        return <FileText className="w-3.5 h-3.5 text-gray-400" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'approved':
        return <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" title="Approved" />;
      case 'changes_requested':
        return <span className="w-1.5 h-1.5 rounded-full bg-amber-400" title="Changes Requested" />;
      case 'rejected':
        return <span className="w-1.5 h-1.5 rounded-full bg-rose-400" title="Rejected" />;
      default:
        return <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" title="Pending Review" />;
    }
  };

  return (
    <div className="h-10 bg-[#080d1a] border-b border-gray-800 flex items-center px-2 gap-1 overflow-x-auto select-none">
      {sessions.map((sess) => {
        const isActive = sess.id === activeSessionId;
        return (
          <button
            key={sess.id}
            onClick={() => onSelectSession(sess.id)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded text-xs transition border ${
              isActive
                ? 'bg-gray-800 text-white border-gray-700 shadow-sm'
                : 'text-gray-400 hover:text-gray-200 hover:bg-gray-900 border-transparent'
            }`}
          >
            {getIcon(sess.artifact.type)}
            <span className="font-medium max-w-[140px] truncate">{sess.title}</span>
            <span className="text-[10px] text-gray-400 font-mono">({sess.agent})</span>
            {getStatusBadge(sess.status)}
            {sess.annotations.length > 0 && (
              <span className="ml-0.5 px-1.5 py-0.2 text-[10px] rounded-full bg-blue-900 text-blue-200 border border-blue-700/50">
                {sess.annotations.length}
              </span>
            )}
          </button>
        );
      })}

      {onNewSession && (
        <button
          onClick={onNewSession}
          className="flex items-center gap-1 px-2 py-1 text-xs text-gray-400 hover:text-white hover:bg-gray-800 rounded transition border border-dashed border-gray-700 ml-1"
          title="Create New Review Session"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>New</span>
        </button>
      )}
    </div>
  );
};
