import React, { useState } from 'react';
import { SessionStatus } from '../types';
import { CheckCircle2, AlertTriangle, XCircle, Send, X } from 'lucide-react';

interface SubmitModalProps {
  isOpen: boolean;
  onClose: () => void;
  status: SessionStatus;
  annotationCount: number;
  onConfirm: (summary: string) => void;
}

export const SubmitModal: React.FC<SubmitModalProps> = ({
  isOpen,
  onClose,
  status,
  annotationCount,
  onConfirm,
}) => {
  const [summary, setSummary] = useState('');

  if (!isOpen) return null;

  const getStatusIcon = () => {
    switch (status) {
      case 'approved':
        return <CheckCircle2 className="w-5 h-5 text-emerald-400" />;
      case 'rejected':
        return <XCircle className="w-5 h-5 text-rose-400" />;
      default:
        return <AlertTriangle className="w-5 h-5 text-amber-400" />;
    }
  };

  const getStatusTitle = () => {
    switch (status) {
      case 'approved':
        return 'Approve Artifact';
      case 'rejected':
        return 'Reject Artifact';
      default:
        return `Submit Changes Requested (${annotationCount} annotations)`;
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConfirm(summary);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 animate-in fade-in duration-150">
      <div className="w-full max-w-md bg-[#0d1424] border border-gray-800 rounded-xl shadow-2xl p-6 text-slate-200">
        <div className="flex items-center justify-between pb-3 mb-3 border-b border-gray-800">
          <div className="flex items-center gap-2">
            {getStatusIcon()}
            <h3 className="font-semibold text-sm text-slate-100">{getStatusTitle()}</h3>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition">
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 text-xs">
          <div>
            <label className="block text-gray-400 mb-1.5 font-medium">
              Reviewer Summary (Sent directly to the AI agent):
            </label>
            <textarea
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              rows={3}
              autoFocus
              className="w-full bg-[#080d1a] border border-gray-700 rounded-lg p-2.5 text-xs text-slate-100 focus:outline-none focus:border-blue-500 font-sans"
              placeholder={
                status === 'approved'
                  ? 'Plan meets all requirements, proceed to execution.'
                  : 'Apply the requested surgical patches and re-submit for verification.'
              }
            />
          </div>

          <div className="bg-[#121b2f] p-3 rounded-lg border border-gray-800 text-[11px] text-gray-400 space-y-1">
            <p>
              • All visual annotations will be compiled into deterministic surgical patch blocks.
            </p>
            <p>
              • The calling agent process will receive the feedback via standard I/O or webhook.
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-3.5 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300 font-medium transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              className={`px-4 py-1.5 rounded-lg font-medium text-white flex items-center gap-1.5 transition ${
                status === 'approved'
                  ? 'bg-emerald-600 hover:bg-emerald-500'
                  : status === 'rejected'
                  ? 'bg-rose-600 hover:bg-rose-500'
                  : 'bg-blue-600 hover:bg-blue-500'
              }`}
            >
              <Send className="w-3.5 h-3.5" />
              <span>Confirm & Dispatch</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
