import React from 'react';
import { X, Keyboard } from 'lucide-react';

interface ShortcutsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ShortcutsModal: React.FC<ShortcutsModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) return null;

  const shortcuts = [
    { key: '?', desc: 'Show / hide keyboard shortcuts' },
    { key: 'p', desc: 'Toggle surgical patch feedback drawer' },
    { key: 'Ctrl + Enter', desc: 'Open feedback submit & dispatch dialog' },
    { key: 'Esc', desc: 'Close any active modal or drawer' },
    { key: '1 / 2 / 3', desc: 'Switch viewport (Mobile / Tablet / Desktop) in UI mode' },
    { key: 'u / s', desc: 'Toggle Unified vs. Split view in Git Diff mode' },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 animate-in fade-in duration-150">
      <div className="w-full max-w-md bg-[#0d1424] border border-gray-800 rounded-xl shadow-2xl p-6 text-slate-200">
        <div className="flex items-center justify-between pb-3 mb-4 border-b border-gray-800">
          <div className="flex items-center gap-2">
            <Keyboard className="w-5 h-5 text-blue-400" />
            <h3 className="font-semibold text-sm text-slate-100">Keyboard Shortcuts</h3>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="space-y-2 text-xs">
          {shortcuts.map((s) => (
            <div
              key={s.key}
              className="flex items-center justify-between py-1.5 px-2.5 rounded bg-[#11192e] border border-gray-800/80"
            >
              <span className="text-gray-300">{s.desc}</span>
              <kbd className="px-2 py-0.5 rounded bg-gray-800 font-mono text-[11px] font-semibold text-blue-300 border border-gray-700 shadow-sm">
                {s.key}
              </kbd>
            </div>
          ))}
        </div>

        <div className="mt-5 pt-3 border-t border-gray-800 flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-medium transition"
          >
            Got it
          </button>
        </div>
      </div>
    </div>
  );
};
